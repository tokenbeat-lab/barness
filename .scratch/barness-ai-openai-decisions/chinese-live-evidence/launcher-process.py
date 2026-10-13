#!/usr/bin/env python3
"""Public launcher regression: rejected transport secrets must never persist."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import zipfile
import shutil


base = Path(__file__).resolve().parent
root = base.parents[2]
phase = sys.argv[1] if len(sys.argv) == 2 else 'check'
if phase not in ('red', 'green', 'check'):
    raise SystemExit('phase must be red, green or check')
results = []
with tempfile.TemporaryDirectory(prefix='barness-launcher-process-') as temporary:
    sandbox = Path(temporary)
    # A source-only checkout has no dotenv, vendor credentials, or Git history.
    # The former launcher builds, rejects through Go, and then leaks its proxy
    # argument; the fixed launcher must reject even before trying to build.
    with zipfile.ZipFile(base / 'execution-source.zip') as source:
        source.extractall(sandbox)
    shutil.copyfile(root / 'ai/live/support-matrix.json', sandbox / 'ai/live/support-matrix.json')
    target = sandbox / '.scratch/barness-ai-openai-decisions/chinese-live-evidence/run.py'
    target.parent.mkdir(parents=True)
    target.write_bytes((base / 'run.py').read_bytes())
    # Supply a deterministic Git identity without relying on a parent repo.
    subprocess.run(['git', 'init', '-q', str(sandbox)], check=True)
    subprocess.run(['git', '-C', str(sandbox), '-c', 'user.name=Process Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '--allow-empty', '-qm', 'fixture'], check=True)
    env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH') if k in os.environ}
    for name, proxy in (
        ('userinfo', 'http://name:launcher-sensitive-fixture@127.0.0.1:7890'),
        ('query', 'http://127.0.0.1:7890?value=launcher-sensitive-fixture'),
        ('fragment', 'http://127.0.0.1:7890#launcher-sensitive-fixture'),
    ):
        run = subprocess.run([sys.executable, str(target), '--live', '--proxy', proxy],
                             cwd=sandbox, env=env, capture_output=True, timeout=120)
        logs = list((sandbox / '.evidence').rglob('*')) if (sandbox / '.evidence').exists() else []
        persisted = any(b'launcher-sensitive-fixture' in p.read_bytes() for p in logs if p.is_file())
        result = {'case': name, 'nonzeroExit': run.returncode != 0,
                  'safeOutput': b'launcher-sensitive-fixture' not in run.stdout + run.stderr,
                  'noExecutionDirectory': not (sandbox / '.evidence').exists(),
                  'noSecretPersisted': not persisted}
        result['pass'] = all(result[k] for k in ('nonzeroExit', 'safeOutput', 'noExecutionDirectory', 'noSecretPersisted'))
        results.append(result)
        # Use a separate subprocess workspace for each failure's artifacts.
        shutil.rmtree(sandbox / '.evidence', ignore_errors=True)
report = {'seam': 'run.py public process; no vendor key supplied', 'phase': phase,
          'launcherSHA256': hashlib.sha256((base / 'run.py').read_bytes()).hexdigest(),
          'command': ['python3', '.scratch/barness-ai-openai-decisions/chinese-live-evidence/launcher-process.py', phase],
          'results': results, 'status': 'PASS' if all(r['pass'] for r in results) else 'FAIL'}
(base / ('launcher-' + phase + '.json')).write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report))
sys.exit(0 if report['status'] == 'PASS' else 1)
