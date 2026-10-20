package ai

import "slices"

// GLM's endpoint capability is independent of catalog compatibility. Validate
// the actual post-sampling/post-hook body even when no callback was installed.
// Authorization runs first, preserving tenant-denied for vendor references.
func validateGLMResponsesControls(m Model, body map[string]any) string {
	if body["store"] != false || body["stream"] != true {
		return "GLM Responses requires stream:true and stateless store:false"
	}
	if choice, present := body["tool_choice"]; present && choice != "none" && choice != "auto" {
		return "GLM Responses supports only default, none or auto tool choice"
	}
	if tools, present := body["tools"]; present {
		list, ok := tools.([]any)
		if !ok {
			return "GLM Responses tools must be function declarations"
		}
		for _, raw := range list {
			tool, ok := raw.(map[string]any)
			if !ok || tool["type"] != "function" {
				return "GLM Responses serves only function tools"
			}
			if _, present := tool["strict"]; present {
				return "GLM Responses strict function tools are not integrated"
			}
			name, _ := tool["name"].(string)
			if name == "" {
				return "GLM function tools must name a function"
			}
			if parameters, present := tool["parameters"]; present {
				if _, ok := parameters.(map[string]any); !ok {
					return "GLM function parameters must be a JSON Schema object"
				}
			}
			for field := range tool {
				if !slices.Contains([]string{"type", "name", "description", "parameters"}, field) {
					return "GLM function tool field is not integrated: " + field
				}
			}
		}
	}
	allowed := []string{"model", "input", "stream", "store", "instructions", "temperature", "top_p", "max_output_tokens", "stop", "tools", "tool_choice", "reasoning", "text", "prompt_cache_key"}
	for field := range body {
		if !slices.Contains(allowed, field) {
			return "GLM Responses field is not integrated: " + field
		}
	}
	if !m.acceptsImages() && glmResponsesNonTextInput(body["input"]) {
		return "the selected GLM model accepts text input only"
	}
	if text, present := body["text"]; present {
		config, ok := text.(map[string]any)
		format, _ := config["format"].(map[string]any)
		if !ok || len(config) != 1 || len(format) != 1 || (format["type"] != "text" && format["type"] != "json_object") {
			return "GLM Responses supports only text or json_object output format"
		}
	}
	return validateGLMResponsesInput(body["input"])
}

func validateGLMResponsesInput(value any) string {
	input, ok := value.([]any)
	if !ok {
		return "GLM Responses requires complete input history"
	}
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			return "GLM Responses input must contain documented input items"
		}
		switch item["type"] {
		case "reasoning":
			content, ok := item["content"].(map[string]any)
			_, text := content["text"].(string)
			if len(item) != 2 || !ok || len(content) != 2 || content["type"] != "reasoning_text" || !text {
				return "GLM InputReasoning requires only a reasoning_text content object"
			}
		case "message":
			if len(item) != 3 || !slices.Contains([]any{"user", "assistant", "system", "developer"}, item["role"]) {
				return "GLM InputMessage requires only type, role and content"
			}
		case "function_call":
			if len(item) != 4 {
				return "GLM function calls require only type, name, call_id and arguments"
			}
		case "function_call_output":
			if len(item) != 3 {
				return "GLM function outputs require only type, call_id and output"
			}
		default:
			return "GLM Responses input item type is not integrated"
		}
	}
	return ""
}

// A final native callback is another input boundary. The public Request image
// placeholder conversion remains unchanged; native images may not bypass it.
func glmResponsesNonTextInput(value any) bool {
	switch value := value.(type) {
	case []any:
		for _, item := range value {
			if glmResponsesNonTextInput(item) {
				return true
			}
		}
	case map[string]any:
		if slices.Contains([]any{"input_image", "input_file", "input_audio", "audio", "video"}, value["type"]) {
			return true
		}
		for _, item := range value {
			if glmResponsesNonTextInput(item) {
				return true
			}
		}
	}
	return false
}
