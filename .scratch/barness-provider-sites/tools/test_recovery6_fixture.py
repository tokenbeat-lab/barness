"""Disposable closed5 originals and fixed6 executables, never real authority."""
import ast
import hashlib
import json
from pathlib import Path
from types import SimpleNamespace
from test_recovery5_fixture import seed as seed4,builder as builder5,approve as approve5,keys,BASE
OLD='provider-sites-final-acceptance-5';NEW='provider-sites-final-acceptance-6'
ROUTES=('chat','responses','messages')
NAMES=('formal-zai-responses',)+tuple('mm-global-formal-m3-'+v for v in ROUTES)+('formal-zai-chat','formal-zai-messages','formal-minimax-cn-chat','formal-minimax-cn-responses')

def seed(test):
    for name in ('ai/live/regional_harness_test.go','ai/internal/testkit/supportmatrix/regional_contract.go'):
        p=test.root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text('package controlled\n')
    test.commit();seed4(test);root=test.root;files={};commit=test.git('rev-parse','HEAD')
    def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
    def put(name,value,raw=False):
        p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(value if raw else (json.dumps(value)+'\n').encode());files[str(name)]=sha(p);return p
    from recovery_flow_source import SourceFreeze
    archive=put(BASE/'recovery5/source-at-start.json',dict(sourceCommit=commit,files=SourceFreeze(SimpleNamespace(ROOT=root)).files()));manifest=sha(archive)
    actual=put(BASE/'recovery5-authorization.json',dict(decision='APPROVED',round=OLD,requestText='CONTROLLED actual5 US36',requestReference='CONTROLLED independent fifth reply'))
    native=Path('.evidence/minimax-renewals')/OLD
    scope=dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId='provider-sites-final-acceptance-2',maxCalls=378,maxAttempts=414,campaignMaxCalls=1358,campaignMaxAttempts=1454,newReservedMicrousd=6900000,retainedReservedMicrousd=28800000,campaignMaximumMicrousd=36000000,actualHumanApprovalSHA256=sha(actual))
    hp=put(native/'human-approval.json',scope);review=put(native/'source-review.json',scope|dict(sourceCommit=commit,sourceManifestSHA256=manifest,binariesManifestSHA256='b'*64,planSHA256='p'*64))
    grant=put(native/'grant.json',dict(schema=6,decision='APPROVED',renewalId=OLD,sourceCommit=commit,executionSourceSha256=hashlib.sha256(b'{"controlled":"x"}').hexdigest(),humanApprovalSha256=sha(hp),sourceReviewSha256=sha(review),sourceManifestSha256=manifest,binariesManifestSha256='b'*64,planSha256='p'*64,maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000))
    put(native/'operator-authority.json',scope);put(native/'launch.lock',b'',True);rows=[];tail=''
    for route,api in zip(ROUTES,('openai-completions','openai-responses','anthropic-messages')):
        identity=dict(route=route,model='MiniMax-M3',purpose='standard',api=api,endpoint='https://api.minimax.io/anthropic' if route=='messages' else 'https://api.minimax.io/v1');session='old5-'+route
        put(native/(route+'.json'),dict(schema=2,renewalId=OLD,grantSha256=sha(grant),session=session,identity=identity,processNonce=hashlib.sha256(route.encode()).hexdigest()))
        for i in range(13):
            for kind in (('reserve','attempt') if i<12 else ('reserve',)):
                row=dict(renewalId=OLD,renewalGrantSha256=sha(grant),session=session,sequence=len(rows)+1,stage='formal',identity=identity,case='case-'+str(i),kind=kind,output=1 if kind=='reserve' else 0,input=2048 if kind=='reserve' else 0,at='2026-10-20T00:00:00Z',previous=tail,hash='');tail=hashlib.sha256(json.dumps(row,separators=(',',':')).encode()).hexdigest();row['hash']=tail;rows.append(row)
    put(native/'formal.jsonl',(''.join(json.dumps(v,separators=(',',':'))+'\n' for v in rows)).encode(),True)
    processes=[]
    for name in NAMES:
        mm=name.startswith('mm-');failed=name=='formal-minimax-cn-responses';calls=13 if mm else 18 if name.startswith('formal-zai-') else 9;attempts=12 if mm else calls;reserve=200000 if mm else 650000 if name.startswith('formal-zai-') else 350000
        budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        base=BASE/'recovery5'/name/'old5-session';leaf='minimax-report.json' if mm else 'regional-report.json'
        rp=put(base/leaf,dict(gitCommit=commit,executionSource={'controlled':'x'} if mm else {},scenarios=[dict(id=str(i),outcome='FAIL' if failed and i==8 else 'PASS') for i in range(calls)]))
        command=put(BASE/'recovery5'/(name+'-command.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,launcherAudit=dict(passed=True,findings=[])))
        consumer=put(BASE/'recovery5'/(name+'-consume-command.json'),dict(state='COMPLETED',exitCode=0,credentialsRead=False)) if not failed else None
        claim=put(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=commit,reservedMicrousd=reserve,budget=budget,commandReceiptSha256=sha(command),reportedFileSha256=sha(rp) if not failed else None,nativeConsumerReceiptSha256=sha(consumer) if consumer else None))
        processes.append(dict(name=name,canonicalClaim=str(claim.relative_to(root)),exitCode=1 if failed else 0,state='COMPLETED',parentBudget=budget,raw=str(base),nativeRecordedCalls=calls,nativeRecorderHTTP=attempts,reservedMicrousd=reserve))
    put(BASE/'recovery5/formal-minimax-cn-responses-usd.json',dict(reservedMicrousd=54238,attempts=9))
    total=sum((root/name).stat().st_size for name in files)
    link=root/BASE/'recovery5/no-key-live/controlled-refusal/seal-link.json';link.parent.mkdir(parents=True);link.symlink_to('seal.json')
    refusals={str(link.relative_to(root)):dict(kind='symlink',mode=link.lstat().st_mode,target='seal.json',followed=False)}
    stop=dict(nonRegularRejectionFixtures=refusals,schema=1,kind='ACTUAL_FIXED5_STOP_AFTER_MINIMAX_CN_RESPONSES_NATIVE_HISTORY_FAILURE',state='STOPPED_NO_FURTHER_PAID_LAUNCH',sourceCommit=commit,actualHumanApprovalSHA256=sha(actual),newPaidProcesses=8,passedProcesses=7,failedProcesses=1,newNativeRecordedCalls=111,newNativeRecorderHTTP=108,campaignNativeRecordedCalls=702,campaignNativeRecorderHTTP=692,campaignClaimedCalls=1160,campaignClaimedHTTP=1232,wholeReservedMicrousd=32050000,refundMicrousd=0,unknownParentProcesses=5,failedParentRemainsUNKNOWN=True,automaticReruns=False,paidContinuationAuthorized=False,processes=processes,notRunProcesses=['formal-minimax-cn-messages','glm-chat','glm-responses','glm-messages','openai-decisions','typesafe-classifier','protocol114','text-jev','text-luna','visual-jev','visual-decisions'],files=dict(files),fileCount=len(files),fileBytes=total)
    sp=put(BASE/'recovery5-stopped/index.json',stop)
    put(BASE/'closed-cohort-5/seal.json',dict(schema=1,kind='CLOSED_FIFTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(BASE/'recovery5-stopped/index.json'),stopIndexSHA256=sha(sp),actualCalls=702,actualHTTPAttempts=692,claimedCalls=1160,claimedHTTPAttempts=1232,wholeReservedMicrousd=32050000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False))
    path=test.tools/'recovery6_history.py';path.write_text(path.read_text().replace('5e8014e7405cfa86da21637b96e47cb294c4ff8e766e79440fe88ff56eef4f73',sha(sp)).replace('STOP_FILE_COUNT=31152','STOP_FILE_COUNT='+str(len(stop['files']))))
    path=test.tools/'recovery6_retained.py';path.write_text(path.read_text().replace('8c8342606390d37dff041f11de7ff3b8d82d12d5',commit))
    reviewed={}
    for name in ('ai/live/regional_harness_test.go','ai/internal/testkit/supportmatrix/regional_contract.go'):
        p=root/name;p.write_text(p.read_text()+'// CONTROLLED legitimate regional repair\n');reviewed[name]=sha(p)
    test.commit();current=test.git('rev-parse','HEAD')
    proof=json.loads((root/BASE/'decisions-budget-diagnosis-dev/compatibility-eight.json').read_text());proof.update(currentSourceCommit=current,reviewedSourceDifferences=reviewed)
    for i,row in enumerate(proof['cases']):
        row['requestHeadersBytesEqual']=True
        for phase in ('baseline','current'):
            wire=json.loads((root/row[phase+'WireFile']).read_text());wire[0]['requestHeaders']={'accept':'application/json'}
            target=put(BASE/'controlled6-compat'/phase/(str(i)+'-wire.json'),wire);row[phase+'WireFile']=str(target.relative_to(root));row[phase+'WireFileSHA256']=sha(target)
    dest=put(BASE/'recovery6-preparation/compatibility-eight.json',proof)
    path=test.tools/'recovery6_retained.py';path.write_text(path.read_text().replace('PENDING_FINAL_ROOT_84_REPROOF',sha(dest)));test.commit()

def builder(test):
    builder5(test);go=test.root/'bin/go';text=go.read_text();tree=ast.parse(text)
    producer=next(n.args[0].value for n in ast.walk(tree) if isinstance(n,ast.Call) and isinstance(n.func,ast.Attribute) and n.func.attr=='write_text' and isinstance(n.args[0],ast.Constant))
    replacement=producer.replace("new='provider-sites-final-acceptance-5'","new='provider-sites-final-acceptance-6'").replace('-recovery5-source','-recovery6-source').replace('-recovery5-preflight=','-recovery6-preflight=').replace("g['schema']==6","g['schema']==7")
    replacement=replacement.replace("prior['previousCohort']['previousCohort']['stopIndexSHA256']","prior['previousCohort']['previousCohort']['previousCohort']['stopIndexSHA256']").replace("prior['previousCohort']['secondCohortSealSHA256']","prior['previousCohort']['previousCohort']['secondCohortSealSHA256']")
    replacement=replacement.replace("prior['previousCohort']['thirdCohort']","prior['previousCohort']['previousCohort']['thirdCohort']").replace("prior['previousCohort']['thirdCohortHistorySHA256']","prior['previousCohort']['previousCohort']['thirdCohortHistorySHA256']")
    replacement=replacement.replace("fourthCohort=prior['fourthCohort'],fourthCohortHistorySha256=prior['fourthCohortHistorySHA256']","fourthCohort=prior['previousCohort']['fourthCohort'],fourthCohortHistorySha256=prior['previousCohort']['fourthCohortHistorySHA256'],fifthCohort=prior['fifthCohort'],fifthCohortHistorySha256=prior['fifthCohortHistorySHA256']")
    go.write_text(text.replace(repr(producer),repr(replacement)));test.commit()

def approve(test,base):
    import datetime
    auth=json.loads((base/'execution-authorization.template.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
    review.update(decision='APPROVED',reviewedBy='CONTROLLED independent',authorizedRoutes=list(json.loads((base/'plan.json').read_text())['jobs']))
    for field in ('standardsReview','specReview'):
        p=base/(field+'.md');p.write_text('CONTROLLED review');review[field]=dict(path=str(p.relative_to(test.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())
    human=json.loads((base/'recovery6-authorization.template.json').read_text());human.update(decision='APPROVED',requestText='CONTROLLED explicit19 US39',requestReference='CONTROLLED independent sixth reply')
    actual=test.root/BASE/'recovery6-authorization.json';actual.write_text(json.dumps(human));auth.update(decision='APPROVED',requestText=human['requestText'],requestReference=human['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(actual.read_bytes()).hexdigest());review['requestReference']=human['requestReference'];(base/'prerelease-source-review.json').write_text(json.dumps(review))
    for name in json.loads((base/'plan.json').read_text())['jobs']:
        if not name.startswith('formal-'):continue
        p=base/'grants'/(name+'.template.json');v=json.loads(p.read_text());v.update(id='ps-final-6-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat());dest=p.with_name(name+'.json');dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    for name in ('protocol114','visual-jev','visual-decisions'):
        leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json';v=json.loads((base/(leaf+'.template.json')).read_text());v.update(accountAlias='controlled-own-standard',regionAlias='controlled-own-region');dest=base/leaf;dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    (base/'execution-authorization.json').write_text(json.dumps(auth))
