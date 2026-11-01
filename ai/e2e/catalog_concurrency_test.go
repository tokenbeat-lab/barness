package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Concurrent readers must not mutate shared input or each other's reports.
func catalogConcurrentLoads(t *testing.T, ev *evidence.Case, size, workers int) {
	t.Helper()
	var models []string
	for i := 0; i < size; i++ {
		models = append(models, strings.Replace(localChatJSON, "gpt-4.1-mini", fmt.Sprintf("fixture-%d", i), 1))
	}
	models = append(models, `{"provider":"openai","api":"openai-responses","id":"invalid","unknown":true}`)
	raw := localCatalogFile(`{"version":"concurrent","models":[` + strings.Join(models, ",") + `]}`)
	original := bytes.Clone(raw)
	type result struct {
		catCount int
		report   catalog.Report
		err      error
		found    bool
	}
	results := make([]result, workers)
	var wg sync.WaitGroup
	start := time.Now()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, r, err := catalog.Load(raw)
			_, found := c.Lookup("openai", "openai-responses", fmt.Sprintf("fixture-%d", size-1))
			results[i] = result{len(c.Models), r, err, found}
		}(i)
	}
	wg.Wait()
	runtime.ReadMemStats(&after)
	ev.Record("load-measurement", map[string]any{"models": size, "workers": workers, "elapsedSeconds": time.Since(start).Seconds(), "allocatedBytes": after.TotalAlloc - before.TotalAlloc})
	ev.Check("shared input remains unchanged", bytes.Equal(raw, original), "input changed")
	for i, r := range results {
		ev.Check(fmt.Sprintf("reader %d loads and queries", i), r.err == nil && r.catCount == size && r.report.Loaded == size && r.report.Failed == 1 && r.found, "%+v", r.err)
	}
	if len(results) > 1 && len(results[0].report.Models) > size && len(results[0].report.Models[size].Raw) > 0 {
		saved := bytes.Clone(results[1].report.Models[size].Raw)
		results[0].report.Models[size].Raw[0] ^= 1
		ev.Check("failed reports are independent", bytes.Equal(saved, results[1].report.Models[size].Raw), "report shared")
	} else {
		ev.Check("failed reports exist", false, "missing failure record")
	}
}

func TestCatalogConcurrentLoad(t *testing.T) {
	ev := run.Case(t, "CATALOG-CONCURRENT-load")
	catalogConcurrentLoads(t, ev, 16, 4)
}

func catalogConcurrentApply(t *testing.T, ev *evidence.Case, workers int) {
	t.Helper()
	binary, store, candidate, check := frozenCheckedCatalog(t, ev)
	root := catalogSnapshotRoot(t)
	output := filepath.Join(t.TempDir(), "runtime.json")
	type result struct {
		output []byte
		stderr string
		exit   int
	}
	results := make([]result, workers)
	gate := make(chan struct{})
	var wg sync.WaitGroup
	started := time.Now()
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-gate
			cmd := exec.Command(binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--out", output, "--confirm")
			cmd.Dir = root
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			out, err := cmd.Output()
			results[i] = result{out, stderr.String(), commandExit(err)}
		}(i)
	}
	close(gate)
	wg.Wait()
	successes := 0
	for i, r := range results {
		var response map[string]any
		_ = json.Unmarshal(r.output, &response)
		if r.exit == 0 {
			successes++
		} else {
			category, _ := response["category"].(string)
			ev.Check(fmt.Sprintf("writer %d rejected busy or stale", i), category == "store_busy" || category == "baseline_changed", "exit=%d stdout=%s stderr=%s", r.exit, r.output, r.stderr)
		}
		ev.Record(fmt.Sprintf("writer-%d", i), map[string]any{"exitCode": r.exit, "stdout": string(r.output), "stderr": r.stderr})
	}
	ev.Record("apply-measurement", map[string]any{"workers": workers, "successes": successes, "elapsedSeconds": time.Since(started).Seconds()})
	ev.Check("one writer commits", successes == 1, "successes=%d", successes)
	raw, err := os.ReadFile(output)
	ev.Check("output equals checked snapshot", err == nil && bytes.Equal(raw, candidate.Snapshot), "err=%v", err)
	out, exit := invokeCatalogCLI(t, ev, binary, "history", "--store", store)
	var history struct {
		History []maintenance.Revision `json:"history"`
	}
	_ = json.Unmarshal(out, &history)
	ev.Check("history has one appended revision", exit == 0 && len(history.History) == 2, "exit=%d %s", exit, out)
}
func TestCatalogConcurrentApply(t *testing.T) {
	ev := run.Case(t, "CATALOG-CONCURRENT-apply")
	catalogConcurrentApply(t, ev, 4)
}
func TestCatalogDesignLoad(t *testing.T) {
	for _, size := range []int{512, 2048} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			ev := pressureCase(t, fmt.Sprintf("CATALOG-PRESSURE-load-%d", size))
			catalogConcurrentLoads(t, ev, size, 8)
		})
	}
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("apply-%d", i), func(t *testing.T) {
			ev := pressureCase(t, fmt.Sprintf("CATALOG-PRESSURE-apply-%d", i))
			catalogConcurrentApply(t, ev, 4)
		})
	}
}
