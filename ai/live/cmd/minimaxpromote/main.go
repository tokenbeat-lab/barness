// minimaxpromote performs a data-only catalog activation after own acceptance.
package main

import (
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
)

func main() {
	receipt := flag.String("receipt", "", "new exclusive candidate-lineage receipt path")
	flag.Parse()
	if flag.NArg() != 3 || *receipt == "" {
		fmt.Fprintln(os.Stderr, "usage: minimaxpromote -receipt <new-receipt.json> <chat-candidate> <responses-candidate> <messages-candidate>")
		os.Exit(2)
	}
	if e := minimaxlive.Activate(flag.Args(), *receipt); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
