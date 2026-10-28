"""Select all affected ordinary packages with retained explicit input closure."""
from pathlib import Path
import json,subprocess,sys
ROOT=Path(__file__).resolve().parents[2]
old=json.load(open(ROOT/'.evidence/minimax-thinking-supplement/full-ordinary-selection-v1.json'))
argv=old['argv'];inputs=[argv[i+1] for i,x in enumerate(argv) if x=='--input']
inputs+=['docs/adr/0055-route-scoped-evidence-applicability.md','docs/agents/verification.md','tools/verification/README.md','.scratch/minimax-thinking-supplement/spec.md','.scratch/minimax-thinking-supplement-plan/plan.md']
inputs+=['.scratch/kimi-composition-promotion/spec.md','ai/live/REGIONAL.md']
inputs=sorted(set(inputs))
cmd=['python3','tools/verification/run.py','--feature','kimi-composition-promotion','--stage','e2e']
for p in inputs:cmd+=['--input',p]
cmd+=['--','go','test','-json','./ai/e2e','./ai','./ai/internal/testkit/release','./ai/internal/testkit/supportmatrix','./ai/internal/testkit/minimaxlive','./ai/internal/testkit/livegrant','./ai/internal/testkit/jevformal','./ai/examples/visualeval','-count=1','-failfast','-timeout=20m']
out=ROOT/'.evidence/kimi-composition-promotion/full-ordinary-selection-v1.json'
with out.open('x') as f:json.dump({'argv':cmd,'inputs':inputs,'basis':'complete previous seven packages plus livegrant; shared grant duration/optional fields, producer and consumer changes; no opt-in stages'},f,indent=2);f.write('\n')
sys.exit(subprocess.run(cmd,cwd=ROOT).returncode)
