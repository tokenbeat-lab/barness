package jevformal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

type Resources struct {
	Calls         int `json:"calls"`
	Attempts      int `json:"attempts"`
	Questions     int `json:"questions"`
	Sources       int `json:"sources"`
	References    int `json:"references"`
	SourceBytes   int `json:"sourceBytes"`
	ExpandedBytes int `json:"expandedBytes"`
	StateBytes    int `json:"stateBytes"`
}

type Protocol struct {
	Schema         int       `json:"schema"`
	Kind           string    `json:"kind"`
	Route          string    `json:"route"`
	Mode           string    `json:"mode"`
	Outcome        string    `json:"outcome"`
	Complete       bool      `json:"complete"`
	AuditPassed    bool      `json:"auditPassed"`
	GitCommit      string    `json:"gitCommit"`
	CatalogVersion string    `json:"catalogVersion"`
	CatalogHash    string    `json:"catalogHash"`
	StartedAt      time.Time `json:"startedAt"`
	FinishedAt     time.Time `json:"finishedAt"`
	Expected       []string  `json:"expected"`
	Authorization  struct {
		AccountAlias string `json:"accountAlias"`
		RegionAlias  string `json:"regionAlias"`
	} `json:"authorization"`
	SourceHashes map[string]string `json:"sourceHashes"`
	Budget       struct {
		Actual               Resources `json:"actual"`
		ReportedInputTokens  int       `json:"reportedInputTokens"`
		ReportedOutputTokens int       `json:"reportedOutputTokens"`
	} `json:"budget"`
}

type Proof struct {
	Protocol                  Protocol
	ProtocolHash, EffectsHash string
}

func repository() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", errors.New("repository root unavailable")
		}
		root = parent
	}
}

// Each verifier has a bounded process lifetime and receives no route credentials.
func command(root, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = root
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOMODCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "PYTHONDONTWRITEBYTECODE=1")
	// Verifiers emit bounded diagnostics only; never return filesystem or content
	// from child output through the matrix's public error.
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return errors.New("independent formal verifier rejected evidence")
	}
	return nil
}

