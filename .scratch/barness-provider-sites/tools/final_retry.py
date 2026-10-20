"""One expressly approved Responses amendment; the original33 scope stays closed."""
import os
import json
from pathlib import Path
import subprocess
import final_contract as r
import final_predecessor as f
c=r.c
BASE=f.BASE
NAME='glm-responses-retry-1'
APPROVAL=BASE/'responses-retry-authorization.json'
STOPPED=BASE/'round2-stopped/index.json'
FAILED='.evidence/minimax-launches/'+r.ROUND+'-glm-responses.json'

def pin(name):return f.file_hash(str(name))[0]
def doc(name):return f.document(str(name))
def relative(path):return Path(path).relative_to(c.common_root()).as_posix()
def local_doc(out,name):return doc(relative(out/name))
def local_pin(out,name):return pin(relative(out/name))
def old_output(value):
    value=value.removeprefix('$REPO/').removeprefix('$COMMON/')
    path=Path(value)
    if path.is_absolute():value=relative(path)
    return f.owned(value)

def facts():
    human=doc(APPROVAL);stop=doc(STOPPED)
    if (human.get('kind')!='ACTUAL_HUMAN_ONE_EXPLICIT_GLM_RESPONSES_RETRY_AND_CONTINUATION' or human.get('decision')!='APPROVED' or
        not human.get('humanText') or not human.get('humanMessageReference') or human.get('originalFailedClaim')!=FAILED or
        human.get('originalFailedClaimSHA256')!=pin(FAILED) or human.get('stoppedFactsSHA256')!=pin(STOPPED) or
        human.get('candidatePromotionPreserved') is not True or human.get('researchCallsAuthorized')!=0):
        raise RuntimeError('one exact human retry approval and original failed claim required')
    scope=dict(provider='zhipu',api='openai-responses',models=['glm-5.3','glm-5.2'],completeScenarios=22,additionalCompleteProcesses=1,retainOriginalFailure=True,continueOriginalRemainingProcessesOnSuccess=True,finalFullRaceOnFreshREADY=True,maxCampaignUSD=30,automaticAdditionalProcessReruns=False)
    allocations=dict(retainPrefixWholeMicrousd=6500000,retryMicrousd=500000,unstartedGLMMessagesMicrousd=500000,allOtherOriginalReservationsUnchanged=True,totalWholeMicrousd=29900000,knownRateOperationalCeilingsNotInvoices=True)
    if (human.get('scope')!=scope or any(human.get('derivedNativeProcessBounds',{}).get(k)!=v for k,v in dict(maxCalls=24,maxHTTPAttempts=32).items()) or
        any(human.get('derivedCampaignCeilings',{}).get(k)!=v for k,v in dict(maxCalls=798,maxHTTPAttempts=842).items()) or human.get('operatorConservativeAllocations')!=allocations):
        raise RuntimeError('fixed24/32 amendment,798/842 campaign and .5/.5 allocations required')
    expected=dict(kind='STOPPED_ACTUAL_ROUND2_AFTER_REAL_GLM_RESPONSES_FAILURE',roundId=r.ROUND,state='STOPPED_NO_AUTOMATIC_PAID_CONTINUATION',completedProcesses=8,newActualCalls=125,newActualHTTPAttempts=125,newClaimedCalls=192,newClaimedHTTPAttempts=208,newWholeReservedMicrousd=5000000,campaignActualCalls=165,campaignActualHTTPAttempts=165,campaignClaimedCalls=244,campaignClaimedHTTPAttempts=260,campaignWholeReservedMicrousd=6500000,refundMicrousd=0,failedParentUNKNOWNUnchanged=True,nativeM3GrantNeverIssued=True)
    if any(stop.get(k)!=v or type(stop.get(k)) is not type(v) for k,v in expected.items()):raise RuntimeError('immutable original8 process accounting and6.5USD required')
    if not 1<=len(stop['files'])<=512:raise RuntimeError('bounded stopped prefix inventory required')
    total=0
    for name,want in stop['files'].items():
        got,size=f.file_hash(name,(128<<20)-total);total+=size
        if got!=want:raise RuntimeError('stopped original changed: '+name)
    old=doc(FAILED);old_out=old_output(old['collectionRoot'])
    report_name=next(v['nativeBundle'] for v in stop['processes'] if v['name']=='glm-responses')+'/formal-glm-report.json'
    report=doc(report_name)
    if (old.get('state')!='COMPLETED' or old.get('exitCode')!=1 or old.get('reservedMicrousd')!=1000000 or
        old.get('budget')!=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) or report.get('outcome')!='FAIL' or
        len(report.get('scenarios',[]))!=22 or report.get('budget',{}).get('callsUsed')!=22 or report['budget'].get('httpAttempts')!=22):
        raise RuntimeError('original complete22 real FAIL/parent UNKNOWN must remain unchanged')
    global_name='.evidence/minimax-renewals/'+r.ROUND+'/operator-authority.json';authority=doc(global_name)
    if f.file_hash('.evidence/minimax-renewals/'+r.ROUND+'/launch.lock')[1]!=0:raise RuntimeError('original canonical launcher lock must remain a regular empty file')
    if (human.get('originalR2ApprovalSHA256')!=pin(BASE/'round2-authorization.json') or
        human.get('originalEligibilityAndFinalRaceApprovalSHA256')!=authority['eligibilityAuthorizationSHA256'] or
        authority.get('maxCalls')!=722 or authority.get('maxAttempts')!=758 or authority.get('maxMicrousd')!=30000000 or
        authority.get('newMaxMicrousd')!=28400000):raise RuntimeError('original R2 authority stays an unchanged independent sub-scope')
    return dict(approvalSHA256=pin(APPROVAL),stoppedSHA256=pin(STOPPED),failedClaimSHA256=pin(FAILED),failedReportSHA256=pin(report_name),originalAuthoritySHA256=pin(global_name),measuredCommit=stop['sourceCommit'],oldOutput=str(old_out),humanText=human['humanText'],humanReference=human['humanMessageReference'])

