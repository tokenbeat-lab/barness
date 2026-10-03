package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// protocolSuite describes one scenario-fixture protocol to the shared
// resource, admission, callback and observability suites
// (protocol_{limits,admission,callbacks,observe}_test.go): the data its
// cases embed — reply scripts, error bodies, generated long outputs, refused
// models — and, as explicit fields, where the protocol deliberately differs
// from the others (ADR-0012, ADR-0013). Every suite runs the same cases on
// each protocol under the protocol's case ID prefix.
type protocolSuite struct {
	proto *fixtureProtocol
	// prefix is the protocol's case ID prefix, e.g. "P02".
	prefix string
	// options is the protocol's empty full options.
	options ai.Options
	// sharedAdapter says another protocol's suite covers the tool JSON,
	// error body, CRLF framing and held-connection bounds of the adapter
	// this one shares (DeepSeek's: P01 Responses, P04 Chat); only the
	// request body, frame and output bounds, which this protocol's
	// capabilities and replies shape, run on it, and no other suite does,
	// so it sets only proto, prefix and options.
	sharedAdapter bool

	// toolArgs is text.json "interleaved"'s tool call argument JSON as the
	// protocol streams it.
	toolArgs string
	// limited is the failures.json scenario whose 429 body is the limited
	// error body.
	limited string
	held    heldScripts
	// longOutput generates a stream of one text block in n deltas of 9
	// bytes each, and the deltas.
	longOutput func(t *testing.T, n int) (provider.Reply, []string)

	// timeoutOptions is the full options carrying the protocol's
	// timeoutMs; nil when the protocol has none, so timeoutMs bounds
	// nothing (pi's Google path).
	timeoutOptions func(ms int) ai.Options
	// headerTimeoutRetried says whether a response header timeout is
	// retried under the binding's retry policy; pi's Google path never
	// retries a request that got no response.
	headerTimeoutRetried bool
	// overloaded is a retried overload reply.
	overloaded provider.Reply
	// deniedModel is a model the binding does not allow.
	deniedModel string
	// credential is k's credential snapshot for the binding.
	credential func(k tenantKey) ai.Credential

	// keyHeader is the header carrying the key, and keyValue its value for
	// k's key.
	keyHeader string
	keyValue  func(k tenantKey) string
	callbacks callbackData

	// authFailure is a 401/403 reply and the vendor request id it carries
	// ("" when the vendor sends none).
	authFailure          provider.Reply
	authFailureRequestID string
	redaction            redactionData
	// foreignOptions are another API's options, which the binding refuses.
	foreignOptions ai.Options
	// rejections are the protocol's preflight refusals beyond the shared
	// binding and credential ones: models, options.
	rejections []rejectCase
	env        environmentData
}

// protocolSuites are the protocols every shared suite runs on.
var protocolSuites = []*protocolSuite{anthropicSuite, geminiSuite, chatSuite}

// forEachSuite runs suite on each of suites as a subtest named by the
// protocol's fixture directory.
func forEachSuite(t *testing.T, suites []*protocolSuite, suite func(*testing.T, *protocolSuite)) {
	for _, s := range suites {
		t.Run(s.proto.dir, func(t *testing.T) { suite(t, s) })
	}
}

// entries are the four public entry points on the protocol's binding.
func (s *protocolSuite) entries() []outcomeEntry { return scenarioEntries(s.options) }

// caseID is the evidence case ID "<P>-<rest>".
func (s *protocolSuite) caseID(rest string) string { return s.prefix + "-" + rest }

// requestID is the call's request ID "req-<p>-<rest>".
func (s *protocolSuite) requestID(rest string) string {
	return "req-" + strings.ToLower(s.prefix) + "-" + rest
}

// heldScripts are the pieces of a reply that exceeds a limit and then holds
// the connection without finishing.
type heldScripts struct {
	// start is complete frames opening the stream.
	start string
	// frameOpen opens a data frame the reply never closes.
	frameOpen string
	// filler is frames carrying no output (pings, comments).
	filler string
	// errorOpen opens an error JSON body the reply never closes.
	errorOpen string
}

// callbackData is a protocol's trusted callback cases.
type callbackData struct {
	// onResponse says whether the response callback runs; pi's Google path
	// never calls onResponse.
	onResponse bool
	replaced   replacedPayload
	// changedInPlace is nil when the protocol's in-place change case is
	// its own, run by more (Anthropic's payload-sees-betas).
	changedInPlace *payloadChange
	// widen are the payload changes a callback cannot make.
	widen           []payloadWidening
	hostedTool      hostedTool
	headerTransform headerTransform
	// more runs the protocol's own callback cases.
	more func(t *testing.T, s *protocolSuite)
}

// headerTransform adds to the shared header transform case: extra runs in
// the transform and check on the request sent; either may be nil.
type headerTransform struct {
	extra func(http.Header)
	check func(*evidence.Case, provider.Request)
}

// replacedPayload is a payload callback's replacement and what is sent.
type replacedPayload struct {
	// name is the case name, which says whether stream is forced back on.
	name string
	body func(p *ai.Payload) map[string]any
	// sent is the body sent, described by about.
	sent, about string
	// check is nil or checks the request further.
	check func(*evidence.Case, provider.Request)
}

// payloadChange is an in-place payload change and how its sent body shows.
type payloadChange struct {
	change func(body map[string]any)
	sent   func(body map[string]any) bool
	about  string
}

type payloadWidening struct {
	name   string
	change func(map[string]any)
}

// hostedTool is a hosted tool the binding allows, as a callback adds it.
type hostedTool struct {
	name  string
	add   func(body map[string]any)
	sent  func(body map[string]any) bool
	about string
}

// redactionData is the protocol's redaction case data.
type redactionData struct {
	// kept is text the authorized Result keeps (signatures, reasoning).
	kept []string
	// echo is an auth error body echoing secret and the prompt, and its
	// status.
	echo func(secret string) (status int, body string)
	// sensitive must reach no observation or log; notInError no error
	// text. The key's secret is in both.
	sensitive, notInError []string
}

// rejectCase is one preflight refusal: entries nil runs the four entries.
type rejectCase struct {
	id      string
	code    ai.Code
	phase   ai.Phase
	target  ai.Target
	arrange func(w *world)
	entries []outcomeEntry
}

// environmentData is what the protocol's SDKs read from the environment
// beyond polluteEnvironment's, and the headers it must not turn into.
type environmentData struct {
	vars   map[string]string
	leaked []string
}

// fullEntries are the stream-full and complete-full entries, each with its
// own full options.
func fullEntries(stream, complete ai.Options) []outcomeEntry {
	return []outcomeEntry{
		{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.Stream(ctx, scope, target, req, stream))
		}},
		{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.Complete(ctx, scope, target, req, complete)
			return outcome{result: res, err: err}
		}},
	}
}
