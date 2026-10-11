#!/usr/bin/env python3
"""Freeze complete bundles, including failures; never change test artifacts."""
import gzip
import hashlib
import io
import json
from pathlib import Path
import shutil
import subprocess
import tarfile

if not __debug__:
    raise SystemExit('optimized Python disables integrity assertions; use normal python3')
ROOT = Path(__file__).resolve().parents[3]
FOLDER = Path(__file__).resolve().parent
CURRENT = ROOT / '.evidence/barness-ai-thinking-sampling/current-release'


def read(path):
    return json.loads(path.read_text())


def sha(path):
    digest = hashlib.sha256()
    with path.open('rb') as source:
        for chunk in iter(lambda: source.read(1 << 20), b''):
            digest.update(chunk)
    return digest.hexdigest()


def write(name, value):
    (FOLDER / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def bundle_under(base):
    matches = list(base.glob('*/manifest.json'))
    if len(matches) != 1:
        raise ValueError('expected exactly one complete bundle: ' + str(base))
    return matches[0].parent


def pack(role, source, manifest_path=None):
    print('packing full archive:', role, flush=True)
    target = FOLDER / (role + '.tar.gz')
    with target.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', mtime=0, filename='') as gz:
        with tarfile.open(fileobj=gz, mode='w') as tar:
            for path in [source] + sorted(source.rglob('*')):
                if path.is_symlink() or not (path.is_file() or path.is_dir()):
                    raise ValueError('unsafe source artifact: ' + str(path))
                info = tar.gettarinfo(str(path), arcname=role + '/' + str(path.relative_to(source)))
                info.uid = info.gid = 0
                info.uname = info.gname = ''
                info.mtime = 0
                if path.is_file():
                    if path.parent == source and path.name in ('release-report.json', 'release-report.md'):
                        # Failed command tails can contain private host paths.
                        # Redact only these diagnostic reports; case bytes and
                        # their original artifact hashes stay untouched.
                        data = redact_paths(path.read_bytes())
                        info.size = len(data)
                        tar.addfile(info, io.BytesIO(data))
                    else:
                        with path.open('rb') as content:
                            tar.addfile(info, content)
                else:
                    tar.addfile(info)
    manifest_path = manifest_path or source / 'manifest.json'
    return {'role': role, 'archive': target.name, 'sha256': sha(target),
            'originalBundle': str(source.relative_to(ROOT)),
            'diagnosticRedaction': 'Top-level release-report.json/md host paths only; case bytes unchanged.',
            **bundle_record(role, source, manifest_path)}


def bundle_record(role, source, manifest_path):
    manifest = read(manifest_path)
    cases = manifest.get('cases', [])
    return {'bundle': role + '/' + str(manifest_path.parent.relative_to(source)),
            'manifestSHA256': sha(manifest_path), 'commit': manifest['git_commit'],
            'cases': len(cases), 'statuses': {s: sum(c['status'] == s for c in cases) for s in sorted({c['status'] for c in cases})},
            'failedCases': [c['id'] for c in cases if c['status'] == 'FAIL']}


def redact_paths(data):
    return data.replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')


report = read(CURRENT / 'gate/release-report.json')
assert read(CURRENT / 'gate-exit.json')['exitCode'] == 1 and not report['pass']
# Freeze failures as well as passes; replay checks each original decision.
assert report['differential']['pending'] == 0
archives = [pack(role, bundle_under(CURRENT / 'gate' / role)) for role in ('offline', 'race')]
for smoke in read(CURRENT / 'smoke-commands.json'):
    archives.append(pack(smoke['combo'], ROOT / smoke['bundle']))
evaluation = ROOT / '.evidence/barness-ai/thinking-sampling-evaluation/20261011T025637.596141000Z'
archives.append(pack('chinese-effect', evaluation))
history = [pack('initial-gate', CURRENT / 'gate-initial', bundle_under(CURRENT / 'gate-initial/offline') / 'manifest.json')]
timed_out = CURRENT / 'gate-audit-timeout'
timeout_history = pack('audit-timeout-gate', timed_out, bundle_under(timed_out / 'offline') / 'manifest.json')
timeout_race = bundle_record('audit-timeout-gate', timed_out, bundle_under(timed_out / 'race') / 'manifest.json')
timeout_race.update(recordedAudit='NOT_RUN', commandStatus='FAIL',
                    reason='Go terminated the command at 11 minutes during the post-test artifact audit. All case verdicts were PASS; the command was FAIL.')
timeout_history['additionalBundles'] = [timeout_race]
history.append(timeout_history)
history_inputs = {
    'migration-red': '.evidence/barness-ai-thinking-sampling/migration/20261011T024344.369361000Z',
    'new-differential-red': '.evidence/barness-ai-thinking-sampling/new-diff-red/20261011T025545.398060000Z',
    'catalog-red': '.evidence/barness-ai/20261011T024127.166058000Z',
    'catalog-validation-red': '.evidence/barness-ai/20261011T024211.835134000Z',
    'snapshot-red': '.evidence/barness-ai/20261011T024512.928546000Z',
    'sampling-red': '.evidence/barness-ai/20261011T024902.498104000Z',
    'duration-red': '.evidence/barness-ai/20261011T024644.500804000Z',
    'lifecycle-fixture-red': '.evidence/barness-ai/20261011T025508.623788000Z',
}
for role, name in history_inputs.items():
    history.append(pack(role, ROOT / name))
history.append(pack('diagnostic-pressure-10-rounds', bundle_under(ROOT / '.evidence/barness-ai-thinking-sampling/pressure-reproduction/bundles')))
shutil.copyfile(ROOT / '.evidence/barness-ai-thinking-sampling/pressure-reproduction/command.json', FOLDER / 'pressure-reproduction-command.json')
for name in ['release-report.json', 'release-report.md', 'fixtures.json', 'catalog-snapshot.json', 'support-matrix.json', 'traceability.json', 'ledger.json']:
    (FOLDER / name).write_bytes(redact_paths((CURRENT / 'gate' / name).read_bytes()))
for name in ['gate-command.json', 'gate-exit.json', 'smoke-commands.json', 'gate-initial-command.json', 'gate-initial-exit.json', 'gate-audit-timeout-command.json', 'gate-audit-timeout-exit.json']:
    shutil.copyfile(CURRENT / name, FOLDER / name)
shutil.copyfile(ROOT / '.evidence/barness-ai-thinking-sampling/host-example.go', FOLDER / 'host-example.go.txt')
logs = FOLDER / 'logs'
logs.mkdir(exist_ok=True)
log_sources = list((ROOT / '.evidence/barness-ai-thinking-sampling').glob('*.log'))
log_sources += list(CURRENT.glob('*.log')) + list(CURRENT.glob('live/*/*.log'))
log_sources += list((ROOT / '.evidence/barness-ai/thinking-sampling-evaluation-logs').glob('*.log'))
log_sources += [ROOT / '.evidence/barness-ai-thinking-sampling/pressure-reproduction/run.log']
for i, path in enumerate(sorted(log_sources)):
    data = path.read_bytes().replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
    # Display logs keep content lines intact; whitespace-only lines carry no
    # diagnostic data. Original per-case artifacts remain byte-for-byte.
    data = b'\n'.join(line if line.strip() else b'' for line in data.split(b'\n'))
    (logs / (str(i) + '-' + path.name)).write_bytes(data)
sources = subprocess.check_output(['git', 'ls-files', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()
excluded = {'ai/README.md', 'ai/README.zh-CN.md', 'ai/live/support-matrix.json'}
for name in sources:
    if name not in excluded:
        committed = subprocess.check_output(['git', 'show', report['commit'] + ':' + name], cwd=ROOT)
        assert hashlib.sha256(committed).hexdigest() == sha(ROOT / name), 'tested source commit mismatch: ' + name
write('source-hashes.json', {name: sha(ROOT / name) for name in sources if name not in excluded})
binary = ROOT / '.evidence/barness-ai-thinking-sampling/live.test'
build = subprocess.check_output(['go', 'version', '-m', str(binary)], cwd=ROOT, text=True)
build = build.replace(str(ROOT), '[REPO]').replace(str(Path.home()), '[HOME]')
write('live-harness.json', {'sha256': sha(binary), 'buildInfo': build,
                          'note': 'Observed build used the implemented uncommitted working tree. buildInfo has no VCS provenance; the build commit cannot be inferred. Manifest git_commit records runtime checkout HEAD, not embedded build metadata. Preserve the original SHA and buildInfo without relabeling the binary as built from the final commit.',
                          'combos': [s['combo'] for s in read(CURRENT / 'smoke-commands.json')]})
write('manifest.json', {'schema': 1, 'testedCommit': report['commit'], 'archives': archives,
                        'history': history, 'initialRace': 'interrupted; NOT_RUN; included raw files cannot certify race',
                        'timedOutRace': 'All cases PASS, original audit NOT_RUN, command FAIL. Complete raw bundle preserved and independently re-audited without changing that verdict.'})
write('sha256.json', {str(p.relative_to(FOLDER)): sha(p) for p in sorted(FOLDER.rglob('*'))
                     if p.is_file() and p.name not in {'sha256.json', 'verification.json'}})
print('packed', len(archives), 'current bundles and', len(history), 'complete failure histories')
