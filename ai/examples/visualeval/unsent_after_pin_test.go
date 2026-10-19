package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestVisualUnsentAfterPin(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		t.Run(route, func(t *testing.T) {
			ev := unsentEvidence.Case(t, "VQ20-after-pin-"+route)
			p, err := load(route)
			if err != nil {
				t.Fatal(err)
			}
			a := arguments{mode: "controlled", route: route, entry: "hooked", calls: 12, workMS: 225000, callMS: 30000}
			r := initial(p, a)
			b, err := newBundle(filepath.Join(unsentEvidence.Dir(), "after-pin-"+route), "synthetic-unsent-credential")
			if err != nil || b.inputs(p, a, nil, &r) != nil {
				t.Fatal("controlled input setup")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hooksRun := 0
			hooks := ai.Hooks{TransformHeaders: func(context.Context, ai.CallScope, http.Header) error {
				hooksRun++
				cancel()
				return nil
			}}
			transport := &noWireTransport{}
			if err = executeSample(ctx, p, a, b, &r, 0, "http://127.0.0.1"+officialBasePath(route), "synthetic-unsent-credential", "synthetic-account", transport, hooks); err != nil {
				t.Fatal(err)
			}
			for i := 1; i < len(r.Samples); i++ {
				r.Samples[i].Status = "cancelled"
			}
			finalize(p, &r)
			s := r.Samples[0]
			ev.Record("public-result", s)
			ev.Check("real hook after credential pin", hooksRun == 1 && s.Result.Metadata.Resolved && s.Result.Metadata.CredentialVersion == "visual-evaluation-v1" && s.Assertions["credentialSnapshot"], "did not reach resolved public lifecycle")
			ev.Check("request cancellation never sends", s.Error != nil && s.Error.Code == ai.CodeCanceled && s.Error.Phase == ai.PhaseRequest && r.Budget.Calls == 1 && r.Budget.Attempts == 0 && transport.calls == 0, "wrong phase or send accounting")
			var access unsentCapture
			accessRaw, accessErr := os.ReadFile(filepath.Join(b.dir, "samples/v01/unsent.json"))
			ev.Check("measured credential read retained", accessErr == nil && strict(accessRaw, &access) == nil && access.CredentialReads == 1, "missing actual credential access")
			var obs []ai.Observation
			raw, err := os.ReadFile(filepath.Join(b.dir, "samples/v01/observations.json"))
			if err != nil || strict(raw, &obs) != nil {
				t.Fatal("missing lifecycle")
			}
			ev.Check("start unresolved terminal resolved", len(obs) == 2 && !obs[0].Call.Resolved && obs[1].Call.Resolved && same(obs[1].Call, s.Result.Metadata), "incorrect observer attribution")
			for _, name := range []string{"request.json", "response.bin"} {
				_, err := os.Lstat(filepath.Join(b.dir, "samples/v01", name))
				ev.Check("absent capture "+name, os.IsNotExist(err), "invented unsent capture")
			}
			err = b.finish(r)
			ev.Check("resolved unsent package audits", err == nil, "finish: %v", err)
			err = verify(b.dir, false)
			ev.Check("resolved unsent failure replays", err == nil, "verify: %v", err)
			ev.Check("failed evidence is never live", verify(b.dir, true) != nil, "promoted failure")
		})
	}
}
