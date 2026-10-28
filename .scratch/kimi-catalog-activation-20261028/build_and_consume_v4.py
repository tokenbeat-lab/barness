"""Build the activated readonly consumer and capture six formal qualifications."""
from pathlib import Path
from datetime import datetime,timezone
import json,sys,time
R=Path(__file__).resolve().parents[2];B=R/'.evidence/kimi-catalog-activation-20261028'
sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,isolate_go_home,capture
load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
assert load(B/'activation.json')['pass_']
out=B/'build-v4';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
source=source_snapshot(R);binary=out/'releasegate';argv=['go','build','-o',str(binary),'./ai/release/cmd/releasegate'];started=datetime.now(timezone.utc).isoformat();tic=time.monotonic()
r=capture(argv,R,env,out,180,[])
save(out/'receipt.json',dict(argv=argv,source=source,startedAt=started,finishedAt=datetime.now(timezone.utc).isoformat(),elapsedSeconds=time.monotonic()-tic,files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r))
assert r['status']=='PASS' and r['exitCode']==0,(out/'stderr.log').read_text()
assert source_snapshot(R)==source
save(out/'binary.json',dict(binary=ref(binary),source=source,kind='readonly consumer embeds activated catalog; never a paid producer'))
completion=load(R/'.evidence/kimi-history-spaced-20261028/completion.json');consumers=B/'formal-consumers-v4';consumers.mkdir();rows=[]
for row in completion['routes']:
 vr=row['composition'];p=R/vr['path'];assert sha(p)==vr['sha256'];bundle=str(p.parent.relative_to(R));target=consumers/row['route'];logs=B/('capture-v4-'+row['route']);logs.mkdir()
 argv=[sys.executable,'tools/verification/consume_live.py','--binary',str(binary),'--bundle',bundle,'--out',str(target.relative_to(R))]
 started=datetime.now(timezone.utc).isoformat();tic=time.monotonic();r=capture(argv,R,env,logs,180,[])
 save(logs/'receipt.json',dict(argv=argv,startedAt=started,finishedAt=datetime.now(timezone.utc).isoformat(),elapsedSeconds=time.monotonic()-tic,files={p.name:sha(p) for p in logs.iterdir() if p.is_file()},**r))
 assert r['status']=='PASS' and r['exitCode']==0,(logs/'stderr.log').read_text()
 outer=load(target/'live-consumer.json');native=load(target/'native/live-consumer.json')
 assert outer['schema']==2 and outer['pass'] and outer['exitCode']==0 and native['pass']
 rows.append(dict(route=row['route'],outer=ref(target/'live-consumer.json'),native=ref(target/'native/live-consumer.json'),capture=ref(logs/'receipt.json'),originalComposition=vr))
 print(json.dumps(dict(route=row['route'],formal='PASS',nativeExitCode=outer['exitCode'],vendorCalls=0)),flush=True)
save(B/'formal-qualification-v4.json',dict(pass_=True,consumer=ref(binary),source=source,consumers=rows,newVendorCalls=0,newGrants=0))
