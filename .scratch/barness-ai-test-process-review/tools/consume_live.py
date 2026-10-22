#!/usr/bin/env python3
"""Capture one read-only native live consumer; never run a verification stage."""
import argparse
from datetime import datetime, timezone
from pathlib import Path
import sys
import time

from receipt_files import json_file, save, sha, source_snapshot, within
from run_process import capture, environment


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', required=True)
    parser.add_argument('--bundle', required=True)
    parser.add_argument('--out', required=True)
    parser.add_argument('--timeout', type=float, default=120)
    args = parser.parse_args()
    if not 0 < args.timeout <= 300:
        parser.error('timeout outside bounds')
    root = Path.cwd().resolve()
    try:
        bundle, out = within(root, args.bundle), within(root, args.out)
        binary = Path(args.binary).resolve()
        if binary.name != 'releasegate' or not bundle.is_dir() or out == bundle or bundle in out.parents:
            raise ValueError('native input and new outside output required')
        executable_hash = sha(binary)
        source = source_snapshot(root)
        env, _ = environment('e2e')
        out.mkdir(mode=0o700)
        argv = [str(binary), '-live-bundle', args.bundle, '-out', str(out / 'native')]
        started = datetime.now(timezone.utc).isoformat()
        tic = time.monotonic()
        receipt = {'schema': 2, 'consumer': 'releasegate/native-live-v1', 'pass': False,
                   'status': 'RUNNING', 'exitCode': None, 'source': source, 'bundle': args.bundle,
                   'inputs': {}, 'argv': argv, 'executableSHA256': executable_hash,
                   'startedAt': started, 'files': {}, 'goEnv': 'off', 'goFlags': ''}
        # If finalization fails this durable initial record stays non-passing.
        save(out / 'live-consumer.json', receipt)
        receipt.update(capture(argv, root, env, out, args.timeout, []))
        if receipt['error'] is None:
            del receipt['error']
        if receipt['exitCode'] is not None:
            (out / 'exit-code').write_text(str(receipt['exitCode']) + '\n')
        native = out / 'native/live-consumer.json'
        try:
            qualification = json_file(native)
            if source_snapshot(root) != source or sha(binary) != executable_hash or qualification['source'] != source or qualification['bundle'] != args.bundle:
                raise ValueError('native source or input changed')
            receipt.update(inputs=qualification['inputs'], reportSHA256=qualification.get('reportSHA256', ''))
            receipt['pass'] = receipt['status'] == 'PASS' and qualification['pass'] is True
        except (OSError, ValueError, KeyError, TypeError):
            receipt.update(status='CAPTURE_FAILED', error='native qualification unavailable or changed', **{'pass': False})
        receipt.update(finishedAt=datetime.now(timezone.utc).isoformat(), elapsedSeconds=time.monotonic() - tic)
        for name in ('stdout.jsonl', 'stderr.log', 'exit-code', 'native/live-consumer.json'):
            if (out / name).is_file():
                receipt['files'][name] = sha(out / name)
        save(out / 'live-consumer.json', receipt)
        sys.stdout.write((out / 'stdout.jsonl').read_text())
        sys.stderr.write((out / 'stderr.log').read_text())
        return 0 if receipt['pass'] else 1
    except (OSError, ValueError):
        print('consume-live: capture could not be finalized', file=sys.stderr)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
