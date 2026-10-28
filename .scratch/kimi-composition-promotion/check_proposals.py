from pathlib import Path
import json,tempfile,sys
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
out=ROOT/'.evidence/kimi-composition-promotion/proposal-validation-v2';out.mkdir()
with tempfile.TemporaryDirectory() as td:
 td=Path(td);(td/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/minimaxproposalinspection\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n');(td/'main.go').write_bytes((ROOT/'.scratch/kimi-composition-promotion/check_proposals.go').read_bytes())
 env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
 argv=['go','run','-mod=mod','.',str(ROOT)];r=capture(argv,td,env,out,180,[])
 inputs={str(p.relative_to(ROOT)):sha(p) for p in (ROOT/'.scratch/kimi-composition-promotion/pending-scopes-v5').glob('*.json')}
 save(out/'receipt.json',dict(source=source_snapshot(ROOT),argv=argv,cwd=str(td),inputs=inputs,files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r));print(r);raise SystemExit(0 if r['exitCode']==0 else 1)
