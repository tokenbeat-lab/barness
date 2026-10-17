package release

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// MiniMax adds six bounded batches alongside the unchanged original4/8 loads.
// Missing/failed cases, altered capacities, report-only PASS, leaked resources,
// wrong own protocol/model, incomplete captures/results or Observer terminals
// must fail. The normal and race bundles independently consume these artifacts.
type miniMaxLoad struct {
	designLoad
	BaselineCase                                                                string
	BaselineCallsUnchanged, BaselinePermitOccupancy, AdditionalCalls, QueuePeak int
	MiniMaxProtocol                                                             string
	OriginalRoutes                                                              []string
}

func miniMaxLoadNames() []string {
	var names []string
	for _, host := range []string{"local", "cloud"} {
		for _, route := range []string{"chat", "responses", "messages"} {
			names = append(names, "minimax-design-load-"+host+"-"+route)
		}
	}
	return names
}

// MiniMaxDesignLoadEvidence verifies the additional actual captures without
// replacing or reducing DesignLoadEvidence's original capacity probes.
func MiniMaxDesignLoadEvidence(b Bundle) []EvidenceCheck {
	var out []EvidenceCheck
	for host, policy := range map[string]*ai.ResourcePolicy{"local": localassembly.MixedPolicy(), "cloud": hostintegration.CloudMixedPolicy()} {
		tenants := 1
		if host == "cloud" {
			tenants = 2
		}
		for phase, route := range []string{"chat", "responses", "messages"} {
			id := "MM05-design-load-" + host + "-" + route
			e := EvidenceCheck{Name: "minimax-design-load-" + host + "-" + route, Bundle: b.Dir, Status: NotRun, Detail: "required additional own design-load case missing"}
			if dir := b.caseDirs[id]; dir != "" {
				e.Status = Fail
				if err := miniMaxLoadProblem(b, id, dir, host, route, phase, tenants, policy); err == nil {
					e.Status, e.Detail = Pass, "own protocol and original routes; full3/6-call captures/results/Observer; memory and75percent headroom; resources zero"
				} else {
					e.Detail = err.Error()
				}
			}
			out = append(out, e)
		}
	}
	slices.SortFunc(out, func(a, b EvidenceCheck) int { return strings.Compare(a.Name, b.Name) })
	return out
}

