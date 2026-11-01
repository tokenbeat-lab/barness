// catalogctl maintains reviewed local catalogs. It never invokes a model API.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func main() {
	result, err := execute(os.Args[1:])
	if err != nil {
		category := "execution_failed"
		var known *maintenance.Error
		if errors.As(err, &known) {
			category = known.Category
		}
		response := map[string]any{"status": "error", "category": category, "message": err.Error()}
		if result != nil {
			response["result"] = result
		}
		if known != nil && len(known.Diagnostics) > 0 {
			response["diagnostics"] = known.Diagnostics
		}
		_ = json.NewEncoder(os.Stdout).Encode(response)
		os.Exit(1)
	}
	if result != nil {
		if raw, ok := result.(rawOutput); ok {
			if _, err := os.Stdout.Write(raw); err != nil {
				fmt.Fprintln(os.Stderr, "write output:", err)
				os.Exit(1)
			}
			return
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, "write result:", err)
			os.Exit(1)
		}
	}
}

type rawOutput []byte

func execute(args []string) (any, error) {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return map[string]any{"status": "help", "commands": []string{"versions", "history", "candidate", "export", "add", "revise", "withdraw", "restore"}}, nil
	}
	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	store := flags.String("store", ".catalog", "local maintenance store")
	_ = flags.Bool("json", false, "emit machine-readable JSON (the default)")
	dryRun := flags.Bool("dry-run", false, "preview without creating files")
	confirm := flags.Bool("confirm", false, "explicitly confirm the previewed change")
	op := flags.String("operation", "", "chat, image or classifier")
	provider := flags.String("provider", "", "barness Provider identity")
	api := flags.String("api", "", "explicit API protocol")
	model := flags.String("model", "", "serving model ID")
	reason := flags.String("reason", "", "reason for the reviewed change")
	id := flags.String("id", "", "immutable candidate identity")
	sourcePath := flags.String("source", "", "retained models.dev input file (otherwise fetch source URL)")
	sourceURL := flags.String("source-url", maintenance.ModelsDevURL, "catalog source URL")
	sourceTime := flags.String("source-time", "", "original RFC3339 fetch time for a retained input file")
	sourceProvider := flags.String("source-provider", "", "models.dev serving Provider key")
	sourceModel := flags.String("source-model", "", "source offering key (defaults to --model)")
	definitionPath := flags.String("definition", "", "explicit manual JSON definition/overlay")
	decisionsPath := flags.String("decisions", "", "JSON pointer field decisions with ownership/reasons")
	exportState := flags.Bool("state", false, "export selected approved maintenance state instead of runtime bytes")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return map[string]any{"status": "help", "command": command}, nil
		}
		return nil, &maintenance.Error{Category: "invalid_argument", Message: err.Error()}
	}
	if flags.NArg() != 0 {
		return nil, &maintenance.Error{Category: "invalid_argument", Message: "unexpected positional arguments"}
	}
	switch command {
	case "versions":
		current, err := maintenance.ReadCurrent(*store)
		if err != nil {
			return nil, err
		}
		versions := map[ai.ProviderID]uint64{}
		for _, p := range current.State.Providers {
			versions[p.ID] = p.Version
		}
		return map[string]any{"status": "current", "revision": current.ID, "versions": versions}, nil
	case "export":
		raw, err := maintenance.Export(*store, *exportState)
		if err != nil {
			return nil, err
		}
		return rawOutput(raw), nil
	case "history":
		history, err := maintenance.History(*store)
		return map[string]any{"status": "history", "history": history}, err
	case "candidate":
		candidate, err := maintenance.ReadCandidate(*store, *id)
		return map[string]any{"status": "awaiting_check", "candidate": candidate, "changes": candidate.Changes}, err
	case "add", "revise", "withdraw", "restore":
		request := maintenance.ChangeRequest{Action: command, Identity: maintenance.Identity{Operation: ai.Operation(*op), Provider: ai.ProviderID(*provider), API: ai.API(*api), ID: *model}, Reason: *reason, SourceProvider: *sourceProvider, SourceModel: *sourceModel}
		if command == "add" || (command == "revise" && (*sourcePath != "" || *sourceProvider != "")) {
			source, err := maintenance.AcquireSource(context.Background(), *sourcePath, *sourceURL, *sourceTime)
			if err != nil {
				return nil, err
			}
			request.Source = &source
		}
		if *definitionPath != "" {
			raw, err := os.ReadFile(*definitionPath)
			if err != nil {
				return nil, &maintenance.Error{Category: "invalid_argument", Message: "manual definition could not be read"}
			}
			request.Definition = raw
		}
		if *decisionsPath != "" {
			raw, err := os.ReadFile(*decisionsPath)
			if err != nil {
				return nil, &maintenance.Error{Category: "invalid_argument", Message: "field decisions could not be read"}
			}
			if err := json.Unmarshal(raw, &request.Decisions); err != nil {
				return nil, &maintenance.Error{Category: "invalid_argument", Message: "field decisions must be a JSON object"}
			}
		}
		proposal, err := maintenance.Plan(*store, request)
		if err != nil {
			return nil, err
		}
		if *dryRun || proposal.Status == "unchanged" {
			return proposal, nil
		}
		if proposal.Candidate == nil {
			return proposal, &maintenance.Error{Category: "unresolved_definition", Message: "required source/manual fields remain unresolved", Diagnostics: proposal.Diagnostics}
		}
		if !*confirm {
			return nil, &maintenance.Error{Category: "confirmation_required", Message: "preview with --dry-run, then explicitly use --confirm"}
		}
		candidate, err := maintenance.Freeze(*store, proposal, *confirm)
		return map[string]any{"status": "awaiting_check", "candidate": candidate, "changes": candidate.Changes}, err
	default:
		return nil, &maintenance.Error{Category: "invalid_argument", Message: "unknown command"}
	}
}
