package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Raw JSON is intentional: a map cannot express duplicate authorization keys.
func TestKimiPacingGrantJSONAmbiguity(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	consumer := kimiQualificationConsumer(t)
	for _, sc := range []struct {
		name, alias, value string
		canonical          bool
	}{
		{"case-duplicate", "CallIntervalMs", "0", true},
		{"exact-duplicate", "callIntervalMs", "0", true},
		{"alias-null", "CallIntervalMs", "null", false},
		{"uppercase", "CALLINTERVALMS", "0", false},
		{"unicode-fold", "callIntervalMſ", "0", false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-pacing-json-"+sc.name)
			combo := "kimi-intl-chat"
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, "openai-completions", kind))
			}
			grant := kimiPacingGrant(t, combo, digest, "json-"+sc.name, 35000)
			if !sc.canonical {
				delete(grant, "callIntervalMs")
			}
			cmd, base := kimiPacingCommand(t, binary, combo, srv.URL(), grant)
			raw := strings.TrimSpace(string(mustMarshal(t, grant)))
			raw = strings.TrimSuffix(raw, "}") + `,"` + sc.alias + `":` + sc.value + `}`
			for _, env := range cmd.Env {
				if path, ok := strings.CutPrefix(env, "BARNESS_REGIONAL_GRANT="); ok {
					if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			ev.Record("raw-grant", raw)
			out, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(out)})
			ev.Record("requests", srv.Requests())
			ev.Check("ambiguous new authorization field refuses before HTTP", err != nil && len(srv.Requests()) == 0, "%v requests=%d", err, len(srv.Requests()))
			if err == nil {
				dir := kimiPacingBundle(t, base)
				verify := exec.CommandContext(ctxFor(t), consumer, "-controlled", dir)
				verify.Dir, verify.Env = "../..", regionalChildEnvironment()
				vr, ve := verify.CombinedOutput()
				ev.Record("consumer", map[string]any{"argv": verify.Args, "error": errString(ve), "output": regionalSafeOutput(vr)})
				ev.Check("consumer independently refuses ambiguous original grant", ve != nil, "%s", regionalSafeOutput(vr))
			}
		})
	}
}
