"""Read-only fixed 13+1 consumption after the one paid producer has passed."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,sys,time
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-thinking-resample';S=R/'.scratch/minimax-thinking-resample'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
from run_process import environment,capture,isolate_go_home
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False);f.write('\n')
info=load(B/'attempt-index.json');assert info['producerPass'] and not (B/'STOP.json').exists();rp=R/info['nativeReceipt']['path'];assert sha(rp)==info['nativeReceipt']['sha256'];receipt=load(rp);assert receipt['exitCode']==0 and receipt['status']=='PASS' and len(receipt['bundles'])==1
native=rp.parent/receipt['bundles'][0]['path'];report=load(native/'minimax-report.json');assert report['outcome']=='PASS' and not report['complete'] and report['mode']=='live' and report['budget']['calls']==report['budget']['attempts']==1 and not load(native/'audit.json')['findings']
files={str(p.relative_to(native)):sha(p) for p in native.rglob('*') if p.is_file()};assert load(native/'minimax-integrity.json')['files']=={n:h for n,h in files.items() if n not in ['minimax-integrity.json','live-report.json']}
save(B/'new-native-files-before-consumption.json',files)
pre=load(B/'preflight.json');freeze=load(R/pre['freeze']['path']);source=source_snapshot(R);assert source==pre['source'];consumer=R/freeze['releasegate']['path'];assert sha(consumer)==freeze['releasegate']['sha256']
old=pre['baseline'];base={'bundle':old['bundle'],'manifestSHA256':old['manifest']['sha256'],'reportSHA256':old['report']['sha256'],'sealSHA256':old['seal']['sha256'],'producer':old['producer']}
single={'bundle':str(native.relative_to(R)),'manifestSHA256':sha(native/'manifest.json'),'reportSHA256':sha(native/'minimax-report.json'),'sealSHA256':sha(native/'minimax-integrity.json'),'producer':ref(rp)}
descriptor=B/'composition-v1';descriptor.mkdir();save(descriptor/'minimax-thinking-composition.json',{'schema':1,'kind':'minimax-messages-thinking-composition','combo':'minimax-messages','model':'MiniMax-M3','baseline':base,'supplement':single})
env,_=environment('e2e')
def execute(out,argv,timeout):
 out.mkdir();isolated=dict(env);isolate_go_home(isolated,'e2e',out);start=datetime.now(timezone.utc);tic=time.monotonic();result=capture(argv,R,isolated,out,timeout,[]);streams={p.name:sha(p) for p in out.iterdir() if p.is_file()};save(out/'receipt.json',{'argv':argv,'source':source,'startedAt':start.isoformat(),'finishedAt':datetime.now(timezone.utc).isoformat(),'elapsedSeconds':time.monotonic()-tic,'files':streams,'vendorCalls':0,**result});return result
inspection=B/'inspection-v1';argv=[str(consumer),'-inspect-bundle',str(descriptor),'-mode','live','-stage','formal'];r=execute(inspection,argv,180)
if r['exitCode']!=0 or r['status']!='PASS':
 save(B/'STOP.json',{'kind':'read-only-fixed-composition-refused','process':ref(inspection/'receipt.json'),'noFurtherPaidAuthority':True});raise SystemExit(1)
q=load(inspection/'stdout.jsonl');assert q['qualification']=='FORMAL_APPLICABLE' and q['mode']=='live' and len(q['measurements'])==2 and len(q['coverage'])==14 and len(q['excludedFailures'])==1
assert q['supplementBilling']['independentCacheWriteFee']=='NOT_APPLICABLE' and q['supplementBilling']['cacheCreationCounter']=='UNREPORTED' and q['supplementBilling']['usageReporting']=='partial'
assert q['measurements'][0]['original']['outcome']=='FAIL' and q['measurements'][1]['original']['outcome']=='PASS' and not q['measurements'][1]['original']['complete']
print(json.dumps({'inspection':'FORMAL_APPLICABLE','coverage':14,'measurements':2,'excludedOriginalFailures':1,'vendorCalls':0}),flush=True)
# Canonical outer capture binds one native consumer and its complete input closure.
consumerOut=B/'consumer-v1';launch=B/'consumer-launch-v1';argv=['python3','tools/verification/consume_live.py','--binary',str(consumer),'--bundle',str(descriptor.relative_to(R)),'--out',str(consumerOut.relative_to(R)),'--timeout','180'];r=execute(launch,argv,200)
proof=load(consumerOut/'live-consumer.json') if (consumerOut/'live-consumer.json').exists() else {}
if r['exitCode']!=0 or r['status']!='PASS' or not proof.get('pass') or proof.get('exitCode')!=0:
 save(B/'STOP.json',{'kind':'outer-read-only-consumer-refused-or-incomplete','process':ref(launch/'receipt.json'),'noFurtherPaidAuthority':True});raise SystemExit(1)
assert source_snapshot(R)==source and proof['source']==source
assert files=={str(p.relative_to(native)):sha(p) for p in native.rglob('*') if p.is_file()}
save(B/'composition-result-v1.json',{'pass':True,'qualification':'FORMAL_APPLICABLE','coverage':14,'measurements':2,'preservedOriginalFailures':1,'descriptor':ref(descriptor/'minimax-thinking-composition.json'),'inspection':ref(inspection/'receipt.json'),'qualificationDetail':ref(inspection/'stdout.jsonl'),'consumer':ref(consumerOut/'live-consumer.json'),'consumerProcess':ref(launch/'receipt.json'),'source':source,'newPaidCalls':1,'newHTTPAttempts':1,'newGrantMaxUSD':'0.01','oldThirteenRecollected':False,'rawCreationCounterMissing':True,'usageReporting':'partial','independentCacheWriteFee':'NOT_APPLICABLE','noZeroFabricated':True,'oldTwoFailedSupplementsNotUsed':True,'noFurtherPaidAuthority':True})
print(json.dumps({'compositionPass':True,'qualification':'FORMAL_APPLICABLE','coverage':14,'consumerInputFiles':len(proof['inputs']),'furtherPaidCalls':0}))
