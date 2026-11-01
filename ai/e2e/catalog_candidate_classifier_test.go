package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func catalogClassifierContract(t *testing.T, ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity, scene string) {
	if scene == "load-query" {
		catalogContractQuery(ev, cat, i)
		return
	}
	m, _ := cat.LookupClassifier(i.Provider, i.API, i.ID)
	w, b := newCatalogContractWorld(t, cat, i)
	b.AllowClassifierImages = true
	w.host.PutBinding(b)
	req := catalogClassifierInput(t, m)
	call := func(req ai.ClassifierRequest, input int64) {
		enqueue(ev, w, catalogClassifierReply(t, i, req, input))
		var opts ai.Options
		if i.API == ai.APITypeSafeSystemOne && len(req.Images) > 0 {
			opts = ai.TypeSafeOptions{ExperimentalImages: true}
		}
		result, err := w.client.Classify(ctxFor(t), textScope("catalog-classifier"), catalogContractTarget(i), req, opts)
		ev.Record("result", result)
		ev.Record("input", req)
		ev.Record("error", errString(err))
		ev.Check("declared classifier request completes", err == nil && len(result.Answers) == len(req.Questions), "%v %+v", err, result)
		if err != nil {
			return
		}
		catalogContractMetadata(ev, cat, i, result.Metadata)
		ev.Check("explicit classifier counters", result.Usage.Input == input && result.Usage.Output == 4 && result.Usage.TotalTokens == input+4, "%+v", result.Usage)
		rate := m.Cost.Input
		highest := int64(-1)
		for _, tier := range m.Cost.Tiers {
			if input > tier.InputTokensAbove && tier.InputTokensAbove >= highest {
				rate = tier.Input
				highest = tier.InputTokensAbove
			}
		}
		ev.Check("independent input-only classifier price", math.Abs(result.Usage.Cost.Total-float64(input)*rate/1e6) < 1e-12, "%+v", result.Usage)
		ev.Check("complete classifier usage", result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", result.Metadata.Attempts[0])
		for name, q := range req.Questions {
			answer := result.Answers[name]
			switch q.(type) {
			case ai.ChoiceQuestion:
				_, ok := answer.(ai.ChoiceAnswer)
				ev.Check("choice answer type", ok, "%T", answer)
			case ai.ScoreQuestion:
				_, ok := answer.(ai.ScoreAnswer)
				ev.Check("score answer type", ok, "%T", answer)
			case ai.BoolQuestion:
				_, ok := answer.(ai.BoolAnswer)
				ev.Check("bool answer type", ok, "%T", answer)
			case ai.BoxQuestion:
				_, ok := answer.(ai.BoxAnswer)
				ev.Check("box answer type", ok, "%T", answer)
			}
		}
		requests := w.provider.Requests()
		body := requests[len(requests)-1].Body
		var wire map[string]json.RawMessage
		mustUnmarshal(t, body, &wire)
		var model string
		_ = json.Unmarshal(wire["model"], &model)
		ev.Check("classifier wire identity", model == i.ID, "%s", body)
	}
	if scene == "authorization" {
		catalogContractDenied(ev, w, i, func(scope ai.CallScope, target ai.Target) error {
			_, err := w.client.Classify(ctxFor(t), scope, target, req, nil)
			return err
		})
		return
	}
	if scene == "questions" {
		call(req, 8)
	}
	if strings.HasPrefix(scene, "tier-") {
		var index int
		var boundary string
		_, err := fmt.Sscanf(scene, "tier-%d-%s", &index, &boundary)
		if err != nil {
			t.Fatal(err)
		}
		input := m.Cost.Tiers[index].InputTokensAbove
		switch boundary {
		case "below":
			input = max(0, input-1)
		case "above":
			input++
		}
		call(req, input)
	}
	if scene == "formats" {
		formats := m.Capabilities.DescriptionFormats
		if len(formats) == 0 {
			formats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString, ai.ClassifierDescriptionObject, ai.ClassifierDescriptionArray}
		}
		for _, format := range formats {
			desc := json.RawMessage(`"Catalog description"`)
			switch format {
			case ai.ClassifierDescriptionObject:
				desc = json.RawMessage(`{"text":"Catalog description"}`)
			case ai.ClassifierDescriptionArray:
				desc = json.RawMessage(`["Catalog description"]`)
			}
			r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog state"`), Questions: map[string]ai.ClassifierQuestion{"q": ai.ChoiceQuestion{Instructions: desc, Criteria: map[string]json.RawMessage{"a": desc, "b": desc}}}}
			call(r, 8)
		}
	}
	if scene == "image-locations" {
		catalogClassifierImages(t, ev, w, i, m, call)
	}
	if scene == "limits" {
		before := len(w.provider.Requests())
		for _, kind := range m.Capabilities.Kinds {
			switch kind {
			case ai.ClassifierQuestionChoice:
				r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog state"`), Questions: map[string]ai.ClassifierQuestion{}}
				for _, n := range []int{max(1, m.Capabilities.MinChoices), m.Capabilities.MaxChoices} {
					criteria := map[string]json.RawMessage{}
					for k := range n {
						criteria[strconv.Itoa(k)] = json.RawMessage(`"Choice"`)
					}
					r.Questions["q"] = ai.ChoiceQuestion{Instructions: json.RawMessage(`"Choose"`), Criteria: criteria}
					call(r, 8)
				}
				criteria := map[string]json.RawMessage{}
				for n := 0; n <= m.Capabilities.MaxChoices; n++ {
					criteria[strconv.Itoa(n)] = json.RawMessage(`"Choice"`)
				}
				r.Questions["q"] = ai.ChoiceQuestion{Instructions: json.RawMessage(`"Choose"`), Criteria: criteria}
				before = len(w.provider.Requests())
				_, err := w.client.Classify(ctxFor(t), textScope("catalog-choice-over"), catalogContractTarget(i), r, nil)
				ev.Check("choice overflow refused before send", err != nil && len(w.provider.Requests()) == before, "%v", err)
			case ai.ClassifierQuestionScore:
				for _, n := range []int{m.Capabilities.MinScoreLevels, m.Capabilities.MaxScoreLevels} {
					criteria := make([]json.RawMessage, n)
					for k := range criteria {
						criteria[k] = json.RawMessage(`"Level"`)
					}
					r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog state"`), Questions: map[string]ai.ClassifierQuestion{"q": ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate"`), Criteria: criteria}}}
					call(r, 8)
				}
			}
		}
	}
	ev.Record("requests", w.provider.Requests())
}
func catalogClassifierInput(t *testing.T, m ai.ClassifierModel) ai.ClassifierRequest {
	r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog state"`), Questions: map[string]ai.ClassifierQuestion{}}
	for _, kind := range m.Capabilities.Kinds {
		switch kind {
		case ai.ClassifierQuestionChoice:
			r.Questions["choice"] = ai.ChoiceQuestion{Instructions: json.RawMessage(`"Choose"`), Criteria: map[string]json.RawMessage{"a": json.RawMessage(`"A"`), "b": json.RawMessage(`"B"`)}}
		case ai.ClassifierQuestionScore:
			r.Questions["score"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate"`), Criteria: []json.RawMessage{json.RawMessage(`"Low"`), json.RawMessage(`"High"`)}}
		case ai.ClassifierQuestionBool:
			r.Questions["bool"] = ai.BoolQuestion{Instructions: json.RawMessage(`"Is this true?"`)}
		case ai.ClassifierQuestionBox:
			r.Questions["box"] = ai.BoxQuestion{Instructions: json.RawMessage(`"Locate object"`)}
			r.Images = jevContextImages(t)[:1]
		}
	}
	return r
}
func catalogClassifierReply(t *testing.T, i catalogContractIdentity, r ai.ClassifierRequest, input int64) provider.Reply {
	answers := map[string]any{}
	ordered := []any{}
	for name, q := range r.Questions {
		var a map[string]any
		switch q := q.(type) {
		case ai.ChoiceQuestion:
			keys := []string{}
			for key := range q.Criteria {
				keys = append(keys, key)
			}
			slices.Sort(keys)
			probabilities := map[string]float64{}
			rows := []any{}
			for n, key := range keys {
				p := 0.
				if n == 0 {
					p = 1
				}
				probabilities[key] = p
				rows = append(rows, map[string]any{"value": key, "probability": p})
			}
			a = map[string]any{"type": "choice", "choice": keys[0], "confidence": 1, "probabilities": probabilities}
			if i.API == ai.APIOpenAIDecisions {
				a["probabilities"] = rows
			}
		case ai.ScoreQuestion:
			probs := map[string]float64{}
			rows := []any{}
			for n, desc := range q.Criteria {
				p := 0.
				if n == 0 {
					p = 1
				}
				probs[strconv.Itoa(n)] = p
				var label string
				_ = json.Unmarshal(desc, &label)
				rows = append(rows, map[string]any{"value": n, "label": label, "probability": p})
			}
			a = map[string]any{"type": "score", "score": 0, "confidence": 1, "probabilities": probs}
			if i.API == ai.APIOpenAIDecisions {
				a["probabilities"] = rows
			}
		case ai.BoolQuestion:
			a = map[string]any{"type": "noul", "noul": 0.75}
			if i.API == ai.APIOpenAIDecisions {
				a = map[string]any{"type": "predicate", "probability": 0.75}
			}
		case ai.BoxQuestion:
			a = map[string]any{"type": "box", "box": map[string]float64{"top": 0.1, "bottom": 0.9, "left": 0.2, "right": 0.8}, "confidence": nil}
		}
		answers[name] = a
		if i.API == ai.APIOpenAIDecisions {
			a["name"] = name
			ordered = append(ordered, a)
		}
	}
	var data any = answers
	if i.API == ai.APIOpenAIDecisions {
		data = ordered
	}
	usage := map[string]any{"input_tokens": input, "output_tokens": 4}
	if i.API == ai.APIOpenAIDecisions {
		usage["total_tokens"] = input + 4
		usage["input_tokens_details"] = map[string]any{"cached_tokens": 0, "cache_write_tokens": 0}
		usage["output_tokens_details"] = map[string]any{"reasoning_tokens": 0}
	}
	body := mustMarshal(t, map[string]any{"model": i.ID, "answers": data, "usage": usage})
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Script: string(body), Chunks: [][]byte{body}}
}
