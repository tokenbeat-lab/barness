//go:build live

package live

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Real smoke entry point, a synthetic credential, and a loopback proxy which
// only rejects CONNECT: this process test cannot send anything to a vendor.
func TestReleaseLiveProxyProcess(t *testing.T) {
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "CONNECT" {
			requests.Add(1)
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()
	for _, mode := range []string{"loopback", "remote", "credential", "query", "invalid-port"} {
		t.Run(mode, func(t *testing.T) {
			cs := run.Case(t, "OD57-live-explicit-proxy-"+mode)
			value := proxy.URL
			switch mode {
			case "remote":
				value = "http://192.0.2.1:7890"
			case "credential":
				value = "http://synthetic:secret@127.0.0.1:7890"
			case "query":
				value = proxy.URL + "?"
			case "invalid-port":
				value = "http://127.0.0.1:65536"
			}
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLive$/^typesafe-classifier$/^single-choice$", "-test.timeout=10s")
			base := t.TempDir()
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), envLive + "=1", envCombo + "=typesafe-classifier", envAlias + "=synthetic@loopback", keyVar("typesafe-classifier") + "=key:controlled-smoke", "BARNESS_AI_LIVE_PROXY=" + value, "BARNESS_AI_EVIDENCE_DIR=" + base}
			before := requests.Load()
			output, err := cmd.CombinedOutput()
			cs.Record("process", map[string]any{"rejected": err != nil, "proxyRequests": requests.Load() - before, "timedOut": ctx.Err() != nil, "configurationRejected": strings.Contains(string(output), "proxy")})
			cs.Check("bounded failure", err != nil && ctx.Err() == nil, "proxy failure not bounded")
			if mode == "loopback" {
				cs.Check("explicit proxy used by actual smoke", requests.Load() > before, "proxy ignored")
			} else {
				cs.Check("invalid proxy refused before network", requests.Load() == before && strings.Contains(string(output), "proxy"), "invalid proxy accepted")
			}
		})
	}
}
