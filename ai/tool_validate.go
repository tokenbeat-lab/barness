package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ValidToolCall is a tool call whose message ended normally with tool use and
// whose arguments are complete, parse as a JSON object and, after pi-ai's
// schema conversion, satisfy the tool's parameter schema. Only ValidateToolCall
// creates one. Whether to execute it stays the host's decision.
type ValidToolCall struct {
	id, name string
	args     string
}

// ID is the call's ToolCall.ID, for the ToolResultMessage.
func (c ValidToolCall) ID() string { return c.id }

// Name is the called tool's name.
func (c ValidToolCall) Name() string { return c.name }

// Arguments returns a copy of the validated, converted arguments object.
func (c ValidToolCall) Arguments() json.RawMessage { return json.RawMessage(c.args) }

// Decode unmarshals the validated arguments into dst.
func (c ValidToolCall) Decode(dst any) error { return json.Unmarshal([]byte(c.args), dst) }

// ToolCallErrorKind says why ValidateToolCall refused a call.
type ToolCallErrorKind string

const (
	// ToolCallNotExecutable: the message did not end with StopReasonToolUse.
	// A call of a truncated (length), failed, aborted or still-running message
	// is never executable, however complete its arguments look.
	ToolCallNotExecutable ToolCallErrorKind = "not_executable"
	// ToolCallNotFound: the content index holds no tool call.
	ToolCallNotFound ToolCallErrorKind = "not_tool_call"
	// ToolCallUnknownTool: no declared tool has the call's name.
	ToolCallUnknownTool ToolCallErrorKind = "unknown_tool"
	// ToolCallInvalidSchema: the tool's Parameters is not a usable JSON Schema,
	// including one that references a remote or file resource.
	ToolCallInvalidSchema ToolCallErrorKind = "invalid_schema"
	// ToolCallMalformedJSON: the arguments are not a complete JSON object,
	// even after repair.
	ToolCallMalformedJSON ToolCallErrorKind = "malformed_json"
	// ToolCallInvalidArguments: the arguments do not satisfy the schema.
	ToolCallInvalidArguments ToolCallErrorKind = "invalid_arguments"
)

// ToolCallError is ValidateToolCall's refusal.
type ToolCallError struct {
	Kind     ToolCallErrorKind
	ToolName string
	// Message describes the refusal; for invalid_arguments it follows pi-ai's
	// "Validation failed for tool …" layout (the per-problem wording is the
	// schema validator's, not TypeBox's), suitable for an isError result.
	Message string
	// Problems lists each schema violation of invalid_arguments.
	Problems []ToolArgumentProblem
}

// ToolArgumentProblem is one schema violation. Path is pi-ai's dotted
// instance path ("root" for the arguments object itself).
type ToolArgumentProblem struct {
	Path    string
	Message string
}

func (e *ToolCallError) Error() string { return "ai: tool call " + string(e.Kind) + ": " + e.Message }

// Is matches another *ToolCallError by Kind.
func (e *ToolCallError) Is(target error) bool {
	t, ok := target.(*ToolCallError)
	return ok && t.Kind == e.Kind
}

// ValidateToolCall decides whether the tool call at msg.Content[contentIndex]
// may be handed to the host for execution. It never executes anything.
//
// The call must belong to a message that ended with StopReasonToolUse: a
// truncated (length), failed, aborted or still-running message yields
// not_executable, however complete its arguments look. The arguments are then
// read from RawArguments (or, for a message stored without it, Arguments) and
// must be a complete JSON object; as in pi-ai, raw control characters and
// invalid escapes inside strings are repaired, nothing else is. Finally pi's
// schema conversion runs (optional nulls dropped, primitives converted to the
// declared type) and the result must satisfy the tool's JSON Schema. Schemas
// can only reference themselves: a remote or file $ref is never loaded.
func ValidateToolCall(msg AssistantMessage, contentIndex int, tools []Tool) (ValidToolCall, error) {
	if msg.StopReason != StopReasonToolUse {
		return ValidToolCall{}, &ToolCallError{Kind: ToolCallNotExecutable,
			Message: fmt.Sprintf("message ended with stop reason %q; only calls of a message that ended with %q can be executed", msg.StopReason, StopReasonToolUse)}
	}
	var call ToolCall
	found := false
	if contentIndex >= 0 && contentIndex < len(msg.Content) {
		call, found = msg.Content[contentIndex].(ToolCall)
	}
	if !found {
		return ValidToolCall{}, &ToolCallError{Kind: ToolCallNotFound, Message: fmt.Sprintf("content index %d holds no tool call", contentIndex)}
	}
	refuse := func(kind ToolCallErrorKind, format string, args ...any) (ValidToolCall, error) {
		return ValidToolCall{}, &ToolCallError{Kind: kind, ToolName: call.Name, Message: fmt.Sprintf(format, args...)}
	}
	i := slices.IndexFunc(tools, func(t Tool) bool { return t.Name == call.Name })
	if i < 0 {
		return refuse(ToolCallUnknownTool, "Tool %q not found", call.Name)
	}
	schema, err := compileToolSchema(tools[i].Parameters)
	if err != nil {
		return refuse(ToolCallInvalidSchema, "tool %q parameters are not a usable JSON Schema: %v", call.Name, err)
	}
	v, err := parseJSONWithRepair(cmp.Or(call.RawArguments, string(call.Arguments), "{}"))
	if err != nil {
		return refuse(ToolCallMalformedJSON, "tool %q arguments are not complete JSON", call.Name)
	}
	args, ok := v.(*jsonObject)
	if !ok {
		return refuse(ToolCallMalformedJSON, "tool %q arguments are not a JSON object", call.Name)
	}
	received := stringifyJSON(args)

	schema.normalizeOptionalNulls(args, schema.doc)
	converted := schema.coerceWithSchema(args, schema.doc)
	problems := schema.problems(converted)
	if _, ok := converted.(*jsonObject); !ok && len(problems) == 0 {
		problems = []ToolArgumentProblem{{Path: "root", Message: "arguments must be an object"}}
	}
	if len(problems) > 0 {
		return ValidToolCall{}, &ToolCallError{Kind: ToolCallInvalidArguments, ToolName: call.Name,
			Message: validationMessage(call.Name, problems, received), Problems: problems}
	}
	return ValidToolCall{id: call.ID, name: call.Name, args: stringifyJSON(converted)}, nil
}

