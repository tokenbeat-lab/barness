#!/usr/bin/env python3
"""Verify preserved pressure evidence, including FAIL runs, without network."""
import hashlib
import json
from pathlib import Path, PurePosixPath
import zipfile

HERE = Path(__file__).resolve().parent

def digest_file(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()

def digest_member(archive, name):
    with archive.open(name) as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()

def safe_name(name):
    path = PurePosixPath(name)
    if path.is_absolute() or '..' in path.parts or '\\' in name:
        raise ValueError('unsafe archive path')
    return str(path)

def verify():
    for line in (HERE / 'SHA256SUMS').read_text().splitlines():
        want, name = line.split('  ', 1)
        if digest_file(HERE / safe_name(name)) != want:
            raise ValueError('outer artifact digest mismatch: ' + name)
    meta = json.loads((HERE / 'manifest.json').read_text())
    root = HERE.parents[2]
    for artifact in meta.get('referencedArtifacts', []):
        if digest_file(root / safe_name(artifact['path'])) != artifact['sha256']:
            raise ValueError('referenced smoke archive changed')
    totals = {'bundles': 0, 'cases': 0, 'historicalFailedCases': 0}
    for artifact in meta['archives']:
        with zipfile.ZipFile(HERE / artifact['file']) as archive:
            names = archive.namelist()
            if len(names) != len(set(names)):
                raise ValueError('duplicate archive entry')
            for name in names:
                safe_name(name)
            for bundle in artifact.get('bundles', []):
                prefix = safe_name(bundle['path']) + '/'
                raw = archive.read(prefix + 'manifest.json')
                if hashlib.sha256(raw).hexdigest() != bundle['manifestSHA256']:
                    raise ValueError('original manifest changed')
                audit_raw = archive.read(prefix + 'audit.json')
                if hashlib.sha256(audit_raw).hexdigest() != bundle['auditSHA256']:
                    raise ValueError('original audit changed')
                if json.loads(audit_raw)['findings']:
                    raise ValueError('audit findings retained in accepted bundle')
                manifest = json.loads(raw)
                cases = manifest['cases']
                counts = {}
                for case in cases:
                    status = case['status']
                    counts[status] = counts.get(status, 0) + 1
                    case_prefix = prefix + safe_name(case['dir']) + '/'
                    listed = case['artifacts']
                    actual = {n[len(case_prefix):] for n in names if n.startswith(case_prefix) and not n.endswith('/')}
                    if actual != set(listed):
                        raise ValueError('case artifact inventory differs')
                    for name, want in listed.items():
                        if digest_member(archive, case_prefix + safe_name(name)) != want:
                            raise ValueError('case artifact digest differs')
                if len(cases) != bundle['cases'] or counts != bundle['statuses']:
                    raise ValueError('original verdicts changed')
                if artifact['role'] == 'full-gate' and counts.get('FAIL', 0):
                    raise ValueError('current full gate contains failed cases')
                if artifact['role'] == 'investigation-history':
                    totals['historicalFailedCases'] += counts.get('FAIL', 0)
                totals['bundles'] += 1
                totals['cases'] += len(cases)
    if totals['historicalFailedCases'] == 0:
        raise ValueError('first failure history missing')
    return {'verified': True, **totals, 'issueResolved': False, 'naturalHistoricalCause': 'UNCONFIRMED'}

if __name__ == '__main__':
    print(json.dumps(verify(), indent=2))
