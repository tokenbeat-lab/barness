package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevFormalAuthorization(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		for _, kind := range []string{"permission", "whitelist", "experimental", "policy", "count", "mime", "apng", "dimensions"} {
			t.Run(entry+"/"+kind, func(t *testing.T) {
				ev := run.Case(t, "UC13-formal-guard-"+entry+"-"+kind)
				u := newJevBoxWorld(t, ev, func(c *ai.Config) {
					*c.Catalog = ai.BuiltinCatalog()
					if kind == "policy" {
						c.Policy.Classifier.Images = nil
					}
					if kind == "count" {
						c.Policy.Classifier.Images.MaxImages = 32
					}
				})
				req := jevBoxInput(t)
				opts := jevContextOptions()
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				switch kind {
				case "permission":
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				case "whitelist":
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"jev-1.13.0"} })
				case "experimental":
					opts = ai.TypeSafeOptions{}
				case "policy":
					phase = ai.PhaseScope
				case "count":
					for len(req.Images) < 25 {
						req.Images = append(req.Images, req.Images[0])
					}
				case "mime":
					req.Images[0] = classifierAssets(t)[2]
				case "apng":
					req.Images[0] = jevAPNG(t)
				case "dimensions":
					req.Images[0] = jevPNG(t, 256, 256)
				}
				scope := textScope("req-formal-guard-" + entry + kind)
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, jevContextTarget(), req, opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), scope, jevContextTarget(), req, opts)
				}
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				ev.Check("formal discovery grants no permission or larger capability", u.host.ReadsFor(scope.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unauthorized send")
				if kind == "count" {
					// All 25 declared sources intentionally contain the same fixture.
					// The older recorder names three distinct fixtures only.
					ev.Record("input", req)
					recordClassifierAssets(ev, req.Images[:1])
					recordImageInputsCall(ev, u, res, err)
					ev.Record("host-reads", u.host.ReadsFor(scope.RequestID))
				} else {
					recordJevContext(ev, u, req, res, err)
				}
			})
		}
	}
}
