"""Bounded immutable localTLS TDD snapshot only; does not execute captures."""
import gzip,hashlib,io,json,os,subprocess,tarfile
from pathlib import Path
ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness');A=ROOT/'.evidence/minimax-glm-host-deadline-1'
BASE=ROOT/'.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment';EXT=COMMON/'.evidence/minimax-retained/ticket-16-glm-header-alignment';SOURCE='d8e7100695a5747a801bf5f8966cef1a2d129e4c';CODE='dcd30f7ce6bf65f5bd1101697c8e5c98155ceb98'
sha=lambda raw:hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,separators=(',',':'))+'\n')
if BASE.exists() or EXT.exists():raise ValueError('fresh retention roots required')
BASE.mkdir(parents=True);EXT.mkdir();selected={};inline={}
def add(p,n):
 if p.is_symlink() or not p.is_file() or p.stat().st_size>1<<20 or n in selected:raise ValueError('bounded regular originals only')
 selected[n]=p
for folder,dirs,files in os.walk(A,followlinks=False):
 for d in dirs:
  if (Path(folder)/d).is_symlink():raise ValueError('no original links')
 for f in files:
  p=Path(folder)/f;add(p,'tdd/'+p.relative_to(A).as_posix())
for f in ('barness-ticket16-deadline-tests.py','barness-ticket16-source-impact.py'):add(Path('/tmp')/f,'controllers/'+f)
for route in ('chat','responses','messages'):
 p=next((COMMON/'.evidence/minimax-global-formal-glm'/('mm-global-formal-m3-'+route)).glob('*/minimax-report.json'));add(p,'historical/m3-'+route+'.json')
 p=next((COMMON/'.evidence/minimax-routed-glm-1'/('glm-'+route)).glob('*/glm-execution-source.json'));add(p,'historical/glm-'+route+'-source.json')
for leaf in ('ai/live/glm_assembly_test.go','ai/live/main_test.go'):
 inline['source/'+Path(leaf).name]=subprocess.check_output(['git','show',CODE+':'+leaf],cwd=ROOT)
names=sorted(set(selected)|set(inline))
if len(names)>32768:raise ValueError('fixed file count')
archive=EXT/'deadline-alignment-raw.tar.gz';pins={}
with archive.open('xb') as output:
 with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
  with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
   for n in names:
    raw=inline[n] if n in inline else selected[n].read_bytes();pins[n]=dict(bytes=len(raw),sha256=sha(raw));member=tarfile.TarInfo('glm-header-alignment/'+n);member.size=len(raw);member.mode=0o600;tar.addfile(member,io.BytesIO(raw))
if archive.stat().st_size>32<<20 or sum(v['bytes'] for v in pins.values())>128<<20:raise ValueError('fixed snapshot bound')
for n,p in selected.items():
 if sha(p.read_bytes())!=pins[n]['sha256']:raise ValueError('original changed during snapshot')
raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(raw,mtime=0);(BASE/'inventory.json.gz').write_bytes(encoded)
sraw=selected['tdd/source-green-headers.json'].read_bytes();sencoded=gzip.compress(sraw,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(sencoded)
capsule=dict(id='controlledGLMHeader120',sourceCommit=SOURCE,archive=dict(file=archive.relative_to(COMMON).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix='glm-header-alignment'),inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(sencoded),decodedBytes=len(sraw),decodedSHA256=sha(sraw),archiveMember='tdd/source-green-headers.json'),observe=[n for n in names if len(n.split('/'))==2 or n.endswith(('/manifest.json','/audit.json','/timing.json','/result.json'))])
dump(BASE/'index.json',dict(schema=1,kind='local-tls-glm-header-policy-alignment',codeCommit=CODE,boundedHelperSHA256='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f',capsule=capsule,originalsUnchangedAfterSnapshot=True,retentionCredentialsRead=False,retentionPaidCalls=0))
for f in ('selected-source-impact.json','setup-copy-note.json'):(BASE/f).write_bytes((A/f).read_bytes())
archive.chmod(0o444)
print(json.dumps(dict(files=len(names),rawBytes=sum(v['bytes'] for v in pins.values()),archiveBytes=archive.stat().st_size,archiveSHA256=capsule['archive']['sha256'],observe=len(capsule['observe'])),indent=2))
