package livegrant

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Consume durably reserves the caller's existing claim identities. Typed route
// contracts own scope validation; all writers share this single crash boundary.
// Every identity is reserved before writing. Any failure remains consumed.
func Consume(store string, names []string, data []byte) error {
	if len(names) == 0 {
		return fmt.Errorf("claim identities missing")
	}
	if err := os.MkdirAll(store, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(store)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("claim store must be a directory")
	}
	root, err := os.OpenRoot(store)
	if err != nil {
		return err
	}
	defer root.Close()
	lock, err := root.OpenFile("lock", os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	var files []*os.File
	for _, name := range names {
		if name == "." || name == "lock" || filepath.Base(name) != name {
			return fmt.Errorf("invalid claim identity")
		}
		f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			return fmt.Errorf("approval already consumed or unavailable")
		}
		defer f.Close()
		files = append(files, f)
	}
	for _, f := range files {
		n, err := f.Write(data)
		if err == nil && n != len(data) {
			err = fmt.Errorf("incomplete claim")
		}
		if err == nil {
			err = f.Sync()
		}
		closed := f.Close()
		if err != nil {
			return err
		}
		if closed != nil {
			return closed
		}
	}
	for _, path := range []string{store, filepath.Dir(store)} {
		dir, err := os.Open(path)
		if err != nil {
			return err
		}
		err = dir.Sync()
		closed := dir.Close()
		if err != nil {
			return err
		}
		if closed != nil {
			return closed
		}
	}
	return nil
}
