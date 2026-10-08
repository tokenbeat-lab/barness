package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

const mixedModel = "same-model"

// Each tenant uses the same names. OpenAI chat/image deliberately share an
// account and credential reference; Google/TypeSafe have independent accounts.
type mixedRoute struct {
	id       string
	op       ai.Operation
	provider ai.ProviderID
	api      ai.API
}

var mixedRoutes = []mixedRoute{
	{"chat", ai.OperationChat, ai.ProviderOpenAI, ai.APIOpenAIResponses},
	{"openai-image", ai.OperationImage, ai.ProviderOpenAI, ai.APIOpenAIImages},
	{"google-image", ai.OperationImage, ai.ProviderGoogle, ai.APIGoogleInteractions},
	{"classifier", ai.OperationClassifier, ai.ProviderTypeSafe, ai.APITypeSafeSystemOne},
}

func (r mixedRoute) prefix(k tenantKey) string { return "/" + k.tenant + "/" + r.id }
func (r mixedRoute) binding(k tenantKey, url string) ai.Binding {
	b := ai.Binding{TenantID: k.tenant, BindingID: r.id, Operation: r.op, Version: "b1", Enabled: true,
		ProviderID: r.provider, API: r.api, Endpoint: url + r.prefix(k), AuthKind: ai.AuthAPIKey,
		AccountScopeID: fmt.Sprintf("acct-%s-%s", k.tenant, r.provider), CredentialRef: string(r.provider), AllowedModels: []string{mixedModel}}
	if r.provider == ai.ProviderGoogle {
		b.Endpoint += "/v1beta"
	}
	return b
}

func mixedCatalog() ai.Catalog {
	c := operationCatalog()
	c.Version = "mixed-15-synthetic"
	c.Models[0].ID = mixedModel
	cfg := ai.Config{Policy: validPolicy()}
	configureImages(&cfg)
	c.ImageModels = cfg.Catalog.ImageModels
	c.ImageModels[0].ID = mixedModel
	configureGoogleImages(&cfg)
	m := cfg.Catalog.ImageModels[0]
	m.ID = mixedModel
	c.ImageModels = append(c.ImageModels, m)
	configureClassifier(&cfg)
	c.ClassifierModels = cfg.Catalog.ClassifierModels
	c.ClassifierModels[0].ID = mixedModel
	return c
}

type mixedWorld struct {
	client   *ai.Client
	host     *host.Host
	provider *provider.Server
	gauges   *probe.Probe
	bodies   *provider.BodyTracker
	rec      *recorder
	adm      *host.Admission
}

func newMixedWorld(t *testing.T, configure func(*ai.Config), tenants ...tenantKey) *mixedWorld {
	t.Helper()
	w := &mixedWorld{host: host.New(), provider: localProvider(t), gauges: probe.New(),
		bodies: provider.TrackBodies(provider.LoopbackTransport()), rec: newRecorder(), adm: host.NewAdmission()}
	cat := mixedCatalog()
	cfg := ai.Config{Policy: hostintegration.CloudMixedPolicy(), Catalog: &cat, Bindings: w.host, Credentials: w.host,
		Transport: w.bodies, Probe: w.gauges, Observer: w.rec, Admission: w.adm, AllowLoopbackHTTP: true}
	if configure != nil {
		configure(&cfg)
	}
	var err error
	w.client, err = ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range tenants {
		w.install(k)
	}
	return w
}

func (w *mixedWorld) install(k tenantKey) {
	for _, r := range mixedRoutes {
		b := r.binding(k, w.provider.URL())
		w.host.PutBinding(b)
		w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: b.CredentialRef, Version: "v1",
			AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(k.secret)})
	}
}

type mixedOutcome struct {
	metadata   ai.CallMetadata
	usage      ai.Usage
	err        error
	chat       ai.Result
	images     ai.ImagesResult
	classifier ai.ClassifierResult
}

func (o mixedOutcome) recordResult(ev *evidence.Case, name string) {
	switch o.metadata.Operation {
	case ai.OperationChat:
		ev.Record(name, o.chat)
	case ai.OperationImage:
		ev.Record(name, o.images)
	case ai.OperationClassifier:
		ev.Record(name, o.classifier)
	}
}

func (r mixedRoute) invoke(ctx context.Context, c *ai.Client, scope ai.CallScope, target ai.Target) mixedOutcome {
	return r.invokeHooked(ctx, c, scope, target, ai.Hooks{})
}

func (r mixedRoute) invokeHooked(ctx context.Context, c *ai.Client, scope ai.CallScope, target ai.Target, hooks ai.Hooks) mixedOutcome {
	h := c.WithHooks(hooks)
	switch r.op {
	case ai.OperationChat:
		res, err := h.Complete(ctx, scope, target, ai.Request{Messages: []ai.Message{ai.UserText("synthetic mixed chat " + scope.TenantID)}}, nil)
		return mixedOutcome{metadata: res.Metadata, usage: res.Message.Usage, err: err, chat: res}
	case ai.OperationImage:
		res, err := h.GenerateImages(ctx, scope, target, ai.ImagesRequest{Prompt: "synthetic mixed image " + scope.TenantID}, nil)
		return mixedOutcome{metadata: res.Metadata, usage: res.Usage, err: err, images: res}
	default:
		input := mixedClassifierInput()
		state, err := json.Marshal(map[string]string{"ticket": "synthetic", "owner": scope.TenantID})
		if err != nil {
			return mixedOutcome{err: err}
		}
		input.State = state
		res, err := h.Classify(ctx, scope, target, input, nil)
		return mixedOutcome{metadata: res.Metadata, usage: res.Usage, err: err, classifier: res}
	}
}

