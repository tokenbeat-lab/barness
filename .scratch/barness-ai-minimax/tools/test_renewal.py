#!/usr/bin/env python3
"""Public executable tests for keyless renewal; all captures are synthetic."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

DRAFT=Path(__file__).resolve().parent
REAL=DRAFT

class Renewal(unittest.TestCase):
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory(prefix='minimax-renewal-test-');self.root=Path(self.tmp.name).resolve()
  self.tools=self.root/'.scratch/barness-ai-minimax/tools';shutil.copytree(REAL,self.tools,ignore=shutil.ignore_patterns('__pycache__'))
  for p in DRAFT.glob('renewal*.py'):shutil.copy(p,self.tools/p.name)
  for p in ('ai/release','ai/live/testdata/glm-formal','ai/live/testdata/glm-candidate','ai/live/testdata/minimax-formal','ai/internal/testkit/supportmatrix','ai/examples/chineseeval/testdata'):(self.root/p).mkdir(parents=True)
  self.write('go.mod',b'module controlled\n');self.write('go.sum',b'')
  self.sources=b'[]\n';self.pin=hashlib.sha256(self.sources).hexdigest()
  self.write('ai/live/testdata/glm-candidate/sources.json',self.sources)
  self.write('ai/internal/testkit/supportmatrix/glm.go',f'package supportmatrix\nconst GLMSourcePin = "{self.pin}"\n'.encode())
  self.snapshot={'hash':'sha256:current','catalog':{'version':'current'}}
  self.put('ai/release/catalog-snapshot.json',self.snapshot)
  self.put('ai/live/support-matrix.json',{'schema':1,'rows':[]})
  self.put('ai/live/testdata/minimax-formal/origins.json',{'schema':1,'entries':{}})
  for r in ('jev','luna'):
   self.put(f'ai/examples/chineseeval/testdata/{r}-route-v1.json',{'catalogHash':'old','catalogVersion':'old','maxCalls':24,'maxAttempts':24,'maxQuestions':24,'maxRetries':0,'callTimeoutSeconds':15,'timeoutSeconds':600,'instructions':'fixed Chinese instructions','datasetHash':'fixed dataset','route':{'model':'jev-1.13.0' if r=='jev' else 'gpt-6-luna'}})
  self.put('ai/examples/chineseeval/testdata/jev-1.13.0-v1.json',{'catalogHash':'old','catalogVersion':'old','maxCalls':24,'maxAttempts':24,'maxQuestions':24,'maxRetries':0,'callTimeoutSeconds':15,'timeoutSeconds':180,'instructions':'legacy fixed Chinese instructions','datasetHash':'legacy dataset'})
  self.put('ai/e2e/testdata/responses/usage.json',{'catalog':{'version':'old','hash':'old'},'scenarios':[{'id':'preserved','usage':{'output':7}}]})
  for n in ('live-captures.json','known-usage.json','support-views.json'):self.put('ai/live/testdata/glm-formal/'+n,{'historical':n})
  self.write('ai/live/testdata/glm-formal/live-captures.sha256',b'historical-pin\n')
  self.bin=self.root/'fake-bin';self.bin.mkdir()
  self.write('fake-bin/go',('#!'+sys.executable+'\nimport sys,os,json,pathlib,shutil\na=sys.argv[1:]\n'
   'if "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);shutil.copy(pathlib.Path(__file__).with_name("producer"),p);p.chmod(0o755)\n'
   'elif "-write-snapshot" in a:\n p=pathlib.Path("ai/release/catalog-snapshot.json");p.write_text(json.dumps({"hash":"sha256:new","catalog":{"version":"new"}}))\n'
   'elif "./ai/live/cmd/glmformal" in a:\n p=pathlib.Path(a[a.index("-out")+1]);p.mkdir();(p/"live-report.json").write_text("{}");(p/"original"/"case").mkdir(parents=True);(p/"original"/"case"/"result.json").write_text("{}")\n'
   'print(json.dumps({"environmentKeys":sorted(os.environ)}))\n').encode());(self.bin/'go').chmod(0o755)
  self.write('fake-bin/producer',('#!'+sys.executable+'\nimport os,json,pathlib,subprocess\nr=os.environ["BARNESS_MINIMAX_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\n'
   'd={"outcome":"NOT_RUN","complete":False,"mode":"live","stage":"candidate","identity":{"route":r,"model":"MiniMax-M3.1-Flash-Preview","purpose":"mplan","api":{"chat":"openai-completions","responses":"openai-responses","messages":"anthropic-messages"}[r],"endpoint":"https://api.minimax.io/"+("anthropic" if r=="messages" else "v1")},"budget":{"calls":0,"reservedAttempts":0,"attempts":0,"requestedOutput":0,"reservedInput":0},"gitCommit":subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip()}\n'
   '(p/"minimax-report.json").write_text(json.dumps(d));print(json.dumps({"environmentKeys":sorted(os.environ)}))\n').encode())
  self.env={'PATH':str(self.bin)+os.pathsep+os.environ['PATH'],'HOME':str(self.root/'home'),'OPENAI_API_KEY':'FORBIDDEN-AMBIENT','HTTP_PROXY':'http://forbidden.invalid'}
  (self.root/'.env').mkdir()
  self.git('init','-q');self.commit()
  self.cli=[sys.executable,str(self.tools/'renewal.py')]
 def tearDown(self):self.tmp.cleanup()
 def write(self,name,raw):p=self.root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(raw)
 def put(self,name,value):self.write(name,(json.dumps(value,indent=2)+'\n').encode())
 def git(self,*args):return subprocess.check_output(['git',*args],cwd=self.root,env=self.env,stderr=subprocess.DEVNULL).decode().strip()
 def commit(self):self.git('add','.');self.git('-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','controlled source')
 def runcli(self,*args,ok=True):
  p=subprocess.run([*self.cli,*args],cwd=self.root,env=self.env,text=True,capture_output=True,timeout=15)
  if ok:self.assertEqual(p.returncode,0,p.stderr[-700:]+p.stdout[:200])
  else:self.assertNotEqual(p.returncode,0,p.stderr[-700:]+p.stdout[:200])
  return p
 def test_catalog_pins_preserve_fixed_business_contract_without_credentials(self):
  paths=[self.root/f'ai/examples/chineseeval/testdata/{r}-route-v1.json' for r in ('jev','luna')]+[self.root/'ai/examples/chineseeval/testdata/jev-1.13.0-v1.json']
  usage=self.root/'ai/e2e/testdata/responses/usage.json';usage_before=json.loads(usage.read_text())
  before=[json.loads(p.read_text()) for p in paths]
  self.runcli('catalog-pins','.evidence/minimax-renewal-pins')
  for p,old in zip(paths,before):
   new=json.loads(p.read_text());self.assertEqual((new.pop('catalogHash'),new.pop('catalogVersion')),('sha256:new','new'))
   old.pop('catalogHash');old.pop('catalogVersion');self.assertEqual(new,old)
  usage_after=json.loads(usage.read_text());self.assertEqual(usage_after.pop('catalog'),{'version':'new','hash':'sha256:new'})
  usage_before.pop('catalog');self.assertEqual(usage_after,usage_before)
  log=(self.root/'.evidence/minimax-renewal-pins/catalog-snapshot.log').read_text()
  self.assertNotIn('OPENAI_API_KEY',log);self.assertNotIn('HTTP_PROXY',log)

 def captures(self):
  commit=self.git('rev-parse','HEAD');result={}
  for route in ('chat','responses','messages'):
   base=f'raw/{route}/session';scenarios=[]
   for n in range(22):
    case=f'glm-5.{3 if n<11 else 2}/case-{n}';scenarios.append({'id':case})
    usage={'input':2,'cacheRead':3,'cacheWrite':1,'output':5,'reasoning':10000} if n<21 else {'input':0,'cacheRead':0,'cacheWrite':0,'output':0}
    self.put(base+'/GLM11-'+case.replace('/','-')+'/attempt-1-result.json',{'result':{'metadata':{'attempts':[{'usage':usage,'usageReporting':'complete' if n<21 else 'unreported'}]}}})
   self.put(base+'/formal-glm-report.json',{'schema':1,'route':route,'mode':'live','kind':'glm-formal','outcome':'PASS','complete':True,'auditPassed':True,'sourceManifestHash':self.pin,'catalogHash':'sha256:current','catalogVersion':'current','gitCommit':commit,'scenarios':scenarios})
   self.put(base+'/live-report.json',{'raw':'unchanged'})
   self.put(base+'/glm-execution-source.json',{'controlled-source':'frozen'})
   self.put(base+'/glm-integrity.json',{'schema':1,'files':{'formal-glm-report.json':hashlib.sha256((self.root/base/'formal-glm-report.json').read_bytes()).hexdigest()}})
   pins={'schema':1}
   for field,leaf in [('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')]:pins[field]=hashlib.sha256((self.root/base/leaf).read_bytes()).hexdigest()
   self.put('.evidence/minimax-launches/glm-'+route+'.json',{'state':'COMPLETED','route':'glm-'+route,'sourceCommit':commit,'exitCode':0,'credentialsRead':True,'launcherAudit':{'passed':True},'capturePins':pins})
   result[route]='raw/'+route
  return result
 def archive(self,out='.evidence/minimax-renewal-glm',ok=True):
  values=self.captures();args=['glm-archive',out]
  for route,value in values.items():args+=['--'+route,value]
  return self.runcli(*args,ok=ok)
 def test_fresh_raw_archives_and_independent_attempt_sums_then_committed_views(self):
  old=self.root/'historical.zip';old.write_bytes(b'IMMUTABLE-HISTORY')
  self.archive()
  index=json.loads((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_text())
  usage=json.loads((self.root/'ai/live/testdata/glm-formal/known-usage.json').read_text())
  self.assertEqual(old.read_bytes(),b'IMMUTABLE-HISTORY')
  for route in ('chat','responses','messages'):
   row=usage['routes'][route];self.assertEqual((row['inputSide'],row['output'],row['unknownUsageAttempts'],row['actualAttempts']),(126,105,1,22))
   import zipfile
   with zipfile.ZipFile(self.root/index['routes'][route]['archive']) as z:
    original=(self.root/'raw'/route/'session/formal-glm-report.json').read_bytes()
    self.assertEqual(z.read(index['routes'][route]['reportPath']),original)
  self.runcli('glm-views','.evidence/minimax-renewal-glm',ok=False)
  self.commit();self.runcli('glm-views','.evidence/minimax-renewal-glm')
  views=json.loads((self.root/'ai/live/testdata/glm-formal/support-views.json').read_text())
  self.assertEqual(set(views['routes']),{'chat','responses','messages'})
  self.assertTrue(all(v['reportPath']=='live-report.json' for v in views['routes'].values()))

 def test_corrective_round_requires_explicit_own_canonical_pins_no_original_fallback(self):
  values=self.captures()
  before={n:(self.root/'ai/live/testdata/glm-formal'/n).read_bytes() for n in ('live-captures.json','known-usage.json')}
  args=['glm-archive','.evidence/minimax-corrective-missing','--round','corrective-1']
  for route,value in values.items():args+=['--'+route,value]
  self.runcli(*args,ok=False)
  for route in ('chat','responses','messages'):
   old=self.root/'.evidence/minimax-launches'/('glm-'+route+'.json')
   value=json.loads(old.read_text());value.update(route='glm-corrective-1-'+route,correctiveRound='corrective-1',humanApprovalSha256='a'*64)
   self.put('.evidence/minimax-launches/glm-corrective-1-'+route+'.json',value)
  args[1]='.evidence/minimax-corrective-explicit';self.runcli(*args)
  index=json.loads((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_text())
  self.assertEqual(index['collectionRound'],'corrective-1')
  self.assertEqual(set(index['launcherReceiptSHA256']),{'chat','responses','messages'})
  # Borrowing one old receipt into the new round is never a fallback.
  mixed=self.root/'.evidence/minimax-launches/glm-corrective-1-messages.json';mixed.write_bytes((self.root/'.evidence/minimax-launches/glm-messages.json').read_bytes())
  fixed=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  args[1]='.evidence/minimax-corrective-mixed';self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),fixed)
 def test_routed_round_requires_own_reviewed_direct_capture_receipts(self):
  values=self.captures();args=['glm-archive','.evidence/minimax-routed-missing','--round','routed-validation-1']
  for route,value in values.items():args+=['--'+route,value]
  before=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),before)
  for route in ('chat','responses','messages'):
   value=json.loads((self.root/'.evidence/minimax-launches'/('glm-'+route+'.json')).read_text())
   value.update(route='routed-validation-1-glm-'+route,baseName='glm-'+route,
    validationRound='routed-validation-1',phase='glm',networkRoute='china-direct',maxCalls=24,maxAttempts=32,
    automaticReruns=False,budget={'state':'REPORTED','status':'PASS','calls':22,'attempts':22})
   value.update({field:'a'*64 for field in ('authorizationSha256','sourceReviewSha256','sourceManifestSha256','binariesManifestSha256','planSha256','commandReceiptSha256')})
   self.put('.evidence/minimax-launches/routed-validation-1-glm-'+route+'.json',value)
  args[1]='.evidence/minimax-routed-complete';self.runcli(*args)
  fixed=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  self.assertEqual(json.loads(fixed)['collectionRound'],'routed-validation-1')
  path=self.root/'.evidence/minimax-launches/routed-validation-1-glm-responses.json';original=path.read_bytes()
  for field,bad in (('networkRoute','international-loopback-proxy'),('authorizationSha256',''),('validationRound','corrective-1'),('budget',{'state':'REPORTED','status':'FAIL','calls':22,'attempts':22})):
   value=json.loads(original);value[field]=bad;path.write_text(json.dumps(value))
   args[1]='.evidence/minimax-routed-bad-'+field;self.runcli(*args,ok=False)
   self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),fixed)
   path.write_bytes(original)

 def repair_receipts(self,values):
  # Synthetic approved historical metadata at the public renewal boundary;
  # no launcher or key acquisition is performed by these fixtures.
  code="""
