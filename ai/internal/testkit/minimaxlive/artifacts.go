package minimaxlive

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyArtifacts(dir string, r Report, files map[string]string) error {
	roots := map[string]bool{"manifest.json": true, "audit.json": true, ReportFile: true, "live-report.json": true, "minimax-catalog.json": true, "minimax-sources.json": true, "minimax-replay-origin.json": true, "minimax-ledger.json": true}
	cases := map[string]bool{}
	for _, c := range r.Calls {
		cases[c.CaseID] = true
	}
	allowed := map[string]bool{}
	for _, s := range []string{"immutable-input.json", "public-request.json", "options.json", "plan.json", "budget-before.json", "result.json", "events.json", "observations.json", "exchanges.json", "metrics.json", "budget-after.json", "request.json", "response.sse", "assertions.json", "incomplete-capability.json", "request-rerun.json"} {
		allowed[s] = true
	}
	for name := range files {
		parts := strings.Split(name, "/")
		if len(parts) == 1 && roots[name] {
			continue
		}
		if len(parts) != 2 || !cases[parts[0]] || !allowed[parts[1]] {
			return fmt.Errorf("unexpected MiniMax artifact")
		}
	}
	f, e := os.Open(dir)
	if e != nil {
		return e
	}
	defer f.Close()
	entries, e := f.ReadDir(1025)
	if len(entries) > 1024 {
		return fmt.Errorf("excessive artifact entries")
	}
	if e != nil && e != io.EOF {
		return e
	}
	for _, entry := range entries {
		if entry.IsDir() && !cases[entry.Name()] {
			return fmt.Errorf("unexpected MiniMax case directory")
		}
	}
	var audit struct {
		Findings []json.RawMessage `json:"findings"`
	}
	raw, e := ReadFile(filepath.Join(dir, "audit.json"), 1<<20)
	if e != nil || json.Unmarshal(raw, &audit) != nil || len(audit.Findings) != 0 {
		return fmt.Errorf("MiniMax redaction audit not clean")
	}
	return nil
}
func verifyEvents(events []json.RawMessage, c Call, result Capture) error {
	if c.ID == "full-complete" || c.ID == "simple-complete" {
		if len(events) != 0 {
			return fmt.Errorf("Complete unexpectedly reports stream events")
		}
		return nil
	}
	if len(events) == 0 {
		return fmt.Errorf("MiniMax stream events missing")
	}
	terminals := 0
	for i, raw := range events {
		var event struct {
			Type     ai.EventType `json:"type"`
			Envelope struct {
				Call  ai.CallAttribution
				Event map[string]json.RawMessage
			} `json:"envelope"`
		}
		if json.Unmarshal(raw, &event) != nil || event.Envelope.Call.RequestID != c.RequestID {
			return fmt.Errorf("MiniMax event attribution differs")
		}
		if i == 0 && event.Type != ai.EventStart && !(c.Failure == "transient-request" && len(events) == 1 && event.Type == ai.EventError) {
			return fmt.Errorf("MiniMax start event absent")
		}
		if event.Type == ai.EventDone || event.Type == ai.EventError {
			terminals++
			if i != len(events)-1 {
				return fmt.Errorf("MiniMax terminal not last")
			}
			var message ai.AssistantMessage
			msg, e := DecodeMessage(event.Envelope.Event["Message"])
			if e != nil {
				return e
			}
			message = msg
			if !JSONEqual(message, result.Result.Message) {
				return fmt.Errorf("MiniMax terminal Result differs")
			}
		}
	}
	if terminals != 1 {
		return fmt.Errorf("MiniMax unique terminal missing")
	}
	return nil
}
