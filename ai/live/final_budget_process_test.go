//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestFinalRoundBudgetProcess(t *testing.T) {
	for _, mode := range []string{"allowed", "declined", "failed-no-refund", "wrong-round", "image-openai", "image-google", "classifier", "decisions", "anthropic-long"} {
		t.Run(mode, func(t *testing.T) {
			cs := run.Case(t, "REGIONAL-final-round-usd-"+mode)
			base := t.TempDir()
			budget := filepath.Join(base, "budget.json")
			maximum := int64(1000000)
			if strings.HasPrefix(mode, "image-") {
				maximum = 2000000
			}
			if mode == "declined" {
				maximum = 1
			}
			raw, _ := json.Marshal(map[string]any{"schema": 1, "round": "provider-sites-final-acceptance-2", "maxMicrousd": maximum, "reservedMicrousd": 0, "attempts": 0})
			os.WriteFile(budget, raw, 0600)
			round := "provider-sites-final-acceptance-2"
			if mode == "wrong-round" {
				round = "provider-sites-final-acceptance-1"
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestFinalRoundBudgetFixture$", "-test.timeout=20s")
			for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
				if value := os.Getenv(name); value != "" {
					cmd.Env = append(cmd.Env, name+"="+value)
				}
			}
			cmd.Env = append(cmd.Env, "BARNESS_FINAL_BUDGET_FIXTURE="+mode, "BARNESS_FINAL_ROUND="+round, "BARNESS_FINAL_BUDGET="+budget, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(base, "native"))
			output, err := cmd.CombinedOutput()
			cs.Record("process", map[string]any{"argv": cmd.Args, "exitSuccess": err == nil, "output": string(output)})
			cs.Check("public client process enforces pre-send known-rate allowance", err == nil, "%v %s", err, output)
			value, _ := os.ReadFile(budget)
			cs.Record("usd-reservation", json.RawMessage(value))
		})
	}
}

