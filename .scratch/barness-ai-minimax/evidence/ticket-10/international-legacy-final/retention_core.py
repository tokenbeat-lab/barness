"""Bounded, no-follow, byte-preserving primitives. No captured command executes."""
from pathlib import Path, PurePosixPath
from collections import Counter
import gzip
import hashlib
import io
import json
import os
import stat
import subprocess
import tarfile


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def digest(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for raw in iter(lambda: f.read(1 << 20), b''):
            h.update(raw)
    return h.hexdigest()


def stable(path, limit=128 << 20):
    path = Path(path)
    if path.name == '.env':
        raise ValueError('credential source is outside retention scope')
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW)
    with os.fdopen(fd, 'rb') as f:
        before = os.fstat(f.fileno())
        if not stat.S_ISREG(before.st_mode) or before.st_size > limit:
            raise ValueError('non-regular or oversized source')
        raw = f.read(limit + 1)
        after = os.fstat(f.fileno())
    final = path.stat(follow_symlinks=False)
    fields = lambda v: (v.st_dev, v.st_ino, v.st_size, v.st_mtime_ns)
    if len(raw) != before.st_size or len(raw) > limit or fields(before) != fields(after) or fields(after) != fields(final):
        raise ValueError('source changed while reading')
    return raw


def safe_name(name):
    p = PurePosixPath(name)
    if not name or name == '.' or p.is_absolute() or '..' in p.parts or len(p.parts) > 24:
        raise ValueError('unsafe archive name')
    return name


def scan(base, excluded=()):
    base = Path(base)
    if base.is_symlink() or not base.is_dir():
        raise ValueError('raw root must be a real directory')
    excluded = {str(Path(p).absolute()) for p in excluded}
    result = {}
    stack = [(base, 0)]
    visited = 0
    total = 0
    while stack:
        folder, depth = stack.pop()
        if depth > 24:
            raise ValueError('directory depth bound')
        with os.scandir(folder) as entries:
            for entry in entries:
                visited += 1
                if visited > 400000:
                    raise ValueError('directory entry bound; partition this raw group')
                p = Path(entry.path)
                if str(p.absolute()) in excluded:
                    continue
                info = entry.stat(follow_symlinks=False)
                if stat.S_ISDIR(info.st_mode):
                    stack.append((p, depth + 1))
                    continue
                key = safe_name(p.relative_to(base).as_posix())
                if stat.S_ISLNK(info.st_mode):
                    raw = os.fsencode(os.readlink(p))
                    if len(raw) > 4096:
                        raise ValueError('symlink text bound')
                    pin = {'bytes': len(raw), 'sha256': sha(raw), 'kind': 'symlink-fixture'}
                else:
                    if not stat.S_ISREG(info.st_mode):
                        raise ValueError('special file outside retention scope')
                    raw = stable(p)
                    pin = {'bytes': len(raw), 'sha256': sha(raw)}
                total += len(raw)
                if total > 4 << 30:
                    raise ValueError('raw group total bound; partition this group')
                result[key] = (p, pin)
    return result


def verify_tar(archive, inventory):
    seen = set()
    total = 0
    with tarfile.open(archive, 'r|gz') as tar:
        for member in tar:
            key = safe_name(member.name)
            if key not in inventory or key in seen:
                raise ValueError('invented or repeated archive member')
            seen.add(key)
            pin = inventory[key]
            if member.issym():
                raw = os.fsencode(member.linkname)
                if pin.get('kind') != 'symlink-fixture' or len(raw) > 4096:
                    raise ValueError('unexpected link fixture')
            else:
                if not member.isfile() or pin.get('kind') or member.size > 128 << 20:
                    raise ValueError('unsafe archive member')
                raw = tar.extractfile(member).read()
            total += len(raw)
            if total > 4 << 30 or len(raw) != pin['bytes'] or sha(raw) != pin['sha256']:
                raise ValueError('archive member byte mismatch/bound')
    if seen != set(inventory):
        raise ValueError('archive omitted files')
    return {'entries': len(seen), 'expandedBytes': total}


