package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLILegacyRevision(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-legacy-reasoning-revision")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	identities := []struct{ provider, api, id string }{
		{"anthropic", "anthropic-messages", "claude-haiku-4-5"},
		{"anthropic", "anthropic-messages", "claude-haiku-4-5-20251001"},
		{"anthropic", "anthropic-messages", "claude-opus-4-5"},
		{"anthropic", "anthropic-messages", "claude-opus-4-5-20251101"},
		{"anthropic", "anthropic-messages", "claude-sonnet-4-5"},
		{"anthropic", "anthropic-messages", "claude-sonnet-4-5-20250929"},
		{"google", "google-generative-ai", "gemini-2.5-flash"},
		{"google", "google-generative-ai", "gemini-2.5-flash-lite"},
		{"google", "google-generative-ai", "gemini-2.5-pro"},
	}
	for _, identity := range identities {
		for _, overlay := range []string{`{"name":"Reviewed legacy name"}`, `{"cost":{"input":7}}`, `{}`} {
			dir := t.TempDir()
			definition := filepath.Join(dir, "overlay.json")
			os.WriteFile(definition, []byte(overlay), 0600)
			out, exit := invokeCatalogCLI(t, ev, binary, "revise", "--store", filepath.Join(dir, "store"), "--operation", "chat", "--provider", identity.provider, "--api", identity.api, "--model", identity.id, "--definition", definition, "--reason", "preserve accepted legacy reasoning defaults", "--confirm")
			var result struct{ Status string }
			json.Unmarshal(out, &result)
			expected := "awaiting_check"
			if overlay == `{}` {
				expected = "unchanged"
			}
			ev.Check("unrelated revision preserves accepted omitted reasoning map", exit == 0 && result.Status == expected, "identity=%+v overlay=%s exit=%d output=%s", identity, overlay, exit, out)
		}
	}
}
