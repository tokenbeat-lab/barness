#!/usr/bin/env python3
"""Copy completed ordinary receipt and review originals; record prior attempt references."""
from pathlib import Path
import hashlib,json,shutil
root=Path(__file__).resolve().parents[4]; work=Path(__file__).resolve().parent
batch=root/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6'
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def copy(src,dst):
 dst.parent.mkdir(parents=True,exist_ok=True)
 with src.open('rb') as a,dst.open('xb') as b: shutil.copyfileobj(a,b)
 assert sha(src)==sha(dst)
 return {'source':str(src.relative_to(root)) if src.is_relative_to(root) else str(src),'destination':str(dst.relative_to(batch)),'sha256':sha(dst)}
assert not (batch/'delivery.json').exists()
receipt=root/'.evidence/catalog-kimi-k31-pro/e2e/20261102T030707.838552Z-x6dz4dj5/receipt.json'
d=json.loads(receipt.read_text());assert d['status']=='PASS' and d['exitCode']==0 and d['source']['commit']=='97e3fef86d6948705eb70abc97976c64c725b442'
copies=[copy(receipt,batch/'verification/ordinary-v3-receipt.json'),copy(Path('/tmp/barness-kimi-k31-spec-independent-v3.md'),batch/'verification/reviews/spec-independent-v3.md'),copy(work/'review-standards-v4.md',batch/'verification/reviews/standards-v4.md'),copy(Path(__file__),batch/'provenance/operations/package-ordinary-v1.py')]
refs=[]
for name in ['.evidence/catalog-kimi-k31-pro/e2e/20261102T024602.035474Z-3jcz6fs1/receipt.json','.evidence/catalog-kimi-k31-pro/e2e/20261102T025343.292240Z-w42x02ym/receipt.json','.evidence/catalog-kimi-k31-pro/e2e/20261102T023927.481768Z-jxy1dhxe/receipt.json','artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/required-cases-v3.json','artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/ordinary-inputs-v3.json','artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/run-ordinary-v3.py']:
 p=root/name;refs.append({'path':name,'sha256':sha(p)})
obj={'status':'PREPARED_NOT_COMPLETE','copies':copies,'originalReferences':refs,'ordinary':{'status':'PASS','exitCode':0,'actualMeasuredCommit':d['source']['commit'],'elapsedSeconds':d['elapsedSeconds'],'requiredCases':'PENDING_VERIFIER_RESULT','skippedCases':'PENDING_VERIFIER_RESULT'},'completeTestCapsulesIncluded':False,'note':'Original cancelled and invalid receipts remain at their source paths; generate/build/vet and final applicability remain pending.'}
p=batch/'verification/ordinary-reference-index-v1.json'
with p.open('x') as f:json.dump(obj,f,indent=2);f.write('\n')
print(json.dumps({'index':str(p.relative_to(root)),'sha256':sha(p),'copies':len(copies),'references':len(refs)}))
