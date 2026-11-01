package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog"
)

var candidateSnapshot = flag.String("catalog-candidate", "ai/catalog_builtin.json", "frozen repository-relative catalog snapshot for controlled contracts")

type catalogContractIdentity struct {
	Operation ai.Operation  `json:"operation"`
	Provider  ai.ProviderID `json:"provider"`
	API       ai.API        `json:"api"`
	ID        string        `json:"id"`
}
type catalogContractCase struct {
	Identity catalogContractIdentity `json:"identity"`
	Scene    string                  `json:"scene"`
	Case     string                  `json:"case"`
	Test     string                  `json:"test"`
}
type catalogContractPlan struct {
	Snapshot string                `json:"snapshot"`
	Version  string                `json:"version"`
	Hash     string                `json:"hash"`
	Cases    []catalogContractCase `json:"cases"`
}

func catalogIdentityKey(i catalogContractIdentity) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", i.Operation, i.Provider, i.API, i.ID)
}
func catalogIdentityName(i catalogContractIdentity) string {
	sum := sha256.Sum256([]byte(catalogIdentityKey(i)))
	return hex.EncodeToString(sum[:12])
}
func catalogSnapshotRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// The fixed entry always uses the final snapshot in Config.Catalog. Its
// explicit per-identity plan is consumed with both Go child events and native
// case artifacts; a parent PASS or a shortened identity list cannot qualify.
func TestCatalogCandidateContracts(t *testing.T) {
	root := catalogSnapshotRoot(t)
	path := *candidateSnapshot
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cat, report, err := catalog.Load(raw)
	ev := run.Case(t, "CATALOG-CANDIDATE-PLAN")
	ev.Record("load-report", report)
	ev.Check("complete frozen snapshot loads", err == nil && report.Failed == 0, "%v %+v", err, report)
	if err != nil || report.Failed != 0 {
		return
	}
	sum := sha256.Sum256(raw)
	plan := catalogContractPlan{Snapshot: "sha256:" + hex.EncodeToString(sum[:]), Version: cat.Version, Hash: report.EffectiveHash}
	add := func(identity catalogContractIdentity, scenes []string) {
		for _, scene := range scenes {
			name := catalogIdentityName(identity)
			plan.Cases = append(plan.Cases, catalogContractCase{identity, scene, "CC-" + name + "-" + scene, t.Name() + "/" + name + "/" + scene})
		}
	}
	for _, m := range cat.Models {
		scenes := []string{"load-query", "basic", "history-tools", "input-modalities", "parameters", "authorization"}
		if m.Reasoning {
			scenes = append(scenes, "reasoning")
		}
		for n := range m.Cost.Tiers {
			for _, boundary := range []string{"below", "equal", "above"} {
				scenes = append(scenes, fmt.Sprintf("tier-%d-%s", n, boundary))
			}
		}
		add(catalogContractIdentity{ai.OperationChat, m.Provider, m.API, m.ID}, scenes)
	}
	for _, m := range cat.ImageModels {
		add(catalogContractIdentity{ai.OperationImage, m.Provider, m.API, m.ID}, []string{"load-query", "generation", "options", "edit-boundaries", "authorization"})
	}
	for _, m := range cat.ClassifierModels {
		scenes := []string{"load-query", "questions", "formats", "image-locations", "limits", "authorization"}
		for n := range m.Cost.Tiers {
			for _, boundary := range []string{"below", "equal", "above"} {
				scenes = append(scenes, fmt.Sprintf("tier-%d-%s", n, boundary))
			}
		}
		add(catalogContractIdentity{ai.OperationClassifier, m.Provider, m.API, m.ID}, scenes)
	}
	ev.Record("plan", plan)
	ev.Check("nonempty complete identity plan", len(plan.Cases) > 0, "cases=%d", len(plan.Cases))
	groups := map[string][]catalogContractCase{}
	order := []string{}
	for _, item := range plan.Cases {
		key := catalogIdentityName(item.Identity)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], item)
	}
	for _, name := range order {
		t.Run(name, func(t *testing.T) {
			for _, item := range groups[name] {
				t.Run(item.Scene, func(t *testing.T) {
					scene := run.Case(t, item.Case)
					scene.Record("identity", item.Identity)
					switch item.Identity.Operation {
					case ai.OperationChat:
						catalogChatContract(t, scene, cat, item.Identity, item.Scene)
					case ai.OperationImage:
						catalogImageContract(t, scene, cat, item.Identity, item.Scene)
					case ai.OperationClassifier:
						catalogClassifierContract(t, scene, cat, item.Identity, item.Scene)
					}
				})
			}
		})
	}
}

func catalogContractQuery(ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity) {
	switch i.Operation {
	case ai.OperationChat:
		m, ok := cat.Lookup(i.Provider, i.API, i.ID)
		ev.Check("exact chat query", ok && m.ID == i.ID, "%+v", m)
		ev.Check("chat list discoverable", len(cat.ModelsOf(i.Provider, i.API)) > 0, "missing list")
	case ai.OperationImage:
		m, ok := cat.LookupImage(i.Provider, i.API, i.ID)
		ev.Check("exact image query", ok && m.ID == i.ID, "%+v", m)
		ev.Check("image list discoverable", len(cat.ImageModelsOf(i.Provider, i.API)) > 0, "missing list")
	case ai.OperationClassifier:
		m, ok := cat.LookupClassifier(i.Provider, i.API, i.ID)
		ev.Check("exact classifier query", ok && m.ID == i.ID, "%+v", m)
		ev.Check("classifier list discoverable", len(cat.ClassifierModelsOf(i.Provider, i.API)) > 0, "missing list")
	}
}
