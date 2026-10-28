"""Close documentation-only confirmation with existing read-only release CLI."""
from pathlib import Path
from datetime import datetime,timezone
import json,hashlib,sys,time
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/kimi-composition-promotion';SCRATCH=ROOT/'.scratch/kimi-composition-promotion';FACTS=BASE/'account-confirmation-v1'
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
from run_process import environment,capture,isolate_go_home
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
freeze=load(BASE/'build-v2/freeze.json');source=source_snapshot(ROOT);assert source==freeze['source']
indexPath=BASE/'pending-scopes-v5-confirmed-v1.json';index=load(indexPath);previous=load(BASE/'pending-scopes-v5.json')
assert index['status']=='NOT_GRANTED' and index['accountAndRegionConfirmation']==index['ratesAndTierConfirmation']=='USER_CONFIRMED'
previousDocs={load(ROOT/r['path'])['route']:(r,load(ROOT/r['path'])) for r in previous['scopeDocuments']}
references={}
def checkrefs(d):
 if isinstance(d,dict):
  if isinstance(d.get('path'),str) and isinstance(d.get('sha256'),str):
   p=ROOT/d['path'];assert p.is_file() and not p.is_symlink(),d
   if d['path'] not in references:references[d['path']]=sha(p)
   assert references[d['path']]==d['sha256'],d['path']
  for v in d.values():checkrefs(v)
 elif isinstance(d,list):
  for v in d:checkrefs(v)
checkrefs(index)
for r in index['scopeDocuments']:
 d=load(ROOT/r['path']);checkrefs(d);beforeRef,before=previousDocs[d['route']]
 allowed={'confirmations','identityToConfirm','status','previousProposal'}
 assert {k for k in set(before)|set(d) if before.get(k)!=d.get(k)}<=allowed
 assert d['confirmations']['accountAndRegion']==d['confirmations']['ratesAndServiceTier']==d['confirmations']['modelPermissionsAndQuota']=='USER_CONFIRMED'
 assert not any(d[k] for k in ['approved','grantIssued','executable'])
 assert d['authorizationStillRequired']['id']==d['authorizationStillRequired']['expiresAt']=='NOT_ISSUED'
 assert d['previousProposal']==beforeRef
 for k in set(before)-allowed:assert before[k]==d[k],(d['route'],k)
identityFields=['accountAlias','accountPurpose','accountScopeID','provider','regionAlias','tenantID']
for route,(r,before) in previousDocs.items():
 d=next(load(ROOT/x['path']) for x in index['scopeDocuments'] if load(ROOT/x['path'])['route']==route)
 assert all(d['identityToConfirm'].get(k)==before['identityToConfirm'].get(k) for k in identityFields)
