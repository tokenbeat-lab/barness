package chineseeval

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

// Answer is the evaluation view, independently converted from public model answers.
type Answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

type SampleResult struct {
	StartedAt     time.Time       `json:"startedAt,omitzero"`
	FinishedAt    time.Time       `json:"finishedAt,omitzero"`
	DurationMS    *float64        `json:"durationMS,omitempty"`
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	ResponseModel string          `json:"responseModel,omitempty"`
	Answer        *Answer         `json:"answer,omitempty"`
	Code          ai.Code         `json:"code,omitempty"`
	Phase         ai.Phase        `json:"phase,omitempty"`
	Metadata      ai.CallMetadata `json:"metadata"`
}

type Budget struct {
	Calls        int     `json:"calls"`
	Questions    int     `json:"questions"`
	Attempts     int     `json:"attempts"`
	Retries      int     `json:"retries"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	KnownCostUSD float64 `json:"knownCostUSD"`
	Unreported   int     `json:"unreported"`
	Partial      int     `json:"partial"`
}

type Report struct {
	Route          *Route         `json:"route,omitempty"`
	RouteHash      string         `json:"routeHash,omitempty"`
	CodeHash       string         `json:"codeHash,omitempty"`
	ProtocolStatus string         `json:"protocolStatus,omitempty"`
	Schema         int            `json:"schema"`
	Kind           string         `json:"kind"`
	Status         string         `json:"status"`
	Reason         string         `json:"reason,omitempty"`
	Audit          string         `json:"audit"`
	DatasetHash    string         `json:"datasetHash"`
	ConfigHash     string         `json:"configHash"`
	Dataset        Dataset        `json:"dataset"`
	Config         Config         `json:"config"`
	AccountAlias   string         `json:"accountAlias"`
	StartedAt      time.Time      `json:"startedAt"`
	FinishedAt     time.Time      `json:"finishedAt"`
	Samples        []SampleResult `json:"samples"`
	Budget         Budget         `json:"budget"`
	Metrics        Metrics        `json:"metrics"`
}

func (p *Plan) initial(kind, alias string) Report {
	// The report is an independent view, so review tools cannot rewrite the plan.
	dataset := p.dataset
	dataset.Samples = slices.Clone(p.dataset.Samples)
	dataset.Labels = maps.Clone(p.dataset.Labels)
	dataset.Distribution = maps.Clone(p.dataset.Distribution)
	r := Report{Schema: 1, Kind: kind, Status: "NOT_RUN", Audit: "PENDING", DatasetHash: p.datasetHash, ConfigHash: p.configHash, Dataset: dataset, Config: p.config, AccountAlias: alias, StartedAt: time.Now().UTC()}
	if p.config.Route != nil {
		route := p.Route()
		r.Schema = 2
		r.Route = &route
		r.RouteHash = route.Hash()
		r.Config = p.Config()
	}
	for _, s := range p.dataset.Samples {
		r.Samples = append(r.Samples, SampleResult{ID: s.ID, Status: "not_run"})
	}
	return r
}

// NotRun preserves every sample, with absent scores rather than a fabricated 0.
func (p *Plan) NotRun(reason string) Report {
	r := p.initial("live", "unreported")
	r.Reason = reason
	r.FinishedAt = time.Now().UTC()
	r.ProtocolStatus = protocolStatus(r)
	r.Metrics = metrics(r)
	return r
}

// Run uses the public Client and a host-authorized binding. There are no host
// retries; the example's binding must also have MaxRetries=0. All actual
// attempts, including failed and cancelled ones, retain their reported usage.
func (p *Plan) Run(ctx context.Context, client *ai.Client, scope ai.CallScope, binding, kind, alias string) Report {
	r := p.initial(kind, alias)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)
	defer cancel()
	for i, s := range p.dataset.Samples {
		sample := &r.Samples[i]
		if ctx.Err() != nil {
			sample.Status = "cancelled"
			continue
		}
		if r.Budget.Calls >= p.config.MaxCalls || r.Budget.Questions >= p.config.MaxQuestions || r.Budget.Attempts >= p.config.MaxAttempts {
			sample.Status = "budget_exhausted"
			continue
		}
		input := p.request(s)
		scope.RequestID = requestid.New()
		callCtx, callCancel := context.WithTimeout(ctx, time.Duration(p.config.CallTimeoutSeconds)*time.Second)
		start := time.Now()
		res, err := client.Classify(callCtx, scope, ai.Target{BindingID: binding, ModelID: p.config.Model}, input, nil)
		callCancel()
		if r.Schema == 2 {
			sample.StartedAt = start.UTC()
			sample.FinishedAt = time.Now().UTC()
			sample.DurationMS = number(float64(sample.FinishedAt.Sub(sample.StartedAt)) / 1e6)
		}
		r.Budget.Calls++
		r.Budget.Questions++
		sample.Metadata = res.Metadata
		sample.ResponseModel, _ = res.ResponseModel.Get()
		sample.Status = "answered"
		r.Budget.add(res.Metadata)
		if err != nil {
			sample.Status = "failed"
			var ae *ai.Error
			if errors.As(err, &ae) {
				sample.Code, sample.Phase = ae.Code, ae.Phase
				if ae.Code == ai.CodeCanceled {
					sample.Status = "cancelled"
				}
			}
		} else if !p.identity(res.Metadata) || (r.Schema == 1 && sample.ResponseModel != p.config.Model) {
			sample.Status = "version_drift"
		} else {
			// nil error includes RefusalAnswer. This closed switch converts at
			// the Host boundary; business review/requestioning belongs here.
			switch a := res.Answers[p.config.QuestionKey].(type) {
			case ai.ChoiceAnswer:
				sample.Answer = &Answer{Type: "choice", Choice: a.Choice, Probabilities: maps.Clone(a.Probabilities), Confidence: a.Confidence}
			case ai.RefusalAnswer:
				sample.Status = "refused"
			case ai.ScoreAnswer, ai.BoolAnswer:
				sample.Status = "missing_answer" // this task requests choice only
			default:
				sample.Status = "missing_answer"
			}
			if r.Schema == 1 && !usageComplete(sample.Metadata) && sample.Status == "answered" {
				sample.Status = "usage_incomplete"
			}
		}
	}
	if r.Budget.Calls == 0 {
		r.Status = "NOT_RUN"
		r.Reason = "cancelled before first logical call"
		for i := range r.Samples {
			r.Samples[i].Status = "not_run"
		}
		r.FinishedAt = time.Now().UTC()
		r.ProtocolStatus = protocolStatus(r)
		r.Metrics = metrics(r)
		return r
	}
	r.Status = "COMPLETE"
	for _, s := range r.Samples {
		if (s.Status != "answered" && s.Status != "refused") || !usageComplete(s.Metadata) {
			r.Status = "FAIL"
		}
	}
	r.FinishedAt = time.Now().UTC()
	r.ProtocolStatus = protocolStatus(r)
	r.Metrics = metrics(r)
	return r
}

func (p *Plan) identity(m ai.CallMetadata) bool {
	return m.Resolved && m.Operation == ai.OperationClassifier && m.ProviderID == p.Route().Provider && m.API == p.Route().API && m.ModelID == p.config.Model && m.CatalogVersion == p.config.CatalogVersion && m.CatalogHash == p.config.CatalogHash && len(m.Attempts) > 0
}

func (b *Budget) add(m ai.CallMetadata) {
	b.Attempts += len(m.Attempts)
	b.Retries += max(0, len(m.Attempts)-1)
	for _, a := range m.Attempts {
		b.InputTokens += a.Usage.Input
		b.OutputTokens += a.Usage.Output
		b.KnownCostUSD += a.Usage.Cost.Total
		if a.UsageReporting == ai.UsageUnreported {
			b.Unreported++
		}
		if a.UsageReporting == ai.UsagePartial {
			b.Partial++
		}
	}
}

func usageComplete(m ai.CallMetadata) bool {
	if len(m.Attempts) != 1 {
		return false
	}
	return m.Attempts[0].UsageReporting == ai.UsageComplete
}
func protocolStatus(r Report) string {
	if r.Schema == 1 {
		return ""
	}
	if r.Budget.Calls == 0 {
		return "NOT_RUN"
	}
	for _, s := range r.Samples {
		if s.Metadata.RequestID != "" && s.Status != "answered" && s.Status != "refused" {
			return "FAIL"
		}
	}
	return "COMPLETE"
}

// Request constructs an independent public request for a recorded sample. The
// execution loop and artifact replay share the task definition at this boundary.

func (p *Plan) Request(sampleID string) (ai.ClassifierRequest, error) {
	for _, s := range p.dataset.Samples {
		if s.ID == sampleID {
			return p.request(s), nil
		}
	}
	return ai.ClassifierRequest{}, errors.New("evaluation: unknown task sample")
}
func (p *Plan) request(s Sample) ai.ClassifierRequest {
	q := ai.ChoiceQuestion{Instructions: jsonString(p.config.Instructions), Criteria: map[string]json.RawMessage{}}
	for key, value := range p.dataset.Labels {
		q.Criteria[key] = jsonString(value)
	}
	return ai.ClassifierRequest{State: jsonString(s.Text), Questions: map[string]ai.ClassifierQuestion{p.config.QuestionKey: q}}
}
