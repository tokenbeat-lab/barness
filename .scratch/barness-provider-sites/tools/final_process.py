"""One fixed selected-secret child, with exactly-one .env definition policy.

Closed launchers keep their historical parser. This finite round deliberately
owns a stricter acquisition boundary; no closed authority or parser is edited.
"""
import datetime as dt
from pathlib import Path
import shlex
import subprocess
import tempfile
import time
import final_contract as r
c=r.c


def selected(variable):
    allowed={job['sourceVariable'] for job in r.jobs(c.ROOT/'.evidence/unused-secret-scope').values()}
    if variable not in allowed:
        raise RuntimeError('fixed selected credential variable required')
    values=[]
    with (c.common_root()/'.env').open() as stream:
        for line in stream:
            name,sep,value=line.strip().removeprefix('export ').partition('=')
            if sep and name.strip()==variable:
                parts=shlex.split(value,comments=True)
                if len(parts)!=1 or not parts[0].strip():
                    raise RuntimeError('selected credential configuration invalid: '+variable)
                values.append(parts[0].strip())
    if len(values)!=1:
        raise RuntimeError('exactly one selected credential definition required: '+variable)
    return values[0]


def run(out,job,argv,env,intent):
    if not intent.is_file():
        raise RuntimeError('canonical finite claim must exist before selected credential acquisition')
    path=out/(job['name']+'-command.json')
    record=dict(state='STARTING',name=job['name'],argv=list(map(str,argv)),cwd=job['cwd'],
        timeoutSeconds=job['timeoutSeconds'],credentialsInherited=False,sourceVariable=job['sourceVariable'],
        injectedCredential=job['injectedCredential'],credentialsRead=False,startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),exitCode=None)
    c.dump(path,c.redact(record),exclusive=True)
    key='';child=None;code=1;began=time.monotonic();childenv=c.ENV|env
    with tempfile.TemporaryFile() as captured:
        try:
            record['credentialsRead']=True
            key=selected(job['sourceVariable']);childenv[job['injectedCredential']]=key
            record['environmentKeys']=sorted(childenv);c.dump(path,c.redact(record,key))
            child=subprocess.Popen(list(map(str,argv)),cwd=job['cwd'],env=childenv,stdout=captured,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                code=child.wait(timeout=job['timeoutSeconds']);record['state']='COMPLETED'
            except subprocess.TimeoutExpired:
                c.stop(child);code=124;record['state']='TIMEOUT'
        except BaseException as error:
            if child is not None:c.stop(child)
            record['state']='INTERRUPTED' if isinstance(error,KeyboardInterrupt) else 'ERROR'
            record['exceptionType']=type(error).__name__
        finally:
            captured.seek(0);raw=captured.read((16<<20)+1)
            if len(raw)>16<<20:
                code=1;record['state']='LOG_LIMIT';raw=raw[:16<<20]
            (out/(job['name']+'.log')).write_text(c.redact(raw.decode(errors='replace'),key))
            findings=[]
            if key:
                from renewal_common import file_bytes
                for base in (out/job['name'],out/'live'/job['name']):
                    if not base.is_dir():continue
                    for file in base.rglob('*'):
                        if not file.is_file():continue
                        try:
                            if key.encode() in file_bytes(file):findings.append(str(file.relative_to(out)))
                        except (OSError,RuntimeError):findings.append(str(file.relative_to(out)))
            if findings:code=1
            record.update(exitCode=code,seconds=time.monotonic()-began,finishedAt=dt.datetime.now(dt.timezone.utc).isoformat(),
                launcherAudit=dict(passed=not findings,findings=findings))
            safe=c.redact(record,key);c.dump(path,safe)
            value=c.doc(intent);value.update(state=safe['state'],exitCode=code,credentialsRead=safe['credentialsRead'],
                launcherAudit=safe['launcherAudit'],finishedAt=safe['finishedAt']);c.dump(intent,value)
            childenv.pop(job['injectedCredential'],None);key=''
    return code
