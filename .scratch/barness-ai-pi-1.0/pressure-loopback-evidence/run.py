#!/usr/bin/env python3
"""Run unchanged public-Client pressure loads and preserve each process verdict."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import socket
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--mode', choices=['isolated', 'fresh', 'full', 'race'], required=True)
    parser.add_argument('--rounds', type=int, default=100)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    if not 1 <= args.rounds <= 1000:
        parser.error('rounds must be from 1 to 1000')
    root = Path(__file__).resolve().parents[3]
    out = args.out.resolve()
    try:
        relative = out.relative_to(root)
    except ValueError:
        parser.error('output must be inside the repository')
    out.mkdir(parents=True, exist_ok=False)
    # No vendor credentials or inherited test switches reach these processes.
    env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG') if k in os.environ}
    env['BARNESS_AI_PRESSURE'] = '1'
    env['BARNESS_AI_PRESSURE_LOOPS'] = str(args.rounds if args.mode != 'fresh' else 1)
    if args.mode in ('full', 'race'):
        env['BARNESS_AI_PIDIFF'] = '1'
    sources = {}
    files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=root, text=True).splitlines()
    for name in files:
        path = root / name
        if path.is_file() and (name in ('go.mod', 'go.sum', 'AGENTS.md') or name.startswith('ai/')):
            sources[name] = hashlib.sha256(path.read_bytes()).hexdigest()
    (out / 'source-hashes.json').write_text(json.dumps(sources, indent=2, sort_keys=True) + '\n')
    commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
    results = []
    processes = args.rounds if args.mode == 'fresh' else 1
    for i in range(processes):
        process_out = out / str(i)
        process_out.mkdir()
        env['BARNESS_AI_EVIDENCE_DIR'] = str(process_out / 'bundles')
        command = ['go', 'test']
        if args.mode == 'race':
            command += ['-race']
        command += ['./...' if args.mode in ('full', 'race') else './ai/e2e']
        if args.mode in ('isolated', 'fresh'):
            command += ['-run', '^TestPressureLoopback$']
        command += ['-count=1', '-timeout=30m']
        result = subprocess.run(command, cwd=root, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        log = result.stdout
        for value, label in ((str(root), '[REPO]'), (str(Path.home()), '[HOME]'), (socket.gethostname(), '[HOST]')):
            log = log.replace(value.encode(), label.encode())
        (process_out / 'go-test.log').write_bytes(log)
        results.append({'process': i, 'command': command, 'exitCode': result.returncode})
        receipt = {'commit': commit, 'mode': args.mode, 'rounds': args.rounds, 'pressure': '1', 'pidiff': env.get('BARNESS_AI_PIDIFF', '0'), 'processes': results}
        (out / 'results.json').write_text(json.dumps(receipt, indent=2) + '\n')
        print(f'{relative}/{i}: exit {result.returncode}', flush=True)
        if result.returncode:
            # A failed round/process is final, never retried into a success.
            return 1
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
