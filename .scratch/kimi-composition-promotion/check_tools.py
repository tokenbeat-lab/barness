from pathlib import Path
import json,sys
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
out=ROOT/'.evidence/kimi-composition-promotion/tools-vet-v2';out.mkdir()
env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off');source=source_snapshot(ROOT)
commands=[['python3','-m','unittest','discover','-s','tools/verification/tests'],['go','vet','./ai/e2e','./ai/internal/testkit/livegrant','./ai/internal/testkit/minimaxlive','./ai/internal/testkit/supportmatrix','./ai/internal/testkit/release','./ai/live/cmd/processgrant','./ai/live/cmd/regionalpromote','./ai/release/cmd/releasegate'],['go','vet','-tags','live','./ai/live']]
results=[]
for i,argv in enumerate(commands):
 d=out/str(i+1);d.mkdir();r=capture(argv,ROOT,env,d,300,[]);save(d/'receipt.json',dict(argv=argv,source=source,compileOnly=argv[1]=='vet',files={p.name:sha(p) for p in d.iterdir() if p.is_file()},**r));results.append(r)
 if r['exitCode']!=0:print(r);raise SystemExit(1)
assert source==source_snapshot(ROOT)
save(out/'receipt.json',dict(source=source,commands=commands,results=results,vendorCalls=0,grantsIssued=0,claims=0));print({'checks':len(results),'exitCodes':[r['exitCode'] for r in results]})
