package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type MultimodalPromotedName struct {
	Image string `json:"MultimodalPromotedName"`
	Extra string `json:"extra"`
}

type multimodalAmbiguousA struct{ Hidden string }
type multimodalAmbiguousB struct{ Hidden string }

type multimodalBorrowedState struct {
	State  string   `json:"state"`
	Images []string `json:"images"`
}

// Borrowed callback trees can be large yet fit the document budget. A much
// smaller image-occurrence budget must reject them before encoded copies.
func TestMultimodalRuntimeBorrowedReferences(t *testing.T) {
	for _, shape := range []string{"map", "raw", "typed", "embedded", "integer-assets", "struct-assets", "wide"} {
		t.Run(shape, func(t *testing.T) {
			ev := run.Case(t, "UC09-borrowed-references-"+shape)
			u := newJevBoxWorld(t, ev, func(c *ai.Config) {
				c.Policy.MaxRequestBytes = 8 << 20
				c.Policy.Classifier.MaxStateBytes = 8 << 20
				c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 1 << 19}
			})
			req := jevBoxInput(t)
			var before, after runtime.MemStats
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				var key string
				for k := range p.Body["assets"].(map[string]any) {
					key = k
				}
				keys := make([]string, 65536)
				for i := range keys {
					keys[i] = key
				}
				state := map[string]any{"state": "bounded borrowed content", "images": keys}
				switch shape {
				case "map":
					p.Body["state"] = state
				case "raw":
					p.Body["state"] = json.RawMessage(mustMarshal(t, state))
				case "typed":
					p.Body["state"] = struct {
						State  string   `json:"state"`
						Images []string `json:"images"`
					}{"bounded borrowed content", keys}
				case "integer-assets":
					key = "424242"
					for i := range keys {
						keys[i] = key
					}
					var asset any
					for _, a := range p.Body["assets"].(map[string]any) {
						asset = a
					}
					p.Body["assets"] = map[int]any{424242: asset}
					p.Body["state"] = map[string]any{"state": "content", "images": keys}
				case "struct-assets":
					key = "photo-private-key"
					for i := range keys {
						keys[i] = key
					}
					var asset any
					for _, a := range p.Body["assets"].(map[string]any) {
						asset = a
					}
					p.Body["assets"] = struct {
						Photo any `json:"photo-private-key"`
					}{asset}
					p.Body["state"] = map[string]any{"state": "content", "images": keys}
				case "wide":
					fields := make([]reflect.StructField, 2048)
					for i := range fields {
						fields[i] = reflect.StructField{Name: fmt.Sprintf("Field%d", i), Type: reflect.TypeFor[int]()}
					}
					business := reflect.New(reflect.StructOf(fields)).Elem().Interface()
					p.Body["state"] = struct {
						State  any      `json:"state"`
						Images []string `json:"images"`
					}{business, keys}

				case "embedded":
					p.Body["state"] = struct{ multimodalBorrowedState }{multimodalBorrowedState{"bounded borrowed content", keys}}
				}
				runtime.ReadMemStats(&before)
				return ai.KeepPayload(), nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-borrowed-references-"+shape), jevContextTarget(), req, jevContextOptions())
			runtime.ReadMemStats(&after)
			allocated := after.TotalAlloc - before.TotalAlloc
			ceiling := uint64(512 << 10)
			if shape == "wide" {
				ceiling = 8 << 20
			}
			ev.Record("allocation", map[string]any{"postCallbackBytes": allocated, "ceiling": ceiling, "references": 65536, "shape": shape})
			ev.Check("reference budget enforced", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}), "got %v", err)
			ev.Check("known amplification rejected before copies", allocated < ceiling, "allocated %d", allocated)
			ev.Check("no admission or Attempt", len(u.adm.Requests()) == 0 && len(res.Metadata.Attempts) == 0, "sent amplified payload")
			recordJevContext(ev, u, req, res, err)
		})
	}
}

func TestMultimodalRuntimeBorrowedReferenceAccounting(t *testing.T) {
	for _, shape := range []string{"map", "raw", "typed", "untagged", "tagged-shadow", "ambiguous-omitted", "number", "promoted-name"} {
		for _, mode := range []string{"exact", "count", "expanded"} {
			t.Run(shape+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "UC09-reference-accounting-"+shape+"-"+mode)
				req := jevBoxInput(t)
				decoded, _ := base64.StdEncoding.DecodeString(req.Images[0].Data)
				u := newJevBoxWorld(t, ev, func(c *ai.Config) {
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 4, MaxExpandedImageBytes: 4 * int64(len(decoded))}
					if mode == "count" {
						c.Policy.Classifier.Images.References.MaxReferences = 3
					}
					if mode == "expanded" {
						c.Policy.Classifier.Images.References.MaxExpandedImageBytes--
					}
				})
				enqueueJevBox(t, ev, u, jevBoxAnswer, "0")
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					var key string
					for k := range p.Body["assets"].(map[string]any) {
						key = k
					}
					state := map[string]any{"state": map[string]any{"nested": []any{key, "business text"}}, "images": []string{key, key}}
					p.Body["questions"].(map[string]any)["exists"].(map[string]any)["instructions"] = []string{"compare", key}
					switch shape {
					case "map":
						p.Body["state"] = state
					case "raw":
						p.Body["state"] = json.RawMessage(mustMarshal(t, state))
					case "typed":
						p.Body["state"] = struct {
							State  map[string]any `json:"state"`
							Images []string       `json:"images"`
						}{state["state"].(map[string]any), []string{key, key}}
					case "tagged-shadow":
						p.Body["state"] = struct {
							State struct {
								Image string
								Alias string `json:"Image"`
							} `json:"state"`
							Images []string `json:"images"`
						}{struct {
							Image string
							Alias string `json:"Image"`
						}{"ordinary", key}, []string{key, key}}
					case "ambiguous-omitted":
						p.Body["state"] = map[string]any{"state": struct {
							multimodalAmbiguousA
							multimodalAmbiguousB
							Image string
						}{multimodalAmbiguousA{key}, multimodalAmbiguousB{key}, key}, "images": []string{key, key}}
					case "number":
						var asset any
						for _, a := range p.Body["assets"].(map[string]any) {
							asset = a
						}
						key = "424242"
						p.Body["assets"] = map[string]any{key: asset}
						p.Body["state"] = map[string]any{"state": map[string]any{"Image": key, "Number": json.Number(key)}, "images": []string{key, key}}
						p.Body["questions"].(map[string]any)["exists"].(map[string]any)["instructions"] = []string{"compare", key}

					case "promoted-name":
						p.Body["state"] = map[string]any{"state": struct{ MultimodalPromotedName }{MultimodalPromotedName{key, "ordinary"}}, "images": []string{key, key}}
					case "untagged":
						p.Body["state"] = struct {
							State struct {
								Image string
								Text  string
							} `json:"state"`
							Images []string `json:"images"`
						}{struct {
							Image string
							Text  string
						}{key, "business text"}, []string{key, key}}
					}
					return ai.KeepPayload(), nil
				}}
				res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-reference-accounting-"+shape+mode), jevContextTarget(), req, jevContextOptions())
				if mode == "exact" {
					ev.Check("exact mixed occurrences accepted", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					ev.Check("independent occurrence budget", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0, "got %v", err)
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
