package livegrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Receipt struct {
	Schema      int       `json:"schema"`
	GrantID     string    `json:"grantId"`
	GrantSHA256 string    `json:"grantSha256"`
	ScopeSHA256 string    `json:"scopeSha256"`
	Session     string    `json:"session"`
	StartedAt   time.Time `json:"startedAt"`
}

func store() (string, error) {
	raw, e := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	p := strings.TrimSpace(string(raw))
	if e != nil || !filepath.IsAbs(p) {
		return "", fmt.Errorf("trusted common Git grant store unavailable")
	}
	return filepath.Join(p, "barness-live-process-grants"), nil
}
func keys(g Grant) (string, string) {
	id := Hash([]byte(g.ID))
	g.ID = ""
	g.ExpiresAt = g.ExpiresAt.UTC()
	raw, _ := json.Marshal(g)
	return "id-" + id + ".json", "scope-" + Hash(raw) + ".json"
}

// Claim burns both the human approval identity and its complete semantic scope
// before HTTP. A partial write is consumed forever; no reset/recovery exists.
func Claim(g Grant, raw []byte, session string, at time.Time) (Receipt, error) {
	if e := g.Check(g.Scope, at); e != nil {
		return Receipt{}, e
	}
	id, scope := keys(g)
	r := Receipt{1, g.ID, Hash(raw), strings.TrimSuffix(strings.TrimPrefix(scope, "scope-"), ".json"), session, at}
	p, e := store()
	if e != nil {
		return r, e
	}
	data, _ := json.Marshal(r)
	if err := Consume(p, []string{id, scope}, data); err != nil {
		return r, fmt.Errorf("fresh process %w", err)
	}
	return r, nil
}

func Verify(g Grant, raw []byte, r Receipt) error {
	id, scope := keys(g)
	if r.Schema != 1 || r.GrantID != g.ID || r.GrantSHA256 != Hash(raw) || "scope-"+r.ScopeSHA256+".json" != scope || r.Session == "" || g.Check(g.Scope, r.StartedAt) != nil {
		return fmt.Errorf("process claim attribution differs")
	}
	p, e := store()
	if e != nil {
		return e
	}
	root, e := os.OpenRoot(p)
	if e != nil {
		return e
	}
	defer root.Close()
	want, _ := json.Marshal(r)
	for _, k := range []string{scope, id} {
		got, e := root.ReadFile(k)
		if e != nil || !bytes.Equal(got, want) {
			return fmt.Errorf("authoritative process claim differs")
		}
	}
	return nil
}
