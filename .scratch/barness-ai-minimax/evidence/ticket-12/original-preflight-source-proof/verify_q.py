"""Read-only bounded verifier. It never extracts links or executes captured commands."""
from pathlib import Path, PurePosixPath
from collections import Counter
import argparse
import gzip
import hashlib
import io
import json
import os
import tarfile

SOURCE = '9dcf896b9bea357fd6c784acbca1b027dc96a177'
ARCHIVE_SHA = '56e08cbf7105f4fe1a1751bae42ba83988bcdbb4d1b1d443dd9bcd9a2d62ae3c'


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for raw in iter(lambda: f.read(1 << 20), b''):
            h.update(raw)
    return h.hexdigest()


def verify_raw(archive, index_path, inventory_path):
    checks = 0

    def need(ok, detail):
        nonlocal checks
        if not ok:
            raise ValueError(detail)
        checks += 1

    need(index_path.stat().st_size <= 1 << 20, 'bounded external metadata')
    idx = json.loads(index_path.read_bytes())
    with inventory_path.open('rb') as stream:
        compressed = stream.read((8 << 20) + 1)
    need(len(compressed) <= 8 << 20, 'compressed inventory bound')
    need(sha(compressed) == idx['inventorySha256'], 'original full inventory identity')
    with gzip.GzipFile(fileobj=io.BytesIO(compressed)) as stream:
        inventory_raw = stream.read((32 << 20) + 1)
    need(len(inventory_raw) <= 32 << 20, 'expanded inventory bound')
    inv = json.loads(inventory_raw)
    need(len(inv) == 27418 and idx['codeCommit'] == SOURCE and idx['actualSupport'] is False,
         'original failed source and inventory')
    need(digest(archive) == idx['archiveSha256'] == ARCHIVE_SHA and
         archive.stat().st_size == idx['archiveBytes'] == 53271279, 'raw archive identity')
    selected = {}
    seen = set()
    regular_bytes = 0
    link_bytes = 0
    links = 0
    with tarfile.open(archive, 'r|gz') as tar:
        for member in tar:
            name = member.name
            parts = PurePosixPath(name).parts
            need(name in inv and name not in seen and not name.startswith('/') and
                 '..' not in parts and len(parts) <= 24, 'safe unique archive entry')
            seen.add(name)
            pin = inv[name]
            if member.issym():
                raw = os.fsencode(member.linkname)
                need(pin.get('kind') == 'intentional-negative-symlink' and
                     member.linkname == pin['target'] and len(raw) == pin['bytes'] <= 4096
                     and sha(raw) == pin['sha256'], 'negative symlink metadata')
                # Original expandedBytes includes link text, never dereferenced bytes.
                links += 1
                link_bytes += len(raw)
                continue
            need(member.isfile() and not pin.get('kind') and member.size <= 128 << 20,
                 'bounded regular archive entry')
            f = tar.extractfile(member)
            h = hashlib.sha256()
            count = 0
            keep = len(parts) == 2 or (name.endswith('/manifest.json') and len(parts) <= 4)
            keep = keep or name == idx['results'][0]['manifest'].replace('manifest.json', 'audit.json')
            chunks = [] if keep else None
            for raw in iter(lambda: f.read(1 << 20), b''):
                count += len(raw)
                need(count <= 128 << 20, 'actual entry read bound')
                h.update(raw)
                if chunks is not None:
                    chunks.append(raw)
            need(count == pin['bytes'] and h.hexdigest() == pin['sha256'], 'exact raw entry bytes')
            regular_bytes += count
            need(regular_bytes + link_bytes <= 512 << 20, 'total expanded bound')
            if chunks is not None:
                selected[name] = b''.join(chunks)
    need(seen == set(inv) and links == idx['negativeSymlinkFixtures'] == 1 and
         len(seen) - links == idx['files'] == 27417 and
         regular_bytes + link_bytes == idx['expandedBytes'] == 253350201,
         'complete exact archive inventory')
    mn = idx['results'][0]['manifest']
    manifest = json.loads(selected[mn])
    outcomes = Counter(case['status'] for case in manifest['cases'])
    need(outcomes == {'PASS': 335, 'NOT_RUN': 62} and len(manifest['cases']) == 397,
         'original failed no-key bundle case outcomes')
    refs = 0
    base = PurePosixPath(mn).parent
    for case in manifest['cases']:
        for name, pin in case.get('artifacts', {}).items():
            key = (base / case['dir'] / name).as_posix()
            need(key in inv and inv[key]['sha256'] == pin, 'main manifest actual artifact reference')
            refs += 1
    need(refs == idx['results'][0]['artifactReferences'] == 15261, 'main artifact inventory count')
    audit = json.loads(selected[(base / 'audit.json').as_posix()])
    need(len(audit['findings']) == 2 and all(f['rule'] == 'environment' for f in audit['findings']),
         'original audit2 remains failed')
    safe_findings = []
    for finding in audit['findings']:
        rel = finding['file'].split('GLM-corrective-1-capture-large/', 1)
        need(len(rel) == 2 and rel[1] in ('assertions.json', 'new-capture-replay-exit.json'),
             'original two diagnostic finding files')
        safe_findings.append('GLM-corrective-1-capture-large/' + rel[1])
    checks_json = json.loads(selected['original-Q/checks.json'])
    need(checks_json == {'passed': False, 'commands': {'vet': 0, 'vet-live': 0, 'no-key-live': 1},
                         'credentialsRead': False}, 'original failed prerelease checks')
    for file, code in [('vet-command.json', 0), ('vet-live-command.json', 0), ('no-key-live-command.json', 1)]:
        cmd = json.loads(selected['original-Q/' + file])
        need(cmd['exitCode'] == code and cmd['credentialsRead'] is False and
             cmd['launcherAudit'] == {'passed': True, 'findings': []}, 'original command receipt')
    for file in ('prerelease-source-review.template.json', 'human-approval.template.json'):
        need(json.loads(selected['original-Q/' + file])['decision'] == 'PENDING', 'original pending template')
    need('original-Q/prerelease-source-review.json' not in inv and 'original-Q/human-approval.json' not in inv,
         'no authoritative approval in original collection')
    source = json.loads(selected['original-Q/source-at-start.json'])
    need(source['sourceCommit'] == SOURCE and len(source['files']) == 1033, 'original source manifest')
    return {'integrity': 'PASS', 'originalPreflight': 'FAIL', 'sourceCommit': SOURCE,
            'checks': checks, 'regularFiles': 27417, 'negativeSymlinkFixtures': links,
            'regularExpandedBytes': regular_bytes, 'negativeSymlinkTextBytes': link_bytes,
            'mainArtifactReferences': refs, 'cases': dict(outcomes), 'auditFindings': 2,
            'auditFindingFiles': safe_findings, 'vetExit': 0, 'vetLiveExit': 0,
            'noKeyGoExit': 1, 'sourceReview': 'PENDING', 'humanApproval': 'PENDING',
            'paidExecutionAuthorized': False, 'sourceFiles': 1033}, selected, idx


