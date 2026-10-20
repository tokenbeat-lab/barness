"""Socket-free synthetic234-file prefix and public producer/consumer executables."""
import ast
import hashlib
import json
import zipfile
import stat
from pathlib import Path
BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
OLD='provider-sites-final-acceptance-2'
NEW='provider-sites-final-acceptance-3'

def seed(test):
    root=test.root;files={}
    def put(name,value,raw=False):
        p=root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(value if raw else (json.dumps(value)+'\n').encode());files[str(name)]=hashlib.sha256(p.read_bytes()).hexdigest()
    combos=[p+'-'+r for p in ('zai','minimax-cn') for r in ('chat','responses','messages')]
    names=['candidate-'+n for n in combos]+['glm-chat','glm-responses','glm-responses-retry-1','glm-messages']+['formal-'+n for n in combos]+['mm-global-formal-m3-chat']
    for name in names:
        reserve=650000 if 'zai-' in name else 350000 if 'minimax-cn-' in name else 1000000 if name in ('glm-chat','glm-responses') else 500000 if name.startswith('glm-') else 200000
        failed=name in ('glm-responses','mm-global-formal-m3-chat');bundle=BASE/'old2'/name/'session'
        budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None) if failed else dict(state='REPORTED',status='PASS',calls=22 if name.startswith('glm-') else 18 if 'zai' in name else 9,attempts=22 if name.startswith('glm-') else 18 if 'zai' in name else 9)
        put(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'),dict(state='COMPLETED',exitCode=1 if failed else 0,validationRound=OLD,sourceCommit=test.git('rev-parse','HEAD'),collectionRoot=str(BASE/'old2'),baseName=name,reservedMicrousd=reserve,budget=budget,nativeConsumerReceiptSha256=None if failed else 'a'*64))
        if name=='mm-global-formal-m3-chat':
            put(bundle/'minimax-report.json',dict(outcome='FAIL',complete=False,auditPassed=True,budget=dict(calls=13,attempts=12),expected=[dict(id=str(i)) for i in range(13)],calls=[dict(id=str(i),outcome='FAIL' if i==0 else 'PASS') for i in range(13)]));native=bundle
    for name in ('grant.json','human-approval.json','source-review.json','chat.json','operator-authority.json'):put(Path('.evidence/minimax-renewals')/OLD/name,dict(synthetic=True))
    put(Path('.evidence/minimax-renewals')/OLD/'formal.jsonl',b'controlled13-reserve12-attempt\n',True);put(Path('.evidence/minimax-renewals')/OLD/'launch.lock',b'',True)
    jobs=['openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','protocol114','mm-global-formal-m3-responses','mm-global-formal-m3-messages','text-jev','text-luna','visual-jev','visual-decisions']
    earlier={n:v for n,v in files.items() if any(n.endswith(OLD+'-'+name+'.json') for name in names[:8])}
    for i in range(131-len(earlier)):
        p=BASE/'old2-earlier'/('original-'+str(i)+'.json');put(p,dict(controlled=i));earlier[str(p)]=files.pop(str(p))
    put(BASE/'round2-stopped/index.json',dict(files=earlier))
    glm={}
    for route,name in [('chat','glm-chat'),('responses','glm-responses-retry-1'),('messages','glm-messages')]:
        p=Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json');receipt=json.loads((root/p).read_text());glm[route]=dict(canonicalReceiptSHA256=files[str(p)],actualSourceCommit=receipt['sourceCommit'],nativeConsumerSHA256=receipt['nativeConsumerReceiptSha256'])
    put(BASE/'responses-retry-preparation/glm-three-routes.json',dict(kind='THREE_GENUINE_FULL22_GLM_PROOFS',routes=glm))
    # Match production:234 directly owns seven new claims, original131 owns
    # the first eight, and the pinned GLM summary owns retry/Messages.
    for name in names[:10]:files.pop(str(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json')))
    for i in range(234-len(files)):put(BASE/'closed-cohort-2'/('controlled-original-'+str(i)+'.json'),dict(synthetic=i))
    stop=dict(kind='ACTUAL_REMAINING_STOP_AFTER_REAL_GLOBAL_MINIMAX_CHAT_THINKING_FAILURE',regionalFormalProcesses=6,regionalFormalScenarios=81,globalMiniMaxActualCalls=13,globalMiniMaxActualHTTP=12,campaignActualCalls=303,campaignActualHTTP=302,campaignClaimedCalls=456,campaignClaimedHTTP=492,wholeReservedMicrousd=10700000,refundMicrousd=0,unknownParentProcesses=2,originalFailedGLMUnchanged=True,automaticReruns=False,unstartedProcesses=jobs,files=dict(files),processes=[dict(name='mm-global-formal-m3-chat',nativeBundle=str(native))])
    put(BASE/'remaining-stopped/index.json',stop)
    put(BASE/'closed-cohort-2/seal.json',dict(schema=1,kind='CLOSED_SECOND_COHORT_HISTORY_SEAL',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',stopIndex=str(BASE/'remaining-stopped/index.json'),stopIndexSHA256=files[str(BASE/'remaining-stopped/index.json')],actualCalls=303,actualHTTPAttempts=302,claimedCalls=456,claimedHTTPAttempts=492,wholeReservedMicrousd=10700000,refundMicrousd=0,noAutomaticReruns=True,paidContinuationAuthorized=False))
    # These controlled native seams do not certify a real GLM publication.
    (test.tools/'final_publication.py').write_text('import final_contract as r\nc=r.c\ndef validate(index,views=None):return dict(controlled=True)\ndef require(route):\n name="glm-responses-retry-1" if route=="responses" else "glm-"+route\n p=c.common_root()/".evidence/minimax-launches"/(r.ROUND+"-"+name+".json")\n return c.doc(p),c.digest(p)\n')
    p=root/'ai/live/testdata/glm-formal';p.mkdir(parents=True);(p/'live-captures.json').write_text('{}')
    rows=[dict(combo=n) for n in jobs if n not in ('protocol114','mm-global-formal-m3-responses','mm-global-formal-m3-messages') and not n.startswith(('text-','visual-'))]+[dict(combo=n) for n in [*combos,'typesafe-classifier-jev-1.14.0',*('zhipu-'+r for r in ('chat','responses','messages')),*('minimax-'+r for r in ('chat','responses','messages'))]]
    (root/'ai/live/support-matrix.json').write_text(json.dumps(dict(schema=3,rows=rows)))
    views={}
    for route in ('chat','responses','messages'):
        dest=root/'.evidence/provider-sites/retained-glm'/route;dest.mkdir(parents=True);raw=json.dumps(dict(combo='zhipu-'+route)).encode();(dest/'live-report.json').write_bytes(raw)
        archive=dest.with_suffix('.zip');info=zipfile.ZipInfo('live-report.json');info.external_attr=(stat.S_IFREG|0o600)<<16
        with zipfile.ZipFile(archive,'w') as z:z.writestr(info,raw)
        views[route]=dict(archive=str(archive.relative_to(root)),archiveSHA256=hashlib.sha256(archive.read_bytes()).hexdigest(),reportPath='live-report.json',files=1)
    (p/'support-views.json').write_text(json.dumps(dict(routes=views)))
    test.commit()

def builder(test):
    test.fake_build();go=test.root/'bin/go';text=go.read_text();tree=ast.parse(text)
    producer=next(n.args[0].value for n in ast.walk(tree) if isinstance(n,ast.Call) and isinstance(n.func,ast.Attribute) and n.func.attr=='write_text' and isinstance(n.args[0],ast.Constant))
    extra=r'''
import hashlib,subprocess,shutil
root=pathlib.Path.cwd()
while not (root/'go.mod').exists():root=root.parent
out=pathlib.Path(sys.argv[0]).parent
new='provider-sites-final-acceptance-3';old='provider-sites-final-acceptance-2'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def dump(p,v):p.parent.mkdir(parents=True,exist_ok=True);p.write_text(json.dumps(v))
if '-recovery-source' in a:
 prior=json.loads((out/'predecessor.json').read_text());native=root/'.evidence/minimax-renewals'/old
 recovery=dict(secondCohortSealSha256=prior['secondCohortSealSHA256'],stopIndexSha256=prior['stopIndexSHA256'],secondCalls=13,secondAttempts=12,secondLedgerRows=25,retainedActualCalls=303,retainedActualAttempts=302,retainedClaimedCalls=456,retainedClaimedAttempts=492,retainedReservedMicrousd=10700000,secondLedgerTailSha256='a'*64)
 for leaf,key in [('grant.json','secondGrantSha256'),('human-approval.json','secondHumanApprovalSha256'),('source-review.json','secondSourceReviewSha256'),('chat.json','secondChatClaimSha256'),('formal.jsonl','secondLedgerSha256')]:recovery[key]=sha(native/leaf)
 print(json.dumps(dict(renewalId=new,sourceCommit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36,closedHistory={str(i):'a'*64 for i in range(15)},closedHistorySha256=prior['firstCohort']['nativeHistorySHA256'],firstCohortSealSha256=prior['firstCohort']['sealSHA256'],recovery=recovery,recoveryHistorySha256=sha(native/'formal.jsonl'))))
if any(v.startswith('-recovery-preflight=') for v in a):
 grant=root/'.evidence/minimax-renewals'/new/'grant.json';g=json.loads(grant.read_text());assert g['schema']==4
 print(json.dumps(dict(route=a[0].split('=')[1],grant=g,grantSha256=sha(grant),facts=dict(renewalId=new))))
if os.getenv('BARNESS_FINAL_BUDGET'):
 evidence=pathlib.Path(os.environ.get('BARNESS_AI_EVIDENCE_DIR',a[a.index('-out')+1] if '-out' in a else '.'))
 name=evidence.name;session=evidence if name.startswith('visual-') else evidence/'controlled-session';session.mkdir(parents=True)
 commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();report=dict(gitCommit=commit,budget=dict(callsUsed=1,httpAttempts=1),expected=['one'],scenarios=[dict(id='one',outcome='PASS')])
 leaf='live-report.json'
 if name.startswith('formal-'):
  report['combo']=name.removeprefix('formal-');leaf='regional-report.json';dump(session/'live-report.json',report)
 elif name.startswith('mm-'):
  grant=root/'.evidence/minimax-renewals'/new/'grant.json';report.update(formalRenewalId=new,formalRenewalGrantSha256=sha(grant),outcome='PASS',complete=True,auditPassed=True,identity=dict(route=os.environ['BARNESS_MINIMAX_ROUTE']),session=session.name,budget=dict(calls=13,attempts=12));leaf='minimax-report.json';dump(session/'minimax-integrity.json',{});dump(session/'minimax-ledger.json',{})
 elif name.startswith('text-'):report.update(status='COMPLETE',budget=dict(calls=1,attempts=1));leaf='evaluation-report.json'
 elif name.startswith('visual-'):report.update(status='COMPLETE',budget=dict(calls=1,attempts=1));leaf='report.json'
 elif name=='protocol114':report['budget']=dict(actual=dict(calls=1,attempts=1));leaf='multimodal-report.json'
 else:report['combo']=os.environ['BARNESS_AI_LIVE_COMBO']
 dump(session/leaf,report)
 usd=pathlib.Path(os.environ['BARNESS_FINAL_BUDGET']);v=json.loads(usd.read_text());assert v['round']==old;v.update(reservedMicrousd=1000,attempts=12 if name.startswith('mm-') else 1)
 if (root/'.evidence/wrong-usd-policy').exists():v['round']=new
 dump(usd,v)
if '-publish' in a and '-formal' in a:
 base=pathlib.Path(a[-1]);report=json.loads((base/'minimax-report.json').read_text());report['combo']='minimax-'+report['identity']['route'];dump(base/'live-report.json',report)
if '-matrix' in a and pathlib.Path(sys.argv[0]).name=='supportmatrix':
 matrix=pathlib.Path(a[a.index('-matrix')+1]);v=json.loads(matrix.read_text());
 for row in v.get('rows',[]):row['allPassedAt']='controlled'
 dump(matrix,v)
'''
    go.write_text(text.replace(repr(producer),repr(producer+extra)));test.commit()
