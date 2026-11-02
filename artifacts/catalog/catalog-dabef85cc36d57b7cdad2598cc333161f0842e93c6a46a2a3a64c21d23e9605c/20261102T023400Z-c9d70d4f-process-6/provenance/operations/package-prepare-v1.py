#!/usr/bin/env python3
"""Prepare evidence copies in an existing delivery batch; never seal or publish it.

The operator supplies a separately reviewed JSON plan with repository-relative
source paths, expected SHA-256 digests and batch-relative destination paths.
A fresh plan is required for each batch. No runtime snapshot or review file is
written by this script. Original evidence remains in place.
"""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import shutil

ROOT = Path(__file__).resolve().parents[4]
CATALOG = ROOT / 'artifacts' / 'catalog'
ALLOWED = {'provenance', 'verification'}


def sha256(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def relative(value):
    path = PurePosixPath(value)
    if path.is_absolute() or not path.parts or any(p in ('.', '..') for p in path.parts):
        raise ValueError(f'unsafe relative path: {value!r}')
    return path


def within(base, path):
    resolved = path.resolve(strict=True)
    if not resolved.is_relative_to(base.resolve(strict=True)):
        raise ValueError(f'path escapes root: {path}')
    return resolved


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--batch', required=True, help='existing repository-relative version/batch directory')
    parser.add_argument('--plan', required=True, help='reviewed repository-relative JSON copy plan')
    args = parser.parse_args()
    batch = within(CATALOG, ROOT / relative(args.batch))
    if not batch.is_dir() or (batch / 'delivery.json').exists():
        raise ValueError('batch must exist and must not have a completion index')
    plan_path = within(ROOT, ROOT / relative(args.plan))
    plan = json.loads(plan_path.read_text(encoding='utf-8'))
    if plan.get('status') != 'READY_FOR_COPY' or not isinstance(plan.get('files'), list):
        raise ValueError('plan must explicitly authorize an enumerated copy')
    pending = []
    seen = set()
    for item in plan['files']:
        source = within(ROOT, ROOT / relative(item['source']))
        dest_rel = relative(item['destination'])
        if dest_rel.parts[0] not in ALLOWED or dest_rel in seen:
            raise ValueError(f'destination not allowed or duplicated: {dest_rel}')
        dest = batch.joinpath(*dest_rel.parts)
        if dest.exists() or not source.is_file():
            raise ValueError(f'existing destination or non-file source: {dest_rel}')
        expected = item['sha256'].removeprefix('sha256:')
        if len(expected) != 64 or sha256(source) != expected:
            raise ValueError(f'source digest mismatch: {source}')
        # Existing parents must remain inside the batch, including symlinks.
        parent = dest.parent
        while not parent.exists():
            parent = parent.parent
        within(batch, parent)
        seen.add(dest_rel)
        pending.append((source, dest, expected))
    for source, dest, expected in pending:
        dest.parent.mkdir(parents=True, exist_ok=True)
        with source.open('rb') as src, dest.open('xb') as dst:
            shutil.copyfileobj(src, dst)
        if sha256(dest) != expected:
            raise ValueError(f'copied digest mismatch: {dest}')
    print(json.dumps({'status': 'COPIED_NOT_COMPLETE', 'files': len(pending), 'completionIndexWritten': False}))


if __name__ == '__main__':
    main()
