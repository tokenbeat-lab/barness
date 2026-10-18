"""Read-only finite ticket14 history snapshot; never executes captured code."""
import gzip,hashlib,io,json,os,stat,subprocess,sys,tarfile
from pathlib import Path
sys.dont_write_bytecode=True
ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness')
BASE=ROOT/'.scratch/barness-ai-minimax/evidence/ticket-14/repair';EXT=COMMON/'.evidence/minimax-retained/ticket-14-release-mapping-pin'
TDD=ROOT/'.evidence/minimax-mapping-pin-repair-1';SOURCE='c2604db6f693f5c7d3a82c656512004ec85d2249'
sha=lambda raw:hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,ensure_ascii=False,separators=(',',':'))+'\n')
if BASE.exists() or EXT.exists():raise RuntimeError('fresh repair retention roots required')
BASE.mkdir(parents=True);EXT.mkdir();selected={};inline={};links={};references=[]
def add(path,name):
 if name in selected or name in inline:raise RuntimeError('duplicate original member')
 if path.is_symlink():
  if not name.startswith(('preflight-1/no-key-live/','preflight-2/no-key-live/')) or not name.endswith('/formal-corrupt-symlink/formal-integrity.json'):raise RuntimeError('only original negative-fixture symlink allowed')
  target=os.readlink(path)
  if len(target.encode())>1024:raise RuntimeError('bounded negative-fixture symlink target')
  links[name]={'target':target,'sha256':sha(target.encode()),'bytes':len(target.encode()),'original':path};return
 if not path.is_file() or path.stat().st_size>64<<20:raise RuntimeError('bounded regular original required')
 selected[name]=path

def tree(path,prefix,exclude=()):
 for folder,dirs,files in os.walk(path,followlinks=False):
  for d in dirs:
   if (Path(folder)/d).is_symlink():raise RuntimeError('symlink original directory')
  for name in files:
   p=Path(folder)/name
   if p.parent==path and name in exclude:
    references.append(dict(location='$REPO/'+p.relative_to(ROOT).as_posix(),bytes=p.stat().st_size,sha256=sha(p.read_bytes()),copied=False,reason='unused prepared paid binaries; immutable binary manifest and build receipts retained'))
   else:add(p,prefix+'/'+p.relative_to(path).as_posix())
tree(TDD,'tdd')
for n in (1,2):tree(ROOT/'.evidence'/('minimax-ticket14-check-only-'+str(n)),'preflight-'+str(n),('live.test','chineseeval','visualeval'))
visual=[]
for p in (ROOT/'.evidence/barness-ai').glob('*/manifest.json'):
 if p.stat().st_size>16<<20:raise RuntimeError('bounded default manifest')
 m=json.loads(p.read_text())
 if m['git_commit']==SOURCE and len(m['cases'])==16 and all(c['id'].startswith('VQ12-') for c in m['cases']):visual.append(p.parent)
