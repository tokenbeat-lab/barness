import collections,gzip,hashlib,json,os,pathlib,shutil,stat,tarfile
R=pathlib.Path('/tmp/barness-minimax-10-global')
A=pathlib.Path('/tmp/barness-minimax-postactivation-red')
COMMON=pathlib.Path('/Users/cyber/RestoX/barness/.evidence/minimax-retained')
DEST=R/'.scratch/barness-ai-minimax/evidence/ticket-10/international-postactivation'
SOURCE='de051931b8d4e30fb6adf5cc5b2a3672f243ba35'
COMMON.mkdir(parents=True,exist_ok=True);DEST.mkdir(parents=True,exist_ok=True)
def sha(raw):return hashlib.sha256(raw).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,ensure_ascii=False,indent=2)+'\n')
def raw_file(p):
 fd=os.open(p,os.O_RDONLY|os.O_NOFOLLOW)
 with os.fdopen(fd,'rb') as f:
  st=os.fstat(f.fileno())
  if not stat.S_ISREG(st.st_mode) or st.st_size>64<<20:raise RuntimeError('unsafe or large file')
  raw=f.read((64<<20)+1)
  if len(raw)>64<<20 or len(raw)!=st.st_size:raise RuntimeError('grown or changed file')
  return raw

def collect(base):
 files=[];stack=[(base,0)];entries=0;total=0
 while stack:
  folder,depth=stack.pop()
  if depth>24:raise RuntimeError('depth bound')
  with os.scandir(folder) as it:
   for entry in it:
    entries+=1
    if entries>400000:raise RuntimeError('entry bound')
    p=pathlib.Path(entry.path);st=entry.stat(follow_symlinks=False)
    if stat.S_ISDIR(st.st_mode):stack.append((p,depth+1));continue
    if not (stat.S_ISREG(st.st_mode) or stat.S_ISLNK(st.st_mode)) or st.st_size>64<<20:raise RuntimeError('unsafe or large file')
    total+=st.st_size
    if total>4<<30:raise RuntimeError('total bound')
    files.append(p)
 return sorted(files)

