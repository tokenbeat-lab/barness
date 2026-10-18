"""Freeze only the completed ordinary subtree and three completed numeric receipts."""
import gzip,hashlib,io,json,os,stat,tarfile
from pathlib import Path
ROOT=Path('/Users/cyber/RestoX/barness');OWN=Path('/tmp/barness-minimax-routed-validation')
BASE=OWN/'.scratch/barness-ai-minimax/evidence/ticket-13/final-gate';EXT=ROOT/'.evidence/minimax-retained/ticket-13-final-gate'
A=ROOT/'.evidence/minimax-routed-final-compiled-1';ORD=A/'gate/offline';L=ROOT/'.evidence/minimax-routed-legacy-1'
sha=lambda raw:hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,ensure_ascii=False,separators=(',',':'))+'\n')
if BASE.exists() or EXT.exists():raise RuntimeError('fresh final gate retention roots required')
BASE.mkdir();EXT.mkdir();selected={}
for folder,dirs,files in os.walk(ORD,followlinks=False):
 for d in dirs:
  if (Path(folder)/d).is_symlink():raise RuntimeError('symlink ordinary source directory')
 for name in files:
  p=Path(folder)/name;selected['ordinary/'+p.relative_to(ORD).as_posix()]=p
for name,expected in [('vet',0),('vet-live',0),('ordinary',1)]:
 p=A/'gate-command-exits'/(name+'.json');record=json.loads(p.read_text())
 if record['state']!='COMPLETED' or record['exitCode']!=expected or record['credentialsRead'] is not False:raise RuntimeError('completed actual numeric receipts required')
 selected['receipts/'+name+'.json']=p
selected['source/source-at-start.json']=L/'source-at-start.json'
if len(selected)>160000 or sum(p.stat().st_size for p in selected.values())>1600<<20:raise RuntimeError('fixed suite count or raw byte limit')
print(json.dumps({'phase':'archiving-completed-ordinary-only','files':len(selected),'rawBytes':sum(p.stat().st_size for p in selected.values())}),flush=True)
inventory={};archive=EXT/'ordinary-raw.tar.gz'
with archive.open('xb') as output:
 with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
  with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
   for count,(name,path) in enumerate(sorted(selected.items()),1):
    info=path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_size>64<<20:raise RuntimeError('bounded regular ordinary source required')
    with path.open('rb') as original:raw=original.read((64<<20)+1)
    if len(raw)!=info.st_size:raise RuntimeError('ordinary source changed')
    inventory[name]={'bytes':len(raw),'sha256':sha(raw)}
    item=tarfile.TarInfo('routed-compiled-ordinary/'+name);item.size=len(raw);item.mode=stat.S_IMODE(info.st_mode);tar.addfile(item,io.BytesIO(raw))
    if count%30000==0:print(json.dumps({'archivedFiles':count}),flush=True)
if archive.stat().st_size>256<<20:raise RuntimeError('fixed compressed suite limit')
for name,path in sorted(selected.items()):
 if sha(path.read_bytes())!=inventory[name]['sha256']:raise RuntimeError('original ordinary bytes changed after snapshot')
raw=(json.dumps(inventory,sort_keys=True,separators=(',',':'))+'\n').encode()
if len(raw)>40<<20:raise RuntimeError('fixed suite inventory limit')
encoded=gzip.compress(raw,mtime=0);(BASE/'ordinary-inventory.json.gz').write_bytes(encoded)
source=selected['source/source-at-start.json'].read_bytes();source_encoded=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(source_encoded)
session='20261018T084252.149580000Z';manifest='ordinary/'+session+'/manifest.json';report=json.loads(selected[manifest].read_text());nonpass=[c for c in report['cases'] if c['status']!='PASS']
observe=[manifest,'ordinary/'+session+'/audit.json','source/source-at-start.json',*[n for n in inventory if n.startswith('receipts/')]]
for row in nonpass:
 for name in row.get('artifacts',{}):
  if name.endswith('.json') and inventory['ordinary/'+session+'/'+row['dir']+'/'+name]['bytes']<=1<<20:observe.append('ordinary/'+session+'/'+row['dir']+'/'+name)
index={'schema':1,'kind':'completed-ordinary-original-failed','sourceCommit':'8fbd17006677774368d4d77d924a8672fd205c51','suite':'ordinary',
 'limits':{'compressedBytes':256<<20,'decodedTarBytes':2<<30,'files':160000,'entries':200000,'memberBytes':64<<20,'observedMetadataBytes':16<<20,'observedTotalBytes':64<<20,'inventoryDecodedBytes':40<<20,'inventoryCompressedBytes':8<<20,'rawMemberBytes':1600<<20},
 'archive':{'file':archive.relative_to(ROOT).as_posix(),'bytes':archive.stat().st_size,'sha256':sha(archive.read_bytes()),'prefix':'routed-compiled-ordinary'},
 'inventory':{'file':'ordinary-inventory.json.gz','compressedSHA256':sha(encoded),'decodedBytes':len(raw),'decodedSHA256':sha(raw),'files':len(inventory)},
 'sourceSnapshot':{'file':'source-at-start.json.gz','compressedSHA256':sha(source_encoded),'decodedBytes':len(source),'decodedSHA256':sha(source),'archiveMember':'source/source-at-start.json'},
 'observe':sorted(observe),'mainManifest':manifest,'originalsUnchangedAfterSnapshot':True,'retentionCredentialsRead':False,'retentionPaidCalls':0,
 'ordinaryExit':1,'gateVerdictCaptured':False,'raceCaptured':False}
dump(BASE/'ordinary-index.json',index)
archive.chmod(0o444)
print(json.dumps({'phase':'ordinary-snapshot-frozen','files':len(inventory),'rawBytes':sum(v['bytes'] for v in inventory.values()),'archiveBytes':archive.stat().st_size,'archiveSHA256':index['archive']['sha256'],'inventoryDecodedBytes':len(raw),'observe':len(observe)},indent=2))
