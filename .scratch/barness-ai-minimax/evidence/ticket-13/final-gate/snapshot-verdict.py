"""Freeze only the closed compiled gate outer metadata, excluding sealed suites."""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
ROOT=Path('/Users/cyber/RestoX/barness');OWN=Path('/tmp/barness-minimax-routed-validation')
BASE=OWN/'.scratch/barness-ai-minimax/evidence/ticket-13/final-gate';EXT=ROOT/'.evidence/minimax-retained/ticket-13-final-gate';A=ROOT/'.evidence/minimax-routed-final-compiled-1'
sha=lambda raw:hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,ensure_ascii=False,separators=(',',':'))+'\n')
summary=json.loads((A/'summary.json').read_text());report=json.loads((A/'gate/release-report.json').read_text())
if summary['gateExit']!=1 or summary['releasePass'] is not False or summary['allFourNumericExitsPresent'] is not True or report['pass'] is not False:raise RuntimeError('closed failed actual verdict required')
selected={}
for folder,dirs,files in os.walk(A,followlinks=False):
 for d in list(dirs):
  p=Path(folder)/d
  if p in (A/'gate/offline',A/'gate/race'):dirs.remove(d)
  elif p.is_symlink():raise RuntimeError('no symlink metadata directory')
 for name in files:
  p=Path(folder)/name
  if p==A/'releasegate':continue
  selected['outer/'+p.relative_to(A).as_posix()]=p
selected['source/source-at-start.json']=ROOT/'.evidence/minimax-routed-legacy-1/source-at-start.json'
selected['diagnosis/postgate-gap-2.json']=Path('/tmp/barness-routed-postgate-gap-2.json')
if len(selected)>256 or sum(p.lstat().st_size for p in selected.values())>32<<20:raise RuntimeError('bounded outer metadata required')
inventory={};archive=EXT/'verdict-raw.tar.gz'
with archive.open('xb') as output:
 with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
  with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
   for name,path in sorted(selected.items()):
    info=path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_size>8<<20:raise RuntimeError('bounded regular metadata original required')
    raw=path.read_bytes();inventory[name]=dict(bytes=len(raw),sha256=sha(raw));member=tarfile.TarInfo('routed-compiled-verdict/'+name);member.size=len(raw);member.mode=0o600;tar.addfile(member,io.BytesIO(raw))
for name,path in selected.items():
 if sha(path.read_bytes())!=inventory[name]['sha256']:raise RuntimeError('original closed metadata changed')
encoded=gzip.compress((json.dumps(inventory,sort_keys=True,separators=(',',':'))+'\n').encode(),mtime=0);(BASE/'verdict-inventory.json.gz').write_bytes(encoded)
raw=gzip.decompress(encoded);source=selected['source/source-at-start.json'].read_bytes();sourcegz=gzip.compress(source,mtime=0)
if sourcegz!=(BASE/'source-at-start.json.gz').read_bytes():raise RuntimeError('original suite sources differ')
limits=json.loads((BASE/'ordinary-index.json').read_text())['limits'];index=dict(schema=1,kind='completed-outer-original-failed',sourceCommit=summary['sourceCommit'],suite='outer',limits=limits,
 archive=dict(file=archive.relative_to(ROOT).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix='routed-compiled-verdict'),
 inventory=dict(file='verdict-inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw),files=len(inventory)),
 sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(sourcegz),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='source/source-at-start.json'),
 observe=sorted(inventory),originalsUnchangedAfterSnapshot=True,retentionCredentialsRead=False,retentionPaidCalls=0,outerCompleted=True)
dump(BASE/'verdict-index.json',index)
binary=A/'releasegate';dump(BASE/'binary-reference.json',dict(schema=1,file='$REPO/'+binary.relative_to(ROOT).as_posix(),bytes=binary.stat().st_size,sha256=sha(binary.read_bytes()),copied=False,
 existingImmutableArchive='.evidence/minimax-retained/ticket-13-keyless-launch-history/keyless-launch-history-raw.tar.gz',existingMember='compiled-prepared/releasegate'))
archive.chmod(0o444);print(json.dumps(dict(files=len(inventory),archiveBytes=archive.stat().st_size,archiveSHA256=index['archive']['sha256']),indent=2))
