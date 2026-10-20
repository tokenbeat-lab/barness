"""Shared private fixed recovery native mechanics; authority values come only from compiled contexts."""
import datetime as dt
import json
from pathlib import Path

class NativeConsumer:

    def __init__(self, flow):
        self.flow = flow

    def templates(self, out):
        import final_grants
        return final_grants.templates(out, 'remaining')

    def verified(self, out, name, argv, timeout=180):
        if type(timeout) is not int or timeout not in (180,360):raise RuntimeError('fixed native consumer resource bound required')
        if any((out/(name+suffix)).exists() or (out/(name+suffix)).is_symlink() for suffix in ('.log','-command.json')):raise RuntimeError('native command output already exists; preserve first outcome')
        if self.flow.s.run(out, name, argv, timeout=timeout) != 0:
            raise RuntimeError('current native consumer refused; preserve failure')
        return self.flow.s.pin(out / (name + '-command.json'))

    def require_component(self, out, job, auth, scope):
        name = job['name']
        if 'combo' not in job:
            if name != 'protocol114' and (not name.startswith('visual-')):
                return
            leaf = 'protocol-authorization.json' if name == 'protocol114' else name + '-authorization.json'
            p = out / leaf
            value = self.flow.s.doc(p)
            template = self.flow.s.doc(out / (leaf + '.template.json'))
            mutable = {'accountAlias', 'regionAlias'}
            if set(value) != set(template) or {k: v for k, v in value.items() if k not in mutable} != {k: v for k, v in template.items() if k not in mutable} or value.get('accountAlias') != scope['accountAlias'] or (value.get('regionAlias') != scope['regionAlias']) or (value.get('dedicatedCredential') is not True) or (value.get('maxCalls') != 12) or (value.get('maxAttempts') != 12) or (self.flow.s.pin(p) != auth['componentAuthorizationSHA256'].get(name)):
                raise RuntimeError('new own bounded protocol/effect authorization required')
            return
        p = Path(job['environment']['BARNESS_REGIONAL_GRANT'])
        template = self.flow.s.doc(out / 'grants' / (name + '.template.json'))
        value = self.flow.s.doc(p)
        mutable = {'id', 'accountAlias', 'regionAlias', 'expiresAt'}
        if set(value) != set(template) or {k: v for k, v in value.items() if k not in mutable} != {k: v for k, v in template.items() if k not in mutable} or value.get('mode') != 'live' or (value.get('accountAlias') != scope['accountAlias']) or (value.get('regionAlias') != scope['regionAlias']) or (not value.get('id', '').startswith(self.flow.r.REGIONAL_GRANT_PREFIX)):
            raise RuntimeError('new own exact recovery regional grant required')
        expires = dt.datetime.fromisoformat(value['expiresAt'].replace('Z', '+00:00'))
        now = dt.datetime.now(dt.timezone.utc)
        if expires <= now or expires - now > dt.timedelta(hours=24) or self.flow.s.pin(p) != auth['componentAuthorizationSHA256'].get(name):
            raise RuntimeError('reviewed current expiring regional grant pin required')

    def preflight(self, out, job, record):
        if not job['name'].startswith('mm-'):
            return record
        name = job['name'] + '-native-preflight'
        self.verified(out, name, [out / 'minimaxverify', self.flow.r.NATIVE_PREFLIGHT_FLAG + job['route']])
        value = self.flow.s.doc(out/(name+'.log'))
        if value.get('route') != job['route'] or value.get('facts', {}).get('renewalId') != self.flow.r.ROUND:
            raise RuntimeError('native own recovery preflight identity required')
        return record | dict(nativePreflightReceiptSha256=self.flow.s.pin(out / (name + '-command.json')), renewalId=self.flow.r.ROUND, renewalGrantSha256=value['grantSha256'])

    def consume(self, out, job, record):
        import final_execute as old
        report_path = old.report_path(out, job)
        report = self.flow.s.doc(report_path)
        base = report_path.parent
        name = job['name']
        if name not in self.flow.r.EFFECTS and report.get('gitCommit', report.get('sourceCommit')) != record['sourceCommit']:
            raise RuntimeError('own current native measured commit required')
        if 'combo' in job:
            argv = [out / 'regionalverify', '-formal', base]
        elif name.startswith('glm-'):
            if report.get('mode')!='live' or report.get('kind')!='glm-formal' or report.get('route')!=job['glmRoute'] or report.get('outcome')!='PASS' or report.get('complete') is not True or report.get('auditPassed') is not True or len(report.get('scenarios',[]))!=22:raise RuntimeError('own complete current live full22 GLM raw capture required')
            record['capturePins']=dict(schema=1,**{key:self.flow.s.pin(base/leaf) for key,leaf in (('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json'))})
            matrix=out/(name+'-validation-matrix.json')
            with matrix.open('xb') as stream:stream.write(self.flow.files.read(self.flow.r.c.ROOT/'ai/live/support-matrix.json'))
            argv=[out/'supportmatrix','-matrix',matrix,base]
        elif name.startswith('mm-'):
            grant = self.flow.r.c.common_root() / '.evidence/minimax-renewals' / self.flow.r.ROUND / 'grant.json'
            if report.get('formalRenewalId') != self.flow.r.ROUND or report.get('formalRenewalGrantSha256') != self.flow.s.pin(grant):
                raise RuntimeError('own current recovery native full report required')
            record.update(renewalId=self.flow.r.ROUND, renewalGrantSha256=self.flow.s.pin(grant), origin=dict(schema=1, stage='formal', identity=report['identity'], session=report['session'], reportSha256=self.flow.s.pin(base / 'minimax-report.json'), sealSha256=self.flow.s.pin(base / 'minimax-integrity.json'), ledgerSha256=self.flow.s.pin(base / 'minimax-ledger.json'), sourceCommit=record['sourceCommit'], sourceManifestSha256=record['sourceManifestSha256'], binariesManifestSha256=record['binariesManifestSha256'], planSha256=record['planSha256'], formalRenewalId=self.flow.r.ROUND, formalRenewalGrantSha256=self.flow.s.pin(grant)))
            self.flow.r.c.dump(self.flow.e.marker(name), record)
            argv = [out / 'minimaxverify', '-formal', '-publish', base]
        elif name.startswith('text-'):
            argv = [out / 'chineseeval', '-route', name[5:], '-verify', base]
        elif name.startswith('visual-'):
            argv = [out / 'visualeval', '-verify', base, '-require-live']
        elif name == 'protocol114':
            return record | dict(reportedFileSha256=self.flow.s.pin(report_path), nativeConsumer='paired-jevformal-pending')
        else:
            matrix = out / (name + '-validation-matrix.json')
            with matrix.open('xb') as stream:stream.write(self.flow.files.read(self.flow.r.c.ROOT/'ai/live/support-matrix.json'))
            argv = [out / 'supportmatrix', '-matrix', matrix, base]
        # Only the trusted fixed factory can assign a selected consumer bound.
        timeout=self.flow.r.jobs(out)[name].get('consumerTimeoutSeconds',180)
        return record | dict(reportedFileSha256=self.flow.s.pin(report_path), nativeConsumerReceiptSha256=self.verified(out, name + '-consume', argv, timeout=timeout))

    def require_consumed(self, out, job, value):
        import final_execute
        name = job['name']
        report = final_execute.report_path(out, job)
        if value.get('sourceCommit') != self.flow.s.doc(out / 'prepared.json')['sourceCommit'] or value.get('exitCode') != 0 or value.get('state') != 'COMPLETED' or (value.get('budget', {}).get('status') != 'PASS') or (value.get('reportedFileSha256') != self.flow.s.pin(report)) or (value.get('commandReceiptSha256') != self.flow.s.pin(out / (name + '-command.json'))):
            raise RuntimeError('prior current recovery complete native process facts required')
        if name == 'protocol114':
            if value.get('nativeConsumer') != 'paired-jevformal-pending':
                raise RuntimeError('explicit paired Jev protocol consumer required')
            return
        if name.startswith('glm-'):
            wanted=dict(schema=1,**{key:self.flow.s.pin(report.parent/leaf) for key,leaf in (('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json'))})
            if value.get('capturePins')!=wanted:raise RuntimeError('current raw GLM capture pins changed')
        p = out / (name + '-consume-command.json')
        command = self.flow.s.doc(p)
        base = report.parent
        if 'combo' in job:
            argv = [out / 'regionalverify', '-formal', base]
        elif name.startswith('mm-'):
            argv = [out / 'minimaxverify', '-formal', '-publish', base]
        elif name.startswith('text-'):
            argv = [out / 'chineseeval', '-route', name[5:], '-verify', base]
        elif name.startswith('visual-'):
            argv = [out / 'visualeval', '-verify', base, '-require-live']
        else:
            argv = [out / 'supportmatrix', '-matrix', out / (name + '-validation-matrix.json'), base]
        if value.get('nativeConsumerReceiptSha256') != self.flow.s.pin(p) or command.get('state') != 'COMPLETED' or type(command.get('exitCode')) is not int or command.get('exitCode') != 0 or command.get('launcherAudit')!=dict(passed=True,findings=[]) or (command.get('credentialsRead') is not False) or (command.get('argv') != self.flow.r.c.redact(list(map(str, argv)))):
            raise RuntimeError('prior actual own successful native consumer required')
