// auditbundle exposes the existing evidence audit to release-verification scripts.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: auditbundle <evidence-directory>")
		os.Exit(2)
	}
	dir := os.Args[1]
	findings, err := audit.Bundle(dir, audit.Environment())
	if err != nil {
		// Filesystem errors can contain private paths; report a stable reason.
		fmt.Fprintln(os.Stderr, "FAIL: evidence audit could not read the bundle")
		os.Exit(1)
	}
	if findings == nil {
		findings = []audit.Finding{}
	}
	data, err := json.MarshalIndent(map[string]any{"findings": findings}, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(dir, "audit.json"), append(data, '\n'), 0o644)
	}
	if err != nil || len(findings) != 0 {
		fmt.Fprintln(os.Stderr, "FAIL: evidence redaction audit")
		os.Exit(1)
	}
	fmt.Println("PASS: evidence redaction audit")
}
