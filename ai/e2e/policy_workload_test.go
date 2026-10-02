package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/metrics"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// designLoad is the load a formal policy example states it is sized for
// (spec I9 "正式示例", ADR-0002): the pressure scenario runs exactly this load
// under the policy and records how close each value comes to its limit.
type designLoad struct {
	// Basis says where the numbers come from: the policy's own comments.
	Basis string `json:"basis"`
	// Tenants share the process; Calls run at once (the policy's process
	// concurrency), spread evenly over the tenants.
	Tenants int `json:"tenants"`
	Calls   int `json:"calls"`
	// OutputTokens is one turn's streamed output, one token (4 characters)
	// per delta frame as the vendors stream it.
	OutputTokens int `json:"outputTokens"`
	// HistoryBytes of prior text and Images × ImageBytes (image bytes, sent
	// as base64) make up the request.
	HistoryBytes int `json:"historyBytes"`
	Images       int `json:"images"`
	ImageBytes   int `json:"imageBytes"`
	// ToolJSONBytes is the argument JSON of the tool call every fourth call
	// makes instead of a text turn.
	ToolJSONBytes int `json:"toolJsonBytes"`
}

// pressureToken is one streamed token: about 4 characters, as the vendors'
// tokenizers average on English text.
const pressureToken = "tok "

// pressureModel takes images and 128K output tokens.
const pressureModel = "gpt-5"

// pressureTenant is the i-th synthetic pressure tenant.
func pressureTenant(i int) tenantKey {
	return tenantKey{tenant: fmt.Sprintf("tenant-p%02d", i), secret: fmt.Sprintf("sk-test-pressure-%02d-0001", i), alias: fmt.Sprintf("key:tenant-p%02d@v1", i)}
}

// pressureWorld is n tenants with the "primary" Responses binding under
// policy, with body tracking and a resource probe.
type pressureWorld struct {
	provider *provider.Server
	client   *ai.Client
	probe    *probe.Probe
	bodies   *provider.BodyTracker
	tenants  []tenantKey
}

func newPressureWorld(t *testing.T, policy ai.ResourcePolicy, n int) pressureWorld {
	t.Helper()
	aliases := map[string]string{}
	h := host.New()
	srv := provider.New(aliases)
	t.Cleanup(srv.Close)
	var tenants []tenantKey
	for i := range n {
		k := pressureTenant(i + 1)
		aliases[k.secret] = k.alias
		run.RedactSecret(k.secret, k.alias)
		h.PutBinding(primaryBinding(k, srv.URL()))
		h.PutCredential(primaryCredential(k, "v1"))
		tenants = append(tenants, k)
	}
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	p := probe.New()
	client, err := ai.NewClient(ai.Config{Policy: &policy, Bindings: h, Credentials: h, Transport: bodies, Probe: p, AllowLoopbackHTTP: true})
	if err != nil {
		t.Fatalf("NewClient with the example policy: %v", err)
	}
	return pressureWorld{provider: srv, client: client, probe: p, bodies: bodies, tenants: tenants}
}

