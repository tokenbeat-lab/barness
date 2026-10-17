package minimaxlive_test

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Failure inventory: an artifact grows after size admission; an import adds
// unbounded directory entries before case-contract validation. These are the
// approved auxiliary import boundary, not protocol implementation internals.
func TestArtifactGrowthCannotEscapeReadBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "growing.json")
	f, e := os.Create(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		payload := make([]byte, 1<<20)
		for !stop.Load() {
			f.Truncate(0)
			f.WriteAt(payload, 0)
		}
	}()
	defer func() { stop.Store(true); <-done }()
	for i := 0; i < 1000; i++ {
		raw, e := minimaxlive.ReadFile(path, 1024)
		if e == nil && len(raw) > 1024 {
			t.Fatalf("unbounded admitted read: %d bytes", len(raw))
		}
	}
}
func TestInventoryEntryBudgetPrecedesHashing(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 600; i++ {
		f, e := os.CreateTemp(dir, "extra-")
		if e != nil {
			t.Fatal(e)
		}
		f.Close()
	}
	if _, e := minimaxlive.Inventory(dir); e == nil {
		t.Fatal("unbounded invented artifact inventory accepted")
	}
}