func miniMaxLoadProblem(b Bundle, id, dir, host, route string, phase, tenants int, policy *ai.ResourcePolicy) error {
	var r miniMaxLoad
	var recorded ai.ResourcePolicy
	var gauges map[string]*int64
	if len(b.Problems) > 0 || !loadCasePassed(b, id) || readLoadJSON(dir, "design-load.json", &r) != nil || readLoadJSON(dir, "policy.json", &recorded) != nil || readLoadJSON(dir, "resources.json", &gauges) != nil {
		return fmt.Errorf("MiniMax load report, inventory, status or resources missing")
	}
	expectedRoutes := [][]string{{"chat", "openai-image"}, {"openai-image", "google-image"}, {"google-image", "classifier"}}[phase]
	l := r.Load
	if !reflect.DeepEqual(recorded, *policy) || r.BaselineCase != "E08-mixed-pressure-"+host+"-design-load" || r.BaselineCallsUnchanged != 4*tenants || r.BaselinePermitOccupancy != 1 || r.AdditionalCalls != 3*tenants || r.MiniMaxProtocol != route || !slices.Equal(r.OriginalRoutes, expectedRoutes) || l.Calls != 4*tenants || l.Tenants != tenants || l.HistoryBytes != 128<<10 || l.ChatOutputBytes != 16<<10 || l.References != 2 || l.InputImageBytes != 256<<10 || l.MaskBytes != 256<<10 || l.OutputImages != 2 || l.OutputImageBytes != 512<<10 || l.StateBytes != 128<<10 || l.Questions != 8 || l.InstructionBytes != 4000 {
		return fmt.Errorf("MiniMax load identity, original baseline, policy or workload changed")
	}
	if r.PeakPermits != int64(3*tenants) || r.QueuePeak != 0 || r.HeapBudget != uint64(256<<20)*uint64(tenants) || r.Heap.BaselineLive == 0 || r.Heap.PeakLive < r.Heap.BaselineLive || r.Heap.PeakLive-r.Heap.BaselineLive > r.HeapBudget || r.TotalAllocDuringCalls == 0 || r.TotalAllocDuringCalls > r.HeapBudget || !loadResourcesReleased(gauges) {
		return fmt.Errorf("MiniMax load concurrency, memory or release proof invalid")
	}
	limits := map[string]int64{"MaxRequestBytes": policy.MaxRequestBytes, "MaxFrameBytes": policy.MaxFrameBytes, "MaxOutputBytes": policy.MaxOutputBytes, "MaxImageBytes": policy.MaxImageBytes, "MaxToolJSONBytes": policy.MaxToolJSONBytes, "MaxErrorBodyBytes": policy.MaxErrorBodyBytes, "MaxQueuedEvents": int64(3 * tenants * policy.MaxQueuedEvents), "MaxQueuedEventBytes": policy.MaxQueuedEventBytes, "MaxConcurrentPerTenant": int64(policy.MaxConcurrentPerTenant), "MaxConcurrentProcess": int64(policy.MaxConcurrentProcess), "MaxAdmissionWaiters": int64(policy.MaxAdmissionWaiters), "MaxQueuedObservationsConservativeAllRecords": int64(policy.MaxQueuedObservations), "Image.MaxInputImages": 4, "Image.MaxOutputImages": 4, "Image.MaxOutputImageBytes": 1 << 20, "Image.MaxTotalOutputImageBytes": 2 << 20, "Classifier.MaxQuestions": 16, "Classifier.MaxStateBytes": 256 << 10, "Classifier.MaxQuestionBytes": 8 << 10}
	if len(r.Headroom) != len(limits) {
		return fmt.Errorf("MiniMax load complete headroom inventory missing")
	}
	for name, limit := range limits {
		h, ok := r.Headroom[name]
		if !ok || h.Design < 0 || h.Limit != limit || h.Ratio < 0 || h.Ratio > .75 || math.IsNaN(h.Ratio) || math.IsInf(h.Ratio, 0) || math.Abs(h.Ratio-float64(h.Design)/float64(limit)) > 1e-9 {
			return fmt.Errorf("MiniMax load headroom invalid: %s", name)
		}
	}

	// Zero queues are valid for Complete, but fixed dimensions and controls must
	// still agree with actual workload sizes rather than editable small ratios.
	designs := map[string]int64{"MaxImageBytes": 256 << 10, "MaxToolJSONBytes": 0, "MaxErrorBodyBytes": 0, "MaxQueuedEventBytes": 0, "MaxAdmissionWaiters": 0, "Image.MaxInputImages": 3, "Image.MaxOutputImages": 2, "Image.MaxOutputImageBytes": 512 << 10, "Image.MaxTotalOutputImageBytes": 1 << 20, "Classifier.MaxQuestions": 8, "Classifier.MaxStateBytes": 128 << 10, "Classifier.MaxQuestionBytes": 4128}
	for name, want := range designs {
		if r.Headroom[name].Design != want {
			return fmt.Errorf("MiniMax load measured dimension changed: %s", name)
		}
	}
	if r.Headroom["MaxConcurrentProcess"].Design != int64(3*tenants) || r.Headroom["MaxConcurrentPerTenant"].Design != 3 || r.Headroom["MaxQueuedEvents"].Design != 0 || r.Headroom["MaxQueuedObservationsConservativeAllRecords"].Design != int64(12*tenants) {
		return fmt.Errorf("MiniMax load concurrency/queue readings changed")
	}
	var captures []provider.Request
	if readLoadJSON(dir, "requests.json", &captures) != nil || len(captures) != 3*tenants {
		return fmt.Errorf("MiniMax load complete requests missing")
	}
	seen := map[string]bool{}
	var largest, total int64
	for _, capture := range captures {
		parts := strings.Split(strings.Trim(capture.Path, "/"), "/")
		if capture.Method != "POST" || len(parts) < 3 || (parts[0] != "tenant-a" && (tenants != 2 || parts[0] != "tenant-b")) {
			return fmt.Errorf("MiniMax load captured request attribution invalid")
		}
		binding := parts[1]
		key := parts[0] + "/" + binding
		if seen[key] || binding != "minimax-"+route && !slices.Contains(expectedRoutes, binding) {
			return fmt.Errorf("MiniMax load captured route duplicated or changed")
		}
		seen[key] = true
		var wire bytes.Buffer
		if json.Compact(&wire, capture.Body) != nil {
			return fmt.Errorf("MiniMax load request JSON invalid")
		}
		largest = max(largest, int64(wire.Len()))
		total += int64(wire.Len())
		if err := miniMaxLoadRequest(capture.Body, binding); err != nil {
			return err
		}
		if err := miniMaxLoadResult(dir, host, route, parts[0], binding); err != nil {
			return err
		}
	}
	if r.Headroom["MaxRequestBytes"].Design != largest || r.TotalAllocDuringCalls < uint64(total) {
		return fmt.Errorf("MiniMax load measured request/allocation capacity inconsistent")
	}
	var largestResponse int64
	for _, binding := range append(slices.Clone(expectedRoutes), "minimax-"+route) {
		name := "response-old-" + binding + ".json"
		if strings.HasPrefix(binding, "minimax-") {
			name = "response-" + route + ".json"
		}
		var reply provider.Reply
		if err := readLoadJSON(dir, name, &reply); err != nil || reply.Status != 200 || len(reply.Script) == 0 {
			return fmt.Errorf("MiniMax load response script missing: %s (status=%d bytes=%d errorType=%T)", name, reply.Status, len(reply.Script), err)
		}
		size := int64(len(reply.Script))
		largestResponse = max(largestResponse, size)
		if !strings.Contains(binding, "image") {
			for _, frame := range strings.Split(reply.Script, "\n\n") {
				if frame != "" && int64(len(frame)) > r.Headroom["MaxFrameBytes"].Design {
					return fmt.Errorf("MiniMax load measured frame capacity understated")
				}
			}
		}
	}
	if r.Headroom["MaxOutputBytes"].Design != largestResponse {
		return fmt.Errorf("MiniMax load measured output capacity inconsistent")
	}
	results, err := filepath.Glob(filepath.Join(dir, "*-result.json"))
	if err != nil || len(results) != 3*tenants {
		return fmt.Errorf("MiniMax load result inventory changed")
	}
	return nil
}

