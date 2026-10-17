package minimaxlive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Enumerate in bounded batches, rather than WalkDir's eager ReadDir(-1).
// Existing own evidence needs <256 files;512/256MiB preserves ADR0036's
// import ceiling while rejecting invented work before hashing or allocation.
func Inventory(dir string) (map[string]string, error) {
	out := map[string]string{}
	entries := 0
	var total int64
	var walk func(string, int) error
	walk = func(base string, depth int) error {
		if depth > 3 {
			return fmt.Errorf("artifact directory depth exceeded")
		}
		before, e := os.Lstat(base)
		if e != nil {
			return e
		}
		if !before.IsDir() {
			return fmt.Errorf("unsafe artifact directory")
		}
		f, e := os.Open(base)
		if e != nil {
			return e
		}
		defer f.Close()
		info, e := f.Stat()
		if e != nil || !info.IsDir() || !os.SameFile(before, info) {
			return fmt.Errorf("artifact directory changed")
		}
		for {
			batch, e := f.ReadDir(64)
			for _, entry := range batch {
				entries++
				if entries > 1024 {
					return fmt.Errorf("artifact entry count exceeded")
				}
				path := filepath.Join(base, entry.Name())
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("symlink artifact forbidden")
				}
				if entry.IsDir() {
					if e := walk(path, depth+1); e != nil {
						return e
					}
					continue
				}
				name, _ := filepath.Rel(dir, path)
				name = filepath.ToSlash(name)
				if name == SealFile || name == "live-report.json" {
					continue
				}
				if len(out) >= 512 {
					return fmt.Errorf("artifact file count exceeded")
				}
				stat, e := entry.Info()
				if e != nil || !stat.Mode().IsRegular() {
					return fmt.Errorf("unsafe artifact entry")
				}
				if stat.Size() > 16<<20 || total+stat.Size() > 256<<20 {
					return fmt.Errorf("artifact aggregate bytes exceeded")
				}
				raw, e := ReadFile(path, min(16<<20, (256<<20)-total))
				if e != nil {
					return e
				}
				total += int64(len(raw))
				out[name] = Hash(raw)
			}
			if e == io.EOF {
				return nil
			}
			if e != nil {
				return e
			}
		}
	}
	e := walk(dir, 0)
	return out, e
}
