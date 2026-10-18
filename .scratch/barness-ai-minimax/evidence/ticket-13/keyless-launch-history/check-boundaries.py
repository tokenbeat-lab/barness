#!/usr/bin/env python3
"""Replay ten fixed launch-history refusals using temporary metadata and synthetic files."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

BASE=Path(__file__).resolve().parent


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args()
    index=json.loads((BASE/'index.json').read_text());archive=index['archive'];results=[]
    env={name:os.environ[name] for name in ('PATH','HOME','TMPDIR') if name in os.environ}
    def invoke(program,root,expected=None):
        process=subprocess.run([sys.executable,'-B',str(program),'--common-root',str(root)],env=env,text=True,capture_output=True,timeout=90)
        if expected is None:
            if process.returncode!=0:raise ValueError('original capsule verification failed')
        elif process.returncode==0 or not any(reason in process.stderr for reason in expected):raise ValueError('integrity refusal differed')
        return process.returncode
    invoke(BASE/'verify.py',a.common_root)
    for kind,reasons in (('wrong-hash',('compressed archive pin differs',)),('symlink',('OSError',)),
        ('oversized',('compressed archive size or limit',)),('symlink-parent',('NotADirectoryError','OSError'))):
        with tempfile.TemporaryDirectory(prefix='launch-history-archive-boundary-') as temporary:
            root=Path(temporary).resolve();path=root/archive['file'];path.parent.mkdir(parents=True)
            if kind=='wrong-hash':
                with path.open('wb') as stream:stream.truncate(archive['bytes'])
            elif kind=='symlink':path.symlink_to(a.common_root/archive['file'])
            elif kind=='oversized':
                with path.open('wb') as stream:stream.truncate((192<<20)+1)
            else:path.parent.rmdir();path.parent.symlink_to((a.common_root/archive['file']).parent,target_is_directory=True)
            code=invoke(BASE/'verify.py',root,reasons);results.append(dict(case=kind,exitCode=code,rejected=True))
    for kind,reason in (('launder-interruption','independent no-certificate history summary differs'),
        ('launder-red-to-green','independent no-certificate history summary differs'),
        ('enlarge-streaming-limit','fixed streaming limits differ'),('inject-live-a2','frozen prepared allowlist differs'),
        ('redirect-archive','fixed history capsule identity/path differs'),('change-bounded-helper','pinned bounded verifier differs')):
        with tempfile.TemporaryDirectory(prefix='launch-history-semantic-boundary-') as temporary:
            root=Path(temporary).resolve();copied=root/'keyless-launch-history';shutil.copytree(BASE,copied)
            (root/'implementation').mkdir();shutil.copyfile(BASE.parent/'implementation/verify.py',root/'implementation/verify.py')
            if kind in ('launder-interruption','launder-red-to-green'):
                value=json.loads((copied/'summary.json').read_text())
                if kind=='launder-interruption':value['originalA'].update(passVerdict=True,releaseReportPresent=True,releaseCertified=True)
                else:value['probe']['redObserverRecordPresent']=True
                (copied/'summary.json').write_text(json.dumps(value))
            elif kind in ('enlarge-streaming-limit','inject-live-a2','redirect-archive'):
                value=json.loads((copied/'index.json').read_text())
                if kind=='enlarge-streaming-limit':value['limits']['decodedTarBytes']=2<<30
                elif kind=='inject-live-a2':value['compiledPreparedAllowlist'].append('gate/release-report.json')
                else:value['archive']['file']='../foreign-archive.tar.gz'
                (copied/'index.json').write_text(json.dumps(value))
            else:
                with (root/'implementation/verify.py').open('a') as stream:stream.write('\n')
            paths=sorted(p for p in copied.rglob('*') if p.is_file() and p.name!='SHA256SUMS')
            (copied/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(copied).as_posix()+'\n' for p in paths))
            code=invoke(copied/'verify.py',a.common_root,(reason,));results.append(dict(case=kind,exitCode=code,rejected=True))
    invoke(BASE/'verify.py',a.common_root)
    print(json.dumps(dict(schema=1,retentionCredentialsRead=False,retentionPaidCalls=0,actualRequestsReplayed=False,
        originalArchiveVerifiedBeforeAndAfter=True,results=results),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
