#!/usr/bin/env python3
"""Copy final execution source evidence; do not seal or modify input originals."""
from pathlib import Path
import hashlib,json,shutil
ROOT=Path(__file__).resolve().parents[4];W=Path(__file__).resolve().parent
B=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T041450Z-10fb1e18'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def cp(src,dst):
 dst.parent.mkdir(parents=True,exist_ok=True)
 with src.open('rb') as a,dst.open('xb') as z:shutil.copyfileobj(a,z)
 assert sha(src)==sha(dst)
 return {'source':str(src.relative_to(ROOT)),'copy':str(dst.relative_to(B)),'sha256':sha(dst)}
assert not (B/'delivery.json').exists()
names=['run-ordinary-v6.py','ordinary-inputs-v6.json','required-cases-v6.json','run-build-v1.py','inspect-ordinary-v6.py','boundary-applicability-v1.py','boundary-applicability-v1.json','historical-applicability-v1.py','historical-applicability-v1.json']+[f'build-v1/{a}.{b}' for a in ('generate','build','vet') for b in ('json','stdout','stderr')]
copies=[cp(W/'verification'/name,B/'provenance/final-execution-v1'/name) for name in names]
for name in ['review-spec-v3.md','review-spec-v4.md','review-standards-v3.md','review-standards-v4.md','inputs/old-name-negative-audit-v2.md']:
 copies.append(cp(W/name,B/'provenance/final-execution-v1'/name))
copies.append(cp(Path(__file__),B/'provenance/final-execution-v1/prepare-final-inputs-v1.py'))
refs=[]
for p in sorted((ROOT/'.evidence/catalog-model-id-policy/e2e').glob('*/receipt.json')):
 data=json.loads(p.read_text());refs.append({'path':str(p.relative_to(ROOT)),'sha256':sha(p),'status':data.get('status'),'exitCode':data.get('exitCode')})
old=ROOT/'.evidence/catalog-kimi-k31-pro/e2e/20261102T030707.838552Z-x6dz4dj5/receipt.json'
obj={'status':'PREPARED_NOT_COMPLETE','copies':copies,'originalReceipts':refs,'previousCatalogEvidence':{'path':str(old.relative_to(ROOT)),'sha256':sha(old),'meaning':'Historical catalog check; not new policy proof'},'completeEvidenceSelfContained':False,'dynamicFixtureAndHistoricalArchiveAudit':'PENDING'}
p=B/'provenance/final-execution-v1/mapping.json'
with p.open('x') as f:json.dump(obj,f,indent=2);f.write('\n')
print(json.dumps({'copies':len(copies),'receipts':len(refs),'mappingSHA256':sha(p)}))
