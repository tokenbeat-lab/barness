package ai

import (
	"encoding/json"
	"math"
)

func googleImagesUsage(raw json.RawMessage, pricing ImagePricing) (UsageReporting, Usage, *Error) {
	if len(raw) == 0 || string(raw) == "null" {
		return UsageUnreported, Usage{}, nil
	}
	u := Usage{Modalities: &ModalityUsage{}}
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil || duplicate {
		return UsagePartial, u, googleImageProtocolFailure()
	}
	complete := true
	count := func(raw json.RawMessage) (Nullable[int64], bool) {
		if len(raw) == 0 || string(raw) == "null" {
			complete = false
			return Nullable[int64]{}, true
		}
		var n int64
		if json.Unmarshal(raw, &n) != nil || n < 0 {
			return Nullable[int64]{}, false
		}
		return Value(n), true
	}
	totals := map[string]Nullable[int64]{}
	for _, key := range []string{"total_input_tokens", "total_output_tokens", "total_thought_tokens", "total_tokens"} {
		n, ok := count(fields[key])
		if !ok {
			return UsagePartial, u, googleImageProtocolFailure()
		}
		totals[key] = n
		switch key {
		case "total_input_tokens":
			u.Input, _ = n.Get()
		case "total_output_tokens":
			u.Output, _ = n.Get()
		case "total_thought_tokens":
			u.Reasoning = n
		case "total_tokens":
			u.TotalTokens, _ = n.Get()
		}
	}
	thought, hasThought := u.Reasoning.Get()
	if u.Output > math.MaxInt64-thought {
		return UsagePartial, u, googleImageProtocolFailure()
	}
	u.Output += thought
	if _, ok := totals["total_tokens"].Get(); !ok {
		if u.Input > math.MaxInt64-u.Output {
			return UsagePartial, u, googleImageProtocolFailure()
		}
		u.TotalTokens = u.Input + u.Output
	}
	// Cache totals are retained for diagnostics, but never guessed into a
	// text/image price split. Input remains Google's reported total input.
	if raw := fields["total_cached_tokens"]; len(raw) > 0 && string(raw) != "null" {
		n, ok := count(raw)
		if !ok {
			return UsagePartial, u, googleImageProtocolFailure()
		}
		u.CacheRead, _ = n.Get()
	}
	for _, part := range []struct {
		key         string
		text, image *Nullable[int64]
	}{
		{"input_tokens_by_modality", &u.Modalities.InputText, &u.Modalities.InputImage},
		{"output_tokens_by_modality", &u.Modalities.OutputText, &u.Modalities.OutputImage},
	} {
		raw := fields[part.key]
		if len(raw) == 0 || string(raw) == "null" {
			complete = false
			continue
		}
		var entries []json.RawMessage
		if raw[0] != '[' || json.Unmarshal(raw, &entries) != nil {
			return UsagePartial, u, googleImageProtocolFailure()
		}
		seen := map[string]bool{}
		for _, entry := range entries {
			record, duplicate, err := unaryJSONEnvelope(entry)
			if err != nil || duplicate {
				return UsagePartial, u, googleImageProtocolFailure()
			}
			modality, ok := rawString(record["modality"])
			if !ok || seen[modality] || (modality != "text" && modality != "image") {
				return UsagePartial, u, googleImageProtocolFailure()
			}
			seen[modality] = true
			n, ok := count(record["tokens"])
			if !ok {
				return UsagePartial, u, googleImageProtocolFailure()
			}
			if modality == "text" {
				*part.text = n
			} else {
				*part.image = n
			}
		}
		if !seen["text"] || !seen["image"] {
			complete = false
		}
	}
	priced := *u.Modalities
	text, hasText := priced.OutputText.Get()
	image, hasImage := priced.OutputImage.Get()
	output, hasOutput := totals["total_output_tokens"].Get()
	extraThought := thought
	if hasText && hasImage && hasOutput && hasThought {
		if text > math.MaxInt64-image {
			return UsagePartial, u, googleImageProtocolFailure()
		}
		sum := text + image
		// The two source-backed fixture variants establish whether the modality
		// breakdown excludes or includes thinking. Preserve raw modality counts;
		// only the pricing view adds excluded thought at the text output rate.
		switch sum {
		case output:
		case u.Output:
			extraThought = 0
		default:
			return UsagePartial, u, googleImageProtocolFailure()
		}
	} else if hasText && text != 0 && thought != 0 {
		// Missing totals/details cannot establish inclusion. Price only reported
		// text, rather than adding thinking twice, and keep reporting partial.
		extraThought = 0
		complete = false
	}
	if text > math.MaxInt64-extraThought {
		return UsagePartial, u, googleImageProtocolFailure()
	}
	if hasText || extraThought > 0 {
		priced.OutputText = Value(text + extraThought)
	}
	u.Cost = pricing.estimate(priced)
	for _, cost := range []float64{u.Cost.Input, u.Cost.Output, u.Cost.Total} {
		if math.IsNaN(cost) || math.IsInf(cost, 0) {
			u.Cost = UsageCost{}
			return UsagePartial, u, googleImageProtocolFailure()
		}
	}
	if complete {
		return UsageComplete, u, nil
	}
	return UsagePartial, u, nil
}
