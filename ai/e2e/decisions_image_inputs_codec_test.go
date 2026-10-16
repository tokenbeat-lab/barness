package e2e

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImageInputsCodec(t *testing.T) {
	for _, sc := range []struct {
		name, raw string
		valid     bool
	}{
		{"png", `{"mimeType":"image/png","sha256":"HASH"}`, true},
		{"jpeg", `{"sha256":"HASH","mimeType":"image/jpeg"}`, true},
		{"webp", `{"mimeType":"image/webp","sha256":"HASH"}`, true},
		{"missing-mime", `{"sha256":"HASH"}`, false},
		{"missing-hash", `{"mimeType":"image/png"}`, false},
		{"null", `null`, false},
		{"null-mime", `{"mimeType":null,"sha256":"HASH"}`, false},
		{"null-hash", `{"mimeType":"image/png","sha256":null}`, false},
		{"alias", `{"MimeType":"image/png","sha256":"HASH"}`, false},
		{"alias-hash", `{"mimeType":"image/png","SHA256":"HASH"}`, false},
		{"duplicate", `{"mimeType":"image/png","sha256":"HASH","sha256":"HASH"}`, false},
		{"escaped-duplicate", `{"mimeType":"image/png","sha256":"HASH","sha\u003256":"HASH"}`, false},
		{"unknown", `{"mimeType":"image/png","sha256":"HASH","data":"private"}`, false},
		{"native-key", `{"mimeType":"image/png","sha256":"HASH","file_id":"private"}`, false},
		{"wrong-type", `{"mimeType":"image/png","sha256":123}`, false},
		{"mime-case", `{"mimeType":"IMAGE/PNG","sha256":"HASH"}`, false},
		{"mime-parameters", `{"mimeType":"image/png;charset=utf-8","sha256":"HASH"}`, false},
		{"mime-alias", `{"mimeType":"image/jpg","sha256":"HASH"}`, false},
		{"unsupported", `{"mimeType":"image/gif","sha256":"HASH"}`, false},
		{"uppercase", `{"mimeType":"image/png","sha256":"UPPER"}`, false},
		{"short", `{"mimeType":"image/png","sha256":"1234"}`, false},
		{"long", `{"mimeType":"image/png","sha256":"HASH0"}`, false},
		{"nonhex", `{"mimeType":"image/png","sha256":"NONHEX"}`, false},
		{"trailing", `{"mimeType":"image/png","sha256":"HASH"} {}`, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC03-codec-"+sc.name)
			u := newReferenceWorld(t, ev, nil)
			req := decisionsInput()
			req.Images = classifierAssets(t)[:1]
			ev.Fixture("public-request.json", mustMarshal(t, req))
			recordClassifierAssets(ev, classifierAssets(t))
			enqueueClassifierImages(t, ev, u)
			res, err := u.client.Classify(ctxFor(t), textScope("req-image-codec-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			if !ev.Check("public Classify provides one image identity", err == nil && len(res.ImageInputs) == 1, "got %v", err) {
				return
			}
			original := res.ImageInputs[0]
			raw := strings.NewReplacer("UPPER", strings.ToUpper(original.SHA256), "NONHEX", strings.Repeat("g", 64), "HASH", original.SHA256).Replace(sc.raw)
			ev.Fixture("metadata-input.txt", []byte(raw))
			decoded := original
			decodeErr := json.Unmarshal([]byte(raw), &decoded)
			ev.Check("closed metadata decoding", (decodeErr == nil) == sc.valid, "unexpected acceptance")
			if sc.valid {
				var round ai.ClassifierImageInput
				mustUnmarshal(t, mustMarshal(t, decoded), &round)
				ev.Check("metadata round trip", reflect.DeepEqual(decoded, round), "changed identity")
			} else {
				ev.Check("failed decode leaves previous identity untouched", decoded == original, "partially mutated metadata")
			}
			for _, invalid := range []ai.ClassifierImageInput{
				{}, {MimeType: "image/gif", SHA256: original.SHA256},
				{MimeType: "image/png", SHA256: strings.ToUpper(original.SHA256)},
				{MimeType: "image/png", SHA256: strings.Repeat("g", 64)},
				{MimeType: "image/png", SHA256: original.SHA256 + "0"},
			} {
				_, encodeErr := json.Marshal(invalid)
				ev.Check("encoding rejects invalid metadata", encodeErr != nil, "invalid record encoded")
			}
			recordImageInputsCall(ev, u, res, err)
		})
	}
}
