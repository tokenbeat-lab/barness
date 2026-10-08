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
	Schema       int            `json:"schema"`
	Kind         string         `json:"kind"`
	Status       string         `json:"status"`
	Reason       string         `json:"reason,omitempty"`
	Audit        string         `json:"audit"`
	DatasetHash  string         `json:"datasetHash"`
	ConfigHash   string         `json:"configHash"`
	Dataset      Dataset        `json:"dataset"`
	Config       Config         `json:"config"`
	AccountAlias string         `json:"accountAlias"`
	StartedAt    time.Time      `json:"startedAt"`
	FinishedAt   time.Time      `json:"finishedAt"`
	Samples      []SampleResult `json:"samples"`
	Budget       Budget         `json:"budget"`
	Metrics      Metrics        `json:"metrics"`
}

func (p *Plan) initial(kind, alias string) Report {
	// The report is an independent view, so review tools cannot rewrite the plan.
	dataset := p.dataset
	dataset.Samples = slices.Clone(p.dataset.Samples)
	dataset.Labels = maps.Clone(p.dataset.Labels)
	dataset.Distribution = maps.Clone(p.dataset.Distribution)
	r := Report{Schema: 1, Kind: kind, Status: "NOT_RUN", Audit: "PENDING", DatasetHash: p.datasetHash, ConfigHash: p.configHash, Dataset: dataset, Config: p.config, AccountAlias: alias, StartedAt: time.Now().UTC()}
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
		question := ai.ChoiceQuestion{Instructions: jsonString(p.config.Instructions), Criteria: map[string]json.RawMessage{}}
		for key, value := range p.dataset.Labels {
			question.Criteria[key] = jsonString(value)
		}
		scope.RequestID = requestid.New()
		callCtx, callCancel := context.WithTimeout(ctx, time.Duration(p.config.CallTimeoutSeconds)*time.Second)
		res, err := client.Classify(callCtx, scope, ai.Target{BindingID: binding, ModelID: p.config.Model}, ai.ClassifierRequest{State: jsonString(s.Text), Questions: map[string]ai.ClassifierQuestion{p.config.QuestionKey: question}}, nil)
		callCancel()
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
		} else if sample.ResponseModel != p.config.Model || !p.identity(res.Metadata) {
			sample.Status = "version_drift"
		} else if a, ok := res.Answers[p.config.QuestionKey].(ai.ChoiceAnswer); !ok {
			sample.Status = "missing_answer"
		} else {
			sample.Answer = &Answer{Type: "choice", Choice: a.Choice, Probabilities: maps.Clone(a.Probabilities), Confidence: a.Confidence}
			for _, a := range res.Metadata.Attempts {
				if a.UsageReporting != ai.UsageComplete {
					sample.Status = "usage_incomplete"
				}
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
		r.Metrics = metrics(r)
		return r
	}
	r.Status = "COMPLETE"
	for _, s := range r.Samples {
		if s.Status != "answered" {
			r.Status = "FAIL"
		}
	}
	r.FinishedAt = time.Now().UTC()
	r.Metrics = metrics(r)
	return r
}

func (p *Plan) identity(m ai.CallMetadata) bool {
	return m.Resolved && m.Operation == ai.OperationClassifier && m.ProviderID == ai.ProviderTypeSafe && m.API == ai.APITypeSafeSystemOne && m.ModelID == p.config.Model && m.CatalogVersion == p.config.CatalogVersion && m.CatalogHash == p.config.CatalogHash && len(m.Attempts) > 0
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
