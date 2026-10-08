package ai

import (
	"context"
	"encoding/json"
	"slices"
)

// ImagesRequest is a prompt followed by ordered, inline reference images.
// References use the existing input Image, never a path, URL or file ID.
type ImagesRequest struct {
	Prompt          string  `json:"prompt"`
	ReferenceImages []Image `json:"referenceImages,omitempty"`
}

// ImageOptions is the closed set of image protocol options. SimpleOptions
// and chat options cannot be passed to GenerateImages. Nil uses defaults.
type ImageOptions interface {
	Options
	imageOptions()
}

// ImageOutput is an independent, closed set of ordered image result blocks.
// It has no chat assistant or native-state semantics.
type ImageOutput interface{ imageOutput() }

type ImageOutputText struct {
	Text string `json:"text"`
}
type ImageOutputImage struct {
	Data     string `json:"data"` // Strict standard base64; never an external reference.
	MimeType string `json:"mimeType"`
}

func (ImageOutputText) imageOutput()  {}
func (ImageOutputImage) imageOutput() {}
func (b ImageOutputText) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}{"text", b.Text})
}
func (b ImageOutputImage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string `json:"type"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType"`
	}{"image", b.Data, b.MimeType})
}

// ImagesResult is returned on success and failure. Content is published only
// after all images pass validation; usage and metadata survive failures.
type ImagesResult struct {
	Content       []ImageOutput    `json:"content"`
	ResponseID    Nullable[string] `json:"responseId,omitzero"`
	ResponseModel Nullable[string] `json:"responseModel,omitzero"`
	Usage         Usage            `json:"usage"`
	StopReason    StopReason       `json:"stopReason"`
	ErrorMessage  string           `json:"errorMessage,omitempty"`
	Metadata      CallMetadata     `json:"metadata"`
}

// GenerateImages synchronously generates the complete image collection.
// It honors cancellation and never emits deltas or stores continuation state.
func (c *Client) GenerateImages(ctx context.Context, scope CallScope, target Target, req ImagesRequest, opts ImageOptions) (ImagesResult, error) {
	return c.generateImages(ctx, scope, target, req, opts, Hooks{})
}
func (h *HookedClient) GenerateImages(ctx context.Context, scope CallScope, target Target, req ImagesRequest, opts ImageOptions) (ImagesResult, error) {
	return h.client.generateImages(ctx, scope, target, req, opts, h.hooks)
}
func (c *Client) generateImages(ctx context.Context, scope CallScope, target Target, req ImagesRequest, opts ImageOptions, hooks Hooks) (ImagesResult, error) {
	ctx, r := c.beginCall(ctx, scope, target, OperationImage)
	defer r.close()
	res := ImagesResult{}
	options := OpenAIImagesOptions{}
	switch o := opts.(type) {
	case OpenAIImagesOptions:
		options = o
	case *OpenAIImagesOptions:
		if o != nil {
			options = *o
		}
	}
	failure := validateScope(scope)
	if failure == nil {
		failure = contextError(ctx, PhaseScope)
	}
	if failure == nil {
		failure = c.policy.checkImageInput(ctx, req, options.Mask, PhaseScope)
	}
	if failure == nil {
		req.ReferenceImages = slices.Clone(req.ReferenceImages)
		// Host resolvers run after this snapshot and cannot change entry options.
		if opts != nil {
			switch o := opts.(type) {
			case OpenAIImagesOptions:
				opts = o.clone().(OpenAIImagesOptions)
			case *OpenAIImagesOptions:
				if o != nil {
					opts = o.clone().(OpenAIImagesOptions)
				}
			}
		}
		failure = c.executeImages(ctx, r, target, req, opts, hooks, &res)
	}
	res.StopReason = StopReasonStop
	if failure != nil {
		res.Content = nil
		res.StopReason = StopReasonError
		res.ErrorMessage = failure.Message
		if failure.aborts() {
			res.StopReason = StopReasonAborted
		}
	}
	r.finish(callOutcome{stop: res.StopReason, usage: res.Usage, failure: failure})
	res.Metadata = r.meta
	if failure != nil {
		return res, failure
	}
	return res, nil
}
func (c *Client) executeImages(ctx context.Context, r *callRuntime, target Target, req ImagesRequest, opts ImageOptions, hooks Hooks, out *ImagesResult) *Error {
	i, failure := r.resolve(ctx, target)
	if failure != nil {
		return failure
	}
	model := c.catalog.ImageModels[i].clone()
	options := OpenAIImagesOptions{}
	if opts != nil {
		switch o := opts.(type) {
		case OpenAIImagesOptions:
			options = o
		case *OpenAIImagesOptions:
			if o == nil {
				return newError(CodeInvalidRequest, PhaseCapability, "image options must not be a typed nil")
			}
			options = *o
		default:
			return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
		}
	}
	if failure := c.policy.checkImageCapabilities(req, options, model.Capabilities, PhaseCapability); failure != nil {
		return failure
	}
	cred, failure := r.pin(ctx, model.ID, hooks)
	if failure != nil {
		return failure
	}
	return c.generateOpenAIImages(ctx, r, cred, model, req, options, out)
}
