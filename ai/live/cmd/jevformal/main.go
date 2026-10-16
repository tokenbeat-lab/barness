// Command jevformal seals/validates fixed Jev 1.14 formal protocol and independent
// effect packages. It makes no provider calls. Merge separately with supportmatrix.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/jevformal"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	protocol := flag.String("protocol", "", "new formal protocol package")
	effects := flag.String("effects", "", "new independent formal effects package")
	out := flag.String("out", "", "new self-contained support package directory")
	verify := flag.String("verify", "", "verify an existing support package without network")
	live := flag.Bool("require-live", false, "require both packages to be complete real evidence")
	flag.Parse()
	if err := run(*protocol, *effects, *out, *verify, *live); err != nil {
		fmt.Fprintln(os.Stderr, "jevformal:", err)
		os.Exit(1)
	}
}

func run(protocol, effects, out, verify string, live bool) error {
	if verify != "" {
		if protocol != "" || effects != "" || out != "" {
			return fmt.Errorf("verify cannot also produce evidence")
		}
		r, err := supportmatrix.LoadReport(filepath.Join(verify, supportmatrix.ReportFile))
		if err != nil {
			return err
		}
		return supportmatrix.VerifyJevReport(r, live)
	}
	if protocol == "" || effects == "" || out == "" {
		return fmt.Errorf("provide protocol, effects and a new output directory")
	}
	if _, err := jevformal.CheckInputs(protocol, effects, live); err != nil {
		return err
	}
	if err := os.Mkdir(out, 0700); err != nil {
		return err
	}
	if err := jevformal.Copy(protocol, filepath.Join(out, "protocol")); err != nil {
		return err
	}
	if err := jevformal.Copy(effects, filepath.Join(out, "effects")); err != nil {
		return err
	}
	r, err := supportmatrix.NewJevReport(out)
	if err != nil {
		return err
	}
	raw, err := supportmatrix.Encode(r)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(out, supportmatrix.ReportFile), raw, 0600); err != nil {
		return err
	}
	if err = jevformal.Seal(out); err != nil {
		return err
	}
	r, err = supportmatrix.LoadReport(filepath.Join(out, supportmatrix.ReportFile))
	if err != nil {
		return err
	}
	return supportmatrix.VerifyJevReport(r, live)
}
