//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// All counters are cumulative within this route/process, including failed calls.
type multimodalResources struct {
	Calls         int `json:"calls"`
	Attempts      int `json:"attempts"`
	Questions     int `json:"questions"`
	Sources       int `json:"sources"`
	References    int `json:"references"`
	SourceBytes   int `json:"sourceBytes"`
	ExpandedBytes int `json:"expandedBytes"`
	StateBytes    int `json:"stateBytes"`
}

var multimodalLimits = multimodalResources{12, 12, 32, 24, 32, 8 << 20, 16 << 20, 4 << 20}

func (r multimodalResources) fits(limit multimodalResources) bool {
	a, b := []int{r.Calls, r.Attempts, r.Questions, r.Sources, r.References, r.SourceBytes, r.ExpandedBytes, r.StateBytes}, []int{limit.Calls, limit.Attempts, limit.Questions, limit.Sources, limit.References, limit.SourceBytes, limit.ExpandedBytes, limit.StateBytes}
	for i, v := range a {
		if v < 0 || v > b[i] {
			return false
		}
	}
	return true
}
func (r multimodalResources) add(v multimodalResources) multimodalResources {
	return multimodalResources{r.Calls + v.Calls, r.Attempts + v.Attempts, r.Questions + v.Questions, r.Sources + v.Sources, r.References + v.References, r.SourceBytes + v.SourceBytes, r.ExpandedBytes + v.ExpandedBytes, r.StateBytes + v.StateBytes}
}

type multimodalExecution struct {
	Schema            int                      `json:"schema"`
	Kind              string                   `json:"kind"`
	Route             string                   `json:"route"`
	Mode              string                   `json:"mode"`
	Outcome           string                   `json:"outcome"`
	Complete          bool                     `json:"complete"`
	AuditPassed       bool                     `json:"auditPassed"`
	Failure           string                   `json:"failure,omitempty"`
	StartedAt         time.Time                `json:"startedAt"`
	FinishedAt        time.Time                `json:"finishedAt"`
	ExecutionDeadline time.Time                `json:"executionDeadline"`
	GitCommit         string                   `json:"gitCommit"`
	SourceHashes      map[string]string        `json:"sourceHashes"`
	CatalogVersion    string                   `json:"catalogVersion"`
	CatalogHash       string                   `json:"catalogHash"`
	CatalogFileHash   string                   `json:"catalogFileHash"`
	Authorization     *multimodalAuthorization `json:"authorization"`
	AuthorizationHash string                   `json:"authorizationHash,omitempty"`
	Budget            struct {
		Limits                     multimodalResources `json:"limits"`
		Reserved                   multimodalResources `json:"reserved"`
		Actual                     multimodalResources `json:"actual"`
		CallsUsed                  int                 `json:"callsUsed"`
		HTTPAttempts               int                 `json:"httpAttempts"`
		MaxNativeBodyBytes         int                 `json:"maxNativeBodyBytes"`
		ProcessSeconds             int                 `json:"processSeconds"`
		CallSeconds                int                 `json:"callSeconds"`
		FinalizationReserveSeconds int                 `json:"finalizationReserveSeconds"`
		Retries                    int                 `json:"retries"`
		ReportedInputTokens        int                 `json:"reportedInputTokens"`
		ReportedOutputTokens       int                 `json:"reportedOutputTokens"`
	} `json:"budget"`
	Expected  []string `json:"expected"`
	Scenarios []string `json:"scenarios"`
	Unknown   []string `json:"unknown"`
}

var multimodalReport *multimodalExecution

