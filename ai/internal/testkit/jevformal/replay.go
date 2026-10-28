package jevformal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

type replayHost struct {
	binding    ai.Binding
	credential ai.Credential
}

func (h *replayHost) ResolveBinding(_ context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
	if scope.TenantID != h.binding.TenantID || id != h.binding.BindingID {
		return ai.Binding{}, errors.New("unapproved replay binding")
	}
	return h.binding, nil
}
func (h *replayHost) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	if scope.TenantID != h.binding.TenantID || b.CredentialRef != h.binding.CredentialRef {
		return ai.Credential{}, errors.New("unapproved replay credential")
	}
	return h.credential, nil
}

type capturedExchange struct {
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	RequestBody     string            `json:"requestBody"`
	Status          int               `json:"status"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ResponseBody    string            `json:"responseBody"`
}

type replayTransport struct{ exchange capturedExchange }

func (t replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	raw, err := io.ReadAll(io.LimitReader(req.Body, (4<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 4<<20 || req.Method != t.exchange.Method || req.URL.String() != t.exchange.URL || !sameJSON(raw, []byte(t.exchange.RequestBody)) {
		return nil, errors.New("final formal replay request differs")
	}
	headers := http.Header{}
	for name, value := range t.exchange.ResponseHeaders {
		headers.Set(name, value)
	}
	return &http.Response{StatusCode: t.exchange.Status, Header: headers, Body: io.NopCloser(strings.NewReader(t.exchange.ResponseBody)), Request: req}, nil
}

func sameJSON(a, b []byte) bool {
	decode := func(raw []byte) (any, error) {
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var out any
		err := d.Decode(&out)
		return out, err
	}
	x, err := decode(a)
	if err != nil {
		return false
	}
	y, err := decode(b)
	if err != nil {
		return false
	}
	ax, err := json.Marshal(x)
	if err != nil {
		return false
	}
	by, err := json.Marshal(y)
	return err == nil && bytes.Equal(ax, by)
}

// ReplayProtocol revalidates captures through the public runtime. Saved PASS
// assertions and matching response/result corruption cannot replace decoding.
// The transport returns captured bytes only; no credential or network is used.
func ReplayProtocol(root string, p Protocol) error {
	raw, err := Read(filepath.Join(root, "MM11-configuration/host-configuration.json"))
	if err != nil {
		return err
	}
	var config struct {
		Binding ai.Binding         `json:"binding"`
		Policy  *ai.ResourcePolicy `json:"policy"`
		Target  ai.Target          `json:"target"`
	}
	if json.Unmarshal(raw, &config) != nil || config.Binding.Endpoint != "https://api.typesafe.ai/v1" || config.Target.ModelID != "jev-1.14.0" {
		return errors.New("invalid formal replay configuration")
	}
	cat, err := evidence.MeasuredCatalog(root, p.CatalogVersion, p.CatalogHash, ai.BuiltinCatalog())
	if err != nil {
		return err
	}
	for _, id := range p.Expected {
		dir := filepath.Join(root, "MM11-"+id)
		raw, err = Read(filepath.Join(dir, "public-request.json"))
		if err != nil {
			return err
		}
		var input ai.ClassifierRequest
		if json.Unmarshal(raw, &input) != nil {
			return errors.New("invalid replay public input")
		}
		raw, err = Read(filepath.Join(dir, "exchanges.json"))
		if err != nil {
			return err
		}
		var exchanges []capturedExchange
		if json.Unmarshal(raw, &exchanges) != nil || len(exchanges) != 1 {
			return errors.New("formal replay needs one captured attempt")
		}
		raw, err = Read(filepath.Join(dir, "result.json"))
		if err != nil {
			return err
		}
		var saved struct {
			Result map[string]json.RawMessage `json:"result"`
		}
		if json.Unmarshal(raw, &saved) != nil {
			return errors.New("invalid replay result")
		}
		var metadata ai.CallMetadata
		if json.Unmarshal(saved.Result["metadata"], &metadata) != nil {
			return errors.New("invalid replay metadata")
		}
		h := &replayHost{binding: config.Binding, credential: ai.Credential{OwnerTenantID: config.Binding.TenantID, CredentialID: config.Binding.CredentialRef, Version: metadata.CredentialVersion, BindingVersion: config.Binding.Version, AccountScopeID: config.Binding.AccountScopeID, Active: true, APIKey: ai.NewSecret("synthetic-protocol-replay")}}
		client, err := ai.NewClient(ai.Config{Catalog: &cat, Policy: config.Policy, Bindings: h, Credentials: h, Transport: replayTransport{exchanges[0]}})
		if err != nil {
			return errors.New("formal replay Client rejected configuration")
		}
		var body map[string]any
		decoder := json.NewDecoder(strings.NewReader(exchanges[0].RequestBody))
		decoder.UseNumber()
		if decoder.Decode(&body) != nil {
			return errors.New("invalid final replay payload")
		}
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
			payload.Body = body
			return ai.KeepPayload(), nil
		}}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		res, callErr := client.WithHooks(hooks).Classify(ctx, ai.CallScope{TenantID: metadata.TenantID, RequestID: metadata.RequestID, ActorID: metadata.ActorID, JobID: metadata.JobID}, config.Target, input, ai.TypeSafeOptions{ExperimentalImages: true})
		cancel()
		if callErr != nil {
			return errors.New("captured formal protocol response does not replay through Classify")
		}
		raw, err = json.Marshal(res)
		if err != nil {
			return err
		}
		var got map[string]json.RawMessage
		if json.Unmarshal(raw, &got) != nil {
			return errors.New("invalid public replay result")
		}
		for _, name := range []string{"answers", "imageInputs", "responseModel", "usage", "stopReason", "errorMessage"} {
			if !sameJSONOrAbsent(got[name], saved.Result[name]) {
				return errors.New("formal capture contradicts public replay result")
			}
		}
	}
	return nil
}

func sameJSONOrAbsent(a, b []byte) bool {
	if len(a) == 0 || len(b) == 0 {
		return len(a) == len(b)
	}
	return sameJSON(a, b)
}
