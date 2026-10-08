"""Verify saved issue07 evidence from the repository root; no network required."""
import hashlib
import json
from pathlib import Path

root = Path.cwd()
evidence = root / '.scratch/barness-ai-pi-1.0/unary-failures-evidence'


def read(path):
    return json.loads(path.read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


manifest = read(evidence / 'manifest.json')
assert manifest['status'] == 'PASS'
for name, expected in manifest['artifacts'].items():
    assert digest(evidence / name) == expected, name
report = read(evidence / 'report.json')
for name, expected in report['code_and_fixtures_files'].items():
    assert digest(root / name) == expected, name
assert read(evidence / 'audit.json')['findings'] == []
resources = read(evidence / 'unary-resources.json')
assert all(value == 0 for gauges in resources['resources'].values() for value in gauges.values())
assert all(value < resources['allocation_ceiling_bytes'] for value in resources['allocations'].values())
assert all(case['status'] == 'PASS' for case in read(evidence / 'unary-cases.json')['cases'])
print('PASS: saved artifact hashes, tested code, resource and allocation records')

run = report['checks']['full_race_with_differential_and_pressure']
bundle = root / run['bundle']
if (bundle / 'manifest.json').exists():
    assert digest(bundle / 'manifest.json') == run['manifest_sha256']
    for case in read(evidence / 'unary-cases.json')['cases']:
        for name, expected in case['artifacts_sha256'].items():
            assert digest(bundle / case['dir'] / name) == expected, (case['id'], name)
    assert read(bundle / 'audit.json')['findings'] == []
    print('PASS: original local bundle manifest and all unary artifact hashes')
else:
    print('Original local bundle absent; replay commands are in unary-cases.json')
