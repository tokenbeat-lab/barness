"""Public fixed-eleven tool E2E; external executables and keys are synthetic."""
import json,os,subprocess,sys,unittest
from pathlib import Path
import test_repair_fixtures as f

class Eleven(f.Isolated,unittest.TestCase):
 SOURCE_REF=None
 INTENT_GLOB="current-source-eleven-renewal-1-*.json"
 EXTRA_ARTIFACTS=(".evidence/minimax-renewals/current-source-eleven-renewal-1",)
 def setUp(self):
  super().setUp()
  source=f.TOOLS.parents[2]
  common=subprocess.check_output(['git','rev-parse','--git-common-dir'],cwd=source,env=self.env,text=True).strip()
  common=(source/common).resolve().parent
  subprocess.run([sys.executable,'-B',str(self.tools/'test_eleven_validation.py'),'--seed',str(common)],cwd=self.root,env=self.env,check=True)
  self.cli=[sys.executable,'-B',str(self.tools/'eleven_validation.py')]
 def selected_producer(self):
  super().selected_producer()
  producer=self.bin/'producer';code=producer.read_text()
  code=code.replace("argv=sys.argv[1:];route=", "argv=sys.argv[1:]\nif '-renewal-source' in argv:\n import hashlib\n ledger=(root/'.evidence/minimax-stages/formal.jsonl').read_bytes();tail=json.loads(ledger.splitlines()[-1])['hash']\n print(json.dumps(dict(renewalId='current-source-eleven-renewal-1',commonRootSha256=hashlib.sha256(str(root).encode()).hexdigest(),sourceCommit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),executionSourceSha256='a'*64,originalLedgerSha256=hashlib.sha256(ledger).hexdigest(),originalLedgerTailSha256=tail,originalCalls=39,originalAttempts=36)));sys.exit(0)\nroute=")
  code=code.replace("if route:\n for name", "if route:\n value.update(sourceManifestHash=__import__('hashlib').sha256((root/'ai/live/testdata/glm-candidate/sources.json').read_bytes()).hexdigest(),catalogHash=json.loads((root/'ai/release/catalog-snapshot.json').read_text())['hash'],catalogVersion=json.loads((root/'ai/release/catalog-snapshot.json').read_text())['catalog']['version'],scenarios=[dict(id='glm-5.3/case-'+str(i)) for i in range(22)])\n (bundle/leaf).write_text(json.dumps(value))\n for i in range(22):\n  d=bundle/('GLM11-glm-5.3-case-'+str(i));d.mkdir();(d/'attempt-1-result.json').write_text(json.dumps(dict(result=dict(metadata=dict(attempts=[dict(usage=dict(input=1,output=1),usageReporting='complete')])))))\n for name")
  code=code.replace("route=os.environ.get('BARNESS_GLM_ROUTE')", "\nif '-verify' in argv:sys.exit(0)\nif '-formal' in argv and '-out' not in argv:\n bundle=pathlib.Path(argv[-1])\n if '-publish' in argv:(bundle/'live-report.json').write_text('{}')\n sys.exit(0)\nroute=os.environ.get('BARNESS_GLM_ROUTE')")
  code=code.replace("elif 'BARNESS_MM_ROUTE' in os.environ:", "elif 'BARNESS_MINIMAX_ROUTE' in os.environ:\n import hashlib\n assert os.environ.get('BARNESS_MINIMAX_PROXY')=='http://127.0.0.1:7890' and 'BARNESS_AI_LIVE_PROXY' not in os.environ\n m=os.environ['BARNESS_MINIMAX_ROUTE'];bundle=pathlib.Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session';leaf='minimax-report.json';failed=m=='responses' and (root/'.fake-bin/fail-m3-responses').exists()\n grant=root/'.evidence/minimax-renewals/current-source-eleven-renewal-1/grant.json'\n value=dict(mode='live',stage='formal',session='session',outcome='FAIL' if failed else 'PASS',complete=not failed,auditPassed=True,gitCommit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),formalRenewalId='current-source-eleven-renewal-1',formalRenewalGrantSha256=hashlib.sha256(grant.read_bytes()).hexdigest(),identity=dict(route=m,model='MiniMax-M3',purpose='standard',api={'chat':'openai-completions','responses':'openai-responses','messages':'anthropic-messages'}[m],endpoint='https://api.minimax.io/'+('anthropic' if m=='messages' else 'v1')),budget=dict(calls=13,attempts=12,reservedAttempts=12,requestedOutput=20000,reservedInput=20000))\nelif 'BARNESS_MM_ROUTE' in os.environ:")
  code=code.replace("print(json.dumps(dict(os.environ)))", "\nif 'BARNESS_MINIMAX_ROUTE' in os.environ:\n for n in ('minimax-integrity.json','minimax-ledger.json'):(bundle/n).write_text('{}')\nprint(json.dumps(dict(os.environ)))")
  code=code.replace("elif 'BARNESS_MM_ROUTE' in os.environ:", "elif 'BARNESS_AI_CHINESE_EVAL' in os.environ:\n bundle=pathlib.Path(argv[argv.index('-out')+1])/'session';leaf='evaluation-report.json';value=dict(status='COMPLETE',budget=dict(calls=24,attempts=24))\nelif 'BARNESS_MM_ROUTE' in os.environ:")
  producer.write_text(code)
  go=self.bin/'go';code=go.read_text();code+="\nif './ai/live/cmd/glmformal' in a:\n import zipfile,hashlib\n route=a[a.index('-route')+1];out=pathlib.Path(a[a.index('-out')+1]);out.mkdir();base=r/'ai/live/testdata/glm-formal';index=json.loads((base/'live-captures.json').read_text());row=index['routes'][route];sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()\n with zipfile.ZipFile(r/row['archive']) as z:\n  for name in z.namelist():\n   target=out/'original'/name.partition('/')[2];target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(z.read(name))\n proof=dict(schema=1,kind='glm-derived-live-support',route=route,rawCaptureIndexSHA256=sha(base/'live-captures.json'),rawArchiveSHA256=row['archiveSHA256'],originalFormalReportSHA256=sha(out/'original/formal-glm-report.json'),originalSupportReportSHA256=sha(out/'original/live-report.json'),originalGitCommit=index['originalLiveSourceCommit'])\n (out/'glm-support-derivation.json').write_text(json.dumps(proof));(out/'live-report.json').write_text('{}')\n"
  code+="\nif './ai/live/cmd/jevformal' in a and '-out' in a:\n p=pathlib.Path(a[a.index('-out')+1]);p.mkdir();(p/'live-report.json').write_text('{}')\nif './ai/live/cmd/supportmatrix' in a and '-matrix' in a:\n p=pathlib.Path(a[a.index('-matrix')+1]);value=json.loads(p.read_text())\n for row in value['rows']:\n  if row['combo'].startswith(('minimax-','zhipu-')) or row['combo']=='typesafe-classifier-jev-1.14.0':row.update(lastRunAt='2100-01-01T00:00:00Z',allPassedAt='2100-01-01T00:00:00Z')\n p.write_text(json.dumps(value))\n"
  code=code.replace('sys.exit(7 if', 'code=7 if').replace(' else 0)\n',' else 0\n')+'\nsys.exit(code)\n';go.write_text(code)
 def commit_data(self):
  subprocess.run(['git','add','ai/live/testdata/glm-formal'],cwd=self.root,env=self.env,check=True)
  subprocess.run(['git','-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','strict own GLM data only'],cwd=self.root,env=self.env,check=True)
 def complete_glm(self):
  self.selected_producer();self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm');self.approve('glm');self.invoke('authorize','glm')
  (self.root/'.env').write_text('GLM_KEY=SYNTHETIC-GLM\nMINIMAX_KEY=SYNTHETIC-M3\nOPENAI_KEY=SYNTHETIC-OA\nTYPESAFE_KEY=SYNTHETIC-TS\n')
  for route in ('chat','responses','messages'):self.invoke('run-selected','glm','current-source-eleven-renewal-1-glm-'+route)
  argv=['glm-archive','.evidence/minimax-eleven-publication','--round','current-source-eleven-renewal-1']
  for route in ('chat','responses','messages'):argv+=['--'+route,'glm/glm-'+route]
  subprocess.run([sys.executable,'-B',str(self.tools/'renewal.py'),*argv],cwd=self.root,env=self.env,check=True,capture_output=True);self.commit_data()
  subprocess.run([sys.executable,'-B',str(self.tools/'renewal.py'),'glm-views','.evidence/minimax-eleven-publication'],cwd=self.root,env=self.env,check=True,capture_output=True);self.commit_data()
 def test_complete_eleven_public_consumers_promote_only_seven_rows_atomically(self):
  self.complete_glm();self.invoke('prepare','remaining','--phase','remaining');self.invoke('check','remaining');self.approve('remaining');self.invoke('grant','remaining')
  plan=json.loads((self.root/'remaining/plan.json').read_text())
  for name in plan['routes']:self.invoke('run-selected','remaining',name)
  original=json.loads((self.root/'ai/live/support-matrix.json').read_text());result=json.loads(self.invoke('finalize','remaining').stdout)
  self.assertTrue(result['strictElevenCurrent']);self.assertFalse(result['releaseCertified']);self.assertEqual(result['updatedRows'],7)
  current=json.loads((self.root/'ai/live/support-matrix.json').read_text());self.assertEqual(len(current['rows']),17)
  before={r['combo']:r for r in original['rows']};after={r['combo']:r for r in current['rows']}
  for name,row in before.items():
   if not name.startswith(('minimax-','zhipu-')) and name!='typesafe-classifier-jev-1.14.0':self.assertEqual(after[name],row)
  self.invoke('finalize','remaining',ok=False)
 def test_m3_grant_exact_types_are_checked_before_any_intent_or_key(self):
  self.complete_glm();self.invoke('prepare','remaining','--phase','remaining');self.invoke('check','remaining');self.approve('remaining');self.invoke('grant','remaining')
  key=self.root/'.env';key.unlink();key.mkdir();target=self.root/'.evidence/minimax-renewals/current-source-eleven-renewal-1/grant.json';raw=target.read_bytes()
  for field,value in [('maxCalls',60.0),('schema',True),('originalCalls',39.0),('maxAttempts',1),('executionSourceSha256','0'*64)]:
   changed=json.loads(raw);changed[field]=value;target.write_text(json.dumps(changed));self.invoke('run-selected','remaining','current-source-eleven-renewal-1-mm-global-formal-m3-chat',ok=False)
   self.assertFalse((self.root/'.evidence/minimax-launches/current-source-eleven-renewal-1-mm-global-formal-m3-chat.json').exists());self.assertTrue(key.is_dir());target.write_bytes(raw)
 def test_native_m3_success_has_own_origin_and_failed_report_settles_without_projection(self):
  self.complete_glm();self.invoke('prepare','remaining','--phase','remaining');self.invoke('check','remaining');self.approve('remaining');self.invoke('grant','remaining')
  self.invoke('run-selected','remaining','current-source-eleven-renewal-1-mm-global-formal-m3-chat')
  root=self.root/'.evidence/minimax-launches';record=json.loads((root/'current-source-eleven-renewal-1-mm-global-formal-m3-chat.json').read_text())
  self.assertEqual((record['budget']['calls'],record['budget']['attempts']),(13,12));self.assertEqual(record['origin']['formalRenewalId'],'current-source-eleven-renewal-1')
  self.assertTrue((self.root/'remaining/mm-global-formal-m3-chat/session/live-report.json').is_file())
  (self.bin/'fail-m3-responses').touch();self.invoke('run-selected','remaining','current-source-eleven-renewal-1-mm-global-formal-m3-responses',ok=False)
  failed=json.loads((root/'current-source-eleven-renewal-1-mm-global-formal-m3-responses.json').read_text())
  self.assertEqual((failed['budget']['state'],failed['budget']['status'],failed['budget']['calls'],failed['budget']['attempts']),('REPORTED','FAIL',13,12));self.assertNotIn('origin',failed)
  self.assertFalse((self.root/'remaining/mm-global-formal-m3-responses/session/live-report.json').exists())
 def test_known_glm_failure_allows_unchanged_remaining_without_publication(self):
  self.selected_producer();(self.bin/'fail-responses').touch();self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm');self.approve('glm');self.invoke('authorize','glm')
  (self.root/'.env').write_text('GLM_KEY=SYNTHETIC-GLM\n')
  before={n:(self.root/'ai/live/testdata/glm-formal'/n).read_bytes() for n in ('live-captures.json','live-captures.sha256','known-usage.json','support-views.json')}
  for route in ('chat','responses','messages'):self.invoke('run-selected','glm','current-source-eleven-renewal-1-glm-'+route,ok=route!='responses')
  self.invoke('prepare','remaining','--phase','remaining')
  transition=json.loads((self.root/'remaining/transition.json').read_text());self.assertFalse(transition['publication']['threeOwnPASS']);self.assertFalse(transition['publication']['published'])
  for n,raw in before.items():self.assertEqual((self.root/'ai/live/testdata/glm-formal'/n).read_bytes(),raw)
 def test_remaining_requires_strict_publication_then_new_native_grant(self):
  (self.root/'.env').mkdir();self.invoke('prepare','early','--phase','remaining',ok=False);(self.root/'.env').rmdir()
  self.complete_glm();self.invoke('prepare','remaining','--phase','remaining');self.invoke('check','remaining');self.approve('remaining')
  plan=json.loads((self.root/'remaining/plan.json').read_text());self.assertEqual(len(plan['routes']),8)
  self.assertEqual(sum(v['maxCalls'] for v in plan['routes'].values()),144);self.assertEqual(sum(v['maxAttempts'] for v in plan['routes'].values()),156)
  for v in plan['routes'].values():
   self.assertEqual(v['networkRoute'],'international-loopback-proxy')
   self.assertTrue('http://127.0.0.1:7890' in v['argv'] or 'http://127.0.0.1:7890' in v['environment'].values())
  self.invoke('grant','remaining')
  grant=json.loads((self.root/'.evidence/minimax-renewals/current-source-eleven-renewal-1/grant.json').read_text())
  self.assertEqual((grant['originalCalls'],grant['originalAttempts']),(39,36));self.assertEqual((grant['maxCalls'],grant['maxAttempts']),(60,72))
  self.assertEqual(grant['sourceCommit'],self.git_head())
  self.invoke('grant','remaining',ok=False)
 def git_head(self):return subprocess.check_output(['git','rev-parse','HEAD'],cwd=self.root,env=self.env,text=True).strip()
 def invoke(self,*args,ok=True):
  self.outputs.add(args[1]);p=subprocess.run([*self.cli,*args],cwd=self.root,env=self.env,text=True,capture_output=True,timeout=120)
  self.journal.append(dict(args=args,exitCode=p.returncode,stdout=p.stdout.replace(str(self.root),'$ISOLATED_REPO'),stderr=p.stderr.replace(str(self.root),'$ISOLATED_REPO')))
  self.assertEqual(p.returncode==0,ok,p.stderr[-1000:]+p.stdout[:200]);return p
 def approve(self,out):
  import hashlib
  base=self.root/out;plan=json.loads((base/'plan.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
  review.update(decision='APPROVED',reviewedBy='synthetic independent reviewer',authorizedRoutes=list(plan['routes']),credentialScopes={j['sourceVariable']:'current-source-eleven-renewal-1/'+j['networkRoute'] for j in plan['routes'].values()})
  (base/'prerelease-source-review.json').write_text(json.dumps(review));auth=json.loads((base/'execution-authorization.template.json').read_text())
  auth.update(decision='APPROVED',authority='explicit-additional-budget-reply',requestReference='synthetic NEW216/252 reply',requestText='Approved fixed eleven additional216/252 only',checksSHA256=hashlib.sha256((base/'checks.json').read_bytes()).hexdigest(),sourceReviewSHA256=hashlib.sha256((base/'prerelease-source-review.json').read_bytes()).hexdigest())
  (base/'execution-authorization.json').write_text(json.dumps(auth))
 def test_keyless_checks_new_approval_and_global_one_shot_precede_selected_secret(self):
  self.selected_producer();(self.root/'.env').mkdir();self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm')
  checks=json.loads((self.root/'glm/checks.json').read_text());self.assertEqual(len(checks['commandReceipts']),6)
  name='current-source-eleven-renewal-1-glm-chat';intent=self.root/'.evidence/minimax-launches'/(name+'.json')
  self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists());self.approve('glm')
  auth=self.root/'glm/execution-authorization.json';raw=auth.read_bytes()
  for field,value in [('round','release-repair-validation-1'),('approvedAdditionalMaxCalls',216.0),('approvedAdditionalMaxAttempts',True),('approvalProposalSHA256','0'*64),('historySHA256','0'*64)]:
   changed=json.loads(raw);changed[field]=value;auth.write_text(json.dumps(changed));self.invoke('authorize','glm',ok=False);auth.write_bytes(raw)
  self.invoke('authorize','glm');self.assertTrue((self.root/'.env').is_dir())
  (self.root/'.env').rmdir();(self.root/'.env').write_text('GLM_KEY=SYNTHETIC-GLM\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')
  self.invoke('run-selected','glm',name);self.invoke('run-selected','glm',name,ok=False)
  record=json.loads(intent.read_text());self.assertEqual(record['budget']['calls'],22)
  self.assertEqual(record['networkRoute'],'china-direct');log=(self.root/'glm/glm-chat.log').read_text()
  self.assertNotIn('SYNTHETIC-GLM',log);self.assertNotIn('MINIMAX',log);self.assertNotIn('HTTP_PROXY',log)
  self.invoke('prepare','other','--phase','glm');self.invoke('check','other');self.approve('other');self.invoke('authorize','other',ok=False)
  self.invoke('run-selected','other',name,ok=False);self.assertFalse((self.root/'other/glm-chat-command.json').exists())
 def test_pending_three_phase_has_one_fixed_eleven_budget_and_immutable_history(self):
  (self.root/'.env').mkdir();self.invoke('prepare','glm','--phase','glm')
  plan=json.loads((self.root/'glm/plan.json').read_text())
  self.assertEqual(plan['round'],'current-source-eleven-renewal-1')
  self.assertEqual([v['baseName'] for v in plan['routes'].values()],['glm-chat','glm-responses','glm-messages'])
  self.assertEqual(plan['budget']['additionalMaxCalls'],216);self.assertEqual(plan['budget']['additionalMaxAttempts'],252)
  self.assertEqual(plan['history']['known'],dict(calls=648,attempts=648))
  self.assertEqual(plan['history']['separateM3Formal'],dict(calls=39,attempts=36))
  self.assertEqual(len(plan['allJobs']),11)
  self.assertEqual(sum(v['maxCalls'] for v in plan['allJobs'].values()),216)
  self.assertEqual(sum(v['maxAttempts'] for v in plan['allJobs'].values()),252)
  for v in plan['allJobs'].values():
   if v['baseName'].startswith('mm-'):
    self.assertEqual(v['environment'].get('BARNESS_MINIMAX_PROXY'),'http://127.0.0.1:7890')
    self.assertNotIn('BARNESS_AI_LIVE_PROXY',v['environment'])
  self.assertTrue(all(v['networkRoute']=='china-direct' for v in plan['routes'].values()))
  template=json.loads((self.root/'glm/execution-authorization.template.json').read_text())
  self.assertEqual(template['decision'],'PENDING');self.assertFalse(template['requestReference'])
  self.assertTrue((self.root/'.env').is_dir())

if __name__=='__main__' and len(sys.argv)>1 and sys.argv[1]=='--seed':
 import paid_core as c,paid_plan as p,routed_evidence as e,repair_contract as r,repair_inputs as inputs
 original=Path(sys.argv[2])
 def copy(reader,n):
  raw=reader.read(n,16<<20);dest=c.ROOT/n;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(raw)
 with e.Root(original) as reader:
  for base in r.NAMES:
   name=r.ROUND+'-'+base;relative='.evidence/minimax-launches/'+name+'.json';copy(reader,relative)
   record=e.document(reader.read(relative));out=record['collectionRoot'].removeprefix('$REPO/')
   for leaf in ('source-at-start.json','binaries.json','plan.json','prerelease-source-review.json','execution-authorization.json','checks.json','history.json','reuse.json','scope.json',base+'-command.json'):
    copy(reader,out+'/'+leaf)
   for leaf in ('vet','vet-live','no-key-live','no-key-visual-proxy','no-key-release-contracts','no-key-release-package'):copy(reader,out+'/'+leaf+'-command.json')
   directory=out+'/'+base
   leaf='formal-glm-report.json' if base.startswith('glm-') else 'multimodal-report.json' if base=='protocol114' else 'report.json'
   path=directory if base.startswith('visual-') else inputs.bundle(reader,directory,leaf);copy(reader,path+'/'+leaf)
  copy(reader,'.evidence/minimax-stages/formal.jsonl')
  base='.evidence/glm-messages-cancel-20261019-1'
  for leaf in ('paid-intent.json','actual-command.json','actual/20261018T230218.566865000Z/diagnostic-glm-messages-cancel-report.json'):copy(reader,base+'/'+leaf)
else:
 if __name__=='__main__':unittest.main()
