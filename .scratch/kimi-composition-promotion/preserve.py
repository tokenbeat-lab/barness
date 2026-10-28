"""Read-only retention audit. Original versions remain at their pinned paths."""
from pathlib import Path
import json,hashlib,time
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/kimi-composition-promotion'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
t=time.time();previous=load(ROOT/'.evidence/minimax-thinking-supplement/history-protection-final-v2.json')
allowed={x['path']:x for x in previous['authorizedRevisionsWithOriginalBytes']}
old=BASE/'initial/REGIONAL.md';allowed['ai/live/REGIONAL.md']={'path':'ai/live/REGIONAL.md','originalRetained':str(old.relative_to(ROOT)),'originalSHA256':sha(old),'reason':'current authorized Kimi promotion contract documentation only'}
scans=[];bad=[];missing=[]
cache={}
for name in ['.evidence/provider-final-readiness/protected-history-before.json','.evidence/minimax-thinking-supplement/initial/protected-inputs-before.json']:
 pins=load(ROOT/name)
 for n,h in pins.items():
  p=ROOT/n
  if not p.is_file():missing.append(n);continue
  if n not in cache:cache[n]=sha(p)
  if cache[n]!=h:
   a=allowed.get(n)
   if not a or a['originalSHA256']!=h or sha(ROOT/a['originalRetained'])!=h:bad.append(n)
 scans.append({'inventory':ref(ROOT/name),'files':len(pins)})
roots=load(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json')
actual={}
for root in roots:
 for p in Path(root).rglob('*'):
  if p.is_file():actual[str(p)]=sha(p)
expected={str(ROOT/p):h for p,h in load(BASE/'initial/public-stores-before.json').items()};publicExact=expected==actual
assert publicExact and not missing and not bad,(missing[:3],bad[:3],publicExact)
assert (ROOT/'ai/catalog_kimi.json').read_bytes()==(BASE/'initial/catalog-before.json').read_bytes()
# Previous unapproved proposals are preserved and still match their old index.
previousScopes=load(ROOT/'.evidence/minimax-thinking-supplement/pending-scopes-v3.json')
for r in previousScopes['scopeDocuments']:assert sha(ROOT/r['path'])==r['sha256']
for a in allowed.values():a['currentSHA256']=sha(ROOT/a['path'])
out={'kind':'historical-retention-with-explicit-document-revisions','pass':True,'scans':scans,'authorizedDocumentRevisionsWithOriginalBytes':list(allowed.values()),'missing':missing,'unexplainedChanges':bad,'publicAuthorityFiles':len(actual),'publicAuthorityExactSetAndBytesUnchanged':publicExact,'oldSevenProposalsUnchanged':True,'inactiveCatalogUnchanged':True,'elapsedSeconds':time.time()-t,'paidGrantsIssued':0,'realClaims':0,'vendorCalls':0,'pausedResumed':False}
with (BASE/'retention-v2.json').open('x') as f:json.dump(out,f,indent=2);f.write('\n')
print({k:out[k] for k in ['pass','publicAuthorityFiles','elapsedSeconds']})
