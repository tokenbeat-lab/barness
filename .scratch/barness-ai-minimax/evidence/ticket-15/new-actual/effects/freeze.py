#!/usr/bin/env python3
"""Exclusive snapshot of closed effects. Distinct fixed custody, no producer runs.

The two reviewed archive protocols remain unchanged; snapshot layout differs
from GLM because protocol/visual bodies and a shared executable have own pins.
"""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;COMMON=Path('/Users/cyber/RestoX/barness');E=COMMON/'.evidence/minimax-release-repair-effects-1';EXT=COMMON/'.evidence/minimax-retained/ticket-15-new-actual'
SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';ROUND='release-repair-validation-1';b=readers.small();full=readers.full();sha=lambda v:hashlib.sha256(v).hexdigest()
LINK='no-key-live/20261018T114438.245219000Z/formal-corrupt-symlink/formal-integrity.json';TARGET='53a6f961cad79b77ca8beaf2409530413dd04b7c38f01560f2a51eb7d3131782';links=[]

def raw(path,limit):
 with b.opened(COMMON,path.relative_to(COMMON).as_posix()) as f:
  size=os.fstat(f.fileno()).st_size
  if size>limit:raise ValueError('snapshot regular member bound')
  data=f.read(limit+1)
 if len(data)!=size or len(data)>limit:raise ValueError('snapshot byte bound/change')
 return data

def scan(root,prefix,selected):
 pending=[(root,0)];entries=0
 while pending:
  folder,depth=pending.pop()
  if depth>32:raise ValueError('snapshot depth bound')
  with os.scandir(folder) as rows:
   for row in rows:
    entries+=1
    if entries>200000:raise ValueError('snapshot whole shape bound')
    p=Path(row.path)
    if row.is_symlink():
     target=os.readlink(os.fsencode(p));name=p.relative_to(E).as_posix()
     if not prefix.startswith('preflight/') or name!=LINK or sha(target)!=TARGET:raise ValueError('unexpected snapshot link forbidden')
     links.append(dict(originalRelativePath=name,type='symlink',targetHex=target.hex(),targetSHA256=sha(target),targetFollowed=False));continue
    if row.is_dir(follow_symlinks=False):pending.append((p,depth+1));continue
    if not stat.S_ISREG(row.stat(follow_symlinks=False).st_mode):raise ValueError('regular originals required')
    name=prefix+'/'+p.relative_to(root).as_posix();b.relative(name)
    if name in selected:raise ValueError('duplicate member')
    selected[name]=p

def snapshot(label,selected,inline,observe,small=False):
 archive=EXT/('effects-'+label+'-raw.tar.gz');prefix='repair-effects-'+label;names=sorted(set(selected)|set(inline));pins={};total=0;limit=16<<20 if small else full.LIMITS['memberBytes']
 if len(names)>(32768 if small else full.LIMITS['files']):raise ValueError('snapshot file bound')
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',mtime=0,compresslevel=6) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for name in names:
     data=inline[name] if name in inline else raw(selected[name],limit);total+=len(data)
     if len(data)>limit or total>(160<<20 if small else full.LIMITS['rawMemberBytes']):raise ValueError('snapshot payload bound')
     pins[name]=dict(bytes=len(data),sha256=sha(data));member=tarfile.TarInfo(prefix+'/'+name);member.mode=0o600;member.size=len(data);tar.addfile(member,io.BytesIO(data))
 if archive.stat().st_size>(32<<20 if small else full.LIMITS['compressedBytes']):raise ValueError('snapshot compressed bound')
 for name,path in selected.items():
  if sha(raw(path,limit))!=pins[name]['sha256']:raise ValueError('original changed after snapshot')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0);leaf='inventory.json.gz' if small else label+'-inventory.json.gz'
 if len(inventory)>(8<<20 if small else full.LIMITS['inventoryDecodedBytes']) or len(encoded)>8<<20:raise ValueError('inventory bound')
 (BASE/leaf).write_bytes(encoded);source=raw(E/'source-at-start.json',1<<20);se=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(se)
 observe=sorted(set(observe))
 if len(observe)>(512 if small else 256) or not set(observe)<=set(pins) or any(pins[n]['bytes']>(1<<20 if small else full.LIMITS['observedMetadataBytes']) for n in observe) or sum(pins[n]['bytes'] for n in observe)>(16<<20 if small else full.LIMITS['observedTotalBytes']):raise ValueError('observed metadata bound')
 cap=dict(sourceCommit=SOURCE,archive=dict(file=archive.relative_to(COMMON).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix=prefix),inventory=dict(file=leaf,compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(se),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='frozen/source-at-start.json'),observe=observe,originalsUnchangedAfterSnapshot=True)
 if small:cap['id']='actualRepairEffects7931'
 else:cap['limits']=full.LIMITS
 archive.chmod(0o444);print(json.dumps(dict(capsule=label,files=len(pins),rawBytes=total,compressedBytes=archive.stat().st_size,sha256=cap['archive']['sha256'])),flush=True);return cap,pins

