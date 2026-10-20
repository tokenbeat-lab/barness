"""One canonical fixed3 claim per job; no counter, namespace or USD reset."""
from contextlib import contextmanager
import fcntl
import os
from pathlib import Path
import stat
import recovery_contract as r
import recovery_history as h
import recovery_state as s
c=r.c
UNKNOWN=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)

def marker(name):
    if name not in r.RESERVES:raise RuntimeError('fixed24 own marker required')
    return h.f.owned('.evidence/minimax-launches/'+r.ROUND+'-'+name+'.json')
def authority():return h.f.owned('.evidence/minimax-renewals/'+r.ROUND)
def sync_dir(path):
    fd=os.open(path,os.O_RDONLY)
    try:os.fsync(fd)
    finally:os.close(fd)
@contextmanager
def locked():
    base=authority();base.mkdir(parents=True,exist_ok=True,mode=0o700);sync_dir(base.parent);lock=base/'launch.lock'
    fd=os.open(lock,os.O_CREAT|os.O_RDWR|os.O_NOFOLLOW|os.O_NONBLOCK,0o600)
    try:
        if not stat.S_ISREG(os.fstat(fd).st_mode):raise RuntimeError('regular canonical lock required')
        fcntl.flock(fd,fcntl.LOCK_EX);yield base
    finally:os.close(fd)

def summarize():
    prior=h.facts();rows=[]
    for name,job in r.jobs(c.ROOT/'.evidence/unused-recovery-summary').items():
        p=marker(name)
        if not p.exists():
            if p.is_symlink():raise RuntimeError('canonical claim symlink refused')
            continue
        value=s.doc(p)
        if any(value.get(k)!=v for k,v in dict(validationRound=r.ROUND,nativeRenewalId=r.ROUND,baseName=name,maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],reservedMicrousd=job['reservedMicrousd'],**r.policy()).items()):raise RuntimeError('canonical3 scope/policy/reserve changed')
        rows.append(dict(name=name,state=value['state'],budget=value['budget'],maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],reservedMicrousd=value['reservedMicrousd']))
    calls=sum(v['maxCalls'] for v in rows);attempts=sum(v['maxAttempts'] for v in rows);usd=sum(v['reservedMicrousd'] for v in rows)
    if calls>506 or attempts>518 or usd>19200000:raise RuntimeError('fixed3 total exceeded')
    return dict(round=r.ROUND,predecessor=prior,routes=rows,originalProcessCount=17,newProcessCount=len(rows),actualProcessCount=17+len(rows),effectiveRoleCount=9+len(rows),retainedUnknownProcesses=2,claimedCalls=456+calls,claimedAttempts=492+attempts,knownReportedCalls=303+sum(v['budget'].get('calls') or 0 for v in rows),knownReportedAttempts=302+sum(v['budget'].get('attempts') or 0 for v in rows),newReservedMicrousd=usd,reservedMicrousd=10700000+usd,campaignMaxCalls=962,campaignMaxAttempts=1010,maxMicrousd=30000000,credentialsRead=False)

def require_checks(out):
    import final_execute
    final_execute.require_checks(out)

def claim(out,job,prepared,auth):
    with locked() as base:
        if marker(job['name']).exists() or marker(job['name']).is_symlink():raise RuntimeError('fixed3 job already claimed in any output or checkout')
        summary=summarize()
        names={v['name'] for v in summary['routes']};index=len(names)
        if index>=len(r.EXECUTION_ORDER) or names!=set(r.EXECUTION_ORDER[:index]) or job['name']!=r.EXECUTION_ORDER[index]:raise RuntimeError('fixed execution order requires M3 Chat full native PASS before any later job')
        if any(v['state']!='COMPLETED' or v['budget']['status']!='PASS' for v in summary['routes']):raise RuntimeError('previous3 process failed or unknown; no automatic continuation')
        for name in r.EXECUTION_ORDER[:index]:
            import recovery_native
            recovery_native.require_consumed(out,r.jobs(out)[name],s.doc(marker(name)))
        identity={k:auth[k] for k in ('round','validationRound','nativeRenewalId','spendingGuardPolicyId','spendingGuardSourceSHA256','requestText','requestReference','actualHumanApprovalSHA256','maxCalls','maxAttempts','campaignMaxCalls','campaignMaxAttempts','newReservedMicrousd','retainedReservedMicrousd','maxUSD','automaticProcessReruns','predecessor','eligibilityAuthorizationSHA256','raceAuthorizationSHA256')}
        proof=base/'operator-authority.json'
        if proof.exists():
            if s.doc(proof)!=identity:raise RuntimeError('same fixed3 actual human authority required across all phases')
        else:c.dump(proof,identity,exclusive=True)
        sync_dir(base)
        record=dict(state='STARTING',route=r.ROUND+'-'+job['name'],baseName=job['name'],validationRound=r.ROUND,nativeRenewalId=r.ROUND,phase='recovery3',sourceCommit=prepared['sourceCommit'],sourceVariable=job['sourceVariable'],credentialsRead=False,maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],reservedMicrousd=job['reservedMicrousd'],automaticReruns=False,collectionRoot=str(out),networkRoute=job['networkRoute'],budget=UNKNOWN.copy(),**r.policy(),originalCampaignApprovalSHA256=auth['eligibilityAuthorizationSHA256'],actualHumanApprovalSHA256=auth['actualHumanApprovalSHA256'],usdLedger=str(out/(job['name']+'-usd.json')))
        for field,key in [('sourceManifestSha256','sourceManifestSHA256'),('binariesManifestSha256','binariesManifestSHA256'),('planSha256','planSHA256')]:record[field]=prepared[key]
        record.update(sourceReviewSha256=s.pin(out/'prerelease-source-review.json'),authorizationSha256=s.pin(out/'execution-authorization.json'))
        c.dump(marker(job['name']),c.redact(record),exclusive=True);sync_dir(marker(job['name']).parent);return marker(job['name'])

