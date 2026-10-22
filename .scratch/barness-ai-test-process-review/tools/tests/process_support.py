"""Public process fixtures; synthetic repositories never use retained evidence."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

TOOLS = Path(__file__).resolve().parents[1]
ARTIFACT = os.environ.get("BARNESS_PROCESS_TEST_ARTIFACT_DIR")


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def write_json(path, value):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, sort_keys=True, indent=2) + "\n")
    return path


class ProcessCase(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="barness-process-e2e-")
        self.addCleanup(self.temp.cleanup)
        self.repo = Path(self.temp.name)
        (self.repo / "ai").mkdir()
        (self.repo / "ai/example.go").write_text("package ai\n")
        (self.repo / "go.mod").write_text("module example.test/fixture\n\ngo 1.26.2\n")
        (self.repo / "state.md").write_text("Synthetic current status\n")
        self.call("git", "init", "-q")
        self.call("git", "add", ".")
        self.call("git", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test",
                  "commit", "-qm", "synthetic source")

    def call(self, *args, env=None, expected=None):
        child_env = {k: os.environ[k] for k in ("PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH") if k in os.environ}
        child_env.update(env or {})
        result = subprocess.run([str(x) for x in args], cwd=self.repo, env=child_env,
                                capture_output=True, text=True, timeout=40)
        if ARTIFACT:
            target = Path(ARTIFACT)
            target.mkdir(parents=True, exist_ok=True)
            with (target / "processes.jsonl").open("a") as out:
                out.write(json.dumps({"test": self.id(), "program": Path(str(args[0])).name,
                    "argv": [str(a).replace(str(self.repo), "[fixture]") for a in args[1:]],
                    "exitCode": result.returncode,
                    "stdoutSHA256": hashlib.sha256(result.stdout.encode()).hexdigest(),
                    "stderrSHA256": hashlib.sha256(result.stderr.encode()).hexdigest()}) + "\n")
        if expected is not None:
            self.assertEqual(result.returncode, expected, result.stderr + result.stdout)
        return result

    def run_stage(self, mode="valid", **kwargs):
        return self.call(sys.executable, TOOLS / "run.py", "--feature", "sample", "--stage", "e2e",
                         "--timeout", "0.3" if mode == "timeout" else "15", "--",
                         sys.executable, TOOLS / "tests/fixture_child.py", mode, **kwargs)

    def attempts(self):
        return sorted((self.repo / ".evidence/sample/e2e").glob("*/receipt.json"))
