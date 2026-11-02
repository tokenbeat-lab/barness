package e2e

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
)

// Failure modes: a same-capability synthetic name is rejected; the selected
// file, typed hash, and Client disagree; a route loses the exact identity or
// Usage; an unauthorized Binding sends HTTP; invalid capacity, modality,
// price, protocol, empty ID, or duplicate full identity becomes accepted.
// These names are synthetic and make no assertion about a vendor offering.
func futureCatalogCases(t *testing.T) []struct {
	name     string
	cat      ai.Catalog
	identity catalogContractIdentity
} {
	t.Helper()
	builtin := ai.BuiltinCatalog()
	cases := []struct {
		name     string
		cat      ai.Catalog
		identity catalogContractIdentity
	}{}
	for _, p := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			m, ok := builtin.Lookup(p, api, "kimi-k3.1-pro")
			if !ok {
				t.Fatal("missing Kimi source model")
			}
			m.ID = "synthetic-future-kimi"
			cases = append(cases, struct {
				name     string
				cat      ai.Catalog
				identity catalogContractIdentity
			}{string(p) + "-" + string(api), ai.Catalog{Version: "synthetic-future", Models: []ai.Model{m}}, catalogContractIdentity{ai.OperationChat, p, api, m.ID}})
		}
	}
	chat, ok := builtin.Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, "grok-4.7")
	if !ok {
		t.Fatal("missing xAI chat source")
	}
	chat.ID = "synthetic-future-xai-chat"
	cases = append(cases, struct {
		name     string
		cat      ai.Catalog
		identity catalogContractIdentity
	}{"xai-chat", ai.Catalog{Version: "synthetic-future", Models: []ai.Model{chat}}, catalogContractIdentity{ai.OperationChat, chat.Provider, chat.API, chat.ID}})
	image, ok := builtin.LookupImage(ai.ProviderXAI, ai.APIXAIImages, "grok-imagine-image-2.0")
	if !ok {
		t.Fatal("missing xAI image source")
	}
	image.ID = "synthetic-future-xai-image"
	cases = append(cases, struct {
		name     string
		cat      ai.Catalog
		identity catalogContractIdentity
	}{"xai-image", ai.Catalog{Version: "synthetic-future", Models: []ai.Model{}, ImageModels: []ai.ImageModel{image}}, catalogContractIdentity{ai.OperationImage, image.Provider, image.API, image.ID}})
	classifier, ok := builtin.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.14.0")
	if !ok {
		t.Fatal("missing classifier source")
	}
	classifier.ID = "synthetic-future-classifier"
	cases = append(cases, struct {
		name     string
		cat      ai.Catalog
		identity catalogContractIdentity
	}{"typesafe-image-box", ai.Catalog{Version: "synthetic-future", Models: []ai.Model{}, ClassifierModels: []ai.ClassifierModel{classifier}}, catalogContractIdentity{ai.OperationClassifier, classifier.Provider, classifier.API, classifier.ID}})
	return cases
}

