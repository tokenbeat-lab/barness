package supportmatrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	return r, readJSON(path, &r)
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
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
