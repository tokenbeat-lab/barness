package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// The ledger is an evidence-consumer contract. Its public classifier must
// reject unlisted identity/field/kind/case combinations, even after the two
// new native routes and four honest non-routes are registered.
func TestRegionalProviderDifferentialLedgerBounds(t *testing.T) {
	ev := run.Case(t, "RP06-ledger-bounds")
	l := loadLedger(t)
	cn := "PIDIFF-RP06-minimax-cn-messages-MiniMax-M3-text-stream"
	z := "PIDIFF-RP06-zai-chat-glm-5.3-text-stream"
	for _, sc := range []struct {
		api, id, path string
		kind          pioracle.DiffKind
		pass          bool
	}{
		{"anthropic-messages", cn, "request.body.max_tokens", pioracle.Changed, true},
		{"anthropic-messages", cn + "-unlisted", "request.body.max_tokens", pioracle.Changed, false},
		{"openai-completions", cn, "request.body.max_tokens", pioracle.Changed, false},
		{"anthropic-messages", cn, "request.body.model", pioracle.Changed, false},
		{"anthropic-messages", cn, "request.body.max_tokens", pioracle.OnlyBarness, false},
		{"openai-completions", z, "request.body.thinking", pioracle.OnlyPi, true},
		{"openai-completions", z, "request.body.model", pioracle.Changed, false},
		{"openai-completions", z, "request.body.thinking", pioracle.OnlyBarness, false},
		{"openai-responses", "PIDIFF-P01-E04-options-full-cache-session-stream", "request.body.prompt_cache_key", pioracle.Changed, true},
		{"openai-responses", "unlisted-cache-case", "request.body.prompt_cache_key", pioracle.Changed, false},
		{"openai-responses", "PIDIFF-P01-E04-options-full-cache-session-stream", "request.body.model", pioracle.Changed, false},
	} {
		v := l.Classify(sc.api, sc.id, []pioracle.Diff{{Path: sc.path, Kind: sc.kind}})
		ev.Check("only exact reviewed tuple can pass", v.Pass == sc.pass, "%+v verdict=%+v", sc, v)
	}
	for _, r := range regionalFrozenRoutes {
		_, ok := l.Route(string(r.provider), string(r.api))
		ev.Check("actual frozen provider/API is never a route exception", !ok, "%+v", r)
	}
	for _, r := range []struct{ p, a string }{{"zai", "openai-responses"}, {"zai", "anthropic-messages"}, {"minimax-cn", "openai-completions"}, {"minimax-cn", "openai-responses"}} {
		entry, ok := l.Route(r.p, r.a)
		ev.Check("each absent frozen route is explicitly classified", ok && entry.Classification == pioracle.Extension, "%+v", r)
	}
}
