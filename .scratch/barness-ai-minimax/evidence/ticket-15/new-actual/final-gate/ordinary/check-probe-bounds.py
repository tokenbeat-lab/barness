#!/usr/bin/env python3
"""Exercise the public fixture creator on finite local untrusted archive inputs."""
import argparse,hashlib,json,os,subprocess,sys,time
from pathlib import Path
BASE=Path(__file__).resolve().parent
REL='.evidence/minimax-retained/ticket-15-new-actual/incomplete-ordinary-raw.tar.gz'
def main(output):
 output.mkdir(parents=True,exist_ok=False);results=[]
 source=BASE/'boolean-receipt-probe.py';raw=source.read_bytes();(output/'creator-before.py.txt').write_bytes(raw)
 for name in ('oversize','wrong-size','fifo','link'):
  root=output/name;archive=root/'common'/REL;archive.parent.mkdir(parents=True)
  if name in ('oversize','wrong-size'):
   with archive.open('xb') as f:f.truncate((256<<20)+1 if name=='oversize' else 1)
  elif name=='fifo':os.mkfifo(archive)
  else:archive.symlink_to('missing-target')
  argv=[sys.executable,'-B',str(source),'--common-root',str(root/'common'),'--work-output',str(root/'work'),'--mode','prepare','--kind','boolean'];start=time.monotonic()
  run=subprocess.run(argv,capture_output=True,timeout=30,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
  (root/'public.log').write_bytes(run.stdout+run.stderr)
  early=b'compressed archive size differs' in run.stderr
  good=run.returncode==1 and (early if name in ('oversize','wrong-size') else b'regular artifact required' in run.stderr or b'Too many levels' in run.stderr)
  results.append(dict(case=name,argv=argv,actualPublicExit=run.returncode,expectedEarlySizeRefusal=name in ('oversize','wrong-size'),actualEarlySizeRefusal=early,guardPassed=good,seconds=time.monotonic()-start,stdoutSHA256=hashlib.sha256(run.stdout).hexdigest(),stderrSHA256=hashlib.sha256(run.stderr).hexdigest()))
 result=dict(creatorSHA256=hashlib.sha256(raw).hexdigest(),cases=results,credentialsRead=False,paidCalls=0,originalNativeTouched=False)
 (output/'results.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
 if not all(v['guardPassed'] for v in results):raise ValueError('public creator lacks bounded early refusal')
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--output',type=Path,required=True);main(p.parse_args().output)
