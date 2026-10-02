// Package requestid mints the RequestIDs barness-ai's examples give their
// logical calls.
//
// A RequestID names one logical call and must never repeat, across tenants
// and calls alike (ADR-0001, GLOSSARY "Logical Call"): an injected admission
// or Observer may key on it alone. barness-ai cannot check that, so the host
// guarantees it. 128 random bits make a collision negligible without any
// coordination between processes; a host with its own globally unique
// identifiers (a database sequence, a ULID service) may use those instead.
package requestid

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a fresh RequestID. It is never derived from a tenant, session or
// turn: a retry inside barness-ai keeps the call's RequestID, and every new
// logical call — a tool round's next turn included — takes a new one.
func New() string {
	var b [16]byte
	// crypto/rand.Read never returns an error on supported platforms and
	// aborts the program if the system source fails (Go 1.24+).
	_, _ = rand.Read(b[:])
	return "req-" + hex.EncodeToString(b[:])
}
