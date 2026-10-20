package main

import (
	"context"
	"errors"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"strings"
	"time"
)

type executionReport struct {
	Schema      int          `json:"schema"`
	Stage       string       `json:"stage"`
	Status      string       `json:"status"`
	Race        string       `json:"race"`
	ReleasePass bool         `json:"releasePass"`
	BlockedBy   string       `json:"blockedBy,omitempty"`
	Error       string       `json:"error,omitempty"`
	Steps       []stepResult `json:"steps"`
}

func newExecution(stage string) executionReport {
	r := executionReport{Schema: 1, Stage: stage, Status: "RUNNING", Race: "NOT_RUN"}
	for _, id := range []string{"go-vet", "go-vet-live", "go-test", "ordinary-evidence", "go-test-race"} {
		r.Steps = append(r.Steps, stepResult{ID: id, Status: "NOT_RUN"})
	}
	return r
}
func (r executionReport) ExitCode() int {
	if r.Status == "READY_FOR_RACE" {
		return 2
	}
	if r.Status == "PASS" && r.ReleasePass {
		return 0
	}
	return 1
}
func (r *executionReport) fail(id string, err error) {
	r.Status = "FAIL"
	r.BlockedBy = id
	if err != nil {
		r.Error = err.Error()
	}
}
func (r executionReport) commands() []release.Command {
	var out []release.Command
	for _, s := range r.Steps {
		if len(s.Argv) == 0 {
			continue
		}
		out = append(out, release.Command{ID: s.ID, Line: strings.Join(s.Argv, " "), Passed: s.Status == "PASS" && s.ExitCode != nil && *s.ExitCode == 0, Duration: time.Duration(s.Seconds * float64(time.Second)), Tail: s.Tail})
	}
	return out
}
func (r executionReport) save(out string) error { return writeJSON(out, "execution-report.json", r) }

// The two callbacks are private collection boundaries. Production supplies
// frozen-source/input checks and native consumers; tests use real controlled
// child processes without creating formal model-support evidence.
func runCheckCommands(ctx context.Context, root, out string, collectAll bool, guard, consume func() error) executionReport {
	r := newExecution("check")
	if err := r.save(out); err != nil {
		r.fail("report", errors.New("execution report unavailable"))
		return r
	}
	for i, id := range []string{"go-vet", "go-vet-live", "go-test", "ordinary-evidence"} {
		if err := ctx.Err(); err != nil {
			r.fail(id, errors.New("verification cancelled"))
			break
		}
		if err := guard(); err != nil {
			r.fail("source-or-input", err)
			break
		}
		r.Steps[i].Status = "RUNNING"
		if err := r.save(out); err != nil {
			r.fail("report", errors.New("execution report unavailable"))
			break
		}
		if id == "ordinary-evidence" {
			s := stepResult{ID: id, Status: "PASS", StartedAt: time.Now().UTC()}
			if err := consume(); err != nil {
				s.Status = "FAIL"
				s.Error = err.Error()
			}
			s.FinishedAt = time.Now().UTC()
			s.Seconds = s.FinishedAt.Sub(s.StartedAt).Seconds()
			r.Steps[i] = s
		} else {
			r.Steps[i] = runTool(ctx, root, out, id)
		}
		if err := guard(); err != nil {
			r.fail("source-or-input", err)
			break
		}
		if r.Steps[i].Status != "PASS" {
			if r.Status != "FAIL" {
				r.fail(id, errors.New("verification stage failed"))
			}
			if !collectAll {
				break
			}
		}
		if err := r.save(out); err != nil {
			r.fail("report", errors.New("execution report unavailable"))
			break
		}
	}
	if r.Status != "FAIL" {
		r.Status = "READY_FOR_RACE"
	}
	if err := r.save(out); err != nil {
		r.fail("report", errors.New("execution report unavailable"))
	}
	return r
}
