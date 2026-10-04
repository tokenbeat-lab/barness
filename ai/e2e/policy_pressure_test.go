package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// pressureEnv opts in to TestPolicyPressure: a design load streams millions
// of frames, too heavy for every `go test ./...`. The release gate sets it
// (go run ./ai/release/cmd/releasegate); without it the cases are NOT_RUN,
// which fails the gate's P0 check.
const pressureEnv = "BARNESS_AI_PRESSURE"

// pressureDeadline bounds one pressure run; a design load that cannot
// finish in it is a finding, not a slow test.
const pressureDeadline = 3 * time.Minute

// maxDesignRatio is the largest share of a byte or queue limit the design
// load may use: vendors' frame envelopes vary (obfuscation padding,
// logprobs, reasoning summaries), so a policy whose design load nearly
// fills a limit would refuse ordinary turns.
const maxDesignRatio = 0.75

// streamedTokensPerSecond is the generation speed the backlog estimate
// assumes: a fast model streaming to a client that stopped reading.
const streamedTokensPerSecond = 100

// TestPolicyPressure runs each formal policy example (spec I9 "正式示例",
// ADR-0002, issue 24) under the load its comments say it is sized for: every
// permit of the process busy with a design-size turn (long history with
// images in, the largest stated output or tool call out) while consumers
// keep up, and then a consumer that stops reading. The design load must
// complete under the policy, its memory must stay within the policy's stated
// budget, and a stalled consumer must end with resource_limit instead of
// holding memory. Each limit's headroom (design value ÷ limit) and the
// measured heap are kept as policy-pressure.json, the numerical basis the
// policies' comments cite. Admission values are measured by
// TestAdmissionBurstPressure.
//
// Opt-in: BARNESS_AI_PRESSURE=1.
//
// Responses carries the run: its terminal frames repeat the whole response,
// so it is the protocol that comes closest to MaxFrameBytes and
// MaxOutputBytes for a given output.
func TestPolicyPressure(t *testing.T) {
	policies := []struct {
		name   string
		policy *ai.ResourcePolicy
		load   designLoad
		// heapBudget is the memory the policy's comments assume for
		// barness-ai's buffers.
		heapBudget uint64
	}{
		{"local", localassembly.LocalPolicy(), designLoad{
			Basis:   "LocalPolicy: one session plus parallel sub-agents (all 8 permits), long history (4 MiB) with three screenshots, a 128K-token turn, a tool writing a 512 KiB file",
			Tenants: 1, Calls: 8, OutputTokens: 128 << 10, HistoryBytes: 4 << 20, Images: 3, ImageBytes: 2 << 20, ToolJSONBytes: 512 << 10,
		}, 1 << 30},
		{"cloud-interactive", hostintegration.CloudInteractivePolicy(), designLoad{
			Basis:   "CloudInteractivePolicy: 8 tenants at their limit of 4 (all 32 permits), agent histories (2 MiB) with two downscaled images, the 64K-token upper bound of an interactive turn or a 256 KiB tool call",
			Tenants: 8, Calls: 32, OutputTokens: 64 << 10, HistoryBytes: 2 << 20, Images: 2, ImageBytes: 1 << 20, ToolJSONBytes: 256 << 10,
		}, 2 << 30},
		{"cloud-batch", hostintegration.CloudBatchPolicy(), designLoad{
			Basis:   "CloudBatchPolicy: 2 tenants at their limit of 16 (all 32 permits), agent histories (2 MiB) with two images, long 128K-token generations or 256 KiB tool calls",
			Tenants: 2, Calls: 32, OutputTokens: 128 << 10, HistoryBytes: 2 << 20, Images: 2, ImageBytes: 1 << 20, ToolJSONBytes: 256 << 10,
		}, 2 << 30},
	}
	for _, p := range policies {
		t.Run(p.name+"/design-load", func(t *testing.T) {
			ev := pressureCase(t, "E08-policy-pressure-"+p.name+"-design-load")
			r := runDesignLoad(t, ev, *p.policy, p.load, p.heapBudget)
			ev.Record("policy-pressure", r)
			t.Logf("%s: %+v", p.name, r)
		})
		t.Run(p.name+"/stalled-consumer", func(t *testing.T) {
			ev := pressureCase(t, "E08-policy-pressure-"+p.name+"-stalled-consumer")
			r := runStalledConsumer(t, ev, *p.policy, p.load)
			ev.Record("policy-pressure", r)
			t.Logf("%s stalled: %+v", p.name, r)
		})
	}
}

