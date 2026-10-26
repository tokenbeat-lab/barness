package livegrant

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
	if e = os.MkdirAll(p, 0700); e != nil {
		return r, e
	}
	root, e := os.OpenRoot(p)
	if e != nil {
		return r, e
	}
	defer root.Close()
	lock, e := root.OpenFile("lock", os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0600)
	if e != nil {
		return r, e
	}
	defer lock.Close()
	if e = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		return r, e
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	// Reserve both identities under the same lock before the first receipt
	// write. Any later failure leaves both exclusive guards consumed.
	var files []*os.File
	for _, k := range []string{id, scope} {
		f, e := root.OpenFile(k, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
		if e != nil {
			return r, fmt.Errorf("fresh process approval already consumed or unavailable")
		}
		defer f.Close()
		files = append(files, f)
	}
	data, _ := json.Marshal(r)
	for _, f := range files {
		n, e := f.Write(data)
		if e == nil && n != len(data) {
			e = fmt.Errorf("incomplete process claim")
		}
		if e == nil {
			e = f.Sync()
		}
		ce := f.Close()
		if e != nil {
			return r, e
		}
		if ce != nil {
			return r, ce
		}
	}
	for _, dir := range []string{p, filepath.Dir(p)} {
		f, e := os.Open(dir)
		if e != nil {
			return r, e
		}
		e = f.Sync()
		ce := f.Close()
		if e != nil {
			return r, e
		}
		if ce != nil {
			return r, ce
		}
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
