package maintenance

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/data"
)

// AdoptedReview retains the exact completed index and inventory in the seed's
// immutable lineage. Future binaries need not keep the previous embedded seed.
// External references remain references, never fabricated check originals.
type AdoptedReview struct {
	Index []byte            `json:"index"`
	Files map[string][]byte `json:"files"`
}

func builtinSource() (State, []byte, *AdoptedReview, error) {
	state, err := readBuiltinState()
	if err != nil {
		return State{}, nil, nil, err
	}
	snapshot, err := Generate(state)
	if err != nil {
		return State{}, nil, nil, err
	}
	adoption, err := builtinAdoption(state, snapshot)
	return state, snapshot, adoption, err
}

func builtinAdoption(state State, snapshot []byte) (*AdoptedReview, error) {
	raw, err := data.Baseline.ReadFile("baseline/review.json")
	if err != nil && !isMissing(err) {
		return nil, failure("source_unavailable", "read adopted baseline: %v", err)
	}
	adoption := &AdoptedReview{Index: raw, Files: map[string][]byte{}}
	walkErr := fs.WalkDir(data.Baseline, "baseline", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path == "baseline/README.md" || path == "baseline/review.json" {
			return nil
		}
		raw, err := data.Baseline.ReadFile(path)
		if err != nil {
			return err
		}
		adoption.Files[strings.TrimPrefix(path, "baseline/")] = raw
		return nil
	})
	if walkErr != nil {
		return nil, failure("source_unavailable", "read adopted baseline inventory: %v", walkErr)
	}
	if len(raw) == 0 {
		for _, provider := range state.Providers {
			for _, entry := range provider.Entries {
				if entry.Source != nil && entry.Source.Kind != "historical" {
					return nil, failure("source_unavailable", "nonhistorical default sources require the complete adopted baseline review")
				}
			}
		}
		if len(adoption.Files) != 0 {
			return nil, failure("source_unavailable", "adopted baseline completion index is missing")
		}
		return nil, nil
	}
	_, approved, err := validateAdoptedReview(adoption)
	if err != nil {
		return nil, err
	}
	if !equalState(approved.State, state) || !bytes.Equal(approved.Snapshot, snapshot) {
		return nil, failure("source_changed", "adopted baseline differs from selected Provider source")
	}
	for _, provider := range state.Providers {
		path := "providers/" + string(provider.ID) + ".json"
		selected, err := data.Providers.ReadFile(path)
		if err != nil || !bytes.Equal(selected, adoption.Files[path]) {
			return nil, failure("source_changed", "adopted baseline Provider file differs: %s", provider.ID)
		}
	}
	composed, err := Compose(state)
	if err != nil {
		return nil, err
	}
	builtin := ai.BuiltinCatalog()
	expectedHash, _ := composed.Hash()
	actualHash, _ := builtin.Hash()
	if composed.Version != builtin.Version || expectedHash != actualHash {
		return nil, failure("source_changed", "adopted Provider source and compiled runtime default differ")
	}
	return adoption, nil
}

func isMissing(err error) bool { return errors.Is(err, fs.ErrNotExist) }

func equalState(a, b State) bool {
	a.Providers = slices.Clone(a.Providers)
	b.Providers = slices.Clone(b.Providers)
	less := func(a, b ProviderRecord) int { return strings.Compare(string(a.ID), string(b.ID)) }
	slices.SortFunc(a.Providers, less)
	slices.SortFunc(b.Providers, less)
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return bytes.Equal(left, right)
}

