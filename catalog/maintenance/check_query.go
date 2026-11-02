package maintenance

import (
	"context"
	"os"
	"path/filepath"
)

// QueryCandidateCheck reads one explicitly selected check reference. It does
// not search for a floating latest result or mutate the candidate. A compatible
// label is returned only after the same strict verification used by Apply.
func QueryCandidateCheck(ctx context.Context, request ApplyRequest) (CheckResult, error) {
	candidate, err := ReadCandidate(request.Store, request.CandidateID)
	if err != nil {
		return CheckResult{}, err
	}
	if request.CheckReference == "" {
		return CheckResult{Status: "awaiting_check", CandidateID: candidate.ID, BaseDigest: candidate.BaseDigest, SnapshotDigest: digest(candidate.Snapshot)}, nil
	}
	root, err := filepath.Abs(request.RepositoryRoot)
	if err != nil {
		return CheckResult{}, err
	}
	relative, err := checkRelative(root, request.CheckReference)
	if err != nil {
		return CheckResult{}, err
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		return CheckResult{}, failure("check_invalid", "selected candidate check is unavailable")
	}
	var result CheckResult
	if decode(raw, &result) != nil || filepath.Base(relative) != "check.json" || result.Reference != relative || result.CandidateID != candidate.ID || result.BaseDigest != candidate.BaseDigest || result.SnapshotDigest != digest(candidate.Snapshot) {
		return CheckResult{Status: "check_invalid", CandidateID: candidate.ID}, failure("check_invalid", "selected check does not bind this exact candidate")
	}
	switch result.Status {
	case "applicable":
		_, verified, err := VerifyCheck(ctx, request)
		if err != nil {
			result.Status = "check_invalid"
			result.Message = err.Error()
			return result, err
		}
		return verified, nil
	case "definition_failed", "code_gap", "execution_failed", "check_invalid", "storage_failed":
		return result, nil
	default:
		return CheckResult{Status: "check_invalid", CandidateID: candidate.ID}, failure("check_invalid", "selected check has no completed recognized outcome")
	}
}
