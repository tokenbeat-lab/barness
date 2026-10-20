#!/usr/bin/env python3
"""Small private-index negatives over the unchanged full regular archive."""
import argparse,hashlib,json,os,shutil,subprocess,sys,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent;REL='.evidence/minimax-retained/ticket-15-new-actual/incomplete-ordinary-raw.tar.gz';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def seal(base):(base/'SHA256SUMS').write_text(''.join(sha(p)+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def probe(base,root):return subprocess.run([sys.executable,'-B',str(base/'verify.py'),'--common-root',str(root)],capture_output=True,text=True,timeout=120,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
def run(common):
 baseline=probe(BASE,common)
 if baseline.returncode:raise ValueError('original failed ordinary must verify: '+baseline.stderr[-500:])
 original=json.loads(baseline.stdout);results=[]
 for name,reason in {'metadata':'local metadata','wrong-source':'fixed original ordinary','wrong-path':'fixed original ordinary','widen-limits':'fixed streaming limits','archive-pin':'compressed archive pin','summary-pass':'independent ordinary','summary-failures':'independent ordinary','observations':'observed metadata'}.items():
  with tempfile.TemporaryDirectory(prefix='repair-ordinary-retention-negative-') as tmp:
   root=Path(tmp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
   for rel in ('implementation/verify.py','final-gate/capsule.py'):
    p=base.parents[3]/'ticket-13'/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(BASE.parents[3]/'ticket-13'/rel,p)
   archive=root/REL;archive.parent.mkdir(parents=True);os.link(common/REL,archive);index=json.loads((base/'index.json').read_text())
   if name=='metadata':(base/'summary.json').write_text('{}')
   elif name=='wrong-source':index['sourceCommit']='f'*40
   elif name=='wrong-path':index['archive']['file']='.evidence/other.tar.gz'
   elif name=='widen-limits':index['limits']['files']+=1
   elif name=='archive-pin':index['archive']['sha256']='0'*64
   elif name=='observations':index['observe']*=30
   else:
    s=json.loads((base/'summary.json').read_text())
    if name=='summary-pass':s['statuses']={'PASS':12451};s['exitCode']=0
    else:s['failures']=[]
    (base/'summary.json').write_text(json.dumps(s))
   (base/'index.json').write_text(json.dumps(index))
   if name!='metadata':seal(base)
   r=probe(base,root)
   if r.returncode!=1 or reason not in r.stderr:raise ValueError('wrong rejection '+name+': '+r.stderr[-500:])
   results.append(dict(case=name,exitCode=r.returncode,reason=r.stderr.strip()))
 for kind in ('boolean','float'):
  with tempfile.TemporaryDirectory(prefix='repair-ordinary-numeric-negative-') as tmp:
   root=Path(tmp)/'private'
   result=subprocess.run([sys.executable,'-B',str(BASE/'boolean-receipt-probe.py'),'--common-root',str(common),'--work-output',str(root),'--mode','prepare','--kind',kind],capture_output=True,text=True,timeout=300,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
   if result.returncode!=0:raise ValueError('typed numeric public guard failed: '+result.stderr[-500:])
   receipt=json.loads(result.stdout);reason=(root/'prepare-public.log').read_text()
   if receipt['publicExitCode']!=1 or receipt['invalidNumericReceiptRefused'] is not True or not reason.startswith('REFUSED:'):raise ValueError('exact numeric type public refusal required')
   results.append(dict(case='exit-code-'+kind,exitCode=1,reason=reason.strip(),fixtureArchiveSHA256=json.loads((root/'fixture-proof.json').read_text())['fixtureArchive']['archiveSHA256']))
 after=probe(BASE,common)
 if after.returncode or json.loads(after.stdout)!=original:raise ValueError('original archive changed')
 return dict(verified=True,negativeGuards=results,originalArchivesMutated=False,archiveHardlinksUsedReadOnly=True,credentialsRead=False,paidCalls=0)
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,default=BASE.parents[6]);a=p.parse_args();print(json.dumps(run(a.common_root),indent=2))
