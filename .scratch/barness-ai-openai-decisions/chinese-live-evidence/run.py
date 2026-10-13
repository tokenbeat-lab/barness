#!/usr/bin/env python3
"""Isolate Luna credentials and preserve a fresh receipt for every execution."""
import argparse
import datetime
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import shlex
import signal
import subprocess
import sys
from urllib.parse import urlsplit


def stamp():
    return datetime.datetime.now(datetime.timezone.utc).isoformat()


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def safe_proxy(raw):
    # Validate at this boundary too: receipts are written by Python even when
    # the Go CLI rejects the command. Never persist untrusted URL credentials.
    try:
        parsed = urlsplit(raw)
        valid = (raw.startswith('http://') and not any(c.isspace() for c in raw)
                 and '?' not in raw and '#' not in raw and parsed.username is None
                 and parsed.password is None and parsed.path in ('', '/')
                 and parsed.port is not None and 1 <= parsed.port <= 65535
                 and ipaddress.ip_address(parsed.hostname).is_loopback)
    except (ValueError, TypeError):
        valid = False
    if not valid:
        raise argparse.ArgumentTypeError('proxy requires a credential-free loopback HTTP URL and valid port')
    return raw


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live', action='store_true', help='enable up to 24 paid Luna calls')
    parser.add_argument('--proxy', type=safe_proxy, default=None, help='explicit loopback HTTP proxy')
    parser.add_argument('--account-alias', default='local-openai@region-unreported')
    args = parser.parse_args()
    if args.proxy and not args.live:
        parser.error('--proxy requires --live')
    root = Path(__file__).resolve().parents[3]
    os.chdir(root)
    execution = Path('.evidence/issue12-executions') / datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ')
    execution.mkdir(parents=True, exist_ok=False, mode=0o700)
    env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH') if k in os.environ}
    binary = execution / 'chineseeval'
    build_command = ['go', 'build', '-o', str(binary), './ai/examples/chineseeval/cmd/chineseeval']
    build = subprocess.run(build_command, env=env, capture_output=True, timeout=120)
    build_output = (build.stdout + build.stderr).replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
    (execution / 'build.log').write_bytes(build_output)
    if build.returncode:
        raise SystemExit('FAIL: build; diagnostics in ' + str(execution))
    key = ''
    if args.live:
        # Parse only the named credential; never execute/source dotenv or pass
        # any other provider key to the evaluation process. Missing is NOT_RUN.
        values = []
        dotenv = root / '.env'
        for line in dotenv.read_text().splitlines() if dotenv.exists() else []:
            name, sep, value = line.strip().removeprefix('export ').partition('=')
            if sep and name.strip() == 'OPENAI_KEY':
                parts = shlex.split(value, comments=True)
                if len(parts) > 1:
                    raise SystemExit('FAIL: ambiguous OPENAI_KEY')
                values.append(parts[0] if parts else '')
        if len(values) > 1:
            raise SystemExit('FAIL: duplicate OPENAI_KEY')
        key = values[0] if values else ''
        env.update(BARNESS_AI_CHINESE_EVAL='1', BARNESS_AI_CHINESE_EVAL_OPENAI_KEY=key,
                   BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=args.account_alias)
    command = [str(binary), '-route', 'luna', '-out', str(execution / 'bundle')]
    if args.live:
        command.append('-live')
    if args.proxy:
        command += ['-proxy', args.proxy]
    matrix = Path('ai/live/support-matrix.json')
    receipt = {'startedAt': stamp(), 'buildCommand': build_command, 'command': command,
               'binarySHA256': sha(binary), 'gitCommit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
               'credentialSource': '.env:OPENAI_KEY' if args.live else 'none',
               'credentialEnvironment': 'BARNESS_AI_CHINESE_EVAL_OPENAI_KEY' if args.live else None,
               'inheritedEnvironmentNames': sorted(env), 'accountAlias': args.account_alias if args.live else None,
               'accountVerification': 'unknown', 'regionVerification': 'unknown',
               'dedicatedKeyVerification': 'unknown', 'permissionVerification': 'unknown',
               'vendorSpendLimitVerification': 'unknown', 'proxy': args.proxy or None,
               'budget': {'calls': 24, 'questions': 24, 'attempts': 24, 'retries': 0,
                          'concurrency': 1, 'totalSeconds': 600, 'callSeconds': 15},
               'matrixBeforeSHA256': sha(matrix)}
    child = subprocess.Popen(command, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    def cancel(signum, _frame):
        if child.poll() is None:
            child.send_signal(signum)
    signal.signal(signal.SIGINT, cancel)
    signal.signal(signal.SIGTERM, cancel)
    receipt['watchdogTriggered'] = False
    try:
        output, _ = child.communicate(timeout=620)
    except subprocess.TimeoutExpired:
        receipt['watchdogTriggered'] = True
        child.send_signal(signal.SIGTERM)
        try:
            output, _ = child.communicate(timeout=15)
        except subprocess.TimeoutExpired:
            child.kill()
            output, _ = child.communicate()
    if key:
        output = output.replace(key.encode(), b'[REDACTED]')
    output = output.replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
    (execution / 'execution.log').write_bytes(output)
    receipt.update(finishedAt=stamp(), exitCode=child.returncode, matrixAfterSHA256=sha(matrix),
                   logSHA256=sha(execution / 'execution.log'))
    reports = list((execution / 'bundle').glob('*/evaluation-report.json'))
    if len(reports) == 1:
        report = json.loads(reports[0].read_text())
        receipt.update(report=str(reports[0]), reportSHA256=sha(reports[0]),
                       status=report['status'], kind=report['kind'], budgetConsumed=report['budget'],
                       codeHash=report['codeHash'])
    (execution / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')
    print(json.dumps({'execution': str(execution), 'status': receipt.get('status', 'FAIL'), 'exitCode': child.returncode}))
    return child.returncode or int(receipt['watchdogTriggered'])


if __name__ == '__main__':
    sys.exit(main())
