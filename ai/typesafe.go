package ai

import (
	"context"
	"io"
	"net/http"
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
	limits := c.policy.byteLimits()
	body, failure = r.hooks.payload(ctx, body, func(final map[string]any) *Error {
		if final["model"] != model.ID {
			return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
		}
		for _, key := range []string{"endpoint", "api_key", "headers", "provider", "api", "operation", "tools", "credential"} {
			if _, ok := final[key]; ok {
				return newError(CodeTenantDenied, PhaseRequest, "payload callback may not add authority fields")
			}
		}
		return c.policy.Classifier.checkPayload(final, limits.request)
	})
	if failure != nil {
		return failure
	}
	if failure := limits.checkRequestBody(body); failure != nil {
		return failure
	}
	final, failure := c.freezeTypeSafeRequest(body, model)
	if failure != nil {
		return failure
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: "x-typesafe-request-id", describe: typeSafeHTTPError}
	res, failure := c.sendUnary(ctx, r, header, body, "/systemone", failures)
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
	failure = decodeTypeSafeResponse(raw, model, r.initial, final, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}

func typeSafeHTTPError(res *http.Response, body []byte) (string, string) {
	// Only the tenant receives this bounded, redacted diagnostic body. The
	// Observer records the classification, status and vendor id alone.
	msg := res.Status + ": " + string(body)
	return msg, msg
}
