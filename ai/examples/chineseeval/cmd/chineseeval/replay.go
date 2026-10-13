package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

// Replay consumes recorded bytes through the real public Client. It cannot
// open sockets and never regenerates the fixture producer's answers.
type replayTransport struct {
	captures []fixtureCapture
	samples  []chineseeval.SampleResult
	next     int
	cancel   context.CancelFunc
}

func (r *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.next >= len(r.captures) {
		return nil, errors.New("unexpected replay attempt")
	}
	c, s := r.captures[r.next], r.samples[r.next]
	r.next++
	raw, err := io.ReadAll(io.LimitReader(req.Body, 16385))
	req.Body.Close()
	var actual, recorded bytes.Buffer
	if err != nil || len(raw) > 16384 || req.URL.Path != c.Path || json.Compact(&actual, raw) != nil || json.Compact(&recorded, c.Request) != nil || !bytes.Equal(actual.Bytes(), recorded.Bytes()) {
		return nil, errors.New("recorded request mismatch")
	}
	if c.Canceled {
		if len(c.Response) != 0 {
			return nil, errors.New("canceled script has response")
		}
		r.cancel()
		return nil, req.Context().Err()
	}
	if c.NetworkFailure {
		return nil, io.ErrUnexpectedEOF
	}
	header := http.Header{"Content-Type": []string{"application/json"}}
	id := s.Metadata.Attempts[0].ProviderRequestID
	header.Set("x-request-id", id)
	header.Set("x-typesafe-request-id", id)
	status := c.Status
	if status == 0 {
		status = 200
	}
	var body io.ReadCloser = io.NopCloser(bytes.NewReader(c.Response))
	if c.ReadFailure {
		body = &failedReplayBody{Reader: bytes.NewReader(c.Response)}
	}
	return &http.Response{StatusCode: status, Header: header, Body: body, Request: req}, nil
}
func verifyFixture(p *chineseeval.Plan, r chineseeval.Report, raw []byte) error {
	var captures []fixtureCapture
	if json.Unmarshal(raw, &captures) != nil || len(captures) != r.Budget.Attempts {
		return errors.New("fixture capture mismatch")
	}
	return verifyCapturedScripts(p, r, captures)
}
func verifyCapturedScripts(p *chineseeval.Plan, r chineseeval.Report, captures []fixtureCapture) error {
	bad := errors.New("recorded public replay mismatch")
	replay := &replayTransport{captures: captures}
	for _, s := range r.Samples {
		if len(s.Metadata.Attempts) > 0 {
			replay.samples = append(replay.samples, s)
		}
	}
	client, transport, err := newClient(p, "synthetic-replay-credential", newObserver(), replay, nil, nil)
	if err != nil {
		return err
	}
	defer transport.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Config().TimeoutSeconds)*time.Second)
	defer cancel()
	for _, want := range replay.samples {
		// Calls cancelled before an Attempt have no wire to replay. Report and
		// Observer already reconcile those calls independently, at their positions.
		input, err := p.Request(want.ID)
		if err != nil {
			return err
		}
		callCtx, callCancel := context.WithTimeout(ctx, time.Duration(p.Config().CallTimeoutSeconds)*time.Second)
		replay.cancel = callCancel
		actual, callErr := client.Classify(callCtx, ai.CallScope{TenantID: tenant, RequestID: requestid.New(), JobID: "evidence-replay"}, ai.Target{BindingID: binding, ModelID: p.Config().Model}, input, nil)
		callCancel()
		interrupted := want.Code == ai.CodeCanceled || want.Code == ai.CodeDeadlineExceeded
		if !interrupted {
			model, _ := actual.ResponseModel.Get()
			if model != want.ResponseModel || !replayedAnswer(actual, want, p.Config().QuestionKey) {
				return bad
			}
			var failure *ai.Error
			if want.Code == "" && callErr != nil || want.Code != "" && (!errors.As(callErr, &failure) || failure.Code != want.Code || failure.Phase != want.Phase) {
				return bad
			}
		}
		if len(actual.Metadata.Attempts) != 1 {
			return bad
		}
		a, b := actual.Metadata.Attempts[0], want.Metadata.Attempts[0]
		if !interrupted && (a.HTTPStatus != b.HTTPStatus || a.Code != b.Code) || a.UsageReporting != b.UsageReporting || !reflect.DeepEqual(a.Usage, b.Usage) {
			return bad
		}
	}
	if replay.next != len(captures) {
		return bad
	}
	return nil
}
func replayedAnswer(actual ai.ClassifierResult, want chineseeval.SampleResult, key string) bool {
	if want.Answer != nil {
		a, ok := actual.Answers[key].(ai.ChoiceAnswer)
		return ok && reflect.DeepEqual(want.Answer, &chineseeval.Answer{Type: "choice", Choice: a.Choice, Probabilities: a.Probabilities, Confidence: a.Confidence})
	}
	if want.Status == "refused" {
		_, ok := actual.Answers[key].(ai.RefusalAnswer)
		return ok
	}
	return len(actual.Answers) == 0
}

type failedReplayBody struct{ *bytes.Reader }

func (b *failedReplayBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}
func (*failedReplayBody) Close() error { return nil }
