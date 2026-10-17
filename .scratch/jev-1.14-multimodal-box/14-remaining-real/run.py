#!/usr/bin/env python3
"""One bounded remaining-provider round; credentials never enter argv or builds."""
import argparse
from contextlib import suppress
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import shlex
import signal
import subprocess
import tempfile
import time

# Independent from the sealed TypeSafe launcher: its bytes/receipts are history,
# and this round has different identities, budgets and actual failure handling.
ROOT = Path(__file__).resolve().parents[3]
TYPESAFE = ROOT/'.evidence/ticket14-typesafe-r2-20261017T061932.423266Z'
ENV = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}
ROUTES = (
    ('openai-responses','openai','OPENAI_KEY',32),
    ('openai-chat','openai','OPENAI_KEY',32),
    ('openai-images','openai','OPENAI_KEY',4),
    ('openai-decisions','openai','OPENAI_KEY',12),
    ('anthropic-messages','anthropic','ANTHROPIC_KEY',32),
    ('google-gemini','google','GEMINI_KEY',32),
    ('google-interactions-image','google','GEMINI_KEY',4),
    ('deepseek-responses','deepseek','DEEPSEEK_KEY',32),
    ('deepseek-chat','deepseek','DEEPSEEK_KEY',32))
CAPS = {r[0]:r[3] for r in ROUTES}|{'text-luna':24,'visual-decisions':12}


def doc(path):
    return json.loads(path.read_text())


def dump(path, value):
    path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def sources():
    paths = list((ROOT/'ai').rglob('*.go'))
    paths += [ROOT/p for p in ('go.mod','go.sum','ai/release/catalog-snapshot.json',
        'ai/release/traceability.json','ai/examples/chineseeval/testdata/jev-route-v1.json',
        'ai/examples/chineseeval/testdata/luna-route-v1.json','ai/examples/visualeval/testdata/experiment.json')]
    return {str(p.relative_to(ROOT)):digest(p) for p in sorted(paths)}


def bundle(base):
    found = list(base.glob('*/manifest.json'))
    if len(found) != 1:
        raise RuntimeError('unique recorder package missing')
    return found[0].parent


