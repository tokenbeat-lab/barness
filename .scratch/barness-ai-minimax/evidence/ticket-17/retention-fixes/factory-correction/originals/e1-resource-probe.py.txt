#!/usr/bin/env python3
"""Private single-counter mutation; public verifier proves exact-integer behavior."""
import argparse,copy,gzip,hashlib,io,json,os,shutil,subprocess,sys,tarfile
from pathlib import Path
BASE=Path(__file__).resolve().parent;FRONT=BASE.parent;REPO=BASE.parents[4];sha=lambda v:hashlib.sha256(v).hexdigest()
def seal(base):
 (base/'SHA256SUMS').write_text(''.join(sha(p.read_bytes())+'  '+p.relative_to(base).as_posix()+'\n' for p in sorted(base.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
def invoke(base,common,out,expect):
 seal(base);argv=[sys.executable,'-B',str(base/'verify.py'),'--common-root',str(common)];run=subprocess.run(argv,capture_output=True,timeout=240,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ});out.with_suffix('.stdout').write_bytes(run.stdout);out.with_suffix('.stderr').write_bytes(run.stderr);row=dict(argv=argv,exitCode=run.returncode,stdoutSHA256=sha(run.stdout),stderrSHA256=sha(run.stderr),expectedExit=expect,credentialsRead=False,paidCalls=0);out.with_suffix('.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True)
 if run.returncode!=expect or (expect==1 and b'original public resource release differs' not in run.stderr):raise ValueError('public resource boundary outcome differs')
def main(a):
 out=Path(a.out).resolve()
 if a.recheck:
  base=out/'clone/.scratch/barness-ai-minimax/evidence/ticket-17';shutil.copyfile(FRONT/'verify.py',base/'verify.py');invoke(base,out/'common',out/'green',1);return
 out.mkdir(parents=True,exist_ok=False);base=out/'clone/.scratch/barness-ai-minimax/evidence/ticket-17';shutil.copytree(FRONT,base,ignore=shutil.ignore_patterns('retention-fixes','__pycache__'))
 for rel in ('implementation/verify.py','final-gate/capsule.py'):
  target=base.parent/'ticket-13'/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(FRONT.parent/'ticket-13'/rel,target)
 sys.path.insert(0,str(FRONT));import readers
 b=readers.small();full=readers.full();index=json.loads((base/'index.json').read_bytes());pins=json.loads(gzip.decompress((base/'inventory.json.gz').read_bytes()));original=copy.deepcopy(pins);common=out/'common';destination=common/index['archive']['file'];destination.parent.mkdir(parents=True)
 root=index['roots']['green-final-native-shapes'];resource=root+'/MM04-final-block-field-text-missing/resources.json';manifest=root+'/manifest.json';old=b.read(REPO,'.evidence/minimax-ticket17/green-final-native-shapes/bundles/20261018T123017.584602000Z/MM04-final-block-field-text-missing/resources.json',1<<20);value=json.loads(old);assert all(type(v) is int and v==0 for v in value.values());value['bodies']=False if a.value=='false' else 0.0;replacement=(json.dumps(value,separators=(',',':'))+'\n').encode();oldsha=pins[resource]['sha256'];newsha=sha(replacement);changed={};seen=set()
 with b.opened(Path(a.common_root),index['archive']['file']) as source:
  info=os.fstat(source.fileno());digest=hashlib.sha256();count=0
  if info.st_size>full.LIMITS['compressedBytes']:raise ValueError('input compressed cap')
  for chunk in iter(lambda:source.read(1<<20),b''):
   count+=len(chunk)
   if count>full.LIMITS['compressedBytes']:raise ValueError('input cumulative cap')
   digest.update(chunk)
  if count!=index['archive']['bytes'] or digest.hexdigest()!=index['archive']['sha256']:raise ValueError('original raw pin')
  source.seek(0)
  with gzip.GzipFile(fileobj=source) as zin,destination.open('xb') as target,gzip.GzipFile(fileobj=target,mode='wb',mtime=0,compresslevel=6) as zout,tarfile.open(fileobj=zout,mode='w|',format=tarfile.PAX_FORMAT) as output:
   limited=b.Limited(zin,full.LIMITS['decodedTarBytes'])
   with tarfile.open(fileobj=limited,mode='r|') as archive:
    for member in archive:
     name='/'.join(b.relative(member.name)[1:])
     if not member.isfile() or name in seen or name not in pins or member.size!=pins[name]['bytes'] or member.size>full.LIMITS['memberBytes']:raise ValueError('regular owned original member required')
     seen.add(name)
     with archive.extractfile(member) as stream:data=stream.read(full.LIMITS['memberBytes']+1)
     if len(data)!=member.size or sha(data)!=pins[name]['sha256']:raise ValueError('original raw member pin')
     if name==resource:data=replacement
     if name==manifest:
      before=json.loads(data);after=copy.deepcopy(before);case=next(v for v in after['cases'] if v['id']=='MM04-final-block-field-text-missing');assert case['artifacts']['resources.json']==oldsha;case['artifacts']['resources.json']=newsha;marker=data.index(b'"dir": "MM04-final-block-field-text-missing"');offset=data.rfind(oldsha.encode(),0,marker);assert offset>=0;data=data[:offset]+newsha.encode()+data[offset+64:];assert json.loads(data)==after
     if name in (resource,manifest):
      changed[name]=dict(before=pins[name],after=dict(bytes=len(data),sha256=sha(data)));pins[name]=changed[name]['after'];(out/('mutated-resource.json' if name==resource else 'mutated-manifest.json')).write_bytes(data)
     item=tarfile.TarInfo(member.name);item.mode=member.mode;item.size=len(data);output.addfile(item,io.BytesIO(data))
   while limited.read(65536):pass
 if seen!=set(pins) or set(changed)!={resource,manifest} or any(v!=original[n] for n,v in pins.items() if n not in changed):raise ValueError('only exact resource and artifact-digest delta allowed')
 raw=(json.dumps(pins,sort_keys=True,separators=(',',':'))+'\n').encode();encoded=gzip.compress(raw,mtime=0);(base/'inventory.json.gz').write_bytes(encoded);index['inventory'].update(compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw));size=destination.stat().st_size
 if size>full.LIMITS['compressedBytes']:raise ValueError('private output cap')
 with destination.open('rb') as f:digest=hashlib.file_digest(f,'sha256').hexdigest()
 index['archive'].update(bytes=size,sha256=digest);(base/'index.json').write_text(json.dumps(index)+'\n');destination.chmod(0o444);(out/'construction.json').write_text(json.dumps(dict(kind='NON_AUTHORITATIVE-negative-resource-fixture',originalArchiveSHA256='5d9b5d5ad032fdcba150d294cf65f0944f6219068ec810f49a86caa45d71ef8d',value=a.value,changes=changed,files=len(pins),unchangedOtherFiles=len(pins)-2,archive=index['archive'],limits=full.LIMITS,credentialsRead=False,paidCalls=0,releaseCertified=False),indent=2)+'\n');invoke(base,common,out/'red',0)
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',default='/Users/cyber/RestoX/barness');p.add_argument('--out',required=True);p.add_argument('--value',choices=('false','float'));p.add_argument('--recheck',action='store_true');main(p.parse_args())
