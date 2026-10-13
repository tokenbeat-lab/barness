package chineseeval

import "slices"

type Latency struct {
	Count  int      `json:"count"`
	MeanMS *float64 `json:"meanMS"`
	P50MS  *float64 `json:"p50MS"`
	P95MS  *float64 `json:"p95MS"`
}

// Latency measures all executed logical calls, including refusals and failures.
// Nearest-rank percentiles use observed values; no empty-bin interpolation.
func latencies(r Report) *Latency {
	l := &Latency{}
	values := []float64{}
	sum := 0.0
	for _, s := range r.Samples {
		if s.Metadata.RequestID != "" && s.DurationMS != nil {
			v := *s.DurationMS
			values = append(values, v)
			sum += v
		}
	}
	l.Count = len(values)
	if l.Count > 0 {
		slices.Sort(values)
		l.MeanMS = number(sum / float64(l.Count))
		l.P50MS = number(values[(l.Count*50+99)/100-1])
		l.P95MS = number(values[(l.Count*95+99)/100-1])
	}
	return l
}
