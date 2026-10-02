package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// isolationFixture is testdata/isolation/interleave.json: per protocol, the
// one logical input both tenants send and each tenant's own reply scripts.
// Markers are strings only that tenant's replies, ids or attribution
// contain; none may reach the other tenant's call.
type isolationFixture struct {
	Case      string              `json:"case"`
	Source    string              `json:"source"`
	Markers   map[string][]string `json:"markers"`
	Protocols []isolationProtocol `json:"protocols"`
}

type isolationProtocol struct {
	ID           string                     `json:"id"`
	BindingID    string                     `json:"bindingId"`
	Model        string                     `json:"model"`
	SystemPrompt string                     `json:"systemPrompt"`
	User         string                     `json:"user"`
	Tenants      map[string]isolationScript `json:"tenants"`
}

// isolationScript is one tenant's replies: its successful SSE events and,
// for tenant A, its failing statuses.
type isolationScript struct {
	Events       []json.RawMessage `json:"events"`
	Unauthorized *isolationReply   `json:"unauthorized"`
	Throttled    *isolationReply   `json:"throttled"`
}

type isolationReply struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
}

func (r isolationReply) reply() provider.Reply {
	return provider.Reply{Status: r.Status, Header: r.Header, Chunks: [][]byte{[]byte(r.Body)}, Script: r.Body}
}

func loadIsolationFixture(t *testing.T) (isolationFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "isolation", "interleave.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f isolationFixture
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("interleave.json: %v", err)
	}
	return f, raw
}

func (p isolationProtocol) anthropic() bool { return p.BindingID == "claude" }

func (p isolationProtocol) gemini() bool { return p.BindingID == "gemini" }

func (p isolationProtocol) chat() bool { return p.BindingID == "chat" }

// deepseek is DeepSeek on the Responses protocol: the Responses adapter
// with DeepSeek's binding, key, account, endpoint and capabilities.
func (p isolationProtocol) deepseek() bool { return p.BindingID == "deepseek" }

// protocolTimeout reports whether p's protocol has a request timeout of its
// own (timeoutMs); pi's Google path has none.
func (p isolationProtocol) protocolTimeout() bool { return !p.gemini() }

func (p isolationProtocol) target() ai.Target {
	return ai.Target{BindingID: p.BindingID, ModelID: p.Model}
}

func (p isolationProtocol) request() ai.Request {
	return ai.Request{SystemPrompt: p.SystemPrompt, Messages: []ai.Message{ai.UserText(p.User)}}
}

// key is the key k's tenant holds for p's binding.
func (p isolationProtocol) key(k tenantKey) tenantKey {
	switch {
	case p.anthropic():
		return anthropicKey(k)
	case p.gemini():
		return googleKey(k)
	case p.deepseek():
		return deepseekKey(k)
	}
	return k
}

// account is the vendor account of k's tenant's binding for p.
func (p isolationProtocol) account(k tenantKey) string {
	switch {
	case p.anthropic():
		return "acct-" + k.tenant + "-anthropic"
	case p.gemini():
		return "acct-" + k.tenant + "-google"
	case p.deepseek():
		return "acct-" + k.tenant + "-deepseek"
	}
	return "acct-" + k.tenant
}

// path is where k's tenant's endpoint receives p's inference requests.
func (p isolationProtocol) path(k tenantKey) string {
	switch {
	case p.anthropic():
		return tenantPrefix(k) + "/v1/messages"
	case p.gemini():
		return tenantPrefix(k) + "/v1beta/models/" + p.Model + ":streamGenerateContent"
	case p.chat():
		return tenantPrefix(k) + "/v1/chat/completions"
	case p.deepseek():
		return tenantPrefix(k) + "/responses"
	}
	return tenantPrefix(k) + "/v1/responses"
}

func (p isolationProtocol) script(k tenantKey) isolationScript { return p.Tenants[k.tenant] }

// success is k's tenant's successful reply, one SSE frame per chunk.
func (p isolationProtocol) success(t *testing.T, k tenantKey) provider.Reply {
	t.Helper()
	switch {
	case p.gemini():
		return dataEvents(t, p.script(k).Events, provider.FramingLF)
	case p.chat():
		return chatEvents(t, p.script(k).Events, provider.FramingLF)
	}
	return sseEvents(t, p.script(k).Events, provider.FramingLF)
}

// firstMarkedChunk is the index of the first chunk of k's success reply
// carrying k's text marker: once it is written, k's call holds content of
// its own in flight.
func (p isolationProtocol) firstMarkedChunk(t *testing.T, f isolationFixture, k tenantKey) int {
	t.Helper()
	marker := f.Markers[k.tenant][0]
	for i, c := range p.success(t, k).Chunks {
		if strings.Contains(string(c), marker) {
			return i
		}
	}
	t.Fatalf("%s: no chunk of %s carries %q", p.ID, k.tenant, marker)
	return 0
}

