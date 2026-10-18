"""Read-only bounded snapshot of gap2 offline TDD and committed six checks."""
import gzip,hashlib,io,json,os,stat,subprocess,sys,tarfile
from pathlib import Path
sys.dont_write_bytecode=True
ROOT=Path('/tmp/barness-minimax-routed-validation').resolve();COMMON=Path('/Users/cyber/RestoX/barness')
BASE=ROOT/'.scratch/barness-ai-minimax/evidence/ticket-14/gap2';EXT=COMMON/'.evidence/minimax-retained/ticket-14-offline-gap2'
TDD=ROOT/'.evidence/minimax-postgate-offline-gap-2';CHECK=ROOT/'.evidence/minimax-ticket14-gap2-check-only-1'
SOURCE='572570bbe4a9322fb02ed740bd2a10c9b8ce792f';CODE='1a0ce3458d5c499b95c03e3994f3e7be35e88ea2'
sha=lambda raw:hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,ensure_ascii=False,separators=(',',':'))+'\n')
if BASE.exists() or EXT.exists():raise RuntimeError('fresh bounded roots required')
BASE.mkdir(parents=True);EXT.mkdir();selected={};inline={};links={};references=[]
def add(path,name):
 if name in selected or name in inline:raise RuntimeError('duplicate original member')
 if path.is_symlink():
  if not name.startswith('preflight/no-key-live/') or not name.endswith('/formal-corrupt-symlink/formal-integrity.json'):raise RuntimeError('only original deliberate negative fixture link permitted')
  target=os.readlink(path)
  if len(target.encode())>1024:raise RuntimeError('bounded negative-fixture link')
  links[name]={'target':target,'sha256':sha(target.encode()),'bytes':len(target.encode()),'original':path};return
 if not path.is_file() or path.stat().st_size>64<<20:raise RuntimeError('bounded regular original required')
 selected[name]=path

def tree(path,prefix,exclude=()):
 for folder,dirs,files in os.walk(path,followlinks=False):
  for d in dirs:
   if (Path(folder)/d).is_symlink():raise RuntimeError('symlink original directory')
  for name in files:
   p=Path(folder)/name
   if p.parent==path and name in exclude:references.append(dict(location='$REPO/'+p.relative_to(ROOT).as_posix(),bytes=p.stat().st_size,sha256=sha(p.read_bytes()),copied=False,reason='unused prepared paid binary; exact manifest and build receipt retained'))
   else:add(p,prefix+'/'+p.relative_to(path).as_posix())
tree(TDD,'tdd');tree(CHECK,'preflight',('live.test','chineseeval','visualeval'))
visual=[]
for p in (ROOT/'.evidence/barness-ai').glob('*/manifest.json'):
 if p.stat().st_size>16<<20:raise RuntimeError('bounded visual manifest')
 m=json.loads(p.read_text())
 if m['git_commit']==SOURCE and len(m['cases'])==16 and all(c['id'].startswith('VQ12-') for c in m['cases']):visual.append(p.parent)
if len(visual)!=1:raise RuntimeError('exact committed visual proxy bundle required')
tree(visual[0],'actual-visual-proxy/'+visual[0].name)
files=['ai/internal/testkit/release/artifacts_test.go','ai/internal/testkit/release/decisions_record_test.go','ai/e2e/release_recorded_bundle_test.go','ai/e2e/release_recorded_bundle_process_test.go','.scratch/barness-ai-minimax/tools/routed_validation.py','.scratch/barness-ai-minimax/tools/test_routed_validation.py']
for name in files:
 for revision,label in [(CODE,'after'),('1dedc575228c1691bed0843ddb54b84c379c5727','before')]:
  if label=='before' and name.endswith('release_recorded_bundle_process_test.go'):continue
  inline['source/'+label+'/'+name]=subprocess.check_output(['git','show',revision+':'+name],cwd=ROOT)
