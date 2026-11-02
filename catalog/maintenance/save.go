package maintenance

import (
	"bytes"
	"os"
	"path/filepath"
)

// Freeze saves a confirmed plan under its content identity. It cannot apply a
// plan, overwrite another object, or change the current revision pointer.
func Freeze(store string, proposal Proposal, confirmed bool) (Candidate, error) {
	if !confirmed {
		return Candidate{}, failure("confirmation_required", "explicit confirmation is required")
	}
	if proposal.Candidate == nil {
		return Candidate{}, failure("unresolved_definition", "proposal has no complete candidate")
	}
	candidate := *proposal.Candidate
	if err := validateCandidate(candidate); err != nil {
		return Candidate{}, err
	}
	current, err := ReadCurrent(store)
	if err != nil {
		return Candidate{}, err
	}
	if candidate.BaseDigest != current.ID {
		return Candidate{}, failure("baseline_changed", "preview baseline is no longer current")
	}
	name, err := objectName(candidate.ID)
	if err != nil {
		return Candidate{}, err
	}
	raw, err := jsonBytes(candidate)
	if err != nil {
		return Candidate{}, err
	}
	if err := writeImmutable(filepath.Join(store, "candidates", name+".json"), raw); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

// writeImmutable installs fully written bytes using an exclusive hard link.
// A crash may leave a temporary file, never a partially readable object at its
// final name. Identical objects are idempotent; different bytes never replace.
func writeImmutable(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return failure("storage_failed", "create object directory: %v", err)
	}
	if info, err := os.Lstat(dir); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return failure("storage_failed", "object directory is not a local directory")
	}
	tmp, err := os.CreateTemp(dir, ".pending-*")
	if err != nil {
		return failure("storage_failed", "create pending object: %v", err)
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
		return failure("storage_failed", "save pending object: %v", err)
	}
	if err = os.Link(tmp.Name(), path); err != nil {
		if os.IsExist(err) {
			existing, readErr := os.ReadFile(path)
			if readErr == nil && bytes.Equal(existing, raw) {
				return nil
			}
			return failure("object_conflict", "immutable object already exists with different bytes")
		}
		return failure("storage_failed", "install immutable object: %v", err)
	}
	return syncDirectory(dir)
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return failure("storage_failed", "open object directory: %v", err)
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return failure("storage_failed", "sync object directory: %v", err)
	}
	return nil
}
