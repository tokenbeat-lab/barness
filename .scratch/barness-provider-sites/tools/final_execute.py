"""Durable one-use claims, conservative USD reservation and isolated child runs."""
from contextlib import contextmanager
from decimal import Decimal
import fcntl
import os
from pathlib import Path
import final_contract as r
import final_state as s
c = r.c


def marker(name):
    return c.common_root() / '.evidence/minimax-launches' / (r.ROUND + '-' + name + '.json')


def authority_root():
    return c.common_root() / '.evidence/minimax-renewals' / r.ROUND


@contextmanager
def locked():
    root = authority_root()
    root.mkdir(parents=True, exist_ok=True, mode=0o700)
    for p in (root, root.parent, root.parent.parent):
        if p.is_symlink():
            raise RuntimeError('canonical authority path may not contain symlinks')
    lock = root / 'launch.lock'
    if lock.is_symlink():
        raise RuntimeError('canonical lock may not be a symlink')
    with lock.open('a+') as stream:
        fcntl.flock(stream.fileno(), fcntl.LOCK_EX)
        yield root


def summarize():
    import final_predecessor
    prior=final_predecessor.disclosure()
    auth_path=authority_root()/'operator-authority.json'
    auth=c.doc(auth_path) if auth_path.exists() else None
    if auth is not None and auth.get('predecessor')!=prior:
        raise RuntimeError('fixed prior cohort changed after new authority')
    rows=[]
    for name,job in r.jobs(c.ROOT/'.evidence/unused-final-summary').items():
        p=marker(name)
        if not p.exists():continue
        receipt=c.doc(p)
        if receipt.get('validationRound')!=r.ROUND or receipt.get('baseName')!=name:
            raise RuntimeError('canonical fixed-round claim changed')
        rows.append(dict(name=name,reservedMicrousd=receipt['reservedMicrousd'],maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],state=receipt.get('state'),budget=receipt.get('budget')))
    base_calls=sum(v['maxCalls'] for v in rows);base_attempts=sum(v['maxAttempts'] for v in rows)
    import final_retry,final_retry_execute
    approved_amendment=final_retry.facts() if (c.common_root()/final_retry.APPROVAL).exists() else None
    amendment=final_retry.continuation() if final_retry_execute.present() else None
    if amendment is not None:
        rows.append(dict(name=final_retry.NAME,reservedMicrousd=amendment['reservedMicrousd'],maxCalls=24,maxAttempts=32,state=amendment['state'],budget=amendment['budget']))
    own_calls=sum(v['maxCalls'] for v in rows);own_attempts=sum(v['maxAttempts'] for v in rows)
    own_usd=sum(v['reservedMicrousd'] for v in rows)
    return dict(round=r.ROUND,predecessor=prior,paidApproved=auth is not None,maxMicrousd=30000000,
        newReservedMicrousd=own_usd,reservedMicrousd=prior['reservedMicrousd']+own_usd,
        ownClaimedCalls=own_calls,ownClaimedAttempts=own_attempts,baseClaimedCalls=base_calls,baseClaimedAttempts=base_attempts,
        amendmentApproved=approved_amendment is not None,amendmentClaimed=amendment is not None,
        campaignMaxCalls=798 if approved_amendment is not None else r.CAMPAIGN_CALLS,campaignMaxAttempts=842 if approved_amendment is not None else r.CAMPAIGN_ATTEMPTS,
        claimedCalls=prior['claimedCalls']+own_calls,claimedAttempts=prior['claimedAttempts']+own_attempts,
        originalFailedKnownCalls=22 if approved_amendment is not None else 0,originalFailedKnownAttempts=22 if approved_amendment is not None else 0,
        knownReportedCalls=prior['actualCalls']+(22 if approved_amendment is not None else 0)+sum((v.get('budget') or {}).get('calls') or 0 for v in rows),
        knownReportedAttempts=prior['actualAttempts']+(22 if approved_amendment is not None else 0)+sum((v.get('budget') or {}).get('attempts') or 0 for v in rows),
        unknownProcesses=sum((v.get('budget') or {}).get('state')=='UNKNOWN' for v in rows),routes=rows,credentialsRead=False,
        basis='Original R2 sub-scope722/758 and774/810 remains independent; expressly approved retry adds24/32 for798/842. Whole prior6.5USD stays occupied; .5 retry plus .5 unstarted Messages keeps total29.9 within30USD.' if approved_amendment is not None else 'Prior40/40 actual and52/52 claimed plus entire1.5USD stay closed; new722/758 and28.4USD, aggregate774/810 and at most30USD; no refund or process rerun.')


