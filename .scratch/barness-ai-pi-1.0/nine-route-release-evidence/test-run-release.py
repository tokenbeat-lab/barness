#!/usr/bin/env python3
"""Controlled process replay: no keys, network or real Provider calls."""
import json
import os
from pathlib import Path
import subprocess
import shutil
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
RUNNER = Path(__file__).with_name('run-release.py')


class ReleaseRunner(unittest.TestCase):
    def test_missing_key_file_still_produces_no_key_evidence(self):
        with tempfile.TemporaryDirectory(dir=ROOT / '.evidence') as base:
            root = Path(base)
            scripts = root / '.scratch/feature/evidence'
            scripts.mkdir(parents=True)
            script = scripts / 'run-live.py'
            shutil.copyfile(RUNNER.with_name('run-live.py'), script)
            binary = root / 'controlled'
            binary.write_text("#!/usr/bin/env python3\nimport os\nfrom pathlib import Path\np=Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'controlled'\np.mkdir(parents=True)\n(p/'live-report.json').write_text('{}')\n")
            binary.chmod(0o755)
            (root / 'ai/live').mkdir(parents=True)
            out = root / 'result'
            child = subprocess.run(['python3', str(script), '--live', '--combo', 'openai-responses', '--binary', str(binary), '--out', str(out), '--alias', 'controlled@offline'], cwd=root, capture_output=True)
            self.assertEqual(child.returncode, 0, 'missing key file aborted the no-key harness')
            self.assertFalse(json.loads((out / 'command.json').read_text())['keyPresent'])

    def test_absolute_output_is_isolated_and_portable(self):
        with tempfile.TemporaryDirectory(dir=ROOT / '.evidence') as base:
            base = Path(base)
            stub = base / 'bin'
            stub.mkdir()
            go = stub / 'go'
            go.write_text('''#!/usr/bin/env python3
import json, os, sys
from pathlib import Path
a = sys.argv[1:]
if a[0] == 'test':
    p = Path(a[a.index('-o')+1])
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text("#!/usr/bin/env python3\\nimport os\\nfrom pathlib import Path\\np=Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'controlled'\\np.mkdir(parents=True)\\n(p/'live-report.json').write_text('{}')\\n")
    p.chmod(0o755)
    Path(os.environ['BUILD_RECORD']).write_text(str(p))
elif 'releasegate' in a[1]:
    print(Path.cwd(), Path.home())
''')
            go.chmod(0o755)
            out = base / 'result'
            record = base / 'build-path'
            # The runner intentionally allowlists env. Encode the record path
            # in the controlled executable rather than weakening that policy.
            go.write_text(go.read_text().replace("os.environ['BUILD_RECORD']", repr(str(record))))
            env = dict(os.environ, PATH=str(stub)+os.pathsep+os.environ['PATH'])
            child = subprocess.run(['python3', str(RUNNER), '--out', str(out)], cwd=ROOT, env=env, capture_output=True)
            self.assertEqual(child.returncode, 0, 'controlled runner failed')
            self.assertTrue(Path(record.read_text()).is_relative_to(out), 'harness executable shared between invocations')
            for name in ('smoke-commands.json', 'gate-command.json', 'gate.log'):
                data = (out / name).read_text()
                self.assertFalse(str(ROOT) in data or str(Path.home()) in data, 'environment path leaked')
            argv = json.loads((out / 'gate-command.json').read_text())['argv']
            self.assertFalse(any(Path(s).is_absolute() for s in argv), 'replay command is not portable')


if __name__ == '__main__':
    unittest.main()
