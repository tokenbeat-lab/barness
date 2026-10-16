package ai

import (
	"context"
	"encoding/json"
	"slices"
)

// ClassifierRequest evaluates one state against all named questions in one
// logical request. State is a JSON string, object or array, or an explicitly
// image-referenced null root on a model supporting State value images.
type ClassifierRequest struct {
	State     json.RawMessage               `json:"state"`
	Questions map[string]ClassifierQuestion `json:"questions"`
	// Images is the single inline source collection. Empty retains text-only wire.
	Images []Image `json:"images,omitempty"`
	// ImageReferences is a complete usage declaration when nonempty. Omitted,
	// null or empty references use every source once as ordered context.
	ImageReferences []ImageReference `json:"imageReferences,omitempty"`
}

// ClassifierResult exists on both success and failure. Answers are published
// together with ImageInputs only after validation; failure retains usage and
// call metadata.
type ClassifierResult struct {
	Answers       map[string]ClassifierAnswer `json:"answers"`
	ImageInputs   []ClassifierImageInput      `json:"imageInputs,omitempty"`
	ResponseModel Nullable[string]            `json:"responseModel,omitzero"`
	Usage         Usage                       `json:"usage"`
	StopReason    StopReason                  `json:"stopReason"`
	ErrorMessage  string                      `json:"errorMessage,omitempty"`
	Metadata      CallMetadata                `json:"metadata"`
}

// TypeSafeOptions selects experimental image transport without granting image authority.
// Nil and the zero value retain the text protocol defaults.
type TypeSafeOptions struct {
	ExperimentalImages bool `json:"experimentalImages,omitempty"`
}

func (TypeSafeOptions) api() API         { return APITypeSafeSystemOne }
func (o TypeSafeOptions) clone() Options { return o }
func (TypeSafeOptions) validate() string { return "" }

// OpenAIDecisionsOptions selects the Decisions protocol defaults. The text
// classifier has no generation or authority options.
type OpenAIDecisionsOptions struct{}

func (OpenAIDecisionsOptions) api() API         { return APIOpenAIDecisions }
func (o OpenAIDecisionsOptions) clone() Options { return o }
func (OpenAIDecisionsOptions) validate() string { return "" }

// Classify synchronously evaluates a set of questions. Options must belong to
// the binding protocol; SimpleOptions cannot be passed to this full entry.
func (c *Client) Classify(ctx context.Context, scope CallScope, target Target, req ClassifierRequest, opts Options) (ClassifierResult, error) {
	return c.classify(ctx, scope, target, req, opts, Hooks{})
}

func (h *HookedClient) Classify(ctx context.Context, scope CallScope, target Target, req ClassifierRequest, opts Options) (ClassifierResult, error) {
	return h.client.classify(ctx, scope, target, req, opts, h.hooks)
}

