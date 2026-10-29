package ai

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
)

// Currency is separate from token completeness. A valid amount survives an
// invalid token report; an invalid amount never erases valid token counters.
func xaiImagesUsage(raw json.RawMessage) (Usage, UsageReporting, Nullable[float64], *Error) {
	var amount Nullable[float64]
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Usage{}, UsageUnreported, amount, nil
	}
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil {
		return Usage{}, UsageUnreported, amount, xaiImagesProtocolFailure()
	}
	var failure *Error
	if duplicate {
		failure = xaiImagesProtocolFailure()
	}
	if value := bytes.TrimSpace(fields["cost_in_usd_ticks"]); len(value) > 0 && !bytes.Equal(value, []byte("null")) {
		ticks, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil || ticks < 0 {
			failure = xaiImagesProtocolFailure()
		} else {
			amount = Value(float64(ticks) / 1e10)
		}
	}
	counters := map[string]xaiCount{}
	known := false
	count := func(name string, raw json.RawMessage) {
		n, ok := xaiUsageCount(raw)
		if !ok {
			failure = xaiImagesProtocolFailure()
			return
		}
		counters[name] = n
		known = known || n.present
	}
	for _, key := range []string{"input_tokens", "output_tokens", "total_tokens"} {
		count(key, fields[key])
	}
	for _, detail := range []struct {
		key  string
		keys []string
	}{{"input_tokens_details", []string{"cached_tokens", "text_tokens", "image_tokens"}}, {"output_tokens_details", []string{"reasoning_tokens", "text_tokens", "image_tokens"}}} {
		raw := bytes.TrimSpace(fields[detail.key])
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			continue
		}
		parts, dup, err := unaryJSONEnvelope(raw)
		if err != nil {
			failure = xaiImagesProtocolFailure()
			continue
		}
		if dup {
			failure = xaiImagesProtocolFailure()
		}
		for _, key := range detail.keys {
			count(detail.key+"."+key, parts[key])
		}
	}
	if !known {
		return Usage{}, UsageUnreported, amount, failure
	}
	i, o, total, cache, reasoning := counters["input_tokens"], counters["output_tokens"], counters["total_tokens"], counters["input_tokens_details.cached_tokens"], counters["output_tokens_details.reasoning_tokens"]
	u := Usage{Input: i.n, Output: o.n, TotalTokens: total.n, CacheRead: cache.n}
	if reasoning.present {
		u.Reasoning = Value(reasoning.n)
	}
	if i.present && cache.n > i.n || o.present && reasoning.n > o.n {
		failure = xaiImagesProtocolFailure()
	}
	if i.present && o.present {
		if i.n > math.MaxInt64-o.n {
			failure = xaiImagesProtocolFailure()
		} else if total.present && total.n != i.n+o.n {
			failure = xaiImagesProtocolFailure()
		} else if !total.present {
			u.TotalTokens = i.n + o.n
		}
	}
	if total.present && (i.n > total.n || o.n > total.n || cache.n > total.n || reasoning.n > total.n) {
		failure = xaiImagesProtocolFailure()
	}
	for _, detail := range []struct {
		prefix string
		total  xaiCount
		reason xaiCount
	}{{"input_tokens_details.", i, xaiCount{}}, {"output_tokens_details.", o, reasoning}} {
		a, b := counters[detail.prefix+"text_tokens"], counters[detail.prefix+"image_tokens"]
		if a.n > math.MaxInt64-b.n || a.n+b.n > math.MaxInt64-detail.reason.n {
			failure = xaiImagesProtocolFailure()
			continue
		}
		sum := a.n + b.n + detail.reason.n
		if detail.total.present && (sum > detail.total.n || (a.present && b.present && (detail.reason.present || detail.prefix == "input_tokens_details.") && sum != detail.total.n)) {
			failure = xaiImagesProtocolFailure()
		}
		if detail.prefix == "input_tokens_details." && a.present && cache.n > a.n {
			failure = xaiImagesProtocolFailure()
		}
	}
	for _, field := range []struct {
		key string
		dst func(*ModalityUsage) *Nullable[int64]
	}{
		{"input_tokens_details.text_tokens", func(m *ModalityUsage) *Nullable[int64] { return &m.InputText }},
		{"input_tokens_details.image_tokens", func(m *ModalityUsage) *Nullable[int64] { return &m.InputImage }},
		{"output_tokens_details.text_tokens", func(m *ModalityUsage) *Nullable[int64] { return &m.OutputText }},
		{"output_tokens_details.image_tokens", func(m *ModalityUsage) *Nullable[int64] { return &m.OutputImage }},
	} {
		if c := counters[field.key]; c.present {
			if u.Modalities == nil {
				u.Modalities = &ModalityUsage{}
			}
			*field.dst(u.Modalities) = Value(c.n)
		}
	}
	reporting := UsagePartial
	if i.present && o.present && total.present && cache.present && reasoning.present {
		reporting = UsageComplete
	}
	return u, reporting, amount, failure
}
func xaiImagesProtocolFailure() *Error {
	return newError(CodeProtocol, PhaseResponse, "invalid xAI Images response")
}
