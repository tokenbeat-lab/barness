package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogCandidateDeclaredCustomSizes(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-declared-custom-size-contract")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	root := catalogSnapshotRoot(t)
	directory := t.TempDir()
	store := filepath.Join(directory, "store")
	definition := filepath.Join(directory, "capabilities.json")
	// These are controlled Host declarations, not new provider measurements.
	patch := []byte(`{"capabilities":{"customSizes":{"edgeMultiple":256,"maxAspectRatio":2,"maxEdge":2048,"minPixels":262144,"maxPixels":4194304,"source":"synthetic controlled capability fixture"},"maxOutputImages":2,"inputFidelity":["high"]}}`)
	if err := os.WriteFile(definition, patch, 0600); err != nil {
		t.Fatal(err)
	}
	ev.Fixture("capabilities.json", patch)
	model := "gpt-image-2.5-sunburst-2026-09-08"
	out, exit := invokeCatalogCLI(t, ev, binary, "revise", "--store", store, "--operation", "image", "--provider", "openai", "--api", "openai-images", "--model", model, "--definition", definition, "--reason", "controlled capability change", "--confirm")
	var frozen struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	mustUnmarshal(t, out, &frozen)
	if !ev.Check("capability change is a confirmed candidate", exit == 0, "%s", out) {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var check maintenance.CheckResult
	_ = json.Unmarshal(out, &check)
	if !ev.Check("all declared capabilities have controlled check", exit == 0 && check.Status == "applicable", "%s", out) {
		return
	}
	var receipt struct {
		Bundles []struct {
			Path string `json:"path"`
		} `json:"bundles"`
	}
	catalogReadFixtureJSON(t, filepath.Join(root, check.Receipt), &receipt)
	bundle := filepath.Join(root, filepath.Dir(check.Receipt), receipt.Bundles[0].Path)
	identity := catalogContractIdentity{Operation: ai.OperationImage, Provider: "openai", API: "openai-images", ID: model}
	caseID := "CC-" + catalogIdentityName(identity) + "-options"
	var requests []struct {
		Body map[string]json.RawMessage `json:"body"`
	}
	catalogReadFixtureJSON(t, filepath.Join(bundle, caseID, "requests.json"), &requests)
	custom, multiple := false, false
	for _, request := range requests {
		var size string
		var n int
		_ = json.Unmarshal(request.Body["size"], &size)
		_ = json.Unmarshal(request.Body["n"], &n)
		custom = custom || size == "512x512"
		multiple = multiple || n == 2
	}
	ev.Check("declared custom minimum actually exercised on public wire", custom, "requests=%d", len(requests))
	ev.Check("declared maximum output count actually exercised", multiple, "requests=%d", len(requests))
}
