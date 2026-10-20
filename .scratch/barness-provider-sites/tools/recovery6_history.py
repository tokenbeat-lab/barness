"""Immutable actual fifth history; current native qualification remains separate."""
import hashlib
import json
import os
from pathlib import Path
import recovery5_history as previous
import recovery6_contract as r
from recovery_flow_files import claim_names
from recovery_flow_source import SourceFreeze
c=r.c;f=previous.f
relative=previous.relative;pin=previous.pin;expand=previous.expand
strict=previous.strict;exact=previous.exact;digest=previous.digest
CLOSED=r.BASE/'closed-cohort-5/seal.json';STOP=r.BASE/'recovery5-stopped/index.json'
STOP_SHA256='5e8014e7405cfa86da21637b96e47cb294c4ff8e766e79440fe88ff56eef4f73'
STOP_FILE_COUNT=31152
HUMAN=r.BASE/'recovery6-authorization.json';OLD='provider-sites-final-acceptance-5'
NAMES=previous.r.EXECUTION_ORDER[:8];AUTHORITY=previous.AUTHORITY

def doc(path):
    name=relative(path) if Path(path).is_absolute() else str(path)
    # The actual fifth index holds31,152 paths and exceeds the older2MiB bound.
    # Only this fixed pinned index gets16MiB; ordinary JSON keeps2MiB.
    limit=16<<20 if name==str(STOP) else 2<<20
    with f.regular(name,limit) as (stream,size):
        raw=stream.read(size+1)
        if len(raw)!=size:raise RuntimeError('closed5 JSON changed during bounded read')
    value=strict(raw)
    if not isinstance(value,dict):raise RuntimeError('closed5 JSON object required')
    return value

