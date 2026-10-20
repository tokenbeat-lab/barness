#!/usr/bin/env python3
"""Restore and verify the complete pre-MiniMax software outside user checkouts."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import tarfile
import tempfile

ROOT=Path(__file__).resolve().parents[4]
HERE=Path(__file__).resolve().parent
BASELINE='ecfd4b4d6ec54db30453f4b5112989e9fdbd0782'
TREE='b9798a4e96dfb08cbfcce45c182915589b42eed1'
# No credential, endpoint, proxy or vendor environment is inherited.
ENV={k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}


def git(*args,cwd=ROOT):
 return subprocess.check_output(['git',*args],cwd=cwd,env=ENV)


def dump(path,value):
 path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')


def digest(raw):
 return hashlib.sha256(raw).hexdigest()


def file_state(path):
 if path.is_symlink():return dict(kind='symlink',sha256=digest(os.fsencode(os.readlink(path))))
 if not path.exists():return dict(kind='absent')
 return dict(kind='file',sha256=digest(path.read_bytes()),mode=oct(stat.S_IMODE(path.stat().st_mode)))


def snapshot(root):
 # Status detects changed clean files; hashes also detect edits that leave an
 # existing M/?? status unchanged. Symlinks are hashed without following them.
 changed=git('diff','--name-only','-z','HEAD',cwd=root).split(b'\0')
 untracked=git('ls-files','--others','--exclude-standard','-z',cwd=root).split(b'\0')
 return dict(head=git('rev-parse','HEAD',cwd=root).decode().strip(),
  branch=git('branch','--show-current',cwd=root).decode().strip(),
  status=git('status','--porcelain=v1','--untracked-files=all',cwd=root).decode(),
  changed={os.fsdecode(n):file_state(root/os.fsdecode(n)) for n in changed if n},
  untracked={os.fsdecode(n):file_state(root/os.fsdecode(n)) for n in untracked if n})


def roots():
 common=(ROOT/Path(git('rev-parse','--git-common-dir').decode().strip())).resolve()
 return dict(caller=snapshot(ROOT),userRoot=snapshot(common.parent))


def baseline():
 # This proof owns its fixed baseline and receipts independently of the GLM
 # historical rollout. Exact blob/mode verification deliberately parallels
 # that proof; importing GLM globals would couple their distinct contracts.
 entries={}
 for row in git('ls-tree','-r','-z',BASELINE).split(b'\0'):
  if not row:continue
  head,name=row.split(b'\t',1);mode,kind,oid=head.decode().split()
  if kind!='blob':raise RuntimeError('unexpected non-blob baseline entry')
  entries[os.fsdecode(name)]=dict(mode=mode,gitBlob=oid)
 raw=subprocess.run(['git','cat-file','--batch'],cwd=ROOT,env=ENV,input=''.join(v['gitBlob']+'\n' for v in entries.values()).encode(),stdout=subprocess.PIPE,check=True,timeout=60).stdout
 blobs={};offset=0
 for name in entries:
  end=raw.index(b'\n',offset);header=raw[offset:end].split();offset=end+1
  if len(header)!=3 or header[1]!=b'blob':raise RuntimeError('missing baseline blob')
  size=int(header[2]);blobs[name]=raw[offset:offset+size];offset+=size+1
 if offset!=len(raw):raise RuntimeError('unexpected blob tail')
 return entries,blobs


def verify(root,entries,blobs):
 names=set()
 for parent,dirs,files in os.walk(root,followlinks=False):
  for name in files+[n for n in dirs if (Path(parent)/n).is_symlink()]:names.add(str((Path(parent)/name).relative_to(root)))
 if names!=set(entries):raise RuntimeError('restored full path set differs')
 rows={}
 for name,row in entries.items():
  path=root/name;raw=os.fsencode(os.readlink(path)) if path.is_symlink() else path.read_bytes()
  mode='120000' if path.is_symlink() else '100755' if path.stat().st_mode&stat.S_IXUSR else '100644'
  if raw!=blobs[name] or mode!=row['mode']:raise RuntimeError('restored blob or mode differs')
  rows[name]=row|dict(sha256=digest(raw),bytes=len(raw))
 return rows


def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('out');args=parser.parse_args()
 if Path(args.out).is_absolute() or '..' in Path(args.out).parts:parser.error('new repository-relative output required')
 out=ROOT/args.out;out.mkdir(parents=True,exist_ok=False,mode=0o700)
 before=roots();dump(out/'root-before.json',before)
 if git('rev-parse',BASELINE+'^{tree}').decode().strip()!=TREE:raise RuntimeError('fixed baseline tree mismatch')
 entries,blobs=baseline();archive=out/'baseline.tar'
 subprocess.run(['git','archive','--format=tar','--output',str(archive),BASELINE],cwd=ROOT,env=ENV,check=True)
 with tempfile.TemporaryDirectory(prefix='barness-minimax08-rollback-') as temp:
  restored=Path(temp)/'restored';restored.mkdir()
  with tarfile.open(archive) as tar:
   for member in tar:
    name=Path(member.name)
    if name.is_absolute() or '..' in name.parts or not(member.isfile() or member.isdir() or member.issym()):raise RuntimeError('unsafe archive entry')
   tar.extractall(restored,filter='data')
  inventory=verify(restored,entries,blobs);dump(out/'restored-tree.json',inventory)
  probe=Path(temp)/'rollback-probe.go';probe.write_bytes((HERE/'rollback-probe.txt').read_bytes())
  def scrub(value):return str(value).replace(str(temp),'$RESTORATION').replace(str(ROOT),'$REPO').replace(str(Path.home()),'$HOME')
  receipts=[]
  for name,argv in (
   ('go-version',['go','version']),
   ('build-public-probe',['go','build','-o',str(out/'baseline-probe'),str(probe)]),
   ('public-no-minimax-probe',[str(out/'baseline-probe')]),
   ('baseline-old-provider-e2e',['go','test','-v','./ai/e2e','-run','^TestLocalAssemblyExample$','-count=1'])):
   env=ENV|{'BARNESS_AI_EVIDENCE_DIR':str(out/'baseline-e2e')}
   proc=subprocess.run(argv,cwd=restored,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=180)
   (out/(name+'.log')).write_text(scrub(proc.stdout.decode(errors='replace')))
   receipts.append(dict(name=name,argv=[scrub(a) for a in argv],cwd='$RESTORATION/restored',exitCode=proc.returncode,environmentKeys=sorted(env),credentialsInherited=False))
   if proc.returncode:raise RuntimeError(name+' failed')
  dump(out/'commands.json',receipts)
  if inventory!=verify(restored,entries,blobs):raise RuntimeError('restored source mutated during proof')
  dump(out/'module-pins.json',{p:digest((restored/p).read_bytes()) for p in ('go.mod','go.sum')})
  selected=('ai/release/catalog-snapshot.json','ai/live/support-matrix.json','ai/release/traceability.json','ai/e2e/testdata/pidiff/ledger.json')
  dump(out/'baseline-release-index.json',{p:inventory[p] for p in selected})
  for path in selected:
   if 'minimax' in (restored/path).read_text().lower():raise RuntimeError('MiniMax remains in baseline release contract')
  manifests=list((out/'baseline-e2e').glob('*/manifest.json'))
  if len(manifests)!=1:raise RuntimeError('old route evidence missing')
  bundle=manifests[0].parent;manifest=json.loads(manifests[0].read_text());audit=json.loads((bundle/'audit.json').read_text())
  cases=manifest.get('cases',[])
  if not cases or any(c['status']!='PASS' for c in cases):raise RuntimeError('old route controlled case failed')
  if audit.get('findings')!=[]:raise RuntimeError('old route redaction audit failed')
  for case in cases:
   for name,expected in case['artifacts'].items():
    if digest((bundle/case['dir']/name).read_bytes())!=expected:raise RuntimeError('old route artifact inventory differs')
  dump(out/'baseline-old-provider-summary.json',dict(cases=cases,audit=audit,manifestSHA256=digest(manifests[0].read_bytes()),source='unaltered restored baseline'))
  binary_hash=digest((out/'baseline-probe').read_bytes())
 after=roots();dump(out/'root-after.json',after)
 if before!=after:raise RuntimeError('caller/user checkout or untracked work changed')
 dump(out/'rollback-report.json',dict(passVerdict=True,baselineCommit=BASELINE,baselineTree=TREE,completeTrackedFiles=len(entries),archiveSHA256=digest(archive.read_bytes()),restoredTreeVerified=True,rootUnchanged=True,publicNoMiniMaxRefusal=True,oldProviderControlledE2E=True,baselineBinarySHA256=binary_hash,ownedTemporaryRestorationRemoved=True,priorLiveEvidence='historical, not renewed formal support'))
 print('complete MiniMax baseline rollback PASS',flush=True)

if __name__=='__main__':main()
