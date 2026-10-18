import os,stat
from pathlib import Path
root=Path('/Users/cyber/RestoX/barness/.evidence/minimax-release-repair-effects-1')
pending=[(root,0)];entries=0
while pending:
 folder,depth=pending.pop()
 if depth>32:raise ValueError('snapshot depth bound')
 with os.scandir(folder) as rows:
  for row in rows:
   entries+=1
   if entries>200000:raise ValueError('whole shape bound')
   if row.is_symlink():raise ValueError('unexpected snapshot link forbidden: '+str(Path(row.path).relative_to(root)))
   if row.is_dir(follow_symlinks=False):pending.append((Path(row.path),depth+1))
   elif not stat.S_ISREG(row.stat(follow_symlinks=False).st_mode):raise ValueError('regular originals required')
print('regular scan complete')
