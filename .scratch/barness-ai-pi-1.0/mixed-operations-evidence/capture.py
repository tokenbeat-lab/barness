#!/usr/bin/env python3
"""Capture issue 15's audited public E2E evidence without large load payloads.

This delivery keeps its own schema so later tickets cannot change its hashes.
Run from the repo with the full run and the existing releasegate output paths.
"""
import hashlib
import json
import subprocess
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent
bundle = root / sys.argv[1]
evaluation = root / sys.argv[2]
command_results = root / sys.argv[3]
base = '211ab89d8a60e1741efb8d25e39c39d20aa6b520'


def read(path):
    return json.loads(path.read_text())


def write(name, value):
    (folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def case_records(run, case):
    paths = sorted((run / case['dir']).iterdir())
    return {**case, 'artifacts': {p.name: sha(p) for p in paths if p.is_file()},
            'records': {p.name: read(p) for p in paths if p.suffix == '.json'}}


manifest = read(bundle / 'manifest.json')
commands = read(command_results)
tested = commands['tested_commit']
if manifest['git_commit'] != tested:
    raise SystemExit('manifest differs from tested code commit')
required = {
    'go build ./...', 'go vet ./...', 'go vet -tags live ./...',
    "BARNESS_AI_PRESSURE=1 go test -race ./ai/e2e -run '^TestMixed' -count=1",
    "go test ./ai/e2e -run '^TestMixedPressure$' -count=1",
    'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1',
}
if len(commands['commands']) != len(required) or {c['command'] for c in commands['commands']} != required:
    raise SystemExit('missing required command result')
for c in commands['commands']:
    if c['exit_code'] != 0 or c['tested_commit'] != tested or c['ending_commit'] != tested:
        raise SystemExit('failed command or inconsistent tested commit')
    log = root / c['log']
    if not log.is_file() or sha(log) != c['log_sha256']:
        raise SystemExit('missing or changed command log')
    (folder / log.name).write_bytes(log.read_bytes())
write('command-results.json', commands)
full = next(c for c in commands['commands'] if c['command'].endswith('go test -race ./... -count=1'))
if full['evidence_bundle'] != str(bundle.relative_to(root)):
    raise SystemExit('input bundle did not come from recorded full run')
if any(c['status'] != 'PASS' for c in manifest['cases']):
    raise SystemExit('full run has failures or unrun cases')
if read(bundle / 'audit.json')['findings']:
    raise SystemExit('full run audit has findings')
mixed = [case_records(bundle, c) for c in manifest['cases'] if c['test'].startswith('TestMixed')]
if len(mixed) != 103:
    raise SystemExit('unexpected mixed scenario count')
resources = [{'id': c['id'], 'gauges': c['records']['resources.json']}
             for c in mixed if 'resources.json' in c['records']]
if any(any(r['gauges'].values()) for r in resources):
    raise SystemExit('resources not released')
pressure = [{'id': c['id'], **c['records']['mixed-pressure.json']}
            for c in mixed if 'mixed-pressure.json' in c['records']]
for p in pressure:
    if p['heapGrowth'] > p['heapBudget'] or p['totalAllocDuringCalls'] > p['heapBudget']:
        raise SystemExit('memory budget exceeded')
    if any(h['ratio'] > 0.75 for h in p['headroom'].values()):
        raise SystemExit('headroom insufficient')

report = read(evaluation / 'release-report.json')
if report['commit'] != tested or report['bundle'] != str(bundle.relative_to(root)):
    raise SystemExit('release evaluation differs from tested full run')
trace_ids = {'T03-mixed-operation-isolation', 'C08-mixed-lifecycle',
             'C09-mixed-observer', 'H5-mixed-design-load', 'C10-mixed-hosts'}
trace = [t for t in report['traceability'] if t['id'] in trace_ids]
if len(trace) != 5 or any(t['status'] != 'PASS' for t in trace):
    raise SystemExit('mixed traceability not proven')

default = next(c for c in commands['commands'] if c['command'] == "go test ./ai/e2e -run '^TestMixedPressure$' -count=1")
not_run = root / default['evidence_bundle']
default_manifest = read(not_run / 'manifest.json')
default_audit = read(not_run / 'audit.json')
if default_manifest['git_commit'] != tested or default_audit['findings']:
    raise SystemExit('default pressure run commit or audit differs')
if len(default_manifest['cases']) != 2 or any(c['status'] != 'NOT_RUN' for c in default_manifest['cases']):
    raise SystemExit('pressure did not default to NOT_RUN')
write('not-run.json', {'bundle': default['evidence_bundle'],
                      'manifest_sha256': sha(not_run / 'manifest.json'),
                      'manifest': default_manifest, 'audit': default_audit})

reds = []
for run_id in ['20261008T095958.781032000Z', '20261008T101950.141748000Z', '20261008T102150.354437000Z']:
    run = root / '.evidence/barness-ai' / run_id
    m = read(run / 'manifest.json')
    reds.append({'bundle': str(run.relative_to(root)), 'manifest_sha256': sha(run / 'manifest.json'),
                 'cases': [case_records(run, c) for c in m['cases']], 'audit': read(run / 'audit.json')})
write('red.json', {'runs': reds, 'meaning': 'Expected failures before policy implementation, tenant attribution fix and measured chat load reduction.'})

paths = subprocess.check_output(['git', 'diff', '--name-only', base, tested, '--',
                                 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()
source_hashes = {}
for p in paths:
    if not (root / p).is_file():
        continue
    committed = subprocess.check_output(['git', 'show', f'{tested}:{p}'], cwd=root)
    h = hashlib.sha256(committed).hexdigest()
    if sha(root / p) != h:
        raise SystemExit(f'source differs from tested commit: {p}')
    source_hashes[p] = h
write('source-hashes.json', source_hashes)
write('mixed-cases.json', {'cases': mixed})
write('resources.json', {'cases': resources, 'all_recorded_gauges_zero': True})
write('pressure.json', {'cases': pressure, 'peak_method': 'Client EOF barrier, forced GC with outputs retained, and call-window TotalAlloc as an upper bound on new reachable memory including temporary copies.'})
(folder / 'manifest.json').write_bytes((bundle / 'manifest.json').read_bytes())
(folder / 'run-audit.json').write_bytes((bundle / 'audit.json').read_bytes())
(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())
write('trace.json', {'tested_commit': manifest['git_commit'], 'items': trace,
                     'full_gates': report['gates'], 'differential': report['differential'],
                     'scope': 'Existing-bundle evaluation; command gates are NOT_RUN there and live bundle gates belong to issue 17. Actual local command results are in verification.json.'})
write('verification.json', {
    'status': 'PASS', 'issue': '15-mixed-operations-isolation-and-load', 'date': '2026-10-08',
    'base_commit': base, 'tested_commit': manifest['git_commit'],
    'bundle': str(bundle.relative_to(root)), 'manifest_sha256': sha(bundle / 'manifest.json'),
    'commands': commands['commands'],
    'passing_cases': len(manifest['cases']), 'failed_cases': 0, 'mixed_cases': len(mixed),
    'differential_cases': sum(c['id'].startswith('PIDIFF-') for c in manifest['cases']),
    'pressure_cases': sum('pressure' in c['id'] for c in manifest['cases']),
    'resource_cases': len(resources), 'run_audit_findings': 0,
    'review': read(folder / 'review-results.json'), 'versions': manifest['versions'],
    'go_version': manifest['go_version'], 'platform': manifest['platform'],
    'started_at': manifest['started_at'], 'finished_at': manifest['finished_at'],
    'live_acceptance': 'Synthetic public protocol E2E and host examples; issue 17 owns nine-route release acceptance.',
})
# Audit first, then rerun capture so audit.json is included in the delivery hash.
write('sha256.json', {p.name: sha(p) for p in sorted(folder.iterdir())
                      if p.is_file() and p.name != 'sha256.json'})
print(f'captured {len(mixed)} mixed cases; {len(manifest["cases"])} full-run PASS cases')
