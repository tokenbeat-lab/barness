"""Reuse the existing two pinned bounded protocols without changing their limits."""
import hashlib,os,stat,types
from pathlib import Path
BASE=Path(__file__).resolve().parent
SMALL='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
FULL='3017cba1d4ac09e0a203c68ac768a4d508b861fa862aa04c9ca4550efebd9328'
def load(relative,pin):
 path=BASE.parent/'ticket-13'/relative
 with os.fdopen(os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK),'rb') as f:
  info=os.fstat(f.fileno());raw=f.read((64<<10)+1)
 if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=pin:raise ValueError('pinned finite reader differs')
 module=types.ModuleType('fixed_reader');module.__file__=str(path);exec(compile(raw,str(path),'exec'),module.__dict__);module.BASE=BASE;return module
def small():return load('implementation/verify.py',SMALL)
def full():return load('final-gate/capsule.py',FULL)
