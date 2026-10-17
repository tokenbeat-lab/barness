"""Preserve owned default E2E outputs without touching ROOT's shared directory.

This separately sealed operation keeps the already executed 11-tree historical
sealer and its hash immutable. Both receipts remain independently repeatable.
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
SOURCE = W / '.evidence/barness-ai'


def digest(path):
    value = hashlib.sha256()
    with path.open('rb') as stream:
        while part := stream.read(1 << 20):
            value.update(part)
    return value.hexdigest()


def inventory():
    result = {}
    for parent, dirs, files in os.walk(SOURCE, followlinks=False):
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
            result[str(path.relative_to(SOURCE))] = item
    return dict(sorted(result.items()))


summary = json.loads((W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes())
assert summary == dict(exitCode=0, passVerdict=True, reportAvailable=True)
root_mode = stat.S_IMODE(SOURCE.lstat().st_mode)
before = inventory()
runs = sorted(p.name for p in SOURCE.iterdir())
assert len(runs) == 34 and all((SOURCE / name).is_dir() for name in runs), runs
out = E / 'default-e2e-history'
out.mkdir(exist_ok=False)
inv = out / 'files.json'
inv.write_text(json.dumps(before, indent=2) + '\n')
archive = out / 'complete.tar.gz'
with tarfile.open(archive, 'w:gz', compresslevel=1, dereference=False) as tf:
    tf.add(SOURCE, arcname='barness-ai', recursive=False)
    for rel in before:
        tf.inodes.clear()
        tf.add(SOURCE / rel, arcname='barness-ai/' + rel, recursive=False)

with tarfile.open(archive, 'r:gz') as tf:
    members = tf.getmembers()
    names = [m.name for m in members]
    assert len(names) == len(set(names)) == len(before) + 1
    assert set(names) == {'barness-ai'} | {'barness-ai/' + p for p in before}
    assert members[0].isdir() and members[0].mode == root_mode
    for member in members[1:]:
        rel = member.name.removeprefix('barness-ai/')
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
assert inventory() == before
assert stat.S_IMODE(SOURCE.lstat().st_mode) == root_mode
receipt = dict(schema=1, finalReleaseEvidence=False,
    scope='34 owned early focused/default E2E originals, mixed historical statuses',
    originalWorkerPath=str(SOURCE), rootSharedDirectoryUnmodified=True,
    entries=len(before), rootMode=root_mode, runs=runs,
    inventory=str(inv.relative_to(E)), inventorySHA256=digest(inv),
    archive=str(archive.relative_to(E)), archiveSHA256=digest(archive),
    completeMemberSetPayloadAndModeVerified=True, noLinkTargetsOrFIFOContentsRead=True,
    specialMetadata={p:v for p,v in before.items() if v['kind'] in ('symlink', 'fifo')},
    tool=dict(source='preserve-default-history.py', sha256=digest(Path(__file__))))
(E / 'default-e2e-history-preservation.json').write_text(json.dumps(receipt, indent=2) + '\n')
print('34 default-output histories preserved;', len(before), 'entries', flush=True)
