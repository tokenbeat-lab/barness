// Package supportmatrix holds barness-ai's live smoke report and support
// matrix (spec Testing Decisions §5): what one live process found for its
// Provider × API combination, and what the maintainers publish as supported.
//
// A live process writes one Report for the one combination it holds a key
// for. Merge folds it into that combination's Row and nothing else: a pass
// through a shared adapter never marks another combination (spec I10). The
// matrix is checked in (ai/live/support-matrix.json) and changed only by
// merging reports.
package supportmatrix

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// SchemaVersion is the report and matrix layout this package reads and
// writes.
const SchemaVersion = 2

// Outcome is a live scenario's result. There are exactly four; a vendor
// fault is a Fail, never a skip.
type Outcome string

const (
	// Pass: the scenario ran and every assertion held.
	Pass Outcome = "PASS"
	// Fail: the scenario ran and an assertion failed, or the vendor or the
	// environment failed within the retry budget; ErrorCategory says which.
	Fail Outcome = "FAIL"
	// NotRun: the scenario was not executed, e.g. no key was injected.
	NotRun Outcome = "NOT_RUN"
	// Unsupported: the combination's protocol or model cannot provide the
	// capability; Note says why.
	Unsupported Outcome = "UNSUPPORTED"
)

func (o Outcome) valid() bool { return o == Pass || o == Fail || o == NotRun || o == Unsupported }

// Report is one live process's result for its combination.
type Report struct {
	Schema int `json:"schema"`
	// Combo names the combination, e.g. "deepseek-chat"; Provider and API
	// are its pi-ai identifiers and must match the matrix row.
	Combo     string `json:"combo"`
	Operation string `json:"operation"`
	Provider  string `json:"provider"`
	API       string `json:"api"`
	Endpoint  string `json:"endpoint"`
	// Model is the combination's default model; a scenario may name another.
	Model string `json:"model"`
	// SDK is the module and version serving the protocol, or "direct HTTP".
	SDK string `json:"sdk"`
	// AccountAlias names the low-privilege test account and region; never a
	// key or account id.
	AccountAlias   string    `json:"accountAlias"`
	GitCommit      string    `json:"gitCommit"`
	CatalogVersion string    `json:"catalogVersion"`
	CatalogHash    string    `json:"catalogHash"`
	StartedAt      time.Time `json:"startedAt"`
	FinishedAt     time.Time `json:"finishedAt"`
	// Budget is what the process allowed itself and used.
	Budget Budget `json:"budget"`
	// Expected are the IDs of every scenario the combination has, so a run
	// narrowed to some of them cannot mark the row all-passed.
	Expected  []string         `json:"expected"`
	Scenarios []ScenarioResult `json:"scenarios"`
}

// Budget bounds one live process's spend.
type Budget struct {
	MaxCalls            int `json:"maxCalls"`
	MaxImages           int `json:"maxImages,omitempty"`
	ImagesUsed          int `json:"imagesUsed,omitempty"`
	MaxImageWidth       int `json:"maxImageWidth,omitempty"`
	MaxImageHeight      int `json:"maxImageHeight,omitempty"`
	MaxQuestions        int `json:"maxQuestions,omitempty"`
	QuestionsUsed       int `json:"questionsUsed,omitempty"`
	HTTPAttempts        int `json:"httpAttempts"`
	MaxStateBytes       int `json:"maxStateBytesPerCall,omitempty"`
	StateBytesSubmitted int `json:"stateBytesSubmitted,omitempty"`
	MaxRetries          int `json:"maxRetriesPerScenario"`
	MaxOutputTokens     int `json:"maxOutputTokensPerCall"`
	CallsUsed           int `json:"callsUsed"`
	OutputTokensUsed    int `json:"outputTokensUsed"`
	InputTokensUsed     int `json:"inputTokensUsed"`
	EnvironmentRetries  int `json:"environmentRetries"`
}

