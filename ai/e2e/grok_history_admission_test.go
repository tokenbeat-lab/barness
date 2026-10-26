package e2e

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGrokHistoryProducerRequiresFreshGrant(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "legacy-live")
	build := exec.Command("go", "test", "-c", "-tags", "live", "-o", binary, "./ai/live")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	if raw, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build producer: %v %s", e, raw)
	}
	var attempts atomic.Int64
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { attempts.Add(1); w.WriteHeader(502) }))
	defer proxy.Close()
	for _, route := range []string{"openai-responses", "openai-chat", "deepseek-responses", "deepseek-chat", "anthropic-messages", "google-gemini", "zhipu-chat", "zhipu-messages", "minimax-chat", "minimax-responses", "minimax-messages"} {
		for _, mode := range []string{"missing", "invalid", "mixed-route", "closed-renewal", "historical-ledger"} {
			t.Run(route+"/"+mode, func(t *testing.T) {
				if (mode == "closed-renewal" || mode == "historical-ledger") && !strings.HasPrefix(route, "minimax-") {
					t.Skip("only international MiniMax owns historical ledgers")
				}
				ev := run.Case(t, "GROK-HISTORY-PREHTTP-"+route+"-"+mode)
				base := t.TempDir()
				grant := filepath.Join(base, "grant.json")
				if mode != "missing" {
					if e := os.WriteFile(grant, []byte(`{"schema":1,"id":"synthetic-invalid","scope":{}}`), 0600); e != nil {
						t.Fatal(e)
					}
				}
				entry := "^TestLive$/^" + route + "$"
				env := []string{"BARNESS_AI_LIVE=1", "BARNESS_AI_EVIDENCE_DIR=" + base}
				switch {
				case strings.HasPrefix(route, "zhipu-"):
					entry = "^TestGLMFormal$"
					env = append(env, "BARNESS_GLM_ROUTE="+strings.TrimPrefix(route, "zhipu-"), "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-no-vendor-key", "BARNESS_GLM_ACCOUNT_ALIAS=synthetic-account", "BARNESS_GLM_REGION_ALIAS=synthetic-region", "BARNESS_GLM_GRANT="+grant, "BARNESS_GLM_PROXY="+proxy.URL)
				case strings.HasPrefix(route, "minimax-"):
					entry = "^TestMiniMaxFormal$"
					env = append(env, "BARNESS_MINIMAX_ROUTE="+strings.TrimPrefix(route, "minimax-"), "BARNESS_MINIMAX_FORMAL=1", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_KEY=synthetic-no-vendor-key", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=synthetic-region", "BARNESS_MINIMAX_GRANT="+grant, "BARNESS_MINIMAX_PROXY="+proxy.URL)
				default:
					env = append(env, "BARNESS_AI_LIVE_COMBO="+route, "BARNESS_AI_LIVE_KEY_"+strings.ToUpper(strings.ReplaceAll(route, "-", "_"))+"=synthetic-no-vendor-key", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic-account", "BARNESS_AI_LIVE_REGION_ALIAS=synthetic-region", "BARNESS_AI_LIVE_PERMISSIONS_FILE="+grant, "BARNESS_AI_LIVE_PROXY="+proxy.URL)
				}
				if mode == "closed-renewal" {
					env = append(env, "BARNESS_MINIMAX_FORMAL_RENEWAL=fixed-recovery7")
				}
				if mode == "historical-ledger" {
					env = append(env, "BARNESS_MINIMAX_LEDGER="+filepath.Join(base, "historical.jsonl"))
				}
				if mode == "mixed-route" {
					env = append(env, "BARNESS_REGIONAL_ROUTE=kimi-cn-chat")
				}
				cmd := exec.CommandContext(ctxFor(t), binary, "-test.run="+entry, "-test.timeout=1800s")
				cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), env...)
				before := attempts.Load()
				raw, e := cmd.CombinedOutput()
				ev.Record("process", map[string]any{"argv": cmd.Args, "output": regionalSafeOutput(raw), "failed": e != nil, "proxyRequests": attempts.Load() - before, "vendorCalls": 0})
				ev.Check("selected producer refuses before any HTTP", e != nil && attempts.Load() == before, "err=%v requests=%d", e, attempts.Load()-before)
			})
		}
	}
}
