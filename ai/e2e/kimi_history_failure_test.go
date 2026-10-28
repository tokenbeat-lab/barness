package e2e

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiHistorySupplementFailure(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary, digest := kimiQualificationBinary(t)
	for _, route := range kimiQualificationRoutes {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KHS-failure-"+route)
			reply := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"type":"internal_server_error","message":"synthetic"}}`)}}
			_, srv, cmd := kimiHistoryCapture(t, ev, binary, digest, route, strings.Repeat("a", 64), true, reply)
			defer srv.Close()
			again := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			again.Dir, again.Env = cmd.Dir, cmd.Env
			out, err := again.CombinedOutput()
			ev.Record("repeat", map[string]any{"error": errString(err), "output": regionalSafeOutput(out), "requests": len(srv.Requests()), "vendorCalls": 0})
			ev.Check("failure leaves authority consumed", err != nil && len(srv.Requests()) == 1, "requests=%d", len(srv.Requests()))
		})
	}
}
