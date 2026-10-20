package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// This Host explicitly shares each real service/account's one permit across
// tenants. Same-named accounts at other services have independent permits;
// this policy uses public Admission metadata, with no new library quota engine.
func TestRegionalProviderAdmissionIsolation(t *testing.T) {
	for _, limited := range regionalRuntimeProviders {
		for _, r := range miniMaxSharedRoutes {
			t.Run(string(limited)+"/"+r.name, func(t *testing.T) {
				ev := run.Case(t, "RP15-account-admission-"+string(limited)+"-"+r.name)
				type accountKey struct {
					provider ai.ProviderID
					account  string
				}
				var mu sync.Mutex
				held := map[accountKey]int{}
				var requests []ai.AdmissionRequest
				base := host.NewAdmission()
				admission := unaryAdmissionFunc(func(ctx context.Context, req ai.AdmissionRequest) (func(), error) {
					key := accountKey{req.ProviderID, req.AccountScopeID}
					mu.Lock()
					requests = append(requests, req)
					if held[key] >= 1 {
						mu.Unlock()
						return nil, errors.New("synthetic service account busy")
					}
					held[key]++
					mu.Unlock()
					release, err := base.Admit(ctx, req)
					if err != nil {
						mu.Lock()
						held[key]--
						mu.Unlock()
						return nil, err
					}
					var once sync.Once
					return func() { once.Do(func() { release(); mu.Lock(); held[key]--; mu.Unlock() }) }, nil
				})
				w := newRegionalRuntimeWorld(t, func(c *ai.Config) { c.Admission = admission })
				w.adm = base
				entered := make(chan struct{})
				hold := provider.Reply{End: provider.EndSilent, OnReceive: func() { close(entered) }}
				ev.Record("held-response-script", hold)
				w.provider.EnqueueAt(regionalRuntimePrefix("regional-a", limited, r.api), hold)
				id := "regional-account-held-" + string(limited) + "-" + r.name
				s := w.client.Stream(ctxFor(t), ai.CallScope{TenantID: "regional-a", RequestID: id}, regionalRuntimeTarget(limited, r.api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic held account.")}}, nil)
				defer s.Close()
				waitFor(ev, "service/account permit is held", entered)
				denied := drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: "regional-b", RequestID: id + "-denied"}, regionalRuntimeTarget(limited, r.api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic shared account.")}}, nil))
				regionalRuntimeRecord(ev, w, denied)
				checkRejected(ev, denied, ai.CodeAdmissionDenied, ai.PhaseAdmission)
				ev.Check("same actual service/account shares explicit Host limit", len(denied.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 1, "%+v", denied.result.Metadata)
				for _, sibling := range regionalRuntimeProviders {
					if sibling == limited {
						continue
					}
					reply := regionalRuntimeReply(t, r.api)
					ev.Record(string(sibling)+"-script", reply)
					w.provider.EnqueueAt(regionalRuntimePrefix("regional-b", sibling, r.api), reply)
					res, err := w.client.Complete(ctxFor(t), ai.CallScope{TenantID: "regional-b", RequestID: id + "-sibling-" + string(sibling)}, regionalRuntimeTarget(sibling, r.api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic other service account.")}}, nil)
					regionalRuntimeRecord(ev, w, outcome{result: res, err: err})
					ev.Check("other service's same account alias is independent", err == nil && res.Metadata.ProviderID == sibling, "%v %+v", err, res.Metadata)
				}
				_ = s.Close()
				subject := drain(s)
				regionalRuntimeRecord(ev, w, subject)
				ev.Check("held call cancels and returns account permit", errors.Is(subject.err, &ai.Error{Code: ai.CodeCanceled}), "%v", subject.err)
				reply := regionalRuntimeReply(t, r.api)
				ev.Record("after-release-script", reply)
				w.provider.EnqueueAt(regionalRuntimePrefix("regional-b", limited, r.api), reply)
				res, err := w.client.Complete(ctxFor(t), ai.CallScope{TenantID: "regional-b", RequestID: id + "-after-release"}, regionalRuntimeTarget(limited, r.api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic released account.")}}, nil)
				regionalRuntimeRecord(ev, w, outcome{result: res, err: err})
				ev.Check("permit reusable after cancellation", err == nil && res.Metadata.ProviderID == limited, "%v", err)
				mu.Lock()
				total := 0
				for _, n := range held {
					total += n
				}
				snapshot := append([]ai.AdmissionRequest(nil), requests...)
				mu.Unlock()
				ev.Record("account-admission-requests", snapshot)
				ev.Record("account-permits", total)
				ev.Check("all service/account leases released", total == 0, "got %d", total)
				ev.Check("one controlled denial and five actual sends", len(snapshot) == 6 && len(w.provider.Requests()) == 5, "requests=%d sends=%d", len(snapshot), len(w.provider.Requests()))
				w.released(ev)
			})
		}
	}
}
