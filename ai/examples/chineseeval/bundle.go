package chineseeval

import (
	"errors"
	"reflect"
	"regexp"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

type manifest struct {
	Module     string            `json:"module"`
	Package    string            `json:"package"`
	GitCommit  string            `json:"git_commit"`
	GoVersion  string            `json:"go_version"`
	Platform   string            `json:"platform"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at"`
	Versions   map[string]string `json:"versions"`
	Cases      []any             `json:"cases"`
}

// VerifyEvidence reconciles required files as well as checking their contents.
// Redaction alone cannot establish provenance or notice a deleted Observer file.
func (p *Plan) VerifyEvidence(r Report, manifestRaw, observations []byte) error {
	bad := errors.New("evaluation: provenance or observation evidence incomplete")
	if p.Verify(r) != nil {
		return bad
	}
	var m manifest
	if len(manifestRaw) > 1<<20 || decode(manifestRaw, &m) != nil || m.Module != "barness-ai" || m.Package != "host Chinese evaluation (not support smoke)" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(m.GitCommit) || m.GoVersion == "" || m.Platform == "" || m.StartedAt.IsZero() || m.FinishedAt.Before(m.StartedAt) || m.StartedAt.After(r.StartedAt) || m.FinishedAt.Before(r.FinishedAt) || m.Versions["dataset_hash"] != p.datasetHash || m.Versions["config_hash"] != p.configHash || m.Versions["model_catalog_version"] != p.config.CatalogVersion || m.Versions["model_catalog_hash"] != p.config.CatalogHash || m.Versions["typesafe_sdk"] != "direct-net-http-fixed-jev-1.13.0" {
		return bad
	}
	if r.Budget.Calls == 0 {
		return nil
	}
	// The CLI always collects the public Observer; fixtures use their own E2E
	// bundle format and do not become vendor evidence via this verifier.
	if r.Kind != "live" || len(observations) > 1<<20 {
		return bad
	}
	var records []ai.Observation
	if decode(observations, &records) != nil {
		return bad
	}
	samples := map[string]SampleResult{}
	for _, s := range r.Samples {
		if s.Metadata.RequestID != "" {
			samples[s.Metadata.RequestID] = s
		}
	}
	starts, finishes := map[string]int{}, map[string]int{}
	attemptStarts, attemptFinishes := map[string]int{}, map[string]int{}
	for _, o := range records {
		s, ok := samples[o.Call.RequestID]
		if !ok || o.Time.Before(r.StartedAt) || o.Time.After(m.FinishedAt) || o.Call.TenantID != s.Metadata.TenantID || o.Call.BindingID != s.Metadata.BindingID || o.Call.Operation != ai.OperationClassifier {
			return bad
		}
		switch o.Kind {
		case ai.ObservationCallStarted:
			starts[o.Call.RequestID]++
		case ai.ObservationCallFinished:
			finishes[o.Call.RequestID]++
			if !reflect.DeepEqual(o.Call, s.Metadata) {
				return bad
			}
			if len(s.Metadata.Attempts) > 0 && !reflect.DeepEqual(o.Usage, s.Metadata.Attempts[len(s.Metadata.Attempts)-1].Usage) {
				return bad
			}
			if (s.Status == "answered" || s.Status == "usage_incomplete") && (o.StopReason != ai.StopReasonStop || o.Error != nil) {
				return bad
			}
			if s.Code != "" && (o.Error == nil || o.Error.Code != s.Code || o.Error.Phase != s.Phase) {
				return bad
			}
		case ai.ObservationAttemptStarted, ai.ObservationAttemptFinished:
			if o.Attempt == nil {
				return bad
			}
			expected := false
			for _, a := range s.Metadata.Attempts {
				if a.AttemptID == o.Attempt.AttemptID {
					expected = true
					if o.Kind == ai.ObservationAttemptFinished && !reflect.DeepEqual(a, *o.Attempt) {
						return bad
					}
				}
			}
			if !expected {
				return bad
			}
			if o.Kind == ai.ObservationAttemptStarted {
				attemptStarts[o.Attempt.AttemptID]++
			} else {
				attemptFinishes[o.Attempt.AttemptID]++
			}
		default:
			return bad
		}
	}
	for id, s := range samples {
		if starts[id] != 1 || finishes[id] != 1 {
			return bad
		}
		for _, a := range s.Metadata.Attempts {
			if attemptStarts[a.AttemptID] != 1 || attemptFinishes[a.AttemptID] != 1 {
				return bad
			}
		}
	}
	return nil
}
