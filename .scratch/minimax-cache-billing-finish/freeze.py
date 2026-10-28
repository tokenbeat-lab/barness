"""Build normal live-tag producers and read-only tools; never execute a producer."""
from pathlib import Path
import hashlib,json,subprocess,sys,time
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
from run_process import environment,isolate_go_home
out=ROOT/'.evidence/minimax-cache-billing-finish/build-v1';out.mkdir()
env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
source=source_snapshot(ROOT);steps=[]
def execute(name,argv):
 d=out/name;d.mkdir();started=time.time();r=subprocess.run(argv,cwd=ROOT,env=env,capture_output=True)
 for n,b in [('stdout',r.stdout),('stderr',r.stderr),('exit-code',(str(r.returncode)+'\n').encode())]:(d/n).write_bytes(b)
 v={'argv':argv,'exitCode':r.returncode,'elapsedSeconds':time.time()-started};(d/'receipt.json').write_text(json.dumps(v,indent=2)+'\n');steps.append(v)
 if r.returncode:raise RuntimeError(name+': '+r.stderr.decode()[:1000])
 return r.stdout
cli=out/'processgrant';execute('01-processgrant',['go','build','-o',str(cli),'./ai/live/cmd/processgrant'])
mm=json.loads(execute('02-minimax-source',[str(cli),'-source','minimax-messages','-stage','candidate']))
regional=json.loads(execute('03-kimi-source',[str(cli),'-source','regional']))
mmBinary=out/'minimax-live.test';kimiBinary=out/'kimi-live.test'
execute('04-minimax-build',['go','test','-c','-tags','live','-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource='+mm['sourceSha256'],'-o',str(mmBinary),'./ai/live'])
execute('05-kimi-build',['go','test','-c','-tags','live','-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource='+regional['sourceSha256'],'-o',str(kimiBinary),'./ai/live'])
release=out/'releasegate';execute('06-releasegate',['go','build','-o',str(release),'./ai/release/cmd/releasegate'])
promote=out/'regionalpromote';execute('11-regionalpromote',['go','build','-o',str(promote),'./ai/live/cmd/regionalpromote'])
execute('12-promote-metadata',['go','version','-m',str(promote)])
for n,p in [('07-minimax-metadata',mmBinary),('08-kimi-metadata',kimiBinary),('09-release-metadata',release)]:execute(n,['go','version','-m',str(p)])
baseline='53cfa61b39f8246d6ac526ed0c3e5b91a32a82f36a49e3f37034108a9b2224fa'
template=json.loads(execute('10-minimax-scope',[str(cli),'-template','minimax-messages','-mode','live','-stage','candidate','-profile','minimax-messages-thinking-on/v1','-baseline-manifest-sha256',baseline,'-binary',str(mmBinary)]))
assert template['id']=='' and template['expiresAt']=='0001-01-01T00:00:00Z'
assert source==source_snapshot(ROOT),'source changed while freezing'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
ref=lambda p:{'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
value={'kind':'normal-live-tag-builds-only','source':source,'minimaxSource':mm,'regionalSource':regional,'minimax':ref(mmBinary),'kimi':ref(kimiBinary),'releasegate':ref(release),'regionalpromote':ref(promote),'processgrant':ref(cli),'minimaxMachineScope':template['scope'],'steps':steps,'vendorCalls':0,'grantsIssued':0,'claims':0,'producerExecutions':0}
(out/'freeze.json').write_text(json.dumps(value,indent=2)+'\n');print(json.dumps({k:value[k] for k in ['minimax','kimi','releasegate','producerExecutions']}))
