#!/usr/bin/env python3
"""Copy bounded source originals and record immutable command references; not a delivery seal."""
import hashlib, json, pathlib, shutil
ROOT = pathlib.Path(__file__).resolve().parents[4]
WORK = pathlib.Path(__file__).resolve().parent
BATCH = ROOT / 'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6'

def digest(p):
    h = hashlib.sha256()
    with p.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''): h.update(chunk)
    return h.hexdigest()

def copy(src, dst):
    dst.parent.mkdir(parents=True, exist_ok=True)
    with src.open('rb') as a, dst.open('xb') as b: shutil.copyfileobj(a, b)
    assert digest(src) == digest(dst)
    return {'source': str(src.relative_to(ROOT)) if src.is_relative_to(ROOT) else str(src), 'destination': str(dst.relative_to(BATCH)), 'sha256': digest(dst)}

assert BATCH.is_dir() and not (BATCH / 'delivery.json').exists()
records = []
for folder in ('inputs', 'definitions-v1'):
    for src in sorted((WORK / folder).rglob('*')):
        if src.is_file(): records.append(copy(src, BATCH / 'provenance' / src.relative_to(WORK)))
for name in ('operation.json', 'research-cn.md', 'candidate-index-v1.json', 'applied-index-v1.json', 'audit-final-v1.json', 'capture-int-v1.py', 'operate-v1.py', 'prepare-six-v1.py', 'preview-five-v1.py', 'apply-six-v1.py', 'package-prepare-v1.py', 'package-metadata-draft-v1.md', 'package-sources-v1.py'):
    src = WORK / name
    records.append(copy(src, BATCH / 'provenance' / 'operations' / name))
spec = pathlib.Path('/tmp/barness-kimi-k31-spec-independent-v1.md')
records.append(copy(spec, BATCH / 'verification/reviews/spec-independent-v1.md'))
references = []
for folder in sorted((WORK / 'commands').iterdir()):
    if not folder.is_dir(): continue
    for name in ('command.json', 'stdout.jsonl', 'stderr.log'):
        src = folder / name
        if src.is_file(): references.append({'path': str(src.relative_to(ROOT)), 'sha256': digest(src)})
for src in sorted((WORK / 'bin').rglob('*')):
    if src.is_file(): references.append({'path': str(src.relative_to(ROOT)), 'sha256': digest(src), 'kind':'binary'})
index = {'status':'PREPARED_NOT_COMPLETE','copies':records,'originalReferences':references,'note':'Original commands and binaries remain at their paths. Test capsules are not included. No delivery marker was written.'}
p = BATCH / 'provenance' / 'source-copy-index-v1.json'
with p.open('x') as f: json.dump(index, f, indent=2); f.write('\n')
print(json.dumps({'copied':len(records), 'referenced':len(references), 'index':str(p.relative_to(ROOT)), 'sha256':digest(p)}))
