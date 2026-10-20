"""Fixed closed3 history; old native PASS is historical, never current qualification."""
import hashlib
import json
from pathlib import Path
import recovery_history as previous
from recovery_flow_files import claim_names
import recovery4_contract as r
f=previous.f
c=r.c
CLOSED=r.BASE/'closed-cohort-3/seal.json'
STOP=r.BASE/'recovery3-stopped/index.json'
HUMAN=r.BASE/'recovery4-authorization.json'
OLD='provider-sites-final-acceptance-3'
NAMES=tuple('mm-global-formal-m3-'+v for v in r.ROUTES)+('formal-zai-chat','formal-zai-responses')
AUTHORITY=('chat.json','formal.jsonl','grant.json','human-approval.json','launch.lock','messages.json','operator-authority.json','responses.json','source-review.json')
relative=previous.relative
pin=previous.pin
expand=previous.expand
preserved=previous.preserved

def strict(raw):
    def pairs(items):
        result={};seen=set()
        for key,value in items:
            folded=key.casefold()
            if folded in seen:raise RuntimeError('duplicate or case-alias recovery JSON field')
            seen.add(folded);result[key]=value
        return result
    return json.loads(raw,object_pairs_hook=pairs)

def doc(path):
    name=relative(path) if Path(path).is_absolute() else str(path)
    with f.regular(name,2<<20) as (stream,size):
        raw=stream.read(size+1)
        if len(raw)!=size:raise RuntimeError('recovery JSON changed during bounded read')
    value=strict(raw)
    if not isinstance(value,dict):raise RuntimeError('recovery JSON object required')
    return value

def exact(value,expected,label):
    if any(value.get(k)!=v or type(value.get(k)) is not type(v) for k,v in expected.items()):raise RuntimeError(label)

