from pathlib import Path
import json,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement';sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
out=BASE/'release-process-v1';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out)
freeze=json.load(open(BASE/'build-v1/freeze.json'));argv=[str(ROOT/freeze['releasegate']['path']),'-record',str(BASE/'record-v1.json'),'-out',str(BASE/'release-v1')]
source=source_snapshot(ROOT);r=capture(argv,ROOT,env,out,180,[]);save(out/'receipt.json',dict(source=source,argv=argv,files={p.name:sha(p) for p in out.iterdir() if p.is_file()},expectedRefusal=True,**r))
report=json.load(open(BASE/'release-v1/release-report.json'));checks={c['id']:c for c in report['checks']};print({'exitCode':r['exitCode'],'pass':report['pass'],'checks':checks})
assert r['exitCode']==1 and not report['pass'] and checks['source']['pass'] and checks['e2e']['pass']
assert all(not checks[k]['pass'] for k in ['live','race','pressure'])
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
save(BASE/'delivery-v1.json',{'kind':'offline-implementation-delivery-paid-work-still-not-authorized','currentState':ref(ROOT/'.scratch/minimax-thinking-supplement/status.md'),'record':ref(BASE/'record-v1.json'),'releaseReport':ref(BASE/'release-v1/release-report.json'),'releaseProcess':ref(out/'receipt.json'),'verification':ref(BASE/'verification-and-retention-v1.json'),'matrix':ref(BASE/'delivery-matrix-v3.json'),'pendingScopes':ref(BASE/'pending-scopes-v3.json'),'implementationComplete':True,'controlledAndOrdinaryPassed':True,'releasePass':False,'realClaims':0,'paidGrantsIssued':0,'vendorCalls':0,'raceExecuted':False,'pressureExecuted':False,'pausedResumed':False})
