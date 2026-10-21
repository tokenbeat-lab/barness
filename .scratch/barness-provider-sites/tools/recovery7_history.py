"""Immutable actual sixth history, including one interrupted unreported M3 child."""
import hashlib
import json
import os
from pathlib import Path
import recovery6_history as previous
import recovery7_contract as r
from recovery_flow_files import claim_names
from recovery_flow_source import SourceFreeze
c=r.c;f=previous.f
relative=previous.relative;pin=previous.pin;expand=previous.expand
strict=previous.strict;exact=previous.exact;digest=previous.digest
CLOSED=r.BASE/'closed-cohort-6/seal.json';STOP=r.BASE/'recovery6-stopped/index.json'
STOP_SHA256='fcb781794736eb8c8cc83582d757ed16e0db5091661690cab51f517bcdbad81e'
STOP_FILE_COUNT=30238
HUMAN=r.BASE/'recovery7-authorization.json';OLD='provider-sites-final-acceptance-6'
NAMES=previous.r.EXECUTION_ORDER[:2]
AUTHORITY=('operator-authority.json','launch.lock','grant.json','human-approval.json','source-review.json','chat.json','formal.jsonl')

def doc(path):
    name=relative(path) if Path(path).is_absolute() else str(path)
    # Only the actual sixth pinned30,238-path index needs the16MiB JSON bound.
    if name!=str(STOP):return previous.doc(path)
    with f.regular(name,16<<20) as (stream,size):
        raw=stream.read(size+1)
        if len(raw)!=size:raise RuntimeError('closed6 JSON changed during bounded read')
    value=strict(raw)
    if not isinstance(value,dict):raise RuntimeError('closed6 JSON object required')
    return value

