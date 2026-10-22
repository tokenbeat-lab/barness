package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var kimiConcurrentKeys = []tenantKey{
	{tenantA.tenant, "synthetic-kimi-host-a-intl", "key:kimi-host-a-intl@v1"},
	{tenantA.tenant, "synthetic-kimi-host-a-cn", "key:kimi-host-a-cn@v1"},
	{tenantB.tenant, "synthetic-kimi-host-b-intl", "key:kimi-host-b-intl@v1"},
	{tenantB.tenant, "synthetic-kimi-host-b-cn", "key:kimi-host-b-cn@v1"},
}

func init() { knownKeys = append(knownKeys, kimiConcurrentKeys...) }

type kimiConcurrentJob struct {
	id               string
	key              tenantKey
	binding          ai.Binding
	request          ai.Request
	sampling         map[string]json.RawMessage
	stream           *ai.Stream
	result           outcome
	envelopes        []ai.EventEnvelope
	input            int64
	minimum, maximum float64
}

// Twelve finite calls share one Client and a deliberately identical account
// label/session, but preserve four separately authorized tenant/market secrets.
// Header and socket barriers prove simultaneous pinned calls without sleeps,
// timing assumptions, throughput load or a race-stage run.
func TestKimiHostConcurrentMarketsAndTenants(t *testing.T) {
	ev := run.Case(t, "KIMI-CN-INTL-K02-K09-K13-K14-concurrent-host-snapshots")
	cat := kimi.CandidateCatalog()
	rec := newRecorder()
	gauges := probe.New()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	resolvers := map[string]ai.CredentialResolver{}
	reads := map[string]int{}
	var readsMu sync.Mutex
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Observer = rec
		c.Policy.MaxQueuedObservations = 64
		c.Probe = gauges
		c.Transport = bodies
		c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
			resolver, ok := resolvers[s.TenantID+"/"+string(b.ProviderID)]
			if !ok {
				return ai.Credential{}, ai.ErrAccessDenied
			}
			return resolver.ResolveCredential(ctx, s, b)
		})
	}, tenantA, tenantB)
	jobs := []*kimiConcurrentJob{}
	// Independent worked USD expectations, from candidate rates and each
	// scripted counter tuple. Chat has80 writes of unknown TTL; Responses/Messages have0.
	lows := []float64{.000330, .000087, .000090, .000336209461474, .0000952097590008, .0000981850639696, .000348, .000105, .000108, .0003540612912868, .0001130615888136, .0001160368937824}
	highs := []float64{.000570, .000087, .000090, .000574233858978, .0000952097590008, .0000981850639696, .000588, .000105, .000108, .0005920856887908, .0001130615888136, .0001160368937824}
	for ti, tenant := range []tenantKey{tenantA, tenantB} {
		for pi, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
			key := kimiConcurrentKeys[ti*2+pi]
			owner := tenant.tenant + "/" + string(market)
			endpoints := map[ai.API]string{}
			for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
				endpoints[api] = w.provider.URL() + "/kimi-concurrent/" + owner + "/" + string(api)
			}
			auth := kimihost.Authorization{ProviderID: market, TenantID: tenant.tenant, AccountScopeID: "same-account-alias", CredentialRef: owner + "-record", Version: "binding-v1", Purpose: kimihost.Standard, Endpoints: endpoints, TrustedGateway: true}
			bindings, err := kimihost.Bindings(auth)
			if err != nil {
				t.Fatal(err)
			}
			resolvers[owner] = kimihost.Credentials{Authorization: auth, Record: kimihost.CredentialRecord{ProviderID: market, TenantID: tenant.tenant, AccountScopeID: auth.AccountScopeID, CredentialRef: auth.CredentialRef, Purpose: kimihost.Standard, Version: "secret-v1", BindingVersion: auth.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}, ReadSecret: func(_ context.Context, ref string) (ai.Secret, error) {
				readsMu.Lock()
				defer readsMu.Unlock()
				if ref != auth.CredentialRef {
					return ai.Secret{}, ai.ErrAccessDenied
				}
				reads[owner]++
				return ai.NewSecret(key.secret), nil
			}}
			for _, binding := range bindings {
				w.host.PutBinding(binding)
				n := len(jobs)
				id := fmt.Sprintf("kimi-concurrent-%d", n)
				jobs = append(jobs, &kimiConcurrentJob{id: id, key: key, binding: binding, request: ai.Request{Messages: []ai.Message{ai.UserText("original-input-" + id)}}, sampling: map[string]json.RawMessage{"temperature": json.RawMessage(`1`)}, input: int64(8 + n), minimum: lows[n], maximum: highs[n]})
			}
		}
	}
	entered := make(chan string, len(jobs))
	resume := make(chan struct{})
	var sockets atomic.Int64
	simultaneous := make(chan int64, 1)
	var arrivals sync.WaitGroup
	arrivals.Add(len(jobs))
	allSent := doneWhen(arrivals.Wait)
	for _, job := range jobs {
		reply := kimiConcurrentReply(t, job)
		ev.Fixture(job.id+"-synthetic.sse", []byte(reply.Script))
		reply.OnReceive = func() {
			if sockets.Add(1) == int64(len(jobs)) {
				simultaneous <- gauges.Permits()
			}
			arrivals.Done()
			waitFor(ev, "all twelve requests are in flight", allSent)
		}
		w.provider.EnqueueAt(strings.TrimPrefix(job.binding.Endpoint, w.provider.URL()), reply)
		var retained map[string]any
		hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("X-Host-Trace", job.id)
			entered <- job.id
			select {
			case <-resume:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			retained = p.Body
			return ai.KeepPayload(), nil
		}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
			retained["model"] = "callback-mutated-after-serialization"
			return nil
		}}
		scope := ai.CallScope{TenantID: job.binding.TenantID, RequestID: job.id, ActorID: "actor-" + job.binding.TenantID, JobID: "job-" + job.id}
		var opts ai.Options = ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
		if job.binding.API == ai.APIOpenAICompletions {
			opts = ai.ChatOptions{MaxTokens: ai.Value(4096), SessionID: "same-private-session", SamplingParams: job.sampling}
		} else if job.binding.API == ai.APIOpenAIResponses {
			opts = ai.ResponsesOptions{MaxTokens: ai.Value(4096), SessionID: "same-private-session"}
		}
		job.stream = w.client.WithHooks(hooks).Stream(ctxFor(t), scope, ai.Target{BindingID: job.binding.BindingID, ModelID: "kimi-k3"}, job.request, opts)
		defer job.stream.Close()
	}
	for range jobs {
		waitValue(ev, "pinned call reached header barrier", entered)
	}
	for _, job := range jobs {
		job.request.Messages[0] = ai.UserText("caller-mutated-input")
		job.sampling["temperature"][0] = '9'
	}
	for i := range cat.Models {
		cat.Models[i].Cost.Input = 999
		cat.Models[i].Input[0] = ai.ModalityImage
	}
	close(resume)
	var done sync.WaitGroup
	for _, job := range jobs {
		done.Go(func() {
			o := outcome{streamed: true}
			for job.stream.Next() {
				o.events = append(o.events, job.stream.Event())
				job.envelopes = append(job.envelopes, job.stream.Envelope())
			}
			o.streamErr = job.stream.Err()
			o.result, o.err = job.stream.Result()
			job.stream.Close()
			job.result = o
		})
	}
	done.Wait()
	held := waitValue(ev, "all authorized sockets reached deterministic in-flight barrier", simultaneous)
	ev.Record("simultaneous-permits", held)
	ev.Check("twelve independently attributed calls overlap under the shared admission policy", held == 12, "held=%d", held)
	captures := w.provider.Requests()
	ev.Record("requests", captures)
	keys := map[string]bool{}
	ownerKeys := map[string]string{}
	for _, job := range jobs {
		o := job.result
		ev.Record(job.id+"-result", o.result)
		ev.Record(job.id+"-events", job.envelopes)
		meta := o.result.Metadata
		b := job.binding
		if !ev.Check("concurrent call retains exact pinned owner", o.err == nil && meta.TenantID == b.TenantID && meta.ProviderID == b.ProviderID && meta.API == b.API && meta.AccountScopeID == "same-account-alias" && meta.RequestID == job.id && meta.BindingVersion == "binding-v1" && meta.CredentialVersion == "secret-v1" && len(meta.Attempts) == 1, "id=%s err=%v meta=%+v", job.id, o.err, meta) {
			continue
		}
		for _, envelope := range job.envelopes {
			ev.Check("every stream envelope retains own call attribution", envelope.Call == meta.CallAttribution, "id=%s envelope=%+v", job.id, envelope)
		}
		ev.Check("visible result belongs to this call", strings.Contains(string(mustMarshal(t, o.result.Message.Content)), "answer-"+job.id), "id=%s content=%+v", job.id, o.result.Message.Content)
		a := meta.Attempts[0]
		low, lok := a.CostEstimate.Minimum.Get()
		high, hok := a.CostEstimate.Maximum.Get()
		expectedStatus := ai.CostResolved
		if b.API == ai.APIOpenAICompletions {
			expectedStatus = ai.CostBounded
		}
		ev.Check("own counters and independently worked cost bounds", a.UsageReporting == ai.UsageComplete && a.Usage.Input == job.input && a.Usage.Output == 4 && a.CostEstimate.Status == expectedStatus && lok && hok && math.Abs(low.Total-job.minimum) < 1e-12 && math.Abs(high.Total-job.maximum) < 1e-12 && a.Usage == o.result.Message.Usage && a.Usage.Cost == low, "id=%s attempt=%+v", job.id, a)
		var capture *provider.Request
		for i := range captures {
			if captures[i].Header["X-Host-Trace"] == job.id {
				capture = &captures[i]
			}
		}
		if ev.Check("selected tenant market secret reaches only its authorized path", capture != nil && capture.KeyAlias == job.key.alias && strings.HasPrefix(capture.Path, strings.TrimPrefix(b.Endpoint, w.provider.URL())), "id=%s capture=%+v", job.id, capture) {
			fields := bodyFields(t, capture.Body)
			ev.Check("caller and callback mutations cannot change serialized request", strings.Contains(string(capture.Body), "original-input-"+job.id) && !strings.Contains(string(capture.Body), "caller-mutated") && !strings.Contains(string(capture.Body), "callback-mutated"), "id=%s body=%s", job.id, capture.Body)
			if b.API != ai.APIAnthropicMessages {
				var key string
				_ = json.Unmarshal(fields["prompt_cache_key"], &key)
				owner := b.TenantID + "/" + string(b.ProviderID)
				prior := ownerKeys[owner]
				ev.Check("cache key separates tenant/market and shares only same-account protocol scope", len(key) == 64 && !strings.Contains(key, b.TenantID) && !strings.Contains(string(capture.Body), "same-private-session") && ((prior == "" && !keys[key]) || prior == key), "id=%s key=%s prior=%s", job.id, key, prior)
				keys[key] = true
				ownerKeys[owner] = key
			} else {
				ev.Check("Messages does not acquire an undeclared session routing field", fields["prompt_cache_key"] == nil, "body=%s", capture.Body)
			}
		}
		scope := ai.CallScope{TenantID: b.TenantID, RequestID: job.id, ActorID: "actor-" + b.TenantID, JobID: "job-" + job.id}
		records := rec.awaitCall(ev, job.id)
		checkCallRecords(ev, records, scope, b.BindingID, o, 1)
		for _, record := range records {
			if record.Kind == ai.ObservationAttemptStarted {
				ev.Check("attempt cost starts unreported for this identity", record.Attempt != nil && record.Attempt.CostEstimate.Status == ai.CostUnreported, "record=%+v", record)
			}
		}
		original := o.result
		o.result.Metadata.Attempts[0].CostEstimate.Status = ai.CostPartial
		o.result.Metadata.Attempts[0].CostEstimate.Minimum = ai.Value(ai.UsageCost{Total: 999})
		o.result.Message.Content = nil
		for _, record := range records {
			if record.Attempt != nil {
				record.Attempt.CostEstimate.Status = ai.CostPartial
				record.Attempt.CostEstimate.Maximum = ai.Value(ai.UsageCost{Total: 777})
			}
			if len(record.Call.Attempts) > 0 {
				record.Call.Attempts[0].CostEstimate.Status = ai.CostUnreported
			}
		}
		again, err := job.stream.Result()
		againLow, _ := again.Metadata.Attempts[0].CostEstimate.Minimum.Get()
		againHigh, _ := again.Metadata.Attempts[0].CostEstimate.Maximum.Get()
		// Preserve original expected values independently of mutable result slices.
		ev.Check("Result and Observer ownership preserve immutable cost snapshots", err == nil && len(again.Message.Content) > 0 && again.Metadata.Attempts[0].CostEstimate.Status == expectedStatus && againLow == low && againHigh == high && reflect.DeepEqual(again.Message.Usage, original.Message.Usage), "id=%s result=%+v", job.id, again)
	}
	ev.Record("secret-reads", reads)
	ev.Check("four separate records each serve exactly three allowed protocols", len(reads) == 4 && len(keys) == 4, "reads=%v keys=%d", reads, len(keys))
	for _, n := range reads {
		ev.Check("secret selection has no fallback or cross-record reuse", n == 3, "reads=%d", n)
	}
	ev.Record("resources", map[string]int64{"calls": gauges.ActiveCalls(), "permits": gauges.Permits(), "queued": gauges.QueuedEvents(), "bodies": bodies.Open()})
	ev.Check("all twelve calls release their resources", len(captures) == 12 && gauges.ActiveCalls() == 0 && gauges.Permits() == 0 && gauges.QueuedEvents() == 0 && bodies.Open() == 0, "captures=%d", len(captures))
}

func kimiConcurrentReply(t *testing.T, job *kimiConcurrentJob) provider.Reply {
	t.Helper()
	if job.binding.API == ai.APIOpenAICompletions {
		frames := []json.RawMessage{
			json.RawMessage(fmt.Sprintf(`{"choices":[{"index":0,"delta":{"content":"answer-%s"},"finish_reason":"stop"}]}`, job.id)),
			json.RawMessage(fmt.Sprintf(`{"choices":[],"usage":{"prompt_tokens":%d,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"completion_tokens_details":{"reasoning_tokens":0}}}`, job.input+100)),
		}
		return chatEvents(t, frames, "")
	}
	reply, _ := kimiHostScript(t, job.binding.API)
	script := strings.ReplaceAll(reply.Script, `"input_tokens":8`, fmt.Sprintf(`"input_tokens":%d`, job.input))
	script = strings.ReplaceAll(script, "Ready.", "answer-"+job.id)
	if job.binding.API == ai.APIOpenAIResponses {
		script = strings.ReplaceAll(script, `"total_tokens":12`, fmt.Sprintf(`"total_tokens":%d`, job.input+4))
	}
	return provider.SSE([][]byte{[]byte(script)})
}
