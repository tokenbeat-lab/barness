"""Fixed schema5 projection; M3 authority follows the first own ZAI Responses PASS."""
import json
import recovery4_contract as r
import recovery4_history as h
c=r.c

def facts(out):
    from recovery_flow import fixed4
    flow=fixed4();flow.n.verified(out,'native-recovery4-facts',[out/'minimaxverify','-recovery4-source'])
    value=h.strict((out/'native-recovery4-facts.log').read_bytes());prior=h.facts();old=prior['previousCohort']
    if value.get('renewalId')!=r.ROUND or value.get('sourceCommit')!=c.git('rev-parse','HEAD') or len(value.get('closedHistory',{}))!=15 or any(value.get(k)!=v for k,v in dict(originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36).items()):raise RuntimeError('exact native4 current source and original15 history required')
    if value.get('thirdCohort')!=prior['thirdCohort'] or value.get('thirdCohortHistorySha256')!=prior['thirdCohortHistorySHA256']:raise RuntimeError('native4 exact21 closed3 fields and digest required')
    recovery=value.get('recovery',{})
    for key,want in dict(secondCohortSealSha256=old['secondCohortSealSHA256'],stopIndexSha256=old['stopIndexSHA256'],secondCalls=13,secondAttempts=12,secondLedgerRows=25,retainedActualCalls=303,retainedActualAttempts=302,retainedClaimedCalls=456,retainedClaimedAttempts=492,retainedReservedMicrousd=10700000).items():
        if recovery.get(key)!=want:raise RuntimeError('unchanged nested closed2 native recovery required')
    native=c.common_root()/'.evidence/minimax-renewals'/r.old.ROUND
    for leaf,key in [('grant.json','secondGrantSha256'),('human-approval.json','secondHumanApprovalSha256'),('source-review.json','secondSourceReviewSha256'),('chat.json','secondChatClaimSha256'),('formal.jsonl','secondLedgerSha256')]:
        if recovery.get(key)!=h.pin(native/leaf):raise RuntimeError('old2 exact native authority pin differs')
    return value

def first_pass(out):
    from recovery_flow import fixed4
    flow=fixed4();name=r.EXECUTION_ORDER[0];job=r.jobs(out)[name]
    value=flow.s.doc(flow.e.marker(name));flow.n.require_consumed(out,job,value)
    if value['budget']!=dict(state='REPORTED',status='PASS',calls=18,attempts=18):raise RuntimeError('first new4 ZAI Responses full18/18 required before native issue')
    report=flow.s.doc(flow.p.own_base(out,name)/'regional-report.json')
    if len(report.get('expected',[]))!=18 or len(report.get('scenarios',[]))!=18 or any(v['outcome']!='PASS' for v in report['scenarios']):raise RuntimeError('first new4 ZAI Responses18 genuine native PASS required')
    return value

def before_claim(out,job):
    native=h.f.owned('.evidence/minimax-renewals/'+r.ROUND+'/grant.json')
    if job['name']==r.EXECUTION_ORDER[0]:
        if native.exists() or native.is_symlink():raise RuntimeError('native4 must remain unissued before the first ZAI Responses')
        if native.parent.exists() and {p.relative_to(native.parent).as_posix() for p in h.f.native_files(native.parent)}-{'operator-authority.json','launch.lock'}:raise RuntimeError('unexpected native4 intent/nonce before first ZAI Responses')
    elif job['name'].startswith('mm-'):
        first_pass(out);grant=h.doc(native)
        if grant.get('schema')!=5 or grant.get('renewalId')!=r.ROUND:raise RuntimeError('new own schema5 grant required before M3 claim')

def create(out):
    from recovery_flow import fixed4
    flow=fixed4();prepared=flow.s.guard(out)
    for route in r.ROUTES:_,auth,_,_=flow.s.approved(out,r.jobs(out)['mm-global-formal-m3-'+route])
    flow.e.require_checks(out);first_pass(out)
    summary=flow.e.summarize()
    if [v['name'] for v in summary['routes']]!=[r.EXECUTION_ORDER[0]]:raise RuntimeError('issue native4 once immediately after first Responses; no later claim may preexist')
    value=facts(out);c.dump(out/'native-recovery4-facts.json',value)
    human=dict(decision='APPROVED',renewalId=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=1070,campaignMaxAttempts=1130,newReservedMicrousd=17300000,retainedReservedMicrousd=12600000,campaignMaximumMicrousd=30000000,firstCohortSealSHA256=value['firstCohortSealSha256'],secondCohortSealSHA256=value['recovery']['secondCohortSealSha256'],thirdCohortSealSHA256=value['thirdCohort']['cohortSealSha256'],historySHA256=value['closedHistorySha256'],recoveryHistorySHA256=value['recoveryHistorySha256'],thirdCohortHistorySHA256=value['thirdCohortHistorySha256'],planSHA256=prepared['planSHA256'],requestReference=auth['requestReference'],requestText=auth['requestText'],actualHumanApprovalSHA256=auth['actualHumanApprovalSHA256'],provider='minimax',model='MiniMax-M3',purpose='standard',routes=list(r.ROUTES),renewalMaxCalls=60,renewalMaxAttempts=72)
    review=flow.s.doc(out/'prerelease-source-review.json')|{k:human[k] for k in ('renewalId','historySHA256','firstCohortSealSHA256','secondCohortSealSHA256','thirdCohortSealSHA256','recoveryHistorySHA256','thirdCohortHistorySHA256')}
    with flow.e.locked() as base:
        if {p.relative_to(base).as_posix() for p in h.f.native_files(base)}!={'operator-authority.json','launch.lock'}:raise RuntimeError('unexpected native4 preexisting authority or nonce')
        for leaf in ('human-approval.json','source-review.json','grant.json'):
            if (base/leaf).exists() or (base/leaf).is_symlink():raise RuntimeError('native4 authority already issued; never replace')
        c.dump(base/'human-approval.json',human,exclusive=True);c.dump(base/'source-review.json',review,exclusive=True)
        grant={k:v for k,v in value.items() if k not in ('closedHistory','recoveryHistorySha256','thirdCohortHistorySha256')}
        grant.update(schema=5,decision='APPROVED',maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000,humanApprovalSha256=flow.s.pin(base/'human-approval.json'),sourceReviewSha256=flow.s.pin(base/'source-review.json'),sourceManifestSha256=prepared['sourceManifestSHA256'],binariesManifestSha256=prepared['binariesManifestSHA256'],planSha256=prepared['planSHA256'])
        c.dump(base/'grant.json',grant,exclusive=True);flow.e.sync_dir(base)
    for route in r.ROUTES:flow.n.verified(out,'native-grant-preflight-'+route,[out/'minimaxverify','-recovery4-preflight='+route])
    return dict(credentialsRead=False,renewalId=r.ROUND,grantSHA256=flow.s.pin(base/'grant.json'),maxCalls=60,maxAttempts=72)
