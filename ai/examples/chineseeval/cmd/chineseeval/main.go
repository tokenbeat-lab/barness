// chineseeval is a host-owned task evaluation, independent of support smoke.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var stage = "arguments"

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL: Chinese evaluation stage="+stage)
		os.Exit(1)
	}
}

func execute() error {
	dataset := flag.String("dataset", "ai/examples/chineseeval/testdata/zh-support-v1.json", "fixed synthetic dataset")
	config := flag.String("config", "", "pinned task and budgets (default preserves Jev schema 1)")
	catalog := flag.String("catalog", "", "host-owned catalog snapshot (default builtin)")
	routeName := flag.String("route", "", "explicit jev or luna route")
	fixture := flag.String("fixture", "", "offline profile: normal, mixed, all-refusal, usage-absent, usage-partial, cancel, budget, response-alias")
	out := flag.String("out", "", "base directory for a fresh bundle")
	live := flag.Bool("live", false, "enable real evaluation with BARNESS_AI_CHINESE_EVAL=1")
	proxy := flag.String("proxy", "", "explicit loopback HTTP proxy for live HTTPS requests")
	verify := flag.String("verify", "", "verify a bundle with its own recorded snapshot")
	flag.Parse()
	if flag.NArg() != 0 || (*fixture != "" && *live) {
		return errors.New("incompatible arguments")
	}
	stage = "proxy"
	proxyURL, err := evaluationProxy(*proxy)
	if err != nil || (*proxy != "" && (!*live || *verify != "")) {
		return errors.New("invalid live proxy configuration")
	}
	if *verify != "" {
		return verifyBundle(*verify, *routeName)
	}
	stage = "task_pins"
	var selected chineseeval.Route
	if *routeName != "" {
		var err error
		selected, err = chineseeval.RouteNamed(*routeName)
		if err != nil {
			return err
		}
	}
	if *config == "" {
		*config = "ai/examples/chineseeval/testdata/jev-1.13.0-v1.json"
		if *routeName != "" {
			*config = "ai/examples/chineseeval/testdata/" + *routeName + "-route-v1.json"
		}
	}
	stage = "dataset_read"
	data, err := readBounded(*dataset, 1<<20)
	if err != nil {
		return err
	}
	stage = "config_read"
	cfg, err := readBounded(*config, 16384)
	if err != nil {
		return err
	}
	stage = "task_pins"
	var plan *chineseeval.Plan
	if *catalog == "" {
		plan, err = chineseeval.Load(data, cfg, ai.BuiltinCatalog())
	} else {
		raw, readErr := readBounded(*catalog, 2<<20)
		if readErr != nil {
			return readErr
		}
		plan, err = chineseeval.LoadRecorded(data, cfg, raw)
	}
	if err != nil {
		return err
	}
	route := plan.Route()
	if *routeName != "" && route != selected {
		return errors.New("route/config mismatch")
	}
	if *fixture != "" && !fixtureValid(*fixture, route) {
		return errors.New("invalid fixture for route")
	}
	stage = "credential_isolation"
	key := ""
	enabled := *live && os.Getenv("BARNESS_AI_CHINESE_EVAL") == "1"
	if enabled {
		if foreignCredentials(route.KeyEnv()) {
			return errors.New("evaluation credentials must be isolated")
		}
		key = strings.TrimSpace(os.Getenv(route.KeyEnv()))
	}
	if *out != "" {
		if err := os.Setenv(evidence.DirEnv, *out); err != nil {
			return err
		}
	}
	stage = "evidence_directory"
	run, err := evidence.NewRun("host Chinese evaluation (not support smoke)")
	if err != nil {
		return err
	}
	if err := os.Chmod(run.Dir(), 0700); err != nil {
		return err
	}
	if key != "" {
		run.RedactSecret(key, "[EVALUATION-KEY]")
	}
	kind := "live"
	if *fixture != "" {
		kind = "fixture"
	}
	run.SetVersion("dataset_hash", chineseeval.Hash(data))
	run.SetVersion("config_hash", chineseeval.Hash(cfg))
	run.SetVersion("model_catalog_version", plan.Config().CatalogVersion)
	run.SetVersion("model_catalog_hash", plan.Config().CatalogHash)
	run.SetVersion("typesafe_sdk", route.SDK())
	run.SetVersion("classifier_sdk", route.SDK())
	run.SetVersion("route_hash", route.Hash())
	run.SetVersion("evidence_kind", kind)
	run.SetVersion("fixture_profile", *fixture)
	for name, raw := range map[string][]byte{"dataset.json": data, "config.json": cfg} {
		if err := os.WriteFile(filepath.Join(run.Dir(), name), raw, 0600); err != nil {
			return err
		}
	}
	if err := run.Record("catalog", plan.Catalog()); err != nil {
		return err
	}
	code, err := codeSnapshot()
	if err != nil {
		return err
	}
	if err := run.Record("code-snapshot", code); err != nil {
		return err
	}
	codeRaw, err := readBounded(filepath.Join(run.Dir(), "code-snapshot.json"), 1<<20)
	if err != nil {
		return err
	}
	codeHash := chineseeval.Hash(codeRaw)
	run.SetVersion("code_hash", codeHash)
	report := plan.NotRun("explicit evaluation switches disabled")
	if enabled && key == "" {
		report = plan.NotRun("isolated evaluation credential missing")
	}
	if *fixture != "" || (enabled && key != "") {
		alias := "synthetic@offline"
		if kind == "live" {
			stage = "account_alias"
			alias = os.Getenv("BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS")
			if !regexp.MustCompile(`^[A-Za-z0-9@._-]{1,96}$`).MatchString(alias) || strings.HasPrefix(alias, "synthetic@") {
				return errors.New("account and region alias required")
			}
		}
		observer := newObserver()
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		var f *fixtureTransport
		var injected http.RoundTripper
		if kind == "fixture" {
			f = &fixtureTransport{route: route, profile: *fixture, cancel: cancel}
			injected = f
			key = "synthetic-fixture-credential"
		}
		var capture *captureTransport
		if plan.Config().Route != nil {
			capture = &captureTransport{kind: kind}
		}
		stage = "client_assembly"
		client, transport, err := newClient(plan, key, observer, injected, capture, proxyURL)
		if err != nil {
			return err
		}
		defer transport.CloseIdleConnections()
		report = plan.Run(ctx, client, ai.CallScope{TenantID: tenant, JobID: "chinese-evaluation"}, binding, kind, alias)
		if report.Schema == 2 {
			report.CodeHash = codeHash
		}
		if err := run.Record("evaluation-report", report); err != nil {
			return err
		}
		stage = "observer_completion"
		records, err := observer.finished(report.Budget.Calls)
		if err != nil {
			return err
		}
		if err := run.Record("observations-evaluation", records); err != nil {
			return err
		}
		if stats := client.ObserverStats(); stats.Dropped != 0 || stats.Failed != 0 {
			return errors.New("observer evidence incomplete")
		}
		if capture != nil {
			if err := run.Record("captures", capture.snapshot()); err != nil {
				return err
			}
		}
		if f != nil {
			if err := run.Record("fixture-captures", f.captures); err != nil {
				return err
			}
		}
	}
	if report.Schema == 2 {
		report.CodeHash = codeHash
	}
	stage = "report_write"
	if err := run.Record("evaluation-report", report); err != nil {
		return err
	}
	if err := run.Finish(); err != nil {
		return err
	}
	stage = "redaction_audit"
	findings, err := run.Audit()
	if err != nil || len(findings) != 0 {
		return errors.New("redaction audit failed")
	}
	report.Audit = "PASS"
	stage = "report_integrity"
	if err := plan.Verify(report); err != nil {
		return err
	}
	if err := run.Record("evaluation-report", report); err != nil {
		return err
	}
	if findings, err = run.Audit(); err != nil || len(findings) != 0 {
		return errors.New("final audit failed")
	}
	if err := freezeFiles(run.Dir()); err != nil {
		return err
	}
	if err := verifyBundle(run.Dir(), *routeName); err != nil {
		return err
	}
	fmt.Println(report.Status + ": independently audited Chinese classification evaluation")
	stage = "sample_results"
	if report.Status == "FAIL" {
		return errors.New("incomplete evaluation")
	}
	return nil
}

func foreignCredentials(allowed string) bool {
	names := regexp.MustCompile(`(?i)(KEY|TOKEN|SECRET|PASSWORD|PASSWD|CREDENTIAL|AUTH)`)
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if name != allowed && value != "" && names.MatchString(name) {
			return true
		}
	}
	return false
}
