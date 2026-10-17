// minimaxverify checks a MiniMax bundle without credentials or vendor traffic.
package main

import (
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
)

func main() {
	formal := flag.Bool("formal", false, "require current own M3 formal live evidence")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: minimaxverify [-formal] <bundle-directory>")
		os.Exit(2)
	}
	r, e := minimaxlive.Verify(flag.Arg(0), *formal)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Printf("PASS: %s %s mode=%s kind=%s calls=%d attempts=%d cumulativeCalls=%d\n", r.Identity.Route, r.Identity.Model, r.Mode, r.Kind, r.Budget.Calls, r.Budget.Attempts, r.Cumulative.Calls)
}
