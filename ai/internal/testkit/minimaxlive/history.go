package minimaxlive

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"path/filepath"
)

func verifyNativeContinuation(dir string, r Report, request map[string]any, result Capture) error {
	var prior Capture
	found := false
	for _, c := range r.Calls {
		if c.ID == "tool-call" && c.Outcome == "PASS" {
			raw, e := ReadFile(filepath.Join(dir, c.CaseID, "result.json"), 2<<20)
			if e != nil {
				return e
			}
			if json.Unmarshal(raw, &prior) != nil {
				return fmt.Errorf("MiniMax prior tool capture invalid")
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("MiniMax native continuation lacks own complete tool turn")
	}
	if result.Result.Metadata.NativeStateDowngrades != (ai.NativeStateDowngrades{}) {
		return fmt.Errorf("MiniMax required native state downgraded")
	}
	contains := func(value string) bool {
		found := false
		var walk func(any)
		walk = func(v any) {
			switch v := v.(type) {
			case string:
				if v == value {
					found = true
				}
			case map[string]any:
				for _, child := range v {
					walk(child)
				}
			case []any:
				for _, child := range v {
					walk(child)
				}
			}
		}
		walk(request)
		return found
	}
	thoughts := 0
	for _, part := range prior.Result.Message.Content {
		if thought, ok := part.(ai.Thinking); ok && thought.Thinking != "" {
			thoughts++
			if !contains(thought.Thinking) {
				return fmt.Errorf("MiniMax continuation omitted original reasoning text")
			}
			if r.Identity.Route == "messages" && thought.Signature != "" && !contains(thought.Signature) {
				return fmt.Errorf("MiniMax continuation omitted original thinking signature")
			}
		}
	}
	if thoughts == 0 {
		return fmt.Errorf("MiniMax own reasoning prerequisite absent")
	}
	// The result must be a completed assistant answer to the Host tool result.
	if result.Result.Message.StopReason != ai.StopReasonStop {
		return fmt.Errorf("MiniMax native continuation did not complete")
	}
	return nil
}
