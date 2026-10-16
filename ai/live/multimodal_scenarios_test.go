//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type multimodalScenario struct {
	id, entry string
	request   ai.ClassifierRequest
}

func multimodalImages(t *testing.T) []ai.Image {
	t.Helper()
	raster := image.NewRGBA(image.Rect(0, 0, 280, 280))
	for y := 0; y < 280; y++ {
		for x := 0; x < 280; x++ {
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			if x >= 70 && x < 210 && y >= 70 && y < 210 {
				c = color.RGBA{R: 230, G: 30, B: 30, A: 255}
			}
			raster.SetRGBA(x, y, c)
		}
	}
	var p, j bytes.Buffer
	if err := png.Encode(&p, raster); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&j, raster, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return []ai.Image{{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(p.Bytes())}, {MimeType: "image/jpeg", Data: base64.StdEncoding.EncodeToString(j.Bytes())}}
}

func multimodalScenarios(t *testing.T, route string) []multimodalScenario {
	files := multimodalImages(t)
	var out []multimodalScenario
	for _, entry := range []string{"plain", "hooked"} {
		names := []string{"default-context", "explicit-context", "ordered-reuse"}
		if route == "jev" {
			names = append(names, "local-state-instructions", "visual-criteria", "box-existence")
		}
		for _, name := range names {
			r := ai.ClassifierRequest{State: json.RawMessage(`{"state":"preserved","images":["business"],"task":"synthetic red rectangle"}`), Images: files[:1], Questions: map[string]ai.ClassifierQuestion{"exists": ai.BoolQuestion{Instructions: json.RawMessage(`"Is a red rectangle present in the first context image?"`)}}}
			switch name {
			case "explicit-context":
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementContext}}
			case "ordered-reuse":
				r.Images = files
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 1, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}, {ImageIndex: 1, Placement: ai.ImagePlacementContext}}
			case "local-state-instructions":
				r.Images = files
				r.State = json.RawMessage(`{"photo":null,"task":"synthetic red rectangle"}`)
				r.Questions["exists"] = ai.BoolQuestion{Instructions: json.RawMessage(`"Is a red rectangle present in the State photo?"`)}
				r.Questions["example"] = ai.BoolQuestion{Instructions: json.RawMessage(`{"task":"Does this example contain a red rectangle?","example":null}`)}
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}, {ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/example/instructions/example"}}
			case "visual-criteria":
				r.Images = files
				r.Questions = map[string]ai.ClassifierQuestion{
					"choice": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Select the matching visual standard"`), Criteria: map[string]json.RawMessage{"match": json.RawMessage(`null`)}},
					"bool":   ai.BoolQuestion{Instructions: json.RawMessage(`"Does the State describe the true visual standard?"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`null`), False: json.RawMessage(`null`)}},
					"score":  ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate using these ordered visual standards"`), Criteria: []json.RawMessage{json.RawMessage(`{"example":null}`), json.RawMessage(`{"example":null}`)}},
				}
				for i, path := range []string{"/questions/choice/criteria/match", "/questions/bool/criteria/true", "/questions/bool/criteria/false", "/questions/score/criteria/0/example", "/questions/score/criteria/1/example"} {
					r.ImageReferences = append(r.ImageReferences, ai.ImageReference{ImageIndex: i % 2, Placement: ai.ImagePlacementValue, Pointer: path})
				}
			case "box-existence":
				r.Questions["location"] = ai.BoxQuestion{Instructions: json.RawMessage(`"Locate the red rectangle in the context image. Answer existence separately."`)}
			}
			out = append(out, multimodalScenario{name + "-" + entry, entry, r})
		}
	}
	return out
}

func multimodalHooks(files []ai.Image) ai.Hooks {
	return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		// Only swap the two known files. Budget reservation covers the larger
		// encoded file at every source/occurrence before the callback runs.
		if assets, ok := p.Body["assets"].(map[string]any); ok {
			for _, v := range assets {
				a := v.(map[string]any)
				if a["data_base64"] == files[0].Data {
					a["data_base64"], a["format"] = files[1].Data, "jpeg"
				} else {
					a["data_base64"], a["format"] = files[0].Data, "png"
				}
			}
		} else {
			var walk func(any)
			walk = func(v any) {
				switch v := v.(type) {
				case map[string]any:
					for k, x := range v {
						if s, ok := x.(string); ok && k == "image_url" {
							for i, img := range files {
								if strings.HasSuffix(s, ","+img.Data) {
									other := files[1-i]
									v[k] = "data:" + other.MimeType + ";base64," + other.Data
									break
								}
							}
						} else {
							walk(x)
						}
					}
				case []any:
					for _, x := range v {
						walk(x)
					}
				}
			}
			walk(p.Body)
		}
		return ai.KeepPayload(), nil
	}}
}
