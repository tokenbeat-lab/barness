"""Disposable interrupted sixth originals and fixed7 no-socket executables."""
import ast
import hashlib
import json
from pathlib import Path
from types import SimpleNamespace
from test_recovery6_fixture import seed as seed5,builder as builder6,keys,BASE
OLD='provider-sites-final-acceptance-6';NEW='provider-sites-final-acceptance-7'
ROUTES=('chat','responses','messages')
M3_FILES=('ai/live/minimax_harness_test.go','ai/live/minimax_capture_test.go','ai/live/minimax_authorization_deadline_test.go')

def seed(test):
    for name in M3_FILES:
        p=test.root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text('package controlled\n')
    test.commit();seed5(test);root=test.root;files={};commit=test.git('rev-parse','HEAD')
    def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
    def put(name,value,raw=False):
        p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(value if raw else (json.dumps(value)+'\n').encode());files[str(name)]=sha(p);return p
    from recovery_flow_source import SourceFreeze
    archive=put(BASE/'recovery6/source-at-start.json',dict(sourceCommit=commit,files=SourceFreeze(SimpleNamespace(ROOT=root)).files()));manifest=sha(archive)
    actual=put(BASE/'recovery6-authorization.json',dict(decision='APPROVED',round=OLD,requestText='CONTROLLED actual6 US39',requestReference='CONTROLLED independent sixth reply'))
    native=Path('.evidence/minimax-renewals')/OLD
    scope=dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId='provider-sites-final-acceptance-2',maxCalls=378,maxAttempts=414,campaignMaxCalls=1538,campaignMaxAttempts=1646,newReservedMicrousd=6900000,retainedReservedMicrousd=32050000,campaignMaximumMicrousd=39000000,actualHumanApprovalSHA256=sha(actual))
    hp=put(native/'human-approval.json',scope);review=put(native/'source-review.json',scope|dict(sourceCommit=commit,sourceManifestSHA256=manifest,binariesManifestSHA256='b'*64,planSHA256='p'*64))
    grant=put(native/'grant.json',dict(schema=7,decision='APPROVED',renewalId=OLD,sourceCommit=commit,humanApprovalSha256=sha(hp),sourceReviewSha256=sha(review),sourceManifestSha256=manifest,binariesManifestSha256='b'*64,planSha256='p'*64,maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000))
    put(native/'operator-authority.json',scope);put(native/'launch.lock',b'',True);rows=[];tail='';identity=dict(route='chat',model='MiniMax-M3',purpose='standard',api='openai-completions',endpoint='https://api.minimax.io/v1');session='old6-chat'
    put(native/'chat.json',dict(schema=2,renewalId=OLD,grantSha256=sha(grant),session=session,identity=identity,processNonce='c'*64))
    for i in range(4):
        for kind in ('reserve','attempt'):
            row=dict(renewalId=OLD,renewalGrantSha256=sha(grant),session=session,sequence=len(rows)+1,stage='formal',identity=identity,case='case-'+str(i),kind=kind,output=1 if kind=='reserve' else 0,input=2048 if kind=='reserve' else 0,at='2026-10-21T00:00:00Z',previous=tail,hash='');tail=hashlib.sha256(json.dumps(row,separators=(',',':')).encode()).hexdigest();row['hash']=tail;rows.append(row)
    put(native/'formal.jsonl',(''.join(json.dumps(v,separators=(',',':'))+'\n' for v in rows)).encode(),True)
    processes=[];unknown=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
    for name in ('formal-zai-responses','mm-global-formal-m3-chat'):
        mm=name.startswith('mm-');budget=unknown if mm else dict(state='REPORTED',status='PASS',calls=18,attempts=18);reserve=200000 if mm else 650000;base=BASE/'recovery6'/name/'old6-session'
        report=None if mm else put(base/'regional-report.json',dict(gitCommit=commit,expected=list(range(18)),scenarios=[dict(outcome='PASS') for _ in range(18)]))
        command=put(BASE/'recovery6'/(name+'-command.json'),dict(state='COMPLETED',exitCode=-15 if mm else 0,launcherAudit=dict(passed=True,findings=[])))
        consumer=None if mm else put(BASE/'recovery6'/(name+'-consume-command.json'),dict(state='COMPLETED',exitCode=0,credentialsRead=False))
        value=dict(state='COMPLETED',exitCode=-15 if mm else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=commit,reservedMicrousd=reserve,budget=budget,commandReceiptSha256=sha(command))
        if not mm:value.update(reportedFileSha256=sha(report),nativeConsumerReceiptSha256=sha(consumer))
        claim=put(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'),value)
        processes.append(dict(name=name,parent=str(claim.relative_to(root)),parentSHA256=sha(claim),parentBudget=budget,exitCode=value['exitCode'],reservedMicrousd=reserve,nativeConsumerReceiptSHA256=sha(consumer) if consumer else None))
    raw=BASE/'recovery6/mm-global-formal-m3-chat/old6-session';case=raw/'MM07-full-stream-rerun-1';put(case/'result.json',dict(errorCode='deadline_exceeded',errorPhase='request'))
    put(BASE/'recovery6/mm-global-formal-m3-chat-usd.json',dict(reservedMicrousd=12884,attempts=4))
    total=sum((root/name).stat().st_size for name in files);link=root/BASE/'recovery6/no-key-live/controlled-refusal/seal-link.json';link.parent.mkdir(parents=True);link.symlink_to('seal.json')
    refusals={str(link.relative_to(root)):dict(kind='symlink',mode=link.lstat().st_mode,target='seal.json',followed=False)}
    rest=['mm-global-formal-m3-responses','mm-global-formal-m3-messages','formal-zai-chat','formal-zai-messages','formal-minimax-cn-chat','formal-minimax-cn-responses','formal-minimax-cn-messages','glm-chat','glm-responses','glm-messages','openai-decisions','typesafe-classifier','protocol114','text-jev','text-luna','visual-jev','visual-decisions']
    stop=dict(schema=1,kind='ACTUAL_FIXED6_STOP_AFTER_M3_CHAT_FINAL_REQUEST_DEADLINE_CASE_FAILURE',state='STOPPED_NO_FURTHER_PAID_LAUNCH',sourceCommit=commit,newPaidProcesses=2,passedProcesses=1,interruptedAfterFailedCaseProcesses=1,completedFullM3Processes=0,m3FinalReport='ABSENT_INTERRUPTED_AFTER_IRREVERSIBLE_CASE_FAILURE',m3Raw=str(raw),failedFinalCase=str(case),failure=dict(case='full-stream-rerun-1',status='FAIL',errorCode='deadline_exceeded',errorPhase='request'),actualHumanApprovalSHA256=sha(actual),newCompletedReportCalls=18,newCompletedReportHTTP=18,campaignCompletedReportKnownCalls=720,campaignCompletedReportKnownHTTP=710,interruptedM3NativeLedgerCalls=4,interruptedM3NativeLedgerHTTPAllocations=4,interruptedM3NativeLedgerRows=8,interruptedM3NativeLedgerTail=tail,interruptedM3USDReservationMicrousd=12884,interruptedM3USDTransportReservations=4,interruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends=True,campaignClaimedCalls=1204,campaignClaimedHTTPAttempts=1280,wholeReservedMicrousd=32900000,refundMicrousd=0,unknownParentProcesses=6,failedParentRemainsUNKNOWN=True,stopSignal='SIGTERM',childNumericExitCode=-15,publicOperatorNumericExitCode=1,automaticReruns=False,paidContinuationAuthorized=False,processes=processes,notRunProcesses=rest,files=dict(files),fileCount=len(files),fileBytes=total,nonRegularRejectionFixtures=refusals)
    sp=put(BASE/'recovery6-stopped/index.json',stop)
    put(BASE/'closed-cohort-6/seal.json',dict(schema=1,kind='CLOSED_SIXTH_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(BASE/'recovery6-stopped/index.json'),stopIndexSHA256=sha(sp),knownRecordedCalls=724,knownRecordedHTTPAllocations=714,completedReportKnownCalls=720,completedReportKnownHTTP=710,interruptedNativeLedgerCalls=4,interruptedNativeLedgerHTTPAllocations=4,interruptedParentBudget='UNKNOWN',interruptedFinalReportAbsent=True,recordedHTTPAllocationsAreNotConfirmedVendorSends=True,claimedCalls=1204,claimedHTTPAttempts=1280,wholeReservedMicrousd=32900000,unknownParentProcesses=6,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False))
    path=test.tools/'recovery7_history.py';path.write_text(path.read_text().replace('fcb781794736eb8c8cc83582d757ed16e0db5091661690cab51f517bcdbad81e',sha(sp)).replace('STOP_FILE_COUNT=30238','STOP_FILE_COUNT='+str(len(stop['files']))))
    baseline=json.loads((root/BASE/'recovery5/source-at-start.json').read_text());path=test.tools/'recovery7_retained.py';path.write_text(path.read_text().replace('8c8342606390d37dff041f11de7ff3b8d82d12d5',baseline['sourceCommit']))
    for name in M3_FILES:
        p=root/name;p.write_text(p.read_text()+'// CONTROLLED own authorization deadline repair\n')
    test.commit();current=test.git('rev-parse','HEAD');sources=SourceFreeze(SimpleNamespace(ROOT=root)).files();reviewed={n:sha for n,sha in sources.items() if n.startswith('ai/') and baseline['files'].get(n)!=sha}
    proof=json.loads((root/BASE/'recovery6-preparation/compatibility-eight.json').read_text());proof.update(currentSourceCommit=current,reviewedSourceDifferences=reviewed)
    dest=put(BASE/'recovery7-preparation/compatibility-eight.json',proof);path=test.tools/'recovery7_retained.py';path.write_text(path.read_text().replace('PENDING_CURRENT_GO_84_REPROOF',sha(dest)));test.commit()

def builder(test):
    builder6(test);go=test.root/'bin/go';text=go.read_text();tree=ast.parse(text)
    producer=next(n.args[0].value for n in ast.walk(tree) if isinstance(n,ast.Call) and isinstance(n.func,ast.Attribute) and n.func.attr=='write_text' and isinstance(n.args[0],ast.Constant))
    replacement=producer.replace("new='provider-sites-final-acceptance-6'","new='provider-sites-final-acceptance-7'").replace('-recovery6-source','-recovery7-source').replace('-recovery6-preflight=','-recovery7-preflight=').replace("g['schema']==7","g['schema']==8")
    for key in ('stopIndexSHA256','secondCohortSealSHA256','thirdCohort','thirdCohortHistorySHA256','fourthCohort','fourthCohortHistorySHA256','fifthCohort','fifthCohortHistorySHA256'):
        import re
        replacement=re.sub(r"prior((?:\['previousCohort'\])*)\['"+key+r"'\]",lambda m:"prior['previousCohort']"+m[1]+"['"+key+"']",replacement)
    replacement=replacement.replace("fifthCohortHistorySha256=prior['previousCohort']['fifthCohortHistorySHA256']","fifthCohortHistorySha256=prior['previousCohort']['fifthCohortHistorySHA256'],sixthCohort=prior['sixthCohort'],sixthCohortHistorySha256=prior['sixthCohortHistorySHA256']")
    go.write_text(text.replace(repr(producer),repr(replacement)));test.commit()

def approve(test,base):
    import datetime
    auth=json.loads((base/'execution-authorization.template.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
    review.update(decision='APPROVED',reviewedBy='CONTROLLED independent',authorizedRoutes=list(json.loads((base/'plan.json').read_text())['jobs']))
    for field in ('standardsReview','specReview'):
        p=base/(field+'.md');p.write_text('CONTROLLED review');review[field]=dict(path=str(p.relative_to(test.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())
    human=json.loads((base/'recovery7-authorization.template.json').read_text());human.update(decision='APPROVED',requestText='CONTROLLED explicit19 US40',requestReference='CONTROLLED independent seventh reply')
    actual=test.root/BASE/'recovery7-authorization.json';actual.write_text(json.dumps(human));auth.update(decision='APPROVED',requestText=human['requestText'],requestReference=human['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(actual.read_bytes()).hexdigest());review['requestReference']=human['requestReference'];(base/'prerelease-source-review.json').write_text(json.dumps(review))
    for name in json.loads((base/'plan.json').read_text())['jobs']:
        if not name.startswith('formal-'):continue
        p=base/'grants'/(name+'.template.json');v=json.loads(p.read_text());v.update(id='ps-final-7-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat());dest=p.with_name(name+'.json');dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    for name in ('protocol114','visual-jev','visual-decisions'):
        leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json';v=json.loads((base/(leaf+'.template.json')).read_text());v.update(accountAlias='controlled-own-standard',regionAlias='controlled-own-region');dest=base/leaf;dest.write_text(json.dumps(v));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    (base/'execution-authorization.json').write_text(json.dumps(auth))
