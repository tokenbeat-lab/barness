package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

type multi []string

func (m *multi) String() string     { return fmt.Sprint([]string(*m)) }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

type options struct {
	Stage, From, Out, Bundle, RaceBundle        string
	Matrix, Extensions, Trace, Ledger, Snapshot string
	Jev, Luna, VisualJev, VisualDecisions       string
	Live                                        multi
	Preflight, CollectAll, WriteSnapshot        bool
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	o := options{}
	f := flag.NewFlagSet("releasegate", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.Stage, "stage", "check", "check (default), or explicitly selected race")
	f.StringVar(&o.From, "from", "", "completed check directory required by stage race")
	f.StringVar(&o.Out, "out", "", "new output directory")
	f.StringVar(&o.Bundle, "bundle", "", "read-only reevaluation of an existing offline bundle")
	f.StringVar(&o.RaceBundle, "race-bundle", "", "independent bundle for read-only reevaluation")
	f.StringVar(&o.Matrix, "matrix", "ai/live/support-matrix.json", "support matrix")
	f.StringVar(&o.Extensions, "extensions", "ai/release/classifier-extensions.json", "classifier extension registry")
	f.StringVar(&o.Trace, "trace", "ai/release/traceability.json", "traceability map")
	f.StringVar(&o.Ledger, "ledger", "ai/e2e/testdata/pidiff/ledger.json", "difference ledger")
	f.StringVar(&o.Snapshot, "snapshot", "ai/release/catalog-snapshot.json", "catalog snapshot")
	f.StringVar(&o.Jev, "evaluation-jev", "", "independent Jev Chinese bundle")
	f.StringVar(&o.Luna, "evaluation-luna", "", "independent Luna Chinese bundle")
	f.StringVar(&o.VisualJev, "visual-evaluation-jev", "", "formal Jev visual bundle")
	f.StringVar(&o.VisualDecisions, "visual-evaluation-decisions", "", "formal Decisions visual bundle")
	f.Var(&o.Live, "live", "live bundle (repeatable)")
	f.BoolVar(&o.Preflight, "preflight", false, "check external inputs without running tests")
	f.BoolVar(&o.CollectAll, "collect-all", false, "collect check-stage diagnostics; never starts race")
	f.BoolVar(&o.WriteSnapshot, "write-snapshot", false, "regenerate the catalog snapshot and exit")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	seen := map[string]bool{}
	f.Visit(func(v *flag.Flag) { seen[v.Name] = true })
	if f.NArg() != 0 || o.Stage != "check" && o.Stage != "race" {
		return o, errors.New("invalid stage or positional arguments")
	}
	if o.Preflight && o.CollectAll {
		return o, errors.New("preflight and collect-all are mutually exclusive")
	}
	if o.Bundle != "" || o.RaceBundle != "" {
		if o.Bundle == "" || seen["stage"] || o.From != "" || o.Preflight || o.CollectAll || o.WriteSnapshot {
			return o, errors.New("bundle reevaluation cannot execute stages")
		}
	}
	if o.Stage == "race" {
		if o.From == "" || o.Preflight || o.CollectAll || o.WriteSnapshot {
			return o, errors.New("race requires from and cannot use check controls")
		}
		for _, key := range []string{"matrix", "extensions", "trace", "ledger", "snapshot", "live", "evaluation-jev", "evaluation-luna", "visual-evaluation-jev", "visual-evaluation-decisions"} {
			if seen[key] {
				return o, errors.New("race uses original check inputs; overrides are forbidden")
			}
		}
	} else if o.From != "" {
		return o, errors.New("from is only valid for race")
	}
	if o.WriteSnapshot && (len(seen) > 2 || seen["out"] || seen["stage"] || o.Preflight || o.CollectAll) {
		return o, errors.New("write-snapshot is a standalone operation")
	}
	if o.From != "" && o.Out != "" {
		a, _ := filepath.Abs(o.From)
		b, _ := filepath.Abs(o.Out)
		if checkpointsOverlap(a, b) || checkpointsOverlap(b, a) {
			return o, errors.New("race output must be independent of check output")
		}
	}
	return o, nil
}
