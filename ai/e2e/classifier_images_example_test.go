package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The executable is the Host around the already agreed public Classify seams.
// Its evidence must prove both actual HTTP calls and pre-send rejections.
func TestClassifierImagesExample(t *testing.T) {
	ev := run.Case(t, "UC10-executable")
	out := filepath.Join(run.Dir(), "classifier-images-host")
	argv := []string{"run"}
	if os.Getenv("BARNESS_EXAMPLE_RACE") == "1" {
		argv = append(argv, "-race")
	}
	argv = append(argv, "./examples/classifierimages", "-out", out)
	cmd := exec.Command("go", argv...)
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if !ev.Check("default executable replays without credentials", err == nil, "example unavailable or replay failed") {
		t.Logf("example: %s", output)
		return
	}
	var manifest struct {
		Mode        string            `json:"mode"`
		RealService string            `json:"realService"`
		Cases       []string          `json:"cases"`
		Artifacts   map[string]string `json:"artifacts"`
	}
	read := func(name string, value any) {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, value); err != nil {
			t.Fatal(err)
		}
	}
	read("manifest.json", &manifest)
	ev.Record("manifest", manifest)
	ev.Check("controlled evidence does not claim live PASS", manifest.Mode == "controlled" && manifest.RealService == "NOT_RUN", "wrong claims")
	ev.Check("complete two-protocol matrix", len(manifest.Cases) == 48, "got %d", len(manifest.Cases))
	for name, hash := range manifest.Artifacts {
		raw, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		ev.Check("sealed "+name, hex.EncodeToString(sum[:]) == hash, "hash mismatch")
	}
	fixtureHashes := map[string]string{}
	for _, name := range []string{"processed.png", "processed.jpg"} {
		raw, err := os.ReadFile(filepath.Join(out, "fixtures", name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		fixtureHashes[name] = hex.EncodeToString(sum[:])
	}
	var audit struct {
		Findings []any `json:"findings"`
	}
	read("audit.json", &audit)
	ev.Check("redaction audit empty", len(audit.Findings) == 0, "findings")
	for _, id := range manifest.Cases {
		t.Run(id, func(t *testing.T) {
			c := run.Case(t, "UC10-"+id)
			var report struct {
				Route, Scenario, Entry, Status string
				Assertions                     []struct {
					Name string
					Pass bool
				}
				Result struct {
					Answers     map[string]json.RawMessage
					ImageInputs []struct{ MimeType, SHA256 string }
					Metadata    struct{ Attempts []json.RawMessage }
					Usage       struct{ Input int }
				}
				Error *struct{ Code, Phase string }
				View  map[string]struct {
					Kind           string
					Probability    *float64
					Confidence     json.RawMessage
					OriginalPixels *struct{ Left, Top, Right, Bottom float64 }
					LegendImages   []struct {
						Pointer, SHA256 string
						ImageIndex      int
					}
				}
			}
			prefix := filepath.Join("cases", id)
			read(filepath.Join(prefix, "report.json"), &report)
			c.Record("report", report)
			c.Check("example assertions pass", report.Status == "PASS" && len(report.Assertions) > 0, "status=%s", report.Status)
			for _, a := range report.Assertions {
				c.Check(a.Name, a.Pass, "example assertion failed")
			}
			var requests []json.RawMessage
			read(filepath.Join(prefix, "requests.json"), &requests)
			var observations []json.RawMessage
			read(filepath.Join(prefix, "observations.json"), &observations)
			c.Check("observer evidence exists", len(observations) >= 2, "missing observations")
			success := report.Scenario == "context" || report.Scenario == "reuse" || report.Scenario == "refusal" ||
				report.Route == "jev" && (report.Scenario == "state" || report.Scenario == "instructions" || report.Scenario == "criteria" || report.Scenario == "box" || report.Scenario == "hook")
			sent := success || report.Route == "jev" && (report.Scenario == "bad-box" || report.Scenario == "bad-legend")
			if sent {
				c.Check("public Classify crossed HTTP", len(requests) == 1 && len(report.Result.Metadata.Attempts) == 1, "missing or duplicate HTTP/Attempt")
			} else {
				c.Check("rejection sends no HTTP", len(requests) == 0 && len(report.Result.Metadata.Attempts) == 0, "sent rejected call")
			}
			if report.Scenario == "context" || report.Scenario == "reuse" {
				expected := []string{fixtureHashes["processed.png"], fixtureHashes["processed.jpg"]}
				if report.Scenario == "reuse" {
					expected = []string{fixtureHashes["processed.jpg"], fixtureHashes["processed.png"]}
					if report.Route == "decisions" {
						expected = append(expected, fixtureHashes["processed.jpg"])
					}
				}
				c.Check("protocol-specific final occurrence count", len(report.Result.ImageInputs) == len(expected), "wrong image count")
				for i, hash := range expected {
					c.Check("final ordered image hash", i < len(report.Result.ImageInputs) && report.Result.ImageInputs[i].SHA256 == hash, "wrong final image")
				}
			}
			if !success {
				c.Check("atomic typed failure", report.Error != nil && len(report.Result.Answers) == 0 && len(report.Result.ImageInputs) == 0, "partial failure")
			}
			if report.Scenario == "box" && report.Route == "jev" {
				a := report.View["location"]
				wantConfidence := "null"
				if report.Entry == "hooked" {
					wantConfidence = "0"
				}
				c.Check("null and numeric zero remain distinct in view", string(a.Confidence) == wantConfidence, "got %s", a.Confidence)
				p := a.OriginalPixels
				c.Check("clockwise orientation and crop inverted", p != nil && p.Left == 200 && p.Top == 200 && p.Right == 440 && p.Bottom == 360, "wrong pixels: %+v", p)
				e := report.View["exists"].Probability
				c.Check("zero existence remains independent of box", e != nil && *e == 0, "missing zero")
			}
			if report.Scenario == "refusal" {
				c.Check("refusal remains explicit", report.View["exists"].Kind == "refusal", "lost refusal")
			}
			if report.Scenario == "criteria" || report.Scenario == "hook" {
				if report.Route == "jev" {
					refs := report.View["score"].LegendImages
					c.Check("legend relative pointers resolved", len(refs) == 2 && refs[0].Pointer == "/0/example" && refs[1].Pointer == "/1/example", "wrong refs")
					for _, ref := range refs {
						c.Check("legend uses final identity", ref.ImageIndex >= 0 && ref.ImageIndex < len(report.Result.ImageInputs) && ref.SHA256 == report.Result.ImageInputs[ref.ImageIndex].SHA256, "stale identity")
					}
				}
			}
			if report.Route == "decisions" && (report.Scenario == "state" || report.Scenario == "instructions" || report.Scenario == "criteria" || report.Scenario == "box") {
				c.Check("unsupported scope is explicit capability rejection", report.Error != nil && report.Error.Code == "invalid_request" && report.Error.Phase == "capability", "wrong error")
			}
			if report.Scenario == "bad-box" || report.Scenario == "bad-legend" {
				c.Check("known usage retained", report.Result.Usage.Input == 100, "usage lost")
			}
		})
	}
}

func TestClassifierImagesExampleLiveOptIn(t *testing.T) {
	for _, args := range [][]string{{"-live"}, {"-config", "missing.json"}, {"-live", "-route", "jev", "-config", "missing.json", "-key-env", "BARNESS_EXAMPLE_EXPLICIT_TEST_KEY"}} {
		ev := run.Case(t, "UC10-live-opt-in-"+args[0]+"-"+string(rune('0'+len(args))))
		out := filepath.Join(t.TempDir(), "bundle")
		argv := append([]string{"run", "./examples/classifierimages", "-out", out}, args...)
		cmd := exec.Command("go", argv...)
		cmd.Dir = ".."
		_, err := cmd.CombinedOutput()
		ev.Check("incomplete live configuration refuses before call", err != nil, "unexpected live success")
		_, statErr := os.Stat(filepath.Join(out, "manifest.json"))
		ev.Check("no manufactured live evidence", os.IsNotExist(statErr), "created misleading evidence")
	}
}
