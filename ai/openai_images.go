package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"unicode/utf8"
)

func (c *Client) generateOpenAIImages(ctx context.Context, r *callRuntime, cred Credential, model ImageModel, req ImagesRequest, opts OpenAIImagesOptions, out *ImagesResult) *Error {
	header, failure := r.hooks.headers(ctx, http.Header{"Authorization": []string{"Bearer " + cred.APIKey.reveal()}, "Content-Type": []string{"application/json"}, "User-Agent": []string{userAgent}})
	if failure != nil {
		return failure
	}
	body, err := json.Marshal(openAIImagesRequest{Model: model.ID, Prompt: req.Prompt, N: opts.N, Size: opts.Size, Quality: opts.Quality, Background: opts.Background, OutputFormat: opts.OutputFormat, OutputCompression: opts.OutputCompression, Moderation: opts.Moderation})
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "image request could not be encoded")
	}
	limits := c.policy.byteLimits()
	if failure := limits.checkRequestBody(body); failure != nil {
		return failure
	}
	body, failure = r.hooks.payload(ctx, body, func(final map[string]any) *Error { return c.policy.checkOpenAIImagesPayload(final, model.ID) })
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
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid image request")
	}
	if final.Model != model.ID {
		return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	if !utf8.ValidString(final.Prompt) || utf8.RuneCountInString(final.Prompt) < 1 || utf8.RuneCountInString(final.Prompt) > 32000 {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid image prompt")
	}
	if problem := final.options().check(model.Capabilities, c.policy.Image); problem != "" {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced invalid image options")
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: requestIDHeaderOf(model.Provider), classifyStatus: openAIImagesHTTPCode, describe: func(res *http.Response, raw []byte) (string, string) {
		return describeHTTPError("OpenAI Images API error", res.StatusCode, raw)
	}}
	response, failure := c.sendUnary(ctx, r, header, body, "/images/generations", failures)
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
