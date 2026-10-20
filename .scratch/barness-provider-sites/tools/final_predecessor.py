"""Read the one fixed closed first cohort; no research or old authority revival."""
from contextlib import contextmanager
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import stat
import final_contract as r
c=r.c
BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
CLOSED=BASE/'closed-cohort-1'
OLD='provider-sites-final-acceptance-1'


def owned(name):
    parts=PurePosixPath(name).parts
    if not parts or parts[0]!='.evidence' or '..' in parts or '\\' in name or '.env' in parts:
        raise RuntimeError('closed cohort owns only fixed evidence-relative facts')
    path=c.common_root()
    for part in parts:
        path=path/part
        if path.is_symlink():raise RuntimeError('closed cohort may not follow symbolic links')
    return path


@contextmanager
def regular(name,limit):
    path=owned(str(name));before=path.stat(follow_symlinks=False)
    if not stat.S_ISREG(before.st_mode) or before.st_size>limit:
        raise RuntimeError('closed cohort requires a bounded regular file')
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    try:
        opened=os.fstat(fd)
        if (not stat.S_ISREG(opened.st_mode) or not os.path.samestat(before,opened) or
            (before.st_size,before.st_mtime_ns)!=(opened.st_size,opened.st_mtime_ns)):
            raise RuntimeError('closed fact changed before nonblocking open')
        with os.fdopen(fd,'rb',closefd=False) as stream:yield stream,before.st_size
        after=path.stat(follow_symlinks=False)
        if (not stat.S_ISREG(after.st_mode) or not os.path.samestat(before,after) or
            (before.st_size,before.st_mtime_ns)!=(after.st_size,after.st_mtime_ns)):
            raise RuntimeError('closed fact changed during bounded read')
    finally:os.close(fd)


def file_hash(name,remaining=512<<20):
    with regular(name,min(64<<20,remaining)) as (stream,size):
        h=hashlib.sha256();read=0
        while True:
            chunk=stream.read(min(65536,size-read+1))
            if not chunk:break
            read+=len(chunk)
            if read>size:raise RuntimeError('closed fact grew during streaming hash')
            h.update(chunk)
        if read!=size:raise RuntimeError('closed fact truncated during streaming hash')
        value=h.hexdigest()
    return value,size


def document(name):
    # Closed history needs nonblocking identity checks for every JSON read;
    # the older publication importer keeps its independently closed contract.
    with regular(name,2<<20) as (stream,size):
        raw=stream.read(size+1)
        if len(raw)!=size:raise RuntimeError('closed JSON changed during bounded read')
    def pairs(items):
        out={}
        for key,value in items:
            if key in out:raise RuntimeError('duplicate closed JSON field')
            out[key]=value
        return out
    value=json.loads(raw,object_pairs_hook=pairs)
    if not isinstance(value,dict):raise RuntimeError('closed JSON object required')
    return value


def native_files(base):
    # Preserve the existing capture inventory ceilings while hashing through
    # this closure's nonblocking opener, without copying whole binary bodies.
    pending=[(base,0)];paths=[];entries=0
    while pending:
        directory,depth=pending.pop()
        with os.scandir(directory) as scanned:
            for entry in scanned:
                entries+=1
                if entries>1024 or depth+1>3 or entry.is_symlink():raise RuntimeError('closed native tree entry/depth/link refused')
                if entry.is_dir(follow_symlinks=False):pending.append((Path(entry.path),depth+1))
                elif entry.is_file(follow_symlinks=False):
                    paths.append(Path(entry.path))
                    if len(paths)>512:raise RuntimeError('closed native tree file ceiling exceeded')
                else:raise RuntimeError('closed native tree regular file required')
    if not paths:raise RuntimeError('closed native tree may not be empty')
    return paths


