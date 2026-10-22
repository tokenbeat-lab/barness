package evidence

import (
	"crypto/sha256"
	"encoding/hex"
)

// Digest identifies immutable fixture and evidence bytes.
func Digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
