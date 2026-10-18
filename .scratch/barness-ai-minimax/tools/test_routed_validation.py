#!/usr/bin/env python3
"""Public routed-launcher CLI checks; synthetic accounts and local processes."""
import json
import hashlib
import shutil
import os
from pathlib import Path
import subprocess
import sys
import unittest
import test_paid as fixtures


class Routed(unittest.TestCase):
    setUp = fixtures.Launcher.setUp
    tearDown = fixtures.Launcher.tearDown
    invoke = fixtures.Launcher.invoke
    history = fixtures.Launcher.history

    def setup_round(self):
        counts={'openai-responses':15,'openai-chat':17,'anthropic-messages':11,
            'google-gemini':11,'deepseek-responses':10,'deepseek-chat':13,
            'openai-images':2,'google-interactions-image':4,'openai-decisions':9,
            'typesafe-classifier':3,'protocol114':1,'text-jev':24,'text-luna':24,
            'visual-jev':12,'visual-decisions':12}
        for prefix in ('glm-','glm-corrective-1-'):
            counts.update({prefix+r:22 for r in ('chat','responses','messages')})
        for name,count in counts.items():
            value=dict(state='COMPLETED',route=name,sourceCommit='8'*40,
                credentialsRead=True,launcherAudit={'passed':True},reportedFileSha256='a'*64,
                budget=dict(state='REPORTED',status='FAIL',calls=count,attempts=count))
            (self.root/'.evidence/minimax-launches'/(name+'.json')).write_text(json.dumps(value))
        self.cli=[sys.executable,str(self.tools/'routed_validation.py')]

    def producer(self):
        (self.bin/'producer').write_text('#!'+sys.executable+'\nimport json,os,pathlib,subprocess,time\n'
            'time.sleep(0.1)\nr=os.environ["BARNESS_GLM_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\n'
            'd=dict(schema=1,kind="glm-formal",mode="live",route=r,outcome="PASS",complete=True,auditPassed=True,gitCommit=subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip(),budget=dict(callsUsed=22,httpAttempts=22,attemptsReserved=22))\n'
            '(p/"formal-glm-report.json").write_text(json.dumps(d))\n'
            'for n in ("glm-integrity.json","live-report.json","glm-execution-source.json"):(p/n).write_text("{}")\n'
            'print(json.dumps(dict(os.environ)))\n')

    def prepared(self,out='glm'):
        self.setup_round();self.producer()
        self.invoke('prepare',out,'--phase','glm');self.invoke('check',out)

    def approve(self,out='glm'):
        base=self.root/out
        value=json.loads((base/'prerelease-source-review.template.json').read_text())
        value.update(decision='APPROVED',reviewedBy='controlled-independent-reviewer',
            authorizedRoutes=['routed-validation-1-glm-'+r for r in ('chat','responses','messages')],
            credentialScopes={'GLM_KEY':'routed-validation-1/china-direct'})
        review=base/'prerelease-source-review.json';review.write_text(json.dumps(value))
        value=json.loads((base/'execution-authorization.template.json').read_text())
        value.update(decision='APPROVED',authority='explicit-user-request',requestReference='synthetic request fixture',
            sourceReviewSHA256=hashlib.sha256(review.read_bytes()).hexdigest())
        (base/'execution-authorization.json').write_text(json.dumps(value))

    def test_fixed_two_phase_plan_is_keyless_and_routes_by_service_region(self):
        self.setup_round();(self.root/'.env').mkdir()
        for phase,count in (('glm',3),('legacy',15)):
            self.invoke('prepare',phase,'--phase',phase)
            plan=json.loads((self.root/phase/'plan.json').read_text())
            self.assertEqual(plan['round'],'routed-validation-1')
            self.assertEqual(len(plan['routes']),count)
            self.assertEqual(plan['budget'],dict(priorCalls=300,priorAttempts=300,
                additionalMaxCalls=374,additionalMaxAttempts=398,totalMaxCalls=674,totalMaxAttempts=698))
            self.assertFalse(plan['automaticProcessReruns'])
            for name,job in plan['routes'].items():
                self.assertEqual(name,'routed-validation-1-'+job['baseName'])
                self.assertNotIn('HTTP_PROXY',job['environment'])
                if job['provider'] in ('zhipu','deepseek'):
                    self.assertEqual(job['networkRoute'],'china-direct')
                    self.assertNotIn('BARNESS_AI_LIVE_PROXY',job['environment'])
                    self.assertNotIn('-proxy',job['argv'])
                else:
                    self.assertEqual(job['networkRoute'],'international-loopback-proxy')
                    if job['baseName'].startswith(('text-','visual-')):
                        self.assertEqual(job['argv'][job['argv'].index('-proxy')+1],'http://127.0.0.1:7890')
                    else:self.assertEqual(job['environment']['BARNESS_AI_LIVE_PROXY'],'http://127.0.0.1:7890')
            preview=json.loads(self.invoke('preview',phase,next(iter(plan['routes']))).stdout)
            self.assertFalse(preview['credentialsRead'])
        self.invoke('prepare','bad','--phase','single',ok=False)

    def test_review_request_and_exact_freeze_precede_key_and_intent(self):
        self.prepared();(self.root/'.env').mkdir()
        name='routed-validation-1-glm-chat';intent=self.root/'.evidence/minimax-launches'/(name+'.json')
        self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists())
        self.approve()
        base=self.root/'glm'
        for leaf,field,value in (
            ('execution-authorization.json','authority','explicit-numeric-reply'),
            ('execution-authorization.json','checksSHA256','0'*64),
            ('execution-authorization.json','sourceReviewSHA256','0'*64),
            ('execution-authorization.json','budget',{}),
            ('prerelease-source-review.json','decision','PENDING')):
            path=base/leaf;raw=path.read_bytes();doc=json.loads(raw);doc[field]=value;path.write_text(json.dumps(doc))
            self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists());path.write_bytes(raw)
        for path in (self.root/'ai/runtime.go',base/'live.test',base/'plan.json',
                     self.root/'.evidence/minimax-launches/glm-responses.json'):
            raw=path.read_bytes();path.write_bytes(raw+b' ')
            self.invoke('run-selected','glm',name,ok=False);self.assertFalse(intent.exists());path.write_bytes(raw)
        (self.root/'.env').rmdir();(self.root/'.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')
        old={p.name:p.read_bytes() for p in intent.parent.glob('*.json')}
        self.invoke('run-selected','glm',name)
        record=json.loads(intent.read_text())
        self.assertEqual(record['validationRound'],'routed-validation-1')
        self.assertEqual((record['budget']['calls'],record['budget']['attempts']),(22,22))
        self.assertIn('capturePins',record)
        log=(base/'glm-chat.log').read_text()
        self.assertIn('BARNESS_GLM_KEY',log);self.assertNotIn('SYNTHETIC-SELECTED',log)
        self.assertNotIn('HTTP_PROXY',log);self.assertNotIn('MINIMAX',log)
        for leaf,raw in old.items():self.assertEqual((intent.parent/leaf).read_bytes(),raw)

    def test_concurrent_and_cross_directory_launch_cannot_reset_unknown_spend(self):
        self.prepared();self.approve();(self.root/'.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')
        name='routed-validation-1-glm-chat';args=[*self.cli,'run-selected','glm',name]
        a=subprocess.Popen(args,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        b=subprocess.Popen(args,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        a.communicate(timeout=10);b.communicate(timeout=10)
        self.assertEqual(sorted([a.returncode,b.returncode]),[0,1])
        summary=json.loads(self.invoke('summarize','glm').stdout)
        self.assertEqual((summary['current']['calls'],summary['total']['calls']),(22,322))
        report=self.root/'glm/glm-chat/session/formal-glm-report.json';raw=report.read_bytes();report.write_bytes(raw+b' ')
        summary=json.loads(self.invoke('summarize','glm').stdout)
        self.assertEqual((summary['current']['conservativeCalls'],summary['current']['conservativeAttempts']),(24,32))
        report.write_bytes(raw)
        self.invoke('prepare','elsewhere','--phase','glm');self.invoke('check','elsewhere');self.approve('elsewhere')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','elsewhere',name,ok=False)
        self.assertFalse((self.root/'elsewhere/glm-chat-command.json').exists())
        self.invoke('run-selected','glm','routed-validation-2-glm-chat',ok=False)
        # A separate checkout shares only the common Git authority directory.
        # Its new source/output/review cannot resurrect the spent canonical ID.
        alternate=self.root/'synthetic-worktree'
        subprocess.run(['git','worktree','add','--detach',str(alternate),'HEAD'],cwd=self.root,env=self.env,check=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        alternate_cli=[sys.executable,str(alternate/'.scratch/barness-ai-minimax/tools/routed_validation.py')]
        for args in (('prepare','foreign','--phase','glm'),('check','foreign')):
            result=subprocess.run([*alternate_cli,*args],cwd=alternate,env=self.env,text=True,capture_output=True,timeout=10)
            self.assertEqual(result.returncode,0,result.stderr[-500:])
        result=subprocess.run([*alternate_cli,'summarize','foreign'],cwd=alternate,env=self.env,text=True,capture_output=True,timeout=10)
        self.assertEqual(result.returncode,0,result.stderr[-500:])
        foreign=next(row for row in json.loads(result.stdout)['routes'] if row['route']==name)
        self.assertEqual((foreign['state'],foreign['conservativeCalls'],foreign['conservativeAttempts']),('UNKNOWN',24,32))
        primary=self.root
        try:self.root=alternate;self.approve('foreign')
        finally:self.root=primary
        result=subprocess.run([*alternate_cli,'run-selected','foreign',name],cwd=alternate,env=self.env,text=True,capture_output=True,timeout=10)
        self.assertNotEqual(result.returncode,0)
        self.assertFalse((alternate/'foreign/glm-chat-command.json').exists())

    def test_legacy_phase_repeats_both_deepseek_jobs_on_a_fresh_source(self):
        self.setup_round();(self.root/'.env').mkdir()
        self.invoke('prepare','glm','--phase','glm')
        source=self.root/'ai/release/new-view.json';source.write_text('{"fresh":true}')
        subprocess.run(['git','add','ai/release/new-view.json'],cwd=self.root,env=self.env,check=True)
        subprocess.run(['git','-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','committed fresh GLM selectors'],cwd=self.root,env=self.env,check=True)
        self.invoke('check','glm',ok=False)
        self.invoke('prepare','legacy','--phase','legacy')
        plan=json.loads((self.root/'legacy/plan.json').read_text())
        self.assertEqual(len(plan['routes']),15)
        for route in ('deepseek-responses','deepseek-chat'):
            job=plan['routes']['routed-validation-1-'+route]
            self.assertEqual(job['evidenceRoot'],'$REPO/legacy/live/'+route)
            self.assertEqual(job['networkRoute'],'china-direct')

    def test_cross_phase_settlement_reopens_pinned_original_proof(self):
        self.prepared();self.approve();(self.root/'.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')
        name='routed-validation-1-glm-chat';self.invoke('run-selected','glm',name)
        source=self.root/'ai/release/fresh-view.json';source.write_text('{"fresh":true}')
        subprocess.run(['git','add','ai/release/fresh-view.json'],cwd=self.root,env=self.env,check=True)
        subprocess.run(['git','-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','fresh selectors after GLM'],cwd=self.root,env=self.env,check=True,stdout=subprocess.PIPE)
        self.invoke('prepare','legacy','--phase','legacy')
        def row():
            summary=json.loads(self.invoke('summarize','legacy').stdout)
            return next(r for r in summary['routes'] if r['route']==name)
        self.assertEqual((row()['calls'],row()['attempts']),(22,22))
        base=self.root/'glm'
        for leaf in ('glm-chat/session/formal-glm-report.json','glm-chat-command.json',
                     'source-at-start.json','plan.json','binaries.json','execution-authorization.json','prerelease-source-review.json','checks.json'):
            path=base/leaf;raw=path.read_bytes()
            for mode in ('changed','missing','symlink'):
                if mode=='changed':path.write_bytes(raw+b' ')
                elif mode=='missing':path.unlink()
                else:
                    target=base/'exact-copy.json';target.write_bytes(raw);path.unlink();path.symlink_to(target)
                actual=row()
                self.assertEqual((actual['state'],actual['conservativeCalls'],actual['conservativeAttempts']),('UNKNOWN',24,32),(leaf,mode))
                if path.is_symlink():path.unlink()
                path.write_bytes(raw)
        path=base/'glm-chat/session/formal-glm-report.json';raw=path.read_bytes()
        path.write_bytes(b'x'*((16<<20)+1))
        self.assertEqual((row()['state'],row()['conservativeCalls'],row()['conservativeAttempts']),('UNKNOWN',24,32))
        path.unlink();os.mkfifo(path)
        self.assertEqual((row()['state'],row()['conservativeCalls'],row()['conservativeAttempts']),('UNKNOWN',24,32))
        path.unlink();path.write_bytes(raw)
        intent=self.root/'.evidence/minimax-launches'/(name+'.json');original=intent.read_bytes()
        shutil.copytree(base,self.root/'copied-proof')
        for field,value in (('collectionRoot','$REPO/copied-proof'),('collectionRoot','$REPO/../glm'),
                            ('commandReceiptSha256','0'*64),('reportedFileSha256','0'*64),
                            ('collectionOwnerSHA256','0'*64),
                            ('sourceCommit','f'*40)):
            doc=json.loads(original);doc[field]=value;intent.write_text(json.dumps(doc))
            self.assertEqual((row()['state'],row()['conservativeCalls'],row()['conservativeAttempts']),('UNKNOWN',24,32),field)
            intent.write_bytes(original)
        self.assertEqual((row()['calls'],row()['attempts']),(22,22))


    # Failure inventory at the existing public CLI/process seam: the two
    # release contracts, optional default diagnostic and full release package
    # must execute keylessly; either failure denies a reviewed paid intent.
    def release_contract_preflight_probe(self):
        args=['test','./ai/e2e','-count=1','-run',
              '^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract|TestReleaseRecordedBundleDiagnosticProcess)$']
        package=['test','./ai/internal/testkit/release','-count=1']
        code='#!'+sys.executable+'\nimport json,os,pathlib,shutil,sys\n'
        code+='root=pathlib.Path('+repr(str(self.root))+');args=sys.argv[1:]\n'
        code+='if args=='+repr(args)+':\n'
        code+=' (root/"release-contract-probe.json").write_text(json.dumps(dict(argv=args,environmentKeys=sorted(os.environ))))\n'
        code+=' sys.exit(7 if (root/"release-contract-fail").exists() else 0)\n'
        code+='if args=='+repr(package)+':\n'
        code+=' (root/"release-package-probe.json").write_text(json.dumps(dict(argv=args,environmentKeys=sorted(os.environ))))\n'
        code+=' sys.exit(7 if (root/"release-package-fail").exists() else 0)\n'
        code+='if "-o" in args:\n target=pathlib.Path(args[args.index("-o")+1]);shutil.copyfile(root/"fake-bin/producer",target);target.chmod(0o755)\n'
        (self.bin/'go').write_text(code)
        return args

    def test_release_contract_preflight_is_executed_before_approval(self):
        self.setup_round();expected=self.release_contract_preflight_probe();(self.root/'.env').mkdir()
        self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm')
        probe=self.root/'release-contract-probe.json'
        self.assertTrue(probe.is_file(),'required public release contracts/optional diagnostic never executed')
        request=json.loads(probe.read_text());self.assertEqual(request['argv'],expected)
        self.assertNotIn('OPENAI_API_KEY',request['environmentKeys']);self.assertNotIn('BARNESS_GLM_KEY',request['environmentKeys'])
        checks=json.loads((self.root/'glm/checks.json').read_text())
        self.assertTrue(checks['passed']);self.assertFalse(checks['credentialsRead'])
        self.assertEqual(checks['commands'],dict(vet=0,**{'vet-live':0,'no-key-live':0,
            'no-key-visual-proxy':0,'no-key-release-contracts':0,'no-key-release-package':0}))
        package=json.loads((self.root/'release-package-probe.json').read_text())
        self.assertEqual(package['argv'],['test','./ai/internal/testkit/release','-count=1'])
        self.assertNotIn('OPENAI_API_KEY',package['environmentKeys']);self.assertNotIn('BARNESS_GLM_KEY',package['environmentKeys'])
        record=json.loads((self.root/'glm/no-key-release-contracts-command.json').read_text())
        self.assertFalse(record['credentialsRead']);self.assertIsNone(record['sourceVariable'])
        self.invoke('run-selected','glm','routed-validation-1-glm-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/routed-validation-1-glm-chat.json').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def release_contract_preflight_denies_intent(self,marker,command):
        self.setup_round();self.release_contract_preflight_probe();(self.root/'.env').mkdir()
        (self.root/marker).write_text('controlled public command failure')
        self.invoke('prepare','glm','--phase','glm');self.invoke('check','glm',ok=False)
        base=self.root/'glm';checks=json.loads((base/'checks.json').read_text())
        self.assertFalse(checks['passed']);self.assertFalse(checks['credentialsRead'])
        self.assertEqual(checks['commands'][command],7)
        self.approve()
        auth=base/'execution-authorization.json';value=json.loads(auth.read_text())
        value['checksSHA256']=hashlib.sha256((base/'checks.json').read_bytes()).hexdigest();auth.write_text(json.dumps(value))
        self.invoke('run-selected','glm','routed-validation-1-glm-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/routed-validation-1-glm-chat.json').exists())
        self.assertFalse((base/'glm-chat-command.json').exists());self.assertTrue((self.root/'.env').is_dir())

    def test_release_contract_preflight_failure_blocks_key_and_intent(self):
        self.release_contract_preflight_denies_intent('release-contract-fail','no-key-release-contracts')

    def test_release_contract_preflight_package_failure_blocks_key_and_intent(self):
        self.release_contract_preflight_denies_intent('release-package-fail','no-key-release-package')


if __name__=='__main__':unittest.main(verbosity=2)
