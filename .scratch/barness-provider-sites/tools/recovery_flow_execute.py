"""Shared private fixed recovery execute mechanics; authority values come only from compiled contexts."""
from contextlib import contextmanager
import fcntl
import json
import os
from pathlib import Path
import stat
from recovery_flow_files import claim_names

UNKNOWN = dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)

# The file lock and whole process reservation remain durable across failures.
class Execution:

    def __init__(self, flow):
        self.flow = flow

    def marker(self, name):
        if name not in self.flow.r.RESERVES:
            raise RuntimeError('fixed own marker required')
        return self.flow.h.f.owned('.evidence/minimax-launches/' + self.flow.r.ROUND + '-' + name + '.json')

    def authority(self):
        return self.flow.h.f.owned('.evidence/minimax-renewals/' + self.flow.r.ROUND)

    def sync_dir(self, path):
        fd = os.open(path, os.O_RDONLY)
        try:
            os.fsync(fd)
        finally:
            os.close(fd)

    @contextmanager
    def locked(self):
        base = self.authority()
        base.mkdir(parents=True, exist_ok=True, mode=448)
        self.sync_dir(base.parent)
        lock = base / 'launch.lock'
        fd = os.open(lock, os.O_CREAT | os.O_RDWR | os.O_NOFOLLOW | os.O_NONBLOCK, 384)
        try:
            if not stat.S_ISREG(os.fstat(fd).st_mode):
                raise RuntimeError('regular canonical lock required')
            fcntl.flock(fd, fcntl.LOCK_EX)
            yield base
        finally:
            os.close(fd)

    def summarize(self):
        prior = self.flow.h.facts()
        if claim_names(self.flow.r.c,self.flow.r.ROUND)-set(self.flow.r.RESERVES):raise RuntimeError('unregistered canonical recovery claim; no ignored allowance')
        rows = []
        for name, job in self.flow.r.jobs(self.flow.r.c.ROOT / '.evidence/unused-recovery-summary').items():
            p = self.marker(name)
            if not p.exists():
                if p.is_symlink():
                    raise RuntimeError('canonical claim symlink refused')
                continue
            value = self.flow.s.doc(p)
            if any((value.get(k) != v for k, v in dict(validationRound=self.flow.r.ROUND, nativeRenewalId=self.flow.r.ROUND, baseName=name, maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'], reservedMicrousd=job['reservedMicrousd'], **self.flow.r.policy()).items())):
                raise RuntimeError('canonical recovery scope/policy/reserve changed')
            rows.append(dict(name=name, state=value['state'], budget=value['budget'], maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'], reservedMicrousd=value['reservedMicrousd']))
        calls = sum((v['maxCalls'] for v in rows))
        attempts = sum((v['maxAttempts'] for v in rows))
        usd = sum((v['reservedMicrousd'] for v in rows))
        if calls > self.flow.r.MAX_CALLS or attempts > self.flow.r.MAX_ATTEMPTS or usd > self.flow.r.NEW_USD:
            raise RuntimeError('fixed recovery total exceeded')
        return dict(round=self.flow.r.ROUND, predecessor=prior, routes=rows, originalProcessCount=self.flow.r.ORIGINAL_PROCESS_COUNT, newProcessCount=len(rows), actualProcessCount=self.flow.r.ORIGINAL_PROCESS_COUNT + len(rows), effectiveRoleCount=self.flow.r.BASE_ROLE_COUNT + len(rows), retainedUnknownProcesses=self.flow.r.UNKNOWN_PROCESSES, claimedCalls=self.flow.r.RETAINED_CALLS + calls, claimedAttempts=self.flow.r.RETAINED_ATTEMPTS + attempts, knownReportedCalls=self.flow.r.RETAINED_ACTUAL_CALLS + sum((v['budget'].get('calls') or 0 for v in rows)), knownReportedAttempts=self.flow.r.RETAINED_ACTUAL_ATTEMPTS + sum((v['budget'].get('attempts') or 0 for v in rows)), newReservedMicrousd=usd, reservedMicrousd=self.flow.r.RETAINED_USD + usd, campaignMaxCalls=self.flow.r.CAMPAIGN_CALLS, campaignMaxAttempts=self.flow.r.CAMPAIGN_ATTEMPTS, maxMicrousd=self.flow.r.MAX_USD*1000000, credentialsRead=False)

    def require_checks(self, out):
        check=self.flow.s.doc(out/'checks.json')
        expected={'vet':['go','vet','./...'],'vet-live':['go','vet','-tags','live','./...'],'no-key-live':['go','test','-tags','live','-count=1','-timeout=20m','./ai/live']}
        if check.get('passed') is not True or check.get('commands')!={name:0 for name in expected}:raise RuntimeError('actual keyless checks required')
        for name,argv in expected.items():
            path=out/(name+'-command.json');value=self.flow.s.doc(path)
            if self.flow.s.pin(path)!=check.get('commandReceipts',{}).get(name) or value.get('argv')!=argv or value.get('state')!='COMPLETED' or type(value.get('exitCode')) is not int or value['exitCode']!=0 or value.get('credentialsRead') is not False or value.get('credentialsInherited') is not False:raise RuntimeError('exact successful credential-free check receipts required')

    def claim(self, out, job, prepared, auth, admission):
        with self.locked() as base:
            if self.marker(job['name']).exists() or self.marker(job['name']).is_symlink():
                raise RuntimeError('fixed job already claimed in any output or checkout')
            summary = self.summarize()
            names = {v['name'] for v in summary['routes']}
            index = len(names)
            if index >= len(self.flow.r.EXECUTION_ORDER) or names != set(self.flow.r.EXECUTION_ORDER[:index]) or job['name'] != self.flow.r.EXECUTION_ORDER[index]:
                raise RuntimeError('fixed execution order requires each prior own full native PASS')
            if any((v['state'] != 'COMPLETED' or v['budget']['status'] != 'PASS' for v in summary['routes'])):
                raise RuntimeError('previous recovery process failed or unknown; no automatic continuation')
            for name in self.flow.r.EXECUTION_ORDER[:index]:
                self.flow.n.require_consumed(out, self.flow.r.jobs(out)[name], self.flow.s.doc(self.marker(name)))
            if summary['reservedMicrousd']+job['reservedMicrousd']>self.flow.r.MAX_USD*1000000 or summary['newReservedMicrousd']+job['reservedMicrousd']>self.flow.r.NEW_USD or summary['claimedCalls']+job['maxCalls']>self.flow.r.CAMPAIGN_CALLS or summary['claimedAttempts']+job['maxAttempts']>self.flow.r.CAMPAIGN_ATTEMPTS:raise RuntimeError('fixed whole-allocation caps exhausted before Key')
            identity = {k: auth[k] for k in ('round', 'validationRound', 'nativeRenewalId', 'spendingGuardPolicyId', 'spendingGuardSourceSHA256', 'requestText', 'requestReference', 'actualHumanApprovalSHA256', 'maxCalls', 'maxAttempts', 'campaignMaxCalls', 'campaignMaxAttempts', 'newReservedMicrousd', 'retainedReservedMicrousd', 'maxUSD', 'automaticProcessReruns', 'predecessor', 'eligibilityAuthorizationSHA256', 'raceAuthorizationSHA256')}
            proof = base / 'operator-authority.json'
            if proof.exists():
                if self.flow.s.doc(proof) != identity:
                    raise RuntimeError('same fixed recovery actual human authority required across all phases')
            else:
                self.flow.r.c.dump(proof, identity, exclusive=True)
            self.sync_dir(base)
            record = dict(state='STARTING', route=self.flow.r.ROUND + '-' + job['name'], baseName=job['name'], validationRound=self.flow.r.ROUND, nativeRenewalId=self.flow.r.ROUND, phase=self.flow.r.PHASE, sourceCommit=prepared['sourceCommit'], sourceVariable=job['sourceVariable'], credentialsRead=False, maxCalls=job['maxCalls'], maxAttempts=job['maxAttempts'], reservedMicrousd=job['reservedMicrousd'], automaticReruns=False, collectionRoot=str(out), networkRoute=job['networkRoute'], budget=UNKNOWN.copy(), **self.flow.r.policy(), originalCampaignApprovalSHA256=auth['eligibilityAuthorizationSHA256'], actualHumanApprovalSHA256=auth['actualHumanApprovalSHA256'], usdLedger=str(out / (job['name'] + '-usd.json')))
            for field, key in [('sourceManifestSha256', 'sourceManifestSHA256'), ('binariesManifestSha256', 'binariesManifestSHA256'), ('planSha256', 'planSHA256')]:
                record[field] = prepared[key]
            record.update(admission)
            record.update(sourceReviewSha256=self.flow.s.pin(out / 'prerelease-source-review.json'), authorizationSha256=self.flow.s.pin(out / 'execution-authorization.json'))
            self.flow.r.c.dump(self.marker(job['name']), self.flow.r.c.redact(record), exclusive=True)
            self.sync_dir(self.marker(job['name']).parent)
            return self.marker(job['name'])

    def settle(self,out,job):
        import final_execute,paid
        try:
            report=self.flow.s.doc(final_execute.report_path(out,job));budget=report['budget']
            if job['name']=='protocol114':budget=budget['actual']
            calls,attempts=(budget['callsUsed'],budget['httpAttempts']) if 'callsUsed' in budget else (budget['calls'],budget['attempts'])
            if type(calls) is not int or type(attempts) is not int or not 0<=calls<=job['maxCalls'] or not 0<=attempts<=job['maxAttempts'] or paid.report_status(report,job)!='PASS':raise RuntimeError('own complete native process counters required')
            return dict(state='REPORTED',status='PASS',calls=calls,attempts=attempts)
        except (OSError,ValueError,KeyError,RuntimeError,TypeError):return UNKNOWN.copy()

    def run_selected(self, out, job):
        import final_process
        prepared, auth, scope, reserve = self.flow.s.approved(out, job)
        self.require_checks(out)
        self.flow.n.require_component(out, job, auth, scope)
        if Path(job['evidenceRoot']).exists() or (out / (job['name'] + '-command.json')).exists() or (out / (job['name'] + '-usd.json')).exists():
            raise RuntimeError('fresh recovery output required; no process restart')
        admission=self.flow.authority.before_claim(out,job) or {}
        intent = self.claim(out, job, prepared, auth, admission)
        record = self.flow.n.preflight(out, job, self.flow.s.doc(intent))
        self.flow.r.c.dump(intent, record)
        usd = out / (job['name'] + '-usd.json')
        env = job['environment'] | dict(BARNESS_FINAL_ROUND=self.flow.r.SPENDING_GUARD_POLICY_SELECTOR, BARNESS_FINAL_BUDGET=str(usd))
        self.flow.r.c.dump(usd, dict(schema=1, round=self.flow.r.SPENDING_GUARD_POLICY_SELECTOR, maxMicrousd=reserve, reservedMicrousd=0, attempts=0, basis=self.flow.r.LEDGER_BASIS+'; executionAuthority='+self.flow.r.ROUND+'; spendingGuardSourceSHA256='+json.dumps(self.flow.r.policy()['spendingGuardSourceSHA256'],sort_keys=True)+'; no refund; operational estimate, not invoice.'), exclusive=True)
        argv = job['argv']
        proxy = job['networkRoute']
        if proxy != 'direct':
            if job['name'].startswith(('text-', 'visual-')):
                argv = [*argv, '-proxy', proxy]
            else:
                env['BARNESS_REGIONAL_PROXY' if 'combo' in job else 'BARNESS_MINIMAX_PROXY' if job['name'].startswith('mm-') else 'BARNESS_AI_LIVE_PROXY'] = proxy
        env.update(BARNESS_AI_LIVE_ACCOUNT_ALIAS=scope['accountAlias'] + '@' + scope['regionAlias'], BARNESS_MINIMAX_ACCOUNT_ALIAS=scope['accountAlias'], BARNESS_MINIMAX_REGION_ALIAS=scope['regionAlias'], BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=scope['accountAlias'] + '@' + scope['regionAlias'])
        job = job | dict(executionProvenance={k: record[k] for k in ('validationRound', 'nativeRenewalId', 'spendingGuardPolicyId', 'spendingGuardSourceSHA256', 'usdLedger', 'originalCampaignApprovalSHA256', 'actualHumanApprovalSHA256')})
        job['executionProvenance'].update(admission)
        self.flow.s.guard(out)
        code = final_process.run(out, job, argv, env, intent)
        record = self.flow.s.doc(intent)
        record.update(commandReceiptSha256=self.flow.s.pin(out / (job['name'] + '-command.json')), budget=self.settle(out, job) if code == 0 else UNKNOWN.copy())
        if code == 0 and record['budget']['status'] == 'PASS':
            value = self.flow.s.doc(usd)
            if value.get('round') != self.flow.r.SPENDING_GUARD_POLICY_SELECTOR or value.get('maxMicrousd') != reserve or type(value.get('attempts')) is not int or (value['attempts'] < record['budget']['attempts']) or (type(value.get('reservedMicrousd')) is not int) or (not 0 < value['reservedMicrousd'] <= reserve):
                record['budget']['status'] = 'FAIL'
        self.flow.r.c.dump(intent, record)
        self.flow.s.guard(out)
        if code == 0 and record['budget']['status'] == 'PASS':
            try:
                record = self.flow.n.consume(out, job, record)
                self.flow.r.c.dump(intent, record)
            except Exception:
                record['budget']['status'] = 'FAIL'
                self.flow.r.c.dump(intent, record)
                raise
        if code != 0 or record['budget']['status'] != 'PASS':
            raise RuntimeError('process failed or unknown; preserve all recovery claims/USD and stop')
        return dict(route=job['name'], exitCode=code, budget=record['budget'], **self.flow.r.policy())
