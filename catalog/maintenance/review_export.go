package maintenance

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/tokenbeat-lab/barness/catalog"
)

type ReviewInput struct {
	Kind        string `json:"kind"`
	Role        string `json:"role"`
	Name        string `json:"name,omitempty"`
	Digest      string `json:"digest"`
	Path        string `json:"path"`
	URL         string `json:"url,omitempty"`
	FetchedAt   string `json:"fetchedAt,omitempty"`
	CatalogHash string `json:"catalogHash,omitempty"`
	CandidateID string `json:"candidateId,omitempty"`
}

type ReviewReference struct {
	Scope  string `json:"scope"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type ReviewExport struct {
	Status     string            `json:"status"`
	Revision   string            `json:"revision"`
	Files      map[string]string `json:"files"`
	Inputs     []ReviewInput     `json:"inputs"`
	References []ReviewReference `json:"references"`
}

var providerFilename = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// ExportReview produces a new review directory from one approved revision. Raw
// source/manual inputs are self-contained and deduplicated; qualification/check
// originals remain explicit external references, never new vendor assertions.
// review.json is written last and is the completion marker. Failure never moves
// current or overwrites a prior review; partial failed directories stay visible.
func ExportReview(store, repositoryRoot, directory string, dryRun, confirmed bool) (ReviewExport, error) {
	current, err := ReadCurrent(store)
	if err != nil {
		return ReviewExport{}, err
	}
	if directory == "" {
		return ReviewExport{}, failure("invalid_argument", "a new review directory is required")
	}
	if _, err := os.Lstat(directory); err == nil {
		return ReviewExport{}, failure("output_exists", "review destination must not already exist")
	} else if !os.IsNotExist(err) {
		return ReviewExport{}, failure("storage_failed", "inspect review destination: %v", err)
	}
	files := map[string][]byte{"snapshot.json": append([]byte(nil), current.Snapshot...)}
	for name, value := range map[string]any{"state.json": current.State, "revision.json": current} {
		raw, err := jsonBytes(value)
		if err != nil {
			return ReviewExport{}, err
		}
		files[name] = raw
	}
	for _, provider := range current.State.Providers {
		if !providerFilename.MatchString(string(provider.ID)) {
			return ReviewExport{}, failure("invalid_definition", "Provider ID cannot be used as a review filename")
		}
		raw, err := jsonBytes(provider)
		if err != nil {
			return ReviewExport{}, err
		}
		files["providers/"+string(provider.ID)+".json"] = raw
	}
	result := ReviewExport{Status: "preview", Revision: current.ID, Files: map[string]string{}, Inputs: []ReviewInput{}, References: []ReviewReference{}}
	if err := reviewProvenance(store, repositoryRoot, &result, files); err != nil {
		return result, err
	}
	for name, raw := range files {
		result.Files[name] = digest(raw)
	}
	if dryRun {
		return result, nil
	}
	if !confirmed {
		return result, failure("confirmation_required", "review export requires explicit --confirm")
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		return result, failure("storage_failed", "create exclusive review directory: %v", err)
	}
	for _, name := range sortedCheckKeys(result.Files) {
		if err := writeImmutable(filepath.Join(directory, filepath.FromSlash(name)), files[name]); err != nil {
			return result, err
		}
	}
	result.Status = "exported"
	raw, err := jsonBytes(result)
	if err != nil {
		return result, err
	}
	if err := writeImmutable(filepath.Join(directory, "review.json"), raw); err != nil {
		result.Status = "export_failed"
		return result, err
	}
	return result, nil
}

func reviewProvenance(store, repositoryRoot string, result *ReviewExport, files map[string][]byte) error {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return err
	}
	history, err := History(store)
	if err != nil {
		return err
	}
	seenHistorical := map[string]bool{}
	needed := map[string]bool{}
	available := map[string]bool{}
	addBlob := func(raw []byte) string {
		sum := digest(raw)
		name, _ := objectName(sum)
		path := "inputs/" + name + ".bin"
		files[path] = append([]byte(nil), raw...)
		available[sum] = true
		return path
	}
	for _, revision := range history {
		revisionName, err := objectName(revision.ID)
		if err != nil {
			return err
		}
		revisionPath := "revisions/" + revisionName + ".json"
		if _, err := os.Stat(filepath.Join(store, filepath.FromSlash(revisionPath))); err == nil {
			if err := addReviewReference(result, "store", store, revisionPath); err != nil {
				return err
			}
		} else if revision.CandidateID != "" {
			return failure("source_unavailable", "committed historical revision object is unavailable")
		}
		for _, provider := range revision.State.Providers {
			for _, entry := range provider.Entries {
				if entry.Source == nil {
					continue
				}
				source := entry.Source
				if source.Kind != "historical" {
					needed[source.Digest] = true
					continue
				}
				key := source.Path + "\x00" + source.Digest
				if seenHistorical[key] {
					continue
				}
				seenHistorical[key] = true
				relative, err := checkRelative(root, source.Path)
				if err != nil {
					return err
				}
				raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
				if err != nil {
					return failure("source_unavailable", "historical source bytes unavailable")
				}
				// Historical links name the original catalog hash, not the outer
				// file's byte digest. Keep both independent identities in the export.
				_, report, err := catalog.Load(raw)
				if err != nil || report.SourceHash != source.Digest {
					return failure("source_changed", "historical catalog identity differs")
				}
				result.Inputs = append(result.Inputs, ReviewInput{Kind: "historical", Role: "catalog", Name: source.Path, Digest: digest(raw), Path: addBlob(raw), CatalogHash: source.Digest})
			}
		}
		if revision.CandidateID == "" {
			continue
		}
		candidate, err := ReadCandidate(store, revision.CandidateID)
		if err != nil {
			return err
		}
		name, err := objectName(candidate.ID)
		if err != nil {
			return err
		}
		if err := addReviewReference(result, "store", store, "candidates/"+name+".json"); err != nil {
			return err
		}
		for _, input := range candidate.Inputs {
			kind := input.Kind
			if kind == "" {
				kind = "models.dev"
			}
			result.Inputs = append(result.Inputs, ReviewInput{Kind: kind, Role: "source", Digest: input.Digest, Path: addBlob(input.Bytes), URL: input.URL, FetchedAt: input.FetchedAt, CandidateID: candidate.ID})
		}
		for _, input := range candidate.LocalInputs {
			result.Inputs = append(result.Inputs, ReviewInput{Kind: "local", Role: input.Role, Name: input.Name, Digest: input.Digest, Path: addBlob(input.Bytes), CandidateID: candidate.ID})
		}
		if revision.CheckReference != "" {
			relative, err := checkRelative(root, revision.CheckReference)
			if err != nil {
				return err
			}
			if err := addReviewReference(result, "repository", root, relative); err != nil {
				return err
			}
		}
	}
	for sum := range needed {
		if !available[sum] {
			return failure("source_unavailable", "approved source capture is not in retained applied history; retain its original source/review inputs before export")
		}
	}
	return nil
}

func addReviewReference(result *ReviewExport, scope, root, path string) error {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return failure("source_unavailable", "retained %s reference is unavailable: %s", scope, path)
	}
	result.References = append(result.References, ReviewReference{Scope: scope, Path: path, Digest: digest(raw)})
	return nil
}
