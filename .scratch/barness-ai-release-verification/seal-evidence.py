#!/usr/bin/env python3
"""Seal this follow-up's actual outputs; a retained FAIL never becomes PASS."""
import argparse
import importlib.util
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
RECEIPT = HERE/'evidence.json'
PARENT = ROOT/'.evidence/release-followup'
FINAL = PARENT/'final-f29703b'
# Reuse the already-reviewed byte/manifest primitives, not its ticket-specific
# verdict rules. Captured/corrupt nested manifests stay opaque retained bytes.
PRIMITIVES = ROOT/'.scratch/jev-1.14-multimodal-box/seal-14-evidence.py'
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('ticket14_bytes', PRIMITIVES)
bytes14 = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bytes14)
raw, sha, doc = bytes14.raw, bytes14.sha, bytes14.doc
relative, bundle = bytes14.relative, bytes14.bundle
inventory, tree_paths = bytes14.inventory, bytes14.tree_paths
stored_identity, visual_bundle = bytes14.stored_identity, bytes14.visual_bundle
# Explicit ownership: other tasks' pressure evidence must not be claimed here.
NATIVE_NAMES = '''native-20261017T021242.169146Z native-20261017T021341.643649Z
native-20261017T021351.787091Z native-20261017T021359.901385Z
native-20261017T021454.049063Z native-20261017T021555.478177Z
native-20261017T021606.639125Z native-20261017T021706.353405Z
native-20261017T021759.631500Z native-20261017T022432.485454Z
native-20261017T022612.488595Z native-20261017T022622.224455Z
native-20261017T022838.956486Z native-20261017T022910.882286Z
native-20261017T023345.189293Z red-replay-20261017T023049.132600Z'''.split()
HISTORY_PATTERNS = ('bundles/*/manifest.json', '*/bundles/*/manifest.json',
                    '*/*/bundles/*/manifest.json')


def source_paths():
    return sorted(relative(p) for p in (ROOT/'ai').rglob('*.go'))


def history_packages(base):
    # Only recorder output locations, never recursively accepted fixture claims.
    paths = {p for pattern in HISTORY_PATTERNS for p in base.glob(pattern)}
    return sorted(p.parent for p in paths if p.parent.name[:8].isdigit())


def check_real_notrun(base):
    rows = []
    for directory in sorted((base/'preflight').iterdir()):
        manifests = list(directory.glob('*/manifest.json'))
        if len(manifests) != 1:
            raise ValueError('unique real-route preflight missing')
        package = manifests[0].parent
        formal = directory.name == 'typesafe-classifier-jev-1.14.0'
        report = doc(package/('multimodal-report.json' if formal else 'live-report.json'))
        budget = report['budget']
        calls, attempts = ((budget['actual']['calls'], budget['actual']['attempts']) if formal
                           else (budget['callsUsed'], budget['httpAttempts']))
        if calls != 0 or attempts != 0:
            raise ValueError('real request unexpectedly spent')
        if formal:
            if report['outcome'] != 'NOT_RUN' or report['complete']:
                raise ValueError('formal preflight status changed')
        elif any(c['outcome'] != 'NOT_RUN' for c in report['scenarios']):
            raise ValueError('route preflight status changed')
        rows.append(dict(route=directory.name, calls=calls, attempts=attempts,
                         package=relative(package)))
    if len(rows) != 11:
        raise ValueError('complete eleven-route preflight required')
    for p in base.glob('visual-*/report.json'):
        r = doc(p)
        if (r['status'] != 'NOT_RUN' or r['realQuality'] != 'NOT_RUN'
                or any(r['budget'][k] != 0 for k in ('calls', 'attempts', 'retries'))):
            raise ValueError('visual effect makes real-quality claim')
    for p in base.glob('text-*/*/evaluation-report.json'):
        r = doc(p)
        if r['status'] != 'NOT_RUN' or any(r['budget'][k] != 0
                for k in ('calls', 'attempts', 'retries')):
            raise ValueError('text effect status changed')
    return rows


