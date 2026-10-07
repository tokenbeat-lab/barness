package e2e

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// TestOptions is the option part of E04 (spec I7 first five items, User
// Stories 21–23): full Responses options and simple options reach the wire
// as frozen pi-ai encodes them — every reasoning level and its clamp, the
// model's level map, the output token budget with its context estimate,
// samplingParams, prompt cache retention and field presence — through Stream
// and Complete (full) or StreamSimple and CompleteSimple (simple). Each
// scenario's expected body is pi's (see options.json and the
// PIDIFF-P01-E04-options-* differential), except the cache key, which is
// derived per tenant and account.
//
// Ways it could fail, listed before the implementation:
//
//	O1 an unset option is sent, or an explicit null or zero is dropped or
//	   confused with unset (temperature, service_tier, tool_choice,
//	   maxTokens on the simple entry)
//	O2 a null or zero is sent where pi sends nothing (full maxTokens 0/null,
//	   reasoningSummary null)
//	O3 a simple level is not clamped to the model's supported levels, clamped
//	   in the wrong direction, or a full effort is clamped at all
//	O4 the level map's null (disabled), remapped and unset branches are
//	   confused, or the off mapping is not sent when no effort is given
//	O5 the simple output budget misses the model default, replaces 0 with
//	   it, mis-estimates the context (UTF-16 length, images, tool JSON,
//	   usage of the last successful turn, failed turns, system sections and
//	   removed tools), loses the 4096 safety margin or the no-context-window
//	   exception, or the 16-token Responses minimum
//	O6 samplingParams lose nulls or are not applied in the same order on
//	   full and simple entries: named fields, model defaults, call overrides
//	O7 cache retention maps wrongly for a compat; the raw session id is sent;
//	   affinity headers appear without a key, or a key without them
//	O8 Responses sends metadata, or the tenant ends up in the request
//	O9 the caller's options value changes during the call
func TestOptions(t *testing.T) {
	f, raw := loadOptionsFixture(t)
	for _, sc := range f.Scenarios {
		for _, e := range optionsEntries {
			t.Run(sc.ID+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E04-options-"+sc.ID+"-"+e.name)
				ev.Fixture("options.json", raw)
				ev.Record("about", sc.About)
				w := newWorldWith(t, sc.configure, tenantA)
				enqueue(ev, w, f.reply(t))
				scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-options-" + sc.ID}
				o, before, after := within(ev, "the call ends", func() optionsOutcome {
					o, before, after := e.invoke(ctxFor(t), t, w, scope, sc, f.request(t, sc))
					return optionsOutcome{o, before, after}
				}).unpack()
				o.record(ev)

				ev.Check("the call succeeds", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop,
					"err=%v stop=%q", o.err, o.result.Message.StopReason)
				reqs := w.provider.Requests()
				if !ev.Check("one inference request", len(reqs) == 1, "got %d", len(reqs)) {
					return
				}
				req := reqs[0]
				key := checkCacheAffinity(t, ev, req.Body, req.Header, sc.Expect.Affinity)
				ev.Check("wire body is pi's", jsonEqual(req.Body, substituteCacheKey(sc.Expect.Body, key)), "got %s", req.Body)
				ev.Check("the tenant is not sent", !strings.Contains(string(req.Body), tenantA.tenant) && !headersMention(req.Header, tenantA.tenant),
					"body %s headers %v", req.Body, req.Header)
				ev.Check("the caller's options are unchanged", before == after, "before %s\nafter  %s", before, after)
			})
		}
	}
}

type optionsOutcome struct {
	o             outcome
	before, after string
}

func (x optionsOutcome) unpack() (outcome, string, string) { return x.o, x.before, x.after }

var hexKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// checkCacheAffinity checks the prompt cache key and affinity headers of a
// request and returns the key ("" when none was sent). A key is never the
// caller's raw session id; the same key is sent as session_id and
// x-client-request-id, as pi sends its session id.
func checkCacheAffinity(t *testing.T, ev *evidence.Case, body []byte, header map[string]string, want bool) string {
	t.Helper()
	key := sentCacheKey(t, body)
	sid, hasSid := header["Session_id"]
	cid, hasCid := header["X-Client-Request-Id"]
	if !want {
		ev.Check("no affinity headers", !hasSid && !hasCid, "Session_id=%q X-Client-Request-Id=%q", sid, cid)
		return key
	}
	ev.Check("a derived cache key is sent", hexKey.MatchString(key), "prompt_cache_key=%q", key)
	ev.Check("affinity headers carry the cache key", sid == key && cid == key,
		"key=%q Session_id=%q X-Client-Request-Id=%q", key, sid, cid)
	return key
}