def context(checkout):
    value=facts();checkout=Path(checkout).resolve()
    if checkout in (c.ROOT.resolve(),c.common_root().resolve()):raise RuntimeError('fixed execution requires a separate clean checkout')
    value['checkout']=str(checkout)
    old=Path(value['oldOutput']);manifest=doc(relative(old/'binaries.json'))
    for name in ('live.test','glmformal'):
        path=old/name
        if pin(relative(path))!=manifest[name]:raise RuntimeError('original frozen GLM binary changed')
        value[name]=dict(path=str(path),sha256=manifest[name])
    chat=doc('.evidence/minimax-launches/'+r.ROUND+'-glm-chat.json')
    bundle=next(v['nativeBundle'] for v in doc(STOPPED)['processes'] if v['name']=='glm-chat')
    value['source']=doc(bundle+'/glm-execution-source.json')
    snapshot=doc(relative(old/'source-at-start.json'))['files']
    value['catalogFiles']={name:snapshot[name] for name in ('ai/catalog_minimax.json','ai/catalog_regional.json','ai/release/catalog-snapshot.json')}
    if chat.get('sourceCommit')!=value['measuredCommit']:raise RuntimeError('successful Chat must retain the measured commit')
    guard_context(value)
    return value

def git_at(root,*args):return subprocess.check_output(['git',*args],cwd=root,env=c.ENV,text=True,timeout=30).strip()
def guard_context(value):
    expected=facts()
    if {k:v for k,v in value.items() if k in expected}!=expected:raise RuntimeError('fixed amendment history or human approval changed')
    old=Path(expected['oldOutput']);manifest=doc(relative(old/'binaries.json'))
    bundle=next(v['nativeBundle'] for v in doc(STOPPED)['processes'] if v['name']=='glm-chat')
    snapshot=doc(relative(old/'source-at-start.json'))['files']
    catalog={n:snapshot[n] for n in ('ai/catalog_minimax.json','ai/catalog_regional.json','ai/release/catalog-snapshot.json')}
    if value.get('source')!=doc(bundle+'/glm-execution-source.json') or value.get('catalogFiles')!=catalog:
        raise RuntimeError('execution context must retain original measured source and catalog')
    for name in ('live.test','glmformal'):
        if value.get(name)!=dict(path=str(old/name),sha256=manifest[name]):raise RuntimeError('only the original measured binary may execute')
    checkout=Path(value['checkout']);common=Path(git_at(checkout,'rev-parse','--git-common-dir'))
    if not common.is_absolute():common=checkout/common
    if (common.resolve().parent!=c.common_root().resolve() or git_at(checkout,'rev-parse','HEAD')!=value['measuredCommit'] or
        git_at(checkout,'status','--porcelain')):raise RuntimeError('same-owner clean measured-commit checkout required')
    for name in ('live.test','glmformal'):
        if pin(relative(Path(value[name]['path'])))!=value[name]['sha256']:raise RuntimeError('frozen measured binary changed')
    for root in (c.ROOT,checkout):
        exported=json.loads(subprocess.check_output([value['glmformal']['path'],'-source'],cwd=root/'ai/live',env=c.ENV,timeout=30))
        if exported!={'files':value['source']}:raise RuntimeError('complete native GLM execution source changed')
        for name,want in value['catalogFiles'].items():
            if c.digest(root/name)!=want:raise RuntimeError('measured catalog bytes changed')

