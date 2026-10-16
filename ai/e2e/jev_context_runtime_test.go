package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestJevContextResponseCancellation(t *testing.T) {
	for _, name := range []string{"cancel", "deadline", "revoke"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-runtime-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, nil)
				req := jevContextInput(t)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				if name == "deadline" {
					var end context.CancelFunc
					ctx, end = context.WithTimeout(ctx, 200*time.Millisecond)
					defer end()
				}
				reply := fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, "")
				if name != "revoke" {
					reply.Chunks = [][]byte{[]byte(`{"answers":`)}
					reply.Script = `{"answers":`
					reply.End = provider.EndHold
					if name == "cancel" {
						reply.OnHold = cancel
					}
				} else {
					reply.OnReceive = func() {
						u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false; b.Version = "revoked" })
					}
				}
				enqueue(ev, u.world, reply)
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctx, textScope("req-jev-runtime-"+name+entry), jevContextTarget(), req, jevContextOptions())
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctx, textScope("req-jev-runtime-"+name+entry), jevContextTarget(), req, jevContextOptions())
				}
				if name == "revoke" {
					ev.Check("pinned image authorization completes", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					code := ai.CodeCanceled
					if name == "deadline" {
						code = ai.CodeDeadlineExceeded
					}
					ev.Check("response failure is atomic", errors.Is(err, &ai.Error{Code: code}) && res.Answers == nil && res.ImageInputs == nil, "got %v", err)
				}
				ev.Check("exactly one Attempt, no implicit retry", len(res.Metadata.Attempts) == 1, "unexpected attempts")
				recordJevContext(ev, u, req, res, err)
				if name == "revoke" {
					next, err := u.client.Classify(ctxFor(t), textScope("req-jev-revoked-"+entry), jevContextTarget(), req, jevContextOptions())
					ev.Check("new call sees revoked image permission", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseCapability}) && len(next.Metadata.Attempts) == 0, "got %v", err)
					ev.Record("next-result", next)
					u.records(ev, next, err)
					u.released(ev)
				}
			})
		}
	}
}

func TestJevContextCandidateContract(t *testing.T) {
	ev := run.Case(t, "UC04-candidate-contract")
	u := newJevContextWorld(t, ev, nil)
	for _, name := range []string{"alias", "old", "webp", "local", "duplicate", "omitted"} {
		cat := u.client.Catalog()
		m := &cat.ClassifierModels[0]
		switch name {
		case "alias":
			m.ID = "jev-latest"
		case "old":
			m.ID = "jev-1.13.0"
		case "webp":
			m.Capabilities.ImageMIMETypes = append(m.Capabilities.ImageMIMETypes, "image/webp")
		case "local":
			m.Capabilities.ImageLocations = append(m.Capabilities.ImageLocations, ai.ClassifierImageLocationCriteria)
		case "duplicate":
			m.Capabilities.ImageLocations = append(m.Capabilities.ImageLocations, ai.ClassifierImageLocationContext)
		case "omitted":
			m.Capabilities.ImageLocations = nil
		}
		p := validPolicy()
		_, err := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: u.host, Credentials: u.host})
		ev.Check("closed explicit candidate "+name, err != nil, "catalog accepted unsupported capability")
	}
	catalog := u.client.Catalog()
	lookup, _ := catalog.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.14.0")
	lookup.Capabilities.ImageLocations[0] = ai.ClassifierImageLocationCriteria
	ev.Check("candidate catalog queries own slices", u.client.Catalog().ClassifierModels[0].Capabilities.ImageLocations[0] == ai.ClassifierImageLocationContext, "mutated catalog")
	builtin := ai.BuiltinCatalog()
	hash, _ := builtin.Hash()
	ev.Check("builtin hash unchanged", hash == builtinCatalogHash, "builtin changed")
	ev.Record("catalog", u.client.Catalog())
	ev.Record("builtin-hash", hash)
}
