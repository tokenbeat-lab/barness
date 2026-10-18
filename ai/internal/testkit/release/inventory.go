package release

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

// Case.Artifacts is recursive (evidence.Run.Finish), including child case
// directories. These are import ceilings, not generation/vendor limits.
const (
	maxInventoryFiles   = 512
	maxInventoryEntries = 2048
	maxInventoryDepth   = 24
	maxInventoryFile    = 64 << 20
	maxInventoryBytes   = 256 << 20
)

func canonicalArtifactPath(name string) bool {
	return name != "." && fs.ValidPath(name) && filepath.IsLocal(name) && !strings.Contains(name, `\`)
}

// Live cases have one directory component. Offline import keeps its already
// opened bundle root through every ancestor instead of reopening a checked path.
func verifyArtifactInventory(dir string, artifacts map[string]string) error {
	root, err := os.OpenRoot(filepath.Dir(dir))
	if err != nil {
		return fmt.Errorf("case artifact directory unavailable")
	}
	defer root.Close()
	child, err := openInventoryCase(root, filepath.Base(dir))
	if err != nil {
		return err
	}
	defer child.Close()
	return verifyCaseArtifactInventory(child, artifacts)
}

// Open each component relative to a pinned descriptor. Root prevents escape
// during replacement; Lstat/SameFile also refuses internal symlink aliases.
func openInventoryCase(root *os.Root, name string) (*os.Root, error) {
	if !canonicalArtifactPath(name) || len(strings.Split(name, "/")) > maxInventoryDepth {
		return nil, fmt.Errorf("invalid case artifact directory")
	}
	parent := root
	for _, part := range strings.Split(name, "/") {
		child, err := openInventoryDirectory(parent, part)
		if parent != root {
			parent.Close()
		}
		if err != nil {
			return nil, err
		}
		parent = child
	}
	return parent, nil
}

func openInventoryDirectory(parent *os.Root, name string) (*os.Root, error) {
	before, err := parent.Lstat(name)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("case artifact directory missing: %w", fs.ErrNotExist)
	}
	if err != nil || !before.IsDir() {
		return nil, fmt.Errorf("case artifact directory is not a regular directory")
	}
	child, err := parent.OpenRoot(name)
	if err != nil {
		return nil, fmt.Errorf("case artifact directory unavailable")
	}
	opened, e1 := child.Stat(".")
	after, e2 := parent.Lstat(name)
	if e1 != nil || e2 != nil || !after.IsDir() || !os.SameFile(before, opened) || !os.SameFile(before, after) {
		child.Close()
		return nil, fmt.Errorf("case artifact directory changed")
	}
	return child, nil
}

func verifyCaseArtifactInventory(root *os.Root, artifacts map[string]string) error {
	if len(artifacts) == 0 || len(artifacts) > maxInventoryFiles {
		return fmt.Errorf("case artifact integrity inventory missing or over file limit")
	}
	dirs := map[string]bool{}
	for name, hash := range artifacts {
		if !canonicalArtifactPath(name) || len(strings.Split(name, "/")) > maxInventoryDepth || len(hash) != 2*sha256.Size || len(audit.File("identifiers.txt", []byte(name), audit.Environment())) != 0 {
			return fmt.Errorf("artifact identifier or path invalid")
		}
		if _, err := hex.DecodeString(hash); err != nil {
			return fmt.Errorf("artifact hash invalid")
		}
		for prefix, _, ok := strings.Cut(name, "/"); ok; {
			dirs[prefix] = true
			start := len(prefix) + 1
			next, _, more := strings.Cut(name[start:], "/")
			prefix, ok = prefix+"/"+next, more
		}
	}
	files, entries := 0, 0
	var total int64
	var walk func(*os.Root, string) error
	walk = func(dir *os.Root, prefix string) error {
		listing, err := dir.Open(".")
		if err != nil {
			return fmt.Errorf("case artifact directory unavailable")
		}
		defer listing.Close()
		for {
			batch, readErr := listing.ReadDir(1)
			if readErr != nil && readErr != io.EOF {
				return fmt.Errorf("case artifact enumeration failed")
			}
			for _, entry := range batch {
				entries++
				if entries > maxInventoryEntries {
					return fmt.Errorf("case artifact directory entry limit exceeded")
				}
				name := prefix + entry.Name()
				if entry.IsDir() {
					if !dirs[name] {
						return fmt.Errorf("case contains an unlisted artifact directory")
					}
					child, err := openInventoryDirectory(dir, entry.Name())
					if err != nil {
						return err
					}
					err = walk(child, name+"/")
					child.Close()
					if err != nil {
						return err
					}
					continue
				}
				if !entry.Type().IsRegular() || artifacts[name] == "" {
					return fmt.Errorf("case contains an unlisted or non-regular artifact")
				}
				n, err := hashInventoryArtifact(dir, entry.Name(), artifacts[name], maxInventoryBytes-total)
				if err != nil {
					return err
				}
				total += n
				files++
			}
			if readErr == io.EOF {
				return nil
			}
		}
	}
	if err := walk(root, ""); err != nil {
		return err
	}
	if files != len(artifacts) {
		return fmt.Errorf("case artifact inventory is incomplete")
	}
	return nil
}

func hashInventoryArtifact(root *os.Root, name, want string, remaining int64) (int64, error) {
	limit := min(int64(maxInventoryFile), remaining)
	before, err := root.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Size() > limit {
		return 0, fmt.Errorf("case artifact missing, non-regular or over byte limit")
	}
	// Nonblocking open also bounds a regular-file → FIFO replacement race.
	f, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return 0, fmt.Errorf("case artifact unavailable")
	}
	defer f.Close()
	opened, e1 := f.Stat()
	after, e2 := root.Lstat(name)
	if e1 != nil || e2 != nil || !opened.Mode().IsRegular() || !after.Mode().IsRegular() || !os.SameFile(before, opened) || !os.SameFile(before, after) || opened.Size() > limit {
		return 0, fmt.Errorf("case artifact changed or over byte limit")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, limit+1))
	if err != nil || n > limit || hex.EncodeToString(h.Sum(nil)) != want {
		return 0, fmt.Errorf("case artifact missing, changed or over byte limit")
	}
	return n, nil
}
