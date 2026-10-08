package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
)

func (c *Client) classifyTypeSafe(ctx context.Context, r *callRuntime, cred Credential, model ClassifierModel, req ClassifierRequest, out *ClassifierResult) *Error {
	merged := http.Header{"Authorization": []string{"Bearer " + cred.APIKey.reveal()}, "Content-Type": []string{"application/json"}, "User-Agent": []string{userAgent}}
	header, failure := r.hooks.headers(ctx, merged)
	if failure != nil {
		return failure
	}
	body, err := encodeTypeSafeRequest(model.ID, req)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "classifier request could not be encoded")
	}
	body, failure = r.hooks.payload(ctx, body, func(final map[string]any) string {
		if final["model"] != model.ID {
			return "payload callback may not change the authorized model"
		}
		for _, key := range []string{"endpoint", "api_key", "headers", "provider", "api", "operation", "tools", "credential"} {
			if _, ok := final[key]; ok {
				return "payload callback may not add authority fields"
			}
		}
		return ""
	})
	if failure != nil {
		return failure
	}
	// Freeze and decode the final payload, independently of the callback's
	// mutable maps. Its final questions are the sole answer-validation source.
	var finalModel string
	if err := decodeTypeSafeRequest(body, &finalModel, &req); err != nil {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid classifier request")
	}
	if finalModel != model.ID {
		return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	if failure := c.policy.Classifier.check(req); failure != nil {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid classifier request")
	}
	if problem := req.checkCapabilities(model.Capabilities); problem != "" {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback exceeded classifier model capabilities")
	}
	limits := c.policy.byteLimits()
	if failure := limits.checkRequestBody(body); failure != nil {
		return failure
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: "x-typesafe-request-id", describe: typeSafeHTTPError}
	var res *http.Response
	failure = r.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.binding.Endpoint, "/")+"/systemone", bytes.NewReader(body))
		if err != nil {
			return attemptOutcome{failure: newError(CodeInvalidRequest, PhaseRequest, "request could not be built")}
		}
		for key, values := range header {
			req.Header[key] = slices.Clone(values)
		}
		response, err := c.http.Do(req)
		if err != nil {
			closeBody(response)
			return failures.request(ctx, err, response)
		}
		limits.limitBody(response, bodyJSON)
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			defer closeBody(response)
			return failures.status(response)
		}
		res = response
		return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: res.Header.Get("x-typesafe-request-id")}
	})
	if failure != nil {
		return failure
	}
	defer closeBody(res)
	if failure := r.hooks.response(ctx, res); failure != nil {
		return failure
	}
	raw, err := io.ReadAll(res.Body)
	if failure := contextError(ctx, PhaseResponse); failure != nil {
		return failure
	}
	if err != nil {
		return unaryReadFailure(err)
	}
	failure = decodeTypeSafeResponse(raw, model, r.initial, req, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}

// The public bool discriminator never leaks onto the native wire. Conversion
// is explicit in both directions and keeps raw numeric descriptions intact.
func encodeTypeSafeRequest(model string, req ClassifierRequest) ([]byte, error) {
	questions := make(map[string]json.RawMessage, len(req.Questions))
	for key, q := range req.Questions {
		var raw []byte
		var err error
		if boolean, ok := q.(BoolQuestion); ok {
			raw, err = json.Marshal(struct {
				Type         string          `json:"type"`
				Instructions json.RawMessage `json:"instructions"`
				Criteria     *BoolCriteria   `json:"criteria,omitempty"`
			}{"noul", boolean.Instructions, boolean.Criteria})
		} else {
			raw, err = json.Marshal(q)
		}
		if err != nil {
			return nil, err
		}
		questions[key] = raw
	}
	return json.Marshal(struct {
		Model     string                     `json:"model"`
		State     json.RawMessage            `json:"state"`
		Questions map[string]json.RawMessage `json:"questions"`
	}{model, req.State, questions})
}

func decodeTypeSafeRequest(body []byte, model *string, req *ClassifierRequest) error {
	var wire struct {
		Model     string          `json:"model"`
		State     json.RawMessage `json:"state"`
		Questions json.RawMessage `json:"questions"`
	}
	if err := decodeClassifierJSON(body, &wire); err != nil {
		return err
	}
	var questions map[string]json.RawMessage
	if err := decodeClassifierJSON(wire.Questions, &questions); err != nil {
		return err
	}
	decoded, err := decodeClassifierQuestions(questions, true)
	if err != nil {
		return err
	}
	*req = ClassifierRequest{State: wire.State, Questions: decoded}
	*model = wire.Model
	return nil
}

func typeSafeHTTPError(res *http.Response, body []byte) (string, string) {
	// Only the tenant receives this bounded, redacted diagnostic body. The
	// Observer records the classification, status and vendor id alone.
	msg := res.Status + ": " + string(body)
	return msg, msg
}
func unaryReadFailure(err error) *Error {
	var limit *limitExceeded
	var expired *timeLimitExpired
	switch {
	case errors.As(err, &limit):
		return limit.failure(PhaseResponse)
	case errors.As(err, &expired):
		return expired.failure(PhaseResponse)
	}
	return newError(CodeTransport, PhaseResponse, "classifier response body could not be read")
}
