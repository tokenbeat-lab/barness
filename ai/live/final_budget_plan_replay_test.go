//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Actual audited request bodies are supplied explicitly by the offline runner.
// Public hooks re-encode them under the same operation/provider/API/model; the
// final-wire recorder/guard and synthetic transport are the ordinary Host chain.
func TestFinalBudgetCapturedPlan(t *testing.T) {
	path := os.Getenv("BARNESS_FINAL_BUDGET_CAPTURE_FIXTURES")
	if path == "" {
		t.Skip("explicit audited offline request fixture only")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil || len(raw) > 16<<20 {
		t.Fatal("bounded captured plan required")
	}
	var cases []struct{ Name, Source, Operation, Provider, API, Endpoint, Model, Body string }
	if json.Unmarshal(raw, &cases) != nil || len(cases) == 0 || len(cases) > 512 {
		t.Fatal("finite captured plan required")
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%03d-%s", i, tc.Name), func(t *testing.T) {
			cs := run.Case(t, fmt.Sprintf("REGIONAL-budget-captured-%03d-%s", i, tc.Name))
			cs.Record("source", tc.Source)
			var body map[string]any
			decoder := json.NewDecoder(strings.NewReader(tc.Body))
			decoder.UseNumber()
			if err := decoder.Decode(&body); err != nil {
				t.Fatal(err)
			}
			budget := filepath.Join(t.TempDir(), "usd.json")
			os.WriteFile(budget, []byte(`{"schema":1,"round":"provider-sites-final-acceptance-2","maxMicrousd":30000000,"reservedMicrousd":0,"attempts":0}`), 0600)
			t.Setenv("BARNESS_FINAL_ROUND", "provider-sites-final-acceptance-2")
			t.Setenv("BARNESS_FINAL_BUDGET", budget)
			count := 0
			rec := newRecorder()
			rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = roundTripFunc(func(*http.Request) (*http.Response, error) {
				count++
				return nil, errors.New("controlled captured-plan transport")
			})
			c := &combo{name: tc.Name, operation: ai.Operation(tc.Operation), provider: ai.ProviderID(tc.Provider), api: ai.API(tc.API), endpoint: tc.Endpoint, model: tc.Model}
			if strings.HasPrefix(tc.Name, "mm-global-formal-m3-") {
				c.name = "minimax-" + strings.TrimPrefix(tc.Name, "mm-global-formal-m3-")
			}
			cat := ai.BuiltinCatalog()
			// This is the existing smoke's deliberate unsupported Chat-model probe:
			// borrow the same model's already sourced Responses price, as original P04.
			if strings.Contains(tc.Source, "pro-model-unavailable") {
				for _, m := range cat.Models {
					if m.Provider == c.provider && m.API == ai.APIOpenAIResponses && m.ID == c.model {
						m.API = c.api
						cat.Models = append(cat.Models, m)
						break
					}
				}
			}
			client, err := newClient(c, "synthetic-budget-credential", rec, cat)
			if err != nil {
				t.Fatal(err)
			}
			hook := client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				if p.Operation != c.operation || p.ProviderID != c.provider || p.API != c.api || p.ModelID != c.model {
					return ai.KeepPayload(), errors.New("captured identity changed")
				}
				return ai.ReplacePayload(body), nil
			}})
			target := ai.Target{BindingID: c.name, ModelID: c.model}
			if c.operation == ai.OperationImage {
				request := ai.ImagesRequest{Prompt: "Controlled captured image request."}
				if c.api == ai.APIOpenAIImages {
					opts := imageSmokeOptions("png", "low")
					if _, ok := body["images"]; ok {
						ref, mask := smokeImageInputs(t)
						request.ReferenceImages = []ai.Image{ref}
						if _, ok := body["mask"]; ok {
							opts.Mask = &mask
						}
					}
					_, err = hook.GenerateImages(context.Background(), newScope(), target, request, opts)
				} else {
					_, err = hook.GenerateImages(context.Background(), newScope(), target, request, googleImageSmokeOptions())
				}
			} else {
				var opts ai.Options
				switch c.api {
				case ai.APIOpenAIResponses:
					responses := ai.ResponsesOptions{MaxTokens: ai.Value(1024)}
					if _, ok := body["prompt_cache_key"]; ok && c.provider == ai.ProviderMiniMax {
						responses.SessionID = "synthetic-fixed-prefix"
						responses.CacheRetention = ai.CacheRetentionShort
					}
					opts = responses
				case ai.APIOpenAICompletions:
					chat := ai.ChatOptions{MaxTokens: ai.Value(1024)}
					if _, ok := body["prompt_cache_key"]; ok {
						chat.SessionID = "live-smoke-session"
						chat.CacheRetention = ai.CacheRetentionLong
					}
					opts = chat
				case ai.APIAnthropicMessages:
					opts = ai.AnthropicOptions{MaxTokens: ai.Value(1024)}
				case ai.APIGoogleGenerativeAI:
					opts = ai.GeminiOptions{MaxTokens: ai.Value(1024)}
				}
				_, err = hook.Complete(context.Background(), newScope(), target, ai.Request{Messages: []ai.Message{ai.UserText("Controlled captured chat request.")}}, opts)
			}
			exchanges := rec.take()
			cs.Record("final-wire", exchanges)
			ledger, _ := os.ReadFile(budget)
			cs.Fixture("usd.json", ledger)
			causes := []string{}
			for cause, n := err, 0; cause != nil && n < 8; cause, n = errors.Unwrap(cause), n+1 {
				causes = append(causes, cause.Error())
			}
			cs.Record("errors", causes)
			cs.Check("actual public operation reaches exactly one synthetic send", err != nil && count == 1, "count=%d err=%v", count, err)
			if len(exchanges) == 1 {
				var original, final any
				json.Unmarshal([]byte(tc.Body), &original)
				json.Unmarshal([]byte(exchanges[0].RequestBody), &final)
				a, _ := json.Marshal(original)
				b, _ := json.Marshal(final)
				cs.Check("captured final payload semantics retained", bytes.Equal(a, b), "payload changed")
			}
		})
	}
}
