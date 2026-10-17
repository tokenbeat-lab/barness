package supportmatrix

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

const glmViewMaxFiles = 512
const glmViewMaxBytes = int64(256 << 20)

// Inspect the finite owned shape before reading any view payload. Directory
// enumeration itself is bounded; WalkDir would first allocate every sibling.
func glmViewPaths(dir string, originals map[string][]byte) (map[string]string, error) {
	allowed := map[string]bool{"live-report.json": true, "glm-support-derivation.json": true, "audit.json": true, "manifest.json": true, "glm-support-integrity.json": true}
	for name := range originals {
		allowed["original/"+name] = true
	}
	directories := map[string]bool{".": true}
	for name := range allowed {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			directories[parent] = true
		}
	}
	queue := []string{"."}
	files := map[string]string{}
	var total int64
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		base := filepath.Join(dir, filepath.FromSlash(current))
		st, e := os.Lstat(base)
		if e != nil || !st.IsDir() {
			return nil, glmBad("unexpected support view artifact directory")
		}
		handle, e := os.Open(base)
		if e != nil {
			return nil, e
		}
		entries, readErr := handle.ReadDir(glmViewMaxFiles + 1)
		handle.Close()
		if readErr != nil && readErr != io.EOF {
			return nil, readErr
		}
		if len(entries) > glmViewMaxFiles {
			return nil, glmBad("unexpected support view artifact count")
		}
		for _, entry := range entries {
			rel := path.Join(current, entry.Name())
			if entry.IsDir() {
				if !directories[rel] {
					return nil, glmBad("unexpected support view artifact directory")
				}
				queue = append(queue, rel)
				continue
			}
			if !allowed[rel] || len(files) >= glmViewMaxFiles {
				return nil, glmBad("unexpected support view artifact")
			}
			info, e := entry.Info()
			if e != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > 16<<20 {
				return nil, glmBad("unexpected support view artifact size/type")
			}
			total += info.Size()
			if total > glmViewMaxBytes {
				return nil, glmBad("unexpected support view artifact aggregate bytes")
			}
			files[rel] = filepath.Join(dir, filepath.FromSlash(rel))
		}
	}
	return files, nil
}

func glmHashViewPaths(paths map[string]string) (map[string]string, error) {
	hashes := map[string]string{}
	var actualBytes int64
	for rel, p := range paths {
		if rel == "glm-support-integrity.json" {
			continue
		}
		raw, e := glmRead(p)
		if e != nil {
			return nil, e
		}
		actualBytes += int64(len(raw))
		if actualBytes > glmViewMaxBytes {
			return nil, glmBad("support view read exceeds aggregate bound")
		}
		hashes[rel] = glmHash(raw)
	}
	return hashes, nil
}

func glmViewInventory(dir string, originals map[string][]byte) (map[string]string, error) {
	paths, e := glmViewPaths(dir, originals)
	if e != nil {
		return nil, e
	}
	return glmHashViewPaths(paths)
}
