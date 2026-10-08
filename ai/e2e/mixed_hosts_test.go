package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

func TestMixedHosts(t *testing.T) {
	for _, name := range []string{"local", "cloud"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "E10-mixed-host-"+name)
			w := newMixedWorld(t, nil, tenantA)
			client := w.client
			var local *localassembly.Operations
			if name == "local" {
				t.Setenv("BARNESS_MIXED_KEY", tenantA.secret)
				cfg := localassembly.OperationsConfig{TenantID: tenantA.tenant, Policy: localassembly.MixedPolicy(), Transport: w.bodies, AllowLoopbackHTTP: true, Catalog: ptrMixedCatalog(), Keys: map[string]localassembly.KeySource{}}
				for _, r := range mixedRoutes {
					b := r.binding(tenantA, w.provider.URL())
					cfg.Bindings = append(cfg.Bindings, b)
					cfg.Keys[b.CredentialRef] = localassembly.KeySource{EnvVar: "BARNESS_MIXED_KEY"}
				}
				var err error
				local, err = localassembly.OpenOperations(cfg)
				if err != nil {
					t.Fatal(err)
				}
				client = local.Client
				// Resolver storage must own every original Binding slice.
				for i := range cfg.Bindings {
					cfg.Bindings[i].AllowedModels[0] = "mutated"
				}
			}
			h := hostintegration.New(client, hostintegration.NewMemoryStore(), requestid.New)
			who := hostintegration.Principal{TenantID: tenantA.tenant, ActorID: "authenticated"}
			results := make([]mixedOutcome, len(mixedRoutes))
			var wg sync.WaitGroup
			for i, r := range mixedRoutes {
				w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
				wg.Go(func() {
					target := ai.Target{BindingID: r.id, ModelID: mixedModel}
					if local != nil {
						results[i] = r.invoke(ctxFor(t), client, local.Scope(), target)
						return
					}
					switch r.op {
					case ai.OperationChat:
						id, err := h.OpenSession(ctxFor(t), who)
						if err != nil {
							t.Error(err)
							return
						}
						res, err := h.Turn(ctxFor(t), who, hostintegration.TurnRequest{SessionID: id, BindingID: r.id, ModelID: mixedModel, Prompt: "synthetic"}, discardMixedSink{})
						results[i] = mixedOutcome{metadata: res.Metadata, usage: res.Message.Usage, chat: res, err: err}
					case ai.OperationImage:
						res, err := h.GenerateImages(ctxFor(t), who, hostintegration.ImageRequest{Target: target, Input: ai.ImagesRequest{Prompt: "synthetic"}})
						results[i] = mixedOutcome{metadata: res.Metadata, usage: res.Usage, images: res, err: err}
					case ai.OperationClassifier:
						res, err := h.Classify(ctxFor(t), who, hostintegration.ClassificationRequest{Target: target, Input: mixedClassifierInput()})
						results[i] = mixedOutcome{metadata: res.Metadata, usage: res.Usage, classifier: res, err: err}
					}
				})
			}
			wg.Wait()
			ids := map[string]bool{}
			for i, o := range results {
				o.recordResult(ev, mixedRoutes[i].id+"-result")
				ev.Check("host returns typed attributed result", o.err == nil && o.metadata.Resolved && o.metadata.Operation == mixedRoutes[i].op && o.metadata.TenantID == who.TenantID, "got %+v / %v", o.metadata, o.err)
				ev.Check("host mints unique call id", o.metadata.RequestID != "" && !ids[o.metadata.RequestID], "id %s", o.metadata.RequestID)
				ids[o.metadata.RequestID] = true
				if name == "cloud" {
					w.record(ev, o)
				}
			}
			if name == "cloud" {
				bad, err := h.Classify(ctxFor(t), who, hostintegration.ClassificationRequest{TenantID: tenantB.tenant, Target: ai.Target{BindingID: "classifier", ModelID: mixedModel}, Input: mixedClassifierInput()})
				ev.Check("host refuses claimed tenant before Client", errors.Is(err, hostintegration.ErrForbidden) && bad.Metadata.RequestID == "", "got %v", err)
				ctx, cancel := context.WithCancel(ctxFor(t))
				cancel()
				res, err := h.GenerateImages(ctx, who, hostintegration.ImageRequest{Target: ai.Target{BindingID: "openai-image", ModelID: mixedModel}, Input: ai.ImagesRequest{Prompt: "synthetic"}})
				ev.Check("downstream cancellation preserved", errors.Is(err, &ai.Error{Code: ai.CodeCanceled}) && len(res.Content) == 0, "got %v", err)
			}
			ev.Record("requests", w.provider.Requests())
			ev.Check("only four intended calls sent", len(w.provider.Requests()) == 4, "got %d", len(w.provider.Requests()))
			w.released(ev)
		})
	}
}

func ptrMixedCatalog() *ai.Catalog { c := mixedCatalog(); return &c }

type discardMixedSink struct{}

func (discardMixedSink) Send(ai.EventEnvelope) error { return nil }
