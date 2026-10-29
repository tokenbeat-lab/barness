package ai

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) executeXAIImages(ctx context.Context, r *callRuntime, model ImageModel, req ImagesRequest, opts ImageOptions, hooks Hooks, out *ImagesResult) *Error {
	options := XAIImagesOptions{}
	switch o := opts.(type) {
	case nil:
	case XAIImagesOptions:
		options = o
	case *XAIImagesOptions:
		if o == nil {
			return newError(CodeInvalidRequest, PhaseCapability, "image options must not be a typed nil")
		}
		options = *o
	default:
		return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
	}
	if problem := options.check(model.Capabilities, c.policy.Image, len(req.ReferenceImages)); problem != "" {
		return imageInputFailure(PhaseCapability, problem)
	}
	if failure := c.policy.checkXAIImageCount(len(req.ReferenceImages), model.Capabilities, PhaseCapability); failure != nil {
		return failure
	}
	wire := xaiImagesRequest{Model: model.ID, Prompt: req.Prompt, ResponseFormat: "b64_json", N: options.N, AspectRatio: options.AspectRatio, Resolution: options.Resolution, Quality: options.Quality}
	extra := int64(0)
	for _, img := range req.ReferenceImages {
		extra += int64(len(img.MimeType)+len(img.Data)+len("data:;base64,")) - 1
		wire.Images = append(wire.Images, xaiImageResource{Type: "image_url", URL: "A"})
	}
	if len(wire.Images) == 1 {
		wire.Image = &wire.Images[0]
		wire.Images = nil
	}
	if failure := c.policy.checkXAIImagesEncoding(wire, extra); failure != nil {
		return failure
	}
	if failure := c.policy.checkImageInput(ctx, req, nil, PhaseScope); failure != nil {
		return failure
	}
	if wire.Image != nil {
		img := req.ReferenceImages[0]
		wire.Image.URL = "data:" + img.MimeType + ";base64," + img.Data
	} else {
		for i, img := range req.ReferenceImages {
			wire.Images[i].URL = "data:" + img.MimeType + ";base64," + img.Data
		}
	}
	cred, failure := r.pin(ctx, model.ID, hooks)
	if failure != nil {
		return failure
	}
	return c.generateXAIImages(ctx, r, cred, model, wire, out)
}
func (c *Client) generateXAIImages(ctx context.Context, r *callRuntime, cred Credential, model ImageModel, wire xaiImagesRequest, out *ImagesResult) *Error {
	headers, failure := r.hooks.headers(ctx, http.Header{"Authorization": {"Bearer " + cred.APIKey.reveal()}, "Content-Type": {"application/json"}, "User-Agent": {userAgent}})
	if failure != nil {
		return failure
	}
	body, err := json.Marshal(wire)
	if err != nil {
		return imageInputFailure(PhaseRequest, "Imagine request could not be encoded")
	}
	editing := wire.Image != nil || len(wire.Images) > 0
	body, failure = r.hooks.payload(ctx, body, func(p map[string]any) *Error { return c.policy.checkXAIImagesPayload(p, model, editing) })
	if failure != nil {
		return failure
	}
	if failure := c.policy.byteLimits().checkRequestBody(body); failure != nil {
		return failure
	}
	final, failure := decodeXAIImagesRequest(ctx, body, model, &c.policy, editing)
	if failure != nil {
		return failure
	}
	path := "/images/generations"
	if editing {
		path = "/images/edits"
	}
	failures := httpFailures{apiKey: cred.APIKey, clock: c.clock, requestIDHeader: "x-request-id", classifyStatus: openAIImagesHTTPCode, describe: func(res *http.Response, raw []byte) (string, string) {
		return describeHTTPError("xAI Images API error", res.StatusCode, raw)
	}}
	response, failure := c.sendUnary(ctx, r, headers, body, path, failures)
	if failure != nil {
		return failure
	}
	defer closeBody(response)
	if failure := r.hooks.response(ctx, response); failure != nil {
		return failure
	}
	if response.StatusCode != http.StatusOK {
		return imageProtocolFailure()
	}
	failure = decodeXAIImagesResponse(ctx, response.Body, model, final, *c.policy.Image, r.initial, out)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	return failure
}
