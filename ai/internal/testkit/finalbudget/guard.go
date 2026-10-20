// Package finalbudget is the fixed provider-sites-final-acceptance-2 Host
// spending guard. It is test/acceptance tooling, never a production AI policy.
package finalbudget

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const Round = "provider-sites-final-acceptance-2"

var mu sync.Mutex

type ledger struct {
	Schema   int          `json:"schema"`
	Round    string       `json:"round"`
	Maximum  int64        `json:"maxMicrousd"`
	Reserved int64        `json:"reservedMicrousd"`
	Attempts int          `json:"attempts"`
	Basis    string       `json:"basis,omitempty"`
	Last     *reservation `json:"last,omitempty"`
}

type reservation struct {
	Model         string  `json:"model"`
	InputReserve  int64   `json:"inputReserve"`
	OutputReserve int64   `json:"outputReserve"`
	InputRate     float64 `json:"inputRate"`
	OutputRate    float64 `json:"outputRate"`
	Microusd      int64   `json:"microusd"`
}

// Reserve runs on the final encoded request before any network send. It keeps
// the entire conservative known-rate reservation after every send, including
// cancellation, rejection, network error and absent usage. It does not claim
// the byte reservation is a tokenizer or a provider invoice guarantee.
func Reserve(req *http.Request) error {
	round, path := os.Getenv("BARNESS_FINAL_ROUND"), os.Getenv("BARNESS_FINAL_BUDGET")
	if round == "" && path == "" {
		return nil
	}
	if round != Round || path == "" {
		return fmt.Errorf("fixed final-round USD configuration required")
	}
	if req.GetBody == nil {
		return fmt.Errorf("final USD guard requires immutable request bytes")
	}
	body, err := req.GetBody()
	if err != nil {
		return fmt.Errorf("final USD request unavailable")
	}
	defer body.Close()
	raw, err := io.ReadAll(io.LimitReader(body, (4<<20)+1))
	if err != nil || len(raw) > 4<<20 {
		return fmt.Errorf("final USD request byte ceiling exceeded")
	}
	estimate, err := estimate(req, raw)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<10 {
		return fmt.Errorf("owned bounded final USD ledger required")
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("final USD ledger read failed")
	}
	var value ledger
	if json.Unmarshal(raw, &value) != nil || value.Schema != 1 || value.Round != Round || value.Maximum <= 0 || value.Maximum > 30000000 || value.Reserved < 0 || value.Attempts < 0 || value.Attempts >= 32 {
		return fmt.Errorf("fixed finite final USD ledger invalid")
	}
	if estimate.Microusd <= 0 || estimate.Microusd > value.Maximum-value.Reserved {
		return fmt.Errorf("remaining conservative final USD reservation exhausted before send")
	}
	value.Reserved += estimate.Microusd
	value.Attempts++
	value.Last = &estimate
	value.Basis = "Final encoded bytes+1024 and explicit image reserve; final output setting or sourced image-model estimate; maximum catalog tier/rate; no refund for unknown usage; conservative known-rate operational estimate, not invoice."
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("final USD ledger encoding failed")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".final-usd-")
	if err != nil {
		return fmt.Errorf("final USD durable reservation failed")
	}
	name := file.Name()
	defer os.Remove(name)
	if err = file.Chmod(0600); err == nil {
		_, err = file.Write(append(encoded, '\n'))
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		return fmt.Errorf("final USD reservation publication failed")
	}
	return nil
}

func positive(value any) int64 {
	n, ok := value.(float64)
	if !ok || n < 0 || math.IsNaN(n) || math.IsInf(n, 0) || n != math.Trunc(n) || n > 1<<31 {
		return 0
	}
	return int64(n)
}
