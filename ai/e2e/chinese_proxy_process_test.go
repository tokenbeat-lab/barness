package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
)

// The OD55 public process seam also owns host transport configuration. These
// failures prevent the live runner from needing source overlays for a proxy.
func TestChineseEvaluationProxyProcess(t *testing.T) {
	for i, proxy := range []string{
		"http://192.0.2.1:7890", "http://localhost:7890", "https://127.0.0.1:7890",
		"http://name:fixture@127.0.0.1:7890", "http://127.0.0.1", "http://127.0.0.1:0",
		"http://127.0.0.1:65536", "http://127.0.0.1:7890/path", "http://127.0.0.1:7890?x=1",
		"http://127.0.0.1:7890#fragment",
	} {
		t.Run(fmt.Sprintf("invalid-%d", i), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("OD55-proxy-invalid-%d", i))
			out := t.TempDir()
			output, err := evaluationCommand(t, "-route", "luna", "-live", "-proxy", proxy, "-out", out)
			entries, _ := os.ReadDir(out)
			ev.Check("invalid proxy rejected before credentials or evidence", err != nil && strings.Contains(string(output), "stage=proxy") && len(entries) == 0, "proxy guard missing")
		})
	}
	t.Run("fixture", func(t *testing.T) {
		ev := run.Case(t, "OD55-proxy-fixture-rejected")
		output, err := evaluationCommand(t, "-route", "luna", "-fixture", "normal", "-proxy", "http://127.0.0.1:7890", "-out", t.TempDir())
		ev.Check("proxy requires live invocation", err != nil && strings.Contains(string(output), "stage=proxy"), "fixture accepts live transport configuration")
	})
	t.Run("connect-failure", func(t *testing.T) {
		ev := run.Case(t, "OD55-proxy-connect-failure")
		var calls, wrong atomic.Int32
		proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			if r.Method != "CONNECT" || r.Host != "api.openai.com:443" || r.Header.Get("Authorization") != "" || r.Header.Get("Proxy-Authorization") != "" {
				wrong.Add(1)
			}
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer proxy.Close()
		out := t.TempDir()
		args := []string{"run", "./ai/examples/chineseeval/cmd/chineseeval", "-route", "luna", "-live", "-proxy", proxy.URL, "-out", out}
		cmd := exec.CommandContext(ctxFor(t), "go", args...)
		cmd.Dir = "../.."
		for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
			if v, ok := os.LookupEnv(name); ok {
				cmd.Env = append(cmd.Env, name+"="+v)
			}
		}
		// Only CONNECT reaches this controlled server; there is no TLS listener
		// or forwarding, and therefore no provider request or billable call.
		cmd.Env = append(cmd.Env, "BARNESS_AI_CHINESE_EVAL=1", "BARNESS_AI_CHINESE_EVAL_OPENAI_KEY=proxy-process-fixture", "BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=controlled-proxy@offline")
		output, err := cmd.CombinedOutput()
		ev.Record("process", map[string]any{"failed": err != nil, "connects": calls.Load(), "wrongConnects": wrong.Load()})
		if !ev.Check("explicit proxy reaches only official TLS target", calls.Load() == 24 && wrong.Load() == 0, "CONNECT counts=%d/%d", calls.Load(), wrong.Load()) {
			return
		}
		ev.Check("transport failure is FAIL", err != nil && strings.Contains(string(output), "stage=sample_results"), "wrong process status")
		dir, report := evaluationBundle(t, out, ev)
		budget := report["budget"].(map[string]any)
		ev.Check("all failures retained without retry or fake quality", report["status"] == "FAIL" && budget["calls"] == float64(24) && budget["attempts"] == float64(24) && budget["retries"] == float64(0) && budget["unreported"] == float64(24), "invalid failure budget")
		_, err = evaluationCommand(t, "-route", "luna", "-verify", dir)
		ev.Check("failed transport evidence remains verifiable", err == nil, "failure bundle does not verify")
	})
}
