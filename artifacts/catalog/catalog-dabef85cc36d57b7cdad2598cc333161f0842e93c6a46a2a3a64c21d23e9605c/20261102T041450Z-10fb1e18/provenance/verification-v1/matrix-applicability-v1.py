#!/usr/bin/env python3
"""Compare the matrix execution source and selected inputs to frozen HEAD."""
import json,pathlib,sys
root=pathlib.Path.cwd();sys.path.insert(0,str(root/'tools/verification'))
from receipt_files import sha,source_snapshot
p=root/'.evidence/catalog-model-id-policy/e2e/20261102T042409.321856Z-84w6k84i/receipt.json';r=json.loads(p.read_text());now=source_snapshot(root);a=r['source']['files'];b=now['files'];delta=[n for n in sorted(set(a)|set(b)) if a.get(n)!=b.get(n)];bad=[]
for n,h in r['selectedInputs'].items():
 try:
  if sha(root/n)!=h:bad.append(n)
 except Exception:bad.append(n)
ok=not delta and not bad and now['commit']=='807afbc548ad9410a090121a8671bdb189f50128'
result={'status':'PASS' if ok else 'FAIL','receipt':str(p.relative_to(root)),'receiptSha256':sha(p),'measuredCommit':r['source']['commit'],'frozenCommit':now['commit'],'sourceCount':len(b),'sourceDifferences':delta,'selectedInputCount':len(r['selectedInputs']),'selectedInputMismatches':bad}
out=root/'artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/matrix-applicability-v1.json';out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),'resultSha',sha(out),'scriptSha',sha(pathlib.Path(__file__)));sys.exit(not ok)
