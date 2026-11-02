package maintenance

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

// The collector owns and cancels its native process group. Forward TERM and
// allow its bounded shutdown to finish before CommandContext force-stops it.
func checkProcess(ctx context.Context, root, dir, name string, argv []string) ([]byte, error) {
	command, err := jsonBytes(map[string]any{"argv": argv})
	if err != nil {
		return nil, err
	}
	if err := writeImmutable(filepath.Join(dir, name+"-command.json"), command); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 8 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exit := -1
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	for filename, raw := range map[string][]byte{name + "-stdout.json": stdout.Bytes(), name + "-stderr.log": stderr.Bytes(), name + "-exit-code": []byte(jsonNumber(exit) + "\n")} {
		if err := writeImmutable(filepath.Join(dir, filename), raw); err != nil {
			return stdout.Bytes(), err
		}
	}
	return stdout.Bytes(), runErr
}
func jsonNumber(value int) string { raw, _ := json.Marshal(value); return string(raw) }
func sortedCheckKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for name := range values {
		keys = append(keys, name)
	}
	slices.Sort(keys)
	return keys
}

type checkInspection struct {
	CandidateID    string   `json:"candidateId"`
	BaseDigest     string   `json:"baseDigest"`
	Status         string   `json:"status"`
	Conclusion     string   `json:"conclusion"`
	MeasuredCommit string   `json:"measuredCommit"`
	RequiredCases  []string `json:"requiredCases"`
	ReceiptSHA256  string   `json:"receiptSHA256"`
	SnapshotSHA256 string   `json:"snapshotSHA256"`
}

func inspectCheck(ctx context.Context, root, dir, prepared, receipt string) (checkInspection, error) {
	var result checkInspection
	argv := []string{"python3", "tools/verification/catalog_check.py", "inspect", "--prepared", prepared, "--receipt", receipt}
	var raw []byte
	var err error
	if dir != "" {
		raw, err = checkProcess(ctx, root, dir, "inspect", argv)
	} else {
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=")
		raw, err = cmd.Output()
	}
	if err != nil || json.Unmarshal(raw, &result) != nil || result.Status != "applicable" || result.Conclusion != "code_compatible" || len(result.RequiredCases) == 0 {
		return result, failure("check_invalid", "controlled check evidence is missing, changed, incomplete, or stale")
	}
	return result, nil
}
