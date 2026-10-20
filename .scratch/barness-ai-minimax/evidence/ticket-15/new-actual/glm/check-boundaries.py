#!/usr/bin/env python3
"""Failure-first public process probes; no captured command or actual Key is read."""
import argparse,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent
REL='.evidence/minimax-retained/ticket-15-new-actual/glm-raw.tar.gz'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def seal(base):
 (base/'SHA256SUMS').write_text(''.join(sha(p)+'  '+p.name+'\n' for p in sorted(base.iterdir()) if p.is_file() and p.name!='SHA256SUMS'))

def probe(base,common):
 return subprocess.run([sys.executable,'-B',str(base/'verify.py'),'--common-root',str(common)],capture_output=True,text=True,timeout=120,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})

def run(common):
 baseline=probe(BASE,common)
 if baseline.returncode!=0:raise ValueError('correct actual FAIL history must verify: '+baseline.stderr[-1000:])
 value=json.loads(baseline.stdout)
 if value.get('collectionOutcome')!='FAIL' or (value.get('calls'),value.get('attempts'))!=(66,66) or value.get('threeRoutePublication') is not False:raise ValueError('correct frozen failed actual result required')
 cases={'metadata-pin':'local metadata','wrong-source':'fixed own','summary-pass':'independent summary','summary-usage':'independent summary','inventory-bound':'inventory decoded','archive-hash':'compressed archive pin','archive-size':'compressed archive size','archive-link':'symbolic links','archive-fifo':'regular artifact','traversal':'owned artifact path','duplicate':'archive path or duplicate','typed-link':'regular archive member','entry-bound':'archive entry limit','observation-bound':'observed metadata','canonical-capture-forgery':'failed Messages','canonical-root-forgery':'canonical root','messages-support-forgery':'failed Messages','secondary-path-forgery':'fixed secondary','refusal-forgery':'strict publication'}
 results=[];errors=[]
 for name,reason in cases.items():
  with tempfile.TemporaryDirectory(prefix='repair-glm-retention-negative-') as tmp:
   root=Path(tmp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/glm';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
   for relative in ('implementation/verify.py','final-gate/capsule.py'):
    source=BASE.parents[2]/'ticket-13'/relative;target=base.parents[2]/'ticket-13'/relative;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,target)
   archive=root/REL;archive.parent.mkdir(parents=True);shutil.copyfile(common/REL,archive);index=json.loads((base/'index.json').read_text());cap=index['capsule']
   if name=='metadata-pin':(base/'summary.json').write_text('{}')
   elif name=='wrong-source':index['actualSource']='f'*40
   elif name.startswith('summary-'):
    v=json.loads((base/'summary.json').read_text())
    if name=='summary-pass':v['collectionOutcome']='PASS';v['threeRoutePublication']=True
    else:v['unknownOrPartialUsageAttempts']=0;v['knownInputSide']=0
    (base/'summary.json').write_text(json.dumps(v))
   elif name=='inventory-bound':
    p=base/'inventory.json.gz';p.write_bytes(gzip.compress(b'x'*((8<<20)+1)));cap['inventory']['compressedSHA256']=sha(p)
   elif name=='archive-hash':cap['archive']['sha256']='0'*64
   elif name=='archive-size':
    with archive.open('wb') as f:f.truncate((32<<20)+1)
    cap['archive']['bytes']=archive.stat().st_size
   elif name in ('archive-link','archive-fifo'):
    archive.unlink()
    if name=='archive-link':archive.symlink_to(common/REL)
    else:os.mkfifo(archive,0o600)
   elif name=='observation-bound':cap['observe']*=3
   elif name=='secondary-path-forgery':index['preflight']['archive']['file']='.evidence/minimax-retained/ticket-15-new-actual/redirected.tar.gz'
   elif name in ('canonical-capture-forgery','canonical-root-forgery','messages-support-forgery','refusal-forgery'):
    files={}
    with tarfile.open(archive,'r:gz') as tar:
     for member in tar:
      if not member.isfile() or member.size>16<<20:raise ValueError('fixed regular mutation fixture')
      files[member.name]=tar.extractfile(member).read()
    member='repair-glm-actual/'+('canonical/messages.json' if name.startswith('canonical-') else index['routes']['messages']+'/live-report.json' if name=='messages-support-forgery' else 'frozen/strict-publication-refused.json');v={} if name=='messages-support-forgery' else json.loads(files[member])
    if name=='canonical-capture-forgery':v['capturePins']={'schema':1}
    elif name=='canonical-root-forgery':v['collectionRoot']='$REPO/.evidence/redirected'
    elif name=='messages-support-forgery':v={'syntheticUnauthorizedSupport':True}
    else:v['unchanged']=False
    files[member]=(json.dumps(v)+'\n').encode();pins=json.loads(gzip.decompress((base/'inventory.json.gz').read_bytes()));key=member.split('/',1)[1];pins[key]=dict(bytes=len(files[member]),sha256=hashlib.sha256(files[member]).hexdigest())
    with tarfile.open(archive,'w:gz') as tar:
     for leaf,raw in sorted(files.items()):
      info=tarfile.TarInfo(leaf);info.size=len(raw);info.mode=0o600;tar.addfile(info,io.BytesIO(raw))
    raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();p=base/'inventory.json.gz';p.write_bytes(gzip.compress(raw));cap['inventory'].update(compressedSHA256=sha(p),decodedSHA256=hashlib.sha256(raw).hexdigest(),decodedBytes=len(raw),files=len(pins));cap['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive))
   else:
    raw=gzip.decompress((base/'source-at-start.json.gz').read_bytes())
    with tarfile.open(archive,'w:gz') as tar:
     for n in range(32769 if name=='entry-bound' else 2 if name=='duplicate' else 1):
      leaf='../escape' if name=='traversal' else 'repair-glm-actual/empty-'+str(n) if name=='entry-bound' else 'repair-glm-actual/frozen/source-at-start.json';member=tarfile.TarInfo(leaf);data=raw
      if name=='entry-bound':member.type=tarfile.DIRTYPE;data=b''
      if name=='typed-link':member.type=tarfile.SYMTYPE;member.linkname='../../outside';data=b''
      member.size=len(data);tar.addfile(member,io.BytesIO(data))
    cap['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive))
   (base/'index.json').write_text(json.dumps(index))
   if name!='metadata-pin':seal(base)
   result=probe(base,root)
   if result.returncode==0 or reason not in result.stderr:errors.append('wrong rejection for '+name+': '+result.stderr[-500:])
   results.append(dict(case=name,exitCode=result.returncode,reason=result.stderr.strip()))
 if errors:raise ValueError(' | '.join(errors))
 after=probe(BASE,common)
 if after.returncode!=0 or json.loads(after.stdout)!=value:raise ValueError('original archives changed after private negative probes')
 return dict(verified=True,negativeGuards=results,originalArchivesMutated=False,retentionCredentialsRead=False,retentionPaidCalls=0)

if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=p.parse_args();print(json.dumps(run(a.common_root),indent=2))
