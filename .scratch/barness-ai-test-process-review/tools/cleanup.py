#!/usr/bin/env python3
"""Maintain an explicitly reviewed evidence batch; no tests or vendor execution."""
import argparse
from datetime import datetime, timedelta, timezone
import json
import os
from pathlib import Path
import shutil
import sys
import tempfile

from cleanup_scan import inventory, scan
from cleanup_state import (batch, baseline_check, fresh_check, journal, load_plan, locked,
                           moves, now, payload_check, sync_dir, verified_state)
from receipt_files import json_file, save, sha, within


def create(root, parent):
    snapshot = scan(root)
    base = Path(tempfile.mkdtemp(prefix=datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ-'), dir=parent))
    meta = base / 'metadata'
    meta.mkdir()
    save(meta / 'identity.json', {'schema': 1, 'batch': base.name, 'repo': str(root)})
    for name, value in [('sources.json', snapshot['sources']), ('baseline.json', snapshot['baseline']),
                        ('preexisting-missing.json', snapshot['missing']), ('snapshot.json', snapshot)]:
        save(meta / name, value)
    for name, rows in [('refs.jsonl', snapshot['refs']), ('classifications.jsonl', snapshot['refs'])]:
        (meta / name).write_text(''.join(json.dumps(row, sort_keys=True) + '\n' for row in rows))
    save(meta / 'state.json', {'status': 'SCANNED', 'createdAt': now(),
                             'baselineSHA256': sha(meta / 'baseline.json')})
    return {'batch': base.name, 'preexistingMissing': len(snapshot['missing'])}


def make_plan(root, meta):
    state = json_file(meta / 'state.json')
    if state['status'] != 'SCANNED':
        raise ValueError('only an unmoved scan can be planned')
    snapshot = json_file(meta / 'snapshot.json')
    selected = [row['path'] for row in snapshot['candidates'] if row['mark'] == 'QUARANTINE']
    fresh_check(root, meta, selected, unchanged=True)
    plan = {'schema': 1, 'candidates': snapshot['candidates'],
            'inventories': {name: inventory(within(root, name)) for name in selected}}
    save(meta / 'plan.json', plan)
    state.update(status='PLANNED', planSHA256=sha(meta / 'plan.json'))
    save(meta / 'state.json', state)
    return {'planSHA256': state['planSHA256'], 'candidates': plan['candidates'], 'new_missing': 0}


def move(root, base, meta, expected):
    state, plan, selected = load_plan(meta, expected)
    if state['status'] != 'PLANNED' or moves(meta):
        raise ValueError('batch is not an untouched reviewed plan')
    fresh_check(root, meta, selected, unchanged=True)
    for name in selected:
        original = within(root, name)
        destination = within(base / 'payload', name)
        if destination.exists() or inventory(original) != plan['inventories'][name]:
            raise ValueError('movement source changed or destination exists')
        if original.stat().st_dev != base.stat().st_dev:
            raise ValueError('cross-filesystem move rejected')
    state.update(status='MOVING', lastMutationAt=now())
    save(meta / 'state.json', state)
    try:
        for name in selected:
            original, destination = within(root, name), within(base / 'payload', name)
            destination.parent.mkdir(parents=True, exist_ok=True)
            journal(meta, {'event': 'MOVE_INTENT', 'path': name})
            original.rename(destination)
            sync_dir(original.parent)
            sync_dir(destination.parent)
            journal(meta, {'event': 'MOVED', 'path': name})
        state.update(status='MOVED', lastMutationAt=now())
        save(meta / 'state.json', state)
    except BaseException:
        state['status'] = 'MOVE_FAILED'
        save(meta / 'state.json', state)
        raise
    return {'status': state['status'], 'moved': len(selected)}


def verify(root, base, meta):
    state, _plan, selected = load_plan(meta)
    if state['status'] not in ('MOVED', 'VERIFIED'):
        raise ValueError('batch is not ready for verification')
    try:
        fresh_check(root, meta, selected)
        payload_check(base, meta, selected)
    except BaseException:
        state['status'] = 'VERIFY_FAILED'
        save(meta / 'state.json', state)
        raise
    # Reinspection must not silently restart an unchanged observation period.
    if state['status'] != 'VERIFIED':
        verified_state(meta, state)
    save(meta / 'verification.json', {'status': 'PASS', 'new_missing': 0, 'at': now(),
                                    'planSHA256': state['planSHA256']})
    return {'status': 'VERIFIED', 'new_missing': 0, 'verifiedAt': state.get('verifiedAt')}


def restore(root, base, meta):
    state, plan, selected = load_plan(meta)
    if state['status'] in ('SCANNED', 'PLANNED', 'PURGED'):
        raise ValueError('batch has no restorable payload')
    for name in reversed(selected):
        original, payload = within(root, name), within(base / 'payload', name)
        if not payload.exists():
            # Intent persisted but rename did not happen, or earlier restore completed.
            if inventory(original) != plan['inventories'][name]:
                raise ValueError('ambiguous missing payload')
            continue
        if original.exists() or inventory(payload) != plan['inventories'][name]:
            raise ValueError('restore would overwrite or payload changed')
    for name in reversed(selected):
        original, payload = within(root, name), within(base / 'payload', name)
        if payload.exists():
            original.parent.mkdir(parents=True, exist_ok=True)
            journal(meta, {'event': 'RESTORE_INTENT', 'path': name})
            payload.rename(original)
            sync_dir(original.parent)
            sync_dir(payload.parent)
            journal(meta, {'event': 'RESTORED', 'path': name})
    baseline_check(root, meta)
    state.update(status='RESTORED', lastMutationAt=now())
    state.pop('verifiedAt', None)
    save(meta / 'state.json', state)
    return {'status': 'RESTORED'}


def inspect(root, base, meta, expected=None):
    state, _plan, selected = load_plan(meta, expected)
    if state['status'] != 'VERIFIED':
        raise ValueError('only verified batches may be purged')
    for field in ('verifiedAt', 'lastMutationAt'):
        stamp = datetime.fromisoformat(state[field])
        if stamp.tzinfo is None or datetime.now(timezone.utc) - stamp < timedelta(days=7):
            raise ValueError('seven-day observation period incomplete')
    fresh_check(root, meta, selected)
    payload_check(base, meta, selected)
    save(meta / 'deletion-check.json', {'status': 'PASS', 'at': now(), 'planSHA256': state['planSHA256']})
    return state


def purge(root, base, meta, expected):
    state = inspect(root, base, meta, expected)
    payload = base / 'payload'
    total = sum(p.stat().st_size for p in payload.rglob('*') if p.is_file())
    state['status'] = 'PURGING'
    save(meta / 'state.json', state)
    try:
        shutil.rmtree(payload)
        baseline_check(root, meta)
        state.update(status='PURGED', purgedAt=now(), reclaimedBytes=total)
    except BaseException:
        state['status'] = 'PURGE_FAILED'
        raise
    finally:
        save(meta / 'state.json', state)
    return {'status': 'PURGED', 'reclaimedBytes': total}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=('scan', 'plan', 'move', 'verify', 'restore', 'inspect', 'purge'))
    parser.add_argument('--repo', type=Path, default=Path.cwd())
    parser.add_argument('--batch')
    parser.add_argument('--plan-sha256')
    args = parser.parse_args()
    root = args.repo.resolve()
    try:
        if args.action != 'scan' and not args.batch:
            raise ValueError('exact batch identity required')
        if args.action in ('move', 'purge') and not args.plan_sha256:
            raise ValueError('reviewed plan hash required')
        with locked(root) as parent:
            if args.action == 'scan':
                result = create(root, parent)
            else:
                base, meta = batch(root, args.batch)
                if args.action == 'plan':
                    result = make_plan(root, meta)
                elif args.action in ('move', 'purge'):
                    result = globals()[args.action](root, base, meta, args.plan_sha256)
                elif args.action == 'inspect':
                    result = inspect(root, base, meta)
                else:
                    result = globals()[args.action](root, base, meta)
            print(json.dumps(result, sort_keys=True))
        return 0
    except (OSError, ValueError, KeyError, TypeError) as error:
        print('cleanup: ' + str(error), file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
