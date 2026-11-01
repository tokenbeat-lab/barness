package maintenance

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ReadCurrent never creates a store. A missing current pointer selects the
// checked-in seed; an existing but invalid pointer fails without fallback.
func ReadCurrent(store string) (Revision, error) {
	raw, err := os.ReadFile(filepath.Join(store, "current.json"))
	if os.IsNotExist(err) {
		return builtinRevision()
	}
	if err != nil {
		return Revision{}, failure("storage_failed", "read current revision: %v", err)
	}
	var pointer struct {
		Revision string `json:"revision"`
	}
	if err := decode(raw, &pointer); err != nil {
		return Revision{}, failure("storage_failed", "invalid current pointer: %v", err)
	}
	return readRevision(store, pointer.Revision)
}

func builtinRevision() (Revision, error) {
	state, err := BuiltinState()
	if err != nil {
		return Revision{}, err
	}
	snapshot, err := Generate(state)
	if err != nil {
		return Revision{}, err
	}
	id, err := stateDigest(state)
	return Revision{ID: id, State: state, Snapshot: snapshot, Change: Change{Action: "seed", Reason: "reviewed historical catalog migration"}}, err
}

func readRevision(store, id string) (Revision, error) {
	name, err := objectName(id)
	if err != nil {
		return Revision{}, err
	}
	raw, err := os.ReadFile(filepath.Join(store, "revisions", name+".json"))
	if err != nil {
		return Revision{}, failure("storage_failed", "read immutable revision: %v", err)
	}
	var revision Revision
	if err := decode(raw, &revision); err != nil {
		return Revision{}, failure("storage_failed", "invalid revision: %v", err)
	}
	digest, err := stateDigest(revision.State)
	if err != nil || digest != id || revision.ID != id {
		return Revision{}, failure("revision_changed", "revision identity does not match its state")
	}
	expected, err := Generate(revision.State)
	if err != nil || !bytes.Equal(expected, revision.Snapshot) {
		return Revision{}, failure("revision_changed", "revision snapshot differs from its approved state")
	}
	return revision, nil
}

func ReadCandidate(store, id string) (Candidate, error) {
	name, err := objectName(id)
	if err != nil {
		return Candidate{}, err
	}
	raw, err := os.ReadFile(filepath.Join(store, "candidates", name+".json"))
	if err != nil {
		return Candidate{}, failure("storage_failed", "read candidate: %v", err)
	}
	var candidate Candidate
	if err := decode(raw, &candidate); err != nil {
		return Candidate{}, failure("candidate_changed", "invalid candidate: %v", err)
	}
	if candidate.ID != id {
		return Candidate{}, failure("candidate_changed", "candidate object differs from the requested identity")
	}
	if err := validateCandidate(candidate); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

// History follows only committed immutable revisions. Orphans from a failed
// save never become effective history and remain available for diagnosis.
func History(store string) ([]Revision, error) {
	current, err := ReadCurrent(store)
	if err != nil {
		return nil, err
	}
	history := []Revision{current}
	seen := map[string]bool{current.ID: true}
	for current.Parent != "" {
		if seen[current.Parent] {
			return nil, failure("storage_failed", "revision history contains a cycle")
		}
		seen[current.Parent] = true
		current, err = readRevision(store, current.Parent)
		if err != nil {
			return nil, err
		}
		history = append(history, current)
	}
	return history, nil
}

func objectName(id string) (string, error) {
	name := strings.TrimPrefix(id, "sha256:")
	raw, err := hex.DecodeString(name)
	if err != nil || len(raw) != 32 || id != "sha256:"+name {
		return "", failure("invalid_argument", "expected a sha256 object identity")
	}
	return name, nil
}

func jsonBytes(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	return append(raw, '\n'), err
}
