package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// inspect is the keyless native-contract seam used by controlled process E2E.
// It writes stdout only, never a support projection into the original bundle.
func inspect(dir, mode, stage string) int {
	if mode != "live" && mode != "controlled" || stage != "candidate" && stage != "formal" || mode == "controlled" && stage == "formal" {
		return 2
	}
	var value any
	var err error
	if supportmatrix.IsMiniMaxThinkingComposition(dir) {
		var q supportmatrix.MiniMaxThinkingQualification
		q, err = supportmatrix.VerifyMiniMaxThinkingComposition(dir, stage == "formal")
		if err == nil && q.Mode != mode {
			err = fmt.Errorf("MiniMax composition mode differs")
		}
		value = q
	} else if supportmatrix.IsKimiComposition(dir) {
		var q supportmatrix.KimiQualification
		q, err = supportmatrix.VerifyKimiComposition(dir, stage == "formal")
		if err == nil && q.Mode != mode {
			err = fmt.Errorf("composition mode differs from requested mode")
		}
		value = supportmatrix.KimiInspection{KimiQualification: q, QualificationAt: time.Now().UTC()}
	} else if mode == "live" && stage == "formal" {
		var root string
		root, err = os.Getwd()
		if err == nil {
			value, err = release.QualifyNativeLive(root, dir)
		}
	} else if _, e := os.Stat(filepath.Join(dir, minimaxlive.ReportFile)); e == nil {
		var r minimaxlive.Report
		r, err = minimaxlive.Load(dir)
		if err == nil && r.Kind == minimaxlive.ThinkingKind {
			r, err = minimaxlive.VerifyThinkingSupplement(dir, stage == "formal")
			if err == nil {
				var billing minimaxlive.ThinkingBilling
				billing, err = minimaxlive.AssessThinkingBilling(dir, r)
				value = minimaxlive.ThinkingInspection{Report: r, Billing: billing}
			}
		} else if err == nil {
			r, err = minimaxlive.Verify(dir, stage == "formal")
			value = r
		}
		if err == nil && r.Mode != mode {
			err = fmt.Errorf("native mode differs from selected mode")
		}
	} else if mode == "controlled" {
		value, err = supportmatrix.VerifyRegionalControlledCapture(dir)
	} else if stage == "candidate" {
		value, err = supportmatrix.VerifyRegionalQualification(dir)
	} else {
		value, err = supportmatrix.NewRegionalReport(dir)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if json.NewEncoder(os.Stdout).Encode(value) != nil {
		return 1
	}
	return 0
}
