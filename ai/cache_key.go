package ai

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// cacheScope is what a call's cache and affinity identifiers are derived
// from besides the caller's session id: tenant, actual Provider and vendor
// account, pinned from the trusted binding snapshot (ADR-0049).
type cacheScope struct {
	tenant   string
	provider ProviderID
	account  string
}

// key derives the identifier sent for session: 64 hex characters, the same
// for the same tenant, Provider, account and session. Credential rotation
// does not change it. Every consumer moves to v2 together; fields
// are length-prefixed so no two inputs share an encoding.
func (s cacheScope) key(session string) string {
	h := sha256.New()
	h.Write([]byte("barness-ai/prompt-cache-key/v2"))
	for _, f := range []string{s.tenant, string(s.provider), s.account, session} {
		h.Write(binary.BigEndian.AppendUint64(nil, uint64(len(f))))
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
