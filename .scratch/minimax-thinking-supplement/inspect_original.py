"""Compile a keyless inspection caller in a temporary internal module, then run
against immutable real history. No overlays, grants, claims or HTTP clients."""
from pathlib import Path
import json,os,sys,subprocess,tempfile,time,hashlib
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'tools/verification'))
from run_process import environment
from receipt_files import source_snapshot
out=ROOT/'.evidence/minimax-thinking-supplement/original-baseline-v5';out.mkdir()
env,_=environment('e2e');env.update({'GOPROXY':'off','GOSUMDB':'off'})
with tempfile.TemporaryDirectory() as td:
 td=Path(td);(td/'go.mod').write_text('module github.com/tokenbeat-lab/barness/ai/minimaxthinkinginspection\n\ngo 1.26\n\nrequire github.com/tokenbeat-lab/barness v0.0.0\nreplace github.com/tokenbeat-lab/barness => '+str(ROOT)+'\n')
 (td/'main.go').write_bytes((ROOT/'.scratch/minimax-thinking-supplement/inspect_original.go').read_bytes())
 binary=out/'inspect-original'
 c=['go','build','-mod=mod','-o',str(binary),'.']
 t=time.time();b=subprocess.run(c,cwd=td,env=env,capture_output=True)
 (out/'build.stdout').write_bytes(b.stdout);(out/'build.stderr').write_bytes(b.stderr)
 if b.returncode:raise RuntimeError(b.stderr.decode())
 base='.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id'
 bundle=base+'/bundles/20261027T000249.211139000Z';cmd=[str(binary),bundle,base+'/receipt.json']
 r=subprocess.run(cmd,cwd=ROOT,env=env,capture_output=True)
 (out/'stdout.json').write_bytes(r.stdout);(out/'stderr.log').write_bytes(r.stderr);(out/'exit-code').write_text(str(r.returncode)+'\n')
 receipt={'argv':cmd,'exitCode':r.returncode,'source':source_snapshot(ROOT),'elapsedSeconds':time.time()-t,'vendorCalls':0,'grantsIssued':0,'claims':0}
 (out/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps({'exitCode':r.returncode,'stderr':r.stderr.decode()[:1500]}));sys.exit(r.returncode)
