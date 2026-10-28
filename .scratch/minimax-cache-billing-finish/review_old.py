from pathlib import Path
import json,sys,tempfile
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-cache-billing-finish';S=R/'.scratch/minimax-cache-billing-finish'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot,sha
from run_process import environment,capture,isolate_go_home
out=B/'old-capture-review-v1';out.mkdir();source=source_snapshot(R)
with tempfile.TemporaryDirectory() as td:
 td=Path(td);(td/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/minimaxofflineassessment\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(R)+'\n')
 (td/'main.go').write_bytes((S/'offline_review_main.go').read_bytes())
 overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(R/'ai/internal/testkit/minimaxlive/offline_assessment.go'):str(S/'offline_assessment.go')}})+'\n')
 env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
 argv=['go','run','-mod=mod','-overlay',str(overlay),'.',str(R)];result=capture(argv,td,env,out,180,[])
 assert source==source_snapshot(R)
 with (out/'receipt.json').open('x') as f:json.dump({'argv':argv,'source':source,'inputs':{str(p.relative_to(R)):sha(p) for p in [S/'offline_assessment.go',S/'offline_review_main.go',Path(__file__)]},'files':{p.name:sha(p) for p in out.iterdir() if p.is_file()},'diagnosticOverlayOnly':True,'networkRequests':0,**result},f,indent=2);f.write('\n')
 print(result);raise SystemExit(0 if result['exitCode']==0 else 1)
