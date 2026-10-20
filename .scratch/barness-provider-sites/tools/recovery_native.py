"""Native current3 facts, exact component grants and full-process consumers."""
import datetime as dt
import json
from pathlib import Path
import recovery_contract as r
import recovery_history as h
import recovery_state as s
c=r.c

def templates(out):
    # Component payload templates are shared business contracts. Their ID3
    # ownership is enforced here; the closed R2 approval importer stays intact.
    import final_grants
    return final_grants.templates(out,'remaining')

def verified(out,name,argv):
    if s.run(out,name,argv,timeout=180)!=0:raise RuntimeError('current native consumer refused; preserve failure')
    return s.pin(out/(name+'-command.json'))

def facts(out):
    verified(out,'native-recovery-facts',[out/'minimaxverify','-recovery-source'])
    value=json.loads((out/'native-recovery-facts.log').read_text());prior=h.facts();recovery=value.get('recovery',{})
    expected=dict(secondCohortSealSha256=prior['secondCohortSealSHA256'],stopIndexSha256=prior['stopIndexSHA256'],secondCalls=13,secondAttempts=12,secondLedgerRows=25,retainedActualCalls=303,retainedActualAttempts=302,retainedClaimedCalls=456,retainedClaimedAttempts=492,retainedReservedMicrousd=10700000)
    if (value.get('renewalId')!=r.ROUND or value.get('sourceCommit')!=c.git('rev-parse','HEAD') or len(value.get('closedHistory',{}))!=15 or any(value.get(k)!=v for k,v in dict(originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36).items()) or any(recovery.get(k)!=v for k,v in expected.items())):raise RuntimeError('exact native3 closed1+failed2 history/source required')
    for name,key in [('grant.json','secondGrantSha256'),('human-approval.json','secondHumanApprovalSha256'),('source-review.json','secondSourceReviewSha256'),('chat.json','secondChatClaimSha256'),('formal.jsonl','secondLedgerSha256')]:
        if recovery[key]!=s.pin(c.common_root()/'.evidence/minimax-renewals'/r.old.ROUND/name):raise RuntimeError('native failed2 authority pin changed')
    return value

def require_component(out,job,auth,scope):
    name=job['name']
    if 'combo' not in job:
        if name!='protocol114' and not name.startswith('visual-'):return
        leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json'
        p=out/leaf;value=s.doc(p);template=s.doc(out/(leaf+'.template.json'));mutable={'accountAlias','regionAlias'}
        if (set(value)!=set(template) or {k:v for k,v in value.items() if k not in mutable}!={k:v for k,v in template.items() if k not in mutable} or
            value.get('accountAlias')!=scope['accountAlias'] or value.get('regionAlias')!=scope['regionAlias'] or value.get('dedicatedCredential') is not True or
            value.get('maxCalls')!=12 or value.get('maxAttempts')!=12 or s.pin(p)!=auth['componentAuthorizationSHA256'].get(name)):
            raise RuntimeError('new own bounded protocol/effect authorization required')
        return
    p=Path(job['environment']['BARNESS_REGIONAL_GRANT']);template=s.doc(out/'grants'/(name+'.template.json'));value=s.doc(p);mutable={'id','accountAlias','regionAlias','expiresAt'}
    if set(value)!=set(template) or {k:v for k,v in value.items() if k not in mutable}!={k:v for k,v in template.items() if k not in mutable} or value.get('mode')!='live' or value.get('accountAlias')!=scope['accountAlias'] or value.get('regionAlias')!=scope['regionAlias'] or not value.get('id','').startswith('ps-final-3-'):raise RuntimeError('new own exact3 regional grant required')
    expires=dt.datetime.fromisoformat(value['expiresAt'].replace('Z','+00:00'));now=dt.datetime.now(dt.timezone.utc)
    if expires<=now or expires-now>dt.timedelta(hours=24) or s.pin(p)!=auth['componentAuthorizationSHA256'].get(name):raise RuntimeError('reviewed current expiring regional grant pin required')

def preflight(out,job,record):
    if not job['name'].startswith('mm-'):return record
    name=job['name']+'-native-preflight';verified(out,name,[out/'minimaxverify','-recovery-preflight='+job['route']]);value=json.loads((out/(name+'.log')).read_text())
    if value.get('route')!=job['route'] or value.get('facts',{}).get('renewalId')!=r.ROUND:raise RuntimeError('native own recovery3 preflight identity required')
    return record|dict(nativePreflightReceiptSha256=s.pin(out/(name+'-command.json')),renewalId=r.ROUND,renewalGrantSha256=value['grantSha256'])