func TestCatalogFutureModelPublicContract(t *testing.T) {
	for _, tc := range futureCatalogCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "FUTURE-MODEL-load-"+tc.name)
			path := filepath.Join(t.TempDir(), "synthetic.json")
			raw := localCatalogFile(string(mustMarshal(t, tc.cat)))
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			loaded, report, loadErr := catalog.LoadFile(path)
			hash, hashErr := tc.cat.Hash()
			w := newWorld(t, tenantA)
			client, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &tc.cat})
			ev.Record("source", tc.cat)
			ev.Record("load-report", report)
			ev.Check("direct catalog hash accepts supported synthetic identity", hashErr == nil && hash != "", "err=%v", hashErr)
			ev.Check("LoadFile accepts exactly the selected synthetic definition", loadErr == nil && report.Loaded == 1 && report.Failed == 0 && reflect.DeepEqual(loaded, tc.cat), "err=%v report=%+v", loadErr, report)
			if !ev.Check("public Client accepts supported synthetic identity", clientErr == nil, "err=%v", clientErr) {
				return
			}
			ev.Check("client retains exact catalog identity", reflect.DeepEqual(client.Catalog(), tc.cat), "catalog differs")
			var scenes []string
			switch tc.identity.Operation {
			case ai.OperationChat:
				scenes = []string{"basic", "cache-usage", "reasoning", "authorization"}
			case ai.OperationImage:
				scenes = []string{"generation", "authorization"}
			case ai.OperationClassifier:
				scenes = []string{"questions", "image-locations", "authorization"}
			}
			if tc.identity.Operation == ai.OperationChat {
				t.Run("response-mismatch", func(t *testing.T) {
					mismatch := run.Case(t, "FUTURE-MODEL-"+tc.name+"-response-mismatch")
					model, _ := tc.cat.Lookup(tc.identity.Provider, tc.identity.API, tc.identity.ID)
					world, _ := newCatalogContractWorld(t, tc.cat, tc.identity)
					reply := catalogChatReply(t, model, 8, false)
					reply.Script = strings.ReplaceAll(reply.Script, tc.identity.ID, "synthetic-wrong-model")
					reply.Chunks = [][]byte{[]byte(reply.Script)}
					enqueue(mismatch, world, reply)
					result, err := world.client.CompleteSimple(ctxFor(t), textScope("future-mismatch"), catalogContractTarget(tc.identity), ai.Request{Messages: []ai.Message{ai.UserText("Controlled input")}}, ai.SimpleOptions{MaxTokens: ai.Value(64)})
					if (tc.identity.Provider == ai.ProviderMoonshot || tc.identity.Provider == ai.ProviderMoonshotCN) && tc.identity.API != ai.APIAnthropicMessages {
						reported := "synthetic-wrong-model"
						if tc.identity.API == ai.APIOpenAIResponses {
							reported = ""
						}
						mismatch.Check("target and protocol-specific reported model remain separate", err == nil && result.Message.Model == tc.identity.ID && result.Message.ResponseModel == reported && result.Metadata.ModelID == tc.identity.ID, "err=%v result=%+v", err, result)
					} else {
						mismatch.Check("strict protocol refuses response model mismatch", err != nil, "err=%v", err)
					}
				})
			}
			t.Run("missing-target", func(t *testing.T) {
				e := run.Case(t, "FUTURE-MODEL-"+tc.name+"-missing-target")
				world, binding := newCatalogContractWorld(t, tc.cat, tc.identity)
				binding.AllowedModels = append(binding.AllowedModels, "synthetic-not-in-catalog")
				world.host.PutBinding(binding)
				target := ai.Target{BindingID: binding.BindingID, ModelID: "synthetic-not-in-catalog"}
				var err error
				switch tc.identity.Operation {
				case ai.OperationChat:
					_, err = world.client.CompleteSimple(ctxFor(t), textScope("future-missing"), target, ai.Request{Messages: []ai.Message{ai.UserText("Controlled input")}}, ai.SimpleOptions{})
				case ai.OperationImage:
					_, err = world.client.GenerateImages(ctxFor(t), textScope("future-missing"), target, ai.ImagesRequest{Prompt: "Controlled input"}, nil)
				case ai.OperationClassifier:
					_, err = world.client.Classify(ctxFor(t), textScope("future-missing"), target, ai.ClassifierRequest{}, nil)
				}
				e.Check("catalog missing target is refused without HTTP", err != nil && len(world.provider.Requests()) == 0, "err=%v", err)
			})
			if tc.identity.Operation == ai.OperationChat {
				t.Run("hook-model", func(t *testing.T) {
					e := run.Case(t, "FUTURE-MODEL-"+tc.name+"-hook-model")
					world, _ := newCatalogContractWorld(t, tc.cat, tc.identity)
					hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						p.Body["model"] = "synthetic-unauthorized"
						return ai.KeepPayload(), nil
					}}
					_, err := world.client.WithHooks(hooks).CompleteSimple(ctxFor(t), textScope("future-hook"), catalogContractTarget(tc.identity), ai.Request{Messages: []ai.Message{ai.UserText("Controlled input")}}, ai.SimpleOptions{})
					e.Check("hook cannot change authorized model or send HTTP", err != nil && len(world.provider.Requests()) == 0, "err=%v", err)
				})
			}
			for _, scene := range scenes {
				t.Run(scene, func(t *testing.T) {
					e := run.Case(t, "FUTURE-MODEL-"+tc.name+"-"+scene)
					switch tc.identity.Operation {
					case ai.OperationChat:
						catalogChatContract(t, e, tc.cat, tc.identity, scene)
					case ai.OperationImage:
						catalogImageContract(t, e, tc.cat, tc.identity, scene)
					case ai.OperationClassifier:
						catalogClassifierContract(t, e, tc.cat, tc.identity, scene)
					}
				})
			}
		})
	}
}

