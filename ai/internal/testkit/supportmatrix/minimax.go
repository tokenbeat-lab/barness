package supportmatrix

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// MiniMaxEvidence binds the public support projection to its sealed own M3
// process. Candidate/controlled/replay processes cannot construct this proof.
type MiniMaxEvidence struct {
	ReportSHA256 string `json:"reportSha256"`
	SealSHA256   string `json:"sealSha256"`
	PlanSHA256   string `json:"planSha256"`
	SourceSHA256 string `json:"sourceSha256"`
	RegionAlias  string `json:"regionAlias"`
	Purpose      string `json:"purpose"`
}

func MiniMaxExpected() []string {
	var ids []string
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	for _, p := range minimaxlive.Plan(id) {
		ids = append(ids, p.ID)
	}
	return ids
}
func NewMiniMaxReport(dir string) (Report, error) {
	own, e := minimaxlive.Verify(dir, true)
	if e != nil {
		return Report{}, fmt.Errorf("MiniMax formal evidence: %w", e)
	}
	r := Report{Schema: SchemaVersion, Combo: "minimax-cn-" + own.Identity.Route, Operation: "chat", Provider: "minimax-cn", API: string(own.Identity.API), Endpoint: own.Identity.Endpoint, Model: own.Identity.Model, SDK: own.SDK, AccountAlias: own.AccountAlias, GitCommit: own.GitCommit, CatalogVersion: own.CatalogVersion, CatalogHash: own.CatalogHash, StartedAt: own.StartedAt, FinishedAt: own.FinishedAt, Expected: MiniMaxExpected(), evidenceDir: dir,
		Budget: Budget{MaxRetries: 1, MaxCalls: 20, MaxAttempts: 24, MaxOutputTokens: 4096, CallSeconds: 120, ProcessSeconds: 1800, CallsUsed: own.Budget.Calls, AttemptsReserved: own.Budget.ReservedAttempts, HTTPAttempts: own.Budget.Attempts, MaxRequestedOutputTokens: 40000, RequestedOutputTokens: own.Budget.RequestedOutput, MaxReservedInputTokens: 120000, ReservedInputTokens: own.Budget.ReservedInput}}
	for _, p := range own.Expected {
		scenario := ScenarioResult{ID: p.ID, Model: r.Model, Outcome: Pass}
		for _, c := range own.Calls {
			if c.ID != p.ID {
				continue
			}
			scenario.CaseID = c.CaseID
			scenario.DurationMS += c.DurationMS
			scenario.Calls++
			scenario.Attempts += len(c.Attempts)
			scenario.Retries += c.Rerun
		}
		r.Scenarios = append(r.Scenarios, scenario)
	}

	report, e := minimaxlive.ReadFile(filepath.Join(dir, minimaxlive.ReportFile), 2<<20)
	if e != nil {
		return Report{}, e
	}
	seal, e := minimaxlive.ReadFile(filepath.Join(dir, minimaxlive.SealFile), 2<<20)
	if e != nil {
		return Report{}, e
	}
	r.MiniMax = &MiniMaxEvidence{ReportSHA256: minimaxlive.Hash(report), SealSHA256: minimaxlive.Hash(seal), PlanSHA256: own.PlanHash, SourceSHA256: own.SourceHash, RegionAlias: own.RegionAlias, Purpose: own.Identity.Purpose}
	return r, nil
}
func VerifyMiniMaxReport(r Report) error {
	if r.evidenceDir == "" || r.MiniMax == nil {
		return refuse("MiniMax requires an owned sealed report")
	}
	want, e := NewMiniMaxReport(r.evidenceDir)
	if e != nil {
		return refuse("MiniMax formal report invalid: %v", e)
	}
	if !reflect.DeepEqual(want, r) {
		return refuse("MiniMax support projection differs from actual evidence")
	}
	return nil
}
func checkMiniMax(r Report) error {
	if !strings.HasPrefix(r.Combo, "minimax-cn-") {
		if r.MiniMax != nil {
			return refuse("foreign report carries MiniMax proof")
		}
		return nil
	}
	return VerifyMiniMaxReport(r)
}
