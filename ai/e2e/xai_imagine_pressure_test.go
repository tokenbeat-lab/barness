package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
)

// Public failure inventory, recorded before implementing the load driver:
// missing/partial ordered output; five-reference or ten-output bounds lost;
// oversized image/request/response accepted; cross-tenant credentials/results;
// repeated 2xx requests; unbounded heap/waiters; cancellation leaks; Observer
// aliases or content disclosure. These are synthetic Host loads, never vendor
// capability or real-price evidence. Ordinary/race runs stop at pressureCase.
func TestXAIImaginePressure(t *testing.T) {
	for _, mode := range []string{"plain", "hooked"} {
		t.Run("candidate-load-"+mode, func(t *testing.T) {
			ev := pressureCase(t, "XAI-IMG-pressure-candidate-load-"+mode)
			imaginePressureLoad(t, ev, mode)
		})
	}
	t.Run("admission-cancel", func(t *testing.T) {
		ev := pressureCase(t, "XAI-IMG-pressure-admission-cancel")
		imaginePressureQueue(t, ev)
	})
	for _, boundary := range []string{"input-image", "input-total", "output-image", "output-total", "output-body"} {
		t.Run("byte-limits-"+boundary, func(t *testing.T) {
			ev := pressureCase(t, "XAI-IMG-pressure-byte-limits-"+boundary)
			imaginePressureBoundary(t, ev, boundary)
		})
	}
}

type imaginePressureWorld struct {
	*world
	gauges  *probe.Probe
	bodies  *provider.BodyTracker
	records *recorder
}

// Route only the already redirected loopback request. Synthetic credentials
// select independent HTTP queues; arrival order never chooses tenant content.
type imaginePressureRoute struct{ next http.RoundTripper }

