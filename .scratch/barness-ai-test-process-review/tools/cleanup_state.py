"""Durable batch identity, inventory checks and append-only movement receipts."""
from contextlib import contextmanager
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import re

from cleanup_scan import inventory, intersects, scan
from receipt_files import json_file, save, sha, within


def now():
    return datetime.now(timezone.utc).isoformat()


def sync_dir(path):
    fd = os.open(path, os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


@contextmanager
def locked(root):
    parent = root / '.evidence/_quarantine'
    within(root, '.evidence/_quarantine')
    parent.mkdir(parents=True, exist_ok=True)
    path = parent / 'LOCK'
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    try:
        os.write(fd, (str(os.getpid()) + '\n').encode())
        os.fsync(fd)
        sync_dir(parent)
        yield parent
    finally:
        os.close(fd)
        path.unlink()
        sync_dir(parent)


def batch(root, identifier):
    if not re.fullmatch(r'[0-9TZ.-]+-[a-z0-9_]+', identifier):
        raise ValueError('invalid batch identity')
    base = within(root, '.evidence/_quarantine/' + identifier)
    metadata = base / 'metadata'
    identity = json_file(metadata / 'identity.json')
    if identity != {'schema': 1, 'batch': identifier, 'repo': str(root)}:
        raise ValueError('batch repository mismatch')
    return base, metadata


def journal(meta, record):
    with (meta / 'moves.jsonl').open('a') as stream:
        stream.write(json.dumps(dict(record, at=now()), sort_keys=True) + '\n')
        stream.flush()
        os.fsync(stream.fileno())
    sync_dir(meta)


def load_plan(meta, expected=None):
    state = json_file(meta / 'state.json')
    actual = sha(meta / 'plan.json')
    if state.get('planSHA256') != actual or expected is not None and expected != actual:
        raise ValueError('plan hash mismatch')
    plan = json_file(meta / 'plan.json')
    if state.get('baselineSHA256') != sha(meta / 'baseline.json'):
        raise ValueError('baseline hash mismatch')
    selected = [r['path'] for r in plan['candidates'] if r['mark'] == 'QUARANTINE']
    if len(selected) != len(set(selected)) or any(intersects(a, b) for i, a in enumerate(selected) for b in selected[i+1:]):
        raise ValueError('overlapping movement selection')
    return state, plan, selected


def baseline_check(root, meta):
    for name, want in json_file(meta / 'baseline.json').items():
        if inventory(within(root, name)) != want:
            raise ValueError('retained baseline missing or changed')


def fresh_check(root, meta, selected, unchanged=False):
    baseline_check(root, meta)
    fresh = scan(root)
    if any(intersects(candidate, keep) for candidate in selected for keep in fresh['protected']):
        raise ValueError('current reference protects selected payload')
    if fresh['unresolved']:
        raise ValueError('unresolved references block maintenance')
    if unchanged and fresh['sources'] != json_file(meta / 'sources.json'):
        raise ValueError('reference inputs changed; review a new batch')
    return fresh


def moves(meta):
    records = []
    path = meta / 'moves.jsonl'
    if path.exists():
        for line in path.read_text().splitlines():
            records.append(json.loads(line))
    return records


def payload_check(base, meta, selected):
    records = moves(meta)
    completed = [r['path'] for r in records if r['event'] == 'MOVED']
    intended = [r['path'] for r in records if r['event'] == 'MOVE_INTENT']
    if completed != selected or intended != selected:
        raise ValueError('movement journal incomplete or ambiguous')
    wanted = json_file(meta / 'plan.json')['inventories']
    actual_files = set()
    payload = base / 'payload'
    for relative in selected:
        path = within(payload, relative)
        if inventory(path) != wanted[relative]:
            raise ValueError('payload inventory mismatch')
        if path.is_file():
            actual_files.add(relative)
        else:
            actual_files.update(str(p.relative_to(payload)) for p in path.rglob('*') if p.is_file())
    found_files = set()
    for parent, dirs, files in os.walk(payload, followlinks=False):
        for name in dirs + files:
            path = Path(parent) / name
            if path.is_symlink():
                raise ValueError('symbolic payload entry rejected')
        found_files.update(str((Path(parent) / name).relative_to(payload)) for name in files)
    if found_files != actual_files:
        raise ValueError('unlisted payload file')


def verified_state(meta, state):
    state.update(status='VERIFIED', verifiedAt=now())
    save(meta / 'state.json', state)
    sync_dir(meta)
