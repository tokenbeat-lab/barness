from pathlib import Path
import sys,json
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-cache-billing-finish';sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import source_snapshot,sha
from run_process import environment,capture,isolate_go_home
out=B/'vet-v1';out.mkdir();source=source_snapshot(R);env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
commands=[['go','vet','./ai/e2e','./ai/internal/testkit/minimaxlive','./ai/internal/testkit/supportmatrix','./ai/release/cmd/releasegate'],['go','vet','-tags','live','./ai/live']];results=[]
for i,argv in enumerate(commands):
 d=out/str(i+1);d.mkdir();r=capture(argv,R,env,d,240,[]);results.append({'argv':argv,**r});assert r['exitCode']==0,r
assert source_snapshot(R)==source
with (out/'receipt.json').open('x') as f:json.dump({'source':source,'compileOnly':True,'results':results,'vendorCalls':0,'testExecutions':0},f,indent=2);f.write('\n')
print({'vetExitCodes':[r['exitCode'] for r in results]})
