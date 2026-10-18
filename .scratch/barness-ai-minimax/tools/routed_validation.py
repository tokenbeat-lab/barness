#!/usr/bin/env python3
"""One finite routed validation, split around committed GLM selector renewal."""
import argparse
import json
from pathlib import Path
import sys
sys.dont_write_bytecode=True
import paid_core as c
import paid_plan as p
import paid
import routed_contract as r
import routed_evidence as e


def prepare(out,phase):
    prior=r.prior();frozen=c.committed_sources()
    out.mkdir(parents=True,exist_ok=False,mode=0o700)
    c.dump(out/'source-at-start.json',frozen)
    c.dump(out/'plan.json',c.redact(dict(schema=1,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,collection=e.location(out),
        priorReceipts=prior,routes=r.jobs(out,phase),automaticProcessReruns=False,
        authorizationBasis='explicit continuation request; finite ceiling is not an explicit numeric reply',
        noSceneSplice=True,proxy=r.PROXY)))
    catalog=c.doc(c.ROOT/'ai/release/catalog-snapshot.json')
    auth=p.authorizations(out,catalog['hash']) if phase=='legacy' else {}
    for name,value in auth.items():
        field='authorizationReference' if name=='protocol-authorization.json' else 'reference'
        value[field]='barness-ai-minimax-ticket13';c.dump(out/name,value)
    builds=[('live.test',['go','test','-tags','live','-c','-o',out/'live.test','./ai/live'])]
    if phase=='legacy':builds += [
        ('chineseeval',['go','build','-o',out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
        ('visualeval',['go','build','-o',out/'visualeval','./ai/examples/visualeval'])]
    for name,argv in builds:
        if c.execute(out,'build-'+name,argv)!=0:raise RuntimeError('credential-free routed build failed')
    c.dump(out/'binaries.json',{name:c.digest(out/name) for name,_ in builds})
    c.dump(out/'prepared.json',dict(state='PREPARED',credentialsRead=False,
        planSHA256=c.digest(out/'plan.json'),authorizationFiles={n:c.digest(out/n) for n in auth}))
    pins=review_pins(out)
    c.dump(out/'prerelease-source-review.template.json',dict(decision='PENDING',reviewedBy='',
        **pins,authorizedRoutes=[],credentialScopes={}))
    c.dump(out/'execution-authorization.template.json',dict(decision='PENDING',authority='',
        requestReference='',requestText=r.REQUEST,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,
        **pins,checksSHA256='',sourceReviewSHA256=''))
    guard(out)
    return dict(state='PREPARED',round=r.ROUND,phase=phase,credentialsRead=False,**pins)


def review_pins(out):
    return dict(sourceCommit=c.doc(out/'source-at-start.json')['sourceCommit'],
        sourceManifestSHA256=c.digest(out/'source-at-start.json'),
        binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'))


def guard(out):
    frozen=c.guard(out);plan=c.doc(out/'plan.json');phase=plan.get('phase')
    if (phase not in ('glm','legacy') or plan.get('round')!=r.ROUND or plan.get('budget')!=r.BUDGET or
        plan.get('priorReceipts')!=r.prior() or plan.get('routes')!=c.redact(r.jobs(out,phase)) or
        plan.get('proxy')!=r.PROXY or plan.get('collection')!=e.location(out) or plan.get('accountingScope')!=r.ACCOUNTING_SCOPE or plan.get('automaticProcessReruns') is not False):
        raise RuntimeError('fixed routed plan or carried prior receipts changed')
    return frozen


def approved(out,job):
    frozen=guard(out);c.reviewed(out,job)
    authorization=c.doc(out/'execution-authorization.json')
    expected=dict(decision='APPROVED',authority='explicit-user-request',
        requestReference=authorization.get('requestReference'),requestText=r.REQUEST,
        round=r.ROUND,phase=c.doc(out/'plan.json')['phase'],budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,**review_pins(out),
        checksSHA256=c.digest(out/'checks.json'),sourceReviewSHA256=c.digest(out/'prerelease-source-review.json'))
    scope=r.ROUND+'/'+job['networkRoute']
    if (authorization!=expected or not isinstance(authorization.get('requestReference'),str) or
        not authorization['requestReference'].strip() or
        c.doc(out/'prerelease-source-review.json').get('credentialScopes',{}).get(job['sourceVariable'])!=scope):
        raise RuntimeError('exact reviewed source/checks and explicit continuation request receipt required')
    return frozen


def counters(out,job):
    """Settle native producer DTOs while the outer intent stays canonical."""
    try:
        command_path=out/(job['baseName']+'-command.json')
        command=c.doc(command_path);intent=c.doc(paid.marker(job['name']))
        if command.get('state')!='COMPLETED' or intent.get('state')!='COMPLETED':
            raise ValueError('interrupted component cannot settle')
        if intent.get('commandReceiptSha256') and c.digest(command_path)!=intent['commandReceiptSha256']:
            raise ValueError('original command receipt changed')
        path=r.report_path(out,job)
        if intent.get('reportedFileSha256') and c.digest(path)!=intent['reportedFileSha256']:
            raise ValueError('original report changed')
        return e.native_counters(c.doc(path),job)
    except (OSError,RuntimeError,ValueError,KeyError,TypeError):
        return dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)


def summarize(out):
    guard(out);rows=[]
    for name,job in r.jobs(out).items():
        marker=paid.marker(name)
        actual=dict(state='NOT_RUN',status='NOT_RUN',calls=0,attempts=0)
        if marker.exists() or marker.is_symlink():
            actual=e.settled(job)
        known=(actual.get('state') in ('REPORTED','NOT_RUN') and actual.get('status') in ('PASS','FAIL','NOT_RUN','UNSUPPORTED') and
            type(actual.get('calls')) is int and 0<=actual['calls']<=job['maxCalls'] and
            type(actual.get('attempts')) is int and 0<=actual['attempts']<=job['maxAttempts'])
        if not known:actual=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
        else:actual={field:actual[field] for field in ('state','status','calls','attempts')}
        rows.append(dict(route=name,baseName=job['baseName'],provider=job['provider'],networkRoute=job['networkRoute'],
            maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],**actual,
            conservativeCalls=actual['calls'] if known else job['maxCalls'],
            conservativeAttempts=actual['attempts'] if known else job['maxAttempts']))
    current=dict(calls=sum(row['calls'] or 0 for row in rows),attempts=sum(row['attempts'] or 0 for row in rows),
        conservativeCalls=sum(row['conservativeCalls'] for row in rows),conservativeAttempts=sum(row['conservativeAttempts'] for row in rows),
        maxCalls=374,maxAttempts=398)
    if current['conservativeCalls']>374 or current['conservativeAttempts']>398:raise RuntimeError('new routed finite budget exceeded')
    result=dict(round=r.ROUND,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,priorReceipts=r.prior(),routes=rows,current=current,
        total=dict(calls=300+current['calls'],attempts=300+current['attempts'],
            conservativeCalls=300+current['conservativeCalls'],conservativeAttempts=300+current['conservativeAttempts'],
            maxCalls=674,maxAttempts=698),automaticProcessReruns=False,credentialsRead=False)
    c.dump(out/'routed-consumption.json',result);return result


def run_selected(out,job):
    frozen=approved(out,job);summarize(out)
    evidence=Path(job['evidenceRoot']);name=job['name'];base=job['baseName'];intent=paid.marker(name)
    if evidence.exists() or evidence.is_symlink() or (out/(base+'-command.json')).exists():
        raise RuntimeError('fresh native output required; stale evidence cannot be reused')
    c.dump(intent,c.redact(dict(state='STARTING',route=name,baseName=base,validationRound=r.ROUND,
        phase=c.doc(out/'plan.json')['phase'],sourceCommit=frozen['sourceCommit'],sourceVariable=job['sourceVariable'],
        credentialsRead=False,maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],networkRoute=job['networkRoute'],
        collectionRoot=str(out),collectionOwnerSHA256=e.location(out)['rootSHA256'],automaticReruns=False,accountingScope=r.ACCOUNTING_SCOPE,priorReceipts=r.prior(),
        authorizationSha256=c.digest(out/'execution-authorization.json'),sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),
        sourceManifestSha256=c.digest(out/'source-at-start.json'),binariesManifestSha256=c.digest(out/'binaries.json'),
        planSha256=c.digest(out/'plan.json'))),exclusive=True)
    # Native command names deliberately match existing artifact audit roots.
    # Only the global authority namespace changes; producer DTOs stay native.
    code=c.execute(out,base,job['argv'],env=job['environment'],timeout=job['timeoutSeconds'],cwd=job['cwd'],
        source=job['sourceVariable'],injected=job['injectedCredential'],receipt=intent)
    guard(out);record=c.doc(intent);record['commandReceiptSha256']=c.digest(out/(base+'-command.json'))
    c.dump(intent,record);actual=counters(out,job);record['budget']=actual
    if actual['state'] in ('REPORTED','NOT_RUN'):record['reportedFileSha256']=c.digest(r.report_path(out,job))
    c.dump(intent,record)
    if code==0 and base.startswith('glm-') and actual['status']=='PASS':
        record['capturePins']=paid.pin_glm_capture(out,job|dict(name=base),record);c.dump(intent,record)
    c.dump(out/(name+'-intent.json'),record);summarize(out)
    if code!=0 or actual['state']=='UNKNOWN':raise RuntimeError('fixed routed process failed; retain consumption, no automatic next round')
    return dict(route=name,baseName=base,exitCode=code,budget=actual)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('prepare','check','preview','run-selected','summarize'))
    parser.add_argument('out');parser.add_argument('route',nargs='?')
    parser.add_argument('--phase',choices=('glm','legacy'))
    a=parser.parse_args();out=paid.out_path(a.out)
    if a.mode=='prepare':
        if not a.phase:raise RuntimeError('prepare requires fixed glm or legacy phase')
        result=prepare(out,a.phase)
    elif a.mode=='check':
        guard(out);result=paid.check(out);guard(out)
        code=c.execute(out,'no-key-visual-proxy',['go','test','./ai/e2e','-run',
            '^TestVisualEvaluationProxyProcess$','-count=1','-timeout=5m'],timeout=360)
        result['commands']['no-key-visual-proxy']=code;result['passed']=result['passed'] and code==0
        c.dump(out/'checks.json',result);guard(out)
        if not result['passed']:raise RuntimeError('no-key visual routing checks failed')
        template=c.doc(out/'execution-authorization.template.json')
        template['checksSHA256']=c.digest(out/'checks.json');c.dump(out/'execution-authorization.template.json',template)
    else:
        if a.phase:raise RuntimeError('phase is fixed by preparation')
        guard(out);plan=c.doc(out/'plan.json');inventory=r.jobs(out,plan['phase'])
        if a.mode=='summarize':result=summarize(out)
        else:
            if a.route not in inventory:raise RuntimeError('one fixed complete routed validation job required')
            job=inventory[a.route]
            if a.mode=='preview':result=job|dict(credentialsRead=False,authorization='PENDING independent review and request receipt')
            else:result=run_selected(out,job)
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
