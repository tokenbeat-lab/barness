package release

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// MultimodalImageEvidence imports ticket 09's four simultaneous two-tenant,
// two-protocol calls separately for each public entry. It augments the original
// local/cloud loads; it never substitutes this smaller batch for their budgets.
func MultimodalImageEvidence(b Bundle) []EvidenceCheck {
	var out []EvidenceCheck
	for _, entry := range []string{"plain", "hooked"} {
		e := EvidenceCheck{Name: "multimodal-images-" + entry, Bundle: b.Dir, Status: NotRun, Detail: "required public-entry mixed image load missing"}
		id := "UC09-design-load-" + entry
		if dir := b.caseDirs[id]; dir != "" {
			e.Status = Fail
			if len(b.Problems) != 0 || !loadCasePassed(b, id) {
				e.Detail = "mixed image case failed or artifact inventory invalid"
			} else if err := verifyMultimodalLoad(dir, id, entry); err != nil {
				e.Detail = err.Error()
			} else {
				e.Status, e.Detail = Pass, "four simultaneous calls, two tenants and protocols; 64 MiB memory ceiling, <=75% limit use, both final results/usage/Observer and zero resources; controlled HTTP only"
			}
		}
		out = append(out, e)
	}
	return out
}

func verifyMultimodalLoad(dir, id, entry string) error {
	bad := func(file string) error { return fmt.Errorf("%s: mixed image contract missing or inconsistent", file) }
	var assertions struct {
		ID         string `json:"case_id"`
		Status     string
		Assertions []struct{ Pass bool }
	}
	if readLoadJSON(dir, "assertions.json", &assertions) != nil || assertions.ID != id || assertions.Status != "PASS" || len(assertions.Assertions) == 0 {
		return bad("assertions.json")
	}
	for _, a := range assertions.Assertions {
		if !a.Pass {
			return bad("assertions.json")
		}
	}
	var load struct {
		Calls, Tenants, Protocols, SourceImagesPerCall, JevOccurrences, DecisionsOccurrences, HeldBodiesAtEOF int
		MemoryCeiling, TotalAlloc, TotalAllocAtEOF                                                            uint64
		HTTPBytesAtEOF                                                                                        int64
		RealService                                                                                           string
		Heap                                                                                                  struct{ BaselineLive, PeakLive uint64 }
		Headroom                                                                                              map[string]loadLimit
	}
	if readLoadJSON(dir, "load.json", &load) != nil || load.Calls != 4 || load.Tenants != 2 || load.Protocols != 2 || load.SourceImagesPerCall != 2 || load.JevOccurrences != 6 || load.DecisionsOccurrences != 3 || load.HeldBodiesAtEOF != 4 || load.MemoryCeiling != 64<<20 || load.TotalAllocAtEOF == 0 || load.TotalAlloc < load.TotalAllocAtEOF || load.TotalAlloc >= load.MemoryCeiling || load.HTTPBytesAtEOF <= 0 || load.RealService != "NOT_RUN" || load.Heap.PeakLive < load.Heap.BaselineLive || load.Heap.PeakLive-load.Heap.BaselineLive >= load.MemoryCeiling || len(load.Headroom) != 7 {
		return bad("load.json")
	}
	for key, h := range load.Headroom {
		limit := int64(64 << 10)
		if key == "expanded" {
			limit = 128 << 10
		} else if key == "references" {
			limit = 8
		}
		if h.Limit != limit || !validLoadRatio(h) {
			return bad("load.json limits")
		}
	}
	if load.Headroom["references"].Design != 6 || load.Headroom["source"].Design <= 0 || load.Headroom["expanded"].Design <= 0 {
		return bad("load.json")
	}
	var resources map[string]*int64
	if readLoadJSON(dir, "resources.json", &resources) != nil || !loadResourcesReleased(resources) {
		return bad("resources.json")
	}
	var config struct {
		Catalog ai.Catalog
		Policy  ai.ResourcePolicy
	}
	if readLoadJSON(dir, "config.json", &config) != nil {
		return bad("config.json")
	}
	p := config.Policy
	if p.MaxRequestBytes != 64<<10 || p.MaxOutputBytes != 16<<10 || p.MaxImageBytes != 32<<10 || p.MaxConcurrentProcess != 4 || p.MaxConcurrentPerTenant != 2 || p.Classifier == nil || p.Classifier.Images == nil || p.Classifier.Images.MaxImages != 2 || p.Classifier.Images.MaxTotalImageBytes != 64<<10 || p.Classifier.Images.References == nil || p.Classifier.Images.References.MaxReferences != 8 || p.Classifier.Images.References.MaxExpandedImageBytes != 128<<10 {
		return bad("config.json")
	}
	hash, err := config.Catalog.Hash()
	if err != nil {
		return bad("config.json")
	}
	var captures []provider.Request
	if readLoadJSON(dir, "requests.json", &captures) != nil || len(captures) != 4 {
		return bad("requests.json")
	}
	seen := map[string]bool{}
	for _, c := range captures {
		parts := strings.Split(c.Path, "/")
		if len(parts) != 4 || (parts[1] != "tenant-a" && parts[1] != "tenant-b") || seen[c.Path] || c.Method != "POST" || c.Query != "" || c.KeyAlias != "key:"+parts[1]+"@v1" {
			return bad("requests.json")
		}
		seen[c.Path] = true
		jev := parts[2] == "classifier" && parts[3] == "systemone"
		if !jev && !(parts[2] == "luna" && parts[3] == "decisions") {
			return bad("requests.json")
		}
		var compact bytes.Buffer
		if json.Compact(&compact, c.Body) != nil {
			return bad("requests.json")
		}
		h, ok := load.Headroom["request-"+c.KeyAlias+c.Path]
		if !ok || h.Design != int64(compact.Len()) {
			return bad("load.json request bytes")
		}
		call := "req-load-" + parts[1] + parts[2] + entry
		if err := verifyMultimodalResult(dir, call, parts[1], parts[2], hash, jev, c); err != nil {
			return err
		}
	}
	return nil
}