import sys,shutil
from pathlib import Path
sys.path.insert(0,'.scratch/barness-ai-minimax/tools')
import paid_core as c,repair_contract as r,routed_evidence as e,paid_checks
out=c.ROOT/'repair-own';out.mkdir();jobs=r.jobs(out,'glm');commit=c.git('rev-parse','HEAD')
for route in ('chat','responses','messages'):
 source=c.ROOT/'raw'/route/'session';target=out/('glm-'+route)/'session';shutil.copytree(source,target)
 value=c.doc(target/'formal-glm-report.json');value['budget']=dict(callsUsed=22,httpAttempts=22,attemptsReserved=22);c.dump(target/'formal-glm-report.json',value)
 seal=c.doc(target/'glm-integrity.json');seal['files']['formal-glm-report.json']=c.digest(target/'formal-glm-report.json');c.dump(target/'glm-integrity.json',seal)
for leaf in r.EXTRA_PINS:c.dump(out/leaf,{})
c.dump(out/'source-at-start.json',dict(sourceCommit=commit,files={}));c.dump(out/'binaries.json',{})
c.dump(out/'plan.json',c.redact(dict(round=r.ROUND,phase='glm',budget=r.BUDGET,collection=e.location(out),routes=jobs)))
pins=dict(sourceCommit=commit,sourceManifestSHA256=c.digest(out/'source-at-start.json'),binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})
checks=dict(passed=True,commands={},commandReceipts={},credentialsRead=False)
for name,argv,_,_ in paid_checks.commands(out):
 c.dump(out/(name+'-command.json'),dict(name=name,argv=argv,cwd='$REPO',state='COMPLETED',exitCode=0,credentialsRead=False,credentialsInherited=False,sourceVariable=None,injectedCredential=None,launcherAudit=dict(passed=True,findings=[])))
 checks['commands'][name]=0;checks['commandReceipts'][name]=c.digest(out/(name+'-command.json'))
