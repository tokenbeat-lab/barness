package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// The fixture's decoded file is the independent source of identity; neither
// the public request index nor the Base64/data URL is an image identity.
func expectedImageInputs(t *testing.T, images []ai.Image) []map[string]string {
	t.Helper()
	var out []map[string]string
	for _, img := range images {
		raw, err := base64.StdEncoding.Strict().DecodeString(img.Data)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		out = append(out, map[string]string{"mimeType": img.MimeType, "sha256": hex.EncodeToString(sum[:])})
	}
	return out
}

func checkImageInputs(ev *evidence.Case, res ai.ClassifierResult, images []ai.Image) {
	t := ev.T()
	var fields map[string]json.RawMessage
	mustUnmarshal(t, mustMarshal(t, res), &fields)
	if len(images) == 0 {
		ev.Check("text/failure omits imageInputs", fields["imageInputs"] == nil, "unexpected metadata")
		return
	}
	want := mustMarshal(t, expectedImageInputs(t, images))
	ev.Check("final ordered decoded file identities; closed records", jsonEqual(fields["imageInputs"], want), "metadata differs from final images")
}

func recordImageInputsCall(ev *evidence.Case, u *unaryWorld, res ai.ClassifierResult, err error) {
	ev.Record("result", res)
	ev.Record("error", errString(err))
	u.records(ev, res, err)
	obs := u.rec.all()
	ev.Record("observer-records", obs)
	wire := observationText(ev.T(), obs)
	diagnostics := string(mustMarshal(ev.T(), res.Metadata))
	identities := expectedImageInputs(ev.T(), classifierAssets(ev.T()))
	for _, img := range res.ImageInputs {
		identities = append(identities, map[string]string{"sha256": img.SHA256})
	}
	for _, img := range identities {
		for _, marker := range []string{img["sha256"], "imageInputs", "image_url", "base64,"} {
			ev.Check("shared observations and diagnostics exclude image content", !strings.Contains(wire, marker) && !strings.Contains(diagnostics, marker), "content leaked")
		}
	}
	u.released(ev)
}

func TestDecisionsImageInputsOwnership(t *testing.T) {
	ev := run.Case(t, "UC03-result-ownership")
	u := newReferenceWorld(t, ev, nil)
	req := decisionsInput()
	assets := classifierAssets(t)
	req.Images = []ai.Image{assets[0], assets[0], assets[1]}
	ev.Fixture("public-request.json", mustMarshal(t, req))
	recordClassifierAssets(ev, assets)
	enqueueClassifierImages(t, ev, u)
	first, err := u.client.Classify(ctxFor(t), textScope("req-image-owner-first"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Check("first call succeeds", err == nil && len(first.ImageInputs) == 3, "got %v", err)
	checkImageInputs(ev, first, req.Images)
	recordImageInputsCall(ev, u, first, err)
	if len(first.ImageInputs) != 3 {
		return
	}
	// Returned records are Host-owned. Changing one duplicate occurrence cannot
	// change another record, a later call, or the frozen evidence of this call.
	saved := first.ImageInputs[1]
	first.ImageInputs[0] = first.ImageInputs[2]
	ev.Check("duplicate occurrence records are independent values", first.ImageInputs[1] == saved, "duplicate alias")
	enqueueClassifierImages(t, ev, u)
	second, secondErr := u.client.Classify(ctxFor(t), textScope("req-image-owner-second"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Check("next call succeeds independently", secondErr == nil, "got %v", secondErr)
	checkImageInputs(ev, second, req.Images)
	second.ImageInputs[1] = second.ImageInputs[2]
	ev.Check("later returned slice cannot mutate earlier result", first.ImageInputs[1] == saved, "cross-call alias")
	ev.Record("second-result-after-host-edit", second)
	u.records(ev, second, secondErr)
	u.released(ev)
}

func TestDecisionsImageInputsOrder(t *testing.T) {
	for _, sc := range []struct {
		name  string
		order []int
	}{
		{"text", nil}, {"png", []int{0}}, {"jpeg", []int{1}}, {"webp", []int{2}},
		{"ordered", []int{0, 1, 2}}, {"references", []int{2, 0, 1, 2}}, {"same-file", []int{0, 0}},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC03-identity-"+sc.name)
			u := newReferenceWorld(t, ev, nil)
			req := decisionsInput()
			assets := classifierAssets(t)
			var ordered []ai.Image
			for _, i := range sc.order {
				ordered = append(ordered, assets[i])
			}
			req.Images = ordered
			if sc.name == "references" {
				req.Images = assets
				for _, i := range sc.order {
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: i, Placement: ai.ImagePlacementContext})
				}
			}
			recordClassifierAssets(ev, assets)
			ev.Fixture("public-request.json", mustMarshal(t, req))
			enqueueClassifierImages(t, ev, u)
			res, err := u.client.Classify(ctxFor(t), textScope("req-image-inputs-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			ev.Check("classification succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
			checkImageInputs(ev, res, ordered)
			requests := u.provider.Requests()
			if ev.Check("one actual Attempt and request", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "request/attempt mismatch") && len(ordered) > 0 {
				var body map[string]any
				mustUnmarshal(t, requests[0].Body, &body)
				ev.Check("metadata order is actual content order", reflect.DeepEqual(body["input"], classifierImageMessages(ordered)), "unexpected final input")
			}
			recordImageInputsCall(ev, u, res, err)
		})
	}
}
