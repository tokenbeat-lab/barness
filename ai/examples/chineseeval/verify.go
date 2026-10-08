package chineseeval

import (
	"errors"
	"math"
	"reflect"
)

// Verify rejects incomplete or internally inconsistent artifacts. This checks
// completeness and provenance, never a business accuracy threshold. A fixture
// report remains a fixture even when this returns nil.
func (p *Plan) Verify(r Report) error {
	bad := errors.New("evaluation: report integrity or audit failed")
	if r.Schema != 1 || (r.Kind != "live" && r.Kind != "fixture") || r.Audit != "PASS" || r.AccountAlias == "" || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.DatasetHash != p.datasetHash || r.ConfigHash != p.configHash || !reflect.DeepEqual(r.Config, p.config) || !reflect.DeepEqual(r.Dataset, p.dataset) || len(r.Samples) != len(p.dataset.Samples) {
		return bad
	}
	var b Budget
	complete, notRun := true, true
	ids := map[string]bool{}
	for i, s := range r.Samples {
		if s.ID != p.dataset.Samples[i].ID {
			return bad
		}
		switch s.Status {
		case "not_run", "budget_exhausted":
			if s.Metadata.RequestID != "" || s.Answer != nil || len(s.Metadata.Attempts) > 0 {
				return bad
			}
		case "cancelled":
			if s.Answer != nil {
				return bad
			}
		case "failed", "version_drift", "missing_answer", "usage_incomplete", "answered":
			if s.Metadata.RequestID == "" {
				return bad
			}
		default:
			return bad
		}
		if s.Metadata.RequestID != "" {
			if ids[s.Metadata.RequestID] {
				return bad
			}
			ids[s.Metadata.RequestID] = true
			b.Calls++
			b.Questions++
			b.add(s.Metadata)
		}
		if s.Status == "answered" || s.Status == "usage_incomplete" {
			if !p.identity(s.Metadata) || s.ResponseModel != p.config.Model || s.Answer == nil || !p.answerValid(*s.Answer) {
				return bad
			}
			incomplete := false
			for _, a := range s.Metadata.Attempts {
				if a.UsageReporting != "complete" {
					incomplete = true
				}
			}
			if incomplete != (s.Status == "usage_incomplete") {
				return bad
			}
		} else if s.Answer != nil {
			return bad
		}
		if s.Status != "answered" {
			complete = false
		}
		if s.Status != "not_run" {
			notRun = false
		}
	}
	if !reflect.DeepEqual(b, r.Budget) || b.Calls > p.config.MaxCalls || b.Questions > p.config.MaxQuestions || b.Attempts > p.config.MaxAttempts || b.Retries > p.config.MaxRetries {
		return bad
	}
	want := "FAIL"
	if complete {
		want = "COMPLETE"
	}
	if notRun {
		want = "NOT_RUN"
	}
	if r.Status != want || !reflect.DeepEqual(metrics(r), r.Metrics) {
		return bad
	}
	return nil
}

func (p *Plan) answerValid(a Answer) bool {
	if a.Type != "choice" || p.dataset.Labels[a.Choice] == "" || len(a.Probabilities) != len(p.dataset.Labels) || !probability(a.Confidence) {
		return false
	}
	sum := 0.0
	for label := range p.dataset.Labels {
		v, ok := a.Probabilities[label]
		if !ok || !probability(v) || v > a.Probabilities[a.Choice] {
			return false
		}
		sum += v
	}
	return math.Abs(sum-1) <= 1e-6
}
func probability(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 1 }

func (p *Plan) ReadReport(raw []byte) (Report, error) {
	var r Report
	if len(raw) > 16<<20 {
		return r, errors.New("evaluation: report exceeds budget")
	}
	if err := decode(raw, &r); err != nil {
		return r, err
	}
	return r, p.Verify(r)
}