class Round:
    def __init__(self, out):
        self.out = ROOT/out

    def guard(self):
        current = sources()
        if current != doc(self.out/'source-at-start.json')['files']:
            raise RuntimeError('frozen source/config changed; paid execution stopped')
        if any(current.get(p) != h for p,h in doc(TYPESAFE/'source-at-start.json')['files'].items()):
            raise RuntimeError('existing TypeSafe evidence expired; a separately bounded fresh round is required')

    @staticmethod
    def secret(variable):
        with (ROOT/'.env').open() as stream:
            for line in stream:
                name,sep,value = line.strip().removeprefix('export ').partition('=')
                if sep and name.strip() == variable:
                    try:
                        parts = shlex.split(value,comments=True)
                    except ValueError:
                        break
                    if len(parts) == 1 and parts[0].strip():
                        return parts[0].strip()
                    break
        raise RuntimeError('selected credential variable is unavailable')

    @staticmethod
    def stop(proc, first=signal.SIGTERM):
        with suppress(ProcessLookupError):
            os.killpg(proc.pid,first)
        try:
            proc.wait(timeout=15)
        except subprocess.TimeoutExpired:
            pass
        with suppress(ProcessLookupError):
            os.killpg(proc.pid,signal.SIGKILL)
        proc.wait()

    def execute(self, name, argv, extra=None, timeout=180, cwd=ROOT, source=None, injected=None):
        receipt = self.out/(name+'-command.json')
        with receipt.open('x') as intent:
            json.dump({'state':'STARTING'},intent)
        secret = ''
        env = ENV | (extra or {})
        record = dict(name=name,argv=[str(v).replace(str(ROOT),'$REPO') for v in argv],
            cwd=str(cwd.relative_to(ROOT)) or '.',environmentKeys=sorted(env),credentialsInherited=False,
            sourceVariable=source,injectedCredential=injected,timeoutSeconds=timeout,
            startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),state='STARTING',exitCode=None)
        began, proc = time.monotonic(), None
        with tempfile.TemporaryFile() as captured:
            try:
                if source:
                    if not injected:
                        raise RuntimeError('credential injection target missing')
                    secret = self.secret(source)
                    env[injected] = secret
                    record['environmentKeys'] = sorted(env)
                dump(receipt,record)
                print(name+': started',flush=True)
                proc = subprocess.Popen([str(v) for v in argv],cwd=cwd,env=env,
                    stdout=captured,stderr=subprocess.STDOUT,start_new_session=True)
                try:
                    record.update(exitCode=proc.wait(timeout=timeout),state='COMPLETED')
                except subprocess.TimeoutExpired:
                    self.stop(proc)
                    record.update(exitCode=124,state='TIMEOUT',terminatedProcessExitCode=proc.returncode)
            except BaseException as error:
                if proc is not None:
                    self.stop(proc,signal.SIGINT if isinstance(error,KeyboardInterrupt) else signal.SIGTERM)
                record.update(state='INTERRUPTED' if isinstance(error,KeyboardInterrupt) else 'ERROR',
                    exceptionType=type(error).__name__,terminatedProcessExitCode=proc.returncode if proc else None)
                raise
            finally:
                captured.seek(0)
                output = captured.read()
                if secret:
                    output = output.replace(secret.encode(),b'[SELECTED-CREDENTIAL]')
                output = output.replace(str(ROOT).encode(),b'$REPO').replace(str(Path.home()).encode(),b'$HOME')
                (self.out/(name+'.log')).write_bytes(output)
                record['seconds'] = time.monotonic()-began
                dump(receipt,record)
                env.pop(injected,None)
                secret = ''
        print(name+': exit '+str(record['exitCode']),flush=True)
        return record['exitCode']

    def prepare(self):
        self.out.mkdir(parents=True,exist_ok=False,mode=0o700)
        dump(self.out/'source-at-start.json',dict(files=sources(),gitCommit=subprocess.check_output(
            ['git','rev-parse','HEAD'],cwd=ROOT,env=ENV,text=True).strip()))
        self.guard()
        dump(self.out/'authorization-scope.json',dict(reference='ticket14-remaining-20261017-r1',
            source='direct user instruction; selected root .env variables',initialMaxCalls=248,initialMaxAttempts=248,
            taskMaxCalls=550,taskMaxAttempts=550,recoveryMaxCalls=302,recoveryRounds=1,
            providerInitialCaps={'openai':116,'anthropic':32,'google':36,'deepseek':64},
            accountRegionPermissions='UNVERIFIED',dollarCap=None,proxy=None,
            routes=[dict(route=r,provider=p,sourceVariable=k,maxCalls=n,maxAttempts=n) for r,p,k,n in ROUTES],
            effects={'text-luna':24,'visual-decisions':12},existingTypeSafe=str(TYPESAFE.relative_to(ROOT))))
        dump(self.out/'visual-authorization.json',dict(route='decisions',model='gpt-6-luna',
            service='https://api.openai.com/v1/decisions',catalogHash=doc(ROOT/'ai/release/catalog-snapshot.json')['hash'],
            accountAlias='openai-local-unverified',regionAlias='unverified',reference='ticket14-remaining-20261017-r1',
            dedicatedCredential=True,maxCalls=12,maxAttempts=12))
        for name,argv in (
            ('build-live',['go','test','-tags','live','-c','-o',self.out/'live.test','./ai/live']),
            ('build-text',['go','build','-o',self.out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
            ('build-visual',['go','build','-o',self.out/'visualeval','./ai/examples/visualeval'])):
            if self.execute(name,argv,timeout=180) != 0:
                raise RuntimeError('credential-free build failed')
        self.guard()
        dump(self.out/'prepared.json',dict(state='PREPARED',credentialsRead=False))

    def summarize(self):
        rows = []
        for name,cap in CAPS.items():
            path = None
            if name == 'visual-decisions':
                path = self.out/'visual-decisions/report.json'
            else:
                pattern = 'text-luna/*/evaluation-report.json' if name == 'text-luna' else 'live/'+name+'/*/live-report.json'
                found = list(self.out.glob(pattern))
                if len(found) == 1:
                    path = found[0]
            if path is None or not path.exists():
                rows.append(dict(name=name,maxCalls=cap,actual='UNKNOWN',conservativeCalls=cap,conservativeAttempts=cap))
                continue
            r = doc(path)
            b = r['budget']
            calls,attempts = (b['calls'],b['attempts']) if name in ('text-luna','visual-decisions') else (b['callsUsed'],b['httpAttempts'])
            if not 0 <= calls <= cap or not 0 <= attempts <= cap:
                raise RuntimeError('actual route budget exceeded')
            rows.append(dict(name=name,calls=calls,attempts=attempts,maxCalls=cap,
                conservativeCalls=calls,conservativeAttempts=attempts,status=r.get('status'),
                scenarios=[{'id':s['id'],'outcome':s['outcome']} for s in r.get('scenarios',[])]))
        totals = dict(calls=sum(r.get('calls',0) for r in rows),attempts=sum(r.get('attempts',0) for r in rows),
            conservativeCalls=sum(r['conservativeCalls'] for r in rows),
            conservativeAttempts=sum(r['conservativeAttempts'] for r in rows),maxCalls=248,maxAttempts=248)
        if max(totals['conservativeCalls'],totals['conservativeAttempts']) > 248:
            raise RuntimeError('bounded initial round exceeded')
        dump(self.out/'consumption.json',dict(routes=rows,total=totals))
        return totals

    def live(self):
        self.guard()
        if not (self.out/'prepared.json').is_file():
            raise RuntimeError('prepare required')
        with (self.out/'paid-round-started.json').open('x') as marker:
            json.dump(dict(maxCalls=248,maxAttempts=248,noAutomaticResend=True),marker)
        codes = {}
        for route,provider,key,cap in ROUTES:
            self.guard()
            injected = 'BARNESS_AI_LIVE_KEY_'+route.upper().replace('-','_')
            extra = {'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':route,
                'BARNESS_AI_LIVE_ACCOUNT_ALIAS':provider+'-local-unverified@region-unverified',
                'BARNESS_AI_EVIDENCE_DIR':str(self.out/'live'/route)}
            formal = route == 'openai-decisions'
            if formal:
                extra['BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT'] = '1'
            selector = '^TestFormalDecisions$' if formal else '^TestLive$/^'+route+'$'
            limit,timeout = ('180s',210) if formal else (('10m',660) if cap==4 else ('30m',2100))
            codes[route] = self.execute(route,[self.out/'live.test','-test.run='+selector,'-test.timeout='+limit],
                extra=extra,timeout=timeout,cwd=ROOT/'ai/live',source=key,injected=injected)
        self.guard()
        codes['text-luna'] = self.execute('text-luna',[self.out/'chineseeval','-route','luna','-live','-out',self.out/'text-luna'],
            extra={'BARNESS_AI_CHINESE_EVAL':'1','BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS':'openai-local-unverified@region-unverified'},
            timeout=630,source='OPENAI_KEY',injected='BARNESS_AI_CHINESE_EVAL_OPENAI_KEY')
        self.guard()
        codes['visual-decisions'] = self.execute('visual-decisions',[self.out/'visualeval','-formal','-mode','live',
            '-route','decisions','-entry','plain','-authorization',self.out/'visual-authorization.json','-out',self.out/'visual-decisions'],
            timeout=270,source='OPENAI_KEY',injected='BARNESS_VISUAL_KEY')
        self.guard()
        totals = self.summarize()
        found = list((self.out/'text-luna').glob('*/manifest.json'))
        codes['verify-text'] = (self.execute('verify-text',[self.out/'chineseeval','-route','luna','-verify',found[0].parent])
            if len(found)==1 else None)
        codes['verify-visual'] = self.execute('verify-visual',[self.out/'visualeval','-verify',self.out/'visual-decisions','-require-live'])
        dump(self.out/'live-summary.json',dict(commands=codes,total=totals,allCommandsSucceeded=all(v==0 for v in codes.values())))
        if not all(v==0 for v in codes.values()):
            raise SystemExit(1)

    def merge(self):
        self.guard()
        with (self.out/'merge-started.json').open('x') as marker:
            json.dump({'noAutomaticReplay':True},marker)
        (self.out/'matrix-before.json').write_bytes((ROOT/'ai/live/support-matrix.json').read_bytes())
        codes = {}
        for route,_,_,_ in ROUTES:
            if len(list((self.out/'live'/route).glob('*/manifest.json'))) != 1:
                continue
            path = bundle(self.out/'live'/route)
            if not (path/'live-report.json').is_file() or not (path/'audit.json').is_file():
                continue
            r = doc(path/'live-report.json')
            command = doc(self.out/(route+'-command.json'))
            accepting = (command['exitCode']==0 and not doc(path/'audit.json')['findings'] and
                r['scenarios'] and all(s['outcome'] in ('PASS','UNSUPPORTED') for s in r['scenarios']))
            if accepting:
                codes[route] = self.execute('merge-'+route,['go','run','./ai/live/cmd/supportmatrix',path])
        (self.out/'matrix-after.json').write_bytes((ROOT/'ai/live/support-matrix.json').read_bytes())
        dump(self.out/'merge-summary.json',dict(commands=codes,allAccepted=len(codes)==9 and all(v==0 for v in codes.values())))

    def gate(self):
        self.guard()
        args = ['go','run','./ai/release/cmd/releasegate','-out',self.out/'gate',
            '-evaluation-jev',bundle(TYPESAFE/'text-jev'),'-evaluation-luna',bundle(self.out/'text-luna'),
            '-visual-evaluation-jev',TYPESAFE/'visual-jev','-visual-evaluation-decisions',self.out/'visual-decisions']
        for route in doc(ROOT/'ai/release/traceability.json')['liveCombos']:
            path = (TYPESAFE/'support' if route=='typesafe-classifier-jev-1.14.0' else
                    bundle(TYPESAFE/'protocol113') if route=='typesafe-classifier' else bundle(self.out/'live'/route))
            args += ['-live',path]
        code = self.execute('releasegate',args,timeout=3600)
        self.guard()
        dump(self.out/'gate-summary.json',dict(exitCode=code,passVerdict=doc(self.out/'gate/release-report.json')['pass']))
        if code != 0:
            raise SystemExit(code)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('prepare','live','merge','gate'))
    parser.add_argument('out')
    args = parser.parse_args()
    if Path(args.out).is_absolute() or '..' in Path(args.out).parts:
        parser.error('new repository-relative evidence root required')
    getattr(Round(args.out),args.mode)()


if __name__ == '__main__':
    main()
