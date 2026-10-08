#!/usr/bin/env python3
"""Harvest issue 10's audited offline run using relative paths and content hashes."""
import hashlib
import json
import subprocess
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent
bundle = (root / sys.argv[1]).resolve()
base = '3b31ec0edb01d2a0a64828fe36acbe8dbb0b7e6b'
command = 'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai/issue10-final" go test -race ./... -count=1'
def read(path):
    return json.loads(path.read_text())
def write(name, value):
    (folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')
def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

manifest = read(bundle / 'manifest.json')
audit = read(bundle / 'audit.json')
cases = manifest['cases']
images = [c for c in cases if c['id'].startswith('P08-')]
if any(c['status'] != 'PASS' for c in cases) or audit['findings']:
    raise SystemExit('run has failures or redaction findings')
index, resources, allocations = [], [], []
for case in images:
    case_dir = bundle / case['dir']
    entry = dict(case)
    entry['artifacts'] = {p.name: sha(p) for p in sorted(case_dir.iterdir()) if p.is_file()}
    index.append(entry)
    path = case_dir / 'resources.json'
    if path.exists():
        resources.append({'id': case['id'], 'gauges': read(path)})
    path = case_dir / 'allocated-bytes.json'
    if path.exists():
        allocations.append({'id': case['id'], 'bytes': read(path)})
if any(any(v for v in r['gauges'].values()) for r in resources):
    raise SystemExit('resource gauges are not zero')
if any(a['bytes'] >= 4 << 20 for a in allocations):
    raise SystemExit('callback allocation exceeds acceptance limit')
paths = subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()
source_hashes = {p: sha(root / p) for p in paths if (root / p).is_file()}
write('image-cases.json', {'cases': index})
write('image-resources.json', {'resources': resources, 'allocations': allocations})
(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())
report = {
    'status': 'PASS', 'date': '2026-10-08', 'issue': '10-openai-json-image-editing',
    'base_commit': base, 'tested_commit': manifest['git_commit'], 'command': command,
    'bundle': str(bundle.relative_to(root)), 'manifest_sha256': sha(bundle / 'manifest.json'),
    'passing_cases': len(cases), 'failed_cases': 0, 'image_cases': len(images),
    'edit_cases': sum('edit-' in c['id'] for c in images),
    'differential_cases': sum(c['id'].startswith('PIDIFF-') for c in cases),
    'pressure_cases': sum('pressure' in c['id'] for c in cases),
    'audit_findings': 0, 'versions': manifest['versions'],
    'go_version': manifest['go_version'], 'platform': manifest['platform'],
    'started_at': manifest['started_at'], 'finished_at': manifest['finished_at'],
    'resource_cases': len(resources), 'all_recorded_gauges_zero': True,
    'allocation_cases': len(allocations), 'allocation_max_bytes': max(a['bytes'] for a in allocations),
    'source_hashes': source_hashes,
    'live_acceptance': 'Issue 11; no image model or support-matrix change',
}
write('report.json', report)
write('delivery-hashes.json', source_hashes)
write('manifest.json', {'artifacts': {p.name: sha(p) for p in sorted(folder.iterdir()) if p.is_file() and p.name not in ('manifest.json', 'audit.json')}})
print(f"harvested {len(cases)} PASS cases; {len(images)} image cases")
