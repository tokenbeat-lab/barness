#!/usr/bin/env python3
"""Verify source, public run artifacts and the portable issue 15 delivery."""
import hashlib
import json
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


for p, h in read(folder / 'source-hashes.json').items():
    assert sha(root / p) == h, f'source hash mismatch: {p}'
for p, h in read(folder / 'sha256.json').items():
    assert sha(folder / p) == h, f'delivery hash mismatch: {p}'
v = read(folder / 'verification.json')
m = read(folder / 'manifest.json')
assert v['tested_commit'] == m['git_commit']
assert sha(folder / 'manifest.json') == v['manifest_sha256']
for p, h in read(folder / 'source-hashes.json').items():
    committed = subprocess.check_output(['git', 'show', f'{v["tested_commit"]}:{p}'], cwd=root)
    assert hashlib.sha256(committed).hexdigest() == h, f'tested commit source mismatch: {p}'
assert len(m['cases']) == v['passing_cases'] and all(c['status'] == 'PASS' for c in m['cases'])
assert not v['failed_cases'] and not v['run_audit_findings']
assert not read(folder / 'audit.json')['findings'] and not read(folder / 'run-audit.json')['findings']
commands = read(folder / 'command-results.json')
assert commands['tested_commit'] == v['tested_commit']
assert commands['commands'] == v['commands']
full = next(c for c in commands['commands'] if c['command'].endswith('go test -race ./... -count=1'))
assert full['evidence_bundle'] == v['bundle']
for c in commands['commands']:
    assert c['exit_code'] == 0 and c['tested_commit'] == c['ending_commit'] == v['tested_commit']
    assert sha(folder / Path(c['log']).name) == c['log_sha256']
assert not v['review']['standards_remaining'] and not v['review']['spec_remaining']
assert all(t['status'] == 'PASS' for t in read(folder / 'trace.json')['items'])
assert read(folder / 'trace.json')['differential']['pending'] == 0
default = next(c for c in commands['commands'] if c['command'] == "go test ./ai/e2e -run '^TestMixedPressure$' -count=1")
default_run = read(folder / 'not-run.json')
assert default_run['bundle'] == default['evidence_bundle']
assert default_run['manifest']['git_commit'] == v['tested_commit']
assert len(default_run['manifest']['cases']) == 2
assert all(c['status'] == 'NOT_RUN' for c in default_run['manifest']['cases'])
assert not default_run['audit']['findings']
default_bundle = root / default_run['bundle']
if default_bundle.exists():
    assert sha(default_bundle / 'manifest.json') == default_run['manifest_sha256']
cases = read(folder / 'mixed-cases.json')['cases']
assert len(cases) == v['mixed_cases'] == 103
bundle = root / v['bundle']
for case in cases:
    assert case['status'] == 'PASS'
    assert all(a['pass'] for a in case['records']['assertions.json']['assertions'])
    if bundle.exists():
        for name, h in case['artifacts'].items():
            assert sha(bundle / case['dir'] / name) == h, f'run artifact mismatch: {case["id"]}/{name}'
for case in read(folder / 'resources.json')['cases']:
    assert not any(case['gauges'].values()), case['id']
pressure = read(folder / 'pressure.json')['cases']
assert len(pressure) == 2
for p in pressure:
    assert p['heapGrowth'] <= p['totalAllocDuringCalls'] <= p['heapBudget']
    assert p['peakPermits'] == p['load']['calls'] == p['captureBoundCalls']
    assert p['load']['calls'] in (4, 8) and p['load']['chatOutputBytes'] == 16 << 10
    assert p['clientUnaryBytesAtEOF'] == 2796928 * p['load']['tenants']
    assert p['inputImageBase64Bytes'] == 349528 and p['outputImageBase64Bytes'] == 699052
    assert all(h['ratio'] <= 0.75 for h in p['headroom'].values())
for run in read(folder / 'red.json')['runs']:
    assert any(c['status'] == 'FAIL' for c in run['cases'])
    assert not run['audit']['findings']
print(f"verified {v['mixed_cases']} mixed scenarios; {v['passing_cases']} full-run scenarios PASS")
print('replay: BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1')