def record():
    report = doc(FINAL/'gate/release-report.json')
    if {name:sha(ROOT/name) for name in source_paths()} != doc(HERE/'native-pressure/source-base.json'):
        raise ValueError('Go source changed since final controlled verification')
    if report['pass'] or any(r['status'] != 'NOT_RUN' for r in report['live']['rows']):
        raise ValueError('overall release or live state was misrepresented')
    for name in ('go-test', 'go-test-race', 'go-vet', 'p0-offline', 'differential', 'redaction-audit'):
        if not next(g['pass'] for g in report['gates'] if g['id'] == name):
            raise ValueError('required controlled gate has not passed: '+name)
    snapshot = next(s for s in report['snapshots'] if s['name'] == 'pressure-stream-integrity')
    before = PARENT/'pressure-tracker-input.md'
    if not snapshot['detail'].endswith('sha256:'+sha(before)):
        raise ValueError('original blocker snapshot drift')
    transition = doc(PARENT/'pressure-tracker-transition.json')
    after = ROOT/transition['after']['path']
    if transition['before']['sha256'] != sha(before) or transition['after']['sha256'] != sha(after):
        raise ValueError('post-run tracker transition drift')
    if sha(ROOT/transition['after']['snapshot']) != sha(after):
        raise ValueError('resolved tracker snapshot does not match original')
    if not transition['after']['resolved'] or '\nStatus: resolved\n' not in raw(after).decode():
        raise ValueError('original tracker acceptance is not resolved')
    receipt = dict(schema=1, kind='current-release-followup', overall='BLOCKED',
        realCalls=0, realAttempts=0, executionCommit=report['commit'], goContentCommit='cc78cfd',
        excludedRebuildableRootBinaries=sorted(bytes14.BINS),
        sourcePaths=source_paths(), parentEntries=sorted(p.name for p in PARENT.iterdir()
            if p.name != 'integrity'), files={}, runs=[], subprocessRuns=[], history=[],
        inventories=[], trees=[], commands=doc(FINAL/'commands.json'), gates=report['gates'],
        live=report['live'], pressureSnapshotAtStart=snapshot, pressureAcceptanceAfterRun=transition,
        preflights=check_real_notrun(FINAL),
        note='Original report and historical failures unchanged. Tracker resolved only after acceptance; '
             'this receipt never retroactively passes its snapshot or the NOT_RUN release. '
             'Captured/corrupt nested fixtures are opaque bytes, not accepted recorder verdicts.')

    def seal(path):
        receipt['files'][relative(path)] = sha(path)

    children = {p.parent for p in sorted((FINAL/'live-harness').glob('*/manifest.json'))[1:]}
    for path in bytes14.recorder_roots(FINAL):
        if path in children:
            receipt['subprocessRuns'].append(bundle(path, allow_incomplete=True))
        else:
            run = bundle(path)
            if run['auditFindings'] != 0:
                raise ValueError('current primary package audit findings')
            receipt['runs'].append(run)
    for name in ('visual-decisions', 'visual-jev'):
        receipt['inventories'].append(visual_bundle(FINAL/name))
    for p in sorted(FINAL.glob('text-*/*/files.json')):
        files = doc(p)
        inventory(p.parent, files)
        receipt['inventories'].append(dict(file=relative(p), base=relative(p.parent), count=len(files)))
    for item in doc(FINAL/'gate/fixtures.json'):
        p = ROOT/item['path']
        if len(raw(p)) != item['bytes'] or sha(p) != item['sha256']:
            raise ValueError('checked-in fixture changed')
        seal(p)
    indexes = PARENT/'integrity'
    indexes.mkdir(exist_ok=True)
    entries = sorted(PARENT.iterdir())
    if any(p.is_symlink() for p in entries):
        raise ValueError('linked top-level evidence entry')
    directories = [p for p in entries if p.is_dir() and p != indexes]
    directories += [ROOT/'.evidence/pressure-integrity'/name for name in NATIVE_NAMES]
    for i, directory in enumerate(directories):
        if not directory.is_dir() or directory.is_symlink():
            raise ValueError('missing or linked evidence root')
        if directory != FINAL:
            receipt['history'].extend(bundle(p, allow_incomplete=True) for p in history_packages(directory))
        # Whole exact tree includes failures, partials and corrupt fixture inputs.
        paths = tree_paths(directory)
        files = {name:stored_identity(directory/name) for name in sorted(paths)}
        index = indexes/(str(i).zfill(2)+'-'+directory.name+'.json')
        index.write_text(json.dumps(files, sort_keys=True, separators=(',', ':'))+'\n')
        seal(index)
        receipt['trees'].append(dict(base=relative(directory), file=relative(index), count=len(files)))
    for p in sorted(PARENT.iterdir()):
        if p.is_file():
            seal(p)
    for name in source_paths():
        seal(ROOT/name)
    for name in ('go.mod', 'go.sum', 'ai/release/traceability.json',
                 'ai/release/classifier-extensions.json', 'ai/release/catalog-snapshot.json',
                 'ai/live/support-matrix.json', 'ai/internal/testkit/release/testdata/multimodal-load.json',
                 'ai/internal/testkit/pioracle/node/package-lock.json',
                 'ai/internal/testkit/pioracle/node/runner.mjs',
                 'ai/internal/testkit/pioracle/node/package.json',
                 'ai/internal/testkit/pioracle/node/provenance.json',
                 'ai/internal/testkit/pioracle/node/model-data.sha256',
                 'ai/e2e/testdata/pidiff/ledger.json',
                 '.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md',
                 '.scratch/jev-1.14-multimodal-box/run-14-verification.py',
                 '.scratch/jev-1.14-multimodal-box/14-evidence.json',
                 '.scratch/jev-1.14-multimodal-box/issues/14-current-catalog-release-integration.md',
                 'docs/barness-ai/current-catalog-release.md'):
        seal(ROOT/name)
    for p in sorted(HERE.rglob('*')):
        if p.is_file() and p != RECEIPT:
            seal(p)
    seal(PRIMITIVES)
    RECEIPT.write_text('{\n'+',\n'.join('  '+json.dumps(k)+': '+json.dumps(v, ensure_ascii=False,
                       separators=(',', ':')) for k,v in receipt.items())+'\n}\n')
    print('sealed', len(receipt['runs']), 'primary packages,', len(receipt['history']),
          'historical packages,', len(receipt['trees']), 'complete trees; release BLOCKED')


