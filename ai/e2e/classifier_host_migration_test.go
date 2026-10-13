package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Both existing Host assembly interfaces are exercised, including the original
// Jev structured descriptions. No migration layer converts incompatible inputs.
func TestClassifierHostMigration(t *testing.T) {
	for _, host := range []string{"local", "cloud"} {
		for _, route := range []string{"jev", "luna"} {
			for _, scenario := range []string{"nil-options", "protocol-options", "structured", "wrong-options", "refusal", "image", "no-image-binding", "no-image-policy", "no-image-model"} {
				if route == "jev" && (scenario == "refusal" || scenario == "image" || scenario == "no-image-binding" || scenario == "no-image-policy" || scenario == "no-image-model") {
					continue
				}
				t.Run(host+"/"+route+"/"+scenario, func(t *testing.T) {
					ev := run.Case(t, "OD56-host-"+host+"-"+route+"-"+scenario)
					w := newWorldWith(t, func(c *ai.Config) {
						c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 1, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 1, MaxTotalImageBytes: 4096}}
						if scenario == "no-image-policy" {
							c.Policy.Classifier.Images = nil
						}
						if scenario == "no-image-model" {
							catalog := ai.BuiltinCatalog()
							c.Catalog = &catalog
							for i := range c.Catalog.ClassifierModels {
								m := &c.Catalog.ClassifierModels[i]
								if m.ID == "gpt-6-luna" {
									m.Input = []ai.Modality{ai.ModalityText}
									m.Capabilities.MaxInputImages = 0
									m.Capabilities.ImageMIMETypes = nil
								}
							}
						}
					}, tenantA)
					b := primaryBinding(tenantA, w.provider.URL())
					b.BindingID, b.Operation = "migration", ai.OperationClassifier
					b.ProviderID, b.API, b.AllowedModels = ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, []string{"jev-1.13.0"}
					var options ai.Options
					if route == "luna" {
						b.ProviderID, b.API, b.AllowedModels = ai.ProviderOpenAI, ai.APIOpenAIDecisions, []string{"gpt-6-luna"}
						if scenario == "protocol-options" {
							options = ai.OpenAIDecisionsOptions{}
						}
						if scenario == "wrong-options" {
							options = ai.TypeSafeOptions{}
						}
					} else {
						if scenario == "protocol-options" {
							options = ai.TypeSafeOptions{}
						}
						if scenario == "wrong-options" {
							options = ai.OpenAIDecisionsOptions{}
						}
					}
					b.AllowClassifierImages = scenario == "image" || scenario == "no-image-policy" || scenario == "no-image-model"
					w.host.PutBinding(b)
					client := w.client
					if host == "local" {
						t.Setenv("BARNESS_HOST_MIGRATION_KEY", tenantA.secret)
						cat := w.client.Catalog()
						policy := validPolicy()
						policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 1, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 1, MaxTotalImageBytes: 4096}}
						if scenario == "no-image-policy" {
							policy.Classifier.Images = nil
						}
						assembly, err := localassembly.OpenOperations(localassembly.OperationsConfig{TenantID: tenantA.tenant, Bindings: []ai.Binding{b}, Keys: map[string]localassembly.KeySource{b.CredentialRef: {EnvVar: "BARNESS_HOST_MIGRATION_KEY"}}, Policy: policy, Catalog: &cat, Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true})
						if err != nil {
							t.Fatal(err)
						}
						client = assembly.Client
					}
					model, ok := client.Catalog().LookupClassifier(b.ProviderID, b.API, b.AllowedModels[0])
					ev.Check("discover independent input capabilities", ok && (route == "jev" && len(model.Input) == 0 || route == "luna" && len(model.Input) > 0), "model capability missing")
					q := ai.ChoiceQuestion{Instructions: json.RawMessage(`"判断当前目标"`), Criteria: map[string]json.RawMessage{"billing": json.RawMessage(`"支付"`), "technical": json.RawMessage(`"故障"`), "delivery": json.RawMessage(`"配送"`), "account": json.RawMessage(`"账号"`)}}
					if scenario == "structured" {
						q.Instructions = json.RawMessage(`{"task":"判断当前目标"}`)
					}
					input := ai.ClassifierRequest{State: json.RawMessage(`"请求退款"`), Questions: map[string]ai.ClassifierQuestion{"intent": q}}
					if scenario == "image" || scenario == "no-image-binding" || scenario == "no-image-policy" || scenario == "no-image-model" {
						raw, err := os.ReadFile("testdata/decisions/images/gopher.webp")
						if err != nil {
							t.Fatal(err)
						}
						input.Images = []ai.Image{{MimeType: "image/webp", Data: base64.StdEncoding.EncodeToString(raw)}}
					}
					reply := chineseReply("billing")
					if route == "luna" {
						answer := `{"type":"choice","name":"intent","choice":"billing","confidence":0.8,"probabilities":[{"value":"billing","probability":0.7},{"value":"technical","probability":0.1},{"value":"delivery","probability":0.1},{"value":"account","probability":0.1}]}`
						if scenario == "refusal" {
							answer = `{"type":"refusal","name":"intent"}`
						}
						raw := []byte(`{"model":"gpt-6-luna","answers":[` + answer + `],"usage":{"input_tokens":100,"output_tokens":10,"total_tokens":110,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}}`)
						reply = provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "synthetic-migration"}, Chunks: [][]byte{raw}, Script: string(raw)}
					}
					w.provider.Enqueue(reply)
					var res ai.ClassifierResult
					var err error
					if host == "local" {
						res, err = client.Classify(context.Background(), ai.CallScope{TenantID: tenantA.tenant, RequestID: requestid.New()}, ai.Target{BindingID: b.BindingID, ModelID: model.ID}, input, options)
					} else {
						h := hostintegration.New(client, hostintegration.NewMemoryStore(), requestid.New)
						res, err = h.Classify(context.Background(), hostintegration.Principal{TenantID: tenantA.tenant}, hostintegration.ClassificationRequest{Target: ai.Target{BindingID: b.BindingID, ModelID: model.ID}, Input: input, Options: options})
					}
					ev.Record("result", res)
					ev.Record("requests", w.provider.Requests())
					ev.Record("error", errString(err))
					invalid := scenario == "wrong-options" || (scenario == "structured" && route == "luna") || scenario == "no-image-binding" || scenario == "no-image-policy" || scenario == "no-image-model"
					if invalid {
						ev.Check("incompatible input never sends", err != nil && len(w.provider.Requests()) == 0, "got %v", err)
						var ae *ai.Error
						ev.Check("typed failure retained", errors.As(err, &ae), "untyped failure")
						return
					}
					if !ev.Check("host closure succeeds", err == nil && len(w.provider.Requests()) == 1, "got %v", err) {
						return
					}
					// The host makes a closed branch before touching an ordinary answer.
					state := ""
					switch a := res.Answers["intent"].(type) {
					case ai.ChoiceAnswer:
						state = "choice"
						ev.Check("choice view", a.Choice == "billing", "wrong choice")
					case ai.ScoreAnswer:
						state = "score"
					case ai.BoolAnswer:
						state = "bool"
					case ai.RefusalAnswer:
						state = "refused"
					default:
						t.Fatal("missing answer")
					}
					want := "choice"
					if scenario == "refusal" {
						want = "refused"
					}
					ev.Check("nil error does not assert ordinary answer", state == want, "state=%s", state)
				})
			}
		}
	}
}