def ledger(native):
    rows=[strict(v) for v in SourceFreeze(c).read(native/'formal.jsonl',4<<20).splitlines()]
    tail='';counts=[0,0];sessions=set();identity=None
    keys=('renewalId','renewalGrantSha256','session','sequence','stage','identity','case','kind','output','input','at','previous','hash')
    for i,row in enumerate(rows):
        if set(row)!=set(keys):raise RuntimeError('closed6 strict native receipt fields required')
        raw={key:row[key] for key in keys};raw['hash']=''
        if row['sequence']!=i+1 or row['previous']!=tail or hashlib.sha256(json.dumps(raw,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()!=row['hash']:raise RuntimeError('closed6 original interrupted native chain differs')
        if row['identity'].get('route')!='chat' or row['renewalId']!=OLD or row['renewalGrantSha256']!=pin(native/'grant.json') or row['stage']!='formal' or row['kind'] not in ('reserve','attempt'):raise RuntimeError('closed6 native authority differs')
        if identity is not None and identity!=row['identity']:raise RuntimeError('closed6 one exact chat identity required')
        identity=row['identity'];counts[row['kind']=='attempt']+=1;sessions.add(row['session']);tail=row['hash']
    if len(rows)!=8 or counts!=[4,4] or len(sessions)!=1:raise RuntimeError('closed6 interrupted4/4 and8rows required')
    value=doc(native/'chat.json');exact(value,dict(schema=2,renewalId=OLD,grantSha256=pin(native/'grant.json'),session=next(iter(sessions)),identity=identity),'closed6 consumed chat nonce differs')
    if len(value.get('processNonce',''))!=64:raise RuntimeError('closed6 nonce identity differs')
    return tail

def original_files(stop):
    from recovery7_files import verify
    verify(stop)
    files=stop['files'];total=0
    if not isinstance(files,dict) or len(files)!=STOP_FILE_COUNT or stop.get('fileCount')!=len(files):raise RuntimeError('closed6 exact original pin set required')
    for name,want in files.items():
        got,size=f.file_hash(name,(640<<20)-total);total+=size
        if got!=want:raise RuntimeError('closed6 original changed: '+name)
    if stop.get('fileBytes')!=total:raise RuntimeError('closed6 original total bytes differ')
    for name,fixture in stop.get('nonRegularRejectionFixtures',{}).items():
        entry=Path(name);path=f.owned(str(entry.parent))/entry.name
        if fixture.get('kind')!='symlink' or fixture.get('followed') is not False or not path.is_symlink() or os.readlink(path)!=fixture.get('target') or path.lstat().st_mode!=fixture.get('mode'):raise RuntimeError('closed6 recorded refusal fixture changed')
    return files

def parents(stop,files):
    pins={};unknown=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
    if claim_names(c,OLD)!=set(NAMES) or len(stop['processes'])!=2 or {v['name'] for v in stop['processes']}!=set(NAMES) or set(stop['notRunProcesses'])!=set(previous.r.EXECUTION_ORDER[2:]):raise RuntimeError('closed6 exact2 parents and17 unstarted required')
    for row in stop['processes']:
        name=row['name'];mm=name.startswith('mm-');p=f.owned('.evidence/minimax-launches/'+OLD+'-'+name+'.json');value=doc(p);budget=unknown if mm else dict(state='REPORTED',status='PASS',calls=18,attempts=18)
        exact(value,dict(state='COMPLETED',exitCode=-15 if mm else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=stop['sourceCommit'],reservedMicrousd=previous.r.RESERVES[name],budget=budget),'closed6 interrupted-15/UNKNOWN or original18PASS differs')
        exact(row,dict(exitCode=value['exitCode'],parentBudget=budget,reservedMicrousd=value['reservedMicrousd'],parentSHA256=pin(p),nativeConsumerReceiptSHA256=None if mm else value['nativeConsumerReceiptSha256']),'closed6 original parent facts differ')
        command=f.owned(str(r.BASE/'recovery6'/(name+'-command.json')));cmd=doc(command)
        if files.get(relative(p))!=pin(p) or files.get(relative(command))!=pin(command) or value.get('commandReceiptSha256')!=pin(command) or cmd.get('state')!='COMPLETED' or cmd.get('exitCode')!=value['exitCode'] or cmd.get('launcherAudit')!=dict(passed=True,findings=[]):raise RuntimeError('closed6 actual command/parent pin differs')
        if mm:
            if value.get('reportedFileSha256') or value.get('nativeConsumerReceiptSha256'):raise RuntimeError('closed6 interrupted child cannot gain a final report/consumer')
            base=expand(stop['m3Raw'])
            for leaf in ('minimax-report.json','minimax-integrity.json'):
                path=base/leaf
                if path.exists() or path.is_symlink():raise RuntimeError('closed6 final M3 report must remain absent')
            case=expand(stop['failedFinalCase']);result=doc(case/'result.json')
            exact(result,dict(errorCode='deadline_exceeded',errorPhase='request'),'closed6 original failed request deadline required')
            if files.get(relative(case/'result.json'))!=pin(case/'result.json'):raise RuntimeError('closed6 failed case outside original pins')
        else:
            reports=[name for name in files if name.startswith(str(r.BASE/'recovery6/formal-zai-responses')+'/') and name.endswith('/regional-report.json')]
            if len(reports)!=1:raise RuntimeError('closed6 one own original full ZAI report required')
            report_path=f.owned(reports[0]);report=doc(report_path);consumer=f.owned(str(r.BASE/'recovery6'/(name+'-consume-command.json')));con=doc(consumer)
            if report.get('gitCommit')!=stop['sourceCommit'] or len(report.get('expected',[]))!=18 or len(report.get('scenarios',[]))!=18 or any(v.get('outcome')!='PASS' for v in report['scenarios']):raise RuntimeError('closed6 original ZAI full18PASS required')
            if value.get('reportedFileSha256')!=pin(report_path) or value.get('nativeConsumerReceiptSha256')!=pin(consumer) or files.get(relative(consumer))!=pin(consumer) or con.get('exitCode')!=0 or con.get('state')!='COMPLETED' or con.get('credentialsRead') is not False:raise RuntimeError('closed6 original ZAI current native consumer differs')
        pins[name]=pin(p)
    return pins

def facts():
    prior=previous.facts();seal=doc(CLOSED);stop=doc(STOP)
    if pin(STOP)!=STOP_SHA256:raise RuntimeError('fixed actual closed6 stop index changed')
    expected=dict(schema=1,kind='CLOSED_SIXTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndexSHA256=pin(STOP),knownRecordedCalls=724,knownRecordedHTTPAllocations=714,completedReportKnownCalls=720,completedReportKnownHTTP=710,interruptedNativeLedgerCalls=4,interruptedNativeLedgerHTTPAllocations=4,interruptedParentBudget='UNKNOWN',interruptedFinalReportAbsent=True,recordedHTTPAllocationsAreNotConfirmedVendorSends=True,claimedCalls=1204,claimedHTTPAttempts=1280,wholeReservedMicrousd=32900000,unknownParentProcesses=6,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False)
    if set(seal)!=set(expected)|{'stopIndex'} or expand(seal['stopIndex'])!=f.owned(str(STOP)):raise RuntimeError('strict closed6 seal fields/path required')
    exact(seal,expected,'closed6 honest completed/interrupted accounting differs')
    exact(stop,dict(schema=1,kind='ACTUAL_FIXED6_STOP_AFTER_M3_CHAT_FINAL_REQUEST_DEADLINE_CASE_FAILURE',state='STOPPED_NO_FURTHER_PAID_LAUNCH',newPaidProcesses=2,passedProcesses=1,interruptedAfterFailedCaseProcesses=1,completedFullM3Processes=0,m3FinalReport='ABSENT_INTERRUPTED_AFTER_IRREVERSIBLE_CASE_FAILURE',failure=dict(case='full-stream-rerun-1',status='FAIL',errorCode='deadline_exceeded',errorPhase='request'),newCompletedReportCalls=18,newCompletedReportHTTP=18,campaignCompletedReportKnownCalls=720,campaignCompletedReportKnownHTTP=710,interruptedM3NativeLedgerCalls=4,interruptedM3NativeLedgerHTTPAllocations=4,interruptedM3NativeLedgerRows=8,interruptedM3USDReservationMicrousd=12884,interruptedM3USDTransportReservations=4,interruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends=True,campaignClaimedCalls=1204,campaignClaimedHTTPAttempts=1280,wholeReservedMicrousd=32900000,refundMicrousd=0,unknownParentProcesses=6,failedParentRemainsUNKNOWN=True,stopSignal='SIGTERM',childNumericExitCode=-15,publicOperatorNumericExitCode=1,automaticReruns=False,paidContinuationAuthorized=False),'closed6 interrupted original facts required')
    files=original_files(stop);parent_pins=parents(stop,files);native=f.owned('.evidence/minimax-renewals/'+OLD)
    if {p.relative_to(native).as_posix() for p in f.native_files(native)}!=set(AUTHORITY):raise RuntimeError('closed6 exact7 interrupted authority files required')
    authority={leaf:pin(native/leaf) for leaf in AUTHORITY}
    if any(files.get(relative(native/leaf))!=sha for leaf,sha in authority.items()):raise RuntimeError('closed6 authority outside original pins')
    grant=doc(native/'grant.json');human=doc(native/'human-approval.json');review=doc(native/'source-review.json');actual=r.BASE/'recovery6-authorization.json'
    exact(grant,dict(schema=7,decision='APPROVED',renewalId=OLD,sourceCommit=stop['sourceCommit'],humanApprovalSha256=pin(native/'human-approval.json'),sourceReviewSha256=pin(native/'source-review.json'),maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000),'closed6 original schema7 native authority required')
    exact(human,dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=378,maxAttempts=414,campaignMaxCalls=1538,campaignMaxAttempts=1646,newReservedMicrousd=6900000,retainedReservedMicrousd=32050000,campaignMaximumMicrousd=39000000,actualHumanApprovalSHA256=pin(actual)),'closed6 original39USD reply cannot become7')
    exact(doc(actual),dict(decision='APPROVED',round=OLD),'closed6 actual direct reply required')
    if stop['actualHumanApprovalSHA256']!=pin(actual) or files.get(str(actual))!=pin(actual):raise RuntimeError('closed6 actual human original changed')
    for field,key in [('sourceCommit','sourceCommit'),('sourceManifestSHA256','sourceManifestSha256'),('binariesManifestSHA256','binariesManifestSha256'),('planSHA256','planSha256')]:
        if review.get(field)!=grant[key]:raise RuntimeError('closed6 original review/source pins differ')
    archive=doc(r.BASE/'recovery6/source-at-start.json')
    if pin(r.BASE/'recovery6/source-at-start.json')!=grant['sourceManifestSha256'] or archive['sourceCommit']!=stop['sourceCommit']:raise RuntimeError('closed6 original source archive differs')
    tail=ledger(native)
    if stop['interruptedM3NativeLedgerTail']!=tail:raise RuntimeError('closed6 exact interrupted ledger tail required')
    exact(doc(r.BASE/'recovery6/mm-global-formal-m3-chat-usd.json'),dict(reservedMicrousd=12884,attempts=4),'closed6 interrupted guard reservation is12884/4')
    sixth=dict(cohortSealSha256=pin(CLOSED),stopIndexSha256=pin(STOP),nativeSourceCommit=stop['sourceCommit'],actualHumanApprovalSHA256=pin(actual),retainedCompletedReportKnownCalls=720,retainedCompletedReportKnownHTTP=710,retainedKnownRecordedCalls=724,retainedKnownRecordedHTTPAllocations=714,retainedClaimedCalls=1204,retainedClaimedAttempts=1280,retainedReservedMicrousd=32900000,retainedUnknownParentProcesses=6,newCompletedReportCalls=18,newCompletedReportHTTP=18,newClaimedCalls=44,newClaimedAttempts=48,newReservedMicrousd=850000,interruptedNativeCalls=4,interruptedNativeHTTPAllocations=4,interruptedNativeLedgerRows=8,interruptedNativeLedgerTailSHA256=tail,interruptedUSDReservationMicrousd=12884,interruptedUSDTransportReservations=4,nativeAuthoritySHA256=authority,parentClaimsSHA256=parent_pins)
    return dict(sixthCohortSealSHA256=pin(CLOSED),sixthCohortHistorySHA256=digest(sixth),sixthCohort=sixth,previousCohort=prior,firstCohort=prior['firstCohort'],actualCalls=724,actualAttempts=714,completedReportKnownCalls=720,completedReportKnownHTTP=710,recordedHTTPAllocationsAreNotConfirmedVendorSends=True,claimedCalls=1204,claimedAttempts=1280,reservedMicrousd=32900000,originalProcessCount=50,authorizationSHA256=prior['authorizationSHA256'],credentialScopes=prior['credentialScopes'])

def prepare_metadata(prior):return dict(sixthCohortHistorySHA256=prior['sixthCohortHistorySHA256'])
def scope_metadata(prepared):return dict(executionOrder=list(r.EXECUTION_ORDER))

def actual_approval(prepared):
    value=doc(HUMAN);expected=dict(decision='APPROVED',round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=378,maxAttempts=414,campaignMaxCalls=1582,campaignMaxAttempts=1694,newReservedMicrousd=6900000,retainedReservedMicrousd=32900000,maxUSD=40,automaticProcessReruns=False,processReservationsMicrousd=r.RESERVES,executionOrder=list(r.EXECUTION_ORDER),originalCampaignApprovalSHA256=prepared['originalCampaignApprovalSHA256'])
    exact(value,expected,'new actual direct human fixed19/US40 approval required')
    for key in ('sourceCommit','sourceManifestSHA256','binariesManifestSHA256','planSHA256','sixthCohortSealSHA256','sixthCohortHistorySHA256','predecessorSHA256','spendingGuardSourceSHA256'):
        if value.get(key)!=prepared[key]:raise RuntimeError('new7 actual human source/binary/plan/history/guard differs')
    if not value.get('requestText') or not value.get('requestReference') or value['requestReference']==doc(r.BASE/'recovery6-authorization.json')['requestReference']:raise RuntimeError('old6 reply cannot authorize relabeled7')
    return value

def retained():
    from recovery7_retained import verify
    return verify()
