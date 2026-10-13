package chineseeval

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// Verify rejects incomplete or internally inconsistent artifacts. This checks
// completeness and provenance, never a business accuracy threshold. A fixture
// report remains a fixture even when this returns nil.
func (p *Plan) Verify(r Report) error {
	bad := errors.New("evaluation: report integrity or audit failed")
	if !p.schemaValid(r) || (r.Kind != "live" && r.Kind != "fixture") || r.Audit != "PASS" || r.AccountAlias == "" || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.DatasetHash != p.datasetHash || r.ConfigHash != p.configHash || !reflect.DeepEqual(r.Config, p.config) || !reflect.DeepEqual(r.Dataset, p.dataset) || len(r.Samples) != len(p.dataset.Samples) {
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
		case "failed", "version_drift", "missing_answer", "usage_incomplete", "answered", "refused":
			if s.Metadata.RequestID == "" {
				return bad
			}
		default:
			return bad
		}
		if r.Schema == 2 && !sampleTimingValid(r, s, i) {
			return bad
		}
		if s.Metadata.RequestID != "" {
			if ids[s.Metadata.RequestID] {
				return bad
			}
			ids[s.Metadata.RequestID] = true
			if !p.attemptsValid(s) {
				return bad
			}
			for _, a := range s.Metadata.Attempts {
				if a.Usage.Input > math.MaxInt64-b.InputTokens || a.Usage.Output > math.MaxInt64-b.OutputTokens {
					return bad
				}
			}
			b.Calls++
			b.Questions++
			b.add(s.Metadata)
		}
		if s.Status == "answered" || s.Status == "usage_incomplete" || s.Status == "refused" {
			if !p.identity(s.Metadata) || (r.Schema == 1 && s.ResponseModel != p.config.Model) || (s.Status == "refused" && (r.Schema != 2 || s.Answer != nil)) || (s.Status != "refused" && (s.Answer == nil || !p.answerValid(*s.Answer))) {
				return bad
			}
			incomplete := false
			for _, a := range s.Metadata.Attempts {
				if a.UsageReporting != "complete" {
					incomplete = true
				}
			}
			if r.Schema == 1 && incomplete != (s.Status == "usage_incomplete") {
				return bad
			}
		} else if s.Answer != nil {
			return bad
		}
		if (s.Status != "answered" && s.Status != "refused") || !usageComplete(s.Metadata) {
			complete = false
		}
		if s.Status != "not_run" {
			notRun = false
		}
	}
	if (r.Status != "NOT_RUN" && b.Calls == 0) || !reflect.DeepEqual(b, r.Budget) || b.Calls > p.config.MaxCalls || b.Questions > p.config.MaxQuestions || b.Attempts > p.config.MaxAttempts || b.Retries > p.config.MaxRetries {
		return bad
	}
	want := "FAIL"
	if complete {
		want = "COMPLETE"
	}
	if notRun {
		want = "NOT_RUN"
	}
	if r.Status != want || r.ProtocolStatus != protocolStatus(r) || !reflect.DeepEqual(metrics(r), r.Metrics) {
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

// Independent validation at the artifact import boundary protects accounting
// even when an edited report also changes its aggregate counters coherently.
func (p *Plan) attemptsValid(s SampleResult) bool {
	m := s.Metadata
	if len(m.Attempts) > 1 {
		return false
	} // pinned zero-retry task
	if len(m.Attempts) > 0 && (!p.identity(m) || m.TenantID == "" || m.BindingID == "" || m.AccountScopeID == "" || m.BindingVersion == "" || m.CredentialVersion == "") {
		return false
	}
	for i, a := range m.Attempts {
		if a.AttemptID != fmt.Sprintf("%s#%d", m.RequestID, i+1) || a.RetryDelay != 0 || a.HTTPStatus < 0 || a.HTTPStatus > 599 {
			return false
		}
		u := a.Usage
		if p.Route().API == ai.APIOpenAIDecisions && !decisionsUsageValid(u, a.UsageReporting) {
			return false
		}
		if a.UsageReporting != ai.UsageComplete && a.UsageReporting != ai.UsagePartial && a.UsageReporting != ai.UsageUnreported {
			return false
		}
		if u.Input < 0 || u.Output < 0 || u.Input > math.MaxInt64-u.Output || (a.UsageReporting == ai.UsageComplete && u.TotalTokens != u.Input+u.Output) || u.TotalTokens < u.Input+u.Output || u.CacheRead != 0 || u.CacheWrite != 0 || u.Modalities != nil {
			return false
		}
		rates := p.price.CostRates
		threshold := int64(-1)
		for _, t := range p.price.Tiers {
			if u.Input > t.InputTokensAbove && t.InputTokensAbove > threshold {
				rates = t.CostRates
				threshold = t.InputTokensAbove
			}
		}
		cost := float64(rates.Input/1e6) * float64(u.Input)
		if !finiteNonnegative(u.Cost.Input) || !finiteNonnegative(u.Cost.Total) || math.Abs(u.Cost.Input-cost) > 1e-12 || u.Cost.Total != u.Cost.Input || u.Cost.Output != 0 || u.Cost.CacheRead != 0 || u.Cost.CacheWrite != 0 {
			return false
		}
		if a.UsageReporting == ai.UsageUnreported && u != (ai.Usage{}) {
			return false
		}
		if (s.Status == "answered" || s.Status == "usage_incomplete" || s.Status == "refused") && (a.HTTPStatus != 200 || a.Code != "" || a.ProviderRequestID == "") {
			return false
		}
	}
	return true
}
func finiteNonnegative(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }

func (p *Plan) schemaValid(r Report) bool {
	if p.config.Route == nil {
		return r.Schema == 1 && r.Route == nil && r.RouteHash == "" && r.CodeHash == "" && r.ProtocolStatus == ""
	}
	return r.Schema == 2 && r.Route != nil && *r.Route == p.Route() && r.RouteHash == p.Route().Hash() && regexp.MustCompile(`^sha256:[a-f0-9]{64}$`).MatchString(r.CodeHash) && r.FinishedAt.Sub(r.StartedAt) <= time.Duration(p.config.TimeoutSeconds)*time.Second+time.Second
}
func sampleTimingValid(r Report, s SampleResult, i int) bool {
	if s.Metadata.RequestID == "" {
		return s.StartedAt.IsZero() && s.FinishedAt.IsZero() && s.DurationMS == nil
	}
	if s.DurationMS == nil || !finiteNonnegative(*s.DurationMS) || s.StartedAt.Before(r.StartedAt) || s.FinishedAt.Before(s.StartedAt) || s.FinishedAt.After(r.FinishedAt) || *s.DurationMS != float64(s.FinishedAt.Sub(s.StartedAt))/1e6 {
		return false
	}
	return i == 0 || r.Samples[i-1].FinishedAt.IsZero() || !s.StartedAt.Before(r.Samples[i-1].FinishedAt)
}

// Partial canonical usage does not retain output presence. A zero Output alone
// cannot establish an upper bound for reasoning when output was omitted.
func decisionsUsageValid(u ai.Usage, reporting ai.UsageReporting) bool {
	if !u.CacheWrite1h.IsZero() || u.Reasoning.IsNull() {
		return false
	}
	reasoning, present := u.Reasoning.Get()
	if reporting == ai.UsageComplete && !present {
		return false
	}
	return !present || (reasoning >= 0 && ((reporting != ai.UsageComplete && u.Output == 0) || reasoning <= u.Output))
}
