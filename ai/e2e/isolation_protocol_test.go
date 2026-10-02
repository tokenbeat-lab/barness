package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// What each protocol of testdata/isolation/interleave.json is: the binding
// that selects it, and the key, account, endpoint path, reply framing,
// provider, API and case prefix that follow from it. The isolation,
// envelope and host/local example suites share these.

func (p isolationProtocol) anthropic() bool { return p.BindingID == "claude" }

func (p isolationProtocol) gemini() bool { return p.BindingID == "gemini" }

func (p isolationProtocol) chat() bool { return p.BindingID == "chat" }

// deepseek is DeepSeek on the Responses protocol: the Responses adapter
// with DeepSeek's binding, key, account, endpoint and capabilities.
func (p isolationProtocol) deepseek() bool { return p.BindingID == "deepseek" }

// deepseekChat is DeepSeek on Chat Completions: the Chat adapter with
// DeepSeek's binding and compat, on the same account and key as deepseek.
func (p isolationProtocol) deepseekChat() bool { return p.BindingID == "deepseek-chat" }

// protocolTimeout reports whether p's protocol has a request timeout of its
// own (timeoutMs); pi's Google path has none.
func (p isolationProtocol) protocolTimeout() bool { return !p.gemini() }

func (p isolationProtocol) target() ai.Target {
	return ai.Target{BindingID: p.BindingID, ModelID: p.Model}
}

func (p isolationProtocol) request() ai.Request {
	return ai.Request{SystemPrompt: p.SystemPrompt, Messages: []ai.Message{ai.UserText(p.User)}}
}

// key is the key k's tenant holds for p's binding.
func (p isolationProtocol) key(k tenantKey) tenantKey {
	switch {
	case p.anthropic():
		return anthropicKey(k)
	case p.gemini():
		return googleKey(k)
	case p.deepseek(), p.deepseekChat():
		return deepseekKey(k)
	}
	return k
}

// account is the vendor account of k's tenant's binding for p.
func (p isolationProtocol) account(k tenantKey) string {
	switch {
	case p.anthropic():
		return "acct-" + k.tenant + "-anthropic"
	case p.gemini():
		return "acct-" + k.tenant + "-google"
	case p.deepseek(), p.deepseekChat():
		return "acct-" + k.tenant + "-deepseek"
	}
	return "acct-" + k.tenant
}

// path is where k's tenant's endpoint receives p's inference requests.
func (p isolationProtocol) path(k tenantKey) string {
	switch {
	case p.anthropic():
		return tenantPrefix(k) + "/v1/messages"
	case p.gemini():
		return tenantPrefix(k) + "/v1beta/models/" + p.Model + ":streamGenerateContent"
	case p.chat():
		return tenantPrefix(k) + "/v1/chat/completions"
	case p.deepseek():
		return tenantPrefix(k) + "/responses"
	case p.deepseekChat():
		return tenantPrefix(k) + "/chat/completions"
	}
	return tenantPrefix(k) + "/v1/responses"
}

func (p isolationProtocol) script(k tenantKey) isolationScript { return p.Tenants[k.tenant] }

// success is k's tenant's successful reply, one SSE frame per chunk.
func (p isolationProtocol) success(t *testing.T, k tenantKey) provider.Reply {
	t.Helper()
	switch {
	case p.gemini():
		return dataEvents(t, p.script(k).Events, provider.FramingLF)
	case p.chat(), p.deepseekChat():
		return chatEvents(t, p.script(k).Events, provider.FramingLF)
	}
	return sseEvents(t, p.script(k).Events, provider.FramingLF)
}

// firstMarkedChunk is the index of the first chunk of k's success reply
// carrying k's text marker: once it is written, k's call holds content of
// its own in flight.
func (p isolationProtocol) firstMarkedChunk(t *testing.T, f isolationFixture, k tenantKey) int {
	t.Helper()
	marker := f.Markers[k.tenant][0]
	for i, c := range p.success(t, k).Chunks {
		if strings.Contains(string(c), marker) {
			return i
		}
	}
	t.Fatalf("%s: no chunk of %s carries %q", p.ID, k.tenant, marker)
	return 0
}

// protocolCase is p's protocol case prefix: P01 Responses, P02 Anthropic,
// P03 Gemini, P04 Chat Completions, P05 DeepSeek Responses, P06 DeepSeek
// Chat Completions.
func protocolCase(p isolationProtocol) string {
	switch {
	case p.anthropic():
		return "P02"
	case p.gemini():
		return "P03"
	case p.chat():
		return "P04"
	case p.deepseek():
		return "P05"
	case p.deepseekChat():
		return "P06"
	}
	return "P01"
}

func (p isolationProtocol) provider() ai.ProviderID {
	switch {
	case p.anthropic():
		return ai.ProviderAnthropic
	case p.gemini():
		return ai.ProviderGoogle
	case p.deepseek(), p.deepseekChat():
		return ai.ProviderDeepSeek
	}
	return ai.ProviderOpenAI
}

func (p isolationProtocol) api() ai.API {
	switch {
	case p.anthropic():
		return ai.APIAnthropicMessages
	case p.gemini():
		return ai.APIGoogleGenerativeAI
	case p.chat(), p.deepseekChat():
		return ai.APIOpenAICompletions
	}
	return ai.APIOpenAIResponses
}
