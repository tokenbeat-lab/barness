// Command releasegate runs barness-ai's release gate (spec Testing Decisions
// §6, issue 24) from the repository root and writes a release bundle:
//
//	go run ./ai/release/cmd/releasegate [-live <live-bundle-dir>]...
//
// It runs go vet (with and without the live tag), the offline E2E with the
// frozen pi differential (BARNESS_AI_PIDIFF=1) and the policy pressure
// scenarios (BARNESS_AI_PRESSURE=1), and the suite under -race, each
// writing its evidence below the release bundle; audits every
// bundle for keys, Authorization values and real content; checks the
// difference ledger, the support matrix, the traceability map and the
// checked-in catalog snapshot; and writes release-report.{json,md} with the
// offline, differential and live results in separate sections. It exits 1
// unless every gate passes.
//
// -bundle evaluates an existing offline bundle without running anything;
// its command gates then fail as not run. -write-snapshot regenerates the
// catalog snapshot and exits.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

// tailBytes is how much of a failed command's output the report keeps.
const tailBytes = 4 << 10

func main() {
	out := flag.String("out", "", "release bundle directory (default .evidence/barness-ai-release/<time>)")
	bundle := flag.String("bundle", "", "evaluate this existing offline bundle instead of running the suite")
	matrixPath := flag.String("matrix", "ai/live/support-matrix.json", "support matrix")
	tracePath := flag.String("trace", "ai/release/traceability.json", "traceability map")
	ledgerPath := flag.String("ledger", "ai/e2e/testdata/pidiff/ledger.json", "difference ledger")
	snapshotPath := flag.String("snapshot", "ai/release/catalog-snapshot.json", "checked-in catalog snapshot")
	writeSnapshot := flag.Bool("write-snapshot", false, "regenerate the catalog snapshot and exit")
	evaluation := flag.String("evaluation", "", "independent complete Chinese evaluation bundle (no live calls are made by the gate)")
	var live multi
	flag.Var(&live, "live", "a live evidence bundle to audit (repeatable)")
	flag.Parse()

	snapshot, err := release.CatalogSnapshot()
	if err != nil {
		fail(err)
	}
	if *writeSnapshot {
		if err := os.WriteFile(*snapshotPath, snapshot, 0o644); err != nil {
			fail(err)
		}
		fmt.Println("wrote", *snapshotPath)
		return
	}
	if *out == "" {
		*out = filepath.Join(".evidence", "barness-ai-release", time.Now().UTC().Format("20060102T150405Z"))
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}
	trace, err := release.LoadTraceMap(*tracePath)
	if err != nil {
		fail(err)
	}
	ledger, err := release.LoadLedger(*ledgerPath)
	if err != nil {
		fail(err)
	}
	matrix, err := supportmatrix.LoadMatrix(*matrixPath)
	if err != nil {
		fail(err)
	}

	modules, err := release.GoModules(".")
	if err != nil {
		fail(err)
	}
	in := release.Inputs{Ledger: ledger, Matrix: matrix, Trace: trace, Modules: modules,
		Snapshots: []release.Snapshot{release.CheckSnapshot("catalog", *snapshotPath, snapshot), release.CheckReleasePlan(trace)}}
	var catalog struct {
		Hash    string
		Catalog struct{ Version string }
	}
	if err := json.Unmarshal(snapshot, &catalog); err != nil {
		fail(err)
	}
	in.CatalogVersion, in.CatalogHash = catalog.Catalog.Version, catalog.Hash
	audited := live
	var missing []release.AuditResult
	if *bundle == "" {
		offlineBase, raceBase := filepath.Join(*out, "offline"), filepath.Join(*out, "race")
		in.Commands = []release.Command{
			run(release.CmdVet, nil, "go", "vet", "./..."),
			run(release.CmdVetLive, nil, "go", "vet", "-tags", "live", "./..."),
			run(release.CmdTest, []string{"BARNESS_AI_PIDIFF=1", "BARNESS_AI_PRESSURE=1", evidenceDir(offlineBase)}, "go", "test", "-count=1", "./..."),
			run(release.CmdRace, []string{"BARNESS_AI_PIDIFF=1", "BARNESS_AI_PRESSURE=1", evidenceDir(raceBase)}, "go", "test", "-race", "-count=1", "./..."),
		}
		if *bundle, err = release.FindBundle(offlineBase); err != nil {
			fail(err)
		}
		if race, err := release.FindBundle(raceBase); err == nil {
			audited = append(audited, race)
		} else {
			// A race bundle that cannot be found cannot be audited: report
			// it rather than audit less.
			missing = append(missing, release.AuditResult{Bundle: raceBase, Err: err.Error()})
		}
	}
	if in.Offline, err = release.LoadBundle(*bundle); err != nil {
		fail(err)
	}
	forbidden := audit.Environment()
	if *evaluation != "" {
		audited = append(audited, *evaluation)
	}
	for _, dir := range append([]string{*bundle}, audited...) {
		in.RequiredAudits = append(in.RequiredAudits, dir)
		in.Audits = append(in.Audits, release.AuditBundle(dir, forbidden))
	}
	in.Audits = append(in.Audits, missing...)
	in.Evidence = append(release.DesignLoadEvidence(in.Offline), release.ChineseEffectEvidence(*evaluation))

	report := release.Evaluate(in)
	fixtures, err := release.Fixtures(".", "ai/e2e/testdata")
	if err != nil {
		fail(err)
	}
	write(*out, "release-report.json", report)
	write(*out, "fixtures.json", fixtures)
	writeRaw(*out, "release-report.md", []byte(report.Markdown()))
	writeRaw(*out, "catalog-snapshot.json", snapshot)
	for _, src := range []string{*matrixPath, *tracePath, *ledgerPath} {
		data, err := os.ReadFile(src)
		if err != nil {
			fail(err)
		}
		writeRaw(*out, filepath.Base(src), data)
	}
	fmt.Println(report.Markdown())
	fmt.Fprintf(os.Stderr, "barness-ai release bundle: %s\n", *out)
	if !report.Pass {
		os.Exit(1)
	}
}

