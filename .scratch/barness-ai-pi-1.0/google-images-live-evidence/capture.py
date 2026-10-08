#!/usr/bin/env python3
"""Capture issue 14's audited artifacts without repeated fixture payloads."""
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent
base = '2aa71180fc1dbc647b94520abdf1be0c8944ea10'


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(name, value):
    (folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def cases_index(bundle, manifest):
    kept = {'result.json', 'error.json', 'requests.json', 'response-script.json',
            'resources.json', 'assertions.json'}
    return [{**case, 'artifacts': {p.name: sha(p) for p in (bundle / case['dir']).iterdir() if p.is_file()},
             'records': {p.name: read(p) for p in (bundle / case['dir']).iterdir() if p.name in kept}}
            for case in manifest['cases']]


for name, relative in {
    'delivery-refused': '.evidence/issue14-live/20261008T090854.560734000Z',
    'prompt-refused': '.evidence/issue14-live/20261008T091316.585690000Z',
    'live-pass': '.evidence/issue14-live/20261008T091619.605370000Z',
    'not-run': '.evidence/issue14-not-run/20261008T091928.075162000Z',
}.items():
    source = root / relative
    if read(source / 'audit.json')['findings']:
        raise SystemExit('source live audit has findings')
    shutil.copytree(source, folder / name, dirs_exist_ok=True)

write('manifest.json', {'module': 'barness-ai', 'scope': 'issue 14 own live reports and offline evidence', 'base_commit': base, 'bundles': ['delivery-refused', 'prompt-refused', 'live-pass', 'not-run']})

if len(sys.argv) > 1:
    bundle = (root / sys.argv[1]).resolve()
    manifest = read(bundle / 'manifest.json')
    cases = manifest['cases']
    if any(c['status'] != 'PASS' for c in cases) or read(bundle / 'audit.json')['findings']:
        raise SystemExit('full offline run has failures or audit findings')
    google = [c for c in cases if c['id'].startswith('P09-')]
    index = cases_index(bundle, {**manifest, 'cases': google})
    write('google-cases.json', {'cases': index})
    write('offline-manifest.json', {**manifest, 'source_bundle': str(bundle.relative_to(root)),
                                   'source_manifest_sha256': sha(bundle / 'manifest.json')})
    write('offline-summary.json', {'passing_cases': len(cases), 'google_cases': len(google),
          'differential_cases': sum(c['id'].startswith('PIDIFF-') for c in cases),
          'pressure_cases': sum('pressure' in c['id'] for c in cases), 'audit_findings': 0})

if len(sys.argv) > 2:
    bundle = (root / sys.argv[2]).resolve()
    manifest = read(bundle / 'manifest.json')
    if any(c['status'] != 'PASS' for c in manifest['cases']) or read(bundle / 'audit.json')['findings']:
        raise SystemExit('live harness run has failures or audit findings')
    write('live-harness-manifest.json', {**manifest, 'source_bundle': str(bundle.relative_to(root)),
                                       'source_manifest_sha256': sha(bundle / 'manifest.json'),
                                       'audit_findings': 0})
    write('live-harness-cases.json', {'cases': cases_index(bundle, manifest)})

paths = subprocess.check_output(['git', 'diff', '--name-only', base, '--', 'ai', 'docs', 'GLOSSARY.md'], cwd=root, text=True).splitlines()
paths += subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard', '--', 'ai', 'docs'], cwd=root, text=True).splitlines()
paths += ['.scratch/barness-ai-pi-1.0/' + name for name in
          ['spec.md', 'design.md', 'issues/14-google-images-live-and-catalog.md']]
write('source-hashes.json', {p: sha(root / p) for p in sorted(set(paths)) if (root / p).is_file()})
write('catalog-snapshot.json', read(root / 'ai/release/catalog-snapshot.json'))
write('matrix.json', read(root / 'ai/live/support-matrix.json'))
# audit.json is independently regenerated; hashing it would form a cycle.
write('sha256.json', {str(p.relative_to(folder)): sha(p) for p in sorted(folder.rglob('*'))
      if p.is_file() and str(p.relative_to(folder)) not in {'sha256.json', 'audit.json'}})
