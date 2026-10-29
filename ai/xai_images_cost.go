package ai

import "math"

// A request count is never evidence of paid output after validation fails.
// This image route deliberately leaves Usage.Cost as token-only zero cost.
func xaiImagesEstimate(model ImageModel, w xaiImagesRequest, outputs int) CostEstimate {
	result := CostEstimate{Status: CostPartial}
	p := model.Pricing.PerImage
	if p == nil {
		return result
	}
	if outputs != w.options().quantity() {
		return result
	}
	size, sizeKnown := w.Resolution.Get()
	quality, qualityKnown := w.Quality.Get()
	if !sizeKnown || !qualityKnown {
		return result
	}
	inputRate, hasInput := p.InputImage.Get()
	if len(w.resources()) > 0 && !hasInput {
		return result
	}
	low, high := math.Inf(1), math.Inf(-1)
	for _, row := range p.Output {
		if row.ImageSize != size || (quality != "auto" && row.Quality != quality) {
			continue
		}
		rate, ok := row.Price.Get()
		if !ok {
			return result
		}
		low = min(low, rate)
		high = max(high, rate)
	}
	if math.IsInf(low, 0) || math.IsInf(high, 0) {
		return result
	}
	minimum := UsageCost{Input: inputRate * float64(len(w.resources())), Output: low * float64(outputs)}
	minimum.Total = minimum.Input + minimum.Output
	maximum := UsageCost{Input: minimum.Input, Output: high * float64(outputs)}
	maximum.Total = maximum.Input + maximum.Output
	if !xaiFiniteCost(minimum) || !xaiFiniteCost(maximum) {
		return result
	}
	result.Status = CostResolved
	if quality == "auto" {
		result.Status = CostBounded
	}
	result.Minimum, result.Maximum = Value(minimum), Value(maximum)
	return result
}
