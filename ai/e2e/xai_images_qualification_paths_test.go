package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestXAIImagesProducerPaths(t *testing.T) {
	f := newImageQualificationFixture(t)
	for _, name := range []string{"absolute-binary-relative-bundle", "absolute-binary-absolute-bundle", "absolute-binary-native-consumer", "outside-binary", "symlink-binary"} {
		t.Run(name, func(t *testing.T) {
			p := imageQualificationPathPacket(t, f)
			binary := filepath.Join(p.attempt, "synthetic-producer")
			switch name {
			case "outside-binary":
				outside := filepath.Join(t.TempDir(), "synthetic-producer")
				bytes, err := os.ReadFile(binary)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(outside, bytes, 0700); err != nil {
					t.Fatal(err)
				}
				rel, _ := filepath.Rel(p.root, outside)
				imageQualificationSetArgv(t, p, outside, filepath.ToSlash(rel))
			case "symlink-binary":
				link := filepath.Join(p.attempt, "producer-link")
				if err := os.Symlink(binary, link); err != nil {
					t.Fatal(err)
				}
				rel, _ := filepath.Rel(p.root, link)
				imageQualificationSetArgv(t, p, link, filepath.ToSlash(rel))
			}
			ev := run.Case(t, "XAIIMG-producer-path-"+name)
			if name == "absolute-binary-native-consumer" {
				proof, code := p.consume(t, "absolute-producer")
				ev.Check("native consumer accepts retained repository-local absolute argv", code == 0 && proof["pass"] == true, "exit=%d pass=%v", code, proof["pass"])
				return
			}
			arg, _ := filepath.Rel(p.root, p.bundle)
			if name == "absolute-binary-absolute-bundle" {
				arg = p.bundle
			}
			cmd := exec.CommandContext(p.ctx, p.binary, "-inspect-bundle", filepath.ToSlash(arg), "-mode", "live", "-stage", "formal")
			cmd.Dir, cmd.Env = p.root, regionalChildEnvironment()
			raw, err := cmd.CombinedOutput()
			code := imagineProcessExit(err)
			want := name == "absolute-binary-relative-bundle" || name == "absolute-binary-absolute-bundle"
			ev.Record("process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(raw), "vendorCalls": 0})
			ev.Check("only retained repository-local regular executable qualifies", (code == 0) == want, "exit=%d want=%v", code, want)
		})
	}
}

// Absolute argv is necessary provenance, but it is not useful to copy the
// host path into the audited main bundle. Keep the original private fixture
// beside that bundle and record its repository-relative inventory and hashes.
func imageQualificationPathPacket(t *testing.T, f imageQualificationFixture) imageQualificationFixture {
	t.Helper()
	p := f.packet(t)
	old := p.attempt
	destination := filepath.Join(filepath.Dir(filepath.Dir(run.Dir())), "path-"+filepath.Base(old))
	if err := os.Rename(old, destination); err != nil {
		t.Fatal(err)
	}
	oldRel, _ := filepath.Rel(p.root, old)
	newRel, _ := filepath.Rel(p.root, destination)
	p.attempt = destination
	p.bundle = filepath.Join(destination, "bundles", filepath.Base(p.bundle))
	imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) {
		selected := map[string]any{}
		for name, h := range m["selectedInputs"].(map[string]any) {
			suffix := strings.TrimPrefix(name, filepath.ToSlash(oldRel)+"/")
			if suffix == name {
				t.Fatal("fixture selected input escaped its private parent")
			}
			selected[filepath.ToSlash(newRel)+"/"+suffix] = h
		}
		m["selectedInputs"] = selected
	})
	imageQualificationSetArgv(t, p, filepath.Join(p.attempt, "synthetic-producer"), "")
	inventory, err := liveSupplementInventory(t, p.attempt)
	if err != nil {
		t.Fatal(err)
	}
	ev := run.Case(t, "XAIIMG-producer-path-fixture-"+filepath.Base(destination))
	ev.Record("retained-fixture", map[string]any{"path": filepath.ToSlash(newRel), "inputs": inventory, "notActualLive": true})
	return p
}
func imageQualificationSetArgv(t *testing.T, p imageQualificationFixture, binary, selected string) {
	t.Helper()
	command := filepath.Join(p.attempt, "command.json")
	imageQualificationMutate(t, command, func(m map[string]any) { m["argv"].([]any)[0] = binary })
	imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) {
		m["argv"].([]any)[0] = binary
		hash := imageQualificationHash(t, command)
		m["files"].(map[string]any)["command.json"] = hash
		m["commandSHA256"] = hash
		if selected != "" {
			inputs := m["selectedInputs"].(map[string]any)
			for name, h := range inputs {
				if filepath.Base(name) == "synthetic-producer" {
					delete(inputs, name)
					inputs[selected] = h
					break
				}
			}
		}
	})
}
