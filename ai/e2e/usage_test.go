package e2e

import (
	"errors"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// TestUsageAndCost covers E11 (spec I10, User Stories 40–41): the message
// keeps pi's Usage, estimated cost included, on the success, failure and
// retry paths with usage missing, null, zero or reported, while every
// attempt separately records how completely its usage was reported, priced
// with a versioned catalog the call's metadata names.
//
// Ways it could fail:
//
//	U1 missing or null usage reads as a complete zero report, so an
//	   unreported attempt looks free
//	U2 a zero report reads as unreported, or loses pi's reasoning: 0
//	U3 usage without its detail objects reads as complete, or is dropped
//	U4 cached or cache-write tokens are not subtracted from input, or are
//	   priced at the input rate
//	U5 reasoning tokens are added to the total or priced on top of output
//	U6 a tier applies at its threshold instead of strictly above it, counts
//	   input alone, or is ignored above it
//	U7 the service tier adjustment uses the request over the response's
//	   tier, misses flex/priority, or is not reflected in the total
//	U8 a failed response's usage reaches the message or its attempt (pi
//	   drops it), or the attempt reads as a complete report rather than
//	   unreported, so the failure looks free
//	U9 a retried call sums its attempts into the message usage, or a
//	   failed attempt borrows the successful one's usage
//	U10 the attempt records reach the Observer without the usage reporting,
//	   or differ from the Result's
//	U11 the metadata does not name the catalog version and hash the cost
//	   was estimated with, or a host catalog's prices are ignored
//	U12 the built-in catalog's content changes without a version change, or
//	   a catalog with a negative or non-finite price is accepted
func TestUsageAndCost(t *testing.T) {
	f, raw := loadUsageFixture(t)
	text, textRaw := loadTextFixture(t, "text-basic.json")
	entries := []struct {
		name   string
		invoke func(*world, usageScenario) outcome
	}{
		{"stream-full", func(w *world, sc usageScenario) outcome {
			return drain(w.client.Stream(ctxFor(t), textScope("req-e11-"+sc.ID+"-stream"), ai.Target{BindingID: "primary", ModelID: f.model(sc)}, textRequest(text), sc.options()))
		}},
		{"complete-full", func(w *world, sc usageScenario) outcome {
			res, err := w.client.Complete(ctxFor(t), textScope("req-e11-"+sc.ID+"-complete"), ai.Target{BindingID: "primary", ModelID: f.model(sc)}, textRequest(text), sc.options())
			return outcome{result: res, err: err}
		}},
	}
	for _, sc := range f.Scenarios {
		for _, e := range entries {
			t.Run(sc.ID+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E11-"+sc.ID+"-"+e.name)
				ev.Fixture("usage.json", raw)
				ev.Fixture("text-basic.json", textRaw)
				rec := newRecorder()
				catalog := ai.BuiltinCatalog()
				w := observedWorld(t, rec, func(c *ai.Config) {
					withModelPatch(f.model(sc), nil, sc.ModelPatch)(c)
					if c.Catalog != nil {
						catalog = *c.Catalog
					}
				}, tenantA)
				w.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = sc.Retry.policy() })
				enqueue(ev, w, f.replies(t, sc)...)
				o := within(ev, "call ends", func() outcome { return e.invoke(w, sc) })
				o.record(ev)
				checkUsageOutcome(ev, sc, o)
				checkPricingSnapshot(ev, o.result.Metadata, catalog.Version, catalogHash(t, catalog))
				obs := rec.awaitCall(ev, o.result.Metadata.RequestID)
				checkCallRecords(ev, obs, textScope(o.result.Metadata.RequestID), "primary", o, len(sc.Expect.Attempts))
			})
		}
	}

	t.Run("builtin-catalog-pinned", func(t *testing.T) {
		ev := run.Case(t, "P01-E11-builtin-catalog-pinned")
		ev.Fixture("usage.json", raw)
		c := ai.BuiltinCatalog()
		hash := catalogHash(t, c)
		ev.Record("catalog", map[string]string{"version": c.Version, "hash": hash})
		ev.Check("built-in catalog version", c.Version == f.Catalog.Version, "got %q want %q", c.Version, f.Catalog.Version)
		ev.Check("built-in catalog content is the pinned one (bump Version, then the pin, on any change)", hash == f.Catalog.Hash,
			"got %s want %s", hash, f.Catalog.Hash)
		ev.Check("hash is stable", hash == catalogHash(t, ai.BuiltinCatalog()), "differs between copies")
		changed := ai.BuiltinCatalog()
		changed.Models[0].Cost.Output += 0.01
		ev.Check("a price change changes the hash", catalogHash(t, changed) != hash, "unchanged")
		changed.Models[0].Cost.Output = math.NaN()
		_, err := changed.Hash()
		ev.Check("an invalid catalog has no hash", errors.Is(err, ai.ErrInvalidConfig), "err=%v", err)
	})

	t.Run("invalid-prices-rejected", func(t *testing.T) {
		ev := run.Case(t, "P01-E11-invalid-prices-rejected")
		for name, change := range map[string]func(*ai.ModelCost){
			"negative-input":      func(c *ai.ModelCost) { c.Input = -1 },
			"nan-output":          func(c *ai.ModelCost) { c.Output = math.NaN() },
			"infinite-cache-read": func(c *ai.ModelCost) { c.CacheRead = math.Inf(1) },
			"negative-tier-rate": func(c *ai.ModelCost) {
				c.Tiers = []ai.CostTier{{InputTokensAbove: 10, CostRates: ai.CostRates{CacheWrite: -1}}}
			},
			"negative-threshold": func(c *ai.ModelCost) { c.Tiers = []ai.CostTier{{InputTokensAbove: -1}} },
		} {
			catalog := ai.BuiltinCatalog()
			change(&catalog.Models[0].Cost)
			h := host.New()
			_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &catalog})
			var ce *ai.ConfigError
			ev.Check(name+": construction fails on the catalog", errors.As(err, &ce) && ce.Field == "Catalog" && errors.Is(err, ai.ErrInvalidConfig), "err=%v", err)
		}
	})
}

