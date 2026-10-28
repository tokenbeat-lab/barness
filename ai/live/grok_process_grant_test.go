//go:build live

package live

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The reviewed frozen build embeds the selected family native source hash.
var legacyBuildSource string

var glmApprovalDeadline time.Time

type processAuthorization struct {
	grant livegrant.Grant
	raw   []byte
	claim livegrant.Receipt
}

func loadProcessAuthorization(route, mode, stage, account, region, path string) (*processAuthorization, error) {
	if path == "" {
		return nil, fmt.Errorf("fresh single-process grant is required")
	}
	raw, e := readDecisionsFile(path, 1<<20)
	if e != nil {
		return nil, fmt.Errorf("fresh single-process grant unreadable")
	}
	keyName := keyVar(route)
	if strings.HasPrefix(route, "zhipu-") {
		keyName = "BARNESS_GLM_KEY"
	}
	if strings.HasPrefix(route, "minimax-") {
		keyName = "BARNESS_MINIMAX_KEY"
	}
	if key := strings.TrimSpace(os.Getenv(keyName)); key != "" && strings.Contains(string(raw), key) {
		return nil, fmt.Errorf("process approval must not contain a credential")
	}
	g, e := livegrant.Decode(raw)
	if e != nil {
		return nil, e
	}
	exe, e := os.Executable()
	if e != nil {
		return nil, e
	}
	binary, e := livegrant.FileHash(exe)
	if e != nil {
		return nil, e
	}
	want, e := supportmatrix.ProcessScope(route, mode, stage, binary)
	if strings.HasPrefix(route, "minimax-") && len(g.Scope.Scenes) == 1 {
		id, err := minimaxlive.Select(strings.TrimPrefix(route, "minimax-"), minimaxReport.Identity.Model, minimaxReport.Identity.Purpose)
		if err != nil || stage != "candidate" {
			return nil, fmt.Errorf("invalid thinking-on stage/identity")
		}
		want, e = minimaxlive.ThinkingOnScope(id, mode, binary, g.Scope.BaselineManifestSHA256)
	}

	if e != nil {
		return nil, e
	}
	if legacyBuildSource == "" || legacyBuildSource != want.SourceSHA256 {
		return nil, fmt.Errorf("fresh producer binary was not built from this frozen native source")
	}
	if g.AccountAlias != account || g.RegionAlias != region {
		return nil, fmt.Errorf("selected account/region differs from fresh process approval")
	}
	if e = g.Check(want, time.Now()); e != nil {
		return nil, e
	}
	return &processAuthorization{grant: g, raw: raw}, nil
}

func (a *processAuthorization) consume() error {
	if a == nil {
		return fmt.Errorf("fresh process approval absent")
	}
	if !a.claim.StartedAt.IsZero() {
		return fmt.Errorf("fresh process approval already claimed in this process")
	}
	source, err := supportmatrix.RegionalExecutionSource()
	if strings.HasPrefix(a.grant.Scope.Route, "minimax-") {
		source, err = minimaxlive.ExecutionSource()
	}
	if strings.HasPrefix(a.grant.Scope.Route, "zhipu-") {
		source, err = supportmatrix.GLMExecutionSource()
	}
	if err != nil {
		return err
	}
	if err := run.Record("execution-source", source); err != nil {
		return err
	}
	r, e := livegrant.Claim(a.grant, a.raw, filepath.Base(run.Dir()), time.Now())
	if e != nil {
		return e
	}
	a.claim = r
	run.SetVersion("process_build_source", legacyBuildSource)
	if e = os.WriteFile(filepath.Join(run.Dir(), "process-grant.json"), a.raw, 0600); e != nil {
		return e
	}
	return run.Record("process-authorization-receipt", r)
}

func claimGenericProcess() error {
	if cfg.authorization == nil {
		return fmt.Errorf("fresh generic process authorization absent")
	}
	return cfg.authorization.consume()
}

func processDeadline() time.Time {
	deadline := started.Add(1780 * time.Second)
	if cfg.authorization != nil && cfg.authorization.grant.ExpiresAt.Before(deadline) {
		deadline = cfg.authorization.grant.ExpiresAt
	}
	return deadline
}