// ScenarioResult is one scenario of a report.
type ScenarioResult struct {
	ID string `json:"id"`
	// CaseID is the evidence case holding the scenario's captures.
	CaseID  string  `json:"caseId"`
	Model   string  `json:"model"`
	Outcome Outcome `json:"outcome"`
	// ErrorCategory is the barness-ai error code, or one of the Category
	// values; required for a Fail.
	ErrorCategory string `json:"errorCategory,omitempty"`
	// Environment marks a Fail as a vendor or environment fault (rate
	// limit, 5xx, transport, timeout) that persisted through its retries.
	Environment        bool        `json:"environment,omitempty"`
	DurationMS         int64       `json:"durationMs"`
	Calls              int         `json:"calls"`
	Attempts           int         `json:"attempts"`
	Questions          int         `json:"questions,omitempty"`
	Images             int         `json:"images,omitempty"`
	ImageCalls         []ImageCall `json:"imageCalls,omitempty"`
	Retries            int         `json:"retries,omitempty"`
	ProviderRequestIDs []string    `json:"providerRequestIds,omitempty"`
	// Note records what was observed beyond pass/fail, and is required for
	// Unsupported.
	Note string `json:"note,omitempty"`
}

// Failure categories that are not barness-ai error codes.
const (
	// CategoryAssertion: the vendor answered and an assertion failed.
	CategoryAssertion = "assertion"
	// CategoryBudget: the process's call budget was used up.
	CategoryBudget = "budget"
	// CategoryConfig: the process was misconfigured and sent nothing.
	// Such a report says nothing about the vendor and is never merged.
	CategoryConfig = "config"
)

// Matrix is the published support matrix: one row per first-phase
// combination.
type Matrix struct {
	Schema int   `json:"schema"`
	Rows   []Row `json:"rows"`
}

