package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

type imaginePressureByteDigest struct {
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

func imaginePressureDigest(raw []byte) imaginePressureByteDigest {
	sum := sha256.Sum256(raw)
	return imaginePressureByteDigest{Bytes: int64(len(raw)), SHA256: hex.EncodeToString(sum[:])}
}

func imaginePressureImageDigests(t *testing.T, images []ai.Image) ([]map[string]any, int64) {
	t.Helper()
	items := make([]map[string]any, 0, len(images))
	var total int64
	for i, img := range images {
		raw, err := base64.StdEncoding.Strict().DecodeString(img.Data)
		if err != nil {
			t.Fatal(err)
		}
		digest := imaginePressureDigest(raw)
		items = append(items, map[string]any{"ordinal": i, "mimeType": img.MimeType, "encodedBytes": len(img.Data), "decoded": digest})
		total += digest.Bytes
	}
	return items, total
}

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
	// Construct the prospective JSON independently at the public wire contract.
	// Input rejections have no actual HTTP payload; do not label this as sent.
	native := map[string]any{"model": imagineModel, "prompt": req.Prompt, "response_format": "b64_json", "n": count, "resolution": "1k", "quality": "low"}
	refs := make([]map[string]string, len(req.ReferenceImages))
	for i, img := range req.ReferenceImages {
		refs[i] = map[string]string{"type": "image_url", "url": "data:" + img.MimeType + ";base64," + img.Data}
	}
	if len(refs) == 1 {
		native["image"] = refs[0]
	} else if len(refs) > 0 {
		native["images"] = refs
	}
	constructed := mustMarshal(t, native)
	inputFacts, inputBytes := imaginePressureImageDigests(t, req.ReferenceImages)
	outputFacts, outputBytes := imaginePressureImageDigests(t, output)
	requestDigest, responseDigest := imaginePressureDigest(constructed), imaginePressureDigest([]byte(reply.Script))
	var limit int64
	p := newImaginePressureWorld(t, func(c *ai.Config) {
		switch kind {
		case "input-image":
			limit = c.Policy.MaxImageBytes
		case "input-total":
			limit = requestDigest.Bytes - 1
			c.Policy.MaxRequestBytes = limit
		case "output-image":
			limit = (512 << 10) - 1
			c.Policy.Image.MaxOutputImageBytes = limit
		case "output-total":
			limit = (5 << 20) - 1
			c.Policy.Image.MaxTotalOutputImageBytes = limit
		case "output-body":
			limit = responseDigest.Bytes - 1
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
	id := "pressure-boundary-" + kind
	result, err := p.client.WithHooks(ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
		encoded++
		return ai.KeepPayload(), nil
	}}).GenerateImages(ctx, imaginePressureScope(id, tenantA.tenant), imagineTarget(), req, opts)
	imaginePressureSummary(ev, "bounded-result", result, err)
	var failure *ai.Error
	ev.Check("one-byte overflow fails atomically", errors.As(err, &failure) && failure.Code == ai.CodeResourceLimit && len(result.Content) == 0, "err=%v", err)
	reads := p.host.ReadsFor(result.Metadata.RequestID).Credentials
	requests := p.provider.Requests()
	actual := int64(0)
	switch kind {
	case "input-image":
		actual = inputBytes
	case "input-total":
		actual = requestDigest.Bytes
	case "output-image":
		actual = outputBytes
	case "output-total":
		actual = outputBytes
	case "output-body":
		actual = responseDigest.Bytes
	}
	ev.Check("boundary differs by exactly one byte", actual == limit+1, "actual=%d limit=%d", actual, limit)
	if inputFailure {
		ev.Check("known input overflow precedes credentials and encoding", reads == 0 && encoded == 0 && len(requests) == 0 && len(result.Metadata.Attempts) == 0, "reads=%d callbacks=%d requests=%d", reads, encoded, len(requests))
	} else {
		ev.Check("2xx overflow has exactly one attempt", reads == 1 && encoded == 1 && len(requests) == 1 && len(result.Metadata.Attempts) == 1, "reads=%d callbacks=%d requests=%d", reads, encoded, len(requests))
		if len(result.Metadata.Attempts) == 1 {
			a := result.Metadata.Attempts[0]
			amount, known := a.ReportedCostUSD.Get()
			ev.Check("known charge and missing tokens survive partial output", known && amount == float64(ticks)/1e10 && a.HTTPStatus == 200 && a.UsageReporting == ai.UsageUnreported && a.Usage == (ai.Usage{}) && result.Usage == (ai.Usage{}) && a.CostEstimate.Status == ai.CostPartial, "attempt=%+v", a)
		}
	}
	captured := make([]map[string]any, 0, len(requests))
	for _, r := range requests {
		ev.Check("bounded HTTP has planned body and credential", jsonEqual(r.Body, constructed) && r.KeyAlias == imaginePressureAlias(tenantA.tenant) && r.Path == imaginePressurePrefix(tenantA.tenant, len(req.ReferenceImages) > 0) && r.Method == "POST" && r.Query == "", "kind=%s alias=%s path=%s", kind, r.KeyAlias, r.Path)
		captured = append(captured, map[string]any{"body": imaginePressureDigest(r.Body), "keyAlias": r.KeyAlias, "path": r.Path})
	}
	p.checkObserver(ev, imaginePressureCall{id: id, tenant: tenantA.tenant, req: req, result: result, err: err}, !inputFailure)
	ev.Record("byte-boundary", map[string]any{"kind": kind, "actualBytes": actual, "limitBytes": limit, "inputImages": inputFacts, "outputImages": outputFacts, "constructedRequest": requestDigest, "scriptedResponse": responseDigest, "capturedRequests": captured, "credentialReads": reads, "payloadCallbacks": encoded, "httpAttempts": len(requests), "realService": "NOT_RUN"})
	p.released(ev)
}
