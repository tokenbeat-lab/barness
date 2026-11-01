package e2e

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/catalog"
)

func TestLocalCatalogFixedSnapshotAndAuthorization(t *testing.T) {
	ev := run.Case(t, "CAT-T12-file-report-retry-and-permission-snapshot")
	extra := strings.Replace(localChatJSON, `gpt-4.1-mini`, `not-whitelisted`, 1)
	raw := localCatalogFile(`{"version":"snapshot-01","models":[` + localChatJSON + `,` + extra + `]}`)
	path := filepath.Join(t.TempDir(), "selected.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, report, err := catalog.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := catalogHash(t, loaded)
	ev.Record("original-report", report)
	fixture, fixtureRaw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("text-basic.json", fixtureRaw)
	retry, retryRaw := loadRetryFixture(t)
	ev.Fixture("retry.json", retryRaw)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &loaded; c.Clock = retry.clock() }, tenantA)
	w.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
	denied := ai.Target{BindingID: "primary", ModelID: "not-whitelisted"}
	_, deniedErr := w.client.CompleteSimple(ctxFor(t), textScope("catalog-not-authorized"), denied, textRequest(fixture), ai.SimpleOptions{})
	ev.Check("loaded definition does not extend Binding permission", errors.Is(deniedErr, &ai.Error{Code: ai.CodeTenantDenied}) && len(w.provider.Requests()) == 0, "err=%v", deniedErr)
	changed := make(chan error, 1)
	failing := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{}`)}, Script: `{}`}
	failing.OnReceive = func() {
		// The Provider has captured the first attempt. These are caller-owned
		// copies, so only this logical call's construction snapshot may price
		// the retry. The channel makes the mutation completion observable.
		loaded.Models[0].Cost.Input = 999
		loaded.Models[0].Input[0] = ai.ModalityImage
		report.Models[0].Status = catalog.Failed
		copy(raw, []byte(`broken`))
		changed <- os.WriteFile(path, []byte(`broken`), 0600)
	}
	enqueue(ev, w, failing, sseReply(t, fixture, provider.FramingLF))
	callContext := ctxFor(t)
	res, callErr := w.client.CompleteSimple(callContext, textScope("catalog-retry"), textTarget(fixture), textRequest(fixture), ai.SimpleOptions{})
	select {
	case mutationErr := <-changed:
		if mutationErr != nil {
			t.Fatal(mutationErr)
		}
	case <-callContext.Done():
		t.Fatal("first attempt never reached the controlled Provider")
	}
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	ev.Check("in-flight retry uses original loaded prices", callErr == nil && len(res.Metadata.Attempts) == 2, "err=%v metadata=%+v", callErr, res.Metadata)
	checkFinalMessage(ev, fixture, res.Message)
	checkPricingSnapshot(ev, res.Metadata, "snapshot-01", hash)
	requests := w.provider.Requests()
	ev.Check("both attempts send original request bytes", len(requests) == 2 && jsonEqual(requests[0].Body, requests[1].Body), "requests=%+v", requests)
	original, _ := w.client.Catalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
	ev.Check("report/source mutation cannot alter Client", original.Cost.Input == 0.4 && original.Input[0] == ai.ModalityText, "model=%+v", original)
	_, _, readErr := catalog.LoadFile(path)
	ev.Check("new explicit file selection observes corruption", readErr != nil, "err=%v", readErr)

	// A vendor rejection stays a call result, with the same effective price
	// identity. It cannot trigger a catalog refresh, removal or model switch.
	enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic auth rejection"}}`)}, Script: `{"error":{"message":"synthetic auth rejection"}}`})
	rejected, rejectErr := w.client.CompleteSimple(ctxFor(t), textScope("catalog-runtime-failure"), textTarget(fixture), textRequest(fixture), ai.SimpleOptions{})
	ev.Record("rejected-call", rejected)
	ev.Check("runtime error keeps loaded catalog and price identity", rejectErr != nil && rejected.Metadata.CatalogHash == hash && catalogHash(t, w.client.Catalog()) == hash, "err=%v", rejectErr)
	ev.Record("source-file-before-mutation", json.RawMessage(localCatalogFile(`{"version":"snapshot-01","models":[`+localChatJSON+`,`+extra+`]}`)))
}
