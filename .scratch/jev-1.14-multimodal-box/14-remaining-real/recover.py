#!/usr/bin/env python3
"""The one reserved post-fix full round; retain every initial verdict."""
import argparse
import importlib.util
import json
from pathlib import Path
import subprocess
import sys

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('remaining', HERE/'run.py')
r = importlib.util.module_from_spec(spec)
spec.loader.exec_module(r)
ROOT = r.ROOT
INITIAL = ROOT/'.evidence/ticket14-remaining-20261017073529Z'
CAPS = r.CAPS | {'protocol113':6,'text-jev':24,'protocol114':12,'visual-jev':12}


class Recovery:
    def __init__(self, out):
        self.commands = r.Round(out)
        self.out = self.commands.out
        self.execute = self.commands.execute

    def guard(self):
        if r.sources() != r.doc(self.out/'source-at-start.json')['files']:
            raise RuntimeError('recovery source/config changed; no further paid execution')
        reservation = r.doc(INITIAL/'recovery-reservation.json')
        if reservation['recoveryRoot'] != str(self.out.relative_to(ROOT)):
            raise RuntimeError('the single recovery round is reserved elsewhere')
        if reservation['initialConsumptionSHA256'] != r.digest(INITIAL/'consumption.json'):
            raise RuntimeError('initial consumption changed')

    def prepare(self):
        total = r.doc(INITIAL/'consumption.json')['total']
        if max(total['conservativeCalls'],total['conservativeAttempts']) > 248:
            raise RuntimeError('initial conservative budget is invalid')
        if any(r.doc(INITIAL/(name+'-command.json'))['state'] != 'COMPLETED' for name in r.CAPS):
            raise RuntimeError('initial paid jobs must have finished before recovery')
        self.out.mkdir(parents=True,exist_ok=False,mode=0o700)
        reservation = dict(recoveryRoot=str(self.out.relative_to(ROOT)),initialConsumption=total,
            initialConsumptionSHA256=r.digest(INITIAL/'consumption.json'),maxCalls=302,maxAttempts=302,
            taskMaxCalls=550,taskMaxAttempts=550,recoveryRounds=1)
        with (INITIAL/'recovery-reservation.json').open('x') as marker:
            json.dump(reservation,marker,indent=2)
        r.dump(self.out/'authorization-scope.json',reservation | dict(
            reason='fixed optional direct transport; one full post-fix evidence round',
            providerCaps={'openai':116,'anthropic':32,'google':36,'deepseek':64,'typesafe':54},
            accountRegionPermissions='UNVERIFIED',dollarCap=None,proxy=None))
        r.dump(self.out/'source-at-start.json',dict(files=r.sources(),gitCommit=subprocess.check_output(
            ['git','rev-parse','HEAD'],cwd=ROOT,env=r.ENV,text=True).strip()))
        cat = r.doc(ROOT/'ai/release/catalog-snapshot.json')
        for route,model,provider,service in (
            ('decisions','gpt-6-luna','openai','https://api.openai.com/v1/decisions'),
            ('jev','jev-1.14.0','typesafe','https://api.typesafe.ai/v1/systemone')):
            proof = dict(route=route,model=model,service=service,catalogHash=cat['hash'],
                accountAlias=provider+'-local-unverified',regionAlias='unverified',
                reference='ticket14-recovery-20261017',dedicatedCredential=True,maxCalls=12,maxAttempts=12)
            r.dump(self.out/('visual-'+route+'-authorization.json'),proof)
            if route == 'jev':
                proof.pop('catalogHash')
                proof['authorizationReference'] = proof.pop('reference')
                r.dump(self.out/'protocol-authorization.json',proof)
        for name,argv in (
            ('build-live',['go','test','-tags','live','-c','-o',self.out/'live.test','./ai/live']),
            ('build-text',['go','build','-o',self.out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
            ('build-visual',['go','build','-o',self.out/'visualeval','./ai/examples/visualeval'])):
            if self.execute(name,argv,timeout=180) != 0:
                raise RuntimeError('credential-free build failed')
        self.guard()
        r.dump(self.out/'prepared.json',dict(state='PREPARED',credentialsRead=False))

    def check(self):
        self.guard()
        code = self.execute('live-harness',['go','test','-tags','live','-count=1','./ai/live'],
            extra={'BARNESS_AI_EVIDENCE_DIR':str(self.out/'live-harness')},timeout=900)
        self.guard()
        r.dump(self.out/'live-harness-accepted.json',dict(command='live-harness',exitCode=code))
        if code != 0:
            raise SystemExit(code)

    def jobs(self):
        # Retain the initial launcher as executed history. Only orchestration is
        # repeated here: this round has a new complete source and 302-call cap.
        for route,provider,key,cap in r.ROUTES:
            extra = {'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':route,
                'BARNESS_AI_LIVE_ACCOUNT_ALIAS':provider+'-local-unverified@region-unverified',
                'BARNESS_AI_EVIDENCE_DIR':str(self.out/'live'/route)}
            formal = route == 'openai-decisions'
            if formal:
                extra['BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT'] = '1'
            selector = '^TestFormalDecisions$' if formal else '^TestLive$/^'+route+'$'
            limit,timeout = ('180s',210) if formal else (('10m',660) if cap==4 else ('30m',2100))
            yield route,[self.out/'live.test','-test.run='+selector,'-test.timeout='+limit],dict(
                extra=extra,timeout=timeout,cwd=ROOT/'ai/live',source=key,
                injected='BARNESS_AI_LIVE_KEY_'+route.upper().replace('-','_'))
        yield 'protocol113',[self.out/'live.test','-test.run=^TestLive$/^typesafe-classifier$','-test.timeout=10m'],dict(
            extra={'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':'typesafe-classifier',
                'BARNESS_AI_LIVE_ACCOUNT_ALIAS':'typesafe-local-unverified@region-unverified',
                'BARNESS_AI_EVIDENCE_DIR':str(self.out/'protocol113')},cwd=ROOT/'ai/live',timeout=660,
            source='TYPESAFE_KEY',injected='BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER')
        yield 'protocol114',[self.out/'live.test','-test.run=^TestMultimodalFormal$','-test.timeout=180s'],dict(
            extra={'BARNESS_MM_ROUTE':'jev','BARNESS_MM_FORMAL':'1','BARNESS_MM_LIVE':'1',
                'BARNESS_MM_AUTHORIZATION':str(self.out/'protocol-authorization.json'),
                'BARNESS_AI_EVIDENCE_DIR':str(self.out/'protocol114')},cwd=ROOT/'ai/live',timeout=210,
            source='TYPESAFE_KEY',injected='BARNESS_MM_KEY')
        for route,key,injected in (('luna','OPENAI_KEY','BARNESS_AI_CHINESE_EVAL_OPENAI_KEY'),
                                    ('jev','TYPESAFE_KEY','BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY')):
            yield 'text-'+route,[self.out/'chineseeval','-route',route,'-live','-out',self.out/('text-'+route)],dict(
                extra={'BARNESS_AI_CHINESE_EVAL':'1','BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS':
                    ('openai' if route=='luna' else 'typesafe')+'-local-unverified@region-unverified'},
                timeout=630,source=key,injected=injected)
        for route,key in (('decisions','OPENAI_KEY'),('jev','TYPESAFE_KEY')):
            yield 'visual-'+route,[self.out/'visualeval','-formal','-mode','live','-route',route,'-entry','plain',
                '-authorization',self.out/('visual-'+route+'-authorization.json'),'-out',self.out/('visual-'+route)],dict(
                timeout=270,source=key,injected='BARNESS_VISUAL_KEY')

    def summarize(self):
        # The first nine supports and two OpenAI effects retain their own 248
        # ceiling, independently from the additional 54-call TypeSafe envelope.
        self.commands.summarize()
        rows = r.doc(self.out/'consumption.json')['routes']
        for name,cap in (('protocol113',6),('protocol114',12),('text-jev',24),('visual-jev',12)):
            path = self.out/'visual-jev/report.json' if name=='visual-jev' else None
            pattern = name+'/*/'+('evaluation-report.json' if name=='text-jev' else
                'multimodal-report.json' if name=='protocol114' else 'live-report.json')
            if path is None:
                found = list(self.out.glob(pattern))
                path = found[0] if len(found)==1 else None
            if path is None or not path.is_file():
                rows.append(dict(name=name,actual='UNKNOWN',conservativeCalls=cap,conservativeAttempts=cap,maxCalls=cap))
                continue
            report = r.doc(path)
            b = report['budget']
            actual = b['actual'] if name=='protocol114' else b
            calls,attempts = ((actual['callsUsed'],actual['httpAttempts']) if name=='protocol113' else
                (actual['calls'],actual['attempts']))
            if not 0 <= calls <= cap or not 0 <= attempts <= cap:
                raise RuntimeError('TypeSafe component budget exceeded')
            rows.append(dict(name=name,calls=calls,attempts=attempts,maxCalls=cap,
                conservativeCalls=calls,conservativeAttempts=attempts,status=report.get('status',report.get('outcome'))))
        total = {k:sum(row[k] for row in rows) for k in ('conservativeCalls','conservativeAttempts')}
        total.update(calls=sum(row.get('calls',0) for row in rows),attempts=sum(row.get('attempts',0) for row in rows),
            maxCalls=302,maxAttempts=302)
        if max(total['conservativeCalls'],total['conservativeAttempts']) > 302:
            raise RuntimeError('recovery budget exceeded')
        initial = r.doc(INITIAL/'consumption.json')['total']
        task = {k:initial[k]+total[k] for k in ('calls','attempts','conservativeCalls','conservativeAttempts')}
        if max(task['conservativeCalls'],task['conservativeAttempts']) > 550:
            raise RuntimeError('task budget exceeded')
        task.update(maxCalls=550,maxAttempts=550,recoveryRounds=1)
        r.dump(self.out/'consumption-all.json',dict(routes=rows,total=total))
        r.dump(self.out/'task-consumption.json',task)
        return total

    def live(self):
        self.guard()
        accepted = r.doc(self.out/'live-harness-accepted.json')
        if (accepted['command'] not in ('live-harness','live-harness-recheck') or accepted['exitCode']!=0
                or r.doc(self.out/(accepted['command']+'-command.json'))['exitCode']!=0):
            raise RuntimeError('complete no-key live harness must pass before paid recovery')
        with (self.out/'paid-round-started.json').open('x') as marker:
            json.dump(dict(maxCalls=302,maxAttempts=302,noAutomaticResend=True),marker)
        codes = {}
        for name,argv,options in self.jobs():
            self.guard()
            codes[name] = self.execute(name,argv,**options)
        self.guard()
        total = self.summarize()
        for route in ('luna','jev'):
            found = list((self.out/('text-'+route)).glob('*/manifest.json'))
            codes['verify-text-'+route] = (self.execute('verify-text-'+route,
                [self.out/'chineseeval','-route',route,'-verify',found[0].parent]) if len(found)==1 else None)
        for route in ('decisions','jev'):
            codes['verify-visual-'+route] = self.execute('verify-visual-'+route,
                [self.out/'visualeval','-verify',self.out/('visual-'+route),'-require-live'])
        found = list((self.out/'protocol114').glob('*/manifest.json'))
        if len(found)==1:
            codes['support'] = self.execute('support',['go','run','./ai/live/cmd/jevformal',
                '-protocol',found[0].parent,'-effects',self.out/'visual-jev','-out',self.out/'support','-require-live'])
            if codes['support']==0:
                codes['verify-support'] = self.execute('verify-support',['go','run','./ai/live/cmd/jevformal',
                    '-verify',self.out/'support','-require-live'])
        else:
            codes['support'] = None
        passed = all(v==0 for v in codes.values())
        r.dump(self.out/'live-summary.json',dict(commands=codes,total=total,allCommandsSucceeded=passed))
        if not passed:
            raise SystemExit(1)

    def live_inputs(self):
        result = []
        for route in r.doc(ROOT/'ai/release/traceability.json')['liveCombos']:
            result.append(self.out/'support' if route=='typesafe-classifier-jev-1.14.0' else
                r.bundle(self.out/'protocol113') if route=='typesafe-classifier' else r.bundle(self.out/'live'/route))
        return result

    def merge(self):
        self.guard()
        if not r.doc(self.out/'live-summary.json')['allCommandsSucceeded']:
            raise RuntimeError('all strict producers must pass before matrix merge')
        with (self.out/'merge-started.json').open('x') as marker:
            json.dump({'wholeBatchAtomic':True},marker)
        matrix = ROOT/'ai/live/support-matrix.json'
        (self.out/'matrix-before.json').write_bytes(matrix.read_bytes())
        code = self.execute('matrix-merge',['go','run','./ai/live/cmd/supportmatrix',*self.live_inputs()])
        (self.out/'matrix-after.json').write_bytes(matrix.read_bytes())
        if code != 0:
            raise SystemExit(code)

    def gate(self):
        self.guard()
        args = ['go','run','./ai/release/cmd/releasegate','-out',self.out/'gate',
            '-evaluation-jev',r.bundle(self.out/'text-jev'),'-evaluation-luna',r.bundle(self.out/'text-luna'),
            '-visual-evaluation-jev',self.out/'visual-jev','-visual-evaluation-decisions',self.out/'visual-decisions']
        for path in self.live_inputs():
            args += ['-live',path]
        code = self.execute('releasegate',args,timeout=3600)
        self.guard()
        r.dump(self.out/'gate-summary.json',dict(exitCode=code,passVerdict=r.doc(self.out/'gate/release-report.json')['pass']))
        if code != 0:
            raise SystemExit(code)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('prepare','check','live','merge','gate'))
    parser.add_argument('out')
    args = parser.parse_args()
    if Path(args.out).is_absolute() or '..' in Path(args.out).parts:
        parser.error('new repository-relative evidence root required')
    getattr(Recovery(args.out),args.mode)()


if __name__ == '__main__':
    main()
