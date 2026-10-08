package chineseeval

import (
	"encoding/json"
	"maps"
	"math"
	"slices"
)

type Bin struct {
	Count    int      `json:"count"`
	Correct  int      `json:"correct"`
	Mean     *float64 `json:"mean"`
	Accuracy *float64 `json:"accuracy"`
}

type Metrics struct {
	Total              int                       `json:"total"`
	Scored             int                       `json:"scored"`
	Correct            int                       `json:"correct"`
	Failed             int                       `json:"failed"`
	Unexecuted         int                       `json:"unexecuted"`
	Accuracy           *float64                  `json:"accuracy"`
	ScoredAccuracy     *float64                  `json:"scoredAccuracy"`
	Confusion          map[string]map[string]int `json:"confusion"`
	StatusCounts       map[string]int            `json:"statusCounts"`
	ConfidenceBins     []Bin                     `json:"confidenceBins"`
	TopProbabilityBins []Bin                     `json:"topProbabilityBins"`
	ConfidenceECE      *float64                  `json:"confidenceECE"`
	TopProbabilityECE  *float64                  `json:"topProbabilityECE"`
	Brier              *float64                  `json:"brier"`
	ErrorIDs           []string                  `json:"errorIDs"`
}

// Accuracy uses ALL labeled samples. ScoredAccuracy is explicitly conditional;
// failed/unevaluated samples remain visible in denominators and confusion rows.
// Confidence is the vendor's separate confidence signal, not a calibrated
// probability of correctness. Both it and the chosen-option probability get
// descriptive reliability bins; no production decision threshold is added.
func metrics(r Report) Metrics {
	m := Metrics{Total: len(r.Dataset.Samples), Confusion: map[string]map[string]int{}, StatusCounts: map[string]int{}, ConfidenceBins: make([]Bin, 10), TopProbabilityBins: make([]Bin, 10), ErrorIDs: []string{}}
	for _, label := range slices.Sorted(maps.Keys(r.Dataset.Labels)) {
		m.Confusion[label] = map[string]int{}
	}
	confidenceSum, probabilitySum := make([]float64, 10), make([]float64, 10)
	brier := 0.0
	for i, s := range r.Samples {
		gold := r.Dataset.Samples[i].Label
		m.StatusCounts[s.Status]++
		if s.Status != "answered" || s.Answer == nil {
			m.Confusion[gold][s.Status]++
			if s.Metadata.RequestID == "" {
				m.Unexecuted++
			} else {
				m.Failed++
			}
			m.ErrorIDs = append(m.ErrorIDs, s.ID)
			continue
		}
		a := s.Answer
		m.Scored++
		m.Confusion[gold][a.Choice]++
		correct := a.Choice == gold
		if correct {
			m.Correct++
		} else {
			m.ErrorIDs = append(m.ErrorIDs, s.ID)
		}
		addBin(m.ConfidenceBins, confidenceSum, a.Confidence, correct)
		addBin(m.TopProbabilityBins, probabilitySum, a.Probabilities[a.Choice], correct)
		for _, label := range slices.Sorted(maps.Keys(r.Dataset.Labels)) {
			truth := 0.0
			if label == gold {
				truth = 1
			}
			delta := a.Probabilities[label] - truth
			brier += delta * delta
		}
	}
	if r.Status != "NOT_RUN" {
		m.Accuracy = number(float64(m.Correct) / float64(m.Total))
	}
	if m.Scored > 0 {
		m.ScoredAccuracy = number(float64(m.Correct) / float64(m.Scored))
		m.ConfidenceECE = finishBins(m.ConfidenceBins, confidenceSum, m.Scored)
		m.TopProbabilityECE = finishBins(m.TopProbabilityBins, probabilitySum, m.Scored)
		m.Brier = number(brier / float64(m.Scored))
	}
	return m
}

func addBin(bins []Bin, sums []float64, value float64, correct bool) {
	i := min(9, int(value*10))
	bins[i].Count++
	sums[i] += value
	if correct {
		bins[i].Correct++
	}
}
func finishBins(bins []Bin, sums []float64, scored int) *float64 {
	ece := 0.0
	for i := range bins {
		b := &bins[i]
		if b.Count == 0 {
			continue
		}
		b.Mean = number(sums[i] / float64(b.Count))
		b.Accuracy = number(float64(b.Correct) / float64(b.Count))
		ece += float64(b.Count) / float64(scored) * math.Abs(*b.Mean-*b.Accuracy)
	}
	return number(ece)
}
func number(value float64) *float64           { return &value }
func jsonString(value string) json.RawMessage { raw, _ := json.Marshal(value); return raw }
