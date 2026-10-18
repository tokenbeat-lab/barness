#!/usr/bin/env python3
"""Read-only actual GLM integrity/accounting; never grants PASS or replay."""
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


def analyse(b,index,payload):
    inventory=b.inventory(index['capsule']);rows=[];scenes=[];source=b.doc(payload['frozen/source-at-start.json'])
    parent=b.doc(payload['parent/plan.json']);consumption=b.doc(payload['frozen/routed-consumption.json'])
    approval=b.doc(payload['frozen/execution-authorization.json']);review=b.doc(payload['frozen/prerelease-source-review.json'])
    checks=b.doc(payload['frozen/checks.json'])
    expected_commit='51c34338cdcabd77625d6265409ab910ae08c8c9'
    if (source['sourceCommit']!=expected_commit or index['actualSource']!=expected_commit or parent['sourceCommit']!=expected_commit or
        approval['authority']!='explicit-user-request' or approval['decision']!='APPROVED' or review['decision']!='APPROVED' or
        checks['passed'] is not True or any(code!=0 for code in checks['commands'].values())):
        raise ValueError('original source or reviewed no-key authority differs')
    pins={'sourceManifestSha256':'source-at-start.json','binariesManifestSha256':'binaries.json','planSha256':'plan.json',
        'authorizationSha256':'execution-authorization.json','sourceReviewSha256':'prerelease-source-review.json'}
    source_pins={'sourceCommit':expected_commit,'sourceManifestSHA256':b.sha(payload['frozen/source-at-start.json']),
        'binariesManifestSHA256':b.sha(payload['frozen/binaries.json']),'planSHA256':b.sha(payload['frozen/plan.json'])}
    if (any(review.get(k)!=v or approval.get(k)!=v for k,v in source_pins.items()) or
        approval['checksSHA256']!=b.sha(payload['frozen/checks.json']) or approval['sourceReviewSHA256']!=b.sha(payload['frozen/prerelease-source-review.json']) or
        review['credentialScopes'].get('GLM_KEY')!='routed-validation-1/china-direct' or parent['automaticReruns'] is not False):
        raise ValueError('independent original source/checks/request review differs')
    for route,root in index['routes'].items():
        prefix=root+'/';report=b.doc(payload[prefix+'formal-glm-report.json']);receipt=b.doc(payload['canonical/'+route+'.json'])
        command=b.doc(payload['frozen/glm-'+route+'-command.json']);audit=b.doc(payload[prefix+'audit.json'])
        expected_outcome='FAIL' if route=='responses' else 'PASS';expected_calls=21 if route=='responses' else 22
        if (report['outcome']!=expected_outcome or report['complete']!=(route!='responses') or report['auditPassed'] is not True or
            report['gitCommit']!=expected_commit or report['route']!=route or receipt['sourceCommit']!=expected_commit or
            receipt['route']!='routed-validation-1-glm-'+route or receipt['phase']!='glm' or receipt['validationRound']!='routed-validation-1' or
            receipt['state']!='COMPLETED' or receipt['credentialsRead'] is not True or receipt['networkRoute']!='china-direct' or
            receipt['automaticReruns'] is not False or receipt['launcherAudit']['passed'] is not True or audit['findings']!=[] or
            command['state']!='COMPLETED' or command['credentialsRead'] is not True or command['exitCode']!=(1 if route=='responses' else 0) or
            command['name']!='glm-'+route or command['sourceVariable']!='GLM_KEY' or command['injectedCredential']!='BARNESS_GLM_KEY' or
            command['launcherAudit']['passed'] is not True or any('PROXY' in name.upper() for name in command['environmentKeys'])):
            raise ValueError('own actual identity, failed status, direct routing or audit differs')
        for field,leaf in pins.items():
            if receipt[field]!=b.sha(payload['frozen/'+leaf]):raise ValueError('canonical original authority pin differs')
        if (receipt['reportedFileSha256']!=b.sha(payload[prefix+'formal-glm-report.json']) or
            receipt['commandReceiptSha256']!=b.sha(payload['frozen/glm-'+route+'-command.json'])):
            raise ValueError('original report or command pin differs')
        seal=b.doc(payload[prefix+'glm-integrity.json']);expected_files={name.removeprefix(prefix):value['sha256'] for name,value in inventory.items() if name.startswith(prefix) and name.removeprefix(prefix) not in ('glm-integrity.json','live-report.json')}
        if seal.get('schema')!=1 or seal['files']!=expected_files:raise ValueError('original sealed member set differs')
        capture=receipt.get('capturePins')
        if route=='responses':
            if capture is not None:raise ValueError('failed Responses cannot acquire PASS capture pins')
        else:
            expected={'schema':1}
            for field,leaf in (('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')):expected[field]=b.sha(payload[prefix+leaf])
            if capture!=expected:raise ValueError('successful own original capture pins differ')
        mapping=b.doc(payload[prefix+'glm-execution-source.json'])
        if len(mapping)!=217 or any(source['files'].get(name)!=digest for name,digest in mapping.items()):raise ValueError('actual selected source differs')
        scenarios=report['scenarios'];ids=[scene['id'] for scene in scenarios]
        if len(ids)!=22 or len(set(ids))!=22 or set(ids)!=set(report['expected']):raise ValueError('complete22 inventory required')
        calls=attempts=records=wire_count=known_input=known_output=0;http=collections.Counter();usage=collections.Counter()
        failures=[]
        for scene in scenarios:
            case=prefix+'GLM11-'+scene['id'].replace('/','-')+'/';result_name=case+'attempt-1-result.json';wire_name=case+'attempt-1-exchanges.json'
            calls+=scene['calls'];attempts+=scene['httpAttempts'];actual=[];wire=[]
            if result_name in inventory:
                result=b.doc(payload[result_name]);actual=result['result']['metadata']['attempts'];wire=json.loads(payload[wire_name])
            if len(actual)!=scene['calls'] or len(wire)!=scene['httpAttempts']:raise ValueError('independent actual record count differs')
            for attempt in actual:
                records+=1;u=attempt.get('usage',{});values=[u.get(k,0) for k in ('input','cacheRead','cacheWrite','output')]
                if any(type(value) is not int or value<0 for value in values):raise ValueError('invalid usage counter')
                known_input+=sum(values[:3]);known_output+=values[3];usage[attempt.get('usageReporting','unreported')]+=1
            for exchange in wire:
                wire_count+=1;http[str(exchange.get('status','unreported'))]+=1
                if not exchange['url'].startswith('https://open.bigmodel.cn/'):raise ValueError('foreign wire target')
            if scene['id'].endswith('/native-history') and scene['outcome']!='PASS':raise ValueError('actual native history differs')
            if scene['outcome']=='FAIL':
                failures.append(scene['id']);assertions=b.doc(payload[case+'assertions.json'])
                if scene['id']=='glm-5.3/tool-result':
                    if actual or wire or scene['calls'] or not any(a['name']=='own tool prerequisite' and not a['pass'] for a in assertions['assertions']):raise ValueError('zero-call failed prerequisite differs')
                else:
                    result=b.doc(payload[result_name]);metadata=result['result']['metadata']
                    if (scene['durationMs']<90000 or metadata['attempts'][0]['code']!='deadline_exceeded' or
                        'ResponseHeaderTimeout (1m30s)' not in result['error'] or 'status' in wire[0]):raise ValueError('original header timeout differs')
            scenes.append(dict(route=route,id=scene['id'],outcome=scene['outcome'],calls=scene['calls'],attempts=scene['httpAttempts'],actualRecords=len(actual),durationMs=scene['durationMs']))
        budget=report['budget'];outcomes=dict(collections.Counter(s['outcome'] for s in scenarios))
        if (calls!=expected_calls or attempts!=expected_calls or records!=expected_calls or wire_count!=expected_calls or
            budget['callsUsed']!=calls or budget['httpAttempts']!=attempts or budget['plannedCalls']!=22 or budget['retries']!=0 or
            budget['maxCalls']!=24 or budget['maxAttempts']!=32 or budget['maxOutputTokens']!=4096 or budget['callSeconds']!=120 or budget['processSeconds']!=1800 or
            receipt['budget']['calls']!=calls or receipt['budget']['attempts']!=attempts or receipt['budget']['status']!=expected_outcome or
            outcomes!=({'PASS':19,'FAIL':3} if route=='responses' else {'PASS':22})):
            raise ValueError('actual scene/record/wire/budget sums differ')
        rows.append(dict(route=route,outcome=expected_outcome,complete=report['complete'],plannedScenes=22,outcomes=outcomes,calls=calls,attempts=attempts,
            endpoint=report['endpoint'],api=report['api'],accountVerification=report['accountVerification'],regionVerification=report['regionVerification'],privilegeVerification=report['privilegeVerification'],
            actualAttemptRecords=records,httpStatus=dict(http),usageReporting=dict(usage),knownInputSide=known_input,knownOutput=known_output,
            failures=failures,selectedFiles=len(mapping),integrityFiles=len(seal['files']),auditFindings=0,successfulCapturePins=capture is not None))
    total_calls=sum(row['calls'] for row in rows);total_attempts=sum(row['attempts'] for row in rows)
    if (total_calls,total_attempts)!=(65,65) or consumption['current']['calls']!=65 or consumption['current']['attempts']!=65 or consumption['total']['calls']!=365 or consumption['total']['attempts']!=365:
        raise ValueError('original phase-closed consumption differs')
    selectors=b.doc(payload['publication-check/unchanged-selectors.json']);publication=b.doc(payload['publication-check/strict-rejected-publication-command.json'])
    if selectors['before']!=selectors['after'] or selectors['allUnchanged'] is not True or publication['exitCode']!=1 or publication['credentialsRead'] is not False:
        raise ValueError('failed public publication or unchanged selectors differ')
    for name,digest in selectors['before'].items():
        original=source['files'].get('ai/live/testdata/glm-formal/'+name)
        if name=='live-captures.sha256':original=b.sha((selectors['before']['live-captures.json']+'\n').encode())
        if original!=digest:raise ValueError('selector original source pins differ')
    retired=b.doc(payload['historical83/retirement.json'])
    if retired['paidIntents']!=0 or retired['credentialsRead'] is not False:raise ValueError('historical preflight authority differs')
    for leaf,digest in retired['originalPins'].items():
        if b.sha(payload['historical83/'+leaf])!=digest:raise ValueError('retired preflight pin differs')
    usage_total=collections.Counter()
    for row in rows:usage_total.update(row['usageReporting'])
    return dict(schema=1,actualSource=expected_commit,collectionOutcome='FAIL',threeRoutePublication=False,jevPublicationInThisPhase=False,
        calls=total_calls,attempts=total_attempts,plannedScenes=66,outcomes={'PASS':63,'FAIL':3},priorCurrentReleaseCalls=300,priorCurrentReleaseAttempts=300,
        currentReleaseCalls=365,currentReleaseAttempts=365,accountingScope=consumption['accountingScope'],automaticReruns=False,routes=rows,
        usageReporting=dict(usage_total),unknownOrPartialUsageAttempts=sum(v for k,v in usage_total.items() if k!='complete'),
        knownInputSide=sum(row['knownInputSide'] for row in rows),knownOutput=sum(row['knownOutput'] for row in rows),
        usageBasis='Independent final Attempt sums: input+cacheRead+cacheWrite once; output once. Reasoning is already output; partial/unreported stays unknown. Known lower bound, not invoice.'),scenes


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
    for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        digest,name=line.split('  ',1)
        if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
    index=b.doc(b.read(BASE,'index.json',1<<20))
    if index.get('boundedHelperSHA256')!=HELPER_SHA or index['capsule']['id']!='actualGLM51' or set(index['routes'])!={'chat','responses','messages'}:
        raise ValueError('fixed own GLM collection index differs')
    archive,payload=b.verify_archive(a.common_root,index['capsule']);summary,scenes=analyse(b,index,payload)
    if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent summary differs')
    expected=[json.loads(line) for line in b.read(BASE,'scenarios.jsonl',1<<20).decode().splitlines()]
    if scenes!=expected:raise ValueError('independent scene inventory differs')
    print(json.dumps(dict(schema=1,verified=True,retentionCredentialsRead=False,retentionPaidCalls=0,archive=archive,collectionOutcome='FAIL',calls=65,attempts=65,
        threeRoutePublication=False,scope='integrity and actual consumption only; failed collection remains failed'),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
