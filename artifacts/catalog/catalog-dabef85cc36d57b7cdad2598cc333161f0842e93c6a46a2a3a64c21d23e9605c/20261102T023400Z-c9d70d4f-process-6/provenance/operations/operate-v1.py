#!/usr/bin/env python3
"""Immutable command capture for this maintenance operation; no shell expansion."""
import argparse
import datetime
import hashlib
import json
import pathlib
import subprocess
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
WORK = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT / 'tools/verification'))
from run_process import capture, environment, isolate_go_home


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--label', required=True)
    parser.add_argument('--timeout', type=int, default=180)
    parser.add_argument('--expect', type=int, default=0)
    parser.add_argument('argv', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    argv = args.argv[1:] if args.argv[:1] == ['--'] else args.argv
    out = WORK / 'commands' / args.label
    out.mkdir(parents=True, exist_ok=False)
    env, secrets = environment('e2e')
    isolate_go_home(env, 'e2e', out)
    env.update(GOWORK='off', GOPROXY='off', GOSUMDB='off')
    started = datetime.datetime.now(datetime.timezone.utc).isoformat()
    tick = time.monotonic()
    state = capture(argv, str(ROOT), env, out, args.timeout, secrets)
    record = dict(schema=1, label=args.label, argv=argv, cwd=str(ROOT),
                  startedAt=started,
                  finishedAt=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                  wallSeconds=time.monotonic()-tick,
                  head=subprocess.check_output(['git','rev-parse','HEAD'], cwd=ROOT, text=True).strip(),
                  script={'path':str(pathlib.Path(__file__).relative_to(ROOT)), 'sha256':digest(pathlib.Path(__file__))},
                  sanitizedEnvironment={'GOENV':'off','GOFLAGS':'','GOWORK':'off','GOPROXY':'off','GOSUMDB':'off','HOME':'<isolated>'},
                  **state)
    if pathlib.Path(argv[0]).is_file():
        record['executableSha256'] = digest(pathlib.Path(argv[0]))
    for name in ['stdout.jsonl','stderr.log']:
        p=out/name
        if p.exists():
            record[name]={'path':str(p.relative_to(ROOT)), 'sha256':digest(p)}
    (out/'command.json').write_text(json.dumps(record,indent=2)+'\n')
    print(json.dumps({'label':args.label,'exitCode':state.get('exitCode'),'status':state.get('status'),'command':str((out/'command.json').relative_to(ROOT))}))
    if state.get('exitCode') != args.expect:
        sys.exit(1)


if __name__ == '__main__':
    main()
