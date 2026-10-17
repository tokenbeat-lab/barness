#!/usr/bin/env python3
"""One authorized TypeSafe round; no credential inheritance or paid retries."""
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

ROOT = Path(__file__).resolve().parents[3]
ENV = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('mode', choices=('prepare','live','gate'))
parser.add_argument('out', help='new repository-relative evidence directory')
parser.add_argument('--authorization-reference', default='ticket14-user-typesafe-20261017')
args = parser.parse_args()
if Path(args.out).is_absolute() or '..' in Path(args.out).parts:
    parser.error('repository-relative output required')
OUT = ROOT/args.out
SECRET = ''


def dump(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def sources():
    paths = list((ROOT/'ai').rglob('*.go'))
    paths += [ROOT/name for name in ('go.mod','go.sum','ai/release/catalog-snapshot.json',
        'ai/release/traceability.json','ai/examples/chineseeval/testdata/jev-route-v1.json',
        'ai/examples/visualeval/testdata/experiment.json')]
    return {str(p.relative_to(ROOT)):digest(p) for p in sorted(paths)}


def guard():
    if sources() != json.loads((OUT/'source-at-start.json').read_text())['files']:
        raise RuntimeError('source/config changed; no further execution authorized by this frozen round')


def stop(proc, first=signal.SIGTERM):
    with suppress(ProcessLookupError):
        os.killpg(proc.pid, first)
    try:
        proc.wait(timeout=15)
    except subprocess.TimeoutExpired:
        pass
    with suppress(ProcessLookupError):
        os.killpg(proc.pid, signal.SIGKILL)
    proc.wait()


def execute(name, argv, extra=None, cwd=ROOT, timeout=180, injected=None):
    receipt = OUT/(name+'-command.json')
    # Exclusive intent marker precedes the subprocess, including build/verify.
    # Crash recovery inspects retained state; it never silently resends a step.
    with receipt.open('x') as intent:
        json.dump({'state':'STARTING'}, intent)
    env = ENV | (extra or {})
    record = dict(name=name, argv=[str(x).replace(str(ROOT),'$REPO') for x in argv],
        cwd=str(cwd.relative_to(ROOT)) or '.', startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),
        timeoutSeconds=timeout, environmentKeys=sorted(env), credentialsInherited=False,
        injectedCredential=injected, exitCode=None, state='STARTING')
    began = time.monotonic()
    proc = None
    with tempfile.TemporaryFile() as captured:
        try:
            dump(receipt, record)
            print(name+': started', flush=True)
            proc = subprocess.Popen([str(x) for x in argv], env=env, cwd=cwd,
                stdout=captured, stderr=subprocess.STDOUT, start_new_session=True)
            try:
                record['exitCode'] = proc.wait(timeout=timeout)
                record['state'] = 'COMPLETED'
            except subprocess.TimeoutExpired:
                stop(proc)
                record.update(exitCode=124, state='TIMEOUT')
        except BaseException as error:
            if proc is not None:
                stop(proc, signal.SIGINT if isinstance(error, KeyboardInterrupt) else signal.SIGTERM)
            record.update(state='INTERRUPTED' if isinstance(error, KeyboardInterrupt) else 'ERROR',
                          exceptionType=type(error).__name__)
            raise
        finally:
            captured.seek(0)
            output = captured.read()
            if SECRET:
                output = output.replace(SECRET.encode(), b'[TYPESAFE-CREDENTIAL]')
            output = output.replace(str(ROOT).encode(), b'$REPO').replace(str(Path.home()).encode(), b'$HOME')
            (OUT/(name+'.log')).write_bytes(output)
            record['seconds'] = time.monotonic()-began
            dump(receipt, record)
    print(name+': exit '+str(record['exitCode']), flush=True)
    return record['exitCode']


