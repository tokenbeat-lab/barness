package ai

import (
	"context"
	"encoding/json"
	"slices"
)

// ClassifierRequest evaluates one state against all named questions in one
// logical request. State is a JSON string, object or array.
type ClassifierRequest struct {
	State     json.RawMessage               `json:"state"`
	Questions map[string]ClassifierQuestion `json:"questions"`
}

// ClassifierResult exists on both success and failure. Answers are published
// together only after validation; failure retains usage and call metadata.
type ClassifierResult struct {
	Answers       map[string]ClassifierAnswer `json:"answers"`
	ResponseModel Nullable[string]            `json:"responseModel,omitzero"`
	Usage         Usage                       `json:"usage"`
	StopReason    StopReason                  `json:"stopReason"`
	ErrorMessage  string                      `json:"errorMessage,omitempty"`
	Metadata      CallMetadata                `json:"metadata"`
}

// TypeSafeOptions is empty: nil and this value both select protocol defaults.
type TypeSafeOptions struct{}

func (TypeSafeOptions) api() API         { return APITypeSafeSystemOne }
func (o TypeSafeOptions) clone() Options { return o }
func (TypeSafeOptions) validate() string { return "" }

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
		failure = c.policy.Classifier.check(req)
	}
	if failure == nil {
		failure = c.policy.byteLimits().checkClassifierInput(req)
	}
	// Bound known sizes before copying any mutable input.
	if failure == nil {
		req = req.clone()
		failure = c.executeClassifier(ctx, r, target, req, opts, hooks, &res)
	}
	res.StopReason = StopReasonStop
	if failure != nil {
		res.Answers = nil
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
	out := ClassifierRequest{State: slices.Clone(r.State), Questions: make(map[string]ClassifierQuestion, len(r.Questions))}
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

func (c *Client) executeClassifier(ctx context.Context, r *callRuntime, target Target, req ClassifierRequest, opts Options, hooks Hooks, res *ClassifierResult) *Error {
	i, failure := r.resolve(ctx, target)
	if failure != nil {
		return failure
	}
	model := c.catalog.ClassifierModels[i].clone()
	if failure := validateOptions(opts, model.API); failure != nil {
		return failure
	}
	if opts != nil {
		opts = opts.clone()
	}
	if problem := req.checkCapabilities(model.Capabilities); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	cred, failure := r.pin(ctx, model.ID, hooks)
	if failure != nil {
		return failure
	}
	// Fixed dispatch: this operation currently has one real protocol, so a
	// dynamic registry or generic adapter interface adds no useful boundary.
	return c.classifyTypeSafe(ctx, r, cred, model, req, res)
}
