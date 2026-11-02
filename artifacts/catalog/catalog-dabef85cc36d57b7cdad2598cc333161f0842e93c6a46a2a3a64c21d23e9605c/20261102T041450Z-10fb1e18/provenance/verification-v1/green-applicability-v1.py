#!/usr/bin/env python3
"""Compare the measured targeted source and inputs to the frozen checkout."""
import json,pathlib,sys
root=pathlib.Path.cwd();sys.path.insert(0,str(root/'tools/verification'))
from receipt_files import sha,source_snapshot
p=root/'.evidence/catalog-model-id-policy/e2e/20261102T042122.569488Z-ulu96n25/receipt.json';r=json.loads(p.read_text());now=source_snapshot(root);before=r['source']['files'];after=now['files'];delta=[n for n in sorted(set(before)|set(after)) if before.get(n)!=after.get(n)];bad=[]
for n,h in r['selectedInputs'].items():
 try:
  if sha(root/n)!=h:bad.append(n)
 except Exception:bad.append(n)
ok=not delta and not bad and now['commit']=='65a1ecbb2cb97f51c37cfee849925712187e5acc'
result={'status':'PASS' if ok else 'FAIL','receipt':str(p.relative_to(root)),'receiptSha256':sha(p),'measuredCommit':r['source']['commit'],'frozenCommit':now['commit'],'sourceCount':len(after),'sourceDifferences':delta,'selectedInputCount':len(r['selectedInputs']),'selectedInputMismatches':bad}
out=root/'artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/green-applicability-v1.json';out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),'resultSha',sha(out),'scriptSha',sha(pathlib.Path(__file__)));sys.exit(not ok)
