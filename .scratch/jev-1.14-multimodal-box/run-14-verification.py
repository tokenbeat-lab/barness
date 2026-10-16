#!/usr/bin/env python3
"""Reproduce ticket 14 without inheriting route credentials or making paid calls."""
import argparse
import datetime as dt
import json
import os
from pathlib import Path
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('out', help='new repository-relative output directory')
parser.add_argument('--gate', action='store_true', help='run full release command after preflights')
a = parser.parse_args()
OUT = ROOT / a.out
OUT.mkdir(parents=True, exist_ok=False)
ENV = {k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}
commands = []


def execute(name, argv, cwd=ROOT, extra=None, timeout=1800):
    env = ENV | (extra or {})
    started = dt.datetime.now(dt.timezone.utc).isoformat()
    clock = time.monotonic()
    with (OUT / (name + '.log')).open('w') as log:
        proc = subprocess.Popen([str(v) for v in argv], cwd=cwd, env=env,
                                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        for line in proc.stdout:
            log.write(line.replace(str(ROOT), '$REPO').replace(str(Path.home()), '$HOME'))
        code = proc.wait(timeout=timeout)
    record = dict(name=name, argv=[str(v).replace(str(ROOT), '$REPO') for v in argv],
                  cwd=str(cwd.relative_to(ROOT)) or '.', startedAt=started,
                  seconds=time.monotonic()-clock, exitCode=code,
                  environmentKeys=sorted(env), credentialsInherited=False)
    commands.append(record)
    (OUT / 'commands.json').write_text(json.dumps(commands, indent=2)+'\n')
    print(f'{name}: exit {code}', flush=True)
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
    formal = ['-formal'] if route == 'jev' else []
    execute('visual-effects-'+route, [OUT/'visualeval',*formal,'-mode','live','-route',route,
                                     '-out',OUT/('visual-'+route)])
execute('live-harness', ['go','test','-tags','live','./ai/live','-count=1','-timeout=10m'],
        extra={'BARNESS_AI_EVIDENCE_DIR':str(OUT/'live-harness')})
if a.gate:
    def text_bundle(route):
        manifests = list((OUT/('text-'+route)).glob('*/manifest.json'))
        if len(manifests) != 1: raise RuntimeError('unique text-effect package missing')
        return str(manifests[0].parent.relative_to(ROOT))
    args = ['go','run','./ai/release/cmd/releasegate','-out',str((OUT/'gate').relative_to(ROOT)),
            '-evaluation-jev',text_bundle('jev'),
            '-evaluation-luna',text_bundle('luna')]
    for bundle in live: args += ['-live',bundle]
    execute('releasegate', args, timeout=2400)
print('No credential was inherited; all real-service preflights must remain NOT_RUN.', flush=True)
