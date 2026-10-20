"""Controlled closed3 originals and executable native contracts; no vendor sockets."""
import ast
import hashlib
import json
from pathlib import Path
from test_recovery_fixture import seed as seed2, builder as builder3, BASE
OLD='provider-sites-final-acceptance-3'
NEW='provider-sites-final-acceptance-4'
NAMES=('mm-global-formal-m3-chat','mm-global-formal-m3-responses','mm-global-formal-m3-messages','formal-zai-chat','formal-zai-responses')

def seed(test):
    seed2(test);root=test.root;files={};commit=test.git('rev-parse','HEAD')
    def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
    def put(name,value,raw=False):
        p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(value if raw else (json.dumps(value)+'\n').encode());files[str(name)]=sha(p);return p
    native=Path('.evidence/minimax-renewals')/OLD
    archive=put(BASE/'recovery3/source-at-start.json',dict(sourceCommit=commit,files={'ai/sentinel.go':sha(root/'ai/sentinel.go')}));manifest=sha(archive)
    human=put(BASE/'recovery3-authorization.json',dict(decision='APPROVED',round=OLD,requestText='controlled actual3 reply',requestReference='controlled actual3'))
    old_scope=dict(decision='APPROVED',renewalId=OLD,validationRound=OLD,nativeRenewalId=OLD,spendingGuardPolicyId='provider-sites-final-acceptance-2',maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,campaignMaximumMicrousd=30000000,actualHumanApprovalSHA256=sha(human))
    hp=put(native/'human-approval.json',old_scope)
    review=put(native/'source-review.json',old_scope|dict(sourceCommit=commit,sourceManifestSHA256=manifest,binariesManifestSHA256='b'*64,planSHA256='p'*64))
    grant=put(native/'grant.json',dict(schema=4,decision='APPROVED',renewalId=OLD,sourceCommit=commit,executionSourceSha256=hashlib.sha256(json.dumps({'controlled':'x'},sort_keys=True,separators=(',',':')).encode()).hexdigest(),humanApprovalSha256=sha(hp),sourceReviewSha256=sha(review),sourceManifestSha256=manifest,binariesManifestSha256='b'*64,planSha256='p'*64,maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000))
    put(native/'operator-authority.json',old_scope);put(native/'launch.lock',b'',True)
    ledger=[];previous='';sessions={};identities={}
    for route,api in [('chat','openai-completions'),('responses','openai-responses'),('messages','anthropic-messages')]:
        session='old3-'+route;sessions[route]=session;identity=dict(route=route,model='MiniMax-M3',purpose='standard',api=api,endpoint='https://api.minimax.io/anthropic' if route=='messages' else 'https://api.minimax.io/v1');identities[route]=identity
        put(native/(route+'.json'),dict(schema=2,renewalId=OLD,grantSha256=sha(grant),session=session,identity=identity,processNonce=hashlib.sha256(route.encode()).hexdigest()))
        for i in range(13):
            for kind in (('reserve','attempt') if i<12 else ('reserve',)):
                row=dict(renewalId=OLD,renewalGrantSha256=sha(grant),session=session,sequence=len(ledger)+1,stage='formal',identity=identity,case='case-'+str(i),kind=kind,output=1 if kind=='reserve' else 0,input=2048 if kind=='reserve' else 0,at='2026-10-20T00:00:00Z',previous=previous,hash='')
                row['hash']=hashlib.sha256(json.dumps(row,separators=(',',':')).encode()).hexdigest();previous=row['hash'];ledger.append(row)
    put(native/'formal.jsonl',(''.join(json.dumps(v,separators=(',',':'))+'\n' for v in ledger)).encode(),True)
    processes=[]
    for name in NAMES:
        mm=name.startswith('mm-');failed=name=='formal-zai-responses';calls=13 if mm else 18;attempts=12 if mm else 18;reserve=200000 if mm else 650000
        budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        base=BASE/'recovery3'/name/('old3-'+name.removeprefix('mm-global-formal-m3-') if mm else 'old3-session')
        report=dict(gitCommit=commit,budget=dict(calls=calls,attempts=attempts) if mm else dict(callsUsed=calls,httpAttempts=attempts),expected=[str(i) for i in range(calls)],calls=[dict(id=str(i),outcome='PASS') for i in range(calls)] if mm else [],scenarios=[dict(id=str(i),outcome='FAIL' if failed and i==0 else 'PASS') for i in range(calls)])
        leaf='minimax-report.json' if mm else 'regional-report.json'
        if mm:report.update(formalRenewalId=OLD,formalRenewalGrantSha256=sha(grant),identity=identities[name.rsplit('-',1)[-1]],session=sessions[name.rsplit('-',1)[-1]],outcome='PASS',complete=True,auditPassed=True,executionSource={'controlled':'x'},budget=dict(calls=13,attempts=12))
        rp=put(base/leaf,report);put(base/'audit.json',dict(findings=[]))
        command=put(BASE/'recovery3'/(name+'-command.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,launcherAudit=dict(passed=True,findings=[])))
        consumer=put(BASE/'recovery3'/(name+'-consume-command.json'),dict(state='COMPLETED',exitCode=0,credentialsRead=False)) if not failed else None
        claim=put(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,baseName=name,validationRound=OLD,nativeRenewalId=OLD,sourceCommit=commit,reservedMicrousd=reserve,maxCalls=20 if mm else 24,maxAttempts=24,budget=budget,reportedFileSha256=sha(rp) if not failed else None,commandReceiptSha256=sha(command),nativeConsumerReceiptSha256=sha(consumer) if consumer else None,sourceManifestSha256=manifest,binariesManifestSha256='b'*64,planSha256='p'*64))
        processes.append(dict(name=name,exitCode=1 if failed else 0,parentBudget=budget,reservedMicrousd=reserve,nativeBundle=str(base),nativeReportSHA256=sha(rp),actualCalls=calls,actualHTTP=attempts,nativeOutcome='PASS' if mm else None))
    for i in range(1012-len(files)):put(BASE/'recovery3-stopped'/('original-'+str(i)+'.json'),dict(controlled=i))
    all_jobs=['mm-global-formal-m3-'+r for r in ('chat','responses','messages')]+['formal-'+p+'-'+r for p in ('zai','minimax-cn') for r in ('chat','responses','messages')]+['openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','protocol114','text-jev','text-luna','visual-jev','visual-decisions']
    stop=dict(schema=1,kind='ACTUAL_RECOVERY3_STOP_AFTER_ZAI_RESPONSES_NATIVE_SOURCE_FAILURE',sourceCommit=commit,actualHumanApprovalSHA256=sha(human),successfulProcesses=4,failedProcesses=1,newActualCalls=75,newActualHTTP=72,campaignActualCalls=378,campaignActualHTTP=374,campaignClaimedCalls=564,campaignClaimedHTTP=612,newWholeReservedMicrousd=1900000,wholeReservedMicrousd=12600000,refundMicrousd=0,unknownParentProcesses=3,originalFailedGLMUnchanged=True,originalFailedM3Unchanged=True,automaticReruns=False,unstartedProcesses=[n for n in all_jobs if n not in NAMES],files=dict(files),processes=processes)
    sp=put(BASE/'recovery3-stopped/index.json',stop)
    put(BASE/'closed-cohort-3/seal.json',dict(schema=1,kind='CLOSED_THIRD_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(BASE/'recovery3-stopped/index.json'),stopIndexSHA256=sha(sp),actualCalls=378,actualHTTPAttempts=374,claimedCalls=564,claimedHTTPAttempts=612,wholeReservedMicrousd=12600000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False))
    test.commit()

def builder(test):
    builder3(test);go=test.root/'bin/go';text=go.read_text();tree=ast.parse(text)
    producer=next(n.args[0].value for n in ast.walk(tree) if isinstance(n,ast.Call) and isinstance(n.func,ast.Attribute) and n.func.attr=='write_text' and isinstance(n.args[0],ast.Constant))
    replacement=producer.replace("new='provider-sites-final-acceptance-3'","new='provider-sites-final-acceptance-4'").replace('-recovery-source','-recovery4-source').replace('-recovery-preflight=','-recovery4-preflight=').replace("g['schema']==4","g['schema']==5")
    replacement=replacement.replace("prior['stopIndexSHA256']","prior['previousCohort']['stopIndexSHA256']")
    replacement=replacement.replace("if '-recovery4-source' in a:", "if '-recovery4-source' in a:\n archive=root/'.evidence/provider-sites/final-acceptance-20261020/recovery3/source-at-start.json';oldgrant=json.loads((root/'.evidence/minimax-renewals/provider-sites-final-acceptance-3/grant.json').read_text());assert sha(archive)==oldgrant['sourceManifestSha256'],'CONTROLLED_NATIVE_ARCHIVE_REJECT'")
    replacement=replacement.replace("recovery=recovery,recoveryHistorySha256=sha(native/'formal.jsonl')","recovery=recovery,recoveryHistorySha256=sha(native/'formal.jsonl'),thirdCohort=prior['thirdCohort'],thirdCohortHistorySha256=prior['thirdCohortHistorySHA256']")
    replacement=replacement.replace("commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();report=", "commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();count=18 if name.startswith('formal-zai-') else 9 if name.startswith('formal-minimax-cn-') else 1;report=")
    replacement=replacement.replace("budget=dict(callsUsed=1,httpAttempts=1),expected=['one'],scenarios=[dict(id='one',outcome='PASS')]", "budget=dict(callsUsed=count,httpAttempts=count),expected=[str(i) for i in range(count)],scenarios=[dict(id=str(i),outcome='PASS') for i in range(count)]")
    replacement=replacement.replace("attempts=12 if name.startswith('mm-') else 1", "attempts=12 if name.startswith('mm-') else count")
    go.write_text(text.replace(repr(producer),repr(replacement)));test.commit()


def approve(test,base):
    import datetime
    auth=json.loads((base/'execution-authorization.template.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
    review.update(decision='APPROVED',reviewedBy='controlled independent',authorizedRoutes=list(json.loads((base/'plan.json').read_text())['jobs']))
    for field in ('standardsReview','specReview'):
        p=base/(field+'.md');p.write_text('Controlled independent ordinary source decision.');review[field]=dict(path=str(p.relative_to(test.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())
    (base/'prerelease-source-review.json').write_text(json.dumps(review))
    human=json.loads((base/'recovery4-authorization.template.json').read_text());human.update(decision='APPROVED',requestText='CONTROLLED explicitly approve new4 complete24/506/518 US30',requestReference='CONTROLLED direct new4 reply')
    actual=test.root/BASE/'recovery4-authorization.json';actual.write_text(json.dumps(human))
    auth.update(decision='APPROVED',requestText=human['requestText'],requestReference=human['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(actual.read_bytes()).hexdigest())
    for name in json.loads((base/'plan.json').read_text())['jobs']:
        if not name.startswith('formal-'):continue
        p=base/'grants'/(name+'.template.json');grant=json.loads(p.read_text());grant.update(id='ps-final-4-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat());dest=p.with_name(name+'.json');dest.write_text(json.dumps(grant));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    for name in ('protocol114','visual-jev','visual-decisions'):
        leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json';value=json.loads((base/(leaf+'.template.json')).read_text());value.update(accountAlias='controlled-own-standard',regionAlias='controlled-own-region');dest=base/leaf;dest.write_text(json.dumps(value));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
    (base/'execution-authorization.json').write_text(json.dumps(auth))

def keys(test):
    (test.root/'.env').rmdir();(test.root/'.env').write_text(''.join(k+'=CONTROLLED-ONLY\n' for k in ('OPENAI_KEY','GEMINI_KEY','ANTHROPIC_KEY','DEEPSEEK_KEY','TYPESAFE_KEY','MINIMAX_KEY','MINIMAX_CN_KEY','ZAI_KEY')))


def reseal(test,changed):
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    stop_path=test.root/BASE/'recovery3-stopped/index.json';stop=json.loads(stop_path.read_text())
    for path in changed:stop['files'][str(path.relative_to(test.root))]=sha(path)
    stop_path.write_text(json.dumps(stop));seal_path=test.root/BASE/'closed-cohort-3/seal.json';seal=json.loads(seal_path.read_text());seal['stopIndexSHA256']=sha(stop_path);seal_path.write_text(json.dumps(seal))
