package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func (c *Client) generateOpenAIImages(ctx context.Context, r *callRuntime, cred Credential, model ImageModel, req ImagesRequest, opts OpenAIImagesOptions, out *ImagesResult) *Error {
	header, failure := r.hooks.headers(ctx, http.Header{"Authorization": []string{"Bearer " + cred.APIKey.reveal()}, "Content-Type": []string{"application/json"}, "User-Agent": []string{userAgent}})
	if failure != nil {
		return failure
	}
	wire := openAIImagesRequest{Model: model.ID, Prompt: req.Prompt, N: opts.N, Size: opts.Size, Quality: opts.Quality, Background: opts.Background, OutputFormat: opts.OutputFormat, OutputCompression: opts.OutputCompression, Moderation: opts.Moderation, InputFidelity: opts.InputFidelity}
	editing := len(req.ReferenceImages) > 0
	path := "/images/generations"
	if editing {
		path = "/images/edits"
		wire.Images = make([]openAIInlineImage, len(req.ReferenceImages))
		for i, img := range req.ReferenceImages {
			wire.Images[i] = openAIImageReference(img)
		}
	}
	if opts.Mask != nil {
		mask := openAIImageReference(*opts.Mask)
		wire.Mask = Value(mask)
	}
	body, err := json.Marshal(wire)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "image request could not be encoded")
	}
	limits := c.policy.byteLimits()
	if failure := limits.checkRequestBody(body); failure != nil {
		return failure
	}
	body, failure = r.hooks.payload(ctx, body, func(final map[string]any) *Error { return c.policy.checkOpenAIImagesPayload(final, model, editing) })
	if failure != nil {
		return failure
	}
	if failure := limits.checkRequestBody(body); failure != nil {
		return failure
	}
	// Decode the independently frozen request. Result MIME/count validation
	// follows the final request rather than the caller's pre-hook options.
	final, err := decodeOpenAIImagesRequest(body)
	if err != nil {
		var failure *Error
		if errors.As(err, &failure) {
			return failure
		}
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid image request")
	}
	if final.Model != model.ID {
		return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	if failure := c.policy.checkOpenAIImagesFinal(ctx, final, model.Capabilities, editing); failure != nil {
		return failure
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: requestIDHeaderOf(model.Provider), classifyStatus: openAIImagesHTTPCode, describe: func(res *http.Response, raw []byte) (string, string) {
		return describeHTTPError("OpenAI Images API error", res.StatusCode, raw)
	}}
	response, failure := c.sendUnary(ctx, r, header, body, path, failures)
	if failure != nil {
		return failure
	}
	defer closeBody(response)
	if failure := r.hooks.response(ctx, response); failure != nil {
		return failure
	}
	if failure := contextError(ctx, PhaseResponse); failure != nil {
		return failure
	}
	failure = decodeOpenAIImagesResponse(ctx, response.Body, model, final.options(), *c.policy.Image, r.initial, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}

func openAIImagesHTTPCode(status int, body []byte) Code {
	var wire struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if (status == http.StatusBadRequest || status == http.StatusUnprocessableEntity) && json.Unmarshal(body, &wire) == nil && wire.Error.Code == "content_policy_violation" {
		return CodeUpstreamError
	}
	return codeForStatus(status)
}
