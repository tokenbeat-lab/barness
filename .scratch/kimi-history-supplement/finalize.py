"""Audit complete current receipts and preserve a new immutable qualification index."""
import collections
import hashlib
import json
from pathlib import Path
import sys
ROOT=Path(__file__).resolve().parents[2]
BASE=ROOT/'.evidence/kimi-history-supplement'
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
load=lambda p:json.loads(Path(p).read_text())
def sha(p):
 p=Path(p);assert p.is_file() and not p.is_symlink(),str(p)
 return hashlib.sha256(p.read_bytes()).hexdigest()
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(name,value):
 p=BASE/name
 with p.open('x') as f:json.dump(value,f,ensure_ascii=False,indent=2,sort_keys=True);f.write('\n')
 return ref(p)
source=source_snapshot(ROOT);before=load(BASE/'source-before.json')
changed=[n for n in sorted(set(source['files'])|set(before['files'])) if source['files'].get(n)!=before['files'].get(n)]
def product(n):
 return n in ('go.mod','go.sum') or n.startswith('ai/examples/requestid/') or n.startswith('ai/') and not any(n.startswith(p) for p in ('ai/e2e/','ai/live/','ai/release/','ai/internal/testkit/','ai/examples/'))
assert not [n for n in changed if product(n)]
selected=load(BASE/'full-ordinary-selection-v1.json')['argv'];command=selected[selected.index('--')+1:]
inputs={selected[i+1]:sha(ROOT/selected[i+1]) for i,x in enumerate(selected) if x=='--input'}
found=[]
for p in (BASE/'e2e').glob('*/receipt.json'):
 r=load(p)
 if r.get('argv')==command and r.get('source')==source and r.get('selectedInputs')==inputs:found.append((p,r))
assert len(found)==1,[(str(p),r['status']) for p,r in found]
path,ordinary=found[0];assert ordinary['status']=='PASS' and ordinary['exitCode']==0
routes=['kimi-'+site+'-'+api for site in ('cn','intl') for api in ('chat','responses','messages')]
required=set()
for route in routes:
 for k in ('composition','capture','failure','template'):required.add('KHS-'+k+'-'+route)
 for k in ('reviewed','later-history','changed-origin','new-runtime','dependency','other-provider'):required.add('KHS-source-'+route+'-'+k)
for route in routes[:3]:required.add('KHS-request-reencoding-'+route)
for k in ('full-grant','old-nine','full-entry','timeout','count','source','binary','baseline','scene','provider','model','calls','attempts','input','output','spend','expired','purpose','interval','duplicate','unknown','null-scene'):required.add('KHS-admission-'+k)
for k in ('source','legacy','check','formal','other-provider','missing-pin'):required.add('KHS-template-scope-'+k)
assert len(required)==91
seen=set();packages=[]
for b in ordinary['bundles']:
 bundle=path.parent/b['path'];assert sha(bundle/'manifest.json')==b['manifestSHA256'] and sha(bundle/'audit.json')==b['auditSHA256']
 m=load(bundle/'manifest.json');assert load(bundle/'audit.json')['findings']==[]
 for c in m['cases']:
  for n,h in c['artifacts'].items():assert sha(bundle/c['dir']/n)==h,(c['id'],n)
  if c['id'].startswith('KHS-'):
   assert c['status']=='PASS',c['id'];seen.add(c['id'])
 counts=dict(collections.Counter(c['status'] for c in m['cases']));assert set(counts)<={'PASS','NOT_RUN'}
 packages.append({'package':b['package'],'cases':len(m['cases']),'statuses':counts,'manifest':ref(bundle/'manifest.json'),'audit':ref(bundle/'audit.json')})
assert seen==required,{'missing':sorted(required-seen),'extra':sorted(seen-required)}
assert sum(x['package']=='./ai/e2e' for x in packages)==1
for n in ('tools-v1','vet-v1','vet-live-v1','public-replay-v2','native-baselines-v2','build-v1'):
 r=load(BASE/n/'receipt.json');assert r['source']==source and r['exitCode']==0 and r['status']=='PASS',n
 for p,h in r.get('inputs',{}).items():assert sha(ROOT/p)==h,p
 for p,h in r.get('files',{}).items():assert sha(BASE/n/p)==h,(n,p)