func miniMaxLoadRequest(raw []byte, binding string) error {
	var doc map[string]any
	if json.Unmarshal(raw, &doc) != nil {
		return fmt.Errorf("MiniMax load request missing")
	}
	model := "same-model"
	if strings.HasPrefix(binding, "minimax-") {
		model = "MiniMax-M3"
	}
	if doc["model"] != model {
		return fmt.Errorf("MiniMax load captured model changed")
	}
	texts := []string{}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case string:
			texts = append(texts, x)
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(doc)
	switch binding {
	case "chat", "minimax-chat", "minimax-responses", "minimax-messages":
		if !slices.Contains(texts, strings.Repeat("h", 128<<10)) {
			return fmt.Errorf("MiniMax load complete history shrunk")
		}
	case "classifier":
		if doc["state"] != strings.Repeat("s", (128<<10)-2) {
			return fmt.Errorf("MiniMax load classifier state shrunk")
		}
		q, ok := doc["questions"].(map[string]any)
		if !ok || len(q) != 8 {
			return fmt.Errorf("MiniMax load classifier questions shrunk")
		}
		count := 0
		for _, s := range texts {
			if s == strings.Repeat("q", 4000) {
				count++
			}
		}
		if count != 8 {
			return fmt.Errorf("MiniMax load classifier instructions shrunk")
		}
	default:
		images := 0
		for _, s := range texts {
			enc := strings.TrimPrefix(s, "data:image/png;base64,")
			if !strings.HasPrefix(enc, "iVBOR") {
				continue
			}
			image, e := base64.StdEncoding.Strict().DecodeString(enc)
			if e != nil || len(image) != 256<<10 {
				return fmt.Errorf("MiniMax load input image bytes changed")
			}
			images++
		}
		expected := 2
		if binding == "openai-image" {
			expected = 3
		}
		if images != expected {
			return fmt.Errorf("MiniMax load input images/mask missing")
		}
	}
	return nil
}

