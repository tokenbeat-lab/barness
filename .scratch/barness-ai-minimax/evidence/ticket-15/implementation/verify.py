#!/usr/bin/env python3
"""Read-only fixed six CLI TDD replay; executes no captured command or vendor request."""
import argparse,hashlib,json,os,re,stat,sys,types
from pathlib import Path
BASE=Path(__file__).resolve().parent
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
CODE='09228e06e6ff6ea5062e3636228cf59724f8041e'
FAILED={'red-plan','green-plan','red-native','red-native-2','green-native','green-native-2','green-native-3','red-preflight','red-core','green-core','red-publication','green-publication','red-certification','red-certification-2','red-shape','red-check-failures','red-publication-payload'}

def helper():
 p=BASE.parents[1]/'ticket-13/implementation/verify.py';fd=os.open(p,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
 with os.fdopen(fd,'rb') as f:info=os.fstat(f.fileno());raw=f.read((64<<10)+1)
 if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:raise ValueError('pinned finite reader differs')
 b=types.ModuleType('bounded_repair');b.__file__=str(p);exec(compile(raw,str(p),'exec'),b.__dict__);b.BASE=BASE;return b

def analyse(b,ix,m):
 pins=b.inventory(ix['capsule']);doc=lambda n:b.doc(m[n]);source=doc('source/committed-source.json');measured=doc('tdd/source-final-tools.json');commands={}
 if source['sourceCommit']!=CODE or measured['measuredHEAD']!=ix['measuredHEAD'] or measured['files']!=source['files'] or doc('tdd/source-final-vets.json')['files']!=source['files']:raise ValueError('measured WIP and exact committed source differ')
 for n,v in pins.items():
  if n.startswith('source/tools/') and source['files'].get('.scratch/barness-ai-minimax/tools/'+n.rsplit('/',1)[1])!=v['sha256']:raise ValueError('committed source bytes differ')
 for n in sorted(m):
  if not n.startswith('tdd/') or len(n.split('/'))!=2 or not n.endswith('-command.json'):continue
  stage=n.removeprefix('tdd/').removesuffix('-command.json');v=doc(n);code=1 if stage in FAILED else 0
  if v['state']!='COMPLETED' or v['exitCode']!=code or v['credentialsRead'] is not False or v['credentialsInherited'] is not False or v['sourceVariable'] is not None or v['injectedCredential'] is not None or v['launcherAudit']!={'passed':True,'findings':[]} or any(any(token in key for token in ('KEY','TOKEN','SECRET')) for key in v['environmentKeys']):raise ValueError('keyless command outcome differs: '+stage)
  commands[stage]=dict(exitCode=code,seconds=v['seconds'])
 if not FAILED<=set(commands) or not {'final-tools','final-vet','final-vet-live','green-publication-payload','green-certification','green-shape','green-check-failures'}<=set(commands):raise ValueError('fixed RED/GREEN history missing')
 match=re.search(rb'Ran ([0-9]+) tests in [0-9.]+s\s+OK\s*$',m['tdd/final-tools.log'])
 if match is None or int(match[1])!=71:raise ValueError('complete tools regression differs')
 original=doc('certified/actual-six-check-observed.json');receipts={n.removesuffix('.json'):h for n,h in original['original18CanonicalPins'].items()};receipts.update({n:v['receiptSHA256'] for n,v in original['original21PriorPins'].items()})
 prefix='tdd/public-cli-artifacts-final/test_release_repair.Repair.';plan_test=prefix+'test_six_plan_is_pending_and_preserves_known545_history/'
 for phase,names in [('glm',{'glm-chat','glm-responses','glm-messages'}),('effects',{'protocol114','visual-jev','visual-decisions'})]:
  plan=doc(plan_test+phase+'/plan.json');auth=doc(plan_test+phase+'/execution-authorization.template.json')
  if plan['round']!='release-repair-validation-1' or plan['phase']!=phase or {v['baseName'] for v in plan['routes'].values()}!=names or plan['budget']!={'priorCalls':545,'priorAttempts':545,'additionalMaxCalls':108,'additionalMaxAttempts':132,'totalMaxCalls':653,'totalMaxAttempts':677} or auth['decision']!='PENDING' or auth['requestReference']!='' or auth['requestText']!='' or {n:v['receiptSHA256'] for n,v in plan['history']['receipts'].items()}!=receipts or plan['history']['known']!={'calls':545,'attempts':545} or len(plan['reuse']['inputs'])!=15 or plan['scope']['matrixSHA256']!='7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f':raise ValueError('pending fixed scope or immutable original39 history differs')
  for v in plan['routes'].values():
   if v['networkRoute']!=('china-direct' if v['provider']=='zhipu' else 'international-loopback-proxy'):raise ValueError('routing differs')
   if v['provider']=='zhipu' and 'BARNESS_AI_LIVE_PROXY' in v['environment'] or v['baseName'].startswith('visual-') and 'http://127.0.0.1:7890' not in v['argv'] or v['baseName']=='protocol114' and v['environment'].get('BARNESS_AI_LIVE_PROXY')!='http://127.0.0.1:7890':raise ValueError('explicit direct/proxy contract differs')
 inventories=[]
 for n in sorted(m):
  if not n.startswith(prefix) or not n.endswith('/inventory.json'):continue
  v=doc(n);base=n.removesuffix('inventory.json')
  if v['paidVendorCalls']!=0 or len(v['files'])>10000 or v['bytes']>96<<20:raise ValueError('controlled artifact scope differs')
  for leaf,pin in v['files'].items():
   b.relative(leaf)
   if leaf=='.env' or pins.get(base+leaf)!=pin:raise ValueError('controlled artifact pin differs')
  inventories.append(dict(test=n.split('/')[2],files=len(v['files']),bytes=v['bytes']))
 if len(inventories)!=8:raise ValueError('eight public CLI artifacts required')
 core=prefix+'test_core_new_approval_one_shot_and_honest_failed_phase_settlement/';journal=doc(core+'inventory.json');events=json.loads(m[core+'public-cli-journal.json']);unknown=[]
 for event in events:
  if event['args'][:2]==['summarize','glm'] and event['exitCode']==0:
   value=json.loads(event['stdout']);unknown.extend(v for v in value['routes'] if v['state']=='UNKNOWN')
 if len(unknown)!=1 or (unknown[0]['conservativeCalls'],unknown[0]['conservativeAttempts'])!=(24,32):raise ValueError('unknown full-cap disclosure differs')
 settled=doc(core+'effects/repair-consumption.json')
 if settled['current']['calls']!=101 or settled['current']['attempts']!=101 or settled['total']['calls']!=646 or settled['total']['attempts']!=646 or len(settled['routes'])!=6 or sum(v['status']=='FAIL' for v in settled['routes'])!=1:raise ValueError('honest failed GLM/effects simulation differs')
 return dict(schema=1,codeCommit=CODE,measuredHEAD=measured['measuredHEAD'],workingTreeAtMeasurement=True,sourceFiles=len(source['files']),commands=commands,tests=71,publicCLIArtifacts=inventories,originalClosedReceipts=39,originalKnownCalls=545,originalKnownAttempts=545,plannedNewJobs=6,additionalMaxCalls=108,additionalMaxAttempts=132,combinedMaxCalls=653,combinedMaxAttempts=677,syntheticSelectedCalls=101,syntheticSelectedAttempts=101,syntheticSelectedCredentialsRead=True,realCredentialsRead=False,vendorCalls=0,newPaidAuthority='PENDING',releaseCertified=False)

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=p.parse_args();b=helper()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 ix=b.doc(b.read(BASE,'index.json',1<<20));cap=ix['capsule']
 if ix['codeCommit']!=CODE or cap['sourceCommit']!=CODE or cap['id']!='fixedSixKeylessImplementation' or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-implementation/release-repair-raw.tar.gz' or cap['archive']['prefix']!='release-repair-implementation' or ix['newPaidAuthority']!='PENDING':raise ValueError('fixed capsule identity differs')
 pins=b.inventory(cap)
 if len(cap['observe'])>512 or len(set(cap['observe']))!=len(cap['observe']) or not set(cap['observe'])<=set(pins) or any(pins[n]['bytes']>1<<20 for n in cap['observe']) or sum(pins[n]['bytes'] for n in cap['observe'])>16<<20:raise ValueError('bounded observation differs')
 result,m=b.verify_archive(a.common_root,cap);summary=analyse(b,ix,m)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent observed summary differs')
 negative=b.doc(b.read(BASE,'negative-results.json',1<<20));rows=negative.get('negativeGuards',[])
 if negative.get('verified') is not True or negative.get('originalArchivesMutated') is not False or len(rows)!=14 or len({r['case'] for r in rows})!=14 or any(type(r.get('exitCode')) is not int or r['exitCode']==0 or not r.get('reason','').startswith('REFUSED:') for r in rows):raise ValueError('actual negative guard evidence differs')
 print(json.dumps(dict(verified=True,codeCommit=CODE,archive=result,negativeGuards=14,realCredentialsRead=False,vendorCalls=0,newPaidAuthority='PENDING',releaseCertified=False),indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
