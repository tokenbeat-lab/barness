"""Committed-source freeze, review guard and isolated selected-secret process."""
from contextlib import suppress
import datetime as dt
import hashlib
import functools
import json
import os
from pathlib import Path
import shlex
import signal
import stat
import subprocess
import tempfile
import time

ROOT=Path(__file__).resolve().parents[3]
ENV={k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}


def git(*args):
    return subprocess.check_output(['git',*args],cwd=ROOT,env=ENV).decode().strip()


@functools.lru_cache(maxsize=1)
def common_root():
    common=Path(git('rev-parse','--git-common-dir'))
    return (ROOT/common).resolve().parent


def doc(path):
    return json.loads(path.read_text())


def dump(path,value,exclusive=False):
    path.parent.mkdir(parents=True,exist_ok=True,mode=0o700)
    raw=json.dumps(value,ensure_ascii=False,indent=2)+'\n'
    if exclusive:
        with path.open('x') as stream:
            stream.write(raw);stream.flush();os.fsync(stream.fileno())
    else:
        fd,name=tempfile.mkstemp(prefix='.receipt-',dir=path.parent)
        try:
            with os.fdopen(fd,'w') as stream:
                stream.write(raw);stream.flush();os.fsync(stream.fileno())
            os.replace(name,path)
        finally:
            with suppress(FileNotFoundError):os.unlink(name)


def digest(path):
    if not stat.S_ISREG(path.lstat().st_mode):raise RuntimeError('regular source/binary required')
    return hashlib.sha256(path.read_bytes()).hexdigest()


def redact(value,selected=''):
    if isinstance(value,dict):return {k:redact(v,selected) for k,v in value.items()}
    if isinstance(value,list):return [redact(v,selected) for v in value]
    if isinstance(value,str):
        # Longest paths first, because common root can contain the worktree.
        for path,label in sorted(((str(ROOT),'$REPO'),(str(common_root()),'$COMMON'),
                                  (str(Path.home()),'$HOME'),(os.environ.get('HOME',''),'$HOME')),
                                 key=lambda x:len(x[0]),reverse=True):
            if path:value=value.replace(path,label)
        if selected:value=value.replace(selected,'[SELECTED-CREDENTIAL]')
    return value


def sources():
    names={'go.mod','go.sum'}
    for base,dirs,files in os.walk(ROOT/'ai'):
        dirs[:]=[n for n in dirs if n not in ('node_modules','.evidence','.git')]
        for name in files:
            path=Path(base)/name;rel=path.relative_to(ROOT).as_posix()
            if (name.endswith(('.go','.json')) or rel.startswith('ai/live/testdata/minimax/')) and rel!='ai/live/support-matrix.json':
                names.add(rel)
    names.update(p.relative_to(ROOT).as_posix() for p in Path(__file__).parent.glob('*.py'))
    return {n:digest(ROOT/n) for n in sorted(names)}


def committed_sources():
    commit=git('rev-parse','HEAD');files=sources()
    for name,sha in files.items():
        raw=subprocess.check_output(['git','show',commit+':'+name],cwd=ROOT,env=ENV,stderr=subprocess.DEVNULL)
        if hashlib.sha256(raw).hexdigest()!=sha:raise RuntimeError('execution source must be committed before freeze')
    return dict(sourceCommit=commit,files=files)


def guard(out):
    frozen=doc(out/'source-at-start.json')
    if sources()!=frozen['files'] or git('rev-parse','HEAD')!=frozen['sourceCommit']:
        raise RuntimeError('frozen source commit/set changed')
    for name,sha in doc(out/'binaries.json').items():
        if digest(out/name)!=sha:raise RuntimeError('frozen binary changed')
    if digest(out/'plan.json')!=doc(out/'prepared.json')['planSHA256']:
        raise RuntimeError('frozen collection plan changed')
    for name,sha in doc(out/'prepared.json')['authorizationFiles'].items():
        if digest(out/name)!=sha:raise RuntimeError('frozen component authorization changed')
    return frozen