func (r imaginePressureRoute) RoundTrip(req *http.Request) (*http.Response, error) {
	tenant := ""
	switch req.Header.Get("Authorization") {
	case "Bearer " + xaiKey.secret:
		tenant = tenantA.tenant
	case "Bearer " + tenantB.secret:
		tenant = tenantB.tenant
	default:
		return nil, fmt.Errorf("unconfigured synthetic pressure credential")
	}
	copy := req.Clone(req.Context())
	u := *req.URL
	u.Path = "/pressure/" + tenant + u.Path
	copy.URL = &u
	return r.next.RoundTrip(copy)
}
func newImaginePressureWorld(t *testing.T, configure func(*ai.Config)) *imaginePressureWorld {
	t.Helper()
	p := &imaginePressureWorld{gauges: probe.New(), records: newRecorder()}
	candidate := xai.ImagesCandidateCatalog()
	p.world = newImagineWorld(t, func(c *ai.Config) {
		c.Catalog = &candidate
		c.Policy.MaxImageBytes = 512 << 10
		c.Policy.MaxRequestBytes = 4 << 20
		c.Policy.MaxOutputBytes = 12 << 20
		c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 5, MaxOutputImages: 10, MaxOutputImageBytes: 1 << 20, MaxTotalOutputImageBytes: 8 << 20}
		c.Policy.MaxConcurrentPerTenant = 2
		c.Policy.MaxConcurrentProcess = 4
		c.Policy.MaxAdmissionWaiters = 2
		c.Policy.AdmissionWait = 30 * time.Second
		c.Policy.CallTimeout = pressureDeadline
		c.Policy.MaxQueuedObservations = 128
		c.Observer = p.records
		c.Probe = p.gauges
		c.Transport.(*imagineTransport).next = imaginePressureRoute{provider.LoopbackTransport()}
		p.bodies = provider.TrackBodies(c.Transport)
		c.Transport = p.bodies
		if configure != nil {
			configure(c)
		}
	})
	b := imagineBinding()
	b.TenantID = tenantB.tenant
	b.AccountScopeID = "pressure-account-b"
	p.host.PutBinding(b)
	p.host.PutCredential(ai.Credential{OwnerTenantID: tenantB.tenant, CredentialID: b.CredentialRef, Version: "b-v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantB.secret)})
	return p
}
func imaginePressurePrefix(tenant string, editing bool) string {
	path := "/v1/images/generations"
	if editing {
		path = "/v1/images/edits"
	}
	return "/pressure/" + tenant + path
}
func imaginePressureScope(id, tenant string) ai.CallScope {
	s := textScope(id)
	s.TenantID = tenant
	return s
}
func imaginePressureOptions(n int) ai.XAIImagesOptions {
	return ai.XAIImagesOptions{N: ai.Value(n), Resolution: ai.Value("1k"), Quality: ai.Value("low")}
}
func imaginePressureImage(t *testing.T, size, tag int) ai.Image {
	t.Helper()
	img := mixedPNG(t, size)
	raw, err := base64.StdEncoding.DecodeString(img.Data)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] = byte(tag)
	img.Data = base64.StdEncoding.EncodeToString(raw)
	return img
}
func imaginePressureHash(data string) string {
	raw, _ := base64.StdEncoding.Strict().DecodeString(data)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func imaginePressureReply(images []ai.Image, ticks int64) provider.Reply {
	entries := make([]map[string]string, len(images))
	for i, img := range images {
		entries[i] = map[string]string{"b64_json": img.Data, "mime_type": img.MimeType}
	}
	raw, _ := json.Marshal(entries)
	return imagineReply(fmt.Sprintf(`{"usage":{"cost_in_usd_ticks":%d},"data":%s}`, ticks, raw))
}
func (p *imaginePressureWorld) released(ev *evidence.Case) {
	eventually(ev, "pressure resources return to zero", func() bool {
		return p.gauges.ActiveCalls() == 0 && p.gauges.Permits() == 0 && p.gauges.AdmissionWaiters() == 0 && p.bodies.Open() == 0
	})
	ev.Record("resources", map[string]int64{"calls": p.gauges.ActiveCalls(), "permits": p.gauges.Permits(), "waiters": p.gauges.AdmissionWaiters(), "bodies": p.bodies.Open()})
}
func imaginePressureSummary(ev *evidence.Case, name string, res ai.ImagesResult, err error) {
	hashes := make([]string, 0, len(res.Content))
	for _, block := range res.Content {
		if img, ok := block.(ai.ImageOutputImage); ok {
			hashes = append(hashes, imaginePressureHash(img.Data))
		}
	}
	ev.Record(name, map[string]any{"metadata": res.Metadata, "usage": res.Usage, "stopReason": res.StopReason, "error": errString(err), "imageHashes": hashes})
}

// Expected identity comes from the planned public call, never a received alias
// or result. CallStarted precedes snapshot resolution, including on success.
func imaginePressureMetadata(ev *evidence.Case, id, tenant string, resolved bool) ai.CallMetadata {
	scope, b := imaginePressureScope(id, tenant), imagineBinding()
	want := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: tenant, RequestID: id, ActorID: scope.ActorID, JobID: scope.JobID, BindingID: b.BindingID, Operation: ai.OperationImage}}
	if resolved {
		cat := xai.ImagesCandidateCatalog()
		hash, err := cat.Hash()
		ev.Check("candidate price identity is valid", err == nil, "err=%v", err)
		want.Resolved, want.ProviderID, want.API, want.ModelID = true, ai.ProviderXAI, ai.APIXAIImages, imagineModel
		want.AccountScopeID, want.BindingVersion, want.CredentialVersion = b.AccountScopeID, b.Version, "v1"
		if tenant == tenantB.tenant {
			want.AccountScopeID, want.CredentialVersion = "pressure-account-b", "b-v1"
		}
		want.CatalogVersion, want.CatalogHash = cat.Version, hash
	}
	return want
}
func imaginePressureAlias(tenant string) string {
	if tenant == tenantB.tenant {
		return tenantB.alias
	}
	return xaiKey.alias
}
func (p *imaginePressureWorld) checkObserver(ev *evidence.Case, c imaginePressureCall, resolved bool) {
	res := c.result
	observations := p.records.awaitCall(ev, c.id)
	want := imaginePressureMetadata(ev, c.id, c.tenant, resolved)
	meta := res.Metadata
	meta.Attempts = nil
	ev.Check("result has independently planned identity", reflect.DeepEqual(meta, want), "call=%s metadata=%+v", c.id, meta)
	ev.Check("complete bounded Observer lifecycle", reflect.DeepEqual(kinds(observations), lifecycle(c.id, len(res.Metadata.Attempts))), "call=%s lifecycle=%v", c.id, kinds(observations))
	// Reporting belongs only to actual attempts. Zero Usage on an unsent
	// call is an empty carrier, never a claim of reported vendor counters.
	if len(res.Metadata.Attempts) == 0 {
		ev.Check("zero-attempt result has no reported consumption", res.Usage == (ai.Usage{}), "call=%s usage=%+v", c.id, res.Usage)
	}
	for i, o := range observations {
		ev.Check("Observer time is present and ordered", !o.Time.IsZero() && (i == 0 || !o.Time.Before(observations[i-1].Time)), "call=%s record=%d", c.id, i)
		switch o.Kind {
		case ai.ObservationCallStarted:
			ev.Check("entry has scope only before resolution", reflect.DeepEqual(o.Call, imaginePressureMetadata(ev, c.id, c.tenant, false)) && o.Attempt == nil, "call=%s metadata=%+v", c.id, o.Call)
		case ai.ObservationAttemptStarted, ai.ObservationAttemptFinished:
			ev.Check("attempt has independently planned complete attribution", reflect.DeepEqual(o.Call, want) && o.Attempt != nil, "call=%s metadata=%+v", c.id, o.Call)
			if o.Attempt == nil {
				continue
			}
			ordinal := (i - 1) / 2
			if !ev.Check("attempt ordinal is bounded", ordinal < len(res.Metadata.Attempts), "call=%s ordinal=%d", c.id, ordinal) {
				continue
			}
			a := res.Metadata.Attempts[ordinal]
			ev.Check("attempt ID follows planned request", o.Attempt.AttemptID == fmt.Sprintf("%s#%d", c.id, ordinal+1), "call=%s", c.id)
			if o.Kind == ai.ObservationAttemptFinished {
				ev.Check("Observer retains all terminal attempt facts", reflect.DeepEqual(*o.Attempt, a) && o.Duration > 0, "call=%s attempt=%+v", c.id, o.Attempt)
			} else {
				ev.Check("attempt starts without invented consumption", o.Attempt.UsageReporting == ai.UsageUnreported && o.Attempt.Usage == (ai.Usage{}) && o.Attempt.ReportedCostUSD == (ai.Nullable[float64]{}), "call=%s", c.id)
			}
		case ai.ObservationCallFinished:
			ev.Check("Observer retains terminal metadata usage and stop", reflect.DeepEqual(o.Call, res.Metadata) && o.Usage == res.Usage && o.StopReason == res.StopReason && o.Attempt == nil && o.Duration > 0, "call=%s", c.id)
			if len(res.Metadata.Attempts) == 0 {
				ev.Check("zero-attempt Observer has no reported consumption", o.Usage == (ai.Usage{}) && len(o.Call.Attempts) == 0 && o.Attempt == nil, "call=%s usage=%+v", c.id, o.Usage)
			}
			var failure *ai.Error
			var classification *ai.ObservedError
			if errors.As(c.err, &failure) {
				classification = &ai.ObservedError{Code: failure.Code, Phase: failure.Phase, HTTPStatus: failure.HTTPStatus, ProviderRequestID: failure.ProviderRequestID, RetryAfter: failure.RetryAfter}
			}
			ev.Check("Observer retains classification without error text", reflect.DeepEqual(o.Error, classification), "call=%s error=%+v", c.id, o.Error)
		}
	}
	before := string(mustMarshal(ev.T(), observations))
	for _, forbidden := range []string{"b64_json", "data:image/", "pressure prompt", "errorMessage", "iVBORw0KGgo", xaiKey.secret, tenantB.secret} {
		ev.Check("Observer contains metadata only", !strings.Contains(before, forbidden), "call=%s content marker present", c.id)
	}
	if len(res.Metadata.Attempts) > 0 {
		original := res.Metadata.Attempts[0].ReportedCostUSD
		res.Metadata.Attempts[0].ReportedCostUSD = ai.Value(999.0)
		ev.Check("result mutation cannot change Observer charge", string(mustMarshal(ev.T(), p.records.forRequest(c.id))) == before, "call=%s Observer aliases result", c.id)
		res.Metadata.Attempts[0].ReportedCostUSD = original
	}
}

