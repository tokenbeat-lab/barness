"""Finalize immutable readiness facts; no authority or verification stage starts."""
from pathlib import Path
import collections,decimal,hashlib,json,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/provider-final-readiness';SCRATCH=ROOT/'.scratch/provider-final-readiness';sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
load=lambda p:json.loads(Path(p).read_text())
def sha(p):
 p=Path(p);assert p.is_file() and not p.is_symlink(),str(p)
 return hashlib.sha256(p.read_bytes()).hexdigest()
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(name,x):
 p=BASE/name
 with p.open('x') as f:json.dump(x,f,ensure_ascii=False,indent=2,sort_keys=True);f.write('\n')
 return ref(p)
source=source_snapshot(ROOT);before=load(BASE/'source-before.json');changed=[p for p in sorted(set(source['files'])|set(before['files'])) if source['files'].get(p)!=before['files'].get(p)]
def product(n):return n in ('go.mod','go.sum') or n.startswith('ai/examples/requestid/') or n.startswith('ai/') and not any(n.startswith(p) for p in ('ai/e2e/','ai/live/','ai/release/','ai/internal/testkit/','ai/examples/'))
assert not [n for n in changed if product(n)]
assert len(changed)==6,changed
for p in changed:
 if p.endswith('.go'):assert len((ROOT/p).read_text().splitlines())<=500,p
selection=load(BASE/'full-ordinary-selection-v2.json')['argv'];argv=selection[selection.index('--')+1:];inputs={selection[i+1]:sha(ROOT/selection[i+1]) for i,a in enumerate(selection) if a=='--input'}
found=[]
for p in (BASE/'e2e').glob('*/receipt.json'):
 r=load(p)
 if r.get('argv')==argv and r.get('source')==source and r.get('selectedInputs')==inputs:found.append((p,r))
assert len(found)==1,[(str(p),r['status']) for p,r in found]
p,ordinary=found[0];assert ordinary['status']=='PASS' and ordinary['exitCode']==0
packages=[];required={'MMREADY-stream-'+s for s in ('read-error','eof','complete','request-eof')}|{'KIMIREADY-producer-selected-inputs'};seen=set()
for b in ordinary['bundles']:
 d=p.parent/b['path'];assert sha(d/'manifest.json')==b['manifestSHA256'] and sha(d/'audit.json')==b['auditSHA256'];m=load(d/'manifest.json');assert load(d/'audit.json')['findings']==[]
 for c in m['cases']:
  for n,h in c['artifacts'].items():assert sha(d/c['dir']/n)==h,(c['id'],n)
  if c['id'] in required:assert c['status']=='PASS';seen.add(c['id'])
 counts=dict(collections.Counter(c['status'] for c in m['cases']));assert set(counts)<={'PASS','NOT_RUN'}
 packages.append({'package':b['package'],'cases':len(m['cases']),'statuses':counts,'manifest':ref(d/'manifest.json'),'audit':ref(d/'audit.json')})
assert seen==required and sum(b['package']=='./ai/e2e' for b in ordinary['bundles'])==1
for n in ('native-analysis-v3','native-facts-v2','kimi-original-producers-green-v3','proposal-validation-v1','build-metadata-v2','vet-v1'):
 r=load(BASE/n/'receipt.json');assert r['exitCode']==0 and r['status']=='PASS' and r['source']==source,n
 for path,h in r.get('inputs',{}).items():assert sha(ROOT/path)==h,(n,path)
 for path,h in r.get('files',{}).items():assert sha(BASE/n/path)==h,(n,path)
originals=load(BASE/'originals-audit-v2.json');assert originals['source']==source and originals['originalExitCode']==1 and originals['checked']==2227 and not originals['missing']
partial=load(BASE/'native-analysis-v3/stdout.jsonl');assert partial['pass'] and not partial['nativeFullQualification'] and partial['originalOutcome']=='FAIL' and not partial['originalComplete'];assert len(partial['cases'])==13 and all(c['pass'] for c in partial['cases']);assert partial['missingActualCases']==['thinking-on']
kimi=load(BASE/'kimi-original-producers-green-v3/stdout.jsonl');assert len(kimi)==6 and all(c['producerProofPass'] for c in kimi)
props=load(BASE/'kimi-pending-scopes-v2.json');total=decimal.Decimal('0');assert props['status']=='NOT_GRANTED' and len(props['scopeDocuments'])==6
facts=load(BASE/'native-facts-v2/stdout.jsonl')
for x in props['scopeDocuments']:
 assert sha(ROOT/x['path'])==x['sha256'];d=load(ROOT/x['path']);assert not d['approved'] and not d['grantIssued'] and not d['executable'] and 'schema' not in d and 'id' not in d and 'expiresAt' not in d
 assert d['ordinarySource']==source and d['nativeSourceSHA256']==facts['nativeSourceSHA256'];assert sha(ROOT/d['frozenProducer']['path'])==d['frozenProducer']['sha256']
 for name in ('manifest','report','seal'):
  x=d['baseline'][name];assert sha(ROOT/x['path'])==x['sha256']
 s=d['executionScope'];assert s['calls']==s['httpAttempts']==1 and s['sdkRetries']==s['environmentRetries']==0 and s['scenarioIDs']==['kimi-k3/xai-origin-history'] and s['maxReservedInputTokens']==s['maxRequestedOutputTokens']==4096 and s['processSeconds']==240 and s['callIntervalMs']==60000
 total+=decimal.Decimal(s['proposedMaxUSD'])