def pack_bytes(path, items):
    with Path(path).open('xb') as target:
        with gzip.GzipFile(filename='', fileobj=target, mode='wb', mtime=0) as gz:
            with tarfile.open(fileobj=gz, mode='w') as tar:
                for key, raw in sorted(items.items()):
                    safe_name(key)
                    info = tarfile.TarInfo(key)
                    info.size, info.mode, info.mtime = len(raw), 0o600, 0
                    tar.addfile(info, io.BytesIO(raw))


def pack_tree(base, path, excluded=()):
    files = scan(base, excluded)
    inventory = {name: pin for name, (_, pin) in files.items()}
    with Path(path).open('xb') as target:
        with gzip.GzipFile(filename='', fileobj=target, mode='wb', mtime=0) as gz:
            with tarfile.open(fileobj=gz, mode='w') as tar:
                for key, (source, pin) in sorted(files.items()):
                    info = tarfile.TarInfo(key)
                    info.mode = 0o600
                    if pin.get('kind'):
                        raw = os.fsencode(os.readlink(source))
                        if len(raw) != pin['bytes'] or sha(raw) != pin['sha256']:
                            raise ValueError('link changed before archive')
                        info.type, info.linkname = tarfile.SYMTYPE, os.fsdecode(raw)
                        tar.addfile(info)
                    else:
                        raw = stable(source)
                        if len(raw) != pin['bytes'] or sha(raw) != pin['sha256']:
                            raise ValueError('file changed before archive')
                        info.size = len(raw)
                        tar.addfile(info, io.BytesIO(raw))
    final_inventory = {name: pin for name, (_, pin) in scan(base, excluded).items()}
    if final_inventory != inventory:
        raise ValueError('raw tree changed during archive')
    summary = verify_tar(path, inventory)
    inv_path = Path(str(path).removesuffix('.tar.gz') + '-files.json.gz')
    raw = json.dumps(inventory, ensure_ascii=False, sort_keys=True, separators=(',', ':')).encode()
    with inv_path.open('xb') as f:
        with gzip.GzipFile(filename='', fileobj=f, mode='wb', mtime=0) as gz:
            gz.write(raw)
    return {'archive': Path(path).name, 'sha256': digest(path), 'bytes': Path(path).stat().st_size,
            'inventory': inv_path.name, 'inventorySha256': digest(inv_path), **summary}, inventory


def main_manifests(base, inventory):
    candidates = []
    for name, pin in inventory.items():
        if PurePosixPath(name).name != 'manifest.json' or len(PurePosixPath(name).parts) > 3 or pin.get('kind'):
            continue
        candidates.append(name)
    results = []
    if not candidates:
        return results
    for name in candidates:
        parent = PurePosixPath(name).parent
        if any(PurePosixPath(other).parent in parent.parents for other in candidates if other != name):
            continue
        value = json.loads(stable(Path(base) / name, 16 << 20))
        if not isinstance(value.get('cases'), list):
            if not isinstance(value.get('artifacts'), dict):
                continue
            refs = 0
            for file, pin in value['artifacts'].items():
                key = safe_name(str(parent / file))
                if key not in inventory or inventory[key]['sha256'] != pin:
                    raise ValueError('outer direct artifact reference differs')
                refs += 1
            results.append({'manifest': name, 'caseOutcomes': {},
                            'artifactReferences': refs, 'auditFindings': None})
            continue
        refs = 0
        for case in value['cases']:
            for file, pin in case.get('artifacts', {}).items():
                key = safe_name(str(parent / case['dir'] / file))
                if key not in inventory or inventory[key]['sha256'] != pin:
                    raise ValueError('outer actual artifact reference differs')
                refs += 1
        audit = Path(base) / str(parent / 'audit.json')
        findings = len(json.loads(stable(audit, 16 << 20))['findings']) if audit.is_file() else None
        results.append({'manifest': name, 'caseOutcomes': dict(Counter(case['status'] for case in value['cases'])),
                        'artifactReferences': refs, 'auditFindings': findings})
    return results


