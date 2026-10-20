package minimaxlive

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Historical source has its own committed archive. Never call the current-source
// verifier here: changed authority/harness code is precisely why recovery4 exists.
func closedHistoricalSource(root *os.Root, files map[string]string, collection string, g RenewalGrant) error {
	var archive struct {
		SourceCommit string            `json:"sourceCommit"`
		Files        map[string]string `json:"files"`
	}
	name := collection + "source-at-start.json"
	raw, e := cohortJSON(root, name, &archive)
	if e != nil || files[name] != Hash(raw) || Hash(raw) != g.SourceManifestSHA256 || archive.SourceCommit != g.SourceCommit || len(archive.Files) == 0 || len(archive.Files) > 8192 {
		return fmt.Errorf("closed third complete source archive differs: %v", e)
	}
	return committedArchive(archive.SourceCommit, archive.Files)
}
func committedArchive(commit string, files map[string]string) error {
	if len(commit) != 40 || strings.Trim(commit, "0123456789abcdef") != "" {
		return fmt.Errorf("closed source commit invalid")
	}
	names := make([]string, 0, len(files))
	for name, pin := range files {
		if filepath.IsAbs(name) || filepath.ToSlash(filepath.Clean(name)) != name || strings.Contains(name, "\\") || strings.ContainsAny(name, "\n\x00") || strings.HasPrefix(name, "-") || strings.HasPrefix(name, "../") || name == ".env" || strings.HasSuffix(name, "/.env") || !regexpHex64(pin) {
			return fmt.Errorf("closed source path/pin invalid")
		}
		names = append(names, name)
	}
	sort.Strings(names)
	owner, e := Root()
	if e != nil {
		return e
	}
	args := append([]string{"archive", "--format=tar", commit, "--"}, names...)
	c := exec.Command("git", args...)
	c.Dir = owner
	c.Env = []string{}
	stream, e := c.StdoutPipe()
	if e != nil {
		return e
	}
	if e = c.Start(); e != nil {
		return fmt.Errorf("closed source commit unavailable")
	}
	done := false
	defer func() {
		if !done {
			c.Process.Kill()
			c.Wait()
		}
	}()
	tr := tar.NewReader(stream)
	seen := map[string]bool{}
	total := int64(0)
	entries := 0
	for {
		entry, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("closed source archive invalid")
		}
		entries++
		if entries > 8192 {
			return fmt.Errorf("closed source entry ceiling")
		}
		if entry.Typeflag == tar.TypeDir || entry.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if entry.Typeflag != tar.TypeReg || entry.Size < 0 || entry.Size > maxCohortFile || total+entry.Size > maxCohortBytes || seen[entry.Name] || files[entry.Name] == "" {
			return fmt.Errorf("closed source archive file selection invalid")
		}
		h := sha256.New()
		n, err := io.Copy(h, io.LimitReader(tr, entry.Size+1))
		if err != nil || n != entry.Size || hex.EncodeToString(h.Sum(nil)) != files[entry.Name] {
			return fmt.Errorf("closed source differs from original commit")
		}
		seen[entry.Name] = true
		total += n
	}
	e = c.Wait()
	done = true
	if e != nil || len(seen) != len(files) {
		return fmt.Errorf("closed source committed archive incomplete")
	}
	return nil
}
