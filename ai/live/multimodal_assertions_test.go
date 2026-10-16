//go:build live

package live

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Read the final native body independently of the adapter and public indices.
func multimodalNative(route multimodalRoute, body string) (map[string]any, []ai.ClassifierImageInput, multimodalResources, error) {
	var native map[string]any
	var images []ai.ClassifierImageInput
	var resources multimodalResources
	if json.Unmarshal([]byte(body), &native) != nil || native["model"] != route.model {
		return nil, nil, resources, fmt.Errorf("wrong native model")
	}
	seen := map[string]bool{}
	appendImage := func(key, mime, data string) error {
		raw, err := base64.StdEncoding.Strict().DecodeString(data)
		if err != nil {
			return err
		}
		resources.References++
		resources.ExpandedBytes += len(raw)
		if !seen[key] {
			resources.Sources++
			resources.SourceBytes += len(raw)
			seen[key] = true
		}
		if route.name == "decisions" || seen[key+"#metadata"] == false {
			images = append(images, ai.ClassifierImageInput{MimeType: mime, SHA256: sha256Text(raw)})
			seen[key+"#metadata"] = true
		}
		return nil
	}
	if route.name == "decisions" {
		qs, ok := native["questions"].([]any)
		if !ok {
			return nil, nil, resources, fmt.Errorf("questions missing")
		}
		resources.Questions = len(qs)
		input, ok := native["input"].([]any)
		if !ok || len(input) != 1 {
			return nil, nil, resources, fmt.Errorf("input missing")
		}
		message, ok := input[0].(map[string]any)
		if !ok || message["role"] != "user" {
			return nil, nil, resources, fmt.Errorf("user input missing")
		}
		blocks, ok := message["content"].([]any)
		if !ok {
			return nil, nil, resources, fmt.Errorf("content missing")
		}
		for i, v := range blocks {
			block := v.(map[string]any)
			if i == 0 {
				if block["type"] != "input_text" {
					return nil, nil, resources, fmt.Errorf("text first required")
				}
				resources.StateBytes = len(block["text"].(string))
				continue
			}
			url, ok := block["image_url"].(string)
			if !ok {
				return nil, nil, resources, fmt.Errorf("image URL missing")
			}
			head, data, ok := strings.Cut(url, ",")
			if !ok {
				return nil, nil, resources, fmt.Errorf("inline image missing")
			}
			mime := strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
			if err := appendImage(url, mime, data); err != nil {
				return nil, nil, resources, err
			}
		}
	} else {
		assets, ok := native["assets"].(map[string]any)
		if !ok {
			return nil, nil, resources, fmt.Errorf("assets missing")
		}
		state, _ := json.Marshal(native["state"])
		resources.StateBytes = len(state)
		var walk func(any) error
		walk = func(v any) error {
			switch v := v.(type) {
			case string:
				if raw, ok := assets[v]; ok {
					a := raw.(map[string]any)
					mime := "image/" + a["format"].(string)
					if mime == "image/jpg" {
						mime = "image/jpeg"
					}
					return appendImage(v, mime, a["data_base64"].(string))
				}
			case map[string]any:
				keys := make([]string, 0, len(v))
				for k := range v {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					if err := walk(v[k]); err != nil {
						return err
					}
				}
			case []any:
				for _, x := range v {
					if err := walk(x); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := walk(native["state"]); err != nil {
			return nil, nil, resources, err
		}
		qs, ok := native["questions"].(map[string]any)
		if !ok {
			return nil, nil, resources, fmt.Errorf("questions missing")
		}
		resources.Questions = len(qs)
		keys := make([]string, 0, len(qs))
		for k := range qs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			q := qs[k].(map[string]any)
			for _, part := range []string{"instructions", "criteria"} {
				if err := walk(q[part]); err != nil {
					return nil, nil, resources, err
				}
			}
		}
	}
	return native, images, resources, nil
}
func multimodalNativeResources(route multimodalRoute, body string) (multimodalResources, error) {
	_, _, r, err := multimodalNative(route, body)
	return r, err
}

func validateMultimodal(cs *evidence.Case, route multimodalRoute, sc multimodalScenario, res ai.ClassifierResult, ex exchange, obs []ai.Observation) {
	native, images, _, err := multimodalNative(route, ex.RequestBody)
	if !cs.Check("final native images parsed", err == nil, "%v", err) {
		return
	}
	cs.Check("final image identities", reflect.DeepEqual(images, res.ImageInputs), "final MIME/hash differs")
	actual, ok := res.ResponseModel.Get()
	cs.Check("real response model recorded", ok && strings.TrimSpace(actual) != "", "missing model")
	cs.Check("full typed result", res.StopReason == ai.StopReasonStop && len(res.Answers) == len(sc.request.Questions), "incomplete answers")
	cs.Check("complete Observer lifecycle", len(obs) == 4 && obs[3].Kind == ai.ObservationCallFinished && reflect.DeepEqual(obs[3].Call, res.Metadata) && reflect.DeepEqual(obs[3].Usage, res.Usage), "observer mismatch")
	cs.Check("usage presence retained", res.Metadata.Attempts[0].UsageReporting != ai.UsageUnreported && reflect.DeepEqual(res.Metadata.Attempts[0].Usage, res.Usage), "usage absent or differs")
	if route.name == "jev" {
		cs.Check("explicit image experiment", ex.RequestHeaders["typesafe-danger-zone-images"] == "true", "missing image opt-in")
	}
	for id, q := range sc.request.Questions {
		switch q.(type) {
		case ai.BoolQuestion:
			_, ok := res.Answers[id].(ai.BoolAnswer)
			cs.Check("typed independent bool", ok, "missing bool")
		case ai.ChoiceQuestion:
			_, ok := res.Answers[id].(ai.ChoiceAnswer)
			cs.Check("typed choice", ok, "missing choice")
		case ai.ScoreQuestion:
			a, ok := res.Answers[id].(ai.ScoreAnswer)
			if !cs.Check("typed visual score with legend", ok && len(a.Legend) == 2 && len(a.LegendImageReferences) == 2, "missing visual legend") {
				continue
			}
			assets := native["assets"].(map[string]any)
			levels := native["questions"].(map[string]any)[id].(map[string]any)["criteria"].([]any)
			for i, level := range levels {
				key := level.(map[string]any)["example"].(string)
				asset := assets[key].(map[string]any)
				raw, _ := base64.StdEncoding.DecodeString(asset["data_base64"].(string))
				idx := -1
				for n, img := range images {
					if img.SHA256 == sha256Text(raw) {
						idx = n
					}
				}
				cs.Check("public legend restores null", jsonEquivalent(a.Legend[i], []byte(`{"example":null}`)), "native key leaked")
				ref := a.LegendImageReferences[i]
				cs.Check("legend points to final image", ref.ImageIndex == idx && ref.Pointer == "/"+strconv.Itoa(i)+"/example" && ref.Placement == ai.ImagePlacementValue, "wrong legend reference")
			}
		case ai.BoxQuestion:
			a, ok := res.Answers[id].(ai.BoxAnswer)
			cs.Check("typed normalized Box with explicit confidence", ok && a.Box.Top >= 0 && a.Box.Bottom <= 1 && a.Box.Top < a.Box.Bottom && a.Box.Left >= 0 && a.Box.Right <= 1 && a.Box.Left < a.Box.Right && !a.Confidence.IsZero(), "invalid Box")
		}
	}
	// Mapping assertions use the public fixture as independent expected content.
	if strings.HasPrefix(sc.id, "default-context") || strings.HasPrefix(sc.id, "explicit-context") || strings.HasPrefix(sc.id, "ordered-reuse") {
		if route.name == "jev" {
			state := native["state"].(map[string]any)
			raw, _ := json.Marshal(state["state"])
			cs.Check("shared State preserved", jsonEquivalent(raw, sc.request.State), "State overwritten")
			refs := state["images"].([]any)
			if strings.HasPrefix(sc.id, "ordered-reuse") {
				cs.Check("ordered reuse survives projection", len(refs) == 3 && refs[0] == refs[2] && refs[0] != refs[1], "order/reuse lost")
			}
		}
	}
	if strings.HasPrefix(sc.id, "local-state-instructions") {
		state := native["state"].(map[string]any)
		qs := native["questions"].(map[string]any)
		cs.Check("local State and own instructions retain positions", state["photo"] != nil && qs["example"].(map[string]any)["instructions"].(map[string]any)["example"] != nil && qs["exists"].(map[string]any)["instructions"] == "Is a red rectangle present in the State photo?", "local scope changed")
	}
}
