package e2e

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func TestXAIImagesProducerOriginalDocuments(t *testing.T) {
	f := newImageQualificationFixture(t)
	for _, name := range []string{"matching-original-git", "missing-original-git", "mismatched-original-git", "missing-runtime", "missing-json"} {
		t.Run(name, func(t *testing.T) {
			p := f.packet(t)
			imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) {
				files := m["source"].(map[string]any)["files"].(map[string]any)
				removed := 0
				for path := range files {
					if strings.HasSuffix(path, ".md") {
						delete(files, path)
						removed++
					}
				}
				if removed != 2 {
					t.Fatalf("expected two collector-excluded original documents, got %d", removed)
				}
				path := ""
				if name == "missing-runtime" {
					path = "ai/images.go"
				}
				if name == "missing-json" {
					path = "ai/internal/testkit/xai/images-candidate.json"
				}
				if path != "" {
					if files[path] == nil {
						t.Fatal("fixture omitted expected source input")
					}
					delete(files, path)
				}
			})
			switch name {
			case "missing-original-git":
				alternates := filepath.Join(os.Getenv("GIT_DIR"), "objects", "info", "alternates")
				raw, err := os.ReadFile(alternates)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(alternates, nil, 0600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.WriteFile(alternates, raw, 0600); err != nil {
						t.Error(err)
					}
				})
			case "mismatched-original-git":
				imageQualificationCorruptPrivateDoc(t, p)
			}
			proof, code := p.consume(t, "history-"+name)
			ev := run.Case(t, "XAIIMG-producer-history-"+name)
			want := name == "matching-original-git"
			ev.Check("only original immutable document bytes may close collector omissions", (code == 0) == want && (proof["pass"] == true) == want, "exit=%d pass=%v", code, proof["pass"])
			if want {
				inputs := proof["inputs"].(map[string]any)
				prefix := "@xai-original-source/724e988132ac5bf6ac393db156e07081eb0e2427/"
				ev.Check("both original Git documents are transitive inputs", len(inputs) == 59 && inputs[prefix+"ai/internal/testkit/kimi/README.md"] == "fb7f44b8a785684ffa68f13fc700060791c14644682b644d6fd7cd8b3b6858b3" && inputs[prefix+"ai/internal/testkit/regional/README.md"] == "4e04f6824e0ed6f9a51d5039684192fdc945c9f217c70cbfc15a2de1d5a1dbab", "document inputs missing or changed")
			}
		})
	}
}

// Shadow one blob only in the owned temporary Git namespace. The genuine
// shared commit/object store stays read-only, as does the current README.
func imageQualificationCorruptPrivateDoc(t *testing.T, p imageQualificationFixture) {
	t.Helper()
	commit := "724e988132ac5bf6ac393db156e07081eb0e2427"
	spec := commit + ":ai/internal/testkit/kimi/README.md"
	git := func(args ...string) ([]byte, error) {
		cmd := exec.CommandContext(p.ctx, "git", args...)
		cmd.Dir, cmd.Env = p.root, regionalChildEnvironment()
		return cmd.Output()
	}
	type object struct {
		kind string
		raw  []byte
	}
	objects := map[string]object{}
	for _, revision := range []string{commit, commit + "^{tree}", commit + ":ai", commit + ":ai/internal", commit + ":ai/internal/testkit", commit + ":ai/internal/testkit/kimi", commit + ":ai/internal/testkit/regional", commit + ":ai/internal/testkit/regional/README.md", spec} {
		id, err := git("rev-parse", revision)
		if err != nil {
			t.Fatal(err)
		}
		oid := strings.TrimSpace(string(id))
		if len(oid) != 40 {
			t.Fatal("private source object unavailable")
		}
		kind, err := git("cat-file", "-t", oid)
		if err != nil {
			t.Fatal(err)
		}
		kindText := strings.TrimSpace(string(kind))
		raw, err := git("cat-file", kindText, oid)
		if err != nil {
			t.Fatal(err)
		}
		objects[oid] = object{kindText, raw}
	}
	raw, err := git("rev-parse", spec)
	if err != nil {
		t.Fatal(err)
	}
	target := strings.TrimSpace(string(raw))
	changed := []byte("synthetic mismatched immutable document\n")
	objects[target] = object{"blob", changed}
	for oid, value := range objects {
		path := filepath.Join(os.Getenv("GIT_DIR"), "objects", oid[:2], oid[2:])
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatal("private source object unexpectedly exists")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		var compressed bytes.Buffer
		writer := zlib.NewWriter(&compressed)
		fmt.Fprintf(writer, "%s %d%c", value.kind, len(value.raw), 0)
		writer.Write(value.raw)
		writer.Close()
		if err := os.WriteFile(path, compressed.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Remove(path); err != nil {
				t.Error(err)
			}
		})
	}
	// Git can prefer packed alternate objects over a loose local object. Remove
	// that read-only alternate after retaining just this path's genuine objects.
	alternates := filepath.Join(os.Getenv("GIT_DIR"), "objects", "info", "alternates")
	original, err := os.ReadFile(alternates)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(alternates, nil, 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(alternates, original, 0600); err != nil {
			t.Error(err)
		}
	})
	got, err := git("cat-file", "blob", spec)
	ev := run.Case(t, "XAIIMG-producer-history-private-mismatch")
	ev.Check("private Git returns different actual document bytes", err == nil && bytes.Equal(got, changed), "private corruption setup failed: %v", err)
	ev.Record("private-document", map[string]any{"sha256": evidence.Digest(got), "notActualLive": true})
}
