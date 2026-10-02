// Package e2e holds barness-ai's offline acceptance tests. They run as an
// external Go program against the public ai.Client, a local controlled Provider
// and a trusted-host double (spec Testing Decisions §1–§3); nothing here reaches
// inside the library.
package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var run *evidence.Run

// builtinCatalogHash is the built-in catalog's hash, computed once.
var builtinCatalogHash string

func TestMain(m *testing.M) {
	var err error
	run, err = evidence.NewRun("./ai/e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	run.SetVersion("openai-go", evidence.ModuleVersion("github.com/openai/openai-go/v3"))
	run.SetVersion("anthropic-sdk-go", evidence.ModuleVersion("github.com/anthropics/anthropic-sdk-go"))
	catalog := ai.BuiltinCatalog()
	run.SetVersion("model_catalog_version", catalog.Version)
	if builtinCatalogHash, err = catalog.Hash(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	run.SetVersion("model_catalog_hash", builtinCatalogHash)
	code := m.Run()
	if err := run.Finish(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	if !auditPassed(run) {
		code = 1
	}
	os.Exit(code)
}

// auditPassed runs the redaction audit over the finished bundle and reports
// every finding (spec Testing Decisions §6).
func auditPassed(run *evidence.Run) bool {
	findings, err := run.Audit()
	if err != nil {
		fmt.Fprintln(os.Stderr, "redaction audit:", err)
		return false
	}
	for _, f := range findings {
		fmt.Fprintf(os.Stderr, "redaction audit: %s [%s] %s\n", f.File, f.Rule, f.Detail)
	}
	return len(findings) == 0
}

// validPolicy is a finite policy comfortably above what the offline fixtures
// need. Its numbers are test values, not deployment guidance.
func validPolicy() *ai.ResourcePolicy {
	return &ai.ResourcePolicy{
		MaxRequestBytes:        1 << 20,
		MaxImageBytes:          1 << 19,
		MaxFrameBytes:          1 << 16,
		MaxToolJSONBytes:       1 << 16,
		MaxErrorBodyBytes:      1 << 14,
		MaxOutputBytes:         1 << 20,
		MaxQueuedEvents:        1024,
		MaxQueuedEventBytes:    1 << 20,
		MaxConcurrentPerTenant: 8,
		MaxConcurrentProcess:   32,
		MaxAdmissionWaiters:    64,
		AdmissionWait:          time.Second,
		CallTimeout:            30 * time.Second,
		ConnectTimeout:         5 * time.Second,
		ResponseHeaderTimeout:  10 * time.Second,
		ReadIdleTimeout:        10 * time.Second,
	}
}
