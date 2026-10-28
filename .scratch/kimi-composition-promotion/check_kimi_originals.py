from pathlib import Path
import json,sys,tempfile
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
out=ROOT/'.evidence/kimi-composition-promotion/kimi-original-producers-v2';out.mkdir();mod=Path(tempfile.mkdtemp(prefix='barness-kimi-producer-audit-'))
(mod/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/kimiproduceraudit\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n')
program=ROOT/'.scratch/provider-final-readiness/check_kimi_originals.go';(mod/'main.go').write_bytes(program.read_bytes());wrapper=ROOT/'.scratch/provider-final-readiness/kimi_producer_audit_v2.go'
save(out/'overlay.json',{'Replace':{str(ROOT/'ai/internal/testkit/supportmatrix/readiness_producer_audit.go'):str(wrapper)}})
env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off');source=source_snapshot(ROOT)
argv=['go','run','-mod=mod','-overlay',str(out/'overlay.json'),'.'];r=capture(argv,mod,env,out,180,[])
inputs={str(p.relative_to(ROOT)):sha(p) for p in [program,wrapper,Path(__file__)]};selection=json.load(open(ROOT/'.evidence/provider-delivery-offline/selected-evidence-v1.json'))
for row in selection['kimi']:
 p=ROOT/row['original']['receipt']['path'];inputs[str(p.relative_to(ROOT))]=sha(p)
 for p in (ROOT/row['bundle']).rglob('*'):
  if p.is_file():inputs[str(p.relative_to(ROOT))]=sha(p)
save(out/'receipt.json',dict(source=source,argv=argv,cwd=str(mod),inputs=inputs,files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r));print(r);raise SystemExit(0 if r['exitCode']==0 else 1)
