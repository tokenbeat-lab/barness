package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/catalog"
)

// CheckRequest selects a fixed controlled public E2E entry. There is no
// arbitrary command, caller-authored PASS, live or qualification option.
type CheckRequest struct {
	Store           string
	CandidateID     string
	RepositoryRoot  string
	EvidenceFeature string
	Timeout         time.Duration
	Confirmed       bool
	DryRun          bool
}

type CheckResult struct {
	Status         string   `json:"status"`
	Conclusion     string   `json:"conclusion,omitempty"`
	CandidateID    string   `json:"candidateId"`
	BaseDigest     string   `json:"baseDigest"`
	SnapshotDigest string   `json:"snapshotDigest"`
	Prepared       string   `json:"prepared,omitempty"`
	Receipt        string   `json:"receipt,omitempty"`
	ReceiptSHA256  string   `json:"receiptSHA256,omitempty"`
	MeasuredCommit string   `json:"measuredCommit,omitempty"`
	RequiredCases  []string `json:"requiredCases,omitempty"`
	Reference      string   `json:"reference,omitempty"`
	Message        string   `json:"message,omitempty"`
}

var checkFeaturePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,100}$`)

// Check freezes exact candidate/base/source bytes, executes the fixed public
// contract, then independently consumes its native artifacts and real exits.
// Every attempt is retained. It never changes the effective revision.
func Check(ctx context.Context, request CheckRequest) (result CheckResult, err error) {
	candidate, err := ReadCandidate(request.Store, request.CandidateID)
	if err != nil {
		return result, err
	}
	current, err := ReadCurrent(request.Store)
	if err != nil {
		return result, err
	}
	result = CheckResult{Status: "awaiting_check", CandidateID: candidate.ID, BaseDigest: candidate.BaseDigest, SnapshotDigest: digest(candidate.Snapshot)}
	if current.ID != candidate.BaseDigest {
		return result, failure("baseline_changed", "candidate baseline is no longer current")
	}
	if request.DryRun {
		return result, nil
	}
	if !request.Confirmed {
		return result, failure("confirmation_required", "check requires explicit --confirm")
	}
	if !checkFeaturePattern.MatchString(request.EvidenceFeature) {
		return result, failure("invalid_argument", "explicit evidence feature is required")
	}
	root, err := filepath.Abs(request.RepositoryRoot)
	if err != nil {
		return result, err
	}
	if request.Timeout == 0 {
		request.Timeout = 5 * time.Minute
	}
	seconds := int(request.Timeout / time.Second)
	if seconds < 1 || seconds > 1200 {
		return result, failure("invalid_argument", "check timeout must be 1s..20m")
	}
	base := filepath.Join(root, ".evidence", request.EvidenceFeature, "check-inputs")
	if err := localDirectory(base); err != nil {
		return result, err
	}
	dir, err := os.MkdirTemp(base, time.Now().UTC().Format("20060102T150405.000000000Z-"))
	if err != nil {
		return result, err
	}
	relative, _ := filepath.Rel(root, dir)
	result.Reference = filepath.ToSlash(filepath.Join(relative, "check.json"))
	defer func() {
		if err != nil {
			if known, ok := err.(*Error); ok {
				result.Status = known.Category
			} else {
				result.Status = "execution_failed"
			}
			result.Message = err.Error()
		}
		raw, encodeErr := jsonBytes(result)
		if encodeErr == nil {
			encodeErr = writeImmutable(filepath.Join(dir, "check.json"), raw)
		}
		if encodeErr != nil {
			result.Status = "execution_failed"
			err = encodeErr
		}
	}()
	for name, value := range map[string]any{"candidate.json": candidate, "baseline.json": current} {
		raw, e := jsonBytes(value)
		if e != nil {
			return result, e
		}
		if e = writeImmutable(filepath.Join(dir, name), raw); e != nil {
			return result, e
		}
	}
	if err = writeImmutable(filepath.Join(dir, "snapshot.json"), candidate.Snapshot); err != nil {
		return result, err
	}
	_, report, loadErr := catalog.Load(candidate.Snapshot)
	if loadErr != nil {
		return result, failure("definition_failed", "candidate file is invalid")
	}
	if report.Failed != 0 {
		category := "definition_failed"
		for _, model := range report.Models {
			if model.Reason == catalog.UnknownField || model.Reason == catalog.UnsupportedDefinition {
				category = "code_gap"
			}
		}
		return result, failure(category, "candidate has %d unavailable definitions", report.Failed)
	}
	preparedOutput, err := checkProcess(ctx, root, dir, "prepare", []string{"python3", "tools/verification/catalog_check.py", "prepare", "--directory", filepath.ToSlash(relative), "--feature", request.EvidenceFeature, "--timeout", fmt.Sprint(seconds)})
	if err != nil {
		return result, failure("execution_failed", "could not freeze actual check inputs")
	}
	var prepared struct {
		Prepared string            `json:"prepared"`
		Inputs   map[string]string `json:"inputs"`
	}
	if json.Unmarshal(preparedOutput, &prepared) != nil || prepared.Prepared == "" || len(prepared.Inputs) == 0 {
		return result, failure("execution_failed", "invalid prepared input manifest")
	}
	result.Prepared = prepared.Prepared
	var manifest struct {
		Argv []string `json:"argv"`
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(prepared.Prepared)))
	if err != nil {
		return result, err
	}
	if json.Unmarshal(raw, &manifest) != nil {
		return result, failure("execution_failed", "invalid execution plan")
	}
	argv := []string{"python3", "tools/verification/run.py", "--feature", request.EvidenceFeature, "--stage", "e2e", "--timeout", fmt.Sprint(seconds)}
	for _, name := range sortedCheckKeys(prepared.Inputs) {
		argv = append(argv, "--input", name)
	}
	argv = append(append(argv, "--"), manifest.Argv...)
	processCtx, cancel := context.WithTimeout(ctx, request.Timeout+15*time.Second)
	defer cancel()
	output, processErr := checkProcess(processCtx, root, dir, "execute", argv)
	var execution struct {
		Attempt  string `json:"attempt"`
		Status   string `json:"status"`
		ExitCode *int   `json:"exitCode"`
	}
	if json.Unmarshal(output, &execution) == nil && execution.Attempt != "" {
		result.Receipt = filepath.ToSlash(filepath.Join(execution.Attempt, "receipt.json"))
	}
	if processErr != nil || execution.Status != "PASS" || execution.ExitCode == nil || *execution.ExitCode != 0 {
		return result, failure("execution_failed", "controlled candidate process did not complete; retained attempt %s", execution.Attempt)
	}
	verified, err := inspectCheck(ctx, root, dir, result.Prepared, result.Receipt)
	if err != nil {
		return result, err
	}
	result.Status = "applicable"
	result.Conclusion = "code_compatible"
	result.ReceiptSHA256 = verified.ReceiptSHA256
	result.MeasuredCommit = verified.MeasuredCommit
	result.RequiredCases = verified.RequiredCases
	return result, nil
}

func localDirectory(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for current := absolute; current != filepath.Dir(current); current = filepath.Dir(current) {
		info, e := os.Lstat(current)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return failure("storage_failed", "evidence ancestor must be a local directory")
		}
	}
	if err := os.MkdirAll(absolute, 0700); err != nil {
		return failure("storage_failed", "create evidence directory: %v", err)
	}
	return nil
}
func checkRelative(root, path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", failure("invalid_argument", "check reference must be inside selected repository")
	}
	return filepath.ToSlash(relative), nil
}
