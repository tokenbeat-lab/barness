package ai

import (
	"encoding/json"
	"reflect"
)

// All mutable protocol facts belong to one call. xAI's stricter snapshot
// contract deliberately leaves the established Responses providers unchanged.
type xaiResponsesState struct {
	items    map[int64]*xaiStreamItem
	ids      map[string]int64
	calls    map[string]int64
	terminal json.RawMessage
	choice   string
	parallel bool
}
type xaiStreamItem struct {
	kind, id, call, name string
	index                int
	done                 bool
	native               json.RawMessage
	parts                map[string]map[int64]*xaiStreamPart
	primary              string
	emitted              string
	args                 *argumentText
	argsDone             bool
}
type xaiStreamPart struct {
	text *argumentText
	done bool
}

func newXAIResponsesState(provider ProviderID, body []byte) *xaiResponsesState {
	if provider != ProviderXAI {
		return nil
	}
	var controls struct {
		Choice   string `json:"tool_choice"`
		Parallel *bool  `json:"parallel_tool_calls"`
	}
	_ = json.Unmarshal(body, &controls)
	return &xaiResponsesState{items: map[int64]*xaiStreamItem{}, ids: map[string]int64{}, calls: map[string]int64{}, choice: controls.Choice, parallel: controls.Parallel == nil || *controls.Parallel}
}
func xaiStreamFailure() *Error {
	return newError(CodeProtocol, PhaseStream, "xAI Responses stream contains malformed or contradictory data")
}
func xaiStreamObject(raw []byte) (map[string]json.RawMessage, bool) {
	obj, duplicate, err := unaryJSONEnvelope(raw)
	return obj, err == nil && !duplicate
}
func xaiStreamEqual(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
func (s *xaiResponsesState) handle(p *responsesParser, raw []byte, eventType string) *Error {
	ev, ok := xaiStreamObject(raw)
	if !ok {
		return xaiStreamFailure()
	}
	typ, ok := rawString(ev["type"])
	if eventType == "error" {
		typ, ok = "error", true
	}
	if !ok {
		return xaiStreamFailure()
	}
	if p.terminal {
		if typ == "response.completed" || typ == "response.incomplete" {
			if xaiStreamEqual(s.terminal, ev["response"]) {
				return nil
			}
		}
		return xaiStreamFailure()
	}
	if seq, has := ev["sequence_number"]; has {
		if _, ok := xaiStreamIndex(seq); !ok {
			return xaiStreamFailure()
		}
	}
	switch typ {
	case "error":
		return newError(CodeUpstreamError, PhaseStream, "xAI Responses reported an upstream error")
	case "response.created", "response.in_progress":
		r, ok := xaiStreamObject(ev["response"])
		if !ok {
			return xaiStreamFailure()
		}
		if model, has := r["model"]; has {
			value, ok := rawString(model)
			if !ok || value != p.usage.model.ID {
				return xaiStreamFailure()
			}
			p.out.responseModel(value)
		}
		if id, has := r["id"]; has {
			value, ok := rawString(id)
			if !ok {
				return xaiStreamFailure()
			}
			p.out.responseID(value)
		}
		return nil
	case "response.completed", "response.incomplete", "response.failed":
		return s.finish(p, typ, ev["response"])
	case "response.output_item.added", "response.output_item.done":
		key, ok := xaiStreamIndex(ev["output_index"])
		if !ok {
			return xaiStreamFailure()
		}
		item, failure := s.item(p, key, ev["item"])
		if failure != nil {
			return failure
		}
		if typ == "response.output_item.done" {
			return s.closeItem(p, item, ev["item"])
		}
		if item.done {
			return xaiStreamFailure()
		}
		return nil
	case "response.output_text.delta", "response.output_text.done", "response.refusal.delta", "response.refusal.done",
		"response.reasoning_text.delta", "response.reasoning_text.done", "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done",
		"response.content_part.added", "response.content_part.done", "response.reasoning_summary_part.added", "response.reasoning_summary_part.done":
		return s.text(p, typ, ev)
	case "response.function_call_arguments.delta", "response.function_call_arguments.done":
		return s.arguments(p, typ, ev)
	default:
		return xaiStreamFailure()
	}
}
func (s *xaiResponsesState) associated(ev map[string]json.RawMessage, kind string) (*xaiStreamItem, bool) {
	key, ok := xaiStreamIndex(ev["output_index"])
	if !ok {
		return nil, false
	}
	item, ok := s.items[key]
	if !ok || item.kind != kind || item.done {
		return nil, false
	}
	if raw, has := ev["item_id"]; has {
		id, ok := rawString(raw)
		if !ok || id != item.id {
			return nil, false
		}
	}
	if raw, has := ev["call_id"]; has {
		id, ok := rawString(raw)
		if !ok || id != item.call {
			return nil, false
		}
	}
	return item, true
}
func (s *xaiResponsesState) finish(p *responsesParser, typ string, raw json.RawMessage) *Error {
	r, ok := xaiStreamObject(raw)
	if !ok {
		return xaiStreamFailure()
	}
	status, ok := rawString(r["status"])
	if !ok || typ != "response."+status {
		return xaiStreamFailure()
	}
	if model, has := r["model"]; has {
		value, ok := rawString(model)
		if !ok || value != p.usage.model.ID {
			return xaiStreamFailure()
		}
		p.out.responseModel(value)
	}
	p.out.rawStopReason(status)
	if status == "failed" {
		p.terminal = true
		return newError(CodeUpstreamError, PhaseStream, "xAI Responses reported a failed response")
	}
	var output []json.RawMessage
	if json.Unmarshal(r["output"], &output) != nil || output == nil {
		return xaiStreamFailure()
	}
	for i, rawItem := range output {
		item, failure := s.item(p, int64(i), rawItem)
		if failure != nil {
			return failure
		}
		if failure := s.closeItem(p, item, rawItem); failure != nil {
			return failure
		}
	}
	for key := range s.items {
		if key >= int64(len(output)) {
			return xaiStreamFailure()
		}
	}
	// The generic finalizer only handles status and usage for this branch; xAI
	// has already reconciled all authoritative output without typed coercions.
	if id, has := r["id"]; has {
		value, ok := rawString(id)
		if !ok {
			return xaiStreamFailure()
		}
		p.out.responseID(value)
	}
	var response struct {
		Usage   json.RawMessage `json:"usage"`
		Tier    json.RawMessage `json:"service_tier"`
		Details struct {
			Reason string `json:"reason"`
		} `json:"incomplete_details"`
	}
	if json.Unmarshal(raw, &response) != nil {
		return xaiStreamFailure()
	}
	u, reporting, estimate, failure := xaiUsage(p.usage.model, response.Usage)
	if failure != nil {
		return failure
	}
	if len(response.Tier) != 0 && string(response.Tier) != "null" {
		tier, ok := rawString(response.Tier)
		if !ok || tier != "default" {
			estimate = CostEstimate{Status: CostPartial}
			p.out.usage(u)
			p.attempt.usageEstimate(reporting, u, estimate)
			return xaiStreamFailure()
		}
	}
	p.out.usage(u)
	p.attempt.usageEstimate(reporting, u, estimate)
	reason := response.Details.Reason
	if reason != "" {
		p.out.rawStopReason(status + "." + reason)
	}
	p.terminal = true
	s.terminal = append(json.RawMessage(nil), raw...)
	if status == "incomplete" {
		if reason != "max_output_tokens" {
			return newError(CodeUpstreamError, PhaseStream, "xAI Responses reported an incomplete response")
		}
		p.stop = StopReasonLength
		p.out.stop(p.stop)
		return nil
	}
	if s.choice == "required" && len(s.calls) == 0 || s.choice == "none" && len(s.calls) > 0 || !s.parallel && len(s.calls) > 1 {
		return xaiStreamFailure()
	}
	p.stop = StopReasonStop
	if len(s.calls) > 0 {
		p.stop = StopReasonToolUse
	}
	p.out.stop(p.stop)
	return nil
}

func xaiStreamIndex(raw json.RawMessage) (int64, bool) {
	var n int64
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &n) != nil || n < 0 {
		return 0, false
	}
	return n, true
}
