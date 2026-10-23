package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

func TestKimiPacingGrantGuards(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	for _, sc := range []struct {
		name, combo string
		interval    any
	}{{"negative", "kimi-intl-chat", -1}, {"too-large", "kimi-intl-chat", 60001}, {"fraction", "kimi-intl-chat", json.Number("1.5")}, {"overflow", "kimi-intl-chat", json.Number("9223372036854775808")}, {"null", "kimi-intl-chat", nil}, {"non-kimi", "zai-chat", 40}} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-pacing-grant-"+sc.name)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, "openai-completions", kind))
			}
			grant := kimiPacingGrant(t, sc.combo, digest, sc.name, sc.interval)
			cmd, _ := kimiPacingCommand(t, binary, sc.combo, srv.URL(), grant)
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw)})
			ev.Record("requests", srv.Requests())
			ev.Check("invalid pacing refuses before any HTTP", err != nil && len(srv.Requests()) == 0, "%v requests=%d %s", err, len(srv.Requests()), regionalSafeOutput(raw))
		})
	}
}
