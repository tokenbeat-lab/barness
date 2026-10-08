#!/usr/bin/env python3
"""Verify portable full artifacts, re-audit and recompute evidence-only gates."""
import hashlib
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile

if not __debug__:
    raise SystemExit('optimized Python disables integrity checks; use normal python3')

ROOT = Path(__file__).resolve().parents[3]
FOLDER = Path(__file__).resolve().parent


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


for name, digest in read(FOLDER / 'sha256.json').items():
    assert sha(FOLDER / name) == digest, 'delivery artifact changed: '+name
manifest = read(FOLDER / 'manifest.json')
verification = read(FOLDER / 'verification.json')
report = read(FOLDER / 'release-report.json')
assert report['commit'] == manifest['git_commit'] == verification['testedCommit']
assert report['pass'] and all(g['pass'] for g in report['gates'])
assert len(report['gates']) == verification['gateCount'] == 10
assert all(c['passed'] for c in report['commands'])
assert read(FOLDER / 'gate-exit.json')['exitCode'] == 0
assert report['differential']['pending'] == verification['differentialPending'] == 0
assert all(t['status'] == 'PASS' for t in report['traceability'])
assert all(not a['findings'] and not a.get('error') for a in report['audits'])
assert not verification['standardsRemaining'] and not verification['specRemaining']
for name, digest in read(FOLDER / 'source-hashes.json').items():
    assert sha(ROOT / name) == digest, 'current source differs from tested source: '+name
    committed = subprocess.check_output(['git', 'show', report['commit']+':'+name], cwd=ROOT)
    assert hashlib.sha256(committed).hexdigest() == digest, 'recorded source commit mismatch'
for adr in read(FOLDER / 'adr-review.json'):
    assert sha(ROOT / adr['path']) == adr['sha256'], 'ADR revision drift'

with tempfile.TemporaryDirectory(prefix='barness-release-verify-') as work:
    work = Path(work)
    bundles = {}
    for a in manifest['archives']:
        assert sha(FOLDER / a['archive']) == a['sha256']
        with tarfile.open(FOLDER / a['archive'], 'r:gz') as tar:
            for member in tar.getmembers():
                path = Path(member.name)
                assert not path.is_absolute() and '..' not in path.parts and (member.isdir() or member.isfile()), 'unsafe archive member'
                assert member.uid == member.gid == 0 and not member.uname and not member.gname, 'archive exposes host ownership'
            tar.extractall(work, filter='data')
        bundle = work / a['bundle']
        bundles[a['role']] = bundle
        assert sha(bundle / 'manifest.json') == a['manifestSHA256']
        recorded = read(bundle / 'audit.json')
        assert not recorded['findings'], 'run audit failed'
        m = read(bundle / 'manifest.json')
        assert len(m.get('cases', [])) == a['cases']
        if a['role'] in ('offline', 'race'):
            assert all(c['status'] in ('PASS','UNSUPPORTED') for c in m['cases']), 'offline/race case failed or absent'
        for c in m.get('cases', []):
            assert c['artifacts'], 'artifact inventory absent'
            for name, digest in c['artifacts'].items():
                assert Path(name).name == name and sha(bundle / c['dir'] / name) == digest, 'case artifact changed'
        subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(bundle)], cwd=ROOT, check=True, stdout=subprocess.DEVNULL)
    assert len(bundles) == 12
    subprocess.run(['go', 'run', './ai/examples/chineseeval/cmd/chineseeval', '-verify', str(bundles['chinese-effect'])], cwd=ROOT, check=True, stdout=subprocess.DEVNULL)
    out = work / 'decision'
    command = ['go', 'run', './ai/release/cmd/releasegate', '-out', str(out), '-bundle', str(bundles['offline']), '-evaluation', str(bundles['chinese-effect']),
               '-matrix', str(FOLDER / 'support-matrix.json'), '-trace', str(FOLDER / 'traceability.json'), '-ledger', str(FOLDER / 'ledger.json'), '-snapshot', str(FOLDER / 'catalog-snapshot.json')]
    for role, bundle in bundles.items():
        if role not in ('offline', 'chinese-effect'): command += ['-live', str(bundle)]
    run = subprocess.run(command, cwd=ROOT, capture_output=True)
    recomputed = read(out / 'release-report.json')
    # -bundle deliberately cannot claim fresh command execution. All artifact
    # decisions must match, backed by the archived successful command run.
    commands = {'go-test', 'go-test-race', 'go-vet'}
    assert run.returncode == 1 and not recomputed['pass']
    assert all(g['pass'] == (g['id'] not in commands) for g in recomputed['gates']), 'artifact gate replay mismatch'

print('verified 12 complete bundles; 10 recorded gates PASS; 9 live routes PASS; zero pending')
print('fresh run: python3 '+str(FOLDER.relative_to(ROOT) / 'run-release.py')+' --live --out .evidence/<fresh-run>')