// tenantPrefix is the path under which k's tenant's endpoints live on the
// shared Provider, so each request's path shows which tenant's endpoint it
// was sent to.
func tenantPrefix(k tenantKey) string { return "/" + k.tenant }

// sharedSession is the session id both tenants pass.
const sharedSession = "session-shared"

// isolationWorld is tenants A and B on one Client and one transport, with
// same-named bindings that point at each tenant's own endpoint on the shared
// Provider. It counts response bodies, permits and calls, records every
// observation and admission request, and backs off on a fake clock.
type isolationWorld struct {
	admissionWorld
	rec      *recorder
	arrivals *arrivals
}

// arrivals wraps the Client's transport and runs a hook as each response
// comes back from it, before the SDK sees it: the one point at which a call
// can deterministically end "as its response arrives".
type arrivals struct {
	next http.RoundTripper
	mu   sync.Mutex
	hook func(*http.Request)
}

func (a *arrivals) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := a.next.RoundTrip(req)
	a.mu.Lock()
	hook := a.hook
	a.mu.Unlock()
	if err == nil && hook != nil {
		hook(req)
	}
	return res, err
}

// onArrival sets the hook; nil removes it.
func (a *arrivals) onArrival(hook func(*http.Request)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.hook = hook
}

// configure adjusts the Client config last, e.g. to replace the policy.
func newIsolationWorld(t *testing.T, configure ...func(*ai.Config)) isolationWorld {
	t.Helper()
	rec := newRecorder()
	arr := &arrivals{}
	aw := newAdmissionWorld(t, nil, host.NewAdmission(), append([]func(*ai.Config){func(c *ai.Config) {
		arr.next = c.Transport
		c.Transport = arr
		c.Observer = rec
		c.Policy.MaxQueuedObservations = 256
		c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
	}}, configure...)...)
	for _, k := range []tenantKey{tenantA, tenantB} {
		for id, suffix := range map[string]string{"primary": "/v1", "claude": "", "gemini": "/v1beta", "chat": "/v1", "deepseek": ""} {
			b := aw.host.Binding(k.tenant, id)
			b.Endpoint = aw.provider.URL() + tenantPrefix(k) + suffix
			aw.host.PutBinding(b)
		}
	}
	return isolationWorld{admissionWorld: aw, rec: rec, arrivals: arr}
}

// enqueueFor scripts replies at k's tenant's endpoint and records them.
func (iw isolationWorld) enqueueFor(ev *evidence.Case, k tenantKey, replies ...provider.Reply) {
	iw.provider.EnqueueAt(tenantPrefix(k)+"/", replies...)
	ev.Record("response-script-"+k.tenant, replies)
}

// requestsAt are the captures received at k's tenant's endpoints.
func (iw isolationWorld) requestsAt(k tenantKey) []provider.Request {
	var out []provider.Request
	for _, r := range iw.provider.Requests() {
		if strings.HasPrefix(r.Path, tenantPrefix(k)+"/") {
			out = append(out, r)
		}
	}
	return out
}

// isolationEntry is one public entry point. Both tenants pass the same
// session id; timeoutMs is the call's own protocol timeout.
type isolationEntry struct {
	name   string
	invoke func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope, timeoutMs int) outcome
}

func fullOptions(p isolationProtocol, timeoutMs int) ai.Options {
	switch {
	case p.anthropic():
		return ai.AnthropicOptions{TimeoutMs: timeoutMs}
	case p.gemini():
		return ai.GeminiOptions{}
	case p.chat():
		// Chat sends a cache key off OpenAI's own endpoint only with long
		// retention.
		return ai.ChatOptions{SessionID: sharedSession, CacheRetention: ai.CacheRetentionLong, TimeoutMs: timeoutMs}
	}
	return ai.ResponsesOptions{SessionID: sharedSession, TimeoutMs: timeoutMs}
}

func simpleOptions(p isolationProtocol, timeoutMs int) ai.SimpleOptions {
	o := ai.SimpleOptions{SessionID: sharedSession, TimeoutMs: timeoutMs}
	if p.chat() {
		o.CacheRetention = ai.CacheRetentionLong
	}
	return o
}

