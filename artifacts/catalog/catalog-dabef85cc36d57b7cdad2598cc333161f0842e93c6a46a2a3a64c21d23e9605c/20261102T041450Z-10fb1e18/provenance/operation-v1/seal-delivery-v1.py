#!/usr/bin/env python3
"""Seal this reviewed policy batch only after explicit approval; execution is not approval."""
from pathlib import Path
import argparse,datetime,hashlib,json
ROOT=Path(__file__).resolve().parents[4]
B=ROOT/'artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T041450Z-10fb1e18'
VERSION='catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c'
def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for c in iter(lambda:f.read(1024*1024),b''):h.update(c)
 return h.hexdigest()
def obj(p):return json.loads((B/p).read_text())
def ref(p):return {'path':p,'sha256':sha(B/p)}
a=argparse.ArgumentParser();a.add_argument('--spec-report',required=True);a.add_argument('--seal',action='store_true');args=a.parse_args()
if not args.seal:raise SystemExit('Explicit --seal is required after approval')
spec=Path(args.spec_report)
if spec.is_absolute() or '..' in spec.parts or not str(spec).startswith('verification/reviews/') or not (B/spec).is_file():raise SystemExit('Spec review must be a retained in-batch report')
assert not (B/'delivery.json').exists()
r=obj('verification/result-v1.json');cov=obj('verification/fixture-coverage-v1.json');eq=obj('verification/post-docs-equality-v1.json');code=obj('code-and-distribution.json');review=obj('review/review.json');snap=B/(VERSION+'.json')
assert r['status']=='PASS' and r['commit']=='8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11' and r['ordinary']['status']=='PASS' and r['ordinary']['exitCode']==0 and r['ordinary']['requiredCount']==329
assert r['ordinary']['native']=={'PASS':17232,'NOT_RUN':226} and len(r['ordinary']['bundles'])==4 and all(x['exitCode']==0 for x in r['build'])
assert cov['status']=='PASS' and not cov['failures'] and eq['status']=='PASS' and not eq['sourceDifferences'] and not eq['selectedInputMismatches'] and eq['sourceCount']==2171 and eq['selectedInputCount']==84
assert sha(snap)=='6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a' and snap.read_bytes()==(B/'review/snapshot.json').read_bytes()==(ROOT/'ai/catalog_builtin.json').read_bytes()
paths={str(p.relative_to(B/'review')) for p in (B/'review').rglob('*') if p.is_file()};assert paths==set(review['files'])|{'review.json'} and all('sha256:'+sha(B/'review'/p)==v for p,v in review['files'].items())
assert sha(B/'review/review.json')=='8a4a8758cffbcc360144465bde2d120aa6429f046aed612cbfbfb948886c887d'
assert code['comparisonBase']=='9584d3d6f095aefa4ab73bdc75c39a705b7931d6' and code['comparisonTarget']==r['commit']
assert sha(B/'provenance/changelog-source-final.md')==code['changelogSource']['sha256'] and sha(B/'verification/current-status-before-seal.md')==code['currentStatusSnapshot']['sha256']
files={str(p.relative_to(B)):sha(p) for p in sorted(B.rglob('*')) if p.is_file()}
m={'schema':1,'status':'complete','createdAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'catalogVersion':VERSION,'catalogHash':'sha256:93a1692f982836d9ad478252313b73e75eeee4904e6d79e3bcc915cd54a44005','maintenanceRevision':'sha256:7f0724d5ae599efb2216a22614c775cbce18de88b967b23f48694cd587f8b063','providerVersions':{'anthropic':1,'deepseek':1,'google':1,'minimax':1,'minimax-cn':1,'moonshot':4,'moonshot-cn':4,'openai':2,'typesafe':1,'xai':1,'zai':1,'zhipu':1},'snapshot':{'path':snap.name,'sha256':sha(snap),'identityCount':125},'review':ref('review/review.json'),'reviewMeaning':'Copied approved baseline; its historical checks are not policy verification','comparisonBase':code['comparisonBase'],'comparisonTarget':code['comparisonTarget'],'actualMeasuredCommit':r['commit'],'coreRuntimeCodeChanged':'YES','coreRuntimeFiles':code['coreRuntimeFiles'],'runtimeDependenciesChanged':'NO','modelSchemaChanged':False,'catalogDataChanged':False,'codeAndDistribution':ref('code-and-distribution.json'),'changelog':{'snapshot':ref('CHANGELOG.md'),'sourceOriginal':ref('provenance/changelog-source-final.md'),'readerNotice':ref('provenance/README.md')},'ordinary':r['ordinary'],'ordinaryReceiptCopy':ref('verification/ordinary-v6-receipt.json'),'inspection':ref('verification/ordinary-inspection-v6c.json'),'fixtureCoverage':ref('verification/fixture-coverage-v1.json'),'skipInventory':ref('verification/skip-index-v1.json'),'buildChecks':r['build'],'postDocsEquality':ref('verification/post-docs-equality-v1.json'),'verificationResult':ref('verification/result-v1.json'),'applicability':code['applicability'],'evidenceSelfContained':False,'externalEvidenceNote':'Complete native capsules, historical archive and original inputs remain at recorded external paths.','reviewReports':{'standards':ref('verification/reviews/review-standards-v4.md'),'spec':ref(str(spec))},'currentStatusSnapshot':ref('verification/current-status-before-seal.md'),'files':files}
with (B/'delivery.json').open('x') as f:json.dump(m,f,ensure_ascii=False,indent=2);f.write('\n')
print(json.dumps({'filesExcludingIndex':len(files),'totalFiles':len(files)+1,'deliverySHA256':sha(B/'delivery.json')}))
