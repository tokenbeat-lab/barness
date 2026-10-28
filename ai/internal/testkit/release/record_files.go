package release

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func recordPath(root, name string) (string, error) {
	if !filepath.IsLocal(name) || filepath.ToSlash(filepath.Clean(name)) != name {
		return "", fmt.Errorf("noncanonical receipt path")
	}
	current := root
	for _, part := range strings.Split(name, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", fmt.Errorf("receipt path unavailable")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symbolic receipt path rejected")
		}
	}
	return current, nil
}
func recordHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return "", fmt.Errorf("bounded regular receipt file required")
	}
	hash := sha256.New()
	n, err := io.Copy(hash, io.LimitReader(file, (64<<20)+1))
	if err != nil || n > 64<<20 {
		return "", fmt.Errorf("receipt file exceeds bound")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func verifyRecordFile(root string, ref FileRef) (string, error) {
	path, err := recordPath(root, ref.Path)
	if err != nil {
		return "", err
	}
	got, err := recordHash(path)
	if err != nil || len(ref.SHA256) != 64 || got != ref.SHA256 {
		return "", fmt.Errorf("receipt file integrity failed")
	}
	return path, nil
}
func recordJSON(path string, v any) error {
	raw, err := readArtifact(path, 16<<20)
	if err != nil || !json.Valid(raw) {
		return fmt.Errorf("invalid receipt JSON")
	}
	// Receipt envelopes have their own schema. This bounded validation deliberately
	// does not import provider grant models or their canonical-field policy.
	d := json.NewDecoder(bytes.NewReader(raw))
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 64 {
			return fmt.Errorf("excessive JSON nesting")
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		seen := map[string]bool{}
		for d.More() {
			if delim == '{' {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name := key.(string)
				if seen[name] {
					return fmt.Errorf("duplicate receipt field")
				}
				seen[name] = true
			}
			if err := walk(depth + 1); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	}
	if err := walk(0); err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	if err := canonicalRecordFields(document, reflect.TypeOf(v)); err != nil {
		return err
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func currentRecordSource(root string) (ExecutionSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	readGit := func(args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
		return cmd.Output()
	}
	commit, err := readGit("rev-parse", "HEAD")
	if err != nil {
		return ExecutionSource{}, err
	}
	out := ExecutionSource{Commit: strings.TrimSpace(string(commit)), Files: map[string]string{}}
	// Independent inventory validation is intentional: the read-only verifier
	// cannot trust or execute the collector being verified. Both use this small
	// published source boundary; prose is excluded unless explicitly selected.
	for _, scope := range []string{"ai", "tools/verification"} {
		base := filepath.Join(root, filepath.FromSlash(scope))
		if _, err := os.Lstat(base); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			switch info.Name() {
			case "node_modules", "__pycache__", ".git", ".evidence":
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symbolic execution source rejected")
			}
			if info.IsDir() {
				return nil
			}
			if info.Name() == ".DS_Store" || info.Name() == "Thumbs.db" {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".md", ".pyc", ".test", ".exe":
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			hash, err := recordHash(path)
			if err != nil {
				return err
			}
			out.Files[filepath.ToSlash(relative)] = hash
			if len(out.Files) > 200000 {
				return fmt.Errorf("execution source inventory exceeds bound")
			}
			return nil
		})
		if err != nil {
			return out, err
		}
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		path := filepath.Join(root, name)
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		}
		path, err := recordPath(root, name)
		if err != nil {
			return out, err
		}
		hash, err := recordHash(path)
		if err != nil {
			return out, err
		}
		out.Files[name] = hash
	}
	if len(out.Files) == 0 {
		return out, fmt.Errorf("empty execution source")
	}
	return out, nil
}

func sameRecordSource(a, b ExecutionSource) bool { return reflect.DeepEqual(a.Files, b.Files) }
func checkInputFiles(root string, files map[string]string) error {
	if files == nil {
		return fmt.Errorf("selected inputs must be explicit")
	}
	for name, hash := range files {
		if _, err := verifyRecordFile(root, FileRef{name, hash}); err != nil {
			return err
		}
	}
	return nil
}

// Measured inputs describe the original execution. Mutable documentation may
// have advanced; its original committed bytes still authenticate the old fact.
// Untracked grants/captures must remain available at their immutable paths.
func checkMeasuredInputFiles(root, commit string, files map[string]string) error {
	if files == nil {
		return fmt.Errorf("selected inputs must be explicit")
	}
	commitBytes, commitErr := hex.DecodeString(commit)
	for name, hash := range files {
		if _, err := verifyRecordFile(root, FileRef{name, hash}); err == nil {
			continue
		}
		bad := fmt.Errorf("original selected input unavailable or changed")
		hashBytes, err := hex.DecodeString(hash)
		if commitErr != nil || len(commitBytes) != 20 || err != nil || len(hashBytes) != 32 || !filepath.IsLocal(name) || filepath.ToSlash(filepath.Clean(name)) != name {
			return bad
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		cmd := exec.CommandContext(ctx, "git", "cat-file", "blob", commit+":"+name)
		cmd.Dir, cmd.Stderr = root, io.Discard
		for _, pair := range os.Environ() {
			if !strings.HasPrefix(pair, "GIT_") {
				cmd.Env = append(cmd.Env, pair)
			}
		}
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			cancel()
			return bad
		}
		if err := cmd.Start(); err != nil {
			cancel()
			return bad
		}
		h := sha256.New()
		n, readErr := io.Copy(h, io.LimitReader(pipe, (64<<20)+1))
		if readErr != nil || n > 64<<20 {
			cancel()
		}
		waitErr := cmd.Wait()
		cancel()
		if readErr != nil || waitErr != nil || n > 64<<20 || hex.EncodeToString(h.Sum(nil)) != hash {
			return bad
		}
	}
	return nil
}

// Field spelling is exact; encoding/json's case aliases are not a receipt contract.
func canonicalRecordFields(value any, kind reflect.Type) error {
	for kind.Kind() == reflect.Pointer {
		kind = kind.Elem()
	}
	if value == nil {
		return nil
	}
	switch object := value.(type) {
	case map[string]any:
		if kind.Kind() == reflect.Map {
			for _, v := range object {
				if err := canonicalRecordFields(v, kind.Elem()); err != nil {
					return err
				}
			}
			return nil
		}
		if kind.Kind() != reflect.Struct {
			return nil
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < kind.NumField(); i++ {
			f := kind.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		for name, v := range object {
			field, ok := fields[name]
			if !ok {
				return fmt.Errorf("unknown or noncanonical receipt field")
			}
			if err := canonicalRecordFields(v, field); err != nil {
				return err
			}
		}
	case []any:
		if kind.Kind() == reflect.Slice || kind.Kind() == reflect.Array {
			for _, v := range object {
				if err := canonicalRecordFields(v, kind.Elem()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