// CheckInputs never derives formal identity from the submitted catalog itself.
func CheckInputs(protocol, effects string, requireLive bool) (Proof, error) {
	var out Proof
	root, err := repository()
	if err != nil {
		return out, err
	}
	protocol, err = filepath.Abs(protocol)
	if err != nil {
		return out, err
	}
	effects, err = filepath.Abs(effects)
	if err != nil {
		return out, err
	}
	// Bound file count and aggregate bytes before either independent subprocess
	// or audit traverses an externally supplied package.
	if _, err = inventory(protocol); err != nil {
		return out, err
	}
	if _, err = inventory(effects); err != nil {
		return out, err
	}
	raw, err := Read(filepath.Join(protocol, "multimodal-report.json"))
	if err != nil {
		return out, err
	}
	if json.Unmarshal(raw, &out.Protocol) != nil {
		return out, errors.New("invalid protocol report")
	}
	p := out.Protocol
	cat, err := evidence.MeasuredCatalog(protocol, p.CatalogVersion, p.CatalogHash, ai.BuiltinCatalog())
	if err != nil {
		return out, err
	}
	if err := evidence.VerifyRouteCatalog(cat, ai.BuiltinCatalog(), ai.OperationClassifier, ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, []string{"jev-1.14.0"}); err != nil {
		return out, err
	}
	hash, err := cat.Hash()
	if err != nil {
		return out, err
	}
	if p.Schema != 1 || p.Kind != "jev-formal-protocol" || p.Route != "jev" || !p.Complete || !p.AuditPassed || p.Outcome != "PASS" || p.CatalogHash != hash || p.CatalogVersion != cat.Version || (p.Mode != "live" && p.Mode != "controlled") || (requireLive && p.Mode != "live") {
		return out, errors.New("wrong fixed formal protocol identity or provenance")
	}
	var effect struct{ CatalogKind, Mode, Route, Model, Status, RealQuality, CatalogHash, CatalogVersion string }
	raw, err = Read(filepath.Join(effects, "report.json"))
	if err != nil {
		return out, err
	}
	if json.Unmarshal(raw, &effect) != nil || effect.CatalogKind != "formal" || effect.Mode != p.Mode || effect.Route != "jev" || effect.Model != "jev-1.14.0" || effect.Status != "COMPLETE" || effect.CatalogHash != hash || effect.CatalogVersion != cat.Version || (requireLive && effect.RealQuality != "COMPLETE") {
		return out, errors.New("wrong independent formal effect identity or completeness")
	}
	// A dirty or incomplete source tree cannot borrow another formal run's PASS.
	want := map[string]string{}
	paths := []string{"../../go.mod", "../../go.sum", "../internal/testkit/jevformal/verify_protocol.py"}
	paths = append(paths, "../live/testdata/jev-formal/sources.json")
	err = filepath.WalkDir(filepath.Join(root, "ai"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			rel, e := filepath.Rel(filepath.Join(root, "ai"), path)
			if e != nil {
				return e
			}
			paths = append(paths, "../"+filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return out, err
	}
	for _, path := range paths {
		raw, err := Read(filepath.Join(root, "ai/live", filepath.FromSlash(path)))
		if err != nil {
			return out, err
		}
		want[path] = Hash(raw)
	}
	measured, current := map[string]string{}, map[string]string{}
	for name, h := range p.SourceHashes {
		measured[filepath.ToSlash(filepath.Clean(filepath.Join("ai/live", name)))] = h
	}
	for name, h := range want {
		current[filepath.ToSlash(filepath.Clean(filepath.Join("ai/live", name)))] = h
	}
	if err := evidence.VerifyRouteSource(measured, current, "typesafe", "typesafe-system-one"); err != nil {
		return out, err
	}
	raw, err = Read(filepath.Join(protocol, "MM11-configuration/formal-sources.json"))
	if err != nil || Hash(raw) != want["../live/testdata/jev-formal/sources.json"] {
		return out, errors.New("formal primary source snapshot drift")
	}
	file, err := os.CreateTemp("", "barness-formal-catalog-*.json")
	if err != nil {
		return out, err
	}
	defer os.Remove(file.Name())
	raw, err = json.Marshal(cat)
	if err == nil {
		_, err = file.Write(raw)
	}
	closeErr := file.Close()
	if err != nil {
		return out, err
	}
	if closeErr != nil {
		return out, closeErr
	}
	args := []string{filepath.Join(root, "ai/internal/testkit/jevformal/verify_protocol.py"), "--formal-catalog", file.Name()}
	if requireLive {
		args = append(args, "--require-live")
	}
	args = append(args, protocol)
	if err = command(root, "python3", args...); err != nil {
		return out, err
	}
	if err = ReplayProtocol(protocol, p); err != nil {
		return out, err
	}
	args = []string{"run", "./ai/examples/visualeval", "-verify", effects}
	if requireLive {
		args = append(args, "-require-live")
	}
	if err = command(root, "go", args...); err != nil {
		return out, err
	}
	raw, err = Read(filepath.Join(protocol, "multimodal-integrity.json"))
	if err != nil {
		return out, err
	}
	out.ProtocolHash = Hash(raw)
	raw, err = Read(filepath.Join(effects, "manifest.json"))
	if err != nil {
		return out, err
	}
	out.EffectsHash = Hash(raw)
	return out, nil
}

type Capture struct {
	Metadata               ai.CallMetadata           `json:"metadata"`
	ImageInputs            []ai.ClassifierImageInput `json:"imageInputs"`
	StateBytes, ImageBytes int
	DurationMS             int64
}

func Result(protocol, id string) (Capture, int, error) {
	var value struct {
		Result Capture `json:"result"`
	}
	raw, err := Read(filepath.Join(protocol, "MM11-"+id, "result.json"))
	if err != nil {
		return value.Result, 0, err
	}
	if err = json.Unmarshal(raw, &value); err != nil {
		return value.Result, 0, err
	}
	var req struct {
		Questions map[string]json.RawMessage `json:"questions"`
	}
	raw, err = Read(filepath.Join(protocol, "MM11-"+id, "public-request.json"))
	if err != nil {
		return value.Result, 0, err
	}
	if err = json.Unmarshal(raw, &req); err != nil {
		return value.Result, 0, fmt.Errorf("invalid public request")
	}
	var duration struct {
		Milliseconds int64 `json:"milliseconds"`
	}
	raw, err = Read(filepath.Join(protocol, "MM11-"+id, "duration.json"))
	if err != nil {
		return value.Result, 0, err
	}
	if json.Unmarshal(raw, &duration) != nil || duration.Milliseconds < 0 || duration.Milliseconds > 30000 {
		return value.Result, 0, errors.New("invalid captured call duration")
	}
	value.Result.DurationMS = duration.Milliseconds
	var exchanges []struct {
		RequestBody string `json:"requestBody"`
	}
	raw, err = Read(filepath.Join(protocol, "MM11-"+id, "exchanges.json"))
	if err != nil {
		return value.Result, 0, err
	}
	if json.Unmarshal(raw, &exchanges) != nil || len(exchanges) != 1 {
		return value.Result, 0, errors.New("missing final native request")
	}
	var native struct {
		State  any `json:"state"`
		Assets map[string]struct {
			Data string `json:"data_base64"`
		} `json:"assets"`
	}
	if json.Unmarshal([]byte(exchanges[0].RequestBody), &native) != nil {
		return value.Result, 0, errors.New("invalid native resource capture")
	}
	state, err := json.Marshal(native.State)
	if err != nil {
		return value.Result, 0, err
	}
	value.Result.StateBytes = len(state)
	for _, asset := range native.Assets {
		decoded, err := base64.StdEncoding.Strict().DecodeString(asset.Data)
		if err != nil {
			return value.Result, 0, err
		}
		value.Result.ImageBytes += len(decoded)
	}
	return value.Result, len(req.Questions), nil
}
