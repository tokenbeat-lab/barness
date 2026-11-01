// cataloggen deterministically verifies or generates a selected maintenance
// revision. Generation never claims candidate compatibility or vendor support.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func main() {
	if err := generate(); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "error", "category": "generation_failed", "message": err.Error()})
		os.Exit(1)
	}
}

func generate() error {
	out := flag.String("out", "", "artifact path; omitted writes bytes to stdout")
	statePath := flag.String("state", "", "explicit exported maintenance State JSON")
	store := flag.String("store", "", "explicit selected approved local maintenance store")
	builtin := flag.Bool("builtin", false, "explicitly select reviewed checked-in Provider source")
	check := flag.Bool("check", false, "read-only comparison with --out")
	confirm := flag.Bool("confirm", false, "explicitly confirm generated artifact replacement")
	dryRun := flag.Bool("dry-run", false, "preview without writing output")
	_ = flag.Bool("json", false, "emit JSON preview/errors (the default)")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	selected := 0
	if *statePath != "" {
		selected++
	}
	if *store != "" {
		selected++
	}
	if *builtin {
		selected++
	}
	if selected > 1 {
		return fmt.Errorf("select exactly one of --state, --store or --builtin")
	}
	if *out != "" && !*check && !*dryRun && (selected != 1 || !*confirm) {
		return fmt.Errorf("artifact writes require explicit source selection and --confirm")
	}
	var state maintenance.State
	var err error
	switch {
	case *statePath != "":
		var raw []byte
		raw, err = os.ReadFile(*statePath)
		if err == nil {
			state, err = maintenance.DecodeState(raw)
		}
	case *store != "":
		var revision maintenance.Revision
		revision, err = maintenance.ReadCurrent(*store)
		state = revision.State
	default:
		state, err = maintenance.BuiltinState()
	}
	if err != nil {
		return err
	}
	raw, err := maintenance.Generate(state)
	if err != nil {
		return err
	}
	if *check {
		if *out == "" {
			return fmt.Errorf("--check requires --out")
		}
		actual, err := os.ReadFile(*out)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, raw) {
			return fmt.Errorf("artifact differs from selected reviewed maintenance source")
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "current", "path": *out})
	}
	if *dryRun {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "preview", "path": *out, "bytes": len(raw)})
	}
	if *out == "" {
		_, err = os.Stdout.Write(raw)
		return err
	}
	return maintenance.WriteGenerated(state, *out)
}
