// Package e2e holds barness-ai's offline acceptance tests. They run as an
// external Go program against the public ai.Client, a local controlled Provider
// and a trusted-host double (spec Testing Decisions §1–§3); nothing here reaches
// inside the library.
package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var run *evidence.Run

func TestMain(m *testing.M) {
	var err error
	run, err = evidence.NewRun("./ai/e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	run.SetVersion("openai-go", evidence.ModuleVersion("github.com/openai/openai-go/v3"))
	catalog := ai.BuiltinCatalog()
	run.SetVersion("model_catalog_version", catalog.Version)
	run.SetVersion("model_catalog_hash", hashJSON(catalog))
	code := m.Run()
	if err := run.Finish(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
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

func hashJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