// validationMessage is pi-ai's "Validation failed for tool …" text.
func validationMessage(tool string, problems []ToolArgumentProblem, received string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Validation failed for tool %q:\n", tool)
	for i, p := range problems {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "  - %s: %s", p.Path, p.Message)
	}
	var indented bytes.Buffer
	if json.Indent(&indented, []byte(received), "", "  ") != nil {
		indented.WriteString(received)
	}
	b.WriteString("\n\nReceived arguments:\n")
	b.Write(indented.Bytes())
	return b.String()
}

// toolSchema is one tool's compiled parameter schema plus pi's standalone
// sub-schema validators, which the conversion consults.
type toolSchema struct {
	doc  any
	root *jsonschema.Schema
	subs map[string]*jsonschema.Schema // by sub-schema JSON; nil when it does not compile alone
}

// schemaURL is where a tool schema lives while compiled. Its scheme has no
// loader, so a reference outside the document fails instead of being fetched.
const schemaURL = "barness-tool:///parameters.json"

func newSchemaCompiler() *jsonschema.Compiler {
	c := jsonschema.NewCompiler()
	c.UseLoader(jsonschema.SchemeURLLoader{})
	return c
}

func compileToolSchema(parameters json.RawMessage) (*toolSchema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(parameters))
	if err != nil {
		return nil, err
	}
	c := newSchemaCompiler()
	if err := c.AddResource(schemaURL, doc); err != nil {
		return nil, err
	}
	root, err := c.Compile(schemaURL)
	if err != nil {
		return nil, err
	}
	return &toolSchema{doc: doc, root: root, subs: map[string]*jsonschema.Schema{}}, nil
}

// acceptance is pi's `validator?.Check(value)` for a sub-schema compiled on
// its own: unknown when it cannot be (for example, a $ref into its parent).
type acceptance int

const (
	schemaUnknown acceptance = iota
	schemaAccepts
	schemaRejects
)

func (s *toolSchema) accepts(sub any, v any) acceptance {
	key, err := json.Marshal(sub)
	if err != nil {
		return schemaUnknown
	}
	compiled, seen := s.subs[string(key)]
	if !seen {
		c := newSchemaCompiler()
		if c.AddResource(schemaURL, sub) == nil {
			compiled, _ = c.Compile(schemaURL)
		}
		s.subs[string(key)] = compiled
	}
	switch {
	case compiled == nil:
		return schemaUnknown
	case compiled.Validate(plainJSON(v)) == nil:
		return schemaAccepts
	}
	return schemaRejects
}

// problems validates v against the root schema and lists each violation with
// pi's dotted path; a missing required property is reported at its own path.
func (s *toolSchema) problems(v any) []ToolArgumentProblem {
	var verr *jsonschema.ValidationError
	if err := s.root.Validate(plainJSON(v)); !errors.As(err, &verr) {
		return nil
	}
	printer := message.NewPrinter(language.English)
	var out []ToolArgumentProblem
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) > 0 {
			for _, c := range e.Causes {
				walk(c)
			}
			return
		}
		msg := e.ErrorKind.LocalizedString(printer)
		if req, ok := e.ErrorKind.(*kind.Required); ok {
			for _, missing := range req.Missing {
				out = append(out, ToolArgumentProblem{Path: dottedPath(append(slices.Clone(e.InstanceLocation), missing)), Message: msg})
			}
			return
		}
		out = append(out, ToolArgumentProblem{Path: dottedPath(e.InstanceLocation), Message: msg})
	}
	walk(verr)
	return out
}

func dottedPath(location []string) string {
	if len(location) == 0 {
		return "root"
	}
	return strings.Join(location, ".")
}

// plainJSON converts a JSON value to the generic form the schema validator
// reads. Object order does not matter to validation.
func plainJSON(v any) any {
	switch v := v.(type) {
	case []any:
		out := make([]any, len(v))
		for i, e := range v {
			out[i] = plainJSON(e)
		}
		return out
	case *jsonObject:
		out := make(map[string]any, len(v.keys))
		for _, k := range v.keys {
			out[k] = plainJSON(v.values[k])
		}
		return out
	}
	return v
}