// Row is one combination's support state.
type Row struct {
	Combo        string `json:"combo"`
	Operation    string `json:"operation"`
	Provider     string `json:"provider"`
	API          string `json:"api"`
	Model        string `json:"model,omitempty"`
	SDK          string `json:"sdk,omitempty"`
	AccountAlias string `json:"accountAlias,omitempty"`
	// LastRunAt is when a report last executed any scenario.
	LastRunAt *time.Time `json:"lastRunAt,omitempty"`
	// AllPassedAt is when a report last executed every scenario with each
	// PASS or UNSUPPORTED: the combination's last complete pass.
	AllPassedAt  *time.Time   `json:"allPassedAt,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

// Capability is one scenario's state within a row.
type Capability struct {
	ID            string     `json:"id"`
	Model         string     `json:"model"`
	Outcome       Outcome    `json:"outcome"`
	ErrorCategory string     `json:"errorCategory,omitempty"`
	Note          string     `json:"note,omitempty"`
	LastRunAt     *time.Time `json:"lastRunAt,omitempty"`
	LastPassedAt  *time.Time `json:"lastPassedAt,omitempty"`
}

// ErrInvalidReport matches every refused merge.
var ErrInvalidReport = errors.New("supportmatrix: invalid report")

func refuse(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidReport, fmt.Sprintf(format, args...))
}

// Merge folds r into its combination's row and returns the new matrix; m is
// not changed. A report where nothing ran leaves the matrix as it was, and a
// scenario that did not run keeps its capability's last state. A complete
// successful report replaces the active capability set; matching capabilities
// keep their pass history, and archived reports retain removed probes.
func Merge(m Matrix, r Report) (Matrix, error) {
	idx, err := check(m, r)
	if err != nil {
		return Matrix{}, err
	}
	out := clone(m)
	executed := slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.Outcome != NotRun })
	if !executed {
		return out, nil
	}
	at := r.FinishedAt
	row := &out.Rows[idx]
	row.Model, row.SDK, row.AccountAlias, row.LastRunAt = r.Model, r.SDK, r.AccountAlias, &at
	complete := true
	for _, id := range r.Expected {
		if !slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.ID == id }) {
			complete = false
		}
	}
	for _, s := range r.Scenarios {
		if s.Outcome == NotRun || s.Outcome == Fail {
			complete = false
		}
		if s.Outcome == NotRun {
			continue
		}
		i := slices.IndexFunc(row.Capabilities, func(c Capability) bool { return c.ID == s.ID })
		if i < 0 {
			row.Capabilities = append(row.Capabilities, Capability{ID: s.ID})
			i = len(row.Capabilities) - 1
		}
		c := &row.Capabilities[i]
		c.Model, c.Outcome, c.ErrorCategory, c.Note, c.LastRunAt = s.Model, s.Outcome, s.ErrorCategory, s.Note, &at
		if s.Outcome == Pass {
			c.LastPassedAt = &at
		}
	}
	if complete {
		row.Capabilities = slices.DeleteFunc(row.Capabilities, func(c Capability) bool {
			return !slices.Contains(r.Expected, c.ID)
		})
		row.AllPassedAt = &at
	}
	return out, nil
}

// check validates r against m and returns its row's index.
func check(m Matrix, r Report) (int, error) {
	if r.Schema != SchemaVersion || m.Schema != SchemaVersion {
		return 0, refuse("schema %d (matrix %d), want %d", r.Schema, m.Schema, SchemaVersion)
	}
	idx := slices.IndexFunc(m.Rows, func(row Row) bool { return row.Combo == r.Combo })
	if idx < 0 {
		return 0, refuse("combination %q is not in the matrix", r.Combo)
	}
	row := m.Rows[idx]
	if !row.ValidIdentity() {
		return 0, refuse("matrix row %q has an inconsistent route identity", row.Combo)
	}
	if r.Operation == "" || row.Operation != r.Operation || row.Provider != r.Provider || row.API != r.API {
		return 0, refuse("combination %q is %s × %s × %s, the report says %s × %s × %s", r.Combo, row.Operation, row.Provider, row.API, r.Operation, r.Provider, r.API)
	}
	if r.FinishedAt.IsZero() {
		return 0, refuse("report has no finish time")
	}
	if row.LastRunAt != nil && !r.FinishedAt.After(*row.LastRunAt) {
		return 0, refuse("report finished %s, not after the row's last run %s", r.FinishedAt.Format(time.RFC3339), row.LastRunAt.Format(time.RFC3339))
	}
	if len(r.Expected) == 0 {
		return 0, refuse("report lists no expected scenarios")
	}
	seen := map[string]bool{}
	for _, s := range r.Scenarios {
		switch {
		case s.ID == "":
			return 0, refuse("scenario without an id")
		case !slices.Contains(r.Expected, s.ID):
			return 0, refuse("scenario %q is not one the combination expects", s.ID)
		case s.ErrorCategory == CategoryConfig:
			return 0, refuse("scenario %q failed on the process's configuration; fix it and run again", s.ID)
		case seen[s.ID]:
			return 0, refuse("scenario %q appears twice", s.ID)
		case !s.Outcome.valid():
			return 0, refuse("scenario %q has outcome %q", s.ID, s.Outcome)
		case s.Outcome == Fail && s.ErrorCategory == "":
			return 0, refuse("failed scenario %q has no error category", s.ID)
		case s.Outcome == Unsupported && s.Note == "":
			return 0, refuse("unsupported scenario %q gives no reason", s.ID)
		}
		seen[s.ID] = true
	}
	if err := checkImages(r); err != nil {
		return 0, err
	}
	if err := checkGoogleImages(r); err != nil {
		return 0, err
	}
	if err := checkClassifier(r); err != nil {
		return 0, err
	}
	return idx, nil
}

func clone(m Matrix) Matrix {
	out := Matrix{Schema: m.Schema, Rows: slices.Clone(m.Rows)}
	for i := range out.Rows {
		out.Rows[i].Capabilities = slices.Clone(out.Rows[i].Capabilities)
	}
	return out
}

// Only the shared set validation lives here. Each operation retains its own
// route, fixed model, counters and optional-capability semantics.
func checkExpected(r Report, expected []string) error {
	if len(r.Expected) != len(expected) {
		return refuse("%s report lacks its complete capability set", r.Combo)
	}
	for _, id := range expected {
		if !slices.Contains(r.Expected, id) {
			return refuse("%s report lacks %q", r.Combo, id)
		}
	}
	return nil
}
