#!/usr/bin/env python3
"""Freeze only the explicitly closed ordinary subtree, receipt and immutable plan."""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;ROOT=Path('/Users/cyber/RestoX/barness');B=ROOT/'.evidence/minimax-release-repair-incomplete-gate-1';SESSION='20261018T120404.152055000Z';ORD=B/'gate/offline'/SESSION;EXT=ROOT/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';PLAN='f97c97c158189bc741a22ea714a1cc64053ea79755d688f8a66450d87fab6985';b=readers.small();full=readers.full();sha=lambda v:hashlib.sha256(v).hexdigest()
def raw(path):
 with b.opened(ROOT,path.relative_to(ROOT).as_posix()) as f:
  size=os.fstat(f.fileno()).st_size
  if size>64<<20:raise ValueError('regular member bound')
  data=f.read((64<<20)+1)
 if len(data)!=size or len(data)>64<<20:raise ValueError('original changed during snapshot')
 return data

def main():
 archive=EXT/'incomplete-ordinary-raw.tar.gz'
 if archive.exists() or (BASE/'index.json').exists():raise ValueError('fresh exclusive ordinary capsule required')
 plan=raw(B/'plan-v2.json');receipt=raw(B/'gate-command-exits/ordinary.json');source=raw(ROOT/'.evidence/minimax-release-repair-glm-1/source-at-start.json');p=b.doc(plan);r=b.doc(receipt)
 if sha(plan)!=PLAN or p['source']!=b.doc(source) or p['source']['sourceCommit']!=SOURCE or len(p['source']['files'])!=1056 or r['state']!='COMPLETED' or type(r['exitCode']) is not int or r['exitCode']!=1 or r['credentialsRead'] is not False or r['argv']!=['/usr/local/go/bin/go','test','-count=1','-timeout=20m','./...']:raise ValueError('closed original ordinary/source/plan required')
 selected={};pending=[(ORD,0)];entries=0;total=0
 while pending:
  folder,depth=pending.pop()
  if depth>32:raise ValueError('snapshot depth bound')
  with os.scandir(folder) as rows:
   for row in rows:
    entries+=1
    if entries>full.LIMITS['entries']:raise ValueError('snapshot whole shape bound')
    if row.is_symlink():raise ValueError('snapshot links forbidden')
    path=Path(row.path)
    if row.is_dir(follow_symlinks=False):pending.append((path,depth+1));continue
    info=row.stat(follow_symlinks=False)
    if not stat.S_ISREG(info.st_mode) or info.st_size>full.LIMITS['memberBytes']:raise ValueError('bounded regular originals required')
    name='ordinary/'+SESSION+'/'+path.relative_to(ORD).as_posix();b.relative(name);selected[name]=path;total+=info.st_size
    if len(selected)>full.LIMITS['files'] or total>full.LIMITS['rawMemberBytes']:raise ValueError('snapshot file/payload bound')
 inline={'source/source-at-start.json':source,'receipts/ordinary.json':receipt,'frozen/plan-v2.json':plan};pins={};prefix='repair-incomplete-ordinary';print(json.dumps(dict(snapshot='ordinary',originalFiles=len(selected),rawBytes=total)),flush=True)
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for count,name in enumerate(sorted(set(selected)|set(inline)),1):
     data=inline[name] if name in inline else raw(selected[name]);pins[name]=dict(bytes=len(data),sha256=sha(data));item=tarfile.TarInfo(prefix+'/'+name);item.mode=0o600;item.size=len(data);tar.addfile(item,io.BytesIO(data))
     if count%30000==0:print(json.dumps(dict(archivedFiles=count)),flush=True)
 if archive.stat().st_size>full.LIMITS['compressedBytes'] or sum(v['bytes'] for v in pins.values())>full.LIMITS['rawMemberBytes']:raise ValueError('snapshot compressed/payload bound')
 for name,path in selected.items():
  if sha(raw(path))!=pins[name]['sha256']:raise ValueError('ordinary original changed after snapshot')
 if plan!=raw(B/'plan-v2.json') or receipt!=raw(B/'gate-command-exits/ordinary.json') or source!=raw(ROOT/'.evidence/minimax-release-repair-glm-1/source-at-start.json'):raise ValueError('frozen original metadata changed after snapshot')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0)
 if len(inventory)>full.LIMITS['inventoryDecodedBytes'] or len(encoded)>full.LIMITS['inventoryCompressedBytes']:raise ValueError('inventory bound')
 (BASE/'inventory.json.gz').write_bytes(encoded);se=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(se);manifest='ordinary/'+SESSION+'/manifest.json';v=b.doc(raw(selected[manifest]));observe=[manifest,'ordinary/'+SESSION+'/audit.json',*inline]
 for case in v['cases']:
  if case['status']=='PASS':continue
  for leaf in case.get('artifacts',{}):
   n='ordinary/'+SESSION+'/'+case['dir']+'/'+leaf
   if leaf.endswith('.json') and pins[n]['bytes']<=1<<20:observe.append(n)
 cap=dict(sourceCommit=SOURCE,suite='ordinary',limits=full.LIMITS,archive=dict(file=archive.relative_to(ROOT).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix=prefix),inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(se),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='source/source-at-start.json'),observe=sorted(set(observe)),mainManifest=manifest,originalsUnchangedAfterSnapshot=True,retentionPaidCalls=0,retentionCredentialsRead=False,gateVerdictCaptured=False,raceCaptured=False)
 (BASE/'index.json').write_text(json.dumps(cap,separators=(',',':'))+'\n');archive.chmod(0o444);print(json.dumps(dict(archive=cap['archive'],files=len(pins),rawBytes=sum(v['bytes'] for v in pins.values()),observations=len(cap['observe']))),flush=True)
if __name__=='__main__':main()
