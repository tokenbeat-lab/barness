"""Native current3 facts, exact component grants and full-process consumers."""
import datetime as dt
import json
from pathlib import Path
import recovery_contract as r
import recovery_history as h
import recovery_state as s
c = r.c

def facts(out):
    verified(out, 'native-recovery-facts', [out / 'minimaxverify', '-recovery-source'])
    value = json.loads((out / 'native-recovery-facts.log').read_text())
    prior = h.facts()
    recovery = value.get('recovery', {})
    expected = dict(secondCohortSealSha256=prior['secondCohortSealSHA256'], stopIndexSha256=prior['stopIndexSHA256'], secondCalls=13, secondAttempts=12, secondLedgerRows=25, retainedActualCalls=303, retainedActualAttempts=302, retainedClaimedCalls=456, retainedClaimedAttempts=492, retainedReservedMicrousd=10700000)
    if value.get('renewalId') != r.ROUND or value.get('sourceCommit') != c.git('rev-parse', 'HEAD') or len(value.get('closedHistory', {})) != 15 or any((value.get(k) != v for k, v in dict(originalCalls=39, originalAttempts=36, previousCalls=39, previousAttempts=36).items())) or any((recovery.get(k) != v for k, v in expected.items())):
        raise RuntimeError('exact native3 closed1+failed2 history/source required')
    for name, key in [('grant.json', 'secondGrantSha256'), ('human-approval.json', 'secondHumanApprovalSha256'), ('source-review.json', 'secondSourceReviewSha256'), ('chat.json', 'secondChatClaimSha256'), ('formal.jsonl', 'secondLedgerSha256')]:
        if recovery[key] != s.pin(c.common_root() / '.evidence/minimax-renewals' / r.old.ROUND / name):
            raise RuntimeError('native failed2 authority pin changed')
    return value

def create(out):
    import recovery_execute as e
    prepared = s.guard(out)
    for route in r.ROUTES:
        _, auth, _, _ = s.approved(out, r.jobs(out)['mm-global-formal-m3-' + route])
    e.require_checks(out)
    value = facts(out)
    c.dump(out / 'native-recovery-facts.json', value)
    human = dict(decision='APPROVED', renewalId=r.ROUND, validationRound=r.ROUND, nativeRenewalId=r.ROUND, spendingGuardPolicyId=r.SPENDING_GUARD_POLICY_SELECTOR, maxCalls=506, maxAttempts=518, campaignMaxCalls=962, campaignMaxAttempts=1010, newReservedMicrousd=19200000, retainedReservedMicrousd=10700000, campaignMaximumMicrousd=30000000, firstCohortSealSHA256=value['firstCohortSealSha256'], secondCohortSealSHA256=value['recovery']['secondCohortSealSha256'], historySHA256=value['closedHistorySha256'], recoveryHistorySHA256=value['recoveryHistorySha256'], planSHA256=prepared['planSHA256'], requestReference=auth['requestReference'], requestText=auth['requestText'], provider='minimax', model='MiniMax-M3', purpose='standard', routes=list(r.ROUTES), renewalMaxCalls=60, renewalMaxAttempts=72, actualHumanApprovalSHA256=auth['actualHumanApprovalSHA256'])
    review = s.doc(out / 'prerelease-source-review.json') | {k: human[k] for k in ('renewalId', 'historySHA256', 'firstCohortSealSHA256', 'secondCohortSealSHA256', 'recoveryHistorySHA256')}
    with e.locked() as base:
        for leaf in ('human-approval.json', 'source-review.json', 'grant.json'):
            if (base / leaf).exists() or (base / leaf).is_symlink():
                raise RuntimeError('native3 authority already issued; never replace')
        c.dump(base / 'human-approval.json', human, exclusive=True)
        c.dump(base / 'source-review.json', review, exclusive=True)
        grant = {k: v for k, v in value.items() if k not in ('closedHistory', 'recoveryHistorySha256')}
        grant.update(schema=4, decision='APPROVED', maxCalls=60, maxAttempts=72, perIdentityCalls=20, perIdentityAttempts=24, maxOutputTokens=4096, maxRequestedOutput=40000, maxReservedInput=120000, humanApprovalSha256=s.pin(base / 'human-approval.json'), sourceReviewSha256=s.pin(base / 'source-review.json'), sourceManifestSha256=prepared['sourceManifestSHA256'], binariesManifestSha256=prepared['binariesManifestSHA256'], planSha256=prepared['planSHA256'])
        c.dump(base / 'grant.json', grant, exclusive=True)
        e.sync_dir(base)
    for route in r.ROUTES:
        verified(out, 'native-grant-preflight-' + route, [out / 'minimaxverify', '-recovery-preflight=' + route])
    return dict(credentialsRead=False, renewalId=r.ROUND, grantSHA256=s.pin(base / 'grant.json'), maxCalls=60, maxAttempts=72)

def templates(out):
    from recovery_flow import fixed3
    return fixed3().n.templates(out)

def verified(out, name, argv):
    from recovery_flow import fixed3
    return fixed3().n.verified(out,name,argv)

def require_component(out, job, auth, scope):
    from recovery_flow import fixed3
    return fixed3().n.require_component(out,job,auth,scope)

def preflight(out, job, record):
    from recovery_flow import fixed3
    return fixed3().n.preflight(out,job,record)

def consume(out, job, record):
    from recovery_flow import fixed3
    return fixed3().n.consume(out,job,record)

def require_consumed(out, job, value):
    from recovery_flow import fixed3
    return fixed3().n.require_consumed(out,job,value)

# Fixed3 issuance had no preceding regional probe; its historical contract stays so.
def before_claim(out,job):return None
