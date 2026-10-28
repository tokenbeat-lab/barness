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
	kimi := flag.Bool("kimi", false, "consume exactly six fixed Kimi nine-plus-one compositions")
	controlled := flag.Bool("controlled", false, "nonproduction controlled proposal only")
	flag.Parse()
	if flag.NArg() != 6 || *controlled && (*proposal == "" || *receipt != "") || !*controlled && (*receipt == "" || *proposal != "") {
		fmt.Fprintln(os.Stderr, "promotion requires six live inputs and -receipt (Kimi: fixed nine-plus-one compositions); controlled captures cannot activate")
		os.Exit(2)
	}
	var err error
	if *kimi {
		if *controlled {
			err = supportmatrix.WriteKimiControlledProposal(flag.Args(), *proposal)
		} else {
			err = supportmatrix.ActivateKimiCatalog(flag.Args(), *receipt)
		}
	} else if *controlled {
		err = supportmatrix.WriteRegionalControlledProposal(flag.Args(), *proposal)
	} else {
		err = supportmatrix.ActivateRegionalCatalog(flag.Args(), *receipt)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
