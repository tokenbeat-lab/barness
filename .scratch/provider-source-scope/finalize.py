"""Read-only receipt/custody assertions, followed by new immutable audit indexes."""
import collections
import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[2]
BASE = ROOT / '.evidence/provider-source-scope'
sys.path.insert(0, str(ROOT / 'tools/verification'))
from receipt_files import source_snapshot


def load(p):
    return json.loads(Path(p).read_text())


def sha(p):
    p = Path(p)
    assert p.is_file() and not p.is_symlink(), str(p)
    return hashlib.sha256(p.read_bytes()).hexdigest()


def save(name, obj):
    p = BASE / name
    with p.open('x') as f:
        f.write(json.dumps(obj, indent=2, sort_keys=True) + '\n')
    return str(p.relative_to(ROOT))


def ref(p):
    p = Path(p)
    if not p.is_absolute():
        p = ROOT / p
    return {'path': str(p.relative_to(ROOT)), 'sha256': sha(p)}


current = source_snapshot(ROOT)
before = load(BASE / 'source-before.json')
changed_source = [n for n in sorted(set(before['files']) | set(current['files'])) if before['files'].get(n) != current['files'].get(n)]

def product(n):
    if n in ('go.mod', 'go.sum'):
        return True
    if n.startswith('ai/examples/requestid/'):
        return True
    return n.startswith('ai/') and not any(n.startswith(p) for p in ('ai/internal/testkit/', 'ai/live/', 'ai/e2e/', 'ai/examples/', 'ai/release/'))

assert not [p for p in changed_source if product(p)]

# Use the exact selected command, not the newest directory name or LATEST.
selection = load(BASE / 'full-ordinary-selection-v3.json')['argv']
command = selection[selection.index('--') + 1:]
inputs = {}
for i, arg in enumerate(selection):
    if arg == '--input':
        inputs[selection[i + 1]] = sha(ROOT / selection[i + 1])
candidates = []
for p in (BASE / 'e2e').glob('*/receipt.json'):
    d = load(p)
    if d.get('argv') == command and d.get('source') == current and d.get('selectedInputs') == inputs:
        candidates.append((p, d))
assert len(candidates) == 1, [(str(p), d.get('status')) for p, d in candidates]
ordinary_path, ordinary = candidates[0]
assert ordinary['status'] == 'PASS' and ordinary['exitCode'] == 0, ordinary['status']
packages = []
source_cases = set()
for bundle in ordinary['bundles']:
    path = ordinary_path.parent / bundle['path']
    assert sha(path / 'manifest.json') == bundle['manifestSHA256']
    assert sha(path / 'audit.json') == bundle['auditSHA256']
    manifest, audit = load(path / 'manifest.json'), load(path / 'audit.json')
    assert audit['findings'] == []
    for case in manifest['cases']:
        for name, h in case['artifacts'].items():
            assert sha(path / case['dir'] / name) == h
        if case['id'].startswith('SOURCE55-'):
            assert case['status'] == 'PASS'
            source_cases.add(case['id'])
    counts = dict(collections.Counter(c['status'] for c in manifest['cases']))
    assert set(counts) <= {'PASS', 'NOT_RUN'}, counts
    packages.append({'package': bundle['package'], 'cases': len(manifest['cases']), 'statuses': counts, 'manifest': ref(path / 'manifest.json'), 'audit': ref(path / 'audit.json')})
assert len(source_cases) == 50, len(source_cases)
assert sum(b['package'] == './ai/e2e' for b in ordinary['bundles']) == 1

for n in ['tools', 'vet']:
    r = load(BASE / 'tool-process-and-vet-v3' / n / 'receipt.json')
    assert r['exitCode'] == 0 and r['status'] == 'PASS' and r['source'] == current
replay = load(BASE / 'public-replay-v5/receipt.json')
assert replay['exitCode'] == 0 and replay['status'] == 'PASS' and replay['source'] == current
assert all(sha(ROOT / p) == h for p, h in replay['inputs'].items())
calls = load(BASE / 'public-replay-v5/stdout.jsonl')
assert len(calls) == 40 and all(x['pass'] and x['requestEqual'] and x['sessionEmpty'] for x in calls)
assert sum(x['responseEqual'] for x in calls) == 36 and sum(x['cancelCapture'] for x in calls) == 4

consumers = load(BASE / 'consumption-v3/report.json')
qualified = []
for row in consumers:
    p = ROOT / row['consumer']
    r = load(p)
    assert row['exitCode'] == 0 and r['pass'] is True and r['exitCode'] == 0 and r['schema'] == 2 and r['source'] == current
    # The capture's numerical exit and sealed output must still be authentic.
    for n, h in r['files'].items():
        assert sha(p.parent / n) == h
    assert (p.parent / 'exit-code').read_text().strip() == '0'
    inner = load(p.parent / 'native/live-consumer.json')
    assert inner['pass'] is True and inner['inputs'] == r['inputs'] and inner['source'] == current
    for n, h in r['inputs'].items():
        assert sha(ROOT / r['bundle'] / n) == h
    report = load(ROOT / r['bundle'] / 'live-report.json')
    assert len(report['scenarios']) == len(report['expected']) and all(s['outcome'] == 'PASS' for s in report['scenarios'])
    qualified.append({'route': row['route'], 'kind': row['kind'], 'qualification': 'APPLICABLE', 'receipt': ref(p), 'originalBundle': r['bundle'], 'originalCommit': report['gitCommit'], 'originalFinishedAt': report['finishedAt'], 'models': sorted({s.get('model', report.get('model')) for s in report['scenarios']}), 'mandatoryCases': len(report['scenarios']), 'effectsSamples': 12 if row['route'].startswith('typesafe') else 0, 'newVendorCalls': 0})

