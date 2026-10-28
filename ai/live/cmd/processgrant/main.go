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
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
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
	profile := flag.String("profile", "", "fixed MiniMax supplement profile")
	purpose := flag.String("purpose", "", "fixed regional supplement purpose")
	baseline := flag.String("baseline-manifest-sha256", "", "original fixed supplement baseline manifest hash")
	claim := flag.Bool("claim", false, "consume this new approval once")
	flag.Parse()
	fail := func(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
	_, regionalTemplate := supportmatrix.RegionalRouteOf(*template)
	if *profile != "" {
		if *profile != minimaxlive.ThinkingProfile || *template != "minimax-messages" || *stage != "candidate" || *purpose != "" || *check != "" || *source != "" || *claim || *route != "" || flag.NArg() != 0 {
			fail(fmt.Errorf("fixed MiniMax profile requires exclusive candidate template"))
		}
		hash, e := livegrant.FileHash(*binary)
		if e != nil {
			fail(e)
		}
		id, _ := minimaxlive.Select("messages", minimaxlive.M3, "standard")
		scope, e := minimaxlive.ThinkingOnScope(id, *mode, hash, *baseline)
		if e != nil {
			fail(e)
		}
		_ = json.NewEncoder(os.Stdout).Encode(livegrant.Grant{Schema: 1, Scope: scope})
		return
	}

	if (*purpose != "" || *baseline != "") && *profile == "" && (!regionalTemplate || *template == "") {
		fail(fmt.Errorf("supplement scope flags require a regional template"))
	}
	if *source == "regional" || *source == "glm" {
		if *template != "" || *check != "" || *claim || *binary != "" || *route != "" || flag.NArg() != 0 {
			fail(fmt.Errorf("source is exclusive"))
		}
		files, err := supportmatrix.RegionalExecutionSource()
		if *source == "glm" {
			files, err = supportmatrix.GLMExecutionSource()
		}
		if err != nil {
			fail(err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"files": files, "sourceSha256": supportmatrix.RegionalSourceHash(files)})
		return
	}
	if _, regional := supportmatrix.RegionalRouteOf(*template); regional && *template != "" {
		if *source != "" || *check != "" || *claim || *route != "" || flag.NArg() != 0 {
			fail(fmt.Errorf("template is exclusive"))
		}
		g, err := supportmatrix.RegionalGrantSnapshot(*template, *stage)
		if *purpose != "" {
			if *purpose != supportmatrix.KimiHistoryPurpose || *stage != "candidate" {
				fail(fmt.Errorf("unknown or noncandidate supplement"))
			}
			g, err = supportmatrix.KimiHistoryGrantSnapshot(*template, *baseline)
		} else if *baseline != "" {
			fail(fmt.Errorf("baseline pin requires supplement purpose"))
		}
		if err != nil {
			fail(err)
		}
		g.ID = "<new human authorization required>"
		g.AccountAlias = "<actual own account alias required>"
		g.RegionAlias = "<actual own region alias required>"
		_ = json.NewEncoder(os.Stdout).Encode(g)
		return
	}
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
