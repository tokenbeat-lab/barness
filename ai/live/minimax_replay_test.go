//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

type miniMaxReplayCall struct {
	result minimaxlive.Capture
	events []json.RawMessage
}

var miniMaxReplayCalls map[string]miniMaxReplayCall

func installMiniMaxReplay(rec *recorder) error {
	dir := os.Getenv("BARNESS_MINIMAX_REPLAY")
	r, e := minimaxlive.Verify(dir, false)
	if e != nil {
		return e
	}
	if r.Mode == "replay" {
		return fmt.Errorf("nested replay cannot replace original capture provenance")
	}
	if r.Identity != minimaxReport.Identity {
		return fmt.Errorf("replay identity differs")
	}
	raw, e := minimaxlive.ReadFile(filepath.Join(dir, "minimax-catalog.json"), 2<<20)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(raw, &minimaxCatalog); e != nil {
		return e
	}
	minimaxReport.CatalogVersion = r.CatalogVersion
	minimaxReport.CatalogHash = r.CatalogHash
	run.SetVersion("model_catalog_version", r.CatalogVersion)
	run.SetVersion("model_catalog_hash", r.CatalogHash)
	origin, e := minimaxlive.Snapshot(dir)
	if e != nil {
		return e
	}
	if e = run.Record("minimax-replay-origin", origin); e != nil {
		return e
	}
	var requests []string
	var exchanges []minimaxlive.Exchange
	var kinds []string
	var failures []ai.Code
	miniMaxReplayCalls = map[string]miniMaxReplayCall{}
	for _, c := range r.Calls {
		raw, e := minimaxlive.ReadFile(filepath.Join(dir, c.CaseID, "result.json"), 2<<20)
		if e != nil {
			return e
		}
		var call miniMaxReplayCall
		if e = json.Unmarshal(raw, &call.result); e != nil {
			return e
		}
		raw, e = minimaxlive.ReadFile(filepath.Join(dir, c.CaseID, "events.json"), 2<<20)
		if e != nil {
			return e
		}
		if e = json.Unmarshal(raw, &call.events); e != nil {
			return e
		}
		miniMaxReplayCalls[c.CaseID] = call
		if c.ID == "preflight-error" {
			continue
		}
		raw, e = minimaxlive.ReadFile(filepath.Join(dir, c.CaseID, "exchanges.json"), 2<<20)
		if e != nil {
			return e
		}
		var ex []minimaxlive.Exchange
		if e = json.Unmarshal(raw, &ex); e != nil || len(ex) != 1 {
			return fmt.Errorf("invalid capture replay attempts")
		}
		requests = append(requests, ex[0].RequestBody)
		exchanges = append(exchanges, ex[0])
		kinds = append(kinds, c.ID)
		failures = append(failures, call.result.ErrorCode)
	}
	index := 0
	sourceRaw, _ := minimaxlive.ReadFile(filepath.Join(dir, minimaxlive.ReportFile), 2<<20)
	minimaxReport.ReplaySource = minimaxlive.Hash(sourceRaw)
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(req.Body)
		req.Body.Close()
		if index >= len(requests) || string(raw) != requests[index] {
			return nil, fmt.Errorf("replay request sequence differs")
		}
		ex := exchanges[index]
		kind := kinds[index]
		failure := failures[index]
		index++
		if ex.Status == 0 {
			if failure == ai.CodeDeadlineExceeded {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}
			if ex.TransportError == "" {
				return nil, fmt.Errorf("replay transport failure proof absent")
			}
			return nil, errors.New(ex.TransportError)
		}
		header := http.Header{}
		for k, v := range ex.ResponseHeaders {
			header.Set(k, v)
		}
		var body io.ReadCloser = io.NopCloser(strings.NewReader(ex.ResponseBody))
		if kind == "cancel" {
			pr, pw := io.Pipe()
			body = pr
			go func() {
				io.WriteString(pw, ex.ResponseBody)
				<-req.Context().Done()
				pw.CloseWithError(req.Context().Err())
			}()
		}
		return &http.Response{StatusCode: ex.Status, Header: header, Body: body, Request: req}, nil
	})
	return nil
}
func compareMiniMaxReplay(cs *evidence.Case, id string, actual minimaxlive.Capture, events []json.RawMessage) {
	if miniMaxReplayCalls == nil {
		return
	}
	expected, ok := miniMaxReplayCalls[id]
	if !cs.Check("replay has exact captured public call", ok, "unknown case") {
		return
	}
	normalize := func(c minimaxlive.Capture) minimaxlive.Capture {
		c.Result.Message.Timestamp = 0
		c.Result.Message.DurationMs = ai.Nullable[int64]{}
		c.Result.Metadata.RequestID = ""
		c.Result.Metadata.Attempts = append([]ai.Attempt(nil), c.Result.Metadata.Attempts...)
		for i := range c.Result.Metadata.Attempts {
			c.Result.Metadata.Attempts[i].AttemptID = ""
		}
		return c
	}
	cs.Check("replayed Result matches capture except Host clock/IDs", minimaxlive.JSONEqual(normalize(expected.result), normalize(actual)), "public Result differs")
	types := func(events []json.RawMessage) []string {
		var out []string
		for _, raw := range events {
			var v struct {
				Type string `json:"type"`
			}
			json.Unmarshal(raw, &v)
			out = append(out, v.Type)
		}
		return out
	}
	cs.Check("replayed event sequence matches capture", minimaxlive.JSONEqual(types(expected.events), types(events)), "event sequence differs")
}
