package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func editXAINativeTerminal(t *testing.T, dir string, edit func(map[string]any)) {
	t.Helper()
	editXAINativeTerminalAt(t, dir, "E17-xai-entry-complete/text-http.json", edit)
}
func editXAINativeTerminalAt(t *testing.T, dir, name string, edit func(map[string]any)) {
	t.Helper()
	editXAINativeArray(t, dir, name, func(v []any) {
		e := v[0].(map[string]any)
		chunks := strings.Split(e["responseBody"].(string), "\n\n")
		found := false
		for i, chunk := range chunks {
			if !strings.Contains(chunk, "event: response.completed\n") {
				continue
			}
			lines := strings.Split(chunk, "\n")
			for j, line := range lines {
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				var frame map[string]any
				mustUnmarshal(t, []byte(strings.TrimPrefix(line, "data: ")), &frame)
				edit(frame["response"].(map[string]any))
				raw, err := json.Marshal(frame)
				if err != nil {
					t.Fatal(err)
				}
				lines[j] = "data: " + string(raw)
				found = true
			}
			chunks[i] = strings.Join(lines, "\n")
		}
		if !found {
			t.Fatal("synthetic terminal frame missing")
		}
		e["responseBody"] = strings.Join(chunks, "\n\n")
	})
}

// Every synthetic rebind uses a fresh exclusive ledger entry. Earlier entries
// are preserved, just like the controlled producer's real claim discipline.
func bindXAINativeGrant(t *testing.T, dir string, detail, grant, claim map[string]any) {
	t.Helper()
	id := fmt.Sprintf("SYNTHETIC_native_%d", time.Now().UnixNano())
	grant["id"] = id
	raw, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(dir, "xai-authorization-grant.json"), raw)
	detail["grantId"], detail["claimId"], detail["grantSha256"] = id, id, evidence.Digest(raw)
	claim["grantId"], claim["grantSha256"], claim["runId"] = id, evidence.Digest(raw), filepath.Base(dir)
	claimRaw, err := json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(dir, "xai-authorization-claim.json"), claimRaw)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = root
	common, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(strings.TrimSpace(string(common)), "barness-xai-grants")
	if err := os.MkdirAll(ledger, 0700); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(ledger, id+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write(claimRaw); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
}
func expireXAINativeFixtureGrant(t *testing.T, dir string, after time.Duration) {
	t.Helper()
	detail := readXAINativeObject(t, filepath.Join(dir, "xai-formal-report.json"))
	grant := readXAINativeObject(t, filepath.Join(dir, "xai-authorization-grant.json"))
	claim := readXAINativeObject(t, filepath.Join(dir, "xai-authorization-claim.json"))
	claimed, err := time.Parse(time.RFC3339Nano, claim["claimedAt"].(string))
	if err != nil {
		t.Fatal(err)
	}
	grant["expiresAt"] = claimed.Add(after)
	bindXAINativeGrant(t, dir, detail, grant, claim)
	writeXAINativeJSON(t, filepath.Join(dir, "xai-formal-report.json"), detail)
}

// Tool provenance comes from native function items even if the provider sends
// no encrypted reasoning item on that response. Reported usage remains intact.
func omitXAIToolReasoningFixture(t *testing.T, dir string) {
	t.Helper()
	editXAINativeTerminalAt(t, dir, "E17-xai-tools-required/calls-http.json", func(v map[string]any) {
		var kept []any
		for _, raw := range v["output"].([]any) {
			if raw.(map[string]any)["type"] != "reasoning" {
				kept = append(kept, raw)
			}
		}
		v["output"] = kept
	})
	remove := func(m map[string]any) {
		var kept []any
		for _, raw := range m["content"].([]any) {
			if _, ok := raw.(map[string]any)["thinking"]; !ok {
				kept = append(kept, raw)
			}
		}
		m["content"] = kept
	}
	editXAINativeJSON(t, dir, "E17-xai-tools-required/calls-result.json", func(v map[string]any) { remove(v["message"].(map[string]any)) })
	p := filepath.Join(dir, "E17-xai-tools-required/calls-events.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	mustUnmarshal(t, raw, &events)
	var kept []map[string]any
	for _, e := range events {
		if strings.HasPrefix(e["type"].(string), "thinking_") {
			continue
		}
		if partial, ok := e["Partial"].(map[string]any); ok {
			remove(partial)
		}
		if m, ok := e["Message"].(map[string]any); ok {
			remove(m)
		}
		if index, ok := e["ContentIndex"].(float64); ok {
			e["ContentIndex"] = index - 1
		}
		kept = append(kept, e)
	}
	writeXAINativeJSON(t, p, kept)
	editXAINativeHTTPBody(t, dir, "E17-xai-tools-required/results-http.json", func(v map[string]any) {
		var input []any
		for _, raw := range v["input"].([]any) {
			if raw.(map[string]any)["type"] != "reasoning" {
				input = append(input, raw)
			}
		}
		v["input"] = input
	})
}
