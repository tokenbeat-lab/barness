package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestHistoryNormalization is the history part of E03 (spec I6 table, User
// Stories 11–14 and 19, P01 "支持模型的图片"): a host passes any authorized
// history — other models' and providers' turns, tool calls and results,
// system and tool declaration changes, images — and barness-ai converts it to
// the Responses wire as frozen pi-ai does, through every public entry point.
// Each scenario's expected body is pi's (see history.json and the
// PIDIFF-P01-E03-history-* differential).
//
// Ways it could fail, listed before the implementation:
//
//	H1 null content crashes, or a null user/assistant turn produces an item
//	H2 same-model thinking loses redacted or signed-empty blocks, or keeps
//	   unsigned blank ones; text loses its item id or phase
//	H3 cross-model thinking leaks a signature or redacted payload, or drops
//	   visible reasoning instead of turning it into text
//	H4 a normalized tool call id leaves its result pointing at the old id,
//	   or another provider's item id is sent as is
//	H5 a missing result is not synthesized, synthesized in the wrong place,
//	   or a held system message separates a call from its results
//	H6 an error/aborted turn is replayed, answered or counted
//	H7 a later system message or tool change is lost, applied out of order,
//	   or sections are ordered unlike pi
//	H8 images reach a text-only model, placeholders are not merged, or a
//	   vision model gets placeholders instead of images
//	H9 the caller's history changes during the call
func TestHistoryNormalization(t *testing.T) {
	f, raw := loadHistoryFixture(t)
	for _, sc := range f.Scenarios {
		for _, e := range outcomeEntries {
			t.Run(sc.ID+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E03-history-"+sc.ID+"-"+e.name)
				ev.Fixture("history.json", raw)
				ev.Record("about", sc.About)
				w := newWorldWith(t, sc.configure, tenantA)
				req := sc.request(t, tenantA)
				before := mustMarshal(t, req)

				enqueue(ev, w, f.reply(t))
				scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-history-" + sc.ID}
				target := ai.Target{BindingID: "primary", ModelID: sc.Model}
				o := within(ev, "the call ends", func() outcome { return e.invoke(ctxFor(t), w, scope, target, req) })
				o.record(ev)

				ev.Check("the call succeeds", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop,
					"err=%v stop=%q", o.err, o.result.Message.StopReason)
				reqs := w.provider.Requests()
				if !ev.Check("one inference request", len(reqs) == 1, "got %d", len(reqs)) {
					return
				}
				ev.Check("wire body is pi's", jsonEqual(reqs[0].Body, sc.Expect.Body), "got %s", reqs[0].Body)
				ev.Check("downgrade counts", o.result.Metadata.NativeStateDowngrades == sc.Expect.Downgrades,
					"got %+v want %+v", o.result.Metadata.NativeStateDowngrades, sc.Expect.Downgrades)
				after := mustMarshal(t, req)
				ev.Check("the caller's history is byte-for-byte unchanged", string(after) == string(before), "before %s\nafter  %s", before, after)
			})
		}
	}
}

// TestHistoryRejects: history the wire cannot express faithfully is
// invalid_request before anything is sent, on every entry point.
//
// Ways it could fail, listed before the implementation:
//
//	R1 a nil message or block panics or is silently skipped
//	R2 an image media type that is not image/<subtype> is spliced into the
//	   data URL
//	R3 a system message names a section twice, or removes a section while
//	   giving it text, and one of the two silently wins
//	R4 a system message's tool declarations escape the checks of
//	   Request.Tools
func TestHistoryRejects(t *testing.T) {
	tool := ai.Tool{Name: "a", Parameters: json.RawMessage(`{"type":"object"}`)}
	cases := map[string]ai.Message{
		"nil-message":           nil,
		"nil-user-block":        ai.UserMessage{Content: []ai.UserContent{nil}},
		"nil-tool-result-block": ai.ToolResultMessage{ToolCallID: "c", Content: []ai.ToolResultContent{nil}},
		"image-not-image":       ai.UserMessage{Content: []ai.UserContent{ai.Image{Data: "AA==", MimeType: "text/plain"}}},
		"image-parameters":      ai.UserMessage{Content: []ai.UserContent{ai.Image{Data: "AA==", MimeType: "image/png;base64,x"}}},
		"tool-image-no-type":    ai.ToolResultMessage{ToolCallID: "c", Content: []ai.ToolResultContent{ai.Image{Data: "AA=="}}},
		"section-twice":         ai.SystemMessage{Sections: []ai.SystemSection{{Name: "s", Text: "1"}, {Name: "s", Text: "2"}}},
		"section-removed-text":  ai.SystemMessage{Sections: []ai.SystemSection{{Name: "s", Text: "1", Removed: true}}},
		"tool-added-twice":      ai.SystemMessage{ToolsAdded: []ai.Tool{tool, tool}},
		"tool-added-no-schema":  ai.SystemMessage{ToolsAdded: []ai.Tool{{Name: "b", Parameters: json.RawMessage(`[]`)}}},
		"tool-removed-no-name":  ai.SystemMessage{ToolsRemoved: []string{""}},
	}
	for id, bad := range cases {
		for _, e := range outcomeEntries {
			t.Run(id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E03-history-reject-"+id+"-"+e.name)
				w := newWorld(t, tenantA)
				req := ai.Request{Messages: []ai.Message{ai.UserText("hi"), bad}}
				o := e.invoke(ctxFor(t), w, textScope("req-history-reject"), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}, req)
				o.record(ev)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseScope)
			})
		}
	}
}
