"""Capture read-only current source/SDK/catalog facts for six proposals."""
from pathlib import Path
import json,sys,tempfile
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
out=ROOT/'.evidence/provider-final-readiness/native-facts-v2';out.mkdir();mod=Path(tempfile.mkdtemp(prefix='barness-readiness-facts-'))
(mod/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/readinessfacts\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n')
program=ROOT/'.scratch/provider-final-readiness/native_facts.go';(mod/'main.go').write_bytes(program.read_bytes())
(out/'go.mod.txt').write_bytes((mod/'go.mod').read_bytes())
env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off');argv=['go','run','-mod=mod','.'];source=source_snapshot(ROOT);r=capture(argv,mod,env,out,180,[])
save(out/'receipt.json',dict(source=source,argv=argv,cwd=str(mod),inputs={str(program.relative_to(ROOT)):sha(program),str(Path(__file__).relative_to(ROOT)):sha(__file__)},files={p.name:sha(p) for p in out.iterdir() if p.is_file()},**r));print(r);assert r['exitCode']==0
