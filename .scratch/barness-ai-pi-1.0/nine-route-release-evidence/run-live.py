#!/usr/bin/env python3
"""One explicit combination per process; the release gate never handles keys."""
import argparse
import json
import os
from pathlib import Path
import shlex
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--combo', required=True, choices=['openai-responses', 'openai-chat', 'anthropic-messages', 'google-gemini', 'deepseek-responses', 'deepseek-chat', 'typesafe-classifier', 'openai-images', 'google-interactions-image'])
parser.add_argument('--live', action='store_true')
parser.add_argument('--binary', type=Path, required=True, help='go test -c -tags live -o <binary> ./ai/live')
parser.add_argument('--out', type=Path, required=True)
parser.add_argument('--alias', required=True, help='host-confirmed account/region alias, or explicit unconfirmed alias')
args = parser.parse_args()
args.out = args.out.resolve()
root = Path(__file__).resolve().parents[3]
key_name = {'openai': 'OPENAI_KEY', 'anthropic': 'ANTHROPIC_KEY', 'google': 'GEMINI_KEY', 'deepseek': 'DEEPSEEK_KEY', 'typesafe': 'TYPESAFE_KEY'}[args.combo.split('-')[0]]
key = ''
if args.live and (root / '.env').is_file():
    # Stream only to the selected name; do not load a dictionary of vendor keys.
    with (root / '.env').open() as source:
        for line in source:
            name, sep, value = line.strip().removeprefix('export ').partition('=')
            if sep and name.strip() == key_name:
                parts = shlex.split(value, comments=True)
                if len(parts) == 1:
                    key = parts[0]
                break
env = {name: os.environ[name] for name in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG', 'SSL_CERT_FILE') if name in os.environ}
env.update(BARNESS_AI_LIVE='1', BARNESS_AI_LIVE_COMBO=args.combo,
           BARNESS_AI_LIVE_ACCOUNT_ALIAS=args.alias,
           BARNESS_AI_EVIDENCE_DIR=str(args.out.resolve()))
if key:
    env['BARNESS_AI_LIVE_KEY_' + args.combo.upper().replace('-', '_')] = key
args.out.mkdir(parents=True, exist_ok=True)
started = time.time()
child = subprocess.run([str(args.binary.resolve()), '-test.run=^TestLive$/^' + args.combo + '$', '-test.v'], cwd=root / 'ai/live', env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
output = child.stdout
if key:
    output = output.replace(key.encode(), b'[LIVE-KEY]')
output = output.replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
(args.out / 'run.log').write_bytes(output)
reports = list(args.out.glob('*/live-report.json'))
result = {'combo': args.combo, 'liveEnabled': args.live, 'keyPresent': bool(key), 'exitCode': child.returncode, 'durationSeconds': time.time()-started, 'reports': [str(path.relative_to(root)) for path in reports]}
(args.out / 'command.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result))
raise SystemExit(child.returncode)
