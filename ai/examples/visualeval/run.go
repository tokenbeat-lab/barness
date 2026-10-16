package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type failure struct {
	Code  ai.Code  `json:"code"`
	Phase ai.Phase `json:"phase"`
}
type resultView struct {
	Answers       map[string]json.RawMessage `json:"answers"`
	ImageInputs   []ai.ClassifierImageInput  `json:"imageInputs"`
	ResponseModel ai.Nullable[string]        `json:"responseModel"`
	Usage         ai.Usage                   `json:"usage"`
	Metadata      ai.CallMetadata            `json:"metadata"`
}
type sampleResult struct {
	ID         string                `json:"id"`
	Group      string                `json:"group"`
	Status     string                `json:"status"`
	View       map[string]answerView `json:"view"`
	Result     resultView            `json:"result"`
	Error      *failure              `json:"error"`
	IoU        *float64              `json:"iou"`
	Consumed   bool                  `json:"consumed"`
	Resources  map[string]int64      `json:"resources"`
	Assertions map[string]bool       `json:"assertions"`
	StartedAt  time.Time             `json:"startedAt"`
	FinishedAt time.Time             `json:"finishedAt"`
}
type budget struct {
	Calls         int     `json:"calls"`
	Attempts      int     `json:"attempts"`
	Retries       int     `json:"retries"`
	Questions     int     `json:"questions"`
	Occurrences   int     `json:"occurrences"`
	SourceBytes   int     `json:"sourceBytes"`
	ExpandedBytes int     `json:"expandedBytes"`
	InputTokens   int64   `json:"inputTokens"`
	OutputTokens  int64   `json:"outputTokens"`
	KnownCostUSD  float64 `json:"knownCostUSD"`
	Unreported    int     `json:"unreported"`
	Partial       int     `json:"partial"`
}
type report struct {
	CatalogKind          string         `json:"catalogKind,omitempty"`
	Schema               int            `json:"schema"`
	Mode                 string         `json:"mode"`
	Route                string         `json:"route"`
	Entry                string         `json:"entry"`
	Status               string         `json:"status"`
	RealQuality          string         `json:"realQuality"`
	Reason               string         `json:"reason"`
	CatalogVersion       string         `json:"catalogVersion"`
	CatalogHash          string         `json:"catalogHash"`
	Model                string         `json:"model"`
	DatasetHash          string         `json:"datasetHash"`
	ConfigHash           string         `json:"configHash"`
	CodeHash             string         `json:"codeHash"`
	ProtocolEvidenceHash string         `json:"protocolEvidenceHash"`
	StartedAt            time.Time      `json:"startedAt"`
	FinishedAt           time.Time      `json:"finishedAt"`
	Samples              []sampleResult `json:"samples"`
	Budget               budget         `json:"budget"`
	Metrics              metrics        `json:"metrics"`
}

