#!/usr/bin/env python3
"""Record finite keyless commands and exact measured production source hashes."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

root = Path(__file__).resolve().parents[4]
out = root / '.evidence/visual-unsent-tdd' / sys.argv[1]
out.mkdir(parents=True, exist_ok=False)
argv = sys.argv[2:]
env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH') if k in os.environ}
env.update({'BARNESS_AI_EVIDENCE_DIR': str(out / 'bundles'), 'BARNESS_EXAMPLE_RACE': '1' if '-race' in argv else '0'})
def snapshot():
    names = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'ai', 'go.mod', 'go.sum'], cwd=root).decode().splitlines()
    names += ['ai/examples/visualeval/unsent_test.go']
    return {n: hashlib.sha256((root / n).read_bytes()).hexdigest() for n in sorted(set(names)) if Path(n).suffix in ('.go', '.json') or n in ('go.mod', 'go.sum')}
before = snapshot()
(out / 'source-before.json').write_text(json.dumps(before, sort_keys=True, indent=2) + '\n')
started = time.time()
with (out / 'stdout.log').open('wb') as log:
    result = subprocess.run(argv, cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=600)
after = snapshot()
(out / 'source-after.json').write_text(json.dumps(after, sort_keys=True, indent=2) + '\n')
receipt = {'argv': argv, 'sourceCommit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root).decode().strip(), 'startedUnix': started, 'finishedUnix': time.time(), 'exitCode': result.returncode, 'sourceStable': before == after, 'keyEnvironment': 'whitelist; no vendor keys', 'stdoutSha256': hashlib.sha256((out / 'stdout.log').read_bytes()).hexdigest()}
(out / 'command.json').write_text(json.dumps(receipt, indent=2) + '\n')
print(json.dumps(receipt))
sys.exit(result.returncode)
