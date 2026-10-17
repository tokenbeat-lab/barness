//go:build live

package live

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// The process E2E retains bounded raw child evidence before TempDir cleanup.
// Its original clean audit and seal are inside the archive; the outer fixture
// checksum protects the archive bytes. No vendor capture is manufactured.
func archiveMiniMaxProcess(dir string) ([]byte, error) {
	files, e := minimaxlive.Inventory(dir)
	if e != nil {
		return nil, e
	}
	if _, err := os.Lstat(filepath.Join(dir, minimaxlive.SealFile)); err == nil {
		files[minimaxlive.SealFile] = ""
	}
	if len(files) > 256 {
		return nil, fmt.Errorf("controlled child artifact count exceeded")
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	total := 0
	for _, name := range names {
		raw, e := minimaxlive.ReadFile(filepath.Join(dir, filepath.FromSlash(name)), 16<<20)
		if e != nil {
			return nil, e
		}
		total += len(raw)
		if total > 16<<20 {
			return nil, fmt.Errorf("controlled child archive exceeds bound")
		}
		if e = tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(raw))}); e != nil {
			return nil, e
		}
		if _, e = tw.Write(raw); e != nil {
			return nil, e
		}
	}
	if e = tw.Close(); e != nil {
		return nil, e
	}
	if e = gz.Close(); e != nil {
		return nil, e
	}
	return buf.Bytes(), nil
}
