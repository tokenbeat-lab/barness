package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func runCheckApply(ctx context.Context, command string, args []string) (any, error) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	store := flags.String("store", ".catalog", "local maintenance store")
	id := flags.String("id", "", "immutable candidate ID")
	root := flags.String("repo", ".", "selected repository with fixed public contract")
	feature := flags.String("evidence-feature", "", "explicit .evidence/<feature> output scope")
	timeout := flags.Duration("timeout", 5*time.Minute, "bounded controlled check duration")
	reference := flags.String("check", "", "retained check.json reference for apply")
	output := flags.String("out", "", "optional complete runtime JSON output")
	dry := flags.Bool("dry-run", false, "read-only preview")
	confirm := flags.Bool("confirm", false, "explicitly confirm this operation")
	_ = flags.Bool("json", false, "machine-readable JSON (default)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return map[string]any{"status": "help", "command": command}, nil
		}
		return nil, &maintenance.Error{Category: "invalid_argument", Message: err.Error()}
	}
	if flags.NArg() != 0 {
		return nil, &maintenance.Error{Category: "invalid_argument", Message: "unexpected positional arguments"}
	}
	if command == "check" {
		return maintenance.Check(ctx, maintenance.CheckRequest{Store: *store, CandidateID: *id, RepositoryRoot: *root, EvidenceFeature: *feature, Timeout: *timeout, Confirmed: *confirm, DryRun: *dry})
	}
	revision, err := maintenance.Apply(ctx, maintenance.ApplyRequest{Store: *store, CandidateID: *id, CheckReference: *reference, RepositoryRoot: *root, OutputPath: *output, Confirmed: *confirm, DryRun: *dry})
	status := "applied"
	if *dry {
		status = "preview"
	}
	return map[string]any{"status": status, "revision": revision}, err
}
