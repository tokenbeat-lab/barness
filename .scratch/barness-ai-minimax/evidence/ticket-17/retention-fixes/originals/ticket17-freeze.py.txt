#!/usr/bin/env python3
"""Exclusively freeze the complete closed offline TDD, with unchanged full bounds."""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
import contract as c,readers
BASE=Path(__file__).resolve().parent;ROOT=BASE.parents[3];RAW=ROOT/'.evidence/minimax-ticket17';COMMON=Path('/Users/cyber/RestoX/barness');sha=lambda v:hashlib.sha256(v).hexdigest()
def main():
 b=readers.small();full=readers.full();archive=COMMON/c.ARCHIVE
 if archive.exists() or (BASE/'index.json').exists():raise ValueError('fresh exclusive retention required')
 expected=set(c.STAGES)|{'dependency-restore','dependency-restore-2'}
 if {p.name for p in RAW.iterdir()}!=expected:raise ValueError('exact original stage roots required')
 selected={};pending=[(RAW,0)];entries=0;total=0
 while pending:
  parent,depth=pending.pop()
  if depth>32:raise ValueError('snapshot depth bound')
  with os.scandir(parent) as rows:
   for row in rows:
    entries+=1
    if entries>full.LIMITS['entries'] or row.is_symlink():raise ValueError('whole shape limit or link')
    path=Path(row.path)
    if row.is_dir(follow_symlinks=False):pending.append((path,depth+1));continue
    info=row.stat(follow_symlinks=False)
    if not stat.S_ISREG(info.st_mode) or info.st_size>full.LIMITS['memberBytes']:raise ValueError('regular finite snapshot member required')
    name='stages/'+path.relative_to(RAW).as_posix();b.relative(name);selected[name]=path;total+=info.st_size
 if len(selected)>full.LIMITS['files'] or total>full.LIMITS['rawMemberBytes']:raise ValueError('snapshot file/payload bound')
 for label,(code,_) in c.STAGES.items():
  v=json.loads((RAW/label/'command.json').read_bytes())
  if v['state']!='COMPLETED' or type(v['exitCode']) is not int or v['exitCode']!=code or not v['sourceUnchangedAfter']:raise ValueError('all original commands must be closed with honest numeric codes')
 metadata=('source-proof.json','final-checks.json','original-spec-review.json','original-standards-review.json','failure-inventory.md','tool-fixture-repair-source-proof.json','tool-guard-reapplied-proof.json','trace-expansion.json','trace-byte-delta.json','trace-before.json.gz','trace-after.json.gz')
 for name in metadata:selected['final/'+name]=BASE/name
 selected['controllers/runner.py']=Path('/tmp/barness-ticket17-run.py').resolve();selected['controllers/freeze.py']=Path(__file__)
 
 if len(selected)>full.LIMITS['files']:raise ValueError('final snapshot file bound')
 pins={};prefix='ticket17-offline-tdd';observed=[]
 def read(path):
  if path.is_relative_to(ROOT):root=ROOT;relative=path.relative_to(ROOT).as_posix()
  else:root=path.parent;relative=path.name
  with b.opened(root,relative) as f:
   size=os.fstat(f.fileno()).st_size
   if size>full.LIMITS['memberBytes']:raise ValueError('member size bound')
   data=f.read(full.LIMITS['memberBytes']+1)
  if len(data)!=size:raise ValueError('snapshot source changed')
  return data
 archive.parent.mkdir(parents=True,exist_ok=True)
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for name,path in sorted(selected.items()):
     data=read(path);pins[name]=dict(bytes=len(data),sha256=sha(data));member=tarfile.TarInfo(prefix+'/'+name);member.mode=0o600;member.size=len(data);tar.addfile(member,io.BytesIO(data))
 if sum(v['bytes'] for v in pins.values())>full.LIMITS['rawMemberBytes'] or archive.stat().st_size>full.LIMITS['compressedBytes']:raise ValueError('fixed final archive envelope')
 for name,path in selected.items():
  if sha(read(path))!=pins[name]['sha256']:raise ValueError('original stage changed after freeze')
 roots={}
 for label in c.STAGES:
  observed.append('stages/'+label+'/command.json');manifests=[n for n in pins if n.startswith('stages/'+label+'/bundles/') and n.endswith('/manifest.json')]
  if not c.STAGES[label][1]:
   if manifests:raise ValueError('unexpected no-bundle stage')
   continue
  if len(manifests)!=1:raise ValueError('exact original single bundle required')
  root=manifests[0].removesuffix('/manifest.json');roots[label]=root;observed.extend([root+'/manifest.json',root+'/audit.json'])
  if label in c.TRUE_RED or label=='green-final-native-shapes':
   manifest=json.loads(read(selected[root+'/manifest.json']))
   for case in manifest['cases']:observed.extend(root+'/'+case['dir']+'/'+leaf for leaf in ('requests.json','result.json'))
 observed+=['final/source-proof.json','final/tool-fixture-repair-source-proof.json','final/tool-guard-reapplied-proof.json','stages/dependency-restore-2/proof.json']
 if len(observed)>256 or len(set(observed))!=len(observed) or sum(pins[n]['bytes'] for n in observed)>full.LIMITS['observedTotalBytes'] or any(pins[n]['bytes']>full.LIMITS['observedMetadataBytes'] for n in observed):raise ValueError('fixed observation count')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0);
 if len(inventory)>full.LIMITS['inventoryDecodedBytes'] or len(encoded)>full.LIMITS['inventoryCompressedBytes']:raise ValueError('fixed inventory bound')
 (BASE/'inventory.json.gz').write_bytes(encoded);source=(BASE/'source-proof.json').read_bytes();source_encoded=gzip.compress(source,mtime=0);(BASE/'source-proof.json.gz').write_bytes(source_encoded)
 with archive.open('rb') as f:digest=hashlib.file_digest(f,'sha256').hexdigest()
 index=dict(schema=1,kind='ticket17-offline-final-payload-tdd',sourceCommit=c.SOURCE,archive=dict(file=c.ARCHIVE,bytes=archive.stat().st_size,sha256=digest,prefix=prefix),limits=full.LIMITS,inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-proof.json.gz',compressedSHA256=sha(source_encoded),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='final/source-proof.json'),observe=sorted(observed),roots=roots,originalsUnchangedAfterSnapshot=True,credentialsRead=False,paidCalls=0,releaseCertified=False)
 (BASE/'index.json').write_text(json.dumps(index,separators=(',',':'))+'\n');archive.chmod(0o444);print(json.dumps(dict(files=len(pins),rawBytes=sum(v['bytes'] for v in pins.values()),archive=index['archive'],observed=len(observed)),indent=2))
if __name__=='__main__':main()
