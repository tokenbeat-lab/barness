package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecisionsEvidence(t *testing.T) {
	base := filepath.Join("..", "..", ".scratch", "barness-ai-openai-decisions", "text-choice-evidence")
	for _, name := range []string{"intact", "missing-archive", "changed-archive", "rewritten-manifest"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD32-evidence-"+name)
			if _, err := os.Stat(filepath.Join(base, "verify.py")); err != nil {
				ev.Check("verifier available", false, "text evidence verifier absent")
				return
			}
			dir := base
			if name != "intact" {
				dir = t.TempDir()
				for _, file := range []string{"manifest.json", "text-choice.zip"} {
					raw, err := os.ReadFile(filepath.Join(base, file))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, file), raw, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				switch name {
				case "missing-archive":
					if err := os.Remove(filepath.Join(dir, "text-choice.zip")); err != nil {
						t.Fatal(err)
					}
				case "changed-archive":
					if err := os.WriteFile(filepath.Join(dir, "text-choice.zip"), []byte("changed evidence"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "rewritten-manifest":
					if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"status":"PASS"}`), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			out, err := exec.CommandContext(ctxFor(t), "python3", filepath.Join(base, "verify.py"), "--bundle", dir).CombinedOutput()
			ev.Record("verification", map[string]any{"output": string(out), "succeeded": err == nil})
			if name == "intact" {
				ev.Check("recorded evidence intact", err == nil && strings.Contains(string(out), "no live execution"), "got %s", out)
			} else {
				ev.Check("missing or rewritten evidence rejected", err != nil && strings.Contains(string(out), "FAIL:"), "got %s", out)
			}
		})
	}
}

// A filesystem-boundary race must not replace the archive used by --audit
// after its contents passed validation. All writes stay in this test's temp
// directory; the child runs the public CLI, not a private parser entry.
func TestDecisionsEvidenceSnapshot(t *testing.T) {
	ev := run.Case(t, "P10-OD32-evidence-snapshot-race")
	base := filepath.Join("..", "..", ".scratch", "barness-ai-openai-decisions", "text-choice-evidence")
	dir := t.TempDir()
	for _, name := range []string{"manifest.json", "text-choice.zip"} {
		raw, err := os.ReadFile(filepath.Join(base, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	script := `import io,json,pathlib,runpy,sys,tempfile,zipfile
verifier,folder=sys.argv[1:]
folder=pathlib.Path(folder)
tempfile.tempdir=str(folder)
evil=io.BytesIO()
with zipfile.ZipFile(evil,'w') as z:
    z.writestr('../escaped.txt','unvalidated bytes')
old_exit=zipfile.ZipFile.__exit__
replaced=False
def swap(self,*args):
    global replaced
    result=old_exit(self,*args)
    if not replaced:
        (folder/'text-choice.zip').write_bytes(evil.getvalue())
        replaced=True
    return result
zipfile.ZipFile.__exit__=swap
sys.argv=[verifier,'--bundle',str(folder),'--audit']
try:
    runpy.run_path(verifier,run_name='__main__')
except SystemExit as e:
    code=e.code
print(json.dumps({'replaced':replaced,'escaped':(folder/'escaped.txt').exists(),'code':code}))
`
	out, err := exec.CommandContext(ctxFor(t), "python3", "-c", script, filepath.Join(base, "verify.py"), dir).CombinedOutput()
	ev.Record("verification", map[string]any{"output": string(out), "succeeded": err == nil})
	ev.Check("replacement exercised", strings.Contains(string(out), `"replaced": true`), "got %s", out)
	ev.Check("audit uses validated immutable bytes", err == nil && strings.Contains(string(out), `"escaped": false`) && strings.Contains(string(out), `"code": 0`), "got %s", out)
}

func TestDecisionsEvidenceManifestBudget(t *testing.T) {
	ev := run.Case(t, "P10-OD32-evidence-manifest-budget")
	base := filepath.Join("..", "..", ".scratch", "barness-ai-openai-decisions", "text-choice-evidence")
	dir := t.TempDir()
	file, err := os.Create(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Sparse input is larger than the advertised manifest budget. The
	// filesystem double rejects unbounded reads before they allocate it.
	if err := file.Truncate(128 << 20); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	script := `import pathlib,runpy,sys
verifier,folder=sys.argv[1:]
original=pathlib.Path.open
class BudgetReader:
    def __init__(self,source): self.source=source
    def __enter__(self): self.source.__enter__(); return self
    def __exit__(self,*args): return self.source.__exit__(*args)
    def read(self,size=-1):
        if size<0 or size>16<<20: raise MemoryError('unbounded artifact read')
        return self.source.read(size)
def open_bound(self,*args,**kwargs):
    source=original(self,*args,**kwargs)
    return BudgetReader(source) if self==pathlib.Path(folder)/'manifest.json' else source
pathlib.Path.open=open_bound
sys.argv=[verifier,'--bundle',folder]
runpy.run_path(verifier,run_name='__main__')
`
	out, err := exec.CommandContext(ctxFor(t), "python3", "-c", script, filepath.Join(base, "verify.py"), dir).CombinedOutput()
	ev.Record("verification", map[string]any{"output": string(out), "failed": err != nil})
	ev.Check("bounded public rejection before allocation", err != nil && strings.Contains(string(out), "FAIL:") && !strings.Contains(string(out), "Traceback"), "got %s", out)
}
