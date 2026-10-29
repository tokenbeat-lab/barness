package supportmatrix

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

const XAIImagesReportFile = "xai-images-candidate-report.json"
const xaiImagesMeasuredSource = "7a77c0cfb9f7abe0a838ebedd5d9d72074758abfff41e1bcdf9e89a35f014113"
const xaiImagesMeasuredCommit = "724e988132ac5bf6ac393db156e07081eb0e2427"

type xaiImagesNative struct {
	Proxy             string            `json:"proxy,omitempty"`
	Schema            int               `json:"schema"`
	Kind              string            `json:"kind"`
	Mode              string            `json:"mode"`
	Stage             string            `json:"stage"`
	Outcome           string            `json:"outcome"`
	Failure           string            `json:"failure,omitempty"`
	Complete          bool              `json:"complete"`
	FormalQualified   bool              `json:"formalQualified"`
	AuditPassed       bool              `json:"auditPassed"`
	StartedAt         time.Time         `json:"startedAt"`
	FinishedAt        time.Time         `json:"finishedAt"`
	GitCommit         string            `json:"gitCommit"`
	Scope             livegrant.Scope   `json:"scope"`
	BuildSourceSHA256 string            `json:"buildSourceSha256"`
	Plan              []xaiimages.Scene `json:"plan"`
	Scenarios         []struct {
		ID      string `json:"id"`
		Outcome string `json:"outcome"`
		CaseID  string `json:"caseId"`
	} `json:"scenarios"`
	CallsUsed          int     `json:"callsUsed"`
	HTTPAttempts       int     `json:"httpAttempts"`
	VendorHTTPAttempts int     `json:"vendorHTTPAttempts"`
	InputImages        int     `json:"inputImages"`
	OutputImages       int     `json:"outputImages"`
	ReservedUSD        float64 `json:"reservedUSD"`
	ReportedUSD        float64 `json:"reportedUSD"`
}

// IsXAIImagesCapture identifies this route without accepting its claims.
func IsXAIImagesCapture(dir string) bool {
	info, e := os.Lstat(filepath.Join(dir, XAIImagesReportFile))
	return e == nil && info.Mode().IsRegular()
}

