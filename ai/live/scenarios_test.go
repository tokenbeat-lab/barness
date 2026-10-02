//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// toolChoice is how a full-options call steers tool use.
type toolChoice int

const (
	// toolsDefault leaves the choice to the protocol's default, automatic.
	toolsDefault toolChoice = iota
	// toolForced forces a call of weatherTool.
	toolForced
	// toolsOff forbids tool use: the turn answering a tool result.
	toolsOff
)

// weatherTool is the smoke's one tool; the test host answers it.
var weatherTool = ai.Tool{
	Name:        "get_weather",
	Description: "Returns the current weather for a city.",
	Parameters: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string","description":"City name"}},` +
		`"required":["city"],"additionalProperties":false}`),
}

const (
	shortPrompt   = "Reply with one short sentence about the sea."
	toolPrompt    = "What is the weather in Paris right now? Use the get_weather tool, then answer in one sentence."
	weatherResult = `{"city":"Paris","condition":"sunny","celsius":18}`
)

var textStream = scenario{id: "text-stream", run: func(s *session) {
	if !s.spend() {
		return
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText(shortPrompt)}}
	d := drain(s.env.client.Stream(s.ctx, newScope(), s.target(), req, s.env.combo.full(512, toolsDefault)), nil)
	s.observe("stream", d.res, d.err)
	if s.ok("stream", d.err) {
		s.events("stream", d)
		s.turn("stream", d.res, ai.StopReasonStop)
	}
}}

var textComplete = scenario{id: "text-complete", run: func(s *session) {
	if !s.spend() {
		return
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText(shortPrompt)}}
	res, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, s.quiet(512))
	s.observe("complete", res, err)
	if s.ok("complete", err) {
		s.turn("complete", res, ai.StopReasonStop)
	}
}}

// toolRoundTrip: the model calls the tool, the test host answers it, and
// the next generation answers from the result (spec §5).
var toolRoundTrip = scenario{id: "tool-round-trip", run: func(s *session) {
	if !s.spend() {
		return
	}
	choice, how := toolsDefault, "automatic choice with a directive prompt"
	if s.env.combo.forcesTool {
		choice, how = toolForced, "forced choice"
	}
	s.note("tool call by %s", how)
	tools := []ai.Tool{weatherTool}
	user := ai.UserText(toolPrompt)
	d := drain(s.env.client.Stream(s.ctx, newScope(), s.target(), ai.Request{Messages: []ai.Message{user}, Tools: tools},
		s.env.combo.full(1024, choice)), nil)
	s.observe("call", d.res, d.err)
	if !s.ok("call", d.err) {
		return
	}
	s.events("call", d)
	if !s.turn("call", d.res, ai.StopReasonToolUse) {
		return
	}
	call, idx, ok := s.weatherCall("call", d.res.Message, tools)
	if !ok {
		return
	}
	ended := slices.ContainsFunc(d.events, func(e ai.Event) bool {
		end, ok := e.(ai.ToolCallEndEvent)
		return ok && end.ContentIndex == idx && end.ToolCall.ID == call.ID && end.ToolCall.Name == call.Name &&
			end.ToolCall.Arguments == call.Arguments
	})
	s.check("call: toolcall_end carries the message's call", ended, "call %+v at %d", call, idx)

	if !s.spend() {
		return
	}
	history := []ai.Message{user, d.res.Message, ai.ToolResultText(call, weatherResult, false)}
	res, err := s.env.client.Complete(s.ctx, newScope(), s.target(), ai.Request{Messages: history, Tools: tools},
		s.env.combo.full(512, toolsOff))
	ex := s.observe("answer", res, err)
	if !s.ok("answer", err) {
		return
	}
	s.turn("answer", res, ai.StopReasonStop)
	// The wire id of the call; for Responses pi's ID is "<call_id>|<item id>".
	wireID, _, _ := strings.Cut(call.ID, "|")
	s.check("answer: the tool result names the call", sentStrings(ex)[wireID], "call id %q not in the replayed request", wireID)
}}

