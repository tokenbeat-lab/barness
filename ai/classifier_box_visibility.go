package ai

// Public visibility follows validated placement declarations, never a count of
// source files or another question's local use. Native visibility is checked
// independently from actual final whole-value occurrences after the Hook.
func (p classifierImagePlan) checkBoxes(req ClassifierRequest) *Error {
	shared := false
	local := map[string]bool{}
	for _, ref := range p {
		if ref.Placement == ImagePlacementContext {
			shared = true
			continue
		}
		parts, _ := imagePointerParts(ref.Pointer) // validated content targets
		if parts[0] == "state" {
			shared = true
		} else if len(parts) >= 3 && parts[0] == "questions" && parts[2] == "instructions" {
			local[parts[1]] = true
		}
	}
	for name, question := range req.Questions {
		if _, box := question.(BoxQuestion); box && !shared && !local[name] {
			return newError(CodeInvalidRequest, PhaseScope, "each Box requires an image in context, State or its own Instructions")
		}
	}
	return nil
}
