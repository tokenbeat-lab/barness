package ai

import (
	"bytes"
	"encoding/json"
)

type kimiUsageObject map[string]json.RawMessage

func kimiObject(raw json.RawMessage) (kimiUsageObject, error) {
	var out kimiUsageObject
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	if raw[0] != '{' || json.Unmarshal(raw, &out) != nil {
		return nil, kimiUsageFailure()
	}
	return out, nil
}

func decodeKimiUsage(api API, raw json.RawMessage) (kimiUsageReport, error) {
	var r kimiUsageReport
	object, err := kimiObject(raw)
	if err != nil {
		return r, err
	}
	// Each schema projects once at its protocol boundary. Unknown fields do
	// not become token counters and cannot affect pricing.
	var detailName, outputDetailName, inputName, outputName, readName, writeName, reasonName string
	switch api {
	case APIOpenAICompletions:
		detailName, outputDetailName = "prompt_tokens_details", "completion_tokens_details"
		inputName, outputName, readName, writeName, reasonName = "prompt_tokens", "completion_tokens", "cached_tokens", "cache_write_tokens", "reasoning_tokens"
	case APIOpenAIResponses:
		detailName, outputDetailName = "input_tokens_details", "output_tokens_details"
		inputName, outputName, readName, writeName, reasonName = "input_tokens", "output_tokens", "cached_tokens", "cache_write_tokens", "reasoning_tokens"
	case APIAnthropicMessages:
		detailName, outputDetailName = "cache_creation", "output_tokens_details"
		inputName, outputName, readName, writeName, reasonName = "input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens", "thinking_tokens"
	default:
		return r, kimiUsageFailure()
	}
	details, err := kimiObject(object[detailName])
	if err != nil {
		return r, err
	}
	outputDetails, err := kimiObject(object[outputDetailName])
	if err != nil {
		return r, err
	}
	readRaw, writeRaw := details[readName], details[writeName]
	if api == APIOpenAICompletions && (len(readRaw) == 0 || bytes.Equal(bytes.TrimSpace(readRaw), []byte("null"))) {
		readRaw = object[readName]
	}
	if api == APIAnthropicMessages {
		readRaw, writeRaw = object[readName], object[writeName]
	}
	fields := []struct {
		raw    json.RawMessage
		target *kimiCount
	}{
		{object[inputName], &r.input}, {object[outputName], &r.output}, {readRaw, &r.read}, {writeRaw, &r.write},
		{object["total_tokens"], &r.total}, {outputDetails[reasonName], &r.reasoning},
	}
	if api == APIAnthropicMessages {
		fields = append(fields, struct {
			raw    json.RawMessage
			target *kimiCount
		}{details["ephemeral_5m_input_tokens"], &r.short}, struct {
			raw    json.RawMessage
			target *kimiCount
		}{details["ephemeral_1h_input_tokens"], &r.long})
	}
	for _, field := range fields {
		count, err := parseKimiCount(field.raw)
		if err != nil {
			return r, err
		}
		*field.target = count
	}
	return r, nil
}
