// Command glmformal derives an audited support view from a trusted immutable
// formal GLM capture. It has no credential or vendor-network path.
package main

import (
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
)

func main() {
	route := flag.String("route", "", "chat, responses or messages")
	destination := flag.String("out", "", "new support view directory")
	flag.Parse()
	if *destination == "" {
		fmt.Fprintln(os.Stderr, "-out is required")
		os.Exit(2)
	}
	r, e := supportmatrix.CreateGLMSupportView(*route, *destination)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Printf("%s reported-known input-side=%d output=%d unknown-attempts=%d\n", r.Combo, r.Budget.InputTokensUsed, r.Budget.OutputTokensUsed, r.GLM.UnknownUsageAttempts)
}
