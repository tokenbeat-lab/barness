#!/usr/bin/env python3
"""Recheck archive integrity, complete case artifacts, source and redaction."""
import hashlib
import json
from pathlib import Path, PurePosixPath
import subprocess
import tarfile
import tempfile


def digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def require(ok, message):
    if not ok:
        raise SystemExit(message)


def child(root, name):
    path = PurePosixPath(name)
    require(not path.is_absolute() and '..' not in path.parts and '\\' not in name, 'unsafe artifact path')
    return root.joinpath(*path.parts)


def main():
    here = Path(__file__).resolve().parent
    root = here.parents[2]
    hashes = json.loads((here / 'sha256.json').read_text())
    for name, expected in hashes.items():
        require(digest(child(here, name)) == expected, f'changed delivery file: {name}')
    report = json.loads((here / 'verification-plan.json').read_text())
    findings = []
    # Fixture hashes identify original inputs; artifact hashes identify bytes
    # after Run's registered-secret redaction (existing evidence semantics).
    fixture_inputs = {digest(p): p for p in (root / 'ai/e2e/testdata').rglob('*') if p.is_file()}
    for archive in report['archives']:
        name = archive['file']
        with tempfile.TemporaryDirectory() as tmp:
            work = Path(tmp)
            with tarfile.open(here / name, 'r:gz') as tar:
                for member in tar.getmembers():
                    target = child(work, member.name)
                    require(member.isfile() or member.isdir(), 'links and special files are refused')
                    if member.isdir():
                        target.mkdir(parents=True, exist_ok=True)
                    else:
                        target.parent.mkdir(parents=True, exist_ok=True)
                        with tar.extractfile(member) as src, target.open('wb') as dst:
                            while block := src.read(1 << 20):
                                dst.write(block)
            manifests = []
            for path in work.rglob('manifest.json'):
                data = json.loads(path.read_text())
                if data.get('module') == 'barness-ai' and 'cases' in data:
                    manifests.append((path, data))
            require(len(manifests) == archive['bundles'], f'wrong bundle count: {name}')
            counts = {}
            redacted_fixtures = 0
            for path, data in manifests:
                for case in data['cases']:
                    counts[case['status']] = counts.get(case['status'], 0) + 1
                    folder = child(path.parent, case['dir'])
                    for file, expected in case['artifacts'].items():
                        require(digest(child(folder, file)) == expected, f'changed case artifact: {name}/{case["id"]}/{file}')
                    for file, expected in case['fixtures'].items():
                        require(file in case['artifacts'], f'fixture artifact absent: {name}/{case["id"]}/{file}')
                        if digest(child(folder, file)) != expected:
                            require(expected in fixture_inputs, f'original fixture input absent: {name}/{case["id"]}/{file}')
                            redacted_fixtures += 1
            require(counts == archive['statuses'], f'wrong verdicts: {name}')
            # Audit the entire extraction, including command logs and receipts.
            # The audit CLI requires a manifest at its input root. Its envelope
            # is separate from the preserved/verified E2E manifests.
            if not (work / 'manifest.json').exists():
                (work / 'manifest.json').write_text(json.dumps({'module': 'barness-pressure-archive-audit', 'archive': name}) + '\n')
            audited = subprocess.run(['go', 'run', './ai/internal/testkit/audit/cmd/auditbundle', str(work)], cwd=root, capture_output=True)
            require(audited.returncode == 0, f'redaction audit failed: {name}')
            if archive.get('currentSource'):
                source = json.loads((work / 'source-hashes.json').read_text())
                files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=root, text=True).splitlines()
                current = {n for n in files if (root / n).is_file() and (n in ('go.mod', 'go.sum', 'AGENTS.md') or n.startswith('ai/'))}
                require(current == set(source), f'changed source set: {name}')
                for file, expected in source.items():
                    require(digest(child(root, file)) == expected, f'changed tested source: {file}')
            findings.append({'archive': name, 'bundles': len(manifests), 'statuses': counts, 'audit': 'PASS', 'redactedFixtureCopies': redacted_fixtures})
    print(json.dumps({'verified': findings}, indent=2))


if __name__ == '__main__':
    main()
