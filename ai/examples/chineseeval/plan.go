// Package chineseeval is a host-owned, fixed-task evaluation example. It adds
// no business decisions or confidence thresholds to the model protocol library.
package chineseeval

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

type Sample struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Label     string `json:"label"`
	Rationale string `json:"rationale"`
}

type Dataset struct {
	Version          string            `json:"version"`
	Source           string            `json:"source"`
	Scope            string            `json:"scope"`
	AnnotationPolicy string            `json:"annotationPolicy"`
	Count            int               `json:"count"`
	Distribution     map[string]int    `json:"distribution"`
	Labels           map[string]string `json:"labels"`
	Samples          []Sample          `json:"samples"`
}

type Config struct {
	Version            string `json:"version"`
	DatasetVersion     string `json:"datasetVersion"`
	DatasetHash        string `json:"datasetHash"`
	Model              string `json:"model"`
	CatalogVersion     string `json:"catalogVersion"`
	CatalogHash        string `json:"catalogHash"`
	QuestionKey        string `json:"questionKey"`
	Instructions       string `json:"instructions"`
	MaxCalls           int    `json:"maxCalls"`
	MaxQuestions       int    `json:"maxQuestions"`
	MaxAttempts        int    `json:"maxAttempts"`
	MaxRetries         int    `json:"maxRetries"`
	TimeoutSeconds     int    `json:"timeoutSeconds"`
	CallTimeoutSeconds int    `json:"callTimeoutSeconds"`
}

// Plan owns validated, independently decoded data. Hashes cover the exact input
// bytes; neither caller mutation nor JSON reformatting silently changes a run.
type Plan struct {
	dataset     Dataset
	config      Config
	datasetHash string
	configHash  string
}

func Hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Load validates the complete task before the host resolves any credentials.
// This example deliberately accepts only synthetic data; using production data
// requires a separate approved data-handling workflow and dataset revision.
func Load(data, config []byte, catalog ai.Catalog) (*Plan, error) {
	p := &Plan{datasetHash: Hash(data), configHash: Hash(config)}
	if len(data) > 1<<20 || len(config) > 16384 || decode(data, &p.dataset) != nil || decode(config, &p.config) != nil {
		return nil, errors.New("evaluation: invalid or oversized input document")
	}
	d, c := p.dataset, p.config
	if d.Version == "" || d.Source != "synthetic" || strings.TrimSpace(d.Scope) == "" || strings.TrimSpace(d.AnnotationPolicy) == "" || d.Count != len(d.Samples) || d.Count < 1 || d.Count > 1000 || len(d.Labels) != 4 {
		return nil, errors.New("evaluation: incomplete synthetic dataset definition")
	}
	for _, key := range []string{"billing", "technical", "delivery", "account"} {
		if strings.TrimSpace(d.Labels[key]) == "" {
			return nil, errors.New("evaluation: task label definition missing")
		}
	}
	counts, ids := map[string]int{}, map[string]bool{}
	for _, s := range d.Samples {
		if s.ID == "" || ids[s.ID] || strings.TrimSpace(s.Text) == "" || len(s.Text) > 4096 || strings.TrimSpace(s.Rationale) == "" || d.Labels[s.Label] == "" {
			return nil, errors.New("evaluation: sample lacks unique identity, text, label or annotation basis")
		}
		ids[s.ID] = true
		counts[s.Label]++
	}
	if len(d.Distribution) != len(d.Labels) {
		return nil, errors.New("evaluation: incomplete label distribution")
	}
	for key := range d.Labels {
		if counts[key] < 1 || counts[key] != d.Distribution[key] {
			return nil, errors.New("evaluation: label count mismatch")
		}
	}
	hash, err := catalog.Hash()
	if err != nil || c.Version == "" || c.DatasetVersion != d.Version || c.DatasetHash != p.datasetHash || c.Model != "jev-1.13.0" || c.CatalogVersion != catalog.Version || c.CatalogHash != hash || c.QuestionKey != "intent" || strings.TrimSpace(c.Instructions) == "" || len(c.Instructions) > 4096 {
		return nil, errors.New("evaluation: task, model, dataset or price snapshot drift")
	}
	found := false
	for _, m := range catalog.ClassifierModels {
		if m.ID == c.Model && m.Provider == ai.ProviderTypeSafe && m.API == ai.APITypeSafeSystemOne {
			found = true
		}
	}
	if !found || c.MaxCalls < 1 || c.MaxCalls > d.Count || c.MaxQuestions < 1 || c.MaxQuestions > d.Count || c.MaxAttempts < 1 || c.MaxAttempts > d.Count || c.MaxRetries != 0 || c.TimeoutSeconds < 1 || c.TimeoutSeconds > 600 || c.CallTimeoutSeconds < 1 || c.CallTimeoutSeconds > 60 {
		return nil, errors.New("evaluation: missing fixed model or invalid finite budget")
	}
	return p, nil
}

// Config returns only value fields, for host assembly of the fixed binding.
func (p *Plan) Config() Config { return p.config }