// pressureCase starts a pressure case, NOT_RUN unless opted in.
func pressureCase(t *testing.T, id string) *evidence.Case {
	ev := run.Case(t, id)
	ev.ReplayEnv(pressureEnv + "=1")
	if os.Getenv(pressureEnv) != "1" {
		t.Skip("opt-in: " + pressureEnv + "=1")
	}
	return ev
}

// headroom is one limit's design value against the policy's limit.
type headroom struct {
	Design int64   `json:"design"`
	Limit  int64   `json:"limit"`
	Ratio  float64 `json:"ratio"`
}

func room(design, limit int64) headroom {
	return headroom{Design: design, Limit: limit, Ratio: float64(design) / float64(limit)}
}

// designReport is a design-load run's outcome.
type designReport struct {
	Load     designLoad          `json:"load"`
	Headroom map[string]headroom `json:"headroom"`
	// PeakPermits is the most attempts admitted at once.
	PeakPermits int64 `json:"peakPermits"`
	// HeapBudget is what the policy assumes for barness-ai's buffers. The
	// heap is measured over the whole process, the local provider double
	// (which keeps every request it received, ProviderCaptured bytes) and
	// the scripted frames included, so it overstates barness-ai's share.
	HeapBudget       uint64  `json:"heapBudget"`
	Heap             heapUse `json:"heap"`
	ProviderCaptured int64   `json:"providerCapturedRequestBytes"`
	Duration         string  `json:"duration"`
	// TokensPerSecond is the aggregate rate the process consumed streamed
	// tokens at: how far barness-ai's own processing is from limiting a
	// vendor's generation speed.
	TokensPerSecond int `json:"processedTokensPerSecond"`
}

