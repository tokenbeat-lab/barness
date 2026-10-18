#!/usr/bin/env python3
"""Read only closed protocol/visual actuals; whole release remains uncertified."""
import argparse,collections,json,posixpath,sys
from pathlib import Path
import readers,provenance,quality
BASE=Path(__file__).resolve().parent;SOURCE=provenance.SOURCE;ROUND=provenance.ROUND;NAMES=('protocol114','visual-jev','visual-decisions');CATALOG='sha256:3718603f6ad6cee8a5b66251bede572b62e58fd290b3d7fc7a49504a1d274e4f';PROXY='http://127.0.0.1:7890'

def usage_of(attempts):
 reporting=collections.Counter();known_input=known_output=0
 for a in attempts:
  reporting[a.get('usageReporting','unreported')]+=1;u=a.get('usage',{});values=[u.get(n,0) for n in ('input','cacheRead','cacheWrite','output')]
  if any(type(v) is not int or v<0 for v in values):raise ValueError('invalid actual usage')
  known_input+=sum(values[:3]);known_output+=values[3]
 return dict(usageReporting=dict(reporting),knownInputSide=known_input,knownOutput=known_output,unknownOrPartial=sum(v for k,v in reporting.items() if k!='complete'))

def references(b,pins,prefix,manifest):
 count=0
 for case in manifest['cases']:
  b.relative(case['dir'])
  for leaf,digest in case.get('artifacts',{}).items():
   b.relative(leaf)
   if pins.get(prefix+case['dir']+'/'+leaf,{}).get('sha256')!=digest:raise ValueError('native manifest artifact reference differs')
   count+=1
 return count

def protocol(b,m,pins,root,source):
 prefix=root+'/';doc=lambda name:b.doc(m[prefix+name]);r=doc('multimodal-report.json');seal=doc('multimodal-integrity.json');manifest=doc('manifest.json');ids=r['expected'];attempts=[];scenes=[]
 if r['outcome']!='PASS' or r['complete'] is not True or r['auditPassed'] is not True or r['mode']!='live' or r['gitCommit']!=SOURCE or r['catalogHash']!=CATALOG or len(ids)!=12 or r['scenarios']!=ids or len(set(ids))!=12:raise ValueError('fixed protocol12 native status/source differs')
 sealed={n.removeprefix(prefix):v['sha256'] for n,v in pins.items() if n.startswith(prefix) and n!=prefix+'multimodal-integrity.json'}
 if seal['files']!=sealed:raise ValueError('protocol whole native seal differs')
 mapping=r['sourceHashes'];proof=b.doc(b.read(BASE,'additional-protocol-source.json',1<<20))
 if proof['sourceCommit']!=SOURCE or proof['originalPath']!='.scratch/jev-1.14-multimodal-box/verify-11-evidence.py' or proof['decodedSHA256']!='0ad4b561cfe0407d8ee2943aa6543d4c56ea0f0e895e772ff111d4699a78f08e':raise ValueError('fixed supplemental protocol source differs')
 b.decoded_gzip(proof);selected=dict(source['files']);selected[proof['originalPath']]=proof['decodedSHA256']
 if len(mapping)!=862 or any(selected.get(posixpath.normpath('ai/live/'+n))!=digest for n,digest in mapping.items()):raise ValueError('protocol selected source differs')
 refs=references(b,pins,prefix,manifest)
 if any(c['status']!='PASS' for c in manifest['cases']) or b.doc(m[prefix+'audit.json'])['findings']:raise ValueError('protocol native audit/case failure')
 for identity in ids:
  case=prefix+'MM11-'+identity+'/';result=b.doc(m[case+'result.json']);wire=json.loads(m[case+'exchanges.json']);assertions=b.doc(m[case+'assertions.json']);actual=result['result']['metadata']['attempts']
  if result['error'] or len(actual)!=1 or len(wire)!=1 or wire[0]['status']!=200 or wire[0]['url']!='https://api.typesafe.ai/v1/systemone' or any(a['pass'] is not True for a in assertions['assertions']):raise ValueError('protocol actual call/Attempt/wire/assertion differs')
  attempts+=actual;scenes.append(dict(collection='protocol114',id=identity,status='PASS',calls=1,attempts=1,duration=b.doc(m[case+'duration.json'])))
 u=usage_of(attempts);budget=r['budget']
 if budget['callsUsed']!=12 or budget['httpAttempts']!=12 or budget['retries']!=0 or budget['reportedInputTokens']!=u['knownInputSide'] or budget['reportedOutputTokens']!=u['knownOutput'] or budget['limits']['calls']!=12 or budget['limits']['attempts']!=12:raise ValueError('protocol independent actual usage/budget differs')
 return dict(collection='protocol114',outcome='PASS',scenarios=12,statuses={'PASS':12},calls=12,attempts=12,selectedFiles=len(mapping),artifactReferences=refs,auditFindings=0,unknown=r['unknown'],budget=budget,**u),scenes

