#!/usr/bin/env python3
"""Read-only actual legacy integrity/accounting and original fixed quality metrics."""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import stat
import sys
import types

BASE=Path(__file__).resolve().parent
HELPER_SHA='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'


def helper():
    path=BASE.parent/'implementation/verify.py'
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:
        info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER_SHA:
        raise ValueError('pinned bounded verifier differs')
    module=types.ModuleType('bounded_capsule');module.__file__=str(path)
    exec(compile(raw,str(path),'exec'),module.__dict__);module.BASE=BASE
    return module


EXPECTED='8fbd17006677774368d4d77d924a8672fd205c51'
COUNTS={'openai-responses':14,'openai-chat':14,'anthropic-messages':20,'google-gemini':8,'deepseek-responses':10,
        'deepseek-chat':13,'openai-images':3,'google-interactions-image':2,'openai-decisions':9,'typesafe-classifier':3,
        'protocol114':12,'text-jev':24,'text-luna':24,'visual-jev':12,'visual-decisions':12}
SCENES={'openai-responses':8,'openai-chat':10,'anthropic-messages':11,'google-gemini':6,'deepseek-responses':8,
        'deepseek-chat':11,'openai-images':3,'google-interactions-image':2,'openai-decisions':9,'typesafe-classifier':3}


