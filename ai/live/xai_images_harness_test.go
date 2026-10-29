//go:build live

package live

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

func xaiImagesEntry() string { return xaiimages.Entry }

func TestXAIImagesCandidate(t *testing.T) {
	p := xaiImagesState
	if p == nil {
		t.Skip("NOT_RUN: select the Imagine candidate")
	}
	gate := run.Case(t, "XAIIMG-live-gate")
	if p.problem != "" {
		gate.Check("candidate configuration", false, "%s", p.problem)
		return
	}
	if p.notRun != "" {
		t.Skip("NOT_RUN: " + p.notRun)
	}
	if err := p.authorization.consume(); err != nil {
		p.problem = "fresh grant already consumed or unavailable"
		gate.Check("claim before calls", false, "%s", p.problem)
		return
	}
	p.report.Outcome = "RUNNING"
	if err := run.Record("xai-images-candidate-report", p.report); err != nil {
		t.Fatal("candidate report unavailable")
	}
	deadline := started.Add(1080 * time.Second)
	if expiry := p.authorization.grant.ExpiresAt; expiry.Before(deadline) {
		deadline = expiry
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	rec := newRecorder()
	rec.requestLimit, rec.responseLimit = 1<<20, 32<<20
	transport := rec.next.(*http.Transport)
	defer transport.CloseIdleConnections()
	var expected xaiimages.Scene
	if p.controlled != "" {
		rec.next = xaiRoundTripper(func(r *http.Request) (*http.Response, error) {
			if p.controlled == "cancel" {
				cancel()
				return nil, r.Context().Err()
			}
			return xaiImagesFixture(r, expected)
		})
	}
	next := rec.next
	rec.next = xaiRoundTripper(func(r *http.Request) (*http.Response, error) {
		path := "/images/generations"
		if expected.References > 0 {
			path = "/images/edits"
		}
		if r.URL.String() != xaiimages.Endpoint+path || p.report.HTTPAttempts >= 6 || !p.authorization.grant.ExpiresAt.After(time.Now()) {
			return nil, fmt.Errorf("Imagine send scope exhausted")
		}
		p.report.HTTPAttempts++
		if p.controlled == "" {
			p.report.VendorHTTPAttempts++
		}
		return next.RoundTrip(r)
	})
	policy := livePolicy()
	policy.MaxRequestBytes = 1 << 20
	policy.MaxOutputBytes = 32 << 20
	policy.MaxImageBytes = 512 << 10
	policy.Image = &ai.ImagePolicy{MaxInputImages: 5, MaxOutputImages: 2, MaxOutputImageBytes: 8 << 20, MaxTotalOutputImageBytes: 16 << 20}
	policy.CallTimeout = 180 * time.Second
	combo := &combo{name: xaiimages.Route, operation: ai.OperationImage, provider: ai.ProviderXAI, api: ai.APIXAIImages, endpoint: xaiimages.Endpoint, model: xaiimages.Model}
	client, err := newClientWithPolicy(combo, p.key, rec, p.cat, policy)
	if !gate.Check("shared Host assembly", err == nil, "%v", err) {
		return
	}
	for index, scene := range p.report.Plan {
		if t.Failed() || ctx.Err() != nil {
			break
		}
		expected = scene
		row := &p.report.Scenarios[index]
		ok := t.Run(scene.ID, func(t *testing.T) {
			cs := run.Case(t, row.CaseID)
			cs.ReplayRun(xaiimages.Entry)
			// Reserve before the logical call; a failed request never refunds approval.
			charge := p.report.ReservedUSD + scene.MaximumUSD
			if !cs.Check("finite reservation", p.report.CallsUsed < 6 && charge <= xaiimages.MaxUSD+1e-12 && p.report.ReportedUSD+scene.MaximumUSD <= xaiimages.MaxUSD+1e-12, "candidate budget exhausted") {
				return
			}
			p.report.CallsUsed++
			p.report.InputImages += scene.References
			p.report.OutputImages += scene.Outputs
			p.report.ReservedUSD = charge
			req := xaiImagesRequest(t, scene)
			opts := ai.XAIImagesOptions{N: ai.Value(scene.Outputs), Resolution: ai.Value(scene.Resolution), Quality: ai.Value(scene.Quality)}
			if scene.AspectRatio != "" {
				opts.AspectRatio = ai.Value(scene.AspectRatio)
			}
			callCtx, callCancel := context.WithTimeout(ctx, 180*time.Second)
			defer callCancel()
			scope := newScope()
			res, callErr := client.GenerateImages(callCtx, scope, ai.Target{BindingID: xaiimages.Route, ModelID: xaiimages.Model}, req, opts)
			exchanges := rec.take()
			cs.Record("result", res)
			cs.Record("error", errText(callErr))
			cs.Record("http", exchanges)
			observed := rec.observationsOf(scope.RequestID)
			cs.Record("observations", observed)
			cs.Check("terminal Observer", len(observed) > 0 && observed[len(observed)-1].Kind == ai.ObservationCallFinished, "missing terminal metadata")
			reported, known := 0.0, false
			if len(res.Metadata.Attempts) > 0 {
				reported, known = res.Metadata.Attempts[len(res.Metadata.Attempts)-1].ReportedCostUSD.Get()
				if known {
					p.report.ReportedUSD += reported
				}
			}
			if !cs.Check("one synchronous accepted attempt", callErr == nil && len(exchanges) == 1 && exchanges[0].Status == 200 && !exchanges[0].Truncated && len(res.Metadata.Attempts) == 1, "call=%v attempts=%d", callErr, len(exchanges)) {
				return
			}
			if !cs.Check("native request, result, and integer charge", validateXAIImagesCapture(exchanges[0], req, scene, res) == nil, "native capture validation failed") {
				return
			}
			if !cs.Check("reported charge controls continuation", known && reported >= 0 && reported <= scene.MaximumUSD+1e-12 && p.report.ReportedUSD <= xaiimages.MaxUSD+1e-12, "missing or exceeded report") {
				return
			}

			for n, block := range res.Content {
				image, ok := block.(ai.ImageOutputImage)
				if !cs.Check("image-only output", ok, "unexpected block") {
					continue
				}
				data, _ := base64.StdEncoding.Strict().DecodeString(image.Data)
				cs.Fixture(fmt.Sprintf("output-%d.image", n), data)
			}
		})
		row.Outcome = "PASS"
		if !ok {
			row.Outcome = "FAIL"
			break
		}
	}
}
