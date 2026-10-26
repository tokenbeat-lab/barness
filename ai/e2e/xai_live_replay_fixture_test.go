package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type xaiRealManifest struct {
	Schema             int    `json:"schema"`
	Kind               string `json:"kind"`
	SourceBundle       string `json:"sourceBundle"`
	SourceReportSHA256 string `json:"sourceReportSHA256"`
	NativeSourceSHA256 string `json:"nativeSourceSHA256"`
	Captures           []struct {
		Scenario, Label, File, SHA256 string
		OriginalFile, OriginalSHA256  string
	} `json:"captures"`
}

type xaiRealCapture struct {
	Scenario, Label string
	Request         json.RawMessage
	ResponseBody    string
	Expected        struct {
		Text      string
		Usage     ai.Usage
		Reasoning []json.RawMessage
		Calls     []struct{ CallID, ItemID, Name, Arguments string }
	}
}

type xaiRealInput struct {
	Type, Role, Output string
	CallID             string `json:"call_id"`
	Content            json.RawMessage
}

type xaiRealRequest struct {
	Input             []xaiRealInput
	Tools             []ai.Tool
	MaxOutputTokens   int    `json:"max_output_tokens"`
	ToolChoice        string `json:"tool_choice"`
	ParallelToolCalls *bool  `json:"parallel_tool_calls"`
	PromptCacheKey    string `json:"prompt_cache_key"`
	Reasoning         struct{ Effort ai.ThinkingLevel }
}

func loadXAIRealCaptures(t *testing.T) (xaiRealManifest, []xaiRealCapture, []byte) {
	t.Helper()
	base := filepath.Join("testdata", "xai-responses")
	raw, err := os.ReadFile(filepath.Join(base, "manifest.json"))
	if err != nil {
		t.Fatalf("real replay assembly: %v", err)
	}
	var manifest xaiRealManifest
	mustUnmarshal(t, raw, &manifest)
	var captures []xaiRealCapture
	for _, entry := range manifest.Captures {
		if filepath.Base(entry.File) != entry.File || entry.File == "." {
			t.Fatal("unsafe replay fixture path")
		}
		data, err := os.ReadFile(filepath.Join(base, entry.File))
		if err != nil {
			t.Fatalf("real replay assembly: %v", err)
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != entry.SHA256 {
			t.Fatal("real replay fixture digest differs from provenance")
		}
		var capture xaiRealCapture
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&capture); err != nil {
			t.Fatalf("replay fixture contains unexpected fields: %v", err)
		}
		if capture.Scenario != entry.Scenario || capture.Label != entry.Label {
			t.Fatal("real replay capture identity differs from provenance")
		}
		captures = append(captures, capture)
	}
	return manifest, captures, raw
}

func checkXAIRealSanitization(t *testing.T, raw []byte) bool {
	t.Helper()
	var value any
	mustUnmarshal(t, raw, &value)
	var safe func(any) bool
	safe = func(v any) bool {
		switch node := v.(type) {
		case []any:
			for _, item := range node {
				if !safe(item) {
					return false
				}
			}
		case map[string]any:
			for key, item := range node {
				text, _ := item.(string)
				switch key {
				case "encrypted_content":
					if !strings.HasPrefix(text, "replay-encrypted-") {
						return false
					}
				case "id", "item_id", "call_id":
					if text != "" && !strings.HasPrefix(text, "replay-id-") {
						return false
					}
				case "prompt_cache_key":
					// Null is the captured absence of a hint, not a scoped value.
					if item != nil && text != "replay-cache-hint" {
						return false
					}
				case "requestHeaders", "responseHeaders", "authorization", "set-cookie", "traceparent":
					return false
				}
				if !safe(item) {
					return false
				}
			}
		}
		return true
	}
	return safe(value)
}

func (c xaiRealCapture) reply() provider.Reply {
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "text/event-stream"}, Script: c.ResponseBody, Chunks: [][]byte{[]byte(c.ResponseBody)}}
}

func (c xaiRealCapture) body(t *testing.T) xaiRealRequest {
	t.Helper()
	var body xaiRealRequest
	mustUnmarshal(t, c.Request, &body)
	return body
}

