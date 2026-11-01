package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func catalogChatContract(t *testing.T, ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity, scene string) {
	m, _ := cat.Lookup(i.Provider, i.API, i.ID)
	if scene == "load-query" {
		catalogContractQuery(ev, cat, i)
		return
	}
	w, _ := newCatalogContractWorld(t, cat, i)
	req := ai.Request{SystemPrompt: "Catalog system.", Messages: []ai.Message{ai.UserText("Catalog input.")}}
	opts := ai.SimpleOptions{MaxTokens: ai.Value(64)}
	if scene == "authorization" {
		catalogContractDenied(ev, w, i, func(scope ai.CallScope, target ai.Target) error {
			_, err := w.client.CompleteSimple(ctxFor(t), scope, target, req, opts)
			return err
		})
		return
	}
	if scene == "reasoning" {
		catalogChatReasoning(t, ev, w, cat, i, m, req)
		return
	}
	if scene == "parameters" {
		catalogChatParameters(t, ev, w, cat, i, m, req)
		return
	}
	input := int64(8)
	if strings.HasPrefix(scene, "tier-") {
		var index int
		var boundary string
		_, err := fmt.Sscanf(scene, "tier-%d-%s", &index, &boundary)
		if err != nil {
			t.Fatal(err)
		}
		input = m.Cost.Tiers[index].InputTokensAbove
		switch boundary {
		case "below":
			input = max(0, input-1)
		case "above":
			input++
		}
	}
	if scene == "input-modalities" {
		req.Messages = []ai.Message{ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: "Catalog input."}, ai.Image{Data: catalogContractPNG, MimeType: "image/png"}}}}
	}
	if scene == "history-tools" {
		catalogChatHistory(t, ev, w, cat, i, m, req, opts)
		return
	}
	enqueue(ev, w, catalogChatReply(t, m, input, false))
	result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-"+scene), catalogContractTarget(i), req, opts)
	ev.Record("result", result)
	ev.Record("requests", w.provider.Requests())
	ev.Record("error", errString(err))
	ev.Check("public chat completes", err == nil && result.Message.StopReason == ai.StopReasonStop, "%v %+v", err, result.Message)
	if err != nil {
		return
	}
	catalogChatResult(ev, cat, i, m, result, input)
	body := catalogChatWire(ev, w, i)
	if scene == "input-modalities" {
		image := strings.Contains(string(body), catalogContractPNG)
		ev.Check("declared image input or text placeholder", image == slices.Contains(m.Input, ai.ModalityImage), "%s", body)
	}
}

func catalogChatWire(ev *evidence.Case, w *world, i catalogContractIdentity) []byte {
	requests := w.provider.Requests()
	ev.Check("controlled request observed", len(requests) > 0, "no request")
	if len(requests) == 0 {
		return nil
	}
	r := requests[len(requests)-1]
	var body map[string]json.RawMessage
	ev.Check("wire is a JSON object", json.Unmarshal(r.Body, &body) == nil, "%s", r.Body)
	if i.API == ai.APIGoogleGenerativeAI {
		ev.Check("Google model in target path", strings.Contains(r.Path, i.ID+":"), "%s", r.Path)
	} else {
		var id string
		_ = json.Unmarshal(body["model"], &id)
		ev.Check("wire model exact", id == i.ID, "%s", r.Body)
	}
	ev.Check("tenant credential", r.KeyAlias == tenantA.alias, "%s", r.KeyAlias)
	return r.Body
}
func catalogChatResult(ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity, m ai.Model, result ai.Result, input int64) {
	catalogContractMetadata(ev, cat, i, result.Metadata)
	u := result.Message.Usage
	ev.Check("scripted counters survive", u.Input == input && u.Output == 4 && u.CacheRead == 0 && u.CacheWrite == 0 && u.TotalTokens == input+4, "%+v", u)
	rates := m.Cost.CostRates
	// Independent whole-request pricing expectation comes from the confirmed
	// snapshot and literal scripted counts, never ai's estimate implementation.
	highest := int64(-1)
	for _, tier := range m.Cost.Tiers {
		if input > tier.InputTokensAbove && tier.InputTokensAbove >= highest {
			rates = tier.CostRates
			highest = tier.InputTokensAbove
		}
	}
	expectedInput := float64(input) * rates.Input / 1e6
	expectedOutput := 4 * rates.Output / 1e6
	ev.Check("independent token cost", math.Abs(u.Cost.Input-expectedInput) < 1e-12 && math.Abs(u.Cost.Output-expectedOutput) < 1e-12 && math.Abs(u.Cost.Total-expectedInput-expectedOutput) < 1e-12, "got=%+v input=%g output=%g", u.Cost, expectedInput, expectedOutput)
	if len(result.Metadata.Attempts) == 1 {
		ev.Check("complete native Usage", result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", result.Metadata.Attempts[0])
	}
}
func catalogChatHistory(t *testing.T, ev *evidence.Case, w *world, cat ai.Catalog, i catalogContractIdentity, m ai.Model, req ai.Request, opts ai.SimpleOptions) {
	req.Tools = []ai.Tool{{Name: "catalog_tool", Description: "Catalog tool", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}}
	enqueue(ev, w, catalogChatReply(t, m, 8, true))
	first, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-tool-first"), catalogContractTarget(i), req, opts)
	ev.Record("first", first)
	ev.Check("tool turn accepted", err == nil && first.Message.StopReason == ai.StopReasonToolUse, "%v %+v", err, first.Message)
	if err != nil {
		return
	}
	var tool ai.ToolCall
	found := false
	for _, block := range first.Message.Content {
		if call, ok := block.(ai.ToolCall); ok {
			tool = call
			found = true
		}
	}
	ev.Check("typed tool result", found && tool.Name == "catalog_tool" && tool.RawArguments == "{}", "%+v", tool)
	if !found {
		return
	}
	req.Messages = append(req.Messages, first.Message, ai.ToolResultText(tool, "Catalog tool result.", false), ai.SystemMessage{Content: "Catalog later system."}, ai.UserText("Catalog followup."))
	enqueue(ev, w, catalogChatReply(t, m, 8, false))
	result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-tool-next"), catalogContractTarget(i), req, opts)
	ev.Record("result", result)
	ev.Record("requests", w.provider.Requests())
	ev.Check("tool history replay accepted", err == nil, "%v", err)
	if err != nil {
		return
	}
	catalogChatResult(ev, cat, i, m, result, 8)
	body := catalogChatWire(ev, w, i)
	ev.Check("tool result and later prompt retained", strings.Contains(string(body), "Catalog tool result.") && strings.Contains(string(body), "Catalog later system.") && strings.Contains(string(body), "Catalog followup."), "%s", body)
	if i.API == ai.APIOpenAIResponses {
		hasStrict := strings.Contains(string(body), `"strict":`)
		ev.Check("declared strict wire", hasStrict == m.Compat.SupportsStrictMode, "%s", body)
	}
}
