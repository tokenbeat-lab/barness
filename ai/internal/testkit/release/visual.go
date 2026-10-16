package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// VisualEffectsEvidence imports ticket 12's independent fixed labeled tasks.
// Text effects and protocol smoke are separate obligations. The existing public
// verifier owns exact data, scoring, source, capture/replay and budget validation.
func VisualEffectsEvidence(jev, decisions string) []EvidenceCheck {
	return []EvidenceCheck{visualEffect("jev", jev), visualEffect("decisions", decisions)}
}
func visualEffect(route, dir string) EvidenceCheck {
	e := EvidenceCheck{Name: "visual-effect-" + route, Bundle: dir, Status: NotRun, Detail: "independent current-formal visual task not supplied"}
	if dir == "" {
		return e
	}
	e.Status = Fail
	var r struct {
		Schema                                                                            int
		CatalogKind, Route, Model, CatalogHash, CatalogVersion, Mode, Status, RealQuality string
		Metrics                                                                           struct {
			Common struct{ Total, ClassCorrect, ExistCorrect int }
		}
	}
	raw, err := readArtifact(filepath.Join(dir, "report.json"), 16<<20)
	if err != nil || json.Unmarshal(raw, &r) != nil {
		e.Detail = "visual report missing or invalid"
		return e
	}
	cat := ai.BuiltinCatalog()
	hash, err := cat.Hash()
	model := "jev-1.14.0"
	if route == "decisions" {
		model = "gpt-6-luna"
	}
	if err != nil || r.Schema != 1 || r.Route != route || r.Model != model || r.CatalogKind != "formal" || r.CatalogHash != hash || r.CatalogVersion != cat.Version {
		e.Detail = "visual route/model/current formal catalog mismatch"
		return e
	}
	root, err := sourceRoot()
	if err != nil {
		e.Detail = "visual source root unavailable"
		return e
	}
	absolute, err := filepath.Abs(dir)
	if err != nil {
		e.Detail = "invalid visual bundle location"
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	temporary, err := os.MkdirTemp("", "barness-visual-verifier-")
	if err != nil {
		e.Detail = "visual verifier workspace unavailable"
		return e
	}
	defer os.RemoveAll(temporary)
	binary := filepath.Join(temporary, "visualeval")
	var environment []string
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if value, ok := os.LookupEnv(name); ok {
			environment = append(environment, name+"="+value)
		}
	}
	// Run the built verifier directly: killing `go run` would leave its child
	// executable alive. Build and verification share one finite deadline.
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./ai/examples/visualeval")
	build.Dir, build.Env = root, environment
	if build.Run() != nil {
		e.Detail = "visual verifier build failed"
		return e
	}
	args := []string{"-verify", absolute}
	live := r.Mode == "live" && r.Status == "COMPLETE" && r.RealQuality == "COMPLETE"
	if live {
		args = append(args, "-require-live")
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir, cmd.Env = root, environment
	if cmd.Run() != nil {
		e.Detail = "visual source, fixed inputs, captures, replay, metrics, budget or audit rejected"
		return e
	}
	e.Detail = fmt.Sprintf("%s/%s; fixed 12-sample task, common %d samples, classification %d correct, existence %d correct; no quality threshold", r.Mode, r.Status, r.Metrics.Common.Total, r.Metrics.Common.ClassCorrect, r.Metrics.Common.ExistCorrect)
	e.Status = NotRun
	if r.Status == "FAIL" {
		e.Status = Fail
	} else if live {
		e.Status = Pass
	}
	return e
}
