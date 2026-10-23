"""Public collector -> independent release CLI source handoff after relocation."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

TOOLS = Path(__file__).resolve().parents[1]
PROJECT = TOOLS.parents[1]


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


class RelocatedToolProcess(unittest.TestCase):
    def test_public_source_handoff_detects_all_relocated_python_mutations(self):
        with tempfile.TemporaryDirectory(prefix='barness-relocated-tools-') as temp:
            repo = Path(temp)
            env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH') if k in os.environ}
            env.update(GOENV='off', GOFLAGS='', PYTHONDONTWRITEBYTECODE='1')

            def call(argv, cwd=repo):
                result = subprocess.run([str(a) for a in argv], cwd=cwd, env=env,
                                        capture_output=True, text=True, timeout=180)
                artifact = os.environ.get('BARNESS_PROCESS_TEST_ARTIFACT_DIR')
                if artifact:
                    out = Path(artifact); out.mkdir(parents=True, exist_ok=True)
                    with (out / 'migration-processes.jsonl').open('a') as log:
                        log.write(json.dumps({'test': self.id(), 'argv': [str(a).replace(str(repo), '[fixture]') for a in argv],
                            'exitCode': result.returncode, 'stdoutSHA256': hashlib.sha256(result.stdout.encode()).hexdigest(),
                            'stderrSHA256': hashlib.sha256(result.stderr.encode()).hexdigest()}) + '\n')
                return result

            (repo / 'ai/internal/testkit/jevformal').mkdir(parents=True)
            (repo / 'ai/example.go').write_text('package ai\n')
            jev = repo / 'ai/internal/testkit/jevformal/verify_protocol.py'
            jev.write_text('# synthetic protocol source\n')
            (repo / 'go.mod').write_text('module example.test/source\n\ngo 1.26.2\n')
            (repo / 'state.md').write_text('Synthetic NOT_RUN state\n')
            (repo / '.gitignore').write_text('ignored_source.py\n')
            copied = repo / 'tools/verification'
            shutil.copytree(TOOLS, copied, ignore=shutil.ignore_patterns('__pycache__'))
            for command in (['git', 'init', '-q'], ['git', 'add', '.'],
                ['git', '-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '-qm', 'source']):
                self.assertEqual(call(command).returncode, 0)
            ignored = copied / 'ignored_source.py'
            ignored.write_text('# ignored Python is still executable input\n')
            collected = call([sys.executable, copied / 'run.py', '--feature', 'migration', '--stage', 'e2e', '--',
                              sys.executable, copied / 'tests/fixture_child.py', 'valid'])
            self.assertEqual(collected.returncode, 0, collected.stdout + collected.stderr)
            receipt = json.loads((repo / json.loads(collected.stdout)['attempt'] / 'receipt.json').read_text())
            self.assertEqual(receipt['source']['files']['tools/verification/run.py'], digest(copied / 'run.py'))
            self.assertEqual(receipt['source']['files']['tools/verification/ignored_source.py'], digest(ignored))
            self.assertEqual(receipt['source']['files']['ai/internal/testkit/jevformal/verify_protocol.py'], digest(jev))

            binary = repo / 'releasegate'
            built = call(['go', 'build', '-o', binary, './ai/release/cmd/releasegate'], cwd=PROJECT)
            self.assertEqual(built.returncode, 0, built.stderr)
            record = {'schema': 1, 'feature': 'migration', 'currentState': {'path': 'state.md', 'sha256': digest(repo / 'state.md')},
                'source': receipt['source'], 'selectedInputs': {},
                'impact': {'liveRoutes': [], 'pressureRequired': False, 'reason': 'synthetic source handoff'},
                'e2e': {'status': 'NOT_RUN', 'reason': 'source handoff only'}, 'race': {'status': 'NOT_RUN', 'reason': 'not selected'},
                'live': {'status': 'NOT_APPLICABLE', 'reason': 'no provider change'},
                'pressure': {'status': 'NOT_APPLICABLE', 'reason': 'no capacity change'}}
            (repo / 'record.json').write_text(json.dumps(record))
            counter = 0

            def source_gate(expected):
                nonlocal counter
                counter += 1
                out = repo / ('report-' + str(counter))
                got = call([binary, '-record', 'record.json', '-out', out])
                self.assertEqual(got.returncode, 1, got.stdout + got.stderr)
                report = json.loads((out / 'release-report.json').read_text())
                self.assertFalse(report['pass'], 'source checks never grant omitted verification stages')
                check = next(c for c in report['checks'] if c['id'] == 'source')
                self.assertEqual(check['pass'], expected, report)

            source_gate(True)
            # Both independent selectors must observe path-set and content mutations.
            for target in (ignored, copied / 'tests/fixture_child.py', jev):
                original = target.read_bytes()
                with self.subTest(path=str(target.relative_to(repo)), mutation='modified'):
                    target.write_bytes(original + b'# changed\n'); source_gate(False)
                target.write_bytes(original)
                with self.subTest(path=str(target.relative_to(repo)), mutation='deleted'):
                    target.unlink(); source_gate(False)
                target.write_bytes(original)
            added = copied / 'new_source.py'; added.write_text('# added\n')
            source_gate(False); added.unlink(); source_gate(True)


if __name__ == '__main__':
    unittest.main(verbosity=2)