assert total==decimal.Decimal('0.60')
metadata=(BASE/'build-metadata-v2/stdout.jsonl').read_text();assert '-tags=live\n' in metadata and 'kimi_baseline_fixture' not in metadata and facts['nativeSourceSHA256'] in metadata
assert all(r['executableGrantRejected'] and not r['issued'] and not r['approved'] for r in load(BASE/'proposal-validation-v1/stdout.jsonl'))
matrix=load(BASE/'delivery-matrix-v2.json');assert matrix['source']==source and len(matrix['paths'])==29 and len({r['provider'] for r in matrix['paths']})==12;assert sum(r['fullPathQualification'] for r in matrix['paths'])==22
for r in matrix['paths']:
 for e in r.get('actualEvidence',[]):assert sha(ROOT/e['consumer']['path'])==e['consumer']['sha256']
protected=load(BASE/'protected-history-before.json')
for path,h in protected.items():assert sha(ROOT/path)==h,path
roots=[Path(p) for p in load(BASE/'store-roots.json')];now={str(p.resolve()):sha(p) for r in roots for p in r.rglob('*') if p.is_file()};old={str((ROOT/p).resolve()):h for p,h in protected.items() if any((ROOT/p).resolve().is_relative_to(r) for r in roots)};assert now==old
protection=save('history-protection-final-v1.json',{'baseline':ref(BASE/'protected-history-before.json'),'protectedFiles':len(protected),'unchanged':len(protected),'missing':[],'changed':[],'publicStoreFiles':len(now),'publicStoreAdded':[],'publicStoreRemoved':[],'publicStoreChanged':[],'newPaidGrants':0,'realClaims':0,'vendorCalls':0,'resumedRounds':0,'controlledSyntheticAuthorities':'private Git namespaces only'})
verification=save('verification-and-retention-v1.json',{'source':source,'sourceChangedFiles':changed,'productSourceChanges':[],'ordinary':ref(p),'elapsedSeconds':ordinary['elapsedSeconds'],'selectedInputs':len(inputs),'packages':packages,'requiredCases':sorted(required),'originalsAudit':ref(BASE/'originals-audit-v2.json'),'partialNativeAnalysis':ref(BASE/'native-analysis-v3/receipt.json'),'realKimiProducerChecks':ref(BASE/'kimi-original-producers-green-v3/receipt.json'),'proposalValidation':ref(BASE/'proposal-validation-v1/receipt.json'),'producerBuildMetadata':ref(BASE/'build-metadata-v2/receipt.json'),'vet':ref(BASE/'vet-v1/receipt.json'),'historyProtection':protection,'raceExecuted':False,'pressureExecuted':False,'vendorLiveExecuted':False,'newPaidGrants':0,'realClaims':0,'releasePass':False})
decisions=save('decisions-v1.json',{'kind':'partial-minimax-and-provider-delivery-readiness','MiniMax':{'route':'minimax-messages','model':'MiniMax-M3','originalOutcome':'FAIL','originalExitCode':1,'applicableUniqueCases':13,'requiredUniqueCases':14,'nativeFullQualification':False,'minimumMissingActualScenes':['thinking-on'],'minimumFactGap':1,'singleSceneProducerAndCompositionReady':False,'paidCallCap':None,'USD':None,'underlyingOriginalReadCause':'NOT_RETAINED','nativeAnalysis':ref(BASE/'native-analysis-v3/receipt.json')},'Kimi':{'scopes':ref(BASE/'kimi-pending-scopes-v2.json'),'proposedCalls':6,'proposedTotalUSD':'0.60','actualNewCalls':0,'paidAuthorization':'NOT_GRANTED','producerInputLayoutIssue':'FIXED_AND_REAL_SIX_OLD_PRODUCERS_VERIFIED'},'matrix':ref(BASE/'delivery-matrix-v2.json'),'verification':verification,'releasePass':False})
previous=ROOT/'.evidence/kimi-history-supplement/current-qualification-index-v4.json';index=load(previous);index.update(previousIndex=ref(previous),currentDecisions=decisions,partialMiniMaxRecords=[{'route':'minimax-messages','qualification':'APPLICABLE_THIRTEEN_ORIGINAL_FULL_FAIL','applicableCases':13,'requiredCases':14,'missingActualCases':['thinking-on'],'nativeFullQualification':False,'analysis':ref(BASE/'native-analysis-v3/receipt.json')}],pendingKimiScopes=ref(BASE/'kimi-pending-scopes-v2.json'),deliveryMatrix=ref(BASE/'delivery-matrix-v2.json'),releasePass=False);index['countingNote']='29 operation/provider/API paths:22 complete applicable,6 Kimi9/10,1 MiniMax Messages13/14.23 complete route/model records include two TypeSafe versions sharing one path.Original FAIL and all permissions/states retained;no paid grant/live/race/pressure/resume.'
latest=save('current-qualification-index-v5.json',index)
attempts=[]
for q in sorted((BASE/'e2e').glob('*/receipt.json')):
 r=load(q);attempts.append({'receipt':ref(q),'status':r['status'],'exitCode':r.get('exitCode'),'argv':r.get('argv')})
save('attempts-index-v1.json',{'attempts':attempts,'notes':['Originals audit v1 omitted selectedInputs and env launcher;v2/final recover all2227 bytes.Original failure log preserved.','Kimi original-producer diagnostics v1/v2 used relative bundle in an analysis wrapper;v3 uses absolute path as production composition does.Those failures are diagnostic,not vendor outcomes.','Producer input-layout public RED is the actual consumer defect;current exact source-plus-input consistency is GREEN.'],'releasePass':False})
print(json.dumps({'ordinary':ref(p),'packages':packages,'changedSourceFiles':changed,'sourceFiles':len(source['files']),'protectedFiles':len(protected),'publicStoreFiles':len(now),'decisions':decisions,'index':latest},ensure_ascii=False,indent=2))
