package ai

import (
	"encoding/json"
	"math"
)

// ModalityUsage preserves unreported versus explicitly zero token counts.
// Only image protocols set Usage.Modalities; chat serialization is unchanged.
type ModalityUsage struct {
	InputText   Nullable[int64] `json:"inputText,omitzero"`
	InputImage  Nullable[int64] `json:"inputImage,omitzero"`
	OutputText  Nullable[int64] `json:"outputText,omitzero"`
	OutputImage Nullable[int64] `json:"outputImage,omitzero"`
}

func (u Usage) clone() Usage {
	if u.Modalities != nil {
		v := *u.Modalities
		u.Modalities = &v
	}
	return u
}
func cloneAttempts(attempts []Attempt) []Attempt {
	if attempts == nil {
		return nil
	}
	out := make([]Attempt, len(attempts))
	for i, a := range attempts {
		a.Usage = a.Usage.clone()
		out[i] = a
	}
	return out
}

func (p ImagePricing) estimate(m ModalityUsage) UsageCost {
	product := func(rate Nullable[float64], tokens Nullable[int64]) float64 {
		r, _ := rate.Get()
		n, _ := tokens.Get()
		// Round each product independently before adding, like chat/pi cost.
		return float64(float64(r/1e6) * float64(n))
	}
	c := UsageCost{Input: product(p.InputText, m.InputText) + product(p.InputImage, m.InputImage), Output: product(p.OutputText, m.OutputText) + product(p.OutputImage, m.OutputImage)}
	c.Total = c.Input + c.Output
	return c
}

func openAIImagesUsage(raw json.RawMessage, pricing ImagePricing) (UsageReporting, Usage, *Error) {
	if len(raw) == 0 || string(raw) == "null" {
		return UsageUnreported, Usage{}, nil
	}
	u := Usage{Modalities: &ModalityUsage{}}
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil || duplicate {
		return UsagePartial, u, imageProtocolFailure()
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
	for _, v := range []struct {
		key string
		dst *int64
	}{{"input_tokens", &u.Input}, {"output_tokens", &u.Output}, {"total_tokens", &u.TotalTokens}} {
		n, ok := count(fields[v.key])
		if !ok {
			return UsagePartial, u, imageProtocolFailure()
		}
		*v.dst, _ = n.Get()
	}
	if raw := fields["total_tokens"]; len(raw) == 0 || string(raw) == "null" {
		if u.Input > math.MaxInt64-u.Output {
			return UsagePartial, u, imageProtocolFailure()
		}
		u.TotalTokens = u.Input + u.Output
	}
	for _, v := range []struct {
		key         string
		text, image *Nullable[int64]
	}{
		{"input_tokens_details", &u.Modalities.InputText, &u.Modalities.InputImage},
		{"output_tokens_details", &u.Modalities.OutputText, &u.Modalities.OutputImage},
	} {
		details := fields[v.key]
		if len(details) == 0 || string(details) == "null" {
			complete = false
			continue
		}
		parts, duplicate, err := unaryJSONEnvelope(details)
		if err != nil || duplicate {
			return UsagePartial, u, imageProtocolFailure()
		}
		var ok bool
		*v.text, ok = count(parts["text_tokens"])
		if !ok {
			return UsagePartial, u, imageProtocolFailure()
		}
		*v.image, ok = count(parts["image_tokens"])
		if !ok {
			return UsagePartial, u, imageProtocolFailure()
		}
	}
	u.Cost = pricing.estimate(*u.Modalities)
	for _, c := range []float64{u.Cost.Input, u.Cost.Output, u.Cost.Total} {
		if math.IsNaN(c) || math.IsInf(c, 0) {
			u.Cost = UsageCost{} // Keep reported counts, never publish non-JSON finite costs.
			return UsagePartial, u, imageProtocolFailure()
		}
	}
	if complete {
		return UsageComplete, u, nil
	}
	return UsagePartial, u, nil
}
