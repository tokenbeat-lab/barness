#!/usr/bin/env python3
"""Create two explicitly derived consumer fixtures; never change native inputs."""
import hashlib
import json
import os
from pathlib import Path
import stat
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / 'ticket-13/implementation'))
import verify as bounded

root = Path(__file__).resolve().parents[4] / '.evidence/visual-unsent-tdd'
source = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()
if not source.is_relative_to(root) or source.name != 'unsent-jev-plain-cancel' or source.parent.parent.name != 'bundles' or out.parent != root or out.name not in ('jev-prepin-red-fixtures', 'jev-prepin-green-fixtures', 'jev-prepin-final-fixtures'):
    raise ValueError('fixed owned controlled fixture paths required')
out.mkdir(exist_ok=False)
sha = lambda raw: hashlib.sha256(raw).hexdigest()
original = {}
entries = total = 0
for folder, dirs, files in os.walk(source, followlinks=False):
    entries += len(dirs) + len(files)
    if entries > 32768 or len(Path(folder).relative_to(source).parts) > 32:
        raise ValueError('fixed fixture whole-shape bound')
    for name in dirs + files:
        p = Path(folder) / name
        if p.is_symlink():
            raise ValueError('links forbidden in controlled fixture source')
    for name in files:
        p = Path(folder) / name
        if not stat.S_ISREG(p.stat().st_mode) or p.stat().st_size > 16 << 20:
            raise ValueError('bounded regular controlled source required')
        original[p.relative_to(source).as_posix()] = sha(bounded.read(source, p.relative_to(source).as_posix(), 16 << 20))
        total += p.stat().st_size
        if len(original) > 32768 or total > 192 << 20: raise ValueError('fixed fixture member/payload bound')
for phase in ('capability', 'request'):
    dest = out / phase
    dest.mkdir()
    for name in original:
        p = dest / name
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_bytes(bounded.read(source, name, 16 << 20))
    changed = []
    for name in ('report.json', 'samples/v01/result.json', 'samples/v01/observations.json'):
        p = dest / name
        value = json.loads(bounded.read(dest, name, 16 << 20))
        if name == 'report.json':
            value['samples'][0]['error']['phase'] = phase
        elif name.endswith('result.json'):
            value['error']['phase'] = phase
        else:
            value[1]['error']['phase'] = phase
            value[1]['stopReason'] = 'aborted' if phase == 'request' else 'error'
        p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')
        changed.append(dict(file=name, originalSHA256=original[name], derivedSHA256=sha(bounded.read(dest, name, 16 << 20))))
    manifest = bounded.doc(bounded.read(dest, 'manifest.json', 16 << 20))
    for name in manifest['artifacts']:
        manifest['artifacts'][name] = sha(bounded.read(dest, name, 16 << 20))
    (dest / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    (out / (phase + '-provenance.json')).write_text(json.dumps(dict(
        kind='DERIVED_CONSUMER_FIXTURE_NOT_NATIVE_OBSERVED_PHASE',
        sourceReportSHA256=original['report.json'], sourceFiles=original,
        exactDelta=changed, manifestResealed=True,
        reason='Jev prepareTypeSafeImages before pin; classifier.go226/typesafe_images.go33,60; read0/unresolved/no wire retained'), indent=2) + '\n')
for name, pin in original.items():
    if sha(bounded.read(source, name, 16 << 20)) != pin:
        raise ValueError('original controlled packet changed')
print(json.dumps(dict(created=['capability', 'request'], nativeQualification=False, originalUnchanged=True)))
