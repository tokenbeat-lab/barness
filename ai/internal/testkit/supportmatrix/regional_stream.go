package supportmatrix

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3/packages/ssestream"
)

func regionalDecode(raw []byte, v any) error {
	if !json.Valid(raw) {
		return refuse("regional captured JSON invalid")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := uniqueFields(d, 0); err != nil {
		return refuse("regional captured JSON ambiguous")
	}
	return json.Unmarshal(raw, v)
}

// Recheck the captured protocol independently from the producer's PASS label.
// A canceled body may end inside its final frame; complete turns may not.
func verifyRegionalStream(r Report, s ScenarioResult, body string) error {
	_, kind, _ := strings.Cut(s.ID, "/")
	cancel := kind == "cancel"
	frames := 0
	terminal, finish := false, false
	var checkModel func(any) bool
	checkModel = func(v any) bool {
		switch v := v.(type) {
		case []any:
			for _, x := range v {
				if !checkModel(x) {
					return false
				}
			}
		case map[string]any:
			if model, ok := v["model"].(string); ok && model != "" && model != s.Model {
				return false
			}
			for _, x := range v {
				if !checkModel(x) {
					return false
				}
			}
		}
		return true
	}

	decoder := ssestream.NewDecoder(&http.Response{Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body))})
	defer decoder.Close()
	for decoder.Next() {
		data := strings.TrimSpace(string(decoder.Event().Data))
		if strings.HasPrefix(data, "[DONE]") {
			continue
		}
		var frame map[string]any
		if regionalDecode([]byte(data), &frame) != nil {
			if cancel && frames > 0 {
				break
			}
			return refuse("regional actual SSE frame invalid")
		}
		frames++
		if !checkModel(frame) {
			return refuse("regional actual response served another model")
		}
		switch r.API {
		case "openai-completions":
			for _, v := range regionalArray(frame["choices"]) {
				choice, _ := v.(map[string]any)
				if reason, ok := choice["finish_reason"].(string); ok && reason != "" {
					finish = true
				}
			}
		case "openai-responses":
			terminal = terminal || frame["type"] == "response.completed"
		case "anthropic-messages":
			terminal = terminal || frame["type"] == "message_stop"
		}
	}
	if r.API == "openai-completions" {
		terminal = finish
	}
	if frames == 0 || !cancel && (!terminal || decoder.Err() != nil) {
		return refuse("regional captured own stream lacks legal terminal")
	}
	return nil
}
