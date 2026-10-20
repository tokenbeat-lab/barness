"""Public fixed5 CLI processes; synthetic repositories only, no vendor sockets."""
import json
import subprocess
import sys
from test_recovery_round import RecoveryRound

class Recovery5Round(RecoveryRound):
    def setUp(self):
        super().setUp()
        self.cli=[sys.executable,'-B',str(self.tools/'recovery5_round.py')]

    def test_five_inventory_exact19_scope_and_dollars(self):
        v=json.loads(self.runcli('inventory').stdout)
        self.assertEqual((v['maxCalls'],v['maxAttempts'],v['campaignMaxCalls'],v['campaignMaxAttempts']),(378,414,1358,1454))
        self.assertEqual((v['newReservedMicrousd'],v['retainedReservedMicrousd'],v['maxMicrousd']),(6900000,28800000,36000000))
        self.assertEqual(len(v['jobs']),19)
        self.assertEqual(sum(j['maxCalls'] for j in v['jobs'].values()),378)
        self.assertEqual(sum(j['maxAttempts'] for j in v['jobs'].values()),414)
        self.assertEqual(sum(j['reservedMicrousd'] for j in v['jobs'].values()),6900000)
        self.assertEqual(v['executionOrder'][:4],['formal-zai-responses']+['mm-global-formal-m3-'+r for r in ('chat','responses','messages')])
        self.assertEqual(sum(j['networkRoute']=='http://127.0.0.1:7890' for j in v['jobs'].values()),13)
        for r in ('chat','responses','messages'):
            j=v['jobs']['glm-'+r];self.assertEqual((j['sourceVariable'],j['injectedCredential'],j['plannedCalls'],j['maxCalls'],j['maxAttempts']),('ZHIPU_KEY','BARNESS_GLM_KEY',22,24,32))
            self.assertEqual(j['networkRoute'],'direct')
        self.assertEqual(v['validationRound'],'provider-sites-final-acceptance-5')
        self.assertEqual(v['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        self.assertTrue((self.root/'.env').is_dir())

    def test_five_missing_history_refuses_before_output_or_key(self):
        self.fake_build();self.runcli('prepare','.evidence/provider-sites/recovery5',ok=False)
        self.assertFalse((self.root/'.evidence/provider-sites/recovery5').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def test_five_old4_all_active_modes_permanently_close(self):
        seal=self.root/'.evidence/provider-sites/final-acceptance-20261020/closed-cohort-4/seal.json';seal.parent.mkdir();seal.write_text('{}')
        old=[sys.executable,'-B',str(self.tools/'recovery4_round.py')]
        for mode in ('prepare','preview','check','native-grant','run-selected','final-plan','finalize'):
            args=[mode,'.evidence/provider-sites/no-reopen']+(['formal-zai-responses'] if mode in ('preview','run-selected') else [])
            v=subprocess.run([*old,*args],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=3)
            self.journal.append(dict(argv=args,exitCode=v.returncode,stderr=v.stderr))
            self.assertEqual(v.returncode,1);self.assertIn('permanently closed',v.stderr)
        self.assertFalse((self.root/'.evidence/provider-sites/no-reopen').exists());self.assertTrue((self.root/'.env').is_dir())

    def ready(self):
        from test_recovery5_fixture import seed,builder
        seed(self);builder(self);self.runcli('prepare','.evidence/provider-sites/recovery5')
        base=self.root/'.evidence/provider-sites/recovery5';self.runcli('check','.evidence/provider-sites/recovery5');return base

    def test_five_prepare_closed4_and_pending_without_native_or_key(self):
        base=self.ready();p=json.loads((base/'predecessor.json').read_text())
        self.assertEqual((p['actualCalls'],p['actualAttempts'],p['claimedCalls'],p['claimedAttempts'],p['reservedMicrousd']),(591,584,980,1040,28800000))
        self.assertEqual(len(p['fourthCohort']),25);self.assertEqual(p['fourthCohort']['nativeLedgerRows'],75)
        self.assertEqual(len(json.loads((base/'retained-generic.json').read_text())['roles']),8)
        self.assertEqual(json.loads((base/'recovery5-authorization.template.json').read_text())['decision'],'PENDING')
        self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-5').exists())

    def test_five_generic_production_unknown_test_sdk_catalog_drift_refuse(self):
        from test_recovery5_fixture import seed,builder
        seed(self);builder(self)
        for name in ('ai/live/recorder_test.go','ai/live/config_test.go','ai/new_unknown_test.go','ai/new_sdk.go','ai/new_catalog.json'):
            p=self.root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text('controlled unexpected source')
            v=self.runcli('prepare','.evidence/provider-sites/drift',ok=False);self.assertIn('complete execution source drift',v.stderr);self.assertFalse((self.root/'.evidence/provider-sites/drift').exists());p.unlink()
        self.assertTrue((self.root/'.env').is_dir())

    def test_five_generic_compatibility_and_old_parent_raw_mutations_refuse(self):
        from test_recovery5_fixture import seed,builder,BASE
        seed(self);builder(self)
        compat=self.root/BASE/'decisions-budget-diagnosis-dev/compatibility-eight.json';raw=compat.read_bytes();v=json.loads(raw);v['count']=83;compat.write_text(json.dumps(v));self.runcli('prepare','.evidence/provider-sites/no-proof',ok=False);compat.write_bytes(raw)
        stop=json.loads((self.root/BASE/'recovery4-stopped/index.json').read_text());p=self.root/stop['processes'][-1]['nativeBundle']/'live-report.json';p.write_text('{}');self.runcli('summarize',ok=False)
        self.assertTrue((self.root/'.env').is_dir())

    def test_five_first18_then_schema6_with_three_exact_proofs_once(self):
        from test_recovery5_fixture import approve,keys
        base=self.ready();approve(self,base);self.runcli('native-grant','.evidence/provider-sites/recovery5',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery5','mm-global-formal-m3-chat',ok=False);keys(self)
        self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses')
        self.runcli('native-grant','.evidence/provider-sites/recovery5')
        g=json.loads((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-5/grant.json').read_text());self.assertEqual(g['schema'],6);self.assertEqual(len(g['firstRouteProof']),3);self.assertEqual(len(g['fourthCohort']),25)
        self.assertEqual(g['spendingGuardSourceSha256'],json.loads((base/'prepared.json').read_text())['spendingGuardSourceSHA256'])
        self.runcli('native-grant','.evidence/provider-sites/recovery5',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery5','mm-global-formal-m3-chat')

    def test_five_old_reply_case_alias_and_36budget_before_key(self):
        from test_recovery5_fixture import approve,BASE
        import hashlib
        base=self.ready();approve(self,base);auth=base/'execution-authorization.json';original=auth.read_bytes()
        for key,value in [('maxUSD',30),('maxCalls',379),('newReservedMicrousd',6900001),('retainedReservedMicrousd',0)]:
            v=json.loads(original);v[key]=value;auth.write_text(json.dumps(v));self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False)
        auth.write_bytes(original);human=self.root/BASE/'recovery5-authorization.json';raw=human.read_bytes();v=json.loads(raw);v['requestReference']='CONTROLLED direct new4 reply';human.write_text(json.dumps(v));a=json.loads(original);a.update(requestReference=v['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(human.read_bytes()).hexdigest());auth.write_text(json.dumps(a));self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False)
        human.write_text(raw.decode().replace('"decision": "APPROVED"','"Decision": "PENDING", "decision": "APPROVED"'));self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-5-*')))

    def test_five_source_fifo_and_symlink_before_key(self):
        import os
        base=self.ready();p=base/'prepared.json';raw=p.read_bytes();p.unlink();os.mkfifo(p);self.runcli('preview','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False);p.unlink();p.symlink_to('plan.json');self.runcli('preview','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False);p.unlink();p.write_bytes(raw)
        self.assertTrue((self.root/'.env').is_dir())

    def test_five_all19_to33_retained59_and_exact27_current_raw_inputs(self):
        import shutil
        from test_recovery5_fixture import approve,keys,BASE
        base=self.ready();approve(self,base);keys(self);plan=json.loads((base/'plan.json').read_text())
        failed=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-4-openai-decisions.json';original=failed.read_bytes()
        for i,name in enumerate(plan['executionOrder']):
            self.runcli('run-selected','.evidence/provider-sites/recovery5',name)
            if i==0:self.runcli('native-grant','.evidence/provider-sites/recovery5')
        summary=json.loads(self.runcli('summarize').stdout)
        self.assertEqual((summary['actualProcessCount'],summary['effectiveRoleCount'],summary['retainedUnknownProcesses']),(59,33,4))
        self.assertEqual((summary['claimedCalls'],summary['claimedAttempts'],summary['reservedMicrousd']),(1358,1454,35700000))
        retained=json.loads((base/'retained-generic.json').read_text())['roles'];live={name:value['base'] for name,value in retained.items()};effects={}
        for name,job in plan['jobs'].items():
            if name=='protocol114':continue
            root=self.root/str(job['evidenceRoot']).replace('$REPO/','');path=root if name.startswith('visual-') else root/'controlled-session'
            if name.startswith(('text-','visual-')):effects[name]=str(path.relative_to(self.root))
            else:live[name.removeprefix('formal-').replace('mm-global-formal-m3-','minimax-').replace('glm-','zhipu-')]=str(path.relative_to(self.root))
        jev=base/'derived-jev';jev.mkdir();shutil.copytree(base/'protocol114/controlled-session',jev/'protocol');shutil.copytree(base/'visual-jev',jev/'effects');(jev/'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')));live['typesafe-classifier-jev-1.14.0']=str(jev.relative_to(self.root))
        inputs=base/'selected.json'
        for name in ('openai-chat','zhipu-responses'):
            forged=base/('forged-'+name);source=self.root/live[name] if not str(live[name]).startswith('/') else __import__('pathlib').Path(live[name]);shutil.copytree(source,forged);bad=dict(live);bad[name]=str(forged.relative_to(self.root));inputs.write_text(json.dumps(dict(live=bad,effects=effects)));self.runcli('finalize','.evidence/provider-sites/recovery5','--inputs',str(inputs.relative_to(self.root)),ok=False);self.assertFalse((base/'finalization').exists())
        inputs.write_text(json.dumps(dict(live=live,effects=effects)));v=json.loads(self.runcli('finalize','.evidence/provider-sites/recovery5','--inputs',str(inputs.relative_to(self.root))).stdout)
        proof=json.loads((base/'finalization/effective33-retained59.json').read_text());self.assertEqual(len(proof['roles']),33);self.assertEqual(len(v['liveInputs']),23);self.assertEqual(len(v['effects']),4);self.assertFalse(v['automaticRace']);self.assertEqual(failed.read_bytes(),original)
        for route in ('chat','responses','messages'):
            receipt=json.loads((self.root/('.evidence/minimax-launches/provider-sites-final-acceptance-5-glm-'+route+'.json')).read_text());self.assertEqual(set(receipt['capturePins']),{'schema','formalReportSha256','integritySha256','supportReportSha256','executionSourceSha256'})

    def test_five_first_failure_retains_whole_unknown_and_never_issues_native(self):
        from test_recovery5_fixture import approve
        base=self.ready();approve(self,base)
        self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-5-formal-zai-responses.json';raw=p.read_bytes();v=json.loads(raw);self.assertEqual(v['reservedMicrousd'],650000);self.assertEqual(v['budget']['state'],'UNKNOWN')
        self.runcli('run-selected','.evidence/provider-sites/recovery5','mm-global-formal-m3-chat',ok=False);self.runcli('native-grant','.evidence/provider-sites/recovery5',ok=False);self.runcli('run-selected','.evidence/provider-sites/recovery5','formal-zai-responses',ok=False);self.assertEqual(p.read_bytes(),raw)
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-5/grant.json').exists())

    def test_five_incomplete_glm21_keeps_whole_usd_and_stops_before_next(self):
        from test_recovery5_fixture import approve,keys
        base=self.ready();approve(self,base);keys(self);plan=json.loads((base/'plan.json').read_text())
        for i,name in enumerate(plan['executionOrder'][:9]):
            self.runcli('run-selected','.evidence/provider-sites/recovery5',name)
            if i==0:self.runcli('native-grant','.evidence/provider-sites/recovery5')
        (self.root/'.evidence/incomplete-glm').touch()
        v=self.runcli('run-selected','.evidence/provider-sites/recovery5','glm-chat',ok=False);self.assertIn('full22',v.stderr)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-5-glm-chat.json';raw=p.read_bytes();v=json.loads(raw);self.assertEqual(v['reservedMicrousd'],1000000);self.assertEqual(v['budget']['status'],'FAIL')
        self.runcli('run-selected','.evidence/provider-sites/recovery5','glm-responses',ok=False);self.runcli('run-selected','.evidence/provider-sites/recovery5','glm-chat',ok=False);self.assertEqual(p.read_bytes(),raw)

def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(Recovery5Round(n) for n in loader.getTestCaseNames(Recovery5Round) if n.startswith('test_five_'))
