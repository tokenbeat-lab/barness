"""Verify saved issue08 evidence from the repo root; no network or keys needed."""
import hashlib
import json
from pathlib import Path

root = Path.cwd()
out = root / '.scratch/barness-ai-pi-1.0/typesafe-live-evidence'


def read(path):
    return json.loads(path.read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


manifest = read(out / 'manifest.json')
assert manifest['status'] == 'PASS'
for name, expected in manifest['artifacts'].items():
    assert digest(out / name) == expected, name
report = read(out / 'report.json')
for name, expected in report['code_and_fixtures_files'].items():
    assert digest(root / name) == expected, name
assert read(out / 'audit.json')['findings'] == []
live = read(out / 'live/live-report.json')
assert live['schema'] == 2 and live['operation'] == 'classifier'
assert live['combo'] == 'typesafe-classifier' and live['model'] == 'jev-1.13.0'
assert live['endpoint'] == 'https://api.typesafe.ai/v1'
assert len(live['scenarios']) == 3
assert all(s['outcome'] == 'PASS' and s['calls'] > 0 and all(s['providerRequestIds']) for s in live['scenarios'])
assert set(live['expected']) == {'mixed-questions', 'single-choice', 'context-422'}
b = live['budget']
assert b['callsUsed'] <= b['maxCalls'] == 6 and b['questionsUsed'] <= b['maxQuestions'] == 10
assert b['callsUsed'] == sum(s['calls'] for s in live['scenarios'])
assert b['questionsUsed'] == sum(s['questions'] for s in live['scenarios'])
assert b['httpAttempts'] == sum(s['attempts'] for s in live['scenarios'])
assert '422 shape UNCONFIRMED' in live['scenarios'][-1]['note']
assert read(out / 'live/audit.json')['findings'] == []
assert all(s['outcome'] == 'NOT_RUN' for s in read(out / 'not-run/live-report.json')['scenarios'])
before, after = read(out / 'matrix-before.json'), read(out / 'matrix-after.json')
assert before['rows'][:6] == after['rows'][:6]
assert after == read(root / 'ai/live/support-matrix.json')
assert after['rows'][6]['operation'] == 'classifier'
assert all(r['operation'] == 'chat' for r in after['rows'][:6])
assert read(out / 'matrix-verification.json')['notRunPreservesBytes']
assert report['checks']['full_race_with_differential_and_pressure']['exitCode'] == 0
cases = read(out / 'offline-cases.json')['cases']
assert all(c['status'] == 'PASS' for c in cases)
for case in cases:
    saved = out / 'offline' / case['dir']
    if saved.exists():
        for name, expected in case['artifacts_sha256'].items():
            assert digest(saved / name) == expected, (case['id'], name)
assert report['checks']['release_evaluation']['typesafeLiveStatus'] == 'PASS'
assert report['checks']['release_evaluation']['typesafeTraceStatus'] == 'PASS'
print('PASS: artifact/source hashes, own live evidence, NOT_RUN preservation, matrix isolation and final tests')

bundle = root / report['checks']['full_race_with_differential_and_pressure']['bundle']
if (bundle / 'manifest.json').exists():
    assert digest(bundle / 'manifest.json') == report['checks']['full_race_with_differential_and_pressure']['manifest_sha256']
    for case in read(out / 'offline-cases.json')['cases']:
        for name, expected in case['artifacts_sha256'].items():
            assert digest(bundle / case['dir'] / name) == expected, (case['id'], name)
    assert read(bundle / 'audit.json')['findings'] == []
    print('PASS: original full race bundle and all saved case hashes')
else:
    print('Original local full bundle absent; replay commands are in offline-cases.json')
