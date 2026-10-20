package supportmatrix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Receipts belong to a new evidence location, outside every input tree. All
// existing parents must be real directories: a symlink cannot redirect a
// receipt into production data or mutate an accepted original.
func regionalPromotionPath(root, path string, inputs []string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	evidence := filepath.Join(root, ".evidence/provider-sites/final-acceptance-20261020")
	within := func(base, name string) bool {
		rel, err := filepath.Rel(base, name)
		return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
	}
	if path == evidence || !within(evidence, path) {
		return "", refuse("regional receipt requires a new current acceptance .evidence location")
	}
	for _, input := range inputs {
		base, err := filepath.Abs(input)
		if err != nil || within(base, path) {
			return "", refuse("regional receipt cannot mutate a candidate bundle")
		}
	}
	parent := filepath.Dir(path)
	for at := parent; ; at = filepath.Dir(at) {
		st, err := os.Lstat(at)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return "", refuse("regional receipt parent must be an existing real directory")
		}
		if at == filepath.Dir(at) {
			break
		}
	}
	return path, nil
}

func regionalPromotionReceipt(path string, raw []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	n, err := f.Write(raw)
	if err == nil && n != len(raw) {
		err = fmt.Errorf("regional promotion receipt incomplete")
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return regionalPromotionSync(filepath.Dir(path))
}

func regionalPromotionSync(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func regionalPromotionReplace(path string, raw []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".regional-activation-")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	n, err := temp.Write(raw)
	if err == nil && n != len(raw) {
		err = fmt.Errorf("regional activation data incomplete")
	}
	if err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return err
	}
	return regionalPromotionSync(filepath.Dir(path))
}
