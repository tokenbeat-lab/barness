package ai

import "math"

// Live Jev 1.13 returns score and probabilities independently at hundredth
// precision (ADR-0021, issue 08). Preserve the wire values; accept rounding
// only when a unit-sum underlying distribution within each +/- .005 interval
// could produce the reported score within its +/- .005 interval. More precise
// values keep the original 1e-6 rule. This is a bounded compatibility inference
// from our own official wire, not a promise of a vendor rounding algorithm.
func typeSafeRoundedScoreConsistent(score float64, probabilities map[int]float64) bool {
	const halfQuantum = 0.005
	hundredth := func(v float64) bool { return math.Abs(v*100-math.Round(v*100)) <= 1e-9 }
	if !hundredth(score) {
		return false
	}
	low, high := make([]float64, len(probabilities)), make([]float64, len(probabilities))
	mass, upperMass := 0.0, 0.0
	for i := range low {
		p := probabilities[i]
		if !hundredth(p) {
			return false
		}
		low[i], high[i] = max(0, p-halfQuantum), min(1, p+halfQuantum)
		mass += low[i]
		upperMass += high[i]
	}
	if mass > 1 || upperMass < 1 {
		return false
	}
	// Linear scores attain their extremes by assigning remaining probability
	// mass to the lowest/highest available levels respectively.
	extreme := func(descending bool) float64 {
		expected := 0.0
		remaining := 1 - mass
		for i, p := range low {
			expected += float64(i) * p
		}
		for step := range low {
			i := step
			if descending {
				i = len(low) - 1 - step
			}
			add := min(remaining, high[i]-low[i])
			remaining -= add
			expected += float64(i) * add
		}
		return expected
	}
	return score+halfQuantum+1e-6 >= extreme(false) && score-halfQuantum-1e-6 <= extreme(true)
}
