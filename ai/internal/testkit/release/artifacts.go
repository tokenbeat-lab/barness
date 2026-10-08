package release

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

type designLoad struct {
	Load     struct{ Calls, Tenants, References, InputImageBytes, MaskBytes, OutputImages, OutputImageBytes, StateBytes, Questions int }
	Headroom map[string]struct {
		Design, Limit int64
		Ratio         float64
	}
	Heap                                                             struct{ BaselineLive, PeakLive uint64 }
	HeapBudget, HeapGrowth, TotalAllocDuringCalls                    uint64
	PeakPermits, ClientUnaryBytesAtEOF, ProviderCapturedRequestBytes int64
	InputImageBase64Bytes, OutputImageBase64Bytes                    int
	Duration                                                         string
	CallsPerSecond                                                   float64
}

// DesignLoadEvidence reads the actual reports and release probes of both
// opt-in mixed loads. A PASS case without these artifacts cannot prove capacity.
func DesignLoadEvidence(b Bundle) []EvidenceCheck {
	var out []EvidenceCheck
	for i, name := range []string{"local", "cloud"} {
		id := "E08-mixed-pressure-" + name + "-design-load"
		e := EvidenceCheck{Name: "mixed-pressure-" + name, Bundle: b.Dir, Status: NotRun}
		dir := b.caseDirs[id]
		var r designLoad
		var gauges map[string]int64
		if dir == "" {
			e.Detail = "required design-load case missing"
			out = append(out, e)
			continue
		}
		e.Status = Fail
		if err := readJSON(filepath.Join(dir, "mixed-pressure.json"), &r); err != nil {
			e.Detail = "design-load report missing or invalid"
			out = append(out, e)
			continue
		}
		if err := readJSON(filepath.Join(dir, "resources.json"), &gauges); err != nil {
			e.Detail = "resource-release report missing or invalid"
			out = append(out, e)
			continue
		}
		valid := r.Load.Calls == 4*(i+1) && r.Load.Tenants == i+1 && r.PeakPermits == int64(r.Load.Calls) && r.HeapBudget == uint64(256<<20)*(uint64(i)+1) && r.Heap.PeakLive >= r.Heap.BaselineLive && r.HeapGrowth == r.Heap.PeakLive-r.Heap.BaselineLive && r.HeapGrowth <= r.HeapBudget && r.TotalAllocDuringCalls > 0 && r.TotalAllocDuringCalls <= r.HeapBudget && r.ClientUnaryBytesAtEOF > 0 && r.ProviderCapturedRequestBytes > 0 && r.CallsPerSecond > 0 && !math.IsInf(r.CallsPerSecond, 0)
		duration, err := time.ParseDuration(r.Duration)
		valid = valid && err == nil && duration > 0 && r.Load.References == 2 && r.Load.MaskBytes == 256<<10 && r.Load.InputImageBytes == 256<<10 && r.Load.OutputImages == 2 && r.Load.OutputImageBytes == 512<<10 && r.Load.StateBytes == 128<<10 && r.Load.Questions == 8 && r.InputImageBase64Bytes == base64.StdEncoding.EncodedLen(r.Load.InputImageBytes) && r.OutputImageBase64Bytes == base64.StdEncoding.EncodedLen(r.Load.OutputImageBytes)
		for _, key := range []string{"MaxRequestBytes", "MaxImageBytes", "MaxFrameBytes (chat only)", "MaxOutputBytes", "Image.MaxInputImages", "Image.MaxOutputImages", "Image.MaxOutputImageBytes", "Image.MaxTotalOutputImageBytes", "Classifier.MaxQuestions", "Classifier.MaxStateBytes", "Classifier.MaxQuestionBytes"} {
			h, ok := r.Headroom[key]
			valid = valid && ok && h.Design > 0 && h.Limit > 0 && h.Ratio > 0 && h.Ratio <= .75 && math.Abs(h.Ratio-float64(h.Design)/float64(h.Limit)) <= 1e-9
		}
		for _, key := range []string{"calls", "permits", "waiters", "events", "bodies", "hostPermits"} {
			n, ok := gauges[key]
			valid = valid && ok && n == 0
		}
		if valid {
			e.Status = Pass
			e.Detail = fmt.Sprintf("%d concurrent calls, heap growth %d / budget %d, allocation upper bound %d; all headroom <=75%%; resources zero", r.Load.Calls, r.HeapGrowth, r.HeapBudget, r.TotalAllocDuringCalls)
		} else {
			e.Detail = "design load, memory/headroom budget, concurrency or resource-release proof invalid"
		}
		out = append(out, e)
	}
	return out
}

// ChineseEffectEvidence imports the independent fixed-task host artifacts.
// Integrity/completeness is the release requirement, never an invented score.
func ChineseEffectEvidence(dir string) EvidenceCheck {
	e := EvidenceCheck{Name: "chinese-effect", Bundle: dir, Status: NotRun, Detail: "independent real evaluation not supplied"}
	if dir == "" {
		return e
	}
	data, err := readArtifact(filepath.Join(dir, "dataset.json"), 1<<20)
	if err != nil {
		return e
	}
	config, err := readArtifact(filepath.Join(dir, "config.json"), 16384)
	if err != nil {
		return e
	}
	e.Status = Fail
	plan, err := chineseeval.Load(data, config, ai.BuiltinCatalog())
	if err != nil {
		e.Detail = "dataset/config/catalog drift"
		return e
	}
	raw, err := readArtifact(filepath.Join(dir, "evaluation-report.json"), 16<<20)
	if err != nil {
		e.Detail = "evaluation report missing"
		return e
	}
	r, err := plan.ReadReport(raw)
	if err != nil {
		e.Detail = "evaluation integrity failed"
		return e
	}
	manifest, err := readArtifact(filepath.Join(dir, "manifest.json"), 1<<20)
	if err != nil {
		e.Detail = "evaluation manifest missing"
		return e
	}
	observations, err := readArtifact(filepath.Join(dir, "observations-evaluation.json"), 1<<20)
	if err != nil {
		e.Detail = "evaluation Observer missing"
		return e
	}
	if plan.VerifyEvidence(r, manifest, observations) != nil {
		e.Detail = "evaluation provenance or observations incomplete"
		return e
	}
	if r.Kind != "live" || r.Status != "COMPLETE" {
		e.Status = NotRun
		if r.Status == "FAIL" {
			e.Status = Fail
		}
		e.Detail = "evaluation is " + r.Kind + "/" + r.Status
		return e
	}
	e.Status = Pass
	e.Detail = fmt.Sprintf("independent synthetic Chinese task, %d/%d scored, accuracy %.6f; dataset %s; config %s; no accuracy threshold", r.Metrics.Scored, r.Dataset.Count, *r.Metrics.Accuracy, r.DatasetHash, r.ConfigHash)
	return e
}

// Import budgets apply before allocation, even if a file changes while read.
func readArtifact(path string, budget int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, budget+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > budget {
		return nil, fmt.Errorf("release: artifact exceeds import budget")
	}
	return raw, nil
}
