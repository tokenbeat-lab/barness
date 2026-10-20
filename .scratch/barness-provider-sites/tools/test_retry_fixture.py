"""Independent synthetic stopped-prefix facts; never authenticates live evidence."""
import hashlib,json,os,shutil,subprocess,sys
from pathlib import Path
import test_final_publication
import test_renewal
BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
ROUND='provider-sites-final-acceptance-2'
def seed(t,count=22):
    def put(path,value):
        path=t.root/path;path.parent.mkdir(parents=True,exist_ok=True);path.write_text(json.dumps(value));return digest(path)
    def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()
    t.put=lambda path,value:put(Path(path),value)
    t.pin=hashlib.sha256(b'[]\n').hexdigest()
    put(Path('ai/release/catalog-snapshot.json'),dict(hash='sha256:current',catalog=dict(version='current')))
    put(Path('ai/catalog_minimax.json'),{});put(Path('ai/catalog_regional.json'),dict(active=True))
    p=t.root/'ai/live/testdata/glm-candidate/sources.json';p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(b'[]\n')
    p=t.root/'ai/internal/testkit/supportmatrix/glm.go';p.parent.mkdir(parents=True,exist_ok=True);p.write_text('package supportmatrix\nconst GLMSourcePin = "'+t.pin+'"\n')
    put(Path('ai/live/support-matrix.json'),dict(schema=1,rows=[]))
    for name in ('live-captures.json','known-usage.json','support-views.json'):put(Path('ai/live/testdata/glm-formal')/name,{})
    (t.root/'ai/live/testdata/glm-formal/live-captures.sha256').write_text('unpublished\n')
    t.commit();commit=t.git('rev-parse','HEAD')
    t.captures=lambda:test_renewal.Renewal.captures(t)
    pub=test_final_publication.FinalPublication();pub.fixture=t;pub.root=t.root;pub.seed_new()
    old=t.root/'.evidence/provider-sites/glm-native';files={}
    source={n:digest(t.root/n) for n in ['ai/sentinel.go','go.mod','go.sum']}
    scanner='''import hashlib,json,pathlib,os,sys,subprocess,shutil
root=pathlib.Path.cwd()
while not (root/'go.mod').is_file():root=root.parent
source={n:hashlib.sha256((root/n).read_bytes()).hexdigest() for n in ['ai/sentinel.go','go.mod','go.sum']}
if '-source' in sys.argv:print(json.dumps(dict(files=source)))
'''
    formal=old/'glmformal';formal.write_text('#!'+sys.executable+'\n'+scanner);formal.chmod(0o755)
    producer=scanner+'''if os.getenv('BARNESS_GLM_KEY'):
 route=os.environ['BARNESS_GLM_ROUTE'];target=pathlib.Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session'
 template=pathlib.Path(TEMPLATE)/route/'session';shutil.copytree(template,target)
 report=json.loads((target/'formal-glm-report.json').read_text());report.update(gitCommit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),budget=dict(callsUsed=22,httpAttempts=22));report['scenarios']=report['scenarios'][:COUNT]
 (target/'formal-glm-report.json').write_text(json.dumps(report));(target/'glm-execution-source.json').write_text(json.dumps(source))
 usd=pathlib.Path(os.environ['BARNESS_FINAL_BUDGET']);value=json.loads(usd.read_text());value.update(reservedMicrousd=414000,attempts=22);usd.write_text(json.dumps(value))
'''
    producer=producer.replace('TEMPLATE',repr(str(t.root/'raw'))).replace('COUNT',str(count))
    live=old/'live.test';live.write_text('#!'+sys.executable+'\n'+producer);live.chmod(0o755)
    put(old.relative_to(t.root)/'binaries.json',{'live.test':digest(live),'glmformal':digest(formal)})
    prepared=json.loads((old/'prepared.json').read_text());prepared['binariesManifestSHA256']=digest(old/'binaries.json')
    snapshot=json.loads((old/'source-at-start.json').read_text());snapshot['files']=source|{n:digest(t.root/n) for n in ['ai/catalog_minimax.json','ai/catalog_regional.json','ai/release/catalog-snapshot.json']};put(old.relative_to(t.root)/'source-at-start.json',snapshot);prepared['sourceManifestSHA256']=digest(old/'source-at-start.json');put(old.relative_to(t.root)/'prepared.json',prepared)
    for leaf in ['execution-authorization.json','prerelease-source-review.json']:
        d=json.loads((old/leaf).read_text());d.update(prepared);put(old.relative_to(t.root)/leaf,d)
    prior=json.loads((old/'predecessor-disclosure.json').read_text())
    put(BASE/'round2-authorization.json',dict(decision='APPROVED',roundId=ROUND))
    identity=dict(round=ROUND,requestText='Synthetic approve new fixed722/758 US30 standard only',requestReference='controlled test reply',maxMicrousd=30000000,newMaxMicrousd=28400000,maxCalls=722,maxAttempts=758,campaignMaxCalls=774,campaignMaxAttempts=810,eligibilityAuthorizationSHA256=prior['authorizationSHA256'],raceAuthorizationSHA256=prior['authorizationSHA256'],automaticProcessReruns=False,predecessor=prior)
    put(Path('.evidence/minimax-renewals')/ROUND/'operator-authority.json',identity)
    (t.root/'.evidence/minimax-renewals'/ROUND/'launch.lock').write_bytes(b'')
    processes=[]
    for route in ['chat','responses']:
        base=old/('glm-'+route)/'session';put(base.relative_to(t.root)/'glm-execution-source.json',source)
        marker=t.root/'.evidence/minimax-launches'/(ROUND+'-glm-'+route+'.json');d=json.loads(marker.read_text());d.update(reservedMicrousd=1000000,sourceManifestSha256=prepared['sourceManifestSHA256'],binariesManifestSha256=prepared['binariesManifestSHA256'],sourceReviewSha256=digest(old/'prerelease-source-review.json'),authorizationSha256=digest(old/'execution-authorization.json'))
        if route=='responses':
            report=json.loads((base/'formal-glm-report.json').read_text());report['outcome']='FAIL';report['scenarios'][0]['outcome']='FAIL';put(base.relative_to(t.root)/'formal-glm-report.json',report)
            command=json.loads((old/'glm-responses-command.json').read_text());command['exitCode']=1;put(old.relative_to(t.root)/'glm-responses-command.json',command)
            d.update(exitCode=1,budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None))
        d['capturePins']={'schema':1,**{k:digest(base/leaf) for k,leaf in [('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')]}}
        put(marker.relative_to(t.root),d)
        processes.append(dict(name='glm-'+route,nativeBundle=str(base.relative_to(t.root)),actualCalls=22,actualHTTPAttempts=22,claimedCalls=24,claimedAttempts=32,wholeReservedMicrousd=1000000,exitCode=d['exitCode']))
    (t.root/'.evidence/minimax-launches'/(ROUND+'-glm-messages.json')).unlink();shutil.rmtree(old/'glm-messages')
    for provider in ['zai','minimax-cn']:
        for route in ['chat','responses','messages']:
            name='candidate-'+provider+'-'+route;reserve=650000 if provider=='zai' else 350000
            put(Path('.evidence/minimax-launches')/(ROUND+'-'+name+'.json'),dict(state='COMPLETED',validationRound=ROUND,baseName=name,exitCode=0,maxCalls=24,maxAttempts=24,reservedMicrousd=reserve,budget=dict(state='REPORTED',status='PASS',calls=18 if provider=='zai' else 9,attempts=18 if provider=='zai' else 9),nativeConsumerReceiptSha256='a'*64))
            processes.append(dict(name=name))
    for path in [*old.rglob('*'),*(t.root/'.evidence/minimax-launches').glob(ROUND+'-*.json'),t.root/'.evidence/minimax-renewals'/ROUND/'operator-authority.json',t.root/BASE/'round2-authorization.json']:
        if path.is_file():files[str(path.relative_to(t.root))]=digest(path)
    stopped=dict(schema=1,kind='STOPPED_ACTUAL_ROUND2_AFTER_REAL_GLM_RESPONSES_FAILURE',roundId=ROUND,state='STOPPED_NO_AUTOMATIC_PAID_CONTINUATION',sourceCommit=commit,processes=processes,completedProcesses=8,newActualCalls=125,newActualHTTPAttempts=125,newClaimedCalls=192,newClaimedHTTPAttempts=208,newWholeReservedMicrousd=5000000,campaignActualCalls=165,campaignActualHTTPAttempts=165,campaignClaimedCalls=244,campaignClaimedHTTPAttempts=260,campaignWholeReservedMicrousd=6500000,refundMicrousd=0,failedParentUNKNOWNUnchanged=True,nativeM3GrantNeverIssued=True,files=files)
    stopped_pin=put(BASE/'round2-stopped/index.json',stopped)
    failed=Path('.evidence/minimax-launches')/(ROUND+'-glm-responses.json')
    approval=dict(schema=1,kind='ACTUAL_HUMAN_ONE_EXPLICIT_GLM_RESPONSES_RETRY_AND_CONTINUATION',decision='APPROVED',humanText='Synthetic approve exactly one whole22 retry and original remaining scope within US30',humanMessageReference='synthetic direct reply',scope=dict(provider='zhipu',api='openai-responses',models=['glm-5.3','glm-5.2'],completeScenarios=22,additionalCompleteProcesses=1,retainOriginalFailure=True,continueOriginalRemainingProcessesOnSuccess=True,finalFullRaceOnFreshREADY=True,maxCampaignUSD=30,automaticAdditionalProcessReruns=False),derivedNativeProcessBounds=dict(maxCalls=24,maxHTTPAttempts=32),derivedCampaignCeilings=dict(maxCalls=798,maxHTTPAttempts=842),operatorConservativeAllocations=dict(retainPrefixWholeMicrousd=6500000,retryMicrousd=500000,unstartedGLMMessagesMicrousd=500000,allOtherOriginalReservationsUnchanged=True,totalWholeMicrousd=29900000,knownRateOperationalCeilingsNotInvoices=True),originalR2ApprovalSHA256=digest(t.root/BASE/'round2-authorization.json'),originalEligibilityAndFinalRaceApprovalSHA256=prior['authorizationSHA256'],originalFailedClaim=str(failed),originalFailedClaimSHA256=digest(t.root/failed),stoppedFactsSHA256=stopped_pin,candidatePromotionPreserved=True,researchCallsAuthorized=0)
    put(BASE/'responses-retry-authorization.json',approval)
    checkout=Path(t.tmp.name).parent/(Path(t.tmp.name).name+'-fixed-source')
    subprocess.run(['git','worktree','add','-q','--detach',str(checkout),commit],cwd=t.root,env=t.env,check=True)
    t.addCleanup(lambda:subprocess.run(['git','worktree','remove','--force',str(checkout)],cwd=t.root,env=t.env,capture_output=True))
    t.commit();return checkout,old
