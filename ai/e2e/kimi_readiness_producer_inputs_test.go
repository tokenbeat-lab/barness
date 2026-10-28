package e2e

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Real run.py producers keep Markdown outside source.files, in selectedInputs.
// A complete union is valid, but omission or disagreement remains a rejection.
func TestKimiReadinessProducerInputs(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary, digest := kimiQualificationBinary(t)
	nine := kimiNineFixtureBinary(t)
	consumer := kimiQualificationConsumer(t)
	ev := run.Case(t, "KIMIREADY-producer-selected-inputs")
	baseline := kimiNineFixture(t, ev, nine, "kimi-cn-chat")
	if baseline == "" {
		return
	}
	raw, _ := os.ReadFile(filepath.Join(baseline, "manifest.json"))
	supplement, srv, _ := kimiHistoryCapture(t, ev, binary, digest, "kimi-cn-chat", evidence.Digest(raw), false)
	defer srv.Close()
	if supplement == "" {
		return
	}
	dir := filepath.Join(filepath.Dir(run.Dir()), "kimi-input-layout", filepath.Base(run.Dir()))
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	receipt := filepath.Join(filepath.Dir(baseline), "producer-receipt.json")
	original, e := os.ReadFile(receipt)
	if e != nil {
		t.Fatal(e)
	}
	defer os.WriteFile(receipt, original, 0600)
	var p map[string]any
	mustUnmarshal(t, original, &p)
	files := p["source"].(map[string]any)["files"].(map[string]any)
	selected := map[string]any{}
	for n, h := range files {
		if strings.HasSuffix(n, ".md") {
			selected[n] = h
			delete(files, n)
		}
	}
	if len(selected) < 2 {
		t.Fatal("fixture lacks original Markdown inputs")
	}
	p["selectedInputs"] = selected
	runCheck := func(wantPass bool, label string) {
		if err := os.WriteFile(receipt, mustMarshal(t, p), 0600); err != nil {
			t.Fatal(err)
		}
		c := map[string]any{"schema": 1, "kind": "kimi-history-composition", "combo": "kimi-cn-chat", "model": "kimi-k3", "baseline": kimiCompositionRef(t, baseline), "supplement": kimiCompositionRef(t, supplement)}
		if err := os.WriteFile(filepath.Join(dir, "kimi-composition.json"), mustMarshal(t, c), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.CommandContext(ctxFor(t), consumer, "-inspect-bundle", dir, "-mode", "controlled", "-stage", "candidate")
		cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
		out, e := cmd.CombinedOutput()
		ev.Record(label, map[string]any{"output": regionalSafeOutput(out), "error": errString(e), "vendorCalls": 0})
		ev.Check(label, (e == nil) == wantPass, "%v %s", e, regionalSafeOutput(out))
	}
	runCheck(true, "accept exact original source-plus-selected-input layout")
	name := ""
	for n := range selected {
		name = n
		break
	}
	h := selected[name]
	delete(selected, name)
	runCheck(false, "reject missing original input")
	selected[name] = h
	selected[name] = strings.Repeat("b", 64)
	runCheck(false, "reject changed original input")
	selected[name] = h
	files[name] = strings.Repeat("c", 64)
	runCheck(false, "reject conflicting producer provenance")
	delete(files, name)
	if err := os.WriteFile(receipt, original, 0600); err != nil {
		t.Fatal(err)
	}
	// Leave a valid descriptor and both original receipts after controlled tamper cases.
	c := map[string]any{"schema": 1, "kind": "kimi-history-composition", "combo": "kimi-cn-chat", "model": "kimi-k3", "baseline": kimiCompositionRef(t, baseline), "supplement": kimiCompositionRef(t, supplement)}
	if err := os.WriteFile(filepath.Join(dir, "kimi-composition.json"), mustMarshal(t, c), 0600); err != nil {
		t.Fatal(err)
	}
}
