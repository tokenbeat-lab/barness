#!/usr/bin/env python3
"""Actual public constructor at real OS opens, with only private FS mutations."""
import argparse,gzip,hashlib,json,os,runpy,shutil,subprocess,sys,time
from pathlib import Path
BASE=Path(__file__).resolve().parent;FRONT=BASE.parents[1];REPO=BASE.parents[5];REL=Path('.scratch/barness-ai-minimax/evidence/ticket-17');sha=lambda v:hashlib.sha256(v).hexdigest()
RESOURCE='stages/green-final-native-shapes/bundles/20261018T123017.584602000Z/MM04-final-block-field-text-missing/resources.json'
def fingerprint(b,root,name,limit):
 with b.opened(root,name) as stream:
  before=os.fstat(stream.fileno());digest=hashlib.sha256();count=0
  if before.st_size>limit:raise ValueError('fingerprint size bound')
  for chunk in iter(lambda:stream.read(1<<20),b''):
   count+=len(chunk)
   if count>limit:raise ValueError('fingerprint byte bound')
   digest.update(chunk)
  after=os.fstat(stream.fileno())
  if (before.st_size,before.st_mtime_ns,before.st_ctime_ns)!=(after.st_size,after.st_mtime_ns,after.st_ctime_ns):raise ValueError('original changed during fingerprint')
 return dict(mode=before.st_mode,bytes=count,sha256=digest.hexdigest())
def child(a):
 producer=Path(a.producer);output=Path(a.output);front=output/'clone'/REL;index=front/'index.json';inventory=front/'inventory.json.gz';summary=front/'summary.json';destination=output/'common/.evidence/minimax-retained/ticket-17-final-payload/offline-tdd-raw-v2.tar.gz';event_file=Path(a.events);injected=False
 def emit(v):
  with event_file.open('a') as f:f.write(json.dumps(v)+'\n')
 def replace(path,data):path.unlink();path.write_bytes(data)
 def audit(event,args):
  nonlocal injected
  if event!='open' or not isinstance(args[0],(str,bytes)):return
  path=Path(os.fsdecode(args[0]));flags=args[2]
  if path==event_file or flags&(os.O_WRONLY|os.O_RDWR):return
  if injected:
   if path.name=='inventory.json.gz':emit(dict(event='later-inventory-open',path=str(path)))
   return
  target=summary if a.case=='seal-metadata' else destination if a.case=='output-grow' else index
  if path.name!=target.name or not target.exists():return
  injected=True
  if a.case=='index-size':
   original=target.read_bytes();data=original+b' '*((1<<20)+1-len(original));replace(target,data);changed=[dict(file='index.json',bytes=len(data),sha256=sha(data),recipe='original JSON + trailing spaces to1MiB+1')]
  elif a.case in ('inventory-compressed','inventory-decoded'):
   value=json.loads(index.read_bytes());encoded=inventory.read_bytes();raw=gzip.decompress(encoded)
   if a.case=='inventory-compressed':encoded+=b'\0'*((8<<20)+1-len(encoded))
   else:raw+=b' '*((40<<20)+1-len(raw));encoded=gzip.compress(raw,mtime=0)
   value['inventory'].update(compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw));replace(inventory,encoded);replace(index,(json.dumps(value)+'\n').encode());changed=[dict(file='inventory.json.gz',bytes=len(encoded),sha256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw),recipe='original gzip + zero padding to8MiB+1' if a.case=='inventory-compressed' else 'original decoded JSON + spaces to40MiB+1, gzip mtime0')]
  elif a.case=='seal-metadata':
   original=target.read_bytes();data=original+b' '*((8<<20)+1-len(original));replace(target,data);changed=[dict(file='summary.json',bytes=len(data),sha256=sha(data),recipe='original JSON + trailing spaces to8MiB+1')]
  else:
   target.unlink()
   with target.open('xb') as f:f.truncate(257<<20)
   changed=[dict(file=str(target.relative_to(output)),bytes=257<<20,recipe='fresh sparse all-zero regular file')]
  emit(dict(event='injected',case=a.case,path=str(target),openMode=args[1],flags=flags,changes=changed))
 sys.addaudithook(audit);sys.argv=[str(producer),'--common-root',a.common_root,'--value','false','--out',str(output)];runpy.run_path(str(producer),run_name='__main__')
