#!/usr/bin/env python3
"""Public process custody negatives specified before the effects consumer."""
import argparse,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent;REL='.evidence/minimax-retained/ticket-15-new-actual/effects-actual-raw.tar.gz';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def seal(base):
 (base/'SHA256SUMS').write_text(''.join(sha(p)+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))

def probe(base,root):return subprocess.run([sys.executable,'-B',str(base/'verify.py'),'--common-root',str(root)],capture_output=True,text=True,timeout=180,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})

def run(common):
 first=probe(BASE,common)
 if first.returncode:raise ValueError('original verification failed: '+first.stderr[-500:])
 original=json.loads(first.stdout);results=[]
 cases={'local-pin':'local metadata','wrong-source':'fixed effects','summary-usage':'independent summary','summary-certified':'independent summary','archive-pin':'compressed archive pin','archive-link':'symbolic links','archive-fifo':'regular artifact','observation-limit':'observed metadata','inventory-limit':'inventory decoded','secondary-path':'fixed secondary','shared-binary':'shared original binary','canonical-root':'canonical root','canonical-phase':'canonical root','visual-quality':'independent visual quality','visual-proxy':'explicit visual proxy'}
 for name,reason in cases.items():
  with tempfile.TemporaryDirectory(prefix='repair-effects-retention-negative-') as temp:
   root=Path(temp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects';base.parent.mkdir(parents=True);shutil.copytree(BASE,base);shutil.copytree(BASE.parent/'glm',base.parent/'glm')
   for rel in ('implementation/verify.py','final-gate/capsule.py'):
    p=base.parents[2]/'ticket-13'/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(BASE.parents[2]/'ticket-13'/rel,p)
   index=json.loads((base/'index.json').read_text())
   for cap in [index[k] for k in ('capsule','preflight','binary')]+[index['sharedLiveBinary']['capsule']]:
    p=root/cap['archive']['file'];p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(common/cap['archive']['file'],p)
   cap=index['capsule'];archive=root/REL
   if name=='local-pin':(base/'summary.json').write_text('{}')
   elif name=='wrong-source':index['actualSource']='f'*40
   elif name.startswith('summary-'):
    v=json.loads((base/'summary.json').read_text());v['knownInputSide']=0 if name=='summary-usage' else v['knownInputSide'];v['releaseCertified']=name=='summary-certified';(base/'summary.json').write_text(json.dumps(v))
   elif name=='archive-pin':cap['archive']['sha256']='0'*64
   elif name in ('archive-link','archive-fifo'):
    archive.unlink()
    if name=='archive-link':archive.symlink_to(common/REL)
    else:os.mkfifo(archive,0o600)
   elif name=='observation-limit':cap['observe']*=4
   elif name=='inventory-limit':
    p=base/'inventory.json.gz';p.write_bytes(gzip.compress(b'x'*((8<<20)+1)));cap['inventory']['compressedSHA256']=sha(p)
   elif name=='secondary-path':index['preflight']['archive']['file']='.evidence/redirected.tar.gz'
   elif name=='shared-binary':index['sharedLiveBinary']['pin']['sha256']='0'*64
   else:
    files={}
    with tarfile.open(archive,'r:gz') as tar:
     for member in tar:
      if not member.isfile() or member.size>16<<20:raise ValueError('fixed regular mutation fixture')
      files[member.name]=tar.extractfile(member).read()
    leaf='canonical/visual-jev.json' if name.startswith('canonical-') else index['roots']['visual-jev']+('/report.json' if name=='visual-quality' else '/run-config.json');member=cap['archive']['prefix']+'/'+leaf;v=json.loads(files[member])
    if name=='canonical-root':v['collectionRoot']='$REPO/.evidence/other'
    elif name=='canonical-phase':v['phase']='legacy'
    elif name=='visual-quality':v['metrics']['box']['meanIoU']=1
    else:v['proxy']='direct'
    files[member]=(json.dumps(v)+'\n').encode();pins=json.loads(gzip.decompress((base/'inventory.json.gz').read_bytes()));pins[leaf]=dict(bytes=len(files[member]),sha256=hashlib.sha256(files[member]).hexdigest())
    with tarfile.open(archive,'w:gz') as tar:
     for key,data in sorted(files.items()):
      info=tarfile.TarInfo(key);info.size=len(data);tar.addfile(info,io.BytesIO(data))
    raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();p=base/'inventory.json.gz';p.write_bytes(gzip.compress(raw));cap['inventory'].update(compressedSHA256=sha(p),decodedSHA256=hashlib.sha256(raw).hexdigest(),decodedBytes=len(raw),files=len(pins));cap['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive))
   (base/'index.json').write_text(json.dumps(index))
   if name!='local-pin':seal(base)
   result=probe(base,root)
   if result.returncode!=1 or reason not in result.stderr:raise ValueError('wrong rejection '+name+': '+result.stderr[-1000:])
   results.append(dict(case=name,exitCode=result.returncode,reason=result.stderr.strip()))
 after=probe(BASE,common)
 if after.returncode or json.loads(after.stdout)!=original:raise ValueError('original changed after private probes')
 return dict(verified=True,negativeGuards=results,originalArchivesMutated=False,credentialsRead=False,paidCalls=0)
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=p.parse_args();print(json.dumps(run(a.common_root),indent=2))
