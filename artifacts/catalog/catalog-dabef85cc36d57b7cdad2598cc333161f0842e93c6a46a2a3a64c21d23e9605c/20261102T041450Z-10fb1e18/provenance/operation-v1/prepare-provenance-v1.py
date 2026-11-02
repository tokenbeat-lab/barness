#!/usr/bin/env python3
"""Collect fixed evidence copies and original receipt references; no completion marker."""
from pathlib import Path
import hashlib,json,shutil
ROOT=Path(__file__).resolve().parents[4];WORK=Path(__file__).resolve().parent
BATCH=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T041450Z-10fb1e18'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def copy(src,dst):
 dst.parent.mkdir(parents=True,exist_ok=True)
 with src.open('rb') as a,dst.open('xb') as b:shutil.copyfileobj(a,b)
 assert sha(src)==sha(dst)
 return {'source':str(src.relative_to(ROOT)),'copy':str(dst.relative_to(BATCH)),'sha256':sha(src)}
assert BATCH.is_dir() and not (BATCH/'delivery.json').exists()
copies=[]
for name in ['run-ordinary-v3.py','ordinary-inputs-v3.json','required-cases-v3.json','run-cli-matrix-v1.py','run-green-v4.py','run-red-v1.py','run-jev-old-red-v1.py','green-applicability-v1.py','green-applicability-v1.json','matrix-applicability-v1.py','matrix-applicability-v1.json']:
 copies.append(copy(WORK/'verification'/name,BATCH/'provenance/verification-v1'/name))
for name in ['operation.json','inputs/model-id-audit-v1.md','prepare-batch-v1.py','prepare-batch-result-v1.json','review-standards-v1.md','review-standards-v2.md','review-spec-v1.md','review-spec-v2.md','prepare-provenance-v1.py']:
 copies.append(copy(WORK/name,BATCH/'provenance/operation-v1'/name))
refs=[]
for name in ['20261102T041548.575496Z-zus0m6ad','20261102T041617.254114Z-aj6ywbrd','20261102T041721.865236Z-tplc27st','20261102T041903.793107Z-ccwlmiya','20261102T042005.753739Z-_exvvpl8','20261102T042109.497968Z-9nmol3cw','20261102T042122.569488Z-ulu96n25','20261102T042409.321856Z-84w6k84i']:
 p=ROOT/'.evidence/catalog-model-id-policy/e2e'/name/'receipt.json';data=json.loads(p.read_text());refs.append({'path':str(p.relative_to(ROOT)),'sha256':sha(p),'status':data['status'],'exitCode':data['exitCode']})
old=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6/delivery.json'
obj={'status':'PREPARED_NOT_COMPLETE','copies':copies,'originalReceipts':refs,'priorApprovedPackage':{'path':str(old.relative_to(ROOT)),'sha256':sha(old),'meaning':'Existing source provenance and historical checks only; not evidence of the new admission policy.'},'completeEvidenceSelfContained':False,'ordinary':'PENDING'}
p=BATCH/'provenance/operation-v1/index.json'
with p.open('x') as f:json.dump(obj,f,indent=2);f.write('\n')
print(json.dumps({'copies':len(copies),'receipts':len(refs),'indexSHA256':sha(p)}))