func xaiRealUser(t *testing.T, raw json.RawMessage) ai.UserMessage {
	t.Helper()
	var blocks []struct{ Type, Text, ImageURL string }
	var wire []map[string]string
	mustUnmarshal(t, raw, &wire)
	for _, b := range wire {
		blocks = append(blocks, struct{ Type, Text, ImageURL string }{b["type"], b["text"], b["image_url"]})
	}
	var user ai.UserMessage
	for _, b := range blocks {
		switch b.Type {
		case "input_text":
			user.Content = append(user.Content, ai.Text{Text: b.Text})
		case "input_image":
			prefix, data, ok := strings.Cut(b.ImageURL, ";base64,")
			if !ok || !strings.HasPrefix(prefix, "data:image/") {
				t.Fatal("real replay image must remain inline")
			}
			user.Content = append(user.Content, ai.Image{Data: data, MimeType: strings.TrimPrefix(prefix, "data:")})
		default:
			t.Fatalf("unknown captured user block %q", b.Type)
		}
	}
	return user
}

func (c xaiRealCapture) initial(t *testing.T) ai.Request {
	t.Helper()
	body := c.body(t)
	req := ai.Request{Tools: body.Tools}
	for _, item := range body.Input {
		switch item.Role {
		case "system":
			mustUnmarshal(t, item.Content, &req.SystemPrompt)
		case "user":
			req.Messages = append(req.Messages, xaiRealUser(t, item.Content))
		default:
			t.Fatal("first captured request unexpectedly contains native history")
		}
	}
	return req
}

func (c xaiRealCapture) options(t *testing.T) ai.ResponsesOptions {
	t.Helper()
	body := c.body(t)
	o := ai.ResponsesOptions{MaxTokens: ai.Value(body.MaxOutputTokens), ReasoningEffort: body.Reasoning.Effort}
	if body.ToolChoice != "" {
		o.ToolChoice = ai.Value(ai.ResponsesToolChoice{Mode: body.ToolChoice})
	}
	if body.ParallelToolCalls != nil {
		o.SamplingParams = map[string]json.RawMessage{"parallel_tool_calls": mustMarshal(t, *body.ParallelToolCalls)}
	}
	if body.PromptCacheKey != "" {
		o.SessionID, o.CacheRetention = "synthetic-real-replay-cache-session", ai.CacheRetentionShort
	}
	return o
}

// This deliberately reconstructs Host work from the previous public Result.
// Captured protocol IDs are only an independent pairing oracle, never the IDs
// passed back as domain ToolCallIDs. The live-tag Host cannot be imported here.
func (c xaiRealCapture) continueRequest(t *testing.T, req ai.Request, previous ai.Result, source xaiRealCapture) ai.Request {
	t.Helper()
	if c.Scenario == "cache-usage" {
		return req
	}
	var envelope *ai.NativeStateEnvelope
	restored, envelope, _ := restore(t, persist(t, previous.Message))
	if envelope == nil {
		t.Fatal("captured Result lost native envelope during Host persistence")
	}
	var err error
	restored.NativeState, err = ai.TrustNativeState(textScope(t.Name()), *envelope)
	if err != nil {
		t.Fatal(err)
	}
	req.Messages = append(req.Messages, restored)
	input := c.body(t).Input
	if c.Scenario == "encrypted-continuation" {
		req.Messages = append(req.Messages, xaiRealUser(t, input[len(input)-1].Content))
		return req
	}
	var actual []ai.ToolCall
	for index, block := range previous.Message.Content {
		if call, ok := block.(ai.ToolCall); ok {
			if _, err := ai.ValidateToolCall(previous.Message, index, req.Tools); err != nil {
				t.Fatal(err)
			}
			actual = append(actual, call)
		}
	}
	for index, item := range input {
		if item.Type != "function_call_output" {
			continue
		}
		found := -1
		for j, call := range source.Expected.Calls {
			if call.CallID == item.CallID {
				found = j
			}
		}
		if found < 0 || found >= len(actual) {
			t.Fatal("captured tool output has no independent public call")
		}
		text := strings.TrimSuffix(item.Output, "\n(see attached image)")
		result := ai.ToolResultText(actual[found], text, false)
		if index+1 < len(input) && input[index+1].Role == "user" {
			for _, block := range xaiRealUser(t, input[index+1].Content).Content {
				image, ok := block.(ai.Image)
				if !ok {
					t.Fatal("tool image projection contains unexpected text")
				}
				result.Content = append(result.Content, image)
			}
		}
		req.Messages = append(req.Messages, result)
	}
	return req
}