// pressureRequest is a design request: HistoryBytes of prior text in 16 KiB
// user turns, then a user turn with the images and a question, and the
// write_file tool.
func pressureRequest(l designLoad) ai.Request {
	var msgs []ai.Message
	const turn = 16 << 10
	for left := l.HistoryBytes; left > 0; left -= turn {
		msgs = append(msgs, ai.UserText(strings.Repeat("h", min(turn, left))))
	}
	last := ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: "Describe these screenshots and write the summary to a file."}}}
	img := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x89, 'P', 'N', 'G'}, l.ImageBytes/4))
	for range l.Images {
		last.Content = append(last.Content, ai.Image{Data: img, MimeType: "image/png"})
	}
	msgs = append(msgs, last)
	return ai.Request{SystemPrompt: "You are a coding agent.", Messages: msgs,
		Tools: []ai.Tool{{Name: "write_file", Description: "Write a file.", Parameters: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}`)}}}
}

// sequencedEvents returns an encoder that numbers each Responses event in
// order, as the vendor's sequence_number does.
func sequencedEvents(t *testing.T) func(map[string]any) json.RawMessage {
	seq := 0
	return func(v map[string]any) json.RawMessage {
		v["sequence_number"] = seq
		seq++
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
}

// pressureText is a design text turn as OpenAI streams it: every delta frame
// carries the envelope fields the vendor sends (logprobs, obfuscation), and
// the terminal frames repeat the whole text, as Responses does.
func pressureText(t *testing.T, tokens int) [][]byte {
	t.Helper()
	ev := sequencedEvents(t)
	text := strings.Repeat(pressureToken, tokens)
	part := map[string]any{"type": "output_text", "text": text, "annotations": []any{}}
	item := map[string]any{"id": "msg_p", "type": "message", "status": "completed", "role": "assistant", "content": []any{part}}
	events := []json.RawMessage{
		ev(map[string]any{"type": "response.created", "response": pressureResponse("in_progress", nil, 0)}),
		ev(map[string]any{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"id": "msg_p", "type": "message", "status": "in_progress", "role": "assistant", "content": []any{}}}),
		ev(map[string]any{"type": "response.content_part.added", "item_id": "msg_p", "output_index": 0, "content_index": 0, "part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}}}),
	}
	for range tokens {
		events = append(events, ev(map[string]any{"type": "response.output_text.delta", "item_id": "msg_p", "output_index": 0, "content_index": 0,
			"delta": pressureToken, "logprobs": []any{}, "obfuscation": "Xk3fQ9"}))
	}
	events = append(events,
		ev(map[string]any{"type": "response.output_text.done", "item_id": "msg_p", "output_index": 0, "content_index": 0, "text": text, "logprobs": []any{}}),
		ev(map[string]any{"type": "response.content_part.done", "item_id": "msg_p", "output_index": 0, "content_index": 0, "part": part}),
		ev(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item}),
		ev(map[string]any{"type": "response.completed", "response": pressureResponse("completed", []any{item}, tokens)}),
	)
	return lfChunks(t, events)
}

// pressureToolCall is a design tool call turn: the arguments stream in
// 64-character deltas and are repeated in the done and terminal frames.
func pressureToolCall(t *testing.T, argBytes int) [][]byte {
	t.Helper()
	ev := sequencedEvents(t)
	prefix := `{"path":"notes/summary.md","content":"`
	args := prefix + strings.Repeat("c", max(0, argBytes-len(prefix)-2)) + `"}`
	item := map[string]any{"id": "fc_p", "type": "function_call", "status": "completed", "call_id": "call_p", "name": "write_file", "arguments": args}
	events := []json.RawMessage{
		ev(map[string]any{"type": "response.created", "response": pressureResponse("in_progress", nil, 0)}),
		ev(map[string]any{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"id": "fc_p", "type": "function_call", "status": "in_progress", "call_id": "call_p", "name": "write_file", "arguments": ""}}),
	}
	for i := 0; i < len(args); i += 64 {
		events = append(events, ev(map[string]any{"type": "response.function_call_arguments.delta", "item_id": "fc_p", "output_index": 0, "delta": args[i:min(i+64, len(args))], "obfuscation": "Xk3fQ9"}))
	}
	events = append(events,
		ev(map[string]any{"type": "response.function_call_arguments.done", "item_id": "fc_p", "output_index": 0, "arguments": args}),
		ev(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item}),
		ev(map[string]any{"type": "response.completed", "response": pressureResponse("completed", []any{item}, len(args)/4)}),
	)
	return lfChunks(t, events)
}

