"""Run the current read-only native baseline validator; no producer or credential."""
import json
from pathlib import Path
import sys
import tempfile
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'tools/verification'))
from run_process import capture,environment
from receipt_files import source_snapshot,save,sha
base=ROOT/'.evidence/kimi-history-supplement/native-baselines-v1';base.mkdir()
module=Path(tempfile.mkdtemp(prefix='barness-kimi-baselines-'))
(module/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/baselineinspection\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n')
program='''package main
import("encoding/json";"os";"fmt";"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix")
func main(){os.Chdir("/Users/cyber/RestoX/barness");raw,_:=os.ReadFile(".evidence/provider-delivery-offline/selected-evidence-v1.json");var s struct{Kimi []struct{Route,Bundle string}};json.Unmarshal(raw,&s);failed:=false;var out []map[string]any;for _,v:=range s.Kimi{r,e:=supportmatrix.VerifyKimiBaselineCapture(v.Bundle,"live");msg:="";if e!=nil{msg=e.Error();failed=true};out=append(out,map[string]any{"route":v.Route,"cases":len(r.Scenarios),"error":msg,"pass":e==nil,"vendorCalls":0})};b,_:=json.MarshalIndent(out,"","  ");fmt.Println(string(b));if failed{os.Exit(1)}}
'''
(module/'main.go').write_text(program);(base/'inspection.go.txt').write_text(program)
inputs={}
for row in json.loads((ROOT/'.evidence/provider-delivery-offline/selected-evidence-v1.json').read_text())['kimi']:
 for p in (ROOT/row['bundle']).rglob('*'):
  if p.is_file():inputs[str(p.relative_to(ROOT))]=sha(p)
env,_=environment('e2e');env.update(GOPROXY='off',GOSUMDB='off');source=source_snapshot(ROOT);argv=['go','run','-mod=mod','.'];r=capture(argv,module,env,base,180,[])
save(base/'receipt.json',dict(argv=argv,cwd=str(module),source=source,inputs=inputs,programSHA256=sha(module/'main.go'),**r))
print((base/'stdout.jsonl').read_text());print((base/'stderr.log').read_text()[-1200:]);raise SystemExit(0 if r['exitCode']==0 else 1)
