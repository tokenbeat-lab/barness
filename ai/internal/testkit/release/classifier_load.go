package release

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
)

type loadLimit struct {
	Design, Limit int64
	Ratio         float64
}
type classifierImageLoad struct {
	Schema   int
	Identity struct {
		Operation                          ai.Operation
		Provider                           ai.ProviderID
		API                                ai.API
		Model, CatalogVersion, CatalogHash string
	}
	Load struct {
		Calls, Tenants, CallsPerTenant, StateBytes, Questions, InstructionBytes, ImageCount, TotalImageBytes int
		ImageBytes                                                                                           []int
	}
	Policy                              ai.ResourcePolicy
	Headroom                            map[string]loadLimit
	HeapBudget                          uint64
	Batches                             []classifierLoadBatch
	ImageBase64Bytes, ImageDataURLBytes []int
	RequestJSONBytes                    int64
	ProviderSharedResponseScriptBytes   int64
}
type classifierLoadBatch struct {
	Kind                                                                                                 string
	Calls                                                                                                int
	Successful, Canceled                                                                                 *int
	PeakPermits, PermitsAtEOF, ClientUnaryBytesAtEOF, ResponseBytesPerCall, ProviderCapturedRequestBytes int64
	Heap                                                                                                 struct{ BaselineLive, PeakLive uint64 }
	HeapGrowth, HeapAtEOF, HeapWithResults, TotalAllocDuringCalls                                        uint64
	Duration                                                                                             string
	CallsPerSecond                                                                                       float64
	Resources                                                                                            map[string]*int64
}