for n,h in load(FACTS/'preserved-v5-before.json').items():assert sha(ROOT/n)==h
record=load(BASE/'record-v2.json');oldRecord=load(BASE/'record-v1.json');assert record['source']==source and record['selectedInputs']==oldRecord['selectedInputs']
assert sha(ROOT/record['currentState']['path'])==record['currentState']['sha256']
assert sha(FACTS/'status-before.md')==oldRecord['currentState']['sha256']
assert all(sha(ROOT/n)==h for n,h in record['selectedInputs'].items())
# The three-protocol boolean is inapplicable to the single covered MiniMax
# Messages route; it makes no assertion about other MiniMax protocol accounts.
# Keep this explicit interpretation with the original confirmation record.
semantics=FACTS/'field-semantics.json';save(semantics,{'factRecord':ref(FACTS/'facts.json'),'minimaxThreeProtocolField':{'applicable':False,'coveredRoutes':['minimax-messages'],'meaning':'No three-protocol account relation is asserted for MiniMax; its group covers only Messages. false denotes no such applicable group assertion, not evidence that MiniMax protocol accounts differ.'},'kimiSameAccountScope':'Each own site only, CN and GLOBAL remain separate Provider/account groups; no equality across markets inferred.'})
# This invokes a verifier only. It cannot launch any test, grant, collector or
# vendor action; exit 1 remains expected because actual supplements are absent.
out=BASE/'release-process-v2';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out)
binary=ROOT/freeze['releasegate']['path'];assert sha(binary)==freeze['releasegate']['sha256']
argv=[str(binary),'-record',str(BASE/'record-v2.json'),'-out',str(BASE/'release-v2')]
started=datetime.now(timezone.utc);clock=time.monotonic();result=capture(argv,ROOT,env,out,240,[]);finished=datetime.now(timezone.utc)
save(out/'receipt.json',{'kind':'read-only-record-validation-no-test-execution','argv':argv,'source':source,'startedAt':started.isoformat(),'finishedAt':finished.isoformat(),'elapsedSeconds':time.monotonic()-clock,'expectedRefusal':True,'files':{p.name:sha(p) for p in out.iterdir() if p.is_file()},**result})
report=load(BASE/'release-v2/release-report.json');checks={c['id']:c for c in report['checks']}
assert result['exitCode']==1 and not report['pass'] and checks['source']['pass'] and checks['e2e']['pass'] and not checks['live']['pass']
assert checks['race']['pass'] and checks['pressure']['pass'] # prior scoped N/A, no execution
publicPins=load(BASE/'initial/public-stores-before.json');roots=load(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json')
actual={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()};assert actual=={str(ROOT/p):h for p,h in publicPins.items()}
assert source_snapshot(ROOT)==source and all(sha(ROOT/n)==h for n,h in record['selectedInputs'].items())
assert sha(ROOT/'ai/catalog_kimi.json')=='c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109'
issue=SCRATCH/'issues/02-v5-account-facts.md';text=issue.read_text().replace('Status: claimed','Status: resolved');issue.write_text(text+'\n## Answer\n\n三个账户组事实确认已留存，七份v5事实确认修订1已完成静态差异/引用/来源与输入相等性核对。原v5及旧状态快照保留。record-v2只读判定source/e2e通过、live未完成，release.pass=false；未签发付费授权或执行测试。详见 ../status.md。\n')
closure={'kind':'user-facts-recorded-seven-v5-proposals-still-unapproved','factConfirmation':ref(FACTS/'facts.json'),'factFieldInterpretation':ref(semantics),'verification':ref(FACTS/'verification.json'),'currentState':ref(SCRATCH/'status.md'),'record':ref(BASE/'record-v2.json'),'releaseProcess':ref(out/'receipt.json'),'releaseReport':ref(BASE/'release-v2/release-report.json'),'priorCurrentStateSnapshot':ref(FACTS/'status-before.md'),'priorDelivery':ref(BASE/'delivery-v1.json'),'pendingScopes':ref(indexPath),'matrix':ref(SCRATCH/'delivery-matrix-v5.md'),'factsConfirmedAccountGroups':3,'factConfirmedScopes':7,'verifiedUniqueReferencePins':len(references),'originalV5Preserved':True,'sourceAndOrdinaryInputsExact':True,'publicAuthorityFiles':len(actual),'publicAuthoritySetAndBytesUnchanged':True,'totalProposedUSD':'0.61','paidGrantsIssued':0,'realClaims':0,'vendorCalls':0,'testExecutions':0,'liveExecuted':False,'raceExecuted':False,'pressureExecuted':False,'catalogActivated':False,'pausedResumed':False,'releasePass':False}
save(BASE/'delivery-v2.json',closure)
print(json.dumps({'groupsConfirmed':3,'scopeFactsConfirmed':7,'publicAuthorityFiles':len(actual),'releaseCLIExitCode':result['exitCode'],'releasePass':False,'sourcePass':checks['source']['pass'],'e2ePass':checks['e2e']['pass'],'paidGrantsIssued':0,'testsExecuted':0,'delivery':str((BASE/'delivery-v2.json').relative_to(ROOT))},ensure_ascii=False))
