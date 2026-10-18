"""Reopen launcher-pinned historical bytes without requiring current HEAD."""
import json
import os
from pathlib import PurePosixPath
import re
import stat
import paid_core as c
import paid
import paid_plan as p
import routed_contract as r
import renewal_common as u

HEX=re.compile('[a-f0-9]{64}')


def location(out):
    return dict(rootSHA256=u.sha(str(c.ROOT).encode()),relativeRoot=out.relative_to(c.ROOT).as_posix())


def parts(value):
    path=PurePosixPath(value)
    if not isinstance(value,str) or path.is_absolute() or not value or any(p in ('','.','..') for p in value.split('/')):
        raise ValueError('owned relative artifact path required')
    return path.parts


class Root:
    """Every read uses the original directory FD, with no symlink component."""
    def __init__(self,path):self.path=path;self.fd=None
    def __enter__(self):
        self.fd=os.open(self.path,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW)
        self.identity=os.fstat(self.fd);return self
    def __exit__(self,*error):
        try:
            current=os.stat(self.path,follow_symlinks=False)
            if (current.st_dev,current.st_ino)!=(self.identity.st_dev,self.identity.st_ino):
                raise ValueError('original evidence root changed')
        finally:os.close(self.fd)
    def open(self,value,directory=False):
        names=parts(value);fd=os.dup(self.fd)
        try:
            for index,name in enumerate(names):
                flags=os.O_RDONLY|os.O_NOFOLLOW
                if directory or index<len(names)-1:flags|=os.O_DIRECTORY
                else:flags|=os.O_NONBLOCK
                next_fd=os.open(name,flags,dir_fd=fd);os.close(fd);fd=next_fd
            result=fd;fd=None;return result
        finally:
            if fd is not None:os.close(fd)
    def read(self,value,limit=2<<20):
        with os.fdopen(self.open(value),'rb') as stream:
            info=os.fstat(stream.fileno())
            if not stat.S_ISREG(info.st_mode) or info.st_size>limit:raise ValueError('bounded regular artifact required')
            raw=stream.read(limit+1)
            if len(raw)>limit:raise ValueError('artifact byte limit exceeded')
            return raw
    def unique(self,directory,leaf):
        fd=self.open(directory,directory=True);found=[]
        try:
            with os.scandir(fd) as entries:
                for count,entry in enumerate(entries):
                    if count>=1024 or entry.is_symlink():raise ValueError('invalid bounded bundle entries')
                    if entry.is_dir(follow_symlinks=False):
                        path=directory+'/'+entry.name+'/'+leaf
                        try:raw=self.read(path,16<<20)
                        except FileNotFoundError:continue
                        found.append(raw)
                        if len(found)>1:raise ValueError('ambiguous original report')
        finally:os.close(fd)
        if len(found)!=1:raise ValueError('original report missing')
        return found[0]


def document(raw):
    def pairs(items):
        result={}
        for key,value in items:
            if key in result:raise ValueError('duplicate artifact field')
            result[key]=value
        return result
    value=json.loads(raw,object_pairs_hook=pairs)
    if not isinstance(value,dict):raise ValueError('artifact object required')
    return value


