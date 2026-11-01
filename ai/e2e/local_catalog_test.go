package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/catalog"
)

// The local loading seam was fixed by the spec before implementation. Unknown
// constraints must isolate their record, never disappear into a usable model.
func TestLocalCatalogLoading(t *testing.T) {
	ev := run.Case(t, "CAT-T08-isolated-loading-and-public-call")
	raw := localCatalogFile(`{"version":"local-01","models":[` + localChatJSON + `,` +
		`{"provider":"openai","api":"openai-responses","id":"future","newConstraint":true}` + `]}`)
	ev.Record("source-file", json.RawMessage(raw))
	loaded, report, err := catalog.Load(raw)
	ev.Record("report", report)
	ev.Check("one accepted and one isolated", err == nil && report.Loaded == 1 && report.Failed == 1, "report=%+v err=%v", report, err)
	if err != nil || len(report.Models) != 2 || len(loaded.Models) != 1 {
		t.FailNow()
	}
	ev.Check("unknown field retained as failed definition", report.Models[1].Status == catalog.Failed && report.Models[1].Reason == catalog.UnknownField && len(report.Models[1].Raw) != 0, "status=%+v", report.Models[1])
	version, hash := loaded.Version, catalogHash(t, loaded)
	ev.Check("source identity distinct from effective identity", report.SourceVersion == "local-01" && report.SourceHash != hash && report.EffectiveHash == hash, "report=%+v", report)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &loaded }, tenantA)
	loaded.Models[0].ID = "changed"
	report.Models[1].Raw[0] = ' '
	raw[0] = ' '
	f, fixture := loadTextFixture(t, "text-basic.json")
	ev.Fixture("text-basic.json", fixture)
	enqueue(ev, w, sseReply(t, f, provider.FramingLF))
	res, callErr := w.client.CompleteSimple(ctxFor(t), textScope("local-catalog"), textTarget(f), textRequest(f), ai.SimpleOptions{})
	ev.Check("public call uses accepted immutable catalog", callErr == nil, "err=%v", callErr)
	checkFinalMessage(ev, f, res.Message)
	checkPricingSnapshot(ev, res.Metadata, version, hash)
	checkTextRequest(ev, w, f, tenantA, true)
	_, ok := w.client.Catalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "future")
	ev.Check("failed record not queryable", !ok, "failed model was exposed")
}

const localChatJSON = `{"provider":"openai","api":"openai-responses","id":"gpt-4.1-mini","name":"GPT-4.1 mini","reasoning":false,"input":["text","image"],"contextWindow":1047576,"maxTokens":32768,"compat":{"supportsStrictMode":true},"cost":{"input":0.4,"output":1.6,"cacheRead":0.1,"cacheWrite":0}}`

// This test fixture signs original JSON bytes independently of ai.Catalog.Hash.
// Unknown fields and number/escape spelling must participate in this digest.
func localCatalogFile(body string) []byte {
	sum := sha256.Sum256([]byte(body))
	return []byte(`{"hash":"sha256:` + hex.EncodeToString(sum[:]) + `","catalog":` + body + `}`)
}
