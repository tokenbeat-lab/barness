#!/usr/bin/env python3
"""Prospective GLM corrective-1: no authority until explicit human approval.

Exactly three new full22 processes; original canonical intents never change.
Distinct preparation is deliberate: this round has a separate approval and
carried-consumption contract, not a reset of the original collection.
"""
import argparse
import json
from pathlib import Path
import re
import sys
sys.dont_write_bytecode=True
import paid_core as c
import paid_plan as p
import paid

ROUND='corrective-1'
ROUTES=('chat','responses','messages')
ACCOUNTING_SCOPE='current MiniMax-release GLM renewal: original66/66 plus corrective-1 only; prior GLM-task paid history remains separate and immutable'
BUDGET=dict(priorCalls=66,priorAttempts=66,additionalMaxCalls=72,additionalMaxAttempts=96,
    lifetimeGLMMaxCalls=138,lifetimeGLMMaxAttempts=162,otherLegacyMaxCalls=302,
    otherLegacyMaxAttempts=302,legacyMaxCalls=440,legacyMaxAttempts=464)


def originals():
    result={};commits=set()
    for route in ROUTES:
        name='glm-'+route;path=paid.marker(name);record=c.doc(path);budget=record.get('budget',{})
        if (record.get('state')!='COMPLETED' or record.get('route')!=name or
            record.get('sourceVariable')!='GLM_KEY' or record.get('credentialsRead') is not True or
            record.get('launcherAudit',{}).get('passed') is not True or
            record.get('maxCalls')!=24 or record.get('maxAttempts')!=32 or
            budget.get('state')!='REPORTED' or budget.get('status') not in ('PASS','FAIL') or
            type(budget.get('calls')) is not int or budget['calls']!=22 or
            type(budget.get('attempts')) is not int or budget['attempts']!=22 or
            re.fullmatch('[a-f0-9]{40}',record.get('sourceCommit','')) is None or
            re.fullmatch('[a-f0-9]{64}',record.get('reportedFileSha256','')) is None):
            raise RuntimeError('immutable original three complete22 consumption receipts required; unknown cannot reset')
        commits.add(record['sourceCommit'])
        result[name]=dict(receiptSHA256=c.digest(path),sourceCommit=record['sourceCommit'],
            reportSHA256=record['reportedFileSha256'],calls=22,attempts=22,status=budget['status'])
    if len(commits)!=1:raise RuntimeError('original GLM round must retain one actual source commit')
    return result


def jobs(out):
    original=p.jobs(c.ROOT,out,c.common_root());result={}
    for route in ROUTES:
        job=original['glm-'+route];name='glm-corrective-1-'+route
        job.update(name=name,glmRoute=route,correctiveRound=ROUND,
            evidenceRoot=str(out/name))
        job['environment']['BARNESS_AI_EVIDENCE_DIR']=str(out/name)
        result[name]=job
    return result