func initialMultimodalReport() *multimodalExecution {
	r := &multimodalExecution{Schema: 1, Kind: "multimodal-candidate-only", Route: os.Getenv("BARNESS_MM_ROUTE"), Mode: "live", Outcome: "NOT_RUN", StartedAt: started, GitCommit: evidence.GitCommit(), Unknown: []string{"visual quality and localization IoU", "absence false positives", "provider image capacity and tokenization", "vendor billing/cache/regional premium", "account/region/permissions are Host attestation only", "Decisions alias weights are not frozen by catalog hash"}}
	if os.Getenv("BARNESS_MM_CONTROLLED") != "" {
		r.Mode = "controlled"
	}
	if os.Getenv("BARNESS_MM_FORMAL") == "1" {
		r.Kind = "jev-formal-protocol"
	}
	r.Budget.Limits = multimodalLimits
	r.Budget.MaxNativeBodyBytes, r.Budget.ProcessSeconds, r.Budget.CallSeconds, r.Budget.FinalizationReserveSeconds = 4<<20, 180, 30, 10
	r.ExecutionDeadline = started.Add(170 * time.Second)
	return r
}

func finishMultimodalReport(code int, auditOK bool) int {
	r := multimodalReport
	r.AuditPassed = auditOK
	r.Complete = false
	deadline := started.Add(180 * time.Second)
	if shorter := r.ExecutionDeadline.Add(10 * time.Second); shorter.Before(deadline) {
		deadline = shorter
	}
	fail := func(reason string) int {
		r.FinishedAt = time.Now().UTC()
		r.Outcome, r.Complete, r.Failure = "FAIL", false, reason
		_ = run.Record("multimodal-report", r)
		fmt.Fprintln(os.Stderr, "multimodal evidence rejected:", reason)
		return 1
	}
	if !auditOK {
		return fail("redaction audit failed")
	}
	if code != 0 {
		r.Outcome = "FAIL"
	}
	if time.Now().After(deadline) {
		return fail("process deadline exceeded before finalization")
	}
	if os.Getenv("BARNESS_MM_CONTROLLED") == "manifest" {
		_ = os.WriteFile(filepath.Join(run.Dir(), "manifest.json"), []byte(`{`), 0600)
	}
	manifest, err := readDecisionsFile(filepath.Join(run.Dir(), "manifest.json"), 4<<20)
	if err != nil || !json.Valid(manifest) {
		return fail("manifest invalid")
	}
	// Stay non-accepting through every expensive audit/read/hash operation.
	if err = run.Record("multimodal-report", r); err != nil {
		return fail("report write failed")
	}
	findings, err := run.Audit()
	if err != nil || len(findings) > 0 {
		r.AuditPassed = false
		return fail("final audit failed")
	}
	files := map[string]string{}
	err = filepath.WalkDir(run.Dir(), func(path string, d os.DirEntry, err error) error {
		if time.Now().After(deadline) {
			return fmt.Errorf("finalization deadline exceeded")
		}
		if err != nil || d.IsDir() {
			return err
		}
		if d.Name() == "multimodal-integrity.json" {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular evidence")
		}
		raw, err := readDecisionsFile(path, 16<<20)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(run.Dir(), path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = sha256Text(raw)
		return nil
	})
	if err != nil {
		return fail("inventory read failed")
	}
	r.FinishedAt = time.Now().UTC()
	r.Complete = r.Outcome == "PASS" && code == 0
	if err = run.Record("multimodal-report", r); err != nil {
		return fail("final report write failed")
	}
	final, err := readDecisionsFile(filepath.Join(run.Dir(), "multimodal-report.json"), 1<<20)
	if err != nil {
		return fail("final report read failed")
	}
	files["multimodal-report.json"] = sha256Text(final)
	if os.Getenv("BARNESS_MM_CONTROLLED") == "inventory" {
		_ = os.Mkdir(filepath.Join(run.Dir(), "multimodal-integrity.json"), 0700)
	}
	if err = run.Record("multimodal-integrity", map[string]any{"schema": 1, "files": files}); err != nil {
		return fail("inventory write failed")
	}
	// Only a controlled test can shorten this last deadline. Real publication
	// checks elapsed time after the final write too; slow disk cannot bless PASS.
	if os.Getenv("BARNESS_MM_CONTROLLED") == "publication-deadline" {
		deadline = time.Now().Add(-time.Nanosecond)
	}
	if time.Now().After(deadline) {
		return fail("process deadline exceeded during publication")
	}
	return code
}
