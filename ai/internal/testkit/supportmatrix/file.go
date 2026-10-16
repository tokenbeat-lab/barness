package supportmatrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ReportName is the artifact name a live process gives its report inside
// its evidence bundle; ReportFile is the file it is written to.
const (
	ReportName = "live-report"
	ReportFile = ReportName + ".json"
)

// LoadMatrix reads a matrix file.
func LoadMatrix(path string) (Matrix, error) {
	var m Matrix
	return m, readJSON(path, &m)
}

// LoadReport reads a report file.
func LoadReport(path string) (Report, error) {
	var r Report
	err := readJSON(path, &r)
	r.evidenceDir = filepath.Dir(path)
	return r, err
}

// SaveMatrix writes m to path through a temporary file, so a reader never
// sees a partly written matrix.
func SaveMatrix(path string, m Matrix) error {
	data, err := Encode(m)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".support-matrix-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// CreateTemp makes the file private; the matrix is a checked-in file.
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// Encode is the files' JSON form: indented, HTML left unescaped.
func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readJSON(path string, v any) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return fmt.Errorf("nonregular or oversized report")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 16<<20 || !json.Valid(data) {
		return fmt.Errorf("invalid or oversized report JSON")
	}
	guard := json.NewDecoder(bytes.NewReader(data))
	guard.UseNumber()
	if err = uniqueFields(guard, 0); err != nil {
		return err
	}
	var document any
	shape := json.NewDecoder(bytes.NewReader(data))
	shape.UseNumber()
	if err = shape.Decode(&document); err != nil {
		return err
	}
	if err = exactFields(document, reflect.TypeOf(v)); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// Struct member spelling is contractual; domain map keys remain case-sensitive
// data. Go's ordinary decoder otherwise accepts MODEL and model as aliases.
func exactFields(value any, t reflect.Type) error {
	if value == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
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
		for name, child := range object {
			field, ok := fields[name]
			if !ok {
				return fmt.Errorf("noncanonical report JSON field")
			}
			if err := exactFields(child, field); err != nil {
				return err
			}
		}
	case reflect.Map:
		if object, ok := value.(map[string]any); ok {
			for _, child := range object {
				if err := exactFields(child, t.Elem()); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		if array, ok := value.([]any); ok {
			for _, child := range array {
				if err := exactFields(child, t.Elem()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func uniqueFields(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("report JSON depth exceeded")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
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
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("duplicate report JSON field")
			}
			seen[name] = true
		}
		if err := uniqueFields(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}
