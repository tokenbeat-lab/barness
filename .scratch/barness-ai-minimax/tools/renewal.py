#!/usr/bin/env python3
"""Task-10 keyless catalog pins, GLM renewal, MiniMax origin index and NOT_RUN."""
import argparse
import json
from pathlib import Path
import re
import sys
sys.dont_write_bytecode=True
import paid_core as c
import renewal_common as u
import paid_plan as p


def catalog_pins(out):
    names=['ai/release/catalog-snapshot.json']+[
        'ai/examples/chineseeval/testdata/'+r+'-route-v1.json' for r in ('jev','luna')]
    before={n:u.file_bytes(c.ROOT/n) for n in names}
    configs={n:u.json_doc(c.ROOT/n) for n in names[1:]}
    for value in configs.values():
        if any(value.get(k)!=v for k,v in {'maxCalls':24,'maxAttempts':24,'maxQuestions':24,'maxRetries':0,'callTimeoutSeconds':15,'timeoutSeconds':600}.items()) or not value.get('instructions') or not value.get('datasetHash'):
            raise RuntimeError('fixed Chinese effect contract must remain unchanged')
    try:
        u.public(out,'catalog-snapshot',['go','run','./ai/release/cmd/releasegate','-write-snapshot'])
        snapshot=u.json_doc(c.ROOT/names[0]);version=snapshot['catalog']['version'];digest=snapshot['hash']
        for name,value in configs.items():
            value['catalogHash']=digest;value['catalogVersion']=version;c.dump(c.ROOT/name,value)
            check=u.json_doc(c.ROOT/name)
            old=json.loads(before[name]);old['catalogHash']=digest;old['catalogVersion']=version
            if check!=old:raise RuntimeError('noncatalog Chinese effect fields changed')
    except BaseException:
        for name,raw in before.items():(c.ROOT/name).write_bytes(raw)
        raise
    return dict(credentialsRead=False,catalogHash=digest,catalogVersion=version,changed=names,
                next='commit source/config pins before prepare/build/review')


def origins(out,values):
    allowed={'mm-global-'+s+'-m3-'+r for s in ('candidate','formal') for r in u.ROUTES}
    selected={}
    for value in values:
        route,sep,path=value.partition('=')
        if not sep or route not in allowed or route in selected:raise RuntimeError('distinct fixed M3 candidate/formal route=bundle required')
        selected[route]=u.bundle(path,'minimax-report.json')
    if not selected:raise RuntimeError('at least one own M3 bundle required')
    target=c.ROOT/'ai/live/testdata/minimax-formal/origins.json';index=u.json_doc(target,1<<20)
    if set(index)!= {'schema','entries'} or index['schema']!=1 or not isinstance(index['entries'],dict):raise RuntimeError('trusted origin index schema invalid')
    additions={}
    for name,base in selected.items():
        receipt_path=c.common_root()/'.evidence/minimax-launches'/(name+'.json')
        raw=u.file_bytes(receipt_path,2<<20);receipt=u.json_doc(receipt_path)
        report=u.json_doc(base/'minimax-report.json');origin=receipt.get('origin',{})
        fields={'schema','stage','identity','session','reportSha256','sealSha256','ledgerSha256','sourceCommit','sourceManifestSha256','binariesManifestSha256','planSha256'}
        if not isinstance(origin,dict) or set(origin)!=fields:raise RuntimeError('exact canonical origin schema required')
        stage=name.split('-')[2];route=name.removeprefix('mm-global-'+stage+'-m3-')
        if (receipt.get('state')!='COMPLETED' or receipt.get('exitCode')!=0 or receipt.get('credentialsRead') is not True or
            receipt.get('launcherAudit',{}).get('passed') is not True or receipt.get('route')!=name or
            report.get('outcome')!='PASS' or report.get('mode')!='live' or report.get('complete') is not True or
            report.get('identity',{}).get('model')!='MiniMax-M3' or report.get('identity',{}).get('purpose')!='standard' or
            report.get('identity')!=p.identity(dict(route=route,model='MiniMax-M3',purpose='standard')) or report.get('stage')!=stage or origin.get('schema')!=1 or
            origin.get('stage')!=stage or origin.get('identity')!=report.get('identity') or origin.get('session')!=report.get('session') or
            origin.get('sourceCommit')!=report.get('gitCommit') or receipt.get('sourceCommit')!=report.get('gitCommit')):
            raise RuntimeError('canonical successful own actual launcher origin required')
        for field,file in (('reportSha256','minimax-report.json'),('sealSha256','minimax-integrity.json'),('ledgerSha256','minimax-ledger.json')):
            if origin.get(field)!=u.sha(u.file_bytes(base/file)):raise RuntimeError('canonical origin raw bytes differ')
        if any(not u.HEX.fullmatch(origin.get(f,'')) for f in ('sourceManifestSha256','binariesManifestSha256','planSha256')):
            raise RuntimeError('canonical reviewed source/binary/plan pins missing')
        argv=['go','run','./ai/live/cmd/minimaxverify']
        if stage=='formal':argv+=['-formal']
        u.public(out,'verify-'+name,[*argv,base])
        if u.file_bytes(receipt_path,2<<20)!=raw:raise RuntimeError('canonical receipt changed during public verification')
        entry=dict(origin=origin,launcherReceiptSha256=u.sha(raw));key=origin['reportSha256']
        if key in index['entries'] and index['entries'][key]!=entry:raise RuntimeError('trusted origin collision')
        additions[key]=entry
    index['entries'].update(additions)
    if len((json.dumps(index,ensure_ascii=False,indent=2)+'\n').encode())>1<<20:raise RuntimeError('trusted origin index exceeds verifier import budget')
    c.dump(target,index)
    return dict(credentialsRead=False,indexed=len(additions),total=len(index['entries']),
                basis='only common-root canonical successful launcher receipts and public verifier')


