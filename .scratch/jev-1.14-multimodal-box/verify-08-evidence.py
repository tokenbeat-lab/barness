#!/usr/bin/env python3
"""Seal/verify ticket 08 controlled-HTTP evidence. --record FULL RACE manifests.

Ticket 07's identity traversal remains a historical implementation; import its
pure identity/hash helpers so this ticket can add Box checks without modifying
an already sealed verifier or duplicating image traversal rules.
"""
import importlib.util
import json
import math
import hashlib
import subprocess
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).parent
REPORT = HERE / '08-evidence.json'
BASELINE = '8ecfdecb8a653e59199fa22378fa75a63939a38b'
spec = importlib.util.spec_from_file_location('criteria_evidence', HERE / 'verify-07-evidence.py')
prior = importlib.util.module_from_spec(spec)
spec.loader.exec_module(prior)
sha, relative, sources, identities = prior.sha, prior.relative, prior.sources, prior.identities


def unit(value):
    return type(value) in (float, int) and math.isfinite(value) and 0 <= value <= 1


def box_valid(answer):
    assert set(answer) == {'type', 'box', 'confidence'} and answer['type'] == 'box'
    box = answer['box']
    assert set(box) == {'top', 'bottom', 'left', 'right'}
    assert all(unit(edge) for edge in box.values())
    assert box['left'] < box['right'] and box['top'] < box['bottom']
    assert answer['confidence'] is None or unit(answer['confidence'])


def contains_image(value, assets):
    if isinstance(value, str):
        return value in assets
    if isinstance(value, list):
        return any(contains_image(item, assets) for item in value)
    if isinstance(value, dict):
        return any(contains_image(item, assets) for item in value.values())
    return False


def check_call(directory):
    result = json.loads((directory / 'result.json').read_text())
    requests = json.loads((directory / 'requests.json').read_text()) or []
    successful = result['stopReason'] == 'stop'
    for name in ('input.json', 'error.json', 'binding.json', 'host-reads.json',
                 'resources.json', 'observer-records.json'):
        assert (directory / name).is_file(), (directory, name)
    assert all(value == 0 for value in json.loads((directory / 'resources.json').read_text()).values())
    assert len(result['metadata'].get('attempts') or []) == len(requests), directory
    if not successful:
        assert result['answers'] is None and 'imageInputs' not in result, directory
    if not requests:
        assert not successful, directory
        # Scope/capability failures happen before credentials; final Hook
        # failures may legitimately occur after a single credential pin.
        if not result['metadata']['resolved'] or directory.name.startswith('UC08-capability-'):
            assert json.loads((directory / 'host-reads.json').read_text())['credentials'] == 0
        return
    assert len(requests) == 1, 'never split batches or replay 2xx'
    final = requests[0]['body']
    expected, _ = identities(final)
    scripts = json.loads((directory / 'response-script.json').read_text())
    if successful:
        assert result.get('imageInputs', []) == expected, directory
        assert set(result['answers']) == set(final['questions']), directory
        native = json.loads(scripts[-1]['script'])
        for name, q in final['questions'].items():
            answer = result['answers'][name]
            assert answer['type'] == {'box': 'box', 'noul': 'bool', 'choice': 'choice', 'score': 'score'}[q['type']]
            if q['type'] == 'box':
                assert set(q) == {'type', 'instructions'}, directory
                assert contains_image(final['state'], final.get('assets', {})) or contains_image(q['instructions'], final.get('assets', {})), 'Box visibility'
                box_valid(answer)
                assert answer == native['answers'][name], 'Box values transformed'
            if q['type'] == 'noul':
                assert answer == {'type': 'bool', 'probability': native['answers'][name]['noul']}, 'existence thresholded'
    else:
        # JSON-invalid response literals cannot provide trustworthy usage.
        try:
            native = json.loads(scripts[-1]['script'], parse_constant=lambda value: (_ for _ in ()).throw(ValueError(value)))
        except ValueError:
            native = None
        if native and 'usage' in native:
            assert result['usage']['input'] == native['usage']['input_tokens']
            assert result['usage']['output'] == native['usage']['output_tokens']
    shared = json.dumps(result['metadata']) + (directory / 'observer-records.json').read_text()
    assert all(item['sha256'] not in shared for item in expected), directory
    assert all(key not in shared and key not in json.dumps(result) for key in final.get('assets', {})), directory


