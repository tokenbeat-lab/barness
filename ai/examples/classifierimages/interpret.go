package main

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"

	"github.com/tokenbeat-lab/barness/ai"
)

type pixels struct{ Left, Top, Right, Bottom float64 }
type legendImage struct {
	Pointer    string `json:"pointer"`
	ImageIndex int    `json:"imageIndex"`
	MimeType   string `json:"mimeType"`
	SHA256     string `json:"sha256"`
	File       string `json:"file"`
}
type answerView struct {
	Kind           string                   `json:"kind"`
	Probability    *float64                 `json:"probability,omitempty"`
	Confidence     ai.Nullable[float64]     `json:"confidence,omitzero"`
	Choice         string                   `json:"choice,omitempty"`
	Score          *float64                 `json:"score,omitempty"`
	Legend         map[int]json.RawMessage  `json:"legend,omitempty"`
	LegendImages   []legendImage            `json:"legendImages,omitempty"`
	Box            *ai.Box                  `json:"box,omitempty"`
	SearchImage    *ai.ClassifierImageInput `json:"searchImage,omitempty"`
	OriginalPixels *pixels                  `json:"originalPixels,omitempty"`
}

// Result indices are local to this final result. A MIME/hash lookup chooses a
// Host-owned synthetic file; it grants neither access nor Provider ownership.
func interpret(res ai.ClassifierResult, files []fixture, search ai.ClassifierImageInput) (map[string]answerView, error) {
	out := map[string]answerView{}
	for name, answer := range res.Answers {
		v := answerView{}
		switch a := answer.(type) {
		case ai.RefusalAnswer:
			v.Kind = "refusal"
		case ai.BoolAnswer:
			v.Kind = "bool"
			v.Probability = &a.Probability
		case ai.ChoiceAnswer:
			v.Kind = "choice"
			v.Choice = a.Choice
			v.Confidence = ai.Value(a.Confidence)
		case ai.ScoreAnswer:
			v.Kind = "score"
			v.Score = &a.Score
			v.Confidence = ai.Value(a.Confidence)
			// The view owns mutable JSON independently of the canonical result.
			if a.Legend != nil {
				v.Legend = make(map[int]json.RawMessage, len(a.Legend))
				for level, raw := range a.Legend {
					v.Legend[level] = slices.Clone(raw)
				}
			}
			for _, ref := range a.LegendImageReferences {
				if ref.ImageIndex < 0 || ref.ImageIndex >= len(res.ImageInputs) {
					return nil, errors.New("legend index outside final ImageInputs")
				}
				identity := res.ImageInputs[ref.ImageIndex]
				file, ok := findFile(files, identity)
				if !ok {
					return nil, errors.New("Host lacks final legend file")
				}
				// This example's two known rubric slots are relative to Legend, not to
				// /questions. Keep the slots null and render images alongside the rubric.
				found := false
				for level, raw := range a.Legend {
					if ref.Pointer == "/"+strconv.Itoa(level)+"/example" {
						var slot struct{ Example json.RawMessage }
						if err := json.Unmarshal(raw, &slot); err != nil {
							return nil, err
						}
						found = string(slot.Example) == "null"
					}
				}
				if !found {
					return nil, errors.New("unexpected example legend slot")
				}
				v.LegendImages = append(v.LegendImages, legendImage{ref.Pointer, ref.ImageIndex, identity.MimeType, identity.SHA256, file.Name})
			}
		case ai.BoxAnswer:
			// Question-to-search-image association is explicit Host state. Neither the
			// first result index nor the independent existence answer defines it.
			if name != "location" {
				return nil, errors.New("Host has no search-image association for Box")
			}
			file, ok := findFile(files, search)
			if !ok {
				return nil, errors.New("Host lacks search-image geometry")
			}
			visible := false
			for _, final := range res.ImageInputs {
				if final == search {
					visible = true
				}
			}
			if !visible {
				return nil, errors.New("search image changed without Host attribution update")
			}
			g := file.Geometry
			if g.ClockwiseDegrees != 90 || g.Width != g.Crop.Dx() || g.Height != g.Crop.Dy() {
				return nil, errors.New("unsupported example preprocessing")
			}
			// Sent normalized -> cropped pixels -> clockwise-oriented pixels -> source
			// edges. Edge coordinates use H-x, whereas raster pixel rotation uses H-1-x.
			left := float64(g.Crop.Min.X) + a.Box.Left*float64(g.Width)
			right := float64(g.Crop.Min.X) + a.Box.Right*float64(g.Width)
			top := float64(g.Crop.Min.Y) + a.Box.Top*float64(g.Height)
			bottom := float64(g.Crop.Min.Y) + a.Box.Bottom*float64(g.Height)
			v.Kind = "box"
			v.Box = &a.Box
			v.Confidence = a.Confidence
			v.SearchImage = &search
			v.OriginalPixels = &pixels{top, float64(g.OriginalHeight) - right, bottom, float64(g.OriginalHeight) - left}
		default:
			return nil, errors.New("unhandled classifier answer")
		}
		out[name] = v
	}
	return out, nil
}
func findFile(files []fixture, identity ai.ClassifierImageInput) (fixture, bool) {
	for _, f := range files {
		if f.Identity == identity {
			return f, true
		}
	}
	return fixture{}, false
}
