package release

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

const liveConsumer = "releasegate/native-live-v1"

type LiveConsumerReceipt struct {
	Schema           int               `json:"schema"`
	Consumer         string            `json:"consumer"`
	Pass             bool              `json:"pass"`
	ExitCode         int               `json:"exitCode"`
	Error            string            `json:"error,omitempty"`
	Source           ExecutionSource   `json:"source"`
	Bundle           string            `json:"bundle"`
	Inputs           map[string]string `json:"inputs"`
	ReportSHA256     string            `json:"reportSHA256"`
	Argv             []string          `json:"argv"`
	ExecutableSHA256 string            `json:"executableSHA256"`
	StartedAt        time.Time         `json:"startedAt"`
	FinishedAt       time.Time         `json:"finishedAt"`
	Files            map[string]string `json:"files"`
}

func liveInventory(dir string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic native input")
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		hash, err := recordHash(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(name)] = hash
		if len(files) > 100000 {
			return fmt.Errorf("native inventory exceeds limit")
		}
		return nil
	})
	return files, err
}

func liveReportHash(report supportmatrix.Report) string {
	raw, _ := json.Marshal(report)
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

// CaptureLiveConsumer is an explicitly invoked read-only consumer. The caller
// supplies a new output directory, outside the sealed native input bundle.
func CaptureLiveConsumer(root, bundle, out string, argv []string) (bool, error) {
	started := time.Now().UTC()
	proof := LiveConsumerReceipt{Schema: 1, Consumer: liveConsumer, ExitCode: 1, Bundle: bundle,
		Argv: argv, StartedAt: started, Files: map[string]string{}}
	var report supportmatrix.Report
	dir, err := recordPath(root, bundle)
	absoluteOut, outErr := filepath.Abs(out)
	if outErr != nil {
		return false, outErr
	}
	relativeOut, outErr := filepath.Rel(root, absoluteOut)
	if outErr != nil {
		return false, outErr
	}
	if !filepath.IsLocal(relativeOut) {
		return false, fmt.Errorf("consumer output must be repository-local")
	}
	if _, outErr = recordPath(root, filepath.ToSlash(filepath.Dir(relativeOut))); outErr != nil {
		return false, outErr
	}
	if err == nil {
		rel, e := filepath.Rel(dir, absoluteOut)
		if e != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return false, fmt.Errorf("consumer output must be outside native bundle")
		}
		proof.Source, err = currentRecordSource(root)
	}
	out = absoluteOut
	if e := os.Mkdir(out, 0700); e != nil {
		return false, e
	}
	if err == nil {
		proof.Inputs, err = liveInventory(dir)
	}
	if err == nil {
		report, err = qualifiedLive(root, dir)
	}
	if err == nil && report.GitCommit != proof.Source.Commit {
		err = fmt.Errorf("live source commit mismatch")
	}
	if err == nil {
		proof.Pass = true
		proof.ExitCode = 0
		proof.ReportSHA256 = liveReportHash(report)
	}
	executable, executableErr := os.Executable()
	if executableErr == nil {
		proof.ExecutableSHA256, executableErr = recordHash(executable)
	}
	if executableErr != nil {
		return false, executableErr
	}
	if err != nil {
		proof.Error = "native live consumer rejected evidence"
	}
	raw, marshalErr := json.MarshalIndent(proof, "", "  ")
	if marshalErr != nil {
		return false, marshalErr
	}
	// Capture actual output and exit separately. Input bytes are never modified.
	for name, content := range map[string][]byte{"stdout.jsonl": append(raw, '\n'), "stderr.log": []byte(proof.Error),
		"exit-code": []byte(fmt.Sprintf("%d\n", proof.ExitCode))} {
		path := filepath.Join(out, name)
		if e := os.WriteFile(path, content, 0600); e != nil {
			return false, e
		}
		proof.Files[name], marshalErr = recordHash(path)
		if marshalErr != nil {
			return false, marshalErr
		}
	}
	if _, e := os.Stdout.Write(append(raw, '\n')); e != nil {
		return false, e
	}
	if _, e := os.Stderr.Write([]byte(proof.Error)); e != nil {
		return false, e
	}
	proof.FinishedAt = time.Now().UTC()
	raw, marshalErr = json.MarshalIndent(proof, "", "  ")
	if marshalErr != nil {
		return false, marshalErr
	}
	if e := os.WriteFile(filepath.Join(out, "live-consumer.json"), append(raw, '\n'), 0600); e != nil {
		return false, e
	}
	return proof.Pass, nil
}

func verifyLiveConsumer(root, dir string, ref FileRef, source ExecutionSource, report supportmatrix.Report) error {
	path, err := verifyRecordFile(root, ref)
	if err != nil {
		return err
	}
	var proof LiveConsumerReceipt
	if recordJSON(path, &proof) != nil {
		return fmt.Errorf("invalid native consumer receipt")
	}
	bundle, err := filepath.Rel(root, dir)
	inputs, inputErr := liveInventory(dir)
	invocation := flag.NewFlagSet("native-live-consumer", flag.ContinueOnError)
	invocation.SetOutput(io.Discard)
	argBundle := invocation.String("live-bundle", "", "")
	argOut := invocation.String("out", "", "")
	validArgs := len(proof.Argv) > 0 && filepath.Base(proof.Argv[0]) == "releasegate"
	if validArgs {
		validArgs = invocation.Parse(proof.Argv[1:]) == nil && invocation.NArg() == 0
	}
	outputDir, _ := filepath.Abs(*argOut)

	if err != nil || inputErr != nil || proof.Schema != 1 || proof.Consumer != liveConsumer || !proof.Pass || proof.ExitCode != 0 || proof.Error != "" ||
		proof.Bundle != filepath.ToSlash(bundle) || proof.Source.Commit != source.Commit || !sameRecordSource(proof.Source, source) ||
		!reflect.DeepEqual(proof.Inputs, inputs) || proof.ReportSHA256 != liveReportHash(report) ||
		outputDir != filepath.Dir(path) || !validArgs || *argBundle != proof.Bundle ||
		len(proof.ExecutableSHA256) != 64 || proof.StartedAt.IsZero() || proof.FinishedAt.Before(proof.StartedAt) || len(proof.Files) != 3 {
		return fmt.Errorf("native consumer disagrees with actual input, result or invocation")
	}
	attempt := filepath.Dir(path)
	for _, name := range []string{"stdout.jsonl", "stderr.log", "exit-code"} {
		if _, err := verifyRecordFile(attempt, FileRef{name, proof.Files[name]}); err != nil {
			return err
		}
	}
	var output LiveConsumerReceipt
	if recordJSON(filepath.Join(attempt, "stdout.jsonl"), &output) != nil {
		return fmt.Errorf("consumer output invalid")
	}
	proof.Files = map[string]string{}
	proof.FinishedAt = time.Time{}
	if !reflect.DeepEqual(proof, output) {
		return fmt.Errorf("consumer output disagrees with receipt")
	}
	code, e := os.ReadFile(filepath.Join(attempt, "exit-code"))
	diagnostic, d := os.ReadFile(filepath.Join(attempt, "stderr.log"))
	if e != nil || d != nil || string(code) != "0\n" || len(diagnostic) != 0 {
		return fmt.Errorf("consumer process not passed")
	}
	return nil
}
