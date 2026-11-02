#!/usr/bin/env python3
"""Compare current source/selected bytes to the completed ordinary receipt."""
import json,pathlib,sys
root=pathlib.Path.cwd();sys.path.insert(0,str(root/'tools/verification'))
from receipt_files import sha,source_snapshot
receipt=root/'.evidence/catalog-kimi-k31-pro/e2e/20261102T030707.838552Z-x6dz4dj5/receipt.json';old=json.loads(receipt.read_text());now=source_snapshot(root);oldfiles=old['source']['files'];newfiles=now['files'];sourceDelta=[{'path':p,'measured':oldfiles.get(p),'current':newfiles.get(p)} for p in sorted(set(oldfiles)|set(newfiles)) if oldfiles.get(p)!=newfiles.get(p)];bad=[]
for p,h in old['selectedInputs'].items():
 try:
  if sha(root/p)!=h:bad.append(p)
 except Exception:bad.append(p)
result={'ordinaryReceipt':str(receipt.relative_to(root)),'ordinaryReceiptSha256':sha(receipt),'measuredCommit':old['source']['commit'],'currentCommit':now['commit'],'sourceCount':len(newfiles),'measuredSourceCount':len(oldfiles),'sourceSetAndHashesEqual':not sourceDelta,'sourceDelta':sourceDelta,'selectedInputCount':len(old['selectedInputs']),'selectedInputMismatches':bad,'status':'PASS' if not sourceDelta and not bad else 'FAIL','comparison':'file set and bytes independently of commit metadata'}
out=root/'artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/post-doc-equality-v1.json';out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'path':str(out.relative_to(root)),'sha256':sha(out),**result}));sys.exit(result['status']!='PASS')
