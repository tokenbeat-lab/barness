#!/usr/bin/env python3
"""Public verifier failures on private metadata copies; never edit the raw archive."""
import argparse,hashlib,json,os,shutil,subprocess,sys,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent
def seal(base):
 (base/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def main(common):
 reasons=dict(source='fixed offline retention identity',limits='fixed streaming limits differ',root='invalid owned artifact path',summary='independent complete history summary differs',qualification='independent complete history summary differs',link='Too many levels',fifo='regular artifact required')
 reasons.update({'archive-pin':'compressed archive pin differs','hidden-failure':'fixed original measured source or stage roots differ','key-claim':'fixed offline retention identity','local-pin':'local metadata pin differs'})
 def original():
  run=subprocess.run([sys.executable,'-B',str(BASE/'verify.py'),'--common-root',str(common)],capture_output=True,timeout=180,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
  if run.returncode:raise ValueError('original public verifier must pass: '+run.stderr.decode())
  return run.stdout
 baseline=original();results=[]
 for name in ('source','limits','archive-pin','root','hidden-failure','summary','qualification','key-claim','local-pin','link','fifo'):
  with tempfile.TemporaryDirectory(prefix='barness-ticket17-negative-') as temp:
   root=Path(temp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-17';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
   for rel in ('implementation/verify.py','final-gate/capsule.py'):
    target=base.parent/'ticket-13'/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(BASE.parent/'ticket-13'/rel,target)
   index=json.loads((base/'index.json').read_bytes());selected_common=common
   if name=='source':index['sourceCommit']='0'*40
   elif name=='limits':index['limits']['memberBytes']+=1
   elif name=='archive-pin':index['archive']['sha256']='0'*64
   elif name=='root':index['roots']['green-final-native-shapes']+='/../foreign'
   elif name=='hidden-failure':del index['roots']['green-final-shapes']
   elif name in ('summary','qualification'):
    summary=json.loads((base/'summary.json').read_bytes());summary['trueMalformedSendREDs']=0 if name=='summary' else summary['trueMalformedSendREDs'];summary['releaseCertified']=name=='qualification';(base/'summary.json').write_text(json.dumps(summary)+'\n')
   elif name=='key-claim':index['credentialsRead']=True
   elif name in ('link','fifo'):
    selected_common=root/'common';path=selected_common/index['archive']['file'];path.parent.mkdir(parents=True)
    if name=='link':path.symlink_to(common/index['archive']['file'])
    else:os.mkfifo(path)
   (base/'index.json').write_text(json.dumps(index)+'\n');seal(base)
   if name=='local-pin':(base/'summary.json').write_text('{}\n')
   argv=[sys.executable,'-B',str(base/'verify.py'),'--common-root',str(selected_common)];run=subprocess.run(argv,capture_output=True,timeout=180,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
   row=dict(case=name,publicExit=run.returncode,refused=run.returncode==1 and reasons[name] in run.stderr.decode(),diagnostic=run.stderr.decode(),stdoutSHA256=hashlib.sha256(run.stdout).hexdigest(),stderrSHA256=hashlib.sha256(run.stderr).hexdigest());results.append(row)
   if not row['refused']:raise ValueError('public negative failed: '+name+' '+str(row))
 if original()!=baseline:raise ValueError('original read-only corpus changed after negatives')
 print(json.dumps(dict(publicNegatives=results,originalBeforeAfterIdentical=True,credentialsRead=False,paidCalls=0),indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);main(p.parse_args().common_root)
