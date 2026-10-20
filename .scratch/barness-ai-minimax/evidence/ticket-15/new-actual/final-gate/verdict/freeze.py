#!/usr/bin/env python3
"""Exclusively freeze all closed outer B originals without duplicating suite bulk."""
import argparse,gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;ROOT=Path('/Users/cyber/RestoX/barness');B=ROOT/'.evidence/minimax-release-repair-incomplete-gate-1';EXT=ROOT/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';b=readers.small();full=readers.full();sha=lambda v:hashlib.sha256(v).hexdigest()
def raw(path):
 with b.opened(ROOT,path.relative_to(ROOT).as_posix()) as f:
  size=os.fstat(f.fileno()).st_size
  if size>64<<20:raise ValueError('fixed regular member cap')
  data=f.read((64<<20)+1)
 if len(data)!=size:raise ValueError('original changed during read')
 return data
def archive_pin(common,archive,limit):
 # Independent standalone retention creators keep their existing reader protocol.
 with b.opened(common,archive.relative_to(common).as_posix()) as stream:
  before=os.fstat(stream.fileno())
  if before.st_size>limit:raise ValueError('compressed archive size bound')
  digest=hashlib.sha256();count=0
  for chunk in iter(lambda:stream.read(1<<20),b''):
   count+=len(chunk)
   if count>limit:raise ValueError('compressed archive byte bound')
   digest.update(chunk)
  after=os.fstat(stream.fileno())
  if count!=before.st_size or (before.st_size,before.st_mtime_ns,before.st_ctime_ns)!=(after.st_size,after.st_mtime_ns,after.st_ctime_ns):raise ValueError('archive changed during final bounded hash')
 return count,digest.hexdigest()

def main(common):
 archive=common/'.evidence/minimax-retained/ticket-15-new-actual/incomplete-outer-raw.tar.gz'
 if archive.exists() or (BASE/'index.json').exists():raise ValueError('exclusive fresh outer capsule required')
 selected={};pending=[(B,0)];entries=0;total=0
 while pending:
  folder,depth=pending.pop()
  if depth>32:raise ValueError('whole snapshot depth cap')
  with os.scandir(folder) as rows:
   for row in rows:
    path=Path(row.path);entries+=1
    if entries>4096 or row.is_symlink():raise ValueError('whole snapshot shape/link cap')
    if row.is_dir(follow_symlinks=False):
     if path not in {B/'gate/offline',B/'gate/race'}:pending.append((path,depth+1))
     continue
    info=row.stat(follow_symlinks=False)
    if not stat.S_ISREG(info.st_mode) or info.st_size>64<<20:raise ValueError('regular outer originals required')
    name='outer/'+path.relative_to(B).as_posix();b.relative(name);selected[name]=path;total+=info.st_size
    if len(selected)>512 or total>64<<20:raise ValueError('outer snapshot file/payload cap')
 if len(selected)!=63:raise ValueError('all exact63 original outer files required')
 done=b.doc(raw(B/'completed.json'))
 if done['state']!='COMPLETED' or type(done['outerExit']) is not int or done['outerExit']!=1 or done['reportPresent'] is not True or done['releasePass'] is not False or done['releaseCertified'] is not False:raise ValueError('closed honest failed outer verdict required')
 source=raw(ROOT/'.evidence/minimax-release-repair-glm-1/source-at-start.json');inline={'source/source-at-start.json':source};pins={};prefix='repair-incomplete-outer'
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for name in sorted(set(selected)|set(inline)):
     data=inline[name] if name in inline else raw(selected[name]);pins[name]=dict(bytes=len(data),sha256=sha(data));member=tarfile.TarInfo(prefix+'/'+name);member.mode=0o600;member.size=len(data);tar.addfile(member,io.BytesIO(data))
 if archive.stat().st_size>32<<20 or sum(v['bytes'] for v in pins.values())>64<<20:raise ValueError('narrow creation envelope')
 for name,path in selected.items():
  if sha(raw(path))!=pins[name]['sha256']:raise ValueError('original outer changed after snapshot')
 if source!=raw(ROOT/'.evidence/minimax-release-repair-glm-1/source-at-start.json'):raise ValueError('original source changed after snapshot')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0);se=gzip.compress(source,mtime=0);(BASE/'inventory.json.gz').write_bytes(encoded);(BASE/'source-at-start.json.gz').write_bytes(se)
 refs={label:dict(frontdoor='../'+label,indexSHA256=sha(b.read(BASE.parent,label+'/index.json',1<<20))) for label in ('ordinary','race')}
 observe=sorted(set(pins)-{'outer/releasegate'})
 archive_bytes,archive_sha256=archive_pin(common,archive,32<<20)
 index=dict(kind='closed-incomplete-gate-verdict',sourceCommit=SOURCE,limits=full.LIMITS,archive=dict(file=archive.relative_to(common).as_posix(),bytes=archive_bytes,sha256=archive_sha256,prefix=prefix),inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(se),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='source/source-at-start.json'),observe=observe,suiteReferences=refs,originalsUnchangedAfterSnapshot=True,releaseCertified=False,credentialsRead=False,paidCalls=0,creationBounds=dict(files=512,entries=4096,depth=32,payloadBytes=64<<20,memberBytes=64<<20,compressedBytes=32<<20))
 (BASE/'index.json').write_text(json.dumps(index,separators=(',',':'))+'\n');archive.chmod(0o444);print(json.dumps(dict(files=len(pins),rawBytes=sum(v['bytes'] for v in pins.values()),archive=index['archive']),indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,default=ROOT);main(p.parse_args().common_root)
