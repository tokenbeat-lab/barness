"""Socket-free public fixed4 process checks; never actual approvals/keys."""
import json
import subprocess
import sys
from pathlib import Path
from test_recovery_round import RecoveryRound

class Recovery4Round(RecoveryRound):
    def setUp(self):
        super().setUp()
        self.cli=[sys.executable,'-B',str(self.tools/'recovery4_round.py')]

    def test_four_inventory_exact_budget_policy_and_first_order(self):
        value=json.loads(self.runcli('inventory').stdout)
        self.assertEqual((value['validationRound'],value['nativeRenewalId']),('provider-sites-final-acceptance-4',)*2)
        self.assertEqual(value['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        self.assertEqual((value['maxCalls'],value['maxAttempts'],value['campaignMaxCalls'],value['campaignMaxAttempts']),(506,518,1070,1130))
        self.assertEqual((value['newReservedMicrousd'],value['retainedReservedMicrousd']),(17300000,12600000))
        self.assertEqual(len(value['jobs']),24)
        self.assertEqual(value['executionOrder'][0],'formal-zai-responses')
        self.assertEqual(value['executionOrder'][1:4],['mm-global-formal-m3-'+r for r in ('chat','responses','messages')])
        self.assertEqual(value['jobs']['google-interactions-image']['reservedMicrousd'],2600000)
        self.assertEqual(sum(j['reservedMicrousd'] for j in value['jobs'].values()),17300000)
        self.assertTrue(all(j['networkRoute']=='http://127.0.0.1:7890' for j in value['jobs'].values() if j['provider'] not in ('minimax-cn','deepseek','zhipu')))
        self.assertTrue((self.root/'.env').is_dir())

    def test_four_old3_all_active_modes_close_before_out_or_key(self):
        seal=self.root/'.evidence/provider-sites/final-acceptance-20261020/closed-cohort-3/seal.json'
        seal.parent.mkdir();seal.write_text('{}')
        old=[sys.executable,'-B',str(self.tools/'recovery_round.py')]
        for mode in ('prepare','preview','check','native-grant','run-selected','final-plan','finalize'):
            p=subprocess.run([*old,mode,'.evidence/provider-sites/no-reopen','formal-zai-responses'] if mode in ('preview','run-selected') else [*old,mode,'.evidence/provider-sites/no-reopen'],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=3)
            self.journal.append(dict(argv=[mode],exitCode=p.returncode,stderr=p.stderr))
            self.assertEqual(p.returncode,1);self.assertIn('permanently closed',p.stderr)
        self.assertEqual(subprocess.run([*old,'inventory'],cwd=self.root,env=self.env,capture_output=True).returncode,0)
        self.assertFalse((self.root/'.evidence/provider-sites/no-reopen').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def ready(self):
        from test_recovery4_fixture import seed,builder
        seed(self);builder(self)
        self.runcli('prepare','.evidence/provider-sites/recovery4')
        base=self.root/'.evidence/provider-sites/recovery4'
        self.runcli('check','.evidence/provider-sites/recovery4')
        return base

    def test_four_prepare_binds1012_closed3_and_all_old_failed_facts(self):
        base=self.ready()
        prepared=json.loads((base/'prepared.json').read_text())
        self.assertEqual(prepared['round'],'provider-sites-final-acceptance-4')
        prior=json.loads((base/'predecessor.json').read_text())
        self.assertEqual((prior['actualCalls'],prior['actualAttempts'],prior['claimedCalls'],prior['claimedAttempts'],prior['reservedMicrousd']),(378,374,564,612,12600000))
        self.assertEqual(prior['thirdCohort']['nativeLedgerRows'],75)
        self.assertEqual(len(prior['thirdCohort']['nativeAuthoritySha256']),9)
        self.assertEqual(len(prior['thirdCohort']['parentClaimsSha256']),5)
        human=json.loads((base/'recovery4-authorization.template.json').read_text())
        self.assertEqual(human['decision'],'PENDING')
        self.assertFalse((self.root/'.evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json').exists())
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4').exists())
        self.assertTrue((self.root/'.env').is_dir())


    def approve_recovery(self,base):
        from test_recovery4_fixture import approve
        approve(self,base)

    def test_four_new_actual_human_and_order_before_claim_or_key(self):
        base=self.ready();self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.approve_recovery(base)
        self.runcli('native-grant','.evidence/provider-sites/recovery4',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','mm-global-formal-m3-chat',ok=False)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')))
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def test_four_first18_native_then_schema5_single_issue_and_policy2(self):
        from test_recovery4_fixture import keys
        base=self.ready();self.approve_recovery(base);keys(self)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses')
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').exists())
        self.runcli('native-grant','.evidence/provider-sites/recovery4')
        grant=json.loads((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').read_text())
        self.assertEqual(grant['schema'],5);self.assertEqual(grant['renewalId'],'provider-sites-final-acceptance-4')
        self.assertEqual((grant['thirdCohort']['nativeCalls'],grant['thirdCohort']['nativeAttempts']),(39,36))
        self.assertNotIn('thirdCohortHistorySha256',grant);self.assertNotIn('recoveryHistorySha256',grant)
        self.runcli('native-grant','.evidence/provider-sites/recovery4',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','mm-global-formal-m3-chat')
        cmd=json.loads((base/'mm-global-formal-m3-chat-command.json').read_text());self.assertEqual(cmd['validationRound'],'provider-sites-final-acceptance-4');self.assertEqual(cmd['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        usd=json.loads((base/'mm-global-formal-m3-chat-usd.json').read_text());self.assertEqual(usd['round'],'provider-sites-final-acceptance-2')
        self.runcli('run-selected','.evidence/provider-sites/recovery4','mm-global-formal-m3-chat',ok=False)

    def test_four_all24_select33_retained46_and27_inputs(self):
        from test_recovery4_fixture import keys
        import shutil
        base=self.ready();self.approve_recovery(base);keys(self)
        old=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-formal-zai-responses.json';original=old.read_bytes()
        plan=json.loads((base/'plan.json').read_text())
        for i,name in enumerate(plan['executionOrder']):
            self.runcli('run-selected','.evidence/provider-sites/recovery4',name)
            if i==0:self.runcli('native-grant','.evidence/provider-sites/recovery4')
        summary=json.loads(self.runcli('summarize').stdout)
        self.assertEqual((summary['actualProcessCount'],summary['effectiveRoleCount'],summary['retainedUnknownProcesses']),(46,33,3))
        self.assertEqual((summary['claimedCalls'],summary['claimedAttempts'],summary['reservedMicrousd']),(1070,1130,29900000))
        live={};effects={}
        for name,job in plan['jobs'].items():
            if name=='protocol114':continue
            root=self.root/str(job['evidenceRoot']).replace('$REPO/','');path=root if name.startswith('visual-') else root/'controlled-session'
            if name.startswith(('text-','visual-')):effects[name]=str(path.relative_to(self.root))
            else:live[name.removeprefix('formal-').replace('mm-global-formal-m3-','minimax-')]=str(path.relative_to(self.root))
        for route in ('chat','responses','messages'):live['zhipu-'+route]='.evidence/provider-sites/retained-glm/'+route
        jev=base/'derived-jev';jev.mkdir();shutil.copytree(base/'protocol114/controlled-session',jev/'protocol');shutil.copytree(base/'visual-jev',jev/'effects');(jev/'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')));live['typesafe-classifier-jev-1.14.0']=str(jev.relative_to(self.root))
        inputs=base/'selected.json';inputs.write_text(json.dumps(dict(live=live,effects=effects)))
        forged=base/'forged-same-bytes';shutil.copytree(self.root/live['zai-responses'],forged);bad=dict(live);bad['zai-responses']=str(forged.relative_to(self.root));inputs.write_text(json.dumps(dict(live=bad,effects=effects)));self.runcli('finalize','.evidence/provider-sites/recovery4','--inputs',str(inputs.relative_to(self.root)),ok=False);self.assertFalse((base/'finalization').exists());inputs.write_text(json.dumps(dict(live=live,effects=effects)))
        result=json.loads(self.runcli('finalize','.evidence/provider-sites/recovery4','--inputs',str(inputs.relative_to(self.root))).stdout)
        proof=json.loads((base/'finalization/effective33-retained46.json').read_text());self.assertEqual(len(proof['roles']),33)
        self.assertEqual(proof['accounting']['actualProcessCount'],46);self.assertEqual(old.read_bytes(),original)
        self.assertFalse(result['automaticRace']);self.assertEqual(len(json.loads((self.root/'ai/live/support-matrix.json').read_text())['rows']),23)


    def test_four_source_open_swap_fifo_bounded_before_key(self):
        base=self.ready();self.approve_recovery(base)
        code="""import os,pathlib,runpy,sys
victim=pathlib.Path('ai/sentinel.go').resolve();original=pathlib.Path.lstat;fired=False
def swapped(self,*a,**kw):
 global fired
 info=original(self,*a,**kw)
 if not fired and self==victim:
  fired=True;self.unlink();os.mkfifo(self);print('FIFO_SWAP_INJECTED',file=sys.stderr)
 return info
pathlib.Path.lstat=swapped
cli=sys.argv[1];sys.path.insert(0,str(pathlib.Path(cli).parent));sys.argv=[cli,'preview','.evidence/provider-sites/recovery4','formal-zai-responses'];runpy.run_path(cli,run_name='__main__')
"""
        try:p=subprocess.run([sys.executable,'-B','-c',code,str(self.tools/'recovery4_round.py')],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=3)
        except subprocess.TimeoutExpired:
            self.journal.append(dict(argv=['preview','source-FIFO-open-swap'],exitCode=None,state='TIMEOUT'));self.fail('source opener blocked on FIFO')
        self.journal.append(dict(argv=['preview','source-FIFO-open-swap'],exitCode=p.returncode,stderr=p.stderr));self.assertEqual(p.returncode,1);self.assertIn('REFUSED',p.stderr);self.assertIn('FIFO_SWAP_INJECTED',p.stderr)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')))


    def test_four_historical_source_cannot_be_coherently_resealed_false(self):
        from test_recovery4_fixture import seed,builder,BASE,reseal
        import hashlib
        seed(self);builder(self)
        stop_path=self.root/BASE/'recovery3-stopped/index.json';stop=json.loads(stop_path.read_text());row=stop['processes'][0]
        report=self.root/row['nativeBundle']/'minimax-report.json';v=json.loads(report.read_text());v['executionSource']={'forged-source.go':'f'*64};report.write_text(json.dumps(v));pin=hashlib.sha256(report.read_bytes()).hexdigest();row['nativeReportSHA256']=pin
        claim=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-mm-global-formal-m3-chat.json';v=json.loads(claim.read_text());v['reportedFileSha256']=pin;claim.write_text(json.dumps(v));stop_path.write_text(json.dumps(stop));reseal(self,[report,claim])
        self.runcli('prepare','.evidence/provider-sites/forged4',ok=False)
        self.assertFalse((self.root/'.evidence/provider-sites/forged4').exists());self.assertTrue((self.root/'.env').is_dir())

    def test_four_extra_unknown_current_claim_not_ignored(self):
        base=self.ready()
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-4-unregistered.json';p.write_text('{}')
        self.runcli('summarize',ok=False)
        self.assertTrue((self.root/'.env').is_dir())


    def test_four_final_report_open_swap_fifo_is_bounded(self):
        base=self.ready();inventory=json.loads(self.runcli('inventory').stdout);names=list(inventory['jobs'])
        live={};effects={}
        identities=('openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','typesafe-classifier-jev-1.14.0','zhipu-chat','zhipu-responses','zhipu-messages','minimax-chat','minimax-responses','minimax-messages','zai-chat','zai-responses','zai-messages','minimax-cn-chat','minimax-cn-responses','minimax-cn-messages')
        for name in identities:
            p=base/'plan-only'/name;p.mkdir(parents=True);(p/'live-report.json').write_text(json.dumps(dict(combo=name)));live[name]=str(p.relative_to(self.root))
        for name in ('text-jev','text-luna','visual-jev','visual-decisions'):
            p=base/'plan-only'/name;p.mkdir();(p/('evaluation-report.json' if name.startswith('text-') else 'report.json')).write_text('{}');effects[name]=str(p.relative_to(self.root))
        selected=base/'plan-only-inputs.json';selected.write_text(json.dumps(dict(live=live,effects=effects)))
        code="""import os,pathlib,runpy,sys
victim=pathlib.Path('.evidence/provider-sites/recovery4/plan-only/openai-responses/live-report.json').resolve();original=pathlib.Path.is_file;fired=False
def swapped(self,*a,**kw):
 global fired
 result=original(self,*a,**kw)
 if not fired and self==victim:
  fired=True;self.unlink();os.mkfifo(self);print('FIFO_SWAP_INJECTED',file=sys.stderr)
 return result
pathlib.Path.is_file=swapped
cli=sys.argv[1];sys.path.insert(0,str(pathlib.Path(cli).parent));sys.argv=[cli,'final-plan','.evidence/provider-sites/recovery4','--inputs','.evidence/provider-sites/recovery4/plan-only-inputs.json'];runpy.run_path(cli,run_name='__main__')
"""
        try:p=subprocess.run([sys.executable,'-B','-c',code,str(self.tools/'recovery4_round.py')],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=3)
        except subprocess.TimeoutExpired:
            self.journal.append(dict(argv=['final-plan','report-FIFO-open-swap'],exitCode=None,state='TIMEOUT'));self.fail('selected report opener blocked on FIFO')
        self.journal.append(dict(argv=['final-plan','report-FIFO-open-swap'],exitCode=p.returncode,stderr=p.stderr));self.assertEqual(p.returncode,1);self.assertIn('FIFO_SWAP_INJECTED',p.stderr)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((base/'finalization').exists())


    def test_four_approval_scope_policy_case_alias_and_proxy_refuse(self):
        base=self.ready();self.approve_recovery(base)
        auth=base/'execution-authorization.json';original=auth.read_bytes()
        cases=[('maxCalls',507),('campaignMaxAttempts',1131),('newReservedMicrousd',19200000),('retainedReservedMicrousd',10700000),('nativeRenewalId','provider-sites-final-acceptance-3'),('spendingGuardPolicyId','provider-sites-final-acceptance-4'),('executionOrder',['openai-chat'])]
        for key,value in cases:
            with self.subTest(field=key):
                v=json.loads(original);v[key]=value;auth.write_text(json.dumps(v));self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        for field,key,value in [('processReservationsMicrousd','google-interactions-image',4500000),('networkRoutes','formal-zai-responses','direct'),('usdLedgers','formal-zai-responses','.evidence/old3-usd.json')]:
            v=json.loads(original);v[field][key]=value;auth.write_text(json.dumps(v));self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        auth.write_bytes(original)
        human=self.root/'.evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json';raw=human.read_text()
        for altered in [raw.replace('"decision": "APPROVED"','"decision": "PENDING", "decision": "APPROVED"'),raw.replace('"decision": "APPROVED"','"Decision": "PENDING", "decision": "APPROVED"'),raw.replace('"decision": "APPROVED"','"decision": "PENDING"')]:
            human.write_text(altered);self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')));self.assertTrue((self.root/'.env').is_dir())

    def test_four_closed3_mutations_extra_intent_fifo_and_old_unknown_refuse(self):
        from test_recovery4_fixture import seed,builder,BASE
        import os
        seed(self);builder(self)
        targets=[BASE/'closed-cohort-3/seal.json',Path('.evidence/minimax-renewals/provider-sites-final-acceptance-3/chat.json'),Path('.evidence/minimax-launches/provider-sites-final-acceptance-3-formal-zai-responses.json'),BASE/'remaining-stopped/index.json']
        for target in targets:
            path=self.root/target;raw=path.read_bytes();path.write_text('{}');self.runcli('summarize',ok=False);path.write_bytes(raw)
        seal=self.root/BASE/'closed-cohort-3/seal.json';raw=seal.read_bytes();seal.unlink();os.mkfifo(seal);self.runcli('summarize',ok=False);seal.unlink();seal.write_bytes(raw)
        for path in [self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-3/additional-intent.json',self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-openai-chat.json']:
            path.write_text('{}');self.runcli('summarize',ok=False);path.unlink()
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4').exists())

    def test_four_failed_first_holds_whole_claim_and_no_native_issue(self):
        base=self.ready();self.approve_recovery(base)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-4-formal-zai-responses.json';raw=p.read_bytes();v=json.loads(raw)
        self.assertEqual(v['reservedMicrousd'],650000);self.assertEqual(v['budget']['state'],'UNKNOWN')
        self.runcli('native-grant','.evidence/provider-sites/recovery4',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.assertEqual(p.read_bytes(),raw);self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').exists())
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual(summary['reservedMicrousd'],13250000);self.assertEqual((summary['claimedCalls'],summary['claimedAttempts']),(588,636))


    def test_four_old3_reply_reference_cannot_authorize_relabelled4(self):
        import hashlib
        base=self.ready();self.approve_recovery(base)
        human=self.root/'.evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json';v=json.loads(human.read_text());v['requestReference']='controlled actual3';human.write_text(json.dumps(v))
        auth=base/'execution-authorization.json';a=json.loads(auth.read_text());a.update(requestReference=v['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(human.read_bytes()).hexdigest());auth.write_text(json.dumps(a))
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')))
        self.assertTrue((self.root/'.env').is_dir())


    def test_four_copied_old_nonce_refuses_first_before_claim_or_key(self):
        import shutil
        base=self.ready();self.approve_recovery(base)
        native=self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4';native.mkdir();shutil.copyfile(native.with_name('provider-sites-final-acceptance-3')/'chat.json',native/'chat.json')
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')))
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((native/'grant.json').exists())

    def test_four_concurrent_first_and_crossout_never_reset_reservation(self):
        import shutil
        base=self.ready();self.approve_recovery(base)
        other=base.with_name('other-recovery4');shutil.copytree(base,other)
        self.runcli('run-selected','.evidence/provider-sites/other-recovery4','formal-zai-responses',ok=False)
        argv=[*self.cli,'run-selected','.evidence/provider-sites/recovery4','formal-zai-responses']
        processes=[subprocess.Popen(argv,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True) for _ in range(2)]
        for child in processes:
            stdout,stderr=child.communicate(timeout=10);self.journal.append(dict(argv=argv,exitCode=child.returncode,stdout=stdout,stderr=stderr));self.assertNotEqual(child.returncode,0)
        claims=list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*'));self.assertEqual(len(claims),1)
        v=json.loads(claims[0].read_text());self.assertEqual(v['reservedMicrousd'],650000);self.assertEqual(v['budget']['state'],'UNKNOWN')
        self.assertTrue((self.root/'.env').is_dir())


def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(Recovery4Round(n) for n in loader.getTestCaseNames(Recovery4Round) if n.startswith('test_four_'))