def analyse(b,index,payload):
    inventory=b.inventory(index['capsule']);rows=[];scenes=[];source=b.doc(payload['frozen/source-at-start.json'])
    review=b.doc(payload['frozen/prerelease-source-review.json']);approval=b.doc(payload['frozen/execution-authorization.json'])
    plan=b.doc(payload['frozen/plan.json']);checks=b.doc(payload['frozen/checks.json'])
    consumption=b.doc(payload['frozen/routed-consumption.json'])
    if (source['sourceCommit']!=EXPECTED or index['actualSource']!=EXPECTED or review['decision']!='APPROVED' or
        approval['decision']!='APPROVED' or approval['authority']!='explicit-user-request' or checks['passed'] is not True or
        any(v!=0 for v in checks['commands'].values()) or plan['phase']!='legacy' or plan['round']!='routed-validation-1' or
        plan['automaticProcessReruns'] is not False or plan['proxy']!='http://127.0.0.1:7890'):
        raise ValueError('original reviewed source, authority or fixed round differs')
    for field,leaf in (('sourceManifestSHA256','source-at-start.json'),('binariesManifestSHA256','binaries.json'),('planSHA256','plan.json')):
        if review[field]!=b.sha(payload['frozen/'+leaf]) or approval[field]!=review[field]:raise ValueError('original source authority pin differs')
    if approval['checksSHA256']!=b.sha(payload['frozen/checks.json']) or approval['sourceReviewSHA256']!=b.sha(payload['frozen/prerelease-source-review.json']):
        raise ValueError('original no-key checks or review pin differs')
    def check_map(prefix,mapping):
        for name,digest in mapping.items():
            b.relative(name)
            if inventory.get(prefix+name,{}).get('sha256')!=digest.removeprefix('sha256:'):raise ValueError('original member seal differs')
    for name,raw in payload.items():
        if name.startswith('actual/') and name.endswith('/audit.json') and b.doc(raw).get('findings')!=[]:raise ValueError('original audit findings differ')
    for name,entry in index['routes'].items():
        prefix=entry['root']+'/';report=b.doc(payload[prefix+entry['report']]);receipt=b.doc(payload['canonical/'+name+'.json'])
        command=b.doc(payload['frozen/'+name+'-command.json']);audit=b.doc(payload[prefix+'audit.json'])
        network='china-direct' if name.startswith('deepseek-') else 'international-loopback-proxy'
        if (receipt['state']!='COMPLETED' or receipt['sourceCommit']!=EXPECTED or receipt['route']!='routed-validation-1-'+name or
            receipt['baseName']!=name or receipt['phase']!='legacy' or receipt['validationRound']!='routed-validation-1' or
            receipt['credentialsRead'] is not True or receipt['networkRoute']!=network or receipt['automaticReruns'] is not False or
            receipt['launcherAudit']['passed'] is not True or command['state']!='COMPLETED' or command['name']!=name or
            command['credentialsRead'] is not True or command['exitCode']!=0 or receipt['exitCode']!=0 or audit['findings'] or
            any(n.upper() in ('HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY') for n in command['environmentKeys'])):
            raise ValueError('own actual identity, routing, collection outcome or audit differs')
        pins={'sourceManifestSha256':'source-at-start.json','binariesManifestSha256':'binaries.json','planSha256':'plan.json',
              'authorizationSha256':'execution-authorization.json','sourceReviewSha256':'prerelease-source-review.json',
              'commandReceiptSha256':name+'-command.json'}
        if any(receipt[field]!=b.sha(payload['frozen/'+leaf]) for field,leaf in pins.items()):raise ValueError('canonical original pin differs')
        if receipt['reportedFileSha256']!=b.sha(payload[prefix+entry['report']]):raise ValueError('own original report pin differs')
        manifest=b.doc(payload[prefix+'manifest.json'])
        if 'artifacts' in manifest:check_map(prefix,manifest['artifacts'])
        else:
            if manifest['git_commit']!=EXPECTED:raise ValueError('actual manifest source differs')
            for case in manifest['cases']:
                if case['status']!='PASS' and not (name=='openai-chat' and case['id']=='LIVE-P04-reasoning-history' and case['status']=='UNSUPPORTED'):raise ValueError('actual assertion case failed')
                check_map(prefix+case['dir']+'/',case.get('artifacts',{}))
        records=[];wire=[];kind=entry['kind'];quality=None;budget=report['budget']
        if kind=='live':
            if (report['combo']!=name or report['gitCommit']!=EXPECTED or len(report['scenarios'])!=SCENES[name] or
                {s['id'] for s in report['scenarios']}!=set(report['expected'])):raise ValueError('fixed generic scene inventory differs')
            statuses=dict(collections.Counter(s['outcome'] for s in report['scenarios']))
            wanted={'PASS':9,'UNSUPPORTED':1} if name=='openai-chat' else {'PASS':SCENES[name]}
            if statuses!=wanted:raise ValueError('generic scene outcomes differ')
            for s in report['scenarios']:
                if s['outcome']=='UNSUPPORTED' and (s['id']!='reasoning-history' or s['calls'] or s['attempts']):raise ValueError('protocol unsupported row differs')
                scenes.append(dict(collection=name,id=s['id'],outcome=s['outcome'],calls=s['calls'],attempts=s['attempts'],model=s['model'],note=s.get('note')))
            calls=sum(s['calls'] for s in report['scenarios']);attempts=sum(s['attempts'] for s in report['scenarios'])
            if calls!=budget['callsUsed'] or attempts!=budget['httpAttempts']:raise ValueError('generic scene budget differs')
        elif kind=='protocol':
            if report['outcome']!='PASS' or report['complete'] is not True or report['auditPassed'] is not True or report['gitCommit']!=EXPECTED or len(report['expected'])!=12 or report['scenarios']!=report['expected']:
                raise ValueError('fixed protocol inventory differs')
            seal=b.doc(payload[prefix+'multimodal-integrity.json']);check_map(prefix,seal['files'])
            calls=budget['callsUsed'];attempts=budget['httpAttempts'];statuses={'PASS':12}
            for s in report['scenarios']:scenes.append(dict(collection=name,id=s,outcome='PASS',calls=1,attempts=1,model='jev-1.14.0',note=None))
            quality={'outcome':report['outcome'],'complete':report['complete'],'unknown':report['unknown'],'actual':budget['actual']}
        else:
            if report['status']!='COMPLETE' or len(report['samples'])!=COUNTS[name] or len({s['id'] for s in report['samples']})!=COUNTS[name]:raise ValueError('fixed effect sample inventory differs')
            calls=budget['calls'];attempts=budget['attempts'];statuses=dict(collections.Counter(s['status'] for s in report['samples']))
            if statuses!={'answered':COUNTS[name]}:raise ValueError('effect sample status differs')
            quality=report['metrics']
            if kind=='text':
                check_map(prefix,b.doc(payload[prefix+'files.json']))
                records=[a for s in report['samples'] for a in s['metadata']['attempts']]
                wire=json.loads(payload[prefix+'captures.json'])
                if len(wire)!=24 or any(s.get('source')!='live' for s in wire):raise ValueError('actual text capture differs')
                truth={s['id']:s['label'] for s in report['dataset']['samples']}
                correct=sum(s['answer'].get('choice')==truth[s['id']] for s in report['samples'])
                if correct!=24 or quality['correct']!=correct or quality['total']!=24 or quality['failed'] or quality['unexecuted']:raise ValueError('strict synthetic text score differs')
            elif report['mode']!='live' or report['realQuality']!='COMPLETE' or b.doc(payload[prefix+'run-config.json']).get('proxy')!='http://127.0.0.1:7890':
                raise ValueError('actual visual mode or proxy differs')
            for s in report['samples']:scenes.append(dict(collection=name,id=s['id'],outcome=s['status'],calls=1,attempts=1,model=report['route']['model'] if kind=='text' else report['model'],note=None))
        if kind!='text':
            for path,raw in payload.items():
                if not path.startswith(prefix) or '/source/' in path:continue
                if path.endswith('result.json'):
                    value=b.doc(raw);records+=value.get('result',{}).get('metadata',{}).get('attempts',[])
                if path.endswith(('exchanges.json','wire.json')):
                    value=json.loads(raw);wire+=value if isinstance(value,list) else [value]
        if (calls!=COUNTS[name] or attempts!=calls or len(records)!=calls or len(wire)!=attempts or
            receipt['budget']['calls']!=calls or receipt['budget']['attempts']!=attempts or receipt['budget']['state']!='REPORTED' or receipt['budget']['status']!='PASS'):
            raise ValueError('independent call, final Attempt, wire or canonical budget differs')
        usage=collections.Counter();known_input=known_output=0;http=collections.Counter()
        for a in records:
            usage[a.get('usageReporting','unreported')]+=1;http[str(a.get('httpStatus','unreported'))]+=1
            u=a.get('usage',{});values=[u.get(k,0) for k in ('input','cacheRead','cacheWrite','output')]
            if any(type(v) is not int or v<0 for v in values):raise ValueError('invalid usage counter')
            known_input+=sum(values[:3]);known_output+=values[3]
        rows.append(dict(collection=name,collectionOutcome='PASS',sceneStatuses=statuses,sceneCount=sum(statuses.values()),calls=calls,attempts=attempts,
            actualAttemptRecords=len(records),actualWireRecords=len(wire),networkRoute=network,model=report.get('model',report.get('route')),
            usageReporting=dict(usage),knownInputSide=known_input,knownOutput=known_output,httpStatus=dict(http),auditFindings=0,quality=quality))
    support='actual/support/';seal=b.doc(payload[support+'formal-integrity.json']);check_map(support,seal)
    expected_seal={n.removeprefix(support):v['sha256'] for n,v in inventory.items() if n.startswith(support) and n!=support+'formal-integrity.json'}
    if seal!=expected_seal:raise ValueError('derived support sealed inventory differs')
    own=b.doc(payload[support+'live-report.json'])
    if own['combo']!='typesafe-classifier-jev-1.14.0' or own['gitCommit']!=EXPECTED or len(own['scenarios'])!=12 or any(s['outcome']!='PASS' for s in own['scenarios']):raise ValueError('derived Jev support differs')
    for child,origin in (('protocol',index['routes']['protocol114']['root']),('effects',index['routes']['visual-jev']['root'])):
        for name,pin in inventory.items():
            prefix=origin+'/'
            if name.startswith(prefix) and inventory.get(support+child+'/'+name.removeprefix(prefix))!=pin:raise ValueError('derived own immutable protocol/effects copy differs')
    for name in ('derive-jev-support','validate-17-temporary-matrix','merge-17-default-matrix'):
        record=b.doc(payload['derivation/'+name+'-command.json'])
        if record['state']!='COMPLETED' or record['exitCode']!=0 or record['credentialsRead'] is not False:raise ValueError('original public derive/import receipt differs')
    matrix=b.doc(payload['derivation/matrix-after.json'])
    identities=set(SCENES)|{'typesafe-classifier-jev-1.14.0'}|{provider+'-'+route for provider in ('zhipu','minimax') for route in ('chat','responses','messages')}
    if len(matrix['rows'])!=17 or {r['combo'] for r in matrix['rows']}!=identities:raise ValueError('fixed17 own matrix identities differ')
    matrix_status={r['combo']:('FAIL' if any(c['outcome']=='FAIL' for c in r['capabilities']) else 'PASS') for r in matrix['rows']}
    if {k for k,v in matrix_status.items() if v=='FAIL'}!={'zhipu-responses'}:raise ValueError('matrix cannot hide actual GLM Responses failure')
    references=b.doc(b.read(BASE,'external-references.json',1<<20));companion=references['companionGLM']
    glm=b.read(BASE.parent,'actual-glm/summary.json',1<<20)
    if b.sha(glm)!=companion['summarySHA256'] or b.doc(glm)['calls']!=65 or b.doc(glm)['attempts']!=65 or b.doc(glm)['collectionOutcome']!='FAIL':raise ValueError('companion immutable GLM summary differs')
    m31=b.doc(payload['frozen/m31-not-run-summary.json'])
    if m31['credentialsRead'] is not False or m31['paidIntentsCreated'] is not False or len(m31['routes'])!=3 or any(r['status']!='NOT_RUN' or r['calls'] or r['attempts'] for r in m31['routes']):raise ValueError('M Plan NOT_RUN differs')
    for route in ('chat','responses','messages'):
        prefix=next(n.rsplit('/',1)[0]+'/' for n in inventory if n.startswith('actual/m31-not-run-'+route+'/') and n.endswith('/minimax-report.json'))
        report=b.doc(payload[prefix+'minimax-report.json']);command=b.doc(payload['frozen/m31-not-run-'+route+'-command.json'])
        if report['outcome']!='NOT_RUN' or report['complete'] is not False or any(report['cumulativeSlotBudget'].values()) or command['credentialsRead'] is not False or command['exitCode']!=0:raise ValueError('original unpaid M Plan status differs')
    total=sum(r['calls'] for r in rows);usage=collections.Counter()
    for row in rows:usage.update(row['usageReporting'])
    if total!=180 or consumption['current']['calls']!=245 or consumption['current']['attempts']!=245 or consumption['total']['calls']!=545 or consumption['total']['attempts']!=545:
        raise ValueError('closed routed round budget differs')
    return dict(schema=1,actualSource=EXPECTED,collections=15,collectionOutcome='PASS',calls=180,attempts=180,rows=rows,
        usageReporting=dict(usage),unknownOrPartialUsageAttempts=sum(v for k,v in usage.items() if k!='complete'),knownInputSide=sum(r['knownInputSide'] for r in rows),knownOutput=sum(r['knownOutput'] for r in rows),
        routedGLMCalls=65,routedGLMAttempts=65,routedRoundCalls=245,routedRoundAttempts=245,priorCurrentReleaseCalls=300,priorCurrentReleaseAttempts=300,currentReleaseCalls=545,currentReleaseAttempts=545,
        accountingScope=consumption['accountingScope'],automaticReruns=False,derivedJevSupport=True,mplanStatus='NOT_RUN',
        usageBasis='Independent final Attempt sums: input+cacheRead+cacheWrite once; output once. Reasoning is already output. Partial/unreported remains unknown; known lower bound is not an invoice.',
        matrixLatestRows=matrix_status,qualityScope='Fixed synthetic text24 and visual12 per route; original metrics retained, not a production accuracy claim.'),scenes


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
    for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        digest,name=line.split('  ',1)
        if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
    index=b.doc(b.read(BASE,'index.json',1<<20))
    if index.get('boundedHelperSHA256')!=HELPER_SHA or index['capsule']['id']!='actualLegacy8fbd' or set(index['routes'])!=set(COUNTS):raise ValueError('fixed own legacy collection index differs')
    archive,payload=b.verify_archive(a.common_root,index['capsule']);summary,scenes=analyse(b,index,payload)
    if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent summary differs')
    expected=[json.loads(line) for line in b.read(BASE,'scenarios.jsonl',1<<20).decode().splitlines()]
    if scenes!=expected:raise ValueError('independent scene inventory differs')
    print(json.dumps(dict(schema=1,verified=True,retentionCredentialsRead=False,retentionPaidCalls=0,archive=archive,collections=15,calls=180,attempts=180,
        scope='immutable legacy integrity/accounting/quality only; GLM Responses failure and full release gate remain separately reported'),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
