#!/usr/bin/env python3
"""Verify complete original offline history; no captured code is executed."""
import argparse,collections,json,sys
from pathlib import Path
import contract as c,readers
BASE=Path(__file__).resolve().parent
RUNNER='17dd303a1ba28882f3af5a8f83661fe98d856232e38e19a31d68fc10e6534032'
def analyse(b,index,pins,m):
 source=b.doc(m['final/source-proof.json']);roots=index['roots'];expected_roots={name for name,(_,statuses) in c.STAGES.items() if statuses}
 if set(roots)!=expected_roots or source['sourceCommit']!=c.SOURCE or source['measuredHEAD']!='ca67bedb001217a4b585e798b0feacfa3ee02c1a' or source['allMeasuredCompilerAndJSONSourceByteIdenticalToCommit'] is not True or source['formalApplicabilityRenewed'] is not False:raise ValueError('fixed original measured source or stage roots differ')
 if {name.split('/')[1] for name in pins if name.startswith('stages/')}!=set(c.STAGES)|{'dependency-restore','dependency-restore-2'} or pins['controllers/runner.py']['sha256']!=RUNNER:raise ValueError('whole original stage inventory differs')
 allcases={};stages={};references=0;fixture_inputs=0;fixture_provenance=[];observed={'final/source-proof.json','final/tool-fixture-repair-source-proof.json','final/tool-guard-reapplied-proof.json','stages/dependency-restore-2/proof.json'}
 for label,(code,statuses) in c.STAGES.items():
  command='stages/'+label+'/command.json';observed.add(command);row=b.doc(m[command]);prefix='stages/'+label+'/'
  if row['state']!='COMPLETED' or type(row['exitCode']) is not int or row['exitCode']!=code or row['credentialsRead'] is not False or row['realProviderRequestsAuthorized'] is not False or row['sourceUnchangedAfter'] is not True or row['changedFiles'] or row['runnerSHA256']!=RUNNER or row['sourceSnapshotSHA256']!=pins[prefix+'source-at-start.json']['sha256'] or row['logSHA256']!=pins[prefix+'command.log']['sha256'] or any('KEY' in key or 'PROXY' in key for key in row['environmentKeys']):raise ValueError('original numeric command/source/log disclosure differs: '+label)
  if label in c.FINAL:
   for name,digest in source['files'].items():
    if pins[prefix+'source/'+name]['sha256']!=digest:raise ValueError('measured final WIP is not committed source: '+label)
  stages[label]=dict(actualExitCode=code,seconds=row['seconds'],statuses=statuses)
  if not statuses:continue
  root=roots[label];parts=b.relative(root)
  if len(parts)!=4 or parts[:3]!=['stages',label,'bundles']:raise ValueError('original owned bundle root differs')
  observed.update({root+'/manifest.json',root+'/audit.json'});manifest=b.doc(m[root+'/manifest.json']);cases=manifest['cases'];audit=b.doc(m[root+'/audit.json'])
  if dict(collections.Counter(row['status'] for row in cases))!=statuses or len({row['id'] for row in cases})!=len(cases) or audit['findings']:raise ValueError('original cases/statuses/audit differ: '+label)
  allcases[label]=cases;stages[label].update(cases=len(cases),startedAt=manifest['started_at'],finishedAt=manifest['finished_at'])
  for row in cases:
   directory='/'.join(b.relative(row['dir'].rstrip('/')))
   for table in ('artifacts','fixtures'):
    for leaf,digest in row.get(table,{}).items():
     path=root+'/'+directory+'/'+'/'.join(b.relative(leaf))
     if table=='artifacts':
      if pins[path]['sha256']!=digest:raise ValueError('original manifest artifact reference differs')
      references+=1
     else:
      fixture_inputs+=1
      if pins[path]['sha256']!=digest:
       # Fixture hashes the raw input before writeFile's existing secret-alias
       # redaction. artifacts hashes saved bytes. Keep both original meanings.
       matches=sorted(n for n,pin in pins.items() if n.startswith(prefix+'source/') and pin['sha256']==digest)
       if not matches:raise ValueError('raw fixture input lacks its original same-stage source')
       fixture_provenance.append(dict(stage=label,case=row['id'],table='fixtures',leaf=leaf,declared=digest,actual=pins[path]['sha256'],declaredRawInputSourceMembers=matches))
   if label in c.TRUE_RED or label=='green-final-native-shapes':observed.update(root+'/'+directory+'/'+leaf for leaf in ('requests.json','result.json'))
  if 'NOT_RUN' in statuses:
   missing={row['id'] for row in cases if row['status']=='NOT_RUN'}
   if missing!=c.NOT_RUN or any(('BARNESS_AI_PRESSURE=1' if row['id'].startswith('MM05-design-load-') else 'BARNESS_AI_PIDIFF=1') not in row['replay'] for row in cases if row['status']=='NOT_RUN'):raise ValueError('exact mandatory not-run disclosure differs')
  if label.startswith('selected-opt-in-'):
   argv=b.doc(m[command])['argv']
   if {row['id'] for row in cases}!=c.NOT_RUN or '/usr/bin/env BARNESS_AI_PRESSURE=1 BARNESS_AI_PIDIFF=1' not in argv:raise ValueError('mandatory opt-in execution differs')
 if set(index['observe'])!=observed:raise ValueError('complete finite observation selection differs')
 true_red=0;resources=[];positives=0;green_rejected=0;null_zero=0
 for label in (*c.TRUE_RED,'green-final-native-shapes'):
  root=roots[label];failed=0
  for row in allcases[label]:
   prefix=root+'/'+row['dir']+'/';requests=json.loads(m[prefix+'requests.json']);result=b.doc(m[prefix+'result.json']);attempts=result['metadata'].get('attempts',[]);positive=row['id'].startswith('MM04-final-business-hook-');count=1 if label in c.TRUE_RED or positive else 0
   if requests is None and count==0:
    # Go records a nil empty request slice as JSON null; preserve this fact.
    null_zero+=1;requests=[]
   if not isinstance(requests,list) or not isinstance(attempts,list) or len(requests)!=count or len(attempts)!=count:raise ValueError('actual Provider/Attempt evidence differs')
   resources.append(prefix+'resources.json')
   if label in c.TRUE_RED:
    if row['status']=='FAIL':failed+=1
    elif not positive or row['status']!='PASS':raise ValueError('wrong RED positive control')
   elif positive:positives+=1
   else:green_rejected+=1
  if label in c.TRUE_RED:
   if failed!=c.TRUE_RED[label]:raise ValueError('exact genuine malformed-send RED count differs')
   true_red+=failed
 if (true_red,green_rejected,positives,null_zero)!=(42,42,2,42):raise ValueError('exact final boundary coverage differs')
 dependency=b.doc(m['stages/dependency-restore-2/proof.json'])
 if dependency['sourceBeforeAfterIdentical'] is not True or dependency['destinationExact'] is not True or dependency['regularFiles']!=11857 or dependency['relativeInternalLinks']!=2 or dependency['counts']['regularBytes']!=68841480 or dependency['sdkLockSHA256']!='e0b4084c372bc764b73faa25f2a311a3649e5b7ffb341352f25c13a891b7de65' or dependency['frozenPiSourceSHA256']!='39a9951108d0e7faf60886302ce4bc953d8c51d5ebad22370301f23b29aa2504' or dependency['frozenModelDataSHA256']!='8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8':raise ValueError('original local frozen dependency proof differs')
 if dict(collections.Counter(v['stage'] for v in fixture_provenance))!={name:47 for name in ('regression-ordinary','regression-ordinary-restored','regression-race-restored')} or fixture_provenance!=json.loads(b.read(BASE,'historical-fixture-hash-divergences.json',1<<20)):raise ValueError('exact141 pre-redaction fixture input provenance differs')
 return dict(schema=1,sourceCommit=c.SOURCE,trueMalformedSendREDs=true_red,finalNegativeLocalRejections=green_rejected,finalBusinessWirePositiveControls=positives,nativeNullZeroRequestCaptures=null_zero,stages=stages,manifestArtifactReferences=references,fixtureInputReferences=fixture_inputs,preRedactionFixtureReferences=141,all141RawInputSourceProofsRetained=True,mandatory8SupplementedSeparately=True,mainNotRunIDs=sorted(c.NOT_RUN),fixtureFailuresRemainHistorical=True,credentialsRead=False,paidCalls=0,formalApplicabilityRenewed=False,releaseCertified=False),resources
def main():
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);a=p.parse_args();b=readers.small();full=readers.full()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20))
 if index['kind']!='ticket17-offline-final-payload-tdd' or index['sourceCommit']!=c.SOURCE or index['archive']['file']!=c.ARCHIVE or index['archive']['prefix']!='ticket17-offline-tdd' or index['credentialsRead'] is not False or type(index['paidCalls']) is not int or index['paidCalls']!=0 or index['releaseCertified'] is not False or index['originalsUnchangedAfterSnapshot'] is not True:raise ValueError('fixed offline retention identity differs')
 pins,m,archive=full.verify(b,BASE,a.common_root,index);summary,resources=analyse(b,index,pins,m)
 follow=dict(index,observe=resources);_,values,_=full.verify(b,BASE,a.common_root,follow)
 for name in resources:
  row=b.doc(values[name]);keys={'bodies','calls','events','hostPermits','permits','waiters'}
  if set(row)!=keys or any(type(value) is not int or value!=0 for value in row.values()):raise ValueError('original public resource release differs')
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent complete history summary differs')
 print(json.dumps(dict(verified=True,archive=archive,resourceRecords=len(resources),summary=summary),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