func mixedClassifierInput() ai.ClassifierRequest {
	return ai.ClassifierRequest{State: json.RawMessage(`{"ticket":"synthetic"}`), Questions: map[string]ai.ClassifierQuestion{
		"intent": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Choose intent"`), Criteria: map[string]json.RawMessage{"refund": json.RawMessage(`"Refund"`), "track": json.RawMessage(`"Tracking"`)}}}}
}

func (r mixedRoute) reply(t *testing.T) provider.Reply {
	t.Helper()
	if r.op == ai.OperationChat {
		f, _ := loadTextFixture(t, "text-basic.json")
		return sseReply(t, f, provider.FramingLF)
	}
	var sc fixtureScenario
	if r.op == ai.OperationClassifier {
		f, _ := loadFixture(t, classifierProtocol, "choice.json")
		sc = f.Scenarios[0]
	} else if r.provider == ai.ProviderGoogle {
		sc = firstGoogleImagesScenario(t)
	} else {
		sc = firstImagesScenario(t)
	}
	return sc.replies(t)[0]
}

func (w *mixedWorld) record(ev *evidence.Case, o mixedOutcome) {
	id := o.metadata.RequestID
	ev.Record(id+"-metadata", o.metadata)
	ev.Record(id+"-usage", o.usage)
	ev.Record(id+"-error", errString(o.err))
	obs := w.rec.awaitCall(ev, id)
	ev.Record("observations-"+id, obs)
	base := o.metadata
	base.Attempts = nil
	starts, finishes := 0, 0
	attemptStarts, attemptFinishes := map[string]int{}, map[string]int{}
	for _, v := range obs {
		ev.Check("observation keeps tenant, operation and binding", v.Call.TenantID == o.metadata.TenantID && v.Call.Operation == o.metadata.Operation && v.Call.BindingID == o.metadata.BindingID, "got %+v", v.Call)
		switch v.Kind {
		case ai.ObservationCallStarted:
			starts++
			want := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: o.metadata.TenantID, RequestID: id, ActorID: o.metadata.ActorID, JobID: o.metadata.JobID, BindingID: o.metadata.BindingID, Operation: o.metadata.Operation}}
			ev.Check("unresolved entry attribution", reflect.DeepEqual(v.Call, want), "got %+v", v.Call)
		case ai.ObservationCallFinished:
			finishes++
			ev.Check("observed immutable metadata and usage", reflect.DeepEqual(v.Call, o.metadata) && reflect.DeepEqual(v.Usage, o.usage), "got %+v", v)
		default:
			ev.Check("attempt record has complete snapshot", reflect.DeepEqual(v.Call, base) && v.Attempt != nil, "got %+v", v.Call)
			if v.Attempt == nil {
				continue
			}
			if v.Kind == ai.ObservationAttemptStarted {
				attemptStarts[v.Attempt.AttemptID]++
			} else {
				attemptFinishes[v.Attempt.AttemptID]++
			}
			if v.Kind == ai.ObservationAttemptFinished {
				found := false
				for _, a := range o.metadata.Attempts {
					if a.AttemptID == v.Attempt.AttemptID {
						found = true
						ev.Check("attempt Observer usage matches own result attempt", reflect.DeepEqual(*v.Attempt, a), "got %+v", v.Attempt)
					}
				}
				ev.Check("no unrelated observed attempt", found, "got %s", v.Attempt.AttemptID)
			}
		}
	}
	ev.Check("one logical call and no invented attempts", starts == 1 && finishes == 1 && len(attemptStarts) == len(o.metadata.Attempts) && len(attemptFinishes) == len(o.metadata.Attempts), "starts=%d finishes=%d", starts, finishes)
	for _, a := range o.metadata.Attempts {
		ev.Check("each own attempt observed exactly once", attemptStarts[a.AttemptID] == 1 && attemptFinishes[a.AttemptID] == 1, "attempt %s", a.AttemptID)
	}
}

func (w *mixedWorld) released(ev *evidence.Case) {
	eventually(ev, "all mixed calls finished", func() bool { return w.gauges.ActiveCalls() == 0 })
	state := map[string]int64{"calls": w.gauges.ActiveCalls(), "permits": w.gauges.Permits(), "waiters": w.gauges.AdmissionWaiters(), "events": w.gauges.QueuedEvents(), "bodies": w.bodies.Open(), "hostPermits": int64(w.adm.Held())}
	ev.Record("resources", state)
	ev.Record("host-admission", w.adm.Requests())
	for name, n := range state {
		ev.Check("released "+name, n == 0, "got %d", n)
	}
}
