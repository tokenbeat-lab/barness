//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestFinalRoundCNBudgetProcess(t *testing.T) {
	for _, route := range []string{"chat", "responses", "messages", "wrong-chat"} {
		t.Run(route, func(t *testing.T) {
			cs := run.Case(t, "REGIONAL-final-round-cn-usd-"+route)
			base := t.TempDir()
			budget := filepath.Join(base, "budget.json")
			raw := []byte(`{"schema":1,"round":"provider-sites-final-acceptance-1","maxMicrousd":1000000,"reservedMicrousd":0,"attempts":0}`)
			if err := os.WriteFile(budget, raw, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestFinalRoundCNBudgetFixture$", "-test.timeout=20s")
			for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
				if value := os.Getenv(name); value != "" {
					cmd.Env = append(cmd.Env, name+"="+value)
				}
			}
			cmd.Env = append(cmd.Env, "BARNESS_CN_BUDGET_FIXTURE="+route, "BARNESS_FINAL_ROUND=provider-sites-final-acceptance-1", "BARNESS_FINAL_BUDGET="+budget, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(base, "native"))
			output, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				code = -1
				if failure, ok := err.(*exec.ExitError); ok {
					code = failure.ExitCode()
				}
			}
			cs.Record("public-host-process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(output)})
			cs.Check("own CN Host reaches guarded transport while wrong host cannot", err == nil, "%v %s", err, output)
			raw, err = os.ReadFile(budget)
			if err != nil {
				t.Fatal(err)
			}
			cs.Record("own-cn-usd-reservation", json.RawMessage(raw))
		})
	}
}

func TestFinalRoundCNBudgetFixture(t *testing.T) {
	route := os.Getenv("BARNESS_CN_BUDGET_FIXTURE")
	if route == "" {
		t.Skip("isolated synthetic CN Host process only")
	}
	wrong := route == "wrong-chat"
	protocol := strings.TrimPrefix(route, "wrong-")
	a := minimaxhost.Authorization{TenantID: liveTenant, AccountScopeID: "controlled-cn-standard", CredentialRef: "controlled-cn", Version: "controlled-v1", ProviderID: ai.ProviderMiniMaxCN, Purpose: minimaxhost.Standard}
	if wrong {
		// A trusted gateway override can reach the Client, but this round's
		// spending guard must still refuse a host outside the own CN service.
		a.TrustedGateway = true
		a.Endpoints = map[ai.API]string{ai.APIOpenAICompletions: "https://api.minimax.chat/v1"}
	}
	bindings, err := minimaxhost.Bindings(a)
	if err != nil {
		t.Fatal(err)
	}
	var binding ai.Binding
	for _, candidate := range bindings {
		if strings.HasSuffix(candidate.BindingID, "-"+protocol) {
			binding = candidate
		}
	}
	want, _ := supportmatrix.RegionalRouteOf("minimax-cn-" + protocol)
	if !wrong && binding.Endpoint != want.Endpoint {
		t.Fatalf("public Host and own regional endpoint differ: %s", binding.Endpoint)
	}
	cat, err := regional.QualificationCatalog(ai.BuiltinCatalog())
	if err != nil {
		t.Fatal(err)
	}
	res := &resolver{binding: binding, credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: a.CredentialRef, Version: a.Version, BindingVersion: a.Version, AccountScopeID: a.AccountScopeID, Active: true, APIKey: ai.NewSecret("synthetic-cn-budget-credential")}}
	rec := newRecorder()
	rec.next.(*http.Transport).CloseIdleConnections()
	calls := 0
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("controlled transport failure; no real network")
	})
	client, err := ai.NewClient(ai.Config{Policy: livePolicy(), Bindings: res, Credentials: res, Transport: rec, Observer: rec, Catalog: &cat})
	if err != nil {
		t.Fatal(err)
	}
	model, _ := cat.Lookup(ai.ProviderMiniMaxCN, binding.API, "MiniMax-M3")
	_, err = client.Complete(context.Background(), newScope(), ai.Target{BindingID: binding.BindingID, ModelID: model.ID}, ai.Request{Messages: []ai.Message{ai.UserText("Controlled own CN spending reservation.")}}, regionalOptions(model, 2048))
	if err == nil || wrong && calls != 0 || !wrong && calls != 1 {
		t.Fatalf("guarded public Host requests=%d wrong=%v error=%v", calls, wrong, err)
	}
	var value struct {
		Reserved int64 `json:"reservedMicrousd"`
		Attempts int   `json:"attempts"`
		Last     struct {
			Model      string  `json:"model"`
			InputRate  float64 `json:"inputRate"`
			OutputRate float64 `json:"outputRate"`
		} `json:"last"`
	}
	raw, err := os.ReadFile(os.Getenv("BARNESS_FINAL_BUDGET"))
	if err != nil || json.Unmarshal(raw, &value) != nil {
		t.Fatal("missing own CN reservation")
	}
	if wrong {
		if value.Reserved != 0 || value.Attempts != 0 {
			t.Fatalf("wrong host must be refused before reservation: %+v", value)
		}
	} else if value.Reserved <= 0 || value.Attempts != 1 || value.Last.Model != "MiniMax-M3" || value.Last.InputRate != 0.6248140434 || value.Last.OutputRate != 2.4992561738 {
		t.Fatalf("three protocols retain own CN highest-tier rates after failed send: %+v", value)
	}
}
