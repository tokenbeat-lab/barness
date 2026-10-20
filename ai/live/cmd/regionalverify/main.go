// regionalverify is keyless: it prints the current build-source binding or
// consumes a native own qualification bundle. It never invokes a provider.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	source := flag.Bool("source", false, "print complete current execution source and digest")
	grant := flag.String("grant-template", "", "print an intentionally incomplete own authorization template")
	qualification := flag.String("qualify", "", "verify own candidate/formal live qualification bundle")
	formal := flag.String("formal", "", "verify own formally promoted live support bundle")
	controlled := flag.String("controlled", "", "verify controlled native capture; never live qualification")
	stage := flag.String("stage", "candidate", "grant stage: candidate or formal")
	flag.Parse()
	selected := 0
	if *source {
		selected++
	}
	for _, v := range []string{*grant, *qualification, *formal, *controlled} {
		if v != "" {
			selected++
		}
	}
	if selected != 1 || flag.NArg() != 0 || *stage != "candidate" && *grant == "" {
		fmt.Fprintln(os.Stderr, "select exactly one consumer mode")
		os.Exit(1)
	}
	var value any
	var err error
	switch {
	case *source:
		var files map[string]string
		files, err = supportmatrix.RegionalExecutionSource()
		value = map[string]any{"sourceSha256": supportmatrix.RegionalSourceHash(files), "files": files}
	case *grant != "":

		var template supportmatrix.RegionalGrant
		template, err = supportmatrix.RegionalGrantSnapshot(*grant, *stage)
		template.ID = "<new human authorization required>"
		template.AccountAlias = "<actual own account alias required>"
		template.RegionAlias = "<actual own region alias required>"
		value = template
	case *controlled != "":
		value, err = supportmatrix.VerifyRegionalControlledCapture(*controlled)
	case *qualification != "":
		value, err = supportmatrix.VerifyRegionalQualification(*qualification)
	case *formal != "":
		value, err = supportmatrix.NewRegionalReport(*formal)
	default:
		err = fmt.Errorf("select exactly one of -source, -grant-template, -qualify, -formal")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err = enc.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