type imaginePressureCall struct {
	id, tenant string
	req        ai.ImagesRequest
	opts       ai.XAIImagesOptions
	reply      provider.Reply
	want       []string
	amount     float64
	result     ai.ImagesResult
	err        error
}

func imaginePressureLoad(t *testing.T, ev *evidence.Case, mode string) {
	calls := make([]imaginePressureCall, 0, 4)
	for n, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		for _, editing := range []bool{false, true} {
			c := imaginePressureCall{id: fmt.Sprintf("pressure-%s-%d-%t", mode, n, editing), tenant: tenant, opts: imaginePressureOptions(10), amount: .40}
			c.req.Prompt = "pressure prompt " + c.id
			if editing {
				c.opts = imaginePressureOptions(1)
				c.amount = .09
				for i := range 5 {
					c.req.ReferenceImages = append(c.req.ReferenceImages, imaginePressureImage(t, 256<<10, 1+n*16+i))
				}
			}
			count, _ := c.opts.N.Get()
			output := make([]ai.Image, 0, count)
			for i := range count {
				img := imaginePressureImage(t, 512<<10, 64+n*16+i)
				output = append(output, img)
				c.want = append(c.want, imaginePressureHash(img.Data))
			}
			ticks := int64(4000000000)
			if editing {
				ticks = 900000000
			}
			c.reply = imaginePressureReply(output, ticks)
			calls = append(calls, c)
		}
	}
	complete, release := make(chan struct{}, 4), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	eof := &mixedEOFTransport{ended: func() { complete <- struct{}{} }, release: release}
	p := newImaginePressureWorld(t, func(c *ai.Config) { eof.next = c.Transport; c.Transport = eof })
	for _, c := range calls {
		p.provider.EnqueueAt(imaginePressurePrefix(c.tenant, len(c.req.ReferenceImages) > 0), c.reply)
	}
	heap := startHeapSampler(p.gauges.Permits)
	defer heap.finish()
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	defer unblock()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	hookCalls := make([]int, len(calls))
	for i := range calls {
		wg.Go(func() {
			c := &calls[i]
			scope := imaginePressureScope(c.id, c.tenant)
			if mode == "hooked" {
				c.result, c.err = p.client.WithHooks(ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
					hookCalls[i]++
					return ai.KeepPayload(), nil
				}}).GenerateImages(ctx, scope, imagineTarget(), c.req, c.opts)
			} else {
				c.result, c.err = p.client.GenerateImages(ctx, scope, imagineTarget(), c.req, c.opts)
			}
		})
	}
	for range 4 {
		waitForWithin(ev, "complete pressure body reaches EOF", pressureDeadline, complete)
	}
	heldPermits, heldBodies := p.gauges.Permits(), p.bodies.Open()
	ev.Check("four simultaneous permits and bodies", heldPermits == 4 && heldBodies == 4, "permits=%d bodies=%d", heldPermits, heldBodies)
	runtime.GC()
	heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
	heap.peakGauge.Store(max(heap.peakGauge.Load(), p.gauges.Permits()))
	unblock()
	waitForWithin(ev, "all bounded pressure calls settle", pressureDeadline, doneWhen(wg.Wait))
	runtime.GC()
	heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
	runtime.ReadMemStats(&after)
	use := heap.finish()
	ev.Check("live heap growth remains below 512 MiB", use.LiveGrowth() < 512<<20, "growth=%d", use.LiveGrowth())
	requests := p.provider.Requests()
	ev.Check("one attempt per call without replay", len(requests) == 4, "requests=%d", len(requests))
	rooms := map[string]headroom{}
	for i, c := range calls {
		imaginePressureSummary(ev, c.id, c.result, c.err)
		if !ev.Check("complete ordered candidate output", c.err == nil && len(c.result.Content) == len(c.want), "call=%s err=%v images=%d", c.id, c.err, len(c.result.Content)) {
			continue
		}
		for j, block := range c.result.Content {
			img, ok := block.(ai.ImageOutputImage)
			ev.Check("ordered output hash", ok && img.MimeType == "image/png" && imaginePressureHash(img.Data) == c.want[j], "call=%s image=%d", c.id, j)
		}
		account := imagineBinding().AccountScopeID
		if c.tenant == tenantB.tenant {
			account = "pressure-account-b"
		}
		ev.Check("tenant account and catalog remain pinned", c.result.Metadata.TenantID == c.tenant && c.result.Metadata.AccountScopeID == account && c.result.Metadata.CatalogVersion == xai.ImagesCandidateCatalog().Version, "wrong attribution")
		if ev.Check("one synchronous attempt", len(c.result.Metadata.Attempts) == 1 && c.result.Metadata.Attempts[0].HTTPStatus == 200, "call=%s", c.id) {
			a := c.result.Metadata.Attempts[0]
			amount, known := a.ReportedCostUSD.Get()
			lo, lok := a.CostEstimate.Minimum.Get()
			hi, hok := a.CostEstimate.Maximum.Get()
			ev.Check("charge and resolved estimate stay separate from missing tokens", known && math.Abs(amount-c.amount) < 1e-12 && lok && hok && a.CostEstimate.Status == ai.CostResolved && math.Abs(lo.Total-c.amount) < 1e-12 && math.Abs(hi.Total-c.amount) < 1e-12 && a.UsageReporting == ai.UsageUnreported && c.result.Usage == (ai.Usage{}), "call=%s attempt=%+v", c.id, a)
		}
		p.checkObserver(ev, c, true)
		wantHooks := 0
		if mode == "hooked" {
			wantHooks = 1
		}
		ev.Check("payload callback executes once", hookCalls[i] == wantHooks, "call=%s count=%d", c.id, hookCalls[i])
		matched := 0
		for _, r := range requests {
			var wire struct {
				Model               string
				N                   int
				Quality, Resolution string
				ResponseFormat      string `json:"response_format"`
				Prompt              string
				Image               *struct{ Type, URL string }
				Images              []struct{ Type, URL string }
			}
			mustUnmarshal(t, r.Body, &wire)
			if wire.Prompt != c.req.Prompt {
				continue
			}
			matched++
			wantN, _ := c.opts.N.Get()
			ev.Check("fixed native charge-bearing request", wire.Model == imagineModel && wire.N == wantN && wire.Quality == "low" && wire.Resolution == "1k" && wire.ResponseFormat == "b64_json", "call=%s", c.id)
			alias := xaiKey.alias
			if c.tenant == tenantB.tenant {
				alias = tenantB.alias
			}
			ev.Check("wire keeps tenant credential and native endpoint", r.KeyAlias == alias && r.Path == imaginePressurePrefix(c.tenant, len(c.req.ReferenceImages) > 0), "call=%s", c.id)
			refs := wire.Images
			if wire.Image != nil {
				refs = append(refs, *wire.Image)
			}
			ev.Check("native reference count", len(refs) == len(c.req.ReferenceImages), "call=%s", c.id)
			for j, ref := range refs {
				if j < len(c.req.ReferenceImages) {
					img := c.req.ReferenceImages[j]
					ev.Check("ordered inline reference", ref.Type == "image_url" && ref.URL == "data:"+img.MimeType+";base64,"+img.Data, "call=%s reference=%d", c.id, j)
				}
			}
			sum := sha256.Sum256(r.Body)
			ev.Record(c.id+"-request", map[string]any{"bytes": len(r.Body), "sha256": hex.EncodeToString(sum[:]), "path": r.Path, "keyAlias": r.KeyAlias})
			rooms[c.id+"-request"] = room(int64(len(r.Body)), 4<<20)
		}
		ev.Check("each frozen request is captured exactly once", matched == 1, "call=%s matches=%d", c.id, matched)
		rooms[c.id+"-response"] = room(int64(len(c.reply.Script)), 12<<20)
	}
	rooms["input-image"] = room(256<<10, 512<<10)
	rooms["output-image"] = room(512<<10, 1<<20)
	rooms["generated-total"] = room(5<<20, 8<<20)
	for name, headroom := range rooms {
		ev.Check("finite byte headroom "+name, headroom.Ratio <= maxDesignRatio, "design=%d limit=%d ratio=%f", headroom.Design, headroom.Limit, headroom.Ratio)
	}
	ev.Record("pressure-load", map[string]any{"heldPermitsAtEOF": heldPermits, "heldBodiesAtEOF": heldBodies, "peakPermits": heap.peakGauge.Load(), "mode": mode, "tenants": 2, "calls": 4, "inputReferencesPerEdit": 5, "generatedImagesPerCall": 10, "inputImageBytes": 256 << 10, "outputImageBytes": 512 << 10, "heap": use, "heapCeiling": 512 << 20, "totalAlloc": after.TotalAlloc - before.TotalAlloc, "clientBytesAtEOF": eof.read.Load(), "headroom": rooms, "duration": time.Since(started).String(), "realService": "NOT_RUN"})
	p.released(ev)
}
