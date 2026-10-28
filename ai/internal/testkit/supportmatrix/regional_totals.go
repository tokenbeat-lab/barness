package supportmatrix

import (
	"encoding/json"
	"os"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
)

// Native captures independently determine reservations and usage. A summary
// cannot conceal unrecorded attempts or lower the budget charged before send.
func verifyRegionalTotals(root *os.Root, r Report) error {
	read := func(name string, v any) error {
		raw, err := regionalRead(root, name)
		if err != nil {
			return err
		}
		return regionalDecode(raw, v)
	}
	var cat ai.Catalog
	rawCatalog, err := regionalRead(root, "regional-catalog.json")
	if err != nil || regionalDecode(rawCatalog, &cat) != nil {
		return refuse("original catalog snapshot unavailable")
	}
	hash, hashErr := cat.Hash()
	if hashErr != nil || hash != r.CatalogHash || cat.Version != r.CatalogVersion {
		return refuse("original catalog snapshot differs")
	}
	var captured ai.Catalog
	var facts any
	if read("regional-catalog.json", &captured) != nil || !reflect.DeepEqual(captured, cat) || read("regional-candidate-facts.json", &facts) != nil || !regionalSameFacts(facts, r.Combo) {
		return refuse("regional catalog or independent source facts differ")
	}
	input, output, reserved, requested := 0, 0, 0, 0
	var usd float64
	for _, s := range r.Scenarios {
		var result regionalResult
		var xs []regionalExchange
		if read(s.CaseID+"/result.json", &result) != nil || read(s.CaseID+"/exchanges.json", &xs) != nil || len(xs) != 1 {
			return refuse("regional actual totals lack native captures")
		}
		var wire map[string]json.RawMessage
		if json.Unmarshal([]byte(xs[0].RequestBody), &wire) != nil {
			return refuse("regional encoded budget request invalid")
		}
		key := "max_tokens"
		if r.API == "openai-responses" {
			key = "max_output_tokens"
		}
		if len(wire["max_completion_tokens"]) > 0 {
			key = "max_completion_tokens"
		}
		var limit int
		model, _ := cat.Lookup(ai.ProviderID(r.Provider), ai.API(r.API), s.Model)
		if json.Unmarshal(wire[key], &limit) != nil || limit != min(4096, model.MaxTokens) {
			return refuse("regional own encoded output reservation differs")
		}
		requested += limit
		wireInput := max(2048, len(xs[0].RequestBody)+1024)
		reserved += wireInput
		amount, e := RegionalReservedUSD(r.Combo, s.Model, wireInput, limit)
		if e != nil {
			return e
		}
		usd += amount
		message, err := result.Message.assistant()
		if err != nil || len(result.Metadata.Attempts) != 1 {
			return refuse("regional raw totals lack Result attempt")
		}
		u, err := verifyRegionalResponse(root, r, s, xs[0], message, result.Metadata.Attempts[0])
		if err != nil {
			return err
		}
		if RegionalIsKimi(r.Combo) && (u.Input+u.CacheRead+u.CacheWrite > int64(wireInput) || u.Output > int64(limit)) {
			return refuse("Kimi actual per-call usage exceeds conservative reservation")
		}
		input += int(u.Input + u.CacheRead + u.CacheWrite)
		output += int(u.Output)
	}
	b := r.Budget
	if b.ReservedInputTokens != reserved || b.RequestedOutputTokens != requested || b.InputTokensUsed != input || b.OutputTokensUsed != output || input < 0 || output < 0 || output > requested || input > reserved {
		return refuse("regional captured usage or encoded reservations differ")
	}
	if RegionalIsKimi(r.Combo) && (usd != r.Regional.ReservedUSD || usd > r.Regional.MaxUSD) {
		return refuse("Kimi conservative USD reservation differs")
	}
	return nil
}
