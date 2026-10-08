#!/usr/bin/env python3
"""Replay nine isolated smokes, atomic matrix merges and the evidence-only gate."""
import argparse
import json
import os
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--live', action='store_true', help='explicitly authorize the bounded real smokes; absent keys produce NOT_RUN')
parser.add_argument('--out', type=Path, required=True, help='fresh ignored output directory')
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
args.out = args.out.resolve()
try:
    portable_out = args.out.relative_to(root)
except ValueError:
    raise SystemExit('output must be inside the repository for portable replay paths')
if args.out.exists():
    raise SystemExit('output already exists; choose a fresh directory to preserve historical evidence')
args.out.mkdir(parents=True)
env = {name: os.environ[name] for name in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG', 'SSL_CERT_FILE') if name in os.environ}
binary = args.out / '.tools/live.test'
binary.parent.mkdir(parents=True, exist_ok=True)
build = subprocess.run(['go', 'test', '-c', '-tags', 'live', '-o', str(binary), './ai/live'], cwd=root, env=env, capture_output=True)
if build.returncode:
    raise SystemExit('live harness build failed; no Provider calls were made')
combos = ['openai-responses', 'anthropic-messages', 'google-gemini', 'openai-chat', 'deepseek-responses', 'deepseek-chat', 'typesafe-classifier', 'openai-images', 'google-interactions-image']
results, bundles = [], []
for combo in combos:
    out = args.out / 'live' / combo
    command = ['python3', str(Path(__file__).with_name('run-live.py')), '--combo', combo, '--binary', str(binary), '--out', str(out), '--alias', 'local-'+combo.split('-')[0]+'@region-permissions-unconfirmed']
    if args.live:
        command.append('--live')
    run = subprocess.run(command, cwd=root, env=env)
    reports = list(out.glob('*/live-report.json'))
    merged = None
    if len(reports) == 1:
        bundles.append(str(reports[0].parent.relative_to(root)))
        merge = subprocess.run(['go', 'run', './ai/live/cmd/supportmatrix', str(reports[0])], cwd=root, env=env, capture_output=True)
        merged = merge.returncode
        (out / 'merge.log').write_bytes((merge.stdout+merge.stderr).replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))
    results.append({'combo':combo, 'exitCode':run.returncode, 'mergeExitCode':merged, 'bundle':bundles[-1] if len(reports)==1 else None})
    (args.out / 'smoke-commands.json').write_text(json.dumps(results, indent=2)+'\n')
    print(combo, 'exit', run.returncode, 'merge', merged, flush=True)
gate = ['go','run','./ai/release/cmd/releasegate','-out',str(portable_out/'gate'), '-evaluation','.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live']
for bundle in bundles:
    gate += ['-live',bundle]
(args.out / 'gate-command.json').write_text(json.dumps({'argv':gate},indent=2)+'\n')
run = subprocess.run(gate,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
(args.out / 'gate.log').write_bytes(run.stdout.replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))
(args.out / 'gate-exit.json').write_text(json.dumps({'exitCode':run.returncode},indent=2)+'\n')
print('release gate exit',run.returncode,'report',str(portable_out/'gate/release-report.md'),flush=True)
raise SystemExit(run.returncode)
