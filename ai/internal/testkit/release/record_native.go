package release

import (
	"fmt"
	"strings"
)

// Auxiliary packages have native inventories without the main offline package
// identity. Verify through the same bounded filesystem boundary.
func verifySecondaryNative(dir, pkg, commit string) ([]Case, error) {
	_, root, err := openArtifactBundle(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	var manifest struct {
		Package string `json:"package"`
		Commit  string `json:"git_commit"`
		Cases   []struct {
			ID, Dir   string
			Status    Status
			Artifacts map[string]string
		}
	}
	if readBundleManifest(root, &manifest) != nil || manifest.Package != pkg || manifest.Commit != commit || len(manifest.Cases) == 0 {
		return nil, fmt.Errorf("auxiliary native identity invalid")
	}
	seen := map[string]bool{}
	var cases []Case
	for _, c := range manifest.Cases {
		physical := strings.TrimSuffix(c.Dir, "/")
		if seen[c.ID] || c.ID == "" || !canonicalArtifactPath(physical) || (c.Status != Pass && c.Status != NotRun && c.Status != Unsupported) {
			return nil, fmt.Errorf("auxiliary native case invalid")
		}
		seen[c.ID] = true
		child, err := openInventoryCase(root, physical)
		if err != nil {
			return nil, err
		}
		err = verifyCaseArtifactInventory(child, c.Artifacts)
		child.Close()
		if err != nil {
			return nil, err
		}
		cases = append(cases, Case{ID: c.ID, Status: c.Status})
	}
	return cases, nil
}
