package ai

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// cacheScope is what a call's cache and affinity identifiers are derived
// from besides the caller's session id: the tenant and the vendor account
// serving the call (spec §9: never the raw session id or TenantID).
type cacheScope struct {
	tenant  string
	account string
}

// key derives the identifier sent for session: 64 hex characters, the same
// for the same tenant, account and session and unrelated otherwise. Fields
// are length-prefixed so no two inputs share an encoding.
func (s cacheScope) key(session string) string {
	h := sha256.New()
	h.Write([]byte("barness-ai/prompt-cache-key/v1"))
	for _, f := range []string{s.tenant, s.account, session} {
		h.Write(binary.BigEndian.AppendUint64(nil, uint64(len(f))))
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
