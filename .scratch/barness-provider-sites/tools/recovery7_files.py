"""No-follow bounded inventory of the two sealed sixth collection namespaces."""
import os
import recovery7_contract as r

ROOTS=(r.BASE/'recovery6',r.BASE/'recovery6-preparation/actual-execution')

def verify(stop):
    files=set(stop['files']);fixtures=set(stop.get('nonRegularRejectionFixtures',{}));count=0
    flags=os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW|os.O_NONBLOCK
    owner=os.open(r.c.common_root(),flags)
    try:
        for base in ROOTS:
            prefix=str(base)+'/';expected={name for name in files if name.startswith(prefix)};links={name for name in fixtures if name.startswith(prefix)};seen=set();seen_links=set()
            fd=os.dup(owner)
            try:
                try:
                    for part in base.parts:
                        child=os.open(part,flags,dir_fd=fd);os.close(fd);fd=child
                except FileNotFoundError:
                    if expected or links:raise RuntimeError('closed6 original collection directory missing')
                    continue
                def walk(directory,label,depth):
                    nonlocal count
                    if depth>32:raise RuntimeError('closed6 directory depth ceiling')
                    with os.scandir(directory) as entries:
                        for entry in entries:
                            count+=1
                            if count>65536:raise RuntimeError('closed6 total entry ceiling')
                            name=label+'/'+entry.name
                            if entry.is_symlink():
                                if name not in links:raise RuntimeError('closed6 unpinned symbolic link')
                                seen_links.add(name)
                            elif entry.is_dir(follow_symlinks=False):
                                child=os.open(entry.name,flags,dir_fd=directory)
                                try:walk(child,name,depth+1)
                                finally:os.close(child)
                            elif entry.is_file(follow_symlinks=False):
                                if name not in expected:raise RuntimeError('closed6 unpinned regular file')
                                seen.add(name)
                            else:raise RuntimeError('closed6 unpinned nonregular entry')
                walk(fd,str(base),0)
                if seen!=expected or seen_links!=links:raise RuntimeError('closed6 original namespace inventory differs')
            finally:os.close(fd)
    finally:os.close(owner)
