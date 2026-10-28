"""Validate this design's arithmetic, retained facts, links and non-execution boundary."""
from pathlib import Path
from decimal import Decimal as D
import hashlib,json,re,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement-plan';SCRATCH=ROOT/'.scratch/minimax-thinking-supplement-plan';sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
load=lambda p:json.loads(Path(p).read_text())
def sha(p):
 p=Path(p);assert p.is_file() and not p.is_symlink(),str(p)
 return hashlib.sha256(p.read_bytes()).hexdigest()
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(n,v):
 with (BASE/n).open('x') as f:json.dump(v,f,ensure_ascii=False,indent=2,sort_keys=True);f.write('\n')
source=source_snapshot(ROOT);before=load(BASE/'source-before.json');assert source==before
prior=load(ROOT/'.evidence/provider-final-readiness/verification-and-retention-v2.json');ordinary=load(ROOT/prior['ordinary']['path']);assert prior['source']==source==ordinary['source'] and ordinary['status']=='PASS' and ordinary['exitCode']==0
assert len(source['files'])==2034 and len(ordinary['selectedInputs'])==182
for n,h in ordinary['selectedInputs'].items():assert sha(ROOT/n)==h,n
x=load(BASE/'scope-and-budget-v2.json');assert not x['approved'] and not x['grantIssued'] and not x['executable'] and x['implementationStatus']=='NOT_IMPLEMENTED' and x['authorizationStatus']=='NOT_GRANTED';assert all(k not in x for k in ('scope','schema','id','expiresAt'))
b=x['baseline'];assert b['originalOutcome']=='FAIL' and b['originalProducerExitCode']==1 and not b['originalComplete'];assert len(b['applicableContributionScenes'])==13 and len({s['id'] for s in b['applicableContributionScenes']})==13
for n in ('manifest','report','seal','producer','priorApplicableAnalysis'):
 r=b[n];assert sha(ROOT/r['path'])==r['sha256'],n
old=load(ROOT/b['report']['path']);actual={c['id'] for c in old['expected']};passed={c['id'] for c in b['applicableContributionScenes']};assert passed.isdisjoint({'thinking-on'}) and passed|{'thinking-on'}==actual and len(actual)==14
control=[c for c in old['calls'] if c['id']=='thinking-control'];assert [(c['rerun'],c['outcome']) for c in control]==[(0,'FAIL'),(1,'PASS')]
p=x['futureFixedProfile'];assert p['maxCalls']==p['maxHTTPAttempts']==p['testCount']==1;assert p['maxSDKRetries']==p['maxEnvironmentRetries']==p['maxRequestReruns']==p['maxToolReruns']==0;assert p['maxOutputTokens']==p['maxRequestedOutputTokens']==p['maxReservedInputTokens']==2048;assert p['callSeconds']==120 and p['processSeconds']==180 and p['firstCallWaitSeconds']==0
ex=load(ROOT/b['bundle']/'MM07-thinking-on/exchanges.json')[0];wire=ex['requestBody'].encode();body=json.loads(wire);assert len(wire)==p['originalWireBytes']==389;assert hashlib.sha256(wire).hexdigest()==p['wireSHA256'];assert max(2048,len(wire)+1024)==2048;assert body['thinking']=={'type':'adaptive'} and body['stream'] is True and body['max_tokens']==2048 and body['model']=='MiniMax-M3'
budget=x['budget'];rates=load(ROOT/budget['priceCatalog']['path']);r=next(m for m in rates['models'] if m['api']=='anthropic-messages')['cost'];hi=max(D(str(r['input'])),D(str(r['cacheRead'])),*[D(str(t['input'])) for t in r['tiers']],*[D(str(t['cacheRead'])) for t in r['tiers']]);ho=max(D(str(r['output'])),*[D(str(t['output'])) for t in r['tiers']]);base=(D(2048)*D(str(r['input']))+D(2048)*D(str(r['output'])))/D(1000000);reserve=(D(2048)*hi+D(2048)*ho)/D(1000000);assert base==D(budget['baseFullTokenCapsUSD'])==D('.003072');assert reserve==D(budget['conservativeKnownRatesReserveUSD'])==D('.006144');assert reserve<D(budget['proposedUSDCap'])==D('.01');assert not budget['invoiceGuarantee'] and budget['positiveCacheWritePrice'].startswith('UNKNOWN_NOT_FREE')
assert x['scopeExtensionProposed']['appendOmitEmptyFields']==['BaselineManifestSHA256','MaxUSD'];assert not x['scopeExtensionProposed']['newRegistryOrWorkflowEngine'];assert x['thisTurn']=={'vendorCalls':0,'newPaidGrants':0,'realClaims':0,'liveStarted':False,'raceExecuted':False,'pressureExecuted':False,'pausedRoundsResumed':0,'implementationStarted':False,'releasePass':False}
links=0
for f in ('plan.md','spec.md'):
 doc=SCRATCH/f;assert len(doc.read_text().splitlines())<=500
 for t in re.findall(r'\]\(([^)]+)\)',doc.read_text()):
  if '://' in t or t.startswith('#'):continue
  assert (doc.parent/t.split('#')[0]).exists(),(f,t);links+=1
protected=load(BASE/'protected-inputs-before.json')
for n,h in protected.items():assert sha(ROOT/n)==h,n
roots=[Path(p) for p in load(BASE/'store-roots.json')];now={str(p.resolve()):sha(p) for r in roots for p in r.rglob('*') if p.is_file()};old={str((ROOT/p).resolve()):h for p,h in protected.items() if any((ROOT/p).resolve().is_relative_to(r) for r in roots)};assert now==old and len(now)==8156
value={'kind':'design-only-offline-verification','clientDate':'2026-10-28','status':'PASS','sourceFiles':len(source['files']),'sourceChangedFiles':[],'latestCompleteOrdinary':prior['ordinary'],'latestOrdinaryInputsUnchanged':len(ordinary['selectedInputs']),'newOrdinaryExecution':'NOT_RUN; no runtime or consumer implementation changed','checks':{'baselineUniqueContributions':13,'missingActualScenes':['thinking-on'],'fixedUnion':14,'originalFailedProducerExit':1,'wireBytes':389,'reservedInput':2048,'outputCap':2048,'conservativeUSD':str(reserve),'proposedUSD':'0.01','proposalIsNotExecutableGrant':True,'localReferences':links},'protectedInputFiles':len(protected),'protectedInputMissing':[],'protectedInputChanged':[],'publicStoreFiles':len(now),'publicStoreAdded':[],'publicStoreRemoved':[],'publicStoreChanged':[],'plan':ref(SCRATCH/'plan.md'),'scope':ref(BASE/'scope-and-budget-v2.json'),'verificationProgram':ref(Path(__file__)),'implementationStarted':False,'newPaidGrants':0,'claims':0,'vendorCalls':0,'liveExecuted':False,'raceExecuted':False,'pressureExecuted':False,'pausedRoundsResumed':0,'releasePass':False}
save('verification-and-retention-v2.json',value);print(json.dumps(value,ensure_ascii=False,indent=2))
