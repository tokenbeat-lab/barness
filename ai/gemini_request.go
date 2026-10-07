package ai

import (
	"cmp"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// Gemini Developer API wire DTOs: the REST body of
// models/{model}:streamGenerateContent. The adapter marshals them itself, so
// the exact request — field presence included — is decided here. Shapes
// follow what pi-ai 1.0.0 sends through @google/genai 2.21.0: pi's
// google-generative-ai buildParams and google-shared convertMessages and
// convertTools build the SDK's parameters, and the SDK's
// generateContentParametersToMldev turns them into this body (system
// instruction, tools and tool config move to the top level, the rest of the
// config becomes generationConfig, null config fields are dropped).
type geminiBody struct {
	Contents          []geminiContent   `json:"contents"`
	SystemInstruction *geminiContent    `json:"systemInstruction,omitempty"`
	Tools             []geminiTool      `json:"tools,omitempty"`
	ToolConfig        *geminiToolConfig `json:"toolConfig,omitempty"`
	// GenerationConfig is always sent, empty included: pi always passes a
	// config object, which the SDK always converts.
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role"`
}

// geminiPart is one part; exactly one of its data fields is set.
type geminiPart struct {
	// Text is a pointer because an empty text part is sent as "".
	Text             *string                 `json:"text,omitempty"`
	Thought          bool                    `json:"thought,omitempty"`
	ThoughtSignature string                  `json:"thoughtSignature,omitempty"`
	InlineData       *geminiBlob             `json:"inlineData,omitempty"`
	FunctionCall     *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

func textPart(text string) geminiPart { return geminiPart{Text: &text} }

type geminiBlob struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
	ID   string          `json:"id,omitempty"`
}

type geminiFunctionResponse struct {
	Name string `json:"name"`
	// Response is {"output": text} or, for a failed tool, {"error": text}.
	Response map[string]string `json:"response"`
	// Parts are images nested in the response (Gemini 3 and later).
	Parts []geminiImagePart `json:"parts,omitempty"`
	ID    string            `json:"id,omitempty"`
}

type geminiImagePart struct {
	InlineData geminiBlob `json:"inlineData"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// ParametersJSONSchema is the tool's JSON Schema, sent as is.
	ParametersJSONSchema json.RawMessage `json:"parametersJsonSchema"`
}

type geminiToolConfig struct {
	FunctionCallingConfig struct {
		Mode string `json:"mode"`
	} `json:"functionCallingConfig"`
}

type geminiGenerationConfig struct {
	Temperature     *float64              `json:"temperature,omitempty"`
	MaxOutputTokens *int                  `json:"maxOutputTokens,omitempty"`
	ThinkingConfig  *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

// geminiThinkingConfig is pi's thinkingConfig, which the SDK copies as is: a
// null budget is sent as null.
type geminiThinkingConfig struct {
	IncludeThoughts bool                `json:"includeThoughts,omitempty"`
	ThinkingLevel   GeminiThinkingLevel `json:"thinkingLevel,omitempty"`
	ThinkingBudget  Nullable[int]       `json:"thinkingBudget,omitzero"`
	// invalid is pi's refusal of the model's level map, met while building.
	invalid string
}

// msgGeminiNoContents is the Google SDK's refusal of a request without
// contents, which pi reports before sending anything.
const msgGeminiNoContents = "contents are required"

// buildGeminiBody encodes the prepared history and options for model, as
// pi's buildParams and the Google SDK do. Its failures are the request's own
// (invalid_request): pi meets them before anything is sent.
func buildGeminiBody(model Model, history transcript, opts GeminiOptions) ([]byte, *Error) {
	var body geminiBody
	msgs := history.messages
	if len(msgs) > 0 {
		if sm, ok := msgs[0].(SystemMessage); ok {
			if text := sm.promptText(); text != "" {
				body.SystemInstruction = &geminiContent{Parts: []geminiPart{textPart(text)}, Role: "user"}
			}
			msgs = msgs[1:]
		}
	}
	contents, err := geminiContents(model, msgs)
	if err != nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if len(contents) == 0 {
		return nil, newError(CodeInvalidRequest, PhaseRequest, msgGeminiNoContents)
	}
	body.Contents = contents

	if len(history.tools) > 0 {
		decls := make([]geminiFunctionDeclaration, len(history.tools))
		for i, t := range history.tools {
			decls[i] = geminiFunctionDeclaration{Name: t.Name, Description: t.Description, ParametersJSONSchema: t.Parameters}
		}
		body.Tools = []geminiTool{{FunctionDeclarations: decls}}
		if mode := geminiFunctionCallingMode(opts.ToolChoice); mode != "" {
			body.ToolConfig = &geminiToolConfig{}
			body.ToolConfig.FunctionCallingConfig.Mode = mode
		}
	}

	gc := &body.GenerationConfig
	if t, ok := opts.Temperature.Get(); ok {
		gc.Temperature = &t
	}
	if n, ok := opts.MaxTokens.Get(); ok {
		gc.MaxOutputTokens = &n
	}
	if th, ok := opts.Thinking.Get(); ok && model.Reasoning {
		if th.Enabled {
			tc := geminiThinkingConfig{IncludeThoughts: true}
			switch {
			case !th.Level.IsZero():
				tc.ThinkingLevel, _ = th.Level.Get()
			case !th.BudgetTokens.IsZero():
				tc.ThinkingBudget = th.BudgetTokens
			}
			gc.ThinkingConfig = &tc
		} else {
			tc := geminiDisabledThinking(model)
			if tc.invalid != "" {
				return nil, newError(CodeInvalidRequest, PhaseRequest, tc.invalid)
			}
			gc.ThinkingConfig = &tc
		}
	}
	out, err := marshalJS(body)
	if err != nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	return out, nil
}

// geminiFunctionCallingMode ports pi's resolveGoogleFunctionCallingMode for
// tools without constrained sampling, which barness cannot declare: a tool
// choice maps to its mode; without one no tool config is sent ("").
func geminiFunctionCallingMode(choice Nullable[GeminiToolChoice]) string {
	c, ok := choice.Get()
	if !ok || c == "" {
		return ""
	}
	switch c {
	case GeminiToolChoiceNone:
		return "NONE"
	case GeminiToolChoiceAny:
		return "ANY"
	}
	return "AUTO"
}

// geminiContents is pi's convertMessages for the conversation after the
// leading system message: user parts as they are, a replayed turn as model
// parts, and tool results as function responses, consecutive ones in one
// user turn.
func geminiContents(model Model, msgs []Message) ([]geminiContent, error) {
	contents := []geminiContent{}
	withIDs := geminiRequiresToolCallID(model.ID)
	nested := geminiNestsResponseImages(model.ID)
	for _, msg := range msgs {
		switch m := msg.(type) {
		case UserMessage:
			var parts []geminiPart
			for _, c := range m.Content {
				switch c := c.(type) {
				case Text:
					parts = append(parts, textPart(c.Text))
				case Image:
					parts = append(parts, geminiPart{InlineData: &geminiBlob{MimeType: c.MimeType, Data: c.Data}})
				}
			}
			if len(parts) > 0 {
				contents = append(contents, geminiContent{Parts: parts, Role: "user"})
			}
		case replayedAssistant:
			parts, err := geminiModelParts(m, withIDs)
			if err != nil {
				return nil, err
			}
			if len(parts) > 0 {
				contents = append(contents, geminiContent{Parts: parts, Role: "model"})
			}
		case ToolResultMessage:
			contents = appendGeminiToolResult(contents, model, m, withIDs, nested)
		}
	}
	return contents, nil
}

// geminiModelParts replays a previous turn prepared by prepareTranscript.
// Only the same model's state reaches here as thinking or signatures; a
// signature is sent only when it is well-formed base64. Blank text and
// thinking are skipped unless they carry a signature, which Gemini needs
// echoed back.
func geminiModelParts(m replayedAssistant, withIDs bool) ([]geminiPart, error) {
	var parts []geminiPart
	for _, block := range m.Content {
		switch b := block.(type) {
		case Text:
			sig := geminiReplaySignature(m.sameModel, b.Signature)
			if isBlank(b.Text) && sig == "" {
				continue
			}
			p := textPart(b.Text)
			p.ThoughtSignature = sig
			parts = append(parts, p)
		case Thinking:
			if !m.sameModel {
				// The cross-model rules already turned visible thinking into
				// text; nothing else of another model's thinking is sent.
				continue
			}
			sig := geminiReplaySignature(true, b.Signature)
			if isBlank(b.Thinking) && sig == "" {
				continue
			}
			p := textPart(b.Thinking)
			p.Thought, p.ThoughtSignature = true, sig
			parts = append(parts, p)
		case ToolCall:
			// Arguments that are not JSON (only a host-built call can have
			// them) make the request unencodable: invalid_request, never a
			// silent {}.
			args, err := parseJSON(string(cmp.Or(b.Arguments, "{}")))
			if err != nil {
				return nil, err
			}
			call := &geminiFunctionCall{Name: b.Name, Args: json.RawMessage(stringifyJSON(args))}
			if withIDs {
				call.ID = b.ID
			}
			sig, _ := b.ThoughtSignature.Get()
			parts = append(parts, geminiPart{FunctionCall: call, ThoughtSignature: geminiReplaySignature(m.sameModel, sig)})
		}
	}
	return parts, nil
}

// appendGeminiToolResult is pi's tool result conversion: the text joined by
// newlines, or "(see attached image)" for images alone, as the output (or
// error) of a function response. Images go inside the response for Gemini 3
// and later and other models, and in a separate user turn for older Gemini.
// Function responses join a preceding user turn of function responses.
func appendGeminiToolResult(contents []geminiContent, model Model, m ToolResultMessage, withIDs, nested bool) []geminiContent {
	var texts []string
	var images []geminiImagePart
	for _, c := range m.Content {
		switch c := c.(type) {
		case Text:
			texts = append(texts, c.Text)
		case Image:
			if model.acceptsImages() {
				images = append(images, geminiImagePart{InlineData: geminiBlob{MimeType: c.MimeType, Data: c.Data}})
			}
		}
	}
	value := strings.Join(texts, "\n")
	if value == "" && len(images) > 0 {
		value = "(see attached image)"
	}
	key := "output"
	if m.IsError {
		key = "error"
	}
	resp := &geminiFunctionResponse{Name: m.ToolName, Response: map[string]string{key: value}}
	if len(images) > 0 && nested {
		resp.Parts = images
	}
	if withIDs {
		resp.ID = m.ToolCallID
	}
	part := geminiPart{FunctionResponse: resp}
	if n := len(contents); n > 0 && contents[n-1].Role == "user" && hasFunctionResponse(contents[n-1].Parts) {
		contents[n-1].Parts = append(contents[n-1].Parts, part)
	} else {
		contents = append(contents, geminiContent{Parts: []geminiPart{part}, Role: "user"})
	}
	if len(images) > 0 && !nested {
		parts := []geminiPart{textPart("Tool result image:")}
		for _, img := range images {
			parts = append(parts, geminiPart{InlineData: &img.InlineData})
		}
		contents = append(contents, geminiContent{Parts: parts, Role: "user"})
	}
	return contents
}

func hasFunctionResponse(parts []geminiPart) bool {
	for _, p := range parts {
		if p.FunctionResponse != nil {
			return true
		}
	}
	return false
}

// geminiSignature is the form Google accepts for a thought signature:
// base64 (the API's bytes type), its length a multiple of four.
var geminiSignature = regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)

// geminiReplaySignature is pi's resolveThoughtSignature: a signature is sent
// back only to the same model, and only when well-formed.
func geminiReplaySignature(sameModel bool, sig string) string {
	if !sameModel || sig == "" || len(sig)%4 != 0 || !geminiSignature.MatchString(sig) {
		return ""
	}
	return sig
}

// geminiMajorVersion is the major version of a Gemini model id, if it is one
// (pi's getGeminiMajorVersion).
var geminiMajorVersion = regexp.MustCompile(`^gemini(?:-live)?-(\d+)`)

func geminiMajor(id string) (int, bool) {
	m := geminiMajorVersion.FindStringSubmatch(strings.ToLower(id))
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

// geminiRequiresToolCallID ports pi's requiresToolCallId: models that need
// the call id echoed in function calls and responses.
func geminiRequiresToolCallID(id string) bool {
	major, gemini := geminiMajor(id)
	return strings.HasPrefix(id, "claude-") || strings.HasPrefix(id, "gpt-oss-") || (gemini && major >= 3)
}

// geminiNestsResponseImages ports pi's supportsMultimodalFunctionResponse:
// Gemini 3 and later, and every model that is not Gemini.
func geminiNestsResponseImages(id string) bool {
	major, gemini := geminiMajor(id)
	return !gemini || major >= 3
}