def m31_not_run(out):
    frozen=c.guard(out)
    rows=[];ledger=c.common_root()/'.evidence/minimax-stages/candidate.jsonl'
    before=u.file_bytes(ledger) if ledger.exists() else None
    for route in u.ROUTES:
        name='m31-not-run-'+route;base=out/name
        if base.exists():raise RuntimeError('fresh NOT_RUN output required')
        env={'BARNESS_MINIMAX_ROUTE':route,'BARNESS_MINIMAX_MODEL':'MiniMax-M3.1-Flash-Preview',
             'BARNESS_MINIMAX_PURPOSE':'mplan','BARNESS_MINIMAX_ACCOUNT_ALIAS':'minimax-global-mplan-unverified',
             'BARNESS_MINIMAX_REGION_ALIAS':'global-unverified','BARNESS_AI_EVIDENCE_DIR':str(base),
             'BARNESS_MINIMAX_LEDGER':str(ledger)}
        u.public(out,name,[out/'live.test','-test.run=^TestMiniMaxCandidate$','-test.timeout=120s'],
                 timeout=150,env=env,cwd=c.ROOT/'ai/live')
        report=u.json_doc(u.bundle(base,'minimax-report.json')/'minimax-report.json')
        if (report.get('gitCommit')!=frozen['sourceCommit'] or report.get('outcome')!='NOT_RUN' or report.get('complete') is not False or report.get('mode')!='live' or report.get('entitlement','')!='' or
            report.get('identity')!=p.identity(dict(route=route,model='MiniMax-M3.1-Flash-Preview',purpose='mplan')) or
            report.get('stage')!='candidate' or any(type(report.get('budget',{}).get(k)) is not int or report['budget'][k]!=0 for k in ('calls','reservedAttempts','attempts','requestedOutput','reservedInput'))):
            raise RuntimeError('M3.1 keyless process must preserve own exact global zero-consumption NOT_RUN')
        rows.append(dict(route=route,model=report['identity']['model'],status='NOT_RUN',credentialsRead=False,calls=0,attempts=0))
        c.guard(out)
    after=u.file_bytes(ledger) if ledger.exists() else None
    if before!=after:raise RuntimeError('NOT_RUN modified canonical paid stage ledger')
    result=dict(credentialsRead=False,entitlementConfirmed=False,paidIntentsCreated=False,routes=rows)
    c.dump(out/'m31-not-run-summary.json',result);return result


def main():
    parser=argparse.ArgumentParser(description=__doc__);sub=parser.add_subparsers(dest='mode',required=True)
    for mode in ('catalog-pins','glm-archive','glm-views','origins','m31-not-run'):
        p=sub.add_parser(mode);p.add_argument('out')
        if mode=='glm-archive':
            for route in u.ROUTES:p.add_argument('--'+route,required=True)
        if mode=='origins':p.add_argument('--bundle',action='append',required=True)
    a=parser.parse_args()
    if a.mode=='m31-not-run':out=u.repo_path(a.out);result=m31_not_run(out)
    else:
        out=u.output(a.out,existing=a.mode=='glm-views')
        if a.mode=='catalog-pins':result=catalog_pins(out)
        elif a.mode=='origins':result=origins(out,a.bundle)
        else:
            import renewal_glm as g
            result=g.capture(out,{r:getattr(a,r) for r in u.ROUTES}) if a.mode=='glm-archive' else g.views(out)
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))

if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