def claim(out, job, prepared, auth, reserve):
    import final_retry,final_retry_execute
    if job['name']==final_retry.NAME:return final_retry_execute.claim(out,job,prepared,auth,reserve)
    import final_predecessor
    prior=c.doc(out/'predecessor-disclosure.json')
    if prior!=final_predecessor.disclosure() or auth.get('predecessor')!=prior:
        raise RuntimeError('fixed closed original cohort required before new claim')
    identity=dict(round=r.ROUND,requestText=auth['requestText'],requestReference=auth['requestReference'],
        maxMicrousd=int(Decimal(str(auth['maxUSD']))*1000000),newMaxMicrousd=r.NEW_RESERVED,
        maxCalls=r.MAX_CALLS,maxAttempts=r.MAX_ATTEMPTS,campaignMaxCalls=r.CAMPAIGN_CALLS,campaignMaxAttempts=r.CAMPAIGN_ATTEMPTS,
        eligibilityAuthorizationSHA256=auth['eligibilityAuthorizationSHA256'],raceAuthorizationSHA256=auth['raceAuthorizationSHA256'],
        automaticProcessReruns=False,predecessor=prior)
    with locked() as root:
        proof = root / 'operator-authority.json'
        if proof.exists():
            if c.doc(proof) != identity:
                raise RuntimeError('new stage must bind the same finite human round, not a replacement authority')
        else:
            c.dump(proof, identity, exclusive=True)
        summary = summarize()
        if (summary['reservedMicrousd']+reserve>identity['maxMicrousd'] or summary['newReservedMicrousd']+reserve>r.NEW_RESERVED or
            summary['baseClaimedCalls']+job['maxCalls']>r.MAX_CALLS or summary['baseClaimedAttempts']+job['maxAttempts']>r.MAX_ATTEMPTS or
            summary['claimedCalls']+job['maxCalls']>summary['campaignMaxCalls'] or summary['claimedAttempts']+job['maxAttempts']>summary['campaignMaxAttempts']):
            raise RuntimeError('remaining round USD/call/attempt reservation exhausted before key')
        if not final_retry_execute.can_continue(summary['routes']):
            raise RuntimeError('previous failed/unknown process retains full allowance and blocks dependent continuation')
        record = dict(state='STARTING', route=r.ROUND + '-' + job['name'], baseName=job['name'],
            validationRound=r.ROUND, phase=job['phase'], sourceCommit=prepared['sourceCommit'],
            predecessorSealSHA256=prior['sealSHA256'], sourceVariable=job['sourceVariable'], credentialsRead=False, maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'],
            reservedMicrousd=reserve, automaticReruns=False, collectionRoot=str(out), networkRoute=auth['networkRoutes'][job['name']],
            sourceManifestSha256=prepared['sourceManifestSHA256'], binariesManifestSha256=prepared['binariesManifestSHA256'],
            planSha256=prepared['planSHA256'], sourceReviewSha256=c.digest(out / 'prerelease-source-review.json'),
            authorizationSha256=c.digest(out / 'execution-authorization.json'),
            budget=dict(state='UNKNOWN', status='UNKNOWN', calls=None, attempts=None))
        if (out/'execution-context.json').exists():
            ctx=final_retry.load(out);record.update(sourceCommit=ctx['measuredCommit'],controlSourceCommit=prepared['sourceCommit'],executionContextSHA256=final_retry.local_pin(out,'execution-context.json'))
        c.dump(marker(job['name']), c.redact(record), exclusive=True)
        for directory in (marker(job['name']).parent, root):
            fd=os.open(directory,os.O_RDONLY)
            try:os.fsync(fd)
            finally:os.close(fd)
    return marker(job['name'])


def checks(out):
    s.guard(out)
    commands = [('vet', ['go', 'vet', './...'], 600), ('vet-live', ['go', 'vet', '-tags', 'live', './...'], 600),
        ('no-key-live', ['go', 'test', '-tags', 'live', '-count=1', '-timeout=20m', './ai/live'], 1260)]
    codes, pins = {}, {}
    for name, argv, timeout in commands:
        codes[name] = s.run(out, name, argv, timeout=timeout,
            env={'BARNESS_AI_EVIDENCE_DIR': str(out / 'no-key-live')})
        pins[name] = c.digest(out / (name + '-command.json'))
        if codes[name] != 0:
            break
    result = dict(passed=len(codes) == len(commands) and all(v == 0 for v in codes.values()), commands=codes,
        commandReceipts=pins, credentialsRead=False)
    c.dump(out / 'checks.json', result)
    s.guard(out)
    if not result['passed']:
        raise RuntimeError('keyless checks failed; no paid intent/key')
    return result


def require_checks(out):
    check = c.doc(out / 'checks.json')
    expected = dict(vet=['go', 'vet', './...'], **{'vet-live': ['go', 'vet', '-tags', 'live', './...'],
        'no-key-live': ['go', 'test', '-tags', 'live', '-count=1', '-timeout=20m', './ai/live']})
    if check.get('passed') is not True or check.get('commands') != {n: 0 for n in expected}:
        raise RuntimeError('actual keyless checks required')
    for name, argv in expected.items():
        p = out / (name + '-command.json')
        command = c.doc(p)
        if (c.digest(p) != check.get('commandReceipts', {}).get(name) or command.get('argv') != argv or
            command.get('state') != 'COMPLETED' or type(command.get('exitCode')) is not int or command['exitCode'] != 0 or
            command.get('credentialsRead') is not False or command.get('credentialsInherited') is not False):
            raise RuntimeError('exact successful credential-free command receipts required')


