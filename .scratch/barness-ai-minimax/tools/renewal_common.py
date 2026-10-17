"""Bounded regular-file imports and keyless public CLI execution."""
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import tempfile
import zipfile
import paid_core as c

ROUTES=('chat','responses','messages')
HEX=re.compile(r'^[a-f0-9]{64}$')


def file_bytes(path,limit=16<<20):
    fd=os.open(path,os.O_RDONLY|getattr(os,'O_NOFOLLOW',0))
    with os.fdopen(fd,'rb') as stream:
        info=os.fstat(stream.fileno())
        if not stat.S_ISREG(info.st_mode) or info.st_size>limit:raise RuntimeError('regular bounded artifact required')
        raw=stream.read(limit+1)
        if len(raw)>limit:raise RuntimeError('artifact import byte budget exceeded')
        return raw


def sha(raw):return hashlib.sha256(raw).hexdigest()


def json_doc(path,limit=2<<20):
    def pairs(items):
        out={}
        for key,value in items:
            if key in out:raise RuntimeError('duplicate artifact JSON field')
            out[key]=value
        return out
    value=json.loads(file_bytes(path,limit),object_pairs_hook=pairs)
    if not isinstance(value,dict):raise RuntimeError('artifact JSON object required')
    return value


def repo_path(value):
    path=Path(value)
    if '..' in path.parts:raise RuntimeError('parent path forbidden')
    path=c.ROOT/path
    if not path.is_relative_to(c.ROOT):raise RuntimeError('repository-contained path required')
    cur=c.ROOT
    for name in path.relative_to(c.ROOT).parts:
        cur=cur/name
        if cur.is_symlink():raise RuntimeError('symlink path forbidden')
    return path


def output(value,existing=False):
    path=repo_path(value);parts=path.relative_to(c.ROOT).parts
    if len(parts)<2 or parts[0]!='.evidence' or 'minimax' not in parts[1]:
        raise RuntimeError('new task-owned .evidence/minimax collection required')
    if not existing:path.mkdir(parents=True,exist_ok=False,mode=0o700)
    elif not path.is_dir():raise RuntimeError('existing prepared renewal directory required')
    return path


def public(out,name,argv,timeout=240,env=None,cwd=None):
    if c.execute(out,name,argv,env=env,timeout=timeout,cwd=cwd or c.ROOT)!=0:
        raise RuntimeError('keyless public validation failed; inspect redacted command receipt')


def tree(base):
    if base.is_symlink() or not base.is_dir():raise RuntimeError('regular capture directory required')
    # Bound the entire shape before reading or hashing any payload. Empty
    # directories count too, and scandir must stop before collecting fanout.
    pending=[(base,0)];paths=[];entries=0
    while pending:
        directory,depth=pending.pop()
        with os.scandir(directory) as scanned:
            for entry in scanned:
                entries+=1
                if entries>1024:raise RuntimeError('capture total entries exceed1024')
                if depth+1>3:raise RuntimeError('capture depth exceeds3')
                if entry.is_symlink():raise RuntimeError('symlink capture entry forbidden')
                path=Path(entry.path)
                if entry.is_dir(follow_symlinks=False):pending.append((path,depth+1))
                elif entry.is_file(follow_symlinks=False):
                    if len(paths)>=512:raise RuntimeError('capture file count exceeds512')
                    paths.append(path)
                else:raise RuntimeError('regular capture entry required')
    files={};total=0
    for path in paths:
        raw=file_bytes(path,min(16<<20,(128<<20)-total));total+=len(raw)
        files[path.relative_to(base).as_posix()]=raw
    if not files:raise RuntimeError('empty capture directory')
    return files


def archive(path,files,prefix=''):
    with zipfile.ZipFile(path,'x',zipfile.ZIP_DEFLATED) as zipped:
        for name,raw in sorted(files.items()):
            info=zipfile.ZipInfo(prefix+name);info.external_attr=(stat.S_IFREG|0o600)<<16
            info.compress_type=zipfile.ZIP_DEFLATED;zipped.writestr(info,raw)
    return c.digest(path)


def bundle(value,leaf):
    path=repo_path(value)
    if (path/leaf).is_file():return path
    # The one-level bundle lookup has its own bounded streaming inventory;
    # an eager glob would allocate every attacker-controlled child first.
    directories=[];entries=0
    with os.scandir(path) as scanned:
        for entry in scanned:
            entries+=1
            if entries>1024:raise RuntimeError('bundle lookup entries exceed1024')
            if entry.is_symlink():raise RuntimeError('symlink bundle entry forbidden')
            if entry.is_dir(follow_symlinks=False):directories.append(Path(entry.path))
    found=[]
    for directory in directories:
        report=directory/leaf
        if report.is_symlink():raise RuntimeError('symlink bundle report forbidden')
        if report.is_file():
            if found:raise RuntimeError('exactly one current own bundle required')
            found.append(directory)
    if len(found)!=1:raise RuntimeError('exactly one current own bundle required')
    return found[0]


def committed(names):
    commit=c.git('rev-parse','HEAD')
    import subprocess
    for name in names:
        raw=subprocess.check_output(['git','show',commit+':'+name],cwd=c.ROOT,env=c.ENV)
        if raw!=file_bytes(c.ROOT/name):raise RuntimeError('renewal selector data must be committed before deriving views')
    return commit


def validate_glm(out,paths,name):
    # The public importer validates a temporary matrix. The actual matrix is
    # untouched, and imported report/capture files stay byte-for-byte unchanged.
    with tempfile.TemporaryDirectory(prefix='matrix-',dir=out) as temporary:
        matrix=Path(temporary)/'matrix.json'
        matrix.write_bytes(file_bytes(c.ROOT/'ai/live/support-matrix.json'))
        public(out,name,['go','run','./ai/live/cmd/supportmatrix','-matrix',matrix,*paths])
