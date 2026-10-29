package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type imagineProcessReport struct {
	Proxy              string          `json:"proxy"`
	Outcome            string          `json:"outcome"`
	Complete           bool            `json:"complete"`
	FormalQualified    bool            `json:"formalQualified"`
	AuditPassed        bool            `json:"auditPassed"`
	Scope              livegrant.Scope `json:"scope"`
	CallsUsed          int             `json:"callsUsed"`
	HTTPAttempts       int             `json:"httpAttempts"`
	VendorHTTPAttempts int             `json:"vendorHTTPAttempts"`
	InputImages        int             `json:"inputImages"`
	OutputImages       int             `json:"outputImages"`
	ReservedUSD        float64         `json:"reservedUSD"`
	ReportedUSD        float64         `json:"reportedUSD"`
	Scenarios          []struct{ ID, Outcome string }
}

// The public seam is the isolated binary, its real exit, and retained report.
// Fresh environments contain neither live permission nor vendor credentials.
func TestXAIImagesLiveProcess(t *testing.T) {
	privateGrantStore(t)
	cs := run.Case(t, "XAIIMG-process-build")
	binary := filepath.Join(t.TempDir(), "xai-images.test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	source, err := supportmatrix.XAIImagesExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	build := exec.CommandContext(ctx, "go", "test", "-c", "-tags=live", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource="+supportmatrix.RegionalSourceHash(source), "-o", binary, "./ai/live")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	raw, err := build.CombinedOutput()
	cs.Record("build", map[string]any{"argv": build.Args, "exitCode": imagineProcessExit(err), "output": string(raw)})
	if err != nil {
		t.Fatalf("compile isolated harness: %v %s", err, raw)
	}
	binaryHash, err := livegrant.FileHash(binary)
	if err != nil {
		t.Fatal(err)
	}
	cs.Record("binary", map[string]any{"sha256": binaryHash, "source": source})
	child := func(t *testing.T, name, entry string, env []string) (imagineProcessReport, string, int) {
		t.Helper()
		cs := run.Case(t, "XAIIMG-process-"+name)
		out, e := os.MkdirTemp(run.Dir(), "xai-images-child-")
		if e != nil {
			t.Fatal(e)
		}
		command := exec.CommandContext(ctx, binary, "-test.run="+entry, "-test.count=1", "-test.timeout=30s")
		command.Dir = "../.."
		command.Env = append(xaiProcessEnv(t), "BARNESS_AI_LIVE_COMBO=xai-images", evidence.DirEnv+"="+out)
		command.Env = append(command.Env, env...)
		raw, e := command.CombinedOutput()
		code := imagineProcessExit(e)
		cs.Record("process", map[string]any{"argv": command.Args, "exitCode": code, "output": string(raw)})
		files, _ := filepath.Glob(filepath.Join(out, "*", "xai-images-candidate-report.json"))
		var report imagineProcessReport
		if !cs.Check("own candidate report exists", len(files) == 1, "reports=%d", len(files)) {
			return report, "", code
		}
		data, e := os.ReadFile(files[0])
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(data, &report); e != nil {
			t.Fatal(e)
		}
		cs.Fixture("candidate-report.json", data)
		cs.Check("candidate is never formal", !report.FormalQualified, "unexpected promotion")
		cs.Check("ordinary process has no vendor HTTP", report.VendorHTTPAttempts == 0, "vendor request escaped")
		return report, filepath.Dir(files[0]), code
	}
	const entry = "^TestXAIImagesCandidate$"
	baseline, _, code := child(t, "missing-permission", entry, nil)
	if !cs.Check("missing permission is honest NOT_RUN", code == 0 && baseline.Outcome == "NOT_RUN" && baseline.CallsUsed == 0, "outcome=%s exit=%d", baseline.Outcome, code) {
		return
	}
	grant := func() livegrant.Grant {
		scope := baseline.Scope
		scope.Mode = "controlled"
		return livegrant.Grant{Schema: 1, ID: fmt.Sprintf("imagine-controlled-%d", time.Now().UnixNano()), AccountAlias: "synthetic", RegionAlias: "global", ExpiresAt: time.Now().UTC().Add(time.Hour), Scope: scope}
	}
	write := func(t *testing.T, g livegrant.Grant) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "grant.json")
		data, _ := json.Marshal(g)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	controlled := func(mode, path string) []string {
		return []string{"BARNESS_XAI_IMAGES_CONTROLLED=" + mode, "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + path}
	}
	for _, name := range []string{"no-key", "no-grant", "source", "binary", "catalog", "identity", "model", "endpoint", "budget", "usd", "expiry", "alias", "entry", "mixed-key", "live-in-controlled", "proxy-userinfo", "proxy-query", "proxy-remote", "proxy-https", "proxy-path", "proxy-fragment", "proxy-mode-missing"} {
		t.Run(name, func(t *testing.T) {
			g := grant()
			selected := entry
			want := "FAIL"
			var extra []string
			switch name {
			case "no-key":
				extra = []string{"BARNESS_AI_LIVE=1"}
				want = "NOT_RUN"
			case "no-grant":
				extra = []string{"BARNESS_XAI_IMAGES_CONTROLLED=fixture"}
				want = "NOT_RUN"
			case "source":
				g.Scope.SourceSHA256 = strings.Repeat("0", 64)
			case "binary":
				g.Scope.BinarySHA256 = strings.Repeat("0", 64)
			case "catalog":
				g.Scope.CatalogHash = "sha256:" + strings.Repeat("0", 64)
			case "identity":
				g.Scope.API = "openai-images"
			case "model":
				g.Scope.Models = []string{"grok-imagine-image"}
			case "endpoint":
				g.Scope.Endpoint = "https://other.api.x.ai/v1"
			case "budget":
				g.Scope.MaxAttempts++
			case "usd":
				g.Scope.MaxUSD = 0.61
			case "expiry":
				g.ExpiresAt = time.Now().Add(-time.Hour)
			case "alias":
				g.RegionAlias = "other"
			case "entry":
				selected = "TestXAIImagesCandidate"
			case "mixed-key":
				extra = []string{"OTHER_API_KEY=synthetic-foreign-credential"}
			case "live-in-controlled":
				extra = []string{"BARNESS_AI_LIVE=1"}
			case "proxy-userinfo":
				extra = []string{"BARNESS_AI_LIVE_PROXY=http://user:pass@127.0.0.1:7890"}
			case "proxy-query":
				extra = []string{"BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890?target=api.x.ai"}
			case "proxy-remote":
				extra = []string{"BARNESS_AI_LIVE_PROXY=http://example.invalid:7890"}
			case "proxy-https":
				extra = []string{"BARNESS_AI_LIVE_PROXY=https://127.0.0.1:7890"}
			case "proxy-path":
				extra = []string{"BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890/other"}
			case "proxy-fragment":
				extra = []string{"BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890#"}
			case "proxy-mode-missing":
				extra = []string{"BARNESS_XAI_IMAGES_CONTROLLED=proxy"}
			}
			env := controlled("fixture", write(t, g))
			if name == "no-key" || name == "no-grant" {
				env = extra
			} else {
				env = append(env, extra...)
			}
			report, _, exit := child(t, name, selected, env)
			cs := run.Case(t, "XAIIMG-process-assert-"+name)
			cs.Check("reject before claim or HTTP", report.Outcome == want && report.HTTPAttempts == 0 && report.CallsUsed == 0 && (exit == 0) == (want == "NOT_RUN"), "outcome=%s HTTP=%d exit=%d", report.Outcome, report.HTTPAttempts, exit)
		})
	}

	t.Run("proxy-configured-fixture", func(t *testing.T) {
		env := append(controlled("fixture", write(t, grant())), "BARNESS_AI_LIVE_PROXY=http://127.0.0.1:1")
		report, _, exit := child(t, "proxy-configured-fixture", entry, env)
		cs := run.Case(t, "XAIIMG-process-assert-proxy-configured-fixture")
		cs.Check("explicit proxy configuration accepted without bypassing finite fixture scope", exit == 0 && report.Outcome == "PASS" && report.Complete && report.CallsUsed == 6 && report.HTTPAttempts == 6 && report.Proxy == "http://127.0.0.1:1", "report=%+v exit=%d", report, exit)
	})
	for _, proxyStatus := range []int{http.StatusBadGateway, http.StatusOK} {
		name := fmt.Sprintf("proxy-connect-%d", proxyStatus)
		t.Run(name, func(t *testing.T) {
			type request struct{ Method, Host, Authorization, ProxyAuthorization string }
			var mu sync.Mutex
			var requests []request
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				requests = append(requests, request{r.Method, r.Host, r.Header.Get("Authorization"), r.Header.Get("Proxy-Authorization")})
				mu.Unlock()
				w.WriteHeader(proxyStatus)
			}))
			defer proxy.Close()
			env := append(controlled("proxy", write(t, grant())), "BARNESS_AI_LIVE_PROXY="+proxy.URL)
			report, _, exit := child(t, name, entry, env)
			cs := run.Case(t, "XAIIMG-process-assert-"+name)
			mu.Lock()
			captured := append([]request(nil), requests...)
			mu.Unlock()
			cs.Record("proxy-requests", captured)
			cs.Check("real transport CONNECT remains in local proxy", len(captured) == 1 && captured[0].Method == "CONNECT" && captured[0].Host == "api.x.ai:443" && captured[0].Authorization == "" && captured[0].ProxyAuthorization == "", "requests=%+v", captured)
			cs.Check("failed local proxy consumes one bounded synthetic attempt", exit == 1 && report.Outcome == "FAIL" && !report.Complete && report.CallsUsed == 1 && report.HTTPAttempts == 1 && report.VendorHTTPAttempts == 0 && report.Proxy == proxy.URL, "report=%+v exit=%d", report, exit)
			for i, row := range report.Scenarios {
				want := "NOT_RUN"
				if i == 0 {
					want = "FAIL"
				}
				cs.Check("proxy scenario "+row.ID, row.Outcome == want, "outcome=%s", row.Outcome)
			}
		})
	}
	t.Run("finite-plan", func(t *testing.T) {
		g := grant()
		path := write(t, g)
		report, dir, exit := child(t, "finite-plan", entry, controlled("fixture", path))
		cs := run.Case(t, "XAIIMG-process-assert-finite-plan")
		cs.Check("complete finite controlled scope", exit == 0 && report.Outcome == "PASS" && report.Complete && report.AuditPassed && report.CallsUsed == 6 && report.HTTPAttempts == 6 && report.InputImages == 8 && report.OutputImages == 7 && len(report.Scenarios) == 6, "report=%+v exit=%d", report, exit)
		for _, row := range report.Scenarios {
			cs.Check("planned scenario "+row.ID, row.Outcome == "PASS", "outcome=%s", row.Outcome)
		}
		cs.Check("independent charge and reservation", report.ReservedUSD > 0.529999 && report.ReservedUSD < 0.530001 && report.ReportedUSD > 0.529999 && report.ReportedUSD < 0.530001, "reserved=%f reported=%f", report.ReservedUSD, report.ReportedUSD)
		capture, err := os.ReadFile(filepath.Join(dir, "XAIIMG-live-generation-low", "http.json"))
		var exchanges []struct {
			ResponseBody string
			Truncated    bool
		}
		parseErr := json.Unmarshal(capture, &exchanges)
		cs.Check("full response above old capture limit", err == nil && parseErr == nil && len(exchanges) == 1 && len(exchanges[0].ResponseBody) > 512<<10 && !exchanges[0].Truncated && strings.Contains(exchanges[0].ResponseBody, "400000000"), "capture error=%v parse=%v", err, parseErr)
		again, _, reused := child(t, "claim-reuse", entry, controlled("fixture", path))
		cs.Check("one-shot approval cannot be reused", reused == 1 && again.Outcome == "FAIL" && again.HTTPAttempts == 0, "report=%+v exit=%d", again, reused)
		consumer := filepath.Join(t.TempDir(), "releasegate")
		buildConsumer := exec.CommandContext(ctx, "go", "build", "-o", consumer, "./ai/release/cmd/releasegate")
		buildConsumer.Dir = "../.."
		buildConsumer.Env = xaiProcessEnv(t)
		output, err := buildConsumer.CombinedOutput()
		if err != nil {
			t.Fatalf("consumer build: %v %s", err, output)
		}
		command := exec.CommandContext(ctx, consumer, "-live-bundle", dir, "-out", filepath.Join(t.TempDir(), "consumer"))
		command.Dir = "../.."
		command.Env = xaiProcessEnv(t)
		output, err = command.CombinedOutput()
		cs.Record("formal-consumer", map[string]any{"argv": command.Args, "exitCode": imagineProcessExit(err), "output": string(output)})
		cs.Check("controlled candidate cannot qualify formal", imagineProcessExit(err) == 1, "formal consumer exit=%d", imagineProcessExit(err))
	})
	for _, mode := range []string{"cancel", "final-write"} {
		t.Run(mode, func(t *testing.T) {
			report, _, exit := child(t, mode, entry, controlled(mode, write(t, grant())))
			cs := run.Case(t, "XAIIMG-process-assert-"+mode)
			cs.Check("retained failure is not success", exit == 1 && report.Outcome == "FAIL" && !report.Complete && report.CallsUsed > 0, "report=%+v exit=%d", report, exit)
		})
	}
}

func imagineProcessExit(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}
