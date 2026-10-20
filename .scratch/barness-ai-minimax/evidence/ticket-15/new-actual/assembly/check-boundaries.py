#!/usr/bin/env python3
"""Public process failures specified before the failed-assembly consumer."""
import argparse,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent;REL='.evidence/minimax-retained/ticket-15-new-actual/assembly-failed-raw.tar.gz';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def seal(base):(base/'SHA256SUMS').write_text(''.join(sha(p)+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def probe(base,root):return subprocess.run([sys.executable,'-B',str(base/'verify.py'),'--common-root',str(root)],capture_output=True,text=True,timeout=120,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
def run(common):
 first=probe(BASE,common)
 if first.returncode:raise ValueError('original failed assembly must verify: '+first.stderr[-500:])
 original=json.loads(first.stdout);results=[];cases={'metadata':'local metadata','summary-pass':'independent failed assembly','wrong-source':'fixed assembly','archive-pin':'compressed archive pin','archive-link':'symbolic links','archive-fifo':'regular artifact','observations':'observed metadata','reference-root':'fixed original native','forge-exit':'original keyless numeric','forge-matrix':'original default/candidate matrix','forge-completed':'closed failed assembly','forge-helper':'original operational helper','forge-report-path':'original six native'}
 for name,reason in cases.items():
  with tempfile.TemporaryDirectory(prefix='repair-assembly-retention-negative-') as temp:
   root=Path(temp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/assembly';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
   for label in ('glm','effects'):shutil.copytree(BASE.parent/label,base.parent/label)
   target=base.parents[1]/'finalizer-design';shutil.copytree(BASE.parents[1]/'finalizer-design',target)
   for rel in ('implementation/verify.py','final-gate/capsule.py'):
    p=base.parents[2]/'ticket-13'/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(BASE.parents[2]/'ticket-13'/rel,p)
   index=json.loads((base/'index.json').read_text());cap=index['capsule']
   for c in [cap]+[json.loads((base.parent/label/'index.json').read_text())['capsule'] for label in ('glm','effects')]:
    p=root/c['archive']['file'];p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(common/c['archive']['file'],p)
   archive=root/REL
   if name=='metadata':(base/'summary.json').write_text('{}')
   elif name=='summary-pass':
    v=json.loads((base/'summary.json').read_text());v['releaseCertified']=True;v['assemblyOutcome']='PASS';(base/'summary.json').write_text(json.dumps(v))
   elif name=='wrong-source':index['sourceCommit']='f'*40
   elif name=='archive-pin':cap['archive']['sha256']='0'*64
   elif name in ('archive-link','archive-fifo'):
    archive.unlink()
    if name=='archive-link':archive.symlink_to(common/REL)
    else:os.mkfifo(archive,0o600)
   elif name=='observations':cap['observe']*=32
   elif name=='reference-root':index['nativeReferences']['effects']['frontdoor']='../other'
   else:
    files={}
    with tarfile.open(archive,'r:gz') as tar:
     for member in tar:files[member.name]=tar.extractfile(member).read()
    prefix=cap['archive']['prefix']+'/';leaf='original/'+('validate-four-candidate-command.json' if name=='forge-exit' else 'matrix-candidate.json' if name=='forge-matrix' else 'assemble-completed.json' if name=='forge-completed' else 'assemble-inputs.json');member=prefix+leaf;v={} if name=='forge-completed' else json.loads(files[member])
    if name=='forge-exit':v['exitCode']=0
    elif name=='forge-matrix':v['rows']=[]
    elif name=='forge-completed':v['status']='COMPLETE'
    elif name=='forge-report-path':v['actuals']['release-repair-validation-1-glm-chat']['report']='.evidence/redirected/formal-glm-report.json'
    else:v['operations']['helperSHA256']='0'*64
    files[member]=(json.dumps(v)+'\n').encode();pins=json.loads(gzip.decompress((base/'inventory.json.gz').read_bytes()));pins[leaf]=dict(bytes=len(files[member]),sha256=hashlib.sha256(files[member]).hexdigest())
    if leaf not in cap['observe']:cap['observe'].append(leaf)
    with tarfile.open(archive,'w:gz') as tar:
     for key,data in sorted(files.items()):
      item=tarfile.TarInfo(key);item.size=len(data);tar.addfile(item,io.BytesIO(data))
    raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();p=base/'inventory.json.gz';p.write_bytes(gzip.compress(raw));cap['inventory'].update(compressedSHA256=sha(p),decodedSHA256=hashlib.sha256(raw).hexdigest(),decodedBytes=len(raw),files=len(pins));cap['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive))
   (base/'index.json').write_text(json.dumps(index))
   if name!='metadata':seal(base)
   r=probe(base,root)
   if r.returncode!=1 or reason not in r.stderr:raise ValueError('wrong rejection '+name+': '+r.stderr[-500:])
   results.append(dict(case=name,exitCode=r.returncode,reason=r.stderr.strip()))
 after=probe(BASE,common)
 if after.returncode or json.loads(after.stdout)!=original:raise ValueError('original changed after probes')
 return dict(verified=True,negativeGuards=results,originalArchivesMutated=False,credentialsRead=False,paidCalls=0)
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=p.parse_args();print(json.dumps(run(a.common_root),indent=2))
