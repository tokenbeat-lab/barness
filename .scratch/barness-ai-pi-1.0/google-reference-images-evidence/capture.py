#!/usr/bin/env python3
"""Capture issue 13's audited run without copying repeated fixture payloads.

Per-issue capture stays independent so prior immutable deliveries keep their
original evidence schema and verification behavior.
"""
import hashlib
import json
import subprocess
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent
bundle = (root / sys.argv[1]).resolve()
base = '52ce4feaa6001fc06895cffd7bf94eb6dcd82986'


def read(path):
    return json.loads(path.read_text())


def write(name, value):
    (folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


manifest = read(bundle / 'manifest.json')
cases = manifest['cases']
if any(c['status'] != 'PASS' for c in cases) or read(bundle / 'audit.json')['findings']:
    raise SystemExit('run has failures or redaction findings')
google = [c for c in cases if c['id'].startswith('P09-')]
if not google:
    raise SystemExit('no P09 evidence')
index, resources, observations = [], [], []
kept = {'result.json', 'error.json', 'requests.json', 'response-script.json', 'resources.json', 'assertions.json', 'allocated-bytes.json'}
for case in google:
    case_dir = bundle / case['dir']
    entry = dict(case)
    entry['artifacts'] = {p.name: sha(p) for p in sorted(case_dir.iterdir()) if p.is_file()}
    entry['records'] = {p.name: read(p) for p in sorted(case_dir.iterdir()) if p.name in kept}
    index.append(entry)
    if (case_dir / 'resources.json').exists():
        resources.append({'id': case['id'], 'gauges': read(case_dir / 'resources.json')})
    if (case_dir / 'observations.json').exists():
        observations.extend(read(case_dir / 'observations.json'))
if any(any(r['gauges'].values()) for r in resources):
    raise SystemExit('resource gauges are not zero')
paths = subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()
source_hashes = {p: sha(root / p) for p in paths if (root / p).is_file()}
write('google-cases.json', {'cases': index})
write('resources.json', {'cases': resources, 'all_recorded_gauges_zero': True})
write('observations-google-images.json', observations)
(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())
write('source-hashes.json', source_hashes)
write('manifest.json', {
    **{k: v for k, v in manifest.items() if k != 'cases'}, 'cases': google,
    'source_bundle': str(bundle.relative_to(root)), 'source_manifest_sha256': sha(bundle / 'manifest.json'),
    'source_passing_cases': len(cases), 'scope': 'P09 offline Google Interactions image generation and ordered reference editing',
})
verification = {
    'status': 'PASS', 'issue': '13-google-reference-images-and-guards', 'date': '2026-10-08',
    'base_commit': base, 'tested_commit': manifest['git_commit'],
    'bundle': str(bundle.relative_to(root)), 'manifest_sha256': sha(bundle / 'manifest.json'),
    'commands': [
        {'command': 'go vet ./...', 'exit_code': 0},
        {'command': 'go vet -tags live ./...', 'exit_code': 0},
        {'command': "go test ./ai/e2e -run '^TestGoogleImages|^TestOpenAIImagesEditingKnownSize|^TestUnaryCallback(AllocationBound|EncodingSemantics|ByteElementEncoding)' -count=1", 'exit_code': 0},
        {'command': 'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1', 'exit_code': 0},
    ],
    'passing_cases': len(cases), 'failed_cases': 0, 'google_cases': len(google),
    'openai_image_cases': sum(c['id'].startswith('P08-') for c in cases),
    'differential_cases': sum(c['id'].startswith('PIDIFF-') for c in cases),
    'pressure_cases': sum('pressure' in c['id'] for c in cases),
    'resource_cases': len(resources), 'all_recorded_gauges_zero': True,
    'run_audit_findings': 0, 'review': read(folder / 'review-results.json'),
    'versions': manifest['versions'], 'go_version': manifest['go_version'], 'platform': manifest['platform'],
    'started_at': manifest['started_at'], 'finished_at': manifest['finished_at'],
    'live_acceptance': 'Issue 14; host catalog fixture only. Generation/reference editing and guards are proven offline.',
}
write('verification.json', verification)
# Re-run capture after editing delivery prose or adding trace evidence.
write('sha256.json', {p.name: sha(p) for p in sorted(folder.iterdir()) if p.is_file() and p.name not in {'sha256.json', 'audit.json'}})
print(f'captured {len(google)} Google cases; {len(cases)} full-run PASS cases')
