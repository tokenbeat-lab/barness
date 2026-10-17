package ai

import "github.com/openai/openai-go/v3/responses"

// reasoningSignature is pi's thinkingSignature for a Responses reasoning item:
// JSON.stringify of the item as received, so key order, escapes and number
// spellings are JavaScript's and same-model replay sends exactly what pi
// sends.
func reasoningSignature(raw string) string {
	item, err := parseJSON(raw)
	if err != nil {
		return raw // the SDK decoded it, so it is JSON; kept as is otherwise
	}
	return stringifyJSON(item)
}

// backfillReasoning ports pi's backfillReasoningSignatures: some services
// omit a reasoning item's encrypted_content from output_item.done and send it
// only in the terminal response's output. It is added to the finished block's
// signature unless that already holds a truthy one, in place when the key is
// present and at the end otherwise, so stateless (store:false) replay keeps
// the encrypted reasoning.
func (p *responsesParser) backfillReasoning(output []responses.ResponseOutputItemUnion) {
	for _, item := range output {
		if item.Type != "reasoning" {
			continue
		}
		encrypted, err := parseJSON(item.JSON.EncryptedContent.Raw())
		i, ok := p.reasoning[item.ID]
		if err != nil || !jsTruthy(encrypted) || !ok {
			continue
		}
		if sig, ok := withEncryptedContent(p.out.thinkingSignature(i), encrypted); ok {
			p.out.setThinkingSignature(i, sig)
		}
	}
}

// withEncryptedContent is pi's {...storedItem, encrypted_content}: the stored
// reasoning item with encrypted set, in place when the key exists. ok is
// false when the signature is empty, not an object, or already holds a
// truthy encrypted_content.
func withEncryptedContent(signature string, encrypted any) (string, bool) {
	stored, err := parseJSON(signature)
	obj, isObject := stored.(*jsonObject)
	if signature == "" || err != nil || !isObject {
		return "", false
	}
	if current, _ := obj.get("encrypted_content"); jsTruthy(current) {
		return "", false
	}
	obj.set("encrypted_content", encrypted)
	return stringifyJSON(obj), true
}

// encodeTextSignature produces pi-ai's TextSignatureV1 JSON: {"v":1,"id":…[,"phase":…]}.
func encodeTextSignature(id, phase string) string {
	sig := struct {
		V     int    `json:"v"`
		ID    string `json:"id"`
		Phase string `json:"phase,omitempty"`
	}{V: 1, ID: id, Phase: phase}
	data, _ := marshalJS(sig)
	return string(data)
}
