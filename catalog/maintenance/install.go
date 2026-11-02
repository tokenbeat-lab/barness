package maintenance

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// installRevision keeps the existing bounded local rollback model. Rename and
// durability are separate outcomes: a failed fsync must not hide installed bytes.
// Immutable revision orphans are deliberately retained for diagnosis/retry.
func installRevision(store string, revision Revision, output string) (Revision, error) {
	var old []byte
	existed := false
	if output != "" {
		var err error
		old, err = os.ReadFile(output)
		if err == nil {
			existed = true
		} else if !os.IsNotExist(err) {
			return Revision{}, failure("storage_failed", "read output before apply: %v", err)
		}
		installed, err := replaceFile(output, revision.Snapshot)
		if err != nil {
			if installed {
				_, restore := restoreFile(output, old, existed)
				return Revision{}, rollbackFailure("runtime output", err, restore)
			}
			return Revision{}, err
		}
	}
	path := filepath.Join(store, "current.json")
	prior, err := os.ReadFile(path)
	hadPointer := err == nil
	if err != nil && !os.IsNotExist(err) {
		_, restore := restoreOutput(output, old, existed)
		return Revision{}, rollbackFailure("read current pointer", err, restore)
	}
	pointer, _ := json.Marshal(map[string]string{"revision": revision.ID})
	installed, err := replaceFile(path, append(pointer, '\n'))
	if err == nil {
		return revision, nil
	}
	var pointerRestore error
	if installed {
		restored, restore := restoreFile(path, prior, hadPointer)
		pointerRestore = restore
		if !restored {
			// The pointer still selects the new snapshot. Keep its matching output and
			// explicitly require repair instead of returning a false unchanged claim.
			return Revision{}, failure("storage_failed", "current pointer installation failed: %v; rollback did not restore the baseline; new revision %s remains selected and requires repair: %v", err, revision.ID, restore)
		}
	}
	_, outputRestore := restoreOutput(output, old, existed)
	if pointerRestore != nil || outputRestore != nil {
		return Revision{}, failure("storage_failed", "current pointer installation failed: %v; rollback durability or repair required (pointer: %v; output: %v)", err, pointerRestore, outputRestore)
	}
	return Revision{}, err
}

func restoreOutput(path string, raw []byte, existed bool) (bool, error) {
	if path == "" {
		return true, nil
	}
	return restoreFile(path, raw, existed)
}

// restored says whether the old visible state was installed, independent of
// durability. Errors remain diagnostic even when visible bytes were restored.
func restoreFile(path string, raw []byte, existed bool) (restored bool, err error) {
	if existed {
		return replaceFile(path, raw)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return true, syncDirectory(filepath.Dir(path))
}

func rollbackFailure(stage string, original, restore error) error {
	if restore != nil {
		return failure("storage_failed", "%s installation failed: %v; rollback durability or repair required: %v", stage, original, restore)
	}
	return failure("storage_failed", "%s installation failed and prior bytes restored: %v", stage, original)
}
