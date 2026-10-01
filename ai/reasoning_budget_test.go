package ai

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// TestThinkingBudgets checks the shared thinking budget rules token-budget
// protocols build on (spec I7: default budgets, custom budgets, answer room).
// They have no effect on the Responses wire, so no E2E reaches them yet; this
// is the one isolated test, and with BARNESS_AI_PIDIFF=1 every expectation is
// also checked against frozen pi-ai's own adjustMaxTokensForThinking.
//
// Ways it could fail:
//
//	B1 a default budget differs from minimal 1024 / low 2048 / medium 8192 /
//	   high 16384, or xhigh/max do not use high's
//	B2 a custom budget is ignored, applied to another level, or a custom 0 is
//	   treated as unset
//	B3 without a caller cap the model cap is not the ceiling; with one, the
//	   budget is not added on top or the model cap is exceeded
//	B4 when the ceiling does not exceed the budget, fewer than 1024 tokens
//	   are left for the answer, or the budget goes negative
func TestThinkingBudgets(t *testing.T) {
	base := func(n int) *int { return &n }
	cases := []struct {
		name         string
		base         *int
		modelMax     int
		level        ThinkingLevel
		budgets      ThinkingBudgets
		wantMax      int
		wantThinking int
	}{
		{"default-minimal", nil, 128000, ThinkingMinimal, ThinkingBudgets{}, 128000, 1024},
		{"default-low", nil, 128000, ThinkingLow, ThinkingBudgets{}, 128000, 2048},
		{"default-medium", nil, 128000, ThinkingMedium, ThinkingBudgets{}, 128000, 8192},
		{"default-high", nil, 128000, ThinkingHigh, ThinkingBudgets{}, 128000, 16384},
		{"xhigh-is-high", nil, 128000, ThinkingXHigh, ThinkingBudgets{High: Value(20000)}, 128000, 20000},
		{"max-is-high", nil, 128000, ThinkingMax, ThinkingBudgets{}, 128000, 16384},
		{"custom-low", nil, 128000, ThinkingLow, ThinkingBudgets{Low: Value(3000)}, 128000, 3000},
		{"custom-zero-kept", nil, 128000, ThinkingMedium, ThinkingBudgets{Medium: Value(0)}, 128000, 0},
		{"custom-other-level", nil, 128000, ThinkingHigh, ThinkingBudgets{Low: Value(3000)}, 128000, 16384},
		{"cap-plus-budget", base(4000), 128000, ThinkingMedium, ThinkingBudgets{}, 12192, 8192},
		{"cap-plus-budget-at-model-max", base(120000), 128000, ThinkingHigh, ThinkingBudgets{}, 128000, 16384},
		{"answer-room", nil, 8192, ThinkingHigh, ThinkingBudgets{}, 8192, 7168},
		{"answer-room-floor", nil, 512, ThinkingMinimal, ThinkingBudgets{}, 512, 0},
		{"ceiling-equals-budget", nil, 2048, ThinkingLow, ThinkingBudgets{}, 2048, 1024},
		{"ceiling-above-budget", nil, 2049, ThinkingLow, ThinkingBudgets{}, 2049, 2048},
		{"cap-zero", base(0), 128000, ThinkingMinimal, ThinkingBudgets{}, 1024, 0},
	}

	var pi []pioracle.BudgetResult
	if pioracle.Enabled() {
		o, err := pioracle.Open()
		if err != nil {
			t.Fatal(err)
		}
		in := make([]pioracle.BudgetCase, len(cases))
		for i, c := range cases {
			budgets, err := json.Marshal(c.budgets)
			if err != nil {
				t.Fatal(err)
			}
			in[i] = pioracle.BudgetCase{BaseMaxTokens: c.base, ModelMaxTokens: c.modelMax, Level: string(c.level), Budgets: budgets}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if pi, err = o.ThinkingBudgets(ctx, in); err != nil {
			t.Fatal(err)
		}
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, hasBase := 0, c.base != nil
			if hasBase {
				b = *c.base
			}
			gotMax, gotThinking := adjustMaxTokensForThinking(b, hasBase, c.modelMax, c.level, c.budgets)
			if gotMax != c.wantMax || gotThinking != c.wantThinking {
				t.Errorf("got (%d, %d), want (%d, %d)", gotMax, gotThinking, c.wantMax, c.wantThinking)
			}
			if pi != nil && (pi[i].MaxTokens != c.wantMax || pi[i].ThinkingBudget != c.wantThinking) {
				t.Errorf("pi gives (%d, %d), the expectation is (%d, %d)", pi[i].MaxTokens, pi[i].ThinkingBudget, c.wantMax, c.wantThinking)
			}
		})
	}
}
