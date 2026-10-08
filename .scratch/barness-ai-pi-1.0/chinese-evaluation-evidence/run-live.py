#!/usr/bin/env python3
"""Run the independent Chinese evaluation with one isolated TypeSafe key."""
import argparse
import os
from pathlib import Path
import shlex
import signal
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument('--live', action='store_true', help='enable paid real evaluation')
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
key = ''
if args.live:
    for line in (root / '.env').read_text().splitlines():
        name, sep, value = line.strip().removeprefix('export ').partition('=')
        if sep and name.strip() == 'TYPESAFE_KEY':
            parts = shlex.split(value, comments=True)
            if len(parts) == 1:
                key = parts[0]
            break
# Only runtime configuration crosses the process boundary. Never source .env
# or inherit another provider's keys, proxies, tokens or smoke switches.
env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG', 'SSL_CERT_FILE') if k in os.environ}
command = []
if args.live:
    env.update(BARNESS_AI_CHINESE_EVAL='1',
               BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY=key,
               BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS='local-typesafe@region-unreported')
    command.append('-live')
command += ['-out', str(root / '.evidence/barness-ai/issue16-evaluation')]
# Run a binary so SIGINT/SIGTERM reaches the evaluated host rather than a go
# wrapper. The host cancels its active call and writes remaining sample states.
with tempfile.TemporaryDirectory(prefix='barness-zh-eval-') as build_dir:
    binary = str(Path(build_dir) / 'chineseeval')
    build = subprocess.run(['go', 'build', '-o', binary, './ai/examples/chineseeval/cmd/chineseeval'], cwd=root, env=env, capture_output=True)
    if build.returncode:
        diagnostics = (build.stdout + build.stderr).replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
        (root / '.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/build-failure.log').write_bytes(diagnostics)
        raise SystemExit('FAIL: evaluation stage=host_build; sanitized diagnostics in build-failure.log')
    child = subprocess.Popen([binary] + command, cwd=root, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    def cancel(signum, _frame):
        child.send_signal(signum)
    signal.signal(signal.SIGINT, cancel)
    signal.signal(signal.SIGTERM, cancel)
    output, _ = child.communicate()
    if key:
        output = output.replace(key.encode(), b'[EVALUATION-KEY]')
    # Logs name evidence relative to the repo; they contain no home-directory
    # paths and can themselves be audited with the delivery bundle.
    output = output.replace(str(root).encode(), b'[REPO]')
    log = root / '.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence' / ('live.log' if args.live else 'not-run.log')
    log.write_bytes(output)
    print(output.decode(), end='')
    raise SystemExit(child.returncode)
