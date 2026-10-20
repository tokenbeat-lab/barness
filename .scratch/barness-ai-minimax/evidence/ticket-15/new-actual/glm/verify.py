#!/usr/bin/env python3
"""Read-only actual G custody/accounting; the original Messages failure stays failed."""
import argparse,collections,json,sys
from pathlib import Path
import readers,provenance
BASE=Path(__file__).resolve().parent;SOURCE=provenance.SOURCE;ROUND=provenance.ROUND

def analyse(b,index,m):
 pins=b.inventory(index['capsule']);source,plan,authority=provenance.authority(b,m);rows=[];scenes=[]
 for route,root in index['routes'].items():
  prefix=root+'/';report=b.doc(m[prefix+'formal-glm-report.json']);receipt=b.doc(m['canonical/'+route+'.json']);command=b.doc(m['frozen/glm-'+route+'-command.json']);audit=b.doc(m[prefix+'audit.json']);outcome='FAIL' if route=='messages' else 'PASS';name=ROUND+'-glm-'+route
  if receipt['collectionRoot']!='$REPO/.evidence/minimax-release-repair-glm-1' or receipt['collectionOwnerSHA256']!=plan['collection']['rootSHA256'] or receipt['baseName']!='glm-'+route or receipt['sourceVariable']!='GLM_KEY' or receipt['maxCalls']!=24 or receipt['maxAttempts']!=32:raise ValueError('canonical root/base/source/caps differ')
  if report['outcome']!=outcome or report['complete']!=(route!='messages') or report['auditPassed'] is not True or report['gitCommit']!=SOURCE or report['route']!=route or receipt['sourceCommit']!=SOURCE or receipt['route']!=name or receipt['phase']!='glm' or receipt['validationRound']!=ROUND or receipt['state']!='COMPLETED' or receipt['credentialsRead'] is not True or receipt['networkRoute']!='china-direct' or receipt['automaticReruns'] is not False or receipt['launcherAudit']!={'passed':True,'findings':[]} or audit['findings']!=[]:raise ValueError('own actual source/status/direct routing/audit differs')
  if command['state']!='COMPLETED' or command['exitCode']!=(1 if route=='messages' else 0) or command['name']!='glm-'+route or command['credentialsRead'] is not True or command['credentialsInherited'] is not False or command['sourceVariable']!='GLM_KEY' or command['injectedCredential']!='BARNESS_GLM_KEY' or command['launcherAudit']!={'passed':True,'findings':[]} or any('PROXY' in n.upper() for n in command['environmentKeys']) or command['argv']!=plan['routes'][name]['argv'] or command['cwd']!=plan['routes'][name]['cwd']:raise ValueError('actual selected numeric command differs')
  fields={'sourceManifestSha256':'source-at-start.json','binariesManifestSha256':'binaries.json','planSha256':'plan.json','authorizationSha256':'execution-authorization.json','sourceReviewSha256':'prerelease-source-review.json','historySHA256':'history.json','reuseSHA256':'reuse.json','scopeSHA256':'scope.json'}
  if any(receipt[k]!=b.sha(m['frozen/'+leaf]) for k,leaf in fields.items()) or receipt['reportedFileSha256']!=b.sha(m[prefix+'formal-glm-report.json']) or receipt['commandReceiptSha256']!=b.sha(m['frozen/glm-'+route+'-command.json']):raise ValueError('canonical original source/auth/command/report pins differ')
  seal=b.doc(m[prefix+'glm-integrity.json']);sealed={n.removeprefix(prefix):v['sha256'] for n,v in pins.items() if n.startswith(prefix) and n.removeprefix(prefix) not in ('glm-integrity.json','live-report.json')}
  if seal['schema']!=1 or seal['files']!=sealed:raise ValueError('original sealed member inventory differs')
  capture=receipt.get('capturePins')
  if route=='messages':
   if capture is not None or prefix+'live-report.json' in pins:raise ValueError('failed Messages cannot acquire success capturePins or a fabricated support report')
  else:
   expected={'schema':1,**{field:b.sha(m[prefix+leaf]) for field,leaf in [('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')]}}
   if capture!=expected:raise ValueError('actual successful capture pins differ')
  mapping=b.doc(m[prefix+'glm-execution-source.json'])
  if len(mapping)!=217 or any(source['files'].get(n)!=sha for n,sha in mapping.items()):raise ValueError('actual selected217 source differs')
  scenarios=report['scenarios'];ids=[s['id'] for s in scenarios]
  if len(ids)!=22 or len(set(ids))!=22 or set(ids)!=set(report['expected']):raise ValueError('complete fixed22 native scenarios required')
  calls=attempts=records=wire_count=known_input=known_output=0;http=collections.Counter();usage=collections.Counter();failures=[]
  for scene in scenarios:
   case=prefix+'GLM11-'+scene['id'].replace('/','-')+'/';result=b.doc(m[case+'attempt-1-result.json']);actual=result['result']['metadata']['attempts'];wire=json.loads(m[case+'attempt-1-exchanges.json']);calls+=scene['calls'];attempts+=scene['httpAttempts']
   if len(actual)!=scene['calls'] or len(wire)!=scene['httpAttempts']:raise ValueError('independent actual Attempt/wire count differs')
   for attempt in actual:
    records+=1;values=[attempt.get('usage',{}).get(k,0) for k in ('input','cacheRead','cacheWrite','output')]
    if any(type(v) is not int or v<0 for v in values):raise ValueError('nonnegative exact usage required')
    known_input+=sum(values[:3]);known_output+=values[3];usage[attempt.get('usageReporting','unreported')]+=1
   for exchange in wire:
    wire_count+=1;http[str(exchange.get('status','unreported'))]+=1
    if not exchange['url'].startswith('https://open.bigmodel.cn/'):raise ValueError('original China wire host differs')
   if scene['id'].endswith('/native-history') and scene['outcome']!='PASS':raise ValueError('native history capability differs')
   if scene['outcome']=='FAIL':
    failures.append(scene['id']);assertions=b.doc(m[case+'assertions.json']);message=result['result']['message'];metadata=result['result']['metadata']
    if route!='messages' or scene['id']!='glm-5.3/cancel' or scene['durationMs']<120000 or len(actual)!=1 or len(wire)!=1 or metadata['attempts'][0]['code']!='deadline_exceeded' or message['content'] is not None or 'status' in wire[0] or wire[0].get('transportError')!='context deadline exceeded' or not any(a['name']=='cancel after observable content' and a['pass'] is False and a['detail']=='classification=deadline_exceeded/request' for a in assertions['assertions']):raise ValueError('original failed120s cancel-before-content evidence differs')
   scenes.append(dict(route=route,id=scene['id'],outcome=scene['outcome'],calls=scene['calls'],attempts=scene['httpAttempts'],actualRecords=len(actual),durationMs=scene['durationMs']))
  budget=report['budget'];outcomes=dict(collections.Counter(s['outcome'] for s in scenarios));record_budget=receipt['budget']
  if (calls,attempts,records,wire_count)!=(22,22,22,22) or budget['callsUsed']!=calls or budget['httpAttempts']!=attempts or budget['plannedCalls']!=22 or budget['retries']!=0 or budget['maxCalls']!=24 or budget['maxAttempts']!=32 or budget['maxOutputTokens']!=4096 or budget['callSeconds']!=120 or budget['processSeconds']!=1800 or record_budget['state']!='REPORTED' or record_budget['calls']!=22 or record_budget['attempts']!=22 or record_budget['status']!=outcome or outcomes!=({'PASS':21,'FAIL':1} if route=='messages' else {'PASS':22}):raise ValueError('actual scenario/usage budget sums differ')
  rows.append(dict(route=route,outcome=outcome,complete=report['complete'],scenarios=22,outcomes=outcomes,calls=calls,attempts=attempts,actualAttemptRecords=records,wireRecords=wire_count,httpStatus=dict(http),usageReporting=dict(usage),knownInputSide=known_input,knownOutput=known_output,failures=failures,selectedFiles=217,integrityFiles=len(seal['files']),auditFindings=0,successfulCapturePins=capture is not None,nativeSupportReportPresent=prefix+'live-report.json' in pins,endpoint=report['endpoint'],api=report['api'],accountVerification=report['accountVerification'],regionVerification=report['regionVerification'],privilegeVerification=report['privilegeVerification']))
 for leaf in ('repair-consumption.json','actual-glm-final-summary.json'):
  v=b.doc(m['frozen/'+leaf]);remaining=[r for r in v['routes'] if not r['baseName'].startswith('glm-')]
  if v['current']['calls']!=66 or v['current']['attempts']!=66 or v['total']['calls']!=611 or v['total']['attempts']!=611 or len(remaining)!=3 or any(r['state']!='NOT_RUN' or r['calls']!=0 or r['attempts']!=0 for r in remaining):raise ValueError('phase-closed known66 and historical545 accounting differs')
 publication=provenance.publication(b,m,source,index['routes'],pins);usage=collections.Counter()
 for row in rows:usage.update(row['usageReporting'])
 return dict(schema=1,actualSource=SOURCE,collectionOutcome='FAIL',threeRoutePublication=False,missingNativeSupportReports=['messages'],calls=66,attempts=66,plannedScenes=66,outcomes={'PASS':65,'FAIL':1},priorCurrentReleaseCalls=545,priorCurrentReleaseAttempts=545,currentReleaseCallsAtGClose=611,currentReleaseAttemptsAtGClose=611,remainingThreeAtGClose='NOT_RUN0/0',automaticReruns=False,routes=rows,usageReporting=dict(usage),unknownOrPartialUsageAttempts=sum(v for k,v in usage.items() if k!='complete'),knownInputSide=sum(r['knownInputSide'] for r in rows),knownOutput=sum(r['knownOutput'] for r in rows),usageBasis='Independent final Attempt sums: input+cacheRead+cacheWrite once; output once; reasoning already output. Partial/unreported remains unknown; known lower bound, not invoice.',custody=authority,publication=publication,retentionCredentialsRead=False,retentionPaidCalls=0,releaseCertified=False),scenes

