package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

func verifyBundle(dir, routeName string) error {
	stage = "source_hashes"
	data, err := readBounded(filepath.Join(dir, "dataset.json"), 1<<20)
	if err != nil {
		return err
	}
	config, err := readBounded(filepath.Join(dir, "config.json"), 16384)
	if err != nil {
		return err
	}
	snapshot, err := readBounded(filepath.Join(dir, "catalog.json"), 2<<20)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	plan, err := chineseeval.LoadRecorded(data, config, snapshot)
	if err != nil {
		return err
	}
	if routeName != "" {
		route, err := chineseeval.RouteNamed(routeName)
		if err != nil || route != plan.Route() {
			return errors.New("recorded route mismatch")
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
	if report.Schema == 2 {
		stage = "file_inventory"
		if err := verifyFiles(dir); err != nil {
			return err
		}
		stage = "code_snapshot"
		if err := verifyCode(dir, report.CodeHash); err != nil {
			return err
		}
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
	captureRaw, captureErr := readBounded(filepath.Join(dir, "fixture-captures.json"), 1<<20)
	if report.Kind == "fixture" {
		stage = "fixture_identity"
		if captureErr != nil || verifyFixture(plan, report, captureRaw) != nil {
			return errors.New("fixture captures invalid")
		}
	} else if captureErr == nil || !os.IsNotExist(captureErr) {
		return errors.New("fixture cannot become live evidence")
	}
	if report.Schema == 2 && report.Budget.Calls > 0 {
		stage = "transport_captures"
		raw, err := readBounded(filepath.Join(dir, "captures.json"), 1<<20)
		if err != nil {
			return err
		}
		if err := verifyWire(plan, report, raw); err != nil {
			return err
		}
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
	fmt.Println(report.Status + ": recorded snapshot and audit verified; evidence kind " + report.Kind)
	return nil
}
