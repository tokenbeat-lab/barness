#!/usr/bin/env python3
"""Stream the complete historical gate custody, preserving its actual false result."""
import argparse,json,sys,types
from pathlib import Path
import readers
BASE=Path(__file__).resolve().parent;SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';PLAN='f97c97c158189bc741a22ea714a1cc64053ea79755d688f8a66450d87fab6985';ORIGINAL='bbf895e5cfab00ddab99ae5ef03479cbfe8ffd141d31596b5d9e2c2465877bfe';MATRIX='7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f';CLI='ee246d0e05346d4b6295f60847e64095db45c10306f85db7ba866fef58e79bca';HELPER='0ee913b5856de200b75f550a33c34a8ab8d4f3e52f445221f9c6b820d4790ee6'
CONSUMERS={'ordinary':'e8fa51864d50d61fa27359d84cd04e85b1967ffb6110cb22aa363e10ced754a3','race':'24698beea29fce131a26a4c8ab6fa308709f185611210065744cf1f014feae4a'}
ARGS={'vet':['vet','./...'],'vet-live':['vet','-tags','live','./...'],'ordinary':['test','-count=1','-timeout=20m','./...'],'race':['test','-race','-count=1','-timeout=75m','./...']}
VERIFY=('verify-text-jev','verify-text-luna','verify-visual-jev','verify-visual-decisions','verify-current-jev','verify-minimax-chat','verify-minimax-responses','verify-minimax-messages')
def numeric(value,wanted):return type(value) is int and value==wanted
def suites(b,index,common,m):
 result={}
 if set(index['suiteReferences'])!=set(CONSUMERS):raise ValueError('fixed original suite reference set differs')
 for label,digest in CONSUMERS.items():
  base=BASE.parent/label;ref=index['suiteReferences'][label]
  if ref['frontdoor']!='../'+label or b.sha(b.read(base,'index.json',1<<20))!=ref['indexSHA256']:raise ValueError('original suite index pin differs')
  ix=b.doc(b.read(base,'index.json',1<<20))
  if ix['sourceCommit']!=SOURCE or ix['suite']!=label or ix['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/incomplete-'+label+'-raw.tar.gz' or ix['gateVerdictCaptured'] is not False:raise ValueError('fixed original suite capsule differs')
  # Import only the pinned, read-only consumer. Its captured Go programs,
  # observer and original operator controllers are never executed here.
  raw=b.read(base,'verify.py',64<<10)
  if b.sha(raw)!=digest:raise ValueError('reviewed native consumer source differs')
  module=types.ModuleType('fixed_'+label+'_consumer');module.__file__=str(base/'verify.py');exec(compile(raw,str(base/'verify.py'),'exec'),module.__dict__)
  pins,values,archive=readers.full().verify(b,base,common,ix);summary=module.analyse(b,ix,pins,values,archive)
  if summary!=b.doc(b.read(base,'summary.json',1<<20)) or values['source/source-at-start.json']!=m['source/source-at-start.json'] or values['frozen/plan-v2.json']!=m['outer/plan-v2.json'] or values['receipts/'+label+'.json']!=m['outer/gate-command-exits/'+label+'.json']:raise ValueError('original suite cross-capsule source/plan/command binding differs')
  result[label]=dict(archive=archive,cases=summary['cases'],statuses=summary['statuses'],artifactReferences=summary['artifactReferences'],truePIDiff=summary['truePIDiff'],allNamedPressureCases=summary['allNamedPressureCases'],requiredPRESSURERegexCases=summary['requiredPRESSURERegexCases'],numericExit=summary['exitCode'],sourceCommit=SOURCE)
 return result
def analyse(b,index,pins,m,native):
 doc=lambda name:b.doc(m['outer/'+name]);source=b.doc(m['source/source-at-start.json']);plan=doc('plan-v2.json');done=doc('completed.json');intent=doc('execution-intent.json');review=doc('keyless-source-review.json');report=doc('gate/release-report.json');cmd=doc('actual-current-releasegate-command.json');build=doc('build-current-public-releasegate-command.json')
 if source['sourceCommit']!=SOURCE or len(source['files'])!=1056 or b.sha(m['source/source-at-start.json'])!='4bb4e654f4f23c7faa6fbda6ef24d3ddd7ab1ef41a095d8a6c6e6ea181b66620' or plan['source']!=source or b.sha(m['outer/plan-v2.json'])!=PLAN or b.sha(m['outer/plan.json'])!=ORIGINAL:raise ValueError('original7931/1056 source or plan chain differs')
 for value in (plan,done):
  if value['releaseCertified'] is not False or value['assemblyComplete'] is not False or value['assemblyStatus']!='REFUSED' or not numeric(value['newRowsImported'],0):raise ValueError('failed assembly cannot become certified or import rows')
 if done['state']!='COMPLETED' or done['mode']!='incomplete-input-diagnostic' or not numeric(done['outerExit'],1) or done['reportPresent'] is not True or done['reportError'] is not None or done['releasePass'] is not False or done['success'] is not False or done['certificationBlocked'] is not True or done['allFourNumeric'] is not True or done['allFourZero'] is not False or done['commandBindingsValid'] is not True or done['reportConsistent'] is not True or done['verificationPass'] is not True or done['credentialsRead'] is not False or not numeric(done['paidCalls'],0):raise ValueError('closed actual failed completion differs')
 if done['planSHA256']!=PLAN or done['originalPlanSHA256']!=ORIGINAL or intent['state']!='STARTING' or intent['planSHA256']!=PLAN or intent['originalPlanSHA256']!=ORIGINAL or intent['reviewSHA256']!=b.sha(m['outer/keyless-source-review.json']):raise ValueError('original STARTING-to-completed review binding differs')
 if report['pass'] is not False or report['commit']!=SOURCE or plan['before']['matrixSHA256']!=MATRIX or b.sha(m['outer/gate/support-matrix.json'])!=MATRIX or b.sha(m['outer/gate/catalog-snapshot.json'])!=source['files']['ai/release/catalog-snapshot.json']:raise ValueError('honest report/source/default matrix differs')
 numeric_exits={}
 for label,args in ARGS.items():
  receipt=doc('gate-command-exits/'+label+'.json');wanted=0 if label.startswith('vet') else 1
  if receipt['state']!='COMPLETED' or not numeric(receipt['exitCode'],wanted) or receipt['argv']!=['/usr/local/go/bin/go',*args] or receipt['cwd']!='$REPO' or receipt['credentialsRead'] is not False or receipt!=done['actualCommands'][label]:raise ValueError('actual four numeric command receipt differs')
  numeric_exits[label]=receipt['exitCode']
 rows=report['commands'];ids=['go-vet','go-vet-live','go-test','go-test-race']
 if len(rows)!=4 or [row['id'] for row in rows]!=ids or [row['passed'] for row in rows]!=[True,True,False,False] or any(type(row['passed']) is not bool for row in rows):raise ValueError('public boolean commands do not bind actual0/0/1/1')
 if set(done['verificationExits'])!=set(VERIFY):raise ValueError('exact eight verifier receipts required')
 for label in VERIFY:
  row=doc(label+'-command.json')
  if row['state']!='COMPLETED' or not numeric(row['exitCode'],0) or not numeric(done['verificationExits'][label],0) or row['credentialsRead'] is not False or row['credentialsInherited'] is not False or row['launcherAudit']!={'passed':True,'findings':[]}:raise ValueError('original keyless verification command differs')
 if review['decision']!='APPROVED' or len(review['reviewedBy'])!=2 or review['sourceCommit']!=SOURCE or review['planSHA256']!=PLAN or review['helperSHA256']!=HELPER or review['compiledCLISHA256']!=CLI or intent['credentialsRead'] is not False or not numeric(intent['paidCalls'],0):raise ValueError('source review grants no paid or release qualification')
 for axis in ('Standards','Spec'):
  if review['reportsSHA256'][axis]!=pins['outer/independent-final-'+axis.lower()+'-review.json']['sha256']:raise ValueError('two-axis exact source review pin differs')
 if pins['outer/releasegate']['sha256']!=CLI or plan['compiledCLISHA256']!=CLI or plan['buildReceiptSHA256']!=b.sha(m['outer/build-current-public-releasegate-command.json']) or not numeric(build['exitCode'],0) or build['argv']!=['/usr/local/go/bin/go','build','-o','$REPO/.evidence/minimax-release-repair-incomplete-gate-1/releasegate','./ai/release/cmd/releasegate']:raise ValueError('actual compiled CLI/build binding differs')
 operations=plan['before']['operations'];operator=doc('operator-assets/index.json')
 for name,pin in operator.items():
  if pins['outer/operator-assets/'+name]!={key:pin[key] for key in ('bytes','sha256')}:raise ValueError('original operator asset index differs')
 if operations['helperSHA256']!=HELPER or pins['outer/operator-assets/barness-repair-incomplete-keyless.py.txt']['sha256']!=HELPER or operations['observerSHA256']!=pins['outer/observer-bin/go']['sha256'] or operations['observerSHA256']!='aec940df7db842009d9c977142dc35f604022b380ce383af7ce2dd85383879d7':raise ValueError('original helper/observer source pins differ')
 for suffix in ('','-v2'):
  pre=doc('preapproval-negative'+suffix+'.json');name='run'+('-v2' if suffix else '')+'-before-source-review-refusal-command.json';receipt=doc(name)
  if pre['expectedRefusal'] is not True or not numeric(pre['actualCLIExitCode'],1) or pre['noExecutionIntent'] is not True or pre['noGateOutput'] is not True or pre['planUnchanged'] is not True or pre['commandReceiptSHA256']!=b.sha(m['outer/'+name]) or not numeric(receipt['exitCode'],1):raise ValueError('preapproval refusal history differs')
 argv=cmd['argv']
 if not numeric(cmd['exitCode'],1) or cmd['state']!='COMPLETED' or cmd['credentialsRead'] is not False or cmd['credentialsInherited'] is not False or cmd['injectedCredential'] is not None or cmd['sourceVariable'] is not None or cmd['launcherAudit']!={'passed':True,'findings':[]} or argv[0]!='$REPO/.evidence/minimax-release-repair-incomplete-gate-1/releasegate' or argv.count('-live')!=17 or any('bundle' in flag for flag in argv if flag.startswith('-')) or argv[argv.index('-matrix')+1]!='$REPO/ai/live/support-matrix.json' or not numeric(plan['suppliedLiveDirectories'],17):raise ValueError('public compiled CLI default four-command invocation differs')
 if any(type(g['pass']) is not bool for g in report['gates']):raise ValueError('original boolean gate outcomes required')
 gates=[dict(id=g['id'],name=g['name'],passed=g['pass'],problems=g.get('problems',[])) for g in report['gates']]
 return dict(schema=1,sourceCommit=SOURCE,sourceFiles=1056,outerExit=1,numericExits=numeric_exits,verifierExits={name:0 for name in VERIFY},reportPresent=True,reportPass=False,reportConsistent=True,commandBindingsValid=True,releaseCertified=False,assemblyStatus='REFUSED',newRowsImported=0,matrixSHA256=MATRIX,compiledCLISHA256=CLI,helperSHA256=HELPER,fullOriginalOuterFiles=63,preapprovalRefusals=2,sourceReviewAxes=2,mainSuites=native,gates=gates,stdoutSHA256=pins['outer/actual-current-releasegate.log']['sha256'],stdoutBytes=pins['outer/actual-current-releasegate.log']['bytes'],originalRawStdoutPreserved=True,retentionCredentialsRead=False,retentionPaidCalls=0,currentSourceApplicabilityClaimed=False)
def main():
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);a=p.parse_args();b=readers.small()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 index=b.doc(b.read(BASE,'index.json',1<<20))
 if index['kind']!='closed-incomplete-gate-verdict' or index['sourceCommit']!=SOURCE or index['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/incomplete-outer-raw.tar.gz' or index['archive']['prefix']!='repair-incomplete-outer' or index['releaseCertified'] is not False or index['credentialsRead'] is not False or not numeric(index['paidCalls'],0) or index['originalsUnchangedAfterSnapshot'] is not True:raise ValueError('fixed original verdict capsule identity differs')
 pins,m,archive=readers.full().verify(b,BASE,a.common_root,index)
 if len(pins)!=64 or index['observe']!=sorted(set(pins)-{'outer/releasegate'}):raise ValueError('all exact outer metadata/compiled artifact inventory required')
 native=suites(b,index,a.common_root,m);summary=analyse(b,index,pins,m,native)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent verdict summary differs')
 print(json.dumps(dict(verified=True,originalSourceCommit=SOURCE,releaseCertified=False,outerArchive=archive,summary=summary),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
