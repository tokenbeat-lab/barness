package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These are structural consumer fixtures derived from a keyless controlled
// public Client process. Their synthetic live-shaped records are never a live
// stage receipt, never use vendor credentials, and never grant release support.
// Failure inventory: raw generic projections, candidate/controlled origins,
// partial/substituted scenes, altered capture/usage/Observer and provenance
// must refuse; the complete synthetic contract must be independently readable.
func TestXAIFormalNativeConsumer(t *testing.T) {
	privateGrantStore(t)
	fixture := newXAIFormalNativeFixture(t)
	cases := []struct {
		name string
		pass bool
		edit func(string)
	}{
		{"complete-synthetic", true, nil},
		{"optional-header-present", true, func(dir string) { addXAINativeFixtureHeaders(t, dir) }},
		{"duplicate-identical-terminal", true, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-complete/text-http.json", func(v []any) {
				e := v[0].(map[string]any)
				body := e["responseBody"].(string)
				chunks := strings.Split(body, "\n\n")
				for _, chunk := range chunks {
					if strings.Contains(chunk, "response.completed") {
						e["responseBody"] = strings.Replace(body, "data: [DONE]", chunk+"\n\ndata: [DONE]", 1)
						break
					}
				}
			})
		}},
		{"raw-reasoning-contradiction", false, func(dir string) {
			editXAINativeTerminal(t, dir, func(v map[string]any) {
				v["usage"].(map[string]any)["output_tokens_details"].(map[string]any)["reasoning_tokens"] = 1200
			})
		}},
		{"raw-cache-over-input", false, func(dir string) {
			editXAINativeTerminal(t, dir, func(v map[string]any) {
				v["usage"].(map[string]any)["input_tokens_details"].(map[string]any)["cached_tokens"] = 33
			})
		}},
		{"raw-usage-missing", false, func(dir string) {
			editXAINativeTerminal(t, dir, func(v map[string]any) { delete(v["usage"].(map[string]any), "total_tokens") })
		}},
		{"raw-fractional-token", false, func(dir string) {
			editXAINativeTerminal(t, dir, func(v map[string]any) { v["usage"].(map[string]any)["input_tokens"] = 32.5 })
		}},

		{"grant-valid-at-claim-expired-now", true, func(dir string) { expireXAINativeFixtureGrant(t, dir, time.Nanosecond) }},
		{"grant-expired-at-claim", false, func(dir string) { expireXAINativeFixtureGrant(t, dir, 0) }},
		{"stream-index-substituted", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-stream/text-events.json", func(v []any) {
				for _, raw := range v {
					e := raw.(map[string]any)
					if x, ok := e["ContentIndex"].(float64); ok {
						e["ContentIndex"] = x + 100
					}
				}
			})
		}},
		{"hosted-tool-substituted", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-tools-required/calls-http.json", func(v map[string]any) { v["tools"].([]any)[0].(map[string]any)["type"] = "mcp" })
		}},
		{"hidden-process-attempt", false, func(dir string) {
			mustWriteXAINative(t, filepath.Join(dir, "E17-xai-process-gate", "hidden-http.json"), []byte("{}\n"))
		}},

		{"observer-attribution-substituted", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-complete/text-observer.json", func(v []any) { v[1].(map[string]any)["call"].(map[string]any)["tenantId"] = "SYNTHETIC_other" })
		}},
		{"stream-block-end-substituted", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-stream/text-events.json", func(v []any) {
				for _, raw := range v {
					e := raw.(map[string]any)
					if e["type"] == "text_end" {
						e["Content"] = "SYNTHETIC_changed"
					}
				}
			})
		}},

		{"entry-visible-text-missing", false, func(dir string) {
			editXAINativeTerminal(t, dir, func(v map[string]any) {
				for _, raw := range v["output"].([]any) {
					item := raw.(map[string]any)
					if item["type"] == "message" {
						for _, part := range item["content"].([]any) {
							part.(map[string]any)["text"] = ""
						}
					}
				}
			})
			editXAINativeJSON(t, dir, "E17-xai-entry-complete/text-result.json", func(v map[string]any) {
				for _, raw := range v["message"].(map[string]any)["content"].([]any) {
					item := raw.(map[string]any)
					if _, ok := item["text"]; ok {
						item["text"] = ""
					}
				}
			})
		}},

		{"input-json-order-equivalent", true, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-encrypted-continuation/continued-http.json", func(v map[string]any) { v["store"] = false })
		}},

		{"manifest-time-exceeds-budget", false, func(dir string) {
			editXAINativeJSON(t, dir, "manifest.json", func(v map[string]any) {
				start, err := time.Parse(time.RFC3339Nano, v["started_at"].(string))
				if err != nil {
					t.Fatal(err)
				}
				v["finished_at"] = start.Add(time.Hour)
			})
		}},

		{"tool-native-without-reasoning-item", true, func(dir string) { omitXAIToolReasoningFixture(t, dir) }},

		{"tool-observed-reasoning-dropped", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-tools-required/results-http.json", func(v map[string]any) {
				var input []any
				for _, raw := range v["input"].([]any) {
					if raw.(map[string]any)["type"] != "reasoning" {
						input = append(input, raw)
					}
				}
				v["input"] = input
			})
		}},

		{"noncanonical-mode", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["Mode"] = v["mode"]; delete(v, "mode") })
		}},
		{"projection-disagrees", false, func(dir string) {
			editXAINativeJSON(t, dir, "live-report.json", func(v map[string]any) { v["budget"].(map[string]any)["outputTokensUsed"] = 1 })
		}},
		{"null-store", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-entry-complete/text-http.json", func(v map[string]any) { v["store"] = nil })
		}},
		{"cost-disagrees", false, func(dir string) {
			editXAINativeJSON(t, dir, "E17-xai-entry-complete/text-result.json", func(v map[string]any) {
				v["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["costEstimate"].(map[string]any)["maximum"].(map[string]any)["total"] = 0
			})
		}},
		{"stream-type-missing", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-stream/text-events.json", func(v []any) { delete(v[0].(map[string]any), "type") })
		}},
		{"native-replay-changed", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-encrypted-continuation/continued-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					item := raw.(map[string]any)
					if item["type"] == "reasoning" {
						item["encrypted_content"] = "synthetic-changed"
						break
					}
				}
			})
		}},
		{"image-url", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-user-png/image-http.json", func(v map[string]any) {
				input := v["input"].([]any)
				blocks := input[0].(map[string]any)["content"].([]any)
				for _, raw := range blocks {
					b := raw.(map[string]any)
					if b["type"] == "input_image" {
						b["image_url"] = "https://example.invalid/picture.png"
					}
				}
			})
		}},
		{"history-prefix-lost", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-encrypted-continuation/continued-http.json", func(v map[string]any) { v["input"] = v["input"].([]any)[1:] })
		}},
		{"parallel-pair-lost", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-tools-parallel/results-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					item := raw.(map[string]any)
					if item["type"] == "function_call_output" {
						item["call_id"] = "SYNTHETIC_unknown"
						break
					}
				}
			})
		}},
		{"unknown-input-part", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-entry-complete/text-http.json", func(v map[string]any) {
				items := v["input"].([]any)
				for _, raw := range items {
					item := raw.(map[string]any)
					if item["role"] == "user" {
						item["content"].([]any)[0].(map[string]any)["type"] = "input_audio"
					}
				}
			})
		}},
		{"native-order-changed", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-encrypted-continuation/continued-http.json", func(v map[string]any) {
				items := v["input"].([]any)
				items[1], items[2] = items[2], items[1]
			})
		}},
		{"tool-image-order-changed", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-tool-result-image/results-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					item := raw.(map[string]any)
					blocks, ok := item["content"].([]any)
					if ok && len(blocks) == 2 && blocks[0].(map[string]any)["type"] == "input_image" {
						blocks[0], blocks[1] = blocks[1], blocks[0]
					}
				}
			})
		}},

		{"scope-substituted", false, func(dir string) {
			editXAINativeJSON(t, dir, "E17-xai-entry-complete/text-result.json", func(v map[string]any) { v["metadata"].(map[string]any)["accountScopeId"] = "another-account" })
		}},
		{"controlled", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["mode"] = "controlled"; v["formalQualified"] = false })
		}},
		{"candidate", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["stage"] = "candidate"; v["kind"] = "xai-candidate-only" })
		}},
		{"generic-only", false, func(dir string) { mustRemoveXAINative(t, filepath.Join(dir, "xai-formal-report.json")) }},
		{"missing-scenario", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["scenarios"] = v["scenarios"].([]any)[1:] })
		}},
		{"duplicate-scenario", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { s := v["scenarios"].([]any); s[1] = s[0] })
		}},
		{"substituted-model", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["model"] = "another-model" })
		}},
		{"wrong-source", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) {
				v["sourceSha256"] = "0000000000000000000000000000000000000000000000000000000000000000"
			})
		}},
		{"wrong-build", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) {
				v["buildSourceSha256"] = "0000000000000000000000000000000000000000000000000000000000000000"
			})
		}},
		{"wrong-binary", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) {
				v["binarySha256"] = "0000000000000000000000000000000000000000000000000000000000000000"
			})
		}},
		{"wrong-sdk", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["sdk"] = "github.com/openai/openai-go/v3 v0.0.0" })
		}},
		{"wrong-catalog", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) {
				v["catalogHash"] = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			})
		}},
		{"wrong-budget", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) { v["budget"].(map[string]any)["maxCalls"] = 32 })
		}},
		{"missing-triplet", false, func(dir string) {
			mustRemoveXAINative(t, filepath.Join(dir, "E17-xai-entry-complete", "text-observer.json"))
		}},
		{"observer-disagrees", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-complete/text-observer.json", func(v []any) { v[len(v)-1].(map[string]any)["usage"].(map[string]any)["output"] = 999 })
		}},
		{"usage-disagrees", false, func(dir string) {
			editXAINativeJSON(t, dir, "E17-xai-entry-complete/text-result.json", func(v map[string]any) { v["message"].(map[string]any)["usage"].(map[string]any)["totalTokens"] = 999 })
		}},
		{"truncated", false, func(dir string) {
			editXAINativeArray(t, dir, "E17-xai-entry-complete/text-http.json", func(v []any) { v[0].(map[string]any)["truncated"] = true })
		}},
		{"stored-conversation", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-entry-complete/text-http.json", func(v map[string]any) { v["store"] = true })
		}},
		{"missing-native-include", false, func(dir string) {
			editXAINativeHTTPBody(t, dir, "E17-xai-entry-complete/text-http.json", func(v map[string]any) { delete(v, "include") })
		}},
		{"extra-hidden-file", false, func(dir string) { mustWriteXAINative(t, filepath.Join(dir, "hidden-attempt.json"), []byte("{}\n")) }},
		{"missing-claim", false, func(dir string) { mustRemoveXAINative(t, filepath.Join(dir, "xai-authorization-claim.json")) }},
		{"wrong-claim", false, func(dir string) {
			editXAINativeJSON(t, dir, "xai-authorization-claim.json", func(v map[string]any) {
				v["grantSha256"] = "0000000000000000000000000000000000000000000000000000000000000000"
			})
		}},
	}
	for _, sc := range cases {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E17-formal-native-"+sc.name)
			dir := fixture.copy(t, sc.name)
			if sc.edit != nil {
				sc.edit(dir)
			}
			// Rebind ordinary fixture manifests so semantic mutations reach the
			// independent checker instead of merely failing a stale file hash.
			resealXAINativeManifest(t, dir)
			code, raw := fixture.consume(t, ev, dir)
			var got struct {
				Pass  bool   `json:"pass"`
				Error string `json:"error"`
			}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal("native consumer did not preserve a result")
			}
			ev.Check("native structural qualification result", (code == 0) == sc.pass && got.Pass == sc.pass, "exit=%d pass=%t want=%t error=%s", code, got.Pass, sc.pass, got.Error)
		})
	}
	verifyXAINativeMatrixFixtures(t, fixture)
}

func mustRemoveXAINative(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
