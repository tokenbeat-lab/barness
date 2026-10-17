package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxHostLocalAndCloud(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		t.Run(r.name, func(t *testing.T) {
			ev := run.Case(t, "MM08-local-cloud-"+r.name)
			w := newMiniMaxHostWorld(t)
			purpose := minimaxhost.Standard
			prefix := miniMaxHostPrefix(tenantA, purpose, r)
			w.provider.EnqueueAt(prefix, miniMaxHostReply(t, r, "MiniMax-M3"))
			local := miniMaxHostInvoke(ctxFor(t), w, tenantA, purpose, r, "complete", miniMaxMessagesRequest(), ai.Hooks{}, miniMaxHostScope(tenantA, "mm08-local"))
			ev.Record("local-result", local.result)
			ev.Check("local assembly uses the same Client contract", local.err == nil, "%v", local.err)
			// A trusted per-deployment output cap keeps the existing cloud example
			// bounded without changing its authenticated principal/session contract.
			cloudClient := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				field := map[ai.API]string{ai.APIOpenAICompletions: "max_completion_tokens", ai.APIOpenAIResponses: "max_output_tokens", ai.APIAnthropicMessages: "max_tokens"}[p.API]
				p.Body[field] = 512
				return ai.KeepPayload(), nil
			}})
			cloud := hostintegration.New(cloudClient, hostintegration.NewMemoryStore(), requestid.New)
			who := hostintegration.Principal{TenantID: tenantA.tenant, ActorID: "authenticated-user"}
			session, err := cloud.OpenSession(ctxFor(t), who)
			if err != nil {
				t.Fatal(err)
			}
			firstReply := miniMaxHostReply(t, r, "MiniMax-M3")
			if r.api == ai.APIAnthropicMessages {
				script := `event: message_start
data: {"type":"message_start","message":{"id":"cloud-native","model":"MiniMax-M3"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"Synthetic cloud reasoning.","signature":"cloud-signature"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}

event: message_stop
data: {"type":"message_stop"}

`
				firstReply = provider.SSE([][]byte{[]byte(script)})
			}
			w.provider.EnqueueAt(prefix, firstReply, miniMaxHostReply(t, r, "MiniMax-M3"))
			target := miniMaxHostTarget(purpose, r)
			input := hostintegration.TurnRequest{SessionID: session, BindingID: target.BindingID, ModelID: target.ModelID, Prompt: "Synthetic cloud input."}
			firstSink := &collectSink{}
			first, err := cloud.Turn(ctxFor(t), who, input, firstSink)
			ev.Record("cloud-first-result", first)
			ev.Record("cloud-first-envelopes", envelopesJSON(firstSink.all()))
			if !ev.Check("authenticated cloud turn succeeds", err == nil, "%v", err) {
				return
			}
			secondSink := &collectSink{}
			second, err := cloud.Turn(ctxFor(t), who, input, secondSink)
			ev.Record("cloud-second-result", second)
			ev.Record("cloud-second-envelopes", envelopesJSON(secondSink.all()))
			ev.Check("tenant-scoped persistence is vouched for after loading", err == nil && second.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "%v %+v", err, second.Metadata)
			requests := w.provider.Requests()
			ev.Record("requests", requests)
			if r.api == ai.APIAnthropicMessages && len(requests) == 3 {
				ev.Check("stored signed native state replays intact", strings.Contains(string(requests[2].Body), `"signature":"cloud-signature"`), "%s", requests[2].Body)
			}
			forged := input
			forged.TenantID = tenantB.tenant
			before := len(requests)
			denied, err := cloud.Turn(ctxFor(t), who, forged, &collectSink{})
			ev.Record("forged-result", denied)
			ev.Check("client tenant claims cannot become trusted scope", errors.Is(err, hostintegration.ErrForbidden) && len(w.provider.Requests()) == before, "%v", err)
			other := hostintegration.Principal{TenantID: tenantB.tenant}
			_, err = cloud.Turn(ctxFor(t), other, input, &collectSink{})
			ev.Check("another tenant cannot read the stored session", errors.Is(err, hostintegration.ErrForbidden) && len(w.provider.Requests()) == before, "%v", err)
			checkReleased(ev, w.limitWorld)
		})
	}
}
