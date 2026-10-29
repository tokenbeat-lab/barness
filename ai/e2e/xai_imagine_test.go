package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

const imagineModel = "grok-imagine-image-2.0"
const imaginePNG = "iVBORw0KGgo="

type imagineTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (r *imagineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || req.URL.Host != "api.x.ai" || (req.URL.Path != "/v1/images/generations" && req.URL.Path != "/v1/images/edits") || req.URL.RawQuery != "" {
		return nil, fmt.Errorf("unexpected Imagine target")
	}
	copy := req.Clone(req.Context())
	u := *req.URL
	u.Scheme = r.target.Scheme
	u.Host = r.target.Host
	copy.URL = &u
	return r.next.RoundTrip(copy)
}
func imagineBinding() ai.Binding {
	b := xaiBinding()
	b.BindingID = "imagine"
	b.Operation = ai.OperationImage
	b.API = ai.API("xai-images")
	b.AllowedModels = []string{imagineModel}
	return b
}
func imagineTarget() ai.Target { return ai.Target{BindingID: "imagine", ModelID: imagineModel} }
func imagineCatalog(t *testing.T) ai.Catalog {
	t.Helper()
	var cat ai.Catalog
	mustUnmarshal(t, []byte(`{"version":"imagine-fixture-2026-10-29","imageModels":[{"provider":"xai","api":"xai-images","id":"grok-imagine-image-2.0","name":"Imagine fixture","input":["text","image"],"output":["image"],"capabilities":{"maxReferenceImages":5,"maxOutputImages":10,"imageSizes":["1k","1.5k","2k"],"qualities":["low","medium","auto"],"aspectRatios":["1:1","16:9","auto"]},"pricing":{"source":"synthetic independent spec rates","perImage":{"inputImage":0.01,"output":[{"quality":"low","imageSize":"1k","price":0.04},{"quality":"low","imageSize":"1.5k","price":0.05},{"quality":"low","imageSize":"2k","price":0.06},{"quality":"medium","imageSize":"1k","price":0.06},{"quality":"medium","imageSize":"1.5k","price":0.07},{"quality":"medium","imageSize":"2k","price":0.08}]}}}]}`), &cat)
	return cat
}
func newImagineWorld(t *testing.T, configure ...func(*ai.Config)) *world {
	t.Helper()
	cat := imagineCatalog(t)
	transport := &imagineTransport{next: provider.LoopbackTransport()}
	w := newWorldWith(t, func(c *ai.Config) {
		configureImages(c)
		c.Catalog = &cat
		c.Transport = transport
		for _, f := range configure {
			f(c)
		}
	}, tenantA)
	transport.target, _ = url.Parse(w.provider.URL())
	b := imagineBinding()
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(xaiKey.secret)})
	return w
}
func imagineReply(body string) provider.Reply {
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "X-Request-Id": "imagine-request"}, Script: body, Chunks: [][]byte{[]byte(body)}}
}
func TestXAIImagineGeneration(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-generation-defaults")
	w := newImagineWorld(t)
	reply := imagineReply(`{"data":[{"b64_json":"` + imaginePNG + `"}]}`)
	enqueue(ev, w, reply)
	res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-defaults"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw a circle."}, nil)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	ev.Record("error", errString(err))
	ev.Check("sync inline success", err == nil && len(res.Content) == 1, "result=%+v err=%v", res, err)
	requests := w.provider.Requests()
	if len(requests) != 1 {
		ev.Check("one request", false, "got %d", len(requests))
		return
	}
	var body map[string]any
	mustUnmarshal(t, requests[0].Body, &body)
	ev.Check("fixed native request", requests[0].Path == "/v1/images/generations" && jsonEqual(requests[0].Body, []byte(`{"model":"grok-imagine-image-2.0","prompt":"Draw a circle.","response_format":"b64_json"}`)), "body=%s", requests[0].Body)
	ev.Check("MIME inferred", res.Content[0].(ai.ImageOutputImage).MimeType == "image/png", "content=%+v", res.Content)
	ev.Check("missing token remains unreported", res.Usage.Input == 0 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "usage=%+v", res.Usage)
}
