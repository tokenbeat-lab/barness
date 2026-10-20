"""Closed fourth originals and the narrowly justified eight generic retained roles."""
import hashlib
import json
import subprocess
from pathlib import Path
import recovery4_history as previous
import recovery4_contract as old
import recovery5_contract as r
from recovery_flow_files import claim_names
from recovery_flow_source import SourceFreeze
c=r.c;f=previous.f
relative=previous.relative;pin=previous.pin;expand=previous.expand
doc=previous.doc;strict=previous.strict;exact=previous.exact;digest=previous.digest
CLOSED=r.BASE/'closed-cohort-4/seal.json';STOP=r.BASE/'recovery4-stopped/index.json'
HUMAN=r.BASE/'recovery5-authorization.json';OLD=old.ROUND
NAMES=old.EXECUTION_ORDER[:18]
AUTHORITY=previous.AUTHORITY
COMPAT=r.BASE/'decisions-budget-diagnosis-dev/compatibility-eight.json'


def ledger(native,grant):
    rows=[strict(v) for v in SourceFreeze(c).read(native/'formal.jsonl',4<<20).splitlines()]
    tail='';counts={route:[0,0] for route in r.ROUTES};sessions={}
    keys=('renewalId','renewalGrantSha256','session','sequence','stage','identity','case','kind','output','input','at','previous','hash')
    for i,row in enumerate(rows):
        if set(row)!=set(keys):raise RuntimeError('closed4 strict native receipt fields required')
        raw={key:row[key] for key in keys};raw['hash']=''
        route=row['identity'].get('route')
        if row['sequence']!=i+1 or row['previous']!=tail or hashlib.sha256(json.dumps(raw,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()!=row['hash']:raise RuntimeError('closed4 original native chain differs')
        if route not in counts or row['renewalId']!=OLD or row['renewalGrantSha256']!=pin(native/'grant.json') or row['stage']!='formal' or row['kind'] not in ('reserve','attempt'):raise RuntimeError('closed4 native identity differs')
        counts[route][row['kind']=='attempt']+=1;sessions.setdefault(route,set()).add(row['session']);tail=row['hash']
    if len(rows)!=75 or any(v!=[13,12] for v in counts.values()) or any(len(v)!=1 for v in sessions.values()):raise RuntimeError('closed4 exact39/36 native receipts required')
    for route in r.ROUTES:
        value=doc(native/(route+'.json'))
        exact(value,dict(schema=2,renewalId=OLD,grantSha256=pin(native/'grant.json'),session=next(iter(sessions[route]))),'closed4 consumed nonce differs')
        if len(value.get('processNonce',''))!=64 or value['identity']!=next(v['identity'] for v in rows if v['identity']['route']==route):raise RuntimeError('closed4 nonce identity differs')
    return tail


def facts():
    prior=previous.facts();seal=doc(CLOSED);stop=doc(STOP)
    expected=dict(schema=1,kind='CLOSED_FOURTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndexSHA256=pin(STOP),actualCalls=591,actualHTTPAttempts=584,claimedCalls=980,claimedHTTPAttempts=1040,wholeReservedMicrousd=28800000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False)
    if set(seal)!=set(expected)|{'stopIndex'} or expand(seal['stopIndex'])!=f.owned(str(STOP)):raise RuntimeError('strict closed4 seal path/fields required')
    exact(seal,expected,'closed4 no refund/no further launch counters differ')
    exact(stop,dict(schema=1,kind='ACTUAL_RECOVERY4_STOP_AFTER_OPENAI_DECISIONS_PRESEND_BUDGET_FAILURE',successfulProcesses=17,failedProcesses=1,newActualCalls=213,newActualHTTP=210,campaignActualCalls=591,campaignActualHTTP=584,campaignClaimedCalls=980,campaignClaimedHTTP=1040,newWholeReservedMicrousd=16200000,wholeReservedMicrousd=28800000,refundMicrousd=0,unknownParentProcesses=4,failedNativeRecordedCalls=9,failedNativeRecordedHTTP=9,failedUSDReservationMicrousd=0,failedUSDTransportReservations=0,automaticReruns=False),'closed4 actual591/584 and four UNKNOWN required')
    files=stop['files'];total=0
    if not isinstance(files,dict) or len(files)!=2032:raise RuntimeError('exact2032 closed4 original pins required')
    for name,want in files.items():
        got,size=f.file_hash(name,(512<<20)-total);total+=size
        if got!=want:raise RuntimeError('closed4 original changed: '+name)
    if claim_names(c,OLD)!=set(NAMES) or {v['name'] for v in stop['processes']}!=set(NAMES) or len(stop['processes'])!=18 or set(stop['unstartedProcesses'])!=set(old.EXECUTION_ORDER[18:]):raise RuntimeError('closed4 exact18 parents and six unstarted processes required')
    native=f.owned('.evidence/minimax-renewals/'+OLD)
    if {p.relative_to(native).as_posix() for p in f.native_files(native)}!=set(AUTHORITY):raise RuntimeError('closed4 exact9 native authority files required')
    authority={leaf:pin(native/leaf) for leaf in AUTHORITY}
    if any(files.get(relative(native/leaf))!=sha for leaf,sha in authority.items()):raise RuntimeError('closed4 native authority not retained')
    grant=doc(native/'grant.json');human=doc(native/'human-approval.json');review=doc(native/'source-review.json');actual=r.BASE/'recovery4-authorization.json'
    exact(grant,dict(schema=5,decision='APPROVED',renewalId=OLD,sourceCommit=stop['sourceCommit'],humanApprovalSha256=pin(native/'human-approval.json'),sourceReviewSha256=pin(native/'source-review.json'),maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000),'closed4 original schema5 authority required')
    exact(human,dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=1070,campaignMaxAttempts=1130,newReservedMicrousd=17300000,retainedReservedMicrousd=12600000,campaignMaximumMicrousd=30000000,actualHumanApprovalSHA256=pin(actual)),'old4 approval cannot be relabeled5')
    exact(doc(actual),dict(decision='APPROVED',round=OLD),'closed4 actual direct reply required')
    if stop.get('actualHumanApprovalSHA256')!=pin(actual) or files.get(str(actual))!=pin(actual):raise RuntimeError('closed4 actual human original pin differs')
    for field,key in [('sourceCommit','sourceCommit'),('sourceManifestSHA256','sourceManifestSha256'),('binariesManifestSHA256','binariesManifestSha256'),('planSHA256','planSha256')]:
        if review.get(field)!=grant[key]:raise RuntimeError('closed4 historical source/review pins differ')
    archive=doc(r.BASE/'recovery4/source-at-start.json')
    if pin(r.BASE/'recovery4/source-at-start.json')!=grant['sourceManifestSha256'] or archive['sourceCommit']!=stop['sourceCommit']:raise RuntimeError('closed4 original source archive pin differs')
    parents={}
    for row in stop['processes']:
        name=row['name'];p=f.owned('.evidence/minimax-launches/'+OLD+'-'+name+'.json');value=doc(p);failed=name=='openai-decisions';report=expand(row['nativeBundle'])/('minimax-report.json' if name.startswith('mm-') else 'regional-report.json' if name.startswith('formal-') else 'live-report.json')
        if files.get(relative(p))!=pin(p) or files.get(relative(report))!=pin(report) or row['nativeReportSHA256']!=pin(report):raise RuntimeError('closed4 parent/native raw pin differs')
        exact(value,dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=stop['sourceCommit'],reservedMicrousd=old.RESERVES[name],budget=row['parentBudget']),'closed4 original full parent state differs')
        if value['budget']!=(dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=row['actualCalls'],attempts=row['actualHTTP'])):raise RuntimeError('closed4 failed UNKNOWN must remain occupied')
        command=f.owned(str(r.BASE/'recovery4'/(name+'-command.json')));cmd=doc(command)
        if files.get(relative(command))!=pin(command) or value['commandReceiptSha256']!=pin(command) or cmd.get('exitCode')!=(1 if failed else 0) or cmd.get('launcherAudit')!=dict(passed=True,findings=[]):raise RuntimeError('closed4 original launcher command required')
        if not failed:
            consumer=f.owned(str(r.BASE/'recovery4'/(name+'-consume-command.json')));con=doc(consumer)
            if files.get(relative(consumer))!=pin(consumer) or value.get('reportedFileSha256')!=pin(report) or value.get('nativeConsumerReceiptSha256')!=pin(consumer) or con.get('exitCode')!=0 or con.get('state')!='COMPLETED' or con.get('credentialsRead') is not False:raise RuntimeError('closed4 original successful native consumer required')
        parents[name]=pin(p)
    fourth=dict(actualHumanApprovalSha256=pin(actual),cohortSealSha256=pin(CLOSED),failedNativeRecordedAttempts=9,failedNativeRecordedCalls=9,failedUSDReservationMicrousd=0,failedUSDTransportReservations=0,nativeAttempts=36,nativeAuthoritySha256=authority,nativeCalls=39,nativeLedgerRows=75,nativeLedgerTailSha256=ledger(native,grant),nativeSourceCommit=stop['sourceCommit'],newActualAttempts=210,newActualCalls=213,newClaimedAttempts=428,newClaimedCalls=416,newReservedMicrousd=16200000,parentClaimsSha256=parents,retainedActualAttempts=584,retainedActualCalls=591,retainedClaimedAttempts=1040,retainedClaimedCalls=980,retainedReservedMicrousd=28800000,retainedUnknownParentProcesses=4,stopIndexSha256=pin(STOP))
    return dict(fourthCohortSealSHA256=pin(CLOSED),fourthCohortHistorySHA256=digest(fourth),fourthCohort=fourth,previousCohort=prior,firstCohort=prior['firstCohort'],actualCalls=591,actualAttempts=584,claimedCalls=980,claimedAttempts=1040,reservedMicrousd=28800000,originalProcessCount=40,authorizationSHA256=prior['authorizationSHA256'],credentialScopes=prior['credentialScopes'])