func initial(p *plan, a arguments) report {
	hash, _ := p.catalog.Hash()
	r := report{Schema: 1, Mode: a.mode, Route: p.route, Entry: a.entry, Status: "NOT_RUN", RealQuality: "NOT_RUN", CatalogVersion: p.catalog.Version, CatalogHash: hash, Model: p.model.ID, DatasetHash: p.pins["dataset.json"], ConfigHash: p.pins["experiment.json"], StartedAt: time.Now().UTC()}
	if p.formal {
		r.CatalogKind = "formal"
	}
	for i, s := range p.dataset.Samples {
		r.Samples = append(r.Samples, sampleResult{ID: s.ID, Group: p.group(i), Status: "not_run", View: map[string]answerView{}})
	}
	return r
}
func (b *budget) add(s sampleResult, req ai.ClassifierRequest) {
	b.Calls++
	b.Questions += len(req.Questions)
	count := max(1, len(req.ImageReferences))
	b.Occurrences += count
	raw, _ := data.ReadFile("testdata/" + s.ID + fileSuffix(req.Images[0].MimeType))
	b.SourceBytes += len(raw)
	b.ExpandedBytes += len(raw) * count
	b.Attempts += len(s.Result.Metadata.Attempts)
	b.Retries += max(0, len(s.Result.Metadata.Attempts)-1)
	for _, a := range s.Result.Metadata.Attempts {
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
func fileSuffix(mime string) string {
	if mime == "image/jpeg" {
		return ".jpg"
	}
	return ".png"
}
func executeCalls(ctx context.Context, p *plan, a arguments, b *bundle, r *report, endpoint, key, alias string, transport http.RoundTripper) error {
	for i := range p.dataset.Samples {
		if ctx.Err() != nil {
			r.Samples[i].Status = "cancelled"
			continue
		}
		req := p.request(i)
		n := max(1, len(req.ImageReferences))
		raw, _ := data.ReadFile("testdata/" + p.dataset.Samples[i].File)
		c := p.experiment
		if r.Budget.Calls >= a.calls || r.Budget.Attempts >= c.MaxAttempts || r.Budget.Questions+len(req.Questions) > c.MaxQuestions || r.Budget.Occurrences+n > c.MaxOccurrences || r.Budget.SourceBytes+len(raw) > c.MaxSourceBytes || r.Budget.ExpandedBytes+len(raw)*n > c.MaxExpandedBytes {
			r.Samples[i].Status = "budget_exhausted"
			continue
		}
		s, wire, obs, err := call(ctx, p, a, req, i, endpoint, key, alias, transport)
		if err != nil {
			return err
		}
		r.Samples[i] = s
		r.Budget.add(s, req)
		prefix := fmt.Sprintf("samples/%s/", s.ID)
		if err = b.record(prefix+"input.json", req); err != nil {
			return err
		}
		if err = b.record(prefix+"wire.json", wire); err != nil {
			return err
		}
		if err = b.file(prefix+"request.json", wire.Request); err != nil {
			return err
		}
		if err = b.file(prefix+"response.bin", wire.Response); err != nil {
			return err
		}
		if err = b.record(prefix+"observations.json", obs); err != nil {
			return err
		}
		if err = b.record(prefix+"result.json", s); err != nil {
			return err
		}
	}
	finalize(p, r)
	return nil
}
func call(ctx context.Context, p *plan, a arguments, req ai.ClassifierRequest, i int, endpoint, key, alias string, transport http.RoundTripper) (sampleResult, wire, []ai.Observation, error) {
	s := sampleResult{ID: p.dataset.Samples[i].ID, Group: p.group(i), Status: "answered", StartedAt: time.Now().UTC(), Assertions: map[string]bool{}}
	h := &host{binding: p.binding(endpoint, alias), key: key}
	o := &observations{done: make(chan struct{})}
	g := probe.New()
	c := &capture{next: transport}
	tracker := provider.TrackBodies(c)
	client, err := ai.NewClient(ai.Config{Catalog: &p.catalog, Policy: policy(time.Duration(a.callMS) * time.Millisecond), Bindings: h, Credentials: h, Observer: o, Transport: tracker, AllowLoopbackHTTP: a.mode == "controlled", Probe: g})
	if err != nil {
		return s, wire{}, nil, err
	}
	var opts ai.Options = ai.OpenAIDecisionsOptions{}
	if p.route == "jev" {
		opts = ai.TypeSafeOptions{ExperimentalImages: true}
	}
	callctx, cancel := context.WithTimeout(ctx, time.Duration(a.callMS)*time.Millisecond)
	scope := ai.CallScope{TenantID: h.binding.TenantID, RequestID: requestid.New()}
	target := ai.Target{BindingID: h.binding.BindingID, ModelID: p.model.ID}
	var result ai.ClassifierResult
	if a.entry == "hooked" {
		result, err = client.WithHooks(ai.Hooks{}).Classify(callctx, scope, target, req, opts)
	} else {
		result, err = client.Classify(callctx, scope, target, req, opts)
	}
	cancel()
	s.FinishedAt = time.Now().UTC()
	answers := map[string]json.RawMessage{}
	for k, v := range result.Answers {
		raw, e := json.Marshal(v)
		if e != nil {
			return s, wire{}, nil, e
		}
		answers[k] = raw
	}
	s.Result = resultView{answers, result.ImageInputs, result.ResponseModel, result.Usage, result.Metadata}
	if err != nil {
		var ae *ai.Error
		if !errors.As(err, &ae) {
			return s, wire{}, nil, errors.New("untyped public error")
		}
		s.Error = &failure{ae.Code, ae.Phase}
	}
	assess(p, i, &s)
	wait, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	records := o.wait(wait)
	s.Assertions["observerTerminal"] = len(records) > 0 && records[len(records)-1].Kind == ai.ObservationCallFinished && reflect.DeepEqual(records[len(records)-1].Call, result.Metadata)
	s.Assertions["observerDelivered"] = client.ObserverStats().Dropped == 0 && client.ObserverStats().Failed == 0
	s.Assertions["credentialSnapshot"] = h.reads == 1
	s.Resources = map[string]int64{"bodies": tracker.Open(), "calls": g.ActiveCalls(), "permits": g.Permits(), "waiters": g.AdmissionWaiters()}
	for k, v := range s.Resources {
		s.Assertions["released-"+k] = v == 0
	}
	return s, c.snapshot(), records, nil
}
func assess(p *plan, i int, s *sampleResult) {
	s.View = map[string]answerView{}
	s.IoU = nil
	s.Consumed = false
	if s.Error != nil {
		s.Status = "failed"
		if s.Error.Code == ai.CodeCanceled || s.Error.Code == ai.CodeDeadlineExceeded {
			s.Status = "cancelled"
		}
		return
	}
	m := s.Result.Metadata
	hash, _ := p.catalog.Hash()
	model, _ := s.Result.ResponseModel.Get()
	if !m.Resolved || m.Operation != ai.OperationClassifier || m.ProviderID != p.model.Provider || m.API != p.model.API || m.ModelID != p.model.ID || m.CatalogVersion != p.catalog.Version || m.CatalogHash != hash || model != p.model.ID {
		s.Status = "version_drift"
		return
	}
	s.Status = "answered"
	s.View = views(s.Result.Answers)
	for k := range p.request(i).Questions {
		if s.View[k].Status == "refused" {
			s.Status = "refused"
		}
		if s.View[k].Status == "" {
			s.Status = "missing_answer"
		}
	}
	box := s.View["location"].Box
	exists := s.View["exists"].Probability
	if box != nil && p.dataset.Samples[i].Box != nil {
		s.IoU = number(iou(*box, *p.dataset.Samples[i].Box))
	}
	s.Consumed = box != nil && exists != nil && *exists >= p.experiment.ExistenceThreshold
}
func finalize(p *plan, r *report) {
	r.Status = "COMPLETE"
	if r.Budget.Calls == 0 {
		r.Status = "NOT_RUN"
		if r.Reason == "" {
			r.Reason = "work cancelled before first logical call"
		}
	} else {
		for _, s := range r.Samples {
			if (s.Status != "answered" && s.Status != "refused") || len(s.Result.Metadata.Attempts) != 1 {
				r.Status = "FAIL"
			}
			for _, ok := range s.Assertions {
				if !ok {
					r.Status = "FAIL"
				}
			}
		}
		if r.Budget.Partial+r.Budget.Unreported > 0 {
			r.Status = "FAIL"
		}
	}
	r.RealQuality = "NOT_RUN"
	if r.Mode == "live" && r.Budget.Calls > 0 {
		r.RealQuality = r.Status
	}
	r.Metrics = score(p, *r)
	r.FinishedAt = time.Now().UTC()
}
