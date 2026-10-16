package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type arguments struct {
	mode, route, entry, out, responses, authorization, verify string
	requireLive                                               bool
	formal                                                    bool
	calls, workMS, callMS                                     int
}
type scriptReply struct {
	Status  int             `json:"status"`
	Body    json.RawMessage `json:"body"`
	DelayMS int             `json:"delayMS,omitempty"`
}

func main() {
	a := arguments{}
	flag.StringVar(&a.mode, "mode", "controlled", "controlled or live (requires explicit isolated authorization)")
	flag.StringVar(&a.route, "route", "jev", "jev or decisions")
	flag.StringVar(&a.entry, "entry", "plain", "plain or hooked")
	flag.StringVar(&a.out, "out", "", "new evidence directory")
	flag.StringVar(&a.responses, "responses", "", "controlled-only native response script")
	flag.StringVar(&a.authorization, "authorization", "", "live Host authorization JSON; sole key is BARNESS_VISUAL_KEY")
	flag.StringVar(&a.verify, "verify", "", "independently verify an existing bundle without network")
	flag.BoolVar(&a.requireLive, "require-live", false, "verification must contain complete real quality evidence")
	flag.BoolVar(&a.formal, "formal", false, "use the current formal catalog for the selected fixed classifier; preserve candidate fixtures")
	flag.IntVar(&a.calls, "calls", 12, "controlled-only lower call limit")
	flag.IntVar(&a.workMS, "work-ms", 225000, "controlled-only shorter work deadline")
	flag.IntVar(&a.callMS, "call-ms", 30000, "controlled-only shorter call deadline")
	flag.Parse()
	var err error
	if a.verify != "" {
		err = verify(a.verify, a.requireLive)
	} else {
		err = execute(a)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "visualeval:", err)
		os.Exit(1)
	}
}
func execute(a arguments) error {
	started := time.Now()
	watchdog := time.AfterFunc(240*time.Second, func() { fmt.Fprintln(os.Stderr, "visualeval: hard process deadline"); os.Exit(124) })
	defer watchdog.Stop()
	if (a.mode != "controlled" && a.mode != "live") || (a.entry != "plain" && a.entry != "hooked") || a.calls < 1 || a.calls > 12 || a.workMS < 1 || a.workMS > 225000 || a.callMS < 1 || a.callMS > 30000 {
		return errors.New("invalid fixed route, entry, mode or finite budget")
	}
	if a.mode == "live" && (a.responses != "" || a.calls != 12 || a.workMS != 225000 || a.callMS != 30000) {
		return errors.New("live requires the unchanged frozen experiment")
	}
	if a.mode == "controlled" && a.authorization != "" {
		return errors.New("authorization requires explicit live mode")
	}
	p, err := load(a.route)
	if err == nil && a.formal {
		err = p.useFormalCatalog()
	}
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// The work deadline includes setup. Fifteen seconds remain for local receipts.
	work, cancel := context.WithDeadline(ctx, started.Add(time.Duration(a.workMS)*time.Millisecond))
	defer cancel()
	if a.out == "" {
		a.out = filepath.Join(".evidence", "visual-evaluation", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	key, alias, reason := "synthetic-"+"visual-evaluation-credential", "synthetic-account", ""
	endpoint := ""
	var authorization *authorization
	if a.mode == "live" {
		authorization, key, reason, err = authorize(p, a)
		if err != nil {
			reason = "invalid isolated Host authorization"
		}
	}
	b, e := newBundle(a.out, key)
	if e != nil {
		return e
	}
	r := initial(p, a)
	if e = b.inputs(p, a, authorization, &r); e != nil {
		return e
	}
	if reason != "" {
		r.Reason = reason
		finalize(p, &r)
	} else {
		var transport *http.Transport
		if a.mode == "controlled" {
			var replies []scriptReply
			raw, _ := data.ReadFile("testdata/" + p.route + "-responses.json")
			if a.responses != "" {
				raw, e = readBounded(a.responses, 1<<20)
				if e != nil {
					return e
				}
			}
			if e = strict(raw, &replies); e != nil || len(replies) != 12 {
				return errors.New("controlled script must retain exactly 12 responses")
			}
			if e = b.file("script.json", raw); e != nil {
				return e
			}
			server := provider.New(map[string]string{key: "key:visual-fixture"})
			defer server.Close()
			endpoint = server.URL() + officialBasePath(p.route)
			for _, v := range replies {
				if v.Status < 100 || v.Status > 599 || v.DelayMS < 0 || v.DelayMS > 30000 {
					return errors.New("invalid bounded response script")
				}
				v := v
				server.Enqueue(provider.Reply{Status: v.Status, Header: map[string]string{"Content-Type": "application/json"}, Script: string(v.Body), Chunks: [][]byte{v.Body}, OnReceive: func() {
					if v.DelayMS > 0 {
						select {
						case <-time.After(time.Duration(v.DelayMS) * time.Millisecond):
						case <-work.Done():
						}
					}
				}})
			}
			transport = provider.LoopbackTransport()
		} else {
			endpoint = official(p.route)
			alias = authorization.AccountAlias
			transport = &http.Transport{TLSHandshakeTimeout: 5 * time.Second, IdleConnTimeout: 30 * time.Second}
		}
		defer transport.CloseIdleConnections()
		if e = executeCalls(work, p, a, b, &r, endpoint, key, alias, transport); e != nil {
			return e
		}
	}
	if time.Since(started) > 240*time.Second {
		return errors.New("process deadline exceeded")
	}
	if e = b.finish(r); e != nil {
		return e
	}
	if err != nil {
		return err
	}
	if r.Status == "FAIL" {
		return errors.New("incomplete evaluation retained; inspect report.json")
	}
	return nil
}
