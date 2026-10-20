"""Issue only this approved round's fixed native60/72 M3 grant, keylessly."""
import json
import final_contract as r
import final_state as s
import final_execute as e
c=r.c


def create(out):
    prepared=s.guard(out)
    if prepared['phase']!='remaining':
        raise RuntimeError('new native M3 grant requires the final remaining-source freeze')
    for route in r.ROUTES:
        job=r.jobs(out)['mm-global-formal-m3-'+route]
        _,auth,_,_=s.approved(out,job)
    e.require_checks(out)
    if s.run(out,'native-renewal-facts',[out/'minimaxverify','-renewal-source'],timeout=120)!=0:
        raise RuntimeError('complete native closed-history/current-source facts refused')
    facts=json.loads((out/'native-renewal-facts.log').read_text())
    if (facts.get('renewalId')!=r.ROUND or facts.get('sourceCommit')!=prepared['sourceCommit'] or
        facts.get('originalCalls')!=39 or facts.get('originalAttempts')!=36 or facts.get('previousCalls')!=39 or facts.get('previousAttempts')!=36 or
        len(facts.get('closedHistory',{}))!=15 or facts.get('firstCohortSealSha256')!=auth['predecessor']['sealSHA256'] or
        facts.get('firstCohortActualCalls')!=40 or facts.get('firstCohortActualAttempts')!=40 or facts.get('firstCohortClaimedCalls')!=52 or facts.get('firstCohortClaimedAttempts')!=52 or facts.get('firstCohortReservedMicrousd')!=1500000 or facts.get('nativeRound1NeverIssued') is not True):
        raise RuntimeError('native exact original+closed previous39/36 histories required')
    c.dump(out/'native-renewal-facts.json',facts)
    # Human/source decisions are reused only from this concrete approved
    # source and actual user text. Adding the independently measured history
    # pin is a projection, never a new permission or a modified old authority.
    human=dict(decision='APPROVED',renewalId=r.ROUND,maxCalls=r.MAX_CALLS,maxAttempts=r.MAX_ATTEMPTS,
        requestReference=auth['requestReference'],requestText=auth['requestText'],historySHA256=facts['closedHistorySha256'],
        planSHA256=prepared['planSHA256'],provider='minimax',model='MiniMax-M3',purpose='standard',routes=list(r.ROUTES),
        renewalMaxCalls=60,renewalMaxAttempts=72,firstCohortSealSHA256=facts['firstCohortSealSha256'],
        campaignMaxCalls=r.CAMPAIGN_CALLS,campaignMaxAttempts=r.CAMPAIGN_ATTEMPTS,newReservedMicrousd=r.NEW_RESERVED)
    # Native schema3 retains the already-approved R2 sub-scope fields. The
    # separate one-process reply changes whole-campaign totals, not M3 caps.
    import final_retry,final_retry_execute
    if final_retry_execute.present():
        if not final_retry_execute.success():raise RuntimeError('the approved whole22 retry must pass before remaining native authority')
        amendment=final_retry.facts()
        human.update(scopeBasis='Original R2 sub-scope only:722/758 and774/810; separate expressly approved Responses retry adds24/32 globally.',
            separateResponsesRetryApprovalSHA256=amendment['approvalSHA256'],actualWholeCampaignMaxCalls=798,actualWholeCampaignMaxAttempts=842)
    review=c.doc(out/'prerelease-source-review.json')|dict(historySHA256=facts['closedHistorySha256'],firstCohortSealSHA256=facts['firstCohortSealSha256'])
    with e.locked() as base:
        for leaf in ('human-approval.json','source-review.json','grant.json'):
            if (base/leaf).exists() or (base/leaf).is_symlink():
                raise RuntimeError('new native grant may be issued only once; preserve existing authority')
        c.dump(base/'human-approval.json',human,exclusive=True)
        c.dump(base/'source-review.json',review,exclusive=True)
        grant={k:v for k,v in facts.items() if k!='closedHistory'}
        grant.update(schema=3,decision='APPROVED',maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,
            maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000,
            humanApprovalSha256=c.digest(base/'human-approval.json'),sourceReviewSha256=c.digest(base/'source-review.json'),
            sourceManifestSha256=prepared['sourceManifestSHA256'],binariesManifestSha256=prepared['binariesManifestSHA256'],planSha256=prepared['planSHA256'])
        c.dump(base/'grant.json',grant,exclusive=True)
    for route in r.ROUTES:
        if s.run(out,'native-grant-preflight-'+route,[out/'minimaxverify','-renewal-preflight='+route],timeout=120)!=0:
            raise RuntimeError('new native grant preflight refused; preserve failure without reissue')
    return dict(credentialsRead=False,renewalId=r.ROUND,grantSHA256=c.digest(base/'grant.json'),
        historySHA256=facts['closedHistorySha256'],maxCalls=60,maxAttempts=72)
