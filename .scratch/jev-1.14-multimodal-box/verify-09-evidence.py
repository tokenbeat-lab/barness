#!/usr/bin/env python3
"""Seal/verify ticket 09 evidence: --record FULL_MANIFEST RACE_MANIFEST.

The ignored local bundles remain reproducible with each case's replay command.
Historical failures are sealed separately and never converted to passing runs.
"""
import base64
import hashlib
import importlib.util
import json
import math
import subprocess
import sys
from collections import Counter
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
REPORT = HERE / '09-evidence.json'
BASELINE = '58ec73c03a3d86c29f1d12ab82bed14184d77e89'
spec = importlib.util.spec_from_file_location('criteria_identity', HERE / 'verify-07-evidence.py')
prior = importlib.util.module_from_spec(spec)
spec.loader.exec_module(prior)
sha, sources, relative = prior.sha, prior.sources, prior.relative


def load(path):
    return json.loads(path.read_bytes())


def image(data, mime):
    return {'mimeType': mime, 'sha256': hashlib.sha256(base64.b64decode(data, validate=True)).hexdigest()}


def identities(body):
    if 'assets' in body:
        return prior.identities(body)[0]
    return [
        image(block['image_url'].split(';base64,', 1)[1],
              block['image_url'].split(';base64,', 1)[0][5:])
        for message in body['input'] for block in message['content'] if block['type'] == 'input_image']


def observations(records, forbidden):
    allowed = {'kind', 'time', 'call', 'attempt', 'duration', 'stopReason', 'error', 'usage'}
    for record in records:
        assert set(record) <= allowed
        assert set(record.get('error', {})) <= {'code', 'phase', 'httpStatus', 'providerRequestId', 'retryAfter'}
    text = json.dumps(records, ensure_ascii=False)
    for marker in forbidden:
        assert marker not in text, 'content in shared Observer'


def check_records(directory, result, forbidden):
    metadata = result['metadata']
    attempts = metadata.get('attempts') or []
    record_path = directory / ('observations-' + metadata['requestId'] + '.json')
    if record_path.is_file():
        records = load(record_path)
        assert Counter(record['kind'] for record in records) == {
            'call_started': 1, 'call_finished': 1,
            **({'attempt_started': len(attempts), 'attempt_finished': len(attempts)} if attempts else {})}
        terminal = records[-1]
        assert terminal['call'] == metadata and terminal.get('usage', result['usage']) == result['usage']
        if 'usage' not in terminal:
            assert result['usage']['input'] == result['usage']['output'] == result['usage']['totalTokens'] == 0
        assert terminal['stopReason'] == result['stopReason']
        assert [record['attempt'] for record in records if record['kind'] == 'attempt_finished'] == attempts
        observations(records, forbidden)
    else:
        assert directory.name.startswith('UC09-observer-'), ('missing Observer', directory)


