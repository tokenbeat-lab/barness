#!/usr/bin/env python3
"""Check delivered hashes and optionally replay artifact gates without vendor calls."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import stat
import subprocess
import tempfile
import zipfile

FOLDER = Path(__file__).resolve().parent
ROOT = FOLDER.parents[2]


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1 << 20), b''):
            h.update(chunk)
    return h.hexdigest()


def require(ok, message):
    if not ok:
        raise ValueError(message)


def extract(archive, work):
    with zipfile.ZipFile(archive) as z:
        seen, total = set(), 0
        require(len(z.infolist()) <= 250000, 'archive member budget')
        for item in z.infolist():
            path = Path(item.filename)
            require(not path.is_absolute() and '..' not in path.parts and str(path) not in ['', '.']
                    and item.filename not in seen, 'invalid or duplicate archive path')
            seen.add(item.filename)
            require(not stat.S_ISLNK(item.external_attr >> 16), 'archive symlink')
            total += item.file_size
            require(item.file_size <= 256 << 20 and total <= 3 << 30, 'archive byte budget')
            target = work / path
            if item.is_dir():
                target.mkdir(parents=True, exist_ok=True)
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            with z.open(item) as source, target.open('xb') as output:
                shutil.copyfileobj(source, output, 1 << 20)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--replay', action='store_true')
    args = p.parse_args()
    hashes = json.loads((FOLDER / 'sha256.json').read_text())
    actual = {str(f.relative_to(FOLDER)) for f in FOLDER.rglob('*') if f.is_file()
              and f.name not in ['sha256.json', 'verification.json']}
    require(actual == set(hashes), 'delivery inventory changed')
    for name, want in hashes.items():
        path = Path(name)
        require(not path.is_absolute() and '..' not in path.parts, 'unsafe inventory path')
        require(sha(FOLDER / path) == want, 'delivery hash mismatch: ' + name)
    recorded = json.loads((FOLDER / 'release-report.json').read_text())
    manifest = json.loads((FOLDER / 'manifest.json').read_text())
    delivery = json.loads((FOLDER / 'delivery-verdict.json').read_text())
    require(delivery['status'] == 'FAIL' and not delivery['completed'] and delivery['blockers'], 'missing known-failure delivery block')
    for blocker in delivery['blockers']:
        require(sha(ROOT / blocker['issue']) == blocker['issueSHA256'], 'historical blocker changed; reevaluate delivery')
        require(sha(ROOT / blocker['originalFailureArchive']) == blocker['archiveSHA256'], 'original failure evidence changed')
    result = {'recordedRelease': recorded['pass'], 'deliveryStatus': delivery['status'], 'deliveryHashes': len(hashes), 'replayed': False}
    if args.replay:
        for name, expected_hash in json.loads((FOLDER / 'source-hashes.json').read_text()).items():
            path = Path(name)
            require(not path.is_absolute() and '..' not in path.parts, 'unsafe source path')
            require(sha(ROOT / path) == expected_hash, 'tested source has changed: ' + name)
        env = {k: os.environ[k] for k in ['PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH'] if k in os.environ}
        with tempfile.TemporaryDirectory(prefix='barness-ten-route-') as temp:
            work = Path(temp)
            for name in ['offline.zip', 'race.zip', 'live-evaluation.zip']:
                extract(FOLDER / name, work)
            bundles = {r['role']: work / r['archivePath'] for r in manifest['bundles']}
            for entry in manifest['bundles']:
                bundle = bundles[entry['role']]
                require(sha(bundle / 'manifest.json') == entry['manifestSHA256'], 'manifest hash changed')
                require(sha(bundle / 'audit.json') == entry['auditSHA256'], 'recorded audit changed')
            command = ['go', 'run', './ai/release/cmd/releasegate', '-bundle', str(bundles['offline']),
                       '-race-bundle', str(bundles['race']), '-out', str(work / 'verdict'),
                       '-evaluation-jev', str(bundles['jev']), '-evaluation-luna', str(bundles['luna'])]
            for flag, name in [('matrix', 'support-matrix.json'), ('trace', 'traceability.json'),
                               ('ledger', 'ledger.json'), ('snapshot', 'catalog-snapshot.json')]:
                command += ['-' + flag, str(FOLDER / name)]
            for role, path in bundles.items():
                if role not in ['offline', 'race', 'jev', 'luna']:
                    command += ['-live', str(path)]
            proc = subprocess.run(command, cwd=ROOT, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=600)
            require(proc.returncode == 1, 'replay must preserve unexecuted command gates')
            replay = json.loads((work / 'verdict/release-report.json').read_text())
            commands = {'go-test', 'go-test-race', 'go-vet'}
            expected = {g['id']: False if g['id'] in commands else g['pass'] for g in recorded['gates']}
            require({g['id']: g['pass'] for g in replay['gates']} == expected, 'artifact verdict drift')
            require({t['id']: t['status'] for t in replay['traceability']} ==
                    {t['id']: t['status'] for t in recorded['traceability']}, 'traceability drift')
            require({e['name']: e['status'] for e in replay['evidence']} ==
                    {e['name']: e['status'] for e in recorded['evidence']}, 'evidence status drift')
            result.update(replayed=True, gates=expected, auditedBundles=len(replay['audits']))
    (FOLDER / 'verification.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result))


if __name__ == '__main__':
    main()