def ledger(native):
    # A separately fixed cohort: never reinterpret its consumed ID5 nonce as6.
    rows=[strict(v) for v in SourceFreeze(c).read(native/'formal.jsonl',4<<20).splitlines()]
    tail='';counts={route:[0,0] for route in r.ROUTES};sessions={}
    keys=('renewalId','renewalGrantSha256','session','sequence','stage','identity','case','kind','output','input','at','previous','hash')
    for i,row in enumerate(rows):
        if set(row)!=set(keys):raise RuntimeError('closed5 strict native receipt fields required')
        raw={key:row[key] for key in keys};raw['hash']='';route=row['identity'].get('route')
        if row['sequence']!=i+1 or row['previous']!=tail or hashlib.sha256(json.dumps(raw,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()!=row['hash']:raise RuntimeError('closed5 original native chain differs')
        if route not in counts or row['renewalId']!=OLD or row['renewalGrantSha256']!=pin(native/'grant.json') or row['stage']!='formal' or row['kind'] not in ('reserve','attempt'):raise RuntimeError('closed5 native authority differs')
        counts[route][row['kind']=='attempt']+=1;sessions.setdefault(route,set()).add(row['session']);tail=row['hash']
    if len(rows)!=75 or any(v!=[13,12] for v in counts.values()) or any(len(v)!=1 for v in sessions.values()):raise RuntimeError('closed5 exact39/36 and75rows required')
    for route in r.ROUTES:
        value=doc(native/(route+'.json'));exact(value,dict(schema=2,renewalId=OLD,grantSha256=pin(native/'grant.json'),session=next(iter(sessions[route]))),'closed5 consumed nonce differs')
        if len(value.get('processNonce',''))!=64 or value['identity']!=next(v['identity'] for v in rows if v['identity']['route']==route):raise RuntimeError('closed5 nonce identity differs')
    return tail

def facts():
    prior=previous.facts();seal=doc(CLOSED);stop=doc(STOP)
    if pin(STOP)!=STOP_SHA256:raise RuntimeError('fixed actual closed5 stop index changed')
    expected=dict(schema=1,kind='CLOSED_FIFTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndexSHA256=pin(STOP),actualCalls=702,actualHTTPAttempts=692,claimedCalls=1160,claimedHTTPAttempts=1232,wholeReservedMicrousd=32050000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False)
    if set(seal)!=set(expected)|{'stopIndex'} or expand(seal['stopIndex'])!=f.owned(str(STOP)):raise RuntimeError('strict closed5 seal fields/path required')
    exact(seal,expected,'closed5 accounting/no continuation differs')
    exact(stop,dict(schema=1,kind='ACTUAL_FIXED5_STOP_AFTER_MINIMAX_CN_RESPONSES_NATIVE_HISTORY_FAILURE',state='STOPPED_NO_FURTHER_PAID_LAUNCH',newPaidProcesses=8,passedProcesses=7,failedProcesses=1,newNativeRecordedCalls=111,newNativeRecorderHTTP=108,campaignNativeRecordedCalls=702,campaignNativeRecorderHTTP=692,campaignClaimedCalls=1160,campaignClaimedHTTP=1232,wholeReservedMicrousd=32050000,refundMicrousd=0,unknownParentProcesses=5,failedParentRemainsUNKNOWN=True,automaticReruns=False,paidContinuationAuthorized=False),'closed5 honest111/108 and five UNKNOWN required')
    from recovery6_files import verify
    verify(stop)
    files=stop['files'];total=0
    if not isinstance(files,dict) or len(files)!=STOP_FILE_COUNT or stop.get('fileCount')!=len(files):raise RuntimeError('closed5 exact original pin set required')
    for name,want in files.items():
        got,size=f.file_hash(name,(1<<30)-total);total+=size
        if got!=want:raise RuntimeError('closed5 original changed: '+name)
    if stop.get('fileBytes')!=total:raise RuntimeError('closed5 original total bytes differ')
    for name,fixture in stop.get('nonRegularRejectionFixtures',{}).items():
        entry=Path(name);path=f.owned(str(entry.parent))/entry.name
        if fixture.get('kind')!='symlink' or fixture.get('followed') is not False or not path.is_symlink() or os.readlink(path)!=fixture.get('target') or path.lstat().st_mode!=fixture.get('mode'):raise RuntimeError('closed5 recorded refusal fixture changed')
    if claim_names(c,OLD)!=set(NAMES) or len(stop['processes'])!=8 or {v['name'] for v in stop['processes']}!=set(NAMES) or set(stop['notRunProcesses'])!=set(previous.r.EXECUTION_ORDER[8:]):raise RuntimeError('closed5 exact8 parents and11 unstarted required')
    native=f.owned('.evidence/minimax-renewals/'+OLD)
    if {p.relative_to(native).as_posix() for p in f.native_files(native)}!=set(AUTHORITY):raise RuntimeError('closed5 exact9 authority files required')
    authority={leaf:pin(native/leaf) for leaf in AUTHORITY}
    if any(files.get(relative(native/leaf))!=sha for leaf,sha in authority.items()):raise RuntimeError('closed5 authority outside original pins')
    grant=doc(native/'grant.json');human=doc(native/'human-approval.json');review=doc(native/'source-review.json');actual=r.BASE/'recovery5-authorization.json'
    exact(grant,dict(schema=6,decision='APPROVED',renewalId=OLD,sourceCommit=stop['sourceCommit'],humanApprovalSha256=pin(native/'human-approval.json'),sourceReviewSha256=pin(native/'source-review.json'),maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000),'closed5 original schema6 native authority required')
    exact(human,dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=378,maxAttempts=414,campaignMaxCalls=1358,campaignMaxAttempts=1454,newReservedMicrousd=6900000,retainedReservedMicrousd=28800000,campaignMaximumMicrousd=36000000,actualHumanApprovalSHA256=pin(actual)),'closed5 original36USD approval cannot become6')
    actual_value=doc(actual);exact(actual_value,dict(decision='APPROVED',round=OLD),'closed5 actual direct reply required')
    if stop['actualHumanApprovalSHA256']!=pin(actual) or files.get(str(actual))!=pin(actual):raise RuntimeError('closed5 actual human original changed')
    for field,key in [('sourceCommit','sourceCommit'),('sourceManifestSHA256','sourceManifestSha256'),('binariesManifestSHA256','binariesManifestSha256'),('planSHA256','planSha256')]:
        if review.get(field)!=grant[key]:raise RuntimeError('closed5 original review/source pins differ')
    archive=doc(r.BASE/'recovery5/source-at-start.json')
    if pin(r.BASE/'recovery5/source-at-start.json')!=grant['sourceManifestSha256'] or archive['sourceCommit']!=stop['sourceCommit']:raise RuntimeError('closed5 original source archive differs')
    parents={}
    for row in stop['processes']:
        name=row['name'];failed=name=='formal-minimax-cn-responses';p=f.owned('.evidence/minimax-launches/'+OLD+'-'+name+'.json');value=doc(p);mm=name.startswith('mm-');calls=13 if mm else 18 if name.startswith('formal-zai-') else 9;attempts=12 if mm else calls
        budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        exact(value,dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=stop['sourceCommit'],reservedMicrousd=previous.r.RESERVES[name],budget=budget),'closed5 real failure/UNKNOWN or original PASS differs')
        exact(row,dict(exitCode=value['exitCode'],parentBudget=budget,nativeRecordedCalls=calls,nativeRecorderHTTP=attempts,reservedMicrousd=value['reservedMicrousd']),'closed5 reported counters differ')
        report_path=expand(row['raw'])/('minimax-report.json' if mm else 'regional-report.json');report=doc(report_path)
        for target in (p,report_path):
            if files.get(relative(target))!=pin(target):raise RuntimeError('closed5 parent/raw pin differs')
        if report.get('gitCommit')!=stop['sourceCommit']:raise RuntimeError('closed5 raw historical commit differs')
        if mm and digest(report['executionSource'])!=grant['executionSourceSha256']:raise RuntimeError('closed5 native original source differs')
        command=f.owned(str(r.BASE/'recovery5'/(name+'-command.json')));cmd=doc(command)
        if files.get(relative(command))!=pin(command) or value['commandReceiptSha256']!=pin(command) or cmd.get('exitCode')!=value['exitCode'] or cmd.get('launcherAudit')!=dict(passed=True,findings=[]):raise RuntimeError('closed5 original command differs')
        if not failed:
            consumer=f.owned(str(r.BASE/'recovery5'/(name+'-consume-command.json')));con=doc(consumer)
            if value.get('reportedFileSha256')!=pin(report_path) or value.get('nativeConsumerReceiptSha256')!=pin(consumer) or files.get(relative(consumer))!=pin(consumer) or con.get('exitCode')!=0 or con.get('state')!='COMPLETED' or con.get('credentialsRead') is not False:raise RuntimeError('closed5 original native consumer differs')
        else:
            if len(report.get('scenarios',[]))!=9 or sum(v['outcome']=='PASS' for v in report['scenarios'])!=8:raise RuntimeError('closed5 honest8PASS/1FAIL native history required')
        parents[name]=pin(p)
    usd=doc(r.BASE/'recovery5/formal-minimax-cn-responses-usd.json');exact(usd,dict(reservedMicrousd=54238,attempts=9),'closed5 failed guard reservation is54238/9, not zero')
    fifth=dict(actualHumanApprovalSha256=pin(actual),cohortSealSha256=pin(CLOSED),failedNativeRecordedAttempts=9,failedNativeRecordedCalls=9,failedUSDReservationMicrousd=54238,failedUSDTransportReservations=9,nativeAttempts=36,nativeAuthoritySha256=authority,nativeCalls=39,nativeLedgerRows=75,nativeLedgerTailSha256=ledger(native),nativeSourceCommit=stop['sourceCommit'],newActualAttempts=108,newActualCalls=111,newClaimedAttempts=192,newClaimedCalls=180,newReservedMicrousd=3250000,parentClaimsSha256=parents,retainedActualAttempts=692,retainedActualCalls=702,retainedClaimedAttempts=1232,retainedClaimedCalls=1160,retainedReservedMicrousd=32050000,retainedUnknownParentProcesses=5,stopIndexSha256=pin(STOP))
    return dict(fifthCohortSealSHA256=pin(CLOSED),fifthCohortHistorySHA256=digest(fifth),fifthCohort=fifth,previousCohort=prior,firstCohort=prior['firstCohort'],actualCalls=702,actualAttempts=692,claimedCalls=1160,claimedAttempts=1232,reservedMicrousd=32050000,originalProcessCount=48,authorizationSHA256=prior['authorizationSHA256'],credentialScopes=prior['credentialScopes'])

def prepare_metadata(prior):return dict(fifthCohortHistorySHA256=prior['fifthCohortHistorySHA256'])
def scope_metadata(prepared):return dict(executionOrder=list(r.EXECUTION_ORDER))

def actual_approval(prepared):
    value=doc(HUMAN);expected=dict(decision='APPROVED',round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=378,maxAttempts=414,campaignMaxCalls=1538,campaignMaxAttempts=1646,newReservedMicrousd=6900000,retainedReservedMicrousd=32050000,maxUSD=39,automaticProcessReruns=False,processReservationsMicrousd=r.RESERVES,executionOrder=list(r.EXECUTION_ORDER),originalCampaignApprovalSHA256=prepared['originalCampaignApprovalSHA256'])
    exact(value,expected,'new actual direct human fixed19/US39 approval required')
    for key in ('sourceCommit','sourceManifestSHA256','binariesManifestSHA256','planSHA256','fifthCohortSealSHA256','fifthCohortHistorySHA256','predecessorSHA256','spendingGuardSourceSHA256'):
        if value.get(key)!=prepared[key]:raise RuntimeError('new6 actual human source/binary/plan/history/guard differs')
    if not value.get('requestText') or not value.get('requestReference') or value['requestReference']==doc(r.BASE/'recovery5-authorization.json')['requestReference']:raise RuntimeError('old5 reply cannot authorize relabeled6')
    return value

def retained():
    from recovery6_retained import verify
    return verify()
