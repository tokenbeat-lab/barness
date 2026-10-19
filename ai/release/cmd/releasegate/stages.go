package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func checkStage(ctx context.Context, o options, in release.Inputs, frozen frozenInputs) int {
	root, err := os.Getwd()
	if err != nil {
		return 1
	}
	guard := func() error { return verifyFrozen(root, o, frozen) }
	bundle := ""
	consume := func() error {
		var err error
		bundle, err = release.FindBundle(filepath.Join(o.Out, "offline"))
		if err != nil {
			return errors.New("ordinary native bundle missing")
		}
		if err = loadOrdinary(&in, bundle); err != nil {
			return err
		}
		checks := release.OrdinaryChecks(in)
		if err = writeJSON(o.Out, "ordinary-report.json", checks); err != nil {
			return errors.New("ordinary report write failed")
		}
		if slices.ContainsFunc(checks, func(g release.Gate) bool { return !g.Pass }) {
			return errors.New("ordinary native evidence or audit failed")
		}
		return nil
	}
	r := runCheckCommands(ctx, root, o.Out, o.CollectAll, guard, consume)
	if r.Status == "READY_FOR_RACE" {
		if err := writeInputCopies(o); err != nil {
			r.fail("metadata", err)
		}
		if r.Status != "FAIL" {
			if err := saveCheckpoint(root, o.Out, o, frozen, bundle); err != nil {
				r.fail("checkpoint", err)
			}
		}
	}
	code := persistStage(o, in, &r)
	if code == 2 {
		fmt.Fprintln(os.Stderr, "READY_FOR_RACE: ordinary verification complete; race NOT_RUN; release not certified")
	}
	return code
}

func resumeRace(ctx context.Context, o options) int {
	r := newExecution("race")
	root, err := os.Getwd()
	in := release.Inputs{}
	reject := func(id string, err error) int { r.fail(id, err); return persistStage(o, in, &r) }
	if err != nil {
		return reject("source", errors.New("source root unavailable"))
	}
	if checkpointsOverlap(o.From, o.Out) || checkpointsOverlap(o.Out, o.From) {
		return reject("output", errors.New("race and check directories must be independent"))
	}
	c, prior, err := loadCheckpoint(root, o.From)
	if err != nil {
		return reject("checkpoint", checkpointError(err))
	}
	r.Steps = slices.Clone(prior.Steps)
	in, err = externalInputs(c.Inputs)
	if err != nil {
		return reject("preflight", err)
	}
	p := release.Preflight(in)
	if err := writeJSON(o.Out, "preflight-report.json", p); err != nil {
		return reject("report", errors.New("preflight report write failed"))
	}
	if !p.Ready {
		return reject("preflight", errors.New("external inputs no longer ready"))
	}
	if err := loadOrdinary(&in, filepath.Join(o.From, filepath.FromSlash(c.Bundle))); err != nil {
		return reject("ordinary-evidence", err)
	}
	if slices.ContainsFunc(release.OrdinaryChecks(in), func(g release.Gate) bool { return !g.Pass }) {
		return reject("ordinary-evidence", errors.New("ordinary native evidence failed revalidation"))
	}
	if _, _, err := loadCheckpoint(root, o.From); err != nil {
		return reject("checkpoint", err)
	}
	r.Steps[4].Status = "RUNNING"
	r.Race = "RUNNING"
	if err := r.save(o.Out); err != nil {
		return reject("report", errors.New("execution report unavailable"))
	}
	r.Steps[4] = runTool(ctx, root, o.Out, "go-test-race")
	r.Race = r.Steps[4].Status
	if r.Race != "PASS" {
		return reject("go-test-race", errors.New("race process or timing failed"))
	}
	if _, _, err := loadCheckpoint(root, o.From); err != nil {
		return reject("checkpoint", err)
	}
	bundle, err := release.FindBundle(filepath.Join(o.Out, "race"))
	if err != nil {
		return reject("race-evidence", errors.New("independent race native bundle missing"))
	}
	in.RequiredAudits = append(in.RequiredAudits, bundle)
	in.Audits = append(in.Audits, release.AuditBundle(bundle, audit.Environment()))
	in.Evidence = append(in.Evidence, release.RaceEvidence(bundle, in.Trace, in.Ledger)...)
	if _, _, err := loadCheckpoint(root, o.From); err != nil {
		return reject("checkpoint", err)
	}
	return persistStage(o, in, &r)
}

func persistStage(o options, in release.Inputs, r *executionReport) int {
	if r.Status == "FAIL" {
		in.Snapshots = append(in.Snapshots, release.Snapshot{Name: "execution-continuity", Detail: r.BlockedBy + ": " + r.Error})
	}
	in.Commands = r.commands()
	pass, err := writeVerdict(o, in)
	if err != nil {
		r.fail("report", errors.New("release report write failed"))
	}
	if r.Stage == "race" && r.Status != "FAIL" {
		if pass {
			r.Status = "PASS"
			r.ReleasePass = true
		} else {
			r.fail("release", errors.New("required release gates did not pass"))
		}
	}
	if err := r.save(o.Out); err != nil {
		fmt.Fprintln(os.Stderr, "releasegate: execution report write failed")
		r.fail("report", errors.New("execution report write failed"))
		r.ReleasePass = false
		in.Snapshots = append(in.Snapshots, release.Snapshot{Name: "execution-report", Detail: "execution report write failed"})
		_, _ = writeVerdict(o, in)
	}
	if r.Status == "FAIL" {
		fmt.Fprintln(os.Stderr, "releasegate:", r.BlockedBy, r.Error)
	}
	return r.ExitCode()
}

func writeInputCopies(o options) error {
	fixtures, err := release.Fixtures(".", "ai/e2e/testdata")
	if err != nil {
		return errors.New("fixture inventory unavailable")
	}
	if err := writeJSON(o.Out, "fixtures.json", fixtures); err != nil {
		return err
	}
	for name, path := range map[string]string{"support-matrix.json": o.Matrix, "traceability.json": o.Trace, "ledger.json": o.Ledger, "catalog-snapshot.json": o.Snapshot, "classifier-extensions.json": o.Extensions} {
		raw, err := readRegular(path, checkpointLimit)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(o.Out, name), raw, 0600); err != nil {
			return err
		}
	}
	return nil
}
