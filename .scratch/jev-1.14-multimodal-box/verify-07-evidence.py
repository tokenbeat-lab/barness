#!/usr/bin/env python3
"""Record/verify ticket 07 artifacts and recompute final native image identities.

Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments
verifies the sealed inventory. The bundles remain local under ignored .evidence.
This verifier deliberately preserves the prior ticket's inventory schema: each
ticket keeps its own historical seal while its final traversal evolves here.
"""
import base64
import hashlib
import json
import subprocess
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REPORT = Path(__file__).with_name('07-evidence.json')
BASELINE = '6c51ee83a270017ad0f1bcd54fb6dc9f98935b8d'


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
    seen, out, indices = set(), [], {}

    def walk(value):
        if isinstance(value, dict):
            for key in sorted(value):
                assert all(asset not in key for asset in assets), key
                walk(value[key])
        elif isinstance(value, list):
            for item in value:
                walk(item)
        elif isinstance(value, str):
            if value not in assets:
                assert all(asset not in value for asset in assets), value
                return
            if value in seen:
                return
            seen.add(value)
            indices[value] = len(out)
            asset = assets[value]
            data = base64.b64decode(asset['data_base64'], validate=True)
            mime = {'png': 'image/png', 'jpeg': 'image/jpeg', 'jpg': 'image/jpeg'}[asset['format']]
            out.append({'mimeType': mime, 'sha256': hashlib.sha256(data).hexdigest()})

    walk(body['state'])
    for name in sorted(body['questions']):
        walk(body['questions'][name]['instructions'])
        walk(body['questions'][name].get('criteria'))
    assert set(assets) == seen, 'unused final asset'
    return out, indices


def verify_result(directory, result_path, requests):
    result = json.loads(result_path.read_text())
    success = result['stopReason'] == 'stop'
    if not success:
        assert result['answers'] is None and 'imageInputs' not in result, directory
    if not requests:
        assert not success, directory
        return
    if result_path.name in ('result-a.json', 'result-b.json'):
        alias = 'key:' + result['metadata']['tenantId'] + '@v1'
        requests = [request for request in requests if request['key_alias'] == alias]
    final = requests[-1]['body']
    expected, indices = identities(final)
    if success:
        assert result.get('imageInputs', []) == expected, directory
        assert set(result['answers']) == set(final['questions']), directory
        for name, question in final['questions'].items():
            kind = {'noul': 'bool', 'choice': 'choice', 'score': 'score'}[question['type']]
            assert result['answers'][name]['type'] == kind, directory
            answer = result['answers'][name]
            if kind == 'score' and 'legend' in answer:
                refs = []

                def restore(value, pointer):
                    if isinstance(value, str) and value in indices:
                        refs.append({'imageIndex': indices[value], 'placement': 'value', 'pointer': pointer})
                        return None
                    if isinstance(value, dict):
                        return {key: restore(value[key], pointer + '/' + key.replace('~', '~0').replace('/', '~1')) for key in sorted(value)}
                    if isinstance(value, list):
                        return [restore(item, pointer + '/' + str(i)) for i, item in enumerate(value)]
                    return value

                legend = {str(i): restore(value, '/' + str(i)) for i, value in enumerate(question['criteria'])}
                assert answer['legend'] == legend, directory
                assert answer.get('legendImageReferences', []) == refs, directory
            elif kind == 'score':
                assert 'legendImageReferences' not in answer, directory
    if not expected or not success:
        assert 'imageInputs' not in result, directory
    observed = directory / 'observer-records.json'
    shared = json.dumps(result['metadata'])
    if observed.is_file():
        shared += json.dumps(json.loads(observed.read_text()))
    assert all(item['sha256'] not in shared for item in expected), directory
    for key in final.get('assets', {}):
        assert key not in shared and key not in json.dumps(result), directory


def verify_bundle(path):
    manifest = json.loads(path.read_text())
    assert manifest['cases'], path
    assert all(case['status'] in ('PASS', 'NOT_RUN') for case in manifest['cases']), path
    artifacts = ticket_cases = 0
    for case in manifest['cases']:
        directory = path.parent / case['dir']
        for name, expected in case['artifacts'].items():
            assert sha(directory / name) == expected, directory / name
            artifacts += 1
        if not case['id'].startswith('UC07-'):
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
        assertions = json.loads((directory / 'assertions.json').read_text())
        assert all(item['pass'] for item in assertions['assertions']), directory
        if (directory / 'resources.json').is_file():
            assert all(value == 0 for value in json.loads(
                (directory / 'resources.json').read_text()).values()), directory
    audit_path = path.parent / 'audit.json'
    assert ticket_cases > 0, 'missing ticket coverage'
    assert json.loads(audit_path.read_text())['findings'] == [], audit_path
    return {'manifest': relative(path), 'manifestSHA256': sha(path),
            'audit': relative(audit_path), 'auditSHA256': sha(audit_path),
            'statuses': dict(Counter(case['status'] for case in manifest['cases'])),
            'ticketCases': ticket_cases, 'verifiedArtifacts': artifacts}


def main():
    if len(sys.argv) == 4 and sys.argv[1] == '--record':
        report = {'baseline': BASELINE, 'realService': 'NOT_RUN',
                  'sourceFiles': sources(),
                  'verificationFiles': {relative(path): sha(path) for path in sorted(REPORT.parent.glob('07-*.md'))}
                  | {relative(path): sha(path) for path in sorted((REPORT.parent / '07-logs').glob('*.txt'))}
                  | {relative(Path(__file__)): sha(Path(__file__))}
                  | {name: sha(ROOT / name) for name in ('docs/adr/0021-barness-ai-typesafe-unary-classification.md', 'docs/barness-ai/jev-criteria-images.md')}
                  | {relative(REPORT.parent / 'issues/07-jev-visual-criteria-public-legends.md'): sha(REPORT.parent / 'issues/07-jev-visual-criteria-public-legends.md')},
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
