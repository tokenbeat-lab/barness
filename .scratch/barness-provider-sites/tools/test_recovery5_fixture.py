"""Synthetic closed4 and executable producer contracts for public fixed5 processes."""
import ast
import hashlib
import json
from pathlib import Path
from types import SimpleNamespace
from test_recovery4_fixture import seed as seed3,builder as builder4,BASE
OLD='provider-sites-final-acceptance-4';NEW='provider-sites-final-acceptance-5'
ROUTES=('chat','responses','messages')
GENERIC=('openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image')
NAMES=('formal-zai-responses',)+tuple('mm-global-formal-m3-'+r for r in ROUTES)+('formal-zai-chat','formal-zai-messages')+tuple('formal-minimax-cn-'+r for r in ROUTES)+GENERIC+('openai-decisions',)

def seed(test):
    seed3(test);root=test.root;files={};commit=test.git('rev-parse','HEAD')
    def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
    def put(name,value,raw=False):
        p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(value if raw else (json.dumps(value)+'\n').encode());files[str(name)]=sha(p);return p
    from recovery_flow_source import SourceFreeze
    oldfiles=SourceFreeze(SimpleNamespace(ROOT=root)).files()
    archive=put(BASE/'recovery4/source-at-start.json',dict(sourceCommit=commit,files=oldfiles));manifest=sha(archive)
    actual=put(BASE/'recovery4-authorization.json',dict(decision='APPROVED',round=OLD,requestText='CONTROLLED actual4',requestReference='CONTROLLED direct new4 reply'))
    native=Path('.evidence/minimax-renewals')/OLD
    scope=dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId='provider-sites-final-acceptance-2',maxCalls=506,maxAttempts=518,campaignMaxCalls=1070,campaignMaxAttempts=1130,newReservedMicrousd=17300000,retainedReservedMicrousd=12600000,campaignMaximumMicrousd=30000000,actualHumanApprovalSHA256=sha(actual))
    hp=put(native/'human-approval.json',scope);review=put(native/'source-review.json',scope|dict(sourceCommit=commit,sourceManifestSHA256=manifest,binariesManifestSHA256='b'*64,planSHA256='p'*64))
    grant=put(native/'grant.json',dict(schema=5,decision='APPROVED',renewalId=OLD,sourceCommit=commit,humanApprovalSha256=sha(hp),sourceReviewSha256=sha(review),sourceManifestSha256=manifest,binariesManifestSha256='b'*64,planSha256='p'*64,maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000))
    put(native/'operator-authority.json',scope);put(native/'launch.lock',b'',True)
    rows=[];tail=''
    for route,api in zip(ROUTES,('openai-completions','openai-responses','anthropic-messages')):
        identity=dict(route=route,model='MiniMax-M3',purpose='standard',api=api,endpoint='https://api.minimax.io/anthropic' if route=='messages' else 'https://api.minimax.io/v1')
        put(native/(route+'.json'),dict(schema=2,renewalId=OLD,grantSha256=sha(grant),session='old4-'+route,identity=identity,processNonce=hashlib.sha256(route.encode()).hexdigest()))
        for i in range(13):
            for kind in (('reserve','attempt') if i<12 else ('reserve',)):
                row=dict(renewalId=OLD,renewalGrantSha256=sha(grant),session='old4-'+route,sequence=len(rows)+1,stage='formal',identity=identity,case='case-'+str(i),kind=kind,output=1 if kind=='reserve' else 0,input=2048 if kind=='reserve' else 0,at='2026-10-20T00:00:00Z',previous=tail,hash='')
                tail=hashlib.sha256(json.dumps(row,separators=(',',':')).encode()).hexdigest();row['hash']=tail;rows.append(row)
    put(native/'formal.jsonl',(''.join(json.dumps(v,separators=(',',':'))+'\n' for v in rows)).encode(),True)
    counts=dict(zip(GENERIC,(14,14,20,8,10,13,3,2)));processes=[];sourcefiles={};cases=[]
    for name in NAMES:
        mm=name.startswith('mm-');failed=name=='openai-decisions';calls=13 if mm else 18 if name.startswith('formal-zai-') else 9 if name.startswith('formal-') or failed else counts[name];attempts=12 if mm else calls
        reserve=200000 if mm or failed else 650000 if name.startswith('formal-zai-') else 350000 if name.startswith('formal-') else dict(zip(GENERIC,(1500000,1500000,1800000,500000,500000,500000,3500000,2600000)))[name]
        base=BASE/'recovery4'/('live' if name in GENERIC or failed else 'raw')/name/'old4-session';leaf='minimax-report.json' if mm else 'regional-report.json' if name.startswith('formal-') else 'live-report.json'
        rp=put(base/leaf,dict(gitCommit=commit,combo=name,budget=dict(calls=calls,attempts=attempts),scenarios=[dict(outcome='FAIL' if failed else 'PASS') for i in range(calls)]));put(base/'audit.json',dict(findings=[]))
        command=put(BASE/'recovery4'/(name+'-command.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,launcherAudit=dict(passed=True,findings=[])))
        consumer=put(BASE/'recovery4'/(name+'-consume-command.json'),dict(state='COMPLETED',exitCode=0,credentialsRead=False)) if not failed else None
        budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        parent=put(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=commit,reservedMicrousd=reserve,budget=budget,commandReceiptSha256=sha(command),reportedFileSha256=sha(rp) if not failed else None,nativeConsumerReceiptSha256=sha(consumer) if consumer else None))
        processes.append(dict(name=name,nativeBundle=str(base),nativeReportSHA256=sha(rp),nativeConsumerReceiptSHA256=sha(consumer) if consumer else None,actualCalls=calls,actualHTTP=attempts,parentBudget=budget,exitCode=1 if failed else 0))
        if name in GENERIC:
            for i in range(calls):
                wire=[dict(method='POST',url='https://controlled.invalid/'+name,requestBody=json.dumps(dict(model=name,index=i)))];source=put(base/('capture-'+str(i)+'.json'),wire);sourcefiles[str(source.relative_to(root))]=sha(source)
                case=dict(source=str(source.relative_to(root))+':0',ledgerBytesEqual=True,methodEqual=True,urlEqual=True,requestBodyBytesEqual=True,baselineRequestSHA256=hashlib.sha256(wire[0]['requestBody'].encode()).hexdigest(),currentRequestSHA256=hashlib.sha256(wire[0]['requestBody'].encode()).hexdigest())
                for phase in ('baseline','current'):
                    for kind,value in [('Wire',wire),('Ledger',dict(usd=1000))]:
                        p=put(BASE/'controlled-compat'/phase/name/(str(i)+kind+'.json'),value);case[phase+kind+'File']=str(p.relative_to(root));case[phase+kind+'FileSHA256']=sha(p)
                cases.append(case)
    for i in range(2032-len(files)):put(BASE/'recovery4-stopped'/('original-'+str(i)+'.json'),dict(controlled=i))
    stop=dict(schema=1,actualHumanApprovalSHA256=sha(actual),kind='ACTUAL_RECOVERY4_STOP_AFTER_OPENAI_DECISIONS_PRESEND_BUDGET_FAILURE',sourceCommit=commit,successfulProcesses=17,failedProcesses=1,newActualCalls=213,newActualHTTP=210,campaignActualCalls=591,campaignActualHTTP=584,campaignClaimedCalls=980,campaignClaimedHTTP=1040,newWholeReservedMicrousd=16200000,wholeReservedMicrousd=28800000,refundMicrousd=0,unknownParentProcesses=4,failedNativeRecordedCalls=9,failedNativeRecordedHTTP=9,failedUSDReservationMicrousd=0,failedUSDTransportReservations=0,automaticReruns=False,processes=processes,unstartedProcesses=['typesafe-classifier','protocol114','text-jev','text-luna','visual-jev','visual-decisions'],files=dict(files))
    sp=put(BASE/'recovery4-stopped/index.json',stop)
    put(BASE/'closed-cohort-4/seal.json',dict(schema=1,kind='CLOSED_FOURTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(BASE/'recovery4-stopped/index.json'),stopIndexSHA256=sha(sp),actualCalls=591,actualHTTPAttempts=584,claimedCalls=980,claimedHTTPAttempts=1040,wholeReservedMicrousd=28800000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False))
    (root/'ai/internal/testkit/finalbudget/estimate.go').write_text('package corrected\n')
    current={n:sha(root/n) for n in ('ai/internal/testkit/finalbudget/guard.go','ai/internal/testkit/finalbudget/estimate.go')}
    si=put(BASE/'controlled-compat/source-index.json',dict(files=sourcefiles))
    put(BASE/'decisions-budget-diagnosis-dev/compatibility-eight.json',dict(schema=1,kind='PUBLIC_CLIENT_FINAL_WIRE_GUARD_COMPATIBILITY',state='PASS',count=84,routes=8,requestBodyAndOwnedLedgerBytesUnchanged=True,baselineSourceCommit=commit,baselineSpendingGuardSourceSHA256={n:oldfiles[n] for n in current},currentSpendingGuardSourceSHA256=current,routeCounts=counts,originalSourceIndex=str(si.relative_to(root)),originalSourceIndexSHA256=sha(si),originalSourceFiles=sourcefiles,cases=cases))
    verifier=test.tools/'recovery5_retained.py';verifier.write_text(verifier.read_text().replace('f5d37cc92abc7d9f4db844e73ad8e00efc63d7c316824ae43afff917280907a7',sha(root/BASE/'decisions-budget-diagnosis-dev/compatibility-eight.json')))
    test.commit()


def builder(test):
    builder4(test);go=test.root/'bin/go';text=go.read_text();tree=ast.parse(text)
    producer=next(n.args[0].value for n in ast.walk(tree) if isinstance(n,ast.Call) and isinstance(n.func,ast.Attribute) and n.func.attr=='write_text' and isinstance(n.args[0],ast.Constant))
    replacement=producer.replace("new='provider-sites-final-acceptance-4'","new='provider-sites-final-acceptance-5'").replace('-recovery4-source','-recovery5-source').replace('-recovery4-preflight=','-recovery5-preflight=').replace("g['schema']==5","g['schema']==6")
    replacement=replacement.replace("prior['secondCohortSealSHA256']","prior['previousCohort']['secondCohortSealSHA256']")
    replacement=replacement.replace("prior['previousCohort']['stopIndexSHA256']","prior['previousCohort']['previousCohort']['stopIndexSHA256']")
    replacement=replacement.replace("thirdCohort=prior['thirdCohort'],thirdCohortHistorySha256=prior['thirdCohortHistorySHA256']","thirdCohort=prior['previousCohort']['thirdCohort'],thirdCohortHistorySha256=prior['previousCohort']['thirdCohortHistorySHA256'],fourthCohort=prior['fourthCohort'],fourthCohortHistorySha256=prior['fourthCohortHistorySHA256'],spendingGuardSourceSha256=json.loads((out/'prepared.json').read_text())['spendingGuardSourceSHA256']")
    replacement=replacement.replace("else 1;report=","else 22 if name.startswith('glm-') else 1;report=")
    replacement=replacement.replace(";report=dict(gitCommit=commit",";count=21 if name.startswith('glm-') and (root/'.evidence/incomplete-glm').exists() else count;report=dict(gitCommit=commit")
    replacement=replacement.replace("elif name.startswith('mm-'):","elif name.startswith('glm-'):\n  report.update(mode='live',kind='glm-formal',route=name[4:],outcome='PASS',complete=True,auditPassed=True);leaf='formal-glm-report.json';dump(session/'live-report.json',dict(combo='zhipu-'+name[4:]));dump(session/'glm-integrity.json',{});dump(session/'glm-execution-source.json',{})\n elif name.startswith('mm-'):")
    go.write_text(text.replace(repr(producer),repr(replacement)));test.commit()


def approve(test,base):
    from test_recovery4_fixture import approve as previous
    # The fixture helper only writes into this disposable synthetic repository.
    import datetime
    auth=json.loads((base/'execution-authorization.template.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
    review.update(decision='APPROVED',reviewedBy='CONTROLLED independent',authorizedRoutes=list(json.loads((base/'plan.json').read_text())['jobs']))
    for field in ('standardsReview','specReview'):
        p=base/(field+'.md');p.write_text('CONTROLLED ordinary review');review[field]=dict(path=str(p.relative_to(test.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())
    human=json.loads((base/'recovery5-authorization.template.json').read_text());human.update(decision='APPROVED',requestText='CONTROLLED explicit19 US36',requestReference='CONTROLLED independent fifth reply')
    actual=test.root/BASE/'recovery5-authorization.json';actual.write_text(json.dumps(human));auth.update(decision='APPROVED',requestText=human['requestText'],requestReference=human['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(actual.read_bytes()).hexdigest())
    review['requestReference']=human['requestReference'];(base/'prerelease-source-review.json').write_text(json.dumps(review))
    for name in json.loads((base/'plan.json').read_text())['jobs']:
        if not name.startswith('formal-'):continue
        p=base/'grants'/(name+'.template.json');v=json.loads(p.read_text());v.update(id='ps-final-5-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat());dest=p.with_name(name+'.json');dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    for name in ('protocol114','visual-jev','visual-decisions'):
        leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json';v=json.loads((base/(leaf+'.template.json')).read_text());v.update(accountAlias='controlled-own-standard',regionAlias='controlled-own-region');dest=base/leaf;dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    (base/'execution-authorization.json').write_text(json.dumps(auth))


def keys(test):
    (test.root/'.env').rmdir();(test.root/'.env').write_text(''.join(k+'=CONTROLLED-ONLY\n' for k in ('OPENAI_KEY','GEMINI_KEY','ANTHROPIC_KEY','DEEPSEEK_KEY','TYPESAFE_KEY','MINIMAX_KEY','MINIMAX_CN_KEY','ZAI_KEY','ZHIPU_KEY')))
