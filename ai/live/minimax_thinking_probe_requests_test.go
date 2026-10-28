//go:build live

package live

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func miniMaxProbeRequests(t *testing.T, cs *evidence.Case, dir, route string) {
	t.Helper()
	read := func(id, file string, dst any) {
		raw, err := os.ReadFile(filepath.Join(dir, "MM07-"+id, file+".json"))
		if err != nil || json.Unmarshal(raw, dst) != nil {
			t.Fatalf("read original %s/%s: %v", id, file, err)
		}
	}
	var request struct {
		Messages []struct{ Content []struct{ Text string } }
	}
	read("thinking-on", "public-request", &request)
	prompt := ""
	if len(request.Messages) == 1 && len(request.Messages[0].Content) == 1 {
		prompt = request.Messages[0].Content[0].Text
	}
	cs.Check("fixed finite task requires independent congruence calculation", len(prompt) < 1024 && strings.Contains(prompt, "divided by 17, 19, and 23") && strings.Contains(prompt, "remainders 7, 11, and 13") && !strings.Contains(prompt, "5464"), "probe task=%q", prompt)
	// An independent finite oracle checks that this fixed task has a unique
	// answer; the peer does not use production request-building logic.
	answers := 0
	for n := 1; n <= 17*19*23; n++ {
		if n%17 == 7 && n%19 == 11 && n%23 == 13 {
			answers++
			cs.Check("independent task answer", n == 5464, "n=%d", n)
		}
	}
	cs.Check("independent task solution is unique", answers == 1, "solutions=%d", answers)
	var baseline struct {
		Messages []struct{ Content []struct{ Text string } }
	}
	for _, id := range []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "thinking-control", "preflight-error"} {
		read(id, "public-request", &baseline)
		cs.Check("unrelated text task unchanged "+id, len(baseline.Messages) == 1 && len(baseline.Messages[0].Content) == 1 && baseline.Messages[0].Content[0].Text == "Synthetic protocol probe. Reply with the integer 42.", "task changed")
	}
	var options map[string]any
	read("thinking-on", "options", &options)
	raw, _ := json.Marshal(options)
	cs.Check("own output budget unchanged", strings.Contains(string(raw), "2048"), "options=%s", raw)
	var requestWire []exchange
	read("thinking-on", "exchanges", &requestWire)
	cs.Check("one own HTTP attempt", len(requestWire) == 1, "requests=%d", len(requestWire))
	if len(requestWire) != 1 {
		return
	}
	var body map[string]any
	if json.Unmarshal([]byte(requestWire[0].RequestBody), &body) != nil {
		t.Fatal("captured wire body absent")
	}
	field := map[string]string{"chat": "max_completion_tokens", "responses": "max_output_tokens", "messages": "max_tokens"}[route]
	cs.Check("fixed model and output ceiling unchanged", body["model"] == "MiniMax-M3" && body[field] == float64(2048), "model=%v output=%v", body["model"], body[field])
	if route != "responses" {
		thinking, _ := body["thinking"].(map[string]any)
		cs.Check("M3 adaptive mode unchanged", thinking["type"] == "adaptive", "thinking=%v", thinking)
		if route == "chat" {
			cs.Check("separate thinking format unchanged", body["reasoning_split"] == true, "reasoning_split=%v", body["reasoning_split"])
		}
	} else {
		cs.Check("M3 responses reasoning remains selected", body["reasoning"] != nil, "reasoning absent")
	}
}
