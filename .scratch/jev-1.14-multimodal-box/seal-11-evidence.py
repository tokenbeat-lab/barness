#!/usr/bin/env python3
"""Recheck and seal the completed ticket-11 runs from the repository root.

Only local artifacts are read. Reconstructed child packages retain original
bytes/timestamps and refuse differing existing files. The four legacy failure
fixtures record pre-redaction input hashes; final artifact hashes seal their
redacted copies independently. New runs must retain new IDs, never edit these.
"""
import collections, hashlib, json, pathlib, subprocess

def require(value, message='evidence mismatch'):
    if not value:
        raise ValueError(message)
ROOT = pathlib.Path.cwd()
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()

def safe_path(root, name):
    relative = pathlib.PurePosixPath(name)
    require(not relative.is_absolute() and '..' not in relative.parts, 'evidence path escapes root')
    target = root.joinpath(*relative.parts)
    require(target.resolve().is_relative_to(root.resolve()), 'evidence path escapes root')
    for node in (target, *target.parents):
        require(not node.is_symlink(), 'symlink in evidence path')
        if node == ROOT:
            break
    return target

def inventory(root):
    m = safe_path(root, 'manifest.json')
    d = json.loads(m.read_bytes())
    require(json.loads(safe_path(root, 'audit.json').read_bytes()) == {'findings': []}, 'evidence mismatch')
    n = 0
    pre_redacted = 0
    original_sources = {}
    source_index = {sha(safe_path(ROOT, str(p.relative_to(ROOT)))): str(p.relative_to(ROOT)) for p in (ROOT / 'ai/e2e/testdata').rglob('*') if p.is_file()}
    for c in d['cases']:
        require(c['status'] != 'FAIL', c['id'])
        for file, h in c['artifacts'].items():
            require(sha(safe_path(root, c['dir']+'/'+file)) == h, 'evidence mismatch')
            n += 1
        for file, h in c['fixtures'].items():
            if sha(safe_path(root, c['dir']+'/'+file)) != h:
                require(h in source_index, 'evidence mismatch')
                original_sources[source_index[h]] = h
                pre_redacted += 1
    return dict(path=str(root.relative_to(ROOT)), manifestSHA256=sha(m), auditSHA256=sha(root / 'audit.json'), artifacts=n, preRedactionFixtureInstances=pre_redacted, originalFixtureSources=original_sources, cases=dict(collections.Counter((c['status'] for c in d['cases']))), ticket11Cases=sum((c['id'].startswith('MM11-') for c in d['cases'])))

def restore(root, route):
    source = root / ('MM11-process-' + route + '-success')
    raw = safe_path(source, 'multimodal-integrity.json').read_bytes()
    d = json.loads(raw)
    dest = ROOT / '.evidence/ticket11-sealed' / root.name / route
    dest.mkdir(parents=True, exist_ok=True)
    for file, h in d['files'].items():
        p = safe_path(dest, file)
        data = safe_path(source, file.replace('/', '__')).read_bytes()
        require(hashlib.sha256(data).hexdigest() == h, 'evidence mismatch')
        p.parent.mkdir(parents=True, exist_ok=True)
        if p.exists():
            require(p.read_bytes() == data, 'evidence mismatch')
        else:
            p.write_bytes(data)
    p = safe_path(dest, 'multimodal-integrity.json')
    if p.exists():
        require(p.read_bytes() == raw, 'evidence mismatch')
    else:
        p.write_bytes(raw)
    verifier = ROOT / '.scratch/jev-1.14-multimodal-box/verify-11-evidence.py'
    output = subprocess.check_output(['python3', str(verifier), str(dest)], text=True).strip()
    rejected = subprocess.run(['python3', str(verifier), '--require-live', str(dest)], capture_output=True, text=True)
    require(rejected.returncode == 1, 'evidence mismatch')
    report = json.loads(safe_path(dest, 'multimodal-report.json').read_bytes())
    return dict(path=str(dest.relative_to(ROOT)), outcome=report['outcome'], mode=report['mode'], catalogHash=report['catalogHash'], reportSHA256=sha(dest / 'multimodal-report.json'), integritySHA256=sha(p), verification=output, liveGate='REJECTED', budget=report['budget'])
receipt = dict(schema=1, kind='ticket-11-controlled-and-NOT_RUN-only', codeCommit='59e2dbc04b7a5aef523a05709f49a9668b35a8d5', formalCatalogUpdated=False, visualQuality='NOT_RUN', runs=[], controlled=[], live=[])
receipt['runs'].append(inventory(ROOT / '.evidence/barness-ai/20261016T115408.429237000Z'))
for name in ['20261016T115409.762160000Z', '20261016T115411.218521000Z']:
    root = ROOT / '.evidence/barness-ai' / name
    receipt['runs'].append(inventory(root))
    for route in ['jev', 'decisions']:
        receipt['controlled'].append(restore(root, route))
for name in ['20261016T115446.169508000Z', '20261016T115446.169967000Z']:
    root = ROOT / '.evidence/ticket11-live-not-run' / name
    r = json.loads(safe_path(root, 'multimodal-report.json').read_bytes())
    require(r['outcome'] == 'NOT_RUN' and (not r['complete']) and (r['mode'] == 'live'), 'evidence mismatch')
    require(r['budget']['callsUsed'] == r['budget']['httpAttempts'] == 0, 'evidence mismatch')
    d = json.loads(safe_path(root, 'multimodal-integrity.json').read_bytes())
    require(all((sha(safe_path(root, f)) == h for f, h in d['files'].items())), 'evidence mismatch')
    require(json.loads(safe_path(root, 'audit.json').read_bytes()) == {'findings': []}, 'evidence mismatch')
    receipt['live'].append(dict(path=str(root.relative_to(ROOT)), route=r['route'], outcome=r['outcome'], reason=r['failure'], reportSHA256=sha(root / 'multimodal-report.json'), integritySHA256=sha(root / 'multimodal-integrity.json'), catalogHash=r['catalogHash'], calls=0, httpAttempts=0))
receipt['sourceHashes'] = {p: sha(safe_path(ROOT, p)) for p in subprocess.check_output(['git', 'diff', '--name-only', '80e8d36...59e2dbc'], text=True).splitlines() if p.endswith(('.go', '.py', '.json'))}
receipt['sourceHashes']['.scratch/jev-1.14-multimodal-box/seal-11-evidence.py'] = sha(ROOT / '.scratch/jev-1.14-multimodal-box/seal-11-evidence.py')
receipt['redAssertions'] = {p: sha(ROOT / p) for p in ['.evidence/ticket11-first-red/ticket11-red/20261016T114005.681943000Z/MM11-process-jev-success/assertions.json', '.evidence/barness-ai/20261016T114522.551049000Z/MM11-tamper-jev-legend-index/assertions.json', '.evidence/barness-ai/20261016T115233.532225000Z/MM11-process-jev-unfiltered/assertions.json', '.evidence/barness-ai/20261016T115247.834151000Z/MM11-tamper-jev-source-seal/assertions.json']}
output_path = safe_path(ROOT, '.scratch/jev-1.14-multimodal-box/11-evidence.json')
encoded = json.dumps(receipt, indent=2) + '\n'
if output_path.exists():
    require(output_path.read_text() == encoded, 'existing receipt differs; preserve it and use a new run index')
else:
    output_path.write_text(encoded)
print(json.dumps(dict(runs=receipt['runs'], controlled=[r['verification'] for r in receipt['controlled']], live=receipt['live']), indent=2))
