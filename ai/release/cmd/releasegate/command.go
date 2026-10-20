package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const tailBytes = 4 << 10

type stepResult struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	Argv         []string  `json:"argv,omitempty"`
	ExitCode     *int      `json:"exitCode,omitempty"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
	FinishedAt   time.Time `json:"finishedAt,omitempty"`
	Seconds      float64   `json:"seconds"`
	LimitSeconds float64   `json:"limitSeconds,omitempty"`
	Log          string    `json:"log,omitempty"`
	Events       string    `json:"events,omitempty"`
	Error        string    `json:"error,omitempty"`
	Tail         string    `json:"tail,omitempty"`
}

func commandArgs(id string) []string {
	switch id {
	case "go-vet":
		return []string{"go", "vet", "./..."}
	case "go-vet-live":
		return []string{"go", "vet", "-tags", "live", "./..."}
	case "go-test":
		return []string{"go", "test", "-json", "-count=1", "-timeout=20m", "./..."}
	case "go-test-race":
		return []string{"go", "test", "-race", "-json", "-count=1", "-timeout=75m", "./..."}
	}
	return nil
}
func buildEnvironment() []string {
	var out []string
	// The frozen toolchain and build settings, with no ambient vendor credential.
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH", "GOROOT", "GOFLAGS", "GOWORK", "GOTOOLCHAIN", "GOOS", "GOARCH", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS"} {
		if v, ok := os.LookupEnv(key); ok {
			out = append(out, key+"="+v)
		}
	}
	return out
}
func runTool(ctx context.Context, root, out, id string) stepResult {
	s := stepResult{ID: id, Status: "FAIL", Argv: commandArgs(id), StartedAt: time.Now().UTC()}
	code := -1
	s.ExitCode = &code
	if len(s.Argv) == 0 {
		s.Error = "unknown verification command"
		return s
	}
	logName := id + ".log"
	log, err := os.OpenFile(filepath.Join(out, logName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		s.Error = "command log unavailable"
		return s
	}
	s.Log = logName
	writer := &commandOutput{dst: io.MultiWriter(log, os.Stderr)}
	limit := 10 * time.Minute
	if id == "go-test" {
		limit = 25 * time.Minute
	}
	if id == "go-test-race" {
		limit = 80 * time.Minute
	}
	s.LimitSeconds = limit.Seconds()
	// Package timeouts exclude builds and tool startup. This outer deadline
	// bounds the complete process while retaining the 20m/75m package limits.
	childCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	stdout := io.Writer(writer)
	var eventFile *os.File
	var events *eventWriter
	if id == "go-test" || id == "go-test-race" {
		s.Events = id + ".events.jsonl"
		eventFile, err = os.OpenFile(filepath.Join(out, s.Events), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			log.Close()
			s.Error = "event output unavailable"
			return s
		}
		events = newEventWriter(eventFile)
		stdout = &eventOutput{log: writer, events: events, cancel: cancel}
	}
	cmd := exec.CommandContext(childCtx, s.Argv[0], s.Argv[1:]...)
	cmd.Dir = root
	cmd.Env = buildEnvironment()
	if events != nil {
		base, absErr := filepath.Abs(filepath.Join(out, map[string]string{"go-test": "offline", "go-test-race": "race"}[id]))
		if absErr != nil {
			log.Close()
			eventFile.Close()
			s.Error = "evidence directory invalid"
			return s
		}
		cmd.Env = append(cmd.Env, "BARNESS_AI_PIDIFF=1", "BARNESS_AI_PRESSURE=1", "BARNESS_AI_EVIDENCE_DIR="+base)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 5 * time.Second
	cmd.Stdout = stdout
	cmd.Stderr = &eventOutput{log: writer, cancel: cancel}
	fmt.Fprintln(os.Stderr, "releasegate:", strings.Join(s.Argv, " "))
	err = cmd.Run()
	if err == nil {
		code = 0
	} else if ex := new(exec.ExitError); errors.As(err, &ex) {
		code = ex.ExitCode()
	}
	if events != nil {
		if e := events.Finish(); e != nil {
			s.Error = e.Error()
		}
		if e := eventFile.Close(); e != nil {
			s.Error = "event output close failed"
		}
	}
	if e := log.Close(); e != nil {
		s.Error = "command log close failed"
	}
	if err != nil && s.Error == "" {
		s.Error = "verification process failed"
	}
	if ctx.Err() != nil {
		s.Error = "verification cancelled"
	}
	if errors.Is(childCtx.Err(), context.DeadlineExceeded) {
		s.Error = "verification process deadline exceeded"
	}
	s.FinishedAt = time.Now().UTC()
	s.Seconds = s.FinishedAt.Sub(s.StartedAt).Seconds()
	if code == 0 && s.Error == "" {
		s.Status = "PASS"
	} else {
		s.Tail = writer.tailString()
	}
	return s
}

type commandOutput struct {
	mu   sync.Mutex
	dst  io.Writer
	tail []byte
}

func (w *commandOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(p) >= tailBytes {
		w.tail = append(w.tail[:0], p[len(p)-tailBytes:]...)
	} else {
		w.tail = append(w.tail, p...)
		if len(w.tail) > tailBytes {
			w.tail = append(w.tail[:0], w.tail[len(w.tail)-tailBytes:]...)
		}
	}
	return w.dst.Write(p)
}
func (w *commandOutput) tailString() string { w.mu.Lock(); defer w.mu.Unlock(); return string(w.tail) }

type eventOutput struct {
	log    io.Writer
	events *eventWriter
	cancel context.CancelFunc
}

func (w *eventOutput) Write(p []byte) (int, error) {
	if _, err := w.log.Write(p); err != nil {
		w.cancel()
		return 0, err
	}
	if w.events == nil {
		return len(p), nil
	}
	n, err := w.events.Write(p)
	if err != nil {
		w.cancel()
	}
	return n, err
}
