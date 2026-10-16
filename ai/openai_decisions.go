package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Client) classifyDecisions(ctx context.Context, r *callRuntime, cred Credential, model ClassifierModel, req ClassifierRequest, plan classifierImagePlan, out *ClassifierResult) *Error {
	header, failure := r.hooks.headers(ctx, http.Header{"Authorization": []string{"Bearer " + cred.APIKey.reveal()}, "Content-Type": []string{"application/json"}, "User-Agent": []string{userAgent}})
	if failure != nil {
		return failure
	}
	body, err := encodeDecisionsRequest(model.ID, req, plan)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "Decisions request could not be encoded")
	}
	if failure := c.policy.byteLimits().checkRequestBody(body); failure != nil {
		return failure
	}
	if _, failure := c.freezeDecisionsRequest(ctx, body, model, r.binding.AllowClassifierImages); failure != nil {
		return failure
	}
	body, failure = r.hooks.payload(ctx, body, func(final map[string]any) *Error { return c.checkDecisionsPayload(final, model) })
	if failure != nil {
		return failure
	}
	if failure := c.policy.byteLimits().checkRequestBody(body); failure != nil {
		return failure
	}
	final, failure := c.freezeDecisionsRequest(ctx, body, model, r.binding.AllowClassifierImages)
	if failure != nil {
		return failure
	}
	res, failure := c.sendUnary(ctx, r, header, body, "/decisions", httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: "x-request-id", describe: decisionsHTTPError})
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
	failure = decodeDecisionsResponse(raw, model, r.initial, final, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}

func decisionsHTTPError(res *http.Response, body []byte) (string, string) {
	// The shared HTTP lifetime bounds body before this conversion and redacts
	// the selected diagnostic afterwards. Only an unambiguous error.message
	// is useful to the tenant; gateway HTML and other contents stay private.
	message := res.Status + ": Decisions request failed"
	fields, duplicate, err := unaryJSONEnvelope(body)
	if err == nil && !duplicate {
		detail, duplicate, err := unaryJSONEnvelope(fields["error"])
		if err == nil && !duplicate {
			var text string
			if json.Unmarshal(detail["message"], &text) == nil && text != "" {
				message = res.Status + ": " + text
			}
		}
	}
	return message, message
}
