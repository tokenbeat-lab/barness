#!/usr/bin/env python3
"""Reproduce ticket 14 without inheriting route credentials or making paid calls."""
import argparse
from contextlib import suppress
import datetime as dt
import json
import os
from pathlib import Path
import subprocess
import signal
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('out', help='new repository-relative output directory')
parser.add_argument('--command-timeout', type=float, default=3600, help='positive overall command cap in seconds (at most 3600)')
parser.add_argument('--gate', action='store_true', help='run full release command after preflights')
a = parser.parse_args()
if not 0 < a.command_timeout <= 3600: parser.error("command timeout must be positive and at most 3600 seconds")
OUT = ROOT / a.out
OUT.mkdir(parents=True, exist_ok=False)
ENV = {k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}
commands = []


def execute(name, argv, cwd=ROOT, extra=None, timeout=1800):
    env = ENV | (extra or {})
    started = dt.datetime.now(dt.timezone.utc).isoformat()
    clock = time.monotonic()
    print(f'{name}: started', flush=True)
    timed_out = False
    # A temporary file decouples output collection from the process deadline;
    # no blocking pipe read may defer wait(timeout) until after child exit.
    with tempfile.TemporaryFile(mode='w+t') as captured:
        proc = subprocess.Popen([str(v) for v in argv], cwd=cwd, env=env,
                                stdout=captured, stderr=subprocess.STDOUT,
                                start_new_session=True)
        try:
            code = proc.wait(timeout=min(timeout, a.command_timeout))
        except subprocess.TimeoutExpired:
            timed_out = True
            with suppress(ProcessLookupError):
                os.killpg(proc.pid, signal.SIGTERM)
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                pass
            # Reap stubborn descendants even if the direct parent exited.
            with suppress(ProcessLookupError):
                os.killpg(proc.pid, signal.SIGKILL)
            proc.wait()
            code = 124
        captured.seek(0)
        with (OUT / (name + '.log')).open('w') as log:
            for line in captured:
                log.write(line.replace(str(ROOT), '$REPO').replace(str(Path.home()), '$HOME'))
    record = dict(name=name, argv=[str(v).replace(str(ROOT), '$REPO') for v in argv],
                  cwd=str(cwd.relative_to(ROOT)) or '.', startedAt=started,
                  seconds=time.monotonic()-clock, exitCode=code, timedOut=timed_out,
                  environmentKeys=sorted(env), credentialsInherited=False)
    commands.append(record)
    (OUT / 'commands.json').write_text(json.dumps(commands, indent=2)+'\n')
    print(f'{name}: exit {code}', flush=True)
    if name.startswith('build-') and code != 0: raise SystemExit(code)
    return code


execute('build-live', ['go','test','-tags','live','-c','-o',OUT/'live.test','./ai/live'])
execute('build-text-effects', ['go','build','-o',OUT/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval'])
execute('build-visual-effects', ['go','build','-o',OUT/'visualeval','./ai/examples/visualeval'])
execute('vet', ['go','vet','./...'])
execute('vet-live', ['go','vet','-tags','live','./...'])
routes = json.loads((ROOT/'ai/release/traceability.json').read_text())['liveCombos']
live = []
for route in routes:
    base = OUT / 'preflight' / route
    extra = {'BARNESS_AI_EVIDENCE_DIR':str(base)}
    test = '^TestLive$'
    if route == 'typesafe-classifier-jev-1.14.0':
        extra |= {'BARNESS_MM_ROUTE':'jev','BARNESS_MM_FORMAL':'1'}
        test = '^TestMultimodalFormal$'
    else:
        extra |= {'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':route}
        if route == 'openai-decisions': test = '^TestFormalDecisions$'
    execute('preflight-'+route, [OUT/'live.test','-test.run='+test,'-test.timeout=180s'],
            cwd=ROOT/'ai/live', extra=extra)
    bundles = list(base.glob('*/manifest.json'))
    if len(bundles) != 1: raise RuntimeError('unique route preflight bundle missing')
    live.append(str(bundles[0].parent.relative_to(ROOT)))
for route in ('jev','luna'):
    execute('text-effects-'+route, [OUT/'chineseeval','-route',route,'-out',OUT/('text-'+route)])
for route in ('jev','decisions'):
    formal = ['-formal']
    execute('visual-effects-'+route, [OUT/'visualeval',*formal,'-mode','live','-route',route,
                                     '-out',OUT/('visual-'+route)])
independent = []
for kind, routes in (('text',('jev','luna')),('visual',('jev','decisions'))):
    for route in routes:
        if kind == 'text':
            manifests = list((OUT/('text-'+route)).glob('*/manifest.json'))
            if len(manifests) != 1: raise RuntimeError('unique text-effect package missing')
            argv = [OUT/'chineseeval','-route',route,'-verify',manifests[0].parent]
        else:
            argv = [OUT/'visualeval','-verify',OUT/('visual-'+route)]
        code = execute('verify-'+kind+'-'+route, argv)
        independent.append(dict(kind=kind,route=route,exitCode=code,seconds=commands[-1]['seconds']))
(OUT/'independent-notrun-verification.json').write_text(json.dumps(independent,indent=2)+'\n')
execute('live-harness', ['go','test','-tags','live','./ai/live','-count=1','-timeout=10m'],
        extra={'BARNESS_AI_EVIDENCE_DIR':str(OUT/'live-harness')})
if a.gate:
    def text_bundle(route):
        manifests = list((OUT/('text-'+route)).glob('*/manifest.json'))
        if len(manifests) != 1: raise RuntimeError('unique text-effect package missing')
        return str(manifests[0].parent.relative_to(ROOT))
    args = ['go','run','./ai/release/cmd/releasegate','-out',str((OUT/'gate').relative_to(ROOT)),
            '-evaluation-jev',text_bundle('jev'),
            '-evaluation-luna',text_bundle('luna'),
            '-visual-evaluation-jev',str((OUT/'visual-jev').relative_to(ROOT)),
            '-visual-evaluation-decisions',str((OUT/'visual-decisions').relative_to(ROOT))]
    for bundle in live: args += ['-live',bundle]
    execute('releasegate', args, timeout=3600)
    capacity = []
    for phase in ('offline','race'):
        manifests = list((OUT/'gate'/phase).glob('*/manifest.json'))
        if len(manifests)!=1: raise RuntimeError('unique capacity package missing')
        for entry in ('plain','hooked'):
            path = manifests[0].parent/('UC09-design-load-'+entry)/'load.json'
            load = json.loads(path.read_text())
            resources = json.loads((path.parent/'resources.json').read_text())
            capacity.append(dict(phase=phase,entry=entry,calls=load['calls'],tenants=load['tenants'],
                protocols=load['protocols'],memoryCeiling=load['memoryCeiling'],totalAlloc=load['totalAlloc'],
                heapGrowth=max(0,load['heap']['peakLive']-load['heap']['baselineLive']),
                maxLimitUse=max(v['ratio'] for v in load['headroom'].values()),
                resourcesReleased=all(v==0 for v in resources.values()),artifact=str(path.relative_to(ROOT))))
    (OUT/'multimodal-capacity-summary.json').write_text(json.dumps(capacity,indent=2)+'\n')
print('No credential was inherited; all real-service preflights must remain NOT_RUN.', flush=True)