func TestFinalRoundBudgetFixture(t *testing.T) {
	mode := os.Getenv("BARNESS_FINAL_BUDGET_FIXTURE")
	if mode == "" {
		t.Skip("isolated synthetic child only")
	}
	count := 0
	rec := newRecorder()
	rec.next.(*http.Transport).CloseIdleConnections()
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count++
		if mode == "failed-no-refund" || strings.HasPrefix(mode, "image-") || mode == "classifier" || mode == "decisions" || mode == "anthropic-long" {
			return nil, errors.New("controlled transport failure")
		}
		body := `data: {"id":"chat_controlled","model":"gpt-5-mini","choices":[{"index":0,"delta":{"content":"accepted"},"finish_reason":"stop"}]}` + "\n\ndata: [DONE]\n\n"
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	combo := &combo{name: "final-budget-chat", operation: ai.OperationChat, provider: ai.ProviderOpenAI, api: ai.APIOpenAICompletions, endpoint: "https://api.openai.com/v1", model: "gpt-5-mini"}
	if strings.HasPrefix(mode, "image-") {
		combo.operation = ai.OperationImage
		if mode == "image-google" {
			combo.provider = ai.ProviderGoogle
			combo.api = ai.APIGoogleInteractions
			combo.endpoint = "https://generativelanguage.googleapis.com/v1beta"
			combo.model = "gemini-nano-banana-2.1"
		} else {
			combo.api = ai.APIOpenAIImages
			combo.model = "gpt-image-2.5-sunburst-2026-09-08"
		}
	}
	if mode == "classifier" {
		combo.operation = ai.OperationClassifier
		combo.provider = ai.ProviderTypeSafe
		combo.api = ai.APITypeSafeSystemOne
		combo.endpoint = "https://api.typesafe.ai/v1"
		combo.model = "jev-1.13.0"
	}
	if mode == "decisions" {
		combo.operation = ai.OperationClassifier
		combo.api = ai.APIOpenAIDecisions
		combo.model = "gpt-6-luna"
	}
	if mode == "anthropic-long" {
		combo.provider = ai.ProviderAnthropic
		combo.api = ai.APIAnthropicMessages
		combo.endpoint = "https://api.anthropic.com"
		combo.model = "claude-haiku-4-5"
	}
	client, err := newClient(combo, "synthetic-budget-credential", rec, ai.BuiltinCatalog())
	if err != nil {
		t.Fatal(err)
	}
	target := ai.Target{BindingID: combo.name, ModelID: combo.model}
	if strings.HasPrefix(mode, "image-") {
		ref, _ := smokeImageInputs(t)
		request := ai.ImagesRequest{Prompt: "Controlled reference edit.", ReferenceImages: []ai.Image{ref}}
		if mode == "image-google" {
			_, err = client.GenerateImages(context.Background(), newScope(), target, request, googleImageSmokeOptions())
		} else {
			_, err = client.GenerateImages(context.Background(), newScope(), target, request, imageSmokeOptions("png", "low"))
		}
	} else if mode == "decisions" {
		_, err = client.Classify(context.Background(), newScope(), target, singleChoiceRequest(), ai.OpenAIDecisionsOptions{})
	} else if mode == "classifier" {
		_, err = client.Classify(context.Background(), newScope(), target, singleChoiceRequest(), ai.TypeSafeOptions{})
	} else if mode == "anthropic-long" {
		_, err = client.Complete(context.Background(), newScope(), target, ai.Request{Messages: []ai.Message{ai.UserText("Controlled long cache prompt.")}}, ai.AnthropicOptions{MaxTokens: ai.Value(1024), CacheRetention: ai.CacheRetentionLong})
	} else {
		_, err = client.Complete(context.Background(), newScope(), ai.Target{BindingID: combo.name, ModelID: combo.model}, ai.Request{Messages: []ai.Message{ai.UserText("Controlled final wire allowance.")}}, ai.ChatOptions{MaxTokens: ai.Value(1024)})
	}
	if mode == "allowed" {
		if err != nil || count != 1 {
			t.Fatalf("allowed request %d %v", count, err)
		}
	} else if mode == "failed-no-refund" || strings.HasPrefix(mode, "image-") || mode == "classifier" || mode == "decisions" || mode == "anthropic-long" {
		if err == nil || count != 1 {
			t.Fatalf("failure after send %d %v", count, err)
		}
	} else if err == nil || count != 0 {
		t.Fatalf("decline must precede transport: requests=%d error=%v", count, err)
	}
	var value struct {
		Reserved int64 `json:"reservedMicrousd"`
		Attempts int   `json:"attempts"`
		Last     struct {
			Input      int64   `json:"inputReserve"`
			Output     int64   `json:"outputReserve"`
			InputRate  float64 `json:"inputRate"`
			OutputRate float64 `json:"outputRate"`
		} `json:"last"`
	}
	raw, _ := os.ReadFile(os.Getenv("BARNESS_FINAL_BUDGET"))
	if json.Unmarshal(raw, &value) != nil {
		t.Fatal("budget artifact missing")
	}
	if mode == "allowed" || mode == "failed-no-refund" || strings.HasPrefix(mode, "image-") || mode == "classifier" || mode == "decisions" || mode == "anthropic-long" {
		if value.Reserved <= 0 || value.Attempts != 1 {
			t.Fatalf("sent attempts retain allowance %+v", value)
		}
		if strings.HasPrefix(mode, "image-") && value.Last.Input < 65536 {
			t.Fatalf("compressed image still needs geometric reserve %+v", value)
		}
		if mode == "image-google" && value.Last.Output != 32768 {
			t.Fatalf("Google total output limit reserve %+v", value)
		}
		if mode == "image-openai" && value.Last.Output != 7024 {
			t.Fatalf("sourced calculator max quality reserve %+v", value)
		}
		if mode == "anthropic-long" && value.Last.InputRate < 2 {
			t.Fatalf("one-hour writes reserve twice input price %+v", value)
		}
		if (mode == "classifier" || mode == "decisions") && value.Last.Output != 0 {
			t.Fatalf("classifier output remains unpriced/free %+v", value)
		}
		if mode == "decisions" && (value.Last.InputRate != 0.2 || value.Last.OutputRate != 0) {
			t.Fatalf("Decisions uses its own highest input tier and free output %+v", value)
		}
	} else if value.Reserved != 0 || value.Attempts != 0 {
		t.Fatalf("pre-send decline consumed allowance %+v", value)
	}
}