protected = {}
for name in ['protected-history-before.json', 'supplemental-protection-before.json']:
    protected.update(load(BASE / name))
for name, h in protected.items():
    assert sha(ROOT / name if not Path(name).is_absolute() else name) == h, name
roots = [Path(n) for n in load(BASE / 'store-roots.json')]
now = {str(p.resolve()): sha(p) for root in roots for p in root.rglob('*') if p.is_file()}
old = {str((ROOT / p).resolve()): h for p, h in protected.items() if any((ROOT / p).resolve().is_relative_to(root) for root in roots)}
assert now == old, {'new': sorted(set(now) - set(old)), 'removed': sorted(set(old) - set(now)), 'changed': [p for p in now.keys() & old.keys() if now[p] != old[p]]}
protection_path = save('history-protection-check-v1.json', {'protectedFiles': len(protected), 'unchanged': len(protected), 'changed': [], 'missing': [], 'publicStoreFiles': len(now), 'publicStoreNewFiles': [], 'publicStoreRemovedFiles': [], 'publicStoreChangedFiles': [], 'realClaims': 0, 'resumedRounds': 0, 'baselines': [ref(BASE / 'protected-history-before.json'), ref(BASE / 'supplemental-protection-before.json')]})

checks = {'source': current, 'sourceChangedFiles': changed_source, 'productSourceChangedFiles': [], 'ordinary': ref(ordinary_path), 'ordinaryElapsedSeconds': ordinary['elapsedSeconds'], 'ordinarySelectedInputs': len(inputs), 'packages': packages, 'requiredSourceCases': sorted(source_cases), 'tools': ref(BASE / 'tool-process-and-vet-v3/tools/receipt.json'), 'vet': ref(BASE / 'tool-process-and-vet-v3/vet/receipt.json'), 'publicReplay': ref(BASE / 'public-replay-v5/receipt.json'), 'historyProtection': ref(protection_path), 'build': ref(BASE / 'build-v4/receipt.json'), 'liveExecuted': False, 'raceExecuted': False, 'pressureExecuted': False, 'releasePass': False}
verification_path = save('verification-and-retention-v1.json', checks)
decisions_path = save('decisions-v1.json', {'kind': 'reviewed-route-source-applicability', 'routes': qualified[:3], 'unchangedGLMRegression': qualified[3:], 'review': ref(ROOT / 'ai/internal/testkit/evidence/reviewed_sources.json'), 'originalOmissionCustody': ref(BASE / 'source-omission-custody-v1.json'), 'verification': ref(verification_path), 'minimumPaidRetestsForTheseRoutes': [], 'newVendorCalls': 0, 'newGrants': 0, 'realClaims': 0, 'releasePass': False})
previous_path = ROOT / '.evidence/provider-delivery-offline/current-qualification-index-v2.json'
previous = load(previous_path)
entries = {r['route']: r for r in previous['qualified']}
for r in qualified:
    old_entry = entries.get(r['route'])
    entries[r['route']] = {'route': r['route'], 'qualification': 'APPLICABLE', 'receipt': r['receipt']['path'], 'receiptSHA256': r['receipt']['sha256'], 'basis': 'current complete native read-only qualification; original measurement unchanged', 'mandatoryCases': r['mandatoryCases']}
    if old_entry:
        entries[r['route']]['previousQualification'] = old_entry
assert len(entries) == 23, len(entries)
index_path = save('current-qualification-index-v3.json', {'kind': 'retained-and-reviewed-route-qualification-evidence', 'previousIndex': ref(previous_path), 'currentDecisions': ref(decisions_path), 'qualified': list(entries.values()), 'qualifiedModelRouteEvidenceEntries': len(entries), 'pendingCurrentFormalRecords': [], 'partialKimiRecords': previous['partialKimiRecords'], 'priorActualFailure': previous['priorActualFailure'], 'productSourceChanges': [], 'verificationSourceChanges': changed_source, 'countingNote': '23 route/model evidence entries is not a 23/29 path delivery ratio. TypeSafe 1.13 and 1.14 share one operation/protocol path; Kimi six added scenarios, prior MiniMax Messages failure and required human stages remain independent.', 'releasePass': False})
print(json.dumps({'ordinary': str(ordinary_path.relative_to(ROOT)), 'packages': packages, 'protectedFiles': len(protected), 'publicStoreFiles': len(now), 'sourceFiles': len(current['files']), 'sourceCases': len(source_cases), 'index': index_path, 'decisions': decisions_path}, indent=2))