func verifyMultimodalResult(dir, call, tenant, binding, hash string, jev bool, capture provider.Request) error {
	bad := fmt.Errorf("mixed image input/result/Attempt/Observer reconciliation failed")
	var input ai.ClassifierRequest
	var result struct {
		Answers     map[string]json.RawMessage
		ImageInputs []ai.ClassifierImageInput
		Usage       ai.Usage
		StopReason  ai.StopReason
		Metadata    ai.CallMetadata
	}
	if readLoadJSON(dir, call+"-input.json", &input) != nil || readLoadJSON(dir, call+"-result.json", &result) != nil || len(input.Images) != 2 || len(result.Answers) != len(input.Questions) || result.StopReason != ai.StopReasonStop || result.Usage.Input <= 0 || result.Usage.Cost.Total <= 0 {
		return bad
	}
	model, api, providerID, count := "host-decisions", ai.APIOpenAIDecisions, ai.ProviderOpenAI, 3
	order := []int{1, 0, 1}
	if jev {
		model, api, providerID, count = "jev-1.14.0", ai.APITypeSafeSystemOne, ai.ProviderTypeSafe, 6
		order = []int{1, 0}
	}
	m := result.Metadata
	if len(input.ImageReferences) != count || len(result.ImageInputs) != len(order) || m.RequestID != call || m.TenantID != tenant || m.BindingID != binding || m.ModelID != model || m.API != api || m.ProviderID != providerID || m.Operation != ai.OperationClassifier || m.CatalogHash != hash || len(m.Attempts) != 1 || m.Attempts[0].HTTPStatus != 200 || m.Attempts[0].UsageReporting != ai.UsageComplete {
		return bad
	}
	for i, index := range order {
		raw, err := base64.StdEncoding.Strict().DecodeString(input.Images[index].Data)
		if err != nil || result.ImageInputs[i].MimeType != input.Images[index].MimeType || result.ImageInputs[i].SHA256 != strings.TrimPrefix(pioracle.SHA256(raw), "sha256:") {
			return bad
		}
	}
	if err := verifyRevocationObservations(dir, call, revocationResult{result.Usage, result.StopReason, m}, ""); err != nil {
		return err
	}
	return verifyMultimodalCapture(dir, call, tenant, binding, capture, result.Answers, result.Usage)
}
