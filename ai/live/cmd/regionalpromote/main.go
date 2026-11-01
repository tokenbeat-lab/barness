// regionalpromote emits only controlled, nonproduction qualification proposals.
// Catalog activation is owned by the maintenance candidate/check/apply flow.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	proposal := flag.String("proposal", "", "new exclusive nonproduction controlled proposal")
	kimi := flag.Bool("kimi", false, "consume exactly six fixed Kimi nine-plus-one compositions")
	controlled := flag.Bool("controlled", false, "required: nonproduction controlled proposal only")
	flag.Parse()
	if !*controlled || flag.NArg() != 6 || *proposal == "" {
		fmt.Fprintln(os.Stderr, "controlled proposal requires -controlled -proposal and six inputs (Kimi: fixed nine-plus-one compositions)")
		os.Exit(2)
	}
	var err error
	if *kimi {
		err = supportmatrix.WriteKimiControlledProposal(flag.Args(), *proposal)
	} else {
		err = supportmatrix.WriteRegionalControlledProposal(flag.Args(), *proposal)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
