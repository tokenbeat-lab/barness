package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// The transport changes only the network destination for xAI. Its Binding
// remains the official endpoint, so the production target guard is exercised.
type catalogContractTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (c *catalogContractTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Hostname() != "api.x.ai" {
		return c.next.RoundTrip(r)
	}
	if r.URL.Scheme != "https" || !strings.HasPrefix(r.URL.Path, "/v1/") || r.URL.RawQuery != "" {
		return nil, fmt.Errorf("unexpected xAI contract endpoint")
	}
	copy := r.Clone(r.Context())
	u := *r.URL
	u.Scheme = c.target.Scheme
	u.Host = c.target.Host
	copy.URL = &u
	return c.next.RoundTrip(copy)
}
func newCatalogContractWorld(t *testing.T, cat ai.Catalog, i catalogContractIdentity) (*world, ai.Binding) {
	t.Helper()
	transport := &catalogContractTransport{next: provider.LoopbackTransport()}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Transport = transport
		c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 256, MaxOutputImages: 10, MaxOutputImageBytes: 1 << 16, MaxTotalOutputImageBytes: 1 << 19}
		c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 256, MaxStateBytes: 1 << 16, MaxQuestionBytes: 1 << 16, Images: &ai.ClassifierImagePolicy{MaxImages: 256, MaxTotalImageBytes: 1 << 19}}
	}, tenantA)
	transport.target, _ = url.Parse(w.provider.URL())
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "catalog-contract", Version: "v1", Enabled: true, Operation: i.Operation, ProviderID: i.Provider, API: i.API, Endpoint: w.provider.URL() + "/v1", AuthKind: ai.AuthAPIKey, AccountScopeID: "catalog-account", CredentialRef: "catalog-key", AllowedModels: []string{i.ID}}
	if i.API == ai.APIGoogleGenerativeAI || i.API == ai.APIGoogleInteractions {
		b.Endpoint = w.provider.URL() + "/v1beta"
	}
	if i.Provider == ai.ProviderXAI {
		b.Endpoint = "https://api.x.ai/v1"
	}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA.secret)})
	return w, b
}
func catalogContractTarget(i catalogContractIdentity) ai.Target {
	return ai.Target{BindingID: "catalog-contract", ModelID: i.ID}
}
func catalogContractMetadata(ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity, m ai.CallMetadata) {
	hash, _ := cat.Hash()
	ev.Check("exact immutable price identity", m.CatalogVersion == cat.Version && m.CatalogHash == hash, "%+v", m)
	ev.Check("resolved operation identity", m.Resolved && m.Operation == i.Operation && m.ProviderID == i.Provider && m.API == i.API && m.ModelID == i.ID, "%+v", m)
	ev.Check("one completed controlled attempt", len(m.Attempts) == 1, "%+v", m.Attempts)
}
func catalogContractDenied(ev *evidence.Case, w *world, i catalogContractIdentity, call func(ai.CallScope, ai.Target) error) {
	before := len(w.provider.Requests())
	for _, item := range []struct {
		scope  ai.CallScope
		target ai.Target
	}{
		{ai.CallScope{TenantID: tenantB.tenant, RequestID: "foreign-tenant"}, catalogContractTarget(i)},
		{textScope("not-allowed"), ai.Target{BindingID: "catalog-contract", ModelID: "catalog-not-allowed"}},
	} {
		err := call(item.scope, item.target)
		ev.Check("unauthorized request refused", err != nil, "scope=%+v err=%v", item.scope, err)
	}
	var cross error
	if i.Operation == ai.OperationChat {
		_, cross = w.client.GenerateImages(ctxFor(ev.T()), textScope("cross-operation"), catalogContractTarget(i), ai.ImagesRequest{Prompt: "Cross-operation input."}, nil)
	} else {
		_, cross = w.client.CompleteSimple(ctxFor(ev.T()), textScope("cross-operation"), catalogContractTarget(i), ai.Request{Messages: []ai.Message{ai.UserText("Cross-operation input.")}}, ai.SimpleOptions{})
	}
	ev.Check("binding grants no cross-operation authority", cross != nil, "%v", cross)
	ev.Check("no unauthorized provider request", len(w.provider.Requests()) == before, "requests=%d", len(w.provider.Requests()))
}
