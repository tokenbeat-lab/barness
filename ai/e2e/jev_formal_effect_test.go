package e2e

import (
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevFormalEffectProcess(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC13-formal-effects-"+entry)
			out := filepath.Join(run.Dir(), "formal-effects-"+entry)
			output, err := visualCommand(t, "-formal", "-route", "jev", "-entry", entry, "-out", out)
			ev.Record("process", map[string]any{"failed": err != nil, "output": string(output)})
			if !ev.Check("formal effect producer uses public Classify", err == nil, "process failed") {
				return
			}
			r := visualRead(t, out, "report.json")
			cat := ai.BuiltinCatalog()
			ev.Record("formal-effect-report", r)
			ev.Check("formal effect pins current catalog independently", r["catalogKind"] == "formal" && r["catalogHash"] == catalogHash(t, cat) && r["catalogVersion"] == cat.Version && r["realQuality"] == "NOT_RUN", "wrong identity or claims")
			_, err = visualCommand(t, "-verify", out)
			ev.Check("formal effects replay independently", err == nil, "replay failed")
			_, err = visualCommand(t, "-verify", out, "-require-live")
			ev.Check("controlled formal effect never real PASS", err != nil, "controlled promoted")
		})
	}
}
