package supportmatrix

import (
	"slices"
	"strings"
	"time"
)

// DecisionsEvidence distinguishes official live acceptance from controlled
// pipeline tests. The candidate has a separate schema and cannot be merged.
type DecisionsEvidence struct {
	Mode                         string `json:"mode"`
	Complete                     bool   `json:"complete"`
	AuditPassed                  bool   `json:"auditPassed"`
	SourceManifestHash           string `json:"sourceManifestHash"`
	AccountVerificationStatus    string `json:"accountVerificationStatus"`
	RegionVerificationStatus     string `json:"regionVerificationStatus"`
	PermissionVerificationStatus string `json:"permissionVerificationStatus"`
	AccountVerificationPolicy    string `json:"accountVerificationPolicy"`
}

func DecisionsExpected() []string {
	return []string{"text-choice", "text-score", "text-bool", "mixed-questions", "image-png", "image-jpeg", "image-webp", "cancel-in-flight", "recovery"}
}

func checkDecisions(r Report) error {
	if r.Combo != "openai-decisions" {
		return nil
	}
	if r.Operation != "classifier" || r.Provider != "openai" || r.API != "openai-decisions" || r.Model != "gpt-6-luna" || r.Endpoint != "https://api.openai.com/v1" {
		return refuse("Decisions report does not name the fixed official route and model")
	}
	if _, err := RouteCatalog(r.evidenceDir, r); err != nil {
		return err
	}
	expected := DecisionsExpected()
	if err := checkExpected(r, expected); err != nil {
		return err
	}
	if len(r.Scenarios) != len(expected) {
		return refuse("Decisions lacks its full scenario results")
	}
	d := r.Decisions
	if d == nil || d.Mode != "live" || !d.AuditPassed || d.SourceManifestHash != "61be0c194e7231aa5c3b986c849c2b23935ae6a1830038413530eb0d3618b8fe" {
		return refuse("Decisions lacks its own audited live source evidence")
	}
	executed := slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.Outcome != NotRun })
	if executed {
		if d.AccountVerificationPolicy != "allow_unknown" && d.AccountVerificationPolicy != "require_host_verification" {
			return refuse("Decisions account verification policy is missing")
		}
		for _, status := range []string{d.AccountVerificationStatus, d.RegionVerificationStatus, d.PermissionVerificationStatus} {
			if status != "host_verified" && !(status == "unknown" && d.AccountVerificationPolicy == "allow_unknown") {
				return refuse("Decisions verification status is inconsistent")
			}
		}
	}
	if r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.FinishedAt.Sub(r.StartedAt) > 180*time.Second || strings.TrimSpace(r.GitCommit) == "" || strings.TrimSpace(r.SDK) == "" || (executed && strings.TrimSpace(r.AccountAlias) == "") {
		return refuse("Decisions lacks bounded execution provenance")
	}
	b := r.Budget
	if b.InputTokensUsed < 0 || b.OutputTokensUsed < 0 || (!executed && (b.InputTokensUsed != 0 || b.OutputTokensUsed != 0)) || b.MaxCalls != 12 || b.MaxAttempts != 12 || b.MaxQuestions != 24 || b.MaxStateBytesTotal != 4<<20 || b.MaxImages != 8 || b.MaxDecodedImageBytesTotal != 4<<20 || b.ProcessSeconds != 180 || b.CallSeconds != 30 || b.FinalizationReserveSeconds != 10 || b.MaxRetries != 0 || b.EnvironmentRetries != 0 || b.MaxOutputTokens != 0 || b.MaxStateBytes != 0 || b.CallsUsed < 0 || b.CallsUsed > 12 || b.AttemptsReserved != b.CallsUsed || b.HTTPAttempts < 0 || b.HTTPAttempts > b.AttemptsReserved || b.QuestionsUsed < 0 || b.QuestionsUsed > 24 || b.StateBytesSubmitted < 0 || b.StateBytesSubmitted > 4<<20 || b.ImagesUsed < 0 || b.ImagesUsed > 8 || b.DecodedImageBytesSubmitted < 0 || b.DecodedImageBytesSubmitted > 4<<20 {
		return refuse("Decisions report has an invalid or exceeded budget")
	}
	calls, attempts, questions, images, state, imageBytes := 0, 0, 0, 0, 0, 0
	allPassed := true
	for _, s := range r.Scenarios {
		if s.Model != r.Model || s.Outcome == Unsupported || s.Retries != 0 || s.Calls < 0 || s.Calls > 1 || s.Attempts < 0 || s.Attempts > s.Calls || s.Questions < 0 || s.Questions > 3 || s.Images < 0 || s.Images > 1 || s.StateBytes < 0 || s.StateBytes > 4<<20 || s.DecodedImageBytes < 0 || s.DecodedImageBytes > 4<<20 || s.DurationMS < 0 || s.DurationMS > 30000 {
			return refuse("Decisions capability %q has invalid fixed model or counters", s.ID)
		}
		q, img := 1, 0
		if s.ID == "mixed-questions" || s.ID == "recovery" {
			q = 3
		}
		if strings.HasPrefix(s.ID, "image-") {
			img = 1
		}
		if s.Calls > 0 && (s.Questions != q || s.Images != img || s.StateBytes == 0 || (img == 1 && s.DecodedImageBytes == 0) || (img == 0 && s.DecodedImageBytes != 0)) {
			return refuse("Decisions capability %q disagrees with its fixed reservation", s.ID)
		}
		if s.Outcome == Pass && (s.Calls != 1 || s.Attempts != 1 || s.CaseID == "" || len(s.ProviderRequestIDs) != 1 || strings.TrimSpace(s.ProviderRequestIDs[0]) == "") {
			return refuse("Decisions capability %q lacks actual call evidence", s.ID)
		}
		if s.Outcome == NotRun && (s.Calls != 0 || s.Attempts != 0 || s.Questions != 0 || s.Images != 0 || s.StateBytes != 0 || s.DecodedImageBytes != 0) {
			return refuse("Decisions NOT_RUN capability %q spent resources", s.ID)
		}
		allPassed = allPassed && s.Outcome == Pass
		calls += s.Calls
		attempts += s.Attempts
		questions += s.Questions
		images += s.Images
		state += s.StateBytes
		imageBytes += s.DecodedImageBytes
	}
	if calls != b.CallsUsed || attempts != b.HTTPAttempts || questions != b.QuestionsUsed || images != b.ImagesUsed || state != b.StateBytesSubmitted || imageBytes != b.DecodedImageBytesSubmitted {
		return refuse("Decisions actual counters disagree with scenarios")
	}
	if d.Complete != allPassed {
		return refuse("Decisions pipeline completeness disagrees with outcomes")
	}
	// check already rejects duplicate scenario IDs; equal lengths and membership
	// then prove every independent required capability is present exactly once.
	for _, id := range expected {
		if !slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.ID == id }) {
			return refuse("Decisions missing %q", id)
		}
	}
	return nil
}
