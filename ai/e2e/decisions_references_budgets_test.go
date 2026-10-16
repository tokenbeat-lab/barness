package e2e

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func setReferencePolicy(t *testing.T, c *ai.Config, count int, bytes int64) {
	t.Helper()
	raw := mustMarshal(t, map[string]any{"References": map[string]any{"MaxReferences": count, "MaxExpandedImageBytes": bytes}})
	if err := json.Unmarshal(raw, c.Policy.Classifier.Images); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionsReferencesBudgets(t *testing.T) {
	img := classifierAssets(t)[0]
	decoded, _ := base64.StdEncoding.DecodeString(img.Data)
	size := int64(len(decoded))
	for _, name := range []string{"inherit-count", "inherit-bytes", "explicit-count", "explicit-bytes", "explicit-at", "explicit-exceeds-sources", "model-count", "128", "129", "known-expansion"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC02-budget-"+name)
			count := 3
			success := false
			phase := ai.PhaseScope
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				switch name {
				case "inherit-count":
					c.Policy.Classifier.Images.MaxImages = 2
				case "inherit-bytes":
					c.Policy.Classifier.Images.MaxTotalImageBytes = size*3 - 1
				case "explicit-count":
					setReferencePolicy(t, c, 2, 1<<19)
				case "explicit-bytes":
					setReferencePolicy(t, c, 256, size*3-1)
				case "explicit-at":
					setReferencePolicy(t, c, 3, size*3)
					success = true
				case "explicit-exceeds-sources":
					c.Policy.Classifier.Images.MaxImages = 1
					c.Policy.Classifier.Images.MaxTotalImageBytes = size
					setReferencePolicy(t, c, 3, size*3)
					success = true
				case "model-count":
					c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.MaxInputImages = 2
					phase = ai.PhaseCapability
				case "128":
					count = 128
					success = true
				case "129":
					count = 129
					phase = ai.PhaseCapability
				case "known-expansion":
					c.Policy.MaxRequestBytes = 1000
					c.Policy.MaxImageBytes = 1000
					c.Policy.Classifier.Images.MaxTotalImageBytes = 1000
					count = 20
					setReferencePolicy(t, c, 20, size*20)
				}
			})
			req := decisionsInput()
			req.Images = []ai.Image{img}
			req.ImageReferences = make([]ai.ImageReference, count)
			for i := range req.ImageReferences {
				req.ImageReferences[i] = ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext}
			}
			ev.Fixture("public-request.json", mustMarshal(t, req))
			ev.Record("decoded-image-bytes", size)
			if success {
				enqueueClassifierImages(t, ev, u)
			}
			res, err := u.client.Classify(ctxFor(t), textScope("req-budget-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			if success {
				ev.Check("expanded request succeeds", err == nil && len(res.Metadata.Attempts) == 1, "got %v", err)
			} else {
				code := ai.CodeResourceLimit
				if phase == ai.PhaseCapability {
					code = ai.CodeInvalidRequest
				}
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				ev.Check("bounded before credentials and Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata)
			}
			ev.Record("result", res)
			ev.Record("error", errString(err))
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsReferencesBeforeCopy(t *testing.T) {
	for _, name := range []string{"count", "pointer"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC02-before-copy-"+name)
			u := newClassifierImagesWorld(t, nil)
			req := decisionsInput()
			req.Images = classifierAssets(t)[:1]
			if name == "count" {
				req.ImageReferences = make([]ai.ImageReference, 1<<20)
			} else {
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/" + strings.Repeat("x", 8<<20)}}
			}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.Classify(ctxFor(t), textScope("req-before-copy-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			runtime.ReadMemStats(&after)
			checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseScope, ai.StopReasonError)
			ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
			ev.Check("reject known reference sizes before proportional allocation", after.TotalAlloc-before.TotalAlloc < 1<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			ev.Check("zero credentials and Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsReferencesDocumentBeforeResolution(t *testing.T) {
	ev := run.Case(t, "UC02-document-before-resolution")
	u := newClassifierImagesWorld(t, nil)
	req := decisionsInput()
	req.Images = classifierAssets(t)[:1]
	req.State = json.RawMessage(`{"photo":null,"tail":"` + strings.Repeat("x", 900<<10) + `"}`)
	req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	res, err := u.client.Classify(ctxFor(t), textScope("req-document-before-resolution"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	runtime.ReadMemStats(&after)
	checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseScope, ai.StopReasonError)
	ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
	ev.Check("per-document bound checked before pointer traversal copies subtrees", after.TotalAlloc-before.TotalAlloc < 128<<10, "allocated %d", after.TotalAlloc-before.TotalAlloc)
	u.records(ev, res, err)
	u.released(ev)
}
