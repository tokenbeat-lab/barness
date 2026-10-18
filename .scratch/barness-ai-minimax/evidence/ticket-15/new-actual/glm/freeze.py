#!/usr/bin/env python3
"""Snapshot only closed G and its refused publisher; no producer command runs."""
import gzip,hashlib,io,json,os,stat,subprocess,sys,tarfile
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;OWN=BASE.parents[5];COMMON=Path('/Users/cyber/RestoX/barness');G=COMMON/'.evidence/minimax-release-repair-glm-1';PUBLICATION=COMMON/'.evidence/minimax-release-repair-glm-publication-1'
EXT=COMMON/'.evidence/minimax-retained/ticket-15-new-actual';SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';ROUND='release-repair-validation-1'
b=readers.small();large=readers.full();sha=lambda raw:hashlib.sha256(raw).hexdigest()
KNOWN_LINK='no-key-live/20261018T110935.233553000Z/formal-corrupt-symlink/formal-integrity.json';LINK_SHA='18ebed3cd4eba35406b844c0abba746151949978df4204c4c5b0578b6eff235c';links=[]

def dump(path,value):path.write_text(json.dumps(value,separators=(',',':'))+'\n')
def raw(path,limit):
 relative=path.relative_to(COMMON).as_posix()
 with b.opened(COMMON,relative) as stream:
  size=os.fstat(stream.fileno()).st_size
  if size>limit:raise ValueError('regular snapshot member limit')
  value=stream.read(limit+1)
 if len(value)!=size or len(value)>limit:raise ValueError('snapshot byte bound/change')
 return value

def scan(root,prefix,selected,maximum,limit):
 pending=[(root,0)];entries=0;size=0
 while pending:
  folder,depth=pending.pop()
  if depth>32:raise ValueError('snapshot depth bound')
  with os.scandir(folder) as rows:
   for row in rows:
    entries+=1
    if entries>maximum:raise ValueError('snapshot entire shape bound')
    path=Path(row.path)
    if row.is_symlink():
     relative=path.relative_to(G).as_posix();target=os.readlink(os.fsencode(path))
     if not prefix.startswith('preflight/') or relative!=KNOWN_LINK or sha(target)!=LINK_SHA:raise ValueError('unexpected snapshot link forbidden')
     links.append(dict(originalRelativePath=relative,type='symlink',targetHex=target.hex(),targetSHA256=sha(target),targetFollowed=False));continue
    if row.is_dir(follow_symlinks=False):pending.append((path,depth+1));continue
    info=row.stat(follow_symlinks=False)
    if not stat.S_ISREG(info.st_mode) or info.st_size>limit:raise ValueError('bounded regular snapshot originals required')
    name=prefix+'/'+path.relative_to(root).as_posix();b.relative(name)
    if name in selected:raise ValueError('snapshot duplicate member')
    selected[name]=path;size+=info.st_size
 return entries,size

def snapshot(label,selected,inline,small=False,observe=()):
 names=sorted(set(selected)|set(inline));limit=16<<20 if small else large.LIMITS['memberBytes'];total=0;pins={};archive=EXT/('glm-'+label+'-raw.tar.gz');prefix='repair-glm-'+label
 if len(names)>(32768 if small else large.LIMITS['files']):raise ValueError('snapshot file count bound')
 with archive.open('xb') as output:
  with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
   with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
    for name in names:
     data=inline[name] if name in inline else raw(selected[name],limit);total+=len(data)
     if total>(160<<20 if small else large.LIMITS['rawMemberBytes']):raise ValueError('snapshot regular payload bound')
     pins[name]=dict(bytes=len(data),sha256=sha(data));member=tarfile.TarInfo(prefix+'/'+name);member.size=len(data);member.mode=0o600;tar.addfile(member,io.BytesIO(data))
 if archive.stat().st_size>(32<<20 if small else large.LIMITS['compressedBytes']):raise ValueError('snapshot compressed bound')
 for name,path in selected.items():
  if sha(raw(path,limit))!=pins[name]['sha256']:raise ValueError('selected original changed during snapshot')
 inventory=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(inventory,mtime=0)
 if len(inventory)>(8<<20 if small else large.LIMITS['inventoryDecodedBytes']):raise ValueError('snapshot inventory bound')
 file='inventory.json.gz' if small else label+'-inventory.json.gz';(BASE/file).write_bytes(encoded)
 source=raw(G/'source-at-start.json',1<<20);sencoded=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(sencoded)
 observation=sorted(set(observe))
 if len(observation)>(512 if small else 256) or not set(observation)<=set(pins) or any(pins[n]['bytes']>(1<<20 if small else large.LIMITS['observedMetadataBytes']) for n in observation) or sum(pins[n]['bytes'] for n in observation)>(16<<20 if small else large.LIMITS['observedTotalBytes']):raise ValueError('snapshot observed metadata bound')
 cap=dict(sourceCommit=SOURCE,archive=dict(file=archive.relative_to(COMMON).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix=prefix),inventory=dict(file=file,compressedSHA256=sha(encoded),decodedBytes=len(inventory),decodedSHA256=sha(inventory),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(sencoded),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='frozen/source-at-start.json'),observe=observation,originalsUnchangedAfterSnapshot=True)
 if small:cap['id']='actualRepairGLM7931'
 else:cap['limits']=large.LIMITS
 archive.chmod(0o444);print(json.dumps(dict(snapshot=label,files=len(pins),rawBytes=total,archiveBytes=archive.stat().st_size,archiveSHA256=cap['archive']['sha256'])),flush=True)
 return cap,pins