replay=load(BASE/'public-replay-v2/stdout.jsonl')
assert len(replay)==54 and all(r['pass'] and r['requestEqual'] and r['attempts']==1 for r in replay)
assert sum(r['responseEqual'] for r in replay)==48 and sum(r['cancelCapture'] for r in replay)==6 and sum(r['cancelPrefix'] for r in replay)==6
baseline=load(BASE/'native-baselines-v2/stdout.jsonl');assert len(baseline)==6 and all(r['pass'] and r['cases']==9 for r in baseline)
protected=load(BASE/'protected-history-before.json')
for n,h in protected.items():assert sha(ROOT/n)==h,n
roots=[Path(n) for n in load(BASE/'store-roots.json')]
now={str(p.resolve()):sha(p) for r in roots for p in r.rglob('*') if p.is_file()}
old={str((ROOT/p).resolve()):h for p,h in protected.items() if any((ROOT/p).resolve().is_relative_to(r) for r in roots)}
assert now==old,{'added':sorted(set(now)-set(old)),'removed':sorted(set(old)-set(now)),'changed':[p for p in now.keys()&old.keys() if now[p]!=old[p]]}
protection=save('history-protection-final-v1.json',{'baseline':ref(BASE/'protected-history-before.json'),'protectedFiles':len(protected),'unchanged':len(protected),'missing':[],'changed':[],'publicStoreFiles':len(now),'publicStoreAdded':[],'publicStoreRemoved':[],'publicStoreChanged':[],'realGrantsIssued':0,'realClaims':0,'resumedRounds':0})
verification=save('verification-and-retention-v1.json',{'source':source,'sourceChangedFiles':changed,'productSourceChanges':[],'ordinary':ref(path),'ordinaryElapsedSeconds':ordinary['elapsedSeconds'],'ordinarySelectedInputs':len(inputs),'packages':packages,'requiredCases':sorted(required),'publicReplay':ref(BASE/'public-replay-v2/receipt.json'),'nativeBaselines':ref(BASE/'native-baselines-v2/receipt.json'),'tools':ref(BASE/'tools-v1/receipt.json'),'vet':ref(BASE/'vet-v1/receipt.json'),'vetLive':ref(BASE/'vet-live-v1/receipt.json'),'build':ref(BASE/'build-v1/receipt.json'),'historyProtection':protection,'newPaidGrants':0,'newVendorCalls':0,'liveStatus':'NOT_RUN','raceStatus':'NOT_RUN','pressureStatus':'NOT_RUN','releasePass':False})
selection=load(ROOT/'.evidence/provider-delivery-offline/selected-evidence-v1.json')['kimi']
originals=[]
for row in selection:
 report=load(ROOT/row['bundle']/'regional-report.json')
 originals.append({'route':row['route'],'baselineQualification':'APPLICABLE_NINE','mandatoryCases':9,'bundle':row['bundle'],'manifest':ref(ROOT/row['bundle']/'manifest.json'),'report':ref(ROOT/row['bundle']/'regional-report.json'),'originalCommit':report['gitCommit'],'originalStartedAt':report['startedAt'],'originalFinishedAt':report['finishedAt'],'missingActualCases':['kimi-k3/xai-origin-history'],'newActualCaseStatus':'NOT_RUN','fullTenQualification':'INCOMPLETE','catalog':'INACTIVE','budgetProposalUSD':0.10,'paidAuthorityIssued':False})
decisions=save('decisions-v1.json',{'kind':'kimi-nine-baseline-current-applicability','routes':originals,'sourceReview':ref(BASE/'source-review-deltas-v1.json'),'reviewedPairs':ref(BASE/'reviewed-sources-added-v1.json'),'verification':verification,'newVendorFacts':0,'minimumActualNewScenarios':6,'proposedTotalUSD':0.60,'authorization':'NOT_ISSUED','releasePass':False})
previous=ROOT/'.evidence/provider-source-scope/current-qualification-index-v3.json';index=load(previous)
index.update(previousIndex=ref(previous),currentDecisions=decisions,partialKimiRecords=originals,productSourceChanges=[],verificationSourceChanges=changed,releasePass=False)
index['countingNote']='23 retained route/model evidence entries is not a path delivery ratio; Kimi six remain 9/10 actual, controlled 9+1 is mechanism proof only. Other provider evidence, prior actual failure, grants and paused states remain retained; no paid/live/race/pressure issued or run.'
newindex=save('current-qualification-index-v4.json',index)
print(json.dumps({'ordinary':ref(path),'packages':packages,'newRequiredCases':len(required),'sourceFiles':len(source['files']),'changedSourceFiles':len(changed),'protectedFiles':len(protected),'publicStoreFiles':len(now),'decisions':decisions,'index':newindex},ensure_ascii=False,indent=2))
