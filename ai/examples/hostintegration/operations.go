package hostintegration

import (
	"context"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
)

// ImageRequest is the host's image call. TenantID is only a claim. A real
// HTTP handler decodes its DTO and converts inline images/options to these
// domain values; the Client owns validation and independent call snapshots.
type ImageRequest struct {
	TenantID string
	Target   ai.Target
	Input    ai.ImagesRequest
	Options  ai.ImageOptions
}

type ClassificationRequest struct {
	TenantID string
	Target   ai.Target
	Input    ai.ClassifierRequest
	Options  ai.Options
}

// GenerateImages passes downstream cancellation through to one logical call.
// A failure returns empty output and typed ai.Error plus attempt/usage metadata;
// never log ErrorMessage to a shared log (it is the caller's content).
func (h *Host) GenerateImages(ctx context.Context, who Principal, in ImageRequest) (ai.ImagesResult, error) {
	scope, err := h.operationScope(who, in.TenantID, in.Target)
	if err != nil {
		return ai.ImagesResult{}, err
	}
	return h.client.GenerateImages(ctx, scope, in.Target, in.Input, in.Options)
}

// Classify returns typed answers without imposing business confidence
// thresholds, routing or human review. The caller makes those decisions and
// handles ai.Error codes such as admission_denied/deadline_exceeded. Repeating
// a failed call can spend again; RequestID is attribution, not deduplication.
func (h *Host) Classify(ctx context.Context, who Principal, in ClassificationRequest) (ai.ClassifierResult, error) {
	scope, err := h.operationScope(who, in.TenantID, in.Target)
	if err != nil {
		return ai.ClassifierResult{}, err
	}
	return h.client.Classify(ctx, scope, in.Target, in.Input, in.Options)
}

func (h *Host) operationScope(who Principal, claimed string, target ai.Target) (ai.CallScope, error) {
	if who.TenantID == "" || (claimed != "" && claimed != who.TenantID) {
		return ai.CallScope{}, fmt.Errorf("%w: unauthenticated or tenant claim mismatch", ErrForbidden)
	}
	if target.BindingID == "" || target.ModelID == "" {
		return ai.CallScope{}, fmt.Errorf("%w: binding and model are required", ErrBadRequest)
	}
	return ai.CallScope{TenantID: who.TenantID, ActorID: who.ActorID, RequestID: h.newRequestID()}, nil
}