def secondary(b,index,common,m):
 for label in ('preflight','binary'):
  cap=index[label]
  if cap['sourceCommit']!=SOURCE or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/glm-'+label+'-raw.tar.gz' or cap['archive']['prefix']!='repair-glm-'+label or cap['inventory']['file']!=label+'-inventory.json.gz' or cap['sourceSnapshot']!=index['capsule']['sourceSnapshot']:raise ValueError('fixed secondary capsule identity differs')
 full=readers.full();ppins,pm,preflight=full.verify(b,BASE,common,index['preflight']);bpins,bm,binary=full.verify(b,BASE,common,index['binary']);source=m['frozen/source-at-start.json']
 if pm['frozen/source-at-start.json']!=source or bm['frozen/source-at-start.json']!=source or bm['frozen/binaries.json']!=m['frozen/binaries.json'] or bm['frozen/build-live.test-command.json']!=m['frozen/build-live.test-command.json']:raise ValueError('preflight/binary original source/build differs')
 binaries=b.doc(m['frozen/binaries.json']);build=b.doc(m['frozen/build-live.test-command.json']);body=bpins['frozen/live.test'];standards=b.doc(m['frozen/independent-standards-review.json'])
 if body['sha256']!=binaries['live.test'] or body!=standards['binaries']['live.test'] or build['state']!='COMPLETED' or build['exitCode']!=0 or build['credentialsRead'] is not False or build['argv']!=['go','test','-tags','live','-c','-o','$REPO/.evidence/minimax-release-repair-glm-1/live.test','./ai/live'] or standards['buildReceiptSHA256']!=b.sha(m['frozen/build-live.test-command.json']):raise ValueError('original compiled binary body/hash/receipt differs')
 observed=b.doc(pm['frozen/actual-keyless-case-summary.json']);manifests=[];refs=0
 for item in observed['manifests']:
  name='preflight/'+item['manifest'].removeprefix('.evidence/minimax-release-repair-glm-1/');v=b.doc(pm[name]);prefix=name.removesuffix('manifest.json');statuses=dict(collections.Counter(c['status'] for c in v['cases']))
  if b.sha(pm[name])!=item['sha256'] or v['git_commit']!=SOURCE or len(v['cases'])!=item['cases'] or statuses!=item['statuses']:raise ValueError('original actual keyless manifest differs')
  if prefix+'audit.json' in pm and b.doc(pm[prefix+'audit.json'])['findings']:raise ValueError('actual keyless audit findings')
  for case in v['cases']:
   for leaf,digest in case.get('artifacts',{}).items():
    b.relative(leaf);b.relative(case['dir'])
    if ppins.get(prefix+case['dir']+'/'+leaf,{}).get('sha256')!=digest:raise ValueError('actual keyless artifact reference differs')
    refs+=1
  manifests.append(dict(member=name,cases=len(v['cases']),statuses=statuses,auditFindings=item['auditFindings']))
 if sorted(v['cases'] for v in manifests)!=[1,1,10,399] or next(v['statuses'] for v in manifests if v['cases']==399)!={'PASS':337,'NOT_RUN':62}:raise ValueError('actual399 keyless inventory/NOT_RUN differs')
 projection=b.doc(pm['projection/known-intentional-link.json']);links=projection['links'];expected='no-key-live/20261018T110935.233553000Z/formal-corrupt-symlink/formal-integrity.json';target='18ebed3cd4eba35406b844c0abba746151949978df4204c4c5b0578b6eff235c'
 if len(links)!=1 or links[0]['originalRelativePath']!=expected or links[0]['type']!='symlink' or links[0]['targetSHA256']!=target or b.sha(bytes.fromhex(links[0]['targetHex']))!=target or links[0]['targetFollowed'] is not False or index['intentionalLinkProjection']!=dict(originalRelativePath=expected,targetSHA256=target,archiveMember='projection/known-intentional-link.json',linksFollowed=False):raise ValueError('exact intentional link metadata projection differs')
 return dict(preflight=preflight,binary=binary,binaryBody=body,manifests=manifests,artifactReferences=refs,intentionalLinkMetadataProjections=1,linksFollowed=False)

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=parser.parse_args();b=readers.small()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20));cap=index['capsule']
 if index['actualSource']!=SOURCE or index['round']!=ROUND or index['boundedHelperSHA256']!=readers.SMALL or index['fullReaderSHA256']!=readers.FULL or cap['id']!='actualRepairGLM7931' or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/glm-raw.tar.gz' or cap['archive']['prefix']!='repair-glm-actual' or set(index['routes'])!={'chat','responses','messages'}:raise ValueError('fixed own G retention identity differs')
 pins=b.inventory(cap)
 if len(cap['observe'])>512 or len(cap['observe'])!=len(set(cap['observe'])) or not set(cap['observe'])<=set(pins) or sum(pins[n]['bytes'] for n in cap['observe'])>16<<20:raise ValueError('observed metadata bound differs')
 archive,m=b.verify_archive(a.common_root,cap);summary,scenes=analyse(b,index,m)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent summary differs')
 if scenes!=[json.loads(line) for line in b.read(BASE,'scenarios.jsonl',1<<20).decode().splitlines()]:raise ValueError('independent scenario inventory differs')
 more=secondary(b,index,a.common_root,m)
 if more!=b.doc(b.read(BASE,'secondary-summary.json',1<<20)):raise ValueError('independent secondary summary differs')
 logs=b.doc(b.read(BASE,'tdd/raw-log-index.json',1<<20))['logs']
 if len(logs)!=4 or len({v['originalFile'] for v in logs})!=4:raise ValueError('fixed lossless raw log inventory differs')
 for row in logs:b.decoded_gzip(row)
 negative=b.doc(b.read(BASE,'negative-results.json',1<<20));guards=negative.get('negativeGuards',[])
 if negative.get('verified') is not True or negative.get('originalArchivesMutated') is not False or len(guards)!=19 or len({v['case'] for v in guards})!=19 or any(v.get('exitCode')!=1 or not v.get('reason','').startswith('REFUSED:') for v in guards):raise ValueError('actual public negative guard evidence differs')
 print(json.dumps(dict(verified=True,actualSource=SOURCE,collectionOutcome='FAIL',calls=66,attempts=66,threeRoutePublication=False,archive=archive,secondary=more,retentionCredentialsRead=False,retentionPaidCalls=0,releaseCertified=False),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
