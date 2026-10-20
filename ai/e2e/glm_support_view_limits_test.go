package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Invalid derived-view paths must be refused before their payloads are read.
func TestGLMSupportViewBounds(t *testing.T) {
	var index struct {
		Routes map[string]struct{ Archive, ArchiveSHA256, ReportPath string }
	}
	raw, err := os.ReadFile("../live/testdata/glm-formal/support-views.json")
	if err != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("current view missing")
	}
	entry := index.Routes["chat"]
	for _, mode := range []string{"unlisted", "many", "deep"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "P0-GLM13-support-view-"+mode)
			dir := extractGLMFormalArchive(t, entry.Archive, entry.ArchiveSHA256)
			base := filepath.Dir(filepath.Join(dir, entry.ReportPath))
			count, size := 1, int64(16<<20)
			if mode == "many" {
				count, size = 513, 256<<10
			}
			for i := 0; i < count; i++ {
				name := fmt.Sprintf("original/unlisted-%04d.json", i)
				if mode == "deep" {
					name = "original/" + strings.Repeat("unowned/", 64) + "payload.json"
				}
				p := filepath.Join(base, name)
				if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
					t.Fatal(err)
				}
				file, e := os.Create(p)
				if e != nil {
					t.Fatal(e)
				}
				e = file.Truncate(size)
				file.Close()
				if e != nil {
					t.Fatal(e)
				}
			}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			_, err = supportmatrix.NewGLMSupportViewReport(base)
			runtime.ReadMemStats(&after)
			allocated := after.TotalAlloc - before.TotalAlloc
			ev.Record("resource-envelope", map[string]any{"unownedFiles": count, "logicalBytesPerFile": size, "allocatedBytes": allocated, "maxAllocatedBytes": 48 << 20, "rejected": err != nil, "error": errString(err)})
			ev.Check("invalid finite-owned shape rejected before payload read", err != nil && strings.Contains(err.Error(), "unexpected support view artifact") && allocated < 48<<20, "err=%v allocated=%d", err, allocated)
		})
	}
}
