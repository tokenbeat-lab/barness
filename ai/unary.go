package ai

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
)

// sendUnary is the shared, pinned HTTP lifetime of classification and image
// calls. Protocols own encoding and parsing; the one Client owns networking,
// retries and bounded bodies. A successful response is never replayed.
func (c *Client) sendUnary(ctx context.Context, r *callRuntime, header http.Header, body []byte, path string, failures httpFailures) (*http.Response, *Error) {
	var response *http.Response
	failure := r.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.binding.Endpoint, "/")+path, bytes.NewReader(body))
		if err != nil {
			return attemptOutcome{failure: newError(CodeInvalidRequest, PhaseRequest, "request could not be built")}
		}
		for key, values := range header {
			req.Header[key] = slices.Clone(values)
		}
		res, err := markConnectionFailures(req, c.http.Do)
		if err != nil {
			closeBody(res)
			return failures.request(ctx, err, res)
		}
		c.policy.byteLimits().limitBody(res, bodyJSON)
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			defer closeBody(res)
			return failures.status(res)
		}
		response = res
		return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: res.Header.Get(failures.requestIDHeader)}
	})
	return response, failure
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
	return newError(CodeTransport, PhaseResponse, "response body could not be read")
}
