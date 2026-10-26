// processgrant prints machine scope or checks/burns an explicitly supplied
// approval. It never reads credentials or sends HTTP.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	source := flag.String("source", "", "route whose current prebuild machine facts to print")
	template := flag.String("template", "", "route whose current machine facts to print")
	check := flag.String("check", "", "explicit approval file to check")
	route := flag.String("route", "", "exact selected route")
	mode := flag.String("mode", "live", "live or controlled")
	stage := flag.String("stage", "formal", "candidate or formal")
	binary := flag.String("binary", "", "frozen producer binary path")
	claim := flag.Bool("claim", false, "consume this new approval once")
	flag.Parse()
	fail := func(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
	if *source != "" {
		if *template != "" || *check != "" || *claim || *binary != "" || *route != "" || flag.NArg() != 0 {
			fail(fmt.Errorf("source is an exclusive read-only mode"))
		}
		scope, e := supportmatrix.ProcessScope(*source, *mode, *stage, "")
		if e != nil {
			fail(e)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"sourceSha256": scope.SourceSHA256, "scope": scope, "buildFlag": "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource=" + scope.SourceSHA256})
		return
	}
	if (*template == "") == (*check == "") || flag.NArg() != 0 {
		fail(fmt.Errorf("select exactly one template or check"))
	}
	if *template != "" {
		*route = *template
	}
	hash, e := livegrant.FileHash(*binary)
	if e != nil {
		fail(e)
	}
	scope, e := supportmatrix.ProcessScope(*route, *mode, *stage, hash)
	if e != nil {
		fail(e)
	}
	if *template != "" {
		if *claim {
			fail(fmt.Errorf("template is not an approval"))
		}
		_ = json.NewEncoder(os.Stdout).Encode(livegrant.Grant{Schema: 1, Scope: scope})
		return
	}
	raw, e := livegrant.ReadFile(*check)
	if e != nil {
		fail(e)
	}
	g, e := livegrant.Decode(raw)
	if e != nil {
		fail(e)
	}
	if e = g.Check(scope, time.Now()); e != nil {
		fail(e)
	}
	if *claim {
		if *mode != "controlled" {
			fail(fmt.Errorf("CLI claim is only for controlled admission tests; vendor claims belong to the selected producer"))
		}
		r, e := livegrant.Claim(g, raw, "keyless-controlled-admission-"+g.ID, time.Now())
		if e != nil {
			fail(e)
		}
		_ = json.NewEncoder(os.Stdout).Encode(r)
	} else {
		fmt.Println("scope valid; no claim or vendor authorization acquired")
	}
}