// cancelAfterFirstFrame cancels the call once the first frame after start
// arrived (spec §5 首帧后取消).
var cancelAfterFirstFrame = scenario{id: "cancel-after-first-frame", run: func(s *session) {
	if !s.spend() {
		return
	}
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	req := ai.Request{Messages: []ai.Message{ai.UserText("Count from 1 to 300, one number per line, nothing else.")}}
	first := ai.EventType("")
	d := drain(s.env.client.Stream(ctx, newScope(), s.target(), req, s.env.combo.full(1024, toolsDefault)), func(e ai.Event) {
		if first == "" && e.Type() != ai.EventStart {
			first = e.Type()
			cancel()
		}
	})
	s.observe("canceled", d.res, d.err)
	if first == "" {
		if s.ok("canceled", d.err) {
			s.check("a frame arrived before the end", false, "the stream ended with %q before any frame", d.res.Message.StopReason)
		}
		return
	}
	s.note("canceled after %s", first)
	last := d.events[len(d.events)-1]
	terminal, isError := last.(ai.ErrorEvent)
	s.check("canceled: terminal error event aborted", isError && terminal.Reason == ai.StopReasonAborted, "last event %T", last)
	var e *ai.Error
	s.check("canceled: classified canceled", errors.As(d.err, &e) && e.Code == ai.CodeCanceled, "error %v", d.err)
	s.check("canceled: result aborted", d.res.Message.StopReason == ai.StopReasonAborted, "stop reason %q", d.res.Message.StopReason)
	s.check("canceled: Err matches Result", errText(d.streamErr) == errText(d.err), "Err %v, Result %v", d.streamErr, d.err)
	s.events("canceled", d)
}}

// reasoningHistory: a reasoning turn calls the tool, and the next turn
// replays its reasoning or signature natively with the tool result (spec §5
// 推理/签名历史).
var reasoningHistory = scenario{id: "reasoning-history", run: func(s *session) {
	c := s.env.combo
	if c.reasoning == "" {
		s.unsupported = c.noReplay
		return
	}
	if m, _ := s.modelInfo(); !m.Reasoning {
		s.unsupported = "model " + s.model + " has no reasoning"
		return
	}
	if !s.spend() {
		return
	}
	opts := ai.SimpleOptions{Reasoning: c.reasoning, MaxTokens: ai.Value(maxOutputTokens)}
	tools := []ai.Tool{weatherTool}
	user := ai.UserText("Use the get_weather tool to find the weather in Paris, then tell me whether to bring an umbrella.")
	res, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), ai.Request{Messages: []ai.Message{user}, Tools: tools}, opts)
	s.observe("reasoning", res, err)
	if !s.ok("reasoning", err) || !s.turn("reasoning", res, ai.StopReasonToolUse) {
		return
	}
	call, _, ok := s.weatherCall("reasoning", res.Message, tools)
	artifacts, kinds := nativeArtifacts(res.Message)
	if !ok || !s.check("reasoning: native reasoning state returned", len(artifacts) > 0, "content %+v", res.Message.Content) {
		return
	}
	s.note("returned %s", strings.Join(kinds, ", "))

	if !s.spend() {
		return
	}
	history := []ai.Message{user, res.Message, ai.ToolResultText(call, weatherResult, false)}
	res2, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), ai.Request{Messages: history, Tools: tools}, opts)
	ex := s.observe("replay", res2, err)
	if !s.ok("replay", err) {
		return
	}
	s.turn("replay", res2, ai.StopReasonStop)
	s.check("replay: native state replayed, not downgraded", res2.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}),
		"downgrades %+v", res2.Metadata.NativeStateDowngrades)
	sent := sentStrings(ex)
	s.check("replay: reasoning state on the wire", slices.ContainsFunc(artifacts, func(a string) bool { return sent[a] }),
		"none of the %d returned values was sent back", len(artifacts))
}}