def run_selected(out,job):
    import final_execute as old
    import final_process
    import recovery_native as native
    prepared,auth,scope,reserve=s.approved(out,job);require_checks(out);native.require_component(out,job,auth,scope)
    if Path(job['evidenceRoot']).exists() or (out/(job['name']+'-command.json')).exists() or (out/(job['name']+'-usd.json')).exists():raise RuntimeError('fresh3 output required; no process restart')
    intent=claim(out,job,prepared,auth);record=native.preflight(out,job,s.doc(intent));c.dump(intent,record)
    usd=out/(job['name']+'-usd.json');env=job['environment']|dict(BARNESS_FINAL_ROUND=r.SPENDING_GUARD_POLICY_SELECTOR,BARNESS_FINAL_BUDGET=str(usd))
    c.dump(usd,dict(schema=1,round=r.SPENDING_GUARD_POLICY_SELECTOR,maxMicrousd=reserve,reservedMicrousd=0,attempts=0,basis='Unchanged campaign2 spending policy; execution authority3, no refund, not invoice.'),exclusive=True)
    argv=job['argv'];proxy=job['networkRoute']
    if proxy!='direct':
        if job['name'].startswith(('text-','visual-')):argv=[*argv,'-proxy',proxy]
        else:env['BARNESS_REGIONAL_PROXY' if 'combo' in job else 'BARNESS_MINIMAX_PROXY' if job['name'].startswith('mm-') else 'BARNESS_AI_LIVE_PROXY']=proxy
    env.update(BARNESS_AI_LIVE_ACCOUNT_ALIAS=scope['accountAlias']+'@'+scope['regionAlias'],BARNESS_MINIMAX_ACCOUNT_ALIAS=scope['accountAlias'],BARNESS_MINIMAX_REGION_ALIAS=scope['regionAlias'],BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=scope['accountAlias']+'@'+scope['regionAlias'])
    job=job|dict(executionProvenance={k:record[k] for k in ('validationRound','nativeRenewalId','spendingGuardPolicyId','spendingGuardSourceSHA256','usdLedger','originalCampaignApprovalSHA256','actualHumanApprovalSHA256')})
    s.guard(out);code=final_process.run(out,job,argv,env,intent);record=s.doc(intent);record.update(commandReceiptSha256=s.pin(out/(job['name']+'-command.json')),budget=old.settle(out,job) if code==0 else UNKNOWN.copy())
    if code==0 and record['budget']['status']=='PASS':
        value=s.doc(usd)
        if value.get('round')!=r.SPENDING_GUARD_POLICY_SELECTOR or value.get('maxMicrousd')!=reserve or type(value.get('attempts')) is not int or value['attempts']<record['budget']['attempts'] or type(value.get('reservedMicrousd')) is not int or not 0<value['reservedMicrousd']<=reserve:record['budget']['status']='FAIL'
    c.dump(intent,record);s.guard(out)
    if code==0 and record['budget']['status']=='PASS':
        try:record=native.consume(out,job,record);c.dump(intent,record)
        except Exception:record['budget']['status']='FAIL';c.dump(intent,record);raise
    if code!=0 or record['budget']['status']!='PASS':raise RuntimeError('process failed or unknown; preserve all3 claims/USD and stop')
    return dict(route=job['name'],exitCode=code,budget=record['budget'],**r.policy())
