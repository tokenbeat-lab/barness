package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// Visual evaluation is a standalone Host. Test its actual public process so
// proxy configuration cannot silently become a direct vendor request.
func TestVisualEvaluationProxyProcess(t *testing.T) {
	for i, value := range []string{
		"http://192.0.2.1:7890", "http://localhost:7890", "https://127.0.0.1:7890",
		"http://name:fixture@127.0.0.1:7890", "http://127.0.0.1", "http://127.0.0.1:0",
		"http://127.0.0.1:65536", "http://127.0.0.1:7890/path", "http://127.0.0.1:7890?x=1",
		"http://127.0.0.1:7890?", "http://127.0.0.1:7890#fragment",
		"http://127.0.0.1:7890#",
	} {
		t.Run(fmt.Sprintf("invalid-%d", i), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("VQ12-proxy-invalid-%d", i))
			out := filepath.Join(t.TempDir(), "absent")
			output, err := visualCommandEnv(t, []string{"OTHER_API_KEY=synthetic-unapproved"}, "-mode", "live", "-proxy", value, "-authorization", "absent", "-out", out)
			_, statErr := os.Stat(out)
			ev.Check("proxy refused before authorization and evidence", err != nil && strings.Contains(string(output), "invalid live proxy configuration") && os.IsNotExist(statErr), "invalid proxy reached later stages")
		})
	}
	for _, mode := range []string{"controlled", "verify"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "VQ12-proxy-"+mode+"-rejected")
			args := []string{"-proxy", "http://127.0.0.1:7890", "-out", filepath.Join(t.TempDir(), "absent")}
			if mode == "verify" {
				args = append(args, "-mode", "live", "-verify", "absent")
			}
			output, err := visualCommand(t, args...)
			ev.Check("proxy requires actual live execution", err != nil && strings.Contains(string(output), "invalid live proxy configuration"), "proxy accepted outside live execution")
		})
	}
	for _, route := range []string{"jev", "decisions"} {
		t.Run("connect-"+route, func(t *testing.T) {
			ev := run.Case(t, "VQ12-proxy-connect-"+route)
			host, model, service := "api.typesafe.ai:443", "jev-1.14.0", "https://api.typesafe.ai/v1/systemone"
			if route == "decisions" {
				host, model, service = "api.openai.com:443", "gpt-6-luna", "https://api.openai.com/v1/decisions"
			}
			var calls, wrong atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != "CONNECT" || r.Host != host || r.Header.Get("Authorization") != "" || r.Header.Get("Proxy-Authorization") != "" {
					wrong.Add(1)
				}
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer proxy.Close()
			proof := filepath.Join(t.TempDir(), "authorization.json")
			visualWrite(t, proof, map[string]any{"route": route, "model": model, "service": service, "catalogHash": builtinCatalogHash, "accountAlias": "synthetic-account", "regionAlias": "synthetic-region", "reference": "proxy-process-fixture", "dedicatedCredential": true, "maxCalls": 12, "maxAttempts": 12})
			out := filepath.Join(run.Dir(), "visual-proxy-"+route)
			// The loopback server only rejects CONNECT, never forwards or serves
			// TLS. No provider request or actual credential leaves this process.
			_, err := visualCommandEnv(t, []string{"BARNESS_VISUAL_KEY=synthetic-proxy-fixture", "HTTPS_PROXY=http://forbidden.invalid:9"}, "-formal", "-mode", "live", "-route", route, "-proxy", proxy.URL, "-authorization", proof, "-out", out)
			ev.Record("process", map[string]any{"failed": err != nil, "connects": calls.Load(), "wrongConnects": wrong.Load()})
			if !ev.Check("explicit proxy reaches only official TLS target", calls.Load() == 12 && wrong.Load() == 0, "CONNECT counts=%d/%d", calls.Load(), wrong.Load()) {
				return
			}
			r := visualRead(t, out, "report.json")
			budget := r["budget"].(map[string]any)
			ev.Check("transport failures retained without retries or invented quality", err != nil && r["status"] == "FAIL" && budget["calls"] == float64(12) && budget["attempts"] == float64(12), "wrong process status or budget")
			config := visualRead(t, out, "run-config.json")
			ev.Check("explicit routing retained", config["proxy"] == proxy.URL, "missing safe routing receipt")
			_, err = visualCommand(t, "-verify", out)
			ev.Check("failed transport evidence independently verifies", err == nil, "failed evidence does not verify")
		})
	}
}
