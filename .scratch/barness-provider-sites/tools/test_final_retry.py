"""Public fixed amendment processes use synthetic keys and a separate Git owner."""
import json,os,sys,unittest
from pathlib import Path
from test_final_round import FinalRound
import test_retry_fixture as fixture
class ResponsesRetry(FinalRound):
    def setUp(self):
        super().setUp();self.fake_build(regional_reports=True)
    def tearDown(self):
        self.doCleanups();super().tearDown()
    def prepare_retry(self,count=22):
        self.checkout,self.old=fixture.seed(self,count)
        self.runcli('prepare-responses-retry','.evidence/provider-sites/retry','--execution-root',str(self.checkout))
        self.out=self.root/'.evidence/provider-sites/retry'
        self.runcli('check','.evidence/provider-sites/retry');self.approve(self.out,reserve=500000)
        auth=json.loads((self.out/'retry-authorization.template.json').read_text());auth['decision']='APPROVED';(self.out/'retry-authorization.json').write_text(json.dumps(auth))
    def test_retry_once_keeps_failed_budget_and_truthful_native_commit(self):
        self.prepare_retry();failed=(self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses.json').read_bytes()
        (self.root/'.env').rmdir();(self.root/'.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')
        self.runcli('run-responses-retry','.evidence/provider-sites/retry')
        self.runcli('run-responses-retry','.evidence/provider-sites/retry',ok=False)
        self.assertEqual((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses.json').read_bytes(),failed)
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual(summary['reservedMicrousd'],7000000);self.assertEqual(summary['claimedCalls'],268)
        marker=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses-retry-1.json'
        measured=json.loads(marker.read_text());prepared=json.loads((self.out/'prepared.json').read_text())
        self.assertNotEqual(measured['sourceCommit'],prepared['sourceCommit'])
        self.assertEqual(measured['controlSourceCommit'],prepared['sourceCommit'])
        self.assertEqual(json.loads((self.out/'glm-responses-retry-1/session/formal-glm-report.json').read_text())['gitCommit'],measured['sourceCommit'])
        self.runcli('run-selected','.evidence/provider-sites/retry','glm-messages')
        summary=json.loads(self.runcli('summarize').stdout);self.assertEqual((summary['reservedMicrousd'],summary['claimedCalls'],summary['claimedAttempts']),(7500000,292,324))
        self.assertEqual(summary['unknownProcesses'],1)
        argv=[sys.executable,'-B',str(self.root/'.scratch/barness-ai-minimax/tools/renewal.py'),'glm-archive','.evidence/minimax-retry-published','--round','provider-sites-final-acceptance-2','--chat',str((self.old/'glm-chat').relative_to(self.root)),'--responses',str((self.out/'glm-responses-retry-1').relative_to(self.root)),'--messages',str((self.out/'glm-messages').relative_to(self.root))]
        import subprocess
        result=subprocess.run(argv,cwd=self.root,env=self.env,text=True,capture_output=True,timeout=30)
        self.journal.append(dict(argv=argv,exitCode=result.returncode,stdout=result.stdout,stderr=result.stderr));self.assertEqual(result.returncode,0,result.stderr)
        index=json.loads((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_text())
        self.assertEqual(index['originalLiveSourceCommit'],measured['sourceCommit'])
        self.assertEqual(index['launcherReceiptSHA256']['responses'],__import__('hashlib').sha256(marker.read_bytes()).hexdigest())

    def test_retry_approved_summary_preserves_known_failed_actuals_before_claim(self):
        self.prepare_retry();summary=json.loads(self.runcli('summarize').stdout)
        self.assertEqual((summary['campaignMaxCalls'],summary['campaignMaxAttempts']),(798,842))
        self.assertEqual((summary['knownReportedCalls'],summary['knownReportedAttempts']),(165,165))
        self.assertEqual(summary['reservedMicrousd'],6500000)
    def test_retry_messages_cannot_resume_from_unmeasured_checkout(self):
        self.prepare_retry();(self.root/'.env').rmdir();(self.root/'.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')
        self.runcli('run-responses-retry','.evidence/provider-sites/retry')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.runcli('prepare','.evidence/provider-sites/unsafe','--phase','glm');unsafe=self.root/'.evidence/provider-sites/unsafe'
        self.runcli('check','.evidence/provider-sites/unsafe');self.approve(unsafe,reserve=500000)
        self.runcli('run-selected','.evidence/provider-sites/unsafe','glm-messages',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-messages.json').exists())
        self.assertTrue((self.root/'.env').is_dir())
    def test_retry_native_source_drift_refuses_before_claim_or_secret(self):
        self.prepare_retry();(self.checkout/'ai/sentinel.go').write_text('package changed\n')
        self.runcli('run-responses-retry','.evidence/provider-sites/retry',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses-retry-1.json').exists())
        self.assertTrue((self.root/'.env').is_dir())
    def test_retry_partial22_and_fake_replacement_cannot_continue(self):
        self.prepare_retry(21);(self.root/'.env').rmdir();(self.root/'.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')
        self.runcli('run-selected','.evidence/provider-sites/retry','glm-messages',ok=False)
        self.runcli('run-responses-retry','.evidence/provider-sites/retry',ok=False)
        self.runcli('run-selected','.evidence/provider-sites/retry','glm-messages',ok=False)
    def test_retry_budget_reset_and_altered_failed_claim_refuse(self):
        self.prepare_retry();auth=self.out/'retry-authorization.json';value=json.loads(auth.read_text());value['reservedMicrousd']=1;auth.write_text(json.dumps(value))
        self.runcli('run-responses-retry','.evidence/provider-sites/retry',ok=False)
        self.assertTrue((self.root/'.env').is_dir())
    def test_retry_altered_failure_and_forged_success_cannot_replace(self):
        self.prepare_retry()
        failed=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses.json';before=failed.read_bytes();value=json.loads(before);value['reservedMicrousd']=0;failed.write_text(json.dumps(value))
        self.runcli('run-responses-retry','.evidence/provider-sites/retry',ok=False);failed.write_bytes(before)
        fake=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-2-glm-responses-retry-1.json';fake.write_text(json.dumps(dict(state='COMPLETED',exitCode=0,budget=dict(state='REPORTED',status='PASS',calls=22,attempts=22))))
        self.runcli('run-selected','.evidence/provider-sites/retry','glm-messages',ok=False)
        self.assertTrue((self.root/'.env').is_dir())
def load_tests(loader,tests,pattern):
    return unittest.TestSuite(ResponsesRetry(n) for n in loader.getTestCaseNames(ResponsesRetry) if n.startswith('test_retry_'))
if __name__=='__main__':unittest.main()
