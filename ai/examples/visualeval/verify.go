package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

func verify(dir string, requireLive bool) error {
	read := func(name string) ([]byte, error) {
		if !fs.ValidPath(name) {
			return nil, errors.New("invalid artifact path")
		}
		return readBounded(filepath.Join(dir, name), 16<<20)
	}
	var m manifest
	raw, err := read("manifest.json")
	if err != nil || strict(raw, &m) != nil || m.Schema != 1 {
		return errors.New("invalid manifest")
	}
	listed := map[string]bool{"manifest.json": true}
	for name, hash := range m.Artifacts {
		raw, err := read(name)
		if err != nil || digest(raw) != hash {
			return fmt.Errorf("inventory mismatch: %s", name)
		}
		listed[name] = true
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink in evidence")
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if !listed[filepath.ToSlash(rel)] {
			return errors.New("unlisted evidence artifact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Reading must reject symlinked parents as well as leaves.
	if info, err := os.Lstat(dir); err != nil || !info.IsDir() {
		return errors.New("bundle root must be a regular directory")
	}
	get := func(name string, v any) error {
		raw, e := read(name)
		if e != nil {
			return e
		}
		return strict(raw, v)
	}
	findings, err := audit.Bundle(dir, audit.Environment())
	if err != nil || len(findings) > 0 {
		return errors.New("independent redaction audit failed")
	}
	var audited struct {
		Findings []audit.Finding `json:"findings"`
	}
	if get("audit.json", &audited) != nil || audited.Findings == nil || len(audited.Findings) > 0 {
		return errors.New("missing or failed audit receipt")
	}
	var r report
	if err = get("report.json", &r); err != nil {
		return err
	}
	p, err := load(r.Route)
	if err != nil {
		return err
	}
	for name, hash := range p.pins {
		raw, e := read("inputs/" + name)
		if e != nil || digest(raw) != hash {
			return errors.New("fixed input drift")
		}
	}
	pinBytes, _ := data.ReadFile("testdata/pins.json")
	raw, err = read("inputs/pins.json")
	if err != nil || !bytes.Equal(pinBytes, raw) {
		return errors.New("input pin drift")
	}
	var config struct {
		Mode          string             `json:"mode"`
		Route         string             `json:"route"`
		Entry         string             `json:"entry"`
		MaxCalls      int                `json:"maxCalls"`
		WorkMS        int                `json:"workMS"`
		CallMS        int                `json:"callMS"`
		Policy        *ai.ResourcePolicy `json:"policy"`
		Authorization *authorization     `json:"authorization"`
	}
	if get("run-config.json", &config) != nil || config.Mode != r.Mode || config.Route != r.Route || config.Entry != r.Entry || config.MaxCalls < 1 || config.MaxCalls > 12 || config.WorkMS < 1 || config.WorkMS > 225000 || config.CallMS < 1 || config.CallMS > 30000 || !same(config.Policy, policy(time.Duration(config.CallMS)*time.Millisecond)) {
		return errors.New("invalid run configuration")
	}
	if (r.Mode != "controlled" && r.Mode != "live") || (r.Entry != "plain" && r.Entry != "hooked") {
		return errors.New("invalid evaluation source")
	}
	if r.Mode == "controlled" && config.Authorization != nil {
		return errors.New("controlled run cannot carry live authorization")
	}
	if r.Mode == "live" && r.Budget.Calls > 0 && (validAuthorization(p, config.Authorization) != nil || config.MaxCalls != 12 || config.WorkMS != 225000 || config.CallMS != 30000) {
		return errors.New("live result lacks fixed route authorization")
	}
	expected := initial(p, arguments{mode: r.Mode, entry: r.Entry})
	if r.Schema != expected.Schema || r.CatalogHash != expected.CatalogHash || r.CatalogVersion != expected.CatalogVersion || r.Model != expected.Model || r.DatasetHash != expected.DatasetHash || r.ConfigHash != expected.ConfigHash || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.FinishedAt.Sub(r.StartedAt) > 240*time.Second {
		return errors.New("identity/time drift")
	}
	raw, err = read("protocol-evidence.json")
	if err != nil || digest(raw) != r.ProtocolEvidenceHash {
		return errors.New("protocol receipt drift")
	}
	var protocol struct {
		Live []struct{ Route, Outcome, CatalogHash string }
	}
	if json.Unmarshal(raw, &protocol) != nil {
		return errors.New("invalid independent protocol receipt")
	}
	linked := false
	for _, v := range protocol.Live {
		if v.Route == p.route && v.CatalogHash == r.CatalogHash {
			linked = true
		}
	}
	if !linked {
		return errors.New("protocol evidence does not identify this route and candidate hash")
	}
	var hashes map[string]string
	if get("source-hashes.json", &hashes) != nil || digest(encoded(hashes)) != r.CodeHash {
		return errors.New("invalid source receipt")
	}
	current, err := sourceFiles()
	if err != nil || len(current) != len(hashes) {
		return errors.New("source set drift")
	}
	for name, hash := range hashes {
		raw, e := read("source/" + name)
		currentRaw, exists := current[name]
		if e != nil || !exists || digest(raw) != hash || digest(currentRaw) != hash {
			return errors.New("source snapshot drift; verify with recorded revision")
		}
	}
	if len(r.Samples) != 12 {
		return errors.New("incomplete sample inventory")
	}
	rebuilt := r
	rebuilt.Budget = budget{}
	ids := map[string]bool{}
	for i, s := range r.Samples {
		if s.ID != p.dataset.Samples[i].ID || s.Group != p.group(i) {
			return errors.New("sample order or group drift")
		}
		if s.Result.Metadata.RequestID == "" {
			for name := range m.Artifacts {
				if strings.HasPrefix(name, "samples/"+s.ID+"/") {
					return errors.New("unexecuted sample hides captured call")
				}
			}
			if s.Status != "not_run" && s.Status != "budget_exhausted" && s.Status != "cancelled" {
				return errors.New("invalid unsent terminal")
			}
			if len(s.View) != 0 || len(s.Result.Answers) != 0 || len(s.Result.Metadata.Attempts) != 0 || s.Error != nil || s.IoU != nil || s.Consumed {
				return errors.New("invented unexecuted prediction")
			}
			continue
		}
		prefix := "samples/" + s.ID + "/"
		var recorded sampleResult
		if get(prefix+"result.json", &recorded) != nil || !same(recorded, s) {
			return errors.New("sample result drift")
		}
		raw, e := read(prefix + "input.json")
		if e != nil || !bytes.Equal(raw, encoded(p.request(i))) {
			return errors.New("public input drift")
		}
		var w wire
		if get(prefix+"wire.json", &w) != nil {
			return errors.New("missing wire capture")
		}
		raw, e = read(prefix + "request.json")
		if e != nil || !bytes.Equal(raw, w.Request) {
			return errors.New("final request drift")
		}
		raw, e = read(prefix + "response.bin")
		if e != nil || !bytes.Equal(raw, w.Response) {
			return errors.New("final response drift")
		}
		md := s.Result.Metadata
		alias := "synthetic-account"
		if r.Mode == "live" && config.Authorization != nil {
			alias = config.Authorization.AccountAlias
		}
		if md.TenantID != "visual-evaluation-host" || md.BindingID != p.route || md.BindingVersion != "visual-evaluation-v1" || md.CredentialVersion != "visual-evaluation-v1" || md.AccountScopeID != alias {
			return errors.New("pinned Host attribution drift")
		}

		if ids[md.RequestID] || len(md.Attempts) != 1 || md.Attempts[0].AttemptID != md.RequestID+"#1" || md.Attempts[0].RetryDelay != 0 || md.Attempts[0].HTTPStatus != w.Status {
			return errors.New("attempt identity or accounting drift")
		}
		ids[md.RequestID] = true
		var obs []ai.Observation
		if get(prefix+"observations.json", &obs) != nil || len(obs) != 4 || obs[0].Kind != ai.ObservationCallStarted || obs[1].Kind != ai.ObservationAttemptStarted || obs[2].Kind != ai.ObservationAttemptFinished || obs[3].Kind != ai.ObservationCallFinished || !same(obs[3].Call, md) || obs[1].Attempt == nil || obs[2].Attempt == nil || obs[1].Attempt.AttemptID != md.Attempts[0].AttemptID || !same(*obs[2].Attempt, md.Attempts[0]) {
			return errors.New("Observer accounting drift")
		}
		terminal := obs[3]
		if !same(terminal.Usage, s.Result.Usage) || (s.Error == nil) != (terminal.Error == nil) {
			return errors.New("Observer terminal usage/error drift")
		}
		if s.Error != nil && (terminal.Error.Code != s.Error.Code || terminal.Error.Phase != s.Error.Phase) {
			return errors.New("Observer terminal error differs from public failure")
		}
		for _, o := range obs {
			if o.Call.RequestID != md.RequestID || o.Call.TenantID != md.TenantID {
				return errors.New("Observer attribution drift")
			}
		}
		if len(s.Assertions) != 7 || len(s.Resources) != 4 {
			return errors.New("incomplete assertions/resources")
		}
		for _, name := range []string{"observerTerminal", "observerDelivered", "credentialSnapshot", "released-bodies", "released-calls", "released-permits", "released-waiters"} {
			if !s.Assertions[name] {
				return errors.New("required assertion missing or failed")
			}
		}
		for _, ok := range s.Assertions {
			if !ok {
				return errors.New("failed assertion")
			}
		}
		for _, n := range s.Resources {
			if n != 0 {
				return errors.New("unreleased resource")
			}
		}
		if s.StartedAt.Before(r.StartedAt) || s.FinishedAt.Before(s.StartedAt) || s.FinishedAt.After(r.FinishedAt) || s.FinishedAt.Sub(s.StartedAt) > 31*time.Second {
			return errors.New("sample time drift")
		}
		// Replay consumes only the captured response through the public Classify
		// runtime, using the original public image/task. There is no network access.
		if err := replaySample(p, i, r.Entry, w, s); err != nil {
			return err
		}
		copy := s
		assess(p, i, &copy)
		if copy.Status != s.Status || !same(copy.View, s.View) || !same(copy.IoU, s.IoU) || copy.Consumed != s.Consumed {
			return errors.New("prediction/scoring drift")
		}
		rebuilt.Budget.add(s, p.request(i))
	}
	if rebuilt.Budget.Calls > config.MaxCalls || rebuilt.Budget.Attempts > 12 || rebuilt.Budget.Retries != 0 || rebuilt.Budget.Questions > 48 || rebuilt.Budget.Occurrences > 24 || rebuilt.Budget.SourceBytes > 8<<20 || rebuilt.Budget.ExpandedBytes > 16<<20 || !same(rebuilt.Budget, r.Budget) {
		return errors.New("cumulative accounting drift")
	}
	finalize(p, &rebuilt)
	if rebuilt.Status != r.Status || rebuilt.RealQuality != r.RealQuality || !same(rebuilt.Metrics, r.Metrics) {
		return errors.New("terminal or metrics drift")
	}
	if r.Status == "NOT_RUN" && (r.Reason == "" || r.Budget.Calls != 0) {
		return errors.New("invalid NOT_RUN")
	}
	if requireLive && (r.Mode != "live" || r.Status != "COMPLETE" || r.RealQuality != "COMPLETE") {
		return errors.New("complete real visual evaluation has not run")
	}
	return nil
}

type replayTransport struct{ w wire }

func (t replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Body.Close()
	if t.w.Status == 0 {
		return nil, errors.New("recorded transport failure")
	}
	h := http.Header{}
	for k, v := range t.w.ResponseHeaders {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: t.w.Status, Header: h, Body: io.NopCloser(bytes.NewReader(t.w.Response)), Request: req}, nil
}
func canonicalRequest(raw []byte) any {
	var v map[string]any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	// Exactly one source exists in this fixed experiment. Only its random Jev
	// private key is normalized; its bytes/format/scope and every task stay exact.
	if assets, ok := v["assets"].(map[string]any); ok && len(assets) == 1 {
		for key := range assets {
			raw = bytes.ReplaceAll(encoded(v), []byte(strings.Trim(string(encoded(key)), "\n\"")), []byte("FIXED_IMAGE"))
			break
		}
		if json.Unmarshal(raw, &v) != nil {
			return nil
		}
	}
	return v
}
func replaySample(p *plan, i int, entry string, w wire, s sampleResult) error {
	a := arguments{mode: "controlled", entry: entry, callMS: 30000}
	got, generated, _, err := call(context.Background(), p, a, p.request(i), i, "http://127.0.0.1", "synthetic-replay-credential", "synthetic-replay-account", replayTransport{w})
	if err != nil {
		return err
	}
	if w.Method != "POST" || w.Path != generated.Path || !same(canonicalRequest(w.Request), canonicalRequest(generated.Request)) || !same(w.Headers, generated.Headers) {
		return errors.New("final native input differs from fixed public task")
	}
	if s.Error != nil && (s.Error.Code == ai.CodeCanceled || s.Error.Code == ai.CodeDeadlineExceeded) {
		if len(s.Result.Answers) != 0 || len(s.Result.ImageInputs) != 0 || (s.Error.Phase != ai.PhaseRequest && s.Error.Phase != ai.PhaseResponse) {
			return errors.New("invalid cancelled outcome")
		}
		ga, sa := got.Result.Metadata.Attempts[0], s.Result.Metadata.Attempts[0]
		// Cancellation may occur before decoding even a fully consumed response.
		// In that case nothing was reported, and every usage/cost field stays zero.
		if sa.UsageReporting == ai.UsageUnreported && same(sa.Usage, ai.Usage{}) && same(s.Result.Usage, ai.Usage{}) && s.Result.ResponseModel.IsZero() {
			return nil
		}
		// A cancellation after decoding can retain consumption, but only what the
		// captured response independently produces at this exact price snapshot.
		if !same(ga.Usage, sa.Usage) || ga.UsageReporting != sa.UsageReporting || !same(got.Result.Usage, s.Result.Usage) || !same(got.Result.ResponseModel, s.Result.ResponseModel) {
			return errors.New("cancelled usage/model unsupported by captured response")
		}
		return nil
	}
	if !same(got.Result.Answers, s.Result.Answers) || !same(got.Result.ImageInputs, s.Result.ImageInputs) || !same(got.Result.Usage, s.Result.Usage) || !same(got.Result.ResponseModel, s.Result.ResponseModel) || !same(got.Error, s.Error) {
		return errors.New("captured response differs from public result")
	}
	ga, sa := got.Result.Metadata.Attempts[0], s.Result.Metadata.Attempts[0]
	if !same(ga.Usage, sa.Usage) || ga.UsageReporting != sa.UsageReporting {
		return errors.New("usage/price snapshot drift")
	}
	return nil
}
