package supportmatrix

import (
	"os"
	"reflect"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyXAIImagesObserver(root *os.Root, path string, n xaiImagesNative, result xaiImagesResult) error {
	var rows []ai.Observation
	if e := xaiReadJSON(root, path, &rows); e != nil || len(rows) != 4 {
		return refuse("Imagine four native lifecycle observations required")
	}
	kinds := []ai.ObservationKind{ai.ObservationCallStarted, ai.ObservationAttemptStarted, ai.ObservationAttemptFinished, ai.ObservationCallFinished}
	meta := result.Metadata
	attempt := meta.Attempts[0]
	bare := meta
	bare.Attempts = nil
	for i, row := range rows {
		if row.Kind != kinds[i] || row.Time.Before(n.StartedAt) || row.Time.After(n.FinishedAt) || i > 0 && row.Time.Before(rows[i-1].Time) || row.Error != nil {
			return refuse("Imagine Observer lifecycle, interval or failure differs")
		}
		if i == 0 {
			want := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: meta.TenantID, RequestID: meta.RequestID, JobID: meta.JobID, BindingID: meta.BindingID, Operation: ai.OperationImage}}
			if !reflect.DeepEqual(row.Call, want) || row.Attempt != nil {
				return refuse("Imagine unresolved Observer attribution differs")
			}
		}
		if i == 1 || i == 2 {
			if !reflect.DeepEqual(row.Call, bare) || row.Attempt == nil {
				return refuse("Imagine attempt Observer attribution differs")
			}
		}
		if i == 1 {
			want := ai.Attempt{AttemptID: attempt.AttemptID, UsageReporting: ai.UsageUnreported, CostEstimate: ai.CostEstimate{Status: ai.CostUnreported}}
			if !reflect.DeepEqual(*row.Attempt, want) {
				return refuse("Imagine starting attempt fabricated consumption")
			}
		}
		if i == 2 && (!reflect.DeepEqual(*row.Attempt, attempt) || row.Duration <= 0 || row.Duration > 180*time.Second) {
			return refuse("Imagine finished attempt consumption or timeout differs")
		}
		if i == 3 && (!reflect.DeepEqual(row.Call, meta) || row.Attempt != nil || row.StopReason != ai.StopReasonStop || !reflect.DeepEqual(row.Usage, result.Usage) || row.Duration <= 0 || row.Duration > 180*time.Second) {
			return refuse("Imagine terminal Observer consumption or attribution differs")
		}
		if i < 3 && (row.StopReason != "" || !reflect.DeepEqual(row.Usage, ai.Usage{})) {
			return refuse("Imagine Observer exposes premature terminal facts")
		}
	}
	return nil
}
