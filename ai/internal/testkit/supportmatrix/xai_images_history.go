package supportmatrix

import (
	"context"
	"io"
	"io/fs"
	"os/exec"
	"path/filepath"
	"time"
)

// Only the frozen execution map's collector-excluded Markdown can use Git.
// Commit, canonical path and SHA-256 remain independently pinned by the native
// source/grant and producer. Use the same Git namespace as trusted claims;
// controlled E2E owns a private namespace with read-only genuine objects.
func xaiImagesOriginalDocument(root, commit, name, hash string) error {
	bad := refuse("Imagine immutable original document absent or changed")
	if commit != xaiImagesMeasuredCommit || !fs.ValidPath(name) || !filepath.IsLocal(name) || filepath.Ext(name) != ".md" || !validManifestHash(hash) {
		return bad
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", root, "cat-file", "blob", commit+":"+name)
	cmd.Stderr = io.Discard
	pipe, e := cmd.StdoutPipe()
	if e != nil || cmd.Start() != nil {
		return bad
	}
	raw, readErr := io.ReadAll(io.LimitReader(pipe, (1<<20)+1))
	if readErr != nil || len(raw) > 1<<20 {
		cancel()
	}
	waitErr := cmd.Wait()
	if readErr != nil || waitErr != nil || len(raw) > 1<<20 || glmHash(raw) != hash {
		return bad
	}
	return nil
}
