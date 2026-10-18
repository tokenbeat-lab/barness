package release

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Recompute image/request readings from the actual captured JSON, rather than
// trusting a report whose positive counters could describe a text-only call.
// Captures are pretty-printed by evidence; compact restores sent wire bytes.
func classifierCaptureProblem(files verifiedCase, r classifierImageLoad, b classifierLoadBatch) error {
	var captures []provider.Request
	if readLoadJSON(files, b.Kind+"-requests.json", &captures) != nil || len(captures) != r.Load.Calls {
		return fmt.Errorf("classifier-image captures missing or count changed")
	}
	tenants := map[string]int{}
	var total int64
	for _, c := range captures {
		var wire bytes.Buffer
		if json.Compact(&wire, c.Body) != nil || int64(wire.Len()) != r.RequestJSONBytes {
			return fmt.Errorf("classifier-image whole JSON capacity inconsistent")
		}
		total += int64(wire.Len())
		var req struct {
			Model string
			Input []struct {
				Role    string
				Content []struct {
					Type, Text string
					ImageURL   string `json:"image_url"`
				}
			}
			Questions []json.RawMessage
		}
		if json.Unmarshal(c.Body, &req) != nil || req.Model != r.Identity.Model || len(req.Input) != 1 || req.Input[0].Role != "user" || len(req.Input[0].Content) != 3 || len(req.Questions) != r.Load.Questions {
			return fmt.Errorf("classifier-image captured route/input invalid")
		}
		content := req.Input[0].Content
		if content[0].Type != "input_text" || content[0].Text != strings.Repeat("s", r.Load.StateBytes-2) {
			return fmt.Errorf("classifier-image captured state shrunk")
		}
		for i, part := range content[1:] {
			enc, ok := strings.CutPrefix(part.ImageURL, "data:image/png;base64,")
			if !ok || part.Type != "input_image" || len(part.ImageURL) != r.ImageDataURLBytes[i] || len(enc) != r.ImageBase64Bytes[i] {
				return fmt.Errorf("classifier-image captured data URL inconsistent")
			}
			image, err := base64.StdEncoding.Strict().DecodeString(enc)
			if err != nil || len(image) != r.Load.ImageBytes[i] || !bytes.HasPrefix(image, []byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10}) {
				return fmt.Errorf("classifier-image captured decoded bytes inconsistent")
			}
		}
		for i, rawQuestion := range req.Questions {
			var native bytes.Buffer
			var q struct{ Type, Name, Instructions string }
			if json.Compact(&native, rawQuestion) != nil || int64(native.Len()) > r.Headroom["Classifier.MaxQuestionBytes"].Design || json.Unmarshal(rawQuestion, &q) != nil {
				return fmt.Errorf("classifier-image native question capacity inconsistent")
			}
			if q.Type != "predicate" || q.Name != fmt.Sprintf("q%d", i) || q.Instructions != strings.Repeat("q", r.Load.InstructionBytes) {
				return fmt.Errorf("classifier-image captured questions shrunk")
			}
		}
		path := strings.Split(strings.Trim(c.Path, "/"), "/")
		if c.Method != "POST" || len(path) != 3 || path[1] != "luna" || path[2] != "decisions" {
			return fmt.Errorf("classifier-image captured endpoint invalid")
		}
		tenants[path[0]]++
	}
	if len(tenants) != r.Load.Tenants || total != b.ProviderCapturedRequestBytes {
		return fmt.Errorf("classifier-image Provider copied bytes or tenants inconsistent")
	}
	for tenant, count := range tenants {
		var binding ai.Binding
		if count != 4 || readLoadJSON(files, "binding-"+tenant+".json", &binding) != nil || binding.TenantID != tenant || binding.BindingID != "luna" || binding.Operation != ai.OperationClassifier || binding.ProviderID != r.Identity.Provider || binding.API != r.Identity.API || !binding.Enabled || !binding.AllowClassifierImages || len(binding.AllowedModels) != 1 || binding.AllowedModels[0] != r.Identity.Model || binding.Retry.MaxRetries != 0 {
			return fmt.Errorf("classifier-image tenant authorization inconsistent")
		}
	}
	seenRequests := map[string]bool{}
	resultTenants := map[string]int{}
	name := "local"
	if r.Load.Tenants == 2 {
		name = "cloud"
	}
	for i := range r.Load.Calls {
		var result struct {
			Answers map[string]struct {
				Type        string
				Probability float64
			}
			Metadata   ai.CallMetadata
			StopReason ai.StopReason
		}
		if readLoadJSON(files, fmt.Sprintf("%s-result-%d.json", b.Kind, i), &result) != nil {
			return fmt.Errorf("classifier-image unary results missing")
		}
		m := result.Metadata
		expectedID := fmt.Sprintf("req-load-%s-%s-%d", name, b.Kind, i)
		if m.RequestID != expectedID || seenRequests[m.RequestID] || m.BindingID != "luna" || m.CatalogVersion != r.Identity.CatalogVersion || m.TenantID == "" {
			return fmt.Errorf("classifier-image result attribution inconsistent")
		}
		seenRequests[m.RequestID] = true
		resultTenants[m.TenantID]++
		if m.Operation != r.Identity.Operation || m.ProviderID != r.Identity.Provider || m.API != r.Identity.API || m.ModelID != r.Identity.Model || m.CatalogHash != r.Identity.CatalogHash || len(m.Attempts) != 1 || tenants[m.TenantID] != 4 {
			return fmt.Errorf("classifier-image result identity/attempt inconsistent")
		}
		attempt := m.Attempts[0]
		if attempt.AttemptID != m.RequestID+"#1" || attempt.HTTPStatus != 200 || attempt.ProviderRequestID != "up-design-load" || attempt.Code != "" || attempt.RetryDelay != 0 {
			return fmt.Errorf("classifier-image attempt attribution inconsistent")
		}
		var observations []ai.Observation
		if readLoadJSON(files, "observations-"+m.RequestID+".json", &observations) != nil || len(observations) != 4 {
			return fmt.Errorf("classifier-image Observer evidence missing")
		}
		finishes := 0
		for _, o := range observations {
			if o.Call.RequestID != m.RequestID || o.Call.TenantID != m.TenantID || o.Call.BindingID != m.BindingID || o.Call.Operation != m.Operation {
				return fmt.Errorf("classifier-image Observer attribution inconsistent")
			}
			if o.Kind == ai.ObservationCallFinished {
				finishes++
				if o.StopReason != result.StopReason || (b.Kind == "cancel" && (o.Error == nil || o.Error.Code != ai.CodeCanceled || o.Error.Phase != ai.PhaseResponse)) || (b.Kind != "cancel" && o.Error != nil) {
					return fmt.Errorf("classifier-image Observer terminal inconsistent")
				}
			}
		}
		if finishes != 1 {
			return fmt.Errorf("classifier-image Observer terminal missing")
		}
		if b.Kind == "cancel" {
			if len(result.Answers) != 0 || result.StopReason != ai.StopReasonAborted {
				return fmt.Errorf("classifier-image canceled result invalid")
			}
		} else {
			if len(result.Answers) != r.Load.Questions || result.StopReason != ai.StopReasonStop {
				return fmt.Errorf("classifier-image published results incomplete")
			}
			for q := range r.Load.Questions {
				answer, ok := result.Answers[fmt.Sprintf("q%d", q)]
				if !ok || answer.Type != "bool" || answer.Probability != .8 {
					return fmt.Errorf("classifier-image published answer changed")
				}
			}
		}
	}
	for tenant := range tenants {
		if resultTenants[tenant] != 4 {
			return fmt.Errorf("classifier-image results missing for tenant")
		}
	}
	return nil
}
