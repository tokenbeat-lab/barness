package maintenance

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type ApplyRequest struct {
	Store          string
	CandidateID    string
	CheckReference string
	RepositoryRoot string
	OutputPath     string
	Confirmed      bool
	DryRun         bool
}

// VerifyCheck is read-only. It validates actual artifacts and complete source
// sets afresh; the status written in a check file is never authority.
func VerifyCheck(ctx context.Context, request ApplyRequest) (Candidate, CheckResult, error) {
	candidate, err := ReadCandidate(request.Store, request.CandidateID)
	if err != nil {
		return candidate, CheckResult{}, err
	}
	current, err := ReadCurrent(request.Store)
	if err != nil {
		return candidate, CheckResult{}, err
	}
	if current.ID != candidate.BaseDigest {
		return candidate, CheckResult{}, failure("baseline_changed", "candidate baseline is no longer current")
	}
	root, err := filepath.Abs(request.RepositoryRoot)
	if err != nil {
		return candidate, CheckResult{}, err
	}
	relative, err := checkRelative(root, request.CheckReference)
	if err != nil {
		return candidate, CheckResult{}, err
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		return candidate, CheckResult{}, failure("check_invalid", "check reference unavailable")
	}
	var check CheckResult
	if decode(raw, &check) != nil || check.Status != "applicable" || check.Conclusion != "code_compatible" || check.Reference != strings.TrimSuffix(check.Prepared, "prepared.json")+"check.json" || check.Reference != relative || check.CandidateID != candidate.ID || check.BaseDigest != candidate.BaseDigest || check.SnapshotDigest != digest(candidate.Snapshot) || check.Prepared == "" || check.Receipt == "" {
		return candidate, check, failure("check_invalid", "check does not bind this exact candidate and baseline")
	}
	inspection, err := inspectCheck(ctx, root, "", check.Prepared, check.Receipt)
	if err != nil {
		return candidate, check, err
	}
	if inspection.CandidateID != candidate.ID || inspection.BaseDigest != candidate.BaseDigest || !slices.Equal(inspection.RequiredCases, check.RequiredCases) || inspection.ReceiptSHA256 != check.ReceiptSHA256 || inspection.MeasuredCommit != check.MeasuredCommit || inspection.SnapshotSHA256 != "" && "sha256:"+inspection.SnapshotSHA256 != check.SnapshotDigest {
		return candidate, check, failure("check_invalid", "check reference binding changed")
	}
	return candidate, check, nil
}

// Apply installs the exact tested Snapshot. No source fetch or remapping is
// performed. A new immutable revision precedes the single current-pointer commit.
func Apply(ctx context.Context, request ApplyRequest) (Revision, error) {
	if !request.Confirmed && !request.DryRun {
		return Revision{}, failure("confirmation_required", "apply requires explicit --confirm")
	}
	if err := checkApplyOutput(request); err != nil {
		return Revision{}, err
	}
	candidate, _, err := VerifyCheck(ctx, request)
	if err != nil {
		return Revision{}, err
	}
	if request.DryRun {
		return Revision{CandidateID: candidate.ID, State: candidate.State, Snapshot: candidate.Snapshot, Change: candidate.Change}, nil
	}
	if err := ctx.Err(); err != nil {
		return Revision{}, failure("execution_failed", "apply cancelled before commit")
	}
	return commitCandidate(request.Store, candidate, request.CheckReference, request.OutputPath)
}

func commitCandidate(store string, candidate Candidate, checkRef, output string) (Revision, error) {
	if err := os.MkdirAll(store, 0700); err != nil {
		return Revision{}, failure("storage_failed", "create maintenance store: %v", err)
	}
	lock := filepath.Join(store, ".apply-lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		return Revision{}, failure("store_busy", "another apply owns the maintenance store")
	}
	defer os.Remove(lock)
	current, err := ReadCurrent(store)
	if err != nil {
		return Revision{}, err
	}
	frozen, err := ReadCandidate(store, candidate.ID)
	if err != nil {
		return Revision{}, err
	}
	if current.ID != candidate.BaseDigest || frozen.BaseDigest != current.ID {
		return Revision{}, failure("baseline_changed", "current revision changed before commit")
	}
	id, err := stateDigest(candidate.State)
	if err != nil {
		return Revision{}, err
	}
	revision := Revision{ID: id, Parent: current.ID, CandidateID: candidate.ID, State: candidate.State, Snapshot: append([]byte(nil), candidate.Snapshot...), Change: candidate.Change, CheckReference: checkRef}
	for _, item := range []Revision{current, revision} {
		name, err := objectName(item.ID)
		if err != nil {
			return Revision{}, err
		}
		raw, err := jsonBytes(item)
		if err != nil {
			return Revision{}, err
		}
		if err := writeImmutable(filepath.Join(store, "revisions", name+".json"), raw); err != nil {
			return Revision{}, err
		}
	}
	// Output installation is checked before advancing the pointer. Keep its old
	// bytes for the bounded rollback if the final local pointer write fails.
	var old []byte
	existed := false
	if output != "" {
		old, err = os.ReadFile(output)
		if err == nil {
			existed = true
		} else if !os.IsNotExist(err) {
			return Revision{}, failure("storage_failed", "read output before apply: %v", err)
		}
		if err := atomicReplace(output, candidate.Snapshot); err != nil {
			return Revision{}, err
		}
	}
	pointer, _ := json.Marshal(map[string]string{"revision": id})
	if err := atomicReplace(filepath.Join(store, "current.json"), append(pointer, '\n')); err != nil {
		if output != "" {
			var restore error
			if existed {
				restore = atomicReplace(output, old)
			} else {
				restore = os.Remove(output)
			}
			if restore != nil {
				return Revision{}, failure("storage_failed", "current pointer failed and output rollback requires repair: %v", restore)
			}
		}
		return Revision{}, err
	}
	return revision, nil
}
