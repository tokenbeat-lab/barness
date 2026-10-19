package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var unsentEvidence *evidence.Run

func TestMain(m *testing.M) {
	// sourceFiles snapshots repository-relative production files. This package's
	// own evidence identity cannot be selected as the release offline E2E bundle.
	if err := os.Chdir("../../.."); err != nil {
		panic(err)
	}
	var err error
	unsentEvidence, err = evidence.NewRun("./ai/examples/visualeval")
	if err != nil {
		panic(err)
	}
	code := m.Run()
	if err := unsentEvidence.Finish(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	if findings, err := unsentEvidence.Audit(); err != nil || len(findings) != 0 {
		fmt.Fprintln(os.Stderr, "visual regression evidence audit failed", err, findings)
		code = 1
	}
	os.Exit(code)
}

type noWireTransport struct{ calls int }

func (n *noWireTransport) RoundTrip(*http.Request) (*http.Response, error) {
	n.calls++
	return nil, fmt.Errorf("unexpected transport invocation")
}

func TestVisualUnsentCancellation(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, entry := range []string{"plain", "hooked"} {
			for _, terminal := range []string{"cancel", "deadline"} {
				t.Run(route+"/"+entry+"/"+terminal, func(t *testing.T) {
					ev := unsentEvidence.Case(t, "VQ20-unsent-"+route+"-"+entry+"-"+terminal)
					p, err := load(route)
					if err != nil {
						t.Fatal(err)
					}
					a := arguments{mode: "controlled", route: route, entry: entry, calls: 12, workMS: 225000, callMS: 30000}
					r := initial(p, a)
					b, err := newBundle(filepath.Join(unsentEvidence.Dir(), "unsent-"+route+"-"+entry+"-"+terminal), "synthetic-unsent-credential")
					if err != nil {
						t.Fatal(err)
					}
					if err = b.inputs(p, a, nil, &r); err != nil {
						t.Fatal(err)
					}
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					want := ai.CodeCanceled
					if terminal == "deadline" {
						ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
						defer cancel()
						want = ai.CodeDeadlineExceeded
					}
					transport := &noWireTransport{}
					// This is the real writer after the outer precheck, with the real
					// public Client observing an already-ended parent context.
					if err = executeSample(ctx, p, a, b, &r, 0, "http://127.0.0.1"+officialBasePath(route), "synthetic-unsent-credential", "synthetic-account", transport, ai.Hooks{}); err != nil {
						t.Fatal(err)
					}
					for i := 1; i < len(r.Samples); i++ {
						r.Samples[i].Status = "cancelled"
					}
					finalize(p, &r)
					s := r.Samples[0]
					ev.Record("public-result", s)
					ev.Check("one logical call and no HTTP", r.Budget.Calls == 1 && r.Budget.Attempts == 0 && r.Budget.Retries == 0 && transport.calls == 0, "incorrect no-send accounting")
					ev.Check("public scope terminal remains unresolved", s.Error != nil && s.Error.Code == want && s.Error.Phase == ai.PhaseScope && !s.Result.Metadata.Resolved && s.Result.Metadata.RequestID != "" && len(s.Result.Metadata.Attempts) == 0, "incorrect public terminal")
					ev.Check("zero credential reads are correct", s.Assertions["credentialSnapshot"], "scope call demanded credential access")
					for name, n := range s.Resources {
						ev.Check("zero resource "+name, n == 0, "resource leaked")
					}
					ev.Check("whole denominator remains honest FAIL", r.Status == "FAIL" && r.RealQuality == "NOT_RUN" && len(r.Samples) == 12 && r.Metrics.Common.Total == 8, "failed call promoted or samples removed")
					for _, name := range []string{"request.json", "response.bin"} {
						_, e := os.Lstat(filepath.Join(b.dir, "samples/v01", name))
						ev.Check("absent capture "+name, os.IsNotExist(e), "invented empty wire artifact")
					}
					var access unsentCapture
					accessRaw, accessErr := os.ReadFile(filepath.Join(b.dir, "samples/v01/unsent.json"))
					ev.Check("measured zero credential reads retained", accessErr == nil && strict(accessRaw, &access) == nil && access.CredentialReads == 0, "missing measured credential access")
					var obs []ai.Observation
					raw, err := os.ReadFile(filepath.Join(b.dir, "samples/v01/observations.json"))
					if err != nil || strict(raw, &obs) != nil {
						t.Fatal("missing observations")
					}
					ev.Check("two exact lifecycle observations", len(obs) == 2 && obs[0].Kind == ai.ObservationCallStarted && obs[1].Kind == ai.ObservationCallFinished && same(obs[1].Call, s.Result.Metadata), "wrong observation sequence")
					err = b.finish(r)
					ev.Check("production finish independently audits", err == nil, "finish: %v", err)
					err = verify(b.dir, false)
					ev.Check("honest unsent failure publicly verifies", err == nil, "verify: %v", err)
					ev.Check("failed controlled evidence cannot be live", verify(b.dir, true) != nil, "failure promoted to live")
					if err == nil && route == "jev" && entry == "plain" && terminal == "cancel" {
						testUnsentTampering(t, b, r)
					}
				})
			}
		}
	}
}

func testUnsentTampering(t *testing.T, original *bundle, originalReport report) {
	t.Helper()
	mutations := map[string]func(*bundle, *report){
		"missing-measured-access": func(b *bundle, _ *report) { b.record("samples/v01/unsent.json", map[string]int{}) },
		"invented-access":         func(b *bundle, _ *report) { b.record("samples/v01/unsent.json", unsentCapture{CredentialReads: 9}) },
		"credential":              func(_ *bundle, r *report) { r.Samples[0].Result.Metadata.CredentialVersion = "invented" },
		"resolved":                func(_ *bundle, r *report) { r.Samples[0].Result.Metadata.Resolved = true },
		"attempt":                 func(_ *bundle, r *report) { r.Samples[0].Result.Metadata.Attempts = []ai.Attempt{{}} },
		"wire":                    func(b *bundle, _ *report) { b.record("samples/v01/wire.json", wire{Method: "POST"}) },
		"missing-terminal": func(b *bundle, _ *report) {
			b.record("samples/v01/observations.json", []ai.Observation{{Kind: ai.ObservationCallStarted}})
		},
		"invented-request": func(b *bundle, _ *report) { b.file("samples/v01/request.json", []byte("{}\n")) },
	}
	for name, mutate := range mutations {
		t.Run("tamper-"+name, func(t *testing.T) {
			ev := unsentEvidence.Case(t, "VQ20-tamper-"+name)
			b, err := newBundle(filepath.Join(unsentEvidence.Dir(), "tamper-"+name), original.secret)
			if err != nil {
				t.Fatal(err)
			}
			for name := range original.files {
				raw, err := os.ReadFile(filepath.Join(original.dir, name))
				if err != nil || b.file(name, raw) != nil {
					t.Fatal("copy controlled fixture")
				}
			}
			var r report
			if strict(encoded(originalReport), &r) != nil {
				t.Fatal("copy report")
			}
			mutate(b, &r)
			if err := b.record("samples/v01/result.json", r.Samples[0]); err != nil {
				t.Fatal(err)
			}
			if err := b.finish(r); err != nil {
				t.Fatal(err)
			}
			err = verify(b.dir, false)
			ev.Check("resealed fabricated evidence rejected", err != nil, "tampered package accepted")
		})
	}
}
