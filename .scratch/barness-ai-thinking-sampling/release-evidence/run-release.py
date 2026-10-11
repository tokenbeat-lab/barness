#!/usr/bin/env python3
"""Re-run the complete gate, with live calls only when explicitly enabled."""
import argparse
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[3]
FOLDER = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--live', action='store_true', help='enable the nine bounded smokes and independent evaluation')
parser.add_argument('--out', type=Path, required=True)
args = parser.parse_args()
out = args.out.resolve()
if not out.is_relative_to(ROOT):
    raise SystemExit('output must be inside the repository for reproducible relative records')
if out.exists():
    raise SystemExit('output must be fresh; never overwrite a recorded run')
out.mkdir(parents=True)
env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG', 'SSL_CERT_FILE') if k in os.environ}
command = ['go', 'run', './ai/release/cmd/releasegate', '-out', str(out / 'gate')]
if args.live:
    binary = out / 'live.test'
    subprocess.run(['go', 'test', '-c', '-tags', 'live', '-o', str(binary), './ai/live'], cwd=ROOT, env=env, check=True)
    subprocess.run(['python3', str(FOLDER / 'run-smokes.py'), '--live', '--out', str(out / 'smokes'), '--binary', str(binary)], cwd=ROOT, env=env, check=True)
    subprocess.run(['python3', str(FOLDER / 'run-evaluation.py'), '--live', '--out', str(out / 'evaluation')], cwd=ROOT, env=env)
    evaluations = list((out / 'evaluation').glob('*/evaluation-report.json'))
    if len(evaluations) == 1:
        command += ['-evaluation', str(evaluations[0].parent)]
    for result in json.loads((out / 'smokes/smoke-commands.json').read_text()):
        if result['bundle']:
            command += ['-live', str(ROOT / result['bundle'])]
(out / 'gate-command.json').write_text(json.dumps({'argv': command, 'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()}, indent=2) + '\n')
with (out / 'gate.log').open('w') as log:
    result = subprocess.run(command, cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT)
(out / 'gate-exit.json').write_text(json.dumps({'exitCode': result.returncode}, indent=2) + '\n')
print('gate exit', result.returncode, '; full log and reports:', out)
raise SystemExit(result.returncode)