def visual(b,m,pins,root,source):
 prefix=root+'/';doc=lambda n:b.doc(m[prefix+n]);r=doc('report.json');route=root.removeprefix('actual/visual-');config=doc('run-config.json');samples=r['samples'];gold=doc('inputs/dataset.json')['samples'];experiment=doc('inputs/experiment.json');scenes=[];attempts=[]
 if config['proxy']!=PROXY or config['route']!=route or config['mode']!='live' or config['formal'] is not True or config['maxCalls']!=12:raise ValueError('explicit visual proxy/native mode differs')
 if r['route']!=route or r['status']!='COMPLETE' or r['realQuality']!='COMPLETE' or r['mode']!='live' or r['catalogKind']!='formal' or r['catalogHash']!=CATALOG or len(samples)!=12 or [s['id'] for s in samples]!=[s['id'] for s in gold] or len({s['id'] for s in samples})!=12:raise ValueError('visual fixed12 native identities/status differ')
 metrics=quality.metrics(r,gold,experiment)
 manifest=doc('manifest.json');sealed={n.removeprefix(prefix):v['sha256'] for n,v in pins.items() if n.startswith(prefix) and n!=prefix+'manifest.json'}
 if manifest['artifacts']!=sealed:raise ValueError('visual whole native manifest differs')
 mapping=doc('source-hashes.json')
 if len(mapping)!=323 or any(source['files'].get(n)!=digest for n,digest in mapping.items()):raise ValueError('visual selected323 source differs')
 if doc('audit.json')['findings']:raise ValueError('visual native audit differs')
 for s in samples:
  case=prefix+'samples/'+s['id']+'/';raw=doc('samples/'+s['id']+'/result.json');wire=doc('samples/'+s['id']+'/wire.json');actual=s['result']['metadata']['attempts']
  if s!=raw or s['error'] is not None or s['status']!='answered' or len(actual)!=1 or wire['status']!=200 or not all(s['assertions'].values()) or any(s['resources'].values()):raise ValueError('visual actual result/Attempt/wire/assertions differs')
  attempts+=actual;scenes.append(dict(collection='visual-'+route,id=s['id'],group=s['group'],status=s['status'],calls=1,attempts=1,iou=s['iou'],consumed=s['consumed'],view=s['view'],startedAt=s['startedAt'],finishedAt=s['finishedAt']))
 u=usage_of(attempts);budget=r['budget']
 if budget['calls']!=12 or budget['attempts']!=12 or budget['retries']!=0 or budget['inputTokens']!=u['knownInputSide'] or budget['outputTokens']!=u['knownOutput'] or budget['partial']!=u['usageReporting'].get('partial',0) or budget['unreported']!=u['usageReporting'].get('unreported',0):raise ValueError('visual independent usage/budget differs')
 return dict(collection='visual-'+route,outcome='PASS',nativeStatus='COMPLETE',samples=12,statuses={'answered':12},calls=12,attempts=12,selectedFiles=323,auditFindings=0,metrics=metrics,budget=budget,accountAttestation=config['authorization'],qualityScope='fixed twelve synthetic samples; no population or vendor invoice inference',**u),scenes

