package release

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
)

type designLoad struct {
	Load                                                             struct{ Calls, Tenants, HistoryBytes, ChatOutputBytes, References, InputImageBytes, MaskBytes, OutputImages, OutputImageBytes, StateBytes, Questions, InstructionBytes int }
	Headroom                                                         map[string]loadLimit
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
		var gauges map[string]*int64
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
		valid := len(b.Problems) == 0 && r.Load.HistoryBytes == 128<<10 && r.Load.ChatOutputBytes == 16<<10 && r.Load.InstructionBytes == 4000 && r.Load.Calls == 4*(i+1) && r.Load.Tenants == i+1 && r.PeakPermits == int64(r.Load.Calls) && r.HeapBudget == uint64(256<<20)*(uint64(i)+1) && r.Heap.PeakLive >= r.Heap.BaselineLive && r.HeapGrowth == r.Heap.PeakLive-r.Heap.BaselineLive && r.HeapGrowth <= r.HeapBudget && r.TotalAllocDuringCalls > 0 && r.TotalAllocDuringCalls <= r.HeapBudget && r.ClientUnaryBytesAtEOF > 0 && r.ProviderCapturedRequestBytes > 0 && r.CallsPerSecond > 0 && !math.IsInf(r.CallsPerSecond, 0)
		duration, err := time.ParseDuration(r.Duration)
		valid = valid && err == nil && duration > 0 && r.Load.References == 2 && r.Load.MaskBytes == 256<<10 && r.Load.InputImageBytes == 256<<10 && r.Load.OutputImages == 2 && r.Load.OutputImageBytes == 512<<10 && r.Load.StateBytes == 128<<10 && r.Load.Questions == 8 && r.InputImageBase64Bytes == base64.StdEncoding.EncodedLen(r.Load.InputImageBytes) && r.OutputImageBase64Bytes == base64.StdEncoding.EncodedLen(r.Load.OutputImageBytes)
		limits := map[string]int64{"MaxRequestBytes": 4 << 20, "MaxImageBytes": 1 << 20, "MaxFrameBytes (chat only)": 256 << 10, "MaxOutputBytes": 8 << 20, "Image.MaxInputImages": 4, "Image.MaxOutputImages": 4, "Image.MaxOutputImageBytes": 1 << 20, "Image.MaxTotalOutputImageBytes": 2 << 20, "Classifier.MaxQuestions": 16, "Classifier.MaxStateBytes": 256 << 10, "Classifier.MaxQuestionBytes": 8 << 10}
		for key, limit := range limits {
			h, ok := r.Headroom[key]
			valid = valid && ok && h.Limit == limit && validLoadRatio(h)
		}
		valid = valid && loadResourcesReleased(gauges)
		imageProblem := classifierImageLoadProblem(b, name, i)
		valid = valid && loadCasePassed(b, id) && imageProblem == nil
		if valid {
			e.Status = Pass
			e.Detail = fmt.Sprintf("%d concurrent calls, heap growth %d / budget %d, allocation upper bound %d; all limit use <=75%%; classifier-image design/cancel/recovery verified; resources zero", r.Load.Calls, r.HeapGrowth, r.HeapBudget, r.TotalAllocDuringCalls)
		} else {
			e.Detail = "design load, memory/limit budget, concurrency or resource-release proof invalid"
			if imageProblem != nil {
				e.Detail = imageProblem.Error()
			}
		}
		out = append(out, e)
	}
	return out
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
