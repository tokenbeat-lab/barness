#!/usr/bin/env python3
"""One finite private full-archive fixture for JSON true versus numeric exit1."""
import argparse,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;REL='.evidence/minimax-retained/ticket-15-new-actual/incomplete-ordinary-raw.tar.gz';EXPECTED='9c8c1bcb79b3ea07c93a276f25f2d97768eb937c1b8008c012ae6fdfe934cacd';sha=lambda raw:hashlib.sha256(raw).hexdigest()
def seal(base):(base/'SHA256SUMS').write_text(''.join(sha(p.read_bytes())+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def prepare(common,root,kind):
 if root.exists():raise ValueError('fresh private fixture required')
 root.mkdir();base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
 for rel in ('implementation/verify.py','final-gate/capsule.py'):
  p=base.parents[3]/'ticket-13'/rel;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(BASE.parents[3]/'ticket-13'/rel,p)
 b=readers.small();index=b.doc((BASE/'index.json').read_bytes());full=readers.full();pins=b.doc(full.decoded(b,BASE,index['inventory'],full.LIMITS['inventoryDecodedBytes']))
 if index['archive']['sha256']!=EXPECTED or index['limits']!=full.LIMITS:raise ValueError('fixed original archive required')
 archive=root/REL;archive.parent.mkdir(parents=True);entries=0;total=0;seen=set();changed=False
 with b.opened(common,REL) as source:
  expected=index['archive']['bytes'];size=os.fstat(source.fileno()).st_size
  if type(expected) is not int or size!=expected or size>full.LIMITS['compressedBytes']:raise ValueError('compressed archive size differs')
  digest=hashlib.sha256();read=0
  for chunk in iter(lambda:source.read(1<<20),b''):
   read+=len(chunk)
   if read>expected or read>full.LIMITS['compressedBytes']:raise ValueError('compressed archive read bound')
   digest.update(chunk)
  if read!=expected:raise ValueError('compressed archive size differs')
  if digest.hexdigest()!=EXPECTED:raise ValueError('original archive pin differs')
  source.seek(0)
  with gzip.GzipFile(fileobj=source) as zipped:
   limited=b.Limited(zipped,full.LIMITS['decodedTarBytes'])
   with tarfile.open(fileobj=limited,mode='r|') as original,archive.open('xb') as output:
    with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as compressed:
     with tarfile.open(fileobj=compressed,mode='w|',format=tarfile.PAX_FORMAT) as replacement:
      for member in original:
       entries+=1;parts=b.relative(member.name);name='/'.join(parts[1:]);pin=pins.get(name)
       if entries>full.LIMITS['entries'] or parts[0]!=index['archive']['prefix'] or not member.isfile() or name in seen or pin is None or member.size!=pin['bytes'] or member.size>full.LIMITS['memberBytes']:raise ValueError('fixed regular mutation member required')
       data=original.extractfile(member).read(full.LIMITS['memberBytes']+1)
       if len(data)!=pin['bytes'] or sha(data)!=pin['sha256']:raise ValueError('original member changed')
       total+=len(data)
       if total>full.LIMITS['rawMemberBytes']:raise ValueError('fixed raw bound')
       if name=='receipts/ordinary.json':
        receipt=b.doc(data)
        if type(receipt['exitCode']) is not int or receipt['exitCode']!=1:raise ValueError('actual original numeric1 required')
        receipt['exitCode']=True if kind=='boolean' else 1.0;data=(json.dumps(receipt,separators=(',',':'))+'\n').encode();pins[name]=dict(bytes=len(data),sha256=sha(data));changed=True
       item=tarfile.TarInfo(member.name);item.mode=0o600;item.size=len(data);replacement.addfile(item,io.BytesIO(data));seen.add(name)
   while limited.read(65536):pass
 if seen!=set(pins) or not changed or archive.stat().st_size>full.LIMITS['compressedBytes']:raise ValueError('complete finite mutation required')
 raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(raw,mtime=0)
 if len(raw)>full.LIMITS['inventoryDecodedBytes'] or len(encoded)>full.LIMITS['inventoryCompressedBytes']:raise ValueError('fixed inventory bound')
 (base/'inventory.json.gz').write_bytes(encoded);index['inventory'].update(compressedSHA256=sha(encoded),decodedSHA256=sha(raw),decodedBytes=len(raw),files=len(pins));index['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()));(base/'index.json').write_text(json.dumps(index)+'\n');_,m,actual=full.verify(b,base,root,index)
 actual_code=b.doc(m['receipts/ordinary.json'])['exitCode']
 if type(actual_code) is not (bool if kind=='boolean' else float) or actual_code!=1:raise ValueError('intentional exact numeric-type fixture differs')
 summary=b.doc((base/'summary.json').read_bytes());summary['archive']=actual;(base/'summary.json').write_text(json.dumps(summary)+'\n');archive.chmod(0o444);seal(base);proof=dict(originalArchiveSHA256=EXPECTED,fixtureArchive=actual,intentionalReceiptType=kind,intentionalReceiptValue=actual_code,originalFilesChanged=False,onlyNativeMutation='receipts/ordinary.json exitCode numeric1→noninteger JSON type',sourceBeforeSHA256=sha((base/'verify.py').read_bytes()),credentialsRead=False,paidCalls=0);(root/'fixture-proof.json').write_text(json.dumps(proof)+'\n');return base

def run(common,root,mode,kind):
 base=prepare(common,root,kind) if mode=='prepare' else root/'.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/final-gate/ordinary'
 if mode=='check':
  proof=json.loads((root/'fixture-proof.json').read_text())
  if proof['originalArchiveSHA256']!=EXPECTED:raise ValueError('fixed original fixture provenance differs')
  shutil.copyfile(BASE/'verify.py',base/'verify.py');seal(base)
 argv=[sys.executable,'-B',str(base/'verify.py'),'--common-root',str(root)];result=subprocess.run(argv,capture_output=True,text=True,timeout=180,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ});(root/(mode+'-public.log')).write_text(result.stdout+result.stderr);receipt=dict(argv=argv,publicExitCode=result.returncode,invalidNumericReceiptRefused=result.returncode==1 and 'numeric exit' in result.stderr,stdoutSHA256=sha(result.stdout.encode()),stderrSHA256=sha(result.stderr.encode()),publicVerifierSHA256=sha((base/'verify.py').read_bytes()),credentialsRead=False,paidCalls=0);(root/(mode+'-public-command.json')).write_text(json.dumps(receipt)+'\n');print(json.dumps(receipt,indent=2),flush=True)
 if not receipt['invalidNumericReceiptRefused']:raise ValueError('public verifier accepted noninteger JSON type as numeric exit1')
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);p.add_argument('--work-output',type=Path,required=True);p.add_argument('--mode',choices=('prepare','check'),required=True);p.add_argument('--kind',choices=('boolean','float'),required=True);a=p.parse_args();run(a.common_root,a.work_output,a.mode,a.kind)
