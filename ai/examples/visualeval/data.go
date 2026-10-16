// Command visualeval is a fixed-task Host evaluation. Truth never enters requests
// or the controlled Provider; only this Host's scorer reads the annotation file.
package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sort"

	"github.com/tokenbeat-lab/barness/ai"
)

//go:embed testdata/*
var data embed.FS

type sample struct {
	ID, File, Source, Mime, Orientation, Transform, Label, Rationale string
	Width, Height                                                    int
	Exists                                                           bool
	Box                                                              *ai.Box
}
type dataset struct {
	Version, Source, Scope, AnnotationPolicy string
	Labels                                   map[string]string
	Samples                                  []sample
}
type experiment struct {
	Version, BoxPolicy, Classification, Existence, Localization                                                                                    string
	CommonCount                                                                                                                                    int
	ExistenceThreshold                                                                                                                             float64
	MaxCalls, MaxAttempts, MaxRetries, MaxQuestions, MaxOccurrences, MaxSourceBytes, MaxExpandedBytes, CallSeconds, ProcessSeconds, CleanupSeconds int
	ProtocolEvidence                                                                                                                               map[string]string
}
type plan struct {
	dataset    dataset
	experiment experiment
	catalog    ai.Catalog
	pins       map[string]string
	route      string
	model      ai.ClassifierModel
}

func digest(raw []byte) string { s := sha256.Sum256(raw); return hex.EncodeToString(s[:]) }
func encoded(v any) []byte {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(raw, '\n')
}
func same(a, b any) bool { return bytes.Equal(encoded(a), encoded(b)) }
func strict(raw []byte, into any) error {
	if len(raw) > 16<<20 {
		return errors.New("document exceeds fixed limit")
	}
	// Independent command decoder: duplicate fields cannot silently change a pin
	// or authorization. Kept local so this example can evolve without a public API.
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 32 {
			return errors.New("JSON depth limit")
		}
		t, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		seen := map[string]bool{}
		for d.More() {
			if delim == '{' {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || seen[s] {
					return errors.New("duplicate JSON field")
				}
				seen[s] = true
			}
			if e := walk(depth + 1); e != nil {
				return e
			}
		}
		_, e = d.Token()
		return e
	}
	if err := walk(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("extra JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(into)
}
func load(route string) (*plan, error) {
	if route != "jev" && route != "decisions" {
		return nil, errors.New("select jev or decisions")
	}
	p := &plan{route: route}
	raw, _ := data.ReadFile("testdata/pins.json")
	if err := strict(raw, &p.pins); err != nil {
		return nil, err
	}
	for name, hash := range p.pins {
		raw, err := data.ReadFile("testdata/" + name)
		if err != nil || digest(raw) != hash {
			return nil, errors.New("frozen input drift")
		}
	}
	for name, dest := range map[string]any{"dataset.json": &p.dataset, "experiment.json": &p.experiment, route + "-catalog.json": &p.catalog} {
		raw, _ := data.ReadFile("testdata/" + name)
		if err := strict(raw, dest); err != nil {
			return nil, err
		}
	}
	if len(p.dataset.Samples) != 12 || len(p.catalog.ClassifierModels) != 1 {
		return nil, errors.New("incomplete fixed plan")
	}
	p.model = p.catalog.ClassifierModels[0]
	seen := map[string]bool{}
	for _, s := range p.dataset.Samples {
		if seen[s.ID] || s.ID == "" || p.dataset.Labels[s.Label] == "" || s.Rationale == "" || s.Orientation == "" || s.Transform == "" || s.Exists != (s.Box != nil) {
			return nil, errors.New("incomplete fixed annotation")
		}
		seen[s.ID] = true
		raw, _ := data.ReadFile("testdata/" + s.File)
		cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil || cfg.Width != s.Width || cfg.Height != s.Height {
			return nil, errors.New("image dimensions drift")
		}
	}
	return p, nil
}
func (p *plan) request(i int) ai.ClassifierRequest {
	s := p.dataset.Samples[i]
	raw, _ := data.ReadFile("testdata/" + s.File)
	criteria := map[string]json.RawMessage{}
	for k, v := range p.dataset.Labels {
		criteria[k] = json.RawMessage(encoded(v))
	}
	// Only task text and image bytes enter the public request. No ID, label,
	// rationale, annotation coordinates or source scene reaches the model.
	r := ai.ClassifierRequest{State: json.RawMessage(`"请观察图片并回答中文问题。"`), Images: []ai.Image{{MimeType: s.Mime, Data: base64.StdEncoding.EncodeToString(raw)}}, Questions: map[string]ai.ClassifierQuestion{
		"kind":   ai.ChoiceQuestion{Instructions: encoded(p.experiment.Classification), Criteria: criteria},
		"exists": ai.BoolQuestion{Instructions: encoded(p.experiment.Existence)},
	}}
	if p.route == "jev" && i >= p.experiment.CommonCount {
		r.Questions["location"] = ai.BoxQuestion{Instructions: encoded(p.experiment.Localization)}
		if i < 10 {
			r.State = json.RawMessage(`{"photo":null,"task":"请观察图片并回答中文问题。"}`)
			r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}}
		} else {
			keys := []string{"exists", "kind", "location"}
			sort.Strings(keys)
			for _, key := range keys {
				text := p.experiment.Existence
				if key == "kind" {
					text = p.experiment.Classification
				}
				if key == "location" {
					text = p.experiment.Localization
				}
				instructions := encoded(map[string]any{"task": text, "photo": nil})
				switch q := r.Questions[key].(type) {
				case ai.ChoiceQuestion:
					q.Instructions = instructions
					r.Questions[key] = q
				case ai.BoolQuestion:
					q.Instructions = instructions
					r.Questions[key] = q
				case ai.BoxQuestion:
					q.Instructions = instructions
					r.Questions[key] = q
				}
				r.ImageReferences = append(r.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: fmt.Sprintf("/questions/%s/instructions/photo", key)})
			}
		}
	}
	return r
}
func (p *plan) group(i int) string {
	if i < 8 {
		return "common"
	}
	if p.route == "jev" {
		return "local"
	}
	return "supplementalContext"
}
