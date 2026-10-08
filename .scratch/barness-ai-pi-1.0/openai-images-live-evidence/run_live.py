#!/usr/bin/env python3
"""Explicit issue-11 live runner: inject only one dotenv credential into Go.

Run from the repository root. Costs at most four 1024-square output requests.
Account permissions/region are host configuration, not inferred from a key.
"""
from pathlib import Path
import os
import re
import shlex
import subprocess

root = Path(__file__).resolve().parents[3]
key = None
for line in (root / '.env').read_text().splitlines():
    match = re.match(r'^\s*(?:export\s+)?OPENAI_KEY\s*=\s*(.*)$', line)
    if match:
        values = shlex.split(match[1], comments=True)
        if len(values) == 1:
            key = values[0]
if not key:
    raise SystemExit('OPENAI_KEY is missing; no live request sent')
base = Path(__file__).resolve().parent
# A whitelist prevents ambient vendor/combination credentials entering the process.
env = {name: os.environ[name] for name in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'GOTOOLCHAIN', 'SSL_CERT_FILE', 'SSL_CERT_DIR') if name in os.environ}
env.update(BARNESS_AI_LIVE='1', BARNESS_AI_LIVE_COMBO='openai-images',
           BARNESS_AI_LIVE_ACCOUNT_ALIAS='dotenv-openai@region-unconfirmed',
           BARNESS_AI_LIVE_KEY_OPENAI_IMAGES=key,
           BARNESS_AI_EVIDENCE_DIR=str(base / 'live'))
with (base / 'live.log').open('w') as log:
    result = subprocess.run(['go', 'test', '-tags', 'live', '-count=1', '-v', './ai/live', '-run', '^TestLive/openai-images$'], cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
print('live exit:', result.returncode)
raise SystemExit(result.returncode)
