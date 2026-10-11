#!/usr/bin/env python3
"""Verify full archives and replay the recorded gates without vendor calls."""
import argparse
import gzip
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile
import time

if not __debug__:
    raise SystemExit('optimized Python disables integrity assertions; use normal python3')
ROOT = Path(__file__).resolve().parents[3]
FOLDER = Path(__file__).resolve().parent


def read(path):
    return json.loads(path.read_text())


def sha(path):
    digest = hashlib.sha256()
    with path.open('rb') as source:
        for chunk in iter(lambda: source.read(1 << 20), b''):
            digest.update(chunk)
    return digest.hexdigest()


def relative(name):
    path = Path(name)
    assert name and not path.is_absolute() and '..' not in path.parts, 'unsafe relative path'
    return path


def same_hash(path, digest):
    assert sha(path) == digest, 'artifact changed: ' + str(path)


def extract(archive, work):
    # Streaming validation checks each header before skipping its payload;
    # do not load an unbounded member list before applying the budgets.
    with tarfile.open(archive, 'r|gz') as tar:
        names = set()
        total = 0
        for member in tar:
            name = str(relative(member.name))
            assert name not in names, 'duplicate archive member'
            names.add(name)
            assert member.isdir() or member.isfile(), 'unsafe archive member type'
            assert member.uid == member.gid == 0 and not member.uname and not member.gname, 'host ownership leaked'
            assert 0 <= member.size <= 256 << 20, 'excessive single artifact'
            total += member.size
            assert total <= 4 << 30 and len(names) <= 300000, 'archive expansion budget exceeded'
    with tarfile.open(archive, 'r:gz') as tar:
        tar.extractall(work, filter='data')


def check_bundle(bundle, entry):
    assert sha(bundle / 'manifest.json') == entry['manifestSHA256'], 'manifest changed'
    manifest = read(bundle / 'manifest.json')
    cases = manifest.get('cases', [])
    assert len(cases) == entry['cases'] and manifest['git_commit'] == entry['commit']
    statuses = {s: sum(c['status'] == s for c in cases) for s in sorted({c['status'] for c in cases})}
    assert statuses == entry['statuses'], 'case verdict counts changed'
    assert [c['id'] for c in cases if c['status'] == 'FAIL'] == entry['failedCases'], 'failure history changed'
    if entry.get('recordedAudit') == 'NOT_RUN':
        assert not (bundle / 'audit.json').exists(), 'missing historical audit was rewritten'
        assert entry['commandStatus'] == 'FAIL', 'audit timeout cannot certify a command'
    else:
        assert not read(bundle / 'audit.json')['findings'], 'recorded audit findings'
    for case in cases:
        directory = bundle / relative(case['dir'])
        assert case['artifacts'], 'case artifacts missing'
        for name, digest in case['artifacts'].items():
            assert relative(name).name == name, 'invalid artifact name'
            assert sha(directory / name) == digest, 'case artifact changed'
    result = subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(bundle)], cwd=ROOT, capture_output=True)
    assert result.returncode == 0, 'fresh artifact audit failed: ' + result.stderr.decode()


def faults():
    rejected = []
    with tempfile.TemporaryDirectory() as tmp:
        work = Path(tmp)
        for name, kind in [('../escape', 'file'), ('/absolute', 'file'), ('link', 'symlink'), ('hard', 'hardlink'), ('duplicate', 'duplicate'), ('large', 'large')]:
            archive = work / 'fault.tar.gz'
            if kind == 'large':
                info = tarfile.TarInfo(name)
                info.size = (256 << 20) + 1
                # A dishonest oversized header must fail without allocating
                # its declared payload or extracting a partial file.
                with gzip.open(archive, 'wb') as gz:
                    gz.write(info.tobuf() + bytes(1024))
            else:
                with tarfile.open(archive, 'w:gz') as tar:
                    info = tarfile.TarInfo(name)
                    if kind == 'symlink':
                        info.type, info.linkname = tarfile.SYMTYPE, '../escape'
                    elif kind == 'hardlink':
                        info.type, info.linkname = tarfile.LNKTYPE, '../escape'
                    tar.addfile(info, io.BytesIO())
                    if kind == 'duplicate':
                        tar.addfile(info, io.BytesIO())
            try:
                extract(archive, work / 'unpack')
            except (AssertionError, tarfile.ReadError):
                rejected.append(kind + ':' + name)
            else:
                raise AssertionError('unsafe fault accepted: ' + kind)
        file = work / 'changed'
        file.write_text('initial')
        digest = sha(file)
        file.write_text('changed')
        try:
            same_hash(file, digest)
        except AssertionError:
            rejected.append('changed file')
        else:
            raise AssertionError('changed file accepted')
        file.unlink()
        try:
            same_hash(file, digest)
        except FileNotFoundError:
            rejected.append('missing file')
        else:
            raise AssertionError('missing file accepted')
    return rejected


parser = argparse.ArgumentParser()
parser.add_argument('--fault-tests', action='store_true')
parser.add_argument('--fault-tests-only', action='store_true')
args = parser.parse_args()
started = time.time()
fault_results = faults() if args.fault_tests or args.fault_tests_only else []
if args.fault_tests_only:
    print(json.dumps({'rejected': fault_results}))
    raise SystemExit(0)
hashes = read(FOLDER / 'sha256.json')
files = {str(p.relative_to(FOLDER)) for p in FOLDER.rglob('*')
         if p.is_file() and p.name not in {'sha256.json', 'verification.json'}}
assert files == set(hashes), 'delivery file inventory changed'
for name, digest in hashes.items():
    same_hash(FOLDER / relative(name), digest)
