package main

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// Failure inventory before checkpoint implementation: source additions,
// removals, test/tool/fixture edits, selected-input drift, receipt/log edits,
// unsafe file types and oversized metadata must refuse continuation. Ordinary
// native consumers remain a separate mandatory boundary in the public stage.
func TestCheckpointRefusesDrift(t *testing.T) {
	for _, mode := range []string{"current", "documentation", "readme-source", "evidence-script", "source", "added", "removed", "test", "input", "receipt", "log", "bundle", "tool"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			out := t.TempDir()
			fakeGoTool(t, root, "ready")
			if err := os.MkdirAll(filepath.Join(root, "ai"), 0700); err != nil {
				t.Fatal(err)
			}
			put := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			put("go.mod", "module controlled\n")
			put("go.sum", "")
			put("ai/main.go", "package controlled\n")
			put("ai/main_test.go", "package controlled\n")
			put("input.json", "{}")
			o := options{Stage: "check", Matrix: filepath.Join(root, "input.json")}
			frozen, err := captureCheckpointInputs(root, o)
			if err != nil {
				t.Fatal(err)
			}
			r := runCheckCommands(context.Background(), root, out, false, func() error { return nil }, func() error { return nil })
			if r.Status != "READY_FOR_RACE" {
				t.Fatal(r)
			}
			bundle := filepath.Join(out, "offline", "run")
			if err := os.MkdirAll(bundle, 0700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"manifest.json", "audit.json"} {
				if err := os.WriteFile(filepath.Join(bundle, name), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := saveCheckpoint(root, out, o, frozen, bundle); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "documentation":
				put("ai/README.md", "updated documentation")
			case "readme-source":
				put("ai/README_test.go", "package controlled\n")
			case "evidence-script":
				if err := os.MkdirAll(filepath.Join(root, ".scratch/fixture/evidence"), 0700); err != nil {
					t.Fatal(err)
				}
				put(".scratch/fixture/evidence/run.py", "print('changed executed fixture')")
			case "source":
				put("ai/main.go", "package changed\n")
			case "added":
				put("ai/new.go", "package controlled\n")
			case "removed":
				if err := os.Remove(filepath.Join(root, "ai/main.go")); err != nil {
					t.Fatal(err)
				}
			case "test":
				put("ai/main_test.go", "package changed\n")
			case "input":
				put("input.json", "{\"changed\":true}")
			case "receipt":
				if err := os.WriteFile(filepath.Join(out, "execution-report.json"), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			case "log":
				if err := os.WriteFile(filepath.Join(out, "go-test.log"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			case "bundle":
				if err := os.WriteFile(filepath.Join(bundle, "manifest.json"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			case "tool":
				if err := os.WriteFile(filepath.Join(root, "go"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			_, _, err = loadCheckpoint(root, out)
			wantPass := mode == "current" || mode == "documentation"
			if (err == nil) != wantPass {
				t.Fatalf("mode=%s accepted=%t error=%v", mode, err == nil, err)
			}
		})
	}
}
func TestCheckpointBoundedFiles(t *testing.T) {
	for _, mode := range []string{"regular", "symlink", "fifo", "oversized", "directory"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "file")
			var err error
			switch mode {
			case "regular":
				err = os.WriteFile(p, []byte("ok"), 0600)
			case "symlink":
				err = os.Symlink("target", p)
			case "fifo":
				err = syscall.Mkfifo(p, 0600)
			case "oversized":
				err = os.WriteFile(p, []byte("too large"), 0600)
			case "directory":
				err = os.Mkdir(p, 0700)
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = readRegular(p, 4)
			if (err == nil) != (mode == "regular") {
				t.Fatalf("mode=%s accepted=%t", mode, err == nil)
			}
		})
	}
}
