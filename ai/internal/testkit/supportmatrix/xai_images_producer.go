package supportmatrix

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

type xaiImagesProducer struct {
	Schema     int       `json:"schema"`
	Feature    string    `json:"feature"`
	Stage      string    `json:"stage"`
	Route      string    `json:"route"`
	Status     string    `json:"status"`
	ExitCode   *int      `json:"exitCode"`
	Argv       []string  `json:"argv"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	Source     struct {
		Commit string            `json:"commit"`
		Files  map[string]string `json:"files"`
	} `json:"source"`
	Files          map[string]string  `json:"files"`
	SelectedInputs map[string]string  `json:"selectedInputs"`
	CommandSHA256  string             `json:"commandSHA256"`
	Toolchain      xaiImagesToolchain `json:"toolchain"`
	Bundles        []struct {
		Path     string `json:"path"`
		Manifest string `json:"manifestSHA256"`
		Audit    string `json:"auditSHA256"`
		Package  string `json:"package"`
	} `json:"bundles"`
}
type xaiImagesToolchain struct {
	Executable string  `json:"executableSHA256"`
	GoEnv      string  `json:"goEnv"`
	GoFlags    *string `json:"goFlags"`
}

func verifyXAIImagesProducer(dir string, n xaiImagesNative, inventory map[string]string) (map[string]string, error) {
	inputs := map[string]string{}
	root, e := minimaxlive.Root()
	if e != nil {
		return nil, e
	}
	if filepath.Base(filepath.Dir(dir)) != "bundles" {
		return nil, refuse("Imagine native bundle lacks original producer parent")
	}
	attempt := filepath.Dir(filepath.Dir(dir))
	opened, e := os.OpenRoot(attempt)
	if e != nil {
		return nil, e
	}
	defer opened.Close()
	native, e := os.OpenRoot(dir)
	if e != nil {
		return nil, e
	}
	defer native.Close()
	grant, grantRaw, claim, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return nil, e
	}
	var strictGrant livegrant.Grant
	var strictClaim livegrant.Receipt
	if xaiReadJSON(native, "process-grant.json", &strictGrant) != nil || xaiReadJSON(native, "process-authorization-receipt.json", &strictClaim) != nil || !reflect.DeepEqual(grant, strictGrant) || !reflect.DeepEqual(claim, strictClaim) {
		return nil, refuse("Imagine original authorization JSON is not canonical")
	}
	candidate := xai.ImagesCandidateCatalog()
	hash, _ := candidate.Hash()
	want := livegrant.Scope{Route: xaiimages.Route, Mode: "live", Stage: "candidate", Provider: "xai", API: "xai-images", Endpoint: xaiimages.Endpoint, SDK: xaiimages.SDK, CatalogVersion: candidate.Version, CatalogHash: hash, SourceSHA256: xaiImagesMeasuredSource, BinarySHA256: n.Scope.BinarySHA256, Models: []string{xaiimages.Model}, Scenes: xaiimages.Expected(), MaxCalls: 6, MaxAttempts: 6, CallSeconds: 180, ProcessSeconds: 1200, MaxUSD: .60}
	if !reflect.DeepEqual(n.Scope, want) || !reflect.DeepEqual(grant.Scope, want) || claim.Session != filepath.Base(dir) {
		return nil, refuse("Imagine original candidate scope or session differs")
	}
	if e = livegrant.VerifyEvidence(grant, grantRaw, claim, want, grant.AccountAlias, grant.RegionAlias, n.StartedAt, n.FinishedAt); e != nil {
		return nil, e
	}
	raw, e := regionalRead(opened, "receipt.json")
	var p xaiImagesProducer
	if e != nil || regionalDecode(raw, &p) != nil || p.Schema != 1 || p.Feature != "barness-ai-xai-imagine" || p.Stage != "live" || p.Route != xaiimages.Route || p.Status != "PASS" || p.ExitCode == nil || *p.ExitCode != 0 || p.Source.Commit != n.GitCommit || p.StartedAt.IsZero() || p.StartedAt.After(n.StartedAt) || p.FinishedAt.Before(n.FinishedAt) || len(p.Argv) != 4 || p.Toolchain.Executable != want.BinarySHA256 || p.Toolchain.GoEnv != "off" || p.Toolchain.GoFlags == nil || *p.Toolchain.GoFlags != "" {
		return nil, refuse("Imagine producer actual exit, invocation or source differs")
	}
	for i, arg := range []string{"-test.run=" + xaiimages.Entry, "-test.count=1", "-test.timeout=1200s"} {
		if p.Argv[i+1] != arg {
			return nil, refuse("Imagine producer entry/count/timeout differs")
		}
	}
	rel, e := filepath.Rel(root, attempt)
	if e != nil || !filepath.IsLocal(rel) {
		return nil, refuse("Imagine producer must be repository local")
	}
	prefix := filepath.ToSlash(rel) + "/"
	inputs["@xai-producer/"+prefix+"receipt.json"] = glmHash(raw)
	if len(p.Files) != 4 {
		return nil, refuse("Imagine producer stream inventory differs")
	}
	for _, name := range []string{"command.json", "stdout.jsonl", "stderr.log", "exit-code"} {
		b, e := regionalRead(opened, name)
		if e != nil || glmHash(b) != p.Files[name] || !validManifestHash(p.Files[name]) || name == "exit-code" && string(b) != "0\n" || name == "stdout.jsonl" && strings.TrimSpace(string(b)) != "PASS" {
			return nil, refuse("Imagine producer stream or numeric exit differs")
		}
		inputs["@xai-producer/"+prefix+name] = glmHash(b)
	}
	var command struct {
		Schema           int
		Argv             []string
		Stage, Route     string
		Timeout          float64 `json:"timeoutSeconds"`
		EnvironmentNames []string
		Toolchain        xaiImagesToolchain
	}
	b, e := regionalRead(opened, "command.json")
	if e != nil || regionalDecode(b, &command) != nil || glmHash(b) != p.CommandSHA256 || p.CommandSHA256 != p.Files["command.json"] || command.Schema != 1 || command.Stage != "live" || command.Route != xaiimages.Route || command.Timeout != 1260 || !reflect.DeepEqual(command.Argv, p.Argv) || !reflect.DeepEqual(command.Toolchain, p.Toolchain) {
		return nil, refuse("Imagine producer command differs")
	}
	if slices.Contains(command.EnvironmentNames, "BARNESS_XAI_IMAGES_CONTROLLED") {
		return nil, refuse("Imagine controlled producer cannot qualify live")
	}
	for _, key := range []string{"BARNESS_AI_LIVE", "BARNESS_AI_LIVE_KEY_XAI_IMAGES", "BARNESS_AI_LIVE_PERMISSIONS_FILE", "GOENV", "GOFLAGS"} {
		if !slices.Contains(command.EnvironmentNames, key) {
			return nil, refuse("Imagine original live environment contract missing")
		}
	}
	if len(p.Bundles) != 1 || p.Bundles[0].Path != "bundles/"+claim.Session || p.Bundles[0].Manifest != inventory["manifest.json"] || p.Bundles[0].Audit != inventory["audit.json"] || p.Bundles[0].Package != "-tags live -timeout 20m0s ./ai/live" {
		return nil, refuse("Imagine producer does not bind unique native bundle")
	}
	var original map[string]string
	if xaiReadJSON(native, "execution-source.json", &original) != nil {
		return nil, refuse("Imagine original source absent")
	}
	for name, h := range original {
		compiled, inSource := p.Source.Files[name]
		selected, inInputs := p.SelectedInputs[name]
		if !inSource && !inInputs || inSource && compiled != h || inInputs && selected != h {
			return nil, refuse("Imagine producer omitted or changed an original execution input")
		}
	}
	binary := p.Argv[0]
	if filepath.IsAbs(binary) {
		binary, e = filepath.Rel(root, binary)
		if e != nil {
			return nil, e
		}
	}
	binary = filepath.ToSlash(binary)
	path, e := minimaxlive.LocalEvidencePath(root, binary)
	if e != nil {
		return nil, e
	}
	binaryHash, e := livegrant.FileHash(path)
	if e != nil || binaryHash != want.BinarySHA256 || p.SelectedInputs[binary] != binaryHash {
		return nil, refuse("Imagine retained original executable differs")
	}
	if len(p.SelectedInputs) != 5 {
		return nil, refuse("Imagine producer selected input set is incomplete")
	}
	required := map[string]bool{"grant.json": false, "user-approval.md": false, "approved-retry-plan.md": false, "proxy-selection.json": false}
	for name, h := range p.SelectedInputs {
		path, e := minimaxlive.LocalEvidencePath(root, name)
		if e != nil {
			return nil, e
		}
		actual, e := livegrant.FileHash(path)
		if e != nil || actual != h {
			return nil, refuse("Imagine selected original input changed")
		}
		inputs["@xai-producer/"+name] = h
		base := filepath.Base(name)
		if _, ok := required[base]; ok {
			required[base] = true
			b, e := os.ReadFile(path)
			if e != nil || len(b) > 1<<20 {
				return nil, refuse("Imagine selected approval evidence exceeds bound")
			}
			if base == "grant.json" && string(b) != string(grantRaw) {
				return nil, refuse("Imagine selected raw grant differs")
			}
			if base == "proxy-selection.json" {
				var proxy struct {
					Proxy                 string
					ExplicitUserSelection bool
					Target                string
					AmbientProxyUsed      bool
				}
				if regionalDecode(b, &proxy) != nil || proxy.Proxy != n.Proxy || proxy.Target != xaiimages.Endpoint || proxy.AmbientProxyUsed || (proxy.Proxy != "") != proxy.ExplicitUserSelection {
					return nil, refuse("Imagine original explicit proxy selection differs")
				}
			}
		}
	}
	for _, present := range required {
		if !present {
			return nil, refuse("Imagine original user approval input absent")
		}
	}
	authority, e := livegrant.ClaimEvidenceInputs(dir)
	if e != nil {
		return nil, e
	}
	for k, h := range authority {
		inputs["@xai-authority/"+k] = h
	}
	return inputs, nil
}

// XAIImagesInputs includes transitive producer and trusted claim originals.
// It neither claims another approval nor manufactures a nonexistent ledger.
func XAIImagesInputs(dir string) (map[string]string, error) {
	if _, e := NewXAIImagesReport(dir); e != nil {
		return nil, e
	}
	inventory, e := xaiInventory(dir)
	if e != nil {
		return nil, e
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		return nil, e
	}
	defer root.Close()
	var n xaiImagesNative
	if e = xaiReadJSON(root, XAIImagesReportFile, &n); e != nil {
		return nil, e
	}
	extra, e := verifyXAIImagesProducer(dir, n, inventory)
	if e != nil {
		return nil, e
	}
	for k, v := range extra {
		inventory[k] = v
	}
	return inventory, nil
}
