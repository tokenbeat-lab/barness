#!/usr/bin/env python3
"""Bounded read-only replay of gap2 fixture/diagnostic/preflight repair; no certificate."""
import argparse,collections,hashlib,json,os,re,stat,sys,types
from pathlib import Path,PurePosixPath
BASE=Path(__file__).resolve().parent
SOURCE='572570bbe4a9322fb02ed740bd2a10c9b8ce792f';CODE='1a0ce3458d5c499b95c03e3994f3e7be35e88ea2'
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
CONTRACT=['go','test','./ai/e2e','-count=1','-run','^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract|TestReleaseRecordedBundleDiagnosticProcess)$']
PACKAGE=['go','test','./ai/internal/testkit/release','-count=1']
CHANGED=['.scratch/barness-ai-minimax/tools/routed_validation.py','.scratch/barness-ai-minimax/tools/test_routed_validation.py','ai/e2e/release_recorded_bundle_process_test.go','ai/e2e/release_recorded_bundle_test.go','ai/internal/testkit/release/artifacts_test.go','ai/internal/testkit/release/decisions_record_test.go']

def helper():
 p=BASE.parents[1]/'ticket-13/implementation/verify.py';fd=os.open(p,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
 with os.fdopen(fd,'rb') as f:info=os.fstat(f.fileno());raw=f.read((64<<10)+1)
 if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:raise ValueError('pinned reader differs')
 b=types.ModuleType('bounded_gap2');b.__file__=str(p);exec(compile(raw,str(p),'exec'),b.__dict__);b.BASE=BASE;return b

def analyse(b,index,pins,m,archive):
 def doc(n):return b.doc(m[n])
 def command(n,code,argv=None):
  v=doc(n)
  if (v['state']!='COMPLETED' or type(v['exitCode']) is not int or v['exitCode']!=code or v['credentialsRead'] is not False or v['credentialsInherited'] is not False or
      v['sourceVariable'] is not None or v['injectedCredential'] is not None or v['launcherAudit']!={'passed':True,'findings':[]} or any(re.search('KEY|TOKEN|SECRET',k) for k in v['environmentKeys'])):raise ValueError('original keyless command differs')
  if argv is not None and v['argv']!=argv:raise ValueError('original argv differs')
  return v
 final=doc('preflight/source-at-start.json');before=doc('tdd/source-before.json')
 if final['sourceCommit']!=SOURCE or len(final['files'])!=1048 or len(before['files'])!=1047:raise ValueError('selected source size/commit differs')
 if sorted(n for n in final['files'] if before['files'].get(n)!=final['files'][n])!=CHANGED:raise ValueError('source scope differs')
 for n in CHANGED:
  if pins['source/after/'+n]['sha256']!=final['files'][n]:raise ValueError('final source bytes differ')
  if n in before['files'] and pins['source/before/'+n]['sha256']!=before['files'][n]:raise ValueError('prior source bytes differ')
 for stage in ('green-release-package','race-release-package','race-release-package-complete','race-diagnostic-process','green-preflight','tools-all','vet','vet-live'):
  if doc('tdd/source-'+stage+'.json')['files']!=final['files']:raise ValueError('GREEN selected source differs')
 for stage in ('green-fixture-semantics','green-diagnostic-process'):
  v=doc('tdd/source-'+stage+'.json')['files'];delta=sorted(n for n in final['files'] if v.get(n)!=final['files'][n])
  if delta!=CHANGED[:2] or any(v[n]!=before['files'][n] for n in delta):raise ValueError('early Go slice/source staging differs')
 command('tdd/red-release-package-command.json',1,PACKAGE)
 for stage in ('red-fixture-semantics','red-diagnostic-process','red-preflight'):command('tdd/'+stage+'-command.json',1)
 for stage in ('green-fixture-semantics','green-diagnostic-process','green-preflight','tools-all','vet','vet-live','race-diagnostic-process'):command('tdd/'+stage+'-command.json',0)
 command('tdd/green-release-package-command.json',0,PACKAGE);race=command('tdd/race-release-package-complete-command.json',0,PACKAGE[:2]+['-race']+PACKAGE[2:])
 if race['timeoutSeconds']!=900 or not 300<race['seconds']<900:raise ValueError('bounded replacement race differs')
 for n,count,outcome in [('red-preflight',3,rb'FAILED \(failures=3\)'),('green-preflight',3,rb'OK'),('tools-all',61,rb'OK')]:
  if re.search(rb'Ran '+str(count).encode()+rb' tests in [0-9.]+s\s+'+outcome+rb'\s*$',m['tdd/'+n+'.log']) is None:raise ValueError('CLI RED/GREEN count differs')
 if m['tdd/red-fixture-semantics.log'].count(b'case artifact integrity inventory missing or over file limit')!=15 or b'NOT_RUN' not in m['tdd/red-diagnostic-process.log']:raise ValueError('behavioral RED differs')
 timeout=doc('tdd/race-timeout-annotation.json');interrupted=doc('tdd/race-release-package-command.json')
 if (interrupted['state']!='STARTING' or interrupted['exitCode']!=1 or interrupted['timeoutSeconds']!=360 or timeout['originalReceiptUnchanged'] is not True or
     timeout['sourceStillIdentical'] is not True or timeout['observedOwnedProcessesNoLongerPresent'] is not True or timeout['credentialsRead'] is not False or timeout['paidCalls']!=0):raise ValueError('original timeout history differs')
 checks=doc('preflight/checks.json');names=['vet','vet-live','no-key-live','no-key-visual-proxy','no-key-release-contracts','no-key-release-package']
 if checks!={'passed':True,'commands':{n:0 for n in names},'credentialsRead':False}:raise ValueError('six actual checks differ')
 for n in names:command('preflight/'+n+'-command.json',0,CONTRACT if n=='no-key-release-contracts' else PACKAGE if n=='no-key-release-package' else None)
 if doc('preflight/execution-authorization.template.json')['decision']!='PENDING' or any(n.endswith('/execution-authorization.json') for n in pins):raise ValueError('invented approval')
 observed=doc('tdd/actual-six-check-observed.json')
 if (observed['sourceCommit']!=SOURCE or observed['checks']!=checks or observed['authorization']!='PENDING' or observed['checkOnly'] is not True or observed['newRoundCreated'] is not False or
     observed['original39Unchanged'] is not True or len(observed['original18CanonicalPins'])!=18 or len(observed['original21PriorPins'])!=21 or observed['credentialsRead'] is not False or observed['paidCalls']!=0):raise ValueError('closed39/no new round proof differs')
 for key,leaf in [('sourceManifestSHA256','source-at-start.json'),('binariesManifestSHA256','binaries.json'),('planSHA256','plan.json'),('checksSHA256','checks.json')]:
  if observed[key]!=pins['preflight/'+leaf]['sha256']:raise ValueError('actual check pin differs')
 for leaf in ('actual-six-check-observed.json','race-timeout-annotation.json'):
  if b.read(BASE,leaf,1<<20)!=m['tdd/'+leaf]:raise ValueError('frontdoor/original proof differs')
 links=doc('negative-fixture-links.json')['links']
 if links!=index['negativeFixtureLinkMetadata'] or len(links)!=1:raise ValueError('negative link projection differs')
 for n,v in links.items():
  if n in pins or not n.startswith('preflight/no-key-live/') or not n.endswith('/formal-corrupt-symlink/formal-integrity.json') or b.sha(v['target'].encode())!=v['sha256'] or len(v['target'].encode())!=v['bytes']:raise ValueError('negative link metadata differs')
 external=b.doc(b.read(BASE,'external-references.json',1<<20))
 if len(external['binaries'])!=3 or external['codeCommit']!=CODE or external['checkedSourceCommit']!=SOURCE:raise ValueError('unused binary references differ')
 for row in external['binaries']:
  leaf=Path(row['location']).name
  if row['copied'] is not False or leaf not in ('live.test','chineseeval','visualeval') or doc('preflight/binaries.json')[leaf]!=row['sha256'] or 'preflight/'+leaf in pins:raise ValueError('unused binary pin differs')
 manifests=[];refs=0
 for n in sorted(m):
  if not n.endswith('/manifest.json'):continue
  value=doc(n);root=n.rsplit('/',1)[0];audit=doc(root+'/audit.json') if root+'/audit.json' in m else None
  if audit is not None and audit['findings']:raise ValueError('native audit findings')
  for case in value['cases']:
   directory=PurePosixPath(case['dir'])
   if directory.is_absolute() or '..' in directory.parts:raise ValueError('unsafe native case path')
   for leaf,digest in case['artifacts'].items():
    b.relative(leaf)
    if pins.get(root+'/'+str(directory/leaf),{}).get('sha256')!=digest:raise ValueError('native artifact pin differs')
    refs+=1
  manifests.append(dict(member=n,cases=len(value['cases']),statuses=dict(collections.Counter(c['status'] for c in value['cases'])),gitCommit=value['git_commit'],auditFindings=0 if audit is not None else None))
 for stage,want in [('red-diagnostic-process',{'PASS':3,'FAIL':1}),('green-diagnostic-process',{'PASS':4}),('race-diagnostic-process',{'PASS':4})]:
  rows=[r for r in manifests if r['member'].startswith('tdd/'+stage+'/')]
  if len(rows)!=1 or rows[0]['statuses']!=want:raise ValueError('public diagnostic process RED/GREEN differs')
 for prefix,want in [('preflight/no-key-release-contracts/',{'PASS':10}),(index['visualProxyRoot']+'/',{'PASS':16})]:
  rows=[r for r in manifests if r['member'].startswith(prefix)]
  if len(rows)!=1 or rows[0]['statuses']!=want or rows[0]['gitCommit']!=SOURCE:raise ValueError('fresh actual contract/visual bundle differs')
 main=[r for r in manifests if r['member'].startswith('preflight/no-key-live/') and r['cases']>100]
 if len(main)!=1 or main[0]['statuses']!={'PASS':336,'NOT_RUN':62} or main[0]['auditFindings']!=0:raise ValueError('fresh keyless live outcome differs')
 return dict(schema=1,codeCommit=CODE,checkedSourceCommit=SOURCE,selectedFiles=1048,sourceFilesChanged=CHANGED,fullReleasePackage={'RED':1,'ordinaryGREEN':0,'raceGREEN':0,'replacementRaceSeconds':race['seconds'],'first360sAttemptSuccessful':False},publicDiagnostic={'RED':{'PASS':3,'FAIL':1},'ordinaryGREEN':{'PASS':4},'raceGREEN':{'PASS':4}},publicPreflightCLI={'RED':3,'GREEN':3,'toolsGREEN':61},actualSixChecks=checks,nativeManifests=manifests,artifactReferencesVerified=refs,negativeFixtureLinksMetadataOnly=1,unusedBinaryHashReferences=3,authorization='PENDING',newRoundCreated=False,original39Unchanged=True,retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,releaseCertified=False,archive=archive)

def main():
 import capsule
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20))
 if (index['kind']!='offline-gap2-fixture-diagnostic-and-six-check-repair' or index['sourceCommit']!=SOURCE or index['codeCommit']!=CODE or index['archive']['file']!='.evidence/minimax-retained/ticket-14-offline-gap2/gap2-raw.tar.gz' or index['archive']['prefix']!='offline-gap2'):raise ValueError('fixed capsule identity differs')
 pins,m,archive=capsule.verify(b,BASE,a.common_root,index);summary=analyse(b,index,pins,m,archive)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent summary differs')
 print(json.dumps(dict(verified=True,codeCommit=CODE,checkedSourceCommit=SOURCE,retentionCredentialsRead=False,retentionPaidCalls=0,releaseCertified=False,archive=archive),indent=2))

if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