// headersMention reports s in a header other than Authorization, which the
// local Provider reports as the key's test alias.
func headersMention(header map[string]string, s string) bool {
	for k, v := range header {
		if k != "Authorization" && strings.Contains(v, s) {
			return true
		}
	}
	return false
}

// TestCacheKeyScope: the prompt cache key and session affinity headers
// barness-ai builds from a session id are scoped to the tenant and the
// vendor account (spec §9), so two tenants sharing a session id — or one
// tenant's session moved to another account — never share a cache or
// affinity identity, while the same tenant, account and session keep
// theirs.
//
// Ways it could fail, listed before the implementation:
//
//	K1 the raw session id or the TenantID is sent
//	K2 two tenants with the same session id on accounts that may be shared
//	   get the same key
//	K3 the same tenant's session gets the same key on another account
//	K4 the key is not stable for the same tenant, account and session, so
//	   provider caching never hits
func TestCacheKeyScope(t *testing.T) {
	f, _ := loadOptionsFixture(t)
	var sc optionsScenario
	for _, s := range f.Scenarios {
		if s.ID == "full-cache-session" {
			sc = s
		}
	}
	ev := run.Case(t, "P01-E04-cache-key-scope")
	w := newWorld(t, tenantA, tenantB)
	call := func(k tenantKey, id string) string {
		enqueue(ev, w, f.reply(t))
		scope := ai.CallScope{TenantID: k.tenant, RequestID: "req-cache-key-" + id}
		target := ai.Target{BindingID: "primary", ModelID: sc.Model}
		res, err := w.client.Complete(ctxFor(t), scope, target, f.request(t, sc), sc.fullOptions(t))
		ev.Check("call "+id+" succeeds", err == nil, "err=%v stop=%q", err, res.Message.StopReason)
		reqs := w.provider.Requests()
		last := reqs[len(reqs)-1]
		return checkCacheAffinity(t, ev, last.Body, last.Header, true)
	}
	a1 := call(tenantA, "a1")
	a2 := call(tenantA, "a2")
	b := call(tenantB, "b")
	// Tenant A's binding and key move to another vendor account.
	w.updateBinding(tenantA, func(b *ai.Binding) { b.AccountScopeID = "acct-elsewhere" })
	w.replaceCredential(tenantA, func(c *ai.Credential) { c.AccountScopeID = "acct-elsewhere" })
	moved := call(tenantA, "a-moved")
	ev.Record("keys", map[string]string{"a1": a1, "a2": a2, "b": b, "a-moved": moved})

	ev.Check("the raw session id is not sent", a1 != "session-1", "key=%q", a1)
	ev.Check("stable for the same tenant, account and session", a1 == a2, "%q vs %q", a1, a2)
	ev.Check("another tenant's same session gets another key", a1 != b, "both %q", a1)
	ev.Check("another account gets another key", a1 != moved, "both %q", a1)
}

