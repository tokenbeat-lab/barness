package e2e

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Failure inventory is fixed in the feature spec before implementation. These
// public CLI/process tests never possess a vendor key or a public claim store.
func TestMiniMaxThinkingSupplementTemplate(t *testing.T) {
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true)
	ev := run.Case(t, "MMTS-template")
	g := mmThinkingTemplate(t, ev, cli, binary, strings.Repeat("a", 64))
	s := g.Scope
	ev.Check("fixed identity and single call", s.Route == "minimax-messages" && s.Provider == "minimax" && s.API == "anthropic-messages" && len(s.Models) == 1 && s.Models[0] == minimaxlive.M3 && len(s.Scenes) == 1 && s.Scenes[0] == "thinking-on" && s.Stage == "candidate", "%+v", s)
	ev.Check("exact finite budget", s.MaxCalls == 1 && s.MaxAttempts == 1 && s.MaxRetries == 0 && s.MaxRequestReruns == 0 && s.MaxOutputTokens == 2048 && s.MaxReservedInputTokens == 2048 && s.MaxRequestedOutputTokens == 2048 && s.CallSeconds == 120 && s.ProcessSeconds == 180, "%+v", s)
	ev.Check("template cannot grant authority", g.ID == "" && g.ExpiresAt.IsZero() && g.Check(s, time.Now()) != nil, "unexpected executable template")
	for name, args := range map[string][]string{
		"source":      {"-source", "minimax-messages", "-profile", "minimax-messages-thinking-on/v1"},
		"check":       {"-check", "missing", "-profile", "minimax-messages-thinking-on/v1"},
		"formal":      {"-template", "minimax-messages", "-profile", "minimax-messages-thinking-on/v1", "-stage", "formal", "-binary", binary, "-baseline-manifest-sha256", strings.Repeat("a", 64)},
		"other":       {"-template", "minimax-chat", "-profile", "minimax-messages-thinking-on/v1", "-stage", "candidate", "-binary", binary, "-baseline-manifest-sha256", strings.Repeat("a", 64)},
		"missing-pin": {"-template", "minimax-messages", "-profile", "minimax-messages-thinking-on/v1", "-stage", "candidate", "-binary", binary},
		"unknown":     {"-template", "minimax-messages", "-profile", "arbitrary", "-stage", "candidate", "-binary", binary, "-baseline-manifest-sha256", strings.Repeat("a", 64)},
	} {
		c := exec.CommandContext(ctxFor(t), cli, args...)
		c.Dir, c.Env = "../..", regionalChildEnvironment()
		out, e := c.CombinedOutput()
		ev.Check("reject "+name, e != nil, "%s", regionalSafeOutput(out))
	}
}

func mmThinkingTemplate(t *testing.T, ev *evidence.Case, cli, binary, base string) livegrant.Grant {
	t.Helper()
	cmd := exec.CommandContext(ctxFor(t), cli, "-template", "minimax-messages", "-mode", "controlled", "-stage", "candidate", "-profile", "minimax-messages-thinking-on/v1", "-baseline-manifest-sha256", base, "-binary", binary)
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	out, e := cmd.CombinedOutput()
	ev.Record("template-process", map[string]any{"argv": cmd.Args, "error": errString(e), "output": regionalSafeOutput(out), "vendorCalls": 0})
	if e != nil {
		t.Fatalf("fixed MiniMax single-item template missing: %v %s", e, out)
	}
	g, e := livegrant.Decode(out)
	if e != nil {
		t.Fatal(e)
	}
	return g
}

func TestMiniMaxThinkingSupplementAdmission(t *testing.T) {
	kimiHistoryPrivateStore(t)
	ev := run.Case(t, "MMTS-admission")
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := kimiRetainedBinary(t, miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true), "minimax-single")
	g := mmThinkingTemplate(t, ev, cli, binary, strings.Repeat("a", 64))
	// Every variant runs a real isolated process; admission failures precede claim.
	for _, kind := range []string{"entry", "count", "timeout", "baseline", "source", "binary", "sdk", "model", "scenes", "usd", "calls", "attempts", "retry", "expiry", "full-with-single", "single-with-full"} {
		t.Run(kind, func(t *testing.T) {
			c := run.Case(t, "MMTS-admission-"+kind)
			var bad map[string]any
			mustUnmarshal(t, mustMarshal(t, g), &bad)
			s := bad["scope"].(map[string]any)
			switch kind {
			case "baseline":
				s["baselineManifestSha256"] = "bad"
			case "source":
				s["sourceSha256"] = strings.Repeat("b", 64)
			case "binary":
				s["binarySha256"] = strings.Repeat("b", 64)
			case "sdk":
				s["sdk"] = "other"
			case "model":
				s["models"] = []string{"other"}
			case "scenes":
				s["scenes"] = []string{"thinking-on", "tool-call"}
			case "usd":
				s["maxUSD"] = 0.02
			case "calls":
				s["maxCalls"] = 2
			case "attempts":
				s["maxAttempts"] = 2
			case "retry":
				s["maxRequestReruns"] = 1
			}
			var grant livegrant.Grant
			mustUnmarshal(t, mustMarshal(t, bad), &grant)
			if kind == "single-with-full" {
				id, _ := minimaxlive.Select("messages", minimaxlive.M3, "standard")
				h, _ := livegrant.FileHash(binary)
				grant.Scope, _ = minimaxlive.ProcessScope(id, "controlled", "candidate", h)
			}
			grant.ID = "synthetic-mmts-" + kind + "-" + filepath.Base(run.Dir())
			grant.AccountAlias, grant.RegionAlias = "synthetic-account", "synthetic-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			if kind == "expiry" {
				grant.ExpiresAt = time.Now().UTC().Add(-time.Second)
			}
			dir, _, cmd := mmThinkingProcess(t, c, binary, grant, "success-messages", kind)
			r, e := minimaxlive.Load(dir)
			if e != nil {
				t.Fatal(e)
			}
			c.Check("admission rejects before reservation or HTTP", cmd.ProcessState.ExitCode() == 1 && r.Budget == (minimaxlive.Budget{}) && len(r.Calls) == 0 && r.ProcessGrantSHA256 == "", "exit=%d failure=%s", cmd.ProcessState.ExitCode(), r.Failure)
		})
	}
}
