package ai

import (
	"encoding/json"
	"fmt"
)

// ImagePlacement distinguishes shared context from an explicit JSON value.
type ImagePlacement string

const (
	ImagePlacementContext ImagePlacement = "context"
	ImagePlacementValue   ImagePlacement = "value"
)

// ImageReference selects one entry in ClassifierRequest.Images. Value pointers
// are RFC 6901 strings relative to the public request, targeting a null slot.
// A nonempty reference list must use every source; repeated context uses retain
// their declared order and each consume the expansion budget.
type ImageReference struct {
	ImageIndex int            `json:"imageIndex"`
	Placement  ImagePlacement `json:"placement"`
	Pointer    string         `json:"pointer,omitempty"`
}

func (r ImageReference) problem() string {
	if r.ImageIndex < 0 {
		return "image reference index must be nonnegative"
	}
	switch r.Placement {
	case ImagePlacementContext:
		if r.Pointer != "" {
			return "context image reference must omit pointer"
		}
	case ImagePlacementValue:
		if r.Pointer == "" {
			return "value image reference requires pointer"
		}
	default:
		return "unknown image reference placement"
	}
	return ""
}

func (r ImageReference) MarshalJSON() ([]byte, error) {
	if problem := r.problem(); problem != "" {
		return nil, fmt.Errorf("%s", problem)
	}
	type record ImageReference
	return json.Marshal(record(r))
}

func (r *ImageReference) UnmarshalJSON(raw []byte) error {
	var wire struct {
		ImageIndex *int            `json:"imageIndex"`
		Placement  *ImagePlacement `json:"placement"`
		Pointer    json.RawMessage `json:"pointer"`
	}
	if err := decodeClassifierJSON(raw, &wire); err != nil {
		return err
	}
	if wire.ImageIndex == nil || wire.Placement == nil {
		return fmt.Errorf("image reference requires imageIndex and placement")
	}
	out := ImageReference{ImageIndex: *wire.ImageIndex, Placement: *wire.Placement}
	if len(wire.Pointer) > 0 {
		if out.Placement == ImagePlacementContext || string(wire.Pointer) == "null" {
			return fmt.Errorf("invalid image reference pointer")
		}
		if err := json.Unmarshal(wire.Pointer, &out.Pointer); err != nil {
			return err
		}
	}
	if problem := out.problem(); problem != "" {
		return fmt.Errorf("%s", problem)
	}
	*r = out
	return nil
}

// Both shorthand and explicit declarations produce this one owned plan.
type classifierImagePlan []ImageReference

func planClassifierImages(req ClassifierRequest) (classifierImagePlan, *Error) {
	refs := req.ImageReferences
	if len(refs) == 0 {
		refs = make([]ImageReference, len(req.Images))
		for i := range refs {
			refs[i] = ImageReference{ImageIndex: i, Placement: ImagePlacementContext}
		}
	}
	seen := make([]bool, len(req.Images))
	plan := make(classifierImagePlan, 0, len(refs))
	for _, ref := range refs {
		if problem := ref.problem(); problem != "" {
			return nil, newError(CodeInvalidRequest, PhaseScope, problem)
		}
		if ref.ImageIndex >= len(req.Images) {
			return nil, newError(CodeInvalidRequest, PhaseScope, "image reference index is out of range")
		}
		seen[ref.ImageIndex] = true
		plan = append(plan, ref)
	}
	for _, used := range seen {
		if !used {
			return nil, newError(CodeInvalidRequest, PhaseScope, "explicit image references must use every source")
		}
	}
	return plan, nil
}