// TestOptionRejects: options barness-ai cannot send faithfully, or that
// would cross the call's authorization boundary, are invalid_request in the
// capability phase before any credential is read or request sent, on both
// invocation styles of their entry (spec I3, I7).
//
// Ways it could fail, listed before the implementation:
//
//	X1 an unknown or "off" reasoning level, summary or cache retention is
//	   passed through to the provider
//	X2 a non-finite temperature or a negative token count reaches the wire
//	X3 an invalid tool choice (no or several forms, unknown mode, a function
//	   or allowed tool without a name) is sent
//	X4 a null or negative thinking budget is silently used
//	X5 samplingParams replace the model, the history (and with it native
//	   state), the declared tools, server-side conversation state, a stored
//	   prompt, the derived cache key or the streaming contract, or make the
//	   vendor store the response — on either entry, from the call or the
//	   model's own samplingParams
//	X6 samplingParams or metadata holding something that is not JSON
func TestOptionRejects(t *testing.T) {
	full := map[string]ai.ResponsesOptions{
		"effort-off":            {ReasoningEffort: "off"},
		"effort-unknown":        {ReasoningEffort: "extreme"},
		"summary-unknown":       {ReasoningSummary: "verbose"},
		"cache-unknown":         {CacheRetention: "forever"},
		"temperature-nan":       {Temperature: ai.Value(math.NaN())},
		"temperature-inf":       {Temperature: ai.Value(math.Inf(1))},
		"max-tokens-negative":   {MaxTokens: ai.Value(-1)},
		"tool-choice-empty":     {ToolChoice: ai.Value(ai.ResponsesToolChoice{})},
		"tool-choice-mode":      {ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "sometimes"})},
		"tool-choice-two-forms": {ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "auto", Function: "get_weather"})},
		"tool-choice-allowed-mode": {ToolChoice: ai.Value(ai.ResponsesToolChoice{
			AllowedTools: &ai.ResponsesAllowedTools{Mode: "none", Functions: []string{"get_weather"}}})},
		"tool-choice-allowed-empty": {ToolChoice: ai.Value(ai.ResponsesToolChoice{
			AllowedTools: &ai.ResponsesAllowedTools{Mode: "auto", Functions: []string{""}}})},
		"sampling-not-json": {SamplingParams: map[string]json.RawMessage{"top_p": json.RawMessage(`{`)}},
		"sampling-nil":      {SamplingParams: map[string]json.RawMessage{"top_p": nil}},
	}
	simple := map[string]ai.SimpleOptions{
		"reasoning-off":        {Reasoning: "off"},
		"reasoning-unknown":    {Reasoning: "extreme"},
		"cache-unknown":        {CacheRetention: "forever"},
		"temperature-nan":      {Temperature: ai.Value(math.NaN())},
		"max-tokens-negative":  {MaxTokens: ai.Value(-5)},
		"tool-choice-required": {ToolChoice: ai.Value(ai.ToolChoice("required"))},
		"budget-null":          {ThinkingBudgets: ai.ThinkingBudgets{Low: ai.Null[int]()}},
		"budget-negative":      {ThinkingBudgets: ai.ThinkingBudgets{High: ai.Value(-1)}},
		"metadata-not-json":    {Metadata: map[string]json.RawMessage{"user_id": json.RawMessage(`nope`)}},
		"sampling-not-json":    {SamplingParams: map[string]json.RawMessage{"top_p": json.RawMessage(`{`)}},
	}
	for _, key := range []string{"model", "input", "stream", "tools", "previous_response_id", "conversation", "prompt", "prompt_cache_key", "store", "background"} {
		params := map[string]json.RawMessage{key: json.RawMessage(`"gpt-4"`)}
		full["sampling-"+key] = ai.ResponsesOptions{SamplingParams: params}
		simple["sampling-"+key] = ai.SimpleOptions{SamplingParams: params}
	}

	reject := func(t *testing.T, id string, configure func(*ai.Config), invoke func(w *world, e optionsEntry, scope ai.CallScope, target ai.Target) outcome) {
		for _, e := range optionsEntries {
			t.Run(id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E04-options-reject-"+id+"-"+e.name)
				w := newWorldWith(t, configure, tenantA)
				o := invoke(w, e, textScope("req-options-reject"), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"})
				o.record(ev)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				reads := w.host.ReadsFor("req-options-reject")
				ev.Check("no credential read", reads.Credentials == 0, "got %d", reads.Credentials)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
			})
		}
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText("hi")}}
	for id, opts := range full {
		reject(t, "full-"+id, func(*ai.Config) {}, func(w *world, e optionsEntry, scope ai.CallScope, target ai.Target) outcome {
			if e.stream {
				return drain(w.client.Stream(ctxFor(t), scope, target, req, opts))
			}
			res, err := w.client.Complete(ctxFor(t), scope, target, req, opts)
			return outcome{result: res, err: err}
		})
	}
	invokeSimple := func(opts ai.SimpleOptions) func(w *world, e optionsEntry, scope ai.CallScope, target ai.Target) outcome {
		return func(w *world, e optionsEntry, scope ai.CallScope, target ai.Target) outcome {
			if e.stream {
				return drain(w.client.StreamSimple(ctxFor(t), scope, target, req, opts))
			}
			res, err := w.client.CompleteSimple(ctxFor(t), scope, target, req, opts)
			return outcome{result: res, err: err}
		}
	}
	for id, opts := range simple {
		reject(t, "simple-"+id, func(*ai.Config) {}, invokeSimple(opts))
	}
	// Model defaults are refused at Client construction (TestModelSamplingRejects).
}
