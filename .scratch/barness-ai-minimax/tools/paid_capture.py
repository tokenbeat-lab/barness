"""Bounded immutable research retention; no credentials or vendor path."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import stat
import tempfile


def bounded(path,maximum=16<<20):
    before=path.lstat()
    if not stat.S_ISREG(before.st_mode):raise RuntimeError('unsafe research capture')
    with path.open('rb') as stream:
        info=os.fstat(stream.fileno())
        if not stat.S_ISREG(info.st_mode) or (before.st_dev,before.st_ino)!=(info.st_dev,info.st_ino) or info.st_size>maximum:
            raise RuntimeError('research capture changed or oversized')
        raw=stream.read(maximum+1)
    if len(raw)>maximum:raise RuntimeError('research capture grew beyond limit')
    return raw


def retain(source,target):
    """Own sealed bytes survive a new prepared source root or Git worktree."""
    report=json.loads(bounded(source/'minimax-report.json',2<<20))
    if not isinstance(report,dict):raise RuntimeError('research report object required')
    session=report.get('session')
    if not isinstance(session,str) or not session or Path(session).name!=session or source.name!=session or report.get('mode')!='live' or report.get('kind')!='minimax-research-only':
        raise RuntimeError('only own actual research capture can be retained')
    seal=bounded(source/'minimax-integrity.json',2<<20)
    inventory=json.loads(seal)
    if not isinstance(inventory,dict) or inventory.get('schema')!=1 or not isinstance(inventory.get('files'),dict) or len(inventory['files'])>512:raise RuntimeError('invalid research inventory')
    files={};total=0;entries=0
    def visit(base,depth):
        nonlocal entries,total
        if depth>2 or base.is_symlink():raise RuntimeError('unsafe research nesting')
        with os.scandir(base) as children:
            for entry in children:
                entries+=1
                if entries>1024:raise RuntimeError('research entry limit exceeded')
                if entry.is_symlink():raise RuntimeError('research symlink forbidden')
                path=base/entry.name
                if entry.is_dir(follow_symlinks=False):visit(path,depth+1);continue
                if len(files)>=512:raise RuntimeError('research file limit exceeded')
                raw=bounded(path,min(16<<20,(12<<20)-total));total+=len(raw)
                files[path.relative_to(source).as_posix()]=raw
    visit(source,0)
    actual={n:hashlib.sha256(b).hexdigest() for n,b in files.items() if n!='minimax-integrity.json'}
    if actual!=inventory['files']:raise RuntimeError('research sealed bytes changed')
    target.parent.mkdir(parents=True,exist_ok=True,mode=0o700)
    lock=target.with_name(target.name+'.lock');lock.mkdir(mode=0o700)
    stage=None
    try:
        if target.exists() or target.is_symlink():raise RuntimeError('research retention already exists')
        stage=Path(tempfile.mkdtemp(prefix='.research-',dir=target.parent))
        for rel,raw in files.items():
            path=stage/session/rel;path.parent.mkdir(parents=True,exist_ok=True,mode=0o700)
            with path.open('xb') as stream:stream.write(raw);stream.flush();os.fsync(stream.fileno())
        os.rename(stage,target);stage=None
    finally:
        if stage is not None:shutil.rmtree(stage)
        lock.rmdir()
    return dict(session=session,reportSha256=hashlib.sha256(files['minimax-report.json']).hexdigest(),
                sealSha256=hashlib.sha256(seal).hexdigest(),files=len(files),bytes=total)
