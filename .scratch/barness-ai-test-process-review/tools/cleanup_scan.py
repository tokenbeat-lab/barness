"""Conservative reference inventory for repository-local evidence maintenance."""
import json
import os
from pathlib import Path
import re
import subprocess
from urllib.parse import unquote, urlsplit

from receipt_files import json_file, sha, within

REQUIRED = tuple('.evidence/provider-sites/final-acceptance-20261020/' + p for p in
                 ('recovery7-preparation/index.json', 'recovery6-stopped/index.json', 'closed-cohort-6/seal.json'))
TEXT = {'.md', '.json', '.jsonl', '.go', '.py', '.txt', '.yaml', '.yml', '.toml', '.sha256', '.sum'}
MAX_ENTRIES = 2000000


def inventory(path):
    if path.is_symlink():
        raise ValueError('symbolic candidate or retained path rejected')
    if path.is_file():
        return {'kind': 'file', 'sha256': sha(path)}
    if not path.is_dir():
        raise ValueError('inventory path missing')
    files = {}
    for parent, dirs, names in os.walk(path, followlinks=False):
        for name in dirs + names:
            if (Path(parent) / name).is_symlink():
                raise ValueError('symbolic inventory entry rejected')
        for name in names:
            member = Path(parent) / name
            files[str(member.relative_to(path))] = sha(member)
            if len(files) > MAX_ENTRIES:
                raise ValueError('inventory entry limit exceeded')
    return {'kind': 'directory', 'files': files}


def intersects(a, b):
    return a == b or a.startswith(b + '/') or b.startswith(a + '/')


def successful(path):
    receipts = list(path.rglob('receipt.json')) if path.is_dir() else []
    manifests = list(path.rglob('manifest.json')) if path.is_dir() else []
    if len(receipts) + len(manifests) > MAX_ENTRIES:
        raise ValueError('candidate state limit exceeded')
    if receipts:
        for receipt in receipts:
            value = json_file(receipt)
            if value.get('status') != 'PASS' or value.get('exitCode') != 0:
                return False
        return True
    if manifests:
        return all((value := json_file(p)).get('cases') and
                   all(c.get('status') in ('PASS', 'UNSUPPORTED') for c in value['cases']) for p in manifests)
    return False


