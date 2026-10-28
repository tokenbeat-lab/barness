package supportmatrix

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// RegionalReservation binds a native capture to the durable Host reservation.
// Its two keys prevent both copied paths and relabeling the same approval from
// resetting a budget. A genuinely new approval has a new ID and approved scope
// (including expiry); the trusted Host owns issuance and this repository store.
type RegionalReservation struct {
	Schema         int       `json:"schema"`
	GrantID        string    `json:"grantId"`
	GrantSHA256    string    `json:"grantSha256"`
	ScopeSHA256    string    `json:"scopeSha256"`
	Combo          string    `json:"combo"`
	Session        string    `json:"session"`
	Mode           string    `json:"mode"`
	StartedAt      time.Time `json:"startedAt"`
	MaxCalls       int       `json:"maxCalls"`
	MaxAttempts    int       `json:"maxAttempts"`
	MaxUSD         float64   `json:"maxUSD,omitempty"`
	BinarySHA256   string    `json:"binarySha256,omitempty"`
	CallIntervalMs int       `json:"callIntervalMs,omitempty"`
}

func regionalReservationStore(combo string) (string, error) {
	root, err := minimaxlive.Root()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = root
	raw, err := cmd.Output()
	if err != nil {
		return "", refuse("regional trusted repository unavailable")
	}
	common := strings.TrimSpace(string(raw))
	if !filepath.IsAbs(common) {
		return "", refuse("regional trusted repository path invalid")
	}
	if RegionalIsKimi(combo) {
		return filepath.Join(common, "barness-kimi-grants"), nil
	}
	return filepath.Join(common, "barness-regional-grants"), nil
}

func regionalReservationKeys(g RegionalGrant) (string, string) {
	id := glmHash([]byte(g.ID))
	g.ID = ""
	// RFC3339 offsets describe an instant, not a different approval. Normalize
	// its representation before deriving the semantic reservation identity.
	g.ExpiresAt = g.ExpiresAt.UTC()
	raw, _ := json.Marshal(g)
	return "id-" + id + ".json", "scope-" + glmHash(raw) + ".json"
}

// ReserveRegionalGrant burns the approval before any HTTP. An interrupted or
// partial write is deliberately never repaired or removed: ambiguity consumes
// the approval, requiring a new explicitly authorized run.
func ReserveRegionalGrant(g RegionalGrant, raw []byte, session string, at time.Time) (RegionalReservation, error) {
	receipt := RegionalReservation{Schema: 2, GrantID: g.ID, GrantSHA256: glmHash(raw), Combo: g.Combo, Session: session, Mode: g.Mode, StartedAt: at, MaxCalls: g.MaxCalls, MaxAttempts: g.MaxAttempts}
	if RegionalIsKimi(g.Combo) {
		receipt.MaxUSD, receipt.BinarySHA256, receipt.CallIntervalMs = g.MaxUSD, g.BinarySHA256, g.CallIntervalMs
	}
	id, scope := regionalReservationKeys(g)
	receipt.ScopeSHA256 = strings.TrimSuffix(strings.TrimPrefix(scope, "scope-"), ".json")
	store, err := regionalReservationStore(g.Combo)
	if err != nil {
		return receipt, err
	}
	data, _ := json.Marshal(receipt)
	if err := livegrant.Consume(store, []string{scope, id}, data); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func verifyRegionalReservation(g RegionalGrant, receipt RegionalReservation) error {
	store, err := regionalReservationStore(g.Combo)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(store)
	if err != nil {
		return refuse("regional authoritative reservation unavailable")
	}
	defer root.Close()
	id, scope := regionalReservationKeys(g)
	expected, _ := json.Marshal(receipt)
	if receipt.Schema != 2 || receipt.ScopeSHA256 != strings.TrimSuffix(strings.TrimPrefix(scope, "scope-"), ".json") {
		return refuse("regional reservation identity differs")
	}
	for _, key := range []string{scope, id} {
		raw, e := regionalRead(root, key)
		var value RegionalReservation
		if e != nil || regionalDecode(raw, &value) != nil {
			return refuse("regional authoritative reservation incomplete")
		}
		if !bytes.Equal(raw, expected) {
			return refuse("regional authoritative reservation differs")
		}
	}
	return nil
}
