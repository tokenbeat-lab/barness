// Command glmformal derives an audited support view from a trusted immutable
// formal GLM capture. It has no credential or vendor-network path.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
)

func main() {
	source := flag.Bool("source", false, "print complete actual GLM execution-source files without credentials")
	route := flag.String("route", "", "chat, responses or messages")
	destination := flag.String("out", "", "new support view directory")
	flag.Parse()
	if *source {
		if *route != "" || *destination != "" || flag.NArg() != 0 {
			fmt.Fprintln(os.Stderr, "source is an exclusive read-only mode")
			os.Exit(2)
		}
		files, err := supportmatrix.GLMExecutionSource()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if json.NewEncoder(os.Stdout).Encode(map[string]any{"files": files}) != nil {
			os.Exit(1)
		}
		return
	}
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
