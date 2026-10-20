"""Concrete native consumers and canonical byte pins for this final round."""
from pathlib import Path
import final_contract as r
import final_state as s
import final_execute as e
c = r.c


def verified(out, name, argv):
    if s.run(out, name, argv, timeout=180) != 0:
        raise RuntimeError('own native consumer refused; retain failure and stop dependent stages')
    return c.digest(out / (name + '-command.json'))


def consume(out, job, record):
    report_path = e.report_path(out, job)
    report = c.doc(report_path)
    base = report_path.parent
    # Effects bind complete source contents through their native CodeHash /
    # code-snapshot consumers; their schema deliberately has no GitCommit.
    if job['name'] not in r.EFFECTS and report.get('gitCommit', report.get('sourceCommit')) != record['sourceCommit']:
        raise RuntimeError('native result must bind the reviewed actual source commit')
    name = job['name']
    if 'combo' in job:
        argv = [out / 'regionalverify', '-qualify' if job['stage'] == 'candidate' else '-formal', base]
    elif name.startswith('glm-'):
        if (report.get('mode') != 'live' or report.get('kind') != 'glm-formal' or
            report.get('route') != job['glmRoute'] or report.get('outcome') != 'PASS' or
            report.get('complete') is not True or report.get('auditPassed') is not True or
            len(report.get('scenarios', [])) != 22):
            raise RuntimeError('own complete live full22 GLM capture required')
        record['capturePins'] = dict(schema=1, **{key: c.digest(base / leaf) for key, leaf in (
            ('formalReportSha256', 'formal-glm-report.json'), ('integritySha256', 'glm-integrity.json'),
            ('supportReportSha256', 'live-report.json'), ('executionSourceSha256', 'glm-execution-source.json'))})
        # The public importer is also used by the raw publisher. A private
        # candidate matrix never modifies production support state.
        matrix = out / (name + '-validation-matrix.json')
        matrix.write_bytes((c.ROOT / 'ai/live/support-matrix.json').read_bytes())
        argv = [out / 'supportmatrix', '-matrix', matrix, base]
    elif name.startswith('mm-'):
        grant = c.common_root() / '.evidence/minimax-renewals' / r.ROUND / 'grant.json'
        if report.get('formalRenewalId') != r.ROUND or report.get('formalRenewalGrantSha256') != c.digest(grant):
            raise RuntimeError('own new native renewal grant required; old round is closed')
        record.update(renewalId=r.ROUND, renewalGrantSha256=c.digest(grant))
        record['origin'] = dict(schema=1, stage='formal', identity=report['identity'], session=report['session'],
            reportSha256=c.digest(base / 'minimax-report.json'), sealSha256=c.digest(base / 'minimax-integrity.json'),
            ledgerSha256=c.digest(base / 'minimax-ledger.json'), sourceCommit=record['sourceCommit'],
            sourceManifestSha256=record['sourceManifestSha256'], binariesManifestSha256=record['binariesManifestSha256'],
            planSha256=record['planSha256'], formalRenewalId=r.ROUND, formalRenewalGrantSha256=c.digest(grant))
        # Native Verify requires the canonical external launcher receipt before
        # publishing; a bundle may never self-assert its own live origin.
        c.dump(e.marker(name), record)
        argv = [out / 'minimaxverify', '-formal', '-publish', base]
    elif name.startswith('text-'):
        argv = [out / 'chineseeval', '-route', name[5:], '-verify', base]
    elif name.startswith('visual-'):
        argv = [out / 'visualeval', '-verify', base, '-require-live']
    elif name == 'protocol114':
        # Jev support derivation must later pair this protocol with its own
        # fresh visual effect; no independent matrix claim is made here.
        return record | dict(reportedFileSha256=c.digest(report_path), nativeConsumer='paired-jevformal-pending')
    else:
        matrix = out / (name + '-validation-matrix.json')
        matrix.write_bytes((c.ROOT / 'ai/live/support-matrix.json').read_bytes())
        argv = [out / 'supportmatrix', '-matrix', matrix, base]
    record.update(reportedFileSha256=c.digest(report_path),
        nativeConsumerReceiptSha256=verified(out, name + '-consume', argv))
    return record


def dependencies(job):
    if job['phase'] == 'candidate':
        return
    names = ['candidate-' + combo for combo in r.REGIONAL]
    if job['phase'] == 'remaining':
        names += ['glm-' + route for route in r.ROUTES]
    for name in names:
        value = c.doc(e.marker(name))
        if (value.get('state') != 'COMPLETED' or value.get('exitCode') != 0 or
            value.get('budget', {}).get('status') != 'PASS' or not value.get('nativeConsumerReceiptSha256')):
            raise RuntimeError('earlier own successful consumed fixed-stage evidence required')
    if job['phase'] == 'remaining':
        index = c.doc(c.ROOT / 'ai/live/testdata/glm-formal/live-captures.json')
        if index.get('collectionRound') != r.ROUND:
            raise RuntimeError('fresh own GLM selector publication required before final freeze')
        from renewal_common import committed
        committed(['ai/live/testdata/glm-formal/' + n for n in (
            'live-captures.json', 'live-captures.sha256', 'known-usage.json', 'support-views.json')])


