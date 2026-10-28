from pathlib import Path
import hashlib,json,time
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement';started=time.time()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
checks={}
for label,path in [('allHistorical','.evidence/provider-final-readiness/protected-history-before.json'),('taskProtected','.evidence/minimax-thinking-supplement/initial/protected-inputs-before.json')]:
 protected=json.load(open(ROOT/path));missing=[];changed=[]
 for n,h in protected.items():
  p=Path(n);p=p if p.is_absolute() else ROOT/p
  if not p.is_file():missing.append(n)
  elif sha(p)!=h:changed.append(n)
 checks[label]={'protected':len(protected),'missing':missing,'changed':changed}
old=json.load(open(BASE/'initial/public-stores-before.json'));roots=json.load(open(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json'))
new={str(p):sha(p) for r in roots for p in Path(r).rglob('*') if p.is_file()}
checks['publicStores']={'protected':len(old),'added':sorted(new.keys()-old.keys()),'missing':sorted(old.keys()-new.keys()),'changed':[n for n,h in old.items() if n in new and new[n]!=h]}
passed=all(not any(v for k,v in c.items() if k!='protected') for c in checks.values())
value={'kind':'historical-evidence-and-public-authority-retention','pass':passed,'checks':checks,'elapsedSeconds':time.time()-started,'vendorCalls':0,'paidGrantsIssued':0,'realClaims':0,'race':False,'pressure':False,'pausedResumed':False}
with (BASE/'history-protection-final-v1.json').open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(value));assert passed
