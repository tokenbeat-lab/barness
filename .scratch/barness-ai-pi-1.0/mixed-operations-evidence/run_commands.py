#!/usr/bin/env python3
"""Run actual verification commands and record exit codes, commits and logs.

Use --checks-only when an independently recorded full run is already available.
Keep output below ignored .evidence so command logs cannot expose local paths.
"""
import argparse
import hashlib
import json
import os
import shlex
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[3]
parser = argparse.ArgumentParser()
parser.add_argument('output')
parser.add_argument('--checks-only', action='store_true')
args = parser.parse_args()
output = root / args.output
output.mkdir(parents=True, exist_ok=True)
commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
lines = [
    ('build', 'go build ./...'),
    ('vet', 'go vet ./...'),
    ('vet-live', 'go vet -tags live ./...'),
    ('mixed-race', "BARNESS_AI_PRESSURE=1 go test -race ./ai/e2e -run '^TestMixed' -count=1"),
    ('not-run', "go test ./ai/e2e -run '^TestMixedPressure$' -count=1"),
]
if not args.checks_only:
    lines.append(('full-race', 'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1'))
records = []
for name, line in lines:
    command = shlex.split(line)
    env = os.environ.copy()
    # Each command has independent evidence; inherited flags cannot enable a
    # supposedly default-NOT_RUN pressure run.
    env.pop('BARNESS_AI_PRESSURE', None)
    env.pop('BARNESS_AI_PIDIFF', None)
    while '=' in command[0]:
        k, value = command.pop(0).split('=', 1)
        env[k] = value
    env['BARNESS_AI_EVIDENCE_DIR'] = str(output / (name + '-bundles'))
    log = output / (name + '.log')
    with log.open('wb') as stream:
        result = subprocess.run(command, cwd=root, env=env, stdout=stream, stderr=subprocess.STDOUT)
    ending = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
    records.append({'command': line, 'exit_code': result.returncode, 'tested_commit': commit,
                    'ending_commit': ending, 'log': str(log.relative_to(root)),
                    'log_sha256': hashlib.sha256(log.read_bytes()).hexdigest()})
    if command[:2] == ['go', 'test']:
        manifests = sorted((output / (name + '-bundles')).glob('*/manifest.json'))
        if not manifests:
            raise SystemExit('test did not produce evidence manifest')
        records[-1]['evidence_bundle'] = str(manifests[-1].parent.relative_to(root))
    (output / 'command-results.json').write_text(json.dumps({'tested_commit': commit, 'commands': records}, indent=2) + '\n')
    print(name, result.returncode, flush=True)
    if result.returncode or ending != commit:
        raise SystemExit('command failed or checkout changed during verification')
