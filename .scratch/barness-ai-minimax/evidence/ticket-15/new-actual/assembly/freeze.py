#!/usr/bin/env python3
"""Exclusive regular-only snapshot of the explicitly closed failed A branch."""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;COMMON=Path('/Users/cyber/RestoX/barness');A=COMMON/'.evidence/minimax-release-repair-final-1';EXT=COMMON/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';b=readers.small();sha=lambda v:hashlib.sha256(v).hexdigest()

def raw(path):
 with b.opened(COMMON,path.relative_to(COMMON).as_posix()) as f:
  size=os.fstat(f.fileno()).st_size
  if size>16<<20:raise ValueError('regular snapshot member limit')
  data=f.read((16<<20)+1)
 if len(data)!=size or len(data)>16<<20:raise ValueError('snapshot original changed')
 return data

def main():
 archive=EXT/'assembly-failed-raw.tar.gz'
 if archive.exists() or (BASE/'index.json').exists():raise ValueError('fresh exclusive assembly retention required')
 selected={};pending=[(A,0)];entries=0
 while pending:
  folder,depth=pending.pop()
  if depth>32:raise ValueError('snapshot depth bound')
  with os.scandir(folder) as rows:
   for row in rows:
    entries+=1
    if entries>32768:raise ValueError('snapshot whole shape bound')
    if row.is_symlink():raise ValueError('snapshot links forbidden')
    p=Path(row.path)
    if row.is_dir(follow_symlinks=False):pending.append((p,depth+1));continue
    if not stat.S_ISREG(row.stat(follow_symlinks=False).st_mode):raise ValueError('regular originals required')
    name='original/'+p.relative_to(A).as_posix();b.relative(name);selected[name]=p
 if len(selected)!=603:raise ValueError('exact closed A603 originals required')
 inputs=b.doc(raw(A/'assemble-inputs.json'))
 if inputs['source']['sourceCommit']!=SOURCE or b.doc(raw(A/'validate-four-candidate-command.json'))['exitCode']!=1 or (A/'assemble-completed.json').exists():raise ValueError('actual failed assembly closure required')
 source=raw(COMMON/'.evidence/minimax-release-repair-effects-1/source-at-start.json');matrix=raw(COMMON/'ai/live/support-matrix.json');inline={'frozen/source-at-start.json':source,'selectors/default-matrix.json':matrix,'controllers/freeze.py':Path(__file__).read_bytes(),'controllers/failure-inventory.md':(BASE/'failure-inventory.md').read_bytes()};pins={};total=0;prefix='repair-assembly-failed'
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for name in sorted(set(selected)|set(inline)):
     data=inline[name] if name in inline else raw(selected[name]);total+=len(data)
     if total>160<<20 or len(data)>16<<20:raise ValueError('snapshot payload bound')
     pins[name]=dict(bytes=len(data),sha256=sha(data));member=tarfile.TarInfo(prefix+'/'+name);member.mode=0o600;member.size=len(data);tar.addfile(member,io.BytesIO(data))
 if archive.stat().st_size>32<<20:raise ValueError('snapshot compressed bound')
 for name,path in selected.items():
  if sha(raw(path))!=pins[name]['sha256']:raise ValueError('original A changed after snapshot')
 if source!=raw(COMMON/'.evidence/minimax-release-repair-effects-1/source-at-start.json') or matrix!=raw(COMMON/'ai/live/support-matrix.json'):raise ValueError('source/default matrix changed after snapshot')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0);(BASE/'inventory.json.gz').write_bytes(encoded);se=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(se)
 observe=[n for n in pins if n.startswith(('frozen/','selectors/','controllers/')) or n.startswith('original/') and n.count('/')==1 or n in {'original/jev-support/'+f for f in ('live-report.json','formal-integrity.json','audit.json','manifest.json')}]
 if len(observe)>512 or sum(pins[n]['bytes'] for n in observe)>16<<20 or any(pins[n]['bytes']>1<<20 for n in observe):raise ValueError('observed metadata bound')
 cap=dict(id='failedRepairAssembly7931',sourceCommit=SOURCE,archive=dict(file=archive.relative_to(COMMON).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix=prefix),inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(se),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='frozen/source-at-start.json'),observe=sorted(observe),originalsUnchangedAfterSnapshot=True)
 refs={label:dict(frontdoor='../'+label,indexSHA256=sha((BASE.parent/label/'index.json').read_bytes())) for label in ('glm','effects')};helper=BASE.parents[1]/'finalizer-design/helper-v4.py.txt'
 index=dict(schema=1,kind='closed-failed-repair-assembly',sourceCommit=SOURCE,capsule=cap,boundedHelperSHA256=readers.SMALL,nativeReferences=refs,operationalHelperReference=dict(file='../../finalizer-design/helper-v4.py.txt',sha256=sha(helper.read_bytes())),originalRegularFiles=603,originalLinks=0,retentionPaidCalls=0,retentionCredentialsRead=False)
 (BASE/'index.json').write_text(json.dumps(index,separators=(',',':'))+'\n');archive.chmod(0o444);print(json.dumps(dict(files=len(pins),originalFiles=603,rawBytes=total,archive=cap['archive']),indent=2))
if __name__=='__main__':main()