def reviewed(out,job):
    frozen=guard(out);review=doc(out/'prerelease-source-review.json')
    if (review.get('decision')!='APPROVED' or not review.get('reviewedBy') or
        review.get('sourceCommit')!=frozen['sourceCommit'] or
        review.get('sourceManifestSHA256')!=digest(out/'source-at-start.json') or
        review.get('binariesManifestSHA256')!=digest(out/'binaries.json') or
        review.get('planSHA256')!=digest(out/'plan.json') or
        job['name'] not in review.get('authorizedRoutes',[]) or
        doc(out/'checks.json').get('passed') is not True):
        raise RuntimeError('APPROVED exact-source review and no-key checks required')
    if job.get('purpose')=='mplan' and review.get('credentialScopes',{}).get('MINIMAX_MPLAN_KEY')!='cn-mplan':
        raise RuntimeError('independent M Plan entitlement and credential review required')
    if job.get('purpose')=='standard' and review.get('credentialScopes',{}).get('MINIMAX_KEY')!='cn-standard':
        raise RuntimeError('China standard credential review required')
    return frozen


def selected_secret(variable):
    # No dotenv loading: parse only the independently selected name.
    with (common_root()/'.env').open() as stream:
        for line in stream:
            name,sep,value=line.strip().removeprefix('export ').partition('=')
            if sep and name.strip()==variable:
                parts=shlex.split(value,comments=True)
                if len(parts)==1 and parts[0].strip():return parts[0].strip()
                break
    raise RuntimeError('selected credential unavailable')


def stop(proc):
    with suppress(ProcessLookupError):os.killpg(proc.pid,signal.SIGTERM)
    try:proc.wait(timeout=15)
    except subprocess.TimeoutExpired:pass
    with suppress(ProcessLookupError):os.killpg(proc.pid,signal.SIGKILL)
    proc.wait()


def execute(out,name,argv,env=None,timeout=240,cwd=ROOT,source=None,injected=None,receipt=None):
    """Exclusive intent exists before this function can acquire a selected key."""
    path=out/(name+'-command.json')
    record=dict(state='STARTING',name=name,argv=list(map(str,argv)),cwd=str(cwd),
        timeoutSeconds=timeout,credentialsInherited=False,sourceVariable=source,
        injectedCredential=injected,credentialsRead=False,startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),exitCode=None)
    dump(path,redact(record),exclusive=True)
    key='';child=None;began=time.monotonic();code=1
    childenv=ENV|(env or {})
    with tempfile.TemporaryFile() as capture:
        try:
            if source:
                record['credentialsRead']=True
                key=selected_secret(source);childenv[injected]=key
            record['environmentKeys']=sorted(childenv)
            dump(path,redact(record,key))
            child=subprocess.Popen(list(map(str,argv)),cwd=cwd,env=childenv,stdout=capture,
                stderr=subprocess.STDOUT,start_new_session=True)
            try:code=child.wait(timeout=timeout);record['state']='COMPLETED'
            except subprocess.TimeoutExpired:
                stop(child);code=124;record['state']='TIMEOUT'
        except BaseException as error:
            if child is not None:stop(child)
            record['state']='INTERRUPTED' if isinstance(error,KeyboardInterrupt) else 'ERROR'
            record['exceptionType']=type(error).__name__
        finally:
            capture.seek(0);output=capture.read().decode(errors='replace')
            (out/(name+'.log')).write_text(redact(output,key))
            record.update(exitCode=code,seconds=time.monotonic()-began)
            # Captures are immutable evidence: audit rather than rewrite them.
            findings=[]
            if key:
                for base in ((out/name),(out/'live'/name)):
                    if base.is_dir():
                        for file in base.rglob('*'):
                            if file.is_file() and key.encode() in file.read_bytes():
                                findings.append(str(file.relative_to(out)))
            record['launcherAudit']=dict(passed=not findings,findings=findings)
            if findings:code=1;record['exitCode']=1
            safe=redact(record,key);dump(path,safe)
            if receipt is not None:
                intent=doc(receipt);intent.update(state=safe['state'],exitCode=code,credentialsRead=safe['credentialsRead'],
                    launcherAudit=safe['launcherAudit'],finishedAt=dt.datetime.now(dt.timezone.utc).isoformat())
                dump(receipt,intent)
            childenv.pop(injected,None);key=''
    return code
