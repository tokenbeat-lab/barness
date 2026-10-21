"""Public fixed7 CLI processes; synthetic repositories only, no vendor sockets."""
import json
import subprocess
import sys
from test_recovery_round import RecoveryRound

class Recovery7Round(RecoveryRound):
    def setUp(self):
        super().setUp()
        self.cli=[sys.executable,'-B',str(self.tools/'recovery7_round.py')]

    def test_seven_inventory_exact19_scope_and_dollars(self):
        v=json.loads(self.runcli('inventory').stdout)
        self.assertEqual((v['maxCalls'],v['maxAttempts'],v['campaignMaxCalls'],v['campaignMaxAttempts']),(378,414,1582,1694))
        self.assertEqual((v['newReservedMicrousd'],v['retainedReservedMicrousd'],v['maxMicrousd']),(6900000,32900000,40000000))
        self.assertEqual(len(v['jobs']),19)
        self.assertEqual(sum(j['maxCalls'] for j in v['jobs'].values()),378)
        self.assertEqual(sum(j['maxAttempts'] for j in v['jobs'].values()),414)
        self.assertEqual(sum(j['reservedMicrousd'] for j in v['jobs'].values()),6900000)
        self.assertEqual(v['executionOrder'][:4],['formal-zai-responses']+['mm-global-formal-m3-'+r for r in ('chat','responses','messages')])
        self.assertEqual(sum(j['networkRoute']=='http://127.0.0.1:7890' for j in v['jobs'].values()),13)
        for r in ('chat','responses','messages'):
            j=v['jobs']['glm-'+r];self.assertEqual((j['sourceVariable'],j['injectedCredential'],j['plannedCalls'],j['maxCalls'],j['maxAttempts']),('ZHIPU_KEY','BARNESS_GLM_KEY',22,24,32))
            self.assertEqual(j['networkRoute'],'direct')
        for route in ('chat','responses','messages'):
            j=v['jobs']['mm-global-formal-m3-'+route];self.assertEqual(j['consumerTimeoutSeconds'],600);self.assertEqual(j['controllerTimeoutSeconds'],4800);self.assertEqual((j['timeoutSeconds'],j['controllerTimeoutSeconds']),(3600,4800))
        self.assertNotIn('controllerTimeoutSeconds',v['jobs']['formal-zai-responses'])
        for script in ('recovery_round.py','recovery4_round.py','recovery5_round.py','recovery6_round.py'):
            raw=subprocess.run([sys.executable,'-B',str(self.tools/script),'inventory'],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=5);self.journal.append(dict(argv=[script,'inventory'],exitCode=raw.returncode,stderr=raw.stderr));self.assertEqual(raw.returncode,0);old=json.loads(raw.stdout)
            for name,job in old['jobs'].items():
                if name not in v['jobs']:continue
                current=v['jobs'][name]
                if name.startswith('mm-'):
                    self.assertEqual(job['timeoutSeconds'],1800);self.assertIn('-test.timeout=1800s',job['argv']);self.assertIn('-test.timeout=3600s',current['argv']);self.assertEqual(job.get('consumerTimeoutSeconds',180),360 if script=='recovery6_round.py' else 180)
                else:
                    self.assertEqual(job['timeoutSeconds'],current['timeoutSeconds']);self.assertNotIn('consumerTimeoutSeconds',current)
            self.assertNotIn('MINIMAX_FINAL_CONSUMER_TIMEOUT_SECONDS',(self.tools/(script.replace('_round.py','_contract.py'))).read_text())
        self.assertEqual(v['validationRound'],'provider-sites-final-acceptance-7')
        self.assertEqual(v['spendingGuardPolicyId'],'provider-sites-final-acceptance-2')
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_missing_history_refuses_before_output_or_key(self):
        self.fake_build();self.runcli('prepare','.evidence/provider-sites/recovery7',ok=False)
        self.assertFalse((self.root/'.evidence/provider-sites/recovery7').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_old6_all_active_modes_permanently_close(self):
        seal=self.root/'.evidence/provider-sites/final-acceptance-20261020/closed-cohort-6/seal.json';seal.parent.mkdir();seal.write_text('{}')
        old=[sys.executable,'-B',str(self.tools/'recovery6_round.py')]
        for mode in ('prepare','preview','check','native-grant','run-selected','final-plan','finalize'):
            args=[mode,'.evidence/provider-sites/no-reopen']+(['formal-zai-responses'] if mode in ('preview','run-selected') else [])
            v=subprocess.run([*old,*args],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=3)
            self.journal.append(dict(argv=args,exitCode=v.returncode,stderr=v.stderr))
            self.assertEqual(v.returncode,1);self.assertIn('permanently closed',v.stderr)
        self.assertFalse((self.root/'.evidence/provider-sites/no-reopen').exists());self.assertTrue((self.root/'.env').is_dir())

    def ready(self):
        from test_recovery7_fixture import seed,builder
        seed(self);builder(self);self.runcli('prepare','.evidence/provider-sites/recovery7')
        base=self.root/'.evidence/provider-sites/recovery7';self.runcli('check','.evidence/provider-sites/recovery7');return base

    def test_seven_prepare_closed6_and_pending_without_native_or_key(self):
        base=self.ready();p=json.loads((base/'predecessor.json').read_text())
        self.assertEqual((p['actualCalls'],p['actualAttempts'],p['claimedCalls'],p['claimedAttempts'],p['reservedMicrousd']),(724,714,1204,1280,32900000))
        self.assertEqual(len(p['sixthCohort']),25);self.assertEqual(p['sixthCohort']['interruptedNativeLedgerRows'],8)
        self.assertEqual(len(json.loads((base/'retained-generic.json').read_text())['roles']),8)
        self.assertEqual(json.loads((base/'recovery7-authorization.template.json').read_text())['decision'],'PENDING')
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual((summary['knownReportedCalls'],summary['knownReportedAttempts']),(720,710));self.assertEqual((summary['knownRecordedCalls'],summary['knownRecordedHTTPAllocations']),(724,714));self.assertTrue(summary['recordedHTTPAllocationsAreNotConfirmedVendorSends'])
        old=subprocess.run([sys.executable,'-B',str(self.tools/'recovery6_round.py'),'summarize'],cwd=self.root,env=self.env,capture_output=True,text=True,timeout=10);self.journal.append(dict(argv=['recovery6_round.py','summarize'],exitCode=old.returncode,stderr=old.stderr));self.assertEqual(old.returncode,0);self.assertEqual(json.loads(old.stdout)['sixthCohort'],p['sixthCohort'])
        self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-7').exists())

    def test_seven_generic_production_unknown_test_sdk_catalog_drift_refuse(self):
        from test_recovery7_fixture import seed,builder
        seed(self);builder(self)
        for name in ('ai/live/recorder_test.go','ai/live/config_test.go','ai/new_unknown_test.go','ai/new_sdk.go','ai/new_catalog.json'):
            p=self.root/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text('controlled unexpected source')
            v=self.runcli('prepare','.evidence/provider-sites/drift',ok=False);self.assertIn('source',v.stderr);self.assertFalse((self.root/'.evidence/provider-sites/drift').exists());p.unlink()
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_generic_compatibility_and_old_parent_raw_mutations_refuse(self):
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self)
        compat=self.root/BASE/'recovery7-preparation/compatibility-eight.json';raw=compat.read_bytes();v=json.loads(raw);v['count']=83;compat.write_text(json.dumps(v));self.runcli('prepare','.evidence/provider-sites/no-proof',ok=False);compat.write_bytes(raw)
        stop=json.loads((self.root/BASE/'recovery6-stopped/index.json').read_text());p=self.root/stop['processes'][-1]['parent'];p.write_text('{}');self.runcli('summarize',ok=False)
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_first18_then_schema8_with_three_exact_proofs_once(self):
        from test_recovery7_fixture import approve,keys
        base=self.ready();approve(self,base);self.runcli('native-grant','.evidence/provider-sites/recovery7',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery7','mm-global-formal-m3-chat',ok=False);keys(self)
        self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses')
        self.runcli('native-grant','.evidence/provider-sites/recovery7')
        g=json.loads((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-7/grant.json').read_text());self.assertEqual(g['schema'],8);self.assertEqual(len(g['firstRouteProof']),3);self.assertEqual(len(g['sixthCohort']),25)
        self.assertEqual(g['spendingGuardSourceSha256'],json.loads((base/'prepared.json').read_text())['spendingGuardSourceSHA256'])
        self.runcli('native-grant','.evidence/provider-sites/recovery7',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/recovery7','mm-global-formal-m3-chat')
        command=json.loads((base/'mm-global-formal-m3-chat-consume-command.json').read_text());self.assertEqual(command['timeoutSeconds'],600)
        regional=json.loads((base/'formal-zai-responses-consume-command.json').read_text());self.assertEqual(regional['timeoutSeconds'],180)
        plan=json.loads((base/'plan.json').read_text());self.assertEqual(plan['jobs']['mm-global-formal-m3-chat']['consumerTimeoutSeconds'],600);self.assertNotIn('consumerTimeoutSeconds',plan['jobs']['formal-zai-responses'])

    def test_seven_old_reply_case_alias_and_39budget_before_key(self):
        from test_recovery7_fixture import approve,BASE
        import hashlib
        base=self.ready();approve(self,base);auth=base/'execution-authorization.json';original=auth.read_bytes()
        for key,value in [('maxUSD',39),('maxCalls',379),('newReservedMicrousd',6900001),('retainedReservedMicrousd',0)]:
            v=json.loads(original);v[key]=value;auth.write_text(json.dumps(v));self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False)
        auth.write_bytes(original);human=self.root/BASE/'recovery7-authorization.json';raw=human.read_bytes();v=json.loads(raw);v['requestReference']='CONTROLLED independent sixth reply';human.write_text(json.dumps(v));a=json.loads(original);a.update(requestReference=v['requestReference'],actualHumanApprovalSHA256=hashlib.sha256(human.read_bytes()).hexdigest());auth.write_text(json.dumps(a));self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False)
        human.write_text(raw.decode().replace('"decision": "APPROVED"','"Decision": "PENDING", "decision": "APPROVED"'));self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-7-*')))

    def test_seven_source_fifo_and_symlink_before_key(self):
        import os
        base=self.ready();p=base/'prepared.json';raw=p.read_bytes();p.unlink();os.mkfifo(p);self.runcli('preview','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False);p.unlink();p.symlink_to('plan.json');self.runcli('preview','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False);p.unlink();p.write_bytes(raw)
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_all19_to33_retained69_and_exact27_current_raw_inputs(self):
        import shutil
        from test_recovery7_fixture import approve,keys,BASE
        base=self.ready();approve(self,base);keys(self);plan=json.loads((base/'plan.json').read_text())
        failed=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-6-mm-global-formal-m3-chat.json';original=failed.read_bytes()
        for i,name in enumerate(plan['executionOrder']):
            self.runcli('run-selected','.evidence/provider-sites/recovery7',name)
            if i==0:self.runcli('native-grant','.evidence/provider-sites/recovery7')
        summary=json.loads(self.runcli('summarize').stdout)
        self.assertEqual((summary['actualProcessCount'],summary['effectiveRoleCount'],summary['retainedUnknownProcesses']),(69,33,6))
        self.assertEqual((summary['claimedCalls'],summary['claimedAttempts'],summary['reservedMicrousd']),(1582,1694,39800000))
        retained=json.loads((base/'retained-generic.json').read_text())['roles'];live={name:value['base'] for name,value in retained.items()};effects={}
        for name,job in plan['jobs'].items():
            if name=='protocol114':continue
            root=self.root/str(job['evidenceRoot']).replace('$REPO/','');path=root if name.startswith('visual-') else root/'controlled-session'
            if name.startswith(('text-','visual-')):effects[name]=str(path.relative_to(self.root))
            else:live[name.removeprefix('formal-').replace('mm-global-formal-m3-','minimax-').replace('glm-','zhipu-')]=str(path.relative_to(self.root))
        jev=base/'derived-jev';jev.mkdir();shutil.copytree(base/'protocol114/controlled-session',jev/'protocol');shutil.copytree(base/'visual-jev',jev/'effects');(jev/'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')));live['typesafe-classifier-jev-1.14.0']=str(jev.relative_to(self.root))
        inputs=base/'selected.json'
        for name in ('openai-chat','zhipu-responses'):
            forged=base/('forged-'+name);source=self.root/live[name] if not str(live[name]).startswith('/') else __import__('pathlib').Path(live[name]);shutil.copytree(source,forged);bad=dict(live);bad[name]=str(forged.relative_to(self.root));inputs.write_text(json.dumps(dict(live=bad,effects=effects)));self.runcli('finalize','.evidence/provider-sites/recovery7','--inputs',str(inputs.relative_to(self.root)),ok=False);self.assertFalse((base/'finalization').exists())
        inputs.write_text(json.dumps(dict(live=live,effects=effects)));v=json.loads(self.runcli('finalize','.evidence/provider-sites/recovery7','--inputs',str(inputs.relative_to(self.root))).stdout)
        proof=json.loads((base/'finalization/effective33-retained69.json').read_text());self.assertEqual(len(proof['roles']),33);self.assertEqual(len(v['liveInputs']),23);self.assertEqual(len(v['effects']),4);self.assertFalse(v['automaticRace']);self.assertEqual(failed.read_bytes(),original)
        final_plan=json.loads((base/'finalization/plan.json').read_text());self.assertEqual(final_plan['verifyTimeoutSeconds'].count(360),3)
        for i,argv in enumerate(final_plan['verify']):
            receipt=json.loads((base/'finalization'/('verify-'+str(i)+'-command.json')).read_text());self.assertEqual(receipt['timeoutSeconds'],360 if argv[0].endswith('/minimaxverify') else 180)
        for route in ('chat','responses','messages'):
            receipt=json.loads((self.root/('.evidence/minimax-launches/provider-sites-final-acceptance-7-glm-'+route+'.json')).read_text());self.assertEqual(set(receipt['capturePins']),{'schema','formalReportSha256','integritySha256','supportReportSha256','executionSourceSha256'})

    def test_seven_first_failure_retains_whole_unknown_and_never_issues_native(self):
        from test_recovery7_fixture import approve
        base=self.ready();approve(self,base)
        self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-7-formal-zai-responses.json';raw=p.read_bytes();v=json.loads(raw);self.assertEqual(v['reservedMicrousd'],650000);self.assertEqual(v['budget']['state'],'UNKNOWN')
        ledger=json.loads((base/'formal-zai-responses-usd.json').read_text());self.assertIn('Corrected operation-specific guard',ledger['basis']);self.assertIn('executionAuthority=provider-sites-final-acceptance-7',ledger['basis'])
        for sha in json.loads((base/'prepared.json').read_text())['spendingGuardSourceSHA256'].values():self.assertIn(sha,ledger['basis'])
        self.runcli('run-selected','.evidence/provider-sites/recovery7','mm-global-formal-m3-chat',ok=False);self.runcli('native-grant','.evidence/provider-sites/recovery7',ok=False);self.runcli('run-selected','.evidence/provider-sites/recovery7','formal-zai-responses',ok=False);self.assertEqual(p.read_bytes(),raw)
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-7/grant.json').exists())

    def test_seven_incomplete_glm21_keeps_whole_usd_and_stops_before_next(self):
        from test_recovery7_fixture import approve,keys
        base=self.ready();approve(self,base);keys(self);plan=json.loads((base/'plan.json').read_text())
        for i,name in enumerate(plan['executionOrder'][:9]):
            self.runcli('run-selected','.evidence/provider-sites/recovery7',name)
            if i==0:self.runcli('native-grant','.evidence/provider-sites/recovery7')
        (self.root/'.evidence/incomplete-glm').touch()
        v=self.runcli('run-selected','.evidence/provider-sites/recovery7','glm-chat',ok=False);self.assertIn('full22',v.stderr)
        p=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-7-glm-chat.json';raw=p.read_bytes();v=json.loads(raw);self.assertEqual(v['reservedMicrousd'],1000000);self.assertEqual(v['budget']['status'],'FAIL')
        self.runcli('run-selected','.evidence/provider-sites/recovery7','glm-responses',ok=False);self.runcli('run-selected','.evidence/provider-sites/recovery7','glm-chat',ok=False);self.assertEqual(p.read_bytes(),raw)

    def test_seven_legitimate_large_stop_and_refused_symlink_fixture(self):
        import hashlib,re
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self)
        stop=self.root/BASE/'recovery6-stopped/index.json';v=json.loads(stop.read_text());before=hashlib.sha256(stop.read_bytes()).hexdigest()
        v['retainedIndexPadding']='controlled bounded original index '*70000;stop.write_text(json.dumps(v));after=hashlib.sha256(stop.read_bytes()).hexdigest()
        seal=self.root/BASE/'closed-cohort-6/seal.json';v=json.loads(seal.read_text());v['stopIndexSHA256']=after;seal.write_text(json.dumps(v))
        p=self.tools/'recovery7_history.py';p.write_text(p.read_text().replace(before,after));self.commit()
        self.runcli('prepare','.evidence/provider-sites/large-index')
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_headers_and_precise_reviewed_source_proof_required(self):
        import hashlib
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self);proof=self.root/BASE/'recovery7-preparation/compatibility-eight.json';original=proof.read_bytes();old=hashlib.sha256(original).hexdigest()
        for mode in ('headers','headers-content','duplicate-case','foreign-capture','unknown-source','missing-reviewed','wrong-blob'):
            v=json.loads(original);changed=None
            if mode=='headers-content':
                row=v['cases'][0];wire=self.root/row['currentWireFile'];changed=(wire,wire.read_bytes());data=json.loads(changed[1]);data[0]['requestHeaders']['accept']='application/changed';wire.write_text(json.dumps(data));row['currentWireFileSHA256']=hashlib.sha256(wire.read_bytes()).hexdigest()
            elif mode=='duplicate-case':v['cases'][-1]=v['cases'][0]
            elif mode=='foreign-capture':
                row=v['cases'][0];foreign_source=row['source'].rsplit(':',1)[0];v['originalSourceFiles'].pop(foreign_source);stop=json.loads((self.root/BASE/'recovery4-stopped/index.json').read_text());foreign=stop['processes'][0]['nativeBundle']+'/regional-report.json';v['originalSourceFiles'][foreign]=stop['files'][foreign];row['source']=foreign+':0'
            elif mode=='headers':v['cases'][0]['requestHeadersBytesEqual']=False
            elif mode=='unknown-source':v['reviewedSourceDifferences']['ai/unknown_test.go']='f'*64
            elif mode=='missing-reviewed':v.pop('reviewedSourceDifferences')
            else:v['reviewedSourceDifferences']['ai/live/regional_harness_test.go']='f'*64
            proof.write_text(json.dumps(v));new=hashlib.sha256(proof.read_bytes()).hexdigest();p=self.tools/'recovery7_retained.py';source=p.read_text();p.write_text(source.replace(old,new));self.commit()
            self.runcli('prepare','.evidence/provider-sites/proof-'+mode,ok=False);self.assertFalse((self.root/'.evidence/provider-sites'/('proof-'+mode)).exists());p.write_text(source)
            if changed:changed[0].write_bytes(changed[1])
        self.assertTrue((self.root/'.env').is_dir())

    def test_seven_closed6_extra_regular_link_fifo_refuse(self):
        import os
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self);p=self.root/BASE/'recovery6/unpinned-original.json'
        for kind in ('regular','link','fifo'):
            if kind=='regular':p.write_text('{}')
            elif kind=='link':p.symlink_to('source-at-start.json')
            else:os.mkfifo(p)
            self.runcli('prepare','.evidence/provider-sites/unpinned-'+kind,ok=False);self.assertFalse((self.root/'.evidence/provider-sites'/('unpinned-'+kind)).exists());p.unlink()
        self.assertTrue((self.root/'.env').is_dir())


    def test_seven_interrupted_originals_and_missing_report_remain_honest(self):
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self);stop=json.loads((self.root/BASE/'recovery6-stopped/index.json').read_text())
        paths=[self.root/stop['m3Raw']/'minimax-report.json',self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-6/responses.json']
        for p in paths:
            p.write_text('{}');self.runcli('summarize',ok=False);p.unlink()
        ledger=self.root/BASE/'recovery6/mm-global-formal-m3-chat-usd.json';raw=ledger.read_bytes();v=json.loads(raw);v['reservedMicrousd']=0;ledger.write_text(json.dumps(v));self.runcli('prepare','.evidence/provider-sites/refund-forgery',ok=False);ledger.write_bytes(raw)
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-7').exists())

    def test_seven_pending_actual84_proof_refuses_before_output_or_key(self):
        from test_recovery7_fixture import seed,builder,BASE
        seed(self);builder(self);p=self.tools/'recovery7_retained.py';text=p.read_text();import re
        p.write_text(re.sub(r"COMPAT_SHA256='[^']+'", "COMPAT_SHA256='PENDING_CURRENT_GO_84_REPROOF'",text));self.commit()
        v=self.runcli('prepare','.evidence/provider-sites/pending-proof',ok=False);self.assertIn('pending',v.stderr);self.assertFalse((self.root/'.evidence/provider-sites/pending-proof').exists());self.assertTrue((self.root/'.env').is_dir())

def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(Recovery7Round(n) for n in loader.getTestCaseNames(Recovery7Round) if n.startswith('test_seven_'))
