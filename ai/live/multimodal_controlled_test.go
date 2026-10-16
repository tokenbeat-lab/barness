//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func controlledMultimodalRecorder(t *testing.T, route multimodalRoute, fault string) *recorder {
	t.Helper()
	if route.name == "decisions" {
		rec := controlledDecisionsRecorder(t, map[string]string{"malformed": "malformed-2xx", "429": "429", "process-deadline": "process-deadline"}[fault])
		rec.requestLimit = 4 << 20
		return rec
	}
	stop := make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-typesafe-request-id", "controlled-jev-mm")
		if fault == "process-deadline" {
			_, _ = io.Copy(io.Discard, r.Body)
			select {
			case <-r.Context().Done():
			case <-stop:
			}
			return
		}
		if fault == "429" {
			w.WriteHeader(429)
			fmt.Fprint(w, `{"error":"controlled rate limit"}`)
			return
		}
		if fault == "malformed" {
			fmt.Fprint(w, `{"answers":`)
			return
		}
		var req struct {
			Model     string
			Questions map[string]struct {
				Type     string
				Criteria json.RawMessage
			}
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid controlled native request")
			return
		}
		answers := map[string]any{}
		for id, q := range req.Questions {
			switch q.Type {
			case "noul":
				answers[id] = map[string]any{"type": "noul", "noul": 0}
			case "box":
				answers[id] = map[string]any{"type": "box", "box": map[string]float64{"top": 0.25, "bottom": 0.75, "left": 0.25, "right": 0.75}, "confidence": nil}
			case "choice":
				var criteria map[string]any
				_ = json.Unmarshal(q.Criteria, &criteria)
				p := map[string]float64{}
				first := ""
				for key := range criteria {
					first = key
					p[key] = 1
				}
				answers[id] = map[string]any{"type": "choice", "choice": first, "confidence": 1, "probabilities": p}
			case "score":
				var levels []any
				_ = json.Unmarshal(q.Criteria, &levels)
				legend := map[string]any{}
				p := map[string]float64{}
				for i, l := range levels {
					k := fmt.Sprint(i)
					legend[k] = l
					p[k] = 0
				}
				p["0"] = 1
				answers[id] = map[string]any{"type": "score", "score": 0, "confidence": 1, "probabilities": p, "legend": legend}
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": req.Model, "answers": answers, "usage": map[string]int{"input_tokens": 100, "output_tokens": 20}})
	}))
	t.Cleanup(func() { close(stop); server.Close() })
	rec := newRecorder()
	transport := rec.next.(*http.Transport)
	t.Cleanup(transport.CloseIdleConnections)
	rec.requestLimit = 4 << 20
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		copy := req.Clone(req.Context())
		u := *req.URL
		u.Host = strings.TrimPrefix(server.URL, "https://")
		copy.URL = &u
		copy.Host = u.Host
		return server.Client().Transport.RoundTrip(copy)
	})
	return rec
}
