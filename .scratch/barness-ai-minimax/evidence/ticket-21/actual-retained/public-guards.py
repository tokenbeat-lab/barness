"""Public CLI custody negatives; private copies only, no native/captured execution."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import types

sys.dont_write_bytecode = True
BASE = Path(__file__).resolve().parent


def fixed_reader():
    path = BASE.parent.parent / 'ticket-13/implementation/verify.py'
    with os.fdopen(os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK), 'rb') as stream:
        info = os.fstat(stream.fileno())
        if not stat.S_ISREG(info.st_mode) or info.st_size > 64 << 10: raise ValueError('bounded fixed reader required')
        raw = stream.read((64 << 10) + 1)
    if len(raw) != info.st_size or hashlib.sha256(raw).hexdigest() != 'b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f': raise ValueError('fixed reader differs')
    module = types.ModuleType('fixed_reader'); module.__file__ = str(path)
    exec(compile(raw, str(path), 'exec'), module.__dict__)
    return module


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--common-root', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    a = p.parse_args(); a.out.mkdir(parents=True, exist_ok=False, mode=0o700)
    # The test is authored before the new semantic consumer. The original
    # full-reader code and limits are unchanged; these test the public wrapper.
    cases = {'valid-control': None, 'schema-bool': 'exact integer measurement differs',
        'source-drift': 'fixed custody source/round differs', 'archive-traversal': 'fixed actual archive/source authority differs',
        'archive-size-float': 'exact integer measurement differs', 'inventory-overlimit': 'artifact size limit',
        'archive-symlink': 'too many levels of symbolic links'}
    records = []; b = fixed_reader()
    for name, reason in cases.items():
        case = a.out / name
        target = case / '.scratch/barness-ai-minimax/evidence/ticket-21/actual-retained'
        target.mkdir(parents=True, mode=0o700)
        for name_to_copy in ('verify.py', 'index.json', 'inventory.json.gz', 'source-proof.json.gz', 'closed-program-authority.json'):
            (target / name_to_copy).write_bytes(b.read(BASE, name_to_copy, 8 << 20))
        evidence = target.parent.parent
        for relative in ('ticket-13/implementation/verify.py', 'ticket-13/final-gate/capsule.py'):
            destination = evidence / relative; destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(b.read(BASE.parent.parent, relative, 64 << 10))
        index = b.doc(b.read(target, 'index.json', 1 << 20))
        common = a.common_root
        if name == 'schema-bool': index['schema'] = True
        elif name == 'source-drift': index['sourceCommit'] = '0' * 40
        elif name == 'archive-traversal': index['capsule']['archive']['file'] = '../outside.tar.gz'
        elif name == 'archive-size-float': index['capsule']['archive']['bytes'] = float(index['capsule']['archive']['bytes'])
        elif name == 'inventory-overlimit': (target / 'inventory.json.gz').write_bytes(b'0' * ((8 << 20) + 1))
        elif name == 'archive-symlink':
            common = case / 'common'
            link = common / index['capsule']['archive']['file']; link.parent.mkdir(parents=True)
            link.symlink_to(a.common_root / index['capsule']['archive']['file'])
        (target / 'index.json').write_text(json.dumps(index, indent=2) + '\n')
        argv = [sys.executable, str(target / 'verify.py'), '--common-root', str(common)]
        result = subprocess.run(argv, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            env={k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR') if k in os.environ} | {'PYTHONDONTWRITEBYTECODE': '1'}, timeout=120)
        (case / 'raw.log').write_bytes(result.stdout)
        passed = result.returncode == 0 if reason is None else result.returncode != 0 and reason in result.stdout.decode(errors='replace').lower()
        records.append(dict(case=name, exitCode=result.returncode, expectedRejection=reason,
            outputSHA256=hashlib.sha256(result.stdout).hexdigest(), passed=passed))
    (a.out / 'results.json').write_text(json.dumps(records, indent=2) + '\n')
    print(json.dumps(records, indent=2))
    sys.exit(0 if all(v['passed'] for v in records) else 1)


if __name__ == '__main__': main()
