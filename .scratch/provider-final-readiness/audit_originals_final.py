"""Verify original failed producer and recover original bytes read-only."""
from pathlib import Path
import collections,hashlib,json,subprocess,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/provider-final-readiness';sys.path.insert(0,str(ROOT/'tools/verification'))
from run_process import environment
from receipt_files import source_snapshot
sha=lambda b:hashlib.sha256(b).hexdigest()
parent=ROOT/'.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id';receipt=json.loads((parent/'receipt.json').read_text());bundle=next((parent/'bundles').glob('*/minimax-report.json')).parent;report=json.loads((bundle/'minimax-report.json').read_text());grant=json.loads((bundle/'process-grant.json').read_text());env,_=environment('e2e')
assert receipt['status']=='FAIL' and receipt['exitCode']==1 and (parent/'exit-code').read_text().strip()=='1'
for n,h in receipt['files'].items():assert sha((parent/n).read_bytes())==h,n
assert receipt['source']['commit']==report['gitCommit']
for n,h in report['executionSource'].items():assert receipt['source']['files'].get(n,receipt.get('selectedInputs',{}).get(n))==h,n
argv=receipt['argv'];binary=Path(argv[3] if argv[:2]==['/usr/bin/env','-C'] else argv[0]);assert binary.is_file();assert sha(binary.read_bytes())==grant['scope']['binarySha256']
for b in receipt['bundles']:
 assert sha((parent/b['path']/'manifest.json').read_bytes())==b['manifestSHA256']
 assert sha((parent/b['path']/'audit.json').read_bytes())==b['auditSHA256']
missing=[];where=collections.Counter();items=[]
for kind,group in [('source',receipt['source']['files']),('input',receipt.get('selectedInputs',{}))]:
 for n,h in group.items():
  if Path(n).name=='.env':raise AssertionError('credential file must never be an evidence input')
  p=ROOT/n;origin=''
  if p.is_file() and not p.is_symlink() and sha(p.read_bytes())==h:origin='retained matching file'
  else:
   cmd=['git','show',receipt['source']['commit']+':'+n];r=subprocess.run(cmd,cwd=ROOT,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
   if r.returncode==0 and sha(r.stdout)==h:origin='original Git blob'
  if not origin:missing.append({'kind':kind,'path':n,'sha256':h})
  else:where[origin]+=1
  items.append({'kind':kind,'path':n,'sha256':h,'origin':origin or 'MISSING'})
value={'kind':'original-failed-producer-custody','source':source_snapshot(ROOT),'receipt':{'path':str((parent/'receipt.json').relative_to(ROOT)),'sha256':sha((parent/'receipt.json').read_bytes())},'originalExitCode':1,'originalOutcome':'FAIL','bundle':str(bundle.relative_to(ROOT)),'binary':{'path':str(binary.relative_to(ROOT)),'sha256':sha(binary.read_bytes())},'counts':dict(where),'checked':len(items),'items':items,'missing':missing,'grantsIssued':0,'claims':0,'vendorCalls':0}
with (BASE/'originals-audit-v2.json').open('x') as f:json.dump(value,f,indent=2,sort_keys=True);f.write('\n')
print(json.dumps({'checked':len(items),'origins':dict(where),'missing':missing}));assert not missing