def scan(root):
    root = root.resolve()
    for name in REQUIRED:
        sha(within(root, name))
    if not (root / '.evidence/minimax-retained').is_dir():
        raise ValueError('required minimax retained directory missing')
    names = subprocess.check_output(['git', '-C', str(root), 'ls-files', '-z', '--cached', '--others', '--exclude-standard'])
    sources, references, missing, unresolved = {}, [], [], []
    protected = set(REQUIRED) | {'.evidence/minimax-retained'}
    pending, seen = [], set()
    critical = {root / name for name in REQUIRED}
    for raw in names.split(b'\0'):
        if not raw:
            continue
        name = os.fsdecode(raw)
        if name.startswith('.evidence/') or '__pycache__' in Path(name).parts:
            continue
        path = within(root, name)
        if path.suffix in TEXT or path.name == 'SHA256SUMS':
            pending.append(path)
    pending += [root / name for name in REQUIRED]

    def reference(raw, origin, field, trim=False):
        raw = raw.strip()
        if len(raw) > 8192:
            raise ValueError('reference exceeds path bound')
        value = unquote(raw.split('#', 1)[0])
        if trim:
            value = value.rstrip('.,;，。；')
        if not value or '://' in value or value.startswith(('http:', 'https:', 'mailto:')):
            return
        if not any(x in value for x in ('.evidence', 'evidence/', '.log')) and not value.endswith(('.json', '.md')):
            return
        row = {'source': str(origin.relative_to(root)), 'field': str(field), 'raw': raw}
        if re.search(r'[<>${}*]|\bNEW_[A-Z_]+\b', value):
            references.append(dict(row, classification='EXAMPLE_OR_PREFIX', reason='explicit placeholder or pattern'))
            return
        if value.startswith('.evidence/') or value.startswith('.scratch/'):
            path = root / value
        elif Path(value).is_absolute():
            path = Path(value)
        else:
            path = origin.parent / value
        path = Path(os.path.abspath(path))
        try:
            relative = str(path.relative_to(root))
        except ValueError:
            return
        candidate = relative.startswith('.evidence/') or (relative.startswith('.scratch/') and '/evidence/' in relative)
        if not candidate or relative.startswith('.evidence/_quarantine/'):
            return
        within(root, relative)
        row['path'] = relative
        protected.add(relative)
        if origin in critical:
            critical.add(path)
            if not path.exists():
                raise ValueError('current required index dependency missing')
        if not path.exists():
            row['classification'] = 'PREEXISTING_MISSING'
            missing.append(row)
        else:
            row['classification'] = 'CONCRETE'
            # A reference to one native artifact retains its complete sealed bundle.
            for ancestor in (path if path.is_dir() else path.parent, *path.parents):
                if ancestor == root or ancestor == root / '.evidence':
                    break
                if (ancestor / 'manifest.json').is_file() or (ancestor / 'receipt.json').is_file():
                    protected.add(str(ancestor.relative_to(root)))
            if path.is_file() and (path.suffix in TEXT or path.name == 'SHA256SUMS'):
                pending.append(path)
            elif path.is_dir():
                for member in path.rglob('*'):
                    if member.is_file() and (member.suffix in TEXT or member.name == 'SHA256SUMS'):
                        if origin in critical:
                            critical.add(member)
                        pending.append(member)
                        if len(pending) > MAX_ENTRIES:
                            raise ValueError('reference expansion exceeds bound')
        references.append(row)

    while pending:
        path = pending.pop()
        if path in seen:
            continue
        seen.add(path)
        if len(seen) > MAX_ENTRIES:
            raise ValueError('reference source limit exceeded')
        relative = str(path.relative_to(root))
        sources[relative] = sha(path)
        if path.suffix == '.json':
            value = json_file(path)
            def walk(obj, field=''):
                if isinstance(obj, dict):
                    for key, member in obj.items():
                        reference(key, path, field + '/[key]')
                        walk(member, field + '/' + key)
                elif isinstance(obj, list):
                    for index, member in enumerate(obj):
                        walk(member, field + '/' + str(index))
                elif isinstance(obj, str):
                    reference(obj, path, field)
            walk(value)
        else:
            try:
                lines = path.read_text().splitlines()
            except UnicodeError:
                raise ValueError('text reference source is not UTF-8')
            for number, line in enumerate(lines, 1):
                matches = re.findall(r'\[[^\]]*\]\(<?([^\s)>]+)>?\)', line)
                matches += re.findall(r'`([^`\n]+)`', line)
                bare = re.findall(r'(?<![\w])(?:\.{0,2}/)*\.(?:evidence|scratch)/[A-Za-z0-9._+%/-]+', line)
                if path.name == 'SHA256SUMS':
                    match = re.match(r'^[a-fA-F0-9]{64}\s+\*?(.+)$', line)
                    if match:
                        matches.append(match[1])
                for match in matches:
                    reference(match, path, number)
                for match in bare:
                    reference(match, path, number, trim=True)
    # Keep only maximal retained directories to avoid duplicated baseline hashing.
    existing = sorted(p for p in protected if (root / p).exists())
    maximal = [p for p in existing if not any(p.startswith(parent + '/') and (root / parent).is_dir()
                                              for parent in existing if parent != p)]
    baseline = {p: inventory(root / p) for p in maximal}
    candidates = []
    evidence = root / '.evidence'
    for path in sorted(evidence.iterdir()):
        if path.name == '_quarantine':
            continue
        members = list(path.iterdir()) if path.name == 'barness-ai' and path.is_dir() else [path]
        for member in members:
            name = str(member.relative_to(root))
            kept = any(intersects(name, p) for p in protected)
            mark = 'KEEP' if kept else ('QUARANTINE' if successful(member) else 'BLOCKED')
            candidates.append({'path': name, 'mark': mark, 'reason': 'referenced' if kept else 'successful' if mark == 'QUARANTINE' else 'unknown or failed'})
    for path in (root / '.scratch').rglob('*.log') if (root / '.scratch').exists() else []:
        relative = str(path.relative_to(root))
        if '/evidence/' not in relative:
            continue
        kept = any(intersects(relative, p) for p in protected)
        # Log text is not proof of process success. Require its enclosing receipt.
        success = next((p for p in path.parents if (p / 'receipt.json').is_file()), None)
        ok = success is not None and successful(success)
        candidates.append({'path': relative, 'mark': 'KEEP' if kept else 'QUARANTINE' if ok else 'BLOCKED',
                           'reason': 'referenced' if kept else 'successful' if ok else 'no successful receipt'})
    return {'sources': sources, 'refs': references, 'missing': missing, 'unresolved': unresolved,
            'protected': sorted(protected), 'baseline': baseline, 'candidates': candidates}
