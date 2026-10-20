"""Public CLI review regressions, isolated repositories and synthetic approvals only."""
import subprocess
import sys
import unittest
from test_recovery_round import RecoveryRound
from test_recovery4_round import Recovery4Round

SWAP="""import os,pathlib,runpy,sys
cli,victim,swap,mode,out,route=sys.argv[1:];victim=pathlib.Path(victim).resolve();original=pathlib.Path.lstat;fired=False
def checked(self,*a,**kw):
 global fired
 info=original(self,*a,**kw)
 if self==victim:
  print('FINGERPRINT_TARGET_LSTAT',file=sys.stderr,flush=True)
  if swap=='yes' and not fired:
   fired=True;self.unlink();os.mkfifo(self);print('FIFO_SWAP_INJECTED',file=sys.stderr,flush=True)
 return info
pathlib.Path.lstat=checked
sys.path.insert(0,str(pathlib.Path(cli).parent))
import final_contract
legacy=final_contract.c.digest
def marked(path):
 if pathlib.Path(path)==victim:print('LEGACY_C_DIGEST_READ',file=sys.stderr,flush=True)
 return legacy(path)
final_contract.c.digest=marked
sys.argv=[cli,mode]+([out,route] if mode=='preview' else []);runpy.run_path(cli,run_name='__main__')
"""

def fingerprints(test,round_id):
    import os
    base=test.ready()
    root=test.root
    guard=root/'ai/internal/testkit/finalbudget'
    approval=root/'.evidence/provider-sites/final-acceptance-20261020/authorization.json'
    for target,mode in ((guard/'guard.go','preview'),(guard/'estimate.go','preview'),(approval,'inventory')):
        for swap in (True,False):
            with test.subTest(path=target.name,swap=swap):
                raw=target.read_bytes()
                if not swap:target.unlink();os.mkfifo(target)
                argv=[sys.executable,'-B','-c',SWAP,test.cli[-1],str(target),'yes' if swap else 'no',mode,str(base.relative_to(root)),'formal-zai-responses']
                try:
                    p=subprocess.run(argv,cwd=root,env=test.env,capture_output=True,text=True,timeout=2)
                    test.journal.append(dict(argv=argv,exitCode=p.returncode,stdout=p.stdout,stderr=p.stderr))
                    test.assertEqual(p.returncode,1);test.assertIn('FINGERPRINT_TARGET_LSTAT',p.stderr)
                    if swap:test.assertIn('FIFO_SWAP_INJECTED',p.stderr)
                except subprocess.TimeoutExpired as error:
                    test.journal.append(dict(argv=argv,exitCode=None,state='TIMEOUT',stderr=(error.stderr or b'').decode()))
                    test.fail('fingerprint read blocked on FIFO after target marker')
                finally:
                    target.unlink();target.write_bytes(raw)
    test.assertTrue((root/'.env').is_dir())
    test.assertFalse(list((root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-'+round_id+'-*')))

class ThreeReview(RecoveryRound):
    def test_review_three_fingerprints_nonblocking(self):fingerprints(self,'3')

    def test_review_three_public_process_needs_no_internal_facades(self):
        for name in ('recovery_execute','recovery_state','recovery_finalize','recovery_publication'):
            (self.tools/(name+'.py')).unlink(missing_ok=True)
        base=self.ready();self.approve_recovery(base)
        self.runcli('native-grant',str(base.relative_to(self.root)))
        (self.root/'.env').rmdir();(self.root/'.env').write_text('MINIMAX_KEY=CONTROLLED-ONLY\n')
        self.runcli('run-selected',str(base.relative_to(self.root)),'mm-global-formal-m3-chat')
        import json
        value=json.loads(self.runcli('summarize').stdout)
        self.assertEqual((value['newProcessCount'],value['reservedMicrousd']),(1,10900000))

class FourReview(Recovery4Round):
    def test_review_four_fingerprints_nonblocking(self):fingerprints(self,'4')

    def test_review_four_native_source_archive_gate_before_claim_or_key(self):
        import json
        from test_recovery4_fixture import seed,builder,reseal,BASE
        seed(self);builder(self)
        archive=self.root/BASE/'recovery3/source-at-start.json';value=json.loads(archive.read_text());value['files']['ai/sentinel.go']='f'*64;archive.write_text(json.dumps(value));reseal(self,[archive])
        self.runcli('prepare','.evidence/provider-sites/recovery4');self.runcli('check','.evidence/provider-sites/recovery4')
        base=self.root/'.evidence/provider-sites/recovery4';self.approve_recovery(base)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses',ok=False)
        self.assertFalse(list((self.root/'.evidence/minimax-launches').glob('provider-sites-final-acceptance-4-*')))
        proof=json.loads((base/'initial-native-history-command.json').read_text());self.assertEqual(proof['exitCode'],1);self.assertFalse(proof['credentialsRead'])
        self.assertIn('CONTROLLED_NATIVE_ARCHIVE_REJECT',(base/'initial-native-history.log').read_text())
        self.assertTrue((self.root/'.env').is_dir());self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4').exists())

    def test_review_four_initial_facts_bound_then_grant_is_distinct(self):
        import json,hashlib
        from test_recovery4_fixture import keys
        base=self.ready();self.approve_recovery(base);keys(self)
        self.runcli('run-selected','.evidence/provider-sites/recovery4','formal-zai-responses')
        prefix='initial-native-history';pin=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
        marker=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-4-formal-zai-responses.json';record=json.loads(marker.read_text());command=json.loads((base/'formal-zai-responses-command.json').read_text())
        for value in (record,command):
            self.assertEqual(value['nativeHistoryFactsReceiptSha256'],pin(base/(prefix+'-command.json')))
            self.assertEqual(value['nativeHistoryFactsSha256'],pin(base/(prefix+'.log')))
        initial=(base/(prefix+'-command.json')).read_bytes()
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').exists())
        log=base/(prefix+'.log');raw=log.read_bytes();log.write_text(log.read_text()+' ')
        self.runcli('native-grant','.evidence/provider-sites/recovery4',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-4/grant.json').exists());log.write_bytes(raw)
        self.runcli('native-grant','.evidence/provider-sites/recovery4')
        self.assertEqual((base/(prefix+'-command.json')).read_bytes(),initial)
        self.assertEqual(json.loads((base/'native-recovery4-facts-command.json').read_text())['exitCode'],0)

def load_tests(loader,tests,pattern):
    return unittest.TestSuite(cls(name) for cls in (ThreeReview,FourReview) for name in loader.getTestCaseNames(cls) if name.startswith('test_review_'))