def prepare_metadata(prior):return dict(fourthCohortHistorySHA256=prior['fourthCohortHistorySHA256'])
def scope_metadata(prepared):return dict(executionOrder=list(r.EXECUTION_ORDER))

def actual_approval(prepared):
    value=doc(HUMAN)
    expected=dict(decision='APPROVED',round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=378,maxAttempts=414,campaignMaxCalls=1358,campaignMaxAttempts=1454,newReservedMicrousd=6900000,retainedReservedMicrousd=28800000,maxUSD=36,automaticProcessReruns=False,processReservationsMicrousd=r.RESERVES,executionOrder=list(r.EXECUTION_ORDER),originalCampaignApprovalSHA256=prepared['originalCampaignApprovalSHA256'])
    exact(value,expected,'new actual direct human fixed19/US36 approval required')
    for key in ('sourceCommit','sourceManifestSHA256','binariesManifestSHA256','planSHA256','fourthCohortSealSHA256','fourthCohortHistorySHA256','predecessorSHA256','spendingGuardSourceSHA256'):
        if value.get(key)!=prepared[key]:raise RuntimeError('new5 human must bind exact source/binary/plan/history/guard')
    if not value.get('requestText') or not value.get('requestReference') or value['requestReference']==doc(r.BASE/'recovery4-authorization.json')['requestReference']:raise RuntimeError('old4 reply cannot authorize relabeled5')
    return value


def retained():
    from recovery5_retained import verify
    return verify()
