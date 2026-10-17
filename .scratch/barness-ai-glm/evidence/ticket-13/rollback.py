#!/usr/bin/env python3
"""Verify and execute a complete baseline restoration outside all worktrees."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import stat
import subprocess
import sys
import tarfile
import tempfile
sys.dont_write_bytecode=True
import commands as c
BASELINE='f5cf5362610a8bdb8f3e170f408f3858c149e540'
TREE='e0e006866c606fdab3bf02be0322bde2eae02300'
USER_ROOT=(c.ROOT/c.COMMON).resolve().parent
HERE=Path(__file__).resolve().parent


def git(*args,cwd=c.ROOT):
 return subprocess.check_output(['git',*args],cwd=cwd,env=c.ENV)


def root_state():
 names=git('ls-files','--others','--exclude-standard','-z',cwd=USER_ROOT).split(b'\0')
 files={}
 for raw in names:
  if not raw:continue
  name=os.fsdecode(raw);path=USER_ROOT/name
  files[name]=dict(sha256=c.digest(path),mode=oct(stat.S_IMODE(path.stat().st_mode)))
 return dict(head=git('rev-parse','HEAD',cwd=USER_ROOT).decode().strip(),
  branch=git('branch','--show-current',cwd=USER_ROOT).decode().strip(),
  status=git('status','--porcelain=v1','--untracked-files=all',cwd=USER_ROOT).decode(),untracked=files)


def baseline_entries():
 rows={}
 for row in git('ls-tree','-r','-z',BASELINE).split(b'\0'):
  if not row:continue
  head,name=row.split(b'\t',1);mode,kind,oid=head.decode().split()
  if kind!='blob':raise RuntimeError('unexpected non-blob baseline entry')
  rows[os.fsdecode(name)]=dict(mode=mode,gitBlob=oid)
 return rows


def baseline_blobs(entries):
 names=list(entries)
 raw=subprocess.run(['git','cat-file','--batch'],cwd=c.ROOT,env=c.ENV,
  input=''.join(entries[n]['gitBlob']+'\n' for n in names).encode(),
  stdout=subprocess.PIPE,check=True,timeout=60).stdout
 blobs={};offset=0
 for name in names:
  end=raw.index(b'\n',offset);header=raw[offset:end].split();offset=end+1
  if len(header)!=3 or header[1]!=b'blob':raise RuntimeError('baseline blob missing')
  size=int(header[2]);blobs[name]=raw[offset:offset+size];offset+=size+1
 if offset!=len(raw):raise RuntimeError('unexpected baseline object tail')
 return blobs


def verify_tree(root,entries,blobs):
 names=set()
 for parent,dirs,files in os.walk(root,followlinks=False):
  for name in files+[d for d in dirs if (Path(parent)/d).is_symlink()]:
   names.add(str((Path(parent)/name).relative_to(root)))
 if names!=set(entries):raise RuntimeError('full restored path set differs')
 result={}
 for name,row in entries.items():
  path=root/name;mode=row['mode'];raw=os.readlink(path).encode() if mode=='120000' else path.read_bytes()
  expected=blobs[name]
  actualmode='120000' if path.is_symlink() else '100755' if path.stat().st_mode&stat.S_IXUSR else '100644'
  if raw!=expected or actualmode!=mode:raise RuntimeError('restored blob/mode differs: '+name)
  result[name]=row|dict(sha256=hashlib.sha256(raw).hexdigest(),bytes=len(raw))
 return result


def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('out');args=p.parse_args()
 if Path(args.out).is_absolute() or '..' in Path(args.out).parts:p.error('new repository-relative output required')
 out=c.ROOT/args.out;out.mkdir(parents=True,exist_ok=False,mode=0o700)
 before=root_state();c.dump(out/'root-before.json',before)
 if git('rev-parse',BASELINE+'^{tree}').decode().strip()!=TREE:raise RuntimeError('baseline tree mismatch')
 archive=out/'baseline.tar'
 subprocess.run(['git','archive','--format=tar','--output',str(archive),BASELINE],cwd=c.ROOT,env=c.ENV,check=True)
 entries=baseline_entries();blobs=baseline_blobs(entries)
 with tempfile.TemporaryDirectory(prefix='barness-glm13-rollback-') as temp:
  restored=Path(temp)/'restored';restored.mkdir()
  with tarfile.open(archive) as tar:
   for member in tar:
    name=Path(member.name)
    if name.is_absolute() or '..' in name.parts or not (member.isfile() or member.isdir() or member.issym()):
     raise RuntimeError('unsafe archive member')
   tar.extractall(restored,filter='data')
  inventory=verify_tree(restored,entries,blobs);c.dump(out/'restored-tree.json',inventory)
  probe=Path(temp)/'rollback-probe.go';probe.write_bytes((HERE/'rollback-probe.txt').read_bytes())
  # execute() normally uses repository-relative CWD in receipts. The restored
  # module is outside it, so explicitly run subprocesses here with exact paths.
  receipts=[]
  for name,argv in (
   ('go-version',['go','version']),
   ('build-public-probe',['go','build','-o',str(out/'baseline-probe'),str(probe)]),
   ('public-no-glm-probe',[str(out/'baseline-probe')]),
   ('baseline-old-provider-e2e',['go','test','./ai/e2e','-run','^TestLocalAssemblyExample$','-count=1'])):
   env=c.ENV|{'BARNESS_AI_EVIDENCE_DIR':str(out/'baseline-e2e')}
   proc=subprocess.run(argv,cwd=restored,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=240)
   (out/(name+'.log')).write_bytes(proc.stdout.replace(str(Path(temp)).encode(),b'$RESTORATION'))
   receipts.append(dict(name=name,argv=[str(a).replace(str(Path(temp)),'$RESTORATION') for a in argv],
    cwd='$RESTORATION/restored',exitCode=proc.returncode,environmentKeys=sorted(env),credentialsRead=False))
   if proc.returncode:raise RuntimeError(name+' failed')
  c.dump(out/'commands.json',receipts)
  if inventory!=verify_tree(restored,entries,blobs):raise RuntimeError('baseline source mutated during proof')
  c.dump(out/'module-pins.json',{p:c.digest(restored/p) for p in ('go.mod','go.sum')})
  selected=('ai/release/catalog-snapshot.json','ai/live/support-matrix.json','ai/release/traceability.json',
   'ai/e2e/testdata/pidiff/ledger.json','.scratch/jev-1.14-multimodal-box/14-remaining-real/evidence.json')
  c.dump(out/'baseline-release-index.json',{p:inventory[p] for p in selected})
 after=root_state();c.dump(out/'root-after.json',after)
 if before!=after:raise RuntimeError('user root/HEAD/tracked or untracked bytes changed')
 c.dump(out/'rollback-report.json',dict(passVerdict=True,baselineCommit=BASELINE,baselineTree=TREE,
  archiveSHA256=c.digest(archive),completeTrackedFiles=len(entries),restoredTreeVerified=True,
  publicNoGLMRefusal=True,oldProviderControlledE2E=True,rootUnchanged=True,
  ownedTemporaryRestorationRemoved=True,baselineBinarySHA256=c.digest(out/'baseline-probe'),
  originalGitCommitFieldUnmodified=True,priorReleaseEvidence='historical at its original source/catalog; no renewed paid claims'))
 print('complete baseline rollback PASS',flush=True)
if __name__=='__main__':main()
