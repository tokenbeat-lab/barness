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
func classifierCaptureProblem(dir string, r classifierImageLoad, b classifierLoadBatch) error {
	var captures []provider.Request
	if readLoadJSON(dir, b.Kind+"-requests.json", &captures) != nil || len(captures) != r.Load.Calls {
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
			Questions []struct{ Type, Name, Instructions string }
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
		for i, q := range req.Questions {
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
		if count != 4 || readLoadJSON(dir, "binding-"+tenant+".json", &binding) != nil || binding.TenantID != tenant || binding.BindingID != "luna" || binding.Operation != ai.OperationClassifier || binding.ProviderID != r.Identity.Provider || binding.API != r.Identity.API || !binding.Enabled || !binding.AllowClassifierImages || len(binding.AllowedModels) != 1 || binding.AllowedModels[0] != r.Identity.Model || binding.Retry.MaxRetries != 0 {
			return fmt.Errorf("classifier-image tenant authorization inconsistent")
		}
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
		if readLoadJSON(dir, fmt.Sprintf("%s-result-%d.json", b.Kind, i), &result) != nil {
			return fmt.Errorf("classifier-image unary results missing")
		}
		m := result.Metadata
		if m.Operation != r.Identity.Operation || m.ProviderID != r.Identity.Provider || m.API != r.Identity.API || m.ModelID != r.Identity.Model || m.CatalogHash != r.Identity.CatalogHash || len(m.Attempts) != 1 || tenants[m.TenantID] != 4 {
			return fmt.Errorf("classifier-image result identity/attempt inconsistent")
		}
		if b.Kind == "cancel" {
			if len(result.Answers) != 0 || result.StopReason != ai.StopReasonAborted {
				return fmt.Errorf("classifier-image canceled result invalid")
			}
		} else {
			if len(result.Answers) != r.Load.Questions || result.StopReason != ai.StopReasonStop {
				return fmt.Errorf("classifier-image published results incomplete")
			}
			for _, answer := range result.Answers {
				if answer.Type != "bool" || answer.Probability != .8 {
					return fmt.Errorf("classifier-image published answer changed")
				}
			}
		}
	}
	return nil
}