def main():
 if (BASE/'index.json').exists() or EXT.exists():raise ValueError('fresh exclusive retention output required')
 source=b.doc(raw(G/'source-at-start.json',1<<20));closed=b.doc(raw(G/'actual-glm-final-summary.json',1<<20));refusal=b.doc(raw(G/'strict-publication-refused.json',1<<20))
 if source['sourceCommit']!=SOURCE or closed['current']['calls']!=66 or closed['current']['attempts']!=66 or refusal['actualExitCode']!=1 or refusal['unchanged'] is not True:raise ValueError('closed original G and strict refusal required')
 EXT.mkdir();paid={};preflight={};binary={};roots={}
 for row in os.scandir(G):
  path=Path(row.path)
  if row.is_symlink():raise ValueError('closed G root link forbidden')
  if row.is_file(follow_symlinks=False):
   (binary if row.name=='live.test' else paid)['frozen/'+row.name]=path
  elif row.name in ('no-key-live','no-key-release-contracts'):scan(path,'preflight/'+row.name,preflight,200000,64<<20)
  elif row.name in ('glm-chat','glm-responses','glm-messages'):
   route=row.name.removeprefix('glm-');scan(path,'actual/'+route,paid,32768,16<<20)
   sessions=[p for p in path.iterdir() if p.is_dir() and not p.is_symlink()]
   if len(sessions)!=1:raise ValueError('one actual own GLM session required')
   roots[route]='actual/'+route+'/'+sessions[0].name
  else:raise ValueError('unclassified closed G directory: '+row.name)
 for route in ('chat','responses','messages'):
  path=COMMON/'.evidence/minimax-launches'/(ROUND+'-glm-'+route+'.json');value=b.doc(raw(path,1<<20))
  if value['state']!='COMPLETED' or value['phase']!='glm' or value['sourceCommit']!=SOURCE:raise ValueError('closed canonical G authority required')
  paid['canonical/'+route+'.json']=path
 scan(PUBLICATION,'publication',paid,1024,16<<20)
 for name in ('source-at-start.json','binaries.json','build-live.test-command.json'):
  binary['frozen/'+name]=G/name
 preflight['frozen/source-at-start.json']=G/'source-at-start.json';preflight['frozen/actual-keyless-case-summary.json']=G/'actual-keyless-case-summary.json'
 inline={}
 for leaf in ('live-captures.json','live-captures.sha256','known-usage.json','support-views.json'):
  name='ai/live/testdata/glm-formal/'+leaf;inline['selectors/'+leaf]=subprocess.check_output(['git','show',SOURCE+':'+name],cwd=OWN)
 inline['references/approved-retention-plan.md']=Path('/tmp/barness-repair-actual-retention-plan.md').read_bytes();inline['references/future-finalizer-design-index.json']=(OWN/'.scratch/barness-ai-minimax/evidence/ticket-15/finalizer-design/index.json').read_bytes()
 inline['controllers/freeze.py']=Path(__file__).read_bytes();inline['controllers/failure-inventory.md']=(BASE/'failure-inventory.md').read_bytes()
 observe=[n for n in paid if n.startswith('frozen/') or n.startswith('canonical/') or n.startswith('actual/') and n.endswith(('/formal-glm-report.json','/glm-integrity.json','/glm-execution-source.json','/live-report.json','/audit.json','/assertions.json','/attempt-1-result.json','/attempt-1-exchanges.json'))]+list(inline)
 cap,pins=snapshot('actual',paid,inline,True,observe);cap['archive']['file']=cap['archive']['file'].replace('glm-actual-raw.tar.gz','glm-raw.tar.gz');(EXT/'glm-actual-raw.tar.gz').rename(EXT/'glm-raw.tar.gz')
 manifest_names=[n for n in preflight if n.startswith('frozen/')]
 for item in b.doc(raw(G/'actual-keyless-case-summary.json',1<<20))['manifests']:
  name='preflight/'+item['manifest'].removeprefix('.evidence/minimax-release-repair-glm-1/');manifest_names.append(name);audit=name.removesuffix('manifest.json')+'audit.json'
  if audit in preflight:manifest_names.append(audit)
 projection=(json.dumps(dict(projection='one exact intentional negative-fixture link as regular metadata only',links=links),separators=(',',':'))+'\n').encode()
 if len(links)!=1 or sha(os.readlink(os.fsencode(G/KNOWN_LINK)))!=LINK_SHA:raise ValueError('original exact link metadata changed')
 pcap,_=snapshot('preflight',preflight,{'projection/known-intentional-link.json':projection},False,manifest_names+['projection/known-intentional-link.json']);bcap,_=snapshot('binary',binary,{},False,[n for n in binary if n!='frozen/live.test'])
 index=dict(schema=1,kind='actual-failed-repair-glm-three',actualSource=SOURCE,round=ROUND,boundedHelperSHA256=readers.SMALL,fullReaderSHA256=readers.FULL,intentionalLinkProjection=dict(originalRelativePath=KNOWN_LINK,targetSHA256=LINK_SHA,archiveMember='projection/known-intentional-link.json',linksFollowed=False),routes=roots,capsule=cap,preflight=pcap,binary=bcap,realCredentialsReadByRetention=False,retentionPaidCalls=0)
 if sha(os.readlink(os.fsencode(G/KNOWN_LINK)))!=LINK_SHA:raise ValueError('original link target changed after snapshot')
 dump(BASE/'index.json',index)
if __name__=='__main__':main()