def create(out):
    import recovery_execute as e
    prepared=s.guard(out)
    for route in r.ROUTES:_,auth,_,_=s.approved(out,r.jobs(out)['mm-global-formal-m3-'+route])
    e.require_checks(out);value=facts(out);c.dump(out/'native-recovery-facts.json',value)
    human=dict(decision='APPROVED',renewalId=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,campaignMaximumMicrousd=30000000,firstCohortSealSHA256=value['firstCohortSealSha256'],secondCohortSealSHA256=value['recovery']['secondCohortSealSha256'],historySHA256=value['closedHistorySha256'],recoveryHistorySHA256=value['recoveryHistorySha256'],planSHA256=prepared['planSHA256'],requestReference=auth['requestReference'],requestText=auth['requestText'],provider='minimax',model='MiniMax-M3',purpose='standard',routes=list(r.ROUTES),renewalMaxCalls=60,renewalMaxAttempts=72,actualHumanApprovalSHA256=auth['actualHumanApprovalSHA256'])
    review=s.doc(out/'prerelease-source-review.json')|{k:human[k] for k in ('renewalId','historySHA256','firstCohortSealSHA256','secondCohortSealSHA256','recoveryHistorySHA256')}
    with e.locked() as base:
        for leaf in ('human-approval.json','source-review.json','grant.json'):
            if (base/leaf).exists() or (base/leaf).is_symlink():raise RuntimeError('native3 authority already issued; never replace')
        c.dump(base/'human-approval.json',human,exclusive=True);c.dump(base/'source-review.json',review,exclusive=True)
        grant={k:v for k,v in value.items() if k not in ('closedHistory','recoveryHistorySha256')}
        grant.update(schema=4,decision='APPROVED',maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000,humanApprovalSha256=s.pin(base/'human-approval.json'),sourceReviewSha256=s.pin(base/'source-review.json'),sourceManifestSha256=prepared['sourceManifestSHA256'],binariesManifestSha256=prepared['binariesManifestSHA256'],planSha256=prepared['planSHA256'])
        c.dump(base/'grant.json',grant,exclusive=True);e.sync_dir(base)
    for route in r.ROUTES:verified(out,'native-grant-preflight-'+route,[out/'minimaxverify','-recovery-preflight='+route])
    return dict(credentialsRead=False,renewalId=r.ROUND,grantSHA256=s.pin(base/'grant.json'),maxCalls=60,maxAttempts=72)

def consume(out,job,record):
    import final_execute as old
    import recovery_execute as e
    report_path=old.report_path(out,job);report=s.doc(report_path);base=report_path.parent;name=job['name']
    if name not in r.EFFECTS and report.get('gitCommit',report.get('sourceCommit'))!=record['sourceCommit']:raise RuntimeError('own current native measured commit required')
    if 'combo' in job:argv=[out/'regionalverify','-formal',base]
    elif name.startswith('mm-'):
        grant=c.common_root()/'.evidence/minimax-renewals'/r.ROUND/'grant.json'
        if report.get('formalRenewalId')!=r.ROUND or report.get('formalRenewalGrantSha256')!=s.pin(grant):raise RuntimeError('own current3 native full report required')
        record.update(renewalId=r.ROUND,renewalGrantSha256=s.pin(grant),origin=dict(schema=1,stage='formal',identity=report['identity'],session=report['session'],reportSha256=s.pin(base/'minimax-report.json'),sealSha256=s.pin(base/'minimax-integrity.json'),ledgerSha256=s.pin(base/'minimax-ledger.json'),sourceCommit=record['sourceCommit'],sourceManifestSha256=record['sourceManifestSha256'],binariesManifestSha256=record['binariesManifestSha256'],planSha256=record['planSha256'],formalRenewalId=r.ROUND,formalRenewalGrantSha256=s.pin(grant)))
        c.dump(e.marker(name),record);argv=[out/'minimaxverify','-formal','-publish',base]
    elif name.startswith('text-'):argv=[out/'chineseeval','-route',name[5:],'-verify',base]
    elif name.startswith('visual-'):argv=[out/'visualeval','-verify',base,'-require-live']
    elif name=='protocol114':return record|dict(reportedFileSha256=s.pin(report_path),nativeConsumer='paired-jevformal-pending')
    else:
        matrix=out/(name+'-validation-matrix.json');matrix.write_bytes((c.ROOT/'ai/live/support-matrix.json').read_bytes());argv=[out/'supportmatrix','-matrix',matrix,base]
    return record|dict(reportedFileSha256=s.pin(report_path),nativeConsumerReceiptSha256=verified(out,name+'-consume',argv))


def require_consumed(out,job,value):
    import final_execute
    name=job['name'];report=final_execute.report_path(out,job)
    if (value.get('sourceCommit')!=s.doc(out/'prepared.json')['sourceCommit'] or value.get('exitCode')!=0 or value.get('state')!='COMPLETED' or value.get('budget',{}).get('status')!='PASS' or value.get('reportedFileSha256')!=s.pin(report) or value.get('commandReceiptSha256')!=s.pin(out/(name+'-command.json'))):raise RuntimeError('prior current3 complete native process facts required')
    if name=='protocol114':
        if value.get('nativeConsumer')!='paired-jevformal-pending':raise RuntimeError('explicit paired Jev protocol consumer required')
        return
    p=out/(name+'-consume-command.json');command=s.doc(p)
    base=report.parent
    if 'combo' in job:argv=[out/'regionalverify','-formal',base]
    elif name.startswith('mm-'):argv=[out/'minimaxverify','-formal','-publish',base]
    elif name.startswith('text-'):argv=[out/'chineseeval','-route',name[5:],'-verify',base]
    elif name.startswith('visual-'):argv=[out/'visualeval','-verify',base,'-require-live']
    else:argv=[out/'supportmatrix','-matrix',out/(name+'-validation-matrix.json'),base]
    if value.get('nativeConsumerReceiptSha256')!=s.pin(p) or command.get('state')!='COMPLETED' or command.get('exitCode')!=0 or command.get('credentialsRead') is not False or command.get('argv')!=c.redact(list(map(str,argv))):raise RuntimeError('prior actual own successful native consumer required')
