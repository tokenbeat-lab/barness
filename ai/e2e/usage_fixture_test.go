package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// usageFixture is testdata/responses/usage.json: E11 scenarios run against
// text-basic.json's logical input. Each successful reply is the fixture's
// prefix (one "ok" text block) followed by the scenario's terminal event.
type usageFixture struct {
	// Catalog pins the built-in catalog's version and content hash.
	Catalog struct {
		Version string `json:"version"`
		Hash    string `json:"hash"`
	} `json:"catalog"`
	Model     string            `json:"model"`
	Prefix    []json.RawMessage `json:"prefix"`
	Scenarios []usageScenario   `json:"scenarios"`
}

type usageScenario struct {
	ID string `json:"id"`
	// Model overrides the fixture's model; ModelPatch replaces top-level
	// model fields on both sides, as a host catalog does.
	Model      string          `json:"model"`
	ModelPatch json.RawMessage `json:"modelPatch"`
	// Options are full-entry options in pi's names; absent means none.
	Options *struct {
		ServiceTier string `json:"serviceTier"`
	} `json:"options"`
	Retry   *retrySettings `json:"retry"`
	Replies []usageReply   `json:"replies"`
	Expect  usageExpect    `json:"expect"`
}

// usageReply is one scripted attempt: the prefix and a terminal event, the
// prefix cut by End, or a plain HTTP reply.
type usageReply struct {
	Terminal json.RawMessage   `json:"terminal"`
	End      provider.End      `json:"end"`
	Status   int               `json:"status"`
	Header   map[string]string `json:"header"`
	Body     string            `json:"body"`
}

type usageExpect struct {
	StopReason ai.StopReason `json:"stopReason"`
	Code       ai.Code       `json:"code"`
	// Usage is the message's pi-compatible usage, cost included.
	Usage    ai.Usage       `json:"usage"`
	Attempts []usageAttempt `json:"attempts"`
}

type usageAttempt struct {
	HTTPStatus     int               `json:"httpStatus"`
	Code           ai.Code           `json:"code"`
	UsageReporting ai.UsageReporting `json:"usageReporting"`
	Usage          ai.Usage          `json:"usage"`
}

func loadUsageFixture(t *testing.T) (usageFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f usageFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("usage.json: %v", err)
	}
	return f, raw
}

func (f usageFixture) model(sc usageScenario) string {
	if sc.Model != "" {
		return sc.Model
	}
	return f.Model
}

func (f usageFixture) replies(t *testing.T, sc usageScenario) []provider.Reply {
	t.Helper()
	out := make([]provider.Reply, 0, len(sc.Replies))
	for _, r := range sc.Replies {
		switch {
		case r.Status != 0:
			header := map[string]string{"Content-Type": "application/json"}
			for k, v := range r.Header {
				header[k] = v
			}
			out = append(out, provider.Reply{Status: r.Status, Header: header, Chunks: [][]byte{[]byte(r.Body)}, Script: r.Body})
		case r.Terminal != nil:
			out = append(out, sseEvents(t, append(append([]json.RawMessage(nil), f.Prefix...), r.Terminal), provider.FramingLF))
		default:
			reply := sseEvents(t, f.Prefix, provider.FramingLF)
			reply.End = r.End
			out = append(out, reply)
		}
	}
	return out
}

// options are sc's full-entry options; nil when it has none.
func (sc usageScenario) options() ai.Options {
	if sc.Options == nil {
		return nil
	}
	return ai.ResponsesOptions{ServiceTier: ai.Value(sc.Options.ServiceTier)}
}

// piOptions are sc's options and retry settings in pi's shape.
func (sc usageScenario) piOptions() map[string]any {
	opts := sc.Retry.piOptions()
	if sc.Options != nil {
		if opts == nil {
			opts = map[string]any{}
		}
		opts["serviceTier"] = sc.Options.ServiceTier
	}
	return opts
}

// usagePidiffScenario is E11 scenario sc of TestUsageAndCost on the full
// entry, with the same model patch, options and retry settings on both sides.
func usagePidiffScenario(t *testing.T, f usageFixture, raw []byte, text textFixture, sc usageScenario) pidiffScenario {
	t.Helper()
	return pidiffScenario{fixture: "usage.json", raw: raw, model: f.model(sc), req: textRequest(text),
		replies: f.replies(t, sc), requests: len(sc.Replies), retry: sc.Retry.policy(),
		full: sc.options(), piOptions: sc.piOptions(), modelPatch: sc.ModelPatch}
}