def verify():
    receipt = doc(RECEIPT)
    inventory(ROOT, receipt['files'])
    report = doc(FINAL/'gate/release-report.json')
    if (receipt['overall'] != 'BLOCKED' or report['pass'] or receipt['realCalls'] != 0
            or receipt['realAttempts'] != 0 or receipt['executionCommit'] != report['commit']):
        raise ValueError('overall or execution identity drift')
    for name, actual in (('commands', doc(FINAL/'commands.json')), ('gates', report['gates']),
                         ('live', report['live'])):
        if receipt[name] != actual:
            raise ValueError('actual command/gate/live verdict drift')
    pressure = next(s for s in report['snapshots'] if s['name'] == 'pressure-stream-integrity')
    if (receipt['pressureSnapshotAtStart'] != pressure or receipt['pressureAcceptanceAfterRun']
            != doc(PARENT/'pressure-tracker-transition.json')):
        raise ValueError('tracker state transition drift')
    entries = list(PARENT.iterdir())
    if any(p.is_symlink() for p in entries) or sorted(p.name for p in entries
            if p.name != 'integrity') != receipt['parentEntries']:
        raise ValueError('retained parent entry set changed')
    if source_paths() != receipt['sourcePaths']:
        raise ValueError('complete Go source set changed')
    roots = {relative(p) for p in PARENT.iterdir() if p.is_dir() and p.name != 'integrity'}
    roots.update(relative(ROOT/'.evidence/pressure-integrity'/name) for name in NATIVE_NAMES)
    if len(receipt['trees']) != len(roots) or {item['base'] for item in receipt['trees']} != roots:
        raise ValueError('retained root coverage drift')
    current = {relative(p) for p in bytes14.recorder_roots(FINAL)}
    claimed = receipt['runs']+receipt['subprocessRuns']
    if len(claimed) != len(current) or {run['path'] for run in claimed} != current:
        raise ValueError('current recorder coverage drift')
    historical = {relative(p) for base in roots if ROOT/base != FINAL
                  for p in history_packages(ROOT/base)}
    if len(receipt['history']) != len(historical) or {run['path'] for run in receipt['history']} != historical:
        raise ValueError('historical recorder coverage drift')
    for run in receipt['runs']:
        if bundle(ROOT/run['path']) != run:
            raise ValueError('current verdict drift')
    for run in receipt['subprocessRuns']+receipt['history']:
        if bundle(ROOT/run['path'], allow_incomplete=True) != run:
            raise ValueError('historical verdict drift')
    for item in receipt['inventories']:
        if item.get('kind') == 'visual':
            if visual_bundle(ROOT/item['base']) != item:
                raise ValueError('visual inventory drift')
        else:
            inventory(ROOT/item['base'], doc(ROOT/item['file']))
    for item in receipt['trees']:
        base = ROOT/item['base']
        if not base.is_dir() or base.is_symlink():
            raise ValueError('missing or linked retained root')
        files = doc(ROOT/item['file'])
        if tree_paths(base) != set(files):
            raise ValueError('complete retained tree set changed')
        inventory(base, files, allow_links=True)
    if check_real_notrun(FINAL) != receipt['preflights']:
        raise ValueError('real-route NOT_RUN state drift')
    print('verified source, exact stored trees, primary inventories, actual verdicts, '
          'audits and post-run tracker transition; real evidence remains NOT_RUN')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    choice = parser.add_mutually_exclusive_group(required=True)
    choice.add_argument('--record', action='store_true')
    choice.add_argument('--verify', action='store_true')
    args = parser.parse_args()
    record() if args.record else verify()
