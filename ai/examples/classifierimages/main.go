// Command classifierimages demonstrates Host-owned image input, final result
// interpretation and explicit revocation through the public Classify entries.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

type arguments struct {
	live                                        bool
	route, scenario, entry, config, keyEnv, out string
}

func main() {
	var a arguments
	flag.BoolVar(&a.live, "live", false, "explicitly opt into one real Host-configured call")
	flag.StringVar(&a.route, "route", "", "jev or decisions; default controlled suite runs both")
	flag.StringVar(&a.scenario, "scenario", "", "one scenario; default controlled suite runs all")
	flag.StringVar(&a.entry, "entry", "", "plain or hooked; default controlled suite runs both")
	flag.StringVar(&a.config, "config", "", "trusted Host JSON config (live only)")
	flag.StringVar(&a.keyEnv, "key-env", "", "sole named Host credential environment variable (live only)")
	flag.StringVar(&a.out, "out", "", "new evidence directory")
	flag.Parse()
	if err := execute(a); err != nil {
		fmt.Fprintln(os.Stderr, "classifierimages:", err)
		os.Exit(1)
	}
}
func execute(a arguments) error {
	selected := scenarios()
	var liveConfig *hostConfig
	key := syntheticKey
	mode := "controlled"
	if a.live {
		if a.config == "" || a.keyEnv == "" || (a.route != "jev" && a.route != "decisions") {
			return errors.New("live requires -route, -config and -key-env")
		}
		if a.scenario == "" {
			a.scenario = "context"
		}
		if a.entry == "" {
			a.entry = "plain"
		}
		allowed := false
		for _, s := range []string{"context", "reuse", "state", "instructions", "criteria", "box"} {
			if s == a.scenario {
				allowed = true
			}
		}
		if !allowed {
			return errors.New("live only accepts ordinary explicit input scenarios")
		}
		f, err := os.Open(a.config)
		if err != nil {
			return errors.New("Host configuration unavailable")
		}
		var cfg hostConfig
		d := json.NewDecoder(io.LimitReader(f, 1<<20))
		d.DisallowUnknownFields()
		err = d.Decode(&cfg)
		var extra any
		end := d.Decode(&extra)
		f.Close()
		if err != nil || end != io.EOF {
			return errors.New("invalid Host configuration")
		}
		api := ai.APITypeSafeSystemOne
		providerID := ai.ProviderTypeSafe
		if a.route == "decisions" {
			api = ai.APIOpenAIDecisions
			providerID = ai.ProviderOpenAI
		}
		if cfg.Binding.API != api || cfg.Binding.ProviderID != providerID || cfg.Target.BindingID != cfg.Binding.BindingID || cfg.Policy == nil {
			return errors.New("Host config does not match selected route")
		}
		if cfg.Binding.Retry.MaxRetries != 0 {
			return errors.New("live example requires an explicit zero-retry Host binding")
		}
		if cfg.Policy.MaxRequestBytes <= 0 || cfg.Policy.MaxRequestBytes > 16<<20 || cfg.Policy.MaxOutputBytes <= 0 || cfg.Policy.MaxOutputBytes > 16<<20 {
			return errors.New("live capture requires positive request/output limits up to 16 MiB")
		}
		// No conventional key name, key file, endpoint or catalog fallback.
		key = os.Getenv(a.keyEnv)
		if key == "" {
			return errors.New("explicit Host credential source is empty")
		}
		liveConfig = &cfg
		mode = "live"
	} else if a.config != "" || a.keyEnv != "" {
		return errors.New("Host credentials/config require explicit -live")
	}
	var chosen []scenario
	for _, s := range selected {
		if (a.route == "" || s.Route == a.route) && (a.scenario == "" || s.Name == a.scenario) && (a.entry == "" || s.Entry == a.entry) {
			chosen = append(chosen, s)
		}
	}
	if len(chosen) == 0 || a.live && len(chosen) != 1 {
		return errors.New("select a known route, scenario and entry")
	}
	files, source, err := fixtures()
	if err != nil {
		return err
	}
	if a.out == "" {
		a.out = filepath.Join(".evidence", "classifier-images", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	b, err := newBundle(a.out, key, mode)
	if err != nil {
		return err
	}
	if err = b.file("fixtures/original.png", source); err != nil {
		return err
	}
	for _, f := range files {
		if err = b.file("fixtures/"+f.Name, f.Bytes); err != nil {
			return err
		}
	}
	if err = b.record("fixtures/provenance.json", files); err != nil {
		return err
	}
	transport := &http.Transport{TLSHandshakeTimeout: 5 * time.Second, IdleConnTimeout: 30 * time.Second}
	defer transport.CloseIdleConnections()
	failed := false
	for _, s := range chosen {
		ok, err := runScenario(context.Background(), b, s, files, liveConfig, key, transport)
		if err != nil {
			return err
		}
		if !ok {
			failed = true
		}
	}
	if err = b.finish(); err != nil {
		return err
	}
	fmt.Printf("%s evidence: %s (visual quality NOT_RUN)\n", mode, a.out)
	if failed {
		return errors.New("one or more scenario assertions failed; inspect report.json")
	}
	return nil
}
