package e2e

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const xaiPrePolicyCommit = "9584d3d6f095aefa4ab73bdc75c39a705b7931d6"

var xaiHistoricalOnce sync.Once
var xaiHistoricalRoot string
var xaiHistoricalError error

// One retained tree serves all synthetic packets in this test process. It is
// exported from Git, never assembled from current files or used as a current
// vendor qualification. No global working directory or environment changes.
func imageQualificationHistoricalRoot(t *testing.T) string {
	t.Helper()
	xaiHistoricalOnce.Do(func() { xaiHistoricalRoot, xaiHistoricalError = exportImageHistoricalSource(t) })
	if xaiHistoricalError != nil {
		t.Fatalf("historical source: %v", xaiHistoricalError)
	}
	return xaiHistoricalRoot
}

func exportImageHistoricalSource(t *testing.T) (string, error) {
	t.Helper()
	actual, err := filepath.Abs("../..")
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(filepath.Dir(run.Dir()))
	root := filepath.Join(parent, "xai-images-historical-9584")
	if err = os.Mkdir(root, 0700); err != nil {
		return "", err
	}
	archive := filepath.Join(parent, "xai-images-historical-9584.tar")
	file, err := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctxFor(t), "git", "archive", "--format=tar", xaiPrePolicyCommit, "ai", "go.mod", "go.sum")
	cmd.Dir = actual
	cmd.Env = regionalChildEnvironment()
	cmd.Stdout = file
	stderr, err := cmd.StderrPipe()
	if err != nil {
		file.Close()
		return "", err
	}
	if err = cmd.Start(); err != nil {
		file.Close()
		return "", err
	}
	message, _ := io.ReadAll(stderr)
	err = cmd.Wait()
	closeErr := file.Close()
	if err != nil {
		return "", fmt.Errorf("git archive: %v %s", err, message)
	}
	if closeErr != nil {
		return "", closeErr
	}
	in, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	defer in.Close()
	reader := tar.NewReader(in)
	inventory := map[string]string{}
	for {
		header, e := reader.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", e
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		name := filepath.Clean(header.Name)
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("unsafe archive path")
		}
		target := filepath.Join(root, name)
		switch header.Typeflag {
		case tar.TypeDir:
			if e = os.MkdirAll(target, 0700); e != nil {
				return "", e
			}
		case tar.TypeReg, tar.TypeRegA:
			if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
				return "", e
			}
			out, e := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				return "", e
			}
			h := sha256.New()
			_, e = io.Copy(io.MultiWriter(out, h), reader)
			c := out.Close()
			if e != nil {
				return "", e
			}
			if c != nil {
				return "", c
			}
			inventory[filepath.ToSlash(name)] = hex.EncodeToString(h.Sum(nil))
		default:
			return "", fmt.Errorf("unsupported archive entry %s", name)
		}
	}
	// Product bytes stay read-only; synthetic packets live outside ai/.
	var directories []string
	if err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			directories = append(directories, path)
			return nil
		}
		return os.Chmod(path, 0400)
	}); err != nil {
		return "", err
	}
	for i := len(directories) - 1; i >= 0; i-- {
		if err = os.Chmod(directories[i], 0500); err != nil {
			return "", err
		}
	}
	source, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, err = io.Copy(h, source)
	source.Close()
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(inventory)
	if err != nil {
		return "", err
	}
	ev := run.Case(t, "XAIIMG-historical-source-context")
	ev.Fixture("historical-tree-inventory.json", raw)
	relative, _ := filepath.Rel(actual, archive)
	ev.Record("historical-export", map[string]any{"argv": cmd.Args, "exitCode": 0, "commit": xaiPrePolicyCommit, "archive": filepath.ToSlash(relative), "archiveSHA256": hex.EncodeToString(h.Sum(nil)), "treeFiles": len(inventory), "notCurrentQualification": true})
	return root, nil
}

func TestXAIImagesCurrentPolicySourceRefusal(t *testing.T) {
	f := newImageQualificationFixture(t)
	p := f.packet(t)
	ev := run.Case(t, "XAIIMG-current-policy-source-refusal")
	cmd := exec.CommandContext(p.ctx, p.binary, "-inspect-bundle", p.bundle, "-mode", "live", "-stage", "formal")
	cmd.Dir = catalogSnapshotRoot(t)
	cmd.Env = regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	code := imagineProcessExit(err)
	ev.Record("current-source-inspection", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(raw), "vendorCalls": 0, "notActualLive": true})
	ev.Check("current policy source refuses inherited image qualification", code != 0 && strings.Contains(string(raw), "route product input changed:"), "exit=%d output=%s", code, raw)
}