manifest = read(FOLDER / 'manifest.json')
report = read(FOLDER / 'release-report.json')
assert report['commit'] == manifest['testedCommit']
assert not report['pass'] and len(report['gates']) == 10
assert read(FOLDER / 'gate-exit.json')['exitCode'] == 1
recorded_gates = {g['id']: g['pass'] for g in report['gates']}
commands = {c['id']: c['passed'] for c in report['commands']}
assert len(commands) == 4, 'command execution record incomplete'
assert recorded_gates['go-test'] == commands['go-test']
assert recorded_gates['go-test-race'] == commands['go-test-race']
assert recorded_gates['go-vet'] == (commands['go-vet'] and commands['go-vet-live'])
assert report['differential']['pending'] == 0
assert all(not a['findings'] and not a.get('error') for a in report['audits'])
sources = read(FOLDER / 'source-hashes.json')
excluded = {'ai/README.md', 'ai/README.zh-CN.md', 'ai/live/support-matrix.json'}
current = set(subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()) - excluded
assert current == set(sources), 'source inventory drift'
for name, digest in sources.items():
    assert sha(ROOT / name) == digest, 'current source drift: ' + name
    committed = subprocess.check_output(['git', 'show', report['commit'] + ':' + name], cwd=ROOT)
    assert hashlib.sha256(committed).hexdigest() == digest, 'tested source commit mismatch'
expected_roles = {'offline', 'race', 'chinese-effect'} | {s['combo'] for s in read(FOLDER / 'smoke-commands.json')}
assert len(expected_roles) == 12 and {a['role'] for a in manifest['archives']} == expected_roles
host = read(FOLDER / 'host-example-compile.json')
assert host['exitCode'] == 0
same_hash(FOLDER / 'host-example.go.txt', host['sourceSHA256'])
receipt = {'testedCommit': report['commit'], 'recordedRelease': 'FAIL', 'faultsRejected': fault_results, 'auditedBundles': []}
with tempfile.TemporaryDirectory(prefix='barness-sampling-verify-') as tmp:
    work = Path(tmp)
    bundles = {}
    for entry in manifest['archives'] + manifest['history']:
        archive = FOLDER / relative(entry['archive'])
        assert sha(archive) == entry['sha256']
        extract(archive, work)
        bundle = work / relative(entry['bundle'])
        check_bundle(bundle, entry)
        receipt['auditedBundles'].append(entry['role'])
        for additional in entry.get('additionalBundles', []):
            check_bundle(work / relative(additional['bundle']), additional)
            receipt['auditedBundles'].append(entry['role'] + '/historical-race')
        if entry in manifest['archives']:
            bundles[entry['role']] = bundle
        if entry['role'] in ('offline', 'race'):
            gate = 'go-test' if entry['role'] == 'offline' else 'go-test-race'
            assert recorded_gates[gate] == (not entry['failedCases']), 'command/case verdict drift'
    selectors = read(FOLDER / 'scenario-artifacts.json')
    case_ids = {case['id'] for case in read(bundles['offline'] / 'manifest.json')['cases']}
    assert set(selectors['requirements']) == {f'TS-R{i:02d}' for i in range(1, 13)}
    assert set(selectors['scenarios']) == {f'S{i:02d}' for i in range(1, 31)}
    for group in ('requirements', 'scenarios'):
        for selector, ids in selectors[group].items():
            # R11 is the whole gate plus nine live/effect artifacts below.
            assert ids or selector == 'TS-R11', 'empty scenario selector: ' + selector
            assert len(ids) == len(set(ids)) and set(ids) <= case_ids, 'scenario case drift: ' + selector
    effect = read(bundles['chinese-effect'] / 'evaluation-report.json')
    assert effect['status'] == 'FAIL' and effect['audit'] == 'PASS' and effect['dataset']['count'] == 24
    evaluation = subprocess.run(['go', 'run', './ai/examples/chineseeval/cmd/chineseeval', '-verify', str(bundles['chinese-effect'])], cwd=ROOT, capture_output=True)
    assert evaluation.returncode == 1, 'failed evaluation was reported as passed'
    command = ['go', 'run', './ai/release/cmd/releasegate', '-out', str(work / 'replayed'), '-bundle', str(bundles['offline']), '-evaluation', str(bundles['chinese-effect']),
               '-matrix', str(FOLDER / 'support-matrix.json'), '-trace', str(FOLDER / 'traceability.json'), '-ledger', str(FOLDER / 'ledger.json'), '-snapshot', str(FOLDER / 'catalog-snapshot.json')]
    for role, bundle in bundles.items():
        if role not in ('offline', 'chinese-effect'):
            command += ['-live', str(bundle)]
    replay = subprocess.run(command, cwd=ROOT, capture_output=True)
    assert replay.returncode == 1
    recomputed = read(work / 'replayed/release-report.json')
    command_gates = {'go-test', 'go-test-race', 'go-vet'}
    actual = {g['id']: g['pass'] for g in recomputed['gates']}
    for gate in report['gates']:
        assert actual[gate['id']] == (False if gate['id'] in command_gates else gate['pass']), 'replayed gate drift: ' + gate['id']
    assert {t['id']: t['status'] for t in report['traceability']} == {t['id']: t['status'] for t in recomputed['traceability']}, 'traceability verdict drift'
    receipt['replayedGates'] = actual
receipt['seconds'] = round(time.time() - started, 3)
(FOLDER / 'verification.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
print('verified full artifact integrity; recorded release FAIL; no vendor calls')