def committed_sources(repo, commit, manifest):
    paths = list(manifest['files'])
    for name in paths:
        safe_name(name)
        if '\n' in name or Path(name).name == '.env':
            raise ValueError('unsafe source query')
    request = ''.join(commit + ':' + name + '\n' for name in paths).encode()
    raw = subprocess.run(['git', 'cat-file', '--batch'], cwd=repo, input=request,
                         stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True).stdout
    offset = 0
    out = {}
    for name in paths:
        end = raw.index(b'\n', offset)
        header = raw[offset:end].split()
        if len(header) != 3 or header[1] != b'blob' or int(header[2]) > 8 << 20:
            raise ValueError('unavailable/large source blob')
        size = int(header[2])
        blob = raw[end + 1:end + 1 + size]
        offset = end + 2 + size
        if sha(blob) != manifest['files'][name]:
            raise ValueError('original committed source pin differs')
        out['source/' + name] = blob
    if offset != len(raw) or sum(map(len, out.values())) > 32 << 20:
        raise ValueError('source blob total/batch bound')
    return out


def retained_references(members, inventory):
    """Validate outer actual producer maps; inner negative fixtures are not authority."""
    roots = [name for name in inventory if PurePosixPath(name).name == 'manifest.json'
             and len(PurePosixPath(name).parts) <= 2 and not inventory[name].get('kind')]
    refs = 0
    checks = []
    for name in roots:
        parent = PurePosixPath(name).parent
        value = json.loads(members[name])
        maps = [(name, value.get('artifacts', {}))]
        for filename in ['files.json', 'minimax-integrity.json', 'multimodal-integrity.json']:
            key = str(parent / filename)
            if key not in inventory:
                continue
            extra = json.loads(members[key])
            maps.append((key, extra.get('files', {}) if filename != 'files.json' else extra))
        for key, mapping in maps:
            count = 0
            for relative, expected in mapping.items():
                if not isinstance(expected, str):
                    raise ValueError('unexpected producer file-map schema')
                target = safe_name(str(parent / relative))
                if target not in inventory or inventory[target]['sha256'] != expected.removeprefix('sha256:'):
                    raise ValueError('outer producer file-map reference differs')
                count += 1
            if count:
                checks.append({'map': key, 'references': count})
                refs += count
    return {'references': refs, 'maps': checks}


def retained_main_manifests(members, inventory):
    candidates = [name for name, pin in inventory.items()
                  if PurePosixPath(name).name == 'manifest.json'
                  and len(PurePosixPath(name).parts) <= 3 and not pin.get('kind')]
    results = []
    for name in candidates:
        parent = PurePosixPath(name).parent
        if any(PurePosixPath(other).parent in parent.parents for other in candidates if other != name):
            continue
        value = json.loads(members[name])
        refs = 0
        cases = value.get('cases')
        if isinstance(cases, list):
            mappings = [(case.get('dir', ''), case.get('artifacts', {})) for case in cases]
            outcomes = dict(Counter(case['status'] for case in cases))
            audit = str(parent / 'audit.json')
            findings = len(json.loads(members[audit])['findings']) if audit in members else None
        elif isinstance(value.get('artifacts'), dict):
            mappings = [('', value['artifacts'])]
            outcomes, findings = {}, None
        else:
            continue
        for folder, mapping in mappings:
            for file, pin in mapping.items():
                key = safe_name(str(parent / folder / file))
                if key not in inventory or inventory[key]['sha256'] != pin:
                    raise ValueError('outer actual artifact reference differs')
                refs += 1
        results.append({'manifest': name, 'caseOutcomes': outcomes,
                        'artifactReferences': refs, 'auditFindings': findings})
    return sorted(results, key=lambda result: result['manifest'])
