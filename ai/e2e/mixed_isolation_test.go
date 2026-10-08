package e2e

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMixedConcurrentIsolation(t *testing.T) {
	ev := run.Case(t, "E07-mixed-same-named-bindings")
	w := newMixedWorld(t, nil, tenantA, tenantB)
	var arrived sync.WaitGroup
	arrived.Add(8)
	ready := doneWhen(arrived.Wait)
	var wg sync.WaitGroup
	results := make([]mixedOutcome, 8)
	for n, k := range []tenantKey{tenantA, tenantB} {
		for i, r := range mixedRoutes {
			reply := r.tenantReply(t, k)
			upstream := "up-" + k.tenant + "-" + r.id
			if reply.Header == nil {
				reply.Header = map[string]string{}
			}
			reply.Header["x-request-id"] = upstream
			if r.op == ai.OperationClassifier {
				reply.Header["x-typesafe-request-id"] = upstream
			}
			reply.OnReceive = func() { arrived.Done(); waitFor(ev, "all mixed requests simultaneously in flight", ready) }
			w.provider.EnqueueAt(r.prefix(k), reply)
			wg.Go(func() {
				results[n*4+i] = r.invoke(ctxFor(t), w.client, ai.CallScope{TenantID: k.tenant, ActorID: "actor-" + k.tenant, JobID: "job-" + k.tenant, RequestID: "req-mixed-" + k.tenant + "-" + r.id}, ai.Target{BindingID: r.id, ModelID: mixedModel})
			})
		}
	}
	wg.Wait()
	for n, k := range []tenantKey{tenantA, tenantB} {
		for i, r := range mixedRoutes {
			o := results[n*4+i]
			o.recordResult(ev, o.metadata.RequestID+"-result")
			ev.Check("complete four-dimensional identity", o.err == nil && o.metadata.Resolved && o.metadata.TenantID == k.tenant && o.metadata.Operation == r.op && o.metadata.ProviderID == r.provider && o.metadata.API == r.api && o.metadata.ModelID == mixedModel && o.metadata.AccountScopeID == r.binding(k, "").AccountScopeID, "got %+v / %v", o.metadata, o.err)
			ev.Check("own attempt response and resolver snapshot", len(o.metadata.Attempts) == 1 && o.metadata.Attempts[0].ProviderRequestID == "up-"+k.tenant+"-"+r.id && w.host.ReadsFor(o.metadata.RequestID).Credentials == 1, "got %+v", o.metadata)
			inputTokens := int64(31)
			if k.tenant == tenantB.tenant {
				inputTokens = 47
			}
			ev.Check("own independent result and attempt usage", o.usage.Input == inputTokens && len(o.metadata.Attempts) == 1 && o.metadata.Attempts[0].Usage.Input == inputTokens, "got %+v", o.usage)
			if r.op == ai.OperationChat {
				ev.Check("own chat body", len(o.chat.Message.Content) == 1 && o.chat.Message.Content[0].(ai.Text).Text == "mixed-output-"+k.tenant, "got %+v", o.chat.Message.Content)
			}
			if r.op == ai.OperationImage {
				ev.Check("own image bytes", len(o.images.Content) == 1 && o.images.Content[0].(ai.ImageOutputImage).Data == mixedTenantImage(k), "wrong image for %s", k.tenant)
			}
			if r.op == ai.OperationClassifier {
				choice := "track"
				if k.tenant == tenantB.tenant {
					choice = "refund"
				}
				a, ok := o.classifier.Answers["intent"].(ai.ChoiceAnswer)
				ev.Check("own classifier answer", ok && a.Choice == choice, "got %+v", a)
			}
			w.record(ev, o)
		}
	}
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	ev.Check("eight actual requests", len(reqs) == 8, "got %d", len(reqs))
	for _, req := range reqs {
		k := tenantA
		if strings.HasPrefix(req.Path, "/"+tenantB.tenant+"/") {
			k = tenantB
		}
		ev.Check("key never crosses tenants", req.KeyAlias == k.alias, "path=%s key=%s", req.Path, req.KeyAlias)
		ev.Check("tenant-specific request body", strings.Contains(string(req.Body), k.tenant), "missing request marker")
		var body map[string]any
		mustUnmarshal(t, req.Body, &body)
		ev.Check("same model id stays within operation", body["model"] == mixedModel, "got %v", body["model"])
	}
	w.released(ev)
}

