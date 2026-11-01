package maintenance

import (
	"os"
	"path/filepath"
	"strings"
)

// An export cannot replace the store's immutable objects/current pointer,
// retained evidence, checked-in maintenance source, or historical snapshots.
// Resolve existing parent links so an alias cannot bypass these concrete paths.
func checkApplyOutput(request ApplyRequest) error {
	if request.OutputPath == "" {
		return nil
	}
	output, err := canonicalCheckPath(request.OutputPath)
	if err != nil {
		return failure("storage_failed", "resolve output: %v", err)
	}
	store, err := canonicalCheckPath(request.Store)
	if err != nil {
		return failure("storage_failed", "resolve maintenance store: %v", err)
	}
	root, err := canonicalCheckPath(request.RepositoryRoot)
	if err != nil {
		return failure("storage_failed", "resolve repository: %v", err)
	}
	for _, protected := range []string{store, filepath.Join(root, ".evidence"), filepath.Join(root, "catalog", "data"), filepath.Join(root, "ai", "release")} {
		if pathWithin(protected, output) {
			return failure("invalid_argument", "runtime output overlaps protected source, store, evidence, or history")
		}
	}
	return nil
}
func pathWithin(base, path string) bool {
	relative, err := filepath.Rel(base, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
func canonicalCheckPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	current := absolute
	missing := []string{}
	for {
		_, err = os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", err
			}
			for n := len(missing) - 1; n >= 0; n-- {
				resolved = filepath.Join(resolved, missing[n])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		missing = append(missing, filepath.Base(current))
		next := filepath.Dir(current)
		if next == current {
			return "", err
		}
		current = next
	}
}
