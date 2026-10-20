"""Public process checks of fixed recovery3; all state is synthetic."""
import json
import sys
from test_final_round import FinalRound

class RecoveryRound(FinalRound):
    def setUp(self):
        super().setUp()
        guard=self.root/'ai/internal/testkit/finalbudget';guard.mkdir(parents=True)
        for name in ('guard.go','estimate.go'):(guard/name).write_text('package controlled\n')
        self.commit()
        self.cli=[sys.executable,'-B',str(self.tools/'recovery_round.py')]

    def test_recovery_inventory_exact_policy_scope_budget_and_proxy(self):
        value=json.loads(self.runcli('inventory').stdout)
        self.assertEqual(value['validationRound'],'provider-sites-final-acceptance-3')
        self.assertEqual(value['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        self.assertEqual((value['maxCalls'],value['maxAttempts'],value['campaignMaxCalls'],value['campaignMaxAttempts']),(506,518,962,1010))
        self.assertEqual(len(value['jobs']),24)
        self.assertTrue(all(j['usdLedger'].endswith(n+'-usd.json') for n,j in value['jobs'].items()))
        self.assertEqual(sum(j['reservedMicrousd'] for j in value['jobs'].values()),19200000)
        self.assertEqual(value['jobs']['anthropic-messages']['reservedMicrousd'],1800000)
        self.assertEqual(value['jobs']['openai-images']['reservedMicrousd'],3500000)
        self.assertTrue(all(j['networkRoute']=='http://127.0.0.1:7890' for j in value['jobs'].values() if j['provider'] not in ('minimax-cn','zhipu','deepseek')))
        self.assertFalse(any(n.startswith(('glm-','candidate-')) for n in value['jobs']))
        self.assertTrue((self.root/'.env').is_dir())

    def test_recovery_missing_closure_refuses_before_out_claim_or_key(self):
        self.fake_build()
        self.runcli('prepare','.evidence/provider-sites/recovery3',ok=False)
        self.assertFalse((self.root/'.evidence/provider-sites/recovery3').exists())
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-3').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def ready(self):
        from test_recovery_fixture import seed,builder
        seed(self);builder(self)
        self.runcli('prepare','.evidence/provider-sites/recovery3')
        base=self.root/'.evidence/provider-sites/recovery3'
        self.runcli('check','.evidence/provider-sites/recovery3')
        return base

    def approve_recovery(self,base):
        import hashlib,datetime
        auth=json.loads((base/'execution-authorization.template.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
        review.update(decision='APPROVED',reviewedBy='controlled independent',authorizedRoutes=list(json.loads((base/'plan.json').read_text())['jobs']))
        for field in ('standardsReview','specReview'):
            p=base/(field+'.md');p.write_text('Controlled independent source decision.');review[field]=dict(path=str(p.relative_to(self.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())
        (base/'prerelease-source-review.json').write_text(json.dumps(review))
        human=json.loads((base/'recovery3-authorization.template.json').read_text());human.update(decision='APPROVED',requestText='Controlled explicitly approve fixed24/506/518 US30',requestReference='controlled direct human reply')
        actual=self.root/'.evidence/provider-sites/final-acceptance-20261020/recovery3-authorization.json';actual.write_text(json.dumps(human))
        auth.update(decision='APPROVED',requestText=human['requestText'],requestReference=human['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(actual.read_bytes()).hexdigest())
        for name in json.loads((base/'plan.json').read_text())['jobs']:
            if not name.startswith('formal-'):continue
            p=base/'grants'/(name+'.template.json');grant=json.loads(p.read_text());grant.update(id='ps-final-3-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat());dest=p.with_name(name+'.json');dest.write_text(json.dumps(grant));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
        for name in ('protocol114','visual-jev','visual-decisions'):
            leaf='protocol-authorization.json' if name=='protocol114' else name+'-authorization.json'
            value=json.loads((base/(leaf+'.template.json')).read_text());value.update(accountAlias='controlled-own-standard',regionAlias='controlled-own-region');dest=base/leaf;dest.write_text(json.dumps(value));auth['componentAuthorizationSHA256'][name]=hashlib.sha256(dest.read_bytes()).hexdigest()
        (base/'execution-authorization.json').write_text(json.dumps(auth))

    def test_recovery_new_actual_approval_policy_and_source_before_key(self):
        base=self.ready()
        self.runcli('run-selected','.evidence/provider-sites/recovery3','openai-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-openai-chat.json').exists())
        self.approve_recovery(base)
        auth=json.loads((base/'execution-authorization.json').read_text());auth['spendingGuardPolicyId']='provider-sites-final-acceptance-3';(base/'execution-authorization.json').write_text(json.dumps(auth))
        self.runcli('run-selected','.evidence/provider-sites/recovery3','openai-chat',ok=False)
        self.assertTrue((self.root/'.env').is_dir())
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-openai-chat.json').exists())

    def test_recovery_claim_once_same_campaign_across_output_and_unknown(self):
        base=self.ready();self.approve_recovery(base);self.runcli('native-grant','.evidence/provider-sites/recovery3')
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-mm-global-formal-m3-chat.json'
        value=json.loads(p.read_text());self.assertEqual(value['reservedMicrousd'],200000);self.assertEqual(value['budget']['state'],'UNKNOWN')
        self.assertEqual(value['validationRound'],'provider-sites-final-acceptance-3');self.assertEqual(value['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False)
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual(summary['reservedMicrousd'],10900000);self.assertEqual((summary['claimedCalls'],summary['claimedAttempts']),(476,516));self.assertEqual(summary['retainedUnknownProcesses'],2)

    def test_recovery_native4_and_selected_child_policy_provenance(self):
        base=self.ready();self.approve_recovery(base)
        self.runcli('native-grant','.evidence/provider-sites/recovery3')
        grant=json.loads((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-3/grant.json').read_text());self.assertEqual(grant['schema'],4);self.assertEqual(grant['recovery']['secondCalls'],13)
        (self.root/'.env').rmdir();(self.root/'.env').write_text('MINIMAX_KEY=CONTROLLED-SELECTED\n')
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat')
        cmd=json.loads((base/'mm-global-formal-m3-chat-command.json').read_text());self.assertEqual(cmd['validationRound'],'provider-sites-final-acceptance-3');self.assertEqual(cmd['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        usd=json.loads((base/'mm-global-formal-m3-chat-usd.json').read_text());self.assertEqual(usd['round'],'provider-sites-final-acceptance-2')
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False)

    def test_recovery_later_routes_never_burn_allowance_before_m3_chat(self):
        base=self.ready();self.approve_recovery(base)
        self.runcli('run-selected','.evidence/provider-sites/recovery3','formal-zai-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-formal-zai-chat.json').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def test_recovery_complete24_retains41_and_selects33_without_old_failure(self):
        import shutil,hashlib,zipfile
        base=self.ready();self.approve_recovery(base);self.runcli('native-grant','.evidence/provider-sites/recovery3')
        plan=json.loads((base/'plan.json').read_text());(self.root/'.env').rmdir();(self.root/'.env').write_text(''.join(k+'=CONTROLLED-'+k+'\n' for k in {j['sourceVariable'] for j in plan['jobs'].values()}))
        old=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-mm-global-formal-m3-chat.json';original=old.read_bytes()
        for name in plan['executionOrder']:self.runcli('run-selected','.evidence/provider-sites/recovery3',name)
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual(summary['actualProcessCount'],41);self.assertEqual(summary['effectiveRoleCount'],33);self.assertEqual(summary['retainedUnknownProcesses'],2);self.assertEqual(summary['reservedMicrousd'],29900000);self.assertEqual((summary['claimedCalls'],summary['claimedAttempts']),(962,1010));self.assertEqual(old.read_bytes(),original)
        live={};effects={};views={}
        for name,job in plan['jobs'].items():
            if name in ('protocol114',):continue
            root=self.root/str(job['evidenceRoot']).replace('$REPO/','')
            path=root if name.startswith('visual-') else root/'controlled-session'
            if name.startswith(('text-','visual-')):effects[name]=str(path.relative_to(self.root))
            else:live[name.removeprefix('formal-').replace('mm-global-formal-m3-','minimax-')]=str(path.relative_to(self.root))
        for route in ('chat','responses','messages'):
            live['zhipu-'+route]='.evidence/provider-sites/retained-glm/'+route
        jev=base/'derived-jev';jev.mkdir()
        shutil.copytree(base/'protocol114/controlled-session',jev/'protocol');shutil.copytree(base/'visual-jev',jev/'effects')
        (jev/'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))
        live['typesafe-classifier-jev-1.14.0']=str(jev.relative_to(self.root))
        inputs=base/'selected.json';inputs.write_text(json.dumps(dict(live=live,effects=effects)))
        forged=base/'forged-formal';shutil.copytree(self.root/live['zai-chat'],forged)
        bad=dict(live);bad['zai-chat']=str(forged.relative_to(self.root));inputs.write_text(json.dumps(dict(live=bad,effects=effects)));self.runcli('finalize','.evidence/provider-sites/recovery3','--inputs',str(inputs.relative_to(self.root)),ok=False);self.assertFalse((base/'finalization').exists());inputs.write_text(json.dumps(dict(live=live,effects=effects)))
        result=json.loads(self.runcli('finalize','.evidence/provider-sites/recovery3','--inputs',str(inputs.relative_to(self.root))).stdout)
        proof=json.loads((base/'finalization/effective33-retained41.json').read_text());self.assertEqual(len(proof['roles']),33);self.assertEqual(proof['accounting']['actualProcessCount'],41);self.assertEqual(old.read_bytes(),original)
        self.assertFalse(result['automaticRace']);self.assertEqual(len(json.loads((self.root/'ai/live/support-matrix.json').read_text())['rows']),23)

    def test_recovery_closed_prefix_tamper_fifo_and_extra_old2_intent_refuse(self):
        from test_recovery_fixture import seed,builder
        import os
        seed(self);builder(self)
        victim=self.root/'.evidence/provider-sites/final-acceptance-20261020/closed-cohort-2/seal.json';original=victim.read_bytes();victim.unlink();os.mkfifo(victim)
        self.runcli('summarize',ok=False);victim.unlink();victim.write_bytes(original)
        extra=self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-2/responses.json';extra.write_text('{}');self.runcli('summarize',ok=False);extra.unlink()
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-mm-global-formal-m3-chat.json';value=json.loads(p.read_text());value['budget']=dict(state='REPORTED',status='PASS',calls=0,attempts=0);p.write_text(json.dumps(value));self.runcli('prepare','.evidence/provider-sites/recovery3',ok=False)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-3').exists())

    def test_recovery_wrong_actual_usd_round_retains_full_allowance_and_halts(self):
        base=self.ready();self.approve_recovery(base);self.runcli('native-grant','.evidence/provider-sites/recovery3')
        (self.root/'.env').rmdir();(self.root/'.env').write_text('MINIMAX_KEY=CONTROLLED-ONLY\n');(self.root/'.evidence/wrong-usd-policy').touch()
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-mm-global-formal-m3-chat.json';receipt=json.loads(p.read_text());self.assertEqual(receipt['budget']['status'],'FAIL');self.assertEqual(receipt['reservedMicrousd'],200000)
        self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-responses',ok=False)
        self.assertFalse(p.with_name('provider-sites-final-acceptance-3-mm-global-formal-m3-responses.json').exists())

    def test_recovery_source_symlink_and_different_out_refuse_before_key(self):
        import shutil
        base=self.ready();self.approve_recovery(base)
        other=base.with_name('other-recovery3');shutil.copytree(base,other)
        self.runcli('run-selected','.evidence/provider-sites/other-recovery3','mm-global-formal-m3-chat',ok=False)
        victim=base/'plan.json';raw=victim.read_bytes();victim.unlink();victim.symlink_to('execution-authorization.json');self.runcli('preview','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False);victim.unlink();victim.write_bytes(raw)
        (self.root/'ai/sentinel.go').write_text('package drift\n');self.runcli('run-selected','.evidence/provider-sites/recovery3','mm-global-formal-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-3-mm-global-formal-m3-chat.json').exists());self.assertTrue((self.root/'.env').is_dir())

    def test_recovery_fifo_swap_window_and_old2_public_entry_refuse(self):
        from test_recovery_fixture import seed,builder
        import subprocess
        seed(self);builder(self)
        old=[sys.executable,'-B',str(self.tools/'final_round.py'),'prepare','.evidence/provider-sites/old-reopened','--phase','remaining'];p=subprocess.run(old,cwd=self.root,env=self.env,capture_output=True,text=True,timeout=5);self.assertNotEqual(p.returncode,0);self.assertIn('permanently closed',p.stderr)
        code="""import os,pathlib,runpy,sys
victim=pathlib.Path('.evidence/provider-sites/final-acceptance-20261020/closed-cohort-2/seal.json').resolve();original=os.open;fired=False
def swapped(path,flags,*a,**kw):
 global fired
 if not fired and pathlib.Path(path)==victim:
  fired=True;victim.unlink();os.mkfifo(victim)
 return original(path,flags,*a,**kw)
os.open=swapped
cli=sys.argv[1];sys.path.insert(0,str(pathlib.Path(cli).parent));sys.argv=[cli,'summarize'];runpy.run_path(cli,run_name='__main__')
"""
        p=subprocess.run([sys.executable,'-B','-c',code,str(self.tools/'recovery_round.py')],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=5);self.assertEqual(p.returncode,1);self.assertIn('REFUSED',p.stderr);self.assertTrue((self.root/'.env').is_dir())

    def test_recovery_component_and_final_selected_fifos_are_bounded_before_key(self):
        import os,subprocess
        base=self.ready();self.approve_recovery(base)
        cases=[('protocol114',base/'protocol-authorization.json'),('visual-jev',base/'visual-jev-authorization.json'),('visual-decisions',base/'visual-decisions-authorization.json'),('selected',base/'selected-fifo.json')]
        for name,victim in cases:
            with self.subTest(name=name):
                raw=victim.read_bytes() if victim.exists() else None
                if victim.exists():victim.unlink()
                os.mkfifo(victim)
                args=['final-plan','.evidence/provider-sites/recovery3','--inputs',str(victim.relative_to(self.root))] if name=='selected' else ['run-selected','.evidence/provider-sites/recovery3',name]
                try:
                    try:p=subprocess.run([*self.cli,*args],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=2)
                    except subprocess.TimeoutExpired:
                        self.journal.append(dict(argv=args,state='TIMEOUT',exitCode=None));self.fail('FIFO import blocked before claim/key')
                    self.journal.append(dict(argv=args,exitCode=p.returncode,stderr=p.stderr));self.assertEqual(p.returncode,1);self.assertIn('REFUSED',p.stderr)
                finally:
                    victim.unlink()
                    if raw is not None:victim.write_bytes(raw)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-3-*')));self.assertTrue((self.root/'.env').is_dir())

def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(RecoveryRound(n) for n in loader.getTestCaseNames(RecoveryRound) if n.startswith('test_recovery_'))
