#!/usr/bin/env python3
"""Read only the closed A failure; successful Jev derivation is not assembly."""
import argparse,json,sys
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';ROUND='release-repair-validation-1';MATRIX='7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f';HELPER='e5dc66b14e65011574c807602a77502a4b28b72fe64c1685953c7d2e53fda836'

def native(b,index,common):
 result={}
 for label in ('glm','effects'):
  ref=index['nativeReferences'][label];base=BASE.parent/label
  if ref['frontdoor']!='../'+label or b.sha(b.read(base,'index.json',1<<20))!=ref['indexSHA256']:raise ValueError('fixed original native reference differs')
  v=b.doc(b.read(base,'index.json',1<<20));helper=readers.small();helper.BASE=base;pins=helper.inventory(v['capsule']);_,m=helper.verify_archive(common,v['capsule']);result[label]=(v,pins,m)
 return result

def analyse(b,index,pins,m,refs):
 prefix='original/';doc=lambda n:b.doc(m[prefix+n]);inputs=doc('assemble-inputs.json');source=b.doc(m['frozen/source-at-start.json']);intent=doc('assemble-intent.json');originals={n:v for n,v in pins.items() if n.startswith(prefix)}
 if prefix+'assemble-completed.json' in pins or prefix+'default-merge-command.json' in pins or intent!={'state':'STARTING','credentialsRead':False,'paidCalls':0} or len(originals)!=603:raise ValueError('closed failed assembly must not acquire success markers')
 expected_roots={'assemble-intent.json','assemble-inputs.json','matrix-before.json','matrix-candidate.json',*[n+s for n in ('derive-jev-support','verify-derived-jev','validate-four-candidate') for s in ('-command.json','.log')]}
 if {n.removeprefix(prefix) for n in originals if n.count('/')==1}!=expected_roots:raise ValueError('closed failed assembly root inventory differs')
 if source['sourceCommit']!=SOURCE or len(source['files'])!=1056 or b.sha(m['frozen/source-at-start.json'])!='4bb4e654f4f23c7faa6fbda6ef24d3ddd7ab1ef41a095d8a6c6e6ea181b66620' or inputs['source']!=source:raise ValueError('original frozen assembly source differs')
 expected_ops=dict(helperSHA256=HELPER,observerSHA256='aec940df7db842009d9c977142dc35f604022b380ce383af7ce2dd85383879d7',realGoSHA256='005640c7ff93028cb704283b0f737f2db3faf8b51b2561170c769b83905da646')
 helper_ref=index['operationalHelperReference']
 if inputs['operations']!=expected_ops or helper_ref!={'file':'../../finalizer-design/helper-v4.py.txt','sha256':HELPER} or b.sha(b.read(BASE.parents[1]/'finalizer-design','helper-v4.py.txt',1<<20))!=HELPER:raise ValueError('original operational helper/Go/observer pins differ')
 gi,gpins,gm=refs['glm'];ei,epins,em=refs['effects'];protocol=ei['roots']['protocol114'];jev=ei['roots']['visual-jev'];groot={n:'$REPO/.evidence/minimax-release-repair-glm-1/glm-'+n+'/'+v.rsplit('/',1)[1] for n,v in gi['routes'].items()};eroot='$REPO/.evidence/minimax-release-repair-effects-1';out='$REPO/.evidence/minimax-release-repair-final-1'
 commands={'derive-jev-support':(0,['/usr/local/go/bin/go','run','./ai/live/cmd/jevformal','-protocol',eroot+'/protocol114/'+protocol.rsplit('/',1)[1],'-effects',eroot+'/visual-jev','-out',out+'/jev-support','-require-live']),'verify-derived-jev':(0,['/usr/local/go/bin/go','run','./ai/live/cmd/jevformal','-verify',out+'/jev-support','-require-live']),'validate-four-candidate':(1,['/usr/local/go/bin/go','run','./ai/live/cmd/supportmatrix','-matrix',out+'/matrix-candidate.json',*[groot[n] for n in ('chat','responses','messages')],out+'/jev-support'])}
 for name,(code,argv) in commands.items():
  v=doc(name+'-command.json')
  if v['state']!='COMPLETED' or type(v['exitCode']) is not int or v['exitCode']!=code or v['argv']!=argv or v['cwd']!='$REPO' or v['credentialsRead'] is not False or v['credentialsInherited'] is not False or v['launcherAudit']!={'passed':True,'findings':[]} or any('KEY' in n or 'PROXY' in n for n in v['environmentKeys']):raise ValueError('original keyless numeric command differs')
 diagnostic=m[prefix+'validate-four-candidate.log'].decode();expected='supportmatrix: lstat '+groot['messages']+'/live-report.json: no such file or directory\nexit status 1\n'
 if diagnostic!=expected or gi['routes']['messages']+'/live-report.json' in gpins or gm['frozen/source-at-start.json']!=m['frozen/source-at-start.json'] or em['frozen/source-at-start.json']!=m['frozen/source-at-start.json']:raise ValueError('honest missing Messages support diagnostic differs')
 if any(b.sha(m[n])!=MATRIX for n in (prefix+'matrix-before.json',prefix+'matrix-candidate.json','selectors/default-matrix.json')) or m[prefix+'matrix-before.json']!=m[prefix+'matrix-candidate.json']:raise ValueError('original default/candidate matrix changed')
 for name in ('history','reuse','scope'):
  if inputs['proofs'][name]!=b.doc(em['frozen/'+name+'.json']):raise ValueError('original assembly reuse/history/scope differs')
 actuals=inputs['actuals'];counts=[]
 for label,ix,nativepins,values in [('glm',gi,gpins,gm),('effects',ei,epins,em)]:
  roots=ix['routes'] if label=='glm' else ix['roots']
  for name,root in roots.items():
   full=ROUND+('-glm-' if label=='glm' else '-')+name;row=actuals[full];receipt=values['canonical/'+name+'.json'];leaf='formal-glm-report.json' if label=='glm' else 'multimodal-report.json' if name=='protocol114' else 'report.json';calls=22 if label=='glm' else 12;status='FAIL' if label=='glm' and name=='messages' else 'PASS'
   expected_path=('.evidence/minimax-release-repair-glm-1/glm-'+name+'/'+root.rsplit('/',1)[1] if label=='glm' else '.evidence/minimax-release-repair-effects-1/'+name+('/'+root.rsplit('/',1)[1] if name=='protocol114' else ''))+'/'+leaf
   if row['report']!=expected_path or row['receiptSHA256']!=b.sha(receipt) or row['reportSHA256']!=b.sha(values[root+'/'+leaf]) or row['sourceCommit']!=SOURCE or row['native']['state']!='REPORTED' or row['native']['status']!=status or row['native']['calls']!=calls or row['native']['attempts']!=calls:raise ValueError('original six native actual references differ')
   counts.append(calls)
 if len(actuals)!=6 or sum(counts)!=102:raise ValueError('fixed six actual102 accounting differs')
 derived='original/jev-support/';seal=b.doc(m[derived+'formal-integrity.json']);expected_seal={n.removeprefix(derived):v['sha256'] for n,v in pins.items() if n.startswith(derived) and n!=derived+'formal-integrity.json'}
 if seal!=expected_seal or b.doc(m[derived+'audit.json'])['findings'] or b.doc(m[derived+'manifest.json'])!={'kind':'jev-formal-support','schema':1}:raise ValueError('complete derived Jev seal/audit differs')
 copied=0
 for dest,origin in [('protocol',protocol),('effects',jev)]:
  actual={n.removeprefix(derived+dest+'/'):v for n,v in pins.items() if n.startswith(derived+dest+'/')};expected={n.removeprefix(origin+'/'):v for n,v in epins.items() if n.startswith(origin+'/')}
  if actual!=expected:raise ValueError('derived protocol/effects native copies differ')
  copied+=len(actual)
 report=b.doc(m[derived+'live-report.json']);scenes=report['scenarios'];budget=report['budget']
 if report['gitCommit']!=SOURCE or len(scenes)!=12 or {s['id'] for s in scenes}!=set(report['expected']) or any(s['outcome']!='PASS' or s['calls']!=1 or s['attempts']!=1 for s in scenes) or budget['callsUsed']!=12 or budget['httpAttempts']!=12 or budget['inputTokensUsed']!=9862 or budget['outputTokensUsed']!=372:raise ValueError('derived native protocol accounting differs')
 return dict(schema=1,sourceCommit=SOURCE,assemblyOutcome='FAIL',assemblyIntent='STARTING',assembleCompletedPresent=False,defaultMergePresent=False,commands={n:v[0] for n,v in commands.items()},failure='Missing original GLM Messages live-report.json; candidate validation refused before matrix writes',missingMessagesSupport=True,derivedJev=dict(deriveExit=0,verifyExit=0,scenarios=12,statuses={'PASS':12},calls=12,attempts=12,copiedNativeFiles=copied,sealFiles=len(seal),auditFindings=0),originalRegularFiles=603,originalLinks=0,matrixBeforeSHA256=MATRIX,matrixCandidateSHA256=MATRIX,defaultMatrixSHA256=MATRIX,matrixChanged=False,sixActualCalls=102,sixActualAttempts=102,retentionPaidCalls=0,retentionCredentialsRead=False,releaseCertified=False)

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[5]);a=p.parse_args();b=readers.small()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20));cap=index['capsule']
 if index['sourceCommit']!=SOURCE or index['kind']!='closed-failed-repair-assembly' or index['boundedHelperSHA256']!=readers.SMALL or cap['id']!='failedRepairAssembly7931' or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/assembly-failed-raw.tar.gz' or cap['archive']['prefix']!='repair-assembly-failed' or set(index['nativeReferences'])!={'glm','effects'}:raise ValueError('fixed assembly retention identity differs')
 pins=b.inventory(cap)
 if len(cap['observe'])>512 or len(cap['observe'])!=len(set(cap['observe'])) or sum(pins[n]['bytes'] for n in cap['observe'])>16<<20:raise ValueError('observed metadata bound differs')
 archive,m=b.verify_archive(a.common_root,cap);refs=native(b,index,a.common_root);summary=analyse(b,index,pins,m,refs)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent failed assembly summary differs')
 for row in b.doc(b.read(BASE,'tdd/raw-log-index.json',1<<20))['logs']:b.decoded_gzip(row)
 print(json.dumps(dict(verified=True,assemblyOutcome='FAIL',missingMessagesSupport=True,matrixChanged=False,releaseCertified=False,retentionPaidCalls=0,retentionCredentialsRead=False,archive=archive,summary=summary),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