def report_path(out, job):
    if 'combo' in job:
        from renewal_common import bundle
        # Candidates have only their native report; formal support is a later projection.
        return bundle(job['evidenceRoot'], 'regional-report.json') / 'regional-report.json'
    if job['name']=='glm-responses-retry-1':
        from renewal_common import bundle
        return bundle(job['evidenceRoot'],'formal-glm-report.json')/'formal-glm-report.json'
    return r.old.report_path(out, job)


def settle(out, job):
    try:
        report = c.doc(report_path(out, job))
        b = report['budget']
        if job['name'] == 'protocol114':
            b = b['actual']
        a, h = ('callsUsed', 'httpAttempts') if 'callsUsed' in b else ('calls', 'attempts')
        calls, attempts = b[a], b[h]
        if (type(calls) is not int or type(attempts) is not int or not 0 <= calls <= job['maxCalls'] or not 0 <= attempts <= job['maxAttempts']):
            raise RuntimeError('invalid native counters')
        import paid
        status = paid.report_status(report, job)
        if status != 'PASS':
            raise RuntimeError('own native fixed process did not PASS')
        return dict(state='REPORTED', status='PASS', calls=calls, attempts=attempts)
    except (OSError, ValueError, KeyError, RuntimeError, TypeError):
        return dict(state='UNKNOWN', status='UNKNOWN', calls=None, attempts=None)


def run_selected(out, job):
    import final_retry,final_retry_execute
    if job['name']==final_retry.NAME:prepared,auth,scope,reserve=final_retry_execute.approved(out,s.guard(out),job)
    else:prepared, auth, scope, reserve = s.approved(out, job)
    require_checks(out)
    import final_native
    final_native.dependencies(job)
    import final_grants
    final_grants.require_component(out,job,auth,scope)
    if Path(job['evidenceRoot']).exists() or (out / (job['name'] + '-command.json')).exists():
        raise RuntimeError('fresh native output required; no process restart')
    intent = claim(out, job, prepared, auth, reserve)
    record = final_grants.native_preflight(out,job,c.doc(intent))
    c.dump(intent,record)
    env = dict(job['environment'])
    env.update(BARNESS_FINAL_ROUND=r.ROUND, BARNESS_FINAL_BUDGET=str(out / (job['name'] + '-usd.json')))
    c.dump(Path(env['BARNESS_FINAL_BUDGET']), dict(schema=1, round=r.ROUND, maxMicrousd=reserve,
        reservedMicrousd=0, attempts=0, basis='Conservative final-wire known-rate estimate, not invoice; no release on unknown usage.'))
    proxy = auth['networkRoute'] if job['name']==final_retry.NAME else auth['networkRoutes'][job['name']]
    if proxy != 'direct':
        if job['name'].startswith(('text-', 'visual-')):
            argv = [*job['argv'], '-proxy', proxy]
        else:
            env['BARNESS_REGIONAL_PROXY' if 'combo' in job else 'BARNESS_MINIMAX_PROXY' if job['name'].startswith('mm-') else 'BARNESS_AI_LIVE_PROXY'] = proxy
            argv = job['argv']
    else:
        argv = job['argv']
    env.update(BARNESS_AI_LIVE_ACCOUNT_ALIAS=scope['accountAlias'] + '@' + scope['regionAlias'],
        BARNESS_MINIMAX_ACCOUNT_ALIAS=scope['accountAlias'], BARNESS_MINIMAX_REGION_ALIAS=scope['regionAlias'],
        BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=scope['accountAlias'] + '@' + scope['regionAlias'])
    import final_process
    if (out/'execution-context.json').exists():s.guard(out)
    code = final_process.run(out,job,argv,env,intent)
    record = c.doc(intent)
    record.update(commandReceiptSha256=c.digest(out / (job['name'] + '-command.json')),
        budget=settle(out, job) if code == 0 else dict(state='UNKNOWN', status='UNKNOWN', calls=None, attempts=None))
    if code==0 and record['budget']['state']=='REPORTED':
        usd=c.doc(Path(env['BARNESS_FINAL_BUDGET']))
        if (usd.get('round')!=r.ROUND or usd.get('maxMicrousd')!=reserve or
            type(usd.get('attempts')) is not int or usd['attempts']<record['budget']['attempts'] or
            type(usd.get('reservedMicrousd')) is not int or not 0<=usd['reservedMicrousd']<=reserve or
            record['budget']['attempts']>0 and usd['reservedMicrousd']==0):
            record['budget']['status']='FAIL'
    c.dump(intent, record)
    s.guard(out)
    if code == 0 and record['budget']['status'] == 'PASS':
        try:
            record = final_native.consume(out, job, record)
            c.dump(intent, record)
        except Exception:
            record['budget']['status'] = 'FAIL'
            c.dump(intent, record)
            raise
    if code != 0 or record['budget']['status'] != 'PASS':
        raise RuntimeError('selected process failed or consumption unknown; retain claim, native files and entire USD reservation')
    return dict(route=job['name'], exitCode=code, budget=record['budget'])