// pressureResponse is a response object; outputTokens > 0 adds its usage, as
// the terminal response carries.
func pressureResponse(status string, output []any, outputTokens int) map[string]any {
	r := map[string]any{"id": "resp_p", "object": "response", "created_at": 1790000000, "status": status, "model": "gpt-5-2025-08-07", "output": []any{}}
	if output != nil {
		r["output"] = output
	}
	if outputTokens > 0 {
		r["usage"] = map[string]any{"input_tokens": 1000, "input_tokens_details": map[string]any{"cached_tokens": 0},
			"output_tokens": outputTokens, "output_tokens_details": map[string]any{"reasoning_tokens": 0}, "total_tokens": 1000 + outputTokens}
	}
	return r
}

// pressureReply is a 200 event stream of chunks without provider.SSE's
// evidence copy of the whole body: a design turn's script is tens of MB per
// call, which would dominate the heap being measured. The evidence records
// the generator's parameters (designLoad) instead.
func pressureReply(chunks [][]byte) provider.Reply {
	return provider.Reply{Status: http.StatusOK, Header: map[string]string{"Content-Type": "text/event-stream"}, Chunks: chunks}
}

// coalesce groups frames into writes of n frames each, as TCP coalesces a
// fast upstream's frames; framing within a write is unchanged.
func coalesce(frames [][]byte, n int) [][]byte {
	var out [][]byte
	for i := 0; i < len(frames); i += n {
		out = append(out, bytes.Join(frames[i:min(i+n, len(frames))], nil))
	}
	return out
}

// heapSampler records the peak live heap (runtime/metrics
// /gc/heap/live:bytes: what the last GC found reachable), the peak
// HeapInuse, which also counts garbage not yet collected and so depends on
// GOGC, and the peak of a gauge such as the permits held.
type heapSampler struct {
	baselineLive, baselineInuse uint64
	peakLive, peakInuse         atomic.Uint64
	peakGauge                   atomic.Int64
	stop, done                  chan struct{}
	stopOnce                    sync.Once
}

var liveHeapMetric = []metrics.Sample{{Name: "/gc/heap/live:bytes"}}

func liveHeap() uint64 {
	s := slices.Clone(liveHeapMetric)
	metrics.Read(s)
	return s[0].Value.Uint64()
}

func startHeapSampler(gauge func() int64) *heapSampler {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	s := &heapSampler{baselineLive: liveHeap(), baselineInuse: m.HeapInuse, stop: make(chan struct{}), done: make(chan struct{})}
	s.peakLive.Store(s.baselineLive)
	s.peakInuse.Store(m.HeapInuse)
	go func() {
		defer close(s.done)
		// The gauge is read every millisecond; the heap, whose reading
		// stops the world, every tenth.
		tick := time.NewTicker(time.Millisecond)
		defer tick.Stop()
		for n := 0; ; n++ {
			select {
			case <-tick.C:
				s.peakGauge.Store(max(s.peakGauge.Load(), gauge()))
				if n%10 != 0 {
					continue
				}
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				s.peakInuse.Store(max(s.peakInuse.Load(), m.HeapInuse))
				s.peakLive.Store(max(s.peakLive.Load(), liveHeap()))
			case <-s.stop:
				return
			}
		}
	}()
	return s
}

// heapUse is a run's heap measurements, in bytes.
type heapUse struct {
	BaselineLive  uint64 `json:"baselineLive"`
	PeakLive      uint64 `json:"peakLive"`
	BaselineInuse uint64 `json:"baselineInuse"`
	PeakInuse     uint64 `json:"peakInuse"`
}

// LiveGrowth is how much the reachable heap grew over the baseline.
func (h heapUse) LiveGrowth() uint64 { return h.PeakLive - min(h.PeakLive, h.BaselineLive) }

// finish stops the sampler; it may be called more than once.
func (s *heapSampler) finish() heapUse {
	s.stopOnce.Do(func() { close(s.stop) })
	<-s.done
	return heapUse{BaselineLive: s.baselineLive, PeakLive: s.peakLive.Load(), BaselineInuse: s.baselineInuse, PeakInuse: s.peakInuse.Load()}
}