// checkUsageOutcome asserts the message's pi usage and each attempt's own
// record against the scenario.
func checkUsageOutcome(ev *evidence.Case, sc usageScenario, o outcome) {
	x, m := sc.Expect, o.result.Message
	ev.Check("stop reason", m.StopReason == x.StopReason, "got %q want %q", m.StopReason, x.StopReason)
	var ae *ai.Error
	if x.Code == "" {
		ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	} else {
		ev.Check("error code", errors.As(o.err, &ae) && ae.Code == x.Code, "err=%v want %s", o.err, x.Code)
	}
	ev.Check("message usage is pi's, cost included", m.Usage == x.Usage, "got %+v want %+v", m.Usage, x.Usage)
	if o.streamed && len(o.events) > 0 {
		var terminal ai.AssistantMessage
		switch e := o.events[len(o.events)-1].(type) {
		case ai.DoneEvent:
			terminal = e.Message
		case ai.ErrorEvent:
			terminal = e.Message
		}
		ev.Check("terminal event carries the same usage", terminal.Usage == m.Usage, "got %+v want %+v", terminal.Usage, m.Usage)
	}

	got := o.result.Metadata.Attempts
	if !ev.Check("one record per attempt", len(got) == len(x.Attempts), "got %+v want %+v", got, x.Attempts) {
		return
	}
	for i, a := range got {
		w := x.Attempts[i]
		ev.Check("attempt outcome", a.HTTPStatus == w.HTTPStatus && a.Code == w.Code, "attempt %d got %+v want %+v", i+1, a, w)
		ev.Check("attempt usage reporting", a.UsageReporting == w.UsageReporting, "attempt %d got %q want %q", i+1, a.UsageReporting, w.UsageReporting)
		ev.Check("attempt usage", a.Usage == w.Usage, "attempt %d got %+v want %+v", i+1, a.Usage, w.Usage)
	}
	if last := got[len(got)-1]; x.Code == "" {
		ev.Check("message usage is the streamed attempt's alone", m.Usage == last.Usage, "message %+v attempt %+v", m.Usage, last.Usage)
	}
}

// checkPricingSnapshot asserts the metadata names the catalog the Client
// priced with.
func checkPricingSnapshot(ev *evidence.Case, meta ai.CallMetadata, version, hash string) {
	ev.Check("metadata names the catalog version", meta.CatalogVersion == version, "got %q want %q", meta.CatalogVersion, version)
	ev.Check("metadata names the catalog hash", meta.CatalogHash == hash, "got %q want %q", meta.CatalogHash, hash)
}

// catalogHash is c's hash; c must be a valid catalog.
func catalogHash(t *testing.T, c ai.Catalog) string {
	t.Helper()
	h, err := c.Hash()
	if err != nil {
		t.Fatal(err)
	}
	return h
}
