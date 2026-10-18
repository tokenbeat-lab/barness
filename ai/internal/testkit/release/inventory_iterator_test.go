package release

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

// Explicitly approved isolated boundary: OS readdir duplicate interleavings
// cannot be forced portably by the public path-only loader. The finite script
// reproduces the previously unchecked repeated entry without a timing gamble.
type inventoryScript struct {
	fs.File
	entries []fs.DirEntry
}

func (s *inventoryScript) ReadDir(int) ([]fs.DirEntry, error) {
	if len(s.entries) == 0 {
		return nil, io.EOF
	}
	next := s.entries[0]
	s.entries = s.entries[1:]
	return []fs.DirEntry{next}, nil
}

func TestInventoryIteratorRejectsDuplicateNames(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		fsys := fstest.MapFS{"a.json": {Data: []byte("a")}, "b.json": {Data: []byte("b")}}
		a, _ := fs.Stat(fsys, "a.json")
		b, _ := fs.Stat(fsys, "b.json")
		entries := []fs.DirEntry{fs.FileInfoToDirEntry(a), fs.FileInfoToDirEntry(b)}
		if duplicate {
			entries[1] = entries[0]
		}
		calls, count := 0, 0
		err := scanInventoryEntries(&inventoryScript{entries: entries}, "", &count, map[string]bool{}, func(fs.DirEntry) error { calls++; return nil })
		if duplicate && (err == nil || calls != 1) {
			t.Fatalf("duplicate stream counted as two distinct names: error=%v visits=%d", err, calls)
		}
		if !duplicate && (err != nil || calls != 2) {
			t.Fatalf("unique finite stream refused: error=%v visits=%d", err, calls)
		}
	}
}
