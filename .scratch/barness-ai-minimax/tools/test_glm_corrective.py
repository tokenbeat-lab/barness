#!/usr/bin/env python3
"""Executable corrective-1 boundary; all accounts/captures are synthetic."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import unittest
from test_paid import Launcher


class Corrective(unittest.TestCase):
    setUp = Launcher.setUp
    tearDown = Launcher.tearDown
    invoke = Launcher.invoke
    history = Launcher.history

    def originals(self):
        for route in ('chat','responses','messages'):
            value=dict(state='COMPLETED',route='glm-'+route,sourceCommit='8'*40,
                sourceVariable='GLM_KEY',maxCalls=24,maxAttempts=32,exitCode=1 if route=='responses' else 0,
                credentialsRead=True,launcherAudit=dict(passed=True),reportedFileSha256='a'*64,
                budget=dict(state='REPORTED',calls=22,attempts=22,status='FAIL' if route=='responses' else 'PASS'))
            (self.root/'.evidence/minimax-launches'/('glm-'+route+'.json')).write_text(json.dumps(value))

    def setup_round(self):
        self.originals()
        (self.bin/'producer').write_text('#!'+sys.executable+'\nimport json,os,pathlib,subprocess,time\n'
            'time.sleep(0.1)\nr=os.environ["BARNESS_GLM_ROUTE"];p=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"session";p.mkdir(parents=True)\n'
            'd=dict(schema=1,kind="glm-formal",mode="live",route=r,outcome="PASS",complete=True,auditPassed=True,gitCommit=subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip(),budget=dict(callsUsed=22,httpAttempts=22,attemptsReserved=22))\n'
            '(p/"formal-glm-report.json").write_text(json.dumps(d))\n'
            'for n in ("glm-integrity.json","live-report.json","glm-execution-source.json"):(p/n).write_text("{}")\n'
            'print(json.dumps(dict(os.environ)))\n')
        self.cli=[sys.executable,str(self.tools/'glm_corrective.py')]
        self.invoke('prepare','corrective')
        self.invoke('check','corrective')

    def approve(self,out='corrective',human=True):
        base=self.root/out
        value=json.loads((base/'prerelease-source-review.template.json').read_text())
        value.update(decision='APPROVED',reviewedBy='controlled-parent',
            authorizedRoutes=['glm-corrective-1-'+r for r in ('chat','responses','messages')],
            credentialScopes={'GLM_KEY':'glm-corrective-1'})
        (base/'prerelease-source-review.json').write_text(json.dumps(value))
        if human:
            value=json.loads((base/'human-approval.template.json').read_text())
            value.update(decision='APPROVED',authority='explicit-user-reply',replyReference='controlled explicit fixture')
            (base/'human-approval.json').write_text(json.dumps(value))

    def test_pending_plan_has_exact_new_round_and_carried_originals(self):
        self.setup_round();(self.root/'.env').mkdir()
        data=json.loads((self.root/'corrective/plan.json').read_text())
        self.assertEqual(data['approval'],'PENDING HUMAN APPROVAL')
        self.assertEqual(set(data['routes']),{'glm-corrective-1-'+r for r in ('chat','responses','messages')})
        self.assertEqual(data['budget'],dict(priorCalls=66,priorAttempts=66,additionalMaxCalls=72,additionalMaxAttempts=96,lifetimeGLMMaxCalls=138,lifetimeGLMMaxAttempts=162,otherLegacyMaxCalls=302,otherLegacyMaxAttempts=302,legacyMaxCalls=440,legacyMaxAttempts=464))
        for job in data['routes'].values():
            self.assertEqual((job['maxCalls'],job['maxAttempts'],job['plannedCalls']),(24,32,22))
            self.assertEqual(job['argv'][1],'-test.run=^TestGLMFormal$')
            self.assertEqual(job['sourceVariable'],'GLM_KEY')
        self.approve(human=False)
        self.invoke('run-selected','corrective','glm-corrective-1-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/glm-corrective-1-chat.json').exists())

    def test_explicit_approval_is_exact_and_original_spend_cannot_disappear(self):
        self.setup_round();self.approve();(self.root/'.env').mkdir()
        file=self.root/'corrective/human-approval.json';raw=file.read_bytes();value=json.loads(raw)
        for field,bad in (('planSHA256','0'*64),('additionalMaxCalls',73),('authority','reviewer'),('replyReference','')):
            altered=value|{field:bad};file.write_text(json.dumps(altered))
            self.invoke('run-selected','corrective','glm-corrective-1-chat',ok=False)
            self.assertFalse((self.root/'.evidence/minimax-launches/glm-corrective-1-chat.json').exists())
        file.write_bytes(raw)
        receipt=self.root/'.evidence/minimax-launches/glm-responses.json';old=receipt.read_bytes();receipt.write_text('{}')
        self.invoke('run-selected','corrective','glm-corrective-1-chat',ok=False);receipt.write_bytes(old)
        self.invoke('run-selected','corrective','glm-corrective-2-chat',ok=False)

    def test_fixed_selected_key_origin_and_once_only_across_outputs(self):
        self.setup_round();self.approve()
        (self.root/'.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\nMINIMAX_KEY=SYNTHETIC-UNSELECTED\n')
        original={p.name:p.read_bytes() for p in (self.root/'.evidence/minimax-launches').glob('glm-*.json')}
        self.invoke('run-selected','corrective','glm-corrective-1-chat')
        log=(self.root/'corrective/glm-corrective-1-chat.log').read_text()
        self.assertIn('BARNESS_GLM_KEY',log);self.assertNotIn('SYNTHETIC-SELECTED',log);self.assertNotIn('MINIMAX',log);self.assertNotIn('HTTP_PROXY',log)
        record=json.loads((self.root/'.evidence/minimax-launches/glm-corrective-1-chat.json').read_text())
        self.assertEqual(record['correctiveRound'],'corrective-1');self.assertEqual(record['budget']['calls'],22)
        self.assertIn('capturePins',record);self.assertIn('humanApprovalSha256',record)
        for name,raw in original.items():self.assertEqual((self.root/'.evidence/minimax-launches'/name).read_bytes(),raw)
        summary=json.loads(self.invoke('summarize','corrective').stdout)
        self.assertEqual((summary['lifetime']['calls'],summary['lifetime']['attempts']),(88,88))
        aggregate=subprocess.run([sys.executable,str(self.tools/'paid.py'),'summarize','corrective'],cwd=self.root,env=self.env,text=True,capture_output=True,timeout=10)
        self.assertEqual(aggregate.returncode,0,aggregate.stderr)
        total=json.loads(aggregate.stdout)['legacy']
        self.assertEqual((total['calls'],total['attempts'],total['maxCalls'],total['maxAttempts']),(88,88,440,464))
        self.invoke('prepare','other');self.invoke('check','other');self.approve('other')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','other','glm-corrective-1-chat',ok=False)
        self.assertFalse((self.root/'other/glm-corrective-1-chat-command.json').exists())

    def test_concurrent_launch_is_exclusive_and_missing_report_spends_full_cap(self):
        self.setup_round();self.approve()
        (self.root/'.env').write_text('GLM_KEY=SYNTHETIC-SELECTED\n')
        args=[*self.cli,'run-selected','corrective','glm-corrective-1-chat']
        a=subprocess.Popen(args,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        b=subprocess.Popen(args,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        a.communicate(timeout=10);b.communicate(timeout=10)
        self.assertEqual(sorted([a.returncode,b.returncode]),[0,1])
        report=self.root/'corrective/glm-corrective-1-chat/session/formal-glm-report.json';report.unlink()
        marker=self.root/'.evidence/minimax-launches/glm-corrective-1-chat.json'
        value=json.loads(marker.read_text());value['budget']={'state':'UNKNOWN'};marker.write_text(json.dumps(value))
        summary=json.loads(self.invoke('summarize','corrective').stdout)
        self.assertEqual((summary['lifetime']['conservativeCalls'],summary['lifetime']['conservativeAttempts']),(90,98))
        self.invoke('run-selected','corrective','glm-corrective-1-chat',ok=False)

    def test_unknown_original_fails_preparation_without_credential_or_intent(self):
        self.originals();self.cli=[sys.executable,str(self.tools/'glm_corrective.py')]
        p=self.root/'.evidence/minimax-launches/glm-messages.json';value=json.loads(p.read_text());value['budget']['state']='UNKNOWN';p.write_text(json.dumps(value))
        (self.root/'.env').mkdir();self.invoke('prepare','out',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/glm-corrective-1-messages.json').exists())


if __name__=='__main__':unittest.main(verbosity=2)
