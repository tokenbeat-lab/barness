#!/usr/bin/env python3
"""Seal full release evidence, initial failures and the single paid recovery."""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess
import sys

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
RECEIPT = HERE/'evidence.json'
INDEXES = ROOT/'.evidence/ticket14-remaining-integrity'


def module(name, path):
    spec = importlib.util.spec_from_file_location(name,path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


recovery = module('remaining_recovery',HERE/'recover.py')
previous = module('previous_typesafe_sealer',HERE.parent/'14-typesafe-real/seal.py')
r, b = recovery.r, previous.b
INITIAL = recovery.INITIAL
DIAG = ROOT/'.evidence/ticket14-decisions-direct-20261017'


def historic_source(base):
    source = b.doc(base/'source-at-start.json')
    names = sorted(source['files'])
    data = subprocess.run(['git','cat-file','--batch'],cwd=ROOT,env=r.ENV,
        input=''.join(source['gitCommit']+':'+name+'\n' for name in names).encode(),
        capture_output=True,check=True,timeout=30).stdout
    for name in names:
        header,data = data.split(b'\n',1)
        fields = header.split()
        if len(fields)!=3 or fields[1]!=b'blob':
            raise ValueError('recorded committed source missing')
        size = int(fields[2])
        if hashlib.sha256(data[:size]).hexdigest()!=source['files'][name] or data[size:size+1]!=b'\n':
            raise ValueError('historical committed source changed')
        data = data[size+1:]
    if data:
        raise ValueError('unexpected source batch tail')
    return dict(commit=source['gitCommit'],files=len(names),verifiedAgainst='recorded Git objects')


def commands(base, initial=False):
    expected = {name:('BARNESS_AI_LIVE_KEY_'+name.upper().replace('-','_'),key)
        for name,_,key,_ in r.ROUTES}
    expected.update({'text-luna':('BARNESS_AI_CHINESE_EVAL_OPENAI_KEY','OPENAI_KEY'),
        'visual-decisions':('BARNESS_VISUAL_KEY','OPENAI_KEY')})
    if not initial:
        expected.update({'protocol113':('BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER','TYPESAFE_KEY'),
            'protocol114':('BARNESS_MM_KEY','TYPESAFE_KEY'),
            'text-jev':('BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY','TYPESAFE_KEY'),
            'visual-jev':('BARNESS_VISUAL_KEY','TYPESAFE_KEY'),
            'deepseek-chat-recheck':('BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT','DEEPSEEK_KEY')})
    rows = [b.doc(p) for p in sorted(base.glob('*-command.json'))]
    paid = set()
    for row in rows:
        injection,source = expected.get(row['name'],(None,None))
        if (row['credentialsInherited'] or row['injectedCredential']!=injection or row['sourceVariable']!=source
                or row['state']!='COMPLETED' or row['exitCode']!=(1 if
                    (initial and row['name']=='openai-decisions') or
                    (not initial and row['name'] in ('live-harness','deepseek-chat')) else 0)):
            raise ValueError('command verdict/isolation changed')
        keys = {name for name in row['environmentKeys'] if 'KEY' in name or 'TOKEN' in name or 'SECRET' in name}
        if keys != ({injection} if injection else set()):
            raise ValueError('unexpected credential environment')
        if injection:
            paid.add(row['name'])
    if paid!=set(expected):
        raise ValueError('paid execution step missing')
    required = {'build-live','build-text','build-visual'} | set(expected)
    required |= ({'verify-text','verify-visual','reuse-typesafe-support','reuse-typesafe-text','reuse-typesafe-visual'} if initial else
        {'live-harness','live-harness-recheck','verify-text-luna','verify-text-jev','verify-visual-decisions','verify-visual-jev',
         'support','verify-support','matrix-merge','releasegate'})
    if {row['name'] for row in rows}!=required:
        raise ValueError('command stage set differs')
    return rows


def producers(base, initial=False):
    expected = {r.bundle(base/'live'/name) for name,_,_,_ in r.ROUTES}
    expected.add(r.bundle(base/'text-luna'))
    if not initial:
        expected.update(r.bundle(base/name) for name in ('protocol113','protocol114','text-jev','gate/offline','gate/race'))
        expected.add(r.bundle(base/'deepseek-chat-recheck'))
    actual = {p.parent for pattern in ('live/*/*/manifest.json','protocol11?/*/manifest.json',
        'text-*/*/manifest.json','gate/offline/*/manifest.json','gate/race/*/manifest.json',
        'deepseek-chat-recheck/*/manifest.json') for p in base.glob(pattern)}
    visuals = {'visual-decisions'} if initial else {'visual-decisions','visual-jev'}
    if actual!=expected or {p.parent.name for p in base.glob('visual-*/manifest.json')}!=visuals:
        raise ValueError('mandatory producer/effect stage coverage differs')
    return sorted(expected)


def consumption(base, full=False):
    caps = recovery.CAPS if full else r.CAPS
    summary = b.doc(base/('consumption-all.json' if full else 'consumption.json'))
    rows = {row['name']:row for row in summary['routes']}
    if len(rows)!=len(summary['routes']) or set(rows)!=set(caps):
        raise ValueError('budget route set differs')
    calls,attempts = 0,0
    for name,cap in caps.items():
        if name.startswith('visual-'):
            path = base/name/'report.json'
        elif name.startswith('text-'):
            path = r.bundle(base/name)/'evaluation-report.json'
        elif name=='protocol114':
            path = r.bundle(base/name)/'multimodal-report.json'
        else:
            path = r.bundle(base/(name if name=='protocol113' else 'live/'+name))/'live-report.json'
        budget = b.doc(path)['budget']
        if name=='protocol114':
            budget = budget['actual']
        count,tries = ((budget['calls'],budget['attempts']) if name.startswith(('text-','visual-')) or name=='protocol114'
            else (budget['callsUsed'],budget['httpAttempts']))
        row = rows[name]
        if (not 0<=count<=cap or not 0<=tries<=cap or row.get('actual')=='UNKNOWN' or row['maxCalls']!=cap
                or row['calls']!=count or row['attempts']!=tries
                or row['conservativeCalls']!=count or row['conservativeAttempts']!=tries):
            raise ValueError('raw actual component consumption differs')
        calls,attempts = calls+count,attempts+tries
    total = dict(calls=calls,attempts=attempts,conservativeCalls=calls,conservativeAttempts=attempts,
        maxCalls=sum(caps.values()),maxAttempts=sum(caps.values()))
    if summary['total']!=total:
        raise ValueError('actual budget summary sum differs')
    return summary


def current_results(base):
    report = b.doc(base/'gate/release-report.json')
    if (not report['pass'] or any(not gate['pass'] for gate in report['gates'])
            or len(report['live']['rows'])!=11 or any(row['status']!='PASS' for row in report['live']['rows'])
            or len(report['audits'])!=17 or any(a['findings'] or a.get('error') for a in report['audits'])):
        raise ValueError('complete actual release PASS missing')
    summary = b.doc(base/'live-summary.json')
    if ({name for name,code in summary['commands'].items() if code!=0}!={'deepseek-chat'}
            or not b.doc(base/'accepted-deepseek-chat.json')['accepting']):
        raise ValueError('actual producer/confirmation verdict differs')
    producers(base)
    producers(INITIAL,True)
    for stage in ('offline','race'):
        path = r.bundle(base/'gate'/stage)
        if b.bundle(path)['cases']!={'PASS':9098}:
            raise ValueError('actual complete normal/race manifest failed')
        cases = b.doc(path/'manifest.json')['cases']
        if (sum(c['id'].startswith('PIDIFF-') for c in cases)!=861
                or sum(c['id'].startswith('UC14-') for c in cases)!=33):
            raise ValueError('Pi or public image acceptance coverage differs')
    budgets = consumption(base,True)
    initial = consumption(INITIAL)['total']
    original_task = b.doc(base/'task-consumption.json')
    target = b.doc(r.bundle(base/'deepseek-chat-recheck')/'live-report.json')
    if (target['combo']!='deepseek-chat' or any(s['outcome']!='PASS' for s in target['scenarios'])
            or not 0<=target['budget']['callsUsed']<=32 or not 0<=target['budget']['httpAttempts']<=32):
        raise ValueError('targeted full route did not pass within its envelope')
    target_calls,target_attempts = target['budget']['callsUsed'],target['budget']['httpAttempts']
    task = b.doc(base/'task-consumption-final.json')
    for key in ('calls','attempts','conservativeCalls','conservativeAttempts'):
        extra = target_calls if key in ('calls','conservativeCalls') else target_attempts
        if (original_task[key]!=initial[key]+budgets['total'][key]
                or task[key]!=original_task[key]+extra or not 0<=task[key]<=550):
            raise ValueError('task budget drift')
    deepseek = [row for row in consumption(INITIAL)['routes']+budgets['routes'] if row['name'].startswith('deepseek-')]
    if (sum(row['calls'] for row in deepseek)+target_calls>128
            or sum(row['attempts'] for row in deepseek)+target_attempts>128):
        raise ValueError('DeepSeek task provider cap exceeded')
    if max(budgets['total']['conservativeCalls'],budgets['total']['conservativeAttempts'])>302:
        raise ValueError('recovery exceeded')
    formal = b.doc(r.bundle(base/'live/openai-decisions')/'decisions-execution-report.json')
    if (formal['hostPermissionVerification'] is not None or
            any(formal[k]!='unknown' for k in ('accountVerificationStatus','regionVerificationStatus','permissionVerificationStatus'))):
        raise ValueError('unverified host permission claims changed')
    if b.doc(ROOT/'ai/live/support-matrix.json')!=b.doc(base/'matrix-after.json'):
        raise ValueError('current matrix differs from actual accepted merge')
    before,after = b.doc(base/'matrix-before.json'),b.doc(base/'matrix-after.json')
    if (len(before['rows'])!=11 or [x['combo'] for x in before['rows']]!=[x['combo'] for x in after['rows']]
            or any(old==new for old,new in zip(before['rows'],after['rows']))):
        raise ValueError('expected eleven independent matrix row updates missing')
    return dict(report=report,consumption=budgets,task=task,
        targetedConfirmation=dict(calls=target_calls,attempts=target_attempts,maxCalls=32),
        typesafe=previous.real_results(base),
        matrixChangedRows=[row['combo'] for row in after['rows']])


def store_tree(base):
    INDEXES.mkdir(exist_ok=True)
    files = {name:b.stored_identity(base/name) for name in sorted(b.tree_paths(base))}
    path = INDEXES/(base.name+'.json')
    if path.exists():
        if b.doc(path)!=files:
            raise ValueError('previously indexed evidence changed')
    else:
        with path.open('x') as stream:
            stream.write(json.dumps(files,sort_keys=True,separators=(',',':'))+'\n')
    return dict(base=b.relative(base),file=b.relative(path),count=len(files))


def record(base):
    if RECEIPT.exists():
        raise ValueError('already sealed; use --verify')
    recovery.Recovery(b.relative(base)).guard()
    result = current_results(base)
    source = b.doc(base/'source-at-start.json')
    b.inventory(ROOT,source['files'])
    paths = sorted(b.relative(p) for p in (ROOT/'ai').rglob('*.go'))
    if paths!=sorted(p for p in source['files'] if p.endswith('.go')):
        raise ValueError('complete current Go set changed')
    receipt = dict(schema=1,kind='ticket14-all-provider-real',overall='PASS',base=b.relative(base),
        files=dict(source['files']),sourcePaths=paths,executionCommit=result['report']['commit'],
        sourceCommit=source['gitCommit'],current=result,commands=commands(base),initialCommands=commands(INITIAL,True),
        historicalSources=[historic_source(INITIAL)],runs=[],subprocessRuns=[],inventories=[],storedTrees=[],
        excludedRebuildableRootBinaries=sorted(b.BINS))
    initial = b.doc(r.bundle(INITIAL/'live/openai-decisions')/'live-report.json')
    if (initial['budget']['callsUsed'] or initial['budget']['httpAttempts']
            or any(s['outcome']!='FAIL' or s['errorCategory']!='config' for s in initial['scenarios'])):
        raise ValueError('initial zero-spend config failure changed')
    receipt['initialVerdict'] = dict(accepting=False,realConsumption=b.doc(INITIAL/'consumption.json'),
        decisionsFailure='optional direct proxy rejected before any HTTP')
    for root in (base,INITIAL):
        for producer in producers(root,root==INITIAL):
            receipt['runs'].append(b.bundle(producer))
        for path in sorted(root.glob('visual-*/manifest.json')):
            receipt['inventories'].append(b.visual_bundle(path.parent))
        for path in sorted(root.glob('text-*/*/files.json')):
            b.inventory(path.parent,b.doc(path))
            receipt['inventories'].append(dict(base=b.relative(path.parent),file=b.relative(path),count=len(b.doc(path))))
    for name in ('live-harness','live-harness-recheck'):
        harness = sorted(base.glob(name+'/*/manifest.json'))
        if len(harness)!=3:
            raise ValueError('live harness parent and deliberate crash children required')
        run = b.bundle(harness[0].parent)
        expected = {'PASS':207,'NOT_RUN':62} if name.endswith('recheck') else {'PASS':206,'NOT_RUN':62,'FAIL':1}
        if run['cases']!=expected:
            raise ValueError('original/rechecked harness verdict drift')
        receipt['runs'].append(run)
        receipt['subprocessRuns'].extend(b.bundle(p.parent,allow_incomplete=True) for p in harness[1:])
    for phase,exitcode in (('red',1),('green',0)):
        if b.doc(DIAG/(phase+'-command.json'))['exitCode']!=exitcode:
            raise ValueError('RED/GREEN process verdict drift')
        receipt['runs'].append(b.bundle(r.bundle(DIAG/phase)))
        b.inventory(DIAG/(phase+'-source'),b.doc(DIAG/(phase+'-source.json')))
    receipt['runs'].append(b.bundle(r.bundle(DIAG/'fifo-single')))
    receipt['inventories'].append(previous.support_inventory(base))
    if any(run['auditFindings'] for run in receipt['runs']):
        raise ValueError('primary redaction audit findings')
    # Old receipts describe earlier current trees. Preserve their exact bytes
    # and all indexed runs; never re-evaluate old matrix/docs as current truth.
    old_receipt = HERE.parent/'14-typesafe-real/evidence.json'
    old = b.doc(old_receipt)
    receipt['storedTrees'] = old['storedTrees']+[store_tree(root) for root in (INITIAL,DIAG,base)]
    for item in old['storedTrees']:
        if b.sha(ROOT/item['file'])!=old['files'][item['file']]:
            raise ValueError('historical stored index changed')
    for item in receipt['storedTrees']:
        receipt['files'][item['file']] = b.sha(ROOT/item['file'])
    for fixture in b.doc(base/'gate/fixtures.json'):
        path = ROOT/fixture['path']
        if b.sha(path)!=fixture['sha256'] or len(b.raw(path))!=fixture['bytes']:
            raise ValueError('checked-in fixture changed')
        receipt['files'][b.relative(path)] = b.sha(path)
    for name in ('ai/live/support-matrix.json','ai/release/classifier-extensions.json',
        'ai/internal/testkit/release/testdata/multimodal-load.json','ai/e2e/testdata/pidiff/ledger.json',
        'ai/internal/testkit/pioracle/node/package-lock.json','ai/internal/testkit/pioracle/node/runner.mjs',
        'ai/internal/testkit/pioracle/node/package.json','ai/internal/testkit/pioracle/node/provenance.json',
        'ai/internal/testkit/pioracle/node/model-data.sha256','ai/live/testdata/jev-formal/sources.json',
        '.scratch/jev-1.14-multimodal-box/verify-11-evidence.py',
        '.scratch/jev-1.14-multimodal-box/issues/14-current-catalog-release-integration.md',
        '.scratch/jev-1.14-multimodal-box/14-live-authorization-checklist.md',
        '.scratch/barness-ai-release-verification/issues/04-decisions-direct-transport.md',
        '.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md',
        'docs/barness-ai/current-catalog-release.md'):
        receipt['files'][name] = b.sha(ROOT/name)
    for path in (old_receipt,HERE.parent/'14-typesafe-real/seal.py',HERE.parent/'seal-14-evidence.py',
                 HERE/'run.py',HERE/'recover.py',HERE/'recheck.py',HERE/'seal.py',HERE/'authorization.md',HERE/'verification.md'):
        receipt['files'][b.relative(path)] = b.sha(path)
    with RECEIPT.open('x') as stream:
        stream.write('{\n'+',\n'.join('  '+json.dumps(k)+': '+json.dumps(v,ensure_ascii=False,
            separators=(',',':')) for k,v in receipt.items())+'\n}\n')
    print('sealed actual full release PASS;',result['task']['calls'],'new Calls;',
          sum(item['count'] for item in receipt['storedTrees']),'retained entries')


def verify():
    receipt = b.doc(RECEIPT)
    b.inventory(ROOT,receipt['files'])
    base = ROOT/receipt['base']
    if sorted(b.relative(p) for p in (ROOT/'ai').rglob('*.go'))!=receipt['sourcePaths']:
        raise ValueError('current complete source set drift')
    if (json.loads(json.dumps(current_results(base)))!=receipt['current']
            or commands(base)!=receipt['commands'] or commands(INITIAL,True)!=receipt['initialCommands']):
        raise ValueError('actual result/command drift')
    if historic_source(INITIAL)!=receipt['historicalSources'][0]:
        raise ValueError('initial historical source drift')
    for run in receipt['runs']+receipt['subprocessRuns']:
        if b.bundle(ROOT/run['path'],allow_incomplete=run in receipt['subprocessRuns'])!=run:
            raise ValueError('producer inventory/verdict drift')
    for item in receipt['inventories']:
        if item.get('kind')=='visual':
            if b.visual_bundle(ROOT/item['base'])!=item:
                raise ValueError('visual inventory drift')
        else:
            b.inventory(ROOT/item['base'],b.doc(ROOT/item['file']))
    for item in receipt['storedTrees']:
        root = ROOT/item['base']
        files = b.doc(ROOT/item['file'])
        if b.tree_paths(root)!=set(files) or len(files)!=item['count']:
            raise ValueError('stored evidence set drift')
        b.inventory(root,files,allow_links=True)
    print('current source, complete retained trees, failures, budgets and actual full PASS verified')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record',help='completed recovery root relative to repository')
    parser.add_argument('--verify',action='store_true')
    args = parser.parse_args()
    if args.verify:
        verify()
    elif args.record and not Path(args.record).is_absolute() and '..' not in Path(args.record).parts:
        record(ROOT/args.record)
    else:
        parser.error('use --record ROOT or --verify')


if __name__ == '__main__':
    main()