func TestMixedPreflight(t *testing.T) {
	for _, r := range mixedRoutes {
		for _, mode := range []string{"wrong-entry", "unknown-operation", "wrong-options", "zero-chat"} {
			t.Run(r.id+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "E07-mixed-preflight-"+r.id+"-"+mode)
				w := newMixedWorld(t, nil, tenantA)
				b := w.host.Binding(tenantA.tenant, r.id)
				entry := r
				code := ai.CodeTenantDenied
				switch mode {
				case "wrong-entry":
					if r.op == ai.OperationChat {
						entry = mixedRoutes[1]
					} else {
						entry = mixedRoutes[0]
					}
				case "unknown-operation":
					b.Operation = "unknown"
					code = ai.CodeInvalidRequest
				case "wrong-options":
					code = ai.CodeInvalidRequest
				case "zero-chat":
					b.Operation = ""
					if r.op == ai.OperationChat {
						w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
					}
				}
				w.host.PutBinding(b)
				id := "req-preflight-" + r.id + "-" + mode
				scope := textScope(id)
				target := ai.Target{BindingID: r.id, ModelID: mixedModel}
				o := mixedOutcome{}
				if mode == "wrong-options" {
					switch r.op {
					case ai.OperationChat:
						res, err := w.client.Complete(ctxFor(t), scope, target, ai.Request{Messages: []ai.Message{ai.UserText("synthetic")}}, ai.ChatOptions{})
						o = mixedOutcome{metadata: res.Metadata, chat: res, err: err}
					case ai.OperationImage:
						var opts ai.ImageOptions = ai.GoogleImagesOptions{}
						if r.provider == ai.ProviderGoogle {
							opts = ai.OpenAIImagesOptions{}
						}
						res, err := w.client.GenerateImages(ctxFor(t), scope, target, ai.ImagesRequest{Prompt: "synthetic"}, opts)
						o = mixedOutcome{metadata: res.Metadata, images: res, err: err}
					case ai.OperationClassifier:
						res, err := w.client.Classify(ctxFor(t), scope, target, mixedClassifierInput(), ai.ResponsesOptions{})
						o = mixedOutcome{metadata: res.Metadata, classifier: res, err: err}
					}
				} else {
					o = entry.invoke(ctxFor(t), w.client, scope, target)
				}
				o.recordResult(ev, "result")
				ev.Record("resolver-reads", w.host.ReadsFor(id))
				if mode == "zero-chat" && r.op == ai.OperationChat {
					ev.Check("old zero operation permits chat", o.err == nil, "got %v", o.err)
				} else {
					ev.Check("refused before credential and Provider", errors.Is(o.err, &ai.Error{Code: code}) && w.host.ReadsFor(id).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.metadata.Attempts) == 0, "got %v / %+v", o.err, o.metadata)
				}
				w.released(ev)
			})
		}
	}
}

func TestMixedSnapshotConflict(t *testing.T) {
	for _, r := range mixedRoutes {
		for _, dimension := range []string{"operation", "provider", "endpoint", "account", "credential"} {
			t.Run(r.id+"/"+dimension, func(t *testing.T) {
				ev := run.Case(t, "E07-mixed-snapshot-conflict-"+r.id+"-"+dimension)
				w := newMixedWorld(t, nil, tenantA)
				w.host.BeforeCredential(func() {
					b := w.host.Binding(tenantA.tenant, r.id)
					b.Version = "b2"
					switch dimension {
					case "operation":
						b.Operation = ai.OperationChat
					case "provider":
						b.ProviderID = ai.ProviderAnthropic
					case "endpoint":
						b.Endpoint = "https://example.invalid"
					case "account":
						b.AccountScopeID = "other"
					case "credential":
						b.CredentialRef = "missing"
					}
					w.host.PutBinding(b)
				})
				id := "req-conflict-" + r.id + "-" + dimension
				o := r.invoke(ctxFor(t), w.client, textScope(id), ai.Target{BindingID: r.id, ModelID: mixedModel})
				w.record(ev, o)
				reads := w.host.ReadsFor(id)
				ev.Record("resolver-reads", reads)
				ev.Check("inconsistent snapshot refuses without reparsing", errors.Is(o.err, &ai.Error{Code: ai.CodeCredentialUnavailable, Phase: ai.PhaseConsistency}) && !o.metadata.Resolved && reads.Bindings == 1 && reads.Credentials == 1 && len(w.provider.Requests()) == 0, "got %v / %+v", o.err, reads)
				w.released(ev)
			})
		}
	}
}