def check_case(directory):
    assert all(value == 0 for value in load(directory / 'resources.json').values()), directory
    assertions = load(directory / 'assertions.json')['assertions']
    assert assertions and all(item['pass'] for item in assertions), directory
    requests = (load(directory / 'requests.json') or []) if (directory / 'requests.json').exists() else []
    exact = {path.stem.removeprefix('request-'): (path, load(path)) for path in directory.glob('request-*.json')}
    for capture_path in directory.glob('*-requests.json'):
        for capture in load(capture_path) or []:
            if capture not in requests:
                requests.append(capture)
    forbidden = {'synthetic private error', 'data_base64', 'image_url', '/state/photo'}
    for request_sha, (request_file, _) in exact.items():
        assert sha(request_file) == request_sha, request_file
    for capture in requests:
        if not directory.name.startswith(('UC09-reference-accounting-', 'UC09-quoted-')):
            assert any(body == capture['body'] for _, body in exact.values()), ('missing exact request', directory)
        body = capture['body']
        forbidden.update(key for key in body.get('assets', {}) if key.startswith('<image:') or len(key) > 16)
        forbidden.update(identity['sha256'] for identity in identities(body))
    for path in directory.glob('*-image-*.bin'):
        assert path.read_bytes().startswith((b'\x89PNG\r\n\x1a\n', b'\xff\xd8\xff')), path
    result_paths = sorted(set(directory.glob('*-result.json')) | set(directory.glob('result.json')))
    for path in result_paths:
        result = load(path)
        metadata = result['metadata']
        check_records(directory, result, forbidden)
        attempts = metadata.get('attempts') or []
        if result['usage']['input'] > 0:
            expected_cost = {'typesafe-system-one': 0.0001, 'openai-decisions': 0.000592}[metadata['api']]
            assert math.isclose(result['usage']['cost']['input'], expected_cost, rel_tol=0, abs_tol=1e-15), 'pinned price changed'
            assert math.isclose(result['usage']['cost']['total'], expected_cost, rel_tol=0, abs_tol=1e-15), 'duplicated image cost'
        for i, attempt in enumerate(attempts, 1):
            assert attempt['attemptId'] == metadata['requestId'] + '#' + str(i)
        relevant = [capture for capture in requests
                    if capture['path'].startswith('/' + metadata['tenantId'] + '/' + metadata['bindingId'] + '/')]
        if result['stopReason'] != 'stop':
            assert result['answers'] is None and 'imageInputs' not in result, path
            if result['usage']['input'] > 0:
                assert any(load(response).get('usage', {}).get('input_tokens') == result['usage']['input']
                           for response in directory.glob('response-*.json')), 'lost legal failure usage'
        else:
            assert attempts and (relevant or directory.name.startswith(('UC09-reference-accounting-', 'UC09-quoted-'))), path
            if directory.name.startswith(('UC09-reference-accounting-', 'UC09-quoted-')):
                assert len(requests) == 1 and identities(requests[0]['body']) == result['imageInputs']
                continue
            candidates = [body for request_sha, (_, body) in exact.items()
                          if body['model'] == metadata['modelId'] and identities(body) == result['imageInputs']
                          and (directory / ('response-' + metadata['bindingId'] + '-' + request_sha + '.json')).is_file()]
            assert candidates, ('final image identity mismatch', path)
            body = candidates[0]
            questions = body['questions']
            names = set(questions) if isinstance(questions, dict) else {q['name'] for q in questions}
            assert set(result['answers']) == names, path
            for request_sha, (_, native) in exact.items():
                if native != body or not (directory / ('response-' + metadata['bindingId'] + '-' + request_sha + '.json')).is_file():
                    continue
                response_path = directory / ('response-' + metadata['bindingId'] + '-' + request_sha + '.json')
                assert response_path.is_file(), ('missing final response', response_path)
                response = load(response_path)
                if isinstance(questions, dict):
                    for name, question in questions.items():
                        if question['type'] == 'box':
                            assert result['answers'][name] == response['answers'][name]
                        if question['type'] == 'noul':
                            assert result['answers'][name]['probability'] == response['answers'][name]['noul']
                        if question['type'] == 'score':
                            indices = prior.identities(body)[1]
                            assert response['answers'][name]['legend'] == {str(i): value for i, value in enumerate(question['criteria'])}
                            assert result['answers'][name]['legend']['0'] is None
                            assert result['answers'][name]['legendImageReferences'] == [{'imageIndex': indices[question['criteria'][0]], 'placement': 'value', 'pointer': '/0'}]
                assert result['usage']['input'] == response['usage']['input_tokens']
    if (directory / 'observer-records.json').exists():
        observations(load(directory / 'observer-records.json'), forbidden)
    if (directory / 'load.json').exists():
        value = load(directory / 'load.json')
        assert (value['calls'], value['tenants'], value['protocols'], value['heldBodiesAtEOF']) == (4, 2, 2, 4)
        assert value['totalAlloc'] < value['memoryCeiling'] == 64 << 20
        heap = value['heap']
        assert heap['peakLive'] - min(heap['baselineLive'], heap['peakLive']) < value['memoryCeiling']
    if (directory / 'allocation.json').exists():
        value = load(directory / 'allocation.json')
        assert value['postCallbackBytes'] < value['ceiling']
    return len(result_paths)


