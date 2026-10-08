#!/usr/bin/env python3
"""Inject only TYPESAFE_KEY into the isolated official classifier smoke."""
import os
from pathlib import Path
import shlex
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
key = ''
for line in (root / '.env').read_text().splitlines():
    name, sep, value = line.strip().removeprefix('export ').partition('=')
    if sep and name.strip() == 'TYPESAFE_KEY':
        parts = shlex.split(value, comments=True)
        if len(parts) == 1:
            key = parts[0]
        break
# Do not inherit ambient secrets or the other vendors' variables. HOME and
# PATH are runtime configuration, not reassigned to task-specific values.
env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG', 'SSL_CERT_FILE') if k in os.environ}
env.update(BARNESS_AI_LIVE='1', BARNESS_AI_LIVE_COMBO='typesafe-classifier',
           BARNESS_AI_LIVE_ACCOUNT_ALIAS=os.environ.get('BARNESS_AI_LIVE_ACCOUNT_ALIAS', 'local-typesafe@region-unreported'),
           BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER=key,
           BARNESS_AI_EVIDENCE_DIR=str(root / '.evidence/barness-ai/issue08-live'))
result = subprocess.run(['go', 'test', '-tags', 'live', './ai/live', '-run', '^TestLive/typesafe-classifier$', '-count=1', '-v'], cwd=root, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
# A final guard prevents a malformed error outside evidence from printing the
# credential into a terminal or log. Actual evidence has its own strict audit.
output = result.stdout.replace(key.encode(), b'[REDACTED]') if key else result.stdout
(root / '.scratch/barness-ai-pi-1.0/typesafe-live-evidence/live.log').write_bytes(output)
sys.stdout.buffer.write(output)
sys.exit(result.returncode)