if len(visual)!=1:raise RuntimeError('exact fresh public visual proxy bundle required')
tree(visual[0],'actual-visual-proxy/'+visual[0].name)
copy=json.loads((TDD/'offline-fixture-restoration.json').read_text())
for row in copy['files']:add(ROOT/row['archive'],'fixtures/'+row['archive'].removeprefix('.evidence/'))
for name in ('glm-candidate','glm-formal'):add(ROOT/'ai/live/testdata'/name/'live-captures.json','fixtures/'+name+'-selector.json')
for revision,leaf,name in [('a3a9097','ai/internal/testkit/release/plan.go','source/plan-before.go'),(SOURCE,'ai/internal/testkit/release/plan.go','source/plan-after.go'),
 ('7111641','.scratch/barness-ai-minimax/tools/routed_validation.py','source/routed-before.py'),(SOURCE,'.scratch/barness-ai-minimax/tools/routed_validation.py','source/routed-after.py'),
 (SOURCE,'.scratch/barness-ai-minimax/tools/test_routed_validation.py','source/test-routed-after.py')]:
 inline[name]=subprocess.check_output(['git','show',revision+':'+leaf],cwd=ROOT,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
for file in ('barness-ticket14-red.py','barness-ticket14-keyless-check.py','barness-ticket14-actual-preflight.py','barness-ticket14-actual-preflight-restored.py'):add(Path('/tmp')/file,'controllers/'+file)
add(Path('/tmp/barness-routed-release-plan-gap.json'),'diagnosis/release-plan-gap.json')
link_projection={name:{k:v for k,v in record.items() if k!='original'} for name,record in links.items()}
inline['negative-fixture-links.json']=(json.dumps({'schema':1,'type':'symlink-metadata-only-no-follow-no-extract','links':link_projection},sort_keys=True,separators=(',',':'))+'\n').encode()
all_names=sorted(set(selected)|set(inline))
if len(all_names)>90000:raise RuntimeError('fixed repair file limit')
print(json.dumps({'phase':'archiving-repair','files':len(all_names)}),flush=True)
inventory={};archive=EXT/'ticket14-repair-raw.tar.gz'
with archive.open('xb') as output:
 with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
  with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
   for count,name in enumerate(all_names,1):
    data=inline[name] if name in inline else selected[name].read_bytes()
    if len(data)>64<<20:raise RuntimeError('fixed repair member limit')
    inventory[name]={'bytes':len(data),'sha256':sha(data)}
    member=tarfile.TarInfo('release-mapping-repair/'+name);member.mode=0o600;member.size=len(data);tar.addfile(member,io.BytesIO(data))
    if count%20000==0:print(json.dumps({'archivedFiles':count}),flush=True)
if archive.stat().st_size>192<<20 or sum(p['bytes'] for p in inventory.values())>900<<20:raise RuntimeError('fixed repair byte bound')
for name,path in selected.items():
 raw=path.read_bytes()
 if sha(raw)!=inventory[name]['sha256']:raise RuntimeError('original repair byte changed after snapshot')
for record in links.values():
 if sha(os.readlink(record['original']).encode())!=record['sha256']:raise RuntimeError('original negative-fixture target changed')
raw=(json.dumps(inventory,sort_keys=True,separators=(',',':'))+'\n').encode()
if len(raw)>24<<20:raise RuntimeError('fixed inventory decoded limit')
encoded=gzip.compress(raw,mtime=0);(BASE/'inventory.json.gz').write_bytes(encoded)
source=selected['preflight-2/source-at-start.json'].read_bytes();source_gz=gzip.compress(source,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(source_gz)
observe=[]
for name,pin in inventory.items():
 parts=name.split('/')
 if pin['bytes']<=1<<20 and (name=='negative-fixture-links.json' or name.startswith(('tdd/','source/','controllers/','diagnosis/','fixtures/')) or
    parts[0].startswith('preflight-') and len(parts)==2 or
    parts[0].startswith('preflight-') and len(parts)==4 and parts[1] in ('no-key-live','no-key-release-contracts') and parts[-1] in ('manifest.json','audit.json') or
    parts[0]=='actual-visual-proxy' and len(parts)==3 and parts[-1] in ('manifest.json','audit.json')):observe.append(name)
for name in inventory:
 if name.startswith(('preflight-1/no-key-live/','preflight-2/no-key-live/')) and len(name.split('/'))==4 and name.endswith('/manifest.json') and name not in observe:observe.append(name)
if sum(inventory[n]['bytes'] for n in observe)>16<<20:raise RuntimeError('fixed observed total byte limit')
index={'schema':1,'kind':'offline-mapping-pin-and-routed-preflight-repair','codeCommit':SOURCE,
 'boundedHelperSHA256':'b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f',
 'limits':{'compressedBytes':192<<20,'decodedTarBytes':1<<30,'files':90000,'entries':100000,'memberBytes':64<<20,'observedMetadataBytes':8<<20,'inventoryDecodedBytes':24<<20,'inventoryCompressedBytes':8<<20},
 'archive':{'file':archive.relative_to(COMMON).as_posix(),'bytes':archive.stat().st_size,'sha256':sha(archive.read_bytes()),'prefix':'release-mapping-repair'},
 'inventory':{'file':'inventory.json.gz','compressedSHA256':sha(encoded),'decodedBytes':len(raw),'decodedSHA256':sha(raw),'files':len(inventory)},
 'sourceCommit':SOURCE,'sourceSnapshot':{'file':'source-at-start.json.gz','compressedSHA256':sha(source_gz),'decodedBytes':len(source),'decodedSHA256':sha(source),'archiveMember':'preflight-2/source-at-start.json'},
 'negativeFixtureLinkMetadata':link_projection,'observe':sorted(observe),'visualProxyRoot':'actual-visual-proxy/'+visual[0].name,'originalsUnchangedAfterSnapshot':True,'retentionCredentialsRead':False,'retentionPaidCalls':0}
dump(BASE/'index.json',index);dump(BASE/'external-references.json',dict(schema=1,codeCommit=SOURCE,binaries=references,unusedPaidPreparedBinaries=True,retentionCredentialsRead=False,retentionPaidCalls=0))
for leaf in ('mapping-delta-proof.json','offline-fixture-restoration.json','actual-five-check-observed-2.json'): (BASE/leaf).write_bytes((TDD/leaf).read_bytes())
archive.chmod(0o444)
print(json.dumps({'phase':'repair-raw-frozen','files':len(inventory),'rawBytes':sum(v['bytes'] for v in inventory.values()),'archiveBytes':archive.stat().st_size,'archiveSHA256':index['archive']['sha256'],'inventoryDecodedBytes':len(raw),'observe':len(observe),'observedBytes':sum(inventory[n]['bytes'] for n in observe)},indent=2))