// This is issue 08's standalone capacity import, not the final ten-route gate.
// It consumes the original mixed batch AND the added bounded image batches.
// Old Jev artifacts cannot attest classifier images. Limits are pinned to the
// actual Host example; report counters are cross-checked against wire captures.
func classifierImageLoadProblem(b Bundle, name string, index int) error {
	id := "E08-mixed-pressure-" + name + "-classifier-images-design-load"
	dir := b.caseDirs[id]
	if dir == "" || !loadCasePassed(b, id) {
		return fmt.Errorf("classifier-image design-load case missing or not PASS")
	}
	var r classifierImageLoad
	if err := readLoadJSON(dir, "classifier-images-pressure.json", &r); err != nil {
		return fmt.Errorf("classifier-image capacity report missing or invalid")
	}
	want := localassembly.MixedPolicy()
	if index == 1 {
		want = hostintegration.CloudMixedPolicy()
	}
	l := r.Load
	if r.Schema != 1 || r.Identity.Operation != ai.OperationClassifier || r.Identity.Provider != ai.ProviderOpenAI || r.Identity.API != ai.APIOpenAIDecisions || r.Identity.Model != "gpt-6-luna" || r.Identity.CatalogVersion != "classifier-load-08-synthetic" {
		return fmt.Errorf("classifier-image route identity invalid")
	}
	if l.Calls != 4*(index+1) || l.Tenants != index+1 || l.CallsPerTenant != 4 || l.StateBytes != 128<<10 || l.Questions != 8 || l.InstructionBytes != 4000 || l.ImageCount != 2 || l.TotalImageBytes != 512<<10 || !reflect.DeepEqual(l.ImageBytes, []int{256 << 10, 256 << 10}) {
		return fmt.Errorf("classifier-image design load changed")
	}
	var policy ai.ResourcePolicy
	if readLoadJSON(dir, "policy.json", &policy) != nil || !reflect.DeepEqual(r.Policy, *want) || !reflect.DeepEqual(policy, *want) {
		return fmt.Errorf("classifier-image Host policy changed")
	}
	var cat ai.Catalog
	if readLoadJSON(dir, "catalog.json", &cat) != nil {
		return fmt.Errorf("classifier-image catalog missing")
	}
	hash, err := cat.Hash()
	if err != nil || hash != r.Identity.CatalogHash || cat.Version != r.Identity.CatalogVersion {
		return fmt.Errorf("classifier-image catalog identity changed")
	}
	modelOK := false
	for _, m := range cat.ClassifierModels {
		if m.ID == r.Identity.Model && m.Provider == r.Identity.Provider && m.API == r.Identity.API {
			modelOK = m.Capabilities.MaxInputImages >= l.ImageCount && reflect.DeepEqual(m.Input, []ai.Modality{ai.ModalityText, ai.ModalityImage})
		}
	}
	if !modelOK {
		return fmt.Errorf("classifier-image model capability missing")
	}
	enc := base64.StdEncoding.EncodedLen(256 << 10)
	if !reflect.DeepEqual(r.ImageBase64Bytes, []int{enc, enc}) || !reflect.DeepEqual(r.ImageDataURLBytes, []int{enc + 22, enc + 22}) || r.RequestJSONBytes <= int64(2*enc+l.StateBytes+l.Questions*l.InstructionBytes) || r.ProviderSharedResponseScriptBytes <= 0 {
		return fmt.Errorf("classifier-image encoded capacity invalid")
	}
	if r.HeapBudget != uint64(256<<20)*uint64(index+1) || len(r.Batches) != 3 {
		return fmt.Errorf("classifier-image budget or bounded batches changed")
	}
	questionBytes := int64(len(`{"type":"predicate","name":"q0","instructions":""}`) + l.InstructionBytes)
	expected := map[string]loadLimit{
		"Classifier.Images.MaxImages": {2, 4, .5}, "MaxImageBytes": {256 << 10, 1 << 20, .25},
		"Classifier.Images.MaxTotalImageBytes": {512 << 10, 1 << 20, .5},
		"MaxRequestBytes":                      {r.RequestJSONBytes, 4 << 20, 0},
		"Classifier.MaxStateBytes":             {128 << 10, 256 << 10, .5}, "Classifier.MaxQuestions": {8, 16, .5},
		"Classifier.MaxQuestionBytes": {questionBytes, 8 << 10, 0}, "MaxOutputBytes": {r.ProviderSharedResponseScriptBytes, 8 << 20, 0},
	}
	for key, h := range expected {
		v, ok := r.Headroom[key]
		if !ok || v.Design != h.Design || v.Limit != h.Limit || !validLoadRatio(v) {
			return fmt.Errorf("classifier-image measured limit use invalid: %s", key)
		}
	}
	var script struct {
		Status int
		Chunks [][]byte
	}
	if readLoadJSON(dir, "response-script.json", &script) != nil || script.Status != 200 || len(script.Chunks) != 1 || int64(len(script.Chunks[0])) != r.ProviderSharedResponseScriptBytes {
		return fmt.Errorf("classifier-image response script missing or inconsistent")
	}
	for i, kind := range []string{"design", "cancel", "recovery"} {
		batch := r.Batches[i]
		if batch.Successful == nil || batch.Canceled == nil {
			return fmt.Errorf("classifier-image outcome counts missing")
		}
		if batch.Kind != kind || batch.Calls != l.Calls || batch.PeakPermits != int64(l.Calls) || batch.PermitsAtEOF != int64(l.Calls) || batch.ResponseBytesPerCall != r.ProviderSharedResponseScriptBytes || batch.ClientUnaryBytesAtEOF != int64(l.Calls)*batch.ResponseBytesPerCall {
			return fmt.Errorf("classifier-image concurrency or complete body proof invalid")
		}
		if kind == "cancel" {
			if *batch.Successful != 0 || *batch.Canceled != l.Calls {
				return fmt.Errorf("classifier-image cancellation missing")
			}
		} else if *batch.Successful != l.Calls || *batch.Canceled != 0 {
			return fmt.Errorf("classifier-image complete/recovery results missing")
		}
		// Captured request copies and the Client's complete unary buffers are
		// simultaneously reachable at EOF and were allocated after baseline.
		// A reported upper bound cannot be smaller than those known bytes.
		copyFloor := uint64(int64(l.Calls) * r.RequestJSONBytes)
		allocationFloor := copyFloor + uint64(batch.ClientUnaryBytesAtEOF)
		h := batch.Heap
		if h.BaselineLive == 0 || h.PeakLive < h.BaselineLive || batch.HeapGrowth != h.PeakLive-h.BaselineLive || batch.HeapGrowth > r.HeapBudget || batch.HeapAtEOF == 0 || batch.HeapWithResults == 0 || h.PeakLive < batch.HeapAtEOF || h.PeakLive < batch.HeapWithResults || batch.HeapAtEOF < allocationFloor || batch.HeapWithResults < copyFloor || batch.TotalAllocDuringCalls < allocationFloor || batch.TotalAllocDuringCalls < batch.HeapGrowth || batch.TotalAllocDuringCalls > r.HeapBudget {
			return fmt.Errorf("classifier-image sampled memory or allocation upper bound invalid")
		}
		d, err := time.ParseDuration(batch.Duration)
		rate := float64(l.Calls) / d.Seconds()
		if err != nil || d <= 0 || batch.CallsPerSecond <= 0 || math.IsInf(batch.CallsPerSecond, 0) || math.IsNaN(batch.CallsPerSecond) || math.Abs(batch.CallsPerSecond-rate) > rate*1e-9 || !loadResourcesReleased(batch.Resources) {
			return fmt.Errorf("classifier-image throughput or release proof invalid")
		}
		if err := classifierCaptureProblem(dir, r, batch); err != nil {
			return err
		}
	}
	var resources map[string]*int64
	if readLoadJSON(dir, "resources.json", &resources) != nil || !loadResourcesReleased(resources) {
		return fmt.Errorf("classifier-image final resources missing or nonzero")
	}
	return nil
}

func validLoadRatio(v loadLimit) bool {
	return v.Design > 0 && v.Limit > 0 && v.Ratio > 0 && v.Ratio <= .75 && !math.IsNaN(v.Ratio) && math.Abs(v.Ratio-float64(v.Design)/float64(v.Limit)) <= 1e-9
}
func loadResourcesReleased(resources map[string]*int64) bool {
	for _, key := range []string{"calls", "permits", "waiters", "events", "bodies", "hostPermits"} {
		n, ok := resources[key]
		if !ok || n == nil || *n != 0 {
			return false
		}
	}
	return true
}
func loadCasePassed(b Bundle, id string) bool {
	for _, c := range b.Cases {
		if c.ID == id {
			return c.Status == Pass
		}
	}
	return false
}
func readLoadJSON(dir, name string, v any) error {
	raw, err := readArtifact(filepath.Join(dir, name), 16<<20)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
