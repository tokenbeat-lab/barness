#!/usr/bin/env python3
"""Explicit host dotenv injection; never reads or prints another provider key."""
import argparse
import os
from pathlib import Path
import re
import shlex
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--dotenv', type=Path, required=True)
parser.add_argument('--evidence-dir', type=Path, required=True)
parser.add_argument('--alias', required=True)
args = parser.parse_args()
key = ''
for line in args.dotenv.read_text().splitlines():
    name, sep, value = line.partition('=')
    if sep and name.strip().removeprefix('export ') == 'GEMINI_KEY':
        parts = shlex.split(value, comments=True)
        if len(parts) == 1:
            key = parts[0]
if not key:
    raise SystemExit('GEMINI_KEY missing; use the documented no-key command for NOT_RUN evidence')
env = {k: v for k, v in os.environ.items()
       if not re.search(r'KEY|TOKEN|SECRET|PASSWORD|BARNESS_AI_LIVE|BARNESS_AI_EVIDENCE', k, re.I)}
env.update(BARNESS_AI_LIVE='1', BARNESS_AI_LIVE_COMBO='google-interactions-image',
           BARNESS_AI_LIVE_ACCOUNT_ALIAS=args.alias,
           BARNESS_AI_LIVE_KEY_GOOGLE_INTERACTIONS_IMAGE=key,
           BARNESS_AI_EVIDENCE_DIR=str(args.evidence_dir.resolve()))
raise SystemExit(subprocess.call(['go', 'test', '-tags', 'live', '-count=1', './ai/live',
    '-run', '^TestLive$/^google-interactions-image$', '-v'], env=env))
