package e2e

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Two valid PNG signatures with distinct trailing synthetic markers. The
// literals are independent expected wire values, not reconstructed from output.
func mixedTenantImage(k tenantKey) string {
	if k.tenant == tenantB.tenant {
		return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jBMsAAAAASUVORK5CYIJ0ZW5hbnQtYg=="
	}
	return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jBMsAAAAASUVORK5CYIJ0ZW5hbnQtYQ=="
}

func (r mixedRoute) tenantReply(t *testing.T, k tenantKey) provider.Reply {
	t.Helper()
	input := 31
	choice := "track"
	probabilities := map[string]float64{"refund": 0.1, "track": 0.9}
	if k.tenant == tenantB.tenant {
		input = 47
		choice = "refund"
		probabilities = map[string]float64{"refund": 0.9, "track": 0.1}
	}
	if r.op == ai.OperationChat {
		chunks := pressureText(t, 1)
		for i, ch := range chunks {
			ch = bytes.ReplaceAll(ch, []byte("tok "), []byte("mixed-output-"+k.tenant))
			ch = bytes.ReplaceAll(ch, []byte(`"input_tokens":1000`), []byte(`"input_tokens":`+strconv.Itoa(input)))
			chunks[i] = bytes.ReplaceAll(ch, []byte(`"total_tokens":1001`), []byte(`"total_tokens":`+strconv.Itoa(input+1)))
		}
		return pressureReply(chunks)
	}
	var body any
	if r.op == ai.OperationClassifier {
		body = map[string]any{"model": "synthetic-response", "answers": map[string]any{"intent": map[string]any{"type": "choice", "choice": choice, "probabilities": probabilities, "confidence": 0.82}}, "usage": map[string]int{"input_tokens": input, "output_tokens": 20}}
	} else if r.provider == ai.ProviderGoogle {
		body = map[string]any{"id": "synthetic-image", "status": "completed", "steps": []any{map[string]any{"type": "model_output", "content": []any{map[string]any{"type": "image", "mime_type": "image/png", "data": mixedTenantImage(k)}}}}, "usage": map[string]int{"total_input_tokens": input, "total_output_tokens": 20, "total_tokens": input + 20}}
	} else {
		body = map[string]any{"data": []any{map[string]any{"b64_json": mixedTenantImage(k)}}, "usage": map[string]any{"input_tokens": input, "output_tokens": 20, "total_tokens": input + 20, "input_tokens_details": map[string]int{"text_tokens": input, "image_tokens": 0}, "output_tokens_details": map[string]int{"text_tokens": 0, "image_tokens": 20}}}
	}
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{mustMarshal(t, body)}}
}
