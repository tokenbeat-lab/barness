package ai

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) executeGoogleImages(ctx context.Context, r *callRuntime, model ImageModel, req ImagesRequest, opts ImageOptions, hooks Hooks, out *ImagesResult) *Error {
	options := GoogleImagesOptions{}
	switch o := opts.(type) {
	case nil:
	case GoogleImagesOptions:
		options = o
	case *GoogleImagesOptions:
		if o == nil {
			return newError(CodeInvalidRequest, PhaseCapability, "image options must not be a typed nil")
		}
		options = *o
	default:
		return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
	}
	if problem := googleImagesTargetProblem(r.binding.Endpoint, model.ID); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if problem := options.check(model.Capabilities); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if failure := c.policy.checkGoogleImageCount(len(req.ReferenceImages), model.Capabilities, PhaseCapability); failure != nil {
		return failure
	}
	wire := googleImagesRequest{Model: model.ID, Store: false, Input: []googleImageInput{{Type: "text", Text: req.Prompt}}, Format: googleImageFormat{Type: "image", Delivery: "inline", AspectRatio: options.AspectRatio, ImageSize: options.ImageSize}}
	for _, img := range req.ReferenceImages {
		wire.Input = append(wire.Input, googleImageInput{Type: "image", MimeType: img.MimeType, Data: img.Data})
	}
	if failure := c.policy.checkGoogleImagesEncoding(wire); failure != nil {
		return failure
	}
	cred, failure := r.pin(ctx, model.ID, hooks)
	if failure != nil {
		return failure
	}
	return c.generateGoogleImages(ctx, r, cred, model, wire, out)
}

// Interactions stays on the Client's sole HTTP path. Unlike generateContent
// chat, this synchronous image operation has a response metadata callback.
func (c *Client) generateGoogleImages(ctx context.Context, r *callRuntime, cred Credential, model ImageModel, wire googleImagesRequest, out *ImagesResult) *Error {
	header, failure := r.hooks.headers(ctx, http.Header{"X-Goog-Api-Key": {cred.APIKey.reveal()}, "Content-Type": {"application/json"}, "User-Agent": {userAgent}})
	if failure != nil {
		return failure
	}
	body, err := json.Marshal(wire)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "Google image request could not be encoded")
	}
	if failure := c.policy.checkGoogleImagesBody(body); failure != nil {
		return failure
	}
	body, failure = r.hooks.payload(ctx, body, func(payload map[string]any) *Error {
		return c.policy.checkGoogleImagesPayload(payload, model.Capabilities)
	})
	if failure != nil {
		return failure
	}
	if failure := c.policy.checkGoogleImagesBody(body); failure != nil {
		return failure
	}
	final, failure := decodeGoogleImagesRequest(body, model, &c.policy)
	if failure != nil {
		return failure
	}
	if failure := c.policy.checkGoogleImageCount(len(final.Input)-1, model.Capabilities, PhaseRequest); failure != nil {
		return failure
	}
	if failure := c.policy.checkImageInput(ctx, final.imagesInput(), nil, PhaseRequest); failure != nil {
		return failure
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: "x-request-id", describe: geminiHTTPError}
	response, failure := c.sendUnary(ctx, r, header, body, "/interactions", failures)
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
	failure = decodeGoogleImagesResponse(ctx, response.Body, model, *c.policy.Image, r.initial, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}
func (p *ResourcePolicy) checkGoogleImagesBody(body []byte) *Error {
	if failure := p.byteLimits().checkRequestBody(body); failure != nil {
		return failure
	}
	if int64(len(body)) > googleImagesRequestLimit {
		return limitFailure(PhaseRequest, "Google.MaxInlineRequestBytes", googleImagesRequestLimit)
	}
	return nil
}
