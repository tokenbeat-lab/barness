// This fixture produces native-shaped evidence only inside a ProcessCase's
// independent synthetic Git repository. It has no credentials or HTTP client.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func main() {
	binary := flag.String("binary", "", "synthetic consumer binary")
	output := flag.String("out", "", "new private evidence bundle")
	flag.Parse()
	if err := writeFixture(*binary, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeFixture(binary, output string) error {
	root, err := os.Getwd()
	if err != nil || !strings.HasPrefix(filepath.Base(root), "barness-process-e2e-") {
		return fmt.Errorf("independent ProcessCase repository required")
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil || physical != root {
		return fmt.Errorf("physical private repository required")
	}
	git := filepath.Join(root, ".git")
	info, err := os.Lstat(git)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("private Git directory required")
	}
	common, err := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil || strings.TrimSpace(string(common)) != git {
		return fmt.Errorf("external Git common directory refused")
	}
	out, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, out)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("output must belong to private repository")
	}
	for parent := out; parent != root; parent = filepath.Dir(parent) {
		info, e := os.Lstat(parent)
		if e == nil && info.Mode()&os.ModeSymlink != 0 || e != nil && !os.IsNotExist(e) {
			return fmt.Errorf("symbolic output ancestor refused")
		}
	}
	hash, err := livegrant.FileHash(binary)
	if err != nil {
		return err
	}
	scope, err := supportmatrix.ProcessScope("google-gemini", "live", "formal", hash)
	if err != nil {
		return err
	}
	// The live/formal shape tests the strict read-only consumer. The guards
	// above make these claims unusable in the real checkout or its common store.
	started := time.Now().UTC()
	grant := livegrant.Grant{Schema: 1, ID: "synthetic-google-" + strconv.FormatInt(started.UnixNano(), 10), AccountAlias: "synthetic", RegionAlias: "synthetic-region", ExpiresAt: started.Add(time.Hour), Scope: scope}
	raw, err := json.Marshal(grant)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(out, 0700); err != nil {
		return err
	}
	receipt, err := livegrant.Claim(grant, raw, "synthetic-private-consumer-fixture", started)
	if err != nil {
		return err
	}
	claim, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	for name, data := range map[string][]byte{"process-grant.json": raw, "process-authorization-receipt.json": claim} {
		file, e := os.OpenFile(filepath.Join(out, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		_, e = file.Write(data)
		closeErr := file.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"sourceSha256": scope.SourceSHA256, "startedAt": started, "finishedAt": time.Now().UTC(), "sdk": scope.SDK})
}
