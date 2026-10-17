#!/usr/bin/env python3
"""Selected-secret subprocess isolation and finite command receipts."""
from contextlib import suppress
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import shlex
import signal
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[4]
COMMON = Path(subprocess.check_output(['git','rev-parse','--git-common-dir'],cwd=ROOT,text=True).strip())
CREDENTIAL_FILE = (ROOT/COMMON).resolve().parent/'.env'
ENV = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}


def doc(path):
    return json.loads(path.read_text())


def dump(path, value):
    path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def sources():
    # Jev protocol intentionally pins every ai Go source. Include all tracked
    # execution/config fixtures too; matrix is generated output, not runtime.
    # Renew GLM capture data first, then prepare a new legacy source freeze.
    names = subprocess.check_output(['git','ls-files','ai'],cwd=ROOT,text=True).splitlines()
    names = [n for n in names if n.endswith(('.go','.json')) and n!='ai/live/support-matrix.json']
    names += ['go.mod','go.sum']
    execution = Path(__file__).resolve().parent
    names += [str((execution/name).relative_to(ROOT)) for name in
        ('commands.py','run.py','glm_launcher.py','rollback.py','rollback-probe.txt')]
    return {name:digest(ROOT/name) for name in sorted(names)}


def bundle(base):
    found = list(base.glob('*/manifest.json'))
    if len(found)!=1:
        raise RuntimeError('unique recorder package missing')
    return found[0].parent


def secret(variable):
    with CREDENTIAL_FILE.open() as stream:
        for line in stream:
            name,sep,value=line.strip().removeprefix('export ').partition('=')
            if sep and name.strip()==variable:
                parts=shlex.split(value,comments=True)
                if len(parts)==1 and parts[0].strip():
                    return parts[0].strip()
                break
    raise RuntimeError('selected credential variable unavailable')


def stop(proc):
    with suppress(ProcessLookupError):
        os.killpg(proc.pid,signal.SIGTERM)
    try:
        proc.wait(timeout=15)
    except subprocess.TimeoutExpired:
        pass
    with suppress(ProcessLookupError):
        os.killpg(proc.pid,signal.SIGKILL)
    proc.wait()


class Commands:
    def __init__(self,out):
        self.out=ROOT/out

    def execute(self,name,argv,extra=None,timeout=180,cwd=ROOT,source=None,injected=None):
        receipt=self.out/(name+'-command.json')
        # Exclusive intent: a process with an unknown outcome cannot resend.
        with receipt.open('x') as stream:
            json.dump({'state':'STARTING'},stream)
        selected='';env=ENV|(extra or {});proc=None;began=time.monotonic()
        record=dict(name=name,argv=[str(v).replace(str(ROOT),'$REPO') for v in argv],
            cwd=str(cwd.relative_to(ROOT)) or '.',environmentKeys=sorted(env),
            credentialsInherited=False,sourceVariable=source,injectedCredential=injected,
            timeoutSeconds=timeout,startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),
            state='STARTING',exitCode=None)
        with tempfile.TemporaryFile() as captured:
            try:
                if source:
                    if source not in ('OPENAI_KEY','ANTHROPIC_KEY','GEMINI_KEY','DEEPSEEK_KEY','TYPESAFE_KEY') or not injected:
                        raise RuntimeError('selected source/injection not authorized')
                    selected=secret(source);env[injected]=selected
                    record['environmentKeys']=sorted(env)
                dump(receipt,record)
                print(name+': started',flush=True)
                proc=subprocess.Popen([str(v) for v in argv],cwd=cwd,env=env,
                    stdout=captured,stderr=subprocess.STDOUT,start_new_session=True)
                try:
                    record.update(exitCode=proc.wait(timeout=timeout),state='COMPLETED')
                except subprocess.TimeoutExpired:
                    stop(proc);record.update(exitCode=124,state='TIMEOUT',terminatedProcessExitCode=proc.returncode)
            except BaseException as error:
                if proc is not None:
                    stop(proc)
                record.update(state='INTERRUPTED' if isinstance(error,KeyboardInterrupt) else 'ERROR',
                    exceptionType=type(error).__name__,terminatedProcessExitCode=proc.returncode if proc else None)
                raise
            finally:
                captured.seek(0);output=captured.read()
                if selected:
                    output=output.replace(selected.encode(),b'[SELECTED-CREDENTIAL]')
                output=output.replace(str(ROOT).encode(),b'$REPO').replace(str(Path.home()).encode(),b'$HOME')
                (self.out/(name+'.log')).write_bytes(output)
                record['seconds']=time.monotonic()-began;dump(receipt,record)
                env.pop(injected,None);selected=''
        print(name+': exit '+str(record['exitCode']),flush=True)
        return record['exitCode']