var isolationEntries = []isolationEntry{
	{"stream-full", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope, ms int) outcome {
		return drain(w.client.Stream(ctx, scope, p.target(), p.request(), fullOptions(p, ms)))
	}},
	{"stream-simple", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope, ms int) outcome {
		return drain(w.client.StreamSimple(ctx, scope, p.target(), p.request(), simpleOptions(p, ms)))
	}},
	{"complete-full", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope, ms int) outcome {
		res, err := w.client.Complete(ctx, scope, p.target(), p.request(), fullOptions(p, ms))
		return outcome{result: res, err: err}
	}},
	{"complete-simple", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope, ms int) outcome {
		res, err := w.client.CompleteSimple(ctx, scope, p.target(), p.request(), simpleOptions(p, ms))
		return outcome{result: res, err: err}
	}},
}

// cue is a one-shot signal in a scenario's choreography. Waits on it are
// bounded, so a choreography that cannot proceed fails the case rather
// than hanging the Provider.
type cue struct {
	ch   chan struct{}
	once sync.Once
}

func newCue() *cue { return &cue{ch: make(chan struct{})} }

func (c *cue) fire() { c.once.Do(func() { close(c.ch) }) }

// await blocks until c fires. It may run on a Provider goroutine, so it
// reports a timeout without stopping the test.
func (c *cue) await(ev *evidence.Case, what string) {
	select {
	case <-c.ch:
	case <-time.After(waitDeadline):
		ev.Check(what, false, "not within %s", waitDeadline)
	}
}

// wireTrace is the order in which the Provider wrote the tenants' frames
// and the calls ended, for the evidence bundle.
type wireTrace struct {
	mu    sync.Mutex
	steps []string
}

func (w *wireTrace) add(format string, args ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.steps = append(w.steps, fmt.Sprintf(format, args...))
}

func (w *wireTrace) all() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string(nil), w.steps...)
}

// lockstep makes the Provider write a's and b's chunks alternately, a's
// first: A:0 B:0 A:1 B:1 … Each reply writes its next chunk only once the
// other wrote its current one (or has none left), so both calls are in
// flight together and every frame of one reaches the shared transport
// between frames of the other.
func lockstep(ev *evidence.Case, trace *wireTrace, a, b *provider.Reply) {
	aWrote, bWrote := cues(len(a.Chunks)), cues(len(b.Chunks))
	a.AfterChunk = func(i int) {
		trace.add("A:%d", i)
		aWrote[i].fire()
		if i+1 < len(aWrote) {
			bWrote[min(i, len(bWrote)-1)].await(ev, "tenant B's frame is written before A's next")
		}
	}
	b.OnReceive = func() { aWrote[0].await(ev, "tenant A's first frame is written before B's") }
	b.AfterChunk = func(i int) {
		trace.add("B:%d", i)
		bWrote[i].fire()
		if i+1 < len(bWrote) {
			aWrote[min(i+1, len(aWrote)-1)].await(ev, "tenant A's frame is written before B's next")
		}
	}
}

// alternating is the trace lockstep produces for replies of na and nb chunks.
func alternating(na, nb int) []string {
	var out []string
	for i := range max(na, nb) {
		if i < na {
			out = append(out, fmt.Sprintf("A:%d", i))
		}
		if i < nb {
			out = append(out, fmt.Sprintf("B:%d", i))
		}
	}
	return out
}

func cues(n int) []*cue {
	out := make([]*cue, n)
	for i := range out {
		out[i] = newCue()
	}
	return out
}

// holdAfter makes r stop after writing chunk n, firing holding, until
// resume fires.
func holdAfter(ev *evidence.Case, trace *wireTrace, r *provider.Reply, n int, holding, resume *cue) {
	r.AfterChunk = func(i int) {
		trace.add("B:%d", i)
		if i == n {
			holding.fire()
			resume.await(ev, "tenant A's call ends while B is held mid-stream")
		}
	}
}

// recordTenant stores one tenant's outcome in the evidence bundle.
func recordTenant(ev *evidence.Case, k tenantKey, o outcome) {
	ev.Record("outcome-"+k.tenant, map[string]any{
		"result": o.result, "error": errString(o.err), "events": eventsJSON(o.events), "streamed": o.streamed,
	})
}

// checkNoForeignContent asserts nothing only the other tenant's replies or
// attribution contain appears in k's outcome: not in its events, message,
// metadata, usage or error text.
func checkNoForeignContent(ev *evidence.Case, k tenantKey, o outcome, markers []string) {
	surface, err := json.Marshal(map[string]any{"result": o.result, "events": eventsJSON(o.events)})
	if !ev.Check(k.tenant+"'s outcome encodes", err == nil, "%v", err) {
		return
	}
	text := string(surface) + errString(o.err) + errString(o.streamErr)
	for _, m := range markers {
		ev.Check(k.tenant+"'s outcome holds nothing of the other tenant", !strings.Contains(text, m), "found %q", m)
	}
}

