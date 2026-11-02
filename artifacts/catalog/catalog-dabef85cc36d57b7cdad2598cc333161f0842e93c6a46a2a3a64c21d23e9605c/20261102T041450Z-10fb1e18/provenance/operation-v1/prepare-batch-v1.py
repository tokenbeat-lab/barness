#!/usr/bin/env python3
"""Copy the currently adopted approved catalog into a fresh, incomplete batch."""
from pathlib import Path
import hashlib,json,shutil
ROOT=Path(__file__).resolve().parents[4];WORK=Path(__file__).resolve().parent
VERSION='catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c'
OLD=ROOT/'artifacts/catalog'/VERSION/'20261102T023400Z-c9d70d4f-process-6'
NEW=ROOT/'artifacts/catalog'/VERSION/'20261102T041450Z-10fb1e18'
def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
assert sha(OLD/'delivery.json')=='93e0bba370ee529108f800c40f32cb1b270596d8fafb3c7d13ab8b6ed171b6c1'
raw=(OLD/(VERSION+'.json')).read_bytes();assert hashlib.sha256(raw).hexdigest()=='6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a'
assert raw==(ROOT/'ai/catalog_builtin.json').read_bytes()
review=OLD/'review';base=ROOT/'catalog/data/baseline';expected={str(p.relative_to(review)) for p in review.rglob('*') if p.is_file()}
assert expected=={str(p.relative_to(base)) for p in base.rglob('*') if p.is_file()}
assert all((review/p).read_bytes()==(base/p).read_bytes() for p in expected)
manifest=json.loads((review/'review.json').read_text());assert expected==set(manifest['files'])|{'review.json'}
assert all('sha256:'+sha(review/p)==v for p,v in manifest['files'].items())
NEW.mkdir()
with (NEW/(VERSION+'.json')).open('xb') as f:f.write(raw)
shutil.copytree(review,NEW/'review')
assert all((NEW/'review'/p).read_bytes()==(review/p).read_bytes() for p in expected)
result={'status':'INCOMPLETE_NO_MARKER','sourceBatch':str(OLD.relative_to(ROOT)),'sourceMarkerSHA256':sha(OLD/'delivery.json'),'snapshotSHA256':sha(NEW/(VERSION+'.json')),'reviewFiles':len(expected),'reviewSHA256':sha(NEW/'review/review.json'),'note':'Exact approved bytes copied; no new re-export, compatibility conclusion or verification stage is claimed.'}
with (WORK/'prepare-batch-result-v1.json').open('x') as f:json.dump(result,f,indent=2);f.write('\n')
print(json.dumps(result))