var imageInput = scenario{id: "image", run: func(s *session) {
	if m, _ := s.modelInfo(); !slices.Contains(m.Input, ai.ModalityImage) {
		s.unsupported = "model " + s.model + " takes text only; its image placeholder is covered offline"
		return
	}
	if !s.spend() {
		return
	}
	data := redSquarePNG()
	req := ai.Request{Messages: []ai.Message{ai.UserMessage{Content: []ai.UserContent{
		ai.Text{Text: "What is the main color of this image? Answer with one word."},
		ai.Image{Data: data, MimeType: "image/png"},
	}}}}
	res, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, s.quiet(256))
	ex := s.observe("image", res, err)
	if !s.ok("image", err) {
		return
	}
	s.turn("image", res, ai.StopReasonStop)
	s.check("image: image bytes on the wire", slices.ContainsFunc(ex, func(e exchange) bool { return strings.Contains(e.RequestBody, data) }),
		"the request carried no image data")
	s.note("answer mentions red: %t", strings.Contains(strings.ToLower(textOf(res.Message)), "red"))
}}

// quiet is the simple options of a turn that is not about reasoning.
func (s *session) quiet(maxTokens int) ai.SimpleOptions {
	return ai.SimpleOptions{Reasoning: s.env.combo.quiet, MaxTokens: ai.Value(maxTokens)}
}

func (s *session) modelInfo() (ai.Model, bool) {
	c := s.env.combo
	i := slices.IndexFunc(catalog.Models, func(m ai.Model) bool { return m.Provider == c.provider && m.API == c.api && m.ID == s.model })
	if i < 0 {
		return ai.Model{}, false
	}
	return catalog.Models[i], true
}

// weatherCall finds the message's get_weather call and validates it as a
// host does before executing it.
func (s *session) weatherCall(label string, m ai.AssistantMessage, tools []ai.Tool) (ai.ToolCall, int, bool) {
	for i, c := range m.Content {
		call, ok := c.(ai.ToolCall)
		if !ok {
			continue
		}
		valid, err := ai.ValidateToolCall(m, i, tools)
		var args struct{ City string }
		ok = s.check(label+": tool call validates", err == nil && valid.Name() == weatherTool.Name &&
			valid.Decode(&args) == nil && args.City != "", "call %+v: %v", call, err)
		return call, i, ok
	}
	s.check(label+": tool call present", false, "content %+v", m.Content)
	return ai.ToolCall{}, 0, false
}

// nativeArtifacts are the reasoning values a message carries that a native
// replay sends back: thinking text, signatures and Responses' encrypted
// reasoning.
func nativeArtifacts(m ai.AssistantMessage) (values, kinds []string) {
	add := func(kind, v string) {
		if v != "" {
			values = append(values, v)
			if !slices.Contains(kinds, kind) {
				kinds = append(kinds, kind)
			}
		}
	}
	for _, c := range m.Content {
		switch b := c.(type) {
		case ai.Thinking:
			add("thinking text", b.Thinking)
			add("thinking signature", b.Signature)
			var item struct {
				EncryptedContent string `json:"encrypted_content"`
			}
			if json.Unmarshal([]byte(b.Signature), &item) == nil {
				add("encrypted reasoning", item.EncryptedContent)
			}
			if b.Redacted {
				add("redacted thinking", b.Signature)
			}
		case ai.ToolCall:
			if sig, ok := b.ThoughtSignature.Get(); ok {
				add("tool call thought signature", sig)
			}
		case ai.Text:
			add("text signature", b.Signature)
		}
	}
	return values, kinds
}

// sentStrings is every JSON string value in the captured request bodies.
func sentStrings(ex []exchange) map[string]bool {
	out := map[string]bool{}
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case string:
			out[x] = true
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	for _, e := range ex {
		var v any
		if json.Unmarshal([]byte(e.RequestBody), &v) == nil {
			walk(v)
		}
	}
	return out
}

// redSquarePNG is a 32×32 solid red PNG in base64.
func redSquarePNG() string {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := range 32 {
		for x := range 32 {
			img.Set(x, y, color.RGBA{R: 220, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
