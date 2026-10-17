"""Seal remaining ticket13 historical trees after the actual final CLI ends.

These are immutable failed/superseded receipts, never current release proof.
Archive special negative fixtures as lstat metadata, without following links or
opening FIFO/device/socket contents. Authoritative report validators are intact.
"""
import hashlib
import json
import os
from pathlib import Path
import stat
import tarfile

W = Path('/private/tmp/barness-glm-13')
R = Path('/Users/cyber/RestoX/barness')
E = R / '.evidence/barness-ai-glm/ticket-13'
NAMES = (
    'glm13-current-r1', 'glm13-current-r2', 'glm13-current-r3',
    'glm13-legacy-current-r1', 'glm13-legacy-current-r2',
    'glm13-consumer-diagnostic-r1', 'glm13-parent-inventory-red',
    'glm13-gate-diagnosis-r1', 'glm13-rollback-r1',
    'glm13-rollback-r2', 'glm13-support-current',
)
OTHER = {
    'glm13-current-r4': 'current-release-seal.json',
    'glm13-corrective-current-r1': 'current-release-seal.json',
    'glm13-consumer-diagnostic-r2': 'consumer-diagnostic-r2-index.json',
    'glm13-legacy-current-r3': 'historical-r3-preservation.json',
}


def digest(path):
    value = hashlib.sha256()
    with path.open('rb') as stream:
        while part := stream.read(1 << 20):
            value.update(part)
    return value.hexdigest()


def dump(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def inventory(root):
    result = {}
    for parent, dirs, files in os.walk(root, followlinks=False):
        for name in sorted(dirs + files):
            path = Path(parent) / name
            info = path.lstat()
            item = dict(mode=stat.S_IMODE(info.st_mode))
            if stat.S_ISREG(info.st_mode):
                item.update(kind='regular', bytes=info.st_size, sha256=digest(path))
            elif stat.S_ISDIR(info.st_mode):
                item.update(kind='directory')
            elif stat.S_ISLNK(info.st_mode):
                item.update(kind='symlink', target=os.readlink(path))
            elif stat.S_ISFIFO(info.st_mode):
                item.update(kind='fifo')
            else:
                raise AssertionError(f'Unexpected special fixture: {path}')
            result[str(path.relative_to(root))] = item
    return dict(sorted(result.items()))


summary = json.loads((W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes())
assert summary == dict(exitCode=0, passVerdict=True, reportAvailable=True), summary
found = {p.name for p in (W / '.evidence').iterdir() if p.name.startswith('glm13-')}
assert found == set(NAMES) | OTHER.keys(), (found - set(NAMES) - OTHER.keys())
target = E / 'remaining-history'
target.mkdir(exist_ok=True)
records = {}
for name in NAMES:
    source = W / '.evidence' / name
    root_mode = stat.S_IMODE(source.lstat().st_mode)
    before = inventory(source)
    assert before, name
    inv = target / f'{name}-files.json'
    assert not inv.exists(), inv
    dump(inv, before)
    record = dict(scope='historical only; no current PASS claim',
        originalWorkerPath=str(source), rootMode=root_mode, entries=len(before),
        inventory=str(inv.relative_to(E)), inventorySHA256=digest(inv),
        regularFiles=sum(x['kind'] == 'regular' for x in before.values()),
        specialMetadata={p:v for p,v in before.items() if v['kind'] in ('symlink', 'fifo')},
        linkTargetsRead=False, fifoContentsRead=False)
    canonical = R / '.evidence' / name
    if canonical.exists():
        assert stat.S_IMODE(canonical.lstat().st_mode) == root_mode
        assert inventory(canonical) == before, name
        record.update(storage='existing exact canonical directory', canonicalPath=str(canonical))
    else:
        archive = target / f'{name}.tar.gz'
        assert not archive.exists(), archive
        with tarfile.open(archive, 'w:gz', compresslevel=1, dereference=False) as tf:
            tf.add(source, arcname=name, recursive=False)
            for rel in before:
                # Preserve each regular path's bytes, including hardlinked
                # fixtures, rather than emitting hardlink ownership aliases.
                tf.inodes.clear()
                tf.add(source / rel, arcname=f'{name}/{rel}', recursive=False)
        # Independently reconcile the archive against the inventory. Symlinks
        # and FIFOs are compared as metadata; extractfile is regular-only.
        with tarfile.open(archive, 'r:gz') as tf:
            members = tf.getmembers()
            assert len(members) == len(before) + 1
            names = [member.name for member in members]
            assert len(names) == len(set(names))
            assert set(names) == {name} | {f'{name}/{rel}' for rel in before}
            assert members[0].name == name and members[0].isdir()
            assert members[0].mode == root_mode
            for member in members[1:]:
                rel = member.name.removeprefix(name + '/')
                item = before[rel]
                assert member.mode == item['mode'], rel
                if item['kind'] == 'regular':
                    assert member.isfile() and member.size == item['bytes'], rel
                    value = hashlib.sha256()
                    with tf.extractfile(member) as stream:
                        while part := stream.read(1 << 20):
                            value.update(part)
                    assert value.hexdigest() == item['sha256'], rel
                elif item['kind'] == 'directory':
                    assert member.isdir(), rel
                elif item['kind'] == 'symlink':
                    assert member.issym() and member.linkname == item['target'], rel
                else:
                    assert member.isfifo(), rel
        assert inventory(source) == before, name
        assert stat.S_IMODE(source.lstat().st_mode) == root_mode
        record.update(storage='verified immutable archive',
            archive=str(archive.relative_to(E)), archiveSHA256=digest(archive),
            archivePayloadAndMetadataVerified=True)
    records[name] = record
    print(name, len(before), record['storage'], flush=True)

dump(E / 'remaining-history-preservation.json', dict(schema=1,
    scope='all otherwise-unsealed ticket13 worker evidence trees; historical only',
    allWorkerGLM13TreesAccountedFor=True, finalReleaseEvidence=False,
    noLinkTargetsOrFIFOContentsRead=True, trees=records, separatelyPreserved=OTHER,
    tool=dict(source='preserve-history.py', sha256=digest(Path(__file__)))))
print('All remaining ticket13 histories preserved', flush=True)
