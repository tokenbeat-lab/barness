package minimaxlive

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
)

const maxCohortFile = int64(64 << 20)
const maxCohortBytes = int64(512 << 20)

func cohortOpen(root *os.Root, name string, limit int64) (*os.File, os.FileInfo, error) {
	parts := strings.Split(name, "/")
	if !strings.HasPrefix(name, ".evidence/") || filepath.IsAbs(name) || filepath.ToSlash(filepath.Clean(name)) != name || strings.Contains(name, "\\") || len(parts) > 32 {
		return nil, nil, fmt.Errorf("closed cohort requires bounded owned evidence paths")
	}
	for i, part := range parts {
		if part == ".env" || part == ".." {
			return nil, nil, fmt.Errorf("closed cohort secret/parent paths forbidden")
		}
		info, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if err != nil || info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() {
			return nil, nil, fmt.Errorf("closed cohort path missing or linked")
		}
	}
	before, err := root.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Size() > limit {
		return nil, nil, fmt.Errorf("closed cohort file missing or exceeds byte ceiling")
	}
	file, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, nil, err
	}
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) || before.Size() != opened.Size() {
		file.Close()
		return nil, nil, fmt.Errorf("closed cohort file changed before read")
	}
	return file, before, nil
}

// Keep the release inventory's stable-file streaming checks here: release
// consumes MiniMax evidence, so importing its private helper would create a
// cycle. Hashing the old binaries must never use the small JSON import limit.
func cohortFileHash(root *os.Root, name string, remaining int64) (string, int64, error) {
	file, before, err := cohortOpen(root, name, min(maxCohortFile, remaining))
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(file, before.Size()+1))
	after, statErr := root.Lstat(name)
	if err != nil || statErr != nil || n != before.Size() || !after.Mode().IsRegular() || !os.SameFile(before, after) || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return "", 0, fmt.Errorf("closed cohort changed during bounded hash")
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func cohortJSON(root *os.Root, name string, value any) ([]byte, error) {
	return cohortJSONLimit(root, name, value, 2<<20)
}

func cohortJSONLimit(root *os.Root, name string, value any, limit int64) ([]byte, error) {
	file, before, err := cohortOpen(root, name, limit)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, before.Size()+1))
	after, statErr := root.Lstat(name)
	if err != nil || statErr != nil || int64(len(raw)) != before.Size() || !os.SameFile(before, after) || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return nil, fmt.Errorf("closed JSON changed during bounded read")
	}
	return raw, StrictJSON(raw, value)
}

func cohortArchive(root *os.Root, prefix string, files map[string]string) error {
	pending := []string{prefix}
	count, found := 0, 0
	for len(pending) > 0 {
		name := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		info, err := root.Lstat(name)
		if err != nil || !info.IsDir() || len(strings.Split(name, "/")) > 32 {
			return fmt.Errorf("closed source archive directory missing or linked")
		}
		directory, err := root.Open(name)
		if err != nil {
			return err
		}
		for {
			entries, e := directory.ReadDir(64)
			for _, entry := range entries {
				count++
				path := name + "/" + entry.Name()
				if count > 8192 || entry.Type()&os.ModeSymlink != 0 {
					directory.Close()
					return fmt.Errorf("closed source archive entry ceiling or link")
				}
				if entry.IsDir() {
					pending = append(pending, path)
				} else if files[path] == "" {
					directory.Close()
					return fmt.Errorf("closed source archive has an unpinned file")
				} else {
					found++
				}
			}
			if e != nil {
				directory.Close()
				if e != io.EOF {
					return e
				}
				break
			}
		}
	}
	want := 0
	for name := range files {
		if strings.HasPrefix(name, prefix+"/") {
			want++
		}
	}
	if found == 0 || found != want {
		return fmt.Errorf("closed source archive selection differs")
	}
	return nil
}

// Reports and launcher receipts carry descriptive envelopes; validate duplicate
// JSON and canonical field spellings before projecting their semantic subset.
func cohortJSONProjection(root *os.Root, name string, target any, required ...string) ([]byte, error) {
	var envelope map[string]json.RawMessage
	raw, e := cohortJSON(root, name, &envelope)
	if e == nil {
		for _, key := range required {
			if value, ok := envelope[key]; !ok || strings.TrimSpace(string(value)) == "null" {
				return raw, fmt.Errorf("required receipt field missing: %s", key)
			}
		}
		var document any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber() // Preserve typed integer rejection; never normalize 0.0 into 0.
		if e = decoder.Decode(&document); e == nil {
			var projected any
			projected, e = cohortProjection(document, reflect.TypeOf(target))
			if e == nil {
				var encoded []byte
				encoded, e = json.Marshal(projected)
				if e == nil {
					e = StrictJSON(encoded, target)
				}
			}
		}
	}
	return raw, e
}

// Receipts carry descriptive extensions outside their typed semantic subset.
// Drop only those extensions, reject aliases of known fields at every typed
// level, then use StrictJSON's existing recursive canonical/type validation.
// Map keys and custom JSON values remain data, not struct field spellings.
func cohortProjection(value any, t reflect.Type) (any, error) {
	if value == nil {
		return nil, nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return value, nil
	}
	switch t.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		out := map[string]any{}
		for name, child := range object {
			field, known := fields[name]
			if !known {
				for canonical := range fields {
					if strings.EqualFold(name, canonical) {
						return nil, fmt.Errorf("noncanonical projected receipt field: %s", name)
					}
				}
				continue
			}
			projected, e := cohortProjection(child, field)
			if e != nil {
				return nil, e
			}
			out[name] = projected
		}
		return out, nil
	case reflect.Map:
		if object, ok := value.(map[string]any); ok {
			out := map[string]any{}
			for name, child := range object {
				projected, e := cohortProjection(child, t.Elem())
				if e != nil {
					return nil, e
				}
				out[name] = projected
			}
			return out, nil
		}
	case reflect.Slice:
		if items, ok := value.([]any); ok {
			out := make([]any, len(items))
			for i, child := range items {
				projected, e := cohortProjection(child, t.Elem())
				if e != nil {
					return nil, e
				}
				out[i] = projected
			}
			return out, nil
		}
	}
	return value, nil
}