c.dump(out/'checks.json',checks)
c.dump(out/'prerelease-source-review.json',dict(decision='APPROVED',reviewedBy='synthetic independent review',**pins,authorizedRoutes=list(jobs),credentialScopes={'GLM_KEY':r.ROUND+'/china-direct'}))
c.dump(out/'execution-authorization.json',dict(decision='APPROVED',authority='explicit-additional-budget-reply',requestReference='synthetic new108/132 question/reply',requestText='approve NEW108/132',approvalProposalSHA256=r.PROPOSAL_SHA256,approvedAdditionalMaxCalls=108,approvedAdditionalMaxAttempts=132,round=r.ROUND,phase='glm',budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,**pins,checksSHA256=c.digest(out/'checks.json'),sourceReviewSHA256=c.digest(out/'prerelease-source-review.json')))
for name,job in jobs.items():
 route=job['glmRoute'];base=job['baseName'];bundle=out/base/'session';planned=c.doc(out/'plan.json')['routes'][name]
 c.dump(out/(base+'-command.json'),dict(name=base,argv=planned['argv'],cwd=planned['cwd'],state='COMPLETED',exitCode=0,credentialsRead=True,launcherAudit=dict(passed=True)))
 capture=dict(schema=1,**{field:c.digest(bundle/leaf) for field,leaf in [('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')]})
 c.dump(c.common_root()/'.evidence/minimax-launches'/(name+'.json'),c.redact(dict(state='COMPLETED',route=name,baseName=base,validationRound=r.ROUND,phase='glm',sourceCommit=commit,sourceVariable='GLM_KEY',maxCalls=24,maxAttempts=32,networkRoute='china-direct',automaticReruns=False,credentialsRead=True,exitCode=0,launcherAudit=dict(passed=True),collectionRoot=str(out),collectionOwnerSHA256=e.location(out)['rootSHA256'],sourceManifestSha256=pins['sourceManifestSHA256'],binariesManifestSha256=pins['binariesManifestSHA256'],planSha256=pins['planSHA256'],sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),authorizationSha256=c.digest(out/'execution-authorization.json'),commandReceiptSha256=c.digest(out/(base+'-command.json')),reportedFileSha256=c.digest(bundle/'formal-glm-report.json'),capturePins=capture,budget=dict(state='REPORTED',status='PASS',calls=22,attempts=22),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})))
