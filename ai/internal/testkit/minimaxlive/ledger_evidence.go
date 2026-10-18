package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

type LedgerEvidence struct {
	Receipts []Receipt `json:"receipts"`
	// Each earlier research process is preserved byte for byte, including its
	// old source applicability and failures. It never becomes current PASS.
	Predecessors map[string]map[string][]byte `json:"predecessors"`
}

func Summarize(receipts []Receipt, id Identity) Budget {
	var b Budget
	for _, r := range receipts {
		if r.Identity.BudgetSlot() == id.BudgetSlot() {
			addBudget(&b, r)
		}
	}
	return b
}
func VerifyLedger(dir string, r Report) error {
	raw, e := ReadFile(filepath.Join(dir, "minimax-ledger.json"), 16<<20)
	if e != nil {
		return e
	}
	var ledger LedgerEvidence
	if e = StrictJSON(raw, &ledger); e != nil {
		return e
	}
	previous := ""
	stageBudget := Budget{}
	identityBudget := map[string]Budget{}
	type exactCase struct {
		Identity Identity
		Case     string
	}
	caseCount := map[exactCase]int{}
	reservations := map[exactCase]int{}
	attempts := map[exactCase]int{}
	own := []Receipt{}
	predecessors := map[string]bool{}
	for i, receipt := range ledger.Receipts {
		if receipt.Sequence != i+1 || receipt.Previous != previous || receipt.Hash != receiptHash(receipt) || receipt.Stage != r.Stage {
			return fmt.Errorf("MiniMax stage receipt chain changed")
		}
		previous = receipt.Hash
		id := receipt.Identity
		if !accountingIdentity(receipt) {
			return fmt.Errorf("MiniMax stage identity changed")
		}
		key := id.BudgetSlot()
		b := identityBudget[key]
		addBudget(&b, receipt)
		identityBudget[key] = b
		addBudget(&stageBudget, receipt)
		caseKey := exactCase{id, receipt.Case}
		if receipt.Kind == "reserve" {
			if receipt.Output < 1 || receipt.Output > 4096 || receipt.Input < 1 || receipt.Input > 120000 {
				return fmt.Errorf("invalid recorded reservation")
			}
			caseCount[caseKey]++
			limit := 1
			if RerunnableCase(receipt.Case) {
				limit = 2
			}
			if caseCount[caseKey] > limit {
				return fmt.Errorf("MiniMax stage repeated a case")
			}
			reservations[caseKey]++
		} else if receipt.Kind == "attempt" {
			attempts[caseKey]++
			if attempts[caseKey] > reservations[caseKey] {
				return fmt.Errorf("MiniMax stage contains unreserved attempt")
			}
		} else {
			return fmt.Errorf("unknown stage receipt")
		}
		if id == r.Identity {
			if receipt.Session == r.Session {
				own = append(own, receipt)
			} else {
				predecessors[receipt.Session] = true
			}
		}
	}
	for _, b := range identityBudget {
		if b.Calls > 20 || b.ReservedAttempts > 24 || b.Attempts > 24 || b.RequestedOutput > 40000 || b.ReservedInput > 120000 {
			return fmt.Errorf("MiniMax cumulative identity budget exceeded")
		}
	}
	maxCalls, maxAttempts, maxIDs := 120, 144, 6
	if r.Stage == "formal" {
		maxCalls, maxAttempts, maxIDs = 60, 72, 3
	}
	if stageBudget.Calls > maxCalls || stageBudget.ReservedAttempts > maxAttempts || stageBudget.Attempts > maxAttempts || len(identityBudget) > maxIDs || !reflect.DeepEqual(identityBudget[r.Identity.BudgetSlot()], r.Cumulative) {
		return fmt.Errorf("MiniMax cumulative stage budget differs")
	}
	var expected []Receipt
	for _, c := range r.Calls {
		expected = append(expected, c.Reservation)
		expected = append(expected, c.Attempts...)
	}
	if !reflect.DeepEqual(own, expected) {
		return fmt.Errorf("MiniMax process hides or invents a reservation/attempt")
	}
	if len(predecessors) != len(ledger.Predecessors) {
		return fmt.Errorf("MiniMax predecessor capture missing or invented")
	}
	for session := range predecessors {
		captures, ok := ledger.Predecessors[session]
		if !ok || session == "" || filepath.Base(session) != session {
			return fmt.Errorf("invalid predecessor session")
		}
		var prior Report
		if StrictJSON(captures[ReportFile], &prior) != nil || prior.Session != session || prior.Identity != r.Identity || prior.Kind != "minimax-research-only" || prior.Stage != "candidate" || prior.Complete {
			return fmt.Errorf("MiniMax predecessor is not bounded research")
		}
		var seal Seal
		if StrictJSON(captures[SealFile], &seal) != nil {
			return fmt.Errorf("MiniMax predecessor seal absent")
		}
		for name, hash := range seal.Files {
			bytes, ok := captures[name]
			if !ok || Hash(bytes) != hash {
				return fmt.Errorf("MiniMax predecessor capture changed")
			}
		}
		if len(captures) != len(seal.Files)+1 {
			return fmt.Errorf("MiniMax predecessor invented files")
		}
		priorReceipts := []Receipt{}
		for _, c := range prior.Calls {
			priorReceipts = append(priorReceipts, c.Reservation)
			priorReceipts = append(priorReceipts, c.Attempts...)
		}
		actual := []Receipt{}
		for _, receipt := range ledger.Receipts {
			if receipt.Session == session && receipt.Identity == r.Identity {
				actual = append(actual, receipt)
			}
		}
		if !reflect.DeepEqual(actual, priorReceipts) {
			return fmt.Errorf("MiniMax research hides attempts")
		}
		if len(prior.Expected) != 3 || !reflect.DeepEqual(prior.Expected, ResearchPlan()) || len(prior.Calls) > 3 {
			return fmt.Errorf("MiniMax research inventory changed")
		}
		for _, c := range prior.Calls {
			if !strings.HasPrefix(c.ID, "research-") {
				return fmt.Errorf("MiniMax predecessor case invented")
			}
			var exchanges []Exchange
			if json.Unmarshal(captures[c.CaseID+"/exchanges.json"], &exchanges) != nil || len(exchanges) != len(c.Attempts) {
				return fmt.Errorf("MiniMax predecessor attempt capture missing")
			}
		}
	}
	return nil
}
