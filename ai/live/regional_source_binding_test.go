//go:build live

package live

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Injected when the explicitly reviewed source snapshot is built, e.g.
// -ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=HASH.
// A stale prebuilt binary cannot claim the source it merely sees at runtime.
var regionalBuildSource string

func regionalBinaryHash() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, io.LimitReader(f, 128<<20)); err != nil {
		return "", err
	}
	st, err := f.Stat()
	if err != nil || st.Size() > 128<<20 {
		return "", fmt.Errorf("regional executable exceeds bound")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
