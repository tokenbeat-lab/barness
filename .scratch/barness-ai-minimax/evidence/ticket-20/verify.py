#!/usr/bin/env python3
"""Bounded read-only custody of ticket20 offline TDD; executes no captured command."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import types

BASE = Path(__file__).resolve().parent
SOURCE = 'd02b5adb862e98d73672b47bfd1e3142a2d6082d'
PINS = {
    'ticket-13/implementation/verify.py': 'b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f',
    'ticket-13/final-gate/capsule.py': '3017cba1d4ac09e0a203c68ac768a4d508b861fa862aa04c9ca4550efebd9328',
}


def readers():
    result = []
    for name, pin in PINS.items():
        path = BASE.parent / name
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        with os.fdopen(fd, 'rb') as stream:
            info = os.fstat(stream.fileno())
            if not stat.S_ISREG(info.st_mode) or info.st_size > 64 << 10:
                raise ValueError('fixed reader must be bounded regular bytes')
            raw = stream.read((64 << 10) + 1)
        if len(raw) != info.st_size or hashlib.sha256(raw).hexdigest() != pin:
            raise ValueError('fixed reader differs')
        module = types.ModuleType(name)
        module.__file__ = str(path)
        exec(compile(raw, str(path), 'exec'), module.__dict__)
        result.append(module)
    return result


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--common-root', type=Path, required=True)
    a = p.parse_args()
    b, full = readers()
    index = b.doc(b.read(BASE, 'index.json', 1 << 20))
    if type(index.get('schema')) is not int or index['schema'] != 1 or index['sourceCommit'] != SOURCE or set(index['capsules']) != {'core', 'ordinary', 'race'}:
        raise ValueError('fixed ticket20 custody identity differs')
    results = []
    for kind, cap in index['capsules'].items():
        if cap['sourceCommit'] != SOURCE or cap['archive']['file'] != '.evidence/minimax-retained/ticket-20-visual-unsent/' + kind + '.tar.gz' or cap['sourceSnapshot']['archiveMember'] != 'metadata/source-proof.json':
            raise ValueError('fixed source/archive authority differs')
        pins, seen, summary = full.verify(b, BASE, a.common_root, cap)
        for stage, expected in index['commands'].items():
            name = 'raw/' + stage + '/command.json'
            if name not in pins:
                continue
            receipt = b.doc(seen[name])
            if type(receipt.get('exitCode')) is not int or receipt['exitCode'] != expected['exitCode'] or receipt['sourceStable'] is not True:
                raise ValueError('actual numeric command/source guard differs')
        for name, expected in cap['mainManifests'].items():
            m = b.doc(seen[name])
            audit = b.doc(seen[name.removesuffix('manifest.json') + 'audit.json'])
            counts = {s: sum(c['status'] == s for c in m['cases']) for s in ('PASS', 'FAIL', 'NOT_RUN')}
            if counts != expected['statuses'] or len(m['cases']) != expected['cases'] or len(audit['findings']) != expected['auditFindings']:
                raise ValueError('actual case/audit history differs')
        proof = b.doc(seen['metadata/source-proof.json'])
        if proof['sourceCommit'] != SOURCE or proof['testedBytesEqualCommitted'] is not True or proof['mismatches'] != [] or len(proof['compilerFiles']) != 1040:
            raise ValueError('measured WIP/committed byte proof differs')
        results.append(dict(capsule=kind, **summary))
    print(json.dumps(dict(verified=True, sourceCommit=SOURCE, capsules=results,
        paidCalls=0, credentialsRead=False, scope='offline TDD history only; no current live or release qualification'), indent=2))


if __name__ == '__main__':
    main()
