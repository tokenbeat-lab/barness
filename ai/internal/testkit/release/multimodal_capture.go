package release

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

//go:embed testdata/multimodal-load.json
var multimodalFixture []byte

// These literal expectations describe only the finite controlled UC09 workload.
// Template binding substitutes exact sentinel nodes, never substrings or user
// text. They cannot grant a native payload additional authority.
func verifyMultimodalCapture(files verifiedCase, call, tenant, binding string, capture provider.Request, answers map[string]json.RawMessage, usage ai.Usage) error {
	bad := errors.New("mixed image captured input/wire/response/named answers differ from fixed workload")
	var fixture map[string]json.RawMessage
	if json.Unmarshal(multimodalFixture, &fixture) != nil {
		return bad
	}
	var spec struct{ Input, Request, Response, Answers, Usage any }
	if json.Unmarshal(fixture[binding], &spec) != nil {
		return bad
	}
	var input ai.ClassifierRequest
	if readLoadJSON(files, call+"-input.json", &input) != nil || len(input.Images) != 2 {
		return bad
	}
	var hashes []string
	if json.Unmarshal(fixture["imageHashes"], &hashes) != nil || len(hashes) != 2 {
		return bad
	}
	if tenant == "tenant-b" {
		hashes[0], hashes[1] = hashes[1], hashes[0]
	}
	state := map[string]any{"owner": tenant, "photo": nil}
	stateText, _ := json.Marshal(state)
	values := map[string]any{"TENANT": tenant, "STATE_TEXT": string(stateText), "STATE_JSON": state}
	for i, img := range input.Images {
		raw, err := base64.StdEncoding.Strict().DecodeString(img.Data)
		if err != nil || pioracle.SHA256(raw) != hashes[i] {
			return bad
		}
		index := string(rune('0' + i))
		values["IMAGE"+index+"_DATA"] = img.Data
		values["IMAGE"+index+"_URL"] = "data:" + img.MimeType + ";base64," + img.Data
		values["IMAGE"+index+"_FORMAT"] = strings.TrimPrefix(img.MimeType, "image/")
	}
	if binding == "classifier" {
		var native struct {
			Assets map[string]struct {
				Data string `json:"data_base64"`
			}
		}
		if json.Unmarshal(capture.Body, &native) != nil || len(native.Assets) != 2 {
			return bad
		}
		for i, img := range input.Images {
			matches := 0
			for key, asset := range native.Assets {
				if asset.Data == img.Data {
					values["IMAGE"+string(rune('0'+i))+"_KEY"] = key
					matches++
				}
			}
			if matches != 1 {
				return bad
			}
		}
	}
	// Source bytes are checked above; compare every other public input field.
	raw, err := json.Marshal(input)
	if err != nil {
		return bad
	}
	var public map[string]any
	if json.Unmarshal(raw, &public) != nil {
		return bad
	}
	public["images"] = []any{}
	if !reflect.DeepEqual(public, bindMultimodalFixture(spec.Input, values)) || !sameMultimodalJSON(capture.Body, bindMultimodalFixture(spec.Request, values)) {
		return bad
	}
	var compact bytes.Buffer
	if json.Compact(&compact, capture.Body) != nil {
		return bad
	}
	response := "response-" + binding + "-" + pioracle.SHA256(compact.Bytes()) + ".json"
	var actualResponse any
	if readLoadJSON(files, response, &actualResponse) != nil || !reflect.DeepEqual(actualResponse, bindMultimodalFixture(spec.Response, values)) {
		return bad
	}
	raw, err = json.Marshal(answers)
	if err != nil || !sameMultimodalJSON(raw, spec.Answers) {
		return bad
	}
	raw, err = json.Marshal(usage)
	if err != nil || !sameMultimodalJSON(raw, spec.Usage) {
		return bad
	}
	return nil
}

func sameMultimodalJSON(raw []byte, want any) bool {
	var got any
	return json.Unmarshal(raw, &got) == nil && reflect.DeepEqual(got, want)
}
func bindMultimodalFixture(value any, values map[string]any) any {
	switch v := value.(type) {
	case string:
		if replacement, ok := values[v]; ok {
			return replacement
		}
		return v
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = bindMultimodalFixture(item, values)
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for key, item := range v {
			if replacement, ok := values[key].(string); ok {
				key = replacement
			}
			out[key] = bindMultimodalFixture(item, values)
		}
		return out
	default:
		return value
	}
}
