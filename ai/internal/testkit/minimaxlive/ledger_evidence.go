package minimaxlive

import (
	"fmt"
	"path/filepath"
	"reflect"
)

type LedgerEvidence struct {
	Receipts []Receipt `json:"receipts"`
	// Retained measurement field; current single-process evidence forbids
	// predecessors. Historical captures remain under their fixed consumer.
	Predecessors map[string]map[string][]byte `json:"predecessors"`
}

func Summarize(receipts []Receipt, id Identity) Budget {
	var b Budget
	for _, r := range receipts {
		if r.Identity == id {
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
	for _, row := range ledger.Receipts {
		if row.Session != r.Session || row.Identity != r.Identity {
			return fmt.Errorf("process cannot borrow another session ledger")
		}
	}
	if len(ledger.Predecessors) != 0 {
		return fmt.Errorf("process cannot borrow historical predecessors")
	}
	previous := ""
	budget := Budget{}
	caseCount := map[string]int{}
	reservations := map[string]int{}
	attempts := map[string]int{}
	own := []Receipt{}
	for i, receipt := range ledger.Receipts {
		if receipt.ProcessGrantSHA256 != r.ProcessGrantSHA256 {
			return fmt.Errorf("MiniMax process receipt authority differs")
		}
		if receipt.Sequence != i+1 || receipt.Previous != previous || receipt.Hash != receiptHash(receipt) || receipt.Stage != r.Stage {
			return fmt.Errorf("MiniMax stage receipt chain changed")
		}
		previous = receipt.Hash
		if selected, err := Select(receipt.Identity.Route, receipt.Identity.Model, receipt.Identity.Purpose); err != nil || selected != receipt.Identity {
			return fmt.Errorf("MiniMax stage identity changed")
		}
		addBudget(&budget, receipt)
		caseKey := receipt.Case
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
		own = append(own, receipt)
	}
	if budget.Calls > 20 || budget.ReservedAttempts > 24 || budget.Attempts > budget.ReservedAttempts || budget.Attempts > 24 || budget.RequestedOutput > 40000 || budget.ReservedInput > 120000 || budget != r.Cumulative {
		return fmt.Errorf("MiniMax process budget differs or exceeded")
	}
	var expected []Receipt
	for _, c := range r.Calls {
		expected = append(expected, c.Reservation)
		expected = append(expected, c.Attempts...)
	}
	if !reflect.DeepEqual(own, expected) {
		return fmt.Errorf("MiniMax process hides or invents a reservation/attempt")
	}
	return nil
}
