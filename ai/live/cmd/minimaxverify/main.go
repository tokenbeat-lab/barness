// minimaxverify checks a MiniMax bundle without credentials or vendor traffic.
package main

import (
	"encoding/json"
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
	facts := flag.Bool("renewal-source", false, "read-only current final-renewal source and complete closed history facts")
	preflight := flag.String("renewal-preflight", "", "read-only current final-renewal authority and unclaimed route guard")
	flag.Parse()
	if *preflight != "" {
		if flag.NArg() != 0 || *formal || *publish || *facts {
			fmt.Fprintln(os.Stderr, "renewal-preflight cannot verify or publish")
			os.Exit(2)
		}
		g, pin, e := minimaxlive.RenewalPreflight(*preflight)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		f, e := minimaxlive.RenewalFacts()
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		if e = json.NewEncoder(os.Stdout).Encode(map[string]any{"route": *preflight, "grant": g, "grantSha256": pin, "facts": f}); e != nil {
			os.Exit(1)
		}
		return
	}
	if *facts {
		if flag.NArg() != 0 || *formal || *publish {
			fmt.Fprintln(os.Stderr, "renewal-source cannot verify or publish")
			os.Exit(2)
		}
		f, e := minimaxlive.RenewalFacts()
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		if e = json.NewEncoder(os.Stdout).Encode(f); e != nil {
			os.Exit(1)
		}
		return
	}
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
