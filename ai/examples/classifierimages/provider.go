package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

const syntheticKey = "synthetic-unified-example-credential"

type responseCapture struct {
	mu      sync.Mutex
	body    []byte
	problem error
}

func (c *responseCapture) snapshot() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]byte(nil), c.body...), c.problem
}

// A controlled vendor echoes only the rubric it received after all Hook guards.
// It does no vision. Business truth / visual quality must be evaluated separately.
func reply(s scenario) (provider.Reply, *responseCapture) {
	capture := &responseCapture{}
	chunks := make([][]byte, 1)
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "synthetic-example-response"}, Chunks: chunks, OnRequest: func(req provider.Request) {
		body, err := script(s, req.Body)
		capture.mu.Lock()
		capture.body, capture.problem = body, err
		capture.mu.Unlock()
		if err != nil {
			body = []byte(`{"error":"invalid controlled request"}`)
		}
		chunks[0] = body
	}}, capture
}
func script(s scenario, raw []byte) ([]byte, error) {
	if s.Route == "decisions" {
		var req struct {
			Model     string
			Questions []struct{ Type, Name string }
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, err
		}
		answers := []any{}
		for _, q := range req.Questions {
			a := map[string]any{"type": "predicate", "name": q.Name, "probability": 0}
			if s.Name == "refusal" {
				a = map[string]any{"type": "refusal", "name": q.Name}
			}
			answers = append(answers, a)
		}
		return json.Marshal(map[string]any{"model": req.Model, "answers": answers, "usage": map[string]any{"input_tokens": 100, "output_tokens": 20, "total_tokens": 120, "input_tokens_details": map[string]int{"cached_tokens": 0, "cache_write_tokens": 0}, "output_tokens_details": map[string]int{"reasoning_tokens": 0}}})
	}
	var req struct {
		Model     string
		Questions map[string]struct {
			Type     string
			Criteria []json.RawMessage
		}
	}
	// Bool criteria are absent in this example; score criteria are ordered arrays.
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	answers := map[string]any{}
	for name, q := range req.Questions {
		switch q.Type {
		case "noul":
			answers[name] = map[string]any{"type": "noul", "noul": 0}
		case "box":
			var confidence any
			if s.Entry == "hooked" {
				confidence = 0
			}
			a := map[string]any{"type": "box", "box": map[string]float64{"top": 0.25, "bottom": 0.75, "left": 0.25, "right": 0.75}, "confidence": confidence}
			if s.Name == "bad-box" {
				delete(a, "confidence")
			}
			answers[name] = a
		case "score":
			if len(q.Criteria) != 2 {
				return nil, fmt.Errorf("controlled rubric must have two levels")
			}
			legend := map[string]json.RawMessage{"0": q.Criteria[0], "1": q.Criteria[1]}
			if s.Name == "bad-legend" {
				legend["0"] = json.RawMessage(`"wrong"`)
			}
			answers[name] = map[string]any{"type": "score", "score": 1, "confidence": 1, "probabilities": map[string]float64{"0": 0, "1": 1}, "legend": legend}
		default:
			return nil, fmt.Errorf("unsupported controlled question kind")
		}
	}
	return json.Marshal(map[string]any{"model": req.Model, "answers": answers, "usage": map[string]int{"input_tokens": 100, "output_tokens": 20}})
}
