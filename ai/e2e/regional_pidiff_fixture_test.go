package e2e

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

type regionalFrozenRoute struct {
	provider ai.ProviderID
	api      ai.API
	name     string
	models   []string
}

var regionalFrozenRoutes = []regionalFrozenRoute{{ai.ProviderZAI, ai.APIOpenAICompletions, "zai-chat", []string{"glm-5.3", "glm-5.2"}}, {ai.ProviderMiniMaxCN, ai.APIAnthropicMessages, "minimax-cn-messages", []string{"MiniMax-M3"}}}

func regionalFrozenWorld(t *testing.T, r regionalFrozenRoute) (*world, ai.Target) {
	t.Helper()
	cat, err := regional.CandidateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	cat.Models = cat.ModelsOf(r.provider, r.api)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	base := "/v1"
	if r.api == ai.APIAnthropicMessages {
		base = "/anthropic"
	}
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "regional-frozen", Version: "regional-frozen-v1", Enabled: true, ProviderID: r.provider, API: r.api, Endpoint: w.provider.URL() + base, AuthKind: ai.AuthAPIKey, AccountScopeID: "regional-frozen-account", CredentialRef: "regional-frozen-credential", AllowedModels: r.models}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA.secret)})
	return w, ai.Target{BindingID: b.BindingID, ModelID: r.models[0]}
}
func regionalFrozenScript(t *testing.T, r regionalFrozenRoute, model, stage string) (provider.Reply, []byte) {
	t.Helper()
	if r.api == ai.APIAnthropicMessages {
		file := "text.sse"
		if stage == "tools" {
			file = "tools.sse"
		}
		raw, err := os.ReadFile("testdata/regional-pidiff/minimax-cn-" + file)
		if err != nil {
			t.Fatal(err)
		}
		script := []byte(strings.ReplaceAll(string(raw), "MODEL", model))
		return provider.SSE([][]byte{script}), script
	}
	raw, err := os.ReadFile("testdata/regional-pidiff/zai-chat.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Text []struct {
			Model  string
			Events []json.RawMessage
		}
		Tools []glmToolScript
	}
	mustUnmarshal(t, raw, &f)
	if stage == "text" {
		for _, s := range f.Text {
			if s.Model == model {
				return chatEvents(t, s.Events, ""), raw
			}
		}
	} else {
		for _, s := range f.Tools {
			if s.Model == model {
				events := s.Tools
				if stage == "tool-replay" {
					events = s.Answer
				}
				return chatEvents(t, events, ""), raw
			}
		}
	}
	t.Fatal("missing own regional script")
	return provider.Reply{}, nil
}
func regionalFrozenCompare(t *testing.T, ev *evidence.Case, o *pioracle.Oracle, w *world, r regionalFrozenRoute, piRun pioracle.Run, pi, b pioracle.Observation, script []byte) {
	t.Helper()
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", b)
	checkObservationDuration(t, ev, "pi", pi)
	checkObservationDuration(t, ev, "barness", b)
	diffs, err := pioracle.Compare(pi, b)
	if !ev.Check("complete wire/events/Result compare", err == nil, "%v", err) {
		return
	}
	v := loadLedger(t).Classify(string(r.api), ev.ID(), diffs)
	ev.Record("pidiff", o.Record(ev.ID(), string(r.api), v, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version(), "openai-go": evidence.ModuleVersion("github.com/openai/openai-go/v3"), "anthropic-sdk-go": evidence.ModuleVersion("github.com/anthropics/anthropic-sdk-go")}, NodeVersion: piRun.Node, ModelCatalogHash: catalogHash(t, w.client.Catalog()), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), PiResult: pi.Result, BarnessResult: b.Result, Frames: script}))
	var pending []string
	for _, d := range v.Findings {
		if d.Classification == pioracle.Pending {
			pending = append(pending, string(d.Kind)+" "+d.Path)
		}
	}
	ev.Check("zero pending exact regional differences", v.Pass, "%s", strings.Join(pending, "; "))
}