func validateAdoptedReview(adoption *AdoptedReview) (ReviewExport, Revision, error) {
	var index ReviewExport
	bad := func(message string) (ReviewExport, Revision, error) {
		return index, Revision{}, failure("source_changed", "invalid adopted baseline: %s", message)
	}
	if decode(adoption.Index, &index) != nil || index.Status != "exported" || len(index.Files) != len(adoption.Files) {
		return bad("completion index or inventory")
	}
	for path, expected := range index.Files {
		if !fs.ValidPath(path) || path == "review.json" || digest(adoption.Files[path]) != expected {
			return bad("indexed file missing or changed: " + path)
		}
	}
	var revision Revision
	if decode(adoption.Files["revision.json"], &revision) != nil {
		return bad("revision unavailable")
	}
	id, err := revisionDigest(revision)
	if err != nil || id != index.Revision || revision.ID != id {
		return bad("revision identity")
	}
	var state State
	if decode(adoption.Files["state.json"], &state) != nil || !equalState(state, revision.State) {
		return bad("state differs from revision")
	}
	snapshot, err := Generate(state)
	if err != nil || !bytes.Equal(snapshot, revision.Snapshot) || !bytes.Equal(snapshot, adoption.Files["snapshot.json"]) {
		return bad("snapshot differs from approved state")
	}
	expectedFiles := map[string]bool{"state.json": true, "snapshot.json": true, "revision.json": true}
	for _, provider := range state.Providers {
		path := "providers/" + string(provider.ID) + ".json"
		var actual ProviderRecord
		if decode(adoption.Files[path], &actual) != nil {
			return bad("Provider file unavailable")
		}
		left, _ := json.Marshal(provider)
		right, _ := json.Marshal(actual)
		if !bytes.Equal(left, right) {
			return bad("Provider file differs from state")
		}
		expectedFiles[path] = true
	}
	available := map[string]bool{}
	sources := map[string][]ReviewInput{}
	historical := map[string]bool{}
	for _, input := range index.Inputs {
		name, err := objectName(input.Digest)
		if err != nil || input.Path != "inputs/"+name+".bin" || digest(adoption.Files[input.Path]) != input.Digest {
			return bad("input identity")
		}
		if input.CandidateID != "" {
			if _, err := objectName(input.CandidateID); err != nil {
				return bad("input candidate identity")
			}
		}
		switch input.Kind {
		case "historical":
			_, report, err := catalog.Load(adoption.Files[input.Path])
			if err != nil || input.Role != "catalog" || !fs.ValidPath(input.Name) || report.SourceHash != input.CatalogHash {
				return bad("historical catalog provenance")
			}
			historical[input.Name+"\x00"+input.CatalogHash] = true
		case "official", "models.dev":
			if input.Role != "source" || strings.TrimSpace(input.URL) == "" {
				return bad("source provenance")
			}
			if _, err := time.Parse(time.RFC3339Nano, input.FetchedAt); err != nil {
				return bad("source fetch time")
			}
			sources[input.Digest] = append(sources[input.Digest], input)
		case "local":
			if (input.Role != "definition" && input.Role != "decisions") || input.Name == "" || input.URL != "" || input.FetchedAt != "" {
				return bad("manual input provenance")
			}
		default:
			return bad("unknown input kind")
		}
		expectedFiles[input.Path] = true
		available[input.Digest] = true
	}
	if len(expectedFiles) != len(index.Files) {
		return bad("unexpected indexed files")
	}
	for path := range index.Files {
		if !expectedFiles[path] {
			return bad("unexpected indexed file")
		}
	}
	for _, reference := range index.References {
		if (reference.Scope != "store" && reference.Scope != "repository") || !fs.ValidPath(reference.Path) {
			return bad("external reference")
		}
		if _, err := objectName(reference.Digest); err != nil {
			return bad("reference digest")
		}
		if reference.OriginReview != "" {
			if _, err := objectName(reference.OriginReview); err != nil {
				return bad("reference origin")
			}
		}
	}
	for _, provider := range state.Providers {
		for _, entry := range provider.Entries {
			if source := entry.Source; source != nil {
				if source.Kind == "historical" {
					if !historical[source.Path+"\x00"+source.Digest] {
						return bad("historical source unavailable")
					}
				} else {
					matched := false
					for _, input := range sources[source.Digest] {
						matched = matched || (input.Kind == source.Kind && input.URL == source.URL && input.FetchedAt == source.FetchedAt)
					}
					if !matched {
						return bad("source capture or provenance differs from maintained association")
					}
				}
			}
			for _, decision := range entry.Fields {
				if strings.HasPrefix(decision.Evidence, "input:") && !available[strings.TrimPrefix(decision.Evidence, "input:")] {
					return bad("manual evidence capture unavailable")
				}
			}
		}
	}
	return index, revision, nil
}
