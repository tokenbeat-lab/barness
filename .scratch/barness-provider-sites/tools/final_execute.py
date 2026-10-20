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
    auth_path = authority_root() / 'operator-authority.json'
    if not auth_path.exists():
        import final_research
        research=final_research.disclosure();done=research['state']=='COMPLETED'
        return dict(round=r.ROUND,research=research,reservedMicrousd=research['reservedMicrousd'] if done else 0,
            claimedCalls=4 if done else 0,claimedAttempts=4 if done else 0,actualResearchCalls=research['calls'],
            actualResearchAttempts=research['attempts'],routes=[],credentialsRead=False)
    auth = c.doc(auth_path)
    rows = []
    for name, job in r.jobs(c.ROOT / '.evidence/unused-final-summary').items():
        p = marker(name)
        if not p.exists():
            continue
        receipt = c.doc(p)
        if receipt.get('validationRound') != r.ROUND or receipt.get('baseName') != name:
            raise RuntimeError('canonical fixed-round claim changed')
        rows.append(dict(name=name, reservedMicrousd=receipt['reservedMicrousd'], maxCalls=job['maxCalls'],
            maxAttempts=job['maxAttempts'], state=receipt.get('state'), budget=receipt.get('budget')))
    return dict(round=r.ROUND, research=auth['research'], actualResearchCalls=4, actualResearchAttempts=4, maxMicrousd=auth['maxMicrousd'],
        reservedMicrousd=auth['diagnosticReservedMicrousd'] + sum(v['reservedMicrousd'] for v in rows),
        claimedCalls=4 + sum(v['maxCalls'] for v in rows), claimedAttempts=4 + sum(v['maxAttempts'] for v in rows),
        knownReportedCalls=4+sum((v.get('budget') or {}).get('calls') or 0 for v in rows),
        knownReportedAttempts=4+sum((v.get('budget') or {}).get('attempts') or 0 for v in rows),
        unknownProcesses=sum((v.get('budget') or {}).get('state')=='UNKNOWN' for v in rows),
        routes=rows, credentialsRead=False, basis='Conservative known-rate operational reservations; unknown/failed/cancelled consumption never releases an allowance; not an invoice guarantee.')


def claim(out, job, prepared, auth, reserve):
    diagnostic = auth.get('diagnosticReservedMicrousd')
    disclosure=c.doc(out/'research-disclosure.json')
    if disclosure.get('state')!='COMPLETED' or disclosure.get('calls')!=4 or disclosure.get('attempts')!=4:
        raise RuntimeError('original one-shot four-call research must be imported; never launch it again')
    if type(diagnostic) is not int or diagnostic != 200000:
        raise RuntimeError('already consumed research keeps its full200000 microusd allocation')
    identity = dict(round=r.ROUND, requestText=auth['requestText'], requestReference=auth['requestReference'],
        maxMicrousd=int(Decimal(str(auth['maxUSD'])) * 1000000), diagnosticReservedMicrousd=diagnostic,
        maxCalls=r.MAX_CALLS, maxAttempts=r.MAX_ATTEMPTS, automaticProcessReruns=False, research=disclosure)
    with locked() as root:
        proof = root / 'operator-authority.json'
        if proof.exists():
            if c.doc(proof) != identity:
                raise RuntimeError('new stage must bind the same finite human round, not a replacement authority')
        else:
            c.dump(proof, identity, exclusive=True)
        summary = summarize()
        if (summary['reservedMicrousd'] + reserve > identity['maxMicrousd'] or
            summary['claimedCalls'] + job['maxCalls'] > r.MAX_CALLS or
            summary['claimedAttempts'] + job['maxAttempts'] > r.MAX_ATTEMPTS):
            raise RuntimeError('remaining round USD/call/attempt reservation exhausted before key')
        if any(v.get('state') != 'COMPLETED' or v.get('budget', {}).get('status') != 'PASS' for v in summary['routes']):
            raise RuntimeError('previous failed/unknown process retains full allowance and blocks dependent continuation')
        record = dict(state='STARTING', route=r.ROUND + '-' + job['name'], baseName=job['name'],
            validationRound=r.ROUND, phase=job['phase'], sourceCommit=prepared['sourceCommit'],
            sourceVariable=job['sourceVariable'], credentialsRead=False, maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'],
            reservedMicrousd=reserve, automaticReruns=False, collectionRoot=str(out), networkRoute=auth['networkRoutes'][job['name']],
            sourceManifestSha256=prepared['sourceManifestSHA256'], binariesManifestSha256=prepared['binariesManifestSHA256'],
            planSha256=prepared['planSHA256'], sourceReviewSha256=c.digest(out / 'prerelease-source-review.json'),
            authorizationSha256=c.digest(out / 'execution-authorization.json'),
            budget=dict(state='UNKNOWN', status='UNKNOWN', calls=None, attempts=None))
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
        return bundle(job['evidenceRoot'], 'live-report.json') / 'live-report.json'
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
    prepared, auth, scope, reserve = s.approved(out, job)
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
    proxy = auth['networkRoutes'][job['name']]
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