def bundle(path, final=True):
    manifest = load(path)
    artifacts = calls = ticket = 0
    for case in manifest['cases']:
        directory = path.parent / case['dir']
        assert directory.resolve().is_relative_to(path.parent.resolve())
        files = {str(item.relative_to(directory)) for item in directory.rglob('*') if item.is_file()}
        assert files == set(case['artifacts']), directory
        for name, expected in case['artifacts'].items():
            assert sha(directory / name) == expected, directory / name
            artifacts += 1
        # Legacy fixture hashes identify source bytes before recorder redaction.
        # Every written file is checked above; ticket09 fixtures contain no secrets.
        if case['id'].startswith('UC09-'):
            for name, expected in case['fixtures'].items():
                assert sha(directory / name) == expected, directory / name
        if final:
            assert case['status'] in ('PASS', 'NOT_RUN'), case['id']
        if final and case['id'].startswith('UC09-'):
            ticket += 1
            assert case['status'] == 'PASS' and case['replay']
            calls += check_case(directory)
    audit_path = path.parent / 'audit.json'
    assert load(audit_path)['findings'] == [], audit_path
    return {'manifest': relative(path), 'manifestSHA256': sha(path), 'gitCommit': manifest['git_commit'],
            'auditSHA256': sha(audit_path), 'statuses': dict(Counter(case['status'] for case in manifest['cases'])),
            'ticketCases': ticket, 'recordedResults': calls, 'verifiedArtifacts': artifacts}


def history():
    paths = ['20261016T090521.854551000Z', '20261016T091346.193090000Z',
             '20261016T091451.100123000Z', '20261016T091926.968515000Z',
             '20261016T091701.294713000Z', '20261016T092616.635786000Z',
             '20261016T092756.694222000Z', '20261016T092927.278425000Z',
             '20261016T093156.841686000Z', '20261016T093259.189528000Z',
             '20261016T093429.413111000Z', '20261016T093522.829653000Z']
    out = []
    for name in paths:
        path = ROOT / '.evidence/barness-ai' / name / 'manifest.json'
        entry = bundle(path, final=False)
        assert entry['statuses'].get('FAIL', 0) > 0
        entry['failedCases'] = [case['id'] for case in load(path)['cases'] if case['status'] == 'FAIL']
        out.append(entry)
    return out


def documentation():
    paths = [HERE / name for name in ('09-failure-modes.md', '09-verification.md', '09-code-review.md', 'verify-09-evidence.py')]
    paths.append(HERE / 'issues/09-mixed-image-runtime-resource-isolation.md')
    paths.extend(sorted((HERE / '09-logs').glob('*.log')))
    return {relative(path): sha(path) for path in paths}


def main():
    if len(sys.argv) == 4 and sys.argv[1] == '--record':
        report = {'baseline': BASELINE, 'realService': 'NOT_RUN', 'visualQuality': 'NOT_RUN',
                  'sourceFiles': sources(), 'verificationFiles': documentation(), 'failedHistory': history(),
                  'runs': {name: bundle(Path(value).resolve()) for name, value in zip(('full', 'race'), sys.argv[2:])}}
        REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')
    elif len(sys.argv) != 1:
        raise SystemExit(__doc__)
    report = load(REPORT)
    assert report['sourceFiles'] == sources(), 'Go source inventory changed'
    assert report['verificationFiles'] == documentation(), 'verification documentation changed'
    assert report['failedHistory'] == history(), 'failed history changed'
    for name, entry in report['runs'].items():
        assert bundle(ROOT / entry['manifest']) == entry
        print(name, entry['statuses'], entry['ticketCases'], 'ticket cases;', entry['verifiedArtifacts'], 'artifact hashes')
    for name in ('ai/catalog_typesafe.go', 'ai/catalog_decisions.go', 'ai/release/catalog-snapshot.json', 'ai/live/support-matrix.json'):
        original = subprocess.check_output(['git', 'show', BASELINE + ':' + name], cwd=ROOT)
        assert hashlib.sha256(original).hexdigest() == sha(ROOT / name), name
    print('Mixed images, exact wire fixtures, attempts, resources and audits verified. Live/visual quality NOT_RUN.')


if __name__ == '__main__':
    main()