def main():
 if (BASE/'index.json').exists() or any(EXT.glob('effects-*-raw.tar.gz')):raise ValueError('fresh exclusive E retention required')
 closed=b.doc(raw(E/'actual-six-final-summary.json',1<<20))
 if closed['current']['calls']!=102 or closed['current']['attempts']!=102 or closed['total']['calls']!=647 or closed['total']['attempts']!=647:raise ValueError('closed actual six required')
 actual={};preflight={};binary={};roots={}
 for row in os.scandir(E):
  p=Path(row.path)
  if row.is_symlink():raise ValueError('root link forbidden')
  if row.is_file(follow_symlinks=False):
   if row.name=='live.test':continue
   (binary if row.name=='visualeval' else actual)['frozen/'+row.name]=p
  elif row.name in ('no-key-live','no-key-release-contracts'):scan(p,'preflight/'+row.name,preflight)
  elif row.name in ('protocol114','visual-jev','visual-decisions'):
   scan(p,'actual/'+row.name,actual);roots[row.name]='actual/'+row.name
   if row.name=='protocol114':
    sessions=[p for p in p.iterdir() if p.is_dir() and not p.is_symlink()]
    if len(sessions)!=1:raise ValueError('one protocol session required')
    roots[row.name]+='/'+sessions[0].name
  else:raise ValueError('unclassified closed E root '+row.name)
 for name in roots:actual['canonical/'+name+'.json']=COMMON/'.evidence/minimax-launches'/(ROUND+'-'+name+'.json')
 for leaf in ('source-at-start.json','binaries.json','build-visualeval-command.json'):binary['frozen/'+leaf]=E/leaf
 preflight['frozen/source-at-start.json']=E/'source-at-start.json';preflight['frozen/actual-keyless-case-summary.json']=E/'actual-keyless-case-summary.json'
 inline={'controllers/freeze.py':Path(__file__).read_bytes(),'controllers/failure-inventory.md':(BASE/'failure-inventory.md').read_bytes()}
 observed=[n for n in actual if n.startswith(('frozen/','canonical/')) or n.startswith('actual/') and n.endswith(('/multimodal-report.json','/multimodal-integrity.json','/manifest.json','/report.json','/run-config.json','/audit.json','/result.json','/exchanges.json','/assertions.json','/wire.json','/dataset.json','/source-hashes.json','/experiment.json'))]+list(inline)
 cap,pins=snapshot('actual',actual,inline,observed,True)
 pobserve=['frozen/source-at-start.json','frozen/actual-keyless-case-summary.json','projection/known-intentional-link.json']
 for row in b.doc(raw(E/'actual-keyless-case-summary.json',1<<20))['manifests']:
  n='preflight/'+row['manifest'].removeprefix('.evidence/minimax-release-repair-effects-1/');pobserve.append(n);a=n.removesuffix('manifest.json')+'audit.json'
  if a in preflight:pobserve.append(a)
 if len(links)!=1 or sha(os.readlink(os.fsencode(E/LINK)))!=TARGET:raise ValueError('exact intentional link metadata required')
 projection=(json.dumps(dict(links=links),separators=(',',':'))+'\n').encode();pcap,_=snapshot('preflight',preflight,{'projection/known-intentional-link.json':projection},pobserve);bcap,_=snapshot('binary',binary,{},[n for n in binary if n!='frozen/visualeval'])
 gindex=b.doc((BASE.parent/'glm/index.json').read_bytes());gp=b.doc(gzip.decompress((BASE.parent/'glm/binary-inventory.json.gz').read_bytes()));live=gp['frozen/live.test']
 if sha(raw(E/'live.test',64<<20))!=live['sha256']:raise ValueError('shared G live executable body differs')
 value=dict(schema=1,kind='actual-repair-effects-three',actualSource=SOURCE,round=ROUND,boundedHelperSHA256=readers.SMALL,fullReaderSHA256=readers.FULL,roots=roots,capsule=cap,preflight=pcap,binary=bcap,sharedLiveBinary=dict(frontdoor='../glm',indexSHA256=sha((BASE.parent/'glm/index.json').read_bytes()),capsule=gindex['binary'],member='frozen/live.test',pin=live),intentionalLinkProjection=dict(originalRelativePath=LINK,targetSHA256=TARGET,archiveMember='projection/known-intentional-link.json',linksFollowed=False),retentionPaidCalls=0,retentionCredentialsRead=False)
 (BASE/'index.json').write_text(json.dumps(value,separators=(',',':'))+'\n')
if __name__=='__main__':main()
