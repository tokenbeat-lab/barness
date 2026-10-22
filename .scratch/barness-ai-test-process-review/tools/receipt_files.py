"""Bounded receipt IO and execution-source snapshots, independent of test policy."""
import hashlib
import json
import os
from pathlib import Path
import subprocess

MAX_FILE = 64 << 20


def sha(path):
    path = Path(path)
    if path.is_symlink() or not path.is_file() or path.stat().st_size > MAX_FILE:
        raise ValueError("regular bounded file required")
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(65536), b''):
            h.update(block)
    return h.hexdigest()


def json_file(path):
    path = Path(path)
    sha(path)
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result:
                raise ValueError("duplicate JSON field")
            result[key] = value
        return result
    return json.loads(path.read_text(), object_pairs_hook=pairs)


def save(path, value):
    path = Path(path)
    raw = (json.dumps(value, sort_keys=True, indent=2, allow_nan=False) + '\n').encode()
    with path.open('wb') as stream:
        stream.write(raw)
        stream.flush()
        os.fsync(stream.fileno())


def within(root, relative):
    rel = Path(relative)
    if rel.is_absolute() or '..' in rel.parts or not rel.parts:
        raise ValueError("canonical relative path required")
    target = root / rel
    for parent in (target, *target.parents):
        if parent == root:
            break
        if parent.is_symlink():
            raise ValueError("symbolic link rejected")
    return target


def source_snapshot(root):
    def git(*args):
        return subprocess.check_output(['git', '-C', str(root), *args], stderr=subprocess.DEVNULL)
    files = {}
    excluded_dirs = {'node_modules', '__pycache__', '.git', '.evidence'}
    excluded_suffixes = {'.md', '.pyc', '.test', '.exe'}
    for scope in ('ai', '.scratch/barness-ai-test-process-review/tools'):
        base = root / scope
        if base.is_symlink():
            raise ValueError('symbolic execution source rejected')
        if not base.exists():
            continue
        for parent, dirs, names in os.walk(base, followlinks=False):
            dirs[:] = [name for name in dirs if name not in excluded_dirs]
            for name in dirs:
                if (Path(parent) / name).is_symlink():
                    raise ValueError('symbolic execution source rejected')
            for name in names:
                path = Path(parent) / name
                if path.suffix.lower() in excluded_suffixes:
                    continue
                relative = str(path.relative_to(root))
                files[relative] = sha(within(root, relative))
                if len(files) > 200000:
                    raise ValueError('execution source inventory exceeds bound')
    for name in ('go.mod', 'go.sum'):
        if (root / name).exists():
            files[name] = sha(within(root, name))
    if not files:
        raise ValueError('empty execution source')
    return {'commit': git('rev-parse', 'HEAD').decode().strip(), 'files': files}


def verify_bundle(bundle, package):
    manifest = json_file(bundle / 'manifest.json')
    audit = json_file(bundle / 'audit.json')
    if manifest.get('package') != package or audit.get('findings') != []:
        raise ValueError('wrong package or failed audit')
    cases = manifest.get('cases')
    if not isinstance(cases, list) or not cases:
        raise ValueError('empty evidence cases')
    seen = set()
    passed = 0
    for case in cases:
        identifier = case.get('id')
        if not identifier or identifier in seen or case.get('status') not in ('PASS', 'NOT_RUN', 'UNSUPPORTED'):
            raise ValueError('invalid case outcome')
        seen.add(identifier)
        passed += case['status'] == 'PASS'
        case_dir = within(bundle, case.get('dir', '').rstrip('/'))
        inventory = case.get('artifacts')
        if not isinstance(inventory, dict) or not inventory:
            raise ValueError('missing artifact inventory')
        for name, want in inventory.items():
            if sha(within(case_dir, name)) != want.removeprefix('sha256:'):
                raise ValueError('artifact hash mismatch')
        actual = {str(p.relative_to(case_dir)) for p in case_dir.rglob('*') if p.is_file()}
        if actual != set(inventory):
            raise ValueError('artifact inventory mismatch')
    if not passed:
        raise ValueError('no passing evidence')
    return manifest
