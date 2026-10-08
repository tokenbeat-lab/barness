#!/usr/bin/env python3
"""Verify source, delivery and available original-run hashes for issue 13."""
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


for path, expected in read(folder / 'source-hashes.json').items():
    if sha(root / path) != expected:
        raise SystemExit(f'source hash mismatch: {path}')
for name, expected in read(folder / 'sha256.json').items():
    if sha(folder / name) != expected:
        raise SystemExit(f'delivery hash mismatch: {name}')
v = read(folder / 'verification.json')
if v['failed_cases'] or v['run_audit_findings'] or read(folder / 'audit.json')['findings']:
    raise SystemExit('failed scenario or audit finding')
if any(c['exit_code'] for c in v['commands']):
    raise SystemExit('failed command')
trace = read(folder / 'trace.json')
if trace['tested_commit'] != v['tested_commit']:
    raise SystemExit('trace commit differs from tested commit')
if any(t['status'] != 'PASS' for t in trace['items']):
    raise SystemExit('unproven P09 trace item')
if v['review']['standards_remaining'] or v['review']['spec_remaining']:
    raise SystemExit('unresolved review finding')
bundle = root / v['bundle']
if bundle.exists():
    if sha(bundle / 'manifest.json') != v['manifest_sha256']:
        raise SystemExit('source manifest hash mismatch')
    for case in read(folder / 'google-cases.json')['cases']:
        for name, expected in case['artifacts'].items():
            if sha(bundle / case['dir'] / name) != expected:
                raise SystemExit(f"run artifact mismatch: {case['id']}/{name}")
print(f"verified {v['google_cases']} Google cases; {v['passing_cases']} full-run cases PASS")
print('full replay: ' + v['commands'][-1]['command'])
print('bundle: ' + v['bundle'])