func (c *Client) classify(ctx context.Context, scope CallScope, target Target, req ClassifierRequest, opts Options, hooks Hooks) (ClassifierResult, error) {
	ctx, r := c.beginCall(ctx, scope, target, OperationClassifier)
	defer r.close()
	res := ClassifierResult{}
	failure := validateScope(scope)
	if failure == nil {
		failure = contextError(ctx, PhaseScope)
	}
	if failure == nil {
		failure = c.policy.checkClassifierImageSizes(req.Images, PhaseScope)
	}
	if failure == nil {
		failure = c.policy.checkClassifierReferenceSizes(ctx, req, c.classifierInlineImages(target))
	}
	if failure == nil && (len(req.Images) > 0 || len(req.ImageReferences) > 0) {
		failure = c.policy.byteLimits().checkClassifierInput(req)
	}
	var plan classifierImagePlan
	if failure == nil {
		plan, failure = planClassifierImages(req)
	}
	if failure == nil {
		failure = c.policy.Classifier.check(req, plan)
	}
	if failure == nil {
		failure = plan.checkValues(req)
	}
	if failure == nil {
		failure = c.policy.byteLimits().checkClassifierInput(req)
	}
	if failure == nil && len(req.Images) > 0 {
		failure = checkClassifierImageEncoding(ctx, req.Images, PhaseScope)
	}
	// Bound known sizes before copying any mutable input.
	if failure == nil {
		req = req.clone()
		// Snapshot the new option before any external resolver can mutate a
		// pointer supplied by the Host. Typed nil retains its existing rejection.
		if o, ok := opts.(*TypeSafeOptions); ok && o != nil {
			opts = o.clone()
		}
		failure = c.executeClassifier(ctx, r, target, req, plan, opts, hooks, &res)
	}
	res.StopReason = StopReasonStop
	if failure != nil {
		res.Answers = nil
		res.ImageInputs = nil
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

func (r ClassifierRequest) clone() ClassifierRequest {
	out := ClassifierRequest{State: slices.Clone(r.State), Questions: make(map[string]ClassifierQuestion, len(r.Questions)), Images: slices.Clone(r.Images), ImageReferences: slices.Clone(r.ImageReferences)}
	for key, q := range r.Questions {
		switch q := q.(type) {
		case ChoiceQuestion:
			q.Instructions = slices.Clone(q.Instructions)
			q.Criteria = cloneJSONValues(q.Criteria)
			out.Questions[key] = q
		case ScoreQuestion:
			q.Instructions = slices.Clone(q.Instructions)
			q.Criteria = slices.Clone(q.Criteria)
			for i, v := range q.Criteria {
				q.Criteria[i] = slices.Clone(v)
			}
			out.Questions[key] = q
		case BoolQuestion:
			q.Instructions = slices.Clone(q.Instructions)
			if q.Criteria != nil {
				q.Criteria = &BoolCriteria{True: slices.Clone(q.Criteria.True), False: slices.Clone(q.Criteria.False)}
			}
			out.Questions[key] = q
		}
	}
	return out
}

func (c *Client) executeClassifier(ctx context.Context, r *callRuntime, target Target, req ClassifierRequest, plan classifierImagePlan, opts Options, hooks Hooks, res *ClassifierResult) *Error {
	i, failure := r.resolve(ctx, target)
	if failure != nil {
		return failure
	}
	model := c.catalog.ClassifierModels[i].clone()
	for _, ref := range plan {
		location := ClassifierImageLocationContext
		if ref.Placement == ImagePlacementValue {
			parts, _ := imagePointerParts(ref.Pointer) // scope already resolved the target
			if model.API != APITypeSafeSystemOne || parts[0] != "state" {
				return newError(CodeInvalidRequest, PhaseCapability, "classifier protocol does not support this image location")
			}
			location = ClassifierImageLocationState
		}
		// Omitted Decisions locations retain the existing context default.
		if !(model.API == APIOpenAIDecisions && len(model.Capabilities.ImageLocations) == 0) && !slices.Contains(model.Capabilities.ImageLocations, location) {
			return newError(CodeInvalidRequest, PhaseCapability, "classifier model does not support this image location")
		}
	}
	if problem := classifierImageCapability(req.Images, model); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if model.API == APIOpenAIDecisions && len(plan) > classifierImageLimit(model) {
		return newError(CodeInvalidRequest, PhaseCapability, "expanded classifier images exceed model or protocol capability")
	}
	if len(req.Images) > 0 && !r.binding.AllowClassifierImages {
		return newError(CodeTenantDenied, PhaseCapability, "binding does not allow classifier images")
	}
	if failure := validateOptions(opts, model.API); failure != nil {
		return failure
	}
	if opts != nil {
		opts = opts.clone()
	}
	if problem := req.checkCapabilities(model.Capabilities); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if model.API == APIOpenAIDecisions {
		if !c.classifierInlineImages(target) {
			if failure := c.policy.checkClassifierReferenceSizes(ctx, req, true); failure != nil {
				return failure
			}
		}
		if problem := decisionsTextCapability(req); problem != "" {
			return newError(CodeInvalidRequest, PhaseCapability, problem)
		}
	}
	if model.API == APITypeSafeSystemOne && typeSafeImagesSelected(opts) {
		if c.policy.Classifier.Images == nil || !slices.Contains(model.Input, ModalityImage) || len(model.Capabilities.ImageLocations) == 0 {
			return newError(CodeInvalidRequest, PhaseCapability, "Jev image experiment requires enabled image policy and model capabilities")
		}
		if !r.binding.AllowClassifierImages {
			return newError(CodeTenantDenied, PhaseCapability, "binding does not allow classifier images")
		}
	}
	var typeSafeBody []byte
	if model.API == APITypeSafeSystemOne && len(req.Images) > 0 {
		if !typeSafeImagesSelected(opts) {
			return newError(CodeInvalidRequest, PhaseCapability, "Jev images require explicit experimental selection")
		}
		typeSafeBody, failure = c.prepareTypeSafeImages(ctx, model, req, plan)
		if failure != nil {
			return failure
		}
	}
	cred, failure := r.pin(ctx, model.ID, hooks)
	if failure != nil {
		return failure
	}
	// Protocol conversion stays independent; both paths own the same pinned
	// call runtime and unary transport rather than a dynamic registry.
	switch model.API {
	case APITypeSafeSystemOne:
		return c.classifyTypeSafe(ctx, r, cred, model, req, opts, typeSafeBody, res)
	case APIOpenAIDecisions:
		return c.classifyDecisions(ctx, r, cred, model, req, plan, res)
	default:
		return newError(CodeInvalidRequest, PhaseCapability, "unsupported classifier API")
	}
}