// checkMatchesSolo asserts k's outcome under interleaving is the one k gets
// running alone on its own Client: same events, message (but for its
// timestamp), usage and cost, attribution, attempts and error.
func checkMatchesSolo(ev *evidence.Case, k tenantKey, got, solo outcome) {
	ev.Check(k.tenant+"'s events equal its run alone", reflect.DeepEqual(eventSummary(got.events), eventSummary(solo.events)),
		"got %q\nalone %q", eventSummary(got.events), eventSummary(solo.events))
	gm, sm := got.result.Message, solo.result.Message
	gm.Timestamp, sm.Timestamp = 0, 0
	ev.Check(k.tenant+"'s message, usage and cost equal its run alone", reflect.DeepEqual(gm, sm), "got %+v\nalone %+v", gm, sm)
	ev.Check(k.tenant+"'s metadata and attempts equal its run alone", reflect.DeepEqual(got.result.Metadata, solo.result.Metadata),
		"got %+v\nalone %+v", got.result.Metadata, solo.result.Metadata)
	ev.Check(k.tenant+"'s error equals its run alone", errString(got.err) == errString(solo.err) && errString(got.streamErr) == errString(solo.streamErr),
		"got %v / %v\nalone %v / %v", got.err, got.streamErr, solo.err, solo.streamErr)
}

// checkSentAsSolo asserts what k's endpoint received is what it receives
// when k runs alone: path, key and body, headers included.
func checkSentAsSolo(ev *evidence.Case, k tenantKey, got, solo []provider.Request) {
	if !ev.Check(k.tenant+" sent as many requests as alone", len(got) == len(solo), "got %d alone %d", len(got), len(solo)) {
		return
	}
	for i := range got {
		ev.Check(fmt.Sprintf("%s's request %d equals its run alone", k.tenant, i), reflect.DeepEqual(got[i], solo[i]),
			"got %+v\nalone %+v", got[i], solo[i])
	}
}

// checkAttribution asserts k's call reports k's scope and, once resolved,
// k's account for p.
func checkAttribution(ev *evidence.Case, p isolationProtocol, k tenantKey, scope ai.CallScope, o outcome) {
	m := o.result.Metadata
	ev.Check(k.tenant+"'s result names its own scope", m.TenantID == k.tenant && m.RequestID == scope.RequestID && m.BindingID == p.BindingID,
		"metadata %+v", m)
	if m.Resolved {
		ev.Check(k.tenant+"'s result names its own account", m.AccountScopeID == p.account(k) && m.ModelID == p.Model, "metadata %+v", m)
	}
}

// checkObserved asserts k's call's observations: the lifecycle with
// attempts attempts, each record attributed to k's tenant and request and,
// once resolved, k's account, and call_finished mirroring k's Result.
func checkObserved(ev *evidence.Case, iw isolationWorld, p isolationProtocol, k tenantKey, scope ai.CallScope, o outcome, attempts int) {
	mine := func() []ai.Observation {
		var out []ai.Observation
		for _, x := range iw.rec.forRequest(scope.RequestID) {
			if x.Call.TenantID == k.tenant {
				out = append(out, x)
			}
		}
		return out
	}
	eventually(ev, "observer received "+k.tenant+"'s call_finished", func() bool {
		obs := mine()
		return len(obs) > 0 && obs[len(obs)-1].Kind == ai.ObservationCallFinished
	})
	obs := mine()
	ev.Record("observations-"+k.tenant, obs)
	checkCallRecords(ev, obs, scope, p.BindingID, o, attempts)
	for _, x := range obs {
		ev.Check(k.tenant+"'s records name its own account", x.Call.AccountScopeID == "" || x.Call.AccountScopeID == p.account(k),
			"%s record: %+v", x.Kind, x.Call)
	}
}

// checkAdmitted asserts the host's admission saw each of k's attempts, in
// order, with k's tenant and k's account for p.
func checkAdmitted(ev *evidence.Case, iw isolationWorld, p isolationProtocol, k tenantKey, scope ai.CallScope, attempts int) {
	var got, want []ai.AdmissionRequest
	for _, r := range iw.admission.Requests() {
		if r.TenantID == k.tenant {
			got = append(got, r)
		}
	}
	for i := 1; i <= attempts; i++ {
		want = append(want, ai.AdmissionRequest{TenantID: k.tenant, AccountScopeID: p.account(k), AttemptID: fmt.Sprintf("%s#%d", scope.RequestID, i)})
	}
	ev.Check(k.tenant+"'s attempts are admitted under its tenant and account", reflect.DeepEqual(got, want), "got %+v want %+v", got, want)
}