func evidenceDir(base string) string {
	abs, err := filepath.Abs(base)
	if err != nil {
		fail(err)
	}
	return "BARNESS_AI_EVIDENCE_DIR=" + abs
}

// run executes one gate command, streaming its output to stderr and keeping
// its tail for the report.
func run(id string, env []string, name string, args ...string) release.Command {
	// The evidence directory is the release bundle's business, not the
	// command's meaning; only the switches are shown.
	var shown []string
	for _, kv := range env {
		if !strings.HasPrefix(kv, "BARNESS_AI_EVIDENCE_DIR=") {
			shown = append(shown, kv)
		}
	}
	line := strings.TrimSpace(strings.Join(shown, " ") + " " + name + " " + strings.Join(args, " "))
	fmt.Fprintf(os.Stderr, "releasegate: %s\n", line)
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	// One writer for both streams, so exec copies them on one goroutine.
	out := &tee{&buf}
	cmd.Stdout, cmd.Stderr = out, out
	started := time.Now()
	err := cmd.Run()
	c := release.Command{ID: id, Line: line, Passed: err == nil, Duration: time.Since(started)}
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		fmt.Fprintln(os.Stderr, "releasegate:", err)
	}
	if !c.Passed {
		tail := buf.Bytes()
		c.Tail = string(tail[max(0, len(tail)-tailBytes):])
	}
	return c
}

type tee struct{ buf *bytes.Buffer }

func (t *tee) Write(p []byte) (int, error) {
	t.buf.Write(p)
	return os.Stderr.Write(p)
}

func write(dir, name string, v any) {
	data, err := release.Encode(v)
	if err != nil {
		fail(err)
	}
	writeRaw(dir, name, data)
}

func writeRaw(dir, name string, data []byte) {
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "releasegate:", err)
	os.Exit(2)
}