def digest(value):return hashlib.sha256(json.dumps(value,sort_keys=True,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()

def ledger(native,grant):
    with f.regular(relative(native/'formal.jsonl'),4<<20) as (stream,size):raw=stream.read(size+1)
    if len(raw)!=size:raise RuntimeError('closed3 ledger changed')
    rows=[strict(line) for line in raw.splitlines()];tail='';counts={v:[0,0] for v in r.ROUTES};sessions={}
    keys=('renewalId','renewalGrantSha256','session','sequence','stage','identity','case','kind','output','input','at','previous','hash')
    for i,row in enumerate(rows):
        if set(row)!=set(keys):raise RuntimeError('closed3 native receipt fields differ')
        wire={key:row[key] for key in keys};wire['hash']=''
        if row['sequence']!=i+1 or row['previous']!=tail or hashlib.sha256(json.dumps(wire,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()!=row['hash']:raise RuntimeError('closed3 native75 chain differs')
        route=row['identity'].get('route')
        if route not in counts or row['renewalId']!=OLD or row['renewalGrantSha256']!=pin(native/'grant.json') or row['stage']!='formal' or row['kind'] not in ('reserve','attempt'):raise RuntimeError('closed3 native authority identity differs')
        counts[route][row['kind']=='attempt']+=1;sessions.setdefault(route,set()).add(row['session']);tail=row['hash']
    if len(rows)!=75 or any(v!=[13,12] for v in counts.values()) or any(len(v)!=1 for v in sessions.values()):raise RuntimeError('closed3 exact successful39/36 required')
    for route in r.ROUTES:
        nonce=doc(native/(route+'.json'))
        exact(nonce,dict(schema=2,renewalId=OLD,grantSha256=pin(native/'grant.json'),session=next(iter(sessions[route]))),'closed3 single-use nonce differs')
        if len(nonce.get('processNonce',''))!=64 or nonce['identity']!=next(row['identity'] for row in rows if row['identity']['route']==route):raise RuntimeError('closed3 original nonce identity missing')
    return tail

def claims(stop,files,grant):
    launch=f.owned('.evidence/minimax-launches');actual=claim_names(c,OLD)
    if actual!=set(NAMES) or {v['name'] for v in stop['processes']}!=set(NAMES) or len(stop['processes'])!=5:raise RuntimeError('exact five closed3 parent processes required; no reopening')
    result={}
    for name in NAMES:
        p=launch/(OLD+'-'+name+'.json');rel=relative(p);value=doc(p);row=next(v for v in stop['processes'] if v['name']==name);mm=name.startswith('mm-');failed=name=='formal-zai-responses'
        calls,attempts=(13,12) if mm else (18,18);budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        if files.get(rel)!=pin(p):raise RuntimeError('closed3 original parent pin differs')
        exact(value,dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=stop['sourceCommit'],reservedMicrousd=200000 if mm else 650000,budget=budget),'closed3 actual FAIL/UNKNOWN and successful original parents required')
        exact(row,dict(exitCode=1 if failed else 0,parentBudget=budget,reservedMicrousd=value['reservedMicrousd'],actualCalls=calls,actualHTTP=attempts),'closed3 parent accounting differs')
        base=expand(row['nativeBundle']);report_path=base/('minimax-report.json' if mm else 'regional-report.json');report=doc(report_path)
        if files.get(relative(report_path))!=pin(report_path) or row['nativeReportSHA256']!=pin(report_path) or report.get('gitCommit')!=stop['sourceCommit'] or doc(base/'audit.json')!=dict(findings=[]):raise RuntimeError('closed3 native original/source/audit differs')
        if mm:
            if digest(report.get('executionSource',{}))!=grant['executionSourceSha256']:raise RuntimeError('closed3 original native source map differs from its historical grant')
            exact(report,dict(formalRenewalId=OLD,formalRenewalGrantSha256=pin(f.owned('.evidence/minimax-renewals/'+OLD+'/grant.json')),outcome='PASS',complete=True,auditPassed=True),'closed3 historical native PASS required')
            if report['budget']['calls']!=13 or report['budget']['attempts']!=12 or len(report['expected'])!=13 or len(report['calls'])!=13 or any(v['outcome']!='PASS' for v in report['calls']):raise RuntimeError('closed3 complete13 M3 calls required')
        else:
            scenarios=report['scenarios']
            if report['budget']['callsUsed']!=18 or report['budget']['httpAttempts']!=18 or len(scenarios)!=18 or sum(v['outcome']!='PASS' for v in scenarios)!=(1 if failed else 0):raise RuntimeError('closed3 honest18 ZAI scenarios required')
        for key,field in [('sourceManifestSha256','sourceManifestSha256'),('binariesManifestSha256','binariesManifestSha256'),('planSha256','planSha256')]:
            if value[key]!=grant[field]:raise RuntimeError('closed3 historical control/source pins differ')
        command=f.owned(str(r.BASE/'recovery3'/(name+'-command.json')));cmd=doc(command)
        if files.get(relative(command))!=pin(command) or value['commandReceiptSha256']!=pin(command) or cmd.get('state')!='COMPLETED' or cmd.get('exitCode')!=(1 if failed else 0) or cmd.get('launcherAudit')!=dict(passed=True,findings=[]):raise RuntimeError('closed3 original command differs')
        if not failed:
            consumer=f.owned(str(r.BASE/'recovery3'/(name+'-consume-command.json')));con=doc(consumer)
            if value.get('reportedFileSha256')!=pin(report_path) or value.get('nativeConsumerReceiptSha256')!=pin(consumer) or files.get(relative(consumer))!=pin(consumer) or con.get('state')!='COMPLETED' or con.get('exitCode')!=0 or con.get('credentialsRead') is not False:raise RuntimeError('closed3 successful original native consumer required')
        result[name]=pin(p)
    return result

def facts():
    prior=previous.facts();seal=doc(CLOSED);stop=doc(STOP)
    expected=dict(schema=1,kind='CLOSED_THIRD_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(STOP),stopIndexSHA256=pin(STOP),actualCalls=378,actualHTTPAttempts=374,claimedCalls=564,claimedHTTPAttempts=612,wholeReservedMicrousd=12600000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False)
    if set(seal)!=set(expected):raise RuntimeError('strict14 closed3 seal fields required')
    exact(seal,expected,'closed3 seal/counters must retain all consumption')
    exact(stop,dict(schema=1,kind='ACTUAL_RECOVERY3_STOP_AFTER_ZAI_RESPONSES_NATIVE_SOURCE_FAILURE',successfulProcesses=4,failedProcesses=1,newActualCalls=75,newActualHTTP=72,campaignActualCalls=378,campaignActualHTTP=374,campaignClaimedCalls=564,campaignClaimedHTTP=612,newWholeReservedMicrousd=1900000,wholeReservedMicrousd=12600000,refundMicrousd=0,unknownParentProcesses=3,originalFailedGLMUnchanged=True,originalFailedM3Unchanged=True,automaticReruns=False),'closed3 honest378/374 with three UNKNOWN required')
    files=stop['files'];total=0
    if not isinstance(files,dict) or len(files)!=1012:raise RuntimeError('exact1012 stopped3 original pins required')
    for name,want in files.items():
        got,size=f.file_hash(name,(512<<20)-total);total+=size
        if got!=want:raise RuntimeError('closed3 original changed: '+name)
    if set(stop['unstartedProcesses'])!=set(r.jobs(c.ROOT/'.evidence/unused-recovery4'))-set(NAMES):raise RuntimeError('all19 unstarted old3 processes must remain absent')
    native=f.owned('.evidence/minimax-renewals/'+OLD)
    if {p.relative_to(native).as_posix() for p in f.native_files(native)}!=set(AUTHORITY):raise RuntimeError('closed3 exact9 authority leaves required; no extra intent')
    authority={name:pin(native/name) for name in AUTHORITY}
    if any(files.get(relative(native/name))!=value for name,value in authority.items()):raise RuntimeError('closed3 native authority must be in1012 original pins')
    grant=doc(native/'grant.json');human=doc(native/'human-approval.json');review=doc(native/'source-review.json');actual=r.BASE/'recovery3-authorization.json'
    exact(grant,dict(schema=4,decision='APPROVED',renewalId=OLD,sourceCommit=stop['sourceCommit'],humanApprovalSha256=pin(native/'human-approval.json'),sourceReviewSha256=pin(native/'source-review.json'),maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000),'closed3 unchanged historical schema4 authority required')
    exact(human,dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,campaignMaximumMicrousd=30000000,actualHumanApprovalSHA256=pin(actual)),'old schema4 approved scope must not be relabeled4')
    exact(review,dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR),'closed3 original approved review identity required')
    old_actual=doc(actual)
    exact(old_actual,dict(decision='APPROVED',round=OLD),'closed3 actual original reply required')
    if not old_actual.get('requestText') or not old_actual.get('requestReference'):raise RuntimeError('closed3 actual direct reply missing')
    for field,key in [('sourceCommit','sourceCommit'),('sourceManifestSHA256','sourceManifestSha256'),('binariesManifestSHA256','binariesManifestSha256'),('planSHA256','planSha256')]:
        if review.get(field)!=grant[key]:raise RuntimeError('closed3 original review/source pins differ')
    if stop['actualHumanApprovalSHA256']!=pin(actual) or files.get(str(actual))!=pin(actual):raise RuntimeError('closed3 actual human original pin differs')
    tail=ledger(native,grant);parents=claims(stop,files,grant)
    third=dict(actualHumanApprovalSha256=pin(actual),cohortSealSha256=pin(CLOSED),nativeAttempts=36,nativeAuthoritySha256=authority,nativeCalls=39,nativeLedgerRows=75,nativeLedgerTailSha256=tail,nativeSourceCommit=stop['sourceCommit'],newActualAttempts=72,newActualCalls=75,newClaimedAttempts=120,newClaimedCalls=108,newReservedMicrousd=1900000,parentClaimsSha256=parents,retainedActualAttempts=374,retainedActualCalls=378,retainedClaimedAttempts=612,retainedClaimedCalls=564,retainedReservedMicrousd=12600000,retainedUnknownParentProcesses=3,stopIndexSha256=pin(STOP))
    return dict(thirdCohortSealSHA256=pin(CLOSED),thirdCohortHistorySHA256=digest(third),thirdCohort=third,previousCohort=prior,firstCohort=prior['firstCohort'],secondCohortSealSHA256=prior['secondCohortSealSHA256'],actualCalls=378,actualAttempts=374,claimedCalls=564,claimedAttempts=612,reservedMicrousd=12600000,originalProcessCount=22,authorizationSHA256=prior['authorizationSHA256'],credentialScopes=prior['credentialScopes'])

def prepare_metadata(prior):return dict(thirdCohortHistorySHA256=prior['thirdCohortHistorySHA256'])
def scope_metadata(prepared):return dict(executionOrder=list(r.EXECUTION_ORDER))

def actual_approval(prepared):
    value=doc(HUMAN)
    expected=dict(decision='APPROVED',round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=1070,campaignMaxAttempts=1130,newReservedMicrousd=17300000,retainedReservedMicrousd=12600000,maxUSD=30,automaticProcessReruns=False,processReservationsMicrousd=r.RESERVES,executionOrder=list(r.EXECUTION_ORDER),originalCampaignApprovalSHA256=prepared['originalCampaignApprovalSHA256'])
    exact(value,expected,'new actual human approval for fixed24 recovery4 required')
    for key in ('sourceCommit','sourceManifestSHA256','binariesManifestSHA256','planSHA256','thirdCohortSealSHA256','thirdCohortHistorySHA256','predecessorSHA256','spendingGuardSourceSHA256'):
        if value.get(key)!=prepared[key]:raise RuntimeError('new4 human must bind this exact source/binary/plan/history/policy')
    if not value.get('requestText') or not value.get('requestReference'):raise RuntimeError('actual new4 direct reply required')
    if value['requestReference']==doc(r.BASE/'recovery3-authorization.json')['requestReference']:raise RuntimeError('old3 reply cannot authorize a relabeled4 scope')
    return value
