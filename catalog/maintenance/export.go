package maintenance

import (
	"os"
	"path/filepath"
)

// Export returns only the selected approved revision. Source export supports
// review and deterministic regeneration without editing generated JSON by hand.
func Export(store string, sourceState bool) ([]byte, error) {
	current, err := ReadCurrent(store)
	if err != nil {
		return nil, err
	}
	if sourceState {
		return jsonBytes(current.State)
	}
	return append([]byte(nil), current.Snapshot...), nil
}

// WriteGenerated writes a complete artifact from an explicitly selected
// maintenance state. It generates data, not a check or support qualification.
func WriteGenerated(state State, path string) error {
	raw, err := Generate(state)
	if err != nil {
		return err
	}
	return atomicReplace(path, raw)
}

// atomicReplace is for replaceable exports/current pointers. Immutable
// candidates and historical revisions instead use exclusive writeImmutable.
func atomicReplace(path string, raw []byte) error {
	_, err := replaceFile(path, raw)
	return err
}

// installed remains true when rename succeeded but directory durability failed.
// Callers committing more than one path must restore installed bytes on error.
func replaceFile(path string, raw []byte) (installed bool, err error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".catalog-pending-*")
	if err != nil {
		return false, failure("storage_failed", "create pending output: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(raw); err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return false, failure("storage_failed", "save pending output: %v", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return false, failure("storage_failed", "install output: %v", err)
	}
	return true, syncDirectory(dir)
}