def analyse(b,index,m):
 pins=b.inventory(index['capsule']);source,plan,authority=provenance.authority(b,m);rows=[];scenes=[]
 for name in NAMES:
  root=index['roots'][name];expected='actual/'+name
  if root!=expected and (name!='protocol114' or not root.startswith(expected+'/') or len(root.split('/'))!=3):raise ValueError('fixed native root differs')
  if name=='protocol114':row,cases=protocol(b,m,pins,root,source);leaf='multimodal-report.json'
  else:row,cases=visual(b,m,pins,root,source);leaf='report.json'
  receipt=b.doc(m['canonical/'+name+'.json']);command=b.doc(m['frozen/'+name+'-command.json']);full_name=ROUND+'-'+name;key='OPENAI_KEY' if name=='visual-decisions' else 'TYPESAFE_KEY';injected='BARNESS_MM_KEY' if name=='protocol114' else 'BARNESS_VISUAL_KEY'
  if receipt['collectionRoot']!='$REPO/.evidence/minimax-release-repair-effects-1' or receipt['collectionOwnerSHA256']!=plan['collection']['rootSHA256'] or receipt['phase']!='effects' or receipt['baseName']!=name or receipt['sourceVariable']!=key or receipt['maxCalls']!=12 or receipt['maxAttempts']!=12:raise ValueError('canonical root/phase/base/source/caps differ')
  if receipt['sourceCommit']!=SOURCE or receipt['route']!=full_name or receipt['validationRound']!=ROUND or receipt['state']!='COMPLETED' or receipt['exitCode']!=0 or receipt['credentialsRead'] is not True or receipt['networkRoute']!='international-loopback-proxy' or receipt['automaticReruns'] is not False or receipt['launcherAudit']!={'passed':True,'findings':[]} or receipt['budget']['state']!='REPORTED' or receipt['budget']['status']!='PASS' or receipt['budget']['calls']!=12 or receipt['budget']['attempts']!=12:raise ValueError('actual canonical outcome/routing differs')
  if command['state']!='COMPLETED' or command['exitCode']!=0 or command['name']!=name or command['credentialsRead'] is not True or command['credentialsInherited'] is not False or command['sourceVariable']!=key or command['injectedCredential']!=injected or command['launcherAudit']!={'passed':True,'findings':[]} or command['argv']!=plan['routes'][full_name]['argv'] or command['cwd']!=plan['routes'][full_name]['cwd']:raise ValueError('actual selected numeric command differs')
  env=command['environmentKeys']
  if any('PROXY' in n.upper() for n in env if n!='BARNESS_AI_LIVE_PROXY') or (name=='protocol114')!=('BARNESS_AI_LIVE_PROXY' in env):raise ValueError('explicit native proxy environment differs')
  if name!='protocol114' and command['argv'][-2:]!=['-proxy',PROXY]:raise ValueError('explicit visual proxy argv differs')
  fields={'sourceManifestSha256':'source-at-start.json','binariesManifestSha256':'binaries.json','planSha256':'plan.json','authorizationSha256':'execution-authorization.json','sourceReviewSha256':'prerelease-source-review.json','historySHA256':'history.json','reuseSHA256':'reuse.json','scopeSHA256':'scope.json'}
  if any(receipt[k]!=b.sha(m['frozen/'+leaf]) for k,leaf in fields.items()) or receipt['reportedFileSha256']!=b.sha(m[root+'/'+leaf]) or receipt['commandReceiptSha256']!=b.sha(m['frozen/'+name+'-command.json']):raise ValueError('canonical original report/command/authority pins differ')
  rows.append(row);scenes+=cases
 closed=b.doc(m['frozen/actual-six-final-summary.json']);current=closed['current'];routes=closed['routes']
 if len(routes)!=6 or len({v['route'] for v in routes})!=6 or any(v['state']!='REPORTED' or v['calls']!=(22 if v['baseName'].startswith('glm-') else 12) or v['attempts']!=v['calls'] or v['status']!=('FAIL' if v['baseName']=='glm-messages' else 'PASS') for v in routes) or sum(v['calls'] for v in routes)!=102 or sum(v['attempts'] for v in routes)!=102 or current['calls']!=102 or current['attempts']!=102 or closed['total']['calls']!=647 or closed['total']['attempts']!=647 or closed['budget']!=provenance.BUDGET or closed['automaticProcessReruns'] is not False or b.doc(m['frozen/repair-consumption.json'])!=closed:raise ValueError('closed six actual102 and prior545 accounting differs')
 reporting=collections.Counter()
 for row in rows:reporting.update(row['usageReporting'])
 return dict(schema=1,sourceCommit=SOURCE,threeCollectionsOutcome='PASS',calls=36,attempts=36,currentSixCalls=102,currentSixAttempts=102,currentSixOutcomes={'PASS':5,'FAIL':1},priorCurrentReleaseCalls=545,priorCurrentReleaseAttempts=545,currentReleaseCalls=647,currentReleaseAttempts=647,rows=rows,usageReporting=dict(reporting),knownInputSide=sum(v['knownInputSide'] for v in rows),knownOutput=sum(v['knownOutput'] for v in rows),unknownOrPartial=sum(v['unknownOrPartial'] for v in rows),custody=authority,retentionCredentialsRead=False,retentionPaidCalls=0,automaticReruns=False,releaseCertified=False),scenes

