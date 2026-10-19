package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Approved isolated seam: actual child processes at the CLI's fixed Go-tool
// boundary. Synthetic success exercises scheduling only, never formal proofs.
// Failure inventory: either vet, ordinary, native consumption, source drift,
// cancellation and report writes fail closed; collect-all never crosses race.
func TestCheckSequenceManualBoundary(t *testing.T) {
	for _, mode := range []string{"ready", "vet", "vet-live", "ordinary", "consume", "collect-all", "source-drift", "report-write"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			out := t.TempDir()
			log := fakeGoTool(t, root, mode)
			if mode == "report-write" {
				if err := os.Mkdir(filepath.Join(out, "execution-report.json"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			guard := func() error {
				if mode == "source-drift" {
					return errors.New("source changed")
				}
				return nil
			}
			consume := func() error {
				if mode == "consume" {
					return errors.New("native evidence invalid")
				}
				return nil
			}
			r := runCheckCommands(context.Background(), root, out, mode == "collect-all", guard, consume)
			raw, _ := os.ReadFile(log)
			if strings.Contains(string(raw), "-race") || r.Race != "NOT_RUN" || r.ReleasePass {
				t.Fatalf("automatic race or certification: %+v", r)
			}
			expected := map[string]int{"ready": 3, "vet": 1, "vet-live": 2, "ordinary": 3, "consume": 3, "collect-all": 3, "source-drift": 0, "report-write": 0}[mode]
			if got := len(strings.FieldsFunc(strings.TrimSpace(string(raw)), func(r rune) bool { return r == '\n' })); got != expected {
				t.Fatalf("commands=%d want=%d", got, expected)
			}
			if mode == "ready" {
				if r.Status != "READY_FOR_RACE" || r.ExitCode() != 2 {
					t.Fatalf("not waiting for human: %+v", r)
				}
			} else if r.Status != "FAIL" || r.ExitCode() != 1 {
				t.Fatalf("failed execution passed: %+v", r)
			}
			for _, s := range r.Steps {
				if s.Status == "PASS" && (s.ExitCode == nil || *s.ExitCode != 0) && s.ID != "ordinary-evidence" {
					t.Fatalf("PASS lacks numeric success: %+v", s)
				}
			}
		})
	}
}

func TestFixedCommandEnvelopes(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	fakeGoTool(t, root, "ready")
	for _, id := range []string{"go-test", "go-test-race"} {
		s := runTool(context.Background(), root, out, id)
		got := strings.Join(s.Argv, " ")
		want := "go test -json -count=1 -timeout=20m ./..."
		if id == "go-test-race" {
			want = "go test -race -json -count=1 -timeout=75m ./..."
		}
		if got != want || s.Status != "PASS" {
			t.Fatalf("%s: %+v", id, s)
		}
	}
}

func TestCommandCancellation(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	fakeGoTool(t, root, "ready")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := runTool(ctx, root, out, "go-test")
	if s.Status != "FAIL" || s.Error == "" {
		t.Fatalf("cancelled execution passed: %+v", s)
	}
}

func fakeGoTool(t *testing.T, root, mode string) string {
	t.Helper()
	log := filepath.Join(root, "commands")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\n", log)
	switch mode {
	case "vet", "collect-all":
		script += "case \"$*\" in 'vet ./...') exit 7;; esac\n"
	case "vet-live":
		script += "case \"$*\" in 'vet -tags live ./...') exit 8;; esac\n"
	case "ordinary":
		script += "case \"$1\" in test) exit 9;; esac\n"
	}
	script += "if [ \"$1\" = test ]; then printf '%s\\n' '{\"Time\":\"2026-10-20T00:00:00Z\",\"Action\":\"pass\",\"Package\":\"controlled/scheduler-fixture\",\"Elapsed\":0.001}'; fi\n"
	if err := os.WriteFile(filepath.Join(root, "go"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}
