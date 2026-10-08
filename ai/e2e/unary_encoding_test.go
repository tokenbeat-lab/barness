package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type UnaryHiddenOne struct {
	Hidden string
}

type unaryEncodedByte uint8

func (*unaryEncodedByte) MarshalJSON() ([]byte, error) { return []byte("0"), nil }

func TestUnaryCallbackByteElementEncoding(t *testing.T) {
	ev := run.Case(t, "P07-E04-unary-callback-custom-byte-element")
	u := newUnaryWorld(t, func(c *ai.Config) { c.Policy.Classifier.MaxStateBytes = 3 })
	enqueue(ev, u.world, u.sc.replies(t)[0])
	req := classifierInput(t, u.sc)
	req.State = json.RawMessage("[]")
	h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		p.Body["state"] = []unaryEncodedByte{255}
		return ai.KeepPayload(), nil
	}}
	res, err := u.client.WithHooks(h).Classify(ctxFor(t), textScope("req-custom-byte-element"), u.sc.target(), req, nil)
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Record("requests", u.provider.Requests())
	ev.Check("exact-cap custom byte array accepted", err == nil && len(res.Answers) == 1 && len(u.provider.Requests()) == 1, "got %v", err)
	if requests := u.provider.Requests(); len(requests) == 1 {
		var body map[string]json.RawMessage
		mustUnmarshal(t, requests[0].Body, &body)
		ev.Check("byte element marshaler selected", string(body["state"]) == "[0]", "got %s", body["state"])
	}
	u.released(ev)
}

type UnaryHiddenTwo struct {
	Hidden string
}
type unaryShortText string

func (unaryShortText) MarshalText() ([]byte, error) { return []byte("short"), nil }

type unaryTextKey struct{ Source string }

func (unaryTextKey) MarshalText() ([]byte, error) { return []byte("short"), nil }

type unaryPointerText string

func (*unaryPointerText) MarshalText() ([]byte, error) { return []byte("short"), nil }

type unaryPointerJSON string

func (*unaryPointerJSON) MarshalJSON() ([]byte, error) { return []byte(`"short"`), nil }

type unaryQuestionJSON map[string]any

func (unaryQuestionJSON) MarshalJSON() ([]byte, error) {
	return []byte(`{"intent":{"type":"choice","instructions":"Choose intent","criteria":{"refund":"Refund","track":"Tracking"}}}`), nil
}

func TestUnaryCallbackEncodingSemantics(t *testing.T) {
	large := strings.Repeat("x", 8<<20)
	for _, name := range []string{"promoted-conflict", "text-value", "text-key", "questions-marshaler", "pointer-text-element", "pointer-json-element", "pointer-array-element"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E04-unary-callback-encoding-"+name)
			u := newUnaryWorld(t, nil)
			enqueue(ev, u.world, u.sc.replies(t)[0])
			res, err := u.call(t, ctxFor(t), "req-encoding-"+name, ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "pointer-text-element":
					p.Body["state"] = []unaryPointerText{unaryPointerText(large)}
				case "pointer-json-element":
					p.Body["state"] = []unaryPointerJSON{unaryPointerJSON(large)}
				case "pointer-array-element":
					p.Body["state"] = &[1]unaryPointerJSON{unaryPointerJSON(large)}
				case "promoted-conflict":
					p.Body["state"] = struct {
						UnaryHiddenOne
						UnaryHiddenTwo
						Visible string `json:"visible"`
					}{UnaryHiddenOne{large}, UnaryHiddenTwo{large}, "ok"}
				case "text-value":
					p.Body["state"] = unaryShortText(large)
				case "text-key":
					p.Body["state"] = map[unaryTextKey]string{{Source: large}: "ok"}
				case "questions-marshaler":
					qs := unaryQuestionJSON{}
					for _, key := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
						qs[key] = large
					}
					p.Body["questions"] = qs
				}
				return ai.KeepPayload(), nil
			}})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", u.provider.Requests())
			ev.Check("final encoded small JSON remains valid", err == nil && len(res.Answers) == 1 && len(u.provider.Requests()) == 1, "got %v", err)
			if reqs := u.provider.Requests(); len(reqs) > 0 {
				ev.Check("large omitted source never sent", len(reqs[0].Body) < 4096, "got %d bytes", len(reqs[0].Body))
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
