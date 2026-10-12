package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Runtime proofs keep their own three-kind, image-capable candidate. This
// leaves the choice-only and Jev fixtures as independent narrow controls.
func runtimeDecisionsWorld(t *testing.T, configure func(*ai.Config)) *unaryWorld {
	t.Helper()
	u := newClassifierImagesWorld(t, func(c *ai.Config) {
		m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
		m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}
		m.Capabilities.MinScoreLevels, m.Capabilities.MaxScoreLevels = 2, 10
		if configure != nil {
			configure(c)
		}
	})
	// newDecisionsWorld enables a fixture retry; runtime tests must start at
	// the public zero-value policy and opt in individually.
	u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{} })
	return u
}

func runtimeDecisionsCall(ctx context.Context, u *unaryWorld, id string, req ai.ClassifierRequest, hooks ai.Hooks) (ai.ClassifierResult, error) {
	return u.client.WithHooks(hooks).Classify(ctx, textScope(id), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
}

func runtimeDecisionsReply(t *testing.T) provider.Reply {
	t.Helper()
	return fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "up-runtime"}, Body: `{"answers":[` + decisionsAnswer + `],` + decisionsFullUsage + `}`}.script(t, decisionsProtocol, "")
}

func recordRuntimeConfig(ev *evidence.Case, u *unaryWorld) {
	ev.Record("binding", u.host.Binding(tenantA.tenant, "classifier"))
	ev.Record("catalog", u.client.Catalog())
}
