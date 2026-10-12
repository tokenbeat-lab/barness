package ai

import (
	"bytes"
	"encoding/json"
	"slices"
)

func classifierModelProblem(m ClassifierModel) string {
	if problem := modalitiesProblem(m.Input); problem != "" {
		return problem
	}
	if len(m.Input) > 0 && !slices.Contains(m.Input, ModalityText) {
		return "classifier input must include text"
	}
	c := m.Capabilities
	image := slices.Contains(m.Input, ModalityImage)
	if c.MaxInputImages < 0 || (image && (c.MaxInputImages == 0 || len(c.ImageMIMETypes) == 0)) || (!image && (c.MaxInputImages != 0 || len(c.ImageMIMETypes) > 0)) {
		return "classifier image declarations are inconsistent"
	}
	if image && (m.Provider != ProviderOpenAI || m.API != APIOpenAIDecisions || c.MaxInputImages > decisionsImageLimit) {
		return "classifier image declaration exceeds the implemented protocol"
	}
	seenMIME := map[string]bool{}
	for _, mime := range c.ImageMIMETypes {
		if !classifierImageMIME(mime) || seenMIME[mime] {
			return "classifier image MIME is unknown or duplicated"
		}
		seenMIME[mime] = true
	}
	if c.MinChoices < 0 || c.MinChoices > c.MaxChoices || (c.MinChoices != 0 && !slices.Contains(c.Kinds, ClassifierQuestionChoice)) {
		return "classifier minimum choices do not match its question kinds or maximum"
	}
	seen := map[ClassifierDescriptionFormat]bool{}
	for _, f := range c.DescriptionFormats {
		if (f != ClassifierDescriptionString && f != ClassifierDescriptionObject && f != ClassifierDescriptionArray) || seen[f] {
			return "classifier description format is unknown or duplicated"
		}
		seen[f] = true
	}
	if m.Provider == ProviderOpenAI && m.API == APIOpenAIDecisions && (c.MinChoices < 2 || !slices.Equal(c.DescriptionFormats, []ClassifierDescriptionFormat{ClassifierDescriptionString})) {
		return "Decisions requires at least two choices and string descriptions"
	}
	return ""
}

func classifierQuestionFormats(q ClassifierQuestion, formats []ClassifierDescriptionFormat) bool {
	if len(formats) == 0 {
		return true
	}
	allowed := func(raw json.RawMessage, nullable bool) bool {
		raw = bytes.TrimSpace(raw)
		if nullable && bytes.Equal(raw, []byte("null")) {
			return true
		}
		if len(raw) == 0 {
			return false
		}
		var format ClassifierDescriptionFormat
		switch raw[0] {
		case '"':
			format = ClassifierDescriptionString
		case '{':
			format = ClassifierDescriptionObject
		case '[':
			format = ClassifierDescriptionArray
		}
		return slices.Contains(formats, format)
	}
	switch q := q.(type) {
	case ChoiceQuestion:
		if !allowed(q.Instructions, false) {
			return false
		}
		for _, raw := range q.Criteria {
			if !allowed(raw, true) {
				return false
			}
		}
	case ScoreQuestion:
		if !allowed(q.Instructions, false) {
			return false
		}
		for _, raw := range q.Criteria {
			if !allowed(raw, false) {
				return false
			}
		}
	case BoolQuestion:
		if !allowed(q.Instructions, false) {
			return false
		}
		if q.Criteria != nil && (!allowed(q.Criteria.True, false) || !allowed(q.Criteria.False, false)) {
			return false
		}
	default:
		return false
	}
	return true
}

func decisionsTextCapability(req ClassifierRequest) string {
	for _, q := range req.Questions {
		if choice, ok := q.(ChoiceQuestion); ok && len(choice.Criteria) < 2 {
			return "Decisions requires at least two choices"
		}
		if !classifierQuestionFormats(q, []ClassifierDescriptionFormat{ClassifierDescriptionString}) {
			return "Decisions requires at least two choices and string descriptions"
		}
	}
	return ""
}
