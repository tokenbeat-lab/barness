package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Copies are explicitly synthetic. Originals and authoritative grant ledger
// stay intact; resealing makes these semantic checks independent of hashing.
func kimiQualificationTamper(t *testing.T, original, combo, consumer string) {
	t.Helper()
	for _, mode := range []string{"cost-state", "attempt-finished-cost", "native-source", "native-request-tail", "usd-reserve", "grant-binary", "raw-write-missing", "tool-pair", "wire-full-effort", "wire-simple-low", "wire-simple-high", "wire-default-effort", "wire-effort-removed", "wire-tool-choice", "wire-budget", "wire-tools", "wire-unsupported", "wire-options-effort"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+combo+"-K16-refuse-"+mode)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			caseID := supportmatrix.RegionalCaseID(combo, "kimi-k3/full-stream")
			var paths []string
			switch mode {
			case "wire-simple-low":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/simple-stream")
				paths = []string{caseID + "/exchanges.json"}
			case "wire-simple-high":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/simple-complete")
				paths = []string{caseID + "/exchanges.json"}
			case "wire-default-effort":
				paths = []string{caseID + "/exchanges.json"}
			case "wire-tool-choice", "wire-tools":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/tool-call")
				paths = []string{caseID + "/exchanges.json"}
			case "wire-full-effort", "wire-effort-removed", "wire-budget", "wire-unsupported", "wire-options-effort":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/full-complete")
				paths = []string{caseID + "/exchanges.json"}
				if mode == "wire-options-effort" {
					paths = append(paths, caseID+"/options.json")
				}
			case "native-request-tail":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/native-history")
				paths = []string{caseID + "/request.json"}
			case "attempt-finished-cost":
				paths = []string{caseID + "/observations.json"}
			case "cost-state":
				paths = []string{caseID + "/result.json", caseID + "/observations.json"}
			case "native-source", "usd-reserve":
				paths = []string{supportmatrix.RegionalReportFile}
			case "grant-binary":
				paths = []string{"regional-grant.json"}
			case "raw-write-missing":
				paths = []string{caseID + "/exchanges.json"}
			case "tool-pair":
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/tool-result")
				paths = []string{caseID + "/exchanges.json"}
			}
			for _, name := range paths {
				raw, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				var value any
				mustUnmarshal(t, raw, &value)
				switch mode {
				case "wire-full-effort", "wire-simple-low", "wire-simple-high", "wire-default-effort", "wire-effort-removed", "wire-tool-choice", "wire-budget", "wire-tools", "wire-unsupported", "wire-options-effort":
					kimiQualificationWireTamper(t, value, combo, mode, filepath.Base(name))
				case "native-request-tail":
					r := value.(map[string]any)
					messages := r["Messages"].([]any)
					r["Messages"] = append(messages, messages[len(messages)-1])
				case "attempt-finished-cost":
					v := value.([]any)[2].(map[string]any)["attempt"]
					kimiMutateCost(v)
				case "cost-state":
					kimiMutateCost(value)
				case "native-source":
					value.(map[string]any)["regional"].(map[string]any)["nativeSourceCases"].(map[string]any)["kimi-k3"] = supportmatrix.RegionalCaseID(combo, "kimi-k3/tool-call")
				case "usd-reserve":
					value.(map[string]any)["regional"].(map[string]any)["reservedUSD"] = 0
				case "grant-binary":
					value.(map[string]any)["binarySha256"] = "forged"
				case "raw-write-missing":
					x := value.([]any)[0].(map[string]any)
					x["responseBody"] = regionalMutateRaw(t, x["responseBody"].(string), "raw-usage-missing")
				case "tool-pair":
					x := value.([]any)[0].(map[string]any)
					var wire any
					mustUnmarshal(t, []byte(x["requestBody"].(string)), &wire)
					regionalChangeToolPair(wire)
					body, _ := json.Marshal(wire)
					x["requestBody"] = string(body)
				}
				updated := mustMarshal(t, value)
				if err := os.WriteFile(filepath.Join(dir, name), updated, 0600); err != nil {
					t.Fatal(err)
				}
				if filepath.Dir(name) != "." {
					kimiQualificationManifest(t, dir, caseID, filepath.Base(name), updated)
				}
			}
			if err := supportmatrix.SealRegionalBundle(dir); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctxFor(t), consumer, "-controlled", dir)
			cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
			raw, err := cmd.CombinedOutput()
			ev.Record("consumer", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "originalRetained": true, "vendorCalls": 0})
			ev.Check("readonly semantic consumer rejects forged native facts", err != nil, "%s", raw)
		})
	}
}

func kimiMutateCost(v any) {
	switch v := v.(type) {
	case []any:
		for _, x := range v {
			kimiMutateCost(x)
		}
	case map[string]any:
		if v["kind"] == "attempt_started" {
			return
		}
		for key, x := range v {
			if key == "costEstimate" {
				v[key] = map[string]any{"status": "resolved", "minimum": map[string]any{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0, "total": 0}, "maximum": map[string]any{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0, "total": 0}}
			} else {
				kimiMutateCost(x)
			}
		}
	}
}
func kimiQualificationManifest(t *testing.T, dir, caseID, name string, raw []byte) {
	t.Helper()
	path := filepath.Join(dir, "manifest.json")
	mraw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	mustUnmarshal(t, mraw, &m)
	for _, v := range m["cases"].([]any) {
		c := v.(map[string]any)
		if c["id"] == caseID {
			c["artifacts"].(map[string]any)[name] = evidence.Digest(raw)
		}
	}
	if err := os.WriteFile(path, mustMarshal(t, m), 0600); err != nil {
		t.Fatal(err)
	}
}
