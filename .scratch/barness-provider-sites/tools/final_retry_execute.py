"""Durable accounting for the one fixed Responses retry, with no refunds."""
import os
from pathlib import Path
import final_retry as a
import final_contract as r
import final_state as s
c=r.c

def present():
    import final_execute as e
    return e.marker(a.NAME).exists()

def success():
    if not present():return False
    value=a.continuation()
    if value.get('state')!='COMPLETED' or value.get('exitCode')!=0 or value.get('budget')!=dict(state='REPORTED',status='PASS',calls=22,attempts=22) or not value.get('nativeConsumerReceiptSha256'):
        return False
    import final_native
    final_native.publication_receipt('responses')
    return True

def claim(out,job,prepared,auth,reserve):
    import final_execute as e
    ctx=a.load(out);facts=a.facts()
    with e.locked():
        if e.marker(a.NAME).exists():raise RuntimeError('the expressly approved retry has already been claimed')
        summary=e.summarize()
        expected={'candidate-'+combo for combo in r.REGIONAL}|{'glm-chat','glm-responses'}
        if {v['name'] for v in summary['routes']}!=expected or summary['reservedMicrousd']!=6500000 or reserve!=500000:
            raise RuntimeError('original eight claims and6.5USD remain occupied before one retry')
        record=dict(state='STARTING',route=r.ROUND+'-'+a.NAME,baseName=a.NAME,validationRound=r.ROUND,phase='glm',
            sourceCommit=ctx['measuredCommit'],controlSourceCommit=prepared['sourceCommit'],executionContextSHA256=a.local_pin(out,'execution-context.json'),
            amendmentFacts=facts,sourceVariable='ZHIPU_KEY',credentialsRead=False,maxCalls=24,maxAttempts=32,reservedMicrousd=500000,
            automaticReruns=False,collectionRoot=str(out),networkRoute='direct',sourceManifestSha256=prepared['sourceManifestSHA256'],
            binariesManifestSha256=prepared['binariesManifestSHA256'],planSha256=prepared['planSHA256'],
            sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),authorizationSha256=a.local_pin(out,'retry-authorization.json'),
            authorizationFile='retry-authorization.json',budget=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None))
        c.dump(e.marker(a.NAME),c.redact(record),exclusive=True)
        fd=os.open(e.marker(a.NAME).parent,os.O_RDONLY)
        try:os.fsync(fd)
        finally:os.close(fd)
        return e.marker(a.NAME)

def can_continue(rows):
    failed=[v for v in rows if v.get('state')!='COMPLETED' or v.get('budget',{}).get('status')!='PASS']
    return not failed or len(failed)==1 and failed[0]['name']=='glm-responses' and success()

def effective(rows):
    if not present():return rows
    if not success():raise RuntimeError('one whole22 retry must pass before dependency substitution')
    return [v for v in rows if v['name']!='glm-responses']

def approved(out,prepared,job):
    auth=a.approved(out,prepared)
    # The original machine authorization supplies the unchanged service/account
    # qualification, while the separate actual reply grants only this retry.
    ordinary=dict(job,name='glm-messages')
    _,base,scope,_=s.approved_metadata(out,ordinary,prepared)
    review=c.doc(out/'prerelease-source-review.json')
    if job['name'] not in review.get('authorizedRoutes',[]):raise RuntimeError('retry requires exact current independent source review')
    return prepared,auth,scope,500000
