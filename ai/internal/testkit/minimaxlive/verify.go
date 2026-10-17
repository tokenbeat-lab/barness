package minimaxlive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Seal struct {
	Schema int               `json:"schema"`
	Files  map[string]string `json:"files"`
}

func Inventory(dir string) (map[string]string, error) {
	out := map[string]string{}
	e := filepath.WalkDir(dir, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path == dir {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink artifact forbidden")
		}
		if d.IsDir() {
			return nil
		}
		name, _ := filepath.Rel(dir, path)
		name = filepath.ToSlash(name)
		if name == SealFile || name == "live-report.json" {
			return nil
		}
		raw, e := ReadFile(path, 16<<20)
		if e != nil {
			return e
		}
		out[name] = Hash(raw)
		return nil
	})
	return out, e
}
func Load(dir string) (Report, error) {
	var r Report
	raw, e := ReadFile(filepath.Join(dir, ReportFile), 2<<20)
	if e == nil {
		e = StrictJSON(raw, &r)
	}
	return r, e
}

// Verify recomputes exact files, source set, finite plan and business capture
// facts. formal=true additionally requires actual live + exact built-in M3.
func Verify(dir string, formal bool) (Report, error) {
	r, e := Load(dir)
	if e != nil {
		return r, e
	}
	raw, e := ReadFile(filepath.Join(dir, SealFile), 2<<20)
	if e != nil {
		return r, e
	}
	var seal Seal
	if e = StrictJSON(raw, &seal); e != nil {
		return r, e
	}
	files, e := Inventory(dir)
	if e != nil || seal.Schema != 1 || !reflect.DeepEqual(files, seal.Files) {
		return r, fmt.Errorf("MiniMax sealed file set changed")
	}
	if e = verifyArtifacts(dir, r, files); e != nil {
		return r, e
	}
	if !r.Complete || !r.AuditPassed || r.Outcome != "PASS" || r.Schema != 1 {
		return r, fmt.Errorf("MiniMax report is incomplete")
	}
	id, e := Select(r.Identity.Route, r.Identity.Model, r.Identity.Purpose)
	if e != nil || id != r.Identity {
		return r, fmt.Errorf("MiniMax identity mismatch")
	}
	source, e := ExecutionSource()
	if e != nil || !reflect.DeepEqual(source, r.ExecutionSource) {
		return r, fmt.Errorf("MiniMax execution source set changed")
	}
	sdk, err := SDKVersion(id)
	if err != nil || r.SDK != sdk {
		return r, fmt.Errorf("MiniMax SDK identity differs")
	}
	plan := Plan(id)
	raw, _ = json.Marshal(plan)
	if !reflect.DeepEqual(plan, r.Expected) || Hash(raw) != r.PlanHash || r.SourceHash != SourcePin {
		return r, fmt.Errorf("MiniMax fixed plan/source mismatch")
	}
	if r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.FinishedAt.Sub(r.StartedAt) > 1800e9 || r.SDK == "" || r.AccountAlias == "" || r.RegionAlias == "" || r.Entitlement != "confirmed" {
		return r, fmt.Errorf("MiniMax provenance incomplete")
	}
	cat := minimax.CandidateCatalog()
	if r.Stage == "formal" {
		cat = ai.BuiltinCatalog()
	}
	hash, err := cat.Hash()
	if err != nil || hash != r.CatalogHash || cat.Version != r.CatalogVersion {
		return r, fmt.Errorf("MiniMax selected catalog/price snapshot is stale")
	}
	if _, ok := cat.Lookup(ai.ProviderMiniMaxCN, id.API, id.Model); !ok {
		return r, fmt.Errorf("MiniMax formal identity is absent")
	}
	if formal && (r.Kind != "minimax-formal" || r.Mode != "live" || r.Stage != "formal" || id.Model != M3) {
		return r, fmt.Errorf("candidate, controlled and replay cannot certify formal support")
	}
	var manifest struct {
		Versions map[string]string `json:"versions"`
		Cases    []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"cases"`
	}
	raw, e = ReadFile(filepath.Join(dir, "manifest.json"), 2<<20)
	if e != nil || json.Unmarshal(raw, &manifest) != nil || len(manifest.Cases) != len(r.Calls) {
		return r, fmt.Errorf("MiniMax manifest inventory differs")
	}
	if manifest.Versions["minimax_mode"] != r.Mode || manifest.Versions["minimax_kind"] != r.Kind || manifest.Versions["minimax_source_hash"] != r.SourceHash || manifest.Versions["minimax_plan_hash"] != r.PlanHash {
		return r, fmt.Errorf("MiniMax manifest provenance differs")
	}
	if r.Mode != "live" && r.Mode != "controlled" && r.Mode != "replay" {
		return r, fmt.Errorf("unknown MiniMax execution mode")
	}
	if r.Kind != "minimax-candidate-only" && r.Kind != "minimax-formal" && r.Kind != "minimax-capture-replay-only" {
		return r, fmt.Errorf("unknown MiniMax evidence kind")
	}
	if formal && (r.GitCommit == "" || r.GitCommit == "unknown") {
		return r, fmt.Errorf("formal MiniMax Git provenance absent")
	}
	budget := Budget{}
	seen := map[string]int{}
	for index, call := range r.Calls {
		var p *Case
		for i := range plan {
			if plan[i].ID == call.ID {
				p = &plan[i]
			}
		}
		validRerun := call.ID == "tool-call" && seen[call.ID] == 1 && call.Rerun == 1
		validInitialFailure := call.ID == "tool-call" && call.Rerun == 0 && call.Outcome == "FAIL" && index+1 < len(r.Calls) && r.Calls[index+1].ID == call.ID && r.Calls[index+1].Rerun == 1 && r.Calls[index+1].Outcome == "PASS"
		if p == nil || seen[call.ID] > 0 && !validRerun || call.Rerun != 0 && !validRerun || call.CaseID != CaseID(call.ID, call.Rerun) || call.Outcome != "PASS" && !validInitialFailure {
			return r, fmt.Errorf("invented or failed MiniMax case")
		}
		seen[call.ID]++
		found := false
		for _, mc := range manifest.Cases {
			found = found || mc.ID == call.CaseID && mc.Status == "PASS"
		}
		if !found {
			return r, fmt.Errorf("MiniMax case absent from manifest")
		}
		if e = verifyCall(dir, r, call, *p); e != nil {
			return r, e
		}
		addBudget(&budget, call.Reservation)
		for _, a := range call.Attempts {
			addBudget(&budget, a)
		}
	}
	if len(seen) != len(plan) || !reflect.DeepEqual(budget, r.Budget) || budget.Calls > 20 || budget.Attempts > 24 || budget.RequestedOutput > 40000 || budget.ReservedInput > 120000 {
		return r, fmt.Errorf("MiniMax budget or complete inventory differs")
	}
	if e = VerifyLedger(dir, r); e != nil {
		return r, e
	}
	sources, e := Sources()
	if e != nil {
		return r, e
	}
	raw, e = ReadFile(filepath.Join(dir, "minimax-sources.json"), 2<<20)
	var captured map[string][]byte
	if e != nil || json.Unmarshal(raw, &captured) != nil || !reflect.DeepEqual(sources, captured) {
		return r, fmt.Errorf("MiniMax source provenance changed")
	}
	return r, nil
}
func StrictJSON(raw []byte, v any) error {
	if !json.Valid(raw) {
		return fmt.Errorf("invalid JSON")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 64 {
			return fmt.Errorf("excessive JSON nesting")
		}
		tok, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		seen := map[string]bool{}
		for d.More() {
			if delim == '{' {
				key, e := d.Token()
				if e != nil {
					return e
				}
				name := key.(string)
				if seen[name] {
					return fmt.Errorf("duplicate JSON field")
				}
				seen[name] = true
			}
			if e := walk(depth + 1); e != nil {
				return e
			}
		}
		_, e = d.Token()
		return e
	}
	if e := walk(0); e != nil {
		return e
	}
	var document any
	if e := json.Unmarshal(raw, &document); e != nil {
		return e
	}
	if e := canonical(document, reflect.TypeOf(v)); e != nil {
		return e
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func canonical(v any, t reflect.Type) error {
	if v == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		if obj, ok := v.(map[string]any); ok {
			for name, child := range obj {
				ft, ok := fields[name]
				if !ok {
					return fmt.Errorf("noncanonical JSON field")
				}
				if e := canonical(child, ft); e != nil {
					return e
				}
			}
		}
	case reflect.Map:
		if obj, ok := v.(map[string]any); ok {
			for _, child := range obj {
				if e := canonical(child, t.Elem()); e != nil {
					return e
				}
			}
		}
	case reflect.Slice:
		if a, ok := v.([]any); ok {
			for _, child := range a {
				if e := canonical(child, t.Elem()); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
