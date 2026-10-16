package main

import (
	"encoding/json"
	"math"

	"github.com/tokenbeat-lab/barness/ai"
)

// answerView is a Host-owned copy, preserving absent/refused and nullable Box
// confidence. Bool probability and Choice confidence are separate statistics.
type answerView struct {
	Status        string             `json:"status"`
	Choice        string             `json:"choice,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    *float64           `json:"confidence"`
	Probability   *float64           `json:"probability"`
	Box           *ai.Box            `json:"box,omitempty"`
}
type metricGroup struct {
	Total          int                       `json:"total"`
	ClassScored    int                       `json:"classScored"`
	ClassCorrect   int                       `json:"classCorrect"`
	ExistScored    int                       `json:"existScored"`
	ExistCorrect   int                       `json:"existCorrect"`
	FalsePositive  int                       `json:"falsePositive"`
	FalseNegative  int                       `json:"falseNegative"`
	ClassAccuracy  *float64                  `json:"classAccuracy"`
	ExistAccuracy  *float64                  `json:"existAccuracy"`
	Confusion      map[string]map[string]int `json:"confusion"`
	Statuses       map[string]int            `json:"statuses"`
	Errors         []string                  `json:"errors"`
	ConfidenceBins []confidenceBin           `json:"confidenceBins"`
	ExistenceBrier *float64                  `json:"existenceBrier"`
}
type confidenceBin struct {
	Count   int     `json:"count"`
	Correct int     `json:"correct"`
	Sum     float64 `json:"sum"`
}
type boxMetrics struct {
	Present                   int      `json:"present"`
	Absent                    int      `json:"absent"`
	IoUCount                  int      `json:"iouCount"`
	MeanIoU                   *float64 `json:"meanIoU"`
	PresentDenominatorMeanIoU *float64 `json:"presentDenominatorMeanIoU"`
	Consumed                  int      `json:"consumed"`
	AbsentFalsePositive       int      `json:"absentFalsePositive"`
	AbsentRawBoxes            int      `json:"absentRawBoxes"`
	LowExistenceBoxes         int      `json:"lowExistenceBoxes"`
	MissingExistenceBoxes     int      `json:"missingExistenceBoxes"`
	MissingPresentBoxes       int      `json:"missingPresentBoxes"`
}
type metrics struct {
	Common       metricGroup  `json:"common"`
	Local        *metricGroup `json:"local"`
	Supplemental *metricGroup `json:"supplementalContext"`
	Box          *boxMetrics  `json:"box"`
}

func number(v float64) *float64 { return &v }
func iou(a, b ai.Box) float64 {
	intersection := math.Max(0, math.Min(a.Right, b.Right)-math.Max(a.Left, b.Left)) * math.Max(0, math.Min(a.Bottom, b.Bottom)-math.Max(a.Top, b.Top))
	return intersection / ((a.Right-a.Left)*(a.Bottom-a.Top) + (b.Right-b.Left)*(b.Bottom-b.Top) - intersection)
}
func views(raw map[string]json.RawMessage) map[string]answerView {
	out := map[string]answerView{}
	for key, b := range raw {
		var a struct {
			Type, Choice  string
			Confidence    *float64
			Probability   *float64
			Box           *ai.Box
			Probabilities map[string]float64
		}
		if json.Unmarshal(b, &a) != nil {
			continue
		}
		status := "answered"
		if a.Type == "refusal" {
			status = "refused"
		}
		out[key] = answerView{status, a.Choice, a.Probabilities, a.Confidence, a.Probability, a.Box}
	}
	return out
}
func score(p *plan, r report) metrics {
	m := metrics{Common: groupMetrics(p, r, 0, 8)}
	extra := groupMetrics(p, r, 8, 12)
	if p.route == "decisions" {
		m.Supplemental = &extra
		return m
	}
	m.Local = &extra
	b := &boxMetrics{}
	sum := 0.0
	for i := 8; i < 12; i++ {
		gold := p.dataset.Samples[i]
		s := r.Samples[i]
		a, ok := s.View["location"]
		if gold.Exists {
			b.Present++
		} else {
			b.Absent++
		}
		if !ok || a.Box == nil {
			if gold.Exists {
				b.MissingPresentBoxes++
			}
			continue
		}
		s.IoU = nil
		if gold.Exists {
			sum += iou(*a.Box, *gold.Box)
			b.IoUCount++
		} else {
			b.AbsentRawBoxes++
		}
		exists := s.View["exists"].Probability
		if exists == nil {
			b.MissingExistenceBoxes++
			continue
		}
		if *exists < p.experiment.ExistenceThreshold {
			b.LowExistenceBoxes++
			continue
		}
		if gold.Exists {
			b.Consumed++
		} else {
			b.AbsentFalsePositive++
		}
	}
	if b.IoUCount > 0 {
		b.MeanIoU = number(sum / float64(b.IoUCount))
	}
	if r.Budget.Calls > 0 {
		b.PresentDenominatorMeanIoU = number(sum / float64(b.Present))
	}
	m.Box = b
	return m
}
func groupMetrics(p *plan, r report, start, end int) metricGroup {
	m := metricGroup{Total: end - start, Confusion: map[string]map[string]int{}, Statuses: map[string]int{}, Errors: []string{}, ConfidenceBins: make([]confidenceBin, 10)}
	brier := 0.0
	for i := start; i < end; i++ {
		s, gold := r.Samples[i], p.dataset.Samples[i]
		m.Statuses[s.Status]++
		a := s.View["kind"]
		pred := a.Choice
		if pred == "" {
			pred = s.Status
			if a.Status != "" {
				pred = a.Status
			}
		}
		if m.Confusion[gold.Label] == nil {
			m.Confusion[gold.Label] = map[string]int{}
		}
		m.Confusion[gold.Label][pred]++
		correct := a.Status == "answered" && a.Choice == gold.Label
		if a.Status == "answered" {
			m.ClassScored++
		}
		if correct {
			m.ClassCorrect++
		}
		if a.Confidence != nil {
			idx := min(9, int(*a.Confidence*10))
			bin := &m.ConfidenceBins[idx]
			bin.Count++
			bin.Sum += *a.Confidence
			if correct {
				bin.Correct++
			}
		}
		e := s.View["exists"].Probability
		existCorrect := false
		if e != nil {
			m.ExistScored++
			predicted := *e >= p.experiment.ExistenceThreshold
			existCorrect = predicted == gold.Exists
			if existCorrect {
				m.ExistCorrect++
			}
			if predicted && !gold.Exists {
				m.FalsePositive++
			}
			if !predicted && gold.Exists {
				m.FalseNegative++
			}
			truth := 0.0
			if gold.Exists {
				truth = 1
			}
			brier += (*e - truth) * (*e - truth)
		}
		if !correct || !existCorrect {
			m.Errors = append(m.Errors, s.ID)
		}
	}
	if r.Budget.Calls > 0 {
		m.ClassAccuracy = number(float64(m.ClassCorrect) / float64(m.Total))
		m.ExistAccuracy = number(float64(m.ExistCorrect) / float64(m.Total))
	}
	if m.ExistScored > 0 {
		m.ExistenceBrier = number(brier / float64(m.ExistScored))
	}
	return m
}
