//go:build live

package live

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func controlledDecisionsRecorder(t *testing.T, name string) *recorder {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model     string `json:"model"`
			Questions []struct {
				Type    string `json:"type"`
				Name    string `json:"name"`
				Choices []struct {
					Value string `json:"value"`
				} `json:"choices"`
				Levels []struct {
					Label string `json:"label"`
				} `json:"levels"`
			} `json:"questions"`
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("controlled Provider received malformed request")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req-controlled-decisions")
		switch name {
		case "429":
			w.WriteHeader(429)
			fmt.Fprint(w, `{"error":{"message":"synthetic rate limit"}}`)
			return
		case "504":
			w.WriteHeader(504)
			fmt.Fprint(w, `{"error":{"message":"synthetic upstream timeout"}}`)
			return
		case "malformed-2xx":
			fmt.Fprint(w, `{"answers":[]`)
			return
		}
		answers := []any{}
		for _, q := range req.Questions {
			if name == "refusal" || (name == "mixed-refusal" && q.Name == "severity") {
				answers = append(answers, map[string]any{"type": "refusal", "name": q.Name})
				continue
			}
			switch q.Type {
			case "choice":
				probs := []any{}
				for i, c := range q.Choices {
					p := 0.0
					if i == 0 {
						p = 1
					}
					probs = append(probs, map[string]any{"value": c.Value, "probability": p})
				}
				if name == "precision-2xx" {
					probs[0].(map[string]any)["probability"] = 0.99
				}
				answers = append(answers, map[string]any{"type": "choice", "name": q.Name, "choice": q.Choices[0].Value, "confidence": 1, "probabilities": probs})
			case "score":
				probs := []any{}
				for i, l := range q.Levels {
					p := 0.0
					if i == 0 {
						p = 1
					}
					probs = append(probs, map[string]any{"value": i, "label": l.Label, "probability": p})
				}
				answers = append(answers, map[string]any{"type": "score", "name": q.Name, "score": 0, "confidence": 1, "probabilities": probs})
			case "predicate":
				answers = append(answers, map[string]any{"type": "predicate", "name": q.Name, "probability": 0.7})
			}
		}
		body := map[string]any{"model": "gpt-6-luna", "answers": answers, "usage": map[string]any{"input_tokens": 100, "output_tokens": 0, "total_tokens": 100, "input_tokens_details": map[string]any{"cached_tokens": 0, "cache_write_tokens": 0}, "output_tokens_details": map[string]any{"reasoning_tokens": 0}}}
		switch name {
		case "price-threshold", "price-long-tier":
			tokens := 272000
			if name == "price-long-tier" {
				tokens = 272001
			}
			usage := body["usage"].(map[string]any)
			usage["input_tokens"] = tokens
			usage["total_tokens"] = tokens
		case "usage-absent":
			delete(body, "usage")
		case "usage-partial":
			body["usage"] = map[string]any{"input_tokens": 100}
		case "usage-zero":
			body["usage"] = map[string]any{"input_tokens": 0, "output_tokens": 0, "total_tokens": 0, "input_tokens_details": map[string]any{"cached_tokens": 0, "cache_write_tokens": 0}, "output_tokens_details": map[string]any{"reasoning_tokens": 0}}
		case "model-alias":
			body["model"] = "controlled-actual-luna-alias"
		case "model-missing":
			delete(body, "model")
		}
		json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(server.Close)
	rec := newRecorder()
	transport := rec.next.(*http.Transport)
	t.Cleanup(transport.CloseIdleConnections)
	// Only the external network boundary is replaced. The Client still builds
	// and authorizes the official route. Production cannot select this transport.
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		copy := req.Clone(req.Context())
		u := *req.URL
		u.Scheme = "https"
		u.Host = strings.TrimPrefix(server.URL, "https://")
		copy.URL = &u
		copy.Host = u.Host
		return server.Client().Transport.RoundTrip(copy)
	})
	return rec
}
func controlledCandidateBudgetScenarios(t *testing.T, name string) []decisionsScenario {
	scenarios := decisionsScenarios(t)
	switch name {
	case "narrowed":
		return scenarios[:1]
	case "calls-budget":
		scenarios = nil
		for i := 0; i < 13; i++ {
			scenarios = append(scenarios, decisionsScenario{id: fmt.Sprintf("calls-budget-%d", i), request: singleChoiceRequest()})
		}
	case "questions-budget":
		scenarios = nil
		for i := 0; i < 9; i++ {
			scenarios = append(scenarios, decisionsScenario{id: fmt.Sprintf("questions-budget-%d", i), request: mixedQuestionsRequest()})
		}
	case "state-budget":
		req := singleChoiceRequest()
		req.State, _ = json.Marshal(strings.Repeat("x", 2<<20))
		scenarios = []decisionsScenario{{id: "state-budget-first", request: req}, {id: "state-budget", request: req}}
	case "images-budget":
		scenarios = nil
		asset := candidateImageAssets(t)[0]
		for i := 0; i < 9; i++ {
			req := singleChoiceRequest()
			req.Images = []ai.Image{asset}
			scenarios = append(scenarios, decisionsScenario{id: fmt.Sprintf("images-budget-%d", i), request: req})
		}
	case "image-bytes-budget":
		req := singleChoiceRequest()
		asset := candidateImageAssets(t)[0]
		raw, _ := decodeCandidateImage(asset)
		raw = append(raw, bytes.Repeat([]byte{0}, (2<<20)+1-len(raw))...)
		asset.Data = base64.StdEncoding.EncodeToString(raw)
		req.Images = []ai.Image{asset}
		scenarios = []decisionsScenario{{id: "image-bytes-budget-first", request: req}, {id: "image-bytes-budget", request: req}}
	}
	return scenarios
}
