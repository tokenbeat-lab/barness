package e2e

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestManagedEffort is issue 27 (ADR-0019): the Anthropic models whose
// effort may change between turns (pi's supportsMidConvoEffort:
// claude-fable-5-1, claude-opus-5, claude-opus-5-5) are listed, record each
// turn's effort as providerThinkingLevel and replay it as pi does, so a host
// changing the effort mid-conversation never breaks the thinking binding.
// testdata/anthropic/managed-effort.json runs through the public Client
// here and through frozen pi in TestPiDifferential.
//
// Ways it could fail:
//
//	M1 a managed model is unlisted, or listed without SupportsMidConvoEffort
//	M2 thinking is not adaptive with drop_block, or is switched off by
//	   thinkingEnabled false or a simple call without a level, or the
//	   request-level effort is not high
//	M3 a temperature is sent (fable-5-1's data does not forbid one; only
//	   the managed branch keeps it out)
//	M4 a replayed assistant turn of the same provider and API lacks the
//	   system message with its effort, a turn of another provider or with
//	   no or a non-Anthropic effort gets one, or the call's effort does not
//	   trail the conversation
//	M5 the marker lands between a tool_use and its tool_result, or moves
//	   the cache breakpoint off the last user turn
//	M6 the result does not record the turn's effort (high by default), a
//	   failed call's message included
//	M7 the effort is treated as native state: dropped from a downgraded
//	   or cross-model turn of the same provider
//	M8 the two betas are missing from the anthropic-beta header
//	M9 the recorded effort is lost when the host serializes the result or
//	   passes it straight into its next call
func TestManagedEffort(t *testing.T) {
	t.Run("listed", func(t *testing.T) {
		ev := run.Case(t, "E03-managed-effort-listed")
		for _, id := range []string{"claude-fable-5-1", "claude-opus-5", "claude-opus-5-5"} {
			i := slices.IndexFunc(ai.BuiltinCatalog().Models, func(m ai.Model) bool {
				return m.API == ai.APIAnthropicMessages && m.ID == id
			})
			ev.Check(id+" listed with managed effort", i >= 0 && ai.BuiltinCatalog().Models[i].Compat.SupportsMidConvoEffort,
				"index %d", i)
		}
	})

	f, raw := loadFixture(t, anthropicProtocol, "managed-effort.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P02-E04-managed-effort-"+sc.ID)
			runScenario(t, ev, sc, raw, modeStream)
		})
	}

	// The host changes the effort between two turns, passing the first
	// turn's Result straight into the second call.
	t.Run("result-replays-its-effort", func(t *testing.T) {
		ev := run.Case(t, "P02-E04-managed-effort-result-replays-its-effort")
		ev.Fixture("fixture.json", raw)
		sc := f.Scenarios[slices.IndexFunc(f.Scenarios, func(sc fixtureScenario) bool { return sc.ID == "fable-5-1-first-turn" })]
		w := scenarioWorld(t, sc)
		enqueue(ev, w, sc.replies(t)...)
		enqueue(ev, w, sc.replies(t)...)
		req := sc.request(t)
		first, err := w.client.Complete(ctxFor(t), textScope("req-managed-first"), sc.target(), req,
			ai.AnthropicOptions{Effort: ai.AnthropicEffortMax})
		if !ev.Check("first turn succeeds", err == nil, "err=%v", err) {
			return
		}
		ev.Record("first", first.Message)
		ev.Check("first turn records max", first.Message.ProviderThinkingLevel == "max", "got %q", first.Message.ProviderThinkingLevel)
		var stored struct {
			ProviderThinkingLevel string `json:"providerThinkingLevel"`
		}
		ev.Check("serialized with the message", json.Unmarshal(mustMarshal(t, first.Message), &stored) == nil &&
			stored.ProviderThinkingLevel == "max", "got %+v", stored)

		req.Messages = append(req.Messages, first.Message, ai.UserText("And now?"))
		second, err := w.client.Complete(ctxFor(t), textScope("req-managed-second"), sc.target(), req,
			ai.AnthropicOptions{Effort: ai.AnthropicEffortLow})
		ev.Check("second turn succeeds and records low", err == nil && second.Message.ProviderThinkingLevel == "low",
			"err=%v level=%q", err, second.Message.ProviderThinkingLevel)
		reqs := w.provider.Requests()
		if !ev.Check("two requests sent", len(reqs) == 2, "got %d", len(reqs)) {
			return
		}
		ev.Record("requests", reqs)
		want := `[
			{"role":"user","content":[{"type":"text","text":"Say hello."}]},
			{"role":"system","content":[],"output_config":{"effort":"max"}},
			{"role":"assistant","content":[{"type":"thinking","thinking":"Hmm.","signature":"sig-r"},{"type":"text","text":"Done."}]},
			{"role":"user","content":[{"type":"text","text":"And now?","cache_control":{"type":"ephemeral"}}]},
			{"role":"system","content":[],"output_config":{"effort":"low"}}]`
		var body struct {
			Messages json.RawMessage `json:"messages"`
		}
		mustUnmarshal(t, reqs[1].Body, &body)
		ev.Check("first turn replayed signed after its effort, the new effort trailing",
			jsonEqual(body.Messages, json.RawMessage(want)), "got %s", body.Messages)
	})
}
