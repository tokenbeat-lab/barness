package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Serving facts are independent literals. The conflicting canonical price is
// deliberately present: it cannot supply a serving offering's rate.
const maintenanceSourceFixture = `{"providers":{"openai":{"id":"openai","models":{
 "fixture-chat":{"id":"fixture-chat","type":"chat","name":"Serving name","reasoning":false,"modalities":{"input":["text"],"output":["text"]},"limit":{"context":4096,"output":128},"cost":{"input":1,"output":2,"cache_read":0,"cache_write":0},"canonical_model_id":"openai/canonical"},
 "missing-price":{"id":"missing-price","type":"chat","name":"Missing","reasoning":false,"modalities":{"input":["text"],"output":["text"]},"limit":{"context":4096,"output":128},"cost":{"input":1,"output":2}},
 "fraction":{"id":"fraction","type":"chat","name":"Fraction","reasoning":false,"modalities":{"input":["text"],"output":["text"]},"limit":{"context":4096.5,"output":128},"cost":{"input":1,"output":2,"cache_read":0,"cache_write":0}},
 "audio":{"id":"audio","name":"Audio","reasoning":false,"modalities":{"input":["audio"],"output":["text"]},"limit":{"context":4096,"output":128},"cost":{"input":1,"output":2,"cache_read":0,"cache_write":0}},
 "embedding":{"id":"embedding","type":"embedding","name":"Embedding"}
}}},"models":{"openai/canonical":{"id":"openai/canonical","name":"Wrong canonical name","cost":{"input":999,"output":999,"cache_read":999,"cache_write":999}}}}`

func TestCatalogMaintenanceCLISourceMapping(t *testing.T) {
	ev := run.Case(t, "CATALOG-T02-cli-serving-mapping")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "models-dev.json")
	if err := os.WriteFile(source, []byte(maintenanceSourceFixture), 0600); err != nil {
		t.Fatal(err)
	}
	ev.Fixture("models-dev.json", []byte(maintenanceSourceFixture))
	store := filepath.Join(dir, "store")
	args := []string{"add", "--store", store, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "fixed source acceptance", "--confirm"}
	out, exit := invokeCatalogCLI(t, ev, binary, args...)
	var result struct {
		Status    string `json:"status"`
		Candidate struct {
			State struct {
				Providers []struct {
					ID      string `json:"id"`
					Entries []struct {
						Definition json.RawMessage                                                         `json:"definition"`
						Source     struct{ Provider, ModelID, CanonicalID, Digest, URL, FetchedAt string } `json:"source"`
						Fields     map[string]struct {
							Ownership string `json:"ownership"`
						} `json:"fields"`
					} `json:"entries"`
				} `json:"providers"`
			} `json:"state"`
			Inputs []struct {
				Bytes  []byte `json:"bytes"`
				Digest string `json:"digest"`
			} `json:"inputs"`
		} `json:"candidate"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("source can add a data-only candidate", exit == 0 && result.Status == "awaiting_check", "exit=%d output=%s", exit, out) {
		return
	}
	found := false
	for _, p := range result.Candidate.State.Providers {
		for _, e := range p.Entries {
			var m struct {
				ID, Name string
				Cost     struct{ Input, Output, CacheRead, CacheWrite float64 }
			}
			_ = json.Unmarshal(e.Definition, &m)
			if m.ID != "fixture-chat" {
				continue
			}
			found = true
			ev.Check("offering supplies wire ID and rates", p.ID == "openai" && m.Name == "Serving name" && m.Cost.Input == 1 && m.Cost.Output == 2 && m.Cost.CacheRead == 0 && m.Cost.CacheWrite == 0, "model=%+v", m)
			ev.Check("explicit zero rate retains source ownership", e.Fields["/cost/cacheRead"].Ownership == "source", "fields=%+v", e.Fields)
			ev.Check("association preserves provider/model/canonical and input identity", e.Source.Provider == "openai" && e.Source.ModelID == "fixture-chat" && e.Source.CanonicalID == "openai/canonical" && e.Source.Digest != "" && e.Source.FetchedAt == "2026-11-01T00:00:00Z", "source=%+v", e.Source)
		}
	}
	ev.Check("new identity exists exactly as serving entry", found, "fixture entry missing")
	ev.Check("original source bytes are retained", len(result.Candidate.Inputs) == 1 && string(result.Candidate.Inputs[0].Bytes) == maintenanceSourceFixture, "inputs=%d", len(result.Candidate.Inputs))
	for _, model := range []string{"missing-price", "fraction", "audio", "embedding"} {
		bad := append([]string(nil), args...)
		for i := range bad {
			if i > 0 && bad[i-1] == "--model" {
				bad[i] = model
			}
		}
		bad[2] = filepath.Join(dir, model+"-absent")
		out, exit := invokeCatalogCLI(t, ev, binary, bad...)
		ev.Check("incomplete or unsupported source does not freeze", exit != 0, "%s exit=%d output=%s", model, exit, out)
		_, err := os.Stat(bad[2])
		ev.Check("rejected source creates no store", os.IsNotExist(err), "%s err=%v", model, err)
	}
}