def disclosure():
    seal=document(CLOSED/'seal.json');index=document(CLOSED/'index.json')
    counters=dict(campaignActualCalls=40,campaignActualHTTPAttempts=40,campaignClaimedCalls=52,campaignClaimedHTTPAttempts=52,campaignWholeReservedMicrousd=1500000)
    if (seal.get('schema')!=1 or seal.get('kind')!='CLOSED_FIRST_COHORT_HISTORY_SEAL' or seal.get('cohortIndex')!='index.json' or
        seal.get('cohortIndexSHA256')!=file_hash(str(CLOSED/'index.json'))[0] or
        any(type(seal.get(k)) is not int or seal[k]!=v for k,v in counters.items()) or
        seal.get('nativeRound1NeverIssued') is not True or seal.get('remainingPaidScopeRequiresNewApproval') is not True):
        raise RuntimeError('fixed closed first-cohort seal/counters required')
    expected=dict(schema=1,kind='CLOSED_FAILED_REGIONAL_CANDIDATE_COHORT',roundId=OLD,state='CLOSED_NO_FURTHER_PAID_LAUNCH',actualCalls=40,actualHTTPAttempts=40,claimedCalls=52,claimedHTTPAttempts=52,wholeReservedMicrousd=1500000,refundMicrousd=0,noProcessRestarts=True,originalNativeM3HistoryUnchanged=True,paidContinuationAuthorized=False,sourceArchive='source')
    if any(index.get(k)!=v or type(index.get(k)) is not type(v) for k,v in expected.items()):
        raise RuntimeError('honest closed failed first cohort with retained whole reservations required')
    files=index['files']
    if not isinstance(files,dict) or not 1<=len(files)<=4096:raise RuntimeError('bounded complete predecessor pins required')
    total=0
    for name,sha in files.items():
        actual,size=file_hash(name,(512<<20)-total);total+=size
        if actual!=sha:raise RuntimeError('closed first-cohort original changed: '+name)
    archive=owned(str(CLOSED/'source'));selected=set();pending=[archive];count=0
    while pending:
        directory=pending.pop()
        if len(directory.relative_to(c.common_root()).parts)>32:raise RuntimeError('closed archive depth ceiling exceeded')
        with os.scandir(directory) as entries:
            for entry in entries:
                count+=1
                if count>8192 or entry.is_symlink():raise RuntimeError('closed source entry ceiling or symbolic link')
                if entry.is_dir(follow_symlinks=False):pending.append(Path(entry.path))
                else:selected.add(Path(entry.path).relative_to(c.common_root()).as_posix())
    if not selected or selected!={n for n in files if n.startswith(str(CLOSED/'source')+'/')}:
        raise RuntimeError('complete immutable first-cohort control/source archive required')
    used={'candidate-zai-chat','candidate-zai-messages'}
    expected_absent={str(Path('.evidence/minimax-launches')/(OLD+'-'+n+'.json')) for n in r.jobs(c.ROOT/'.evidence/unused-closed-scope') if n not in used}
    if set(index.get('unstartedCanonicalClaims',[]))!=expected_absent or len(index['unstartedCanonicalClaims'])!=31:
        raise RuntimeError('all31 unstarted first-cohort jobs must remain absent')
    for name in expected_absent:
        path=owned(name)
        if path.exists() or path.is_symlink():raise RuntimeError('closed first-cohort job acquired unexpected authority')
    launches=owned('.evidence/minimax-launches')
    if {p.name for p in launches.glob(OLD+'-*.json')}!={OLD+'-'+n+'.json' for n in used}:
        raise RuntimeError('closed cohort canonical launcher set changed')
    native=owned('.evidence/minimax-renewals/'+OLD)
    lock=owned('.evidence/minimax-renewals/'+OLD+'/launch.lock')
    if {p.name for p in native.iterdir()}!={'operator-authority.json','launch.lock'} or file_hash(str(lock.relative_to(c.common_root())))[1]!=0:
        raise RuntimeError('unused native1 has unexpected grant/ledger/nonce/intent')
    wanted_native={str(Path('.evidence/minimax-renewals')/OLD/n) for n in ('grant.json','human-approval.json','source-review.json','formal.jsonl','chat.json','responses.json','messages.json')}
    if set(index.get('nativeMiniMaxRound1Absent',[]))!=wanted_native:raise RuntimeError('explicit native1 non-consumption facts required')
    for name in wanted_native:
        if owned(name).exists():raise RuntimeError('native1 authority appeared after closure')
    rows=index.get('processes',[])
    if len(rows)!=2 or {v['name'] for v in rows}!=used:raise RuntimeError('both original completed first-cohort children required')
    for row in rows:
        name=row['name'];passed=name=='candidate-zai-messages';code=0 if passed else 1
        marker=str(Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json'))
        command=str(BASE/'candidate-final'/(name+'-command.json'));usd_name=str(BASE/'candidate-final'/(name+'-usd.json'))
        if row.get('canonicalClaim')!=marker or row.get('command')!=command or marker not in files or command not in files or usd_name not in files:
            raise RuntimeError('own original fixed parent/command/USD pins required')
        claim=document(marker);cmd=document(command);usd=document(usd_name)
        native_base=owned(row['nativeBundle']);report=document(str((native_base/'regional-report.json').relative_to(c.common_root())))
        if not native_base.is_relative_to(owned(str(BASE/'candidate-final'/name))):raise RuntimeError('closed own native directory required')
        total=0
        for path in native_files(native_base):
            name=path.relative_to(c.common_root()).as_posix();sha,size=file_hash(name,min(16<<20,(128<<20)-total));total+=size
            if files.get(name)!=sha:raise RuntimeError('entire original native tree must be closed')
        if (row.get('exitCode')!=code or row.get('qualification')!=('PASS' if passed else 'FAIL') or
            row.get('actualCalls')!=18 or row.get('actualHTTPAttempts')!=18 or row.get('claimedCalls')!=24 or row.get('claimedAttempts')!=24 or row.get('wholeReservedMicrousd')!=650000 or
            claim.get('state')!='COMPLETED' or claim.get('exitCode')!=code or claim.get('sourceCommit')!=index['sourceCommit'] or claim.get('reservedMicrousd')!=650000 or
            file_hash(marker)[0]!=row.get('canonicalClaimSHA256') or cmd.get('state')!='COMPLETED' or cmd.get('exitCode')!=code or cmd.get('launcherAudit')!=dict(passed=True,findings=[]) or
            usd.get('round')!=OLD or usd.get('attempts')!=18 or usd.get('maxMicrousd')!=650000 or not 0<usd.get('reservedMicrousd',0)<=650000 or
            report.get('gitCommit')!=index['sourceCommit'] or report.get('budget',{}).get('callsUsed')!=18 or report.get('budget',{}).get('httpAttempts')!=18 or len(report.get('scenarios',[]))!=18 or
            document(str((native_base/'audit.json').relative_to(c.common_root())))!=dict(findings=[])):
            raise RuntimeError('closed actual40/40 and claimed52/52 must retain honest child/audit/USD facts')
        failed=[v['id'] for v in report['scenarios'] if v.get('outcome')!='PASS']
        if failed!=row.get('failedScenarios') or (not passed and len(failed)!=1) or passed and failed:
            raise RuntimeError('genuine first-cohort candidate failure may not be promoted')
    history=seal['nativeHistory']
    if (len(history['closedHistory'])!=15 or any(history.get(k)!=v for k,v in dict(originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36).items()) or
        history.get('countsAreHistoricalNotAdditionalCampaignConsumption') is not True):
        raise RuntimeError('separate original15 M3 histories39/36 twice required')
    for name,sha in history['closedHistory'].items():
        if file_hash('.evidence/'+name)[0]!=sha:raise RuntimeError('original M3 history changed')
    for key in ('facts','command'):
        path=str(CLOSED/history[key])
        if file_hash(path)[0]!=history[key+'SHA256']:raise RuntimeError('closed native history receipt changed')
    if document(CLOSED/history['command']).get('exitCode')!=0:raise RuntimeError('closed native history verification was not successful')
    approval_name=str(BASE/'authorization.json');approval=document(approval_name)
    scopes_name=str(BASE/'candidate-final/execution-authorization.json')
    scope=approval.get('approvedScope',{})
    if scope.get('budgetUSD')!=30 or scope.get('nineProviderStandardAPITestAccountEligibilityConfirmed') is not True or not scope.get('race'):
        raise RuntimeError('original30USD, nine-service eligibility and unique final race approval required')
    if approval.get('decision')!='APPROVED' or approval.get('roundId')!=OLD or files.get(approval_name)!=file_hash(approval_name)[0] or scopes_name not in files:
        raise RuntimeError('original eligibility/race approval must remain immutable')
    return dict(round=OLD,state='CLOSED',sealSHA256=file_hash(str(CLOSED/'seal.json'))[0],indexSHA256=seal['cohortIndexSHA256'],actualCalls=40,actualAttempts=40,claimedCalls=52,claimedAttempts=52,reservedMicrousd=1500000,
        authorizationSHA256=files[approval_name],credentialScopes=document(scopes_name)['standardCredentialScopes'],nativeHistorySHA256=history['closedHistorySHA256'],nativeRound1Unused=True)
