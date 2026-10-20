"""Import the already consumed one-shot raw research; never starts a helper."""
from pathlib import Path, PurePosixPath
import final_contract as r
import renewal_common as u
c=r.c


def disclosure():
    base=c.common_root()/'.evidence/provider-sites/final-acceptance-20261020'
    index_path=base/'research-index.json'
    if not index_path.exists():
        return dict(state='NOT_RUN',calls=0,attempts=0,reservedMicrousd=200000,automaticResearchRestart=False)
    index=u.json_doc(index_path);authorization=u.json_doc(base/'authorization.json');command=u.json_doc(base/'research-command.receipt.json')
    approval_sha=c.digest(base/'authorization.json')
    if (authorization.get('decision')!='APPROVED' or authorization.get('roundId')!=r.ROUND or
        index.get('calls')!=4 or index.get('httpAttempts')!=4 or index.get('retries')!=0 or
        index.get('classification')!='RAW_OWN_WIRE_RESEARCH_NOT_QUALIFICATION' or index.get('qualificationPassClaimed') is not False or
        index.get('keyAudit',{}).get('findings')!=0 or index.get('budgetUSDReserved')!=.2 or
        type(command.get('exitCode')) is not int or command['exitCode']!=0 or command.get('authorizationSha256')!=approval_sha):
        raise RuntimeError('original four-call research/approval/audit facts changed; no research restart')
    for name,sha in index['files'].items():
        parts=PurePosixPath(name).parts
        if not parts or '..' in parts or PurePosixPath(name).is_absolute():
            raise RuntimeError('owned original research artifact path required')
        path=base
        for part in parts:
            path=path/part
            if path.is_symlink():raise RuntimeError('research original may not be a symlink')
        if u.sha(u.file_bytes(path))!=sha:
            raise RuntimeError('immutable original research artifact changed')
    common=Path(c.git('rev-parse','--git-common-dir'))
    if not common.is_absolute():common=c.ROOT/common
    claim=common.resolve()/'barness-regional-research'/(authorization['researchPlanSha256']+'.json')
    original=u.json_doc(claim)
    if original.get('planSha256')!=authorization['researchPlanSha256'] or original.get('authorizationSha256')!=approval_sha:
        raise RuntimeError('original one-shot common Git research claim unavailable')
    return dict(state='COMPLETED',calls=4,attempts=4,reservedMicrousd=200000,automaticResearchRestart=False,
        authorizationSHA256=approval_sha,indexSHA256=c.digest(index_path),commandReceiptSHA256=c.digest(base/'research-command.receipt.json'),
        originalCommonClaimSHA256=c.digest(claim),artifactFiles=index['files'])
