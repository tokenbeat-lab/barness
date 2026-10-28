package release

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

// EvaluateRecord only reads source, receipts and native evidence. It never runs
// verification commands or grants permission to run expensive stages.
func EvaluateRecord(root, path string) RecordReport {
	report := RecordReport{Checks: []RecordCheck{}}
	check := func(id string, err error) {
		c := RecordCheck{ID: id, Pass: err == nil}
		if err != nil {
			c.Problem = err.Error()
		}
		report.Checks = append(report.Checks, c)
	}
	var record Record
	if err := recordJSON(path, &record); err != nil {
		check("record", fmt.Errorf("invalid release record"))
		return report
	}
	valid := regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,100}$`).MatchString(record.Feature) && record.Schema == 2 && regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(record.Source.Commit)
	if !valid {
		check("record", fmt.Errorf("invalid record identity"))
		return report
	}
	report.Feature = record.Feature
	report.Commit = record.Source.Commit
	current, err := currentRecordSource(root)
	report.ReleaseCommit = current.Commit
	if err == nil && !sameRecordSource(current, record.Source) {
		err = fmt.Errorf("execution source drift")
	}
	if err == nil {
		_, err = verifyRecordFile(root, record.CurrentState)
	}
	if err == nil {
		err = checkInputFiles(root, record.SelectedInputs)
	}
	if err == nil && (strings.TrimSpace(record.Impact.Reason) == "" || record.Impact.LiveRoutes == nil || record.Impact.PressureRequired == nil) {
		err = fmt.Errorf("explicit impact analysis required")
	}
	check("source", err)
	check("e2e", verifyStageRecord(root, record, "e2e", record.E2E, ""))
	if record.Impact.RaceRequired != nil && !*record.Impact.RaceRequired && record.Race.Status == "NOT_APPLICABLE" {
		check("race", validNA(record.Race))
	} else {
		check("race", verifyStageRecord(root, record, "race", record.Race, ""))
	}
	if record.Impact.PressureRequired != nil && !*record.Impact.PressureRequired && record.Pressure.Status == "NOT_APPLICABLE" {
		check("pressure", validNA(record.Pressure))
	} else {
		check("pressure", verifyStageRecord(root, record, "pressure", record.Pressure, ""))
	}
	if len(record.Impact.LiveRoutes) == 0 && record.Live.Status == "NOT_APPLICABLE" {
		check("live", validNA(record.Live))
	} else {
		err := error(nil)
		seen := map[string]bool{}
		if len(record.Impact.LiveRoutes) == 0 || record.Live.Status != "PASS" || len(record.Live.Routes) != len(record.Impact.LiveRoutes) {
			err = fmt.Errorf("required live routes missing")
		}
		for _, route := range record.Impact.LiveRoutes {
			if seen[route] || !regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{0,100}$`).MatchString(route) {
				err = fmt.Errorf("invalid live scope")
				break
			}
			seen[route] = true
			if e := verifyStageRecord(root, record, "live", record.Live.Routes[route], route); e != nil {
				err = e
			}
		}
		check("live", err)
	}
	report.Pass = !slices.ContainsFunc(report.Checks, func(c RecordCheck) bool { return !c.Pass })
	return report
}
func validNA(stage StageRecord) error {
	if strings.TrimSpace(stage.Reason) == "" || stage.Receipt.Path != "" || len(stage.Routes) > 0 {
		return fmt.Errorf("invalid not-applicable evidence")
	}
	return nil
}
func verifyStageRecord(root string, record Record, name string, stage StageRecord, route string) error {
	if stage.Status != "PASS" || len(stage.RequiredCases) == 0 {
		return fmt.Errorf("stage missing, incomplete or not passed")
	}
	if name == "race" || name == "pressure" || name == "live" {
		if stage.Selection == nil || stage.Selection.By != "human" || len(stage.Selection.Argv) == 0 {
			return fmt.Errorf("explicit human selection required")
		}
	}
	path, err := verifyRecordFile(root, stage.Receipt)
	if err != nil {
		return err
	}
	var receipt StageReceipt
	if err := recordJSON(path, &receipt); err != nil {
		return fmt.Errorf("invalid stage receipt")
	}
	expectedPrefix := ".evidence/" + receipt.Feature + "/" + name + "/"
	if route != "" {
		expectedPrefix += route + "/"
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,100}$`).MatchString(receipt.Feature) || !strings.HasPrefix(stage.Receipt.Path, expectedPrefix) {
		return fmt.Errorf("receipt stage identity mismatch")
	}
	if receipt.Schema != 1 || receipt.Stage != name || receipt.Status != "PASS" || receipt.ExitCode == nil || *receipt.ExitCode != 0 || receipt.Error != nil || name != "live" && !sameRecordSource(record.Source, receipt.Source) {
		return fmt.Errorf("stage receipt disagrees with selected source or outcome")
	}
	if name == "live" {
		if err := checkMeasuredInputFiles(root, receipt.Source.Commit, receipt.SelectedInputs); err != nil {
			return err
		}
	} else if err := checkInputFiles(root, receipt.SelectedInputs); err != nil {
		return err
	}
	started, a := time.Parse(time.RFC3339Nano, receipt.StartedAt)
	finished, b := time.Parse(time.RFC3339Nano, receipt.FinishedAt)
	if a != nil || b != nil || finished.Before(started) || receipt.ElapsedSeconds < 0 {
		return fmt.Errorf("stage timing invalid")
	}
	if route != "" && (receipt.Route == nil || *receipt.Route != route) {
		return fmt.Errorf("live route mismatch")
	}
	if stage.Selection != nil && !slices.Equal(stage.Selection.Argv, receipt.Argv) {
		return fmt.Errorf("selected command mismatch")
	}
	attempt := filepath.Dir(path)
	for _, required := range []string{"command.json", "stdout.jsonl", "stderr.log", "exit-code"} {
		if receipt.Files[required] == "" {
			return fmt.Errorf("command capture missing")
		}
	}
	for member, hash := range receipt.Files {
		if _, err := verifyRecordFile(attempt, FileRef{member, hash}); err != nil {
			return err
		}
	}
	flags, flagsPresent := receipt.Toolchain["goFlags"]
	if receipt.Toolchain["goEnv"] != "off" || !flagsPresent || flags != "" {
		return fmt.Errorf("Go environment isolation policy missing or unsafe")
	}
	if len(receipt.Toolchain["executableSHA256"]) != 64 || receipt.Toolchain["collectorPython"] == "" {
		return fmt.Errorf("toolchain identity missing")
	}
	if receipt.CommandSHA256 != receipt.Files["command.json"] {
		return fmt.Errorf("command identity mismatch")
	}
	var command struct {
		Toolchain   map[string]string `json:"toolchain,omitempty"`
		Schema      int               `json:"schema"`
		Argv        []string          `json:"argv"`
		Stage       string            `json:"stage"`
		Route       *string           `json:"route"`
		Timeout     float64           `json:"timeoutSeconds"`
		Environment []string          `json:"environmentNames"`
	}
	if recordJSON(filepath.Join(attempt, "command.json"), &command) != nil || command.Schema != 1 || command.Stage != name || !slices.Equal(command.Argv, receipt.Argv) || !reflect.DeepEqual(command.Toolchain, receipt.Toolchain) {
		return fmt.Errorf("recorded command mismatch")
	}
	code, err := os.ReadFile(filepath.Join(attempt, "exit-code"))
	if err != nil || strings.TrimSpace(string(code)) != "0" {
		return fmt.Errorf("native process did not exit zero")
	}
	for _, variable := range command.Environment {
		if name != "live" && strings.HasPrefix(variable, "BARNESS_") && variable != "BARNESS_AI_EVIDENCE_DIR" && variable != "BARNESS_RUN_COMMIT" && !(name == "pressure" && (variable == "BARNESS_AI_PRESSURE" || variable == "BARNESS_AI_PRESSURE_LOOPS")) {
			return fmt.Errorf("ambient execution switch in ordinary receipt")
		}
	}
	if name == "race" {
		raw, err := os.ReadFile(filepath.Join(attempt, "stderr.log"))
		if err != nil || strings.Contains(string(raw), "DATA RACE") {
			return fmt.Errorf("race diagnostic in capture")
		}
	}
	if name != "live" {
		if len(receipt.Argv) < 3 || filepath.Base(receipt.Argv[0]) != "go" || receipt.Argv[1] != "test" || !slices.Contains(receipt.Argv, "-json") || slices.Contains(receipt.Argv, "-race") != (name == "race") {
			return fmt.Errorf("unexpected test command")
		}
		if len(receipt.Tests) == 0 {
			return fmt.Errorf("no executed tests")
		}
		if err := verifyTestEvents(attempt, receipt.Tests); err != nil {
			return err
		}
	}
	var primary []NativeBundleRef

	seenBundles := map[string]bool{}
	for _, ref := range receipt.Bundles {
		if seenBundles[ref.Path] || !strings.HasPrefix(ref.Path, "bundles/") {
			return fmt.Errorf("ambiguous native bundle path")
		}
		seenBundles[ref.Path] = true
		dir, err := recordPath(attempt, ref.Path)
		if err != nil {
			return err
		}
		if _, err = verifyRecordFile(dir, FileRef{"manifest.json", ref.ManifestSHA256}); err != nil {
			return err
		}
		if _, err = verifyRecordFile(dir, FileRef{"audit.json", ref.AuditSHA256}); err != nil {
			return err
		}
		if name != "live" {
			audited := AuditBundle(dir, audit.Environment())
			if audited.Err != "" || len(audited.Findings) != 0 {
				return fmt.Errorf("native evidence audit failed")
			}
		}
		if name == "live" {
			primary = append(primary, ref)
		} else if ref.Package == OfflinePackage {
			primary = append(primary, ref)
		} else {
			_, err := verifySecondaryNative(dir, ref.Package, receipt.Source.Commit)
			if err != nil {
				return err
			}

		}
	}
	if len(primary) != 1 {
		return fmt.Errorf("ambiguous primary native bundle")
	}
	bundleRef := primary[0]
	bundle, err := recordPath(attempt, bundleRef.Path)
	if err != nil {
		return err
	}
	if name == "live" {
		return verifySelectedLive(root, bundle, route, stage, record.Source, receipt.Source)
	}
	native, err := LoadBundle(bundle)
	if err != nil || len(native.Problems) > 0 || native.Commit != receipt.Source.Commit {
		return fmt.Errorf("native offline bundle invalid")
	}

	for _, c := range native.Cases {
		if c.Status == Fail {
			return fmt.Errorf("native failed case")
		}
	}
	seen := map[string]bool{}
	for _, id := range stage.RequiredCases {
		if id == "" || seen[id] || !slices.ContainsFunc(native.Cases, func(c Case) bool { return c.ID == id && c.Status == Pass }) {
			return fmt.Errorf("required case not passed")
		}
		seen[id] = true
	}
	return nil
}
func verifyTestEvents(attempt string, want []string) error {
	path, err := recordPath(attempt, "stdout.jsonl")
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 1<<20)
	seen := map[string]bool{}
	packages := 0
	for scanner.Scan() {
		var event struct{ Action, Package, Test string }
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			continue
		}
		if event.Action == "fail" {
			return fmt.Errorf("failed test event")
		}
		if event.Action == "pass" {
			if event.Test != "" {
				seen[event.Test] = true
			} else {
				packages++
			}
		}
	}
	if scanner.Err() != nil || packages == 0 {
		return fmt.Errorf("incomplete test events")
	}
	for _, name := range want {
		if !seen[name] {
			return fmt.Errorf("unexecuted test claimed")
		}
	}
	return nil
}