def publication_receipt(route):
    """Read-only strict current historical launcher check after later commits."""
    name = 'glm-' + route
    value = c.doc(e.marker(name))
    out = s.path(value['collectionRoot'].replace('$REPO/', ''))
    prepared = c.doc(out / 'prepared.json')
    command = c.doc(out / (name + '-command.json'))
    expected = r.jobs(out)[name]
    fields = dict(sourceManifestSha256='sourceManifestSHA256', binariesManifestSha256='binariesManifestSHA256', planSha256='planSHA256')
    if (value.get('validationRound') != r.ROUND or value.get('baseName') != name or value.get('phase') != 'glm' or
        value.get('state') != 'COMPLETED' or value.get('exitCode') != 0 or value.get('credentialsRead') is not True or
        value.get('sourceVariable') != 'ZHIPU_KEY' or value.get('networkRoute') != 'direct' or
        value.get('automaticReruns') is not False or value.get('maxCalls') != 24 or value.get('maxAttempts') != 32 or
        value.get('budget') != dict(state='REPORTED', status='PASS', calls=22, attempts=22) or
        value.get('sourceCommit') != prepared.get('sourceCommit') or command.get('argv') != c.redact(expected['argv']) or
        command.get('state') != 'COMPLETED' or command.get('exitCode') != 0 or command.get('credentialsRead') is not True or
        c.digest(out / (name + '-command.json')) != value.get('commandReceiptSha256') or
        value.get('launcherAudit', {}).get('passed') is not True):
        raise RuntimeError('new own reviewed direct full22 canonical launcher receipt required')
    for leaf, field in [('source-at-start.json', 'sourceManifestSha256'), ('binaries.json', 'binariesManifestSha256'),
        ('plan.json', 'planSha256'), ('prerelease-source-review.json', 'sourceReviewSha256'), ('execution-authorization.json', 'authorizationSha256')]:
        if c.digest(out / leaf) != value.get(field):
            raise RuntimeError('new canonical source/binary/review/approval/plan bytes changed')
    if any(value[k] != prepared[v] for k, v in fields.items()):
        raise RuntimeError('new canonical prepared source binding changed')
    e.require_checks(out)
    s.approved_metadata(out, expected, prepared)
    receipt_path = out / (name + '-consume-command.json')
    consumer = c.doc(receipt_path)
    if (c.digest(receipt_path) != value.get('nativeConsumerReceiptSha256') or consumer.get('state') != 'COMPLETED' or
        consumer.get('exitCode') != 0 or consumer.get('credentialsRead') is not False):
        raise RuntimeError('actual successful keyless native consumer receipt required')
    return value, c.digest(e.marker(name))


def promote(out):
    prepared=s.guard(out)
    if prepared['phase']!='candidate':
        raise RuntimeError('promotion requires this frozen six-candidate phase')
    from renewal_common import bundle
    paths=[]
    for combo in r.REGIONAL:
        name='candidate-'+combo;receipt=c.doc(e.marker(name))
        if (receipt.get('state')!='COMPLETED' or receipt.get('exitCode')!=0 or
            receipt.get('budget',{}).get('status')!='PASS' or not receipt.get('nativeConsumerReceiptSha256') or
            receipt.get('sourceCommit')!=prepared['sourceCommit']):
            raise RuntimeError('all six own native-consumed current candidate processes must PASS')
        paths.append(bundle(out/name,'live-report.json'))
    before=s.sources()
    receipt_root=c.ROOT/'.evidence/provider-sites/final-acceptance-20261020/promotion'
    receipt_root.mkdir(parents=True,exist_ok=True,mode=0o700)
    lineage=receipt_root/'regional-promotion-lineage.json'
    verified(out,'promote-six',[out/'regionalpromote','-receipt',lineage,*paths])
    after=s.sources();allowed='ai/catalog_regional.json'
    if set(before)!=set(after) or {n for n in before if before[n]!=after[n]}!={allowed}:
        raise RuntimeError('data activation changed unexpected frozen source; retain actual state and stop')
    return dict(credentialsRead=False,activatedFile=allowed,lineageSHA256=c.digest(lineage),
        next='Run the reviewed keyless catalog-pins operation; commit activation/catalog pins, then prepare a new GLM-source phase. Candidate captures remain unchanged historical measurements.')