def load(out):
    value=local_doc(out,'execution-context.json');guard_context(value);return value

def jobs(out):
    ctx=local_doc(out,'execution-context.json');base=r.jobs(out)
    result={}
    for name,route in ((NAME,'responses'),('glm-messages','messages')):
        job=base['glm-'+route].copy();job.update(name=name,baseName=name,cwd=str(Path(ctx['checkout'])/'ai/live'),evidenceRoot=str(out/name),measuredCommit=ctx['measuredCommit'])
        job['environment']=job['environment']|dict(BARNESS_AI_EVIDENCE_DIR=str(out/name))
        result[name]=job
    return result

def plan(out):return r.inventory(out)|dict(phase='glm',amendment=NAME,amendmentMaxCalls=24,amendmentMaxAttempts=32,amendedCampaignMaxCalls=798,amendedCampaignMaxAttempts=842,routes=jobs(out))

def template(out,prepared):
    ctx=local_doc(out,'execution-context.json')
    return prepared|dict(decision='PENDING',amendment=NAME,humanApprovalSHA256=ctx['approvalSHA256'],requestText=ctx['humanText'],requestReference=ctx['humanReference'],failedClaimSHA256=ctx['failedClaimSHA256'],networkRoute='direct',reservedMicrousd=500000,maxCalls=24,maxAttempts=32,campaignMaxCalls=798,campaignMaxAttempts=842,maxUSD=30,automaticAdditionalProcessReruns=False)

def approved(out,prepared):
    auth=local_doc(out,'retry-authorization.json');expected=template(out,prepared)|dict(decision='APPROVED')
    if auth!=expected:raise RuntimeError('exact current reviewed .5USD one-process amendment required')
    return auth

def selected_name(route):
    import final_execute as e
    if route!='responses' or not e.marker(NAME).exists():return 'glm-'+route
    return NAME

def continuation():
    import final_execute as e
    marker=e.marker(NAME)
    if not marker.exists():return None
    value=doc(relative(marker));out=old_output(value['collectionRoot']);ctx=load(out)
    if value.get('amendmentFacts')!=c.redact(facts()) or value.get('executionContextSHA256')!=local_pin(out,'execution-context.json'):
        raise RuntimeError('canonical one-use amendment facts changed')
    return value


def copy_live(out):
    ctx=local_doc(out,'execution-context.json')
    with f.regular(relative(Path(ctx['live.test']['path'])),64<<20) as (source,size):
        with (out/'live.test').open('xb') as target:
            read=0
            while True:
                chunk=source.read(min(65536,size-read+1))
                if not chunk:break
                read+=len(chunk)
                if read>size:raise RuntimeError('original measured binary grew during bounded copy')
                target.write(chunk)
            if read!=size:raise RuntimeError('original measured binary was truncated')
            target.flush();os.fsync(target.fileno())
    if local_pin(out,'live.test')!=ctx['live.test']['sha256']:raise RuntimeError('exclusive measured binary copy differs')
