package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Retain immutable identities/hashes, never unchecked paths or long-lived FDs.
// Every later consumer reopens the same bundle inode and authenticates the
// actual bounded artifact bytes; a post-load filesystem mutation cannot borrow
// the initial inventory verdict. No generated evidence is copied or rewritten.
type artifactBundle struct {
	dir            string
	info, pathInfo os.FileInfo
}

type verifiedCase struct {
	bundle    *artifactBundle
	dir       string
	artifacts map[string]string
	verified  bool
}

func openArtifactBundle(dir string) (*artifactBundle, *os.Root, error) {
	before, err := os.Lstat(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("release: evidence bundle unavailable")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("release: evidence bundle unavailable")
	}
	info, e1 := root.Stat(".")
	after, e2 := os.Lstat(dir)
	if e1 != nil || e2 != nil || !info.IsDir() || !os.SameFile(before, after) {
		root.Close()
		return nil, nil, fmt.Errorf("release: evidence bundle changed")
	}
	return &artifactBundle{dir: dir, info: info, pathInfo: before}, root, nil
}

func (b *artifactBundle) open() (*os.Root, error) {
	current, root, err := openArtifactBundle(b.dir)
	if err != nil {
		return nil, err
	}
	if !os.SameFile(b.info, current.info) || !os.SameFile(b.pathInfo, current.pathInfo) {
		root.Close()
		return nil, fmt.Errorf("release: evidence bundle identity changed")
	}
	return root, nil
}

func (c verifiedCase) read(name string, budget int64) ([]byte, error) {
	if !c.verified || c.bundle == nil || !canonicalArtifactPath(name) || c.artifacts[name] == "" {
		return nil, fmt.Errorf("release: artifact is not verified")
	}
	root, err := c.bundle.open()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	child, err := openInventoryCase(root, c.dir)
	if err != nil {
		return nil, err
	}
	defer func() { child.Close() }()
	// Open each artifact ancestor too; os.Root contains opens while Lstat
	// rejects internal symlinks, not just links escaping the selected case.
	parts := strings.Split(name, "/")
	for _, part := range parts[:len(parts)-1] {
		next, e := openInventoryDirectory(child, part)
		if e != nil {
			return nil, e
		}
		child.Close()
		child = next
	}
	var raw bytes.Buffer
	if _, err := copyInventoryArtifact(child, parts[len(parts)-1], c.artifacts[name], budget, &raw); err != nil {
		return nil, err
	}
	return raw.Bytes(), nil
}

func (c verifiedCase) json(name string, budget int64, value any) error {
	raw, err := c.read(name, budget)
	if err != nil || json.Unmarshal(raw, value) != nil {
		return fmt.Errorf("%s: required authenticated JSON missing or invalid", name)
	}
	return nil
}

// The immutable declared inventory owns these names. A nested child's file
// cannot satisfy a parent's root-level capture, result or differential proof.
func (c verifiedCase) matches(pattern string) []string {
	var out []string
	if !c.verified || c.bundle == nil {
		return out
	}
	for name := range c.artifacts {
		if strings.Contains(name, "/") {
			continue
		}
		if ok, _ := filepath.Match(pattern, name); ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}
