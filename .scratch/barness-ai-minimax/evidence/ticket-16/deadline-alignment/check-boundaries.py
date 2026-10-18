#!/usr/bin/env python3
"""Finite public integrity refusals: corrupt/oversize/linked archive or parent,
forged local91s outcome/paid authority/GLM expiry, redirect/source/helper, or tar link.
All mutations are temporary; no captured test, key or vendor call is executed.
"""
import argparse,gzip,hashlib,json,os,shutil,subprocess,sys,tarfile,tempfile
from pathlib import Path
BASE=Path(__file__).resolve().parent

def seal(base):
 (base/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(base.iterdir()) if p.is_file() and p.name!='SHA256SUMS'))

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();ix=json.loads((BASE/'index.json').read_text());archive=ix['capsule']['archive'];results=[]
 env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ}
 def invoke(program,common,reasons=None):
  p=subprocess.run([sys.executable,'-B',str(program),'--common-root',str(common)],env=env,capture_output=True,text=True,timeout=45)
  if reasons is None:
   if p.returncode:raise ValueError('original refused: '+p.stderr[-300:])
  elif p.returncode==0 or not any(r in p.stderr for r in reasons):raise ValueError('refusal differs: '+p.stderr[-300:])
  return p.returncode
 invoke(BASE/'verify.py',a.common_root)
 for kind,reasons in [('wrong-hash',['compressed archive pin']),('oversized',['compressed archive size']),('symlink',['OSError']),('symlink-parent',['OSError','NotADirectoryError'])]:
  with tempfile.TemporaryDirectory(prefix='glm-header-archive-') as folder:
   root=Path(folder);p=root/archive['file'];p.parent.mkdir(parents=True)
   if kind=='wrong-hash':
    with p.open('wb') as f:f.truncate(archive['bytes'])
   elif kind=='oversized':
    with p.open('wb') as f:f.truncate((32<<20)+1)
   elif kind=='symlink':p.symlink_to(a.common_root/archive['file'])
   else:p.parent.rmdir();p.parent.symlink_to((a.common_root/archive['file']).parent,target_is_directory=True)
   results.append(dict(case=kind,exitCode=invoke(BASE/'verify.py',root,reasons),rejected=True))
 for kind,reason in [('launder-red','independent controlled summary'),('invent-paid','independent controlled summary'),('reuse-expired-glm','independent controlled summary'),('redirect-archive','fixed capsule identity'),('change-source','fixed capsule identity'),('change-helper','pinned reader'),('wrong-m3-source','M3 formal selected files'),('tar-link','regular archive member required')]:
  with tempfile.TemporaryDirectory(prefix='glm-header-semantic-') as folder:
   root=Path(folder);copy=root/'evidence/ticket-16/deadline-alignment';shutil.copytree(BASE,copy);helper=root/'evidence/ticket-13/implementation';helper.mkdir(parents=True);shutil.copyfile(BASE.parents[1]/'ticket-13/implementation/verify.py',helper/'verify.py');common=a.common_root
   if kind in ('launder-red','invent-paid','reuse-expired-glm'):
    p=copy/'summary.json';v=json.loads(p.read_text())
    if kind=='launder-red':v['commands']['red-headers']['exitCode']=0
    elif kind=='invent-paid':v['paidCalls']=1
    else:v['priorGLMSelectedMapsExpired']=0
    p.write_text(json.dumps(v))
   elif kind=='change-helper':
    with (helper/'verify.py').open('a') as f:f.write('\n')
   elif kind=='wrong-m3-source':(copy/'M3-selected-README.md').write_text('changed')
   else:
    p=copy/'index.json';v=json.loads(p.read_text())
    if kind=='redirect-archive':v['capsule']['archive']['file']='.evidence/foreign/raw.tar.gz'
    elif kind=='change-source':v['codeCommit']='0'*40
    else:
     common=root/'common';target=common/archive['file'];target.parent.mkdir(parents=True)
     with target.open('wb') as f:
      with gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as zipped:
       with tarfile.open(fileobj=zipped,mode='w|') as tar:
        member=tarfile.TarInfo('glm-header-alignment/tdd/source-green-headers.json');member.type=tarfile.SYMTYPE;member.linkname='/untrusted';tar.addfile(member)
     v['capsule']['archive'].update(bytes=target.stat().st_size,sha256=hashlib.sha256(target.read_bytes()).hexdigest())
    p.write_text(json.dumps(v))
   seal(copy);results.append(dict(case=kind,exitCode=invoke(copy/'verify.py',common,[reason]),rejected=True))
 invoke(BASE/'verify.py',a.common_root)
 print(json.dumps(dict(schema=1,results=results,originalArchiveVerifiedBeforeAfter=True,credentialsRead=False,paidCalls=0,vendorCalls=0,capturedCommandsExecuted=False),indent=2))

if __name__=='__main__':
 try:main()
 except Exception as e:print('REFUSED: '+type(e).__name__+': '+str(e),file=sys.stderr);sys.exit(1)
