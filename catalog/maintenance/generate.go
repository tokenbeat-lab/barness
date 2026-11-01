package maintenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

type orderedEntry struct {
	provider ai.ProviderID
	entry    Entry
}
type compositionPart struct {
	Provider ai.ProviderID `json:"provider"`
	Version  uint64        `json:"version"`
	Hash     string        `json:"hash"`
}

// Compose produces a runtime catalog from the selected Provider versions.
// Provenance, human decisions and withdrawn definitions remain in State.
func Compose(state State) (ai.Catalog, error) {
	if strings.TrimSpace(state.RulesRevision) == "" {
		return ai.Catalog{}, fmt.Errorf("rulesRevision is required")
	}
	providers := slices.Clone(state.Providers)
	slices.SortFunc(providers, func(a, b ProviderRecord) int { return strings.Compare(string(a.ID), string(b.ID)) })
	parts := make([]compositionPart, 0, len(providers))
	entries := make([]orderedEntry, 0)
	positions := map[string]bool{}
	for i, p := range providers {
		if i > 0 && providers[i-1].ID == p.ID {
			return ai.Catalog{}, fmt.Errorf("duplicate Provider %s", p.ID)
		}
		hash, err := ProviderHash(p)
		if err != nil {
			return ai.Catalog{}, err
		}
		parts = append(parts, compositionPart{p.ID, p.Version, hash})
		for _, e := range p.Entries {
			key := fmt.Sprintf("%s/%d", e.Operation, e.Position)
			if positions[key] {
				return ai.Catalog{}, fmt.Errorf("duplicate collection position %s", key)
			}
			positions[key] = true
			if e.Lifecycle == Active {
				entries = append(entries, orderedEntry{p.ID, e})
			}
		}
	}
	raw, err := json.Marshal(parts)
	if err != nil {
		return ai.Catalog{}, err
	}
	catalog := ai.Catalog{Version: "catalog-" + strings.TrimPrefix(digest(raw), "sha256:"), Models: []ai.Model{}}
	slices.SortFunc(entries, func(a, b orderedEntry) int {
		if cmp := strings.Compare(string(a.entry.Operation), string(b.entry.Operation)); cmp != 0 {
			return cmp
		}
		return a.entry.Position - b.entry.Position
	})
	for _, item := range entries {
		if err := appendDefinition(&catalog, item.entry); err != nil {
			return ai.Catalog{}, err
		}
	}
	if _, err := catalog.Hash(); err != nil {
		return ai.Catalog{}, err
	}
	return catalog, nil
}

// ProviderHash identifies active runtime content and its stable collection
// positions. It excludes the version and provenance: metadata refresh alone
// cannot change a runtime price snapshot or impersonate a content version.
func ProviderHash(provider ProviderRecord) (string, error) {
	if strings.TrimSpace(string(provider.ID)) == "" || provider.Version == 0 {
		return "", fmt.Errorf("Provider ID and positive version are required")
	}
	entries := slices.Clone(provider.Entries)
	slices.SortFunc(entries, func(a, b Entry) int {
		if cmp := strings.Compare(string(a.Operation), string(b.Operation)); cmp != 0 {
			return cmp
		}
		return a.Position - b.Position
	})
	type contentEntry struct {
		Operation  ai.Operation    `json:"operation"`
		Position   int             `json:"position"`
		Definition json.RawMessage `json:"definition"`
	}
	content := make([]contentEntry, 0, len(entries))
	identities := map[string]bool{}
	for _, entry := range entries {
		if entry.Position < 0 {
			return "", fmt.Errorf("negative collection position")
		}
		if entry.Lifecycle != Active && entry.Lifecycle != Withdrawn {
			return "", fmt.Errorf("invalid lifecycle %q", entry.Lifecycle)
		}
		var identity struct {
			Provider ai.ProviderID `json:"provider"`
			API      ai.API        `json:"api"`
			ID       string        `json:"id"`
		}
		if err := json.Unmarshal(entry.Definition, &identity); err != nil {
			return "", fmt.Errorf("invalid model identity: %w", err)
		}
		if identity.Provider != provider.ID || strings.TrimSpace(string(identity.API)) == "" || strings.TrimSpace(identity.ID) == "" {
			return "", fmt.Errorf("incomplete or mismatched model identity")
		}
		key := fmt.Sprintf("%s\x00%s\x00%s", entry.Operation, identity.API, identity.ID)
		if identities[key] {
			return "", fmt.Errorf("duplicate model identity")
		}
		identities[key] = true
		single := ai.Catalog{Version: "maintenance-validation"}
		if err := appendDefinition(&single, entry); err != nil {
			return "", err
		}
		if _, err := single.Hash(); err != nil {
			return "", err
		}
		for field, decision := range entry.Fields {
			if !strings.HasPrefix(field, "/") || strings.TrimSpace(decision.Reason) == "" {
				return "", fmt.Errorf("field decisions require a JSON pointer and reason")
			}
			var definition map[string]any
			if json.Unmarshal(entry.Definition, &definition) != nil || !hasField(definition, field) {
				return "", fmt.Errorf("field decision %q does not name a defined field", field)
			}
			switch decision.Ownership {
			case OwnershipSource, OwnershipManual, OwnershipHistorical:
			default:
				return "", fmt.Errorf("invalid field ownership %q", decision.Ownership)
			}
		}
		if entry.Lifecycle == Active {
			// Marshal the typed definition so harmless whitespace/property ordering
			// in the maintenance JSON cannot change a Provider content digest.
			var typed any
			switch entry.Operation {
			case ai.OperationChat:
				typed = single.Models[0]
			case ai.OperationImage:
				typed = single.ImageModels[0]
			case ai.OperationClassifier:
				typed = single.ClassifierModels[0]
			}
			definition, err := json.Marshal(typed)
			if err != nil {
				return "", err
			}
			content = append(content, contentEntry{entry.Operation, entry.Position, definition})
		}
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return digest(raw), nil
}

func appendDefinition(c *ai.Catalog, e Entry) error {
	if err := validateDefinitionJSON(e.Definition, e.Operation); err != nil {
		return err
	}
	switch e.Operation {
	case ai.OperationChat:
		var m ai.Model
		if err := decode(e.Definition, &m); err != nil {
			return fmt.Errorf("chat definition: %w", err)
		}
		c.Models = append(c.Models, m)
	case ai.OperationImage:
		var m ai.ImageModel
		if err := decode(e.Definition, &m); err != nil {
			return fmt.Errorf("image definition: %w", err)
		}
		c.ImageModels = append(c.ImageModels, m)
	case ai.OperationClassifier:
		var m ai.ClassifierModel
		if err := decode(e.Definition, &m); err != nil {
			return fmt.Errorf("classifier definition: %w", err)
		}
		c.ClassifierModels = append(c.ClassifierModels, m)
	default:
		return fmt.Errorf("unsupported operation %q", e.Operation)
	}
	return nil
}

// Generate emits the exact runtime envelope used for both external files and
// the embedded builtin. It contains no timestamps or maintenance metadata.
func Generate(state State) ([]byte, error) {
	catalog, err := Compose(state)
	if err != nil {
		return nil, err
	}
	hash, err := catalog.Hash()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}{hash, catalog})
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
