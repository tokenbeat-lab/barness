"""Current-source offline/controlled checks, with actual subprocess receipts."""
from datetime import datetime, timezone
import json
from pathlib import Path
import sys
import time
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot, save, sha
from run_process import capture, environment, isolate_go_home
BASE=ROOT/'.evidence/kimi-history-supplement'
source=source_snapshot(ROOT)
selection=json.loads((ROOT/'.evidence/provider-delivery-offline/selected-evidence-v1.json').read_text())['kimi']
inputs={'.evidence/provider-delivery-offline/selected-evidence-v1.json':sha(ROOT/'.evidence/provider-delivery-offline/selected-evidence-v1.json')}
for row in selection:
 for p in (ROOT/row['bundle']).rglob('*'):
  if p.is_file(): inputs[str(p.relative_to(ROOT))]=sha(p)
checks=[('tools-v1',[sys.executable,'-m','unittest','discover','-s','tools/verification/tests','-v'],360,{}),('vet-v1',['go','vet','./ai','./ai/e2e','./ai/internal/testkit/release','./ai/internal/testkit/supportmatrix','./ai/internal/testkit/jevformal','./ai/internal/testkit/minimaxlive','./ai/examples/visualeval','./ai/live/cmd/processgrant','./ai/release/cmd/releasegate'],240,{}),('vet-live-v1',['go','vet','-tags','live,kimi_baseline_fixture','./ai/live'],240,{}),('public-replay-v2',['go','run','./.scratch/kimi-history-supplement/replay.go'],180,dict(inputs,**{'.scratch/kimi-history-supplement/replay.go':sha(ROOT/'.scratch/kimi-history-supplement/replay.go')}))]
for name,argv,timeout,selected in checks:
 out=BASE/name;out.mkdir()
 env,_=environment('e2e');isolate_go_home(env,'e2e',out)
 started=datetime.now(timezone.utc).isoformat();tic=time.monotonic()
 result=capture(argv,ROOT,env,out,timeout,[])
 files={p.name:sha(p) for p in out.iterdir() if p.is_file()}
 save(out/'receipt.json',dict(source=source,argv=argv,inputs=selected,startedAt=started,finishedAt=datetime.now(timezone.utc).isoformat(),elapsedSeconds=time.monotonic()-tic,files=files,**result))
 print(json.dumps({'check':name,**result}),flush=True)
 if result['exitCode']!=0: raise SystemExit(1)
 assert source_snapshot(ROOT)==source, 'source changed during checks'
