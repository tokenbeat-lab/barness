#!/usr/bin/env python3
"""Fourteen finite public CLI refusals; originals and captured commands stay untouched.

Failure inventory: corrupted/oversized/link archive and linked parent, altered
outer verdict/numeric exit/unit failure, redirected archive, enlarged limits,
tar link, changed helper/source/binary and concealed suite failure count.
"""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile
import tempfile

BASE=Path(__file__).resolve().parent


def seal(base):
    files=sorted(p for p in base.iterdir() if p.is_file() and p.name!='SHA256SUMS')
    (base/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files))


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args()
    ordinary=json.loads((BASE/'ordinary-index.json').read_text())['archive'];verdict=json.loads((BASE/'verdict-index.json').read_text())['archive'];results=[]
    env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ}
    def invoke(program,root,reasons=None):
        p=subprocess.run([sys.executable,'-B',str(program),'--common-root',str(root)],env=env,text=True,capture_output=True,timeout=180)
        if reasons is None:
            if p.returncode!=0:raise ValueError('original public verification failed: '+p.stderr[-300:])
        elif p.returncode==0 or not any(v in p.stderr for v in reasons):raise ValueError('public refusal differed: '+p.stderr[-300:])
        return p.returncode
    invoke(BASE/'verify.py',a.common_root)
    for kind,reasons in [('wrong-hash',['compressed archive pin']),('symlink',['OSError']),('oversized',['compressed archive size']),('symlink-parent',['OSError','NotADirectoryError'])]:
        with tempfile.TemporaryDirectory(prefix='fullgate-archive-boundary-') as folder:
            root=Path(folder).resolve();path=root/ordinary['file'];path.parent.mkdir(parents=True);shutil.copyfile(a.common_root/verdict['file'],root/verdict['file'])
            if kind=='wrong-hash':
                with path.open('wb') as stream:stream.truncate(ordinary['bytes'])
            elif kind=='symlink':path.symlink_to(a.common_root/ordinary['file'])
            elif kind=='oversized':
                with path.open('wb') as stream:stream.truncate((256<<20)+1)
            else:(root/verdict['file']).unlink();path.parent.rmdir();path.parent.symlink_to((a.common_root/ordinary['file']).parent,target_is_directory=True)
            results.append(dict(case=kind,exitCode=invoke(BASE/'verify.py',root,reasons),rejected=True))
    cases=[('launder-release-pass','independent outer failed verdict'),('launder-race-exit','independent outer failed verdict'),('hide-extra-unit-failure','independent outer failed verdict'),
        ('redirect-archive','fixed final-gate capsule identity/path'),('enlarge-limit','fixed streaming limits'),('tar-symlink','owned regular tar member'),
        ('change-helper','pinned bounded verifier'),('change-source','fixed final-gate capsule identity/path'),('drop-failure-count','independent complete gate summary'),('change-binary','compiled binary reference')]
    for kind,reason in cases:
        with tempfile.TemporaryDirectory(prefix='fullgate-semantic-boundary-') as folder:
            root=Path(folder).resolve();copied=root/'final-gate';shutil.copytree(BASE,copied);helper=root/'implementation';helper.mkdir()
            shutil.copyfile(BASE.parent/'implementation/verify.py',helper/'verify.py');common=a.common_root
            if kind in ('launder-release-pass','launder-race-exit','hide-extra-unit-failure','drop-failure-count'):
                path=copied/'summary.json';value=json.loads(path.read_text())
                if kind=='launder-release-pass':value['outer']['releasePass']=True
                elif kind=='launder-race-exit':value['outer']['actualNumericExits']['race']=0
                elif kind=='hide-extra-unit-failure':value['outer']['additionalUnregisteredUnitFailure']='none'
                else:value['suites'][0]['statuses']['FAIL']=0
                path.write_text(json.dumps(value))
            elif kind=='change-helper':
                with (helper/'verify.py').open('a') as stream:stream.write('\n')
            elif kind=='change-binary':
                path=copied/'binary-reference.json';value=json.loads(path.read_text());value['sha256']='0'*64;path.write_text(json.dumps(value))
            else:
                path=copied/'verdict-index.json';value=json.loads(path.read_text())
                if kind=='redirect-archive':value['archive']['file']='.evidence/foreign/raw.tar.gz'
                elif kind=='enlarge-limit':value['limits']['decodedTarBytes']+=1
                elif kind=='change-source':value['sourceCommit']='c2604db6f693f5c7d3a82c656512004ec85d2249'
                else:
                    common=root/'synthetic-common';destination=common/verdict['file'];destination.parent.mkdir(parents=True)
                    with destination.open('wb') as output:
                        with gzip.GzipFile(fileobj=output,mode='wb',mtime=0) as zipped:
                            with tarfile.open(fileobj=zipped,mode='w|') as tar:
                                member=tarfile.TarInfo('routed-compiled-verdict/outer/summary.json');member.type=tarfile.SYMTYPE;member.linkname='/untrusted';tar.addfile(member)
                    raw=destination.read_bytes();value['archive']['bytes']=len(raw);value['archive']['sha256']=hashlib.sha256(raw).hexdigest()
                path.write_text(json.dumps(value))
            seal(copied);results.append(dict(case=kind,exitCode=invoke(copied/'verify.py',common,[reason]),rejected=True))
    invoke(BASE/'verify.py',a.common_root)
    print(json.dumps(dict(schema=1,retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,originalArchivesVerifiedBeforeAndAfter=True,results=results),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
