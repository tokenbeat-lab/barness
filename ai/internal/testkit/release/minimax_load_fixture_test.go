package release

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"sync"
	"testing"
	"time"
)

// Existing isolated gate tests need real positive MiniMax load captures;
// invented PASS counters would defeat the new public artifact consumer.
var miniMaxFixture struct {
	once   sync.Once
	base   string
	bundle Bundle
	err    error
}

func TestMain(m *testing.M) {
	code := m.Run()
	if miniMaxFixture.base != "" {
		_ = os.RemoveAll(miniMaxFixture.base)
	}
	os.Exit(code)
}

func actualMiniMaxLoadFixture(t *testing.T) Bundle {
	t.Helper()
	miniMaxFixture.once.Do(func() {
		miniMaxFixture.base, miniMaxFixture.err = os.MkdirTemp("", "barness-release-load-fixture-")
		if miniMaxFixture.err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, "go", "test", "-count=1", "-timeout=90s", "./ai/e2e", "-run", "^TestMiniMaxSharedDesignLoad$")
		command.Dir = "../../../.."
		for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
			if v, ok := os.LookupEnv(key); ok {
				command.Env = append(command.Env, key+"="+v)
			}
		}
		command.Env = append(command.Env, "BARNESS_AI_PRESSURE=1", "BARNESS_AI_EVIDENCE_DIR="+miniMaxFixture.base)
		if output, err := command.CombinedOutput(); err != nil {
			miniMaxFixture.err = fmt.Errorf("public MiniMax load fixture failed: %w (%s)", err, output)
			return
		}
		dir, err := FindBundle(miniMaxFixture.base)
		if err != nil {
			miniMaxFixture.err = err
			return
		}
		miniMaxFixture.bundle, miniMaxFixture.err = LoadBundle(dir)
		if miniMaxFixture.err != nil {
			return
		}
		for _, e := range MiniMaxDesignLoadEvidence(miniMaxFixture.bundle) {
			if e.Status != Pass {
				miniMaxFixture.err = fmt.Errorf("public MiniMax load fixture invalid: %s", e.Detail)
				return
			}
		}
	})
	if miniMaxFixture.err != nil {
		t.Fatal(miniMaxFixture.err)
	}
	b := miniMaxFixture.bundle
	b.Cases = slices.Clone(b.Cases)
	b.caseFiles = maps.Clone(b.caseFiles)
	return b
}
