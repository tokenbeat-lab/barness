package ai

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func (c *Client) classifyTypeSafe(ctx context.Context, r *callRuntime, cred Credential, model ClassifierModel, req ClassifierRequest, opts Options, body []byte, out *ClassifierResult) *Error {
	merged := http.Header{"Authorization": []string{"Bearer " + cred.APIKey.reveal()}, "Content-Type": []string{"application/json"}, "User-Agent": []string{userAgent}}
	if typeSafeImagesSelected(opts) {
		merged.Set(typeSafeImagesHeader, "true")
	}
	header, failure := r.hooks.headers(ctx, merged)
	if failure != nil {
		return failure
	}
	if failure := checkTypeSafeHeaders(header, typeSafeImagesSelected(opts)); failure != nil {
		return failure
	}
	if body == nil {
		var err error
		body, err = encodeTypeSafeRequest(model.ID, req)
		if err != nil {
			return newError(CodeInvalidRequest, PhaseRequest, "classifier request could not be encoded")
		}
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
		if _, present := final["assets"]; present && (!r.binding.AllowClassifierImages || !typeSafeImagesSelected(opts)) {
			return newError(CodeTenantDenied, PhaseRequest, "Jev images are not authorized")
		}
		return c.checkTypeSafePayload(final, limits.request)
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
	if len(final.Assets) > 0 && int64(len(body)) > typeSafeBodyBytes {
		return limitFailure(PhaseRequest, "Jev.MaxRequestBytes", typeSafeBodyBytes)
	}
	images, failure := c.freezeTypeSafeImages(ctx, final, model, r.binding.AllowClassifierImages, typeSafeImagesSelected(opts), typeSafeContextProjection(req, final, model.Capabilities.ImageLocations))
	if failure != nil {
		return failure
	}
	imageInputs, failure := classifierImageInputs(ctx, images.images)
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
	failure = decodeTypeSafeResponse(raw, model, r.initial, final, images.indices, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	if failure == nil {
		out.ImageInputs = imageInputs
	}
	return failure
}

func typeSafeHTTPError(res *http.Response, body []byte) (string, string) {
	// Only the tenant receives this bounded, redacted diagnostic body. The
	// Observer records the classification, status and vendor id alone.
	msg := res.Status + ": " + string(body)
	return msg, msg
}

// Header names were canonicalized by the shared Hook boundary. Images and
// Long Context are protocol authority; callbacks cannot select experiments.
func checkTypeSafeHeaders(header http.Header, selected bool) *Error {
	values := header.Values(typeSafeImagesHeader)
	if (selected && (len(values) != 1 || values[0] != "true")) || (!selected && len(values) != 0) {
		return newError(CodeTenantDenied, PhaseRequest, "header transform may not change the Jev image experiment")
	}
	for name, values := range header {
		if strings.HasPrefix(strings.ToLower(name), "typesafe-danger-zone-") && (name != typeSafeImagesHeader || len(values) != 1 || values[0] != "true") {
			return newError(CodeTenantDenied, PhaseRequest, "unsupported Jev experiment header")
		}
	}
	return nil
}
