// Command supportmatrix merges live smoke reports into barness-ai's support
// matrix (ai/live/support-matrix.json). Each report is one live process's
// result for one Provider × API combination and changes only that
// combination's row. Reports are applied oldest first; if any is refused
// the matrix file is left unchanged.
//
//	go run ./ai/live/cmd/supportmatrix [-matrix ai/live/support-matrix.json] report.json...
//
// A report path may also be an evidence bundle directory holding
// live-report.json.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	matrixPath := flag.String("matrix", "ai/live/support-matrix.json", "support matrix file to update")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: supportmatrix [-matrix path] report.json|bundle-dir...")
		os.Exit(2)
	}
	if err := run(*matrixPath, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "supportmatrix:", err)
		os.Exit(1)
	}
}

func run(matrixPath string, paths []string) error {
	m, err := supportmatrix.LoadMatrix(matrixPath)
	if err != nil {
		return err
	}
	reports := make([]supportmatrix.Report, 0, len(paths))
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			p = filepath.Join(p, supportmatrix.ReportFile)
		}
		r, err := supportmatrix.LoadReport(p)
		if err != nil {
			return err
		}
		reports = append(reports, r)
	}
	slices.SortStableFunc(reports, func(a, b supportmatrix.Report) int { return a.FinishedAt.Compare(b.FinishedAt) })
	for _, r := range reports {
		if m, err = supportmatrix.Merge(m, r); err != nil {
			return fmt.Errorf("%s report finished %s: %w", r.Combo, r.FinishedAt.Format("2006-01-02T15:04:05Z07:00"), err)
		}
		fmt.Printf("%s: merged %d scenarios\n", r.Combo, len(r.Scenarios))
	}
	return supportmatrix.SaveMatrix(matrixPath, m)
}
