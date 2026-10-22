// Command releasegate reads a feature record and verifies existing receipts.
// It never starts ordinary tests, race, pressure, vet, Node or vendor calls.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"os"
	"path/filepath"
)

func main() { os.Exit(cli(os.Args[1:])) }
func cli(args []string) int {
	flags := flag.NewFlagSet("releasegate", flag.ContinueOnError)
	record := flags.String("record", "", "immutable feature receipt index")
	out := flags.String("out", "", "new report directory")
	if flags.Parse(args) != nil || flags.NArg() != 0 || *record == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "releasegate: -record and -out required")
		return 2
	}
	if err := os.Mkdir(*out, 0700); err != nil {
		fmt.Fprintln(os.Stderr, "releasegate: new output directory required")
		return 1
	}
	root, err := os.Getwd()
	if err != nil {
		return 1
	}
	report := release.EvaluateRecord(root, *record)
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return 1
	}
	if err := os.WriteFile(filepath.Join(*out, "release-report.json"), append(raw, '\n'), 0600); err != nil {
		return 1
	}
	text := fmt.Sprintf("# Feature release record\n\nFeature: %s\n\nrelease.pass: %t\n", report.Feature, report.Pass)
	for _, c := range report.Checks {
		text += fmt.Sprintf("\n- %s: %t", c.ID, c.Pass)
		if c.Problem != "" {
			text += " (" + c.Problem + ")"
		}
		text += "\n"
	}
	if err := os.WriteFile(filepath.Join(*out, "release-report.md"), []byte(text), 0600); err != nil {
		return 1
	}
	if report.Pass {
		return 0
	}
	return 1
}
