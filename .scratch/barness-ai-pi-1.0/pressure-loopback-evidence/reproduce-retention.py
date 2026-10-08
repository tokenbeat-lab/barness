#!/usr/bin/env python3
"""Replay the confirmed retention defect in a disposable source copy."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[3]
    out = args.out.resolve()
    try:
        out.relative_to(root / '.evidence')
    except ValueError:
        parser.error('choose a fresh output directory under .evidence/')
    out.mkdir(parents=True, exist_ok=False)
    baseline = 'a6c163af07663fed878de3ca4f18e75eeac2bc2a'
    legacy = subprocess.check_output(['git', 'show', baseline + ':ai/internal/testkit/evidence/evidence.go'], cwd=root)
    files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=root, text=True).splitlines()
    env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'LANG') if k in os.environ}
    env.update(BARNESS_AI_PRESSURE='1', BARNESS_AI_PRESSURE_LOOPS='20', BARNESS_AI_EVIDENCE_DIR=str(out / 'bundles'), GOTOOLCHAIN='go1.26.2')
    command = ['go', 'test', './ai/e2e', '-run', '^TestPressureLoopback$', '-count=1', '-timeout=30m']
    with tempfile.TemporaryDirectory() as tmp:
        source = Path(tmp)
        for name in files:
            path = root / name
            if path.is_file() and (name in ('go.mod', 'go.sum') or name.startswith('ai/')):
                target = source / name
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(path.read_bytes())
        (source / 'ai/internal/testkit/evidence/evidence.go').write_bytes(legacy)
        hashes = {str(p.relative_to(source)): hashlib.sha256(p.read_bytes()).hexdigest() for p in source.rglob('*') if p.is_file()}
        (out / 'source-hashes.json').write_text(json.dumps(hashes, indent=2, sort_keys=True) + '\n')
        result = subprocess.run(command, cwd=source, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        log = result.stdout
        for value, label in ((str(source), '[SOURCE]'), (str(root), '[REPO]'), (str(Path.home()), '[HOME]'), (socket.gethostname(), '[HOST]')):
            log = log.replace(value.encode(), label.encode())
        (out / 'go-test.log').write_bytes(log)
    confirmed = False
    for path in (out / 'bundles').rglob('assertions.json'):
        data = json.loads(path.read_text())
        confirmed |= any(a['name'] == 'completed fixtures do not accumulate beyond the unchanged heap budget' and not a['pass'] for a in data['assertions'])
    verdict = {'command': command, 'legacyEvidenceCommit': baseline, 'sourceParentCommit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(), 'toolchain': 'go1.26.2', 'goTestExitCode': result.returncode, 'retentionRegressionFailed': confirmed}
    (out / 'result.json').write_text(json.dumps(verdict, indent=2) + '\n')
    # The audit accepts a bundle root; this manifest covers receipts/logs as
    # well as every nested E2E bundle, without changing their manifests.
    (out / 'manifest.json').write_text(json.dumps({'module': 'barness-pressure-reproduction', **verdict}, indent=2) + '\n')
    if result.returncode == 0 or not confirmed:
        raise SystemExit('FAIL: retention regression was not reproduced; inspect preserved output')
    audit = subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(out)], cwd=root, env=env, capture_output=True)
    if audit.returncode:
        raise SystemExit('FAIL: redaction audit')
    print('PASS: unchanged design load reproduces retained-fixture budget failure; evidence preserved')


if __name__ == '__main__':
    main()
