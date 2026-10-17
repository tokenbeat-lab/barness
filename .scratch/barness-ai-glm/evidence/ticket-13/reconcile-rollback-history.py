"""One-time read-only reconciliation of the baseline's selected release evidence."""
import collections,hashlib,json,os,subprocess
from pathlib import Path
ROOT=Path('/Users/cyber/RestoX/barness'); BASE='f5cf5362610a8bdb8f3e170f408f3858c149e540'
OUT=ROOT/'.evidence/barness-ai-glm/ticket-13/rollback-history'; OUT.mkdir(parents=True,exist_ok=False)
def digest(p):
 h=hashlib.sha256()
 with open(p,'rb') as f:
  while chunk:=f.read(1<<20):h.update(chunk)
 return h.hexdigest()
def show(p):return subprocess.check_output(['git','show',BASE+':'+p],cwd=ROOT)
def doc(p):return json.loads(p.read_bytes())
def dump(p,v):p.write_text(json.dumps(v,indent=2)+'\n')
indexpath='.scratch/jev-1.14-multimodal-box/14-remaining-real/evidence.json'
raw=show(indexpath); index=json.loads(raw);base=ROOT/index['base']; recorded=index['current']['report'];report=doc(base/'gate/release-report.json')
assert index['overall']=='PASS' and report==recorded and report['pass']
assert len(report['live']['rows'])==11 and all(r['status']=='PASS' for r in report['live']['rows'])
assert all(g['pass'] for g in report['gates']) and len(report['gates'])==10
assert len(report['audits'])==17 and all(not a['findings'] for a in report['audits'])
# Exactly the original selected live inputs, plus four effects and normal/race.
selected={}
for combo in [r['combo'] for r in report['live']['rows']]:
 if combo=='deepseek-chat': p=ROOT/doc(base/'accepted-deepseek-chat.json')['selectedBundle']
 elif combo=='typesafe-classifier-jev-1.14.0':p=base/'support'
 else:
  parent=base/('protocol113' if combo=='typesafe-classifier' else 'live/'+combo)
  dirs=list(parent.glob('*/manifest.json'));assert len(dirs)==1
  p=dirs[0].parent
 selected['live:'+combo]=p
for e in report['evidence']:
 if e['name'].startswith(('chinese-effect-','visual-effect-')) or e['name'] in ['race-offline']:
  assert e['status']=='PASS';selected[e['name']]=Path(e['bundle'])
selected['normal-offline']=Path(report['bundle'])
assert len(selected)==17
# Original inventories are read once. Verify exact selected path sets and bytes;
# unrelated failed runs are neither selected nor copied into the rollback proof.
originals={}; inventorypins=[]
for row in index['storedTrees']:
 if not any(p.is_relative_to(ROOT/row['base']) for p in selected.values()):continue
 inventory=doc(ROOT/row['file']);assert len(inventory)==row['count']
 assert digest(ROOT/row['file'])==index['files'][row['file']]
 originals[row['base']]=inventory
 inventorypins.append(dict(path=row['file'],sha256=digest(ROOT/row['file']),files=len(inventory)))
verified={};summaries=[]
for name,path in selected.items():
 rel=str(path.relative_to(ROOT)); owner=next(k for k in originals if path.is_relative_to(ROOT/k)); prefix=str(path.relative_to(ROOT/owner))+'/'
 expected={p[len(prefix):]:sha for p,sha in originals[owner].items() if p.startswith(prefix)}
 actual={str((Path(p)/f).relative_to(path)) for p,dirs,files in os.walk(path) for f in files}
 assert actual==set(expected),(name,len(actual),len(expected))
 for p,sha in expected.items():
  file=path/p
  assert file.is_file() and not file.is_symlink() and digest(file)==sha,str(file)
  verified[str(file.relative_to(ROOT))]=sha
 record=dict(name=name,path=rel,files=len(expected),inventoryVerified=True)
 if (path/'manifest.json').is_file():
  m=doc(path/'manifest.json');record['manifestSHA256']=digest(path/'manifest.json')
  record['cases']=dict(collections.Counter(c['status'] for c in m.get('cases',[])))
  if 'versions' in m:record.update(gitCommit=m['git_commit'],catalog=m['versions'])
 for reportname in ['live-report.json','evaluation-report.json','report.json','formal-report.json','audit.json']:
  if (path/reportname).is_file():record.setdefault('reportHashes',{})[reportname]=digest(path/reportname)
 summaries.append(record)
 print(name,len(expected),'verified',flush=True)
# Verify the recorded complete measured source set from immutable Git objects.
names=index['sourcePaths']; source=index['sourceCommit']
obj=subprocess.check_output(['git','ls-tree','-r','-z',source],cwd=ROOT);oids={}
for row in obj.split(b'\0'):
 if row:
  head,p=row.split(b'\t',1);oids[os.fsdecode(p)]=head.split()[2].decode()
proc=subprocess.run(['git','cat-file','--batch'],cwd=ROOT,input=''.join(oids[n]+'\n' for n in names).encode(),stdout=subprocess.PIPE,check=True)
b=proc.stdout;off=0
for name in names:
 end=b.index(b'\n',off);header=b[off:end].split();off=end+1;size=int(header[2]);rawfile=b[off:off+size];off+=size+1
 assert hashlib.sha256(rawfile).hexdigest()==index['files'][name],name
assert off==len(b)
dump(OUT/'selected-file-hashes.json',verified)
(OUT/'baseline-evidence-index.json').write_bytes(raw)
dump(OUT/'reconciliation.json',dict(passVerdict=True,baselineCommit=BASE,baselineTree=subprocess.check_output(['git','rev-parse',BASE+'^{tree}'],cwd=ROOT,text=True).strip(),
 originalTrackedIndex=indexpath,trackedIndexSHA256=hashlib.sha256(raw).hexdigest(),
 releaseReport=dict(path=str((base/'gate/release-report.json').relative_to(ROOT)),sha256=digest(base/'gate/release-report.json'),executionCommit=report['commit'],gates=report['gates'],snapshots=report['snapshots']),
 measuredSourceCommit=source,completeMeasuredSourceFiles=len(names),allSourceHashesVerifiedAgainstGit=True,
 priorInventories=inventorypins,selected=summaries,selectedFileCount=len(verified),selectedInventorySHA256=digest(OUT/'selected-file-hashes.json'),
 applicability='Historical PASS at its recorded source/catalog; no paid calls, no current-source renewal or billing claim',credentialsRead=False,readOnlyOriginals=True))
print('historical selected rollback evidence PASS',len(verified),flush=True)
