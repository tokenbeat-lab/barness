#!/usr/bin/env python3
"""Seal this reviewed Kimi catalog batch; originals and post-seal logs remain external."""
from pathlib import Path
import datetime,hashlib,json
root=Path(__file__).resolve().parents[4]
b=root/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6'
version='catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c'
def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
 return h.hexdigest()
def obj(path):return json.loads((b/path).read_text())
def ref(path):return {'path':path,'sha256':sha(b/path)}
assert not (b/'delivery.json').exists()
result=obj('verification/result.json');eq=obj('verification/post-doc-equality-v1.json');code=obj('code-and-distribution.json');review=obj('review/review.json');snapshot=b/(version+'.json')
assert result['status']=='PASS' and result['ordinary']['status']=='PASS' and result['ordinary']['exitCode']==0 and result['ordinary']['commit']=='97e3fef86d6948705eb70abc97976c64c725b442'
assert result['ordinary']['requiredCount']==203 and not result['ordinary']['requiredFailures'] and len(result['ordinary']['bundles'])==4 and all(x['auditFindings']==0 for x in result['ordinary']['bundles'])
assert all(x['exitCode']==0 for x in result['buildChecks']) and result['finalPlanIdentityCount']==125 and result['finalSnapshotSha256']==sha(snapshot)
assert eq['status']=='PASS' and eq['sourceSetAndHashesEqual'] and not eq['selectedInputMismatches'] and eq['sourceCount']==2166 and eq['selectedInputCount']==167
assert snapshot.read_bytes()==(b/'review/snapshot.json').read_bytes()==(root/'ai/catalog_builtin.json').read_bytes()
review_files={str(p.relative_to(b/'review')) for p in (b/'review').rglob('*') if p.is_file()};assert review_files==set(review['files'])|{'review.json'}
for p,h in review['files'].items():assert 'sha256:'+sha(b/'review'/p)==h
assert sha(b/'review/review.json')=='8a4a8758cffbcc360144465bde2d120aa6429f046aed612cbfbfb948886c887d'
assert code['comparisonBase']=='ceabec8b58c42377f4f643ccb87c1bc1d4beb456' and code['comparisonTarget']==result['ordinary']['commit']
for p in ['verification/reviews/spec-final-v2.md','verification/reviews/standards-v5.md','provenance/verification-final-v1/mapping.json','provenance/README.md','provenance/changelog-source-final.md','verification/current-status-before-seal.md']:
 assert (b/p).is_file()
files={str(p.relative_to(b)):sha(p) for p in sorted(b.rglob('*')) if p.is_file()}
marker={'schema':1,'status':'complete','createdAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'catalogVersion':version,'catalogHash':'sha256:93a1692f982836d9ad478252313b73e75eeee4904e6d79e3bcc915cd54a44005','maintenanceRevision':'sha256:7f0724d5ae599efb2216a22614c775cbce18de88b967b23f48694cd587f8b063','providerVersions':{'anthropic':1,'deepseek':1,'google':1,'minimax':1,'minimax-cn':1,'moonshot':4,'moonshot-cn':4,'openai':2,'typesafe':1,'xai':1,'zai':1,'zhipu':1},'snapshot':{'path':snapshot.name,'sha256':sha(snapshot),'identities':125},'review':ref('review/review.json'),'changelog':{'snapshot':ref('CHANGELOG.md'),'sourceOriginal':ref('provenance/changelog-source-final.md'),'readerNotice':ref('provenance/README.md')},'codeAndDistribution':ref('code-and-distribution.json'),'comparisonBase':code['comparisonBase'],'comparisonTarget':code['comparisonTarget'],'actualMeasuredCommit':result['ordinary']['commit'],'coreRuntimeCodeChanged':'YES','runtimeDependenciesChanged':'NO','exampleCodeChanged':'YES','oldProgramProbe':code['oldProgram'],'final125Check':code['finalSnapshotCheck'],'ordinary':{'receipt':ref('verification/ordinary-v3-receipt.json'),'originalReceipt':result['ordinary']['receipt'],'status':'PASS','exitCode':0,'argv':result['ordinary']['argv'],'elapsedSeconds':result['ordinary']['elapsedSeconds'],'requiredCount':203,'requiredFailures':[],'bundles':result['ordinary']['bundles'],'goSkipEvents':result['goSkipEventCount'],'nativeNotRunCount':result['nativeNotRunCount'],'skipInventory':ref('verification/skip-inventory.json')},'buildChecks':result['buildChecks'],'postDocEquality':ref('verification/post-doc-equality-v1.json'),'applicability':code['applicability'],'verificationResult':ref('verification/result.json'),'evidenceSelfContained':False,'externalEvidenceNote':'Original complete case capsules, command inputs and historical receipts remain at their recorded paths.','reviewReports':{'standards':ref('verification/reviews/standards-v5.md'),'spec':ref('verification/reviews/spec-final-v2.md')},'currentStatusSnapshot':ref('verification/current-status-before-seal.md'),'files':files}
with (b/'delivery.json').open('x') as f:json.dump(marker,f,ensure_ascii=False,indent=2);f.write('\n')
print(json.dumps({'filesExcludingIndex':len(files),'totalFiles':len(files)+1,'deliverySHA256':sha(b/'delivery.json')}))
