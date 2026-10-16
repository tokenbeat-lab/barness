package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var multimodalRoutes = []mixedRoute{
	{"classifier", ai.OperationClassifier, ai.ProviderTypeSafe, ai.APITypeSafeSystemOne},
	{"luna", ai.OperationClassifier, ai.ProviderOpenAI, ai.APIOpenAIDecisions},
}

func multimodalModel(r mixedRoute) string {
	if r.api == ai.APITypeSafeSystemOne {
		return "jev-1.14.0"
	}
	return "host-decisions"
}

func newMultimodalWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	u := newJevBoxWorld(t, ev, func(c *ai.Config) {
		c.Catalog.Version = "host-mixed-images-runtime-09"
		m := c.Catalog.ClassifierModels[0]
		m.ID, m.Provider, m.API = "host-decisions", ai.ProviderOpenAI, ai.APIOpenAIDecisions
		m.Capabilities = ai.ClassifierCapabilities{Kinds: []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}, MinChoices: 2, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10, DescriptionFormats: []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}, ImageLocations: []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}, ImageMIMETypes: []string{"image/png", "image/jpeg"}, MaxInputImages: 8}
		m.Cost.Input = 2
		c.Catalog.ClassifierModels = append(c.Catalog.ClassifierModels, m)
		// Finite contract shared by normal and race. Four in-flight calls,
		// two per tenant, two files and up to eight occurrences per call.
		c.Policy.MaxRequestBytes, c.Policy.MaxOutputBytes = 64<<10, 16<<10
		c.Policy.MaxImageBytes = 32 << 10
		c.Policy.MaxToolJSONBytes = 16 << 10
		c.Policy.Classifier.Images.MaxImages = 2
		c.Policy.Classifier.Images.MaxTotalImageBytes = 64 << 10
		c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 128 << 10}
		c.Policy.MaxConcurrentProcess, c.Policy.MaxConcurrentPerTenant = 4, 2
		c.Policy.MaxQueuedObservations = 256
		if change != nil {
			change(c)
		}
	})
	for _, k := range []tenantKey{tenantA, tenantB} {
		for _, r := range multimodalRoutes {
			b := r.binding(k, u.provider.URL())
			b.AllowedModels = []string{multimodalModel(r)}
			b.AllowClassifierImages = true
			u.host.PutBinding(b)
			u.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(k.secret)})
			ev.Record("binding-"+k.tenant+"-"+r.id, b)
		}
	}
	return u
}

