package ai

import "context"

// ClassifierImageReferencePolicy bounds each image occurrence, including reuse.
// Nil References inherits the finite source count and total decoded byte limits.
type ClassifierImageReferencePolicy struct {
	MaxReferences         int
	MaxExpandedImageBytes int64
}

func (p *ClassifierImagePolicy) referenceLimits() ClassifierImageReferencePolicy {
	if p.References != nil {
		return *p.References
	}
	return ClassifierImageReferencePolicy{MaxReferences: p.MaxImages, MaxExpandedImageBytes: p.MaxTotalImageBytes}
}

// Source sizes were checked separately. This pass reads borrowed entries only,
// so even a huge reference table or pointer fails before plan allocation/copy.
func (p *ResourcePolicy) checkClassifierReferenceSizes(ctx context.Context, req ClassifierRequest) *Error {
	count := len(req.ImageReferences)
	if count == 0 {
		count = len(req.Images)
	}
	if count == 0 {
		return nil
	}
	if p.Classifier == nil || p.Classifier.Images == nil {
		return imageInputFailure(PhaseScope, "the client policy does not enable classifier images")
	}
	limits := p.Classifier.Images.referenceLimits()
	if count > limits.MaxReferences {
		return limitFailure(PhaseScope, "Classifier.Images.References.MaxReferences", int64(limits.MaxReferences))
	}
	remaining := limits.MaxExpandedImageBytes
	encoded := p.MaxRequestBytes
	// Shorthand has no repeated expansion: the existing raw-input check already
	// counts every source once. Explicit references need a separate projection,
	// with cheap raw bounds first so oversized documents are never scanned.
	if len(req.ImageReferences) > 0 {
		if failure := p.byteLimits().checkClassifierInput(req); failure != nil {
			return failure
		}
		if failure := p.Classifier.checkReferenceDocumentSizes(req); failure != nil {
			return failure
		}
		if failure := contextError(ctx, PhaseScope); failure != nil {
			return failure
		}
		stateSize := classifierJSONTextLowerBound(req.State)
		if stateSize > encoded {
			return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
		}
		encoded -= stateSize
		for key, question := range req.Questions {
			if failure := contextError(ctx, PhaseScope); failure != nil {
				return failure
			}
			if int64(len(key)) > encoded {
				return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
			}
			encoded -= int64(len(key))
			size, ok := classifierQuestionTextLowerBound(question, encoded)
			if !ok {
				return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
			}
			encoded -= size
		}
	}
	for i := 0; i < count; i++ {
		if failure := contextError(ctx, PhaseScope); failure != nil {
			return failure
		}
		index := i
		if len(req.ImageReferences) > 0 {
			ref := req.ImageReferences[i]
			if problem := ref.problem(); problem != "" {
				return newError(CodeInvalidRequest, PhaseScope, problem)
			}
			index = ref.ImageIndex
		}
		if index < 0 || index >= len(req.Images) {
			return newError(CodeInvalidRequest, PhaseScope, "image reference index is out of range")
		}
		img := req.Images[index]
		size := base64Size(img.Data)
		if size > remaining {
			return limitFailure(PhaseScope, "Classifier.Images.References.MaxExpandedImageBytes", limits.MaxExpandedImageBytes)
		}
		remaining -= size
		// The final inline expansion cannot fit if even its unescaped strings do
		// not fit. The full native framing is checked again at the encoder boundary.
		for _, n := range []int{len(img.Data), len(img.MimeType), len("data:;base64,")} {
			if int64(n) > encoded {
				return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
			}
			encoded -= int64(n)
		}
	}
	return nil
}

func (p *ResourcePolicy) checkClassifierExpandedImages(images []Image, phase Phase) *Error {
	if len(images) == 0 {
		return nil
	}
	limits := p.Classifier.Images.referenceLimits()
	if len(images) > limits.MaxReferences {
		return limitFailure(phase, "Classifier.Images.References.MaxReferences", int64(limits.MaxReferences))
	}
	remaining := limits.MaxExpandedImageBytes
	for _, img := range images {
		size := base64Size(img.Data)
		if size > p.MaxImageBytes {
			return limitFailure(phase, "MaxImageBytes", p.MaxImageBytes)
		}
		if size > remaining {
			return limitFailure(phase, "Classifier.Images.References.MaxExpandedImageBytes", limits.MaxExpandedImageBytes)
		}
		remaining -= size
	}
	return nil
}

// Eager raw bounds protect projection traversal. Full semantic validation still
// runs at the ordinary public boundary with the normalized image plan.
func (p *ClassifierPolicy) checkReferenceDocumentSizes(req ClassifierRequest) *Error {
	if len(req.Questions) > p.MaxQuestions {
		return limitFailure(PhaseScope, "Classifier.MaxQuestions", int64(p.MaxQuestions))
	}
	if int64(len(req.State)) > p.MaxStateBytes {
		return limitFailure(PhaseScope, "Classifier.MaxStateBytes", p.MaxStateBytes)
	}
	for _, q := range req.Questions {
		if _, ok := classifierQuestionSize(q, p.MaxQuestionBytes); !ok {
			return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
		}
	}
	return nil
}
