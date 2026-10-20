"""Read the one fixed closed first cohort; no research or old authority revival."""
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import stat
import final_contract as r
import renewal_common as u
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


def file_hash(name,remaining=512<<20):
    path=owned(name);before=path.stat(follow_symlinks=False)
    if not stat.S_ISREG(before.st_mode) or before.st_size>min(64<<20,remaining):
        raise RuntimeError('closed cohort file/total byte ceiling exceeded')
    with path.open('rb') as stream:
        opened=os.fstat(stream.fileno())
        if not os.path.samestat(before,opened):raise RuntimeError('closed fact changed before open')
        h=hashlib.sha256();read=0
        while True:
            chunk=stream.read(min(65536,before.st_size-read+1))
            if not chunk:break
            read+=len(chunk)
            if read>before.st_size:raise RuntimeError('closed fact grew during streaming hash')
            h.update(chunk)
        value=h.hexdigest()
    after=path.stat(follow_symlinks=False)
    if not os.path.samestat(before,after) or (before.st_size,before.st_mtime_ns)!=(after.st_size,after.st_mtime_ns):
        raise RuntimeError('closed fact changed during streaming hash')
    return value,before.st_size


def document(name):return u.json_doc(owned(str(name)),2<<20)


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
    if {p.name for p in native.iterdir()}!={'operator-authority.json','launch.lock'} or not lock.is_file() or lock.stat().st_size!=0:
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
        native_base=owned(row['nativeBundle']);report=u.json_doc(native_base/'regional-report.json')
        if not native_base.is_relative_to(owned(str(BASE/'candidate-final'/name))):raise RuntimeError('closed own native directory required')
        raw=u.tree(native_base)
        if any(files.get((native_base/leaf).relative_to(c.common_root()).as_posix())!=u.sha(payload) for leaf,payload in raw.items()):
            raise RuntimeError('entire original native tree must be closed')
        if (row.get('exitCode')!=code or row.get('qualification')!=('PASS' if passed else 'FAIL') or
            row.get('actualCalls')!=18 or row.get('actualHTTPAttempts')!=18 or row.get('claimedCalls')!=24 or row.get('claimedAttempts')!=24 or row.get('wholeReservedMicrousd')!=650000 or
            claim.get('state')!='COMPLETED' or claim.get('exitCode')!=code or claim.get('sourceCommit')!=index['sourceCommit'] or claim.get('reservedMicrousd')!=650000 or
            file_hash(marker)[0]!=row.get('canonicalClaimSHA256') or cmd.get('state')!='COMPLETED' or cmd.get('exitCode')!=code or cmd.get('launcherAudit')!=dict(passed=True,findings=[]) or
            usd.get('round')!=OLD or usd.get('attempts')!=18 or usd.get('maxMicrousd')!=650000 or not 0<usd.get('reservedMicrousd',0)<=650000 or
            report.get('gitCommit')!=index['sourceCommit'] or report.get('budget',{}).get('callsUsed')!=18 or report.get('budget',{}).get('httpAttempts')!=18 or len(report.get('scenarios',[]))!=18 or
            u.json_doc(native_base/'audit.json')!=dict(findings=[])):
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
