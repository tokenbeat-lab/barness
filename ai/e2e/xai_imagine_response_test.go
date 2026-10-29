package e2e

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIImagineResponseContract(t *testing.T) {
	for _, name := range []string{"mime-null", "jpeg", "webp", "mime-conflict", "invalid-base64", "url-only", "url-and-data", "empty", "refusal", "moderation-false", "async", "duplicate-data", "bad-json", "bad-count", "atomic", "cost-only", "cost-zero", "cost-null", "cost-negative", "cost-fraction", "cost-overflow", "cost-duplicate", "token-complete", "token-partial", "token-inconsistent", "failure-retains-facts", "reported-identity", "missing-identity", "response-202"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-response-"+name)
			w := newImagineWorld(t)
			entry := `{"b64_json":"` + imaginePNG + `"}`
			usage := ""
			extra := ""
			success := false
			code := ai.CodeProtocol
			status := 200
			wantReported := false
			amount := 0.0
			wantUsage := ai.UsageUnreported
			switch name {
			case "mime-null":
				entry = `{"b64_json":"` + imaginePNG + `","mime_type":null}`
				success = true
			case "jpeg":
				entry = `{"b64_json":"/9j/2Q=="}`
				success = true
			case "webp":
				entry = `{"b64_json":"UklGRgAAAABXRUJQ"}`
				success = true
			case "mime-conflict":
				entry = `{"b64_json":"` + imaginePNG + `","mime_type":"image/jpeg"}`
			case "invalid-base64":
				entry = `{"b64_json":"%%%"}`
			case "url-only":
				entry = `{"url":"https://example.invalid/a.png"}`
			case "url-and-data":
				entry = `{"b64_json":"` + imaginePNG + `","url":"https://example.invalid/a.png"}`
			case "empty":
				entry = ""
			case "refusal":
				entry = ""
				extra = `,"error":{"code":"content_policy_violation"}`
				code = ai.CodeUpstreamError
			case "moderation-false":
				entry = `{"b64_json":"` + imaginePNG + `","respect_moderation":false}`
				code = ai.CodeUpstreamError
			case "async":
				extra = `,"request_id":"deferred-request"`
			case "duplicate-data":
				extra = `,"data":[]`
			case "bad-json":
				extra = "BROKEN"
			case "bad-count":
				entry = entry + "," + entry
			case "atomic":
				entry = entry + `,{"b64_json":"bad"}`
			case "cost-only", "cost-zero", "cost-null", "cost-negative", "cost-fraction", "cost-overflow", "cost-duplicate", "failure-retains-facts":
				ticks := "600000000"
				if name == "cost-zero" {
					ticks = "0"
				}
				if name == "cost-null" {
					ticks = "null"
				}
				if name == "cost-negative" {
					ticks = "-1"
				}
				if name == "cost-fraction" {
					ticks = "1.5"
				}
				if name == "cost-overflow" {
					ticks = "9223372036854775808"
				}
				if name == "cost-duplicate" {
					ticks = `1,"cost_in_usd_ticks":2`
				}
				usage = `,"usage":{"cost_in_usd_ticks":` + ticks + `}`
				success = name == "cost-only" || name == "cost-zero" || name == "cost-null"
				wantReported = name == "cost-only" || name == "cost-zero" || name == "failure-retains-facts"
				if ticks == "600000000" {
					amount = .06
				}
				if name == "failure-retains-facts" {
					entry = `{"b64_json":"%%%"}`
				}
			case "token-complete":
				usage = `,"usage":{"input_tokens":10,"output_tokens":8,"total_tokens":18,"input_tokens_details":{"cached_tokens":2},"output_tokens_details":{"reasoning_tokens":3}}`
				success = true
				wantUsage = ai.UsageComplete
			case "token-partial":
				usage = `,"usage":{"input_tokens":10}`
				success = true
				wantUsage = ai.UsagePartial
			case "token-inconsistent":
				usage = `,"usage":{"input_tokens":10,"output_tokens":8,"total_tokens":17}`
			case "reported-identity":
				extra = `,"id":"native-id","model":"serving-model"`
				success = true
			case "missing-identity":
				success = true
			case "response-202":
				status = 202
			}
			body := `{"data":[` + entry + `]` + usage + extra + `}`
			reply := imagineReplyStatus(status, body)
			enqueue(ev, w, reply)
			result, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-response-"+name), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, ai.XAIImagesOptions{Resolution: ai.Value("1k"), Quality: ai.Value("low")})
			ev.Record("result", result)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("response outcome", (err == nil) == success, "err=%v", err)
			ev.Check("one paid attempt only", len(w.provider.Requests()) == 1 && len(result.Metadata.Attempts) == 1, "metadata=%+v", result.Metadata)
			if !success {
				var ae *ai.Error
				ev.Check("failure classification and atomic content", errors.As(err, &ae) && ae.Code == code && len(result.Content) == 0, "err=%v content=%+v", err, result.Content)
			}
			if len(result.Metadata.Attempts) > 0 {
				a := result.Metadata.Attempts[0]
				var fields map[string]json.RawMessage
				mustUnmarshal(t, mustMarshal(t, a), &fields)
				reported := fields["reportedCostUsd"]
				ev.Check("reported amount presence", (len(reported) > 0) == wantReported, "attempt=%s", mustMarshal(t, a))
				if wantReported {
					var actual float64
					mustUnmarshal(t, reported, &actual)
					ev.Check("ticks conversion", math.Abs(actual-amount) < 1e-12, "amount=%g", actual)
				}
				if success {
					ev.Check("token reporting independent of money", a.UsageReporting == wantUsage, "attempt=%+v", a)
				}
			}
			if name == "reported-identity" {
				id, _ := result.ResponseID.Get()
				model, _ := result.ResponseModel.Get()
				ev.Check("native identity retained", id == "native-id" && model == "serving-model", "result=%+v", result)
			}
			if name == "missing-identity" {
				ev.Check("identity remains missing", result.ResponseID.IsZero() && result.ResponseModel.IsZero(), "result=%+v", result)
			}
			if name == "token-complete" {
				ev.Check("tokens counted once", result.Usage.Input == 10 && result.Usage.Output == 8 && result.Usage.TotalTokens == 18, "usage=%+v", result.Usage)
			}
		})
	}
}