def prepare(out):
    prior=originals();frozen=c.committed_sources()
    out.mkdir(parents=True,exist_ok=False,mode=0o700)
    c.dump(out/'source-at-start.json',frozen)
    c.dump(out/'plan.json',c.redact(dict(schema=1,round=ROUND,accountingScope=ACCOUNTING_SCOPE,approval='PENDING HUMAN APPROVAL',
        budget=BUDGET,originalReceipts=prior,routes=jobs(out),automaticProcessReruns=False,
        noSceneSplice=True,maxOutputTokens=4096,callSeconds=120,processSeconds=1800,defaultRetries=0)))
    if c.execute(out,'build-live',['go','test','-tags','live','-c','-o',out/'live.test','./ai/live'])!=0:
        raise RuntimeError('credential-free corrective build failed')
    c.dump(out/'binaries.json',{'live.test':c.digest(out/'live.test')})
    c.dump(out/'prepared.json',dict(state='PREPARED',credentialsRead=False,
        planSHA256=c.digest(out/'plan.json'),authorizationFiles={}))
    pins=dict(sourceCommit=frozen['sourceCommit'],sourceManifestSHA256=c.digest(out/'source-at-start.json'),
        binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'))
    c.dump(out/'prerelease-source-review.template.json',dict(decision='PENDING',reviewedBy='',
        **pins,authorizedRoutes=[],credentialScopes={}))
    c.dump(out/'human-approval.template.json',dict(decision='PENDING',authority='',replyReference='',
        round=ROUND,accountingScope=ACCOUNTING_SCOPE,**pins,**BUDGET,originalReceipts=prior))
    c.guard(out)
    return dict(state='PREPARED',approval='PENDING HUMAN APPROVAL',credentialsRead=False,**pins)


def guard(out):
    frozen=c.guard(out);plan=c.doc(out/'plan.json')
    if (plan.get('round')!=ROUND or plan.get('accountingScope')!=ACCOUNTING_SCOPE or plan.get('budget')!=BUDGET or
        plan.get('originalReceipts')!=originals() or plan.get('routes')!=c.redact(jobs(out))):
        raise RuntimeError('fixed corrective plan or carried original receipts changed')
    return frozen


def approved(out,job):
    frozen=guard(out);c.reviewed(out,job)
    approval=c.doc(out/'human-approval.json');template=c.doc(out/'human-approval.template.json')
    expected=template|dict(decision='APPROVED',authority='explicit-user-reply',replyReference=approval.get('replyReference'))
    if (approval!=expected or not isinstance(approval.get('replyReference'),str) or not approval['replyReference'].strip() or
        any(approval.get(k)!=v for k,v in BUDGET.items()) or
        approval.get('originalReceipts')!=originals() or
        c.doc(out/'prerelease-source-review.json').get('credentialScopes',{}).get('GLM_KEY')!='glm-corrective-1'):
        raise RuntimeError('explicit human approval of exact corrective source/binary/plan and additional72/96 required')
    # Template is advisory, not authority: independently recompute every pin.
    pins=dict(sourceCommit=frozen['sourceCommit'],sourceManifestSHA256=c.digest(out/'source-at-start.json'),
        binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'))
    if any(approval.get(k)!=v for k,v in pins.items()) or approval.get('round')!=ROUND or approval.get('accountingScope')!=ACCOUNTING_SCOPE:
        raise RuntimeError('human approval exact pins differ')
    if approval.get('checksSHA256')!=c.digest(out/'checks.json'):
        raise RuntimeError('explicit human approval must pin successful exact checks')
    return frozen


def rows():
    result=[]
    for route in ROUTES:
        name='glm-corrective-1-'+route;path=paid.marker(name)
        if not path.exists():continue
        receipt=c.doc(path);budget=receipt.get('budget',{})
        known=(receipt.get('state')=='COMPLETED' and receipt.get('correctiveRound')==ROUND and
            budget.get('state') in ('REPORTED','NOT_RUN') and
            type(budget.get('calls')) is int and 0<=budget['calls']<=24 and
            type(budget.get('attempts')) is int and 0<=budget['attempts']<=32)
        result.append(dict(route=name,provider='zhipu',maxCalls=24,maxAttempts=32,
            state=budget['state'] if known else 'UNKNOWN',status=budget.get('status','UNKNOWN') if known else 'UNKNOWN',
            calls=budget['calls'] if known else None,attempts=budget['attempts'] if known else None,
            conservativeCalls=budget['calls'] if known else 24,conservativeAttempts=budget['attempts'] if known else 32))
    return result


def summarize(out):
    guard(out);current=rows()
    consumed=dict(calls=66+sum(r['calls'] or 0 for r in current),attempts=66+sum(r['attempts'] or 0 for r in current),
        conservativeCalls=66+sum(r['conservativeCalls'] for r in current),conservativeAttempts=66+sum(r['conservativeAttempts'] for r in current),
        maxCalls=138,maxAttempts=162)
    if consumed['conservativeCalls']>138 or consumed['conservativeAttempts']>162:raise RuntimeError('GLM lifetime cap exceeded')
    result=dict(round=ROUND,accountingScope=ACCOUNTING_SCOPE,originalReceipts=originals(),routes=current,lifetime=consumed,
        budget=BUDGET,automaticReruns=False,credentialsRead=False)
    c.dump(out/'corrective-consumption.json',result);return result


def disclosure():
    """Later ordinary plans disclose inherited authority and conservative spend."""
    current=rows()
    if not current:return None
    prior=originals();receipts={}
    for row in current:
        path=paid.marker(row['route']);record=c.doc(path)
        if (record.get('correctiveRound')!=ROUND or record.get('route')!=row['route'] or
            record.get('originalReceipts')!=prior or
            any(re.fullmatch('[a-f0-9]{64}',record.get(field,'')) is None for field in
                ('humanApprovalSha256','sourceManifestSha256','binariesManifestSha256','planSha256'))):
            raise RuntimeError('inherited corrective budget requires canonical approved intent pins')
        receipts[row['route']]=dict(receiptSHA256=c.digest(path),sourceCommit=record['sourceCommit'])
    lifetime=dict(calls=66+sum(r['calls'] or 0 for r in current),attempts=66+sum(r['attempts'] or 0 for r in current),
        conservativeCalls=66+sum(r['conservativeCalls'] for r in current),
        conservativeAttempts=66+sum(r['conservativeAttempts'] for r in current))
    return dict(accountingScope=ACCOUNTING_SCOPE,budget=BUDGET,originalReceipts=prior,
        correctiveReceipts=receipts,routes=current,lifetime=lifetime,
        spentRoutes=['glm-'+r for r in ROUTES]+[r['route'] for r in current])


def run_selected(out,job):
    frozen=approved(out,job);summarize(out)
    evidence=Path(job['evidenceRoot']);intent=paid.marker(job['name'])
    if evidence.exists() or evidence.is_symlink() or (out/(job['name']+'-command.json')).exists():
        raise RuntimeError('fresh corrective output required')
    c.dump(intent,c.redact(dict(state='STARTING',route=job['name'],correctiveRound=ROUND,
        sourceCommit=frozen['sourceCommit'],sourceVariable='GLM_KEY',credentialsRead=False,
        maxCalls=24,maxAttempts=32,collectionRoot=str(out),automaticReruns=False,
        originalReceipts=originals(),humanApprovalSha256=c.digest(out/'human-approval.json'),
        sourceManifestSha256=c.digest(out/'source-at-start.json'),binariesManifestSha256=c.digest(out/'binaries.json'),
        planSha256=c.digest(out/'plan.json'))),exclusive=True)
    code=c.execute(out,job['name'],job['argv'],env=job['environment'],timeout=1800,cwd=job['cwd'],
        source='GLM_KEY',injected='BARNESS_GLM_KEY',receipt=intent)
    guard(out);record=c.doc(intent);actual=paid.counters(out,job);record['budget']=actual
    if actual['state'] in ('REPORTED','NOT_RUN'):record['reportedFileSha256']=c.digest(p.report_path(out,job))
    c.dump(intent,record)
    if code==0 and actual['status']=='PASS':
        record['capturePins']=paid.pin_glm_capture(out,job,record);c.dump(intent,record)
    c.dump(out/(job['name']+'-intent.json'),record);summarize(out)
    if code!=0 or actual['state']=='UNKNOWN':raise RuntimeError('corrective process failed; immutable failure consumes allowance, no R2')
    return dict(route=job['name'],exitCode=code,budget=actual)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('prepare','check','preview','run-selected','summarize'))
    parser.add_argument('out');parser.add_argument('route',nargs='?');a=parser.parse_args();out=paid.out_path(a.out)
    if a.mode=='prepare':result=prepare(out)
    elif a.mode=='check':
        guard(out);result=paid.check(out)
        template=c.doc(out/'human-approval.template.json');template['checksSHA256']=c.digest(out/'checks.json')
        c.dump(out/'human-approval.template.json',template)
    elif a.mode=='summarize':result=summarize(out)
    else:
        inventory=jobs(out)
        if a.route not in inventory:raise RuntimeError('one fixed full22 corrective-1 route required; no R2 or scene selector')
        job=inventory[a.route]
        if a.mode=='preview':guard(out);result=job|dict(approval='PENDING HUMAN APPROVAL',credentialsRead=False)
        else:result=run_selected(out,job)
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
