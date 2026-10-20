// regionalpromote consumes six own candidates before a data-only activation.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	receipt := flag.String("receipt", "", "new exclusive live candidate-lineage receipt under repository .evidence")
	proposal := flag.String("proposal", "", "new exclusive nonproduction controlled proposal")
	controlled := flag.Bool("controlled", false, "nonproduction controlled proposal only")
	flag.Parse()
	if flag.NArg() != 6 || *controlled && (*proposal == "" || *receipt != "") || !*controlled && (*receipt == "" || *proposal != "") {
		fmt.Fprintln(os.Stderr, "regional promotion requires six live candidates and -receipt; controlled captures cannot activate")
		os.Exit(2)
	}
	var err error
	if *controlled {
		err = supportmatrix.WriteRegionalControlledProposal(flag.Args(), *proposal)
	} else {
		err = supportmatrix.ActivateRegionalCatalog(flag.Args(), *receipt)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