def secondary(b,index,common,m):
 full=readers.full();out={};source=m['frozen/source-at-start.json'];binaries=b.doc(m['frozen/binaries.json']);standards=b.doc(m['frozen/independent-standards-review.json'])
 for label in ('preflight','binary'):
  cap=index[label]
  if cap['sourceCommit']!=SOURCE or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/effects-'+label+'-raw.tar.gz' or cap['archive']['prefix']!='repair-effects-'+label or cap['inventory']['file']!=label+'-inventory.json.gz' or cap['sourceSnapshot']!=index['capsule']['sourceSnapshot']:raise ValueError('fixed secondary capsule identity differs')
  pins,values,summary=full.verify(b,BASE,common,cap)
  if values['frozen/source-at-start.json']!=source:raise ValueError('secondary source differs')
  out[label]=summary
  if label=='binary':
   body=pins['frozen/visualeval'];build=b.doc(values['frozen/build-visualeval-command.json'])
   if body!=standards['binaries']['visualeval'] or body['sha256']!=binaries['visualeval'] or values['frozen/build-visualeval-command.json']!=m['frozen/build-visualeval-command.json'] or build['exitCode']!=0 or build['credentialsRead'] is not False or build['argv']!=['go','build','-o','$REPO/.evidence/minimax-release-repair-effects-1/visualeval','./ai/examples/visualeval']:raise ValueError('original visual binary/build differs')
   out['visualBinary']=body
  else:
   observed=b.doc(values['frozen/actual-keyless-case-summary.json']);manifests=[];refs=0
   for row in observed['manifests']:
    name='preflight/'+row['manifest'].removeprefix('.evidence/minimax-release-repair-effects-1/');v=b.doc(values[name]);prefix=name.removesuffix('manifest.json');statuses=dict(collections.Counter(c['status'] for c in v['cases']))
    if b.sha(values[name])!=row['sha256'] or v['git_commit']!=SOURCE or len(v['cases'])!=row['cases'] or statuses!=row['statuses'] or prefix+'audit.json' in values and b.doc(values[prefix+'audit.json'])['findings']:raise ValueError('actual keyless native inventory differs')
    refs+=references(b,pins,prefix,v);manifests.append(dict(member=name,cases=len(v['cases']),statuses=statuses,auditFindings=row['auditFindings']))
   if sorted(v['cases'] for v in manifests)!=[1,1,10,399] or next(v['statuses'] for v in manifests if v['cases']==399)!={'PASS':337,'NOT_RUN':62}:raise ValueError('actual399 keyless inventory differs')
   projection=b.doc(values['projection/known-intentional-link.json'])['links'];expected=index['intentionalLinkProjection'];link=projection[0] if len(projection)==1 else {}
   if expected!=dict(originalRelativePath='no-key-live/20261018T114438.245219000Z/formal-corrupt-symlink/formal-integrity.json',targetSHA256='53a6f961cad79b77ca8beaf2409530413dd04b7c38f01560f2a51eb7d3131782',archiveMember='projection/known-intentional-link.json',linksFollowed=False) or link.get('originalRelativePath')!=expected['originalRelativePath'] or link.get('type')!='symlink' or link.get('targetFollowed') is not False or b.sha(bytes.fromhex(link['targetHex']))!=expected['targetSHA256'] or link['targetSHA256']!=expected['targetSHA256']:raise ValueError('exact intentional link projection differs')
   out.update(manifests=manifests,artifactReferences=refs,intentionalLinkMetadataProjections=1,linksFollowed=False)
 shared=index['sharedLiveBinary'];gbase=BASE.parent/'glm';gi=b.doc(b.read(gbase,'index.json',1<<20))
 if shared['frontdoor']!='../glm' or shared['indexSHA256']!=b.sha(b.read(gbase,'index.json',1<<20)) or shared['capsule']!=gi['binary'] or shared['member']!='frozen/live.test' or shared['pin']!=standards['binaries']['live.test'] or shared['pin']['sha256']!=binaries['live.test']:raise ValueError('shared original binary reference differs')
 gp,gm,gs=full.verify(b,gbase,common,shared['capsule'])
 if gp[shared['member']]!=shared['pin'] or gm['frozen/source-at-start.json']!=source:raise ValueError('shared original binary body differs')
 out['sharedLiveBinary']=dict(pin=shared['pin'],capsule=gs,duplicateExecutableCopies=0);return out

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=parser.parse_args();b=readers.small()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20));cap=index['capsule']
 if index['actualSource']!=SOURCE or index['round']!=ROUND or index['boundedHelperSHA256']!=readers.SMALL or index['fullReaderSHA256']!=readers.FULL or cap['id']!='actualRepairEffects7931' or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/effects-actual-raw.tar.gz' or cap['archive']['prefix']!='repair-effects-actual' or set(index['roots'])!=set(NAMES):raise ValueError('fixed effects retention identity differs')
 pins=b.inventory(cap)
 if len(cap['observe'])>512 or len(cap['observe'])!=len(set(cap['observe'])) or not set(cap['observe'])<=set(pins) or sum(pins[n]['bytes'] for n in cap['observe'])>16<<20:raise ValueError('observed metadata bound differs')
 archive,m=b.verify_archive(a.common_root,cap);summary,scenes=analyse(b,index,m);more=secondary(b,index,a.common_root,m)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent summary differs')
 if scenes!=[json.loads(line) for line in b.read(BASE,'scenarios.jsonl',1<<20).decode().splitlines()]:raise ValueError('independent scenario inventory differs')
 if more!=b.doc(b.read(BASE,'secondary-summary.json',1<<20)):raise ValueError('independent secondary summary differs')
 for row in b.doc(b.read(BASE,'tdd/raw-log-index.json',1<<20))['logs']:b.decoded_gzip(row)
 print(json.dumps(dict(verified=True,sourceCommit=SOURCE,calls=36,attempts=36,currentSixCalls=102,currentReleaseCalls=647,threeCollectionsOutcome='PASS',releaseCertified=False,archive=archive,secondary=more,retentionCredentialsRead=False,retentionPaidCalls=0),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
