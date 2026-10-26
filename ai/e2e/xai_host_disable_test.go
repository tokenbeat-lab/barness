package e2e

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIHostDisableAndSnapshots(t *testing.T) {
	ev := run.Case(t, "XAI-E18-disable-preserves-pinned-history-and-price")
	var callerCatalog *ai.Catalog
	w := newXAIRuntimeWorld(t, func(c *ai.Catalog) { callerCatalog = c })
	reply := xaiReply(t, xaiTerminal(xaiReasoning+`,`+xaiMessage, xaiUsage))
	enqueue(ev, w.world, reply)
	prior, err := w.client.Complete(ctxFor(t), textScope("xai-prior"), xaiTarget(), xaiRequest(), nil)
	ev.Record("prior-result", prior)
	_, hasNative := prior.Message.NativeState.Envelope()
	if !ev.Check("prior history has native state and priced consumption", err == nil && hasNative && prior.Message.Usage.Cost.Total == 0.000118, "%v %+v", err, prior) {
		return
	}
	saved := mustMarshal(t, prior)
	arrived, resume := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(resume) }) }
	defer unblock()
	active := reply
	active.OnReceive = func() { close(arrived); <-resume }
	enqueue(ev, w.world, active)
	done := make(chan outcome, 1)
	go func() {
		res, e := w.client.Complete(ctxFor(t), textScope("xai-active"), xaiTarget(), xaiRequest(), nil)
		done <- outcome{result: res, err: e}
	}()
	waitFor(ev, "inflight attempt already sent", arrived)
	b := xaiBinding()
	b.Enabled, b.Version = false, "disabled-v2"
	w.host.PutBinding(b)
	callerCatalog.Version = "caller-mutated-version"
	callerCatalog.Models[0].Cost.Input = 100
	before := len(w.provider.Requests())
	denied, deniedErr := w.client.Complete(ctxFor(t), textScope("xai-denied"), xaiTarget(), xaiRequest(), nil)
	ev.Record("disabled-binding", b)
	ev.Record("denied-result", denied)
	ev.Record("denied-error", errString(deniedErr))
	ev.Record("denied-reads", w.host.ReadsFor("xai-denied"))
	ev.Check("disabled binding rejects before credential and HTTP with no fallback", errors.Is(deniedErr, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseBinding}) && w.host.ReadsFor("xai-denied").Credentials == 0 && len(denied.Metadata.Attempts) == 0 && len(w.provider.Requests()) == before, "%v %+v", deniedErr, denied.Metadata)
	unblock()
	o := waitValue(ev, "pinned call completes after disable", done)
	o.record(ev)
	ev.Check("inflight call retains original binding and price snapshot", o.err == nil && o.result.Metadata.BindingVersion == "v1" && o.result.Metadata.CatalogVersion == prior.Metadata.CatalogVersion && o.result.Metadata.CatalogHash == prior.Metadata.CatalogHash && o.result.Message.Usage.Cost == prior.Message.Usage.Cost && len(o.result.Metadata.Attempts) == 1, "%v %+v", o.err, o.result.Metadata)
	ev.Check("saved history and original cost remain byte-identical", bytes.Equal(saved, mustMarshal(t, prior)), "%+v", prior)
	b.Enabled, b.Version = true, "restored-v3"
	w.host.PutBinding(b)
	enqueue(ev, w.world, reply)
	req := xaiRequest()
	req.Messages = append(req.Messages, prior.Message, ai.UserText("Continue saved history."))
	restored, restoreErr := w.client.Complete(ctxFor(t), textScope("xai-restored"), xaiTarget(), req, nil)
	ev.Record("restored-result", restored)
	ev.Record("requests", w.provider.Requests())
	ev.Check("explicit restore resumes saved xAI history on same frozen catalog", restoreErr == nil && restored.Metadata.BindingVersion == "restored-v3" && restored.Metadata.CatalogHash == prior.Metadata.CatalogHash && restored.Message.Usage.Cost == prior.Message.Usage.Cost, "%v %+v", restoreErr, restored.Metadata)
	reqs := w.provider.Requests()
	if len(reqs) == 3 {
		ev.Check("saved ciphertext reaches only bound xAI target", bytes.Contains(reqs[2].Body, []byte("synthetic-cipher")) && reqs[2].KeyAlias == xaiKey.alias && reqs[2].Path == "/v1/responses", "%s", reqs[2].Body)
	}
	w.released(ev)
}
