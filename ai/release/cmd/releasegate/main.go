// Command releasegate checks ordinary evidence first. An explicit -stage race
// -from <check-dir> invocation performs the independent final race stage.
// Read-only -bundle reevaluation cannot certify unexecuted command gates.
package main

import (
	"context"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(cli(ctx, os.Args[1:]))
}
func cli(ctx context.Context, args []string) int {
	o, err := parseOptions(args, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "releasegate:", err)
		return 2
	}
	if o.WriteSnapshot {
		raw, err := release.CatalogSnapshot()
		if err == nil {
			err = os.WriteFile(o.Snapshot, raw, 0644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "releasegate: snapshot write failed")
			return 1
		}
		return 0
	}
	if o.Out == "" {
		o.Out = filepath.Join(".evidence", "barness-ai-release", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	if err := newOutput(o.Out); err != nil {
		fmt.Fprintln(os.Stderr, "releasegate:", err)
		return 1
	}
	if o.Stage == "race" {
		return resumeRace(ctx, o)
	}
	root, rootErr := os.Getwd()
	var frozen frozenInputs
	var freezeErr error
	if o.Bundle == "" && rootErr == nil {
		frozen, freezeErr = captureCheckpointInputs(root, o)
	}
	in, inputErr := externalInputs(o)
	if inputErr == nil && (rootErr != nil || freezeErr != nil) {
		inputErr = fmt.Errorf("source/input freeze unavailable")
	}
	if o.Bundle != "" {
		if err := loadOrdinary(&in, o.Bundle); err != nil {
			in.Offline.Problems = append(in.Offline.Problems, err.Error())
		}
		if o.RaceBundle != "" {
			in.RequiredAudits = append(in.RequiredAudits, o.RaceBundle)
			in.Audits = append(in.Audits, release.AuditBundle(o.RaceBundle, audit.Environment()))
		}
		in.Evidence = append(in.Evidence, release.RaceEvidence(o.RaceBundle, in.Trace, in.Ledger)...)
		_, err := writeVerdict(o, in)
		if err != nil {
			fmt.Fprintln(os.Stderr, "releasegate: report write failed")
		}
		return 1
	}
	p := release.Preflight(in)
	if inputErr != nil {
		p.Ready = false
		p.Checks = append(p.Checks, release.Gate{ID: "input-loading", Name: "load fixed release inputs", Problems: []string{inputErr.Error()}})
	}
	if err := writeJSON(o.Out, "preflight-report.json", p); err != nil {
		fmt.Fprintln(os.Stderr, "releasegate: preflight report write failed")
		return 1
	}
	if !p.Ready || o.Preflight {
		r := newExecution("check")
		if !p.Ready {
			r.fail("preflight", fmt.Errorf("external inputs not ready"))
		} else {
			r.Status = "NOT_RUN"
			r.BlockedBy = "preflight-only"
		}
		if err := r.save(o.Out); err != nil {
			fmt.Fprintln(os.Stderr, "releasegate: execution report write failed")
			return 1
		}
		_, err := writeVerdict(o, in)
		if err != nil {
			fmt.Fprintln(os.Stderr, "releasegate: report write failed")
			return 1
		}
		if p.Ready {
			fmt.Fprintln(os.Stderr, "preflight ready; release remains unverified")
			return 0
		}
		fmt.Fprintln(os.Stderr, "preflight failed; no verification command started")
		return 1
	}
	return checkStage(ctx, o, in, frozen)
}
func newOutput(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("output directory unavailable")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		return fmt.Errorf("output directory must be empty")
	}
	return nil
}
