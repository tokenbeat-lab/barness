#!/usr/bin/env python3
"""Freeze full ten-route evidence and failure history without relabeling runs."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

FOLDER = Path(__file__).resolve().parent
ROOT = FOLDER.parents[2]


def read(path):
    return json.loads(path.read_text())


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1 << 20), b''):
            h.update(chunk)
    return h.hexdigest()


def write(name, value):
    (FOLDER / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def redact(raw):
    return raw.replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]')


def bundle_under(base):
    manifests = list(base.glob('*/manifest.json'))
    if len(manifests) != 1:
        raise ValueError('expected one completed bundle: ' + str(base.relative_to(ROOT)))
    return manifests[0].parent


def archive_tree(z, prefix, source, diagnostics=False, excluded=()):
    for path in sorted(source.rglob('*')):
        if path.relative_to(source).parts[0] in excluded:
            continue
        if path.is_symlink():
            raise ValueError('symlink in evidence')
        if not path.is_file():
            continue
        name = prefix + '/' + str(path.relative_to(source))
        if diagnostics and path.parent == source and path.name in ['release-report.json', 'release-report.md']:
            z.writestr(name, redact(path.read_bytes()))
        else:
            z.write(path, name)


def bundle_record(role, archive, prefix, source):
    m = read(source / 'manifest.json')
    cases = m.get('cases', [])
    recorded_audit = read(source / 'audit.json')
    return {'role': role, 'archive': archive, 'archivePath': prefix,
            'originalBundle': str(source.relative_to(ROOT)),
            'manifestSHA256': sha(source / 'manifest.json'), 'auditSHA256': sha(source / 'audit.json'),
            'recordedAuditFindings': len(recorded_audit['findings']),
            'recordedGitCommit': m.get('git_commit'), 'caseCount': len(cases),
            'caseStatuses': {s: sum(c['status'] == s for c in cases) for s in sorted({c['status'] for c in cases})}}


def seal():
    env = {k: os.environ[k] for k in ['PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH'] if k in os.environ}
    subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(FOLDER)],
                   cwd=ROOT, env=env, check=True, timeout=180)
    write('sha256.json', {str(p.relative_to(FOLDER)): sha(p) for p in sorted(FOLDER.rglob('*'))
                         if p.is_file() and p.name not in ['sha256.json', 'verification.json']})


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--gate', type=Path)
    p.add_argument('--collection', type=Path)
    p.add_argument('--seal', action='store_true')
    a = p.parse_args()
    if a.seal:
        seal()
        return
    if not a.gate or not a.collection:
        p.error('--gate and --collection are required for packing')
    gate = (ROOT / a.gate).resolve()
    collection = (ROOT / a.collection).resolve()
    if not gate.is_relative_to(ROOT) or not collection.is_relative_to(ROOT):
        raise ValueError('inputs must belong to this repository')
    report, exit_record = read(gate / 'release-report.json'), read(gate / 'exit.json')
    if report['pass'] != (exit_record['exitCode'] == 0):
        raise ValueError('verdict and actual exit disagree')
    bundles = []
    for role in ['offline', 'race']:
        source = bundle_under(gate / role)
        name = role + '.zip'
        with zipfile.ZipFile(FOLDER / name, 'w', zipfile.ZIP_DEFLATED) as z:
            archive_tree(z, role, source)
        bundles.append(bundle_record(role, name, role, source))
    records = read(collection)
    with zipfile.ZipFile(FOLDER / 'live-evaluation.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for r in records:
            source = ROOT / r['bundle']
            role = r['route']
            archive_tree(z, role, source)
            bundles.append(bundle_record(role, 'live-evaluation.zip', role, source))
        source = ROOT / '.scratch/barness-ai-openai-decisions/chinese-live-evidence/live'
        archive_tree(z, 'luna', source)
        bundles.append(bundle_record('luna', 'live-evaluation.zip', 'luna', source))
    history = []
    with zipfile.ZipFile(FOLDER / 'history.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for source in sorted((ROOT / '.evidence').glob('issue13-*')):
            if source == gate or source == collection.parent or not source.is_dir():
                continue
            # Preserve all original case files and manifests. Only top-level
            # failed command diagnostics replace private host paths.
            excluded = ()
            if source.name == 'issue13-gate':
                # Separate full normal/race histories to keep each tracked
                # archive below hosting single-file limits; no cases are dropped.
                excluded = ('offline', 'race')
                for role in excluded:
                    name = 'history-initial-' + role + '.zip'
                    with zipfile.ZipFile(FOLDER / name, 'w', zipfile.ZIP_DEFLATED) as part:
                        archive_tree(part, source.name + '/' + role, source / role)
            archive_tree(z, source.name, source, diagnostics=True, excluded=excluded)
            history.append(str(source.relative_to(ROOT)))
    for name in ['release-report.json', 'release-report.md', 'fixtures.json', 'catalog-snapshot.json',
                 'support-matrix.json', 'traceability.json', 'ledger.json', 'command.json', 'command.log', 'exit.json']:
        (FOLDER / name).write_bytes(redact((gate / name).read_bytes()))
    shutil.copyfile(collection, FOLDER / 'collection.json')
    shutil.copyfile(collection.parent / 'execution-source.zip', FOLDER / 'live-execution-source.zip')
    logs = FOLDER / 'logs'
    logs.mkdir(exist_ok=True)
    for path in sorted(Path('/tmp').glob('barness-issue13*.log')):
        (logs / path.name).write_bytes(redact(path.read_bytes()))
    for record in records:
        folder = ROOT / record['bundle']
        for name in ['execution.log', 'merge.log', 'receipt.json']:
            path = folder.parent.parent / name
            if path.exists():
                (logs / (record['route'] + '-' + name)).write_bytes(redact(path.read_bytes()))
    source_hashes = {}
    names = subprocess.check_output(['git', 'ls-files', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()
    with zipfile.ZipFile(FOLDER / 'tested-source.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for name in names:
            path = ROOT / name
            if path.suffix not in ['.go', '.json'] and name not in ['go.mod', 'go.sum']:
                continue
            committed = subprocess.check_output(['git', 'show', report['commit'] + ':' + name], cwd=ROOT)
            if hashlib.sha256(committed).hexdigest() != sha(path):
                raise ValueError('tested source commit mismatch: ' + name)
            source_hashes[name] = sha(path)
            z.write(path, name)
    write('source-hashes.json', source_hashes)
    write('manifest.json', {
        'schema': 1, 'testedCommit': report['commit'], 'actualExitCode': exit_record['exitCode'],
        'bundles': bundles, 'retainedHistory': history,
        'provenance': 'Original bundle bytes, hashes, timestamps and recorded HEAD labels are preserved. Live binaries were built before the implementation commit; collection receipts and live-execution-source.zip identify that working tree. Runtime git_commit is a checkout label, not a binary build attestation. The first gate started before review fixes and failed the stale current-matrix migration test; final gate runs the committed fixes. Only top-level command diagnostics redact host paths.',
        'lunaReuse': 'Original issue 12 live package; current execution source set and public capture replay verified without rewriting its original code hash or time.',
        'accountLimits': 'Account region, permissions, dedicated key and provider spending limits were not independently verified; original aliases record unknown.'})
    blocker = ROOT / '.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md'
    original = ROOT / '.scratch/barness-ai-thinking-sampling/release-evidence/audit-timeout-gate.tar.gz'
    status = next((line.split(':', 1)[1].strip() for line in blocker.read_text().splitlines()
                   if line.startswith('Status:')), 'unknown')
    if status == 'resolved':
        raise ValueError('historical blocker changed; review its resolution before regenerating delivery')
    (FOLDER / 'known-pressure-blocker.md').write_bytes(blocker.read_bytes())
    write('delivery-verdict.json', {
        'schema': 1, 'status': 'FAIL', 'completed': False,
        'measuredReleaseGatePass': report['pass'], 'measuredReleaseGateExitCode': exit_record['exitCode'],
        'issue13Status': 'claimed', 'adr0026Status': 'proposed',
        'blockers': [{
            'issue': str(blocker.relative_to(ROOT)), 'issueSHA256': sha(blocker), 'status': status,
            'originalFailureArchive': str(original.relative_to(ROOT)), 'archiveSHA256': sha(original),
            'case': 'E08-policy-pressure-cloud-interactive-design-load',
            'reason': 'Unconfirmed pressure stream corruption from the earlier release remains unresolved. Passing current normal/race or ten diagnostic loops does not establish root cause or a fix.',
            'closure': 'Resolve the original issue with confirmed cause, regression evidence and unchanged design-load checks; then rerun final release acceptance.'}],
        'contract': 'Issue 13 requires old unresolved failures to remain recorded and block final completion. This delivery verdict does not rewrite measured command or artifact results.'})
    seal()
    print('packed', len(bundles), 'current bundles and', len(history), 'history trees')


if __name__ == '__main__':
    main()
