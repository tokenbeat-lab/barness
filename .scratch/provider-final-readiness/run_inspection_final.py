"""Run existing native per-case checks without weakening full qualification."""
from pathlib import Path
import json,sys,tempfile
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
base=ROOT/'.evidence/provider-final-readiness/native-analysis-v2';base.mkdir()
module=Path(tempfile.mkdtemp(prefix='barness-readiness-inspection-'))
(module/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/readinessinspection\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n')
(module/'main.go').write_bytes((ROOT/'.scratch/provider-final-readiness/inspect.go').read_bytes())
overlay={"Replace":{str(ROOT/'ai/internal/testkit/minimaxlive/readiness_audit.go'):str(ROOT/'.scratch/provider-final-readiness/audit_overlay.go')}}
save(base/'overlay.json',overlay)
inputs={str(p.relative_to(ROOT)):sha(p) for p in (ROOT/'.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id').rglob('*') if p.is_file()}
for n in ['audit_overlay.go','inspect.go','run_inspection.py']:p=ROOT/'.scratch/provider-final-readiness'/n;inputs[str(p.relative_to(ROOT))]=sha(p)
env,_=environment('e2e');isolate_go_home(env,'e2e',base);env.update(GOPROXY='off',GOSUMDB='off');source=source_snapshot(ROOT)
argv=['go','run','-mod=mod','-overlay',str(base/'overlay.json'),'.'];r=capture(argv,module,env,base,180,[])
save(base/'receipt.json',dict(source=source,inputs=inputs,argv=argv,cwd=str(module),files={p.name:sha(p) for p in base.iterdir() if p.is_file()},**r));print(r);raise SystemExit(0 if r['exitCode']==0 else 1)
