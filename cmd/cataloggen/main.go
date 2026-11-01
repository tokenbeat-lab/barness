// cataloggen deterministically generates ai's embedded default from the
// reviewed Provider maintenance source. It performs no network access.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func main() {
	out := flag.String("out", "", "output file (stdout when omitted)")
	flag.Parse()
	state, err := maintenance.BuiltinState()
	if err == nil {
		var raw []byte
		raw, err = maintenance.Generate(state)
		if err == nil {
			if *out == "" {
				_, err = os.Stdout.Write(raw)
			} else {
				err = os.WriteFile(*out, raw, 0644)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
