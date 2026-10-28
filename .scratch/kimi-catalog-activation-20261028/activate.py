"""Use the frozen native activation CLI with six pinned real compositions."""
from pathlib import Path
from datetime import datetime,timezone
import json,sys,time
R=Path(__file__).resolve().parents[2];S=R/'.scratch/kimi-catalog-activation-20261028';B=R/'.evidence/kimi-catalog-activation-20261028'
sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import sha,source_snapshot,save
from run_process import environment,capture
load=lambda p:json.loads(p.read_text())
ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
f=load(B/'freeze.json');approval=load(B/'approval.json')
assert sha(R/f['regionalpromote']['path'])==f['regionalpromote']['sha256']
assert sha(R/'ai/catalog_kimi.json')==approval['catalogBefore']['sha256']=='c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109'
current=source_snapshot(R);before=load(B/'source-before.json')
changes=[n for n in before['files'].keys()|current['files'].keys() if before['files'].get(n)!=current['files'].get(n)]
expected=['ai/e2e/kimi_qualification_promotion_test.go','ai/e2e/kimi_host_rollback_test.go','ai/e2e/kimi_composition_promotion_test.go','ai/e2e/kimi_history_composition_test.go']
assert set(changes)==set(expected),changes
dirs=[]
for vr in approval['candidateCompositions']:
 p=R/vr['path'];assert sha(p)==vr['sha256'];dirs.append(str(p.parent))
receiptDir=R/'.evidence/barness-ai-kimi/promotion'/('activation-'+datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ'));receiptDir.mkdir(mode=0o700,parents=True)
intent=receiptDir/'activation-intent.json'
out=B/'native-activation';out.mkdir();argv=[str(R/f['regionalpromote']['path']),'-kimi','-receipt',str(intent),*dirs]
env,_=environment('e2e');start=datetime.now(timezone.utc).isoformat();tic=time.monotonic()
r=capture(argv,R,env,out,180,[])
if r['exitCode'] is not None:(out/'exit-code').write_text(str(r['exitCode'])+'\n')
save(out/'receipt.json',dict(argv=argv,sourceBeforeActivation=current,binary=f['regionalpromote'],approval=ref(B/'approval.json'),startedAt=start,finishedAt=datetime.now(timezone.utc).isoformat(),elapsedSeconds=time.monotonic()-tic,files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r))
assert r['status']=='PASS' and r['exitCode']==0,(out/'stderr.log').read_text()
proof=load(intent);catalog=load(R/'ai/catalog_kimi.json')
assert proof['schema']==2 and proof['mode']=='live' and proof['kind']=='verified-kimi-catalog-activation-intent'
assert len(proof['candidates'])==6 and sum(len(c['qualification']['measurements']) for c in proof['candidates'])==12
assert all(len(c['qualification']['coverage'])==10 for c in proof['candidates'])
assert catalog['active'] is True and len(catalog['models'])==6 and sha(R/'ai/catalog_kimi.json')==proof['catalogSha256']
save(B/'activation.json',dict(pass_=True,nativeReceipt=ref(out/'receipt.json'),intent=ref(intent),catalogBefore=approval['catalogBefore'],catalogAfter=ref(R/'ai/catalog_kimi.json'),catalogVersion=catalog['version'],identities=[dict(provider=m['provider'],api=m['api'],model=m['id']) for m in catalog['models']],candidateCompositions=6,independentMeasurements=12,vendorCalls=0,grantsIssued=0))
(S/'status.md').write_text('# 当前状态\n\n目录原生激活成功：active=true，kimi-k3-formal-2026-10-22.1，六个Kimi自身身份；六组真实9＋1及十二份测量均已由冻结regionalpromote验证。原目录精确备份与独占晋升证明保留，实际退出0。正在重建内嵌新目录的只读工具，并执行正式离线消费和受影响普通验证。未读取凭据、签发grant或请求厂商；race/压力及暂停恢复未执行。\n')
print(json.dumps(dict(activated=True,version=catalog['version'],models=len(catalog['models']),intent=str(intent.relative_to(R)),exitCode=r['exitCode'],vendorCalls=0)))