def verify_source(index_path):
    index = json.loads(index_path.read_bytes())
    archive = index_path.parent / index['archive']
    if digest(archive) != index['archiveSha256'] or archive.stat().st_size != index['archiveBytes']:
        raise ValueError('source proof archive identity')
    files = {}
    total = 0
    with tarfile.open(archive, 'r|gz') as tar:
        for member in tar:
            name = member.name
            if (not member.isfile() or name.startswith('/') or '..' in PurePosixPath(name).parts
                    or name in files or member.size > 8 << 20 or len(files) >= 1100):
                raise ValueError('source proof entry bounds')
            raw = tar.extractfile(member).read()
            total += len(raw)
            if total > 32 << 20:
                raise ValueError('source proof total bound')
            pin = index['files'][name]
            if len(raw) != pin['bytes'] or sha(raw) != pin['sha256']:
                raise ValueError('source proof member hash')
            files[name] = raw
    if set(files) != set(index['files']):
        raise ValueError('source proof omitted entry')
    manifest = json.loads(files['source-at-start.json'])
    if manifest['sourceCommit'] != SOURCE or len(manifest['files']) != 1033:
        raise ValueError('source proof original manifest')
    for name, pin in manifest['files'].items():
        if sha(files['source/' + name]) != pin:
            raise ValueError('source proof original pin')
    if json.loads(files['preflight-checks.json'])['passed'] is not False:
        raise ValueError('source proof relabelled preflight')
    for name in ('source-review.template.json', 'human-approval.template.json'):
        if json.loads(files[name])['decision'] != 'PENDING':
            raise ValueError('source proof relabelled approval')
    if index['originalRawArchive']['sha256'] != ARCHIVE_SHA:
        raise ValueError('source proof external raw boundary')
    return {'sourceByteProofVerified': True, 'committedSourceFiles': 1033,
            'sourceProofArchiveFiles': len(files), 'sourceProofExpandedBytes': total}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--raw-archive', type=Path, required=True)
    parser.add_argument('--raw-index', type=Path, required=True)
    parser.add_argument('--raw-inventory', type=Path, required=True)
    parser.add_argument('--source-index', type=Path)
    args = parser.parse_args()
    result, _, _ = verify_raw(args.raw_archive, args.raw_index, args.raw_inventory)
    if args.source_index:
        result.update(verify_source(args.source_index))
    print(json.dumps(result, ensure_ascii=False))