"""
  subprocess.run([sys.executable,'-B','-c',code],cwd=self.root,env=self.env,check=True,capture_output=True,timeout=15)
  return {r:'repair-own/glm-'+r for r in values}

 def test_release_repair_round_has_no_old_fallback_or_weaker_publication(self):
  values=self.repair_receipts(self.captures());args=['glm-archive','.evidence/minimax-release-repair-raw','--round','release-repair-validation-1']
  for route,value in values.items():args+=['--'+route,value]
  self.runcli(*args)
  fixed=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes();self.assertEqual(json.loads(fixed)['collectionRound'],'release-repair-validation-1')
  receipt=self.root/'.evidence/minimax-launches/release-repair-validation-1-glm-responses.json';original=receipt.read_bytes()
  for field,bad in [('phase','legacy'),('authorizationSha256','0'*64),('sourceCommit','f'*40),('validationRound','routed-validation-1')]:
   value=json.loads(original);value[field]=bad;receipt.write_text(json.dumps(value));args[1]='.evidence/minimax-release-repair-bad-'+field;self.runcli(*args,ok=False)
   self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),fixed);receipt.write_bytes(original)
  report=self.root/'repair-own/glm-responses/session/formal-glm-report.json';value=json.loads(report.read_text());value['outcome']='FAIL';report.write_text(json.dumps(value));args[1]='.evidence/minimax-release-repair-failed';self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),fixed)

 def test_release_repair_views_reopen_original_canonical_and_usage_bytes(self):
  code="""import sys,json,pathlib,zipfile,hashlib
