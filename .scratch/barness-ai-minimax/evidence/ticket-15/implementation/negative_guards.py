#!/usr/bin/env python3
"""Public verifier guard checks in private copies; immutable originals stay read-only."""
import argparse,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent
REL='.evidence/minimax-retained/ticket-15-implementation/release-repair-raw.tar.gz'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def seal(base):
 names=[p for p in base.iterdir() if p.is_file() and p.name!='SHA256SUMS']
 (base/'SHA256SUMS').write_text(''.join(sha(p)+'  '+p.name+'\n' for p in sorted(names)))

def probe(base,common):
 return subprocess.run([sys.executable,'-B',str(base/'verify.py'),'--common-root',str(common)],capture_output=True,text=True,timeout=30,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})

def run(common):
 results=[];cases=('metadata-pin','wrong-source','wrong-summary','inventory-pin','inventory-decoded-bound','archive-hash','archive-size-bound','archive-symlink','archive-fifo','traversal','duplicate','typed-link','entry-bound','observation-bound')
 helper=BASE.parents[1]/'ticket-13/implementation/verify.py'
 for name in cases:
  with tempfile.TemporaryDirectory(prefix='repair-capsule-negative-') as tmp:
   root=Path(tmp);base=root/'.scratch/barness-ai-minimax/evidence/ticket-15/implementation';base.parent.mkdir(parents=True);shutil.copytree(BASE,base)
   target=base.parents[1]/'ticket-13/implementation/verify.py';target.parent.mkdir(parents=True);shutil.copyfile(helper,target)
   archive=root/REL;archive.parent.mkdir(parents=True);shutil.copyfile(common/REL,archive);index=json.loads((base/'index.json').read_text())
   if name=='metadata-pin':(base/'summary.json').write_text('{}')
   elif name=='wrong-source':index['codeCommit']='f'*40
   elif name=='wrong-summary':
    v=json.loads((base/'summary.json').read_text());v['newPaidAuthority']='APPROVED';(base/'summary.json').write_text(json.dumps(v))
   elif name.startswith('inventory-'):
    raw=b'{}' if name=='inventory-pin' else b'x'*((8<<20)+1);p=base/'inventory.json.gz';p.write_bytes(gzip.compress(raw));index['capsule']['inventory']['compressedSHA256']=sha(p)
   elif name=='archive-hash':index['capsule']['archive']['sha256']='0'*64
   elif name=='archive-size-bound':
    with archive.open('wb') as f:f.truncate((32<<20)+1)
    index['capsule']['archive']['bytes']=archive.stat().st_size
   elif name in ('archive-symlink','archive-fifo'):
    archive.unlink()
    if name=='archive-symlink':archive.symlink_to(common/REL)
    else:os.mkfifo(archive,0o600)
   elif name=='observation-bound':index['capsule']['observe']*=3
   else:
    source=gzip.decompress((base/'source-at-start.json.gz').read_bytes())
    with archive.open('wb') as f:
     with gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as z:
      with tarfile.open(fileobj=z,mode='w|') as tar:
       for number in range(32769 if name=='entry-bound' else 2 if name=='duplicate' else 1):
        leaf='../escape' if name=='traversal' else 'release-repair-implementation/empty-'+str(number) if name=='entry-bound' else 'release-repair-implementation/source/committed-source.json'
        member=tarfile.TarInfo(leaf);raw=source
        if name=='entry-bound':member.type=tarfile.DIRTYPE;raw=b''
        if name=='typed-link':member.type=tarfile.SYMTYPE;member.linkname='../../outside';raw=b''
        member.size=len(raw);tar.addfile(member,io.BytesIO(raw))
    index['capsule']['archive'].update(bytes=archive.stat().st_size,sha256=sha(archive))
   (base/'index.json').write_text(json.dumps(index))
   if name!='metadata-pin':seal(base)
   result=probe(base,root)
   if result.returncode==0:raise ValueError('unsafe negative was accepted: '+name)
   results.append(dict(case=name,exitCode=result.returncode,reason=result.stderr.strip()))
 return dict(verified=True,negativeGuards=results,realCredentialsRead=False,vendorCalls=0,originalArchivesMutated=False)

if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=p.parse_args();print(json.dumps(run(a.common_root),indent=2))