// NewXAIImagesReport qualifies original live candidate facts offline. Its
// projection retains the measured identity/time/catalog; new consumer time and
// current source live in the outer receipt. No original file is rewritten.
func NewXAIImagesReport(dir string) (Report, error) {
	zero := Report{}
	before, e := xaiInventory(dir)
	if e != nil {
		return zero, e
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		return zero, e
	}
	defer root.Close()
	for name := range before {
		// The exact candidate catalog legitimately encodes its unused chat
		// model list as null; its whole value is checked independently.
		if strings.HasSuffix(name, ".json") && name != "model-catalog.json" {
			if e = xaiImagesNoNull(root, name); e != nil {
				return zero, e
			}
		}
	}
	var n xaiImagesNative
	if e = xaiReadJSON(root, XAIImagesReportFile, &n); e != nil {
		return zero, e
	}
	if e = xaiImagesNativeFields(root, n); e != nil {
		return zero, e
	}
	manifest, e := verifyXAIImagesOrigin(root, dir, n, before)
	if e != nil {
		return zero, e
	}
	if _, e = verifyXAIImagesProducer(dir, n, before); e != nil {
		return zero, e
	}
	result := Report{Schema: SchemaVersion, Combo: xaiimages.Route, Operation: "image", Provider: "xai", API: "xai-images", Endpoint: xaiimages.Endpoint, Model: xaiimages.Model, SDK: xaiimages.SDK, GitCommit: n.GitCommit, CatalogVersion: n.Scope.CatalogVersion, CatalogHash: n.Scope.CatalogHash, StartedAt: n.StartedAt, FinishedAt: n.FinishedAt, Expected: xaiimages.Expected(), Budget: Budget{MaxCalls: 6, MaxAttempts: 6, HTTPAttempts: 6, CallsUsed: 6, MaxImages: 7, ImagesUsed: 7, CallSeconds: 180, ProcessSeconds: 1200}, evidenceDir: dir}
	grant, _, _, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return zero, e
	}
	result.AccountAlias = grant.AccountAlias
	facts, e := xaiImagesMeasured()
	if e != nil {
		return zero, e
	}
	ticks := int64(0)
	seen := map[string]bool{}
	for i, scene := range xaiimages.Plan() {
		row, total, e := verifyXAIImagesCase(root, n, scene, manifest.Cases, seen, facts[scene.ID])
		if e != nil {
			return zero, e
		}
		if ticks > int64(^uint64(0)>>1)-total {
			return zero, refuse("Imagine reported ticks overflow")
		}
		ticks += total
		if n.Scenarios[i].ID != scene.ID || n.Scenarios[i].CaseID != row.CaseID || n.Scenarios[i].Outcome != "PASS" {
			return zero, refuse("Imagine scene coverage differs")
		}
		result.Scenarios = append(result.Scenarios, row)
	}
	if !xaiAmountEqual(float64(ticks)/1e10, n.ReportedUSD) || ticks != 5100000000 {
		return zero, refuse("Imagine original tick totals differ")
	}
	if e = VerifyReportShape(result); e != nil {
		return zero, e
	}
	after, e := xaiInventory(dir)
	if e != nil || !reflect.DeepEqual(before, after) {
		return zero, refuse("Imagine original evidence changed during qualification")
	}
	return result, nil
}
func xaiImagesNativeFields(root *os.Root, n xaiImagesNative) error {
	raw, e := regionalRead(root, XAIImagesReportFile)
	var fields map[string]json.RawMessage
	if e != nil || regionalDecode(raw, &fields) != nil {
		return refuse("Imagine report unavailable")
	}
	for _, k := range []string{"schema", "kind", "mode", "stage", "outcome", "complete", "formalQualified", "auditPassed", "startedAt", "finishedAt", "gitCommit", "scope", "buildSourceSha256", "plan", "scenarios", "callsUsed", "httpAttempts", "vendorHTTPAttempts", "inputImages", "outputImages", "reservedUSD", "reportedUSD"} {
		if len(fields[k]) == 0 || string(fields[k]) == "null" {
			return refuse("Imagine required report field absent")
		}
	}
	if n.Schema != 1 || n.Kind != "xai-images-candidate-only" || n.Mode != "live" || n.Stage != "candidate" || n.Outcome != "PASS" || !n.Complete || n.FormalQualified || !n.AuditPassed || n.Failure != "" || n.GitCommit != xaiImagesMeasuredCommit || n.BuildSourceSHA256 != xaiImagesMeasuredSource || n.StartedAt.IsZero() || n.FinishedAt.Before(n.StartedAt) || n.FinishedAt.Sub(n.StartedAt) > 1200*time.Second || !reflect.DeepEqual(n.Plan, xaiimages.Plan()) || len(n.Scenarios) != 6 || n.CallsUsed != 6 || n.HTTPAttempts != 6 || n.VendorHTTPAttempts != 6 || n.InputImages != 8 || n.OutputImages != 7 || !xaiAmountEqual(n.ReservedUSD, .53) || !xaiAmountEqual(n.ReportedUSD, .51) {
		return refuse("Imagine original candidate identity, coverage or budget differs")
	}
	// Recheck the original harness's explicit loopback-only metadata. This
	// validator does not create a transport or consult ambient proxy settings.
	if n.Proxy != "" {
		u, e := url.Parse(n.Proxy)
		if e != nil {
			return refuse("Imagine explicit proxy metadata invalid")
		}
		port, e := strconv.Atoi(u.Port())
		ip := net.ParseIP(u.Hostname())
		if e != nil || port < 1 || port > 65535 || ip == nil || !ip.IsLoopback() || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
			return refuse("Imagine explicit proxy metadata invalid")
		}
	}
	return nil
}

// This frozen native contract never emitted JSON null. A null scalar must not
// acquire a Go zero value and silently become an absent or successful fact.
func xaiImagesNoNull(root *os.Root, name string) error {
	raw, e := regionalRead(root, name)
	if e != nil {
		return e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	for {
		token, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil || token == nil {
			return refuse("Imagine explicit null or invalid native JSON")
		}
	}
}

// VerifyXAIImagesReport authenticates a projection against its native capture.
// A serialized generic PASS never authenticates its own native captures.
func VerifyXAIImagesReport(r Report) error {
	if r.Combo != "xai-images" {
		return nil
	}
	if r.evidenceDir == "" {
		return refuse("Imagine support needs its original native capture")
	}
	actual, e := NewXAIImagesReport(r.evidenceDir)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(actual, r) {
		return refuse("Imagine projection changed after native qualification")
	}
	return nil
}
