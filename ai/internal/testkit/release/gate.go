// Package release verifies selected immutable execution receipts and native artifacts.
package release

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type Status string

const (
	Pass        Status = "PASS"
	Fail        Status = "FAIL"
	NotRun      Status = "NOT_RUN"
	Unsupported Status = "UNSUPPORTED"
	NoEvidence  Status = "NO_EVIDENCE"
)

type Case struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}
type Bundle struct {
	Dir       string `json:"dir"`
	Commit    string `json:"commit"`
	Cases     []Case `json:"-"`
	caseFiles map[string]verifiedCase
	Problems  []string `json:"problems,omitempty"`
}
type AuditResult struct {
	Bundle   string                `json:"bundle"`
	Combo    string                `json:"combo,omitempty"`
	Findings []audit.Finding       `json:"findings"`
	Err      string                `json:"error,omitempty"`
	Live     *supportmatrix.Report `json:"-"`
}
type Snapshot struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
	Detail  string `json:"detail,omitempty"`
}
type EvidenceCheck struct {
	Name   string `json:"name"`
	Bundle string `json:"bundle"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}
