package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tokenbeat-lab/barness/ai"
)

type scenario struct{ Route, Name, Entry string }

func scenarios() []scenario {
	var all []scenario
	for _, route := range []string{"jev", "decisions"} {
		for _, entry := range []string{"plain", "hooked"} {
			for _, name := range []string{"context", "reuse", "state", "instructions", "criteria", "box", "image-revoked", "binding-revoked", "budget", "canceled"} {
				all = append(all, scenario{route, name, entry})
			}
			if route == "jev" {
				for _, name := range []string{"bad-box", "bad-legend", "experiment-off"} {
					all = append(all, scenario{route, name, entry})
				}
			}
		}
	}
	return append(all, scenario{"jev", "hook", "hooked"}, scenario{"decisions", "refusal", "plain"})
}
func (s scenario) id() string { return s.Route + "-" + s.Name + "-" + s.Entry }

// This common request is independent of protocol. Switching route changes
// only Target and Options; local placements are deliberately rejected by Decisions.
func input(name string, files []fixture) ai.ClassifierRequest {
	r := ai.ClassifierRequest{State: json.RawMessage(`{"task":"synthetic red rectangle"}`), Images: []ai.Image{files[0].Image, files[1].Image}, Questions: map[string]ai.ClassifierQuestion{
		"exists": ai.BoolQuestion{Instructions: json.RawMessage(`"Is a red rectangle present in the first context image?"`)},
	}}
	switch name {
	case "reuse", "budget":
		r.ImageReferences = []ai.ImageReference{{ImageIndex: 1, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}, {ImageIndex: 1, Placement: ai.ImagePlacementContext}}
	case "state":
		r.State = json.RawMessage(`{"photo":null,"task":"synthetic red rectangle"}`)
		r.ImageReferences = []ai.ImageReference{{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}}
	case "instructions":
		r.Questions["exists"] = ai.BoolQuestion{Instructions: json.RawMessage(`{"task":"Does the example contain a rectangle?","example":null}`)}
		r.ImageReferences = []ai.ImageReference{{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/exists/instructions/example"}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}}
	case "criteria", "hook", "bad-legend":
		r.State = json.RawMessage(`{"photo":null}`)
		r.Questions["exists"] = ai.BoolQuestion{Instructions: json.RawMessage(`{"task":"Is this rectangle present?","example":null}`)}
		r.Questions["score"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate using the ordered visual standards"`), Criteria: []json.RawMessage{json.RawMessage(`{"example":null}`), json.RawMessage(`{"example":null}`)}}
		r.ImageReferences = []ai.ImageReference{
			{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"},
			{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/exists/instructions/example"},
			{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/score/criteria/0/example"},
			{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/score/criteria/1/example"},
		}
	case "box", "bad-box":
		r.Questions["location"] = ai.BoxQuestion{Instructions: json.RawMessage(`"Locate the red rectangle in the first context image only. The second image is a reference. Answer existence separately."`)}
	}
	return r
}

// Hooks are trusted, protocol-aware Host callbacks. Common input builders never
// create native keys. Discover the adapter's keys from its body, then swap the
// actual encoded files; the final state/legend indices must be read from Result.
func swapImages(files []fixture) ai.Hooks {
	return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		assets, ok := p.Body["assets"].(map[string]any)
		if !ok {
			return ai.KeepPayload(), errors.New("expected Jev assets")
		}
		for _, value := range assets {
			asset, ok := value.(map[string]any)
			if !ok {
				return ai.KeepPayload(), errors.New("expected inline asset")
			}
			switch asset["data_base64"] {
			case files[0].Image.Data:
				asset["data_base64"] = files[1].Image.Data
				asset["format"] = "jpeg"
			case files[1].Image.Data:
				asset["data_base64"] = files[0].Image.Data
				asset["format"] = "png"
			default:
				return ai.KeepPayload(), errors.New("unknown synthetic image")
			}
		}
		return ai.KeepPayload(), nil
	}}
}
