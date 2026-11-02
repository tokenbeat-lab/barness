#!/usr/bin/env python3
"""Capture adopted-builtin evidence references without changing frozen inputs."""
from pathlib import Path
import hashlib,json,shutil
ROOT=Path(__file__).resolve().parents[4]; WORK=Path(__file__).resolve().parent
BATCH=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6'
def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def copy(src,dst):
 dst.parent.mkdir(parents=True,exist_ok=True)
 with src.open('rb') as a,dst.open('xb') as b: shutil.copyfileobj(a,b)
 assert digest(src)==digest(dst)
 return {'source':str(src.relative_to(ROOT)), 'destination':str(dst.relative_to(BATCH)), 'sha256':digest(src)}
assert BATCH.is_dir() and not (BATCH/'delivery.json').exists()
paths=['adopt-builtin-v1.py','commands/40-adopt-builtin/command.json','commands/40-adopt-builtin/stdout.jsonl','commands/40-adopt-builtin/stderr.log','status-before-builtin.md','review-spec-builtin-v1.md','review-standards-v3.md','verification/builtin-green-applicability-v1.json','verification/run-builtin-green-v1.py','verification/required-cases-v2.json']
copies=[copy(WORK/name,BATCH/'provenance/builtin-v1'/name) for name in paths]
copies.append(copy(Path(__file__),BATCH/'provenance/builtin-v1/package-builtin-index-v1.py'))
refs=[]
for name in ['.evidence/catalog-kimi-k31-pro/e2e/20261102T025153.030653Z-us31ebf4/receipt.json','.evidence/catalog-kimi-k31-pro/e2e/20261102T025230.476314Z-_wb4hku3/receipt.json','ai/e2e/testdata/catalog/pre-kimi-k31-builtin.json','ai/e2e/testdata/catalog/kimi-k31-approved.json','ai/examples/chineseeval/testdata/jev-1.13.0-v5.json','ai/examples/chineseeval/testdata/jev-route-v5.json','ai/examples/chineseeval/testdata/luna-route-v5.json']:
 p=ROOT/name;refs.append({'path':name,'sha256':digest(p)})
obj={'status':'PREPARED_NOT_COMPLETE','copies':copies,'originalReferences':refs,'ordinary':'PENDING','note':'Original receipts and complete case capsules remain external; no completion marker.'}
p=BATCH/'verification/builtin-evidence-index-v1.json'
with p.open('x') as f:json.dump(obj,f,indent=2);f.write('\n')
print(json.dumps({'copies':len(copies),'references':len(refs),'index':str(p.relative_to(ROOT)),'sha256':digest(p)}))
