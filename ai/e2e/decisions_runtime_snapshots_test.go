package e2e

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
)

func TestDecisionsRuntimePinnedSnapshot(t *testing.T) {
	ev := run.Case(t, "P10-OD45-runtime-pinned-snapshot")
	delay := 20 * time.Millisecond
	clk := clock.NewFake(time.Unix(1790000000, 0), 0)
	var source ai.Catalog
	u := runtimeDecisionsWorld(t, func(c *ai.Config) {
		c.Clock = clk
		source = *c.Catalog
		bindings, credentials := c.Bindings, c.Credentials
		c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
			b, e := bindings.ResolveBinding(ctx, s, id)
			b.Retry = ai.RetryPolicy{MaxRetries: 1, MaxRetryDelay: &delay}
			return b, e
		})
		c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
			// BindingResolver has returned and the library must already own its
			// slices and delay. The credential backend receives another copy.
			b.AllowedModels[0] = "foreign-model"
			*b.Retry.MaxRetryDelay = time.Millisecond
			delay = time.Millisecond
			b.AllowClassifierImages = false
			return credentials.ResolveCredential(ctx, s, b)
		})
	})
	pinned := u.client.Catalog()
	hash, err := pinned.Hash()
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("catalog-before", pinned)
	recordClassifierAssets(ev, classifierAssets(t))
	req := decisionsInput()
	req.Images = classifierAssets(t)[:1]
	var body map[string]any
	var headers http.Header
	counts := [3]int{}
	fail := fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Body: `{"error":{"message":"limited"}}`}.script(t, decisionsProtocol, "")
	fail.OnReceive = func() {
		// Actual receipt proves final headers and payload bytes have been copied.
		u.host.PutCredential(primaryCredential(tenantA2, "v2"))
		u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) {
			b.Enabled = false
			b.AllowClassifierImages = false
			b.AllowedModels = nil
			b.Version = "b2"
			b.Endpoint += "/rotated"
			b.Retry.MaxRetries = 0
		})
		source.Version = "host-mutated"
		source.ClassifierModels[len(source.ClassifierModels)-1].Cost.Input = 999
		body["input"].([]any)[0].(map[string]any)["content"].([]any)[1].(map[string]any)["image_url"] = "changed"
		body["questions"].([]any)[0].(map[string]any)["name"] = "changed"
		headers.Set("X-Frozen", "changed")
		req.Images[0] = ai.Image{Data: "changed", MimeType: "image/gif"}
		req.State[1] = 'X'
		req.Questions["intent"] = nil
	}
	enqueue(ev, u.world, fail, runtimeDecisionsReply(t))
	res, e := runtimeDecisionsCall(ctxFor(t), u, "req-pinned", req, ai.Hooks{
		TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			counts[0]++
			h.Set("X-Frozen", "original")
			headers = h
			return nil
		},
		OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			counts[1]++
			body = p.Body
			return ai.KeepPayload(), nil
		},
		OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
			counts[2]++
			r.Header.Set("x-request-id", "mutated")
			return nil
		},
	})
	ev.Record("result", res)
	ev.Record("sleeps", clk.Sleeps())
	ev.Record("callbacks", counts)
	ev.Check("retry survives backend and borrowed-input mutations", e == nil && len(res.Answers) == 1 && counts == [3]int{1, 1, 1}, "got %v counts=%v", e, counts)
	captures := u.provider.Requests()
	ev.Check("every sent byte header target and key fixed", len(captures) == 2 && reflect.DeepEqual(captures[0], captures[1]) && captures[0].KeyAlias == tenantA.alias && captures[0].Header["X-Frozen"] == "original", "got %d requests", len(captures))
	ev.Check("binding credential price and permission remain pinned", res.Metadata.BindingVersion == "b1" && res.Metadata.CredentialVersion == "v1" && res.Metadata.CatalogVersion == pinned.Version && res.Metadata.CatalogHash == hash && res.Usage.Cost.Input == .000296, "got %+v", res.Metadata)
	ev.Check("resolver delay mutation cannot reduce pinned budget", reflect.DeepEqual(clk.Sleeps(), []time.Duration{10 * time.Millisecond}), "got %v", clk.Sleeps())
	reads := u.host.ReadsFor(res.Metadata.RequestID)
	ev.Record("host-reads", reads)
	ev.Check("no config reread", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)
	if len(res.Metadata.Attempts) == 2 {
		ev.Check("response header isolated from callback", res.Metadata.Attempts[1].ProviderRequestID == "up-runtime", "got %+v", res.Metadata.Attempts)
	}
	u.records(ev, res, e)
	u.released(ev)
}
