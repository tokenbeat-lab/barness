#!/usr/bin/env python3
"""Seal/recheck ticket 10's public E2E and standalone Host bundles, offline."""
import argparse
import base64
import collections
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
RECEIPT = Path(__file__).with_name('10-evidence.json')


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load(path):
    return json.loads(path.read_text())


def relative(path):
    return str(path.resolve().relative_to(ROOT))


def assert_hashes(root, inventory):
    for name, digest in inventory.items():
        path = (root / name).resolve()
        assert path.is_relative_to(root.resolve()), name
        assert path.is_file() and sha(path) == digest, name
    return len(inventory)


def validate_host(root):
    manifest = load(root / 'manifest.json')
    assert manifest['mode'] == 'controlled'
    assert manifest['realService'] == manifest['visualQuality'] == 'NOT_RUN'
    assert len(manifest['cases']) == 48
    actual = {str(p.relative_to(root)) for p in root.rglob('*') if p.is_file()}
    assert actual == set(manifest['artifacts']) | {'manifest.json'}
    count = assert_hashes(root, manifest['artifacts'])
    assert not load(root / 'audit.json')['findings']
    identities = {}
    for f in load(root / 'fixtures/provenance.json'):
        identity = f['identity']
        assert sha(root / 'fixtures' / f['name']) == identity['sha256']
        assert f['geometry']['width'] == 320 and f['geometry']['height'] == 480
        identities[identity['sha256']] = identity['mimeType']
    for case in manifest['cases']:
        p = root / 'cases' / case
        r = load(p / 'report.json')
        req = load(p / 'input.json')
        assert r['status'] == 'PASS' and all(a['pass'] for a in r['assertions'])
        assert not any(load(p / 'resources.json').values())
        for image in req['images']:
            digest = hashlib.sha256(base64.b64decode(image['data'], validate=True)).hexdigest()
            assert identities[digest] == image['mimeType']
        obs = load(p / 'observations.json')
        assert obs[-1]['kind'] == 'call_finished'
        assert obs[-1]['call'] == r['result']['metadata']
        requests = load(p / 'requests.json') or []
        assert len(requests) == len(r['result']['metadata'].get('attempts', []))
        if r['error']:
            assert not r['result']['answers'] and not r['result'].get('imageInputs')
        else:
            images = r['result']['imageInputs']
            for image in images:
                assert identities[image['sha256']] == image['mimeType']
            for answer in r['view'].values():
                for ref in answer.get('legendImages', []):
                    final = images[ref['imageIndex']]
                    assert ref['sha256'] == final['sha256']
                    assert ref['mimeType'] == final['mimeType']
                    level, field = ref['pointer'].lstrip('/').split('/')
                    assert answer['legend'][level][field] is None
            if r['scenario'] == 'box':
                a = r['view']['location']
                assert a['originalPixels'] == dict(Left=200, Top=200, Right=440, Bottom=360)
                assert r['view']['exists']['probability'] == 0
                assert a['confidence'] == (None if r['entry'] == 'plain' else 0)
                assert a['searchImage'] == images[0]
        if requests:
            response = load(p / 'response.json')
            assert load(p / 'response-meta.json')['status'] == 200
            assert r['result']['usage']['input'] == response['usage']['input_tokens']
        # The same common input switches route by Target/Options alone.
        if r['route'] == 'jev' and r['scenario'] in ('context', 'reuse'):
            other = root / 'cases' / case.replace('jev-', 'decisions-', 1) / 'input.json'
            assert req == load(other)
    return count


def validate_run(path):
    manifest = load(path)
    root = path.parent
    counts = collections.Counter(c['status'] for c in manifest['cases'])
    assert not counts['FAIL']
    count = 0
    ticket = 0
    for c in manifest['cases']:
        count += assert_hashes(root / c['dir'], c['artifacts'])
        if c['id'].startswith('UC10-'):
            ticket += 1
            assert c['status'] == 'PASS'
    assert ticket == 52, ticket
    assert not load(root / 'audit.json')['findings']
    host = root / 'classifier-images-host'
    count += validate_host(host)
    return dict(manifest=relative(path), sha256=sha(path), auditSHA256=sha(root / 'audit.json'),
                hostManifestSHA256=sha(host / 'manifest.json'), statuses=dict(counts),
                ticketCases=ticket, verifiedArtifacts=count)


def sources():
    paths = list((ROOT / 'ai/examples/classifierimages').glob('*'))
    paths += [ROOT / 'ai/e2e/classifier_images_example_test.go', ROOT / 'go.mod', ROOT / 'go.sum',
              ROOT / 'ai/catalog_typesafe.go', ROOT / 'ai/release/catalog-snapshot.json',
              ROOT / 'ai/live/support-matrix.json', Path(__file__)]
    return {relative(p): sha(p) for p in paths if p.is_file()}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--record', nargs='+', type=Path)
    args = parser.parse_args()
    if args.record:
        runs = [validate_run(p.resolve()) for p in args.record]
        reds = []
        for name in ('20261016T112243.588000000Z', '20261016T112642.767633000Z'):
            path = ROOT / '.evidence/barness-ai' / name / 'manifest.json'
            manifest = load(path)
            assert any(c['status'] == 'FAIL' for c in manifest['cases'])
            for c in manifest['cases']:
                assert_hashes(path.parent / c['dir'], c['artifacts'])
            reds.append(dict(manifest=relative(path), sha256=sha(path)))
        receipt = dict(schema=1, realService='NOT_RUN', visualQuality='NOT_RUN',
                       runs=runs, reds=reds, sources=sources())
        RECEIPT.write_text(json.dumps(receipt, indent=2) + '\n')
    receipt = load(RECEIPT)
    assert_hashes(ROOT, receipt['sources'])
    for expected in receipt['runs']:
        assert validate_run(ROOT / expected['manifest']) == expected
    for red in receipt['reds']:
        assert sha(ROOT / red['manifest']) == red['sha256']
    print(json.dumps(dict(ticketCases=sum(r['ticketCases'] for r in receipt['runs']),
                          verifiedArtifacts=sum(r['verifiedArtifacts'] for r in receipt['runs']),
                          redManifests=len(receipt['reds']), realService='NOT_RUN',
                          visualQuality='NOT_RUN'), indent=2))


if __name__ == '__main__':
    main()
