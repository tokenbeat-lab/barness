package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var glmB = tenantKey{tenant: "tenant-b", secret: "synthetic-glm-tenant-b", alias: "key:tenant-b-glm@v1"}

func init() { knownKeys = append(knownKeys, glmB) }

type glmSharedRoute struct {
	name string
	api  ai.API
}

var glmSharedRoutes = []glmSharedRoute{{"chat", ai.APIOpenAICompletions}, {"responses", ai.APIOpenAIResponses}, {"messages", ai.APIAnthropicMessages}}

func (r glmSharedRoute) prefix(k tenantKey) string { return "/" + k.tenant + "/glm-" + r.name }
func (r glmSharedRoute) target(model string) ai.Target {
	return ai.Target{BindingID: "glm-" + r.name, ModelID: model}
}
func (r glmSharedRoute) options() ai.Options {
	switch r.name {
	case "chat":
		return ai.ChatOptions{SessionID: "same-session"}
	case "responses":
		return ai.ResponsesOptions{SessionID: "same-session"}
	default:
		return ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
	}
}
func (r glmSharedRoute) script(t *testing.T, model string) (ai.Request, provider.Reply, provider.Reply, []byte) {
	t.Helper()
	switch r.name {
	case "chat":
		sc, raw := glmToolFixture(t, model)
		return glmToolRequest(model), chatEvents(t, sc.Tools, ""), chatEvents(t, sc.Answer, ""), raw
	case "responses":
		scenarios, raw := glmResponsesToolFixtures(t)
		for _, sc := range scenarios {
			if sc.Model == model && sc.Name == "schema-native-array" {
				return glmResponsesToolRequest(), glmResponsesReply(t, sc.Tools), glmResponsesReply(t, sc.Answer), raw
			}
		}
	case "messages":
		first, last := glmMessagesCaptureByName(t, model+"-tool"), glmMessagesCaptureByName(t, model+"-tool-continuation")
		return glmMessagesCapturedRequest(t, first), first.reply(t), last.reply(t), mustMarshal(t, []any{first, last})
	}
	t.Fatalf("missing own fixture %s %s", r.name, model)
	return ai.Request{}, provider.Reply{}, provider.Reply{}, nil
}
func newGLMSharedWorld(t *testing.T, configure func(*ai.Config)) *mixedWorld {
	t.Helper()
	network := provider.LoopbackTransport()
	network.MaxIdleConnsPerHost = 2 // bounded idle pool is separate from active call cleanup
	t.Cleanup(network.CloseIdleConnections)
	w := &mixedWorld{host: host.New(), provider: localProvider(t), gauges: probe.New(), bodies: provider.TrackBodies(network), rec: newRecorder(), adm: host.NewAdmission()}
	cat := glm.ChatCatalog()
	cat.Version = "glm09-shared-synthetic"
	cat.Models = append(cat.Models, glm.ResponsesCatalog().Models...)
	cat.Models = append(cat.Models, glm.MessagesCatalog().Models...)
	cfg := ai.Config{Catalog: &cat, Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Transport: w.bodies, Probe: w.gauges, Observer: w.rec, Admission: w.adm, AllowLoopbackHTTP: true}
	cfg.Policy.MaxQueuedObservations = 512
	if configure != nil {
		configure(&cfg)
	}
	var err error
	w.client, err = ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []tenantKey{glmA, glmB} {
		for _, r := range glmSharedRoutes {
			var b ai.Binding
			switch r.name {
			case "chat":
				b = glm.ChatBinding(k.tenant, "acct-"+k.tenant, "same-credential", w.provider.URL()+r.prefix(k))
			case "responses":
				b = glm.ResponsesBinding(k.tenant, "acct-"+k.tenant, "same-credential", w.provider.URL()+r.prefix(k))
			case "messages":
				b = glm.MessagesBinding(k.tenant, "acct-"+k.tenant, "same-credential", w.provider.URL()+r.prefix(k))
			}
			w.host.PutBinding(b)
		}
		w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: "same-credential", AccountScopeID: "acct-" + k.tenant, Version: "v1", Active: true, APIKey: ai.NewSecret(k.secret)})
	}
	return w
}
func glmSharedRecord(ev *evidence.Case, w *mixedWorld, o outcome) {
	o.record(ev)
	w.record(ev, mixedOutcome{metadata: o.result.Metadata, usage: o.result.Message.Usage, chat: o.result, err: o.err})
	wire := observationText(ev.T(), w.rec.all())
	for _, value := range []string{"Check both.", "BRN42", `"arguments"`, "reasoning_content", "Synthetic input", glmA.secret, glmB.secret} {
		ev.Check("Observer excludes private content", !strings.Contains(wire, value), "unexpected content")
	}
}
func glmSharedContinue(req ai.Request, m ai.AssistantMessage) ai.Request {
	req.Messages = append(req.Messages, m)
	for _, block := range m.Content {
		if c, ok := block.(ai.ToolCall); ok {
			req.Messages = append(req.Messages, ai.ToolResultText(c, "BRN42", false))
		}
	}
	return req
}
func glmSharedWirePairs(t *testing.T, r glmSharedRoute, raw []byte) (native int, calls, results []string) {
	t.Helper()
	body := bodyFields(t, raw)
	var input []map[string]any
	if r.name == "responses" {
		mustUnmarshal(t, body["input"], &input)
	} else {
		mustUnmarshal(t, body["messages"], &input)
	}
	for _, item := range input {
		if r.name == "responses" {
			switch item["type"] {
			case "reasoning":
				native++
			case "function_call":
				calls = append(calls, item["call_id"].(string))
			case "function_call_output":
				results = append(results, item["call_id"].(string))
			}
			continue
		}
		if r.name == "chat" {
			if v, ok := item["reasoning_content"].(string); ok && v != "" {
				native++
			}
			if list, ok := item["tool_calls"].([]any); ok {
				for _, v := range list {
					calls = append(calls, v.(map[string]any)["id"].(string))
				}
			}
			if item["role"] == "tool" {
				results = append(results, item["tool_call_id"].(string))
			}
			continue
		}
		content, _ := item["content"].([]any)
		for _, v := range content {
			b := v.(map[string]any)
			switch b["type"] {
			case "thinking":
				native++
			case "tool_use":
				calls = append(calls, b["id"].(string))
			case "tool_result":
				results = append(results, b["tool_use_id"].(string))
			}
		}
	}
	return
}

// Explicit JSON copy for mutated restored Host records, never trust inferred
// from JSON. The stored envelope is separately attested via TrustNativeState.
func glmSharedRestore(t *testing.T, m ai.AssistantMessage) (ai.AssistantMessage, *ai.NativeStateEnvelope) {
	t.Helper()
	v, e, _ := restore(t, persist(t, m))
	return v, e
}
