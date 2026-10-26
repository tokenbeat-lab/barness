package supportmatrix

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyXAIRequest(request map[string]json.RawMessage) error {
	allowed := map[string]bool{}
	for _, name := range []string{"model", "stream", "store", "include", "input", "max_output_tokens", "reasoning", "tool_choice", "tools", "parallel_tool_calls", "prompt_cache_key"} {
		allowed[name] = true
	}
	for key := range request {
		if !allowed[key] {
			return refuse("xAI unsupported final request control")
		}
	}
	var model string
	var stream, store bool
	var include []string
	if json.Unmarshal(request["model"], &model) != nil || model != xaiModel || json.Unmarshal(request["stream"], &stream) != nil || !stream || string(bytes.TrimSpace(request["stream"])) != "true" || json.Unmarshal(request["store"], &store) != nil || store || string(bytes.TrimSpace(request["store"])) != "false" || json.Unmarshal(request["include"], &include) != nil || len(include) != 1 || include[0] != "reasoning.encrypted_content" {
		return refuse("xAI stateless request controls differ")
	}
	maximum, err := xaiCount(request["max_output_tokens"])
	if err != nil || maximum != 1024 {
		return refuse("xAI visible output reservation differs")
	}
	if r := request["reasoning"]; len(r) > 0 {
		var effort map[string]json.RawMessage
		if regionalDecode(r, &effort) != nil || len(effort) != 1 {
			return refuse("xAI reasoning controls differ")
		}
		var value string
		json.Unmarshal(effort["effort"], &value)
		if value != "low" && value != "medium" && value != "high" && value != "xhigh" {
			return refuse("xAI reasoning effort invalid")
		}
	}
	if v := request["parallel_tool_calls"]; len(v) > 0 {
		var flag bool
		if json.Unmarshal(v, &flag) != nil || string(bytes.TrimSpace(v)) != "true" && string(bytes.TrimSpace(v)) != "false" {
			return refuse("xAI parallel flag invalid")
		}
	}
	if v := request["tool_choice"]; len(v) > 0 {
		var mode string
		if json.Unmarshal(v, &mode) != nil || mode != "auto" && mode != "required" && mode != "none" {
			return refuse("xAI tool choice invalid")
		}
	}
	if v := request["prompt_cache_key"]; len(v) > 0 && string(v) != "null" {
		var key string
		if json.Unmarshal(v, &key) != nil || !xaiHash(key) {
			return refuse("xAI derived cache hint invalid")
		}
	}
	if tools, ok := request["tools"]; ok {
		if _, err := xaiFunctionTools(tools); err != nil {
			return err
		}
	}
	var input []map[string]json.RawMessage
	if regionalDecode(request["input"], &input) != nil || len(input) == 0 {
		return refuse("xAI full input history missing")
	}
	for _, item := range input {
		if err := verifyXAIInputItem(item); err != nil {
			return err
		}
	}
	return nil
}
func verifyXAIScenario(root *os.Root, s ScenarioResult, calls []xaiCall) error {
	first := calls[0]
	if strings.HasPrefix(s.ID, "entry-") || strings.HasPrefix(s.ID, "effort-") {
		visible := ""
		for _, b := range first.message.Content {
			if text, ok := b.(ai.Text); ok {
				visible += text.Text
			}
		}
		if strings.TrimSpace(visible) == "" {
			return refuse("xAI text entry has no visible answer")
		}
		effort := "low"
		if strings.HasPrefix(s.ID, "effort-") {
			effort = strings.TrimPrefix(s.ID, "effort-")
			if effort == "default" {
				effort = ""
			}
		}
		var r map[string]string
		if v := first.request["reasoning"]; len(v) > 0 {
			json.Unmarshal(v, &r)
		}
		if r["effort"] != effort {
			return refuse("xAI omitted/explicit effort differs")
		}
		if s.ID == "effort-high" {
			v, _ := first.message.Usage.Reasoning.Get()
			if v <= 0 {
				return refuse("xAI own nonzero reasoning proof absent")
			}
		}
	}
	if s.ID == "encrypted-continuation" {
		encrypted := false
		for _, item := range first.output {
			var kind string
			json.Unmarshal(item["type"], &kind)
			encrypted = encrypted || kind == "reasoning"
		}
		if !encrypted || len(calls) != 2 || !xaiNativeReplay(first, calls[1]) {
			return refuse("xAI encrypted whole native continuation differs")
		}
		if err := verifyXAIPersistence(root, s, first); err != nil {
			return err
		}
	}
	if strings.HasPrefix(s.ID, "tools-") || s.ID == "tool-result-image" {
		mode := strings.TrimPrefix(s.ID, "tools-")
		if s.ID == "tool-result-image" {
			mode = "required"
		}
		choice := mode
		if mode == "parallel" {
			choice = "required"
		}
		var actual string
		json.Unmarshal(first.request["tool_choice"], &actual)
		if actual != choice {
			return refuse("xAI fixed tool choice differs")
		}
		want := 1
		if mode == "none" {
			want = 0
		}
		if mode == "parallel" {
			want = 2
			var parallel bool
			if json.Unmarshal(first.request["parallel_tool_calls"], &parallel) != nil || !parallel {
				return refuse("xAI parallel tool proof absent")
			}
		}
		definitions, err := xaiFunctionTools(first.request["tools"])
		if err != nil || len(definitions) != 1 {
			return refuse("xAI Host tool definition missing")
		}
		count := 0
		for i, b := range first.message.Content {
			if _, ok := b.(ai.ToolCall); ok {
				if _, err := ai.ValidateToolCall(first.message, i, definitions); err != nil {
					return refuse("xAI tool was not executable by Host")
				}
				count++
			}
		}
		if count != want {
			return refuse("xAI tool policy not fulfilled")
		}
		if want > 0 {
			if len(calls) != 2 || xaiWireString(calls[1].request["tool_choice"]) != "none" {
				return refuse("xAI tool result turn must disable further calls")
			}
			if len(calls) != 2 || !xaiNativeReplay(first, calls[1]) || !xaiToolReplay(first, calls[1], s.ID == "tool-result-image") {
				return refuse("xAI paired tool result or image order differs")
			}
		}
	}
	if s.ID == "user-png" || s.ID == "user-jpeg" {
		var input []map[string]json.RawMessage
		json.Unmarshal(first.request["input"], &input)
		var images []string
		for _, item := range input {
			v, err := xaiImages(item["content"])
			if err != nil {
				return err
			}
			images = append(images, v...)
		}
		mime := "image/png"
		if s.ID == "user-jpeg" {
			mime = "image/jpeg"
		}
		if len(images) != 1 || !strings.HasPrefix(images[0], "data:"+mime+";base64,") {
			return refuse("xAI own PNG/JPEG user proof absent")
		}
	}
	if s.ID == "cache-usage" {
		var a, b string
		json.Unmarshal(first.request["prompt_cache_key"], &a)
		json.Unmarshal(calls[1].request["prompt_cache_key"], &b)
		if !xaiHash(a) || a != b {
			return refuse("xAI scoped stable cache hint proof absent")
		}
	}
	return nil
}
func xaiToolReplay(first, next xaiCall, images bool) bool {
	var input []map[string]json.RawMessage
	if regionalDecode(next.request["input"], &input) != nil {
		return false
	}
	var outputs []map[string]json.RawMessage
	for _, i := range first.output {
		var kind string
		json.Unmarshal(i["type"], &kind)
		if kind == "function_call" {
			outputs = append(outputs, i)
		}
	}
	matched := 0
	for ordinal, tool := range outputs {
		var id string
		json.Unmarshal(tool["call_id"], &id)
		callIndex, resultIndex := -1, -1
		for n, in := range input {
			var kind, call string
			json.Unmarshal(in["type"], &kind)
			json.Unmarshal(in["call_id"], &call)
			if kind == "function_call" && call == id {
				if callIndex != -1 || !xaiFunctionProjection(tool, in) {
					return false
				}
				callIndex = n
			}
			if kind == "function_call_output" && call == id {
				if resultIndex != -1 {
					return false
				}
				resultIndex = n
				var text string
				marker := "synthetic-host-result-" + string(rune('a'+ordinal))
				if images {
					marker += "\n(see attached image)"
				}
				if json.Unmarshal(in["output"], &text) != nil || text != marker {
					return false
				}
			}
		}
		if callIndex < 0 || resultIndex <= callIndex {
			return false
		}
		if images {
			if resultIndex+1 >= len(input) {
				return false
			}
			actual, err := xaiImages(input[resultIndex+1]["content"])
			if err != nil || len(actual) != 2 || !strings.HasPrefix(actual[0], "data:image/png;base64,") || !strings.HasPrefix(actual[1], "data:image/jpeg;base64,") {
				return false
			}
		}
		matched++
	}
	return matched > 0
}
func verifyXAIPersistence(root *os.Root, s ScenarioResult, first xaiCall) error {
	var stored struct {
		Metadata regionalMessage        `json:"metadata"`
		Envelope ai.NativeStateEnvelope `json:"envelope"`
		Blocks   []struct {
			Kind string          `json:"kind"`
			Data json.RawMessage `json:"data"`
		} `json:"blocks"`
	}
	if xaiReadJSON(root, s.CaseID+"/persisted-turn.json", &stored) != nil || stored.Envelope.TenantID != first.result.Metadata.TenantID || stored.Envelope.AccountScopeID != first.result.Metadata.AccountScopeID || stored.Envelope.ProviderID != ai.ProviderXAI || stored.Envelope.API != ai.APIOpenAIResponses || stored.Envelope.ModelID != xaiModel {
		return refuse("xAI persisted Host envelope differs")
	}
	expectedMetadata := first.result.Message
	expectedMetadata.Content = nil
	if !xaiEqual(stored.Metadata, expectedMetadata) {
		return refuse("xAI persisted message metadata changed")
	}
	if len(stored.Blocks) != len(first.message.Content) {
		return refuse("xAI persisted blocks missing")
	}
	for i, b := range stored.Blocks {
		wantKind := ""
		switch first.message.Content[i].(type) {
		case ai.Text:
			wantKind = "text"
		case ai.Thinking:
			wantKind = "thinking"
		case ai.ToolCall:
			wantKind = "tool"
		}
		if b.Kind != wantKind {
			return refuse("xAI persisted block kind changed")
		}
		raw, _ := json.Marshal(first.message.Content[i])
		if !xaiJSONEqual(raw, b.Data) {
			return refuse("xAI persisted complete block changed")
		}
	}
	return nil
}
func xaiImages(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || raw[0] == '"' {
		return nil, nil
	}
	var blocks []map[string]json.RawMessage
	if regionalDecode(raw, &blocks) != nil {
		return nil, refuse("xAI image content malformed")
	}
	var images []string
	for _, b := range blocks {
		var kind string
		json.Unmarshal(b["type"], &kind)
		if kind != "input_image" {
			continue
		}
		var value string
		if json.Unmarshal(b["image_url"], &value) != nil {
			return nil, refuse("xAI inline image missing")
		}
		prefix, data, ok := strings.Cut(value, ",")
		if !ok || prefix != "data:image/png;base64" && prefix != "data:image/jpeg;base64" {
			return nil, refuse("xAI inline image format differs")
		}
		decoded, err := base64.StdEncoding.Strict().DecodeString(data)
		if err != nil || len(decoded) > 4096 || strings.ContainsAny(data, " \r\n\t") {
			return nil, refuse("xAI bounded image bytes invalid")
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(decoded))
		if err != nil || config.Width != 32 || config.Height != 32 || prefix != "data:image/"+format+";base64" {
			return nil, refuse("xAI valid PNG/JPEG input absent")
		}
		images = append(images, value)
	}
	return images, nil
}

func xaiFunctionProjection(output, input map[string]json.RawMessage) bool {
	if len(input) != 5 {
		return false
	}
	for _, k := range []string{"type", "id", "call_id", "name", "arguments"} {
		if !xaiJSONEqual(output[k], input[k]) {
			return false
		}
	}
	return true
}