for file in ('barness-postgate-gap2-red.py','barness-postgate-gap2-tests.py','barness-postgate-gap2-actual-preflight.py'):add(Path('/tmp')/file,'controllers/'+file)
add(Path('/tmp/barness-routed-postgate-gap-2.json'),'diagnosis/postgate-gap-2.json')
projection={n:{k:v for k,v in r.items() if k!='original'} for n,r in links.items()}
inline['negative-fixture-links.json']=(json.dumps({'schema':1,'type':'symlink-metadata-only-no-follow-no-extract','links':projection},sort_keys=True,separators=(',',':'))+'\n').encode()
names=sorted(set(selected)|set(inline))
if len(names)>90000:raise RuntimeError('fixed file cap')
print(json.dumps({'phase':'gap2-snapshot','files':len(names)}),flush=True)
archive=EXT/'gap2-raw.tar.gz';inventory={}
with archive.open('xb') as output:
 with gzip.GzipFile(fileobj=output,mode='wb',compresslevel=6,mtime=0) as zipped:
  with tarfile.open(fileobj=zipped,mode='w|',format=tarfile.PAX_FORMAT) as tar:
   for count,n in enumerate(names,1):
    raw=inline[n] if n in inline else selected[n].read_bytes()
    if len(raw)>64<<20:raise RuntimeError('fixed member cap')
    inventory[n]={'bytes':len(raw),'sha256':sha(raw)}
    member=tarfile.TarInfo('offline-gap2/'+n);member.mode=0o600;member.size=len(raw);tar.addfile(member,io.BytesIO(raw))
    if count%20000==0:print(json.dumps({'archivedFiles':count}),flush=True)
if archive.stat().st_size>192<<20 or sum(p['bytes'] for p in inventory.values())>900<<20:raise RuntimeError('fixed snapshot cap')
for n,p in selected.items():
 if sha(p.read_bytes())!=inventory[n]['sha256']:raise RuntimeError('original mutated after snapshot')
for r in links.values():
 if sha(os.readlink(r['original']).encode())!=r['sha256']:raise RuntimeError('original link target mutated')
raw=(json.dumps(inventory,sort_keys=True,separators=(',',':'))+'\n').encode()
if len(raw)>24<<20:raise RuntimeError('fixed inventory cap')
encoded=gzip.compress(raw,mtime=0);(BASE/'inventory.json.gz').write_bytes(encoded)
sraw=selected['preflight/source-at-start.json'].read_bytes();sencoded=gzip.compress(sraw,mtime=0);(BASE/'source-at-start.json.gz').write_bytes(sencoded)
observe=[]
for n,pin in inventory.items():
 parts=n.split('/')
 if pin['bytes']<=1<<20 and (n=='negative-fixture-links.json' or n.startswith(('tdd/','source/','controllers/','diagnosis/')) or parts[0]=='preflight' and len(parts)==2 or
    n.startswith(('preflight/no-key-live/','preflight/no-key-release-contracts/')) and len(parts)==4 and parts[-1] in ('manifest.json','audit.json') or parts[0]=='actual-visual-proxy' and len(parts)==3 and parts[-1] in ('manifest.json','audit.json')):observe.append(n)
for n in inventory:
 if n.startswith('preflight/no-key-live/') and len(n.split('/'))==4 and n.endswith('/manifest.json') and n not in observe:observe.append(n)
if len(observe)>256 or sum(inventory[n]['bytes'] for n in observe)>16<<20:raise RuntimeError('fixed observation cap')
index=dict(schema=1,kind='offline-gap2-fixture-diagnostic-and-six-check-repair',codeCommit=CODE,sourceCommit=SOURCE,boundedHelperSHA256='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f',limits={'compressedBytes':192<<20,'decodedTarBytes':1<<30,'files':90000,'entries':100000,'memberBytes':64<<20,'observedMetadataBytes':8<<20,'inventoryDecodedBytes':24<<20,'inventoryCompressedBytes':8<<20},archive=dict(file=archive.relative_to(COMMON).as_posix(),bytes=archive.stat().st_size,sha256=sha(archive.read_bytes()),prefix='offline-gap2'),inventory=dict(file='inventory.json.gz',compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw),files=len(inventory)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(sencoded),decodedBytes=len(sraw),decodedSHA256=sha(sraw),archiveMember='preflight/source-at-start.json'),negativeFixtureLinkMetadata=projection,observe=sorted(observe),visualProxyRoot='actual-visual-proxy/'+visual[0].name,originalsUnchangedAfterSnapshot=True,retentionCredentialsRead=False,retentionPaidCalls=0)
dump(BASE/'index.json',index);dump(BASE/'external-references.json',dict(schema=1,codeCommit=CODE,checkedSourceCommit=SOURCE,binaries=references,retentionCredentialsRead=False,retentionPaidCalls=0))
for leaf in ('actual-six-check-observed.json','race-timeout-annotation.json'): (BASE/leaf).write_bytes((TDD/leaf).read_bytes())
archive.chmod(0o444)
print(json.dumps({'phase':'gap2-raw-frozen','files':len(inventory),'rawBytes':sum(p['bytes'] for p in inventory.values()),'archiveBytes':archive.stat().st_size,'archiveSHA256':index['archive']['sha256'],'observe':len(observe)},indent=2))
