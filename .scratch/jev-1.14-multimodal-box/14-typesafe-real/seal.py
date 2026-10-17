#!/usr/bin/env python3
"""Seal the authorized TypeSafe round without changing earlier evidence."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[3]
HERE = Path(__file__).resolve().parent
BASE = ROOT/'.evidence/ticket14-typesafe-r2-20261017T061932.423266Z'
OLD = ROOT/'.evidence/ticket14-typesafe-20261017T055619.971418Z'
DIAG = ROOT/'.evidence/ticket14-typesafe-matrix-repro-20261017'
RECEIPT = HERE/'evidence.json'
INDEX = ROOT/'.evidence/ticket14-typesafe-integrity'/ (BASE.name+'.json')
PRIMITIVES = HERE.parent/'seal-14-evidence.py'
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('ticket14_bytes', PRIMITIVES)
b = importlib.util.module_from_spec(spec)
spec.loader.exec_module(b)
raw, sha, doc, relative = b.raw, b.sha, b.doc, b.relative
PAID = {'protocol113':'BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER',
        'text113':'BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY',
        'protocol114':'BARNESS_MM_KEY','visual114':'BARNESS_VISUAL_KEY'}
PATTERNS = ('protocol113/*/manifest.json','protocol114/*/manifest.json','text-jev/*/manifest.json',
    'gate/offline/*/manifest.json','gate/race/*/manifest.json',
    'no-key/live-harness/*/manifest.json','no-key/preflight/*/*/manifest.json',
    'no-key/text-*/*/manifest.json')


def package(base):
    paths = list(base.glob('*/manifest.json'))
    if len(paths) != 1:
        raise ValueError('unique producer bundle missing')
    return paths[0].parent


def producer_paths(base=BASE, partial=False):
    expected = {package(base/name) for name in ('protocol113','protocol114','text-jev',
        'gate/offline','no-key/text-jev','no-key/text-luna')}
    race = list((base/'gate/race').iterdir())
    if len(race) != 1 or not race[0].is_dir():
        raise ValueError('unique race producer directory required')
    if not partial or (race[0]/'manifest.json').exists():
        expected.add(package(base/'gate/race'))
    routes = {r['combo'] for r in doc(base/'matrix-before.json')['rows']}
    preflight = base/'no-key/preflight'
    if {p.name for p in preflight.iterdir()} != routes:
        raise ValueError('mandatory preflight route set differs')
    expected.update(package(preflight/name) for name in routes)
    harness = sorted((base/'no-key/live-harness').glob('*/manifest.json'))
    if len(harness) != 3:
        raise ValueError('live harness parent/two deliberate crash children missing')
    expected.update(p.parent for p in harness)
    actual = {p.parent for pattern in PATTERNS for p in base.glob(pattern)}
    if actual != expected:
        raise ValueError('mandatory producer stage coverage differs')
    return sorted(expected)


def real_results(base=BASE):
    old = doc(package(base/'protocol113')/'live-report.json')
    text = doc(package(base/'text-jev')/'evaluation-report.json')
    proto = doc(package(base/'protocol114')/'multimodal-report.json')
    visual = doc(base/'visual-jev/report.json')
    if (old['combo'] != 'typesafe-classifier' or old['model'] != 'jev-1.13.0'
            or any(s['outcome'] != 'PASS' for s in old['scenarios'])):
        raise ValueError('Jev1.13 support did not pass')
    if text['kind'] != 'live' or text['status'] != 'COMPLETE' or text['protocolStatus'] != 'COMPLETE':
        raise ValueError('real Chinese effects incomplete')
    if (proto['kind'] != 'jev-formal-protocol' or proto['mode'] != 'live'
            or proto['outcome'] != 'PASS' or not proto['complete'] or not proto['auditPassed']):
        raise ValueError('real formal protocol incomplete')
    if (visual['catalogKind'] != 'formal' or visual['mode'] != 'live'
            or visual['status'] != 'COMPLETE' or visual['realQuality'] != 'COMPLETE'):
        raise ValueError('real formal visual effects incomplete')
    counters = [(old['budget']['callsUsed'],old['budget']['httpAttempts'],old['budget']['environmentRetries']),
        (text['budget']['calls'],text['budget']['attempts'],text['budget']['retries']),
        (proto['budget']['actual']['calls'],proto['budget']['actual']['attempts'],proto['budget']['retries']),
        (visual['budget']['calls'],visual['budget']['attempts'],visual['budget']['retries'])]
    for count, ceiling in zip(counters,(6,24,12,12)):
        if not 0 <= count[0] <= ceiling or not 0 <= count[1] <= ceiling or count[2] != 0:
            raise ValueError('actual paid budget drift')
    calls, attempts = sum(c[0] for c in counters), sum(c[1] for c in counters)
    if calls > 54 or attempts > 54:
        raise ValueError('authorized round exceeded')
    return dict(calls=calls,attempts=attempts,retries=0,componentCounters=counters,
        textBudget=text['budget'],textMetrics=text['metrics'],protocolBudget=proto['budget'],
        visualBudget=visual['budget'],visualMetrics=visual['metrics'],
        catalogHash=proto['catalogHash'],accountRegionPermissions='UNVERIFIED')


def support_inventory(base=BASE):
    support = base/'support'
    if doc(support/'manifest.json') != {'kind':'jev-formal-support','schema':1}:
        raise ValueError('formal support identity changed')
    files = doc(support/'formal-integrity.json')
    if b.tree_paths(support) != set(files)|{'formal-integrity.json'}:
        raise ValueError('formal inventory file set differs')
    b.inventory(support,files)
    if doc(support/'audit.json')['findings']:
        raise ValueError('formal audit findings')
    for name, original in (('protocol',package(base/'protocol114')),('effects',base/'visual-jev')):
        actual = {p:sha(original/p) for p in b.tree_paths(original)}
        copied = {p.removeprefix(name+'/'):v for p,v in files.items() if p.startswith(name+'/')}
        if actual != copied:
            raise ValueError('formal copied package differs from raw producer')
    return dict(base=relative(support),file=relative(support/'formal-integrity.json'),count=len(files))


def matrix_transition():
    before, after = doc(BASE/'matrix-before.json'),doc(BASE/'matrix-after.json')
    if before['schema'] != after['schema'] or len(before['rows']) != len(after['rows']):
        raise ValueError('matrix structure drift')
    changed = []
    for old, new in zip(before['rows'],after['rows']):
        if old['combo'] != new['combo']:
            raise ValueError('matrix identity/order drift')
        if old != new:
            changed.append(old['combo'])
    if set(changed) != {'typesafe-classifier','typesafe-classifier-jev-1.14.0'}:
        raise ValueError('unexpected matrix rows changed')
    if doc(ROOT/'ai/live/support-matrix.json') != after:
        raise ValueError('current matrix differs from accepted transition')
    return dict(changed=changed,unchangedRows=9,beforeSHA256=sha(BASE/'matrix-before.json'),
                afterSHA256=sha(BASE/'matrix-after.json'))


def commands(base=BASE, interrupted=False):
    result = []
    for path in sorted(base.glob('*-command.json')):
        r = doc(path)
        expected = PAID.get(r['name'])
        if r['injectedCredential'] != expected or r['credentialsInherited']:
            raise ValueError('credential isolation receipt drift')
        if expected and expected not in r['environmentKeys']:
            raise ValueError('selected credential variable absent')
        stopped = interrupted and r['name'] == 'releasegate'
        if stopped:
            if r['state'] != 'INTERRUPTED' or r['exitCode'] is not None or r.get('exceptionType') != 'KeyboardInterrupt':
                raise ValueError('historical interruption was misrepresented')
        elif r['state'] != 'COMPLETED' or r['exitCode'] != (1 if r['name']=='releasegate' else 0):
            raise ValueError('required command did not finish with its actual expected verdict')
        result.append(r)
    names = {r['name'] for r in result}
    if not set(PAID)|{'verify-text','verify-visual','support','verify-support','matrix-merge',
                     'no-key-verification','releasegate'} <= names:
        raise ValueError('required execution step missing')
    return result


def no_key_results(base=BASE):
    routes = []
    for directory in sorted((base/'no-key/preflight').iterdir()):
        run = package(directory)
        formal = directory.name == 'typesafe-classifier-jev-1.14.0'
        r = doc(run/('multimodal-report.json' if formal else 'live-report.json'))
        budget = r['budget']
        calls, attempts = ((budget['actual']['calls'],budget['actual']['attempts']) if formal
                           else (budget['callsUsed'],budget['httpAttempts']))
        if calls or attempts:
            raise ValueError('credential-free preflight spent real requests')
        if formal:
            if r['outcome'] != 'NOT_RUN' or r['complete']:
                raise ValueError('formal preflight claims acceptance')
        elif any(s['outcome'] != 'NOT_RUN' for s in r['scenarios']):
            raise ValueError('legacy preflight claims acceptance')
        routes.append(directory.name)
    if len(routes) != 11:
        raise ValueError('eleven isolated no-key preflights required')
    effects = []
    for pattern in ('no-key/text-*/*/evaluation-report.json','no-key/visual-*/report.json'):
        for path in sorted(base.glob(pattern)):
            r = doc(path)
            quality = r.get('protocolStatus',r.get('realQuality'))
            if (r['status'] != 'NOT_RUN' or quality != 'NOT_RUN'
                    or r['budget']['calls'] or r['budget']['attempts'] or r['budget']['retries']):
                raise ValueError('credential-free effects claim real quality/consumption')
            effects.append(relative(path))
    if len(effects) != 4:
        raise ValueError('four independent no-key effect placeholders required')
    return dict(routes=routes,effects=effects,calls=0,attempts=0)


def actual_report():
    r = doc(BASE/'gate/release-report.json')
    if r['pass']:
        raise ValueError('unexpected overall PASS without other provider evidence')
    states = {row['combo']:row['status'] for row in r['live']['rows']}
    if len(states) != 11 or {name for name,status in states.items() if status=='PASS'} != {
            'typesafe-classifier','typesafe-classifier-jev-1.14.0'}:
        raise ValueError('actual accepted live route set differs')
    if any(status != 'NOT_RUN' for name,status in states.items() if not name.startswith('typesafe-classifier')):
        raise ValueError('unapproved provider made an accepting claim')
    for name in ('go-test','go-test-race','go-vet','p0-offline','differential','redaction-audit','snapshots'):
        if not next(g['pass'] for g in r['gates'] if g['id']==name):
            raise ValueError('required current gate did not pass: '+name)
    if ({g['id'] for g in r['gates'] if not g['pass']} != {'live','traceability','required-artifacts'}
            or len(r['audits']) != 17 or any(a['findings'] for a in r['audits'])):
        raise ValueError('actual remaining blockers/audits differ')
    return r


def historic_source():
    source = doc(OLD/'source-at-start.json')
    names = sorted(source['files'])
    env = {k:os.environ[k] for k in ('PATH','HOME') if k in os.environ}
    data = subprocess.run(['git','cat-file','--batch'],cwd=ROOT,env=env,
        input=''.join(source['gitCommit']+':'+p+'\n' for p in names).encode(),
        capture_output=True,check=True,timeout=30).stdout
    for name in names:
        header, data = data.split(b'\n',1)
        fields = header.split()
        if len(fields) != 3 or fields[1] != b'blob':
            raise ValueError('historical committed source missing')
        size = int(fields[2])
        if hashlib.sha256(data[:size]).hexdigest() != source['files'][name] or data[size:size+1] != b'\n':
            raise ValueError('historical source differs from its recorded commit')
        data = data[size+1:]
    if data or sorted(p for p in names if p.endswith('.go')) != sorted(
            relative(p) for p in (ROOT/'ai').rglob('*.go')):
        raise ValueError('historical complete source set differs')
    return dict(commit=source['gitCommit'],files=len(names),verifiedAgainst='recorded Git objects')


def history():
    runs = []
    children = {p.parent for p in sorted((OLD/'no-key/live-harness').glob('*/manifest.json'))[1:]}
    for path in producer_paths(OLD,partial=True):
        runs.append(b.bundle(path,allow_incomplete=path in children))
    normal = b.bundle(package(OLD/'gate/offline'))
    if normal['cases'] != {'PASS':9097,'FAIL':1}:
        raise ValueError('historical failed normal run changed')
    race = next((OLD/'gate/race').iterdir())
    if (race/'manifest.json').exists() or (race/'audit.json').exists() or (OLD/'gate/release-report.json').exists():
        raise ValueError('interrupted race/release was promoted to completed')
    if doc(OLD/'matrix-after.json') != doc(BASE/'matrix-before.json'):
        raise ValueError('independent second-round matrix continuity differs')
    old_files, new_files = (doc(path/'source-at-start.json')['files'] for path in (OLD,BASE))
    if {p for p in old_files if old_files[p] != new_files[p]} != {'ai/e2e/jev_formal_matrix_test.go'}:
        raise ValueError('unexpected source/config differences between paid rounds')
    failed = {c['id'] for c in doc(package(OLD/'gate/offline')/'manifest.json')['cases'] if c['status']=='FAIL'}
    if failed != {'UC13-model-specific-matrix-history'}:
        raise ValueError('historical failure identity differs')
    return dict(source=historic_source(),real=real_results(OLD),commands=commands(OLD,interrupted=True),
        noKey=no_key_results(OLD),support=support_inventory(OLD),runs=runs,normal=normal,
        race=dict(path=relative(race),state='INTERRUPTED',manifest='MISSING',audit='MISSING'),
        currentSourceAcceptance=False)


def store_tree(base):
    if not base.is_dir() or base.is_symlink():
        raise ValueError('missing or linked root')
    index = INDEX.parent/(base.name+'.json')
    files = {name:b.stored_identity(base/name) for name in sorted(b.tree_paths(base))}
    index.parent.mkdir(exist_ok=True)
    if index.exists():
        if doc(index) != files:
            raise ValueError('retained tree changed since its first index')
    else:
        with index.open('x') as stream:
            stream.write(json.dumps(files,sort_keys=True,separators=(',',':'))+'\n')
    return dict(base=relative(base),file=relative(index),count=len(files))


def record():
    if RECEIPT.exists():
        raise ValueError('receipt already sealed; use --verify')
    r = actual_report()
    files = doc(BASE/'source-at-start.json')['files']
    b.inventory(ROOT,files)
    source_paths = sorted(relative(p) for p in (ROOT/'ai').rglob('*.go'))
    if source_paths != sorted(p for p in files if p.endswith('.go')):
        raise ValueError('execution-time complete Go source set differs')
    receipt = dict(schema=1,kind='ticket14-typesafe-real',overall='BLOCKED',files=dict(files),
        sourcePaths=source_paths,
        executionCommit=r['commit'],real=real_results(),commands=commands(),matrix=matrix_transition(),
        gates=r['gates'],live=r['live'],noKey=no_key_results(),history=history(),runs=[],subprocessRuns=[],inventories=[],
        excludedRebuildableRootBinaries=sorted(b.BINS))
    receipt['diagnosis'] = [b.bundle(package(DIAG/name)) for name in ('red','green')]
    rejection = doc(DIAG/'old-source-rejection.json')
    if (rejection['exitCode'] != 1 or rejection['credentialsInherited'] or
            b'formal source snapshot drift' not in raw(DIAG/'old-source-rejection.log')):
        raise ValueError('old-source public refusal evidence differs')
    receipt['oldCurrentRejection'] = rejection
    receipt['diagnosisFiles'] = sorted(b.tree_paths(DIAG))
    for name in receipt['diagnosisFiles']:
        receipt['files'][relative(DIAG/name)] = sha(DIAG/name)
    children = {p.parent for p in sorted((BASE/'no-key/live-harness').glob('*/manifest.json'))[1:]}
    for path in producer_paths():
        run = b.bundle(path,allow_incomplete=path in children)
        if path in children:
            receipt['subprocessRuns'].append(run)
        else:
            if run['auditFindings'] != 0:
                raise ValueError('primary audit findings')
            receipt['runs'].append(run)
    for path in (BASE/'visual-jev',BASE/'no-key/visual-jev',BASE/'no-key/visual-decisions'):
        receipt['inventories'].append(b.visual_bundle(path))
    receipt['inventories'].append(support_inventory())
    for p in [package(BASE/'text-jev')/'files.json',*sorted(BASE.glob('no-key/text-*/*/files.json'))]:
        inv = doc(p)
        b.inventory(p.parent,inv)
        receipt['inventories'].append(dict(base=relative(p.parent),file=relative(p),count=len(inv)))
    for item in doc(BASE/'gate/fixtures.json'):
        p = ROOT/item['path']
        if len(raw(p)) != item['bytes'] or sha(p) != item['sha256']:
            raise ValueError('checked-in fixture changed')
        receipt['files'][relative(p)] = sha(p)
    receipt['storedTrees'] = [store_tree(path) for path in (OLD,BASE)]
    receipt['totalPaid'] = dict(calls=receipt['real']['calls']+receipt['history']['real']['calls'],
        attempts=receipt['real']['attempts']+receipt['history']['real']['attempts'],maxCalls=108,maxAttempts=108)
    for item in receipt['storedTrees']:
        receipt['files'][item['file']] = sha(ROOT/item['file'])
    for name in ('ai/release/classifier-extensions.json',
                 'ai/internal/testkit/release/testdata/multimodal-load.json',
                 'ai/internal/testkit/pioracle/node/package-lock.json',
                 'ai/internal/testkit/pioracle/node/runner.mjs',
                 'ai/internal/testkit/pioracle/node/package.json',
                 'ai/internal/testkit/pioracle/node/provenance.json',
                 'ai/internal/testkit/pioracle/node/model-data.sha256',
                 'ai/e2e/testdata/pidiff/ledger.json',
                 '.scratch/jev-1.14-multimodal-box/run-14-verification.py',
                 'docs/barness-ai/current-catalog-release.md'):
        receipt['files'][name] = sha(ROOT/name)
    for p in [PRIMITIVES,ROOT/'ai/live/support-matrix.json',
              ROOT/'.scratch/jev-1.14-multimodal-box/14-live-authorization-checklist.md',
              ROOT/'.scratch/jev-1.14-multimodal-box/issues/14-current-catalog-release-integration.md',
              ROOT/'.scratch/barness-ai-release-verification/evidence.json',
              ROOT/'.scratch/barness-ai-release-verification/issues/03-jev-matrix-history-current-runs.md',
              ROOT/'.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md']:
        receipt['files'][relative(p)] = sha(p)
    for p in HERE.rglob('*'):
        if p.is_file() and p != RECEIPT:
            receipt['files'][relative(p)] = sha(p)
    RECEIPT.write_text('{\n'+',\n'.join('  '+json.dumps(k)+': '+json.dumps(v,ensure_ascii=False,
        separators=(',',':')) for k,v in receipt.items())+'\n}\n')
    print('sealed',len(receipt['runs']),'primary packages;',sum(x['count'] for x in receipt['storedTrees']),
          'stored entries; real',receipt['totalPaid']['calls'],'Calls across two rounds; release BLOCKED')


def verify():
    receipt = doc(RECEIPT)
    b.inventory(ROOT,receipt['files'])
    if sorted(relative(p) for p in (ROOT/'ai').rglob('*.go')) != receipt['sourcePaths']:
        raise ValueError('complete source set drift')
    if (sorted(b.tree_paths(DIAG)) != receipt['diagnosisFiles'] or
            [b.bundle(ROOT/run['path']) for run in receipt['diagnosis']] != receipt['diagnosis']):
        raise ValueError('isolated failure reproduction drift')
    if (receipt['oldCurrentRejection'] != doc(DIAG/'old-source-rejection.json') or
            receipt['excludedRebuildableRootBinaries'] != sorted(b.BINS)):
        raise ValueError('source-refusal/exclusion summary drift')
    for name, value in (('real',real_results()),('commands',commands()),('matrix',matrix_transition()),
                        ('noKey',no_key_results()),('history',history())):
        # JSON canonicalization equates recorded list counters to Python tuples.
        if json.dumps(receipt[name],sort_keys=True) != json.dumps(value,sort_keys=True):
            raise ValueError('actual paid verdict/consumption drift')
    r = actual_report()
    if (receipt['overall'] != 'BLOCKED' or receipt['gates'] != r['gates'] or receipt['live'] != r['live']
            or receipt['executionCommit'] != r['commit']):
        raise ValueError('actual overall/gate/live drift')
    rows = receipt['runs']+receipt['subprocessRuns']
    if len(rows) != len(producer_paths()) or {x['path'] for x in rows} != {relative(p) for p in producer_paths()}:
        raise ValueError('producer coverage drift')
    children = {run['path'] for run in receipt['subprocessRuns']}
    for run in rows:
        if b.bundle(ROOT/run['path'],allow_incomplete=run['path'] in children) != run:
            raise ValueError('manifest/audit verdict drift')
    for item in receipt['inventories']:
        if item.get('kind') == 'visual':
            if b.visual_bundle(ROOT/item['base']) != item:
                raise ValueError('visual inventory drift')
        else:
            b.inventory(ROOT/item['base'],doc(ROOT/item['file']))
    support_inventory()
    for item in receipt['storedTrees']:
        base = ROOT/item['base']
        if not base.is_dir() or base.is_symlink():
            raise ValueError('missing or linked retained root')
        files = doc(ROOT/item['file'])
        if b.tree_paths(base) != set(files):
            raise ValueError('complete stored file set drift')
        b.inventory(base,files,allow_links=True)
    totals = receipt['totalPaid']
    if (totals['maxCalls'] != 108 or totals['maxAttempts'] != 108 or
            totals['calls'] != receipt['real']['calls']+receipt['history']['real']['calls'] or
            totals['attempts'] != receipt['real']['attempts']+receipt['history']['real']['attempts'] or
            max(totals['calls'],totals['attempts']) > 108):
        raise ValueError('two explicitly authorized rounds exceeded')
    print('verified',receipt['real']['calls'],'real TypeSafe Calls,',receipt['real']['attempts'],
          'Attempts; source, manifests, complete stored bytes, actual gates and matrix transition; '
          'nine routes/two effects remain NOT_RUN')


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    g = p.add_mutually_exclusive_group(required=True)
    g.add_argument('--record',action='store_true')
    g.add_argument('--verify',action='store_true')
    a = p.parse_args()
    record() if a.record else verify()
