#!/usr/bin/env python3
"""Collect current live evidence in isolated processes; never overwrite a run."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import shlex
import subprocess
import zipfile

ROOT = Path(__file__).resolve().parents[3]
COMBOS = ['openai-responses', 'anthropic-messages', 'google-gemini', 'openai-chat',
          'deepseek-responses', 'deepseek-chat', 'typesafe-classifier', 'openai-images',
          'google-interactions-image', 'openai-decisions']
KEYS = {'openai': 'OPENAI_KEY', 'anthropic': 'ANTHROPIC_KEY', 'google': 'GEMINI_KEY',
        'deepseek': 'DEEPSEEK_KEY', 'typesafe': 'TYPESAFE_KEY'}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--live', action='store_true')
    args = p.parse_args()
    os.chdir(ROOT)
    out = args.out.resolve()
    if not out.is_relative_to(ROOT) or out.exists():
        raise SystemExit('output must be a fresh directory inside the repository')
    out.mkdir(parents=True, mode=0o700)
    env = {k: os.environ[k] for k in ['PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH'] if k in os.environ}
    binaries = {'smoke': out / 'live.test', 'evaluation': out / 'chineseeval'}
    for command in [['go', 'test', '-c', '-tags', 'live', '-o', str(binaries['smoke']), './ai/live'],
                    ['go', 'build', '-o', str(binaries['evaluation']), './ai/examples/chineseeval/cmd/chineseeval']]:
        subprocess.run(command, env=env, check=True, timeout=180)
    # Preserve the exact source bytes corresponding to this execution; a commit
    # alone does not identify a working tree. No credentials enter this archive.
    with zipfile.ZipFile(out / 'execution-source.zip', 'x', zipfile.ZIP_DEFLATED) as archive:
        for file in sorted((ROOT / 'ai').rglob('*')):
            if file.is_file() and (file.suffix in ['.go', '.json'] or file.name in ['go.mod', 'go.sum']):
                archive.write(file, str(file.relative_to(ROOT)))
        for name in ['go.mod', 'go.sum']:
            archive.write(ROOT / name, name)
    results = []
    jobs = [('smoke', combo) for combo in COMBOS] + [('evaluation', 'jev')]
    for kind, name in jobs:
        provider = name.split('-')[0] if kind == 'smoke' else 'typesafe'
        key_name = KEYS[provider]
        key = ''
        if args.live:
            values = []
            for line in (ROOT / '.env').read_text().splitlines() if (ROOT / '.env').exists() else []:
                k, sep, value = line.strip().removeprefix('export ').partition('=')
                if sep and k.strip() == key_name:
                    parts = shlex.split(value, comments=True)
                    if len(parts) > 1:
                        raise SystemExit('ambiguous selected credential')
                    values.append(parts[0] if parts else '')
            if len(values) > 1:
                raise SystemExit('duplicate selected credential')
            key = values[0] if values else ''
        folder = out / kind / name
        folder.mkdir(parents=True)
        child_env = dict(env)
        alias = 'local-' + provider + '@region-permissions-unconfirmed'
        if kind == 'smoke':
            child_env.update(BARNESS_AI_LIVE='1', BARNESS_AI_LIVE_COMBO=name,
                             BARNESS_AI_LIVE_ACCOUNT_ALIAS=alias,
                             BARNESS_AI_EVIDENCE_DIR=str(folder / 'bundles'),
                             BARNESS_AI_LIVE_PROXY='http://127.0.0.1:7890')
            if key:
                child_env['BARNESS_AI_LIVE_KEY_' + name.upper().replace('-', '_')] = key
            if name == 'openai-decisions':
                child_env['BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT'] = '1'
                test, timeout = '^TestFormalDecisions$', 210
            else:
                test, timeout = '^TestLive$/^' + name + '$', 620
            command = [str(binaries[kind]), '-test.run=' + test, '-test.timeout=' + ('180s' if name == 'openai-decisions' else '10m')]
            report_name = 'live-report.json'
        else:
            child_env.update(BARNESS_AI_CHINESE_EVAL='1', BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY=key,
                             BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS=alias)
            command = [str(binaries[kind]), '-route', name, '-out', str(folder / 'bundles')]
            if args.live:
                command += ['-live', '-proxy', 'http://127.0.0.1:7890']
            timeout, report_name = 620, 'evaluation-report.json'
        started = datetime.datetime.now(datetime.timezone.utc).isoformat()
        try:
            proc = subprocess.run(command, env=child_env, cwd=ROOT / 'ai/live' if kind == 'smoke' else ROOT,
                                  stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=timeout)
            output, code = proc.stdout, proc.returncode
        except subprocess.TimeoutExpired as error:
            output, code = error.stdout or b'', 124
        if key:
            output = output.replace(key.encode(), b'[REDACTED]')
        output = output.replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')
        (folder / 'execution.log').write_bytes(output)
        reports = list((folder / 'bundles').glob('*/' + report_name))
        receipt = {'kind': kind, 'route': name, 'startedAt': started,
                   'finishedAt': datetime.datetime.now(datetime.timezone.utc).isoformat(),
                   'command': [str(Path(arg).relative_to(ROOT)) if arg.startswith(str(ROOT)) else arg for arg in command],
                   'exitCode': code, 'credentialSource': '.env:' + key_name if args.live else 'none',
                   'keyPresent': bool(key), 'accountAlias': alias,
                   'accountRegionPermissionsDedicatedKeyVerification': 'unknown',
                   'environmentNames': sorted(child_env), 'binarySHA256': digest(binaries[kind]),
                   'bundle': str(reports[0].parent.relative_to(ROOT)) if len(reports) == 1 else None}
        if kind == 'smoke' and len(reports) == 1:
            merge = subprocess.run(['go', 'run', './ai/live/cmd/supportmatrix', str(reports[0])], cwd=ROOT, env=env,
                                   stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=120)
            receipt['mergeExitCode'] = merge.returncode
            (folder / 'merge.log').write_bytes(merge.stdout.replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))
        (folder / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')
        results.append(receipt)
        (out / 'collection.json').write_text(json.dumps(results, indent=2) + '\n')
        print(kind, name, 'exit', code, 'merged', receipt.get('mergeExitCode'), flush=True)


if __name__ == '__main__':
    main()