def verify_bundle(path):
    manifest = json.loads(path.read_text())
    assert manifest['cases'] and all(case['status'] in ('PASS', 'NOT_RUN') for case in manifest['cases'])
    artifacts = cases = calls = 0
    for case in manifest['cases']:
        directory = path.parent / case['dir']
        for name, expected in case['artifacts'].items():
            assert sha(directory / name) == expected, directory / name
            artifacts += 1
        if not case['id'].startswith('UC08-'):
            continue
        cases += 1
        assert case['status'] == 'PASS' and case.get('replay'), case['id']
        assertions = json.loads((directory / 'assertions.json').read_text())['assertions']
        assert assertions and all(item['pass'] for item in assertions), directory
        if (directory / 'result.json').is_file():
            calls += 1
            check_call(directory)
        else:
            assert case['id'].startswith(('UC08-catalog-', 'UC08-closed-json-')), directory
    assert cases > 0 and calls > 0, 'missing Box evidence'
    audit = path.parent / 'audit.json'
    assert json.loads(audit.read_text())['findings'] == [], audit
    return {'manifest': relative(path), 'manifestSHA256': sha(path),
            'audit': relative(audit), 'auditSHA256': sha(audit),
            'gitCommit': manifest['git_commit'],
            'statuses': dict(Counter(case['status'] for case in manifest['cases'])),
            'ticketCases': cases, 'publicCalls': calls, 'verifiedArtifacts': artifacts}


def documentation():
    paths = list(HERE.glob('08-*.md')) + list((HERE / '08-logs').glob('*.txt'))
    paths += [Path(__file__), HERE / 'verify-07-evidence.py',
              HERE / 'issues/08-jev-box-and-existence-answers.md',
              ROOT / 'docs/adr/0021-barness-ai-typesafe-unary-classification.md',
              ROOT / 'docs/barness-ai/jev-box.md']
    return {relative(path): sha(path) for path in sorted(paths)}


def verify_red_history():
    path = ROOT / '.evidence/barness-ai/20261016T083409.188062000Z/manifest.json'
    manifest = json.loads(path.read_text())
    assert len(manifest['cases']) == 2
    artifacts = 0
    for case in manifest['cases']:
        assert case['status'] == 'FAIL' and case['id'].startswith('UC08-public-')
        for name, expected in case['artifacts'].items():
            assert sha(path.parent / case['dir'] / name) == expected
            artifacts += 1
    assert json.loads((path.parent / 'audit.json').read_text())['findings'] == []
    return {'manifest': relative(path), 'manifestSHA256': sha(path),
            'cases': 2, 'verifiedArtifacts': artifacts}


def main():
    if len(sys.argv) == 4 and sys.argv[1] == '--record':
        report = {'baseline': BASELINE, 'realService': 'NOT_RUN', 'visualQuality': 'NOT_RUN',
                  'sourceFiles': sources(), 'verificationFiles': documentation(),
                  'redHistory': verify_red_history(),
                  'runs': {name: verify_bundle(Path(path).resolve()) for name, path in zip(('full', 'race'), sys.argv[2:])}}
        REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')
    elif len(sys.argv) != 1:
        raise SystemExit(__doc__)
    report = json.loads(REPORT.read_text())
    assert report['sourceFiles'] == sources(), 'source inventory changed'
    assert report['redHistory'] == verify_red_history(), 'red artifacts changed'
    for name, expected in report['verificationFiles'].items():
        assert sha(ROOT / name) == expected, name
    for name, entry in report['runs'].items():
        assert verify_bundle(ROOT / entry['manifest']) == entry, name
        print(f"{name}: {entry['statuses']}, {entry['ticketCases']} ticket cases, {entry['verifiedArtifacts']} artifact hashes verified")
    for name in ('ai/catalog_typesafe.go', 'ai/catalog_decisions.go', 'ai/release/catalog-snapshot.json', 'ai/live/support-matrix.json'):
        original = subprocess.check_output(['git', 'show', BASELINE + ':' + name], cwd=ROOT)
        assert hashlib.sha256(original).hexdigest() == sha(ROOT / name), name
    print('Box equality/visibility, final images, atomic usage, observations, sources and zero-finding audits verified; live NOT_RUN')


if __name__ == '__main__':
    main()