def main(a):
 out=Path(a.out).resolve();out.mkdir(parents=True,exist_ok=False);clone=out/'source';front=clone/REL;shutil.copytree(FRONT,front,ignore=shutil.ignore_patterns('retention-fixes','__pycache__'));producer=front/'retention-fixes/resource-probe.py';producer.parent.mkdir();shutil.copyfile(FRONT/'retention-fixes/resource-probe.py',producer)
 for rel in ('implementation/verify.py','final-gate/capsule.py'):
  target=front.parent/'ticket-13'/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(FRONT.parent/'ticket-13'/rel,target)
 sys.path.insert(0,str(FRONT));import readers
 b=readers.small();full=readers.full();index=b.doc(b.read(FRONT,'index.json',1<<20));view=dict(index,observe=[RESOURCE]);_,values,_=full.verify(b,FRONT,Path(a.common_root),view);resource=clone/'.evidence/minimax-ticket17/green-final-native-shapes/bundles/20261018T123017.584602000Z/MM04-final-block-field-text-missing/resources.json';resource.parent.mkdir(parents=True);resource.write_bytes(values[RESOURCE]);before=dict(index=sha(b.read(FRONT,'index.json',1<<20)),inventory=sha(b.read(FRONT,'inventory.json.gz',8<<20)),originalArchive=fingerprint(b,Path(a.common_root),index['archive']['file'],256<<20))
 output=out/'work';events=out/'events.jsonl';argv=[sys.executable,'-B',str(Path(__file__).resolve()),'--child','--case',a.case,'--producer',str(producer),'--output',str(output),'--events',str(events),'--common-root',a.common_root];start=time.monotonic()
 with (out/'stdout.log').open('xb') as stdout,(out/'stderr.log').open('xb') as stderr:run=subprocess.run(argv,stdout=stdout,stderr=stderr,timeout=240,start_new_session=True,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
 front_out=output/'clone'/REL;logs=[json.loads(v) for v in events.read_text().splitlines()] if events.exists() else [];after=dict(index=sha(b.read(FRONT,'index.json',1<<20)),inventory=sha(b.read(FRONT,'inventory.json.gz',8<<20)),originalArchive=fingerprint(b,Path(a.common_root),index['archive']['file'],256<<20));verification=next((p for p in (output/'red.json',output/'verification.json') if p.exists()),None);reached=(output/'construction.json').exists();initial_seal=sha(b.read(front,'SHA256SUMS',1<<20));actual_seal=sha(b.read(front_out,'SHA256SUMS',1<<20));diagnostic=(out/'stderr.log').read_text();row=dict(case=a.case,expectation=a.expect,argv=argv,exitCode=run.returncode,seconds=time.monotonic()-start,injected=any(v['event']=='injected' for v in logs),constructionPresent=reached,verificationInvoked=verification is not None,sealRewritten=initial_seal!=actual_seal,diagnostic=diagnostic,sourceSHA256=sha(b.read(producer.parent,producer.name,64<<10)),testSHA256=sha(Path(__file__).read_bytes()),originalBefore=before,originalAfter=after,originalsUnchanged=before==after,credentialsRead=False,paidCalls=0,releaseCertified=False)
 if verification is not None:row['publicVerification']=b.doc(b.read(output,verification.name,1<<20))
 destination=output/'common'/index['archive']['file']
 if destination.exists() and destination.stat().st_size<=256<<20:row['privateOutput']=fingerprint(b,output/'common',index['archive']['file'],256<<20)
 (out/'receipt.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row,indent=2),flush=True)
 if not row['injected'] or before!=after or run.returncode!=1:raise ValueError('genuine injected failure with immutable originals required')
 if a.expect=='unsafe':valid=reached and verification is not None
 else:
  reasons={'index-size':'artifact size limit','inventory-compressed':'artifact size limit','inventory-decoded':'decoded metadata pin or limit','seal-metadata':'artifact size limit','output-grow':'output archive size bound'};valid=verification is None and reasons[a.case] in diagnostic and (not row['sealRewritten'] if a.case=='seal-metadata' else not reached)
 if not valid:raise ValueError('public factory early/late boundary outcome differs')
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--case',choices=('index-size','inventory-compressed','inventory-decoded','seal-metadata','output-grow'),required=True);p.add_argument('--expect',choices=('unsafe','bounded'));p.add_argument('--out');p.add_argument('--common-root',default='/Users/cyber/RestoX/barness');p.add_argument('--child',action='store_true');p.add_argument('--producer');p.add_argument('--output');p.add_argument('--events');a=p.parse_args();child(a) if a.child else main(a)
