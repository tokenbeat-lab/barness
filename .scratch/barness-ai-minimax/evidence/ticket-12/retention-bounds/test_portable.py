"""Public CLI resource/failure E2E; synthetic inputs never qualify live support."""
from pathlib import Path
import argparse
import gzip
import hashlib
import io
import json
import os
import subprocess
import sys
import tarfile
import tempfile

HERE = Path(__file__).resolve().parent
BASE = HERE.parent
ROOT = BASE.parents[3]
COMMON = ROOT / '.evidence/minimax-retained'
Q = BASE / 'original-preflight-source-proof/verify_q.py'
ACTUAL = BASE / 'actual-corrective-1/verify_actual.py'
RAW_INDEX = BASE / 'glm-corrective-q-9dc-audit-failure.json'
RAW_INVENTORY = BASE / 'glm-corrective-q-9dc-audit-failure-files.json.gz'
RAW_ARCHIVE = COMMON / 'glm-corrective-q-9dc-audit-failure.tar.gz'
PREPAY = BASE / 'actual-corrective-1/corrective-approved-prepay-5bf2c16.json'
PREPAY_ARCHIVE = COMMON / 'corrective-approved-prepay-5bf2c16.tar.gz'
PREPAY_INVENTORY = BASE / 'actual-corrective-1/corrective-approved-prepay-5bf2c16-files.json.gz'
ACTUAL_INDEX = BASE / 'actual-corrective-1/glm-corrective-actual-5bf2c16.json'
WRAPPER = '''import atexit,json,resource,runpy,sys
atexit.register(lambda: print('__RESOURCE__'+json.dumps({'peakRSSBytes':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss*(1 if sys.platform=='darwin' else 1024)})))
sys.argv=sys.argv[1:]
runpy.run_path(sys.argv[0],run_name='__main__')
'''


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def dump(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def invoke(label, program, args, out, expected, memory=None):
    argv = [sys.executable, '-B', '-c', WRAPPER, str(program), *map(str, args)]
    result = subprocess.run(argv, capture_output=True, text=True, timeout=90,
                            env={'PATH': os.defpath, 'PYTHONDONTWRITEBYTECODE': '1'})
    stdout = result.stdout
    resource_line = next(x for x in stdout.splitlines() if x.startswith('__RESOURCE__'))
    peak = json.loads(resource_line.removeprefix('__RESOURCE__'))['peakRSSBytes']
    combined = result.stdout + result.stderr
    ok = result.returncode != 0 and expected in combined
    if memory:
        ok = ok and peak <= memory
    safe = combined.replace(str(ROOT), '$REPO').replace(str(COMMON), '$COMMON')
    safe = safe.replace(str(out), '$TEST_OUT')
    (out / (label + '.log')).write_text(safe)
    return {'case': label, 'pass': ok, 'exitCode': result.returncode,
            'expected': expected, 'peakRSSBytes': peak,
            'maxPeakRSSBytes': memory}


def actual_fixture(folder, inventory):
    """Change only an external-index copy and its outer forensic byte pins."""
    prepay = json.loads(PREPAY.read_bytes())
    prepay['inventorySha256'] = sha(inventory.read_bytes())
    raw = (json.dumps(prepay, indent=2) + '\n').encode()
    pp = folder / 'controlled-prepay-index.json'
    pp.write_bytes(raw)
    idx = json.loads(ACTUAL_INDEX.read_bytes())
    old_archive = ACTUAL_INDEX.parent / idx['archive']
    new_archive = folder / 'controlled-actual.tar.gz'
    with tarfile.open(old_archive, 'r|gz') as old, tarfile.open(new_archive, 'w:gz') as new:
        for member in old:
            if not member.isfile() or member.size > 16 << 20:
                raise ValueError('unexpected original fixture member')
            data = old.extractfile(member).read((16 << 20) + 1)
            if member.name == 'lineage/approved-prepay-index.json':
                data = raw
                idx['files'][member.name] = {'bytes': len(raw), 'sha256': sha(raw)}
            info = tarfile.TarInfo(member.name)
            info.size = len(data)
            new.addfile(info, io.BytesIO(data))
    idx['archive'] = new_archive.name
    idx['archiveBytes'] = new_archive.stat().st_size
    idx['archiveSha256'] = sha(new_archive.read_bytes())
    index = folder / 'controlled-actual-index.json'
    dump(index, idx)
    return [index, '--prepay-index', pp, '--prepay-archive', PREPAY_ARCHIVE,
            '--prepay-inventory', inventory]


def main():
    global COMMON, RAW_ARCHIVE, PREPAY_ARCHIVE
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('out', type=Path)
    parser.add_argument('--archives', type=Path, default=COMMON)
    parser.add_argument('--historical', action='store_true')
    args = parser.parse_args()
    COMMON = args.archives.resolve()
    RAW_ARCHIVE = COMMON / RAW_ARCHIVE.name
    PREPAY_ARCHIVE = COMMON / PREPAY_ARCHIVE.name
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=False)
    results = []
    programs = [Q, ACTUAL]
    historical = args.historical
    if historical:
        with tarfile.open(HERE / 'historical-verifier-metadata.tar.gz', 'r|gz') as tar:
            for member in tar:
                if member.name in ['original-preflight-source-proof/verify_q.py',
                                   'actual-corrective-1/verify_actual.py']:
                    target = out / Path(member.name).name
                    target.write_bytes(tar.extractfile(member).read())
        programs = [out / 'verify_q.py', out / 'verify_actual.py']
    source = {p.name: sha(p.read_bytes()) for p in programs}
    dump(out / 'verifier-source.json', source)
    with tempfile.TemporaryDirectory(prefix='portable-boundary-') as name:
        folder = Path(name)
        bomb = folder / 'bomb.gz'
        with gzip.open(bomb, 'wb', compresslevel=6) as stream:
            for _ in range(96):
                stream.write(b'0' * (1 << 20))
        idx = json.loads(RAW_INDEX.read_bytes())
        bad = folder / 'bad.gz'
        bad.write_bytes(b'not gzip')
        args = ['--raw-archive', RAW_ARCHIVE, '--raw-index', RAW_INDEX, '--raw-inventory', bad]
        results.append(invoke('q-hash-before-decode', programs[0], args, out, 'original full inventory identity'))
        idx['inventorySha256'] = sha(bomb.read_bytes())
        qi = folder / 'controlled-q-index.json'
        dump(qi, idx)
        args = ['--raw-archive', RAW_ARCHIVE, '--raw-index', qi, '--raw-inventory', bomb]
        results.append(invoke('q-expanded-bomb', programs[0], args, out, 'expanded inventory bound', 128 << 20))
        results.append(invoke('actual-expanded-bomb', programs[1], actual_fixture(folder, bomb),
                              out, 'passed prepay inventory expanded bound', 192 << 20))
        concatenated = folder / 'concatenated.gz'
        with concatenated.open('wb') as target:
            for _ in range(2):
                with gzip.GzipFile(fileobj=target, mode='wb') as stream:
                    for _ in range(24):
                        stream.write(b'0' * (1 << 20))
        idx['inventorySha256'] = sha(concatenated.read_bytes())
        dump(qi, idx)
        results.append(invoke('q-concatenated-bound', programs[0],
            ['--raw-archive', RAW_ARCHIVE, '--raw-index', qi, '--raw-inventory', concatenated],
            out, 'expanded inventory bound'))
        extra = folder / 'compressed-overcap.gz'
        extra.write_bytes(b'x' * ((8 << 20) + 1))
        results.append(invoke('q-compressed-overcap', programs[0],
            ['--raw-archive', RAW_ARCHIVE, '--raw-index', qi, '--raw-inventory', extra],
            out, 'bound'))
        truncated = folder / 'truncated.gz'
        truncated.write_bytes(gzip.compress(b'{}')[:-4])
        idx['inventorySha256'] = sha(truncated.read_bytes())
        dump(qi, idx)
        results.append(invoke('q-truncated', programs[0],
            ['--raw-archive', RAW_ARCHIVE, '--raw-index', qi, '--raw-inventory', truncated],
            out, 'EOFError'))
        dump(out / 'fixture.json', {'syntheticOnly': True, 'expandedBombBytes': 96 << 20,
                                   'compressedBombBytes': bomb.stat().st_size,
                                   'compressedBombSHA256': sha(bomb.read_bytes()),
                                   'concatenatedExpandedBytes': 48 << 20,
                                   'historicalVerifier': historical})
    dump(out / 'results.json', {'cases': results, 'pass': all(x['pass'] for x in results),
                                'credentialsRead': False, 'networkUsed': False})
    print(json.dumps(results))
    return 0 if all(x['pass'] for x in results) else 1


if __name__ == '__main__':
    sys.exit(main())
