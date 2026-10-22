package release

// Record is an immutable projection of a feature's single current status.
// It names actual receipts; claimed PASS values never substitute for artifacts.
type Record struct {
	Schema         int               `json:"schema"`
	Feature        string            `json:"feature"`
	CurrentState   FileRef           `json:"currentState"`
	Source         ExecutionSource   `json:"source"`
	SelectedInputs map[string]string `json:"selectedInputs"`
	Impact         Impact            `json:"impact"`
	E2E            StageRecord       `json:"e2e"`
	Live           StageRecord       `json:"live"`
	Race           StageRecord       `json:"race"`
	Pressure       StageRecord       `json:"pressure"`
}
type FileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type ExecutionSource struct {
	Commit string            `json:"commit"`
	Files  map[string]string `json:"files"`
}
type Impact struct {
	LiveRoutes       []string `json:"liveRoutes"`
	PressureRequired *bool    `json:"pressureRequired"`
	Reason           string   `json:"reason"`
}
type Selection struct {
	By   string   `json:"by"`
	Argv []string `json:"argv"`
}
type StageRecord struct {
	Status        string                 `json:"status"`
	Reason        string                 `json:"reason,omitempty"`
	Receipt       FileRef                `json:"receipt,omitempty"`
	RequiredCases []string               `json:"requiredCases,omitempty"`
	Selection     *Selection             `json:"selection,omitempty"`
	Routes        map[string]StageRecord `json:"routes,omitempty"`
	Consumers     []FileRef              `json:"consumers,omitempty"`
}
type NativeBundleRef struct {
	Path           string `json:"path"`
	Package        string `json:"package"`
	ManifestSHA256 string `json:"manifestSHA256"`
	AuditSHA256    string `json:"auditSHA256"`
}
type StageReceipt struct {
	Toolchain      map[string]string `json:"toolchain,omitempty"`
	Schema         int               `json:"schema"`
	Feature        string            `json:"feature"`
	Stage          string            `json:"stage"`
	Route          *string           `json:"route"`
	Status         string            `json:"status"`
	ExitCode       *int              `json:"exitCode"`
	Source         ExecutionSource   `json:"source"`
	SelectedInputs map[string]string `json:"selectedInputs"`
	StartedAt      string            `json:"startedAt"`
	FinishedAt     string            `json:"finishedAt"`
	ElapsedSeconds float64           `json:"elapsedSeconds"`
	Argv           []string          `json:"argv"`
	Bundles        []NativeBundleRef `json:"bundles"`
	CommandSHA256  string            `json:"commandSHA256"`
	Tests          []string          `json:"tests"`
	Files          map[string]string `json:"files"`
	Error          *string           `json:"error,omitempty"`
}
type RecordCheck struct {
	ID      string `json:"id"`
	Pass    bool   `json:"pass"`
	Problem string `json:"problem,omitempty"`
}
type RecordReport struct {
	ReleaseCommit string        `json:"releaseCommit,omitempty"`
	Pass          bool          `json:"pass"`
	Feature       string        `json:"feature,omitempty"`
	Commit        string        `json:"commit,omitempty"`
	Checks        []RecordCheck `json:"checks"`
}
