package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVisualEvaluationAuthorization(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, name := range []string{"missing_key", "wrong_service", "wrong_model", "wrong_hash", "wrong_budget", "duplicate_key", "foreign_credential", "foreign_token"} {
			t.Run(route+"/"+name, func(t *testing.T) {
				ev := run.Case(t, "VQ12-authorization-"+route+"-"+name)
				model, service, hash := "jev-1.14.0", "https://api.typesafe.ai/v1/systemone", "sha256:2cec32a720f2a8ebea657ae59ba3b47b6c088f3025965862dee454d7f1bfa49d"
				if route == "decisions" {
					model, service, hash = "gpt-6-luna", "https://api.openai.com/v1/decisions", "sha256:b247ae003f71145c7f5be5e00424a2a298acd4255bc9a25a7d222aba7d33ec65"
				}
				proof := map[string]any{"route": route, "model": model, "service": service, "catalogHash": hash, "accountAlias": "synthetic-account", "regionAlias": "synthetic-region", "reference": "test-authorization", "dedicatedCredential": true, "maxCalls": 12, "maxAttempts": 12}
				switch name {
				case "wrong_service":
					proof["service"] = "https://unapproved.invalid/endpoint"
				case "wrong_model":
					proof["model"] = "unapproved"
				case "wrong_hash":
					proof["catalogHash"] = "changed"
				case "wrong_budget":
					proof["maxCalls"] = 13
				}
				path := filepath.Join(t.TempDir(), "authorization.json")
				visualWrite(t, path, proof)
				if name == "duplicate_key" {
					raw, _ := os.ReadFile(path)
					os.WriteFile(path, append([]byte(`{"maxCalls":13,`), raw[1:]...), 0600)
				}
				extra := []string{}
				if name == "foreign_credential" {
					extra = []string{"OTHER_API_KEY=synthetic-foreign-credential"}
				}
				if name == "foreign_token" {
					extra = []string{"ACCESS_TOKEN=synthetic-foreign-token"}
				}
				out := filepath.Join(run.Dir(), "visual-auth-"+route+"-"+name)
				_, err := visualCommandEnv(t, extra, "-mode", "live", "-route", route, "-authorization", path, "-out", out)
				ev.Check("invalid authorization cannot execute", (err == nil) == (name == "missing_key"), "incorrect authorization exit")
				r := visualRead(t, out, "report.json")
				ev.Record("report", r)
				b := r["budget"].(map[string]any)
				ev.Check("isolated zero-spend NOT_RUN retained", r["status"] == "NOT_RUN" && b["calls"] == float64(0) && b["attempts"] == float64(0) && len(r["samples"].([]any)) == 12, "spent or invented quality")
				_, err = visualCommand(t, "-verify", out)
				ev.Check("configuration failure receipt verifies", err == nil, "invalid zero-spend receipt")
			})
		}
	}
}
