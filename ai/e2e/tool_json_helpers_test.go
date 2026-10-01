package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestToolJSONHelpers covers the public tool JSON helpers in isolation (spec
// I6, Testing Decisions §6: "先枚举所有失败方式并写行为测试"). They are a system of
// their own: the stream never calls them, and nothing in barness-ai executes a
// tool. Every way they can refuse a call is enumerated here first:
//
// Executability gate (truncated or unfinished turns are never executable):
//
//	G1 message ended with StopReasonLength, arguments complete → not_executable
//	G2 message ended with StopReasonError → not_executable
//	G3 message ended with StopReasonAborted → not_executable
//	G4 message ended with StopReasonStop (no tool use) → not_executable
//	G5 live snapshot (StopReasonPending) → not_executable
//
// Addressing:
//
//	A1 content index negative or past the end → not_tool_call
//	A2 content index holds a text block → not_tool_call
//
// Tool lookup and schema:
//
//	S1 no declared tool has the call's name (including a blank one) → unknown_tool, pi's text
//	S2 Parameters is not JSON → invalid_schema
//	S3 Parameters violates the JSON Schema metaschema → invalid_schema
//	S4 Parameters $ref points at an http resource → invalid_schema, never fetched
//	S5 Parameters $ref points at a file → invalid_schema
//
// Argument JSON:
//
//	J1 truncated object → malformed_json (repair never completes JSON)
//	J2 malformed beyond repair (trailing comma, single quotes, trailing data) → malformed_json
//	J3 valid JSON that is not an object (array, string, number, null) → malformed_json
//	J4 raw JSON truncated although the display arguments look complete → malformed_json
//
// Schema conversion and validation (pi-ai's plain-JSON-schema rules):
//
//	V1 required property missing → invalid_arguments, path names the property
//	V2 value of a non-convertible type → invalid_arguments
//	V3 conversions pi rejects (boolean "1"/"0", null "null", integer "42.1") → invalid_arguments
//	V4 additionalProperties false with an extra key → invalid_arguments
//
// Accepted inputs, whose outcome must not be a refusal:
//
//	OK1 conversions pi applies (string→number, bool→number, null→0, …)
//	OK2 null for an optional non-nullable property is dropped; nullable keeps null
//	OK3 union: a value matching an arm is kept; otherwise the first converting arm wins
//	OK4 repairable JSON (raw control character, invalid escape) is repaired
//	OK5 no raw text: the display arguments are validated (message read from pi-shaped storage)
//	OK6 inputs are not modified and the result shares no storage with them
//
// RepairToolJSON:
//
//	R1 valid JSON unchanged; R2 raw control characters in strings escaped;
//	R3 invalid escapes doubled; R4 trailing lone backslash doubled;
//	R5 \u without four hex digits kept (pi counts "u" as a valid escape), so
//	   the text stays invalid; R6 nothing outside strings changes;
//	R7 truncated JSON stays truncated.
func TestToolJSONHelpers(t *testing.T) {
	weather := ai.Tool{Name: "get_weather", Description: "Current weather", Parameters: json.RawMessage(
		`{"type":"object","properties":{"city":{"type":"string"},"unit":{"type":"string","enum":["c","f"]}},"required":["city"],"additionalProperties":false}`)}

	t.Run("gate", func(t *testing.T) {
		for _, stop := range []ai.StopReason{ai.StopReasonLength, ai.StopReasonError, ai.StopReasonAborted, ai.StopReasonStop, ai.StopReasonPending} {
			t.Run(string(stop), func(t *testing.T) {
				ev := run.Case(t, "E03-tooljson-gate-"+string(stop))
				msg := toolMessage(stop, ai.ToolCall{ID: "call_1|fc_1", Name: "get_weather", Arguments: `{"city":"Paris"}`, RawArguments: `{"city":"Paris"}`})
				_, err := ai.ValidateToolCall(msg, 0, []ai.Tool{weather})
				expectRefusal(ev, err, ai.ToolCallNotExecutable)
			})
		}
	})

	t.Run("addressing", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-addressing")
		msg := toolMessage(ai.StopReasonToolUse, ai.Text{Text: "Checking."}, ai.ToolCall{ID: "call_1|fc_1", Name: "get_weather", RawArguments: `{"city":"Paris"}`})
		for _, i := range []int{-1, 2, 0} {
			_, err := ai.ValidateToolCall(msg, i, []ai.Tool{weather})
			expectRefusal(ev, err, ai.ToolCallNotFound)
		}
	})

	t.Run("unknown-tool", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-unknown-tool")
		msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "get_time", RawArguments: `{}`})
		_, err := ai.ValidateToolCall(msg, 0, []ai.Tool{weather})
		if e := expectRefusal(ev, err, ai.ToolCallUnknownTool); e != nil {
			ev.Check("pi's message", e.Message == `Tool "get_time" not found`, "got %q", e.Message)
		}
		_, err = ai.ValidateToolCall(toolMessage(ai.StopReasonToolUse, ai.ToolCall{}), 0, []ai.Tool{weather})
		expectRefusal(ev, err, ai.ToolCallUnknownTool)
	})

	t.Run("invalid-schema", func(t *testing.T) {
		var fetched atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fetched.Add(1)
			_, _ = w.Write([]byte(`{"type":"string"}`))
		}))
		defer srv.Close()
		// A valid schema on disk: loading it would succeed, so only a refusal
		// to load file references can make this case fail compilation.
		cityFile := filepath.Join(t.TempDir(), "city.json")
		if err := os.WriteFile(cityFile, []byte(`{"type":"string"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		schemas := map[string]string{
			"not-json":    `{"type":`,
			"metaschema":  `{"type":12}`,
			"remote-ref":  `{"type":"object","properties":{"city":{"$ref":"` + srv.URL + `/city.json"}}}`,
			"file-ref":    `{"type":"object","properties":{"city":{"$ref":"file://` + filepath.ToSlash(cityFile) + `"}}}`,
			"missing-ref": `{"type":"object","properties":{"city":{"$ref":"#/$defs/absent"}}}`,
		}
		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "E03-tooljson-invalid-schema-"+name)
				tool := ai.Tool{Name: "get_weather", Parameters: json.RawMessage(schema)}
				msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "get_weather", RawArguments: `{"city":"Paris"}`})
				_, err := ai.ValidateToolCall(msg, 0, []ai.Tool{tool})
				expectRefusal(ev, err, ai.ToolCallInvalidSchema)
			})
		}
		if fetched.Load() != 0 {
			t.Errorf("schema validation fetched a remote $ref %d times", fetched.Load())
		}
	})

	t.Run("malformed-json", func(t *testing.T) {
		raws := map[string]struct{ raw, display string }{
			"truncated":        {`{"city":"Par`, ""},
			"truncated-nested": {`{"city":"Paris","opts":{"unit":"c"`, ""},
			"trailing-comma":   {`{"city":"Paris",}`, ""},
			"single-quotes":    {`{'city':'Paris'}`, ""},
			"trailing-data":    {`{"city":"Paris"} {"city":"Rome"}`, ""},
			"array":            {`["Paris"]`, ""},
			"string":           {`"Paris"`, ""},
			"number":           {`42`, ""},
			"null":             {`null`, ""},
			// The display parse completed the truncated text; the raw text decides.
			"display-complete": {`{"city":"Paris"`, `{"city":"Paris"}`},
		}
		for name, c := range raws {
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "E03-tooljson-malformed-"+name)
				msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "get_weather", Arguments: ai.UncheckedArguments(c.display), RawArguments: c.raw})
				_, err := ai.ValidateToolCall(msg, 0, []ai.Tool{weather})
				expectRefusal(ev, err, ai.ToolCallMalformedJSON)
			})
		}
	})

	t.Run("invalid-arguments", func(t *testing.T) {
		cases := map[string]struct {
			raw   string
			paths []string
		}{
			"missing-required": {`{"unit":"c"}`, []string{"city"}},
			"wrong-type":       {`{"city":{"name":"Paris"}}`, []string{"city"}},
			"enum":             {`{"city":"Paris","unit":"kelvin"}`, []string{"unit"}},
			"additional":       {`{"city":"Paris","country":"FR"}`, nil},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "E03-tooljson-invalid-arguments-"+name)
				msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "get_weather", RawArguments: c.raw})
				_, err := ai.ValidateToolCall(msg, 0, []ai.Tool{weather})
				e := expectRefusal(ev, err, ai.ToolCallInvalidArguments)
				if e == nil {
					return
				}
				ev.Check("problems reported", len(e.Problems) > 0, "got none")
				var paths []string
				for _, p := range e.Problems {
					paths = append(paths, p.Path)
				}
				for _, want := range c.paths {
					ev.Check("problem path "+want, contains(paths, want), "got %q", paths)
				}
				ev.Check("pi's message prefix", len(e.Message) > 0 && e.Message[:len(`Validation failed for tool "get_weather":`)] == `Validation failed for tool "get_weather":`, "got %q", e.Message)
			})
		}
	})

	// V3, OK1–OK3: pi-ai's validation.test.ts conversion tables, on a
	// {"value": <schema>} wrapper as pi builds them.
	t.Run("conversion", func(t *testing.T) {
		cases := []struct {
			name, schema, input, want string // want "" means refused
		}{
			{"number-from-string", `{"type":"number"}`, `"42"`, `42`},
			{"number-from-true", `{"type":"number"}`, `true`, `1`},
			{"number-from-null", `{"type":"number"}`, `null`, `0`},
			{"integer-from-string", `{"type":"integer"}`, `"42"`, `42`},
			{"boolean-from-true", `{"type":"boolean"}`, `"true"`, `true`},
			{"boolean-from-false", `{"type":"boolean"}`, `"false"`, `false`},
			{"boolean-from-1", `{"type":"boolean"}`, `1`, `true`},
			{"boolean-from-0", `{"type":"boolean"}`, `0`, `false`},
			{"string-from-null", `{"type":"string"}`, `null`, `""`},
			{"string-from-true", `{"type":"string"}`, `true`, `"true"`},
			{"null-from-empty", `{"type":"null"}`, `""`, `null`},
			{"null-from-0", `{"type":"null"}`, `0`, `null`},
			{"null-from-false", `{"type":"null"}`, `false`, `null`},
			{"union-keeps-match", `{"type":["number","string"]}`, `"1"`, `"1"`},
			{"union-converts", `{"type":["boolean","number"]}`, `"1"`, `1`},
			{"anyof-keeps-null", `{"anyOf":[{"type":"number"},{"type":"null"}]}`, `null`, `null`},
			{"oneof-keeps-null", `{"oneOf":[{"type":"number"},{"type":"null"}]}`, `null`, `null`},
			{"anyof-converts", `{"anyOf":[{"type":"number"},{"type":"null"}]}`, `"42"`, `42`},
			{"nullable-array", `{"type":["array","null"],"items":{"type":"string"}}`, `null`, `null`},
			{"refuse-boolean-from-1", `{"type":"boolean"}`, `"1"`, ``},
			{"refuse-boolean-from-0", `{"type":"boolean"}`, `"0"`, ``},
			{"refuse-null-from-string", `{"type":"null"}`, `"null"`, ``},
			{"refuse-integer-from-fraction", `{"type":"integer"}`, `"42.1"`, ``},
			{"refuse-number-from-word", `{"type":"number"}`, `"forty-two"`, ``},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				ev := run.Case(t, "E03-tooljson-conversion-"+c.name)
				tool := ai.Tool{Name: "echo", Parameters: json.RawMessage(`{"type":"object","properties":{"value":` + c.schema + `},"required":["value"]}`)}
				msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "echo", RawArguments: `{"value":` + c.input + `}`})
				got, err := ai.ValidateToolCall(msg, 0, []ai.Tool{tool})
				if c.want == "" {
					expectRefusal(ev, err, ai.ToolCallInvalidArguments)
					return
				}
				ev.Check("accepted", err == nil, "err=%v", err)
				ev.Check("converted value", jsonEqual(got.Arguments(), []byte(`{"value":`+c.want+`}`)), "got %s want {\"value\":%s}", got.Arguments(), c.want)
			})
		}
	})

	t.Run("optional-nulls", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-optional-nulls")
		tool := ai.Tool{Name: "echo", Parameters: json.RawMessage(`{"type":"object","properties":{` +
			`"path":{"type":"string"},"offset":{"type":"number"},` +
			`"nullable":{"anyOf":[{"type":"string"},{"type":"null"}]},` +
			`"metadata":{"type":"object","properties":{"enabled":{"type":"boolean"}}},` +
			`"ref":{"$ref":"#/$defs/value"}},` +
			`"required":["path","metadata"],"$defs":{"value":{"anyOf":[{"type":"number"},{"type":"null"}]}}}`)}
		msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "echo",
			RawArguments: `{"path":"file.txt","offset":null,"nullable":null,"metadata":{"enabled":null},"ref":null}`})
		got, err := ai.ValidateToolCall(msg, 0, []ai.Tool{tool})
		ev.Check("accepted", err == nil, "err=%v", err)
		ev.Check("optional nulls dropped, nullable and $ref nulls kept", jsonEqual(got.Arguments(),
			[]byte(`{"path":"file.txt","nullable":null,"metadata":{},"ref":null}`)), "got %s", got.Arguments())
	})

	t.Run("repairable", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-repairable")
		tool := ai.Tool{Name: "echo", Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`)}
		msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "echo", RawArguments: "{\"text\":\"line1\nline2 C:\\dir\"}"})
		got, err := ai.ValidateToolCall(msg, 0, []ai.Tool{tool})
		ev.Check("accepted", err == nil, "err=%v", err)
		var args struct{ Text string }
		ev.Check("decoded", got.Decode(&args) == nil, "")
		ev.Check("repaired value", args.Text == "line1\nline2 C:\\dir", "got %q", args.Text)
	})

	t.Run("display-only-message", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-display-only")
		var call ai.ToolCall
		err := json.Unmarshal([]byte(`{"id":"call_1|fc_1","name":"get_weather","arguments":{"city": "Paris"}}`), &call)
		ev.Check("pi-shaped tool call decodes", err == nil, "err=%v", err)
		ev.Check("arguments kept in JSON.stringify form", call.Arguments == `{"city":"Paris"}`, "got %q", call.Arguments)
		got, err := ai.ValidateToolCall(toolMessage(ai.StopReasonToolUse, call), 0, []ai.Tool{weather})
		ev.Check("accepted", err == nil, "err=%v", err)
		ev.Check("identity", got.ID() == "call_1|fc_1" && got.Name() == "get_weather", "got %q %q", got.ID(), got.Name())
		ev.Check("arguments", string(got.Arguments()) == `{"city":"Paris"}`, "got %s", got.Arguments())
	})

	t.Run("inputs-untouched", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-inputs-untouched")
		tool := ai.Tool{Name: "echo", Parameters: json.RawMessage(`{"type":"object","properties":{"n":{"type":"number"},"o":{"type":"number"}}}`)}
		msg := toolMessage(ai.StopReasonToolUse, ai.ToolCall{ID: "call_1|fc_1", Name: "echo", Arguments: `{"n":"7","o":null}`, RawArguments: `{"n":"7","o":null}`})
		before, toolBefore := msg.Content[0], append([]byte(nil), tool.Parameters...)
		got, err := ai.ValidateToolCall(msg, 0, []ai.Tool{tool})
		ev.Check("accepted", err == nil, "err=%v", err)
		ev.Check("converted", string(got.Arguments()) == `{"n":7}`, "got %s", got.Arguments())
		ev.Check("message unchanged", reflect.DeepEqual(msg.Content[0], before), "got %+v", msg.Content[0])
		ev.Check("schema unchanged", string(tool.Parameters) == string(toolBefore), "got %s", tool.Parameters)
		a := got.Arguments()
		a[1] = 'X'
		ev.Check("Arguments returns a copy", string(got.Arguments()) == `{"n":7}`, "got %s", got.Arguments())
	})

	t.Run("repair", func(t *testing.T) {
		ev := run.Case(t, "E03-tooljson-repair")
		cases := []struct{ name, in, want string }{
			{"valid-unchanged", `{"a":"b\n\u00e9","c":[1,true,null]}`, `{"a":"b\n\u00e9","c":[1,true,null]}`},
			{"control-chars", "{\"a\":\"x\ty\nz\x01\"}", `{"a":"x\ty\nz\u0001"}`},
			{"invalid-escape", `{"path":"C:\dir\x"}`, `{"path":"C:\\dir\\x"}`},
			{"trailing-backslash", `{"a":"b\`, `{"a":"b\\`},
			{"short-unicode-kept", `{"a":"\u12"}`, `{"a":"\u12"}`},
			{"outside-strings", "{\n\t\"a\" :\t1 }", "{\n\t\"a\" :\t1 }"},
			{"truncated-stays", `{"a":"b`, `{"a":"b`},
		}
		for _, c := range cases {
			got := ai.RepairToolJSON(c.in)
			ev.Check(c.name, got == c.want, "RepairToolJSON(%q) = %q, want %q", c.in, got, c.want)
		}
	})
}

// toolMessage is a finished (or live, for StopReasonPending) assistant message
// holding blocks.
func toolMessage(stop ai.StopReason, blocks ...ai.AssistantContent) ai.AssistantMessage {
	return ai.AssistantMessage{Content: blocks, API: ai.APIOpenAIResponses, Provider: ai.ProviderOpenAI, Model: "gpt-4.1-mini", StopReason: stop}
}

// expectRefusal checks err is a *ai.ToolCallError of kind and returns it.
func expectRefusal(ev interface {
	Check(string, bool, string, ...any) bool
}, err error, kind ai.ToolCallErrorKind) *ai.ToolCallError {
	var e *ai.ToolCallError
	if !ev.Check("refused as "+string(kind), errors.As(err, &e) && e.Kind == kind, "err=%v", err) {
		return nil
	}
	ev.Check("errors.Is matches the kind", errors.Is(err, &ai.ToolCallError{Kind: kind}), "")
	return e
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
