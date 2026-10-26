// Package minimaxlive owns MiniMax's finite single-identity evidence contract.
// It does not read credentials or issue HTTP requests.
package minimaxlive

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

const M3 = "MiniMax-M3"
const M31 = "MiniMax-M3.1-Flash-Preview"
const ReportFile = "minimax-report.json"
const SealFile = "minimax-integrity.json"
const UnpricedM31Warning = "M3.1 price unknown; usage is raw provider evidence; Cost is synthetic mechanism output, not M Plan fees"

type Identity struct {
	Route    string `json:"route"`
	Model    string `json:"model"`
	Purpose  string `json:"purpose"`
	API      ai.API `json:"api"`
	Endpoint string `json:"endpoint"`
}

func Select(route, model, purpose string) (Identity, error) {
	id := Identity{Route: route, Model: model, Purpose: purpose}
	if model != M3 && model != M31 {
		return id, fmt.Errorf("unknown MiniMax model")
	}
	if purpose != "standard" && purpose != "mplan" {
		return id, fmt.Errorf("unknown MiniMax account purpose")
	}
	if model == M3 && purpose != "standard" || model == M31 && purpose != "mplan" {
		return id, fmt.Errorf("model requires its independently authorized account purpose")
	}
	switch route {
	case "chat":
		id.API, id.Endpoint = ai.APIOpenAICompletions, minimax.ChatEndpoint
	case "responses":
		id.API, id.Endpoint = ai.APIOpenAIResponses, minimax.ResponsesEndpoint
	case "messages":
		id.API, id.Endpoint = ai.APIAnthropicMessages, minimax.MessagesEndpoint
	default:
		return id, fmt.Errorf("unknown MiniMax protocol")
	}
	return id, nil
}

// BudgetSlot is finite across the international target replacement. Historical
// CN consumption shares this cap; exact Identity still governs proof and cases.
func (id Identity) BudgetSlot() string { return id.Route + "/" + id.Model + "/" + id.Purpose }

type Case struct {
	ID       string `json:"id"`
	Output   int    `json:"output"`
	Research string `json:"research"`
}

// Fixed for each of six identities. At most three initial transient REQUEST
// failures may repeat their scene once, plus one missing-tool/reasoning rerun.
// No started stream is replayed. All reruns consume the same durable ledger.
func Plan(id Identity) []Case {
	plan := []Case{
		{"full-stream", 1024, "default wire and terminal, raw actual model/ID and usage"},
		{"full-complete", 1024, "full Complete text"},
		{"simple-stream", 1024, "simple Stream default selection"},
		{"simple-complete", 1024, "simple Complete default selection"},
		{"thinking-on", 2048, "own explicit reasoning on/high or mandatory high"},
		{"thinking-control", 1024, "binary off or mandatory max"},
		{"tool-call", 4096, "own automatic tool and reasoning in the same assistant turn"},
		{"tool-result-native", 2048, "Host validated tool and original trusted reasoning/native state continuation"},
		{"truncate", 1, "length terminal retains actual known usage"},
		{"cache-first", 1024, "fixed repeated synthetic prefix, observed usage presence only"},
		{"cache-repeat", 1024, "identical prefix; no promised cache hit"},
		{"cancel", 2048, "cancel after visible content, unknown usage remains unknown"},
		{"preflight-error", 1, "invalid output request fails before any Attempt"},
	}
	if id.Model == M3 && id.Purpose == "standard" {
		plan = append(plan, Case{grokhistory.Scenario, 1024, "fixed synthetic xAI collision and cross-turn reuse history"})
	}
	return plan
}
func CaseID(id string, rerun int) string {
	if rerun > 0 {
		return fmt.Sprintf("MM07-%s-rerun-%d", id, rerun)
	}
	return "MM07-" + id
}

// ResearchPlan is the optional bounded pre-convergence frontier. It precedes
// full candidate acceptance and can never certify completeness or support.
func ResearchPlan() []Case {
	return []Case{{"research-full-stream", 1024, "bounded unknown default wire"}, {"research-tool-call", 4096, "own unknown reasoning plus automatic tool"}, {"research-tool-result-native", 2048, "own unknown native continuation"}}
}

// CandidateCatalog separates sourced M3 standard rates from unpriced M3.1
// synthetic mechanism fixtures. Neither catalog is a formal support claim.
func CandidateCatalog(id Identity) ai.Catalog {
	if id.Model == M3 {
		return minimax.PricedM3Catalog()
	}
	return minimax.CandidateCatalog()
}