def selected_secret():
    key = ''
    try:
        with (ROOT/'.env').open() as stream:
            for line in stream:
                name, sep, value = line.strip().removeprefix('export ').partition('=')
                if sep and name.strip() == 'TYPESAFE_KEY':
                    parts = shlex.split(value, comments=True)
                    if len(parts) == 1:
                        key = parts[0].strip()
                    break
    except (OSError, ValueError):
        pass
    if not key:
        raise RuntimeError('authorized TYPESAFE_KEY is unavailable')
    return key


def package(base):
    found = list(base.glob('*/manifest.json'))
    if len(found) != 1:
        raise RuntimeError('unique recorder package missing')
    return found[0].parent


def prepare():
    OUT.mkdir(parents=True, exist_ok=False, mode=0o700)
    dump(OUT/'source-at-start.json', {'gitCommit':subprocess.check_output(
        ['git','rev-parse','HEAD'],cwd=ROOT,env=ENV,text=True).strip(), 'files':sources()})
    cat = json.loads((ROOT/'ai/release/catalog-snapshot.json').read_text())
    proof = dict(route='jev',model='jev-1.14.0',service='https://api.typesafe.ai/v1/systemone',
        accountAlias='local-typesafe-unverified',regionAlias='unverified',
        authorizationReference=args.authorization_reference, dedicatedCredential=True,
        maxCalls=12,maxAttempts=12)
    dump(OUT/'protocol-authorization.json', proof)
    proof['reference'] = proof.pop('authorizationReference')
    proof['catalogHash'] = cat['hash']
    dump(OUT/'visual-authorization.json', proof)
    dump(OUT/'authorization-scope.json', dict(source='user instruction; root .env:TYPESAFE_KEY',
        providers=['typesafe'], models=['jev-1.13.0','jev-1.14.0'], maxCalls=54,maxAttempts=54,
        componentCalls={'protocol113':6,'text113':24,'protocol114':12,'visual114':12},
        accountRegionPermissions='UNVERIFIED', dollarCap=None, proxy=None,
        note='DedicatedCredential attests isolated injection only; no provider permission introspection.'))
    for name, argv in (
        ('build-live',['go','test','-tags','live','-c','-o',OUT/'live.test','./ai/live']),
        ('build-text',['go','build','-o',OUT/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
        ('build-visual',['go','build','-o',OUT/'visualeval','./ai/examples/visualeval'])):
        if execute(name, argv, timeout=180) != 0:
            raise RuntimeError('credential-free build failed')
    guard()
    dump(OUT/'prepared.json', {'state':'PREPARED','noCredentialsRead':True})


def live():
    global SECRET
    guard()
    if not (OUT/'prepared.json').is_file():
        raise RuntimeError('prepare must finish first')
    with (OUT/'paid-round-started.json').open('x') as marker:
        json.dump({'state':'STARTED','maxCalls':54,'noAutomaticRerun':True},marker)
    SECRET = selected_secret()
    alias = 'local-typesafe-unverified@region-unverified'
    steps = (
        ('protocol113', [OUT/'live.test','-test.run=^TestLive$/^typesafe-classifier$','-test.timeout=10m'],
         { 'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':'typesafe-classifier',
          'BARNESS_AI_LIVE_ACCOUNT_ALIAS':alias,'BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER':SECRET,
          'BARNESS_AI_EVIDENCE_DIR':str(OUT/'protocol113')}, ROOT/'ai/live',660,
         'BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER'),
        ('text113',[OUT/'chineseeval','-route','jev','-live','-out',OUT/'text-jev'],
         {'BARNESS_AI_CHINESE_EVAL':'1','BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY':SECRET,
          'BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS':alias},ROOT,630,'BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY'),
        ('protocol114',[OUT/'live.test','-test.run=^TestMultimodalFormal$','-test.timeout=180s'],
         {'BARNESS_MM_ROUTE':'jev','BARNESS_MM_FORMAL':'1','BARNESS_MM_LIVE':'1','BARNESS_MM_KEY':SECRET,
          'BARNESS_MM_AUTHORIZATION':str(OUT/'protocol-authorization.json'),
          'BARNESS_AI_EVIDENCE_DIR':str(OUT/'protocol114')},ROOT/'ai/live',210,'BARNESS_MM_KEY'),
        ('visual114',[OUT/'visualeval','-formal','-mode','live','-route','jev','-entry','plain',
          '-authorization',OUT/'visual-authorization.json','-out',OUT/'visual-jev'],
         {'BARNESS_VISUAL_KEY':SECRET},ROOT,270,'BARNESS_VISUAL_KEY'))
    codes = {}
    for name, argv, env, cwd, timeout, injected in steps:
        guard()
        codes[name] = execute(name,argv,env,cwd,timeout,injected)
    SECRET = ''
    guard()
    protocol = package(OUT/'protocol114')
    text = package(OUT/'text-jev')
    codes['verify-text'] = execute('verify-text',[OUT/'chineseeval','-route','jev','-verify',text])
    codes['verify-visual'] = execute('verify-visual',[OUT/'visualeval','-verify',OUT/'visual-jev','-require-live'])
    codes['support'] = execute('support',['go','run','./ai/live/cmd/jevformal','-protocol',protocol,
        '-effects',OUT/'visual-jev','-out',OUT/'support','-require-live'])
    if codes['support'] == 0:
        codes['verify-support'] = execute('verify-support',['go','run','./ai/live/cmd/jevformal',
            '-verify',OUT/'support','-require-live'])
    # Only audited, accepting reports are merged; failed siblings stay retained.
    reports = []
    if codes['protocol113'] == 0:
        reports.append(package(OUT/'protocol113'))
    if codes.get('verify-support') == 0:
        reports.append(OUT/'support')
    if reports:
        (OUT/'matrix-before.json').write_bytes((ROOT/'ai/live/support-matrix.json').read_bytes())
        codes['matrix-merge'] = execute('matrix-merge',['go','run','./ai/live/cmd/supportmatrix',*reports])
        (OUT/'matrix-after.json').write_bytes((ROOT/'ai/live/support-matrix.json').read_bytes())
    dump(OUT/'live-summary.json',{'commands':codes,'allCommandsSucceeded':all(v==0 for v in codes.values()),
        'noOtherProviderCredentialRead':True,'accountRegionPermissions':'UNVERIFIED'})
    if not all(v == 0 for v in codes.values()):
        raise SystemExit(1)


def gate():
    guard()
    with (OUT/'gate-started.json').open('x') as marker:
        json.dump({'state':'STARTED','providerCredentials':False},marker)
    code = execute('no-key-verification',['python3',ROOT/'.scratch/jev-1.14-multimodal-box/run-14-verification.py',
        str((OUT/'no-key').relative_to(ROOT))],timeout=900)
    if code != 0:
        raise RuntimeError('credential-free producer verification failed')
    inputs = []
    for route in json.loads((ROOT/'ai/release/traceability.json').read_text())['liveCombos']:
        chosen = package(OUT/'no-key/preflight'/route)
        if route == 'typesafe-classifier':
            chosen = package(OUT/'protocol113')
        elif route == 'typesafe-classifier-jev-1.14.0':
            chosen = OUT/'support' if (OUT/'support/live-report.json').is_file() else package(OUT/'protocol114')
        inputs.append(chosen)
    command = ['go','run','./ai/release/cmd/releasegate','-out',str((OUT/'gate').relative_to(ROOT)),
        '-evaluation-jev',package(OUT/'text-jev'),'-evaluation-luna',package(OUT/'no-key/text-luna'),
        '-visual-evaluation-jev',OUT/'visual-jev','-visual-evaluation-decisions',OUT/'no-key/visual-decisions']
    for path in inputs:
        command += ['-live',path]
    guard()
    execute('releasegate',command,timeout=3600)
    guard()
    report = json.loads((OUT/'gate/release-report.json').read_text())
    dump(OUT/'gate-inputs.json',{'live':[str(p.relative_to(ROOT)) for p in inputs],
        'realTypeSafeOnly':True,'overall':report['pass'],'collectionExitDoesNotMeanReleasePass':True})


if args.mode == 'prepare':
    prepare()
elif args.mode == 'live':
    live()
else:
    gate()
