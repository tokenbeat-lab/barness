package minimaxlive

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"time"
)

type Call struct {
	ID               string            `json:"id"`
	CaseID           string            `json:"caseId"`
	Rerun            int               `json:"rerun"`
	Outcome          string            `json:"outcome"`
	Failure          string            `json:"failure,omitempty"`
	RequestID        string            `json:"requestId"`
	Reservation      Receipt           `json:"reservation"`
	Attempts         []Receipt         `json:"attemptReceipts"`
	InputBytes       int               `json:"inputBytes"`
	DurationMS       int64             `json:"durationMs"`
	RawModels        []string          `json:"rawModels"`
	UsagePresence    []json.RawMessage `json:"usagePresence"`
	RequestIDSources []string          `json:"requestIdSources"`
}
type Report struct {
	FormalRenewalID          string            `json:"formalRenewalId,omitempty"`
	FormalRenewalGrantSHA256 string            `json:"formalRenewalGrantSha256,omitempty"`
	Session                  string            `json:"session"`
	Cumulative               Budget            `json:"cumulativeSlotBudget"`
	Schema                   int               `json:"schema"`
	Kind                     string            `json:"kind"`
	Mode                     string            `json:"mode"`
	Stage                    string            `json:"stage"`
	Identity                 Identity          `json:"identity"`
	Outcome                  string            `json:"outcome"`
	Failure                  string            `json:"failure,omitempty"`
	Complete                 bool              `json:"complete"`
	AuditPassed              bool              `json:"auditPassed"`
	SDK                      string            `json:"sdk"`
	AccountAlias             string            `json:"accountAlias"`
	RegionAlias              string            `json:"regionAlias"`
	Entitlement              string            `json:"entitlement"`
	CatalogVersion           string            `json:"catalogVersion"`
	CatalogHash              string            `json:"catalogHash"`
	SourceHash               string            `json:"sourceHash"`
	PlanHash                 string            `json:"planHash"`
	GitCommit                string            `json:"gitCommit"`
	ExecutionSource          map[string]string `json:"executionSource"`
	StartedAt                time.Time         `json:"startedAt"`
	FinishedAt               time.Time         `json:"finishedAt"`
	Deadline                 time.Time         `json:"deadline"`
	Budget                   Budget            `json:"budget"`
	Expected                 []Case            `json:"expected"`
	Calls                    []Call            `json:"calls"`
	Unknown                  []string          `json:"unknown"`
	ReplaySource             string            `json:"replaySource,omitempty"`
}
type Capture struct {
	Result     ai.Result `json:"result"`
	ErrorCode  ai.Code   `json:"errorCode,omitempty"`
	ErrorPhase ai.Phase  `json:"errorPhase,omitempty"`
}