func miniMaxLoadResult(dir, host, route, tenant, binding string) error {
	id := "mm05-load-" + host + "-" + route + "-old-" + binding + "-" + tenant
	if binding == "minimax-"+route {
		id = "mm05-load-" + host + route + tenant
	}
	var result struct {
		Metadata ai.CallMetadata
		Usage    ai.Usage
		Message  struct {
			Content    []struct{ Type, Text string }
			StopReason ai.StopReason
			Usage      ai.Usage
		}
		Content    []struct{ Type, Data string }
		Answers    map[string]json.RawMessage
		StopReason ai.StopReason
	}
	var meta ai.CallMetadata
	var callError string
	if readLoadJSON(dir, id+"-result.json", &result) != nil || readLoadJSON(dir, id+"-metadata.json", &meta) != nil || readLoadJSON(dir, id+"-error.json", &callError) != nil || callError != "" || !reflect.DeepEqual(meta, result.Metadata) || meta.RequestID != id || meta.TenantID != tenant || meta.BindingID != binding || len(meta.Attempts) != 1 || meta.Attempts[0].AttemptID != id+"#1" || meta.Attempts[0].HTTPStatus != 200 || meta.Attempts[0].Code != "" || meta.Attempts[0].RetryDelay != 0 {
		return fmt.Errorf("MiniMax load own result/attempt/attribution missing")
	}
	model := "same-model"
	providerID := ai.ProviderOpenAI
	api := ai.APIOpenAIResponses
	operation := ai.OperationChat
	if strings.HasPrefix(binding, "minimax-") {
		model, providerID = "MiniMax-M3", ai.ProviderMiniMaxCN
		api = map[string]ai.API{"chat": ai.APIOpenAICompletions, "responses": ai.APIOpenAIResponses, "messages": ai.APIAnthropicMessages}[route]
	} else if binding == "openai-image" {
		api, operation = ai.APIOpenAIImages, ai.OperationImage
	} else if binding == "google-image" {
		api, providerID, operation = ai.APIGoogleInteractions, ai.ProviderGoogle, ai.OperationImage
	} else if binding == "classifier" {
		api, providerID, operation = ai.APITypeSafeSystemOne, ai.ProviderTypeSafe, ai.OperationClassifier
	}
	if meta.ModelID != model || meta.ProviderID != providerID || meta.API != api || meta.Operation != operation || !catalogHash.MatchString(meta.CatalogHash) {
		return fmt.Errorf("MiniMax load result identity changed")
	}
	stop := result.StopReason
	usage := result.Usage
	if operation == ai.OperationChat {
		text := ""
		for _, block := range result.Message.Content {
			text += block.Text
		}
		if len(text) != 16<<10 || result.Message.StopReason != ai.StopReasonStop {
			return fmt.Errorf("MiniMax load complete chat output missing")
		}
		stop, usage = result.Message.StopReason, result.Message.Usage
	} else if operation == ai.OperationImage {
		if len(result.Content) != 2 {
			return fmt.Errorf("MiniMax load whole image results missing")
		}
		for _, image := range result.Content {
			raw, e := base64.StdEncoding.Strict().DecodeString(image.Data)
			if e != nil || image.Type != "image" || len(raw) != 512<<10 {
				return fmt.Errorf("MiniMax load whole image bytes missing")
			}
		}
	} else if len(result.Answers) != 8 {
		return fmt.Errorf("MiniMax load complete classifier results missing")
	}
	var obs []ai.Observation
	if readLoadJSON(dir, "observations-"+id+".json", &obs) != nil || len(obs) != 4 {
		return fmt.Errorf("MiniMax load Observer evidence missing")
	}
	kinds := map[ai.ObservationKind]int{}
	for _, o := range obs {
		kinds[o.Kind]++
		if o.Call.RequestID != id || o.Call.TenantID != tenant || o.Call.BindingID != binding || o.Call.Operation != operation {
			return fmt.Errorf("MiniMax load Observer attribution inconsistent")
		}
		if o.Kind == ai.ObservationCallFinished && (!reflect.DeepEqual(o.Call, meta) || !reflect.DeepEqual(o.Usage, usage) || o.StopReason != stop || o.Error != nil) {
			return fmt.Errorf("MiniMax load Observer terminal inconsistent")
		}
		if o.Kind == ai.ObservationAttemptFinished && (o.Attempt == nil || !reflect.DeepEqual(*o.Attempt, meta.Attempts[0])) {
			return fmt.Errorf("MiniMax load observed attempt inconsistent")
		}
	}
	for _, kind := range []ai.ObservationKind{ai.ObservationCallStarted, ai.ObservationAttemptStarted, ai.ObservationAttemptFinished, ai.ObservationCallFinished} {
		if kinds[kind] != 1 {
			return fmt.Errorf("MiniMax load Observer lifecycle incomplete")
		}
	}
	return nil
}
