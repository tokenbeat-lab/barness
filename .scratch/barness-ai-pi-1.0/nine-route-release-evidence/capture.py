#!/usr/bin/env python3
"""Package a completed release without altering any recorded run artifact."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tarfile

if not __debug__:
    raise SystemExit('optimized Python disables integrity checks; use normal python3')

ROOT = Path(__file__).resolve().parents[3]
DEST = Path(__file__).resolve().parent


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(name, value):
    (DEST / name).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def anonymous_header(info):
    info.uid = info.gid = 0
    info.uname = info.gname = ''
    return info


parser = argparse.ArgumentParser()
parser.add_argument('run', type=Path)
args = parser.parse_args()
base = args.run.resolve()
gate = base / 'gate'
report = read(gate / 'release-report.json')
assert report['pass'] and all(g['pass'] for g in report['gates'])
assert read(base / 'gate-exit.json')['exitCode'] == 0
assert all(c['passed'] for c in report['commands'])
assert all(not a['findings'] and not a.get('error') for a in report['audits'])
assert len(report['live']['rows']) == 9
assert all(r['status'] == 'PASS' for r in report['live']['rows'])
assert all(c['exitCode'] == c['mergeExitCode'] == 0 for c in read(base / 'smoke-commands.json'))

for p in gate.iterdir():
    if p.is_file(): shutil.copyfile(p, DEST / p.name)
for name in ('gate.log', 'gate-command.json', 'gate-exit.json', 'smoke-commands.json'):
    shutil.copyfile(base / name, DEST / name)
for source, name in (('.evidence/issue17-live-race-final.log', 'live-harness-race.log'),
                     ('.evidence/issue17-node-final.log', 'oracle-node.log')):
    raw = (ROOT / source).read_text()
    (DEST / name).write_text(raw.replace(str(ROOT), '[REPO]').replace(str(Path.home()), '[HOME]'))

archives = []
for a in report['audits']:
    bundle = ROOT / a['bundle']
    role = a.get('combo') or ('chinese-effect' if 'chinese-evaluation' in a['bundle'] else ('race' if '/race/' in a['bundle'] else 'offline'))
    archive = DEST / (role + '.tar.gz')
    with tarfile.open(archive, 'w:gz', compresslevel=6) as tar:
        tar.add(bundle, arcname=a['bundle'], filter=anonymous_header)
    manifest = read(bundle / 'manifest.json')
    archives.append({'role': role, 'archive': archive.name, 'bundle': a['bundle'], 'sha256': sha(archive),
                     'manifestSHA256': sha(bundle / 'manifest.json'), 'commit': manifest.get('git_commit'),
                     'cases': len(manifest.get('cases', []))})
    if a.get('combo'):
        parent = bundle.parent
        logs = DEST / 'live-commands' / role
        logs.mkdir(parents=True, exist_ok=True)
        for name in ('command.json', 'run.log', 'merge.log'):
            shutil.copyfile(parent / name, logs / name)

offline = ROOT / report['bundle']
manifest = read(offline / 'manifest.json')
pressure, rollback = [], []
for c in manifest['cases']:
    path = offline / c['dir']
    if c['id'].startswith('E08-mixed-pressure-'):
        pressure.append({'id': c['id'], 'status': c['status'], 'report': read(path / 'mixed-pressure.json'), 'resources': read(path / 'resources.json')})
    if c['id'].startswith('E07-mixed-') and any(x in c['id'] for x in ('disabled', 'rotation', 'zero-chat')):
        rollback.append({'id': c['id'], 'status': c['status'], 'artifacts': c['artifacts']})
write('pressure.json', pressure)
write('rollback.json', rollback)

# Critical source is unchanged since the command suite's recorded commit.
# The mutable matrix is its own captured release input, never a source hash.
excluded = {'ai/README.md', 'ai/README.zh-CN.md', 'ai/live/support-matrix.json'}
tracked = set(subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', report['commit'], '--', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()) - excluded
current = set(subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()) - excluded
assert current == tracked, 'tested source inventory drift'
sources = {}
for name in sorted(tracked):
    path = ROOT / name
    assert path.is_file(), 'tested source missing'
    committed = subprocess.check_output(['git', 'show', report['commit']+':'+name], cwd=ROOT)
    assert hashlib.sha256(committed).hexdigest() == sha(path), 'tested source drift'
    sources[name] = sha(path)
write('source-hashes.json', sources)

numbers = [1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 16, 17, 18, 20, 21, 22, 23, 24]
adrs = []
for n in numbers:
    path = next((ROOT / 'docs/adr').glob(f'{n:04d}-*.md'))
    revisions = [s for s in path.read_text().splitlines() if s.startswith('## ') and ('2026-10-08' in s or '工单' in s)]
    adrs.append({'path': str(path.relative_to(ROOT)), 'sha256': sha(path), 'revisions': revisions})
write('adr-review.json', adrs)

write('manifest.json', {'schema': 1, 'package': 'issue17-nine-route-release-delivery', 'git_commit': report['commit'], 'archives': archives})
write('verification.json', {'testedCommit': report['commit'], 'liveHarnessCommit': next(a['commit'] for a in archives if a['role']=='openai-responses'),
                           'liveHarnessSHA256': sha(base / '.tools/live.test'),
                           'gateCount': len(report['gates']), 'traceItems': len(report['traceability']),
                           'offlineCases': len(manifest['cases']), 'differentialPending': report['differential']['pending'],
                           'standardsRemaining': 0, 'specRemaining': 0})
print('packaged', len(archives), 'complete audited bundles;', len(manifest['cases']), 'offline cases')