def archive(name,groups):
 target=COMMON/(name+'.tar.gz');partial=target.with_suffix('.partial')
 if target.exists() or partial.exists():raise RuntimeError('fresh archive required')
 paths={};inventory={};results=[]
 for label,base in groups:
  files=collect(base)
  for p in files:
   rel=p.relative_to(base).as_posix();key=label+'/'+rel
   if key in paths or rel.startswith('/') or '..' in pathlib.PurePosixPath(rel).parts:raise RuntimeError('unsafe duplicate name')
   paths[key]=p
   if p.is_symlink():
    target_text=os.readlink(p);data=os.fsencode(target_text)
    if len(data)>4096:raise RuntimeError('link text bound')
    inventory[key]={'bytes':len(data),'sha256':sha(data),'kind':'intentional-negative-symlink','target':target_text}
   else:
    data=raw_file(p);inventory[key]={'bytes':len(data),'sha256':sha(data)}
  manifests=[]
  for p in files:
   rel=p.relative_to(base)
   if p.name=='manifest.json' and len(rel.parts)<=3:
    d=json.loads(raw_file(p));manifests.append((len(d.get('cases',[])),p,d))
  if manifests:
   largest=max(x[0] for x in manifests)
   for count,p,d in manifests:
    if count!=largest:continue
    counts=collections.Counter(x['status'] for x in d.get('cases',[]));refs=0
    for case in d.get('cases',[]):
     for n,h in case.get('artifacts',{}).items():
      key=(pathlib.PurePosixPath(label)/p.parent.relative_to(base).as_posix()/case['dir']/n).as_posix()
      if key not in inventory or inventory[key]['sha256']!=h:raise RuntimeError('actual Artifact byte mismatch '+key)
      refs+=1
    audit=p.parent/'audit.json';findings=None
    if audit.is_file():findings=len(json.loads(raw_file(audit))['findings'])
    results.append({'group':label,'manifest':label+'/'+p.relative_to(base).as_posix(),'caseOutcomes':dict(counts),'artifactReferences':refs,'auditFindings':findings,'executionGitCommit':d.get('git_commit')})
 with partial.open('xb') as raw_tar, tarfile.open(fileobj=raw_tar,mode='w:gz',compresslevel=6) as tar:
  for key,p in paths.items():
   info=tarfile.TarInfo(key);info.mode=0o600
   if inventory[key].get('kind')=='intentional-negative-symlink':
    info.type=tarfile.SYMTYPE;info.linkname=inventory[key]['target'];tar.addfile(info)
   else:
    info.size=inventory[key]['bytes']
    with p.open('rb') as f:tar.addfile(info,f)
 seen=set()
 with tarfile.open(partial,'r|gz') as tar:
  for member in tar:
   if member.name not in inventory or member.name in seen:raise RuntimeError('archive member mismatch')
   if member.issym():
    expected=inventory[member.name]
    if expected.get('kind')!='intentional-negative-symlink' or member.linkname!=expected['target'] or sha(os.fsencode(member.linkname))!=expected['sha256']:raise RuntimeError('link metadata mismatch')
    seen.add(member.name);continue
   if not member.isfile() or inventory[member.name].get('kind'):raise RuntimeError('archive member mismatch')
   f=tar.extractfile(member);h=hashlib.sha256();count=0
   while True:
    b=f.read(1<<20)
    if not b:break
    count+=len(b);h.update(b)
    if count>64<<20:raise RuntimeError('archive actual read bound')
   if inventory[member.name]!={'bytes':count,'sha256':h.hexdigest()}:raise RuntimeError('archive bytes mismatch')
   seen.add(member.name)
 if seen!=set(inventory):raise RuntimeError('archive omitted bytes')
 os.link(partial,target);partial.unlink()
 digest=hashlib.sha256()
 with target.open('rb') as f:
  while b:=f.read(1<<20):digest.update(b)
 index_name=name+'-files.json.gz'
 with gzip.open(DEST/index_name,'wt',encoding='utf-8') as f:json.dump(inventory,f,ensure_ascii=False,sort_keys=True,separators=(',',':'))
 small=target.stat().st_size<=8<<20
 if small:shutil.copyfile(target,DEST/target.name)
 index={'schema':1,'name':name,'fixCheckpoint':SOURCE,'credentialsRead':False,'vendorCalls':0,'actualSupport':False,
 'archive':target.name if small else '$COMMON/.evidence/minimax-retained/'+target.name,'commonRetention':'$COMMON/.evidence/minimax-retained/'+target.name,
 'archiveBytes':target.stat().st_size,'archiveSha256':digest.hexdigest(),'files':sum(not x.get('kind') for x in inventory.values()),'negativeSymlinkFixtures':sum(bool(x.get('kind')) for x in inventory.values()),'entries':len(inventory),'expandedBytes':sum(x['bytes'] for x in inventory.values()),'inventory':index_name,'inventorySha256':sha((DEST/index_name).read_bytes()),'byteVerified':True,'results':results}
 dump(DEST/(name+'.json'),index);print(json.dumps({'name':name,'archiveBytes':index['archiveBytes'],'files':index['files'],'results':results}),flush=True)
 return index
if __name__=='__main__':
 import sys
 groups={
 'red':[('initial-overlay',R/'.evidence/mm10-postactivation-red'),('complete-activation-live',A/'.evidence/minimax-postactivation-complete-red'),('activation-e2e',A/'.evidence/minimax-postactivation-e2e-red'),('snapshot-boundary',R/'.evidence/minimax-postactivation-snapshot-red')],
 'business-red':[('business-red',A/'.evidence/minimax-postactivation-business-red')],
 'business-active':[('business-green',A/'.evidence/minimax-postactivation-business-green'),('business-race',A/'.evidence/minimax-postactivation-business-race')],
 'business-inactive':[('business-inactive',R/'.evidence/minimax-postactivation-business-inactive-df342b9')],
 'focused':[('focused-live-normal',R/'.evidence/minimax-postactivation-focused-green'),('inactive-affected-e2e',R/'.evidence/minimax-postactivation-inactive-green'),('retained-checks',R/'.evidence/minimax-postactivation-retained-checks')],
 'ordinary':[('ordinary',R/'.evidence/minimax-postactivation-mm-complete-de05193/ordinary')],
 'race':[('race',R/'.evidence/minimax-postactivation-mm-complete-de05193/race')],
 'activated-live':[('activated-live',A/'.evidence/minimax-postactivation-complete-green')],
 'activated-whole':[('activated-whole',A/'.evidence/minimax-postactivation-whole-ordinary-de05193')],
 }
 for mode in sys.argv[1:]:
  SOURCE='df342b9' if mode.startswith('business-') else 'de051931b8d4e30fb6adf5cc5b2a3672f243ba35'
  archive('preactivation-'+mode+('-df342b9' if mode.startswith('business-') else '-de05193'),groups[mode])
