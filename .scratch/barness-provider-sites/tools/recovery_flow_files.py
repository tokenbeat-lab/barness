"""Bounded common-owner claim inventory for the compiled fixed contexts."""
import json
import os
from recovery_flow_source import SourceFreeze
from pathlib import Path

def claim_names(core,round_id):
    root=core.common_root()/'.evidence/minimax-launches'
    if root.is_symlink() or not root.is_dir():raise RuntimeError('canonical launch directory required')
    result=set();count=0;prefix=round_id+'-'
    with os.scandir(root) as entries:
        for entry in entries:
            count+=1
            if count>4096:raise RuntimeError('canonical claim inventory ceiling exceeded')
            if not entry.name.startswith(prefix):continue
            if entry.is_symlink() or not entry.is_file(follow_symlinks=False) or not entry.name.endswith('.json'):raise RuntimeError('canonical own regular claim required')
            result.add(entry.name.removeprefix(prefix).removesuffix('.json'))
    return result


class Artifacts:
    def __init__(self,core):self.source=SourceFreeze(core)
    def read(self,path,limit=16<<20):return self.source.read(path,limit)
    def digest(self,path):return self.source.digest(path)
    def doc(self,path):
        def pairs(items):
            value={};seen=set()
            for key,item in items:
                if key.casefold() in seen:raise RuntimeError('duplicate or case-alias artifact JSON field')
                seen.add(key.casefold());value[key]=item
            return value
        value=json.loads(self.read(path,2<<20),object_pairs_hook=pairs)
        if not isinstance(value,dict):raise RuntimeError('artifact JSON object required')
        return value
    def tree(self,base):
        if base.is_symlink() or not base.is_dir():raise RuntimeError('regular capture directory required')
        pending=[(base,0)];paths=[];entries=0
        while pending:
            directory,depth=pending.pop()
            with os.scandir(directory) as scanned:
                for entry in scanned:
                    entries+=1
                    if entries>1024 or depth+1>3 or entry.is_symlink():raise RuntimeError('capture entry/depth/link ceiling')
                    path=Path(entry.path)
                    if entry.is_dir(follow_symlinks=False):pending.append((path,depth+1))
                    elif entry.is_file(follow_symlinks=False):
                        paths.append(path)
                        if len(paths)>512:raise RuntimeError('capture file ceiling')
                    else:raise RuntimeError('regular capture entry required')
        files={};total=0
        for path in paths:
            raw=self.read(path,min(16<<20,(128<<20)-total));total+=len(raw);files[path.relative_to(base).as_posix()]=raw
        if not files:raise RuntimeError('empty capture directory')
        return files
