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
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

// ChineseEffectsEvidence requires two independently attributed current runs.
// Checking identities also rejects byte-identical copies under different paths.
func ChineseEffectsEvidence(jev, luna string) []EvidenceCheck {
	out := []EvidenceCheck{ChineseEffectEvidence("jev", jev), ChineseEffectEvidence("luna", luna)}
	if jev != "" && luna != "" {
		a, errA := os.Stat(jev)
		b, errB := os.Stat(luna)
		if errA == nil && errB == nil && os.SameFile(a, b) {
			for i := range out {
				out[i].Status = Fail
				out[i].Detail = "the two routes cannot reuse one evaluation bundle"
			}
		}
	}
	return out
}

// ChineseEffectEvidence imports the fixed task at the public verifier seam.
// The recorded-snapshot verifier remains useful independently for old Jev runs;
// release applicability additionally pins the current task, catalog and source.
func ChineseEffectEvidence(routeName, dir string) EvidenceCheck {
	e := EvidenceCheck{Name: "chinese-effect-" + routeName, Bundle: dir, Status: NotRun, Detail: "independent real evaluation not supplied"}
	if dir == "" {
		return e
	}
	e.Status = Fail
	data, err := readArtifact(filepath.Join(dir, "dataset.json"), 1<<20)
	if err != nil {
		e.Detail = "dataset missing"
		return e
	}
	config, err := readArtifact(filepath.Join(dir, "config.json"), 16384)
	if err != nil {
		e.Detail = "config missing"
		return e
	}
	plan, err := chineseeval.Load(data, config, ai.BuiltinCatalog())
	if err != nil {
		e.Detail = "dataset/config/catalog drift"
		return e
	}
	route, err := chineseeval.RouteNamed(routeName)
	if err != nil || plan.Route() != route {
		e.Detail = "evaluation route identity mismatch"
		return e
	}
	// These independent fixed-task pins exclude only the two evolving global
	// catalog fields already validated by Plan.Load. Every business field,
	// including instructions, route and exact timeout/budget, stays pinned.
	var businessConfig map[string]any
	if json.Unmarshal(config, &businessConfig) != nil {
		e.Detail = "fixed dataset/config drift"
		return e
	}
	delete(businessConfig, "catalogVersion")
	delete(businessConfig, "catalogHash")
	fixedConfig, err := json.Marshal(businessConfig)
	if err != nil {
		e.Detail = "fixed dataset/config drift"
		return e
	}
	configHash := map[string]string{"jev": "sha256:97b38ad7127e2af34089ed051f3605c175ecaf3a4a97ce8c9258fb60f82a930f", "luna": "sha256:de6491e41406eb18c95a68ca10d74e6b03c7417578a30a7b096560ef61ab8ea1"}[routeName]
	if chineseeval.Hash(data) != "sha256:3380b6a4033680008446320daefd516d92e38ac9f04d618e6bdc9af25de6f0eb" || chineseeval.Hash(fixedConfig) != configHash {
		e.Detail = "fixed dataset/config drift"
		return e
	}
	raw, err := readArtifact(filepath.Join(dir, "evaluation-report.json"), 16<<20)
	if err != nil {
		e.Detail = "evaluation report missing"
		return e
	}
	r, err := plan.ReadReport(raw)
	if err != nil || r.Schema != 2 {
		e.Detail = "evaluation integrity or current schema failed"
		return e
	}
	root, err := sourceRoot()
	if err != nil {
		e.Detail = "source root unavailable"
		return e
	}
	if err = currentEvaluationSource(root, dir, r.CodeHash); err != nil {
		e.Detail = "evaluation execution source is stale or incomplete"
		return e
	}
	// Go is already required by releasegate. Run the existing public verifier in
	// a credential-free, bounded process; it reconciles file inventory, captures,
	// public-client replay, Observer, recorded source hash and both audit layers.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	absolute, err := filepath.Abs(dir)
	if err != nil {
		e.Detail = "invalid bundle location"
		return e
	}
	cmd := exec.CommandContext(ctx, "go", "run", "./ai/examples/chineseeval/cmd/chineseeval", "-route", routeName, "-verify", absolute)
	cmd.Dir = root
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	if cmd.Run() != nil {
		e.Detail = "evaluation bundle inventory, captures, provenance or audit failed"
		return e
	}
	e.Detail = fmt.Sprintf("independent Chinese task; %d/%d answered; accuracy %s; conditional accuracy %s; ECE %s; Brier %s; cost USD %.9f; no quality threshold", r.Metrics.Scored, r.Dataset.Count, metricJSON(r.Metrics.Accuracy), metricJSON(r.Metrics.ScoredAccuracy), metricJSON(r.Metrics.ConfidenceECE), metricJSON(r.Metrics.Brier), r.Budget.KnownCostUSD)
	if r.Kind != "live" || r.Status != "COMPLETE" {
		e.Status = NotRun
		if r.Status == "FAIL" {
			e.Status = Fail
		}
		e.Detail = r.Kind + "/" + r.Status + "; " + e.Detail
		return e
	}
	e.Status = Pass
	return e
}
func metricJSON(v *float64) string { raw, _ := json.Marshal(v); return string(raw) }