func runDesignLoad(t *testing.T, ev *evidence.Case, policy ai.ResourcePolicy, l designLoad, heapBudget uint64) designReport {
	t.Helper()
	pw := newPressureWorld(t, policy, l.Tenants)
	textFrames, toolFrames := pressureText(t, l.OutputTokens), pressureToolCall(t, l.ToolJSONBytes)
	textChunks, toolChunks := coalesce(textFrames, 64), coalesce(toolFrames, 64)
	isTool := func(i int) bool { return i%4 == 0 }
	// Every reply waits until every call's request has arrived, so all
	// permits are held at once however fast a turn streams: a tool call
	// streams in milliseconds and would otherwise end before the last call
	// is admitted. The wait is bounded so a lost request fails the checks
	// below instead of hanging the run.
	var arrived sync.WaitGroup
	arrived.Add(l.Calls)
	allArrived := doneWhen(arrived.Wait)
	barrier := func() {
		arrived.Done()
		select {
		case <-allArrived:
		case <-time.After(pressureDeadline):
		}
	}
	for i := range l.Calls {
		reply := pressureReply(textChunks)
		if isTool(i) {
			reply = pressureReply(toolChunks)
		}
		reply.OnReceive = barrier
		pw.provider.Enqueue(reply)
	}
	req := pressureRequest(l)
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	defer cancel()

	// A run that times out fails at once: canceling ends every stream, and
	// the sampler stops, before the test returns.
	heap := startHeapSampler(pw.probe.Permits)
	defer heap.finish()
	started := time.Now()
	type ended struct {
		res    ai.Result
		err    error
		events int
	}
	results := make([]ended, l.Calls)
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	for i := range l.Calls {
		k := pw.tenants[i%l.Tenants]
		s := pw.client.Stream(ctx, scopeFor(k, fmt.Sprintf("req-pressure-%d", i)), ai.Target{BindingID: "primary", ModelID: pressureModel}, req, nil)
		wg.Go(func() {
			defer s.Close()
			n := 0
			for s.Next() {
				n++
			}
			res, err := s.Result()
			results[i] = ended{res, err, n}
		})
	}
	waitForWithin(ev, "every design-load call ends", pressureDeadline, doneWhen(wg.Wait))
	elapsed := time.Since(started)
	use := heap.finish()
	peakPermits := heap.peakGauge.Load()

	texts, tools := 0, 0
	for i, e := range results {
		m := e.res.Message
		if !ev.Check("design-load call completes", e.err == nil && m.ErrorMessage == "", "call %d: err %v, message %q", i, e.err, m.ErrorMessage) {
			continue
		}
		for _, c := range m.Content {
			switch c := c.(type) {
			case ai.Text:
				ev.Check("the whole design turn arrived", len(c.Text) == l.OutputTokens*len(pressureToken), "call %d: %d bytes", i, len(c.Text))
				texts++
			case ai.ToolCall:
				ev.Check("the whole tool call arrived", len(c.RawArguments) == l.ToolJSONBytes, "call %d: %d bytes", i, len(c.RawArguments))
				tools++
			}
		}
	}
	ev.Check("every call delivered its turn", texts+tools == l.Calls, "%d text, %d tool of %d", texts, tools, l.Calls)
	ev.Check("the process ran at its concurrency limit", peakPermits == int64(policy.MaxConcurrentProcess), "peak %d of %d", peakPermits, policy.MaxConcurrentProcess)

	var largestRequest, captured int64
	for _, r := range pw.provider.Requests() {
		largestRequest = max(largestRequest, int64(len(r.Body)))
		captured += int64(len(r.Body))
	}
	textLargest, textTotal := largestAndTotal(textFrames)
	toolLargest, toolTotal := largestAndTotal(toolFrames)
	tokens := (l.Calls-tools)*l.OutputTokens + tools*l.ToolJSONBytes/4
	rep := designReport{Load: l, PeakPermits: peakPermits, HeapBudget: heapBudget, Heap: use,
		ProviderCaptured: captured, Duration: elapsed.Round(time.Millisecond).String(), TokensPerSecond: int(float64(tokens) / elapsed.Seconds()),
		Headroom: map[string]headroom{
			"MaxRequestBytes":  room(largestRequest, policy.MaxRequestBytes),
			"MaxImageBytes":    room(int64(l.ImageBytes), policy.MaxImageBytes),
			"MaxFrameBytes":    room(int64(max(textLargest, toolLargest)), policy.MaxFrameBytes),
			"MaxOutputBytes":   room(int64(max(textTotal, toolTotal)), policy.MaxOutputBytes),
			"MaxToolJSONBytes": room(int64(l.ToolJSONBytes), policy.MaxToolJSONBytes),
			// The probe's queue gauge is process-wide: every stream's
			// queue together, against every stream's limit together.
			"MaxQueuedEvents (all streams)": room(pw.probe.PeakQueuedEvents(), int64(l.Calls*policy.MaxQueuedEvents)),
			"MaxConcurrentPerTenant":        room(int64(l.Calls/l.Tenants), int64(policy.MaxConcurrentPerTenant)),
			"MaxConcurrentProcess":          room(peakPermits, int64(policy.MaxConcurrentProcess)),
		}}
	for name, h := range rep.Headroom {
		if strings.HasPrefix(name, "MaxConcurrent") {
			// The design load runs every permit by construction.
			ev.Check("the design load uses every permit of "+name, h.Ratio == 1, "%+v", h)
			continue
		}
		ev.Check("the design load leaves headroom under "+name, h.Ratio <= maxDesignRatio, "%+v", h)
	}
	ev.Check("the reachable heap stays within the policy's budget", use.LiveGrowth() <= heapBudget, "grew %d MiB, budget %d MiB", use.LiveGrowth()>>20, heapBudget>>20)
	checkPressureReleased(ev, pw)
	return rep
}

