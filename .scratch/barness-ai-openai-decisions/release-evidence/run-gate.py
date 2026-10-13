#!/usr/bin/env python3
"""Run the actual full gate using a completed isolated evidence collection."""
import argparse
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--collection', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
os.chdir(root)
if a.out.exists() or a.out.is_absolute() or not a.out.resolve().is_relative_to(root):
    raise SystemExit('output must be fresh and repository-relative')
records = json.loads(a.collection.read_text())
command = ['go', 'run', './ai/release/cmd/releasegate', '-out', str(a.out)]
for record in records:
    if record.get('bundle'):
        command += ['-live' if record['kind'] == 'smoke' else '-evaluation-jev', record['bundle']]
command += ['-evaluation-luna', '.scratch/barness-ai-openai-decisions/chinese-live-evidence/live']
a.out.mkdir(parents=True)
(a.out / 'command.json').write_text(json.dumps(command, indent=2) + '\n')
env = {k: os.environ[k] for k in ['PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH'] if k in os.environ}
with (a.out / 'command.log').open('w') as log:
    child = subprocess.Popen(command, cwd=root, env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    for line in child.stdout:
        log.write(line.replace(str(root), '[REPO]').replace(str(Path.home()), '[HOME]'))
        log.flush()
    code = child.wait()
(a.out / 'exit.json').write_text(json.dumps({'exitCode': code}) + '\n')
print('releasegate exit', code, '; bundle', a.out)
raise SystemExit(code)
