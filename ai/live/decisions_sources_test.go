//go:build live

package live

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

const decisionsSourceHash = "61be0c194e7231aa5c3b986c849c2b23935ae6a1830038413530eb0d3618b8fe"
const candidateVersion = "decisions-candidate-2026-10-13.1"
const candidateReportName = "candidate-decisions-report"

var decisionsCandidate = combo{name: "openai-decisions-candidate", spec: "P10", operation: ai.OperationClassifier, provider: ai.ProviderOpenAI, api: ai.APIOpenAIDecisions, endpoint: "https://api.openai.com/v1", model: "gpt-6-luna", sdk: func() string { return "direct HTTP (OpenAI Decisions; ADR-0026)" }}

type candidateSource struct {
	File   string `json:"file"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}
type candidateSources struct {
	Schema              int               `json:"schema"`
	RetrievedAt         time.Time         `json:"retrievedAt"`
	Sources             []candidateSource `json:"sources"`
	CandidateFileSHA256 string            `json:"candidateFileSHA256"`
	Claims              map[string]string `json:"claims"`
	Pi                  json.RawMessage   `json:"pi"`
	ReferenceExtraction json.RawMessage   `json:"referenceExtraction"`
}

func sha256Text(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// The manifest itself is pinned, as well as each complete snapshot. A rewritten
// manifest cannot bless changed sources. Updating evidence requires a new pin.
func loadDecisionsCandidate(cs *evidence.Case) (ai.Catalog, candidateSources, error) {
	read := func(name string) ([]byte, error) {
		path := filepath.Join("testdata", "decisions", name)
		raw, err := readDecisionsFile(path, 1<<20)
		if err == nil {
			cs.Fixture(name, raw)
		}
		return raw, err
	}
	raw, err := read("sources.json")
	if err != nil {
		return ai.Catalog{}, candidateSources{}, err
	}
	var sources candidateSources
	if sha256Text(raw) != decisionsSourceHash || json.Unmarshal(raw, &sources) != nil || sources.Schema != 1 {
		return ai.Catalog{}, sources, fmt.Errorf("candidate source manifest differs from fixed pin")
	}
	for _, s := range sources.Sources {
		raw, err = read(s.File)
		if err != nil || sha256Text(raw) != s.SHA256 {
			return ai.Catalog{}, sources, fmt.Errorf("candidate source %s differs from fixed snapshot", s.File)
		}
	}
	raw, err = read("candidate.json")
	if err != nil || sha256Text(raw) != sources.CandidateFileSHA256 {
		return ai.Catalog{}, sources, fmt.Errorf("candidate catalog differs from fixed snapshot")
	}
	var cat ai.Catalog
	if err = json.Unmarshal(raw, &cat); err != nil {
		return cat, sources, err
	}
	if cat.Version != candidateVersion || len(cat.Models) != 0 || len(cat.ImageModels) != 0 || len(cat.ClassifierModels) != 1 {
		return cat, sources, fmt.Errorf("candidate catalog is not independent")
	}
	return cat, sources, nil
}

func readDecisionsFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("required candidate input is missing")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("candidate input is not a bounded regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, fmt.Errorf("candidate input exceeds its read budget")
	}
	return raw, nil
}
