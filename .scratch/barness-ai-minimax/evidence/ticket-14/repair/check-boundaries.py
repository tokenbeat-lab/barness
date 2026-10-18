#!/usr/bin/env python3
"""Replay finite read-only integrity refusals against temporary synthetic inputs.

Failure inventory: modified/oversized/link archive, linked parent, redirected
archive, enlarged limits, laundered RED or missing-fixture result, invented paid
authority, changed link metadata, tar link, wrong binary reference, changed helper
and changed source identity. No captured command, key, provider or network runs.
"""
import argparse
import gzip
import hashlib
import io
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
    files=sorted(p for p in base.rglob('*') if p.is_file() and p.name!='SHA256SUMS')
    (base/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(base).as_posix()+'\n' for p in files))


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args()
    index=json.loads((BASE/'index.json').read_text());archive=index['archive'];results=[]
    env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ}
    def invoke(program,root,reasons=None):
        process=subprocess.run([sys.executable,'-B',str(program),'--common-root',str(root)],env=env,text=True,capture_output=True,timeout=45)
        if reasons is None:
            if process.returncode!=0:raise ValueError('original capsule refused: '+process.stderr[-300:])
        elif process.returncode==0 or not any(v in process.stderr for v in reasons):raise ValueError('refusal differed: '+process.stderr[-300:])
        return process.returncode
    invoke(BASE/'verify.py',a.common_root)
    for kind,reasons in [('wrong-hash',['compressed archive pin']),('symlink',['OSError']),('oversized',['compressed archive size']),('symlink-parent',['OSError','NotADirectoryError'])]:
        with tempfile.TemporaryDirectory(prefix='ticket14-archive-boundary-') as folder:
            root=Path(folder).resolve();path=root/archive['file'];path.parent.mkdir(parents=True)
            if kind=='wrong-hash':
                with path.open('wb') as stream:stream.truncate(archive['bytes'])
            elif kind=='symlink':path.symlink_to(a.common_root/archive['file'])
            elif kind=='oversized':
                with path.open('wb') as stream:stream.truncate((192<<20)+1)
            else:path.parent.rmdir();path.parent.symlink_to((a.common_root/archive['file']).parent,target_is_directory=True)
            results.append(dict(case=kind,exitCode=invoke(BASE/'verify.py',root,reasons),rejected=True))
    cases=[('launder-code-red','independent repair summary differs'),('launder-missing-fixture','independent repair summary differs'),
        ('invent-paid-approval','independent repair summary differs'),('redirect-archive','fixed repair capsule identity/path'),
        ('enlarge-limit','fixed streaming limits'),('change-link-target','negative-fixture metadata projection'),
        ('tar-symlink','owned regular tar member required'),('change-unused-binary','unused binary hash'),
        ('change-helper','pinned bounded verifier'),('change-source','fixed repair capsule identity/path')]
    for kind,reason in cases:
        with tempfile.TemporaryDirectory(prefix='ticket14-semantic-boundary-') as folder:
            root=Path(folder).resolve();copied=root/'evidence/ticket-14/repair';shutil.copytree(BASE,copied)
            helper=root/'evidence/ticket-13/implementation';helper.mkdir(parents=True)
            shutil.copyfile(BASE.parents[1]/'ticket-13/implementation/verify.py',helper/'verify.py')
            common=a.common_root
            if kind.startswith('launder-') or kind=='invent-paid-approval':
                path=copied/'summary.json';value=json.loads(path.read_text())
                if kind=='launder-code-red':value['publicContractRED']['FAIL']=0
                elif kind=='launder-missing-fixture':value['actualPreflightFirst']['passed']=True
                else:value['authorization']='APPROVED'
                path.write_text(json.dumps(value))
            elif kind=='change-helper':
                with (helper/'verify.py').open('a') as stream:stream.write('\n')
            elif kind=='change-unused-binary':
                path=copied/'external-references.json';value=json.loads(path.read_text());value['binaries'][0]['sha256']='0'*64;path.write_text(json.dumps(value))
            else:
                path=copied/'index.json';value=json.loads(path.read_text())
                if kind=='redirect-archive':value['archive']['file']='.evidence/foreign/raw.tar.gz'
                elif kind=='enlarge-limit':value['limits']['decodedTarBytes']+=1
                elif kind=='change-source':value['codeCommit']='8fbd17006677774368d4d77d924a8672fd205c51'
                elif kind=='change-link-target':next(iter(value['negativeFixtureLinkMetadata'].values()))['target']+='changed'
                else:
                    common=root/'synthetic-common';destination=common/archive['file'];destination.parent.mkdir(parents=True)
                    with destination.open('wb') as output:
                        with gzip.GzipFile(fileobj=output,mode='wb',mtime=0) as compressed:
                            with tarfile.open(fileobj=compressed,mode='w|') as tar:
                                member=tarfile.TarInfo('release-mapping-repair/negative-fixture-links.json');member.type=tarfile.SYMTYPE;member.linkname='/untrusted-target';tar.addfile(member)
                    raw=destination.read_bytes();value['archive']['bytes']=len(raw);value['archive']['sha256']=hashlib.sha256(raw).hexdigest()
                path.write_text(json.dumps(value))
            seal(copied);results.append(dict(case=kind,exitCode=invoke(copied/'verify.py',common,[reason]),rejected=True))
    invoke(BASE/'verify.py',a.common_root)
    print(json.dumps(dict(schema=1,retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,
        originalArchiveVerifiedBeforeAndAfter=True,results=results),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
