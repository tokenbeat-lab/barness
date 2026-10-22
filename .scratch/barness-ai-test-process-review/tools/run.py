#!/usr/bin/env python3
"""Capture one explicitly selected verification process; never schedule another."""
import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import re
import shutil
import platform
import subprocess
import sys
import tempfile
import time

from receipt_files import json_file, save, sha, source_snapshot, verify_bundle, within
from run_process import capture, environment


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--feature', required=True)
    parser.add_argument('--stage', choices=('e2e', 'race', 'pressure', 'live'), required=True)
    parser.add_argument('--route')
    parser.add_argument('--timeout', type=float)
    parser.add_argument('--input', action='append', default=[], help='selected repository-relative input path')
    parser.add_argument('command', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if not re.fullmatch(r'[a-z0-9][a-z0-9-]{0,100}', args.feature):
        parser.error('invalid feature slug')
    if (args.stage == 'live') != bool(args.route) or (args.route and not re.fullmatch(r'[a-z0-9][a-z0-9.-]{0,100}', args.route)):
        parser.error('route is required only for live')
    argv = args.command[1:] if args.command[:1] == ['--'] else args.command
    if not argv:
        parser.error('command required')
    if any(x.startswith('-race') for x in argv) and args.stage != 'race':
        parser.error('race is permitted only in the explicitly selected race stage')
    if any(x.startswith('-tags=live') or x == 'live' for x in argv) and args.stage != 'live':
        parser.error('live build selection is not an ordinary or pressure command')
    timeout = args.timeout or {'e2e': 1500, 'race': 4800, 'pressure': 1500, 'live': 1900}[args.stage]
    if not 0 < timeout <= 5400:
        parser.error('timeout outside bounds')
    root = Path.cwd().resolve()
    try:
        source = source_snapshot(root)
        if len(args.input) != len(set(args.input)):
            raise ValueError('duplicate selected input')
        selected_inputs = {name: sha(within(root, name)) for name in args.input}
        env, secrets = environment(args.stage, args.route)
        if any(secret and secret in part for secret in secrets for part in argv):
            raise ValueError('credential in command arguments')
        base = root / '.evidence' / args.feature / args.stage
        if args.route:
            base /= args.route
        # All ancestors are locally owned; reject symlinks before allocating.
        for parent in (base, *base.parents):
            if parent == root:
                break
            if parent.is_symlink():
                raise ValueError('symbolic evidence root rejected')
        base.mkdir(parents=True, exist_ok=True)
        prefix = datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ-')
        out = Path(tempfile.mkdtemp(prefix=prefix, dir=base))
        bundles = out / 'bundles'
        bundles.mkdir()
        env['BARNESS_AI_EVIDENCE_DIR'] = str(bundles)
        env['BARNESS_RUN_COMMIT'] = source['commit']
        executable = shutil.which(argv[0], path=env.get('PATH'))
        toolchain = {'collectorPython': platform.python_version(), 'platform': platform.system(),
                     'architecture': platform.machine(),
                     'executableSHA256': sha(Path(executable).resolve()) if executable else ''}
        started = datetime.now(timezone.utc).isoformat()
        monotonic = time.monotonic()
        command = {'schema': 1, 'argv': argv, 'stage': args.stage, 'route': args.route,
                   'timeoutSeconds': timeout, 'environmentNames': sorted(env), 'toolchain': toolchain}
        save(out / 'command.json', command)
        receipt = {'schema': 1, 'feature': args.feature, 'stage': args.stage, 'route': args.route,
                   'status': 'RUNNING', 'exitCode': None, 'source': source, 'selectedInputs': selected_inputs, 'toolchain': toolchain,
                   'startedAt': started, 'argv': argv, 'bundles': [], 'commandSHA256': sha(out / 'command.json')}
        save(out / 'receipt.json', receipt)
        try:
            receipt.update(capture(argv, root, env, out, timeout, secrets))
        except (OSError, ValueError, subprocess.TimeoutExpired):
            receipt.update(status='CAPTURE_FAILED', error='process capture could not be finalized')
        if receipt['exitCode'] is not None:
            (out / 'exit-code').write_text(str(receipt['exitCode']) + '\n')
        try:
            if source_snapshot(root) != source:
                raise ValueError('execution source changed during stage')
            if any(sha(within(root, name)) != value for name, value in selected_inputs.items()):
                raise ValueError('selected input changed during stage')
            expected = './ai/e2e' if args.stage != 'live' else './ai/live'
            found = []
            for path in sorted(bundles.iterdir()):
                if not path.is_dir() or path.is_symlink() or not (path / 'manifest.json').is_file():
                    continue
                manifest = json_file(path / 'manifest.json')
                package = manifest.get('package', '')
                verify_bundle(path, package)
                if manifest.get('git_commit') != source['commit']:
                    raise ValueError('bundle source commit mismatch')
                found.append({'path': str(path.relative_to(out)), 'package': package,
                    'manifestSHA256': sha(path / 'manifest.json'), 'auditSHA256': sha(path / 'audit.json')})
            primary = [item for item in found if item['package'] == expected]
            if len(primary) != 1:
                raise ValueError('expected exactly one primary native bundle')
            receipt['bundles'] = found
            if args.stage != 'live':
                tests = set()
                for line in (out / 'stdout.jsonl').open():
                    try:
                        event = json.loads(line)
                    except ValueError:
                        continue
                    if event.get('Action') == 'pass' and event.get('Test'):
                        tests.add(event['Test'])
                if not tests:
                    raise ValueError('no completed test events')
                receipt['tests'] = sorted(tests)
        except (OSError, ValueError, TypeError) as error:
            if receipt['status'] == 'PASS':
                receipt['status'] = 'INVALID_EVIDENCE'
            receipt['error'] = str(error)
        receipt['files'] = {}
        try:
            for name in ('command.json', 'stdout.jsonl', 'stderr.log', 'exit-code'):
                path = out / name
                if path.is_file():
                    receipt['files'][path.name] = sha(path)
        except (OSError, ValueError):
            receipt.update(status='CAPTURE_FAILED', error='capture inventory could not be finalized')
        receipt.update(finishedAt=datetime.now(timezone.utc).isoformat(), elapsedSeconds=time.monotonic() - monotonic)
        save(out / 'receipt.json', receipt)
        print(json.dumps({'attempt': str(out.relative_to(root)), 'status': receipt['status'], 'exitCode': receipt['exitCode']}))
        return 0 if receipt['status'] == 'PASS' else 1
    except (OSError, ValueError, KeyError) as error:
        print('run: ' + str(error), file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
