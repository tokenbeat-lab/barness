// minimaxverify checks a MiniMax bundle without credentials or vendor traffic.
package main

import (
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"path/filepath"
)

func main() {
	formal := flag.Bool("formal", false, "require current own M3 formal live evidence")
	publish := flag.Bool("publish", false, "exclusively publish own formal support projection after trusted origin")
	flag.Parse()
	if flag.NArg() != 1 || *publish && !*formal {
		fmt.Fprintln(os.Stderr, "usage: minimaxverify [-formal] <bundle-directory>")
		os.Exit(2)
	}
	r, e := minimaxlive.Verify(flag.Arg(0), *formal)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if *publish {
		projection, e := supportmatrix.NewMiniMaxReport(flag.Arg(0))
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		raw, e := supportmatrix.Encode(projection)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		f, e := os.OpenFile(filepath.Join(flag.Arg(0), "live-report.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		_, e = f.Write(raw)
		if e == nil {
			e = f.Sync()
		}
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	}
	fmt.Printf("PASS: %s %s mode=%s kind=%s calls=%d attempts=%d cumulativeCalls=%d\n", r.Identity.Route, r.Identity.Model, r.Mode, r.Kind, r.Budget.Calls, r.Budget.Attempts, r.Cumulative.Calls)
}