func TestCatalogFutureModelIdentityRefusals(t *testing.T) {
	for _, tc := range futureCatalogCases(t) {
		for _, kind := range []string{"empty-id", "duplicate"} {
			t.Run(tc.name+"/"+kind, func(t *testing.T) {
				ev := run.Case(t, "FUTURE-MODEL-reject-"+tc.name+"-"+kind)
				cat := tc.cat
				switch tc.identity.Operation {
				case ai.OperationChat:
					cat.Models = append([]ai.Model(nil), cat.Models...)
					if kind == "empty-id" {
						cat.Models[0].ID = ""
					} else {
						cat.Models = append(cat.Models, cat.Models[0])
					}
				case ai.OperationImage:
					cat.ImageModels = append([]ai.ImageModel(nil), cat.ImageModels...)
					if kind == "empty-id" {
						cat.ImageModels[0].ID = ""
					} else {
						cat.ImageModels = append(cat.ImageModels, cat.ImageModels[0])
					}
				case ai.OperationClassifier:
					cat.ClassifierModels = append([]ai.ClassifierModel(nil), cat.ClassifierModels...)
					if kind == "empty-id" {
						cat.ClassifierModels[0].ID = ""
					} else {
						cat.ClassifierModels = append(cat.ClassifierModels, cat.ClassifierModels[0])
					}
				}
				path := filepath.Join(t.TempDir(), "invalid.json")
				os.WriteFile(path, localCatalogFile(string(mustMarshal(t, cat))), 0600)
				_, report, loadErr := catalog.LoadFile(path)
				_, hashErr := cat.Hash()
				w := newWorld(t, tenantA)
				_, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
				ev.Check("LoadFile refuses invalid identity", loadErr == nil && report.Loaded == 0 && report.Failed > 0, "err=%v report=%+v", loadErr, report)
				ev.Check("direct Hash refuses invalid identity", hashErr != nil, "err=%v", hashErr)
				ev.Check("direct Client refuses invalid identity", clientErr != nil, "err=%v", clientErr)
			})
		}
	}
}

func TestCatalogFutureModelCapabilityRefusals(t *testing.T) {
	for _, tc := range futureCatalogCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "FUTURE-MODEL-capability-refusal-"+tc.name)
			cat := tc.cat
			switch tc.identity.Operation {
			case ai.OperationChat:
				cat.Models = append([]ai.Model(nil), cat.Models...)
				cat.Models[0].ContextWindow++
			case ai.OperationImage:
				cat.ImageModels = append([]ai.ImageModel(nil), cat.ImageModels...)
				cat.ImageModels[0].Capabilities.Mask = true
			case ai.OperationClassifier:
				cat.ClassifierModels = append([]ai.ClassifierModel(nil), cat.ClassifierModels...)
				cat.ClassifierModels[0].Capabilities.ImageMIMETypes = []string{"image/webp"}
			}
			path := filepath.Join(t.TempDir(), "invalid-capability.json")
			os.WriteFile(path, localCatalogFile(string(mustMarshal(t, cat))), 0600)
			_, report, err := catalog.LoadFile(path)
			_, hashErr := cat.Hash()
			w := newWorld(t, tenantA)
			_, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Check("unsupported capability remains refused by loader", err == nil && report.Loaded == 0 && report.Failed == 1, "err=%v report=%+v", err, report)
			ev.Check("unsupported capability remains refused by direct catalog and Client", hashErr != nil && clientErr != nil, "hash=%v client=%v", hashErr, clientErr)
		})
	}
}

// The same model name on distinct Providers or APIs is a distinct identity.
func TestCatalogFutureModelFullIdentity(t *testing.T) {
	ev := run.Case(t, "FUTURE-MODEL-distinct-identities")
	candidates := futureCatalogCases(t)
	first := candidates[0].cat.Models[0]
	second := candidates[1].cat.Models[0]
	third := candidates[3].cat.Models[0]
	cat := ai.Catalog{Version: "synthetic-identities", Models: []ai.Model{first, second, third}}
	path := filepath.Join(t.TempDir(), "identities.json")
	os.WriteFile(path, localCatalogFile(string(mustMarshal(t, cat))), 0600)
	loaded, report, err := catalog.LoadFile(path)
	_, hashErr := cat.Hash()
	w := newWorld(t, tenantA)
	_, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
	ev.Check("same ID across distinct provider and API identities loads", err == nil && report.Loaded == 3 && report.Failed == 0 && len(loaded.Models) == 3, "err=%v report=%+v", err, report)
	ev.Check("same ID across distinct provider and API identities hashes and configures", hashErr == nil && clientErr == nil, "hash=%v client=%v", hashErr, clientErr)
	for _, m := range cat.Models {
		got, ok := loaded.Lookup(m.Provider, m.API, m.ID)
		ev.Check("exact lookup preserves four-dimensional identity", ok && got.Provider == m.Provider && got.API == m.API && got.ID == m.ID, "model=%+v", m)
	}
}
