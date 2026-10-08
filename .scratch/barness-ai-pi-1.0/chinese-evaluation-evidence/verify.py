#!/usr/bin/env python3
"""Verify delivery/source hashes and independent evaluation evidence offline."""
import hashlib
import json
from pathlib import Path
import subprocess

base = Path(__file__).resolve().parent
repo = base.parents[2]
manifest = json.loads((base / 'manifest.json').read_text())
def digest(path):
    return 'sha256:' + hashlib.sha256(path.read_bytes()).hexdigest()
for name, expected in manifest['files'].items():
    assert digest(base / name) == expected, 'delivery hash: ' + name
for name, expected in manifest['sources'].items():
    assert digest(repo / name) == expected, 'source hash: ' + name
assert digest(repo / 'ai/live/support-matrix.json') == manifest['supportMatrixHash'], 'support matrix changed'
for name in ('live-initial', 'live', 'not-run'):
    subprocess.run(['go', 'run', './ai/examples/chineseeval/cmd/chineseeval', '-verify', str(base / name)], cwd=repo, check=True)
history = json.loads((base / 'historical-runs.json').read_text())
assert history['calls'] == 48 and history['attempts'] == 48
for row in history['runs']:
    report = json.loads((base / row['evidence'] / 'evaluation-report.json').read_text())
    assert row['budget'] == report['budget'] and row['correct'] == report['metrics']['correct']
full = json.loads((base / 'full-suite.json').read_text())
assert full['exit'] == 0 and not full['failedTests']
subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(base)], cwd=repo, check=True)
print('PASS: independent live results, NOT_RUN, fixture artifacts, full suite, hashes and delivery audit')
