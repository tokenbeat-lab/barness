package e2e

import "github.com/tokenbeat-lab/barness/ai"

// deepseekResponsesSuite is DeepSeek through Responses (P05) to the byte
// limit suite: only the request body DeepSeek's capabilities shape, and the
// SSE frame and streamed output of its replies, differ from OpenAI's
// Responses calls; the tool JSON and error body bounds are the shared
// adapter's, asserted on P01 by TestByteLimits.
var deepseekResponsesSuite = &protocolSuite{
	proto: deepseekResponsesProtocol, prefix: "P05", options: ai.ResponsesOptions{}, sharedAdapter: true,
}

// deepseekChatSuite is DeepSeek through Chat Completions (P06) to the byte
// limit suite: the same bounds on DeepSeek's Chat calls, whose request body
// DeepSeek's compat shapes; the tool JSON and error body bounds are the
// shared Chat adapter's, asserted on P04.
var deepseekChatSuite = &protocolSuite{
	proto: deepseekChatProtocol, prefix: "P06", options: ai.ChatOptions{}, sharedAdapter: true,
}
