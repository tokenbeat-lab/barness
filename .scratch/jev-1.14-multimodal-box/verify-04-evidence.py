#!/usr/bin/env python3
"""Record/verify ticket 04 artifacts and recompute final native image identities.

Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments
verifies the sealed inventory. The bundles remain local under ignored .evidence.
"""
import base64
import hashlib
import json
import subprocess
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REPORT = Path(__file__).with_name('04-evidence.json')
BASELINE = '490e2f901d711ded59df97f4156fa6ffb43d47f5'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def relative(path):
    return str(path.resolve().relative_to(ROOT))


def sources():
    paths = subprocess.check_output(
        ['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'],
        cwd=ROOT).decode().split('\0')
    return {name: sha(ROOT / name) for name in sorted(set(paths))
            if name.endswith('.go') or name in ('go.mod', 'go.sum')}


def identities(body):
    assets = body.get('assets', {})
    if not assets:
        return []
    seen, out = set(), []
    for key in body['state']['images']:
        if key in seen:
            continue
        seen.add(key)
        asset = assets[key]
        data = base64.b64decode(asset['data_base64'], validate=True)
        mime = {'png': 'image/png', 'jpeg': 'image/jpeg', 'jpg': 'image/jpeg'}[asset['format']]
        out.append({'mimeType': mime, 'sha256': hashlib.sha256(data).hexdigest()})
    if set(assets) != seen:
        raise ValueError('unused final asset')
    return out


def verify_result(directory, result_path, requests):
    result = json.loads(result_path.read_text())
    if result['stopReason'] != 'stop':
        assert result['answers'] is None and 'imageInputs' not in result, directory
        return
    assert requests, directory
    if result_path.name in ('result-a.json', 'result-b.json'):
        alias = 'key:' + result['metadata']['tenantId'] + '@v1'
        requests = [request for request in requests if request['key_alias'] == alias]
    final = requests[-1]['body']
    expected = identities(final)
    assert result.get('imageInputs', []) == expected, directory
    if not expected:
        assert 'imageInputs' not in result, directory
    observed = directory / 'observer-records.json'
    shared = json.dumps(result['metadata'])
    if observed.is_file():
        shared += observed.read_text()
    assert all(item['sha256'] not in shared for item in expected), directory
    for key in final.get('assets', {}):
        assert key not in shared and key not in result_path.read_text(), directory


def verify_bundle(path):
    manifest = json.loads(path.read_text())
    artifacts = ticket_cases = 0
    for case in manifest['cases']:
        directory = path.parent / case['dir']
        for name, expected in case['artifacts'].items():
            assert sha(directory / name) == expected, directory / name
            artifacts += 1
        if not case['id'].startswith('UC04-'):
            continue
        assert case['status'] == 'PASS', case['id']
        ticket_cases += 1
        request_file = directory / 'requests.json'
        requests = json.loads(request_file.read_text()) if request_file.is_file() else []
        for result_name in ('result.json', 'result-a.json', 'result-b.json'):
            result_file = directory / result_name
            if result_file.is_file():
                verify_result(directory, result_file, requests)
        assert (directory / 'assertions.json').is_file(), directory
    audit_path = path.parent / 'audit.json'
    assert json.loads(audit_path.read_text())['findings'] == [], audit_path
    return {'manifest': relative(path), 'manifestSHA256': sha(path),
            'audit': relative(audit_path), 'auditSHA256': sha(audit_path),
            'statuses': dict(Counter(case['status'] for case in manifest['cases'])),
            'ticketCases': ticket_cases, 'verifiedArtifacts': artifacts}


def main():
    if len(sys.argv) == 4 and sys.argv[1] == '--record':
        report = {'baseline': BASELINE, 'realService': 'NOT_RUN',
                  'sourceFiles': sources(),
                  'verificationFiles': {relative(path): sha(path) for path in sorted(REPORT.parent.glob('04-*.md'))}
                  | {relative(path): sha(path) for path in sorted((REPORT.parent / '04-logs').glob('*.txt'))}
                  | {relative(Path(__file__)): sha(Path(__file__))}
                  | {relative(REPORT.parent.parent / 'unary-timeout-priority/issues/01-unary-timeout-priority.md'): sha(REPORT.parent.parent / 'unary-timeout-priority/issues/01-unary-timeout-priority.md')}
                  | {relative(REPORT.parent / 'issues/04-jev-context-images-json.md'): sha(REPORT.parent / 'issues/04-jev-context-images-json.md')},
                  'runs': {name: verify_bundle(Path(path).resolve())
                           for name, path in zip(('full', 'race'), sys.argv[2:])}}
        REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')
    elif len(sys.argv) != 1:
        raise SystemExit(__doc__)
    report = json.loads(REPORT.read_text())
    assert report['sourceFiles'] == sources(), 'Go source inventory changed'
    for name, expected in report['verificationFiles'].items():
        assert sha(ROOT / name) == expected, name
    for name, entry in report['runs'].items():
        actual = verify_bundle(ROOT / entry['manifest'])
        assert actual == entry, name
        print(f"{name}: {entry['statuses']}, {entry['ticketCases']} ticket cases, "
              f"{entry['verifiedArtifacts']} artifact hashes verified")
    for name in ('ai/catalog_typesafe.go', 'ai/release/catalog-snapshot.json'):
        original = subprocess.check_output(['git', 'show', BASELINE + ':' + name], cwd=ROOT)
        assert hashlib.sha256(original).hexdigest() == sha(ROOT / name), name
    print('Final Jev identities, atomic failures, private observations, source and audits verified; live NOT_RUN')


if __name__ == '__main__':
    main()