func multimodalInput(t *testing.T, r mixedRoute, k tenantKey) ai.ClassifierRequest {
	req := decisionsMixedInput()
	req.Images = jevContextImages(t)
	if k == tenantB {
		req.Images[0], req.Images[1] = req.Images[1], req.Images[0]
	}
	req.State = json.RawMessage(`{"owner":"` + k.tenant + `","photo":null}`)
	for _, i := range []int{1, 0, 1} {
		req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: i, Placement: ai.ImagePlacementContext})
	}
	if r.api == ai.APITypeSafeSystemOne {
		req.Questions["location"] = ai.BoxQuestion{Instructions: json.RawMessage(`"locate the mug"`)}
		req.Questions["approved"] = ai.BoolQuestion{Instructions: json.RawMessage(`["compare",null]`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`"present"`), False: json.RawMessage(`"absent"`)}}
		req.Questions["urgency"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"visual rubric"`), Criteria: []json.RawMessage{json.RawMessage(`null`), json.RawMessage(`"different"`)}}
		req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}, ai.ImageReference{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/approved/instructions/1"}, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/urgency/criteria/0"})
	}
	return req
}

func multimodalCall(ctx context.Context, u *unaryWorld, r mixedRoute, k tenantKey, entry, id string, req ai.ClassifierRequest, hooks ai.Hooks) (ai.ClassifierResult, error) {
	scope := textScope(id)
	scope.TenantID = k.tenant
	target := ai.Target{BindingID: r.id, ModelID: multimodalModel(r)}
	var opts ai.Options
	if r.api == ai.APITypeSafeSystemOne {
		opts = jevContextOptions()
	}
	if entry == "plain" {
		return u.client.Classify(ctx, scope, target, req, opts)
	}
	return u.client.WithHooks(hooks).Classify(ctx, scope, target, req, opts)
}

// Jev legend must echo the actual final native rubric, whose random keys belong
// to this call. Dynamic responses and their hashes are recorded before writing.
func multimodalReply(t *testing.T, ev *evidence.Case, r mixedRoute, broken bool) provider.Reply {
	chunks := make([][]byte, 1)
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "up-mixed", "x-typesafe-request-id": "up-mixed"}, Chunks: chunks, OnRequest: func(capture provider.Request) {
		body := []byte(`{"answers":[` + decisionsAnswer + `,` + decisionsScore + `,` + decisionsPredicate + `],` + decisionsFullUsage + `}`)
		if r.api == ai.APITypeSafeSystemOne {
			var native struct {
				Questions map[string]struct{ Criteria json.RawMessage }
			}
			mustUnmarshal(t, capture.Body, &native)
			var levels []json.RawMessage
			mustUnmarshal(t, native.Questions["urgency"].Criteria, &levels)
			body = mustMarshal(t, map[string]any{"model": "jev-1.14.0", "answers": map[string]any{"approved": map[string]any{"type": "noul", "noul": 0}, "intent": map[string]any{"type": "choice", "choice": "track", "confidence": .7, "probabilities": map[string]float64{"track": .9, "other": .1}}, "location": json.RawMessage(jevBoxAnswer), "urgency": map[string]any{"type": "score", "score": 1, "confidence": 1, "probabilities": map[string]float64{"0": 0, "1": 1}, "legend": map[string]json.RawMessage{"0": levels[0], "1": levels[1]}}}, "usage": map[string]int{"input_tokens": 100, "output_tokens": 20}})
		}
		if broken {
			body = []byte(strings.Replace(string(body), `"answers":`, `"answers":null,"invalid":`, 1))
		}
		chunks[0] = body
		ev.Record(fmt.Sprintf("response-%s-%x", r.id, sha256.Sum256(capture.Body)), json.RawMessage(body))
	}}
}

func multimodalEnqueue(ev *evidence.Case, u *unaryWorld, r mixedRoute, k tenantKey, replies ...provider.Reply) {
	for i, reply := range replies {
		ev.Record(fmt.Sprintf("script-%s-%s-%d-%d", k.tenant, r.id, reply.Status, i), reply)
	}
	u.provider.EnqueueAt(r.prefix(k), replies...)
}

func recordMultimodalCall(ev *evidence.Case, u *unaryWorld, req ai.ClassifierRequest, res ai.ClassifierResult, err error) {
	id := res.Metadata.RequestID
	ev.Record(id+"-input", req)
	ev.Record(id+"-result", res)
	ev.Record(id+"-host-reads", u.host.ReadsFor(id))
	for i, img := range req.Images {
		raw, e := base64.StdEncoding.Strict().DecodeString(img.Data)
		if e != nil {
			ev.T().Fatal(e)
		}
		ev.Fixture(fmt.Sprintf("%s-image-%d.bin", id, i), raw)
	}
	w := mixedWorld{rec: u.rec}
	w.record(ev, mixedOutcome{metadata: res.Metadata, usage: res.Usage, classifier: res, err: err})
	observations := u.rec.forRequest(id)
	for _, o := range observations {
		if o.Kind == ai.ObservationCallFinished {
			var failure *ai.Error
			expected := errors.As(err, &failure)
			ev.Check("Observer terminal classification", o.StopReason == res.StopReason && ((expected && o.Error != nil && o.Error.Code == failure.Code && o.Error.Phase == failure.Phase) || (!expected && o.Error == nil)), "terminal changed")
		}
	}
	obs := observationText(ev.T(), observations)
	for _, marker := range []string{"data_base64", "image_url", "instructions", "criteria", "legend", "confidence", `"top":`, "/state/photo", "synthetic private error", tenantA.secret, tenantB.secret} {
		ev.Check("Observer excludes content "+marker, !strings.Contains(obs, marker), "content leaked")
	}
	for _, img := range req.Images {
		ev.Check("Observer excludes file", !strings.Contains(obs, img.Data), "file leaked")
	}
	for _, img := range res.ImageInputs {
		ev.Check("Observer excludes hash", !strings.Contains(obs, img.SHA256), "hash leaked")
	}
	for _, capture := range u.provider.Requests() {
		var native struct{ Assets map[string]json.RawMessage }
		_ = json.Unmarshal(capture.Body, &native)
		for key := range native.Assets {
			ev.Check("Observer excludes private key", !strings.Contains(obs, key), "key leaked")
		}
	}
	ev.Record("requests", u.provider.Requests())
}
