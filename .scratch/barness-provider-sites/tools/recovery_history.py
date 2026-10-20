"""Read-only closed2 facts; all old failures and occupied allowance survive."""
import json
from pathlib import Path
import final_predecessor as f
import recovery_contract as r
c=r.c
CLOSED=r.BASE/'closed-cohort-2/seal.json'
STOP=r.BASE/'remaining-stopped/index.json'
HUMAN=r.BASE/'recovery3-authorization.json'

def relative(path):return Path(path).relative_to(c.common_root()).as_posix()
def doc(path):return f.document(relative(path) if Path(path).is_absolute() else str(path))
def pin(path):return f.file_hash(relative(path) if Path(path).is_absolute() else str(path))[0]
def expand(value):
    for label,path in (('$REPO',c.ROOT),('$COMMON',c.common_root()),('$HOME',Path.home())):value=value.replace(label,str(path))
    return f.owned(relative(Path(value))) if Path(value).is_absolute() else f.owned(value)
def expected_old_names():return {'candidate-'+n for n in r.REGIONAL}|{'glm-chat','glm-responses','glm-responses-retry-1','glm-messages'}|{'formal-'+n for n in r.REGIONAL}|{'mm-global-formal-m3-chat'}

def facts():
    first=f.disclosure();seal=doc(CLOSED);stop=doc(STOP)
    expected=dict(schema=1,kind='CLOSED_SECOND_COHORT_HISTORY_SEAL',roundId=r.old.ROUND,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(STOP),stopIndexSHA256=pin(STOP),actualCalls=303,actualHTTPAttempts=302,claimedCalls=456,claimedHTTPAttempts=492,wholeReservedMicrousd=10700000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False)
    if any(seal.get(k)!=v or type(seal.get(k)) is not type(v) for k,v in expected.items()):raise RuntimeError('honest fixed closed2 seal with retained10.7USD required')
    expected_stop=dict(kind='ACTUAL_REMAINING_STOP_AFTER_REAL_GLOBAL_MINIMAX_CHAT_THINKING_FAILURE',regionalFormalProcesses=6,regionalFormalScenarios=81,globalMiniMaxActualCalls=13,globalMiniMaxActualHTTP=12,campaignActualCalls=303,campaignActualHTTP=302,campaignClaimedCalls=456,campaignClaimedHTTP=492,wholeReservedMicrousd=10700000,refundMicrousd=0,unknownParentProcesses=2,originalFailedGLMUnchanged=True,automaticReruns=False)
    if any(stop.get(k)!=v or type(stop.get(k)) is not type(v) for k,v in expected_stop.items()):raise RuntimeError('closed2 actual303/302 claimed456/492 must not be refunded or rewritten')
    files=stop['files'];total=0
    if not isinstance(files,dict) or len(files)!=234:raise RuntimeError('exact234 closed2 original/control pins required')
    for name,want in files.items():
        got,size=f.file_hash(name,(512<<20)-total);total+=size
        if got!=want:raise RuntimeError('closed2 original changed: '+name)
    launch=f.owned('.evidence/minimax-launches');prefix=r.old.ROUND+'-'
    actual={p.name.removeprefix(prefix).removesuffix('.json') for p in launch.glob(prefix+'*.json')}
    if actual!=expected_old_names():raise RuntimeError('exact17 old2 claims required; old2 is permanently closed')
    earlier=doc(r.BASE/'round2-stopped/index.json');earlier_files=earlier['files']
    if files.get(str(r.BASE/'round2-stopped/index.json'))!=pin(r.BASE/'round2-stopped/index.json') or len(earlier_files)!=131:raise RuntimeError('closed2 must bind the original131-file stopped prefix')
    total=0
    for name,want in earlier_files.items():
        got,size=f.file_hash(name,(512<<20)-total);total+=size
        if got!=want:raise RuntimeError('earlier original131 prefix changed: '+name)
    glm_path=r.BASE/'responses-retry-preparation/glm-three-routes.json';glm=doc(glm_path)
    if files.get(str(glm_path))!=pin(glm_path) or glm.get('kind')!='THREE_GENUINE_FULL22_GLM_PROOFS' or set(glm.get('routes',{}))!=set(r.ROUTES):raise RuntimeError('immutable genuine three-route GLM continuation pins required')
    all_pins=earlier_files|files
    for route,name in [('chat','glm-chat'),('responses','glm-responses-retry-1'),('messages','glm-messages')]:
        all_pins['.evidence/minimax-launches/'+prefix+name+'.json']=glm['routes'][route]['canonicalReceiptSHA256']
    receipts={}
    for name in sorted(actual):
        p=launch/(prefix+name+'.json');rel=relative(p)
        if all_pins.get(rel)!=pin(p):raise RuntimeError('all17 old canonical claims must be pinned by their original immutable layer')
        receipts[name]=doc(p)
    commits=set()
    for route,name in [('chat','glm-chat'),('responses','glm-responses-retry-1'),('messages','glm-messages')]:
        receipt=receipts[name];row=glm['routes'][route];commits.add(row['actualSourceCommit'])
        if receipt.get('sourceCommit')!=row['actualSourceCommit'] or receipt.get('budget')!=dict(state='REPORTED',status='PASS',calls=22,attempts=22) or receipt.get('nativeConsumerReceiptSha256')!=row['nativeConsumerSHA256']:raise RuntimeError('honest unchanged three full22 canonical GLM processes required')
    if len(commits)!=1:raise RuntimeError('preserved GLM must retain one actual measured commit')
    failed=receipts['mm-global-formal-m3-chat']
    row=next(v for v in stop['processes'] if v['name']=='mm-global-formal-m3-chat');report=doc(Path(row['nativeBundle'])/'minimax-report.json')
    if (failed.get('state')!='COMPLETED' or failed.get('exitCode')!=1 or failed.get('reservedMicrousd')!=200000 or failed.get('budget')!=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) or
        report.get('outcome')!='FAIL' or report.get('complete') is not False or report.get('auditPassed') is not True or report.get('budget',{}).get('calls')!=13 or report['budget'].get('attempts')!=12 or len(report.get('expected',[]))!=13 or len(report.get('calls',[]))!=13):raise RuntimeError('genuine failed native13/12 and UNKNOWN parent required')
    unknown=[v for v in receipts.values() if v.get('budget',{}).get('state')=='UNKNOWN']
    if len(unknown)!=2:raise RuntimeError('both historical UNKNOWN receipts must remain')
    unstarted=set(r.jobs(c.ROOT/'.evidence/unused-recovery').keys())-{'formal-'+n for n in r.REGIONAL}-{'mm-global-formal-m3-chat'}
    if set(stop['unstartedProcesses'])!=unstarted:raise RuntimeError('all17 unstarted processes must remain absent in old2')
    native=f.owned('.evidence/minimax-renewals/'+r.old.ROUND)
    native_names={str(p.relative_to(c.common_root())) for p in native.iterdir()}
    if native_names!={n for n in files if n.startswith(relative(native)+'/')}:raise RuntimeError('old2 native authority gained an extra intent or nonce')
    return dict(secondCohortSealSHA256=pin(CLOSED),stopIndexSHA256=pin(STOP),firstCohort=first,actualCalls=303,actualAttempts=302,claimedCalls=456,claimedAttempts=492,reservedMicrousd=10700000,originalProcessCount=17,oldClaimsSHA256={n:pin(launch/(prefix+n+'.json')) for n in sorted(actual)},authorizationSHA256=first['authorizationSHA256'],credentialScopes=first['credentialScopes'])

def preserved():
    # Historical GLM remains R2/b384, with its independently closed publisher.
    import final_publication
    index=c.doc(c.ROOT/'ai/live/testdata/glm-formal/live-captures.json')
    final_publication.validate(index)
    return {route:final_publication.require(route)[1] for route in r.ROUTES}

def actual_approval(prepared):
    value=doc(HUMAN)
    expected=dict(decision='APPROVED',validationRound=r.ROUND,nativeRenewalId=r.ROUND,spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR,maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,maxUSD=30,automaticProcessReruns=False)
    if any(value.get(k)!=v or type(value.get(k)) is not type(v) for k,v in expected.items()) or not value.get('requestText') or not value.get('requestReference'):raise RuntimeError('new actual human approval for fixed recovery3 required')
    for k in ('sourceCommit','sourceManifestSHA256','binariesManifestSHA256','planSHA256','secondCohortSealSHA256','spendingGuardSourceSHA256'):
        if value.get(k)!=prepared[k]:raise RuntimeError('actual new human reply must bind this source/binary/plan/policy')
    return value

# No new metadata is retroactively added to the approved fixed3 contract.
def prepare_metadata(prior):return {}
def scope_metadata(prepared):return {}
