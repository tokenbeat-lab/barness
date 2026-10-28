"""This fixed attempt's keyless preflight. No grants, claims or vendor requests."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,subprocess,sys,tempfile,time
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-thinking-resample';S=R/'.scratch/minimax-thinking-resample'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
from run_process import environment,capture,isolate_go_home
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False);f.write('\n')
a=load(B/'approval.json');assert datetime.now(timezone.utc)<datetime.fromisoformat(a['expiresAt'].replace('Z','+00:00'))
assert sha(R/a['freeze']['path'])==a['freeze']['sha256'];freeze=load(R/a['freeze']['path']);source=source_snapshot(R);assert source==freeze['source']
v=load(R/'.evidence/minimax-cache-billing-finish/verification-v1.json');ordinaryPath=R/v['ordinary']['path'];assert sha(ordinaryPath)==v['ordinary']['sha256'];ordinary=load(ordinaryPath)
assert ordinary['status']=='PASS' and ordinary['exitCode']==0 and ordinary['source']==source;assert all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
for key in ['minimax','processgrant','releasegate']:assert sha(R/freeze[key]['path'])==freeze[key]['sha256']
assert freeze['minimax']==a['producer'] and freeze['minimaxMachineScope']['sourceSha256']==a['nativeSourceSHA256']
old=load(R/'.scratch/kimi-composition-promotion/pending-scopes-v5-confirmed-v1/minimax-messages.json')
assert old['confirmations']['accountAndRegion']==old['confirmations']['ratesAndServiceTier']=='USER_CONFIRMED'
assert old['identityToConfirm']['accountAlias']=='MINIMAX' and old['identityToConfirm']['regionAlias']=='global'
for name in ['manifest','report','seal','producer']:
 d=old['baseline'][name];assert sha(R/d['path'])==d['sha256']
assert old['baseline']['manifest']['sha256']==a['baselineManifestSHA256']
protected={n:sha(R/n) for n in [d['path'] for d in [old['baseline'][k] for k in ['manifest','report','seal','producer']]]}
for n,h in load(R/old['baseline']['seal']['path'])['files'].items():
 p=R/old['baseline']['bundle']/n;assert sha(p)==h;protected[str(p.relative_to(R))]=h
for root in ['.evidence/provider-minimal-paid-supplement','.evidence/minimax-cache-billing-finish']:
 for p in (R/root).rglob('*'):
  if p.is_file():protected[str(p.relative_to(R))]=sha(p)
save(B/'historical-files-before.json',protected)
roots=load(R/'.evidence/minimax-thinking-supplement-plan/store-roots.json');public={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()};assert public==load(R/'.evidence/minimax-cache-billing-finish/initial/public-authority.json');save(B/'authority-before.json',public)
out=B/'preflight';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
common=subprocess.check_output(['git','rev-parse','--path-format=absolute','--git-common-dir'],cwd=R,env=env,text=True).strip();assert Path(common).resolve()==(R/'.git').resolve()
(R0:=B/'git-status-before.txt').write_bytes(subprocess.check_output(['git','status','--porcelain=v1','--untracked-files=all'],cwd=R,env=env))
cli=R/freeze['processgrant']['path'];binary=R/freeze['minimax']['path']
def check(name,argv,cwd=R,timeout=120):
 d=out/name;d.mkdir();started=datetime.now(timezone.utc);tic=time.monotonic();r=capture(argv,cwd,env,d,timeout,[]);files={p.name:sha(p) for p in d.iterdir() if p.is_file()};save(d/'receipt.json',{'argv':argv,'startedAt':started.isoformat(),'finishedAt':datetime.now(timezone.utc).isoformat(),'elapsedSeconds':time.monotonic()-tic,'files':files,**r});assert r['exitCode']==0 and r['status']=='PASS',name;return d
scopeDir=check('scope',[str(cli),'-template','minimax-messages','-mode','live','-stage','candidate','-profile',a['profile'],'-baseline-manifest-sha256',a['baselineManifestSHA256'],'-binary',str(binary)])
template=load(scopeDir/'stdout.jsonl');assert template['id']=='' and template['expiresAt']=='0001-01-01T00:00:00Z' and template['scope']==freeze['minimaxMachineScope'];save(B/'machine-template.json',template)
# Tiny retained caller uses existing public native checks against current source.
main=(R/'.scratch/minimax-thinking-supplement/inspect_original.go').read_text();main=main.replace('pass:=[]string{};', 'current,e:=minimaxlive.ExecutionSource();if e!=nil {panic(e)};pass:=[]string{};').replace('"pass":len(pass)', '"currentExecutionSource":current,"pass":len(pass)')
(S/'inspect_baseline.go').write_text(main)
with tempfile.TemporaryDirectory() as td:
 td=Path(td);(td/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/minimaxresampleinspection\n\ngo 1.26\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(R)+'\n');(td/'main.go').write_text(main)
 probe=out/'inspect-baseline';check('build-baseline-probe',['go','build','-mod=mod','-o',str(probe),'.'],td)
proofDir=check('baseline',[str(probe),old['baseline']['bundle'],old['baseline']['producer']['path']]);proof=load(proofDir/'stdout.jsonl');assert proof['pass'] and len(proof['applicable'])==13
native=proof['currentExecutionSource'];assert all(sha(R/n)==h for n,h in native.items());assert hashlib.sha256(json.dumps(native,sort_keys=True,separators=(',',':')).encode()).hexdigest()==a['nativeSourceSHA256'];save(B/'native-source.json',native)
assert source_snapshot(R)==source
save(B/'preflight.json',{'pass':True,'source':source,'ordinary':ref(ordinaryPath),'freeze':a['freeze'],'baseline':old['baseline'],'baselineProof':ref(proofDir/'receipt.json'),'nativeSource':ref(B/'native-source.json'),'machineTemplate':ref(B/'machine-template.json'),'authorityFiles':len(public),'protectedHistoryFiles':len(protected),'catalogSHA256':sha(R/'ai/catalog_kimi.json'),'trustedGitCommonDir':common,'vendorCalls':0,'grantsIssued':0,'claims':0})
print(json.dumps({'preflight':'PASS','sourceFiles':len(source['files']),'ordinaryInputs':len(ordinary['selectedInputs']),'baselineApplicableItems':13,'protectedHistoryFiles':len(protected),'publicAuthorityFiles':len(public),'grantsIssued':0,'vendorCalls':0}))
