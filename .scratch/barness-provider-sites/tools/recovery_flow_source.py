"""Exact existing runtime/operator selector with bounded nonblocking file hashes."""
from contextlib import contextmanager
import hashlib
import os
from pathlib import Path
import stat
import subprocess

class SourceFreeze:
    def __init__(self,core):self.c=core

    @contextmanager
    def regular(self,path,limit):
        path=Path(path);before=path.lstat()
        if not stat.S_ISREG(before.st_mode) or before.st_size>limit:raise RuntimeError('bounded regular source/artifact required')
        fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
        try:
            opened=os.fstat(fd)
            if not stat.S_ISREG(opened.st_mode) or not os.path.samestat(before,opened) or (before.st_size,before.st_mtime_ns)!=(opened.st_size,opened.st_mtime_ns):raise RuntimeError('source/artifact changed before nonblocking open')
            yield fd,before.st_size
            after=path.lstat()
            if not stat.S_ISREG(after.st_mode) or not os.path.samestat(before,after) or (before.st_size,before.st_mtime_ns)!=(after.st_size,after.st_mtime_ns):raise RuntimeError('source/artifact changed during bounded read')
        finally:os.close(fd)

    def digest(self,path):
        with self.regular(path,64<<20) as (fd,length):
            h=hashlib.sha256();size=0
            while True:
                raw=os.read(fd,min(65536,length-size+1))
                if not raw:break
                size+=len(raw)
                if size>length:raise RuntimeError('source/tool grew while hashing')
                h.update(raw)
            if size!=length:raise RuntimeError('source/tool truncated during hash')
        return h.hexdigest()

    def read(self,path,limit=16<<20):
        with self.regular(path,limit) as (fd,length):
            chunks=[];size=0
            while True:
                raw=os.read(fd,min(65536,length-size+1))
                if not raw:break
                chunks.append(raw);size+=len(raw)
                if size>length:raise RuntimeError('artifact grew during read')
            if size!=length:raise RuntimeError('artifact truncated during read')
        return b''.join(chunks)

    def files(self):
        # This is the existing c.sources + final_state.sources selector, including
        # all own operator code. No execution dependency is removed to reuse proof.
        root=self.c.ROOT;names={'go.mod','go.sum'};entries=0
        for base,dirs,files in os.walk(root/'ai'):
            dirs[:]=[n for n in dirs if n not in ('node_modules','.evidence','.git')]
            for name in dirs:
                if (Path(base)/name).is_symlink():raise RuntimeError('source directory symlink refused')
            entries+=len(dirs)+len(files)
            if entries>32768:raise RuntimeError('bounded source inventory exceeded')
            for name in files:
                path=Path(base)/name;rel=path.relative_to(root).as_posix()
                if (name.endswith(('.go','.json')) or rel.startswith('ai/live/testdata/minimax/') or rel=='ai/internal/testkit/minimax/pricing-source.md') and rel not in ('ai/live/support-matrix.json','ai/live/testdata/minimax-formal/origins.json'):names.add(rel)
        names.update(p.relative_to(root).as_posix() for p in (root/'.scratch/barness-ai-minimax/tools').glob('*.py'))
        names.update(p.relative_to(root).as_posix() for p in (root/'.scratch/barness-provider-sites/tools').iterdir() if p.suffix in ('.py','.md'))
        if len(names)>8192:raise RuntimeError('bounded source file count exceeded')
        result={};total=0
        for name in sorted(names):
            result[name]=self.digest(root/name);total+=(root/name).lstat().st_size
            if total>512<<20:raise RuntimeError('bounded complete source bytes exceeded')
        return result

    def freeze(self):
        commit=self.c.git('rev-parse','HEAD');files=self.files()
        for name,sha in files.items():
            raw=subprocess.check_output(['git','show',commit+':'+name],cwd=self.c.ROOT,env=self.c.ENV,stderr=subprocess.DEVNULL)
            if hashlib.sha256(raw).hexdigest()!=sha:raise RuntimeError('all operator/runtime source must be committed before freeze')
        return dict(sourceCommit=commit,files=files)
