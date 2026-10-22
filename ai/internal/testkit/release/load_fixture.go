package release

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed testdata/minimax-load.jsonl.gz
var miniMaxLoadFixture []byte

// LoadMiniMaxDesignLoadFixture imports retained controlled bytes into an empty
// caller-owned directory. It proves consumer behavior, never present capacity.
// The origin and replacement conditions live next to the fixture.
func LoadMiniMaxDesignLoadFixture(dir string) (Bundle, error) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		return Bundle{}, fmt.Errorf("fixture destination must be empty")
	}
	stream, err := gzip.NewReader(bytes.NewReader(miniMaxLoadFixture))
	if err != nil {
		return Bundle{}, err
	}
	defer stream.Close()
	decoder := json.NewDecoder(io.LimitReader(stream, 128<<20))
	count := 0
	for {
		var record struct {
			Path string `json:"path"`
			Data []byte `json:"data"`
		}
		err := decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Bundle{}, err
		}
		count++
		if count > 512 || !canonicalArtifactPath(record.Path) || len(record.Data) > 16<<20 {
			return Bundle{}, fmt.Errorf("fixture inventory exceeds bounds")
		}
		path := filepath.Join(dir, filepath.FromSlash(record.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return Bundle{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return Bundle{}, err
		}
		_, err = file.Write(record.Data)
		closeErr := file.Close()
		if err != nil {
			return Bundle{}, err
		}
		if closeErr != nil {
			return Bundle{}, closeErr
		}
	}
	return LoadBundle(dir)
}
