#!/usr/bin/env python3
"""Verify only the completed ordinary command; race/outer remain separate."""
import argparse,collections,json,re,sys
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';SESSION='20261018T120404.152055000Z';PLAN='f97c97c158189bc741a22ea714a1cc64053ea79755d688f8a66450d87fab6985';FAILURES=sorted('GLM12-formal-import-chat GLM12-formal-import-messages GLM12-formal-import-responses GLM12-reported-known-usage-chat GLM12-reported-known-usage-messages GLM12-reported-known-usage-responses'.split())

def analyse(b,index,pins,m,archive):
 name=index['mainManifest'];prefix='ordinary/'+SESSION+'/';manifest=b.doc(m[name]);source=b.doc(m['source/source-at-start.json']);plan=b.doc(m['frozen/plan-v2.json']);receipt=b.doc(m['receipts/ordinary.json']);cases=manifest['cases']
 if source['sourceCommit']!=SOURCE or len(source['files'])!=1056 or plan['source']!=source or b.sha(m['frozen/plan-v2.json'])!=PLAN or plan['releaseCertified'] is not False or plan['assemblyComplete'] is not False or type(plan['newRowsImported']) is not int or plan['newRowsImported']!=0:raise ValueError('original immutable source7931/plan-v2 differs')
 if receipt['state']!='COMPLETED' or type(receipt['exitCode']) is not int or receipt['exitCode']!=1 or receipt['credentialsRead'] is not False or receipt['cwd']!='$REPO' or receipt['argv']!=['/usr/local/go/bin/go','test','-count=1','-timeout=20m','./...']:raise ValueError('original actual ordinary numeric exit differs')
 statuses=dict(collections.Counter(c['status'] for c in cases));failed=sorted(c['id'] for c in cases if c['status']=='FAIL')
 if manifest['git_commit']!=SOURCE or len(cases)!=12451 or len({c['id'] for c in cases})!=12451 or statuses!={'PASS':12445,'FAIL':6} or failed!=FAILURES or b.doc(m[prefix+'audit.json'])['findings']:raise ValueError('actual complete ordinary failures/counts/audit differ')
 refs=0
 for case in cases:
  directory=case['dir'].rstrip('/');b.relative(directory)
  for leaf,digest in case['artifacts'].items():
   b.relative(leaf)
   if pins.get(prefix+directory+'/'+leaf,{}).get('sha256')!=digest:raise ValueError('original full native artifact reference differs')
   refs+=1
 pi=[c for c in cases if c['id'].startswith('PIDIFF-')];pressure=[c for c in cases if 'pressure' in c['id'].lower()];required_pressure=[c for c in cases if re.match('^E08-(policy-pressure|admission-burst-pressure|mixed-pressure)-',c['id'])]
 if len(pi)!=868 or any(c['status']!='PASS' for c in pi) or len(pressure)!=29 or len(required_pressure)!=15 or any(c['status']!='PASS' for c in pressure):raise ValueError('actual true PI/pressure counts differ')
 failures=[]
 for c in cases:
  if c['status']=='PASS':continue
  path=prefix+c['dir']+'/assertions.json';v=b.doc(m[path]);failures.append(dict(id=c['id'],status=c['status'],assertionsSHA256=pins[path]['sha256'],assertions=v))
 return dict(schema=1,suite='ordinary',sourceCommit=SOURCE,sourceFiles=1056,exitCode=1,seconds=receipt['seconds'],cases=12451,statuses=statuses,truePIDiff=868,allNamedPressureCases=29,requiredPRESSURERegexCases=15,pressureIDs=[c['id'] for c in pressure],requiredPressureIDs=[c['id'] for c in required_pressure],artifactReferences=refs,auditFindings=0,failures=failures,archive=archive,gateVerdictCaptured=False,raceCaptured=False,packageLevelOuterLogCaptured=False,retentionPaidCalls=0,retentionCredentialsRead=False,releaseCertified=False)

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[6]);a=p.parse_args();b=readers.small()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20))
 if index['sourceCommit']!=SOURCE or index['suite']!='ordinary' or index['mainManifest']!='ordinary/'+SESSION+'/manifest.json' or index['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/incomplete-ordinary-raw.tar.gz' or index['archive']['prefix']!='repair-incomplete-ordinary' or index['raceCaptured'] is not False or index['gateVerdictCaptured'] is not False:raise ValueError('fixed original ordinary capsule identity differs')
 pins,m,archive=readers.full().verify(b,BASE,a.common_root,index);summary=analyse(b,index,pins,m,archive)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent ordinary summary differs')
 for row in b.doc(b.read(BASE,'tdd/raw-log-index.json',1<<20))['logs']:b.decoded_gzip(row)
 print(json.dumps(dict(verified=True,sourceCommit=SOURCE,ordinaryExit=1,cases=summary['cases'],statuses=summary['statuses'],truePIDiff=868,allNamedPressureCases=29,requiredPRESSURERegexCases=15,artifactReferences=summary['artifactReferences'],archive=archive,raceCaptured=False,gateVerdictCaptured=False,releaseCertified=False,credentialsRead=False,paidCalls=0),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
