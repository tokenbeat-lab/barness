"""New finite authority public CLI failure inventory before implementation.

Seam: isolated subprocesses, fixed native DTOs and canonical root-owned receipts.
No producer/vendor request or actual credential is needed for planning/preflight.
"""
import json,unittest,subprocess
import test_repair_fixtures as f

class Repair(f.Isolated,unittest.TestCase):
 def test_six_plan_is_pending_and_preserves_known545_history(self):
  (self.root/'.env').mkdir()
  for phase,names in [('glm',['glm-chat','glm-responses','glm-messages']),('effects',['protocol114','visual-jev','visual-decisions'])]:
   self.invoke('prepare',phase,'--phase',phase)
   plan=json.loads((self.root/phase/'plan.json').read_text())
   self.assertEqual(plan['round'],'release-repair-validation-1')
   self.assertEqual([v['baseName'] for v in plan['routes'].values()],names)
   self.assertEqual(plan['budget'],dict(priorCalls=545,priorAttempts=545,additionalMaxCalls=108,additionalMaxAttempts=132,totalMaxCalls=653,totalMaxAttempts=677))
   self.assertEqual(len(plan['history']['receipts']),39)
   self.assertEqual(plan['history']['known'],dict(calls=545,attempts=545))
   self.assertEqual(len(plan['reuse']['inputs']),15)
   self.assertEqual(plan['scope']['baseline'],f.BASE)
   for j in plan['routes'].values():
    self.assertTrue(j['name'].startswith('release-repair-validation-1-'))
    self.assertEqual(j['networkRoute'],'china-direct' if j['provider']=='zhipu' else 'international-loopback-proxy')
    if j['provider']=='zhipu':self.assertNotIn('BARNESS_AI_LIVE_PROXY',j['environment'])
    elif j['baseName'].startswith('visual-'):self.assertIn('http://127.0.0.1:7890',j['argv'])
    else:self.assertEqual(j['environment']['BARNESS_AI_LIVE_PROXY'],'http://127.0.0.1:7890')
   template=json.loads((self.root/phase/'execution-authorization.template.json').read_text())
   self.assertEqual(template['decision'],'PENDING');self.assertEqual(template['requestReference'],'')
   preview=json.loads(self.invoke('preview',phase,next(iter(plan['routes']))).stdout)
   self.assertFalse(preview['credentialsRead'])
  self.assertTrue((self.root/'.env').is_dir())
  self.assertFalse(any((self.root/'.evidence/minimax-launches').glob('release-repair-validation-1-*')))

 def test_native_capture_inventory_and_exact_scope_deny_laundering(self):
  (self.root/'.env').mkdir()
  bundle=next((self.root/'.evidence/minimax-routed-legacy-1/live/openai-chat').iterdir());manifest=json.loads((bundle/'manifest.json').read_text());case=manifest['cases'][0];capture=bundle/case['dir']/next(iter(case['artifacts']));raw=capture.read_bytes()
  for mode in ('edit','delete','link'):
   try:
    with self.subTest(native=mode):
     if mode=='edit':capture.write_bytes(raw+b' ')
     elif mode=='delete':capture.unlink()
     else:
      copy=capture.parent/'exact-copy.json';copy.write_bytes(raw);capture.unlink();capture.symlink_to(copy)
     self.invoke('prepare','bad-native-'+mode,'--phase','glm',ok=False)
   finally:
    if capture.is_symlink():capture.unlink()
    capture.write_bytes(raw)
  target=self.root/'ai/internal/testkit/release/plan.go';original=target.read_bytes();target.write_bytes(original+b'\n')
  subprocess.run(['git','add',str(target)],cwd=self.root,env=self.env,check=True)
  subprocess.run(['git','-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','wrong named offline delta'],cwd=self.root,env=self.env,check=True)
  self.invoke('prepare','bad-named-source','--phase','glm',ok=False)
  self.assertTrue((self.root/'.env').is_dir())
  self.assertFalse(any((self.root/'.evidence/minimax-launches').glob('release-repair-validation-1-*')))

 def test_six_real_preflight_commands_are_keyless_and_pending(self):
  (self.root/'.env').mkdir();self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm')
  checks=json.loads((self.root/'glm/checks.json').read_text())
  self.assertEqual(checks['commands'],dict(vet=0,**{'vet-live':0,'no-key-live':0,'no-key-visual-proxy':0,'no-key-release-contracts':0,'no-key-release-package':0}))
  self.assertEqual(len(checks['commandReceipts']),6);self.assertTrue(checks['passed']);self.assertFalse(checks['credentialsRead'])
  probes=[json.loads(line) for line in (self.root/'go-probe.jsonl').read_text().splitlines()]
  self.assertEqual(len(probes),7)
  self.assertEqual(probes[-1]['argv'],['test','./ai/internal/testkit/release','-count=1'])
  self.assertEqual(probes[-2]['argv'],['test','./ai/e2e','-count=1','-run','^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract|TestReleaseRecordedBundleDiagnosticProcess)$'])
  self.assertTrue(all(not any('KEY' in n or 'PROXY' in n for n in p['environmentKeys']) for p in probes))
  self.assertEqual(json.loads((self.root/'glm/execution-authorization.template.json').read_text())['decision'],'PENDING')
  self.invoke('run-selected','glm','release-repair-validation-1-glm-chat',ok=False)
  self.assertFalse((self.root/'.evidence/minimax-launches/release-repair-validation-1-glm-chat.json').exists())
  self.assertTrue((self.root/'.env').is_dir())

 def test_core_new_approval_one_shot_and_honest_failed_phase_settlement(self):
  self.selected_producer();(self.root/'.env').mkdir();self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm')
  name='release-repair-validation-1-glm-chat';intent=self.root/'.evidence/minimax-launches'/(name+'.json')
  self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists());self.approve('glm')
  auth=self.root/'glm/execution-authorization.json';raw=auth.read_bytes()
  for field,value in [('authority','explicit-user-request'),('round','routed-validation-1'),('approvedAdditionalMaxCalls',72),('reuseSHA256','0'*64),('scopeSHA256','0'*64),('checksSHA256','0'*64)]:
   changed=json.loads(raw);changed[field]=value;auth.write_text(json.dumps(changed));self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists());auth.write_bytes(raw)
  (self.root/'.env').rmdir();(self.root/'.env').write_text('GLM_KEY=SYNTHETIC-GLM\nTYPESAFE_KEY=SYNTHETIC-TS\nOPENAI_KEY=SYNTHETIC-OA\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')
  (self.bin/'fail-responses').touch();self.invoke('run-selected','glm',name)
  self.invoke('run-selected','glm','release-repair-validation-1-glm-responses',ok=False)
  self.invoke('run-selected','glm','release-repair-validation-1-glm-messages')
  summary=json.loads(self.invoke('summarize','glm').stdout);self.assertEqual(summary['current']['calls'],65);self.assertEqual(summary['total']['calls'],610)
  chat=self.root/'glm/glm-chat/session/formal-glm-report.json';original=chat.read_bytes();chat.unlink()
  summary=json.loads(self.invoke('summarize','glm').stdout);row=next(v for v in summary['routes'] if v['route']==name)
  self.assertEqual((row['state'],row['conservativeCalls'],row['conservativeAttempts']),('UNKNOWN',24,32))
  self.invoke('prepare','effects','--phase','effects');self.invoke('check','effects');self.approve('effects')
  target='release-repair-validation-1-protocol114';self.invoke('run-selected','effects',target,ok=False)
  self.assertFalse((intent.parent/(target+'.json')).exists());chat.write_bytes(original)
  for base in ('protocol114','visual-jev','visual-decisions'):self.invoke('run-selected','effects','release-repair-validation-1-'+base)
  summary=json.loads(self.invoke('summarize','effects').stdout);self.assertEqual((summary['current']['calls'],summary['current']['attempts']),(101,101));self.assertEqual((summary['total']['calls'],summary['total']['attempts']),(646,646))
  self.assertEqual(len(list(intent.parent.glob('release-repair-validation-1-*'))),6)
  self.invoke('prepare','elsewhere','--phase','glm');self.invoke('check','elsewhere');self.approve('elsewhere');(self.root/'.env').unlink();(self.root/'.env').mkdir()
  self.invoke('run-selected','elsewhere',name,ok=False);self.assertFalse((self.root/'elsewhere/glm-chat-command.json').exists())
  log=(self.root/'glm/glm-chat.log').read_text();self.assertNotIn('SYNTHETIC-GLM',log);self.assertNotIn('MINIMAX',log);self.assertNotIn('HTTP_PROXY',log)

 def test_each_failed_keyless_check_or_receipt_blocks_before_intent_and_key(self):
  import hashlib
  (self.root/'.env').mkdir()
  failures=['vet','vet-live','no-key-live','no-key-visual-proxy','no-key-release-contracts','no-key-release-package']
  for index,name in enumerate(failures):
   with self.subTest(check=name):
    out='failed-check-'+str(index);self.invoke('prepare',out,'--phase','glm');(self.root/'fail-check-name').write_text(name)
    self.invoke('check',out,ok=False);self.approve(out)
    checks=json.loads((self.root/out/'checks.json').read_text());self.assertFalse(checks['passed']);self.assertEqual(checks['commands'][name],7)
    self.invoke('run-selected',out,'release-repair-validation-1-glm-chat',ok=False)
    self.assertFalse((self.root/'.evidence/minimax-launches/release-repair-validation-1-glm-chat.json').exists());(self.root/'fail-check-name').unlink()
  self.invoke('prepare','valid-checks','--phase','glm');self.invoke('check','valid-checks');self.approve('valid-checks')
  base=self.root/'valid-checks';receipt=base/'no-key-release-package-command.json';raw=receipt.read_bytes();v=json.loads(raw);v['exitCode']=0;v['argv']=['go','version'];receipt.write_text(json.dumps(v))
  checks=json.loads((base/'checks.json').read_text());checks['commandReceipts']['no-key-release-package']=hashlib.sha256(receipt.read_bytes()).hexdigest();(base/'checks.json').write_text(json.dumps(checks));self.approve('valid-checks')
  self.invoke('run-selected','valid-checks','release-repair-validation-1-glm-chat',ok=False)
  self.assertFalse((self.root/'.evidence/minimax-launches/release-repair-validation-1-glm-chat.json').exists());self.assertTrue((self.root/'.env').is_dir())

 def test_empty_directory_fanout_is_bounded_even_when_certified_files_match(self):
  (self.root/'.env').mkdir();bundle=next((self.root/'.evidence/minimax-routed-legacy-1/live/openai-chat').iterdir())
  for n in range(32769):(bundle/('empty-'+str(n))).mkdir()
  self.invoke('prepare','bad-empty-fanout','--phase','glm',ok=False)
  self.assertTrue((self.root/'.env').is_dir());self.assertFalse(any((self.root/'.evidence/minimax-launches').glob('release-repair-validation-1-*')))

 def test_certified_receipts_policy_and_m3_origin_cannot_be_freshly_rehashed(self):
  (self.root/'.env').mkdir()
  paths=['.evidence/minimax-launches/routed-validation-1-openai-chat.json','.evidence/minimax-launches/mm-global-formal-m3-chat.json','.scratch/barness-ai-minimax/evidence/ticket-13/postgate-source-policy-review.json']
  for index,name in enumerate(paths):
   path=self.root/name;raw=path.read_bytes()
   try:
    with self.subTest(certified=name):
     value=json.loads(raw);value['unexpected']='changed';path.write_text(json.dumps(value));self.invoke('prepare','changed-certified-'+str(index),'--phase','glm',ok=False)
   finally:path.write_bytes(raw)
  self.assertFalse(any((self.root/'.evidence/minimax-launches').glob('release-repair-validation-1-*')))

 def test_core_prepaid_matrix_drift_is_not_runtime_or_new_authority(self):
  matrix=self.root/'ai/live/support-matrix.json';value=json.loads(matrix.read_text());value['rows'][0]['model']='wrong';matrix.write_text(json.dumps(value))
  (self.root/'.env').mkdir();self.invoke('prepare','wrong-matrix','--phase','glm',ok=False)
  self.assertFalse(any((self.root/'.evidence/minimax-launches').glob('release-repair-validation-1-*')))

if __name__=='__main__':unittest.main(verbosity=2)
