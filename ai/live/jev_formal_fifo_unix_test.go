//go:build live && unix

package live

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestJevFormalNonRegularReport(t *testing.T) {
	ev := run.Case(t, "UC13-fifo-report")
	base := t.TempDir()
	bin := filepath.Join(base, "jevformal")
	cmd := exec.Command("go", "build", "-o", bin, "./ai/live/cmd/jevformal")
	cmd.Dir = "../.."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	if err := syscall.Mkfifo(filepath.Join(base, "live-report.json"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, bin, "-verify", base)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	_, err := cmd.CombinedOutput()
	ev.Record("file-attack", map[string]any{"kind": "fifo", "rejected": err != nil, "timedOut": ctx.Err() != nil})
	ev.Check("nonregular report refused before blocking open", err != nil && ctx.Err() == nil, "FIFO blocked until deadline")
}
