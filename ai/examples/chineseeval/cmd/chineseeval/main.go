// chineseeval runs the host's synthetic Chinese task, independently of live
// smoke and the support matrix. It writes a fresh, audited evidence bundle.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var stage = "arguments"

const keyEnv = "BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY"

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL: Chinese evaluation stage="+stage)
		os.Exit(1)
	}
}

func execute() error {
	dataset := flag.String("dataset", "ai/examples/chineseeval/testdata/zh-support-v1.json", "versioned synthetic dataset")
	config := flag.String("config", "ai/examples/chineseeval/testdata/jev-1.13.0-v1.json", "pinned task and budgets")
	out := flag.String("out", "", "base directory for a new evidence bundle")
	live := flag.Bool("live", false, "enable real evaluation with BARNESS_AI_CHINESE_EVAL=1")
	verify := flag.String("verify", "", "verify an existing evidence bundle without network calls")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected arguments")
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
	plan, err := chineseeval.Load(data, cfg, ai.BuiltinCatalog())
	if err != nil {
		return err
	}
	if *verify != "" {
		return verifyBundle(plan, *verify)
	}
	stage = "credential_isolation"
	key := strings.TrimSpace(os.Getenv(keyEnv))
	if *live && os.Getenv("BARNESS_AI_CHINESE_EVAL") == "1" && foreignCredentials() {
		return errors.New("evaluation credentials must be isolated")
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
	run.SetVersion("dataset_hash", chineseeval.Hash(data))
	run.SetVersion("config_hash", chineseeval.Hash(cfg))
	run.SetVersion("model_catalog_version", plan.Config().CatalogVersion)
	run.SetVersion("model_catalog_hash", plan.Config().CatalogHash)
	run.SetVersion("typesafe_sdk", "direct-net-http-fixed-jev-1.13.0")
	// Keep exact source bytes, so reformatting cannot masquerade as the same hash.
	for name, raw := range map[string][]byte{"dataset.json": data, "config.json": cfg} {
		if err := os.WriteFile(filepath.Join(run.Dir(), name), raw, 0600); err != nil {
			return err
		}
	}
	report := plan.NotRun("explicit evaluation switches disabled")
	if *live && os.Getenv("BARNESS_AI_CHINESE_EVAL") == "1" {
		if key == "" {
			report = plan.NotRun("isolated evaluation credential missing")
		} else {
			stage = "account_alias"
			alias := os.Getenv("BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS")
			if !regexp.MustCompile(`^[A-Za-z0-9@._-]{1,96}$`).MatchString(alias) {
				return errors.New("account and region alias required")
			}
			observer := newObserver()
			stage = "client_assembly"
			client, transport, err := newClient(plan.Config(), key, observer)
			if err != nil {
				return err
			}
			defer transport.CloseIdleConnections()
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			report = plan.Run(ctx, client, ai.CallScope{TenantID: tenant, JobID: "chinese-evaluation"}, binding, "live", alias)
			// Preserve sent-attempt accounting before post-run evidence checks.
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
		}
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
	// Even FAIL/NOT_RUN need honest, complete accounting; no successful-only score.
	stage = "report_integrity"
	if err := plan.Verify(report); err != nil {
		return err
	}
	stage = "report_write"
	if err := run.Record("evaluation-report", report); err != nil {
		return err
	}
	if findings, err = run.Audit(); err != nil || len(findings) != 0 {
		return errors.New("final audit failed")
	}
	fmt.Println(report.Status + ": independently audited Chinese classification evaluation")
	stage = "sample_results"
	if report.Status == "FAIL" {
		return errors.New("incomplete evaluation")
	}
	return nil
}

func verifyBundle(plan *chineseeval.Plan, dir string) error {
	stage = "source_hashes"
	for name, want := range map[string]string{"dataset.json": plan.NotRun("verify").DatasetHash, "config.json": plan.NotRun("verify").ConfigHash} {
		raw, err := readBounded(filepath.Join(dir, name), 1<<20)
		if err != nil || chineseeval.Hash(raw) != want {
			return errors.New("source hash mismatch")
		}
	}
	stage = "report_integrity"
	raw, err := readBounded(filepath.Join(dir, "evaluation-report.json"), 16<<20)
	if err != nil {
		return err
	}
	report, err := plan.ReadReport(raw)
	if err != nil {
		return err
	}
	stage = "provenance_observations"
	manifestRaw, err := readBounded(filepath.Join(dir, "manifest.json"), 1<<20)
	if err != nil {
		return err
	}
	var observations []byte
	if report.Budget.Calls > 0 {
		observations, err = readBounded(filepath.Join(dir, "observations-evaluation.json"), 1<<20)
		if err != nil {
			return err
		}
	}
	if err := plan.VerifyEvidence(report, manifestRaw, observations); err != nil {
		return err
	}
	stage = "redaction_audit"
	auditRaw, err := readBounded(filepath.Join(dir, "audit.json"), 1<<20)
	if err != nil {
		return err
	}
	var result struct {
		Findings []audit.Finding `json:"findings"`
	}
	if json.Unmarshal(auditRaw, &result) != nil || result.Findings == nil || len(result.Findings) != 0 {
		return errors.New("missing successful audit")
	}
	findings, err := audit.Bundle(dir, audit.Environment())
	if err != nil || len(findings) != 0 {
		return errors.New("bundle audit failed")
	}
	fmt.Println(report.Status + ": artifact completeness and audit verified; evidence kind " + report.Kind)
	return nil
}

func readBounded(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil || int64(len(raw)) > max {
		return nil, errors.New("oversized input")
	}
	return raw, nil
}

func foreignCredentials() bool {
	names := regexp.MustCompile(`(?i)(KEY|TOKEN|SECRET|PASSWORD|PASSWD|CREDENTIAL|AUTH)`)
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if name != keyEnv && value != "" && names.MatchString(name) {
			return true
		}
	}
	return false
}
