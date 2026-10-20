package supportmatrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// RegionalReservation binds a native capture to the durable Host reservation.
// Its two keys prevent both copied paths and relabeling the same approval from
// resetting a budget. A genuinely new approval has a new ID and approved scope
// (including expiry); the trusted Host owns issuance and this repository store.
type RegionalReservation struct {
	Schema      int       `json:"schema"`
	GrantID     string    `json:"grantId"`
	GrantSHA256 string    `json:"grantSha256"`
	ScopeSHA256 string    `json:"scopeSha256"`
	Combo       string    `json:"combo"`
	Session     string    `json:"session"`
	Mode        string    `json:"mode"`
	StartedAt   time.Time `json:"startedAt"`
	MaxCalls    int       `json:"maxCalls"`
	MaxAttempts int       `json:"maxAttempts"`
}

func regionalReservationStore() (string, error) {
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
	return filepath.Join(common, "barness-regional-grants"), nil
}

func regionalReservationKeys(g RegionalGrant) (string, string) {
	id := glmHash([]byte(g.ID))
	g.ID = ""
	raw, _ := json.Marshal(g)
	return "id-" + id + ".json", "scope-" + glmHash(raw) + ".json"
}

// ReserveRegionalGrant burns the approval before any HTTP. An interrupted or
// partial write is deliberately never repaired or removed: ambiguity consumes
// the approval, requiring a new explicitly authorized run.
func ReserveRegionalGrant(g RegionalGrant, raw []byte, session string, at time.Time) (RegionalReservation, error) {
	receipt := RegionalReservation{Schema: 2, GrantID: g.ID, GrantSHA256: glmHash(raw), Combo: g.Combo, Session: session, Mode: g.Mode, StartedAt: at, MaxCalls: g.MaxCalls, MaxAttempts: g.MaxAttempts}
	id, scope := regionalReservationKeys(g)
	receipt.ScopeSHA256 = strings.TrimSuffix(strings.TrimPrefix(scope, "scope-"), ".json")
	store, err := regionalReservationStore()
	if err != nil {
		return receipt, err
	}
	if err = os.MkdirAll(store, 0700); err != nil {
		return receipt, err
	}
	root, err := os.OpenRoot(store)
	if err != nil {
		return receipt, err
	}
	defer root.Close()
	lock, err := root.OpenFile("lock", os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return receipt, err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return receipt, err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	data, _ := json.Marshal(receipt)
	for _, key := range []string{scope, id} {
		f, e := root.OpenFile(key, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
		if e != nil {
			return receipt, fmt.Errorf("regional authorization consumed or reservation unavailable")
		}
		n, e := f.Write(data)
		if e == nil && n != len(data) {
			e = fmt.Errorf("regional reservation incomplete")
		}
		if e == nil {
			e = f.Sync()
		}
		closeErr := f.Close()
		if e != nil {
			return receipt, e
		}
		if closeErr != nil {
			return receipt, closeErr
		}
	}
	dir, err := os.Open(store)
	if err != nil {
		return receipt, err
	}
	defer dir.Close()
	if err = dir.Sync(); err != nil {
		return receipt, err
	}
	// A new ledger directory must itself survive a crash, not merely the
	// entries inside it. Sync its trusted Git parent before authorizing send.
	parent, err := os.Open(filepath.Dir(store))
	if err != nil {
		return receipt, err
	}
	defer parent.Close()
	if err = parent.Sync(); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func verifyRegionalReservation(g RegionalGrant, receipt RegionalReservation) error {
	store, err := regionalReservationStore()
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
