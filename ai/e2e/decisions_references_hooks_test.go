package e2e

import (
	"context"
	"encoding/base64"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsReferencesHooks(t *testing.T) {
	assets := classifierAssets(t)
	raw, _ := base64.StdEncoding.DecodeString(assets[0].Data)
	for _, name := range []string{"reorder", "add", "replace", "denied", "policy", "count", "bytes", "custom-count", "custom-bytes", "model-count", "bad-order"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC02-hook-"+name)
			success := name == "reorder" || name == "add" || name == "replace"
			code := ai.CodeResourceLimit
			u := newReferenceWorld(t, ev, func(c *ai.Config) {
				setReferencePolicy(t, c, 4, 1<<19)
				switch name {
				case "count", "custom-count":
					setReferencePolicy(t, c, 1, 1<<19)
				case "bytes", "custom-bytes":
					setReferencePolicy(t, c, 4, int64(len(raw)))
				case "model-count":
					c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.MaxInputImages = 1
				case "policy":
					c.Policy.Classifier.Images = nil
				}
			})
			if name == "denied" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
			}
			if name == "policy" || name == "denied" {
				code = ai.CodeTenantDenied
			}
			if name == "bad-order" {
				code = ai.CodeCallbackFailed
			}
			req := decisionsInput()
			expected := []ai.Image{assets[0], assets[0]}
			// Reorder a real reference expansion; other cases add native blocks to a
			// text request, proving the Hook cannot rely on initial-request authority.
			if name == "reorder" {
				req.Images = assets[:2]
				req = decisionsReferenceRequest(t, req, `[{"imageIndex":1,"placement":"context"},{"imageIndex":0,"placement":"context"},{"imageIndex":1,"placement":"context"}]`)
				expected = []ai.Image{assets[0], assets[1], assets[1]}
			}
			ev.Fixture("public-request.json", mustMarshal(t, req))
			recordClassifierAssets(ev, assets)
			if success {
				enqueueClassifierImages(t, ev, u)
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				messages := classifierImageMessages(expected)
				if name == "bad-order" {
					content := messages[0].(map[string]any)["content"].([]any)
					content[0], content[1] = content[1], content[0]
				}
				if name == "custom-count" || name == "custom-bytes" {
					p.Body["input"] = decisionsRaw(mustMarshal(t, messages))
				} else {
					p.Body["input"] = messages
				}
				if name == "replace" {
					return ai.ReplacePayload(p.Body), nil
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-reference-hook-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			if success {
				ev.Check("authorized final image edits succeed", err == nil, "got %v", err)
				requests := u.provider.Requests()
				if ev.Check("one request", len(requests) == 1, "got %d", len(requests)) {
					var final map[string]any
					mustUnmarshal(t, requests[0].Body, &final)
					ev.Check("final submitted image order matches Hook", jsonEqual(mustMarshal(t, final["input"]), mustMarshal(t, classifierImageMessages(expected))), "wrong final input")
				}
			} else {
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("final recheck prevents any Attempt", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			}
			ev.Record("result", res)
			ev.Record("error", errString(err))
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsReferencesFrozenCall(t *testing.T) {
	ev := run.Case(t, "UC02-frozen-call")
	req := decisionsInput()
	assets := classifierAssets(t)
	req.Images = assets[:2]
	req = decisionsReferenceRequest(t, req, `[{"imageIndex":1,"placement":"context"},{"imageIndex":0,"placement":"context"},{"imageIndex":1,"placement":"context"}]`)
	original := []ai.Image{req.Images[1], req.Images[0], req.Images[1]}
	var mutate func()
	u := newReferenceWorld(t, ev, func(c *ai.Config) {
		previous := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
			if mutate != nil {
				mutate()
			}
			return previous.ResolveBinding(ctx, s, id)
		})
	})
	ev.Fixture("public-request.json", mustMarshal(t, req))
	mutate = func() {
		req.ImageReferences[0].ImageIndex = 0
		req.ImageReferences[1].Placement = ai.ImagePlacementValue
		req.ImageReferences[1].Pointer = "/state"
		req.Images[0] = ai.Image{Data: "changed", MimeType: "image/gif"}
		req.State[1] = 'X'
		req.Questions["intent"] = nil
	}
	fail := fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":{"message":"limited"}}`}.script(t, decisionsProtocol, "")
	enqueue(ev, u.world, fail, runtimeDecisionsReply(t))
	calls := [2]int{}
	res, err := u.client.WithHooks(ai.Hooks{
		TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { calls[0]++; return nil },
		OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
			calls[1]++
			return ai.KeepPayload(), nil
		},
	}).Classify(ctxFor(t), textScope("req-reference-frozen"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	captures := u.provider.Requests()
	ev.Check("copied public reference plan survives resolver mutation and retry", err == nil && len(captures) == 2 && len(res.Metadata.Attempts) == 2 && calls == [2]int{1, 1}, "got %v", err)
	checkImageInputs(ev, res, original)
	if len(captures) == 2 {
		ev.Check("retry freezes native bytes and credentials", reflect.DeepEqual(captures[0], captures[1]), "retry changed request")
		var final map[string]any
		mustUnmarshal(t, captures[0].Body, &final)
		ev.Check("original references and sources remain ordered", jsonEqual(mustMarshal(t, final["input"]), mustMarshal(t, classifierImageMessages(original))), "input mutated")
	}
	ev.Record("result", res)
	ev.Record("error", errString(err))
	recordRuntimeConfig(ev, u)
	u.records(ev, res, err)
	u.released(ev)
}
