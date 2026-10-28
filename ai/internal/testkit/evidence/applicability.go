package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// VerifyRouteSource compares product inputs, not collectors or other suites.
// Unknown runtime files are shared: new product code never defaults to harmless.
// The original complete map remains bound to its producer grant and manifest.
func VerifyRouteSource(measured, current map[string]string, provider, api string) error {
	if len(measured) == 0 || len(current) == 0 {
		return fmt.Errorf("execution source missing")
	}
	for _, files := range []map[string]string{measured, current} {
		for name := range files {
			if measured[name] != current[name] && routeInput(name, provider, api) && !reviewedSourceChange(name, measured[name], current[name], provider, api, measured) {
				return fmt.Errorf("route product input changed: %s", name)
			}
		}
	}
	return nil
}

func routeInput(name, provider, api string) bool {
	if name == "go.mod" || name == "go.sum" {
		return true
	}
	if strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "tools/") || strings.HasPrefix(name, "ai/internal/testkit/") || strings.HasPrefix(name, "ai/live/") || strings.HasPrefix(name, "ai/e2e/") {
		return false
	}
	if strings.HasPrefix(name, "ai/examples/requestid/") {
		return true
	}
	if strings.HasPrefix(name, "ai/internal/") {
		return true
	}
	if filepath.Dir(name) != "ai" {
		return strings.HasPrefix(name, "ai/") && !strings.HasPrefix(name, "ai/examples/") && !strings.HasPrefix(name, "ai/release/")
	}
	base := filepath.Base(name)
	// These files only construct catalog data, compared by exact used models.
	// Other catalog code (validation, guards, loaders) remains runtime input.
	switch base {
	case "chat.go":
		return api == "openai-completions"
	case "anthropic.go":
		return api == "anthropic-messages"
	case "responses.go":
		return api == "openai-responses"
	case "catalog.go", "catalog_builtin.go", "catalog_glm.go", "catalog_images.go", "catalog_decisions.go", "catalog_typesafe.go", "catalog_minimax.json", "catalog_regional.json", "catalog_kimi.json":
		return false
	case "catalog_xai.go":
		return provider == "xai"
	case "catalog_kimi_data.go":
		return strings.HasPrefix(provider, "moonshot")
	case "catalog_minimax.go":
		return provider == "minimax"
	case "catalog_regional.go":
		return provider == "zai" || provider == "minimax-cn"
	}
	for _, family := range []struct{ prefix, owner, protocol string }{
		{"responses_xai", "xai", "openai-responses"}, {"xai_", "xai", ""},
		{"responses_kimi", "kimi", "openai-responses"}, {"anthropic_kimi", "kimi", "anthropic-messages"}, {"chat_kimi", "kimi", "openai-completions"}, {"kimi_", "kimi", ""},
		{"responses_minimax", "minimax", "openai-responses"}, {"anthropic_minimax", "minimax", "anthropic-messages"}, {"chat_minimax", "minimax", "openai-completions"}, {"minimax_", "minimax", ""},
	} {
		if strings.HasPrefix(base, family.prefix) {
			owns := provider == family.owner || family.owner == "kimi" && strings.HasPrefix(provider, "moonshot") || family.owner == "minimax" && provider == "minimax-cn"
			return owns && (family.protocol == "" || api == family.protocol)
		}
	}
	for _, family := range []struct{ prefix, protocol string }{{"responses_", "openai-responses"}, {"chat_", "openai-completions"}, {"anthropic_", "anthropic-messages"}, {"gemini_", "google-generative-ai"}, {"typesafe_", "typesafe-system-one"}, {"openai_decisions", "openai-decisions"}, {"openai_images", "openai-images"}, {"google_images", "google-interactions"}} {
		if strings.HasPrefix(base, family.prefix) {
			return api == family.protocol
		}
	}
	return true
}

// VerifyRouteCatalog compares every model actually used by a native contract.
// Names are display metadata; all capability, limit and pricing fields remain.
func VerifyRouteCatalog(old, current ai.Catalog, operation ai.Operation, provider ai.ProviderID, api ai.API, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("route model set missing")
	}
	for _, id := range ids {
		var same bool
		switch operation {
		case ai.OperationChat:
			a, ok := old.Lookup(provider, api, id)
			b, found := current.Lookup(provider, api, id)
			a.Name, b.Name = "", ""
			same = ok && found && reflect.DeepEqual(a, b)
		case ai.OperationImage:
			a, ok := old.LookupImage(provider, api, id)
			b, found := current.LookupImage(provider, api, id)
			a.Name, b.Name = "", ""
			same = ok && found && reflect.DeepEqual(a, b)
		case ai.OperationClassifier:
			a, ok := old.LookupClassifier(provider, api, id)
			b, found := current.LookupClassifier(provider, api, id)
			a.Name, b.Name = "", ""
			same = ok && found && reflect.DeepEqual(a, b)
		}
		if !same {
			return fmt.Errorf("route model data changed or missing: %s", id)
		}
	}
	return nil
}

// MeasuredCatalog reads the original snapshot; it never relabels its hash.
// Older bundles without a snapshot are usable only while the full hash matches.
func MeasuredCatalog(dir, version, hash string, current ai.Catalog) (ai.Catalog, error) {
	if h, err := current.Hash(); err == nil && h == hash && current.Version == version {
		return current, nil
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return ai.Catalog{}, err
	}
	defer root.Close()
	for _, name := range []string{"model-catalog.json", "regional-catalog.json", "minimax-catalog.json", "catalog.json", "MM11-configuration/formal-catalog.json"} {
		info, err := root.Lstat(name)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 4<<20 {
			continue
		}
		raw, err := root.ReadFile(name)
		if err != nil {
			continue
		}
		var c ai.Catalog
		if json.Unmarshal(raw, &c) != nil {
			continue
		}
		if h, err := c.Hash(); err == nil && h == hash && c.Version == version {
			return c, nil
		}
	}
	return measuredGitCatalog(dir, version, hash)
}
