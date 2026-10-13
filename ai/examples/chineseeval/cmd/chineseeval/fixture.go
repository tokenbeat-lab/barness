package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

func fixtureValid(profile string, r chineseeval.Route) bool {
	switch profile {
	case "normal", "usage-absent", "usage-partial", "cancel", "budget", "response-alias", "dual-request-id", "pre-attempt-cancel", "http-failure":
		return true
	case "mixed", "all-refusal", "all-refusal-usage-absent", "all-refusal-usage-partial":
		return r.API == ai.APIOpenAIDecisions
	}
	return false
}

type fixtureCapture struct {
	Status         int             `json:"status,omitempty"`
	NetworkFailure bool            `json:"networkFailure,omitempty"`
	ReadFailure    bool            `json:"readFailure,omitempty"`
	Path           string          `json:"path"`
	Request        json.RawMessage `json:"request"`
	Response       json.RawMessage `json:"response,omitempty"`
	Canceled       bool            `json:"canceled,omitempty"`
}

// This synthetic transport cannot open a socket. It is independent of labels:
// every ordinary answer is billing, so the fixture cannot imply model quality.
type fixtureTransport struct {
	route    chineseeval.Route
	profile  string
	cancel   context.CancelFunc
	captures []fixtureCapture
}

func (f *fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != "POST" || req.URL.Host != reqHost(f.route) {
		return nil, errors.New("fixture route mismatch")
	}
	raw, err := io.ReadAll(io.LimitReader(req.Body, 16385))
	req.Body.Close()
	if err != nil || len(raw) > 16384 {
		return nil, errors.New("fixture request exceeds budget")
	}
	path := "/v1/systemone"
	if f.route.API == ai.APIOpenAIDecisions {
		path = "/v1/decisions"
	}
	if req.URL.Path != path {
		return nil, errors.New("fixture API mismatch")
	}
	var request struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(raw, &request) != nil || request.Model != f.route.Model {
		return nil, errors.New("fixture model mismatch")
	}
	capture := fixtureCapture{Path: path, Request: raw}
	if f.profile == "cancel" {
		capture.Canceled = true
		f.captures = append(f.captures, capture)
		f.cancel()
		return nil, req.Context().Err()
	}
	refusal := strings.HasPrefix(f.profile, "all-refusal") || (f.profile == "mixed" && len(f.captures)%2 == 1)
	model := f.route.Model
	if f.profile == "response-alias" {
		model = "synthetic-response-alias"
	}
	answer := map[string]any{"type": "choice", "choice": "billing", "confidence": .8}
	response := map[string]any{"model": model}
	header := http.Header{"Content-Type": []string{"application/json"}}
	if f.route.API == ai.APIOpenAIDecisions {
		header.Set("x-request-id", fmt.Sprintf("synthetic-zh-%d", len(f.captures)+1))
		answer["name"] = "intent"
		answer["probabilities"] = []any{map[string]any{"value": "billing", "probability": .7}, map[string]any{"value": "technical", "probability": .1}, map[string]any{"value": "delivery", "probability": .1}, map[string]any{"value": "account", "probability": .1}}
		if refusal {
			answer = map[string]any{"type": "refusal", "name": "intent"}
		}
		response["answers"] = []any{answer}
		response["usage"] = map[string]any{"input_tokens": 100, "output_tokens": 10, "total_tokens": 110, "input_tokens_details": map[string]int{"cached_tokens": 0, "cache_write_tokens": 0}, "output_tokens_details": map[string]int{"reasoning_tokens": 0}}
	} else {
		header.Set("x-typesafe-request-id", fmt.Sprintf("synthetic-zh-%d", len(f.captures)+1))
		answer["probabilities"] = map[string]float64{"billing": .7, "technical": .1, "delivery": .1, "account": .1}
		response["answers"] = map[string]any{"intent": answer}
		response["usage"] = map[string]int{"input_tokens": 100, "output_tokens": 10}
	}
	if strings.HasSuffix(f.profile, "usage-absent") {
		delete(response, "usage")
	}
	if strings.HasSuffix(f.profile, "usage-partial") {
		response["usage"] = map[string]int{"input_tokens": 100}
	}
	status := 200
	if f.profile == "dual-request-id" {
		header.Set("x-request-id", "synthetic-different-header")
	}
	if f.profile == "http-failure" {
		status = 503
		response = map[string]any{"error": map[string]string{"message": "synthetic outage", "type": "server_error"}}
	}
	capture.Status = status
	body, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	capture.Response = body
	f.captures = append(f.captures, capture)
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(bytes.NewReader(body)), Request: req}, nil
}
func reqHost(r chineseeval.Route) string {
	if r.Provider == ai.ProviderOpenAI {
		return "api.openai.com"
	}
	return "api.typesafe.ai"
}
