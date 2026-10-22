package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiHostDisableAndRollback(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(service), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("KIMI-%s-K02-K15-host-disable-inflight-rollback", service))
			w := newKimiHostWorld(t)
			b := w.bindings[service][0]
			target := ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}
			request := ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}
			reply, raw := kimiHostScript(t, b.API)
			ev.Fixture("synthetic-script.json", raw)
			reply.OnReceive = func() {
				disabled := b
				disabled.Enabled = false
				disabled.Version = "binding-v2"
				w.host.PutBinding(disabled)
			}
			w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
			pinned, err := kimihost.Turn(ctxFor(t), w.client, textScope("inflight"), target, request, b.API, kimihost.FullComplete, ai.Hooks{})
			ev.Record("inflight", pinned.Result)
			ev.Check("already sent call retains pinned version while Host disables route", err == nil && pinned.Result.Metadata.BindingVersion == "binding-v1" && len(pinned.Result.Metadata.Attempts) == 1, "err=%v result=%+v", err, pinned.Result)
			refused, e := w.client.Complete(ctxFor(t), textScope("disabled"), target, request, nil)
			ev.Record("disabled", refused)
			f, ok := e.(*ai.Error)
			ev.Check("new call refuses before secret or attempt", ok && f.Code == ai.CodeTenantDenied && f.Phase == ai.PhaseBinding && w.reads[service] == 1 && len(refused.Metadata.Attempts) == 0, "err=%v reads=%v", e, w.reads)
			other := ai.ProviderMoonshot
			if service == other {
				other = ai.ProviderMoonshotCN
			}
			sibling := w.bindings[other][0]
			reply, _ = kimiHostScript(t, sibling.API)
			w.provider.EnqueueAt(strings.TrimPrefix(sibling.Endpoint, w.provider.URL()), reply)
			survived, e := kimihost.Turn(ctxFor(t), w.client, textScope("other-market"), ai.Target{BindingID: sibling.BindingID, ModelID: "kimi-k3"}, request, sibling.API, kimihost.SimpleComplete, ai.Hooks{})
			ev.Record("other-market", survived.Result)
			ev.Check("other market remains independently callable", e == nil && survived.Result.Metadata.ProviderID == other && w.reads[other] == 1, "err=%v reads=%v", e, w.reads)
			// Disable first, then load an old snapshot with no Kimi identities. A stale
			// enabled binding cannot acquire a model from another market or protocol.
			prior := ai.BuiltinCatalog()
			old, buildErr := ai.NewClient(ai.Config{Policy: validPolicy(), Catalog: &prior, Bindings: w.host, Credentials: w.host, Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true})
			if buildErr != nil {
				t.Fatal(buildErr)
			}
			w.host.PutBinding(b)
			rollback, e := old.Complete(ctxFor(t), textScope("rollback"), target, request, nil)
			ev.Record("rollback", rollback)
			ev.Record("requests", w.provider.Requests())
			f, ok = e.(*ai.Error)
			ev.Check("old catalog refuses removed identity without credentials or fallback", ok && f.Code == ai.CodeInvalidRequest && f.Phase == ai.PhaseCapability && len(rollback.Metadata.Attempts) == 0 && w.host.ReadsFor("rollback").Credentials == 0 && len(w.provider.Requests()) == 2, "err=%v result=%+v", e, rollback)
		})
	}
}
