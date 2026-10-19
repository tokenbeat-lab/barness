package main

import (
	"context"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"time"
)

// unsentCapture records actual Host access independently of whether pin
// completed. It is local evidence, not an invented HTTP request or an Attempt.
type unsentCapture struct {
	CredentialReads int `json:"credentialReads"`
}

func initialMetadata(route, request string) ai.CallMetadata {
	return ai.CallMetadata{CallAttribution: ai.CallAttribution{
		TenantID: "visual-evaluation-host", RequestID: request,
		BindingID: route, Operation: ai.OperationClassifier,
	}}
}

// Context termination can occur before resolution, during credential access,
// after pin, or before admission produces an Attempt. Only these finite states
// are admitted; the transmitted branch keeps its existing exact checks.
func unsentTermination(p *plan, alias string, s sampleResult, reads int) bool {
	md := s.Result.Metadata
	if s.Error == nil || (s.Error.Code != ai.CodeCanceled && s.Error.Code != ai.CodeDeadlineExceeded) || md.RequestID == "" {
		return false
	}
	expected := initialMetadata(p.route, md.RequestID)
	switch s.Error.Phase {
	case ai.PhaseScope, ai.PhaseBinding:
		if reads != 0 || md.Resolved {
			return false
		}
	case ai.PhaseCapability:
		if p.route != "jev" || reads != 0 || md.Resolved {
			return false
		}
	case ai.PhaseCredential:
		if (reads != 0 && reads != 1) || md.Resolved {
			return false
		}
	case ai.PhaseRequest:
		// Jev validates and projects images before pin; Decisions prepares its
		// request after pin. The initial identity must remain unresolved here.
		if !md.Resolved {
			if p.route != "jev" || reads != 0 {
				return false
			}
			break
		}
		fallthrough
	case ai.PhaseAdmission:
		if reads != 1 || !md.Resolved {
			return false
		}
		expected.Resolved = true
		expected.ProviderID, expected.API, expected.ModelID = p.model.Provider, p.model.API, p.model.ID
		expected.AccountScopeID = alias
		expected.BindingVersion, expected.CredentialVersion = "visual-evaluation-v1", "visual-evaluation-v1"
		expected.CatalogVersion = p.catalog.Version
		expected.CatalogHash, _ = p.catalog.Hash()
	default:
		return false
	}
	return same(md, expected) && len(s.Result.Answers) == 0 && len(s.Result.ImageInputs) == 0 &&
		same(s.Result.ResponseModel, ai.Nullable[string]{}) && same(s.Result.Usage, ai.Usage{})
}

func verifyUnsent(p *plan, i int, entry, alias string, s sampleResult, w wire, obs []ai.Observation, access unsentCapture) error {
	if !unsentTermination(p, alias, s, access.CredentialReads) || !same(w, wire{}) || len(obs) != 2 ||
		!same(s.Resources, map[string]int64{"bodies": 0, "calls": 0, "permits": 0, "waiters": 0}) {
		return errors.New("invalid invoked-but-unsent termination")
	}
	start, end := obs[0], obs[1]
	stop := ai.StopReasonError
	if s.Error.Phase == ai.PhaseRequest || s.Error.Phase == ai.PhaseAdmission {
		stop = ai.StopReasonAborted
	}
	if start.Kind != ai.ObservationCallStarted || end.Kind != ai.ObservationCallFinished || !same(start.Call, initialMetadata(p.route, s.Result.Metadata.RequestID)) || !same(end.Call, s.Result.Metadata) ||
		start.Attempt != nil || end.Attempt != nil || start.Duration != 0 || start.StopReason != "" || start.Error != nil || !same(start.Usage, ai.Usage{}) ||
		end.StopReason != stop || end.Error == nil || end.Error.Code != s.Error.Code || end.Error.Phase != s.Error.Phase ||
		start.Time.Before(s.StartedAt) || end.Time.Before(start.Time) || end.Time.After(s.FinishedAt) || end.Duration < 0 || end.Duration > s.FinishedAt.Sub(s.StartedAt) {
		return errors.New("unsent lifecycle drift")
	}
	// There are no vendor bytes to replay. Re-run the fixed public task under
	// the recorded context terminal; exact legal stage/access/attribution above
	// is checked separately, since the original scheduling boundary is not data.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if s.Error.Code == ai.CodeDeadlineExceeded {
		ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()
	}
	a := arguments{mode: "controlled", entry: entry, callMS: 30000}
	got, generated, _, reads, err := call(ctx, p, a, p.request(i), i, "http://127.0.0.1"+officialBasePath(p.route), "synthetic-replay-credential", "synthetic-replay-account", replayTransport{}, ai.Hooks{})
	if err != nil || !unsentTermination(p, "synthetic-replay-account", got, reads) || got.Error.Code != s.Error.Code || !same(generated, wire{}) {
		return errors.New("unsent public task replay differs")
	}
	return nil
}