func TestMixedRotationAndRetry(t *testing.T) {
	for _, r := range mixedRoutes {
		t.Run(r.id, func(t *testing.T) {
			ev := run.Case(t, "E07-mixed-snapshot-rotation-retry-"+r.id)
			w := newMixedWorld(t, nil, tenantA, tenantB)
			b := w.host.Binding(tenantA.tenant, r.id)
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			c := ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v2", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA2.secret)}
			fail := provider.Reply{Status: 429, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":"synthetic retry"}`)}, OnReceive: func() {
				b.Version = "b2"
				b.Retry.MaxRetries = 0
				b.Endpoint = w.provider.URL() + r.prefix(tenantA) + "/rotated"
				if r.provider == ai.ProviderGoogle {
					b.Endpoint += "/v1beta"
				}
				w.host.PutBinding(b)
				w.host.PutCredential(c)
			}}
			w.provider.EnqueueAt(r.prefix(tenantA), fail, r.reply(t), r.reply(t))
			w.provider.EnqueueAt(r.prefix(tenantB), r.reply(t))
			other := make(chan mixedOutcome, 1)
			go func() {
				other <- r.invoke(ctxFor(t), w.client, scopeFor(tenantB, "req-other-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
			}()
			old := r.invoke(ctxFor(t), w.client, textScope("req-rotation-old-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
			next := r.invoke(ctxFor(t), w.client, textScope("req-rotation-new-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
			unaffected := <-other
			w.record(ev, old)
			w.record(ev, next)
			w.record(ev, unaffected)
			ev.Check("retry fixed original snapshot", old.err == nil && len(old.metadata.Attempts) == 2 && old.metadata.CredentialVersion == "v1" && old.metadata.BindingVersion == "b1" && w.host.ReadsFor(old.metadata.RequestID).Credentials == 1, "got %+v / %v", old.metadata, old.err)
			ev.Check("new call sees rotation and other tenant unaffected", next.err == nil && next.metadata.CredentialVersion == "v2" && next.metadata.BindingVersion == "b2" && unaffected.err == nil && unaffected.metadata.CredentialVersion == "v1", "new=%+v other=%+v", next.metadata, unaffected.metadata)
			var a []provider.Request
			for _, req := range w.provider.Requests() {
				if strings.HasPrefix(req.Path, r.prefix(tenantA)) {
					a = append(a, req)
				}
			}
			ev.Record("requests", w.provider.Requests())
			ev.Check("same endpoint/key for retries; new key/endpoint next call", len(a) == 3 && a[0].KeyAlias == tenantA.alias && a[1].KeyAlias == tenantA.alias && a[0].Path == a[1].Path && a[2].KeyAlias == tenantA2.alias && strings.Contains(a[2].Path, "/rotated/"), "got %d requests", len(a))
			c.Active = false
			w.host.PutCredential(c)
			revoked := r.invoke(ctxFor(t), w.client, textScope("req-revoked-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
			w.record(ev, revoked)
			ev.Check("revocation stops a new call without fallback", errors.Is(revoked.err, &ai.Error{Code: ai.CodeCredentialUnavailable}) && len(w.provider.Requests()) == 4, "got %v", revoked.err)
			w.released(ev)
		})
	}
}

func TestMixedDisableBindings(t *testing.T) {
	for _, disabled := range mixedRoutes[1:] {
		t.Run(disabled.id, func(t *testing.T) {
			ev := run.Case(t, "E07-mixed-binding-disabled-"+disabled.id)
			w := newMixedWorld(t, nil, tenantA)
			arrived, release := make(chan struct{}), make(chan struct{})
			defer close(release)
			reply := disabled.reply(t)
			reply.OnReceive = func() { close(arrived); <-release }
			w.provider.EnqueueAt(disabled.prefix(tenantA), reply)
			old := make(chan mixedOutcome, 1)
			go func() {
				old <- disabled.invoke(ctxFor(t), w.client, textScope("req-disable-inflight"), ai.Target{BindingID: disabled.id, ModelID: mixedModel})
			}()
			waitFor(ev, "request already sent", arrived)
			b := w.host.Binding(tenantA.tenant, disabled.id)
			b.Enabled = false
			b.Version = "b2"
			w.host.PutBinding(b)
			denied := disabled.invoke(ctxFor(t), w.client, textScope("req-disabled-new"), ai.Target{BindingID: disabled.id, ModelID: mixedModel})
			w.record(ev, denied)
			ev.Check("new calls refused before credential", errors.Is(denied.err, &ai.Error{Code: ai.CodeTenantDenied}) && w.host.ReadsFor(denied.metadata.RequestID).Credentials == 0, "got %v", denied.err)
			for _, r := range mixedRoutes {
				if r.id != disabled.id {
					w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
					o := r.invoke(ctxFor(t), w.client, textScope("req-still-enabled-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
					w.record(ev, o)
					ev.Check("other operation still usable", o.err == nil, "%s: %v", r.id, o.err)
				}
			}
			// Release without promising that disabling could cancel an already
			// transmitted request. Its attempt and spend remain attributed.
			release <- struct{}{}
			o := <-old
			w.record(ev, o)
			ev.Check("old snapshot finishes with recorded consumption", o.err == nil && len(o.metadata.Attempts) == 1 && o.usage.Input > 0, "got %+v / %v", o.metadata, o.err)
			ev.Record("requests", w.provider.Requests())
			w.released(ev)
		})
	}
}