args=sys.argv[1:];root=pathlib.Path.cwd();base=root/'ai/live/testdata/glm-formal';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
if './ai/live/cmd/glmformal' in args:
 route=args[args.index('-route')+1];out=pathlib.Path(args[args.index('-out')+1]);out.mkdir();index=json.loads((base/'live-captures.json').read_text());row=index['routes'][route]
 with zipfile.ZipFile(root/row['archive']) as archive:
  for name in archive.namelist():
   target=out/'original'/name.partition('/')[2];target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(archive.read(name))
 proof=dict(schema=1,kind='glm-derived-live-support',route=route,rawCaptureIndexSHA256=sha(base/'live-captures.json'),rawArchiveSHA256=row['archiveSHA256'],originalFormalReportSHA256=sha(out/'original/formal-glm-report.json'),originalSupportReportSHA256=sha(out/'original/live-report.json'),originalGitCommit=index['originalLiveSourceCommit'])
 if (root/'bad-derivation').exists():proof['rawArchiveSHA256']='0'*64
 (out/'glm-support-derivation.json').write_text(json.dumps(proof));(out/'live-report.json').write_text('{}')
"""
  self.write('fake-bin/go',('#!'+sys.executable+'\n'+code).encode());(self.bin/'go').chmod(0o755)
  values=self.repair_receipts(self.captures());out='.evidence/minimax-repair-views';args=['glm-archive',out,'--round','release-repair-validation-1']
  for route,value in values.items():args+=['--'+route,value]
  self.runcli(*args);seed=self.root/'view-seed';shutil.copytree(self.root/out,seed);self.commit();self.runcli('glm-views',out)
  target=self.root/'ai/live/testdata/glm-formal/support-views.json';fixed=target.read_bytes();self.assertEqual(set(json.loads(fixed)['routes']),{'chat','responses','messages'})
  usage=self.root/'ai/live/testdata/glm-formal/known-usage.json';original=usage.read_bytes()
  for mode in ('raw-payload','known-usage','derivation'):
   with self.subTest(invalid=mode):
    clone='.evidence/minimax-repair-bad-view-'+mode;shutil.copytree(seed,self.root/clone)
    if mode=='known-usage':
     value=json.loads(original);value['routes']['chat']['output']+=1;usage.write_text(json.dumps(value))
     state=json.loads((self.root/clone/'raw-published.json').read_text());state['selectorSHA256']['ai/live/testdata/glm-formal/known-usage.json']=hashlib.sha256(usage.read_bytes()).hexdigest();self.put(clone+'/raw-published.json',state);self.commit()
    elif mode=='raw-payload':
     import io,zipfile
     index_path=self.root/'ai/live/testdata/glm-formal/live-captures.json';index_raw=index_path.read_bytes();index=json.loads(index_raw);archive=self.root/index['routes']['chat']['archive'];archive_raw=archive.read_bytes()
     with zipfile.ZipFile(io.BytesIO(archive_raw)) as z:files={n:z.read(n) for n in z.namelist()}
     leaf=next(n for n in files if n.endswith('attempt-1-result.json'));files[leaf]+=b' '
     with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED) as z:
      for n,raw in files.items():
       import stat
       info=zipfile.ZipInfo(n);info.external_attr=(stat.S_IFREG|0o600)<<16;z.writestr(info,raw)
     digest=hashlib.sha256(archive.read_bytes()).hexdigest();index['routes']['chat']['archiveSHA256']=digest;index_path.write_text(json.dumps(index));(index_path.parent/'live-captures.sha256').write_text(hashlib.sha256(index_path.read_bytes()).hexdigest()+'\n')
     value=json.loads(original);value['routes']['chat']['archiveSHA256']=digest;usage.write_text(json.dumps(value));state=json.loads((self.root/clone/'raw-published.json').read_text());state['selectorSHA256']={n:hashlib.sha256((self.root/n).read_bytes()).hexdigest() for n in state['selectorSHA256']};self.put(clone+'/raw-published.json',state);self.commit()
    else:(self.root/'bad-derivation').touch()
    self.runcli('glm-views',clone,ok=False);self.assertEqual(target.read_bytes(),fixed)
    usage.write_bytes(original)
    if mode=='raw-payload':
     index_path.write_bytes(index_raw);(index_path.parent/'live-captures.sha256').write_text(hashlib.sha256(index_raw).hexdigest()+'\n');archive.write_bytes(archive_raw)
    self.commit()

 def test_raw_capture_symlink_and_mixed_source_refuse_without_selector_mutation(self):
  values=self.captures();target=self.root/'raw/chat/session/link';target.symlink_to(self.root/'go.mod')
  original=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  args=['glm-archive','.evidence/minimax-renewal-invalid']
  for route,value in values.items():args+=['--'+route,value]
  self.runcli(*args,ok=False);self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)
  target.unlink();p=self.root/'raw/messages/session/formal-glm-report.json';r=json.loads(p.read_text());r['gitCommit']='f'*40;p.write_text(json.dumps(r))
  args[1]='.evidence/minimax-renewal-mixed';self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)
 def test_origin_index_only_accepts_successful_canonical_receipt(self):
  base='own-mm/session';identity={'route':'chat','model':'MiniMax-M3','purpose':'standard','api':'openai-completions','endpoint':'https://api.minimax.io/v1'}
  commit=self.git('rev-parse','HEAD')
  self.put(base+'/minimax-report.json',{'mode':'live','outcome':'PASS','complete':True,'stage':'candidate','identity':identity,'session':'session','gitCommit':commit})
  self.put(base+'/minimax-integrity.json',{'fixture':'seal'});self.put(base+'/minimax-ledger.json',{'fixture':'ledger'})
  origin={'schema':1,'stage':'candidate','identity':identity,'session':'session','sourceCommit':commit,'sourceManifestSha256':'a'*64,'binariesManifestSha256':'b'*64,'planSha256':'c'*64}
  for key,name in [('reportSha256','minimax-report.json'),('sealSha256','minimax-integrity.json'),('ledgerSha256','minimax-ledger.json')]:origin[key]=hashlib.sha256((self.root/base/name).read_bytes()).hexdigest()
  receipt={'state':'COMPLETED','exitCode':0,'credentialsRead':True,'route':'mm-global-candidate-m3-chat','sourceCommit':commit,'launcherAudit':{'passed':True},'origin':origin}
  # A bundle-local receipt is never accepted as actual-origin authority.
  self.put(base+'/receipt.json',receipt)
  self.runcli('origins','.evidence/minimax-origin-invalid','--bundle','mm-global-candidate-m3-chat='+base,ok=False)
  path='.evidence/minimax-launches/mm-global-candidate-m3-chat.json';self.put(path,receipt)
  self.runcli('origins','.evidence/minimax-origin-current','--bundle','mm-global-candidate-m3-chat='+base)
  index=json.loads((self.root/'ai/live/testdata/minimax-formal/origins.json').read_text())
  entry=index['entries'][origin['reportSha256']]
  self.assertEqual(entry['origin'],origin)
  self.assertEqual(entry['launcherReceiptSha256'],hashlib.sha256((self.root/path).read_bytes()).hexdigest())
  receipt['credentialsRead']=False;self.put(path,receipt)
  self.runcli('origins','.evidence/minimax-origin-failed','--bundle','mm-global-candidate-m3-chat='+base,ok=False)
 def test_global_origin_names_cannot_borrow_cn_identity_or_old_name(self):
  base='cn-own/session';identity={'route':'chat','model':'MiniMax-M3','purpose':'standard','api':'openai-completions','endpoint':'https://api.minimax.cn/v1'}
  commit=self.git('rev-parse','HEAD')
  self.put(base+'/minimax-report.json',{'mode':'live','outcome':'PASS','complete':True,'stage':'candidate','identity':identity,'session':'session','gitCommit':commit})
  self.put(base+'/minimax-integrity.json',{});self.put(base+'/minimax-ledger.json',{})
  origin={'schema':1,'stage':'candidate','identity':identity,'session':'session','sourceCommit':commit,'sourceManifestSha256':'a'*64,'binariesManifestSha256':'b'*64,'planSha256':'c'*64}
  for key,name in [('reportSha256','minimax-report.json'),('sealSha256','minimax-integrity.json'),('ledgerSha256','minimax-ledger.json')]:origin[key]=hashlib.sha256((self.root/base/name).read_bytes()).hexdigest()
  receipt={'state':'COMPLETED','exitCode':0,'credentialsRead':True,'route':'mm-global-candidate-m3-chat','sourceCommit':commit,'launcherAudit':{'passed':True},'origin':origin}
  self.put('.evidence/minimax-launches/mm-global-candidate-m3-chat.json',receipt)
  before=(self.root/'ai/live/testdata/minimax-formal/origins.json').read_bytes()
  self.runcli('origins','.evidence/minimax-cn-spoof','--bundle','mm-global-candidate-m3-chat='+base,ok=False)
  self.put('.evidence/minimax-launches/mm-candidate-m3-chat.json',receipt)
  self.runcli('origins','.evidence/minimax-old-name','--bundle','mm-candidate-m3-chat='+base,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/minimax-formal/origins.json').read_bytes(),before)

 def test_three_not_run_processes_never_acquire_key_or_paid_intents(self):
  command=[sys.executable,str(self.tools/'paid.py'),'prepare','prepared']
  subprocess.run(command,cwd=self.root,env=self.env,check=True,capture_output=True,timeout=15)
  self.runcli('m31-not-run','prepared')
  summary=json.loads((self.root/'prepared/m31-not-run-summary.json').read_text())
  self.assertEqual([r['status'] for r in summary['routes']],['NOT_RUN']*3)
  self.assertFalse(summary['credentialsRead']);self.assertFalse((self.root/'.evidence/minimax-launches').exists())
  self.assertFalse((self.root/'.evidence/minimax-stages/candidate.jsonl').exists())
  for route in ('chat','responses','messages'):
   text=(self.root/f'prepared/m31-not-run-{route}.log').read_text()
   self.assertNotIn('OPENAI_API_KEY',text);self.assertNotIn('HTTP_PROXY',text)
   self.assertNotIn('BARNESS_MINIMAX_KEY',text);self.assertNotIn('BARNESS_MINIMAX_ENTITLEMENT',text)

 def test_keyless_not_run_cannot_report_archived_cn_identity(self):
  producer=self.bin/'producer';producer.write_text(producer.read_text().replace('https://api.minimax.io/','https://api.minimax.cn/'))
  subprocess.run([sys.executable,str(self.tools/'paid.py'),'prepare','prepared'],cwd=self.root,env=self.env,check=True,capture_output=True,timeout=15)
  self.runcli('m31-not-run','prepared',ok=False)
  self.assertFalse((self.root/'.evidence/minimax-launches').exists())
  self.assertFalse((self.root/'.evidence/minimax-stages/candidate.jsonl').exists())

 def test_empty_directory_fanout_and_depth_refuse_before_selector_writes(self):
  values=self.captures();base=self.root/'raw/chat/session'
  original=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  for n in range(1025):(base/f'empty-{n}').mkdir()
  args=['glm-archive','.evidence/minimax-empty-fanout']
  for route,value in values.items():args+=['--'+route,value]
  self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)
  for n in range(1025):(base/f'empty-{n}').rmdir()
  (base/'nested/a/b/c').mkdir(parents=True)
  args[1]='.evidence/minimax-excess-depth';self.runcli(*args,ok=False)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)

 def test_capture_file_count_per_file_and_expanded_bounds(self):
  values=self.captures();original=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  base=self.root/'raw/chat/session'
  for mode in ('count','file','total'):
   extras=[]
   if mode=='count':
    for n in range(513):p=base/f'extra-{n}';p.write_bytes(b'');extras.append(p)
   else:
    for n in range(1 if mode=='file' else 9):
     p=base/f'extra-{n}';extras.append(p)
     with p.open('wb') as stream:stream.truncate((16<<20)+1 if mode=='file' else 15<<20)
   args=['glm-archive','.evidence/minimax-bound-'+mode]
   for route,value in values.items():args+=['--'+route,value]
   self.runcli(*args,ok=False)
   self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)
   for p in extras:p.unlink()

 def test_coherent_raw_edits_cannot_borrow_canonical_successful_receipt(self):
  values=self.captures();base=self.root/'raw/chat/session'
  selectors=['live-captures.json','live-captures.sha256','known-usage.json']
  before={n:(self.root/'ai/live/testdata/glm-formal'/n).read_bytes() for n in selectors}
  receipt=self.root/'.evidence/minimax-launches/glm-chat.json'
  names=['formal-glm-report.json','glm-integrity.json','live-report.json','glm-execution-source.json']
  originals={n:(base/n).read_bytes() for n in names};old_receipt=receipt.read_bytes()
  for mutation in ('coherent-report-seal','missing-pins','support-report','execution-source'):
   for n,raw in originals.items():(base/n).write_bytes(raw)
   receipt.write_bytes(old_receipt)
   if mutation=='coherent-report-seal':
    formal=json.loads(originals['formal-glm-report.json']);formal['accountAlias']='edited-without-new-process';(base/'formal-glm-report.json').write_text(json.dumps(formal))
    support=json.loads(originals['live-report.json']);support['accountAlias']=formal['accountAlias'];(base/'live-report.json').write_text(json.dumps(support))
    seal=json.loads(originals['glm-integrity.json']);seal['files']['formal-glm-report.json']=hashlib.sha256((base/'formal-glm-report.json').read_bytes()).hexdigest();seal['files']['live-report.json']=hashlib.sha256((base/'live-report.json').read_bytes()).hexdigest();(base/'glm-integrity.json').write_text(json.dumps(seal))
   elif mutation=='missing-pins':
    value=json.loads(old_receipt);value.pop('capturePins');receipt.write_text(json.dumps(value))
   elif mutation=='support-report':(base/'live-report.json').write_text('{"raw":"changed"}')
   else:(base/'glm-execution-source.json').write_text('{"controlled-source":"edited"}')
   args=['glm-archive','.evidence/minimax-borrow-'+mutation]
   for route,value in values.items():args+=['--'+route,value]
   self.runcli(*args,ok=False)
   for name,raw in before.items():self.assertEqual((self.root/'ai/live/testdata/glm-formal'/name).read_bytes(),raw)

 def test_bundle_lookup_fanout_is_bounded_before_capture_reads(self):
  values=self.captures();parent=self.root/'raw/chat'
  original=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
  for n in range(1025):(parent/f'empty-{n}').mkdir()
  args=['glm-archive','.evidence/minimax-bundle-fanout']
  for route,value in values.items():args+=['--'+route,value]
  result=self.runcli(*args,ok=False)
  self.assertIn('bundle lookup entries exceed1024',result.stderr)
  self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),original)

if __name__=='__main__':unittest.main(verbosity=2)
