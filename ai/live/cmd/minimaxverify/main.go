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
	recoveryFacts := flag.Bool("recovery-source", false, "read-only fixed recovery3 source and closed failed second cohort facts")
	recoveryPreflight := flag.String("recovery-preflight", "", "read-only recovery3 authority and unclaimed route guard")
	recovery4Facts := flag.Bool("recovery4-source", false, "read-only fixed recovery4 source and closed third cohort facts")
	recovery4Preflight := flag.String("recovery4-preflight", "", "read-only recovery4 authority and unclaimed route guard")
	recovery5Facts := flag.Bool("recovery5-source", false, "read-only fixed recovery5 source and closed fourth cohort facts")
	recovery5Preflight := flag.String("recovery5-preflight", "", "read-only recovery5 authority and unclaimed route guard")
	recovery6Facts := flag.Bool("recovery6-source", false, "read-only fixed recovery6 source and closed fifth cohort facts")
	recovery6Preflight := flag.String("recovery6-preflight", "", "read-only recovery6 authority and unclaimed route guard")
	flag.Parse()
	if *recovery6Facts || *recovery6Preflight != "" {
		if flag.NArg() != 0 || *formal || *publish || *facts || *preflight != "" || *recoveryFacts || *recoveryPreflight != "" || *recovery4Facts || *recovery4Preflight != "" || *recovery5Facts || *recovery5Preflight != "" || *recovery6Facts && *recovery6Preflight != "" {
			fmt.Fprintln(os.Stderr, "recovery6-source/preflight cannot verify or publish")
			os.Exit(2)
		}
		if e := recovery6(*recovery6Facts, *recovery6Preflight); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}
	if *recovery5Facts || *recovery5Preflight != "" {
		if flag.NArg() != 0 || *formal || *publish || *facts || *preflight != "" || *recoveryFacts || *recoveryPreflight != "" || *recovery4Facts || *recovery4Preflight != "" || *recovery5Facts && *recovery5Preflight != "" {
			fmt.Fprintln(os.Stderr, "recovery5-source/preflight cannot verify or publish")
			os.Exit(2)
		}
		if e := recovery5(*recovery5Facts, *recovery5Preflight); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}
	if *recovery4Facts || *recovery4Preflight != "" {
		if flag.NArg() != 0 || *formal || *publish || *facts || *preflight != "" || *recoveryFacts || *recoveryPreflight != "" || *recovery4Facts && *recovery4Preflight != "" {
			fmt.Fprintln(os.Stderr, "recovery4-source/preflight cannot verify or publish")
			os.Exit(2)
		}
		if e := recovery4(*recovery4Facts, *recovery4Preflight); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}
	if *recoveryFacts || *recoveryPreflight != "" {
		if flag.NArg() != 0 || *formal || *publish || *facts || *preflight != "" || *recoveryFacts && *recoveryPreflight != "" {
			fmt.Fprintln(os.Stderr, "recovery-source/preflight cannot verify or publish")
			os.Exit(2)
		}
		if e := recovery(*recoveryFacts, *recoveryPreflight); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}
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

func recovery(facts bool, route string) error {
	if facts {
		f, e := minimaxlive.RecoveryFacts()
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(f)
	}
	g, pin, e := minimaxlive.RecoveryPreflight(route)
	if e != nil {
		return e
	}
	f, e := minimaxlive.RecoveryFacts()
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"route": route, "grant": g, "grantSha256": pin, "facts": f})
}

func recovery4(facts bool, route string) error {
	if facts {
		f, e := minimaxlive.Recovery4Facts()
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(f)
	}
	g, pin, e := minimaxlive.Recovery4Preflight(route)
	if e != nil {
		return e
	}
	f, e := minimaxlive.Recovery4Facts()
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"route": route, "grant": g, "grantSha256": pin, "facts": f})
}

func recovery5(facts bool, route string) error {
	if facts {
		f, e := minimaxlive.Recovery5Facts()
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(f)
	}
	g, pin, e := minimaxlive.Recovery5Preflight(route)
	if e != nil {
		return e
	}
	f, e := minimaxlive.Recovery5Facts()
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"route": route, "grant": g, "grantSha256": pin, "facts": f})
}

func recovery6(facts bool, route string) error {
	if facts {
		f, e := minimaxlive.Recovery6Facts()
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(f)
	}
	g, pin, e := minimaxlive.Recovery6Preflight(route)
	if e != nil {
		return e
	}
	f, e := minimaxlive.Recovery6Facts()
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"route": route, "grant": g, "grantSha256": pin, "facts": f})
}
