"""Session scaffolding; frozen tools own native scope, production and inspection."""
from pathlib import Path
from datetime import datetime, timezone
import hashlib, json, os, subprocess, sys
R=Path(__file__).resolve().parents[2]
B=R/'.evidence/kimi-history-paid-20261028'
S=R/'.scratch/kimi-history-paid-20261028'
sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import source_snapshot, sha
from run_process import environment, capture
ORDER=['kimi-cn-chat','kimi-intl-chat','kimi-cn-responses','kimi-intl-responses','kimi-cn-messages','kimi-intl-messages']
PINS=dict(zip(ORDER,['7833ffc441fdbf4a358d9236d50f0438016a2d4638b36dee87143b02fe2cbf80','27ae799e6cfcf80ffdd8a3b91d022b3b247468b7d4ddedce3944167a0acbb9a2','aa3b5cd103bb50a6607c50f0ea318819dea109147ba421a3ff635a9adf6f07d0','29edfe07d759da1f51517045b1c8ff9a811c33d0737f6b7d1744252d7e58965a','9c1f45795318e7acd4edfad709b69779bc78a0799227ce95be33eb240ed2c82a','bb05c4a746dc518b3a8fe32e5fd0c3979e9d7bc182280e3695c6b31d141a1cf9']))
FREEZE=R/'.evidence/minimax-cache-billing-finish/build-v1/freeze.json'
NATIVE='a2e26473a567b8bef8a7f21e03e6e2db95b7f625f41a8ab3bce101f6352fe5c2'
load=lambda p:json.loads(Path(p).read_text())
def save(p,d):
 p=Path(p)
 with p.open('x') as f:
  json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n');f.flush();os.fsync(f.fileno())
def ref(p):
 p=Path(p);return {'path':str(p.relative_to(R)),'sha256':sha(p)}
def check_ref(v):
 p=R/v['path'];assert sha(p)==v['sha256'],str(p);return p

def state(text):
 (S/'status.md').write_text('# 当前状态\n\n'+text+'\n')
def stop(reason,route=None):
 if not (B/'STOP.json').exists():save(B/'STOP.json',dict(kind='stop-all-remaining-paid-work',recordedAt=datetime.now(timezone.utc).isoformat(),reason=reason,route=route,newExplicitApprovalRequired=True,noRetryOrReplacementGrant=True))
 state('本轮已停止。'+reason+'。不得签发后续grant或重试；保留全部实际消费、失败、授权与账本。证据见[STOP](../../.evidence/kimi-history-paid-20261028/STOP.json)。race、压力、目录激活、正式消费与暂停恢复均未执行。')
def frozen():
 assert sha(FREEZE)=='5c84442aa5534aef5951d4c48e31bb081f18fdc80f37ce3c4453464f9c7e007c','freeze mismatch'
 f=load(FREEZE);assert source_snapshot(R)==f['source'],'current source mismatch'
 for k in ['kimi','processgrant','releasegate','regionalpromote']:check_ref(f[k])
 assert f['kimi']['sha256']=='d0919abbd1844775aa4e789aa2b6e7be638ca41203fd74de11789a7df6716777'
 assert f['regionalSource']['sourceSha256']==NATIVE
 for n,h in f['regionalSource']['files'].items():assert sha(R/n)==h,n
 return f

def readonly(name,argv,timeout=120):
 out=B/name;out.mkdir(parents=True);env,_=environment('e2e')
 r=capture(argv,R,env,out,timeout,[])
 save(out/'receipt.json',dict(argv=argv,binary=ref(Path(argv[0])),source=source_snapshot(R),files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r))
 assert r['status']=='PASS' and r['exitCode']==0,name+': '+(out/'stderr.log').read_text()[:500]
 return json.loads((out/'stdout.jsonl').read_text())

def stores():
 roots=load(R/'.evidence/kimi-history-supplement-plan/store-roots.json')
 return {str(p.relative_to(R)):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}