def settled(job):
    """Unknown retains its whole cap; stored budget counters are never proof."""
    unknown=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
    try:
        marker=paid.marker(job['name'])
        with Root(c.common_root()) as common:
            receipt=document(common.read(marker.relative_to(c.common_root()).as_posix()))
        phase='glm' if job['baseName'].startswith('glm-') else 'legacy'
        if (receipt.get('state')!='COMPLETED' or receipt.get('route')!=job['name'] or
            receipt.get('validationRound')!=r.ROUND or receipt.get('phase')!=phase or receipt.get('automaticReruns') is not False or
            receipt.get('credentialsRead') is not True or receipt.get('launcherAudit',{}).get('passed') is not True or
            receipt.get('collectionOwnerSHA256')!=location(c.ROOT)['rootSHA256'] or
            any(receipt.get(k)!=job[k] for k in ('baseName','sourceVariable','maxCalls','maxAttempts','networkRoute'))):
            return unknown
        root=receipt.get('collectionRoot','')
        if not isinstance(root,str) or not root.startswith('$REPO/'):return unknown
        relative=root.removeprefix('$REPO/');parts(relative)
        with Root(c.ROOT) as evidence:
            def pinned(leaf,field):
                raw=evidence.read(relative+'/'+leaf)
                if HEX.fullmatch(receipt.get(field,'')) is None or u.sha(raw)!=receipt[field]:
                    raise ValueError('original pinned metadata changed')
                return document(raw)
            source=pinned('source-at-start.json','sourceManifestSha256')
            pinned('binaries.json','binariesManifestSha256')
            plan=pinned('plan.json','planSha256')
            review=pinned('prerelease-source-review.json','sourceReviewSha256')
            authorization=pinned('execution-authorization.json','authorizationSha256')
            command=pinned(job['baseName']+'-command.json','commandReceiptSha256')
            pins=dict(sourceCommit=source['sourceCommit'],sourceManifestSHA256=receipt['sourceManifestSha256'],
                binariesManifestSHA256=receipt['binariesManifestSha256'],planSHA256=receipt['planSha256'])
            expected=dict(decision='APPROVED',authority='explicit-user-request',requestReference=authorization.get('requestReference'),
                requestText=r.REQUEST,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,**pins,
                checksSHA256=u.sha(evidence.read(relative+'/checks.json')),sourceReviewSHA256=receipt['sourceReviewSha256'])
            if (authorization!=expected or not isinstance(authorization.get('requestReference'),str) or not authorization['requestReference'].strip() or
                source['sourceCommit']!=receipt.get('sourceCommit') or review.get('decision')!='APPROVED' or not review.get('reviewedBy') or
                any(review.get(k)!=v for k,v in pins.items()) or job['name'] not in review.get('authorizedRoutes',[]) or
                review.get('credentialScopes',{}).get(job['sourceVariable'])!=r.ROUND+'/'+job['networkRoute'] or
                plan.get('round')!=r.ROUND or plan.get('phase')!=phase or plan.get('budget')!=r.BUDGET or
                plan.get('collection')!=dict(rootSHA256=receipt['collectionOwnerSHA256'],relativeRoot=relative)):
                return unknown
            checks=document(evidence.read(relative+'/checks.json'));planned=plan['routes'][job['name']]
            if (checks.get('passed') is not True or command.get('state')!='COMPLETED' or
                command.get('name')!=job['baseName'] or command.get('exitCode')!=receipt.get('exitCode') or
                command.get('credentialsRead') is not True or command.get('launcherAudit',{}).get('passed') is not True or
                command.get('argv')!=planned['argv'] or command.get('cwd')!=planned['cwd'] or
                any(planned.get(k)!=job[k] for k in ('baseName','sourceVariable','maxCalls','maxAttempts','networkRoute'))):
                return unknown
            base=job['baseName'];directory=relative+('/live/' if base in {v[0] for v in p.GENERIC} else '/')+base
            leaf=('evaluation-report.json' if base.startswith('text-') else 'multimodal-report.json' if base=='protocol114' else
                'formal-glm-report.json' if base.startswith('glm-') else 'live-report.json')
            raw=evidence.read(directory+'/report.json',16<<20) if base.startswith('visual-') else evidence.unique(directory,leaf)
            if HEX.fullmatch(receipt.get('reportedFileSha256','')) is None or u.sha(raw)!=receipt['reportedFileSha256']:
                return unknown
            return native_counters(document(raw),job)
    except (OSError,RuntimeError,ValueError,KeyError,TypeError,AttributeError):return unknown


def native_counters(report,job):
    budget=report['budget'];base=job['baseName']
    if base=='protocol114':budget=budget['actual']
    fields=('calls','attempts') if base.startswith(('text-','visual-')) or base=='protocol114' else ('callsUsed','httpAttempts')
    calls,attempts=(budget[key] for key in fields);status=paid.report_status(report,job|dict(name=base))
    if (type(calls) is not int or type(attempts) is not int or not 0<=calls<=job['maxCalls'] or not 0<=attempts<=job['maxAttempts'] or
        status not in ('PASS','FAIL','NOT_RUN','UNSUPPORTED') or (status=='NOT_RUN' and (calls or attempts))):
        raise ValueError('invalid native status or counters')
    safe={k:v for k,v in budget.items() if k in ('calls','attempts','callsUsed','httpAttempts','reservedAttempts','attemptsReserved')}
    if any(type(v) is not int or v<0 for v in safe.values()):raise ValueError('invalid reservation counters')
    return dict(state='NOT_RUN' if status=='NOT_RUN' else 'REPORTED',status=status,calls=calls,attempts=attempts,currentReportBudget=safe)
