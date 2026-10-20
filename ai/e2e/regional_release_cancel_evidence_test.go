package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// A canceled Result describes emitted content, while the recorder can retain
// valid read-ahead or an incomplete final frame. It cannot invent a prefix.
func regionalCaptureCancellationPrefixes(t *testing.T, original, combo string) {
	for _, mode := range []string{"read-ahead", "partial-frame", "wrong-text"} {
		t.Run("cancel-"+mode, func(t *testing.T) {
			ev := run.Case(t, "RP18-cancel-evidence-"+combo+"-"+mode)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			caseID := supportmatrix.RegionalCaseID(combo, supportmatrix.RegionalModels(combo)[0]+"/cancel")
			path := filepath.Join(dir, caseID, "exchanges.json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var xs []map[string]any
			if json.Unmarshal(raw, &xs) != nil {
				t.Fatal("invalid capture")
			}
			body := xs[0]["responseBody"].(string)
			switch mode {
			case "wrong-text":
				body = strings.ReplaceAll(body, "known-content", "synthetic incorrect prefix")
			case "partial-frame":
				body += "event: synthetic\ndata: {"
			case "read-ahead":
				route, _ := supportmatrix.RegionalRouteOf(combo)
				tail := `{"choices":[{"index":0,"delta":{"content":" more"},"finish_reason":"stop"}],"usage":{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}}`
				if route.API == "openai-responses" {
					tail = `{"type":"response.completed","response":{"id":"regional-response","model":"` + supportmatrix.RegionalModels(combo)[0] + `","status":"completed","output":[],"usage":{"input_tokens":0,"output_tokens":0,"total_tokens":0}}}`
				}
				if route.API == "anthropic-messages" {
					tail = `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":0}}`
				}
				body += "data: " + tail + "\n\n"
			}
			xs[0]["responseBody"] = body
			raw = mustMarshal(t, xs)
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			var manifest map[string]any
			mraw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
			if json.Unmarshal(mraw, &manifest) != nil {
				t.Fatal("invalid manifest")
			}
			for _, v := range manifest["cases"].([]any) {
				c := v.(map[string]any)
				if c["id"] == caseID {
					c["artifacts"].(map[string]any)["exchanges.json"] = pioracle.SHA256(raw)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), mustMarshal(t, manifest), 0600); err != nil {
				t.Fatal(err)
			}
			if err := supportmatrix.SealRegionalBundle(dir); err != nil {
				t.Fatal(err)
			}
			_, err = supportmatrix.VerifyRegionalControlledCapture(dir)
			ev.Record("synthetic-cancel-evidence", map[string]any{"mode": mode, "refused": err != nil, "error": err, "originalRetained": true})
			ev.Check("observed cancellation prefix", (err != nil) == (mode == "wrong-text"), "mode=%s error=%v", mode, err)
		})
	}
}
