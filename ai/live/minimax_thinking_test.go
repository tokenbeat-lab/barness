//go:build live

package live

import (
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestMiniMaxThinkingOnSupplement(t *testing.T) { driveMiniMax(t) }

// Validate entry BEFORE claim, even though TestMain also validates normal
// isolated entries. The scene set is trusted only after exact grant.Check.
func selectMiniMaxProcess(a *processAuthorization) error {
	r := minimaxReport
	if len(a.grant.Scope.Scenes) == 1 {
		if flag.Lookup("test.run").Value.String() != "^TestMiniMaxThinkingOnSupplement$" || flag.Lookup("test.count").Value.String() != "1" || flag.Lookup("test.timeout").Value.String() != "3m0s" {
			return fmt.Errorf("exact thinking-on entry, count=1 and timeout=180s required")
		}
		r.Kind = minimaxlive.ThinkingKind
		r.Expected = minimaxlive.ThinkingPlan(r.Identity)
		raw, _ := json.Marshal(r.Expected)
		r.PlanHash = minimaxlive.Hash(raw)
		r.Deadline = started.Add(180 * time.Second)
	} else if flag.Lookup("test.run").Value.String() != minimaxEntry() {
		return fmt.Errorf("full MiniMax grant requires its exact full entry")
	}
	return nil
}
