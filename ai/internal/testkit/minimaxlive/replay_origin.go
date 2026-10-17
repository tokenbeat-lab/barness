package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
)

// Snapshot preserves a verified source bundle without retaining a filesystem
// pointer. Every byte remains bounded and independently inventoried.
func Snapshot(dir string) (map[string][]byte, error) {
	files, e := Inventory(dir)
	if e != nil {
		return nil, e
	}
	files[SealFile] = ""
	out := map[string][]byte{}
	total := 0
	for name := range files {
		raw, e := ReadFile(filepath.Join(dir, filepath.FromSlash(name)), 16<<20)
		if e != nil {
			return nil, e
		}
		total += len(raw)
		if total > 12<<20 {
			return nil, fmt.Errorf("MiniMax source snapshot exceeds bound")
		}
		out[name] = raw
	}
	return out, nil
}
func verifyReplayOrigin(dir string, r Report) error {
	raw, e := ReadFile(filepath.Join(dir, "minimax-replay-origin.json"), 16<<20)
	if r.Mode != "replay" {
		if e == nil {
			return fmt.Errorf("non-replay carries replay provenance")
		}
		return nil
	}
	if e != nil {
		return e
	}
	var origin map[string][]byte
	if json.Unmarshal(raw, &origin) != nil {
		return fmt.Errorf("MiniMax replay source invalid")
	}
	var source Report
	if StrictJSON(origin[ReportFile], &source) != nil || Hash(origin[ReportFile]) != r.ReplaySource || source.Identity != r.Identity || !source.Complete || source.Outcome != "PASS" || source.Mode == "replay" || source.CatalogHash != r.CatalogHash || !reflect.DeepEqual(source.ExecutionSource, r.ExecutionSource) {
		return fmt.Errorf("MiniMax replay source provenance differs")
	}
	var seal Seal
	if StrictJSON(origin[SealFile], &seal) != nil || len(origin) != len(seal.Files)+1 {
		return fmt.Errorf("MiniMax replay source seal incomplete")
	}
	for name, hash := range seal.Files {
		bytes, ok := origin[name]
		if !ok || Hash(bytes) != hash {
			return fmt.Errorf("MiniMax replay source bytes changed")
		}
	}
	if len(source.Calls) != len(r.Calls) {
		return fmt.Errorf("MiniMax replay source call inventory differs")
	}
	for i, call := range r.Calls {
		prior := source.Calls[i]
		if call.ID != prior.ID || call.Rerun != prior.Rerun {
			return fmt.Errorf("MiniMax replay source case order differs")
		}
		if call.ID == "preflight-error" {
			continue
		}
		var original []Exchange
		if json.Unmarshal(origin[prior.CaseID+"/exchanges.json"], &original) != nil || len(original) != 1 {
			return fmt.Errorf("MiniMax replay source Attempt missing")
		}
		current, e := ReadFile(filepath.Join(dir, call.CaseID, "exchanges.json"), 2<<20)
		if e != nil {
			return e
		}
		var got []Exchange
		if json.Unmarshal(current, &got) != nil || len(got) != 1 || got[0].RequestBody != original[0].RequestBody || got[0].ResponseBody != original[0].ResponseBody || got[0].Status != original[0].Status || !reflect.DeepEqual(got[0].ResponseHeaders, original[0].ResponseHeaders) {
			return fmt.Errorf("MiniMax replay capture differs from source")
		}
	}
	return nil
}
