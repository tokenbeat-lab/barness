#!/usr/bin/env python3
"""Adopt the completed approved review as the source and builtin seed."""
import hashlib,json,pathlib,shutil,sys
ROOT=pathlib.Path.cwd(); W=ROOT/'artifacts/catalog/work/20261102T023400Z-c9d70d4f'; R=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6/review'
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
index=R/'review.json'; assert sha(index)=='8a4a8758cffbcc360144465bde2d120aa6429f046aed612cbfbfb948886c887d'; manifest=json.loads(index.read_bytes());assert manifest['status']=='exported' and manifest['revision']=='sha256:7f0724d5ae599efb2216a22614c775cbce18de88b967b23f48694cd587f8b063'
files={str(p.relative_to(R)) for p in R.rglob('*') if p.is_file()};assert files==set(manifest['files'])|{'review.json'} and len(files)==34
for name,digest in manifest['files'].items(): assert 'sha256:'+sha(R/name)==digest,name
snapshot=R/'snapshot.json';assert sha(snapshot)=='6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a'
old=json.loads((ROOT/'ai/catalog_builtin.json').read_bytes())['catalog']; new=json.loads(snapshot.read_bytes())['catalog']
assert new['version']=='catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c'
for k in ['models','imageModels','classifierModels']:
 key=lambda x:(x['provider'],x['api'],x['id']); ids={key(x) for x in old[k]}; assert [x for x in new[k] if key(x) in ids]==old[k]
assert [len(new[k]) for k in ['models','imageModels','classifierModels']]==[119,3,3]
providers=ROOT/'catalog/data/providers'; source=R/'providers';assert {p.name for p in source.iterdir()}=={p.name for p in providers.iterdir() if p.suffix=='.json'}
baseline=ROOT/'catalog/data/baseline'; assert baseline.is_dir()
# Replace rather than merge so inherited blobs cannot enter the new review.
shutil.rmtree(baseline);shutil.copytree(R,baseline)
for p in source.glob('*.json'): shutil.copyfile(p,providers/p.name)
shutil.copyfile(snapshot,ROOT/'ai/catalog_builtin.json')
assert {str(p.relative_to(baseline)) for p in baseline.rglob('*') if p.is_file()}==files
for name in files: assert (baseline/name).read_bytes()==(R/name).read_bytes(),name
for p in source.glob('*.json'): assert (providers/p.name).read_bytes()==p.read_bytes(),p.name
assert (ROOT/'ai/catalog_builtin.json').read_bytes()==snapshot.read_bytes()
print(json.dumps({'status':'PASS','revision':manifest['revision'],'catalogVersion':new['version'],'snapshotSHA256':sha(snapshot),'reviewSHA256':sha(index),'reviewFiles':len(files),'providers':len(list(source.glob('*.json'))),'old119DefinitionsAndOrder':'identical'},ensure_ascii=False))
