package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Each finite boundary is exactly one byte above the selected allowance.
// Valid leading charge metadata survives output rejection. Input rejection
// precedes credentials, trusted payload encoding, and HTTP.
func imaginePressureBoundary(t *testing.T, ev *evidence.Case, kind string) {
	req := ai.ImagesRequest{Prompt: "pressure prompt byte boundary"}
	opts := imaginePressureOptions(10)
	count := 10
	if kind == "input-image" || kind == "input-total" {
		count = 1
		opts = imaginePressureOptions(1)
	}
	if kind == "output-image" {
		count = 1
		opts = imaginePressureOptions(1)
	}
	if kind == "input-image" {
		req.ReferenceImages = []ai.Image{imaginePressureImage(t, (512<<10)+1, 1)}
	}
	if kind == "input-total" {
		for i := range 5 {
			req.ReferenceImages = append(req.ReferenceImages, imaginePressureImage(t, 256<<10, 16+i))
		}
	}
	output := make([]ai.Image, count)
	for i := range output {
		output[i] = imaginePressureImage(t, 512<<10, 32+i)
	}
	ticks := int64(4000000000)
	if count == 1 {
		ticks = 400000000
	}
	reply := imaginePressureReply(output, ticks)
	var limit int64
	p := newImaginePressureWorld(t, func(c *ai.Config) {
		switch kind {
		case "input-image":
			limit = c.Policy.MaxImageBytes
		case "input-total":
			refs := make([]map[string]string, len(req.ReferenceImages))
			for i, img := range req.ReferenceImages {
				refs[i] = map[string]string{"type": "image_url", "url": "data:" + img.MimeType + ";base64," + img.Data}
			}
			native := map[string]any{"model": imagineModel, "prompt": req.Prompt, "response_format": "b64_json", "n": 1, "resolution": "1k", "quality": "low", "images": refs}
			limit = int64(len(mustMarshal(t, native))) - 1
			c.Policy.MaxRequestBytes = limit
		case "output-image":
			limit = (512 << 10) - 1
			c.Policy.Image.MaxOutputImageBytes = limit
		case "output-total":
			limit = (5 << 20) - 1
			c.Policy.Image.MaxTotalOutputImageBytes = limit
		case "output-body":
			limit = int64(len(reply.Script)) - 1
			c.Policy.MaxOutputBytes = limit
			c.Policy.Image.MaxTotalOutputImageBytes = 6 << 20
		}
	})
	inputFailure := kind == "input-image" || kind == "input-total"
	if !inputFailure {
		p.provider.EnqueueAt(imaginePressurePrefix(tenantA.tenant, false), reply)
	}
	encoded := 0
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	defer cancel()
	result, err := p.client.WithHooks(ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
		encoded++
		return ai.KeepPayload(), nil
	}}).GenerateImages(ctx, imaginePressureScope("pressure-boundary-"+kind, tenantA.tenant), imagineTarget(), req, opts)
	imaginePressureSummary(ev, "bounded-result", result, err)
	var failure *ai.Error
	ev.Check("one-byte overflow fails atomically", errors.As(err, &failure) && failure.Code == ai.CodeResourceLimit && len(result.Content) == 0, "err=%v", err)
	reads := p.host.ReadsFor(result.Metadata.RequestID).Credentials
	requests := p.provider.Requests()
	actual := int64(0)
	switch kind {
	case "input-image":
		actual = (512 << 10) + 1
	case "input-total":
		actual = limit + 1
	case "output-image":
		actual = 512 << 10
	case "output-total":
		actual = 5 << 20
	case "output-body":
		actual = int64(len(reply.Script))
	}
	ev.Check("boundary differs by exactly one byte", actual == limit+1, "actual=%d limit=%d", actual, limit)
	if inputFailure {
		ev.Check("known input overflow precedes credentials and encoding", reads == 0 && encoded == 0 && len(requests) == 0 && len(result.Metadata.Attempts) == 0, "reads=%d callbacks=%d requests=%d", reads, encoded, len(requests))
	} else {
		ev.Check("2xx overflow has exactly one attempt", reads == 1 && encoded == 1 && len(requests) == 1 && len(result.Metadata.Attempts) == 1, "reads=%d callbacks=%d requests=%d", reads, encoded, len(requests))
		if len(result.Metadata.Attempts) == 1 {
			a := result.Metadata.Attempts[0]
			amount, known := a.ReportedCostUSD.Get()
			ev.Check("known charge survives output limits", known && amount == float64(ticks)/1e10 && a.CostEstimate.Status != ai.CostResolved, "attempt=%+v", a)
		}
	}
	p.records.awaitCall(ev, result.Metadata.RequestID)
	ev.Record("byte-boundary", map[string]any{"kind": kind, "actualBytes": actual, "limitBytes": limit, "credentialReads": reads, "payloadCallbacks": encoded, "httpAttempts": len(requests), "realService": "NOT_RUN"})
	p.released(ev)
}
