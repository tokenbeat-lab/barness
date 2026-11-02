package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// This exercises the actual public default, not a separately injected candidate.
func TestTerraBuiltinCatalog(t *testing.T) {
	cat := ai.BuiltinCatalog()
	identity := catalogContractIdentity{Operation: ai.OperationChat, Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ID: "gpt-6.1-terra"}
	ev := run.Case(t, "CATALOG-TERRA-builtin")
	m, ok := cat.Lookup(identity.Provider, identity.API, identity.ID)
	if !ev.Check("Terra is discoverable in the public builtin", ok, "identity=%+v", identity) {
		return
	}
	off, offOK := m.ThinkingLevelMap.Off.Get()
	ev.Check("approved reasoning and modalities", m.Reasoning && offOK && off == "none" && m.ThinkingLevelMap.Minimal.IsNull() && m.DefaultThinkingLevel == ai.ThinkingMedium && slices.Equal(m.Input, []ai.Modality{ai.ModalityText, ai.ModalityImage}), "model=%+v", m)
	ev.Check("approved standard and whole-request tier rates", m.Cost.CostRates == (ai.CostRates{Input: 1, Output: 5, CacheRead: .05, CacheWrite: 1.25}) && len(m.Cost.Tiers) == 1 && m.Cost.Tiers[0] == (ai.CostTier{InputTokensAbove: 272000, CostRates: ai.CostRates{Input: 2, Output: 7.5, CacheRead: .1, CacheWrite: 2.5}}), "cost=%+v", m.Cost)
	for _, scene := range []string{"load-query", "basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization", "tier-0-below", "tier-0-equal", "tier-0-above"} {
		t.Run(scene, func(t *testing.T) { catalogChatContract(t, run.Case(t, "CATALOG-TERRA-"+scene), cat, identity, scene) })
	}
}

func TestTerraAdoptedBuiltinSource(t *testing.T) {
	ev := run.Case(t, "CATALOG-TERRA-adopted-source")
	state, err := maintenance.BuiltinState()
	if !ev.Check("adopted maintenance state loads", err == nil, "%v", err) {
		return
	}
	generated, err := maintenance.Generate(state)
	if !ev.Check("maintenance source generates", err == nil, "%v", err) {
		return
	}
	builtin, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := os.ReadFile("../../catalog/data/baseline/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("source, builtin and adopted baseline use identical bytes", bytes.Equal(generated, builtin) && bytes.Equal(baseline, builtin), "generated=%d builtin=%d baseline=%d", len(generated), len(builtin), len(baseline))
	var snapshot struct {
		Catalog ai.Catalog `json:"catalog"`
	}
	mustUnmarshal(t, baseline, &snapshot)
	ev.Check("public builtin equals adopted snapshot", reflect.DeepEqual(ai.BuiltinCatalog(), snapshot.Catalog), "catalog differs")
}

// A clean seed must export its own retained provenance without the old store.
func TestTerraAdoptedSeedExport(t *testing.T) {
	ev := run.Case(t, "CATALOG-TERRA-seed-export")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "absent-store")
	snapshot, err := os.ReadFile("../../catalog/data/baseline/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	ev.Check("fresh absent store exports approved snapshot", exit == 0 && bytes.Equal(out, snapshot), "exit=%d", exit)
	review := filepath.Join(dir, "review")
	out, exit = invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--repo", catalogSnapshotRoot(t), "--review-dir", review, "--confirm")
	if !ev.Check("fresh seed exports retained review", exit == 0, "exit=%d output=%s", exit, out) {
		return
	}
	var manifest maintenance.ReviewExport
	if err := json.Unmarshal(out, &manifest); err != nil {
		t.Fatal(err)
	}
	ev.Fixture("reexport-review.json", out)

	baselineDir := filepath.Join(catalogSnapshotRoot(t), "catalog", "data", "baseline")
	originalIndex, err := os.ReadFile(filepath.Join(baselineDir, "review.json"))
	if err != nil {
		t.Fatal(err)
	}
	var original maintenance.ReviewExport
	if err := json.Unmarshal(originalIndex, &original); err != nil {
		t.Fatal(err)
	}
	ev.Fixture("adopted-review.json", originalIndex)
	inputs := make(map[maintenance.ReviewInput]int)
	for _, input := range manifest.Inputs {
		inputs[input]++
	}
	for _, input := range original.Inputs {
		inputs[input]--
		oldBytes, oldErr := os.ReadFile(filepath.Join(baselineDir, filepath.FromSlash(input.Path)))
		newBytes, newErr := os.ReadFile(filepath.Join(review, filepath.FromSlash(input.Path)))
		ev.Check("original input metadata and bytes retained", oldErr == nil && newErr == nil && bytes.Equal(oldBytes, newBytes), "input=%+v old=%v new=%v", input, oldErr, newErr)
	}
	sameInputs := len(manifest.Inputs) == len(original.Inputs)
	for _, count := range inputs {
		sameInputs = sameInputs && count == 0
	}
	ev.Check("all original inputs retained exactly once", sameInputs, "old=%d new=%d", len(original.Inputs), len(manifest.Inputs))
	originSum := sha256.Sum256(originalIndex)
	origin := "sha256:" + hex.EncodeToString(originSum[:])
	references := make(map[maintenance.ReviewReference]int)
	for _, reference := range manifest.References {
		references[reference]++
	}
	for _, reference := range original.References {
		if reference.OriginReview == "" {
			reference.OriginReview = origin
		}
		references[reference]--
	}
	sameReferences := len(manifest.References) == len(original.References)
	for _, count := range references {
		sameReferences = sameReferences && count == 0
	}
	ev.Check("all original references and origins retained", sameReferences, "old=%d new=%d", len(original.References), len(manifest.References))
	files := map[string]bool{"review.json": true}
	for name := range manifest.Files {
		files[name] = true
	}
	err = filepath.WalkDir(review, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			name, relErr := filepath.Rel(review, path)
			if relErr != nil {
				return relErr
			}
			name = filepath.ToSlash(name)
			ev.Check("reexport contains no unlisted file", files[name], "path=%s", name)
			delete(files, name)
		}
		return nil
	})
	ev.Check("reexport inventory is complete", err == nil && len(files) == 0, "err=%v missing=%v", err, files)
	for path, digest := range manifest.Files {
		raw, err := os.ReadFile(filepath.Join(review, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		ev.Check("reexport inventory matches retained bytes", "sha256:"+hex.EncodeToString(sum[:]) == digest, "path=%s", path)
	}
	again, err := os.ReadFile(filepath.Join(review, "snapshot.json"))
	ev.Check("reexport preserves approved runtime bytes", err == nil && bytes.Equal(again, snapshot), "err=%v", err)
	_, err = os.Stat(store)
	ev.Check("fresh export does not create a store", os.IsNotExist(err), "err=%v", err)
}
