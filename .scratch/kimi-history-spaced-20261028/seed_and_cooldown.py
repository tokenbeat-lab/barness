"""Keyless adoption and cooldown observations; never launch a paid producer."""
from common import *
from datetime import timedelta
import math

command=sys.argv[1]
if command=='seed':
 pre=load(B/'preflight.json');assert pre['pass_']
 f=frozen();results=B/'results';results.mkdir()
 for route in ORDER[:2]:
  old=R/'.evidence/kimi-history-live-20261028/results'/f'{route}.json';r=load(old)
  for field in ['producer','grant','candidateReceipt','candidateResult','transitiveInputs','composition']:check_ref(r[field])
  for n,h in load(R/r['transitiveInputs']['path']).items():assert sha(R/n)==h,n
  composition=R/r['composition']['path']
  q=readonly('reused-chat-inspections/'+route,[str(R/f['releasegate']['path']),'-inspect-bundle',str(composition.parent),'-mode','live','-stage','candidate'])
  assert q['qualification']=='CANDIDATE_COMPLETE' and len(q['coverage'])==10 and len(q['measurements'])==2
  inspection=B/'reused-chat-inspections'/route
  r.update(reusedFrom=ref(old),newGrantIssued=False,candidateReceipt=ref(inspection/'receipt.json'),candidateResult=ref(inspection/'stdout.jsonl'),measurements=q['measurements'],coverage=q['coverage'])
  save(results/f'{route}.json',r)
 # Keep the entire last paid sequence, including its overload and unknown bill.
 protected={}
 for path in (R/'.evidence/kimi-history-live-20261028').rglob('*'):
  if path.is_file():
   assert not path.is_symlink()
   with path.open('rb') as stream:protected[str(path.relative_to(R))]=hashlib.file_digest(stream,'sha256').hexdigest()
 save(B/'previous-paid-history-before.json',protected)
 print(json.dumps(dict(reusedChatCandidates=2,newCalls=0,newGrants=0,preservedPreviousFiles=len(protected))))
elif command=='cooldown':
 route=sys.argv[2];assert route in REMAINING
 done=[r for r in ORDER if (B/'results'/f'{r}.json').exists()]
 assert done==ORDER[:len(done)] and route==ORDER[len(done)]
 assert not (B/'STOP.json').exists() and not (B/'grants'/f'{route}.json').exists()
 if route==REMAINING[0]:
  anchor=ref(B/'approval.json');at=datetime.fromisoformat(load(B/'approval.json')['recordedAt'])
 else:
  prev=load(B/'results'/f'{ORDER[ORDER.index(route)-1]}.json')
  pp=check_ref(prev['producer']);anchor=ref(pp);at=datetime.fromisoformat(load(pp)['finishedAt'])
 earliest=at+timedelta(seconds=COOLDOWN_SECONDS);now=datetime.now(timezone.utc)
 remaining=max(0,math.ceil((earliest-now).total_seconds()))
 if remaining==0:
  d=B/'cooldowns';d.mkdir(exist_ok=True);p=d/f'{route}.json'
  if not p.exists():save(p,dict(route=route,status='ELAPSED',requiredSeconds=COOLDOWN_SECONDS,anchor=anchor,anchorTime=at.isoformat(),earliestGrantAt=earliest.isoformat(),observedAt=now.isoformat(),elapsedSeconds=(now-at).total_seconds(),grantNotYetIssued=True))
 print(json.dumps(dict(route=route,remainingSeconds=remaining,earliestGrantAt=earliest.isoformat(),ready=remaining==0,grantIssued=False)))
else:raise ValueError('select seed or cooldown')