// stalledReport is a stalled-consumer run's outcome.
type stalledReport struct {
	OutputTokens int    `json:"outputTokens"`
	Code         string `json:"code"`
	Phase        string `json:"phase"`
	// PeakQueuedEvents is the most events held unread; BacklogMinutes is
	// how long a client that stopped reading has before MaxQueuedEvents,
	// assuming the model streams AssumedTokensPerSecond.
	PeakQueuedEvents       int64   `json:"peakQueuedEvents"`
	MaxQueuedEvents        int     `json:"maxQueuedEvents"`
	BacklogMinutes         float64 `json:"backlogMinutes"`
	AssumedTokensPerSecond int     `json:"assumedTokensPerSecond"`
	// TextKeptBytes is the text the failed result still holds.
	TextKeptBytes int `json:"textKeptBytes"`
}

func runStalledConsumer(t *testing.T, ev *evidence.Case, policy ai.ResourcePolicy, l designLoad) stalledReport {
	t.Helper()
	pw := newPressureWorld(t, policy, 1)
	pw.provider.Enqueue(pressureReply(coalesce(pressureText(t, l.OutputTokens), 64)))
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	defer cancel()
	// The consumer never calls Next: only the Result is awaited.
	s := pw.client.Stream(ctx, scopeFor(pw.tenants[0], "req-pressure-stalled"), ai.Target{BindingID: "primary", ModelID: pressureModel},
		ai.Request{Messages: []ai.Message{ai.UserText("Write a long report.")}}, nil)
	defer s.Close()
	ended := resultWithin(ev, "the stalled stream ends", s)
	res, err := ended.res, ended.err
	var ae *ai.Error
	rep := stalledReport{OutputTokens: l.OutputTokens, PeakQueuedEvents: pw.probe.PeakQueuedEvents(), MaxQueuedEvents: policy.MaxQueuedEvents,
		BacklogMinutes: float64(policy.MaxQueuedEvents) / streamedTokensPerSecond / 60, AssumedTokensPerSecond: streamedTokensPerSecond}
	if errors.As(err, &ae) {
		rep.Code, rep.Phase = string(ae.Code), string(ae.Phase)
	}
	for _, c := range res.Message.Content {
		if text, ok := c.(ai.Text); ok {
			rep.TextKeptBytes += len(text.Text)
		}
	}
	ev.Check("a stalled consumer ends the call with resource_limit in the event queue", rep.Code == string(ai.CodeResourceLimit) && rep.Phase == string(ai.PhaseEventQueue), "err %v", err)
	// The terminal's place is reserved beyond the limit (ADR-0007).
	ev.Check("the queue held no more than its limit and the terminal", rep.PeakQueuedEvents <= int64(policy.MaxQueuedEvents)+1, "peak %d", rep.PeakQueuedEvents)
	ev.Check("the result keeps what was received", rep.TextKeptBytes > 0, "no text kept")
	ev.Check("a stalled client has minutes before the limit", rep.BacklogMinutes >= 1, "%.1f minutes", rep.BacklogMinutes)
	checkPressureReleased(ev, pw)
	return rep
}

func checkPressureReleased(ev *evidence.Case, pw pressureWorld) {
	if errs := pw.failures.list(); len(errs) > 0 {
		ev.Record("transport-errors", errs)
	}
	eventually(ev, "every call ended", func() bool { return pw.probe.ActiveCalls() == 0 })
	ev.Check("no permit is held", pw.probe.Permits() == 0, "%d held", pw.probe.Permits())
	ev.Check("every response body is closed", pw.bodies.Open() == 0, "%d open", pw.bodies.Open())
}

// waitForWithin is waitFor with a scenario's own deadline.
func waitForWithin(ev *evidence.Case, what string, d time.Duration, ch <-chan struct{}) {
	select {
	case <-ch:
		ev.Check(what, true, "")
	case <-time.After(d):
		ev.Check(what, false, "not within %s", d)
		ev.T().FailNow()
	}
}
