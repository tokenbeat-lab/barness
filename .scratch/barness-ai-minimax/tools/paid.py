#!/usr/bin/env python3
"""Task-10 paid collection: prepare/check/preview/run-selected/summarize/keyless-plan.

Approval is a parent-written exact-source receipt. No command creates APPROVED.
"""
import argparse
import json
from pathlib import Path
import sys
import subprocess
sys.dont_write_bytecode=True
import paid_core as c
import paid_plan as p


def out_path(value):
    path=Path(value)
    if path.is_absolute() or '..' in path.parts:raise RuntimeError('repository-relative collection root required')
    result=(c.ROOT/path).resolve()
    if not result.is_relative_to(c.ROOT) or result==c.ROOT:raise RuntimeError('collection must stay inside repository')
    return result


def prepare(out):
    frozen=c.committed_sources()
    out.mkdir(parents=True,exist_ok=False,mode=0o700)
    c.dump(out/'source-at-start.json',frozen)
    inventory=p.jobs(c.ROOT,out,c.common_root())
    c.dump(out/'plan.json',c.redact(dict(schema=1,reference='barness-ai-minimax-ticket10',
        maxLegacyCalls=374,maxLegacyAttempts=398,providerCaps=p.PROVIDER_CAPS,
        miniMaxCandidate=dict(maxCalls=120,maxAttempts=144,maxIdentities=6),
        miniMaxFormal=dict(maxCalls=60,maxAttempts=72,maxIdentities=3),
        automaticProcessReruns=False,previousTaskBalanceCarried=False,routes=inventory)))
    catalog=c.doc(c.ROOT/'ai/release/catalog-snapshot.json')
    auth=p.authorizations(out,catalog['hash'])
    for name,value in auth.items():c.dump(out/name,value)
    for name,argv in (
        ('build-live',['go','test','-tags','live','-c','-o',out/'live.test','./ai/live']),
        ('build-text',['go','build','-o',out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
        ('build-visual',['go','build','-o',out/'visualeval','./ai/examples/visualeval'])):
        if c.execute(out,name,argv)!=0:raise RuntimeError('credential-free build failed')
    c.dump(out/'binaries.json',{n:c.digest(out/n) for n in ('live.test','chineseeval','visualeval')})
    prepared=dict(state='PREPARED',credentialsRead=False,planSHA256=c.digest(out/'plan.json'),
                  authorizationFiles={name:c.digest(out/name) for name in auth})
    c.dump(out/'prepared.json',prepared);c.guard(out)
    c.dump(out/'prerelease-source-review.template.json',dict(decision='PENDING',reviewedBy='',
        sourceCommit=frozen['sourceCommit'],sourceManifestSHA256=c.digest(out/'source-at-start.json'),
        binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=prepared['planSHA256'],
        authorizedRoutes=[],credentialScopes={}))
    return prepared


def check(out):
    c.guard(out)
    codes={}
    for name,args,timeout in (
        ('vet',['go','vet','./...'],600),('vet-live',['go','vet','-tags','live','./...'],600),
        ('no-key-live',['go','test','-tags','live','-count=1','-timeout=20m','./ai/live'],1260)):
        codes[name]=c.execute(out,name,args,env={'BARNESS_AI_EVIDENCE_DIR':str(out/'no-key-live')},timeout=timeout)
        if codes[name]!=0:break
    c.guard(out);accepted=dict(passed=len(codes)==3 and all(code==0 for code in codes.values()),
                              commands=codes,credentialsRead=False)
    c.dump(out/'checks.json',accepted)
    if not accepted['passed']:raise RuntimeError('no-key prerelease checks failed')
    return accepted


def marker(name):
    return c.common_root()/'.evidence/minimax-launches'/(name+'.json')


def counters(out,job):
    try:
        command=c.doc(out/(job['name']+'-command.json'))
        intent=c.doc(marker(job['name']))
        if command.get('state')!='COMPLETED' or intent.get('state')!='COMPLETED':
            raise ValueError('partial report cannot settle interrupted process')
        report=c.doc(p.report_path(out,job));b=report['budget']
        if job['name']=='protocol114':b=b['actual']
        if job['name'].startswith(('mm-','text-','visual-')) or job['name']=='protocol114':
            calls,attempts=b['calls'],b['attempts']
        else:calls,attempts=b['callsUsed'],b['httpAttempts']
        if (type(calls) is not int or type(attempts) is not int or
            not 0<=calls<=job['maxCalls'] or not 0<=attempts<=job['maxAttempts']):
            raise ValueError('invalid actual component budget')
        status=report.get('outcome',report.get('status','UNKNOWN'))
        if status not in ('PASS','FAIL','NOT_RUN','UNSUPPORTED'):raise ValueError('unknown report status')
        fields=('calls','attempts','callsUsed','httpAttempts','reservedAttempts','requestedOutput','reservedInput')
        safe={k:v for k,v in b.items() if k in fields}
        if any(type(v) is not int or v<0 for v in safe.values()):raise ValueError('invalid budget counter')
        cumulative=report.get('cumulativeIdentityBudget')
        if cumulative is not None:
            cumulative={k:v for k,v in cumulative.items() if k in fields}
            if any(type(v) is not int or v<0 for v in cumulative.values()):raise ValueError('invalid cumulative counter')
        if status=='NOT_RUN' and (calls or attempts):raise ValueError('NOT_RUN consumed resources')
        return dict(state='NOT_RUN' if status=='NOT_RUN' else 'REPORTED',calls=calls,attempts=attempts,
                    status=status,currentReportBudget=safe,cumulativeIdentityBudget=cumulative)
    except (OSError,RuntimeError,ValueError,KeyError,TypeError):
        return dict(state='UNKNOWN',calls=None,attempts=None,status='UNKNOWN')


def run_selected(out,job):
    frozen=c.reviewed(out,job)
    evidence=Path(job['evidenceRoot'])
    if evidence.exists() or evidence.is_symlink() or (out/(job['name']+'-command.json')).exists():
        raise RuntimeError('selected route output already exists; stale evidence cannot be reused')
    # Fail closed across output directories/worktrees, including pre-key crashes.
    if job.get('phase') in ('candidate','research'):
        sibling='mm-'+('research' if job['phase']=='candidate' else 'candidate')+'-'+(
            'm3' if job['purpose']=='standard' else 'm31')+'-'+job['route']
        if marker(sibling).exists():
            prior=c.doc(marker(sibling))
            if job['phase']=='research' or prior.get('budget',{}).get('state') not in ('REPORTED','NOT_RUN'):
                raise RuntimeError('prior stage outcome unknown or research order invalid; full cap consumed')
    intent=marker(job['name'])
    c.dump(intent,c.redact(dict(state='STARTING',route=job['name'],sourceCommit=frozen['sourceCommit'],
        maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],automaticReruns=False,
        collectionRoot=str(out),sourceVariable=job['sourceVariable'],credentialsRead=False)),exclusive=True)
    code=c.execute(out,job['name'],job['argv'],env=job['environment'],timeout=job['timeoutSeconds'],
        cwd=job['cwd'],source=job['sourceVariable'],injected=job['injectedCredential'],receipt=intent)
    c.guard(out)
    actual=counters(out,job)
    record=c.doc(intent);record['budget']=actual;c.dump(intent,record)
    c.dump(out/(job['name']+'-intent.json'),record)
    summarize(out)
    if code!=0:raise RuntimeError('selected process failed; inspect redacted command receipt')
    if actual['state']=='UNKNOWN':raise RuntimeError('selected process produced no usable current report')
    return dict(route=job['name'],exitCode=code,budget=actual)


def summarize(out):
    rows=[]
    for job in p.jobs(c.ROOT,out,c.common_root()).values():
        m=marker(job['name'])
        if not m.exists():actual=dict(state='NOT_RUN',calls=0,attempts=0,status='NOT_RUN')
        else:
            current=out/(job['name']+'-command.json')
            receipt=c.doc(m)
            actual=counters(out,job) if current.exists() else (
                receipt.get('budget',dict(state='UNKNOWN',calls=None,attempts=None,status='UNKNOWN'))
                if receipt.get('state')=='COMPLETED' else dict(state='UNKNOWN',calls=None,attempts=None,status='UNKNOWN'))
        known=actual['state'] in ('REPORTED','NOT_RUN')
        row=dict(route=job['name'],provider=job['provider'],maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],
            **actual,conservativeCalls=actual['calls'] if known else job['maxCalls'],
            conservativeAttempts=actual['attempts'] if known else job['maxAttempts'])
        if 'stage' in job:row.update(stage=job['stage'],identity=job['route']+'/'+job['model']+'/'+job['purpose'])
        rows.append(row)
    legacy=[r for r in rows if r['provider']!='minimax']
    total=dict(calls=sum(r['calls'] or 0 for r in legacy),attempts=sum(r['attempts'] or 0 for r in legacy),
               conservativeCalls=sum(r['conservativeCalls'] for r in legacy),
               conservativeAttempts=sum(r['conservativeAttempts'] for r in legacy),maxCalls=374,maxAttempts=398)
    if total['conservativeCalls']>374 or total['conservativeAttempts']>398:raise RuntimeError('legacy budget exceeded')
    for provider,cap in p.PROVIDER_CAPS.items():
        own=[r for r in legacy if r['provider']==provider]
        if sum(r['conservativeCalls'] for r in own)>cap:raise RuntimeError('provider budget exceeded')
        attemptcap=96 if provider=='zhipu' else cap
        if sum(r['conservativeAttempts'] for r in own)>attemptcap:raise RuntimeError('provider attempt budget exceeded')
    mini=[]
    for stage,calls,attempts in (('candidate',120,144),('formal',60,72)):
        grouped={}
        for row in rows:
            if row.get('stage')!=stage:continue
            own=grouped.setdefault(row['identity'],dict(calls=0,attempts=0,unknown=False))
            own['calls']+=row['calls'] or 0;own['attempts']+=row['attempts'] or 0
            own['unknown']|=row['state']=='UNKNOWN'
        if any(v['calls']>20 or v['attempts']>24 for v in grouped.values()):
            raise RuntimeError('MiniMax identity cumulative budget exceeded')
        # Unknown consumes the whole identity allowance, not one cap per phase.
        summary=dict(stage=stage,maxCalls=calls,maxAttempts=attempts,
            conservativeCalls=sum(20 if v['unknown'] else v['calls'] for v in grouped.values()),
            conservativeAttempts=sum(24 if v['unknown'] else v['attempts'] for v in grouped.values()),identities=grouped)
        if summary['conservativeCalls']>calls or summary['conservativeAttempts']>attempts:
            raise RuntimeError('MiniMax stage budget exceeded')
        mini.append(summary)
    result=dict(routes=rows,legacy=total,minimax=mini,credentialsRead=False)
    c.dump(out/'consumption.json',result)
    return result


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('prepare','check','preview','run-selected','summarize','keyless-plan'))
    parser.add_argument('out');parser.add_argument('route',nargs='?')
    args=parser.parse_args();out=out_path(args.out)
    if args.mode=='prepare':result=prepare(out)
    elif args.mode=='check':result=check(out)
    elif args.mode=='summarize':result=summarize(out)
    elif args.mode=='keyless-plan':c.guard(out);result=p.keyless_plan(c.ROOT,out)
    else:
        inventory=p.jobs(c.ROOT,out,c.common_root())
        if args.route not in inventory:raise RuntimeError('one fixed selected route required')
        job=inventory[args.route]
        if args.mode=='preview':c.guard(out);result=job|dict(credentialsRead=False,automaticReruns=False)
        else:result=run_selected(out,job)
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))


if __name__=='__main__':
    try:main()
    except (OSError,RuntimeError,ValueError,KeyError,subprocess.SubprocessError) as error:
        print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
