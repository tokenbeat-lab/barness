#!/usr/bin/env python3
"""Controlled executable-boundary checks; credentials are synthetic fixtures."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

TOOLS = Path(__file__).resolve().parent


class Launcher(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='minimax-launcher-')
        self.root = Path(self.tmp.name).resolve()
        self.tools = self.root/'.scratch/barness-ai-minimax/tools'
        shutil.copytree(TOOLS, self.tools, ignore=shutil.ignore_patterns('__pycache__'))
        (self.root/'ai/live').mkdir(parents=True)
        (self.root/'ai/release').mkdir()
        (self.root/'ai/runtime.go').write_text('package ai\n')
        (self.root/'go.mod').write_text('module controlled\n')
        (self.root/'go.sum').write_text('')
        (self.root/'ai/release/catalog-snapshot.json').write_text('{"hash":"sha256:controlled"}')
        self.bin = self.root/'fake-bin'
        self.bin.mkdir()
        (self.bin/'go').write_text('#!/bin/sh\nif [ "$1" = test ] || [ "$1" = build ]; then\n'
            '  while [ "$#" -gt 0 ]; do\n    if [ "$1" = -o ]; then shift; cp "'+str(self.bin/'producer')+'" "$1"; chmod +x "$1"; exit 0; fi\n    shift\n  done\nfi\nexit 0\n')
        (self.bin/'go').chmod(0o755)
        (self.bin/'producer').write_text('#!'+sys.executable+'\nimport json,os\nprint(json.dumps(dict(os.environ)))\n')
        self.env = {'PATH':str(self.bin)+os.pathsep+os.environ['PATH'], 'HOME':str(self.root/'home'),
                    'OPENAI_API_KEY':'AMBIENT-SECRET', 'HTTP_PROXY':'http://forbidden.invalid'}
        for args in (['init','-q'], ['add','.'], ['-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','fixture']):
            subprocess.run(['git',*args],cwd=self.root,env=self.env,check=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        self.cli = [sys.executable,str(self.tools/'paid.py')]

    def tearDown(self):
        self.tmp.cleanup()

    def invoke(self,*args,ok=True):
        p = subprocess.run([*self.cli,*args],cwd=self.root,env=self.env,text=True,capture_output=True,timeout=10)
        if ok: self.assertEqual(p.returncode,0,p.stderr[-500:]+p.stdout[:200])
        else: self.assertNotEqual(p.returncode,0,p.stderr[-500:]+p.stdout[:200])
        return p

    def test_prepare_check_preview_never_acquire_credentials(self):
        # A directory at the credential path makes every attempted open fail.
        (self.root/'.env').mkdir()
        self.invoke('prepare','out')
        self.invoke('check','out')
        view=json.loads(self.invoke('preview','out','mm-research-m3-chat').stdout)
        self.assertEqual(view['maxCalls'],20)
        self.assertEqual(view['plannedCalls'],3)
        self.assertFalse(view['credentialsRead'])
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').exists())

    def approve(self,out='out',routes=None):
        base=self.root/out
        review=json.loads((base/'prerelease-source-review.template.json').read_text())
        review.update(decision='APPROVED',reviewedBy='controlled-parent',
                      authorizedRoutes=routes or ['mm-research-m3-chat'],
                      credentialScopes={'MINIMAX_KEY':'cn-standard'})
        (base/'prerelease-source-review.json').write_text(json.dumps(review))

    def producer(self,report=True):
        code='#!'+sys.executable+'\nimport json,os\nfrom pathlib import Path\nprint(json.dumps(dict(os.environ)))\n'
        if report:
            code += "base=Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session';base.mkdir(parents=True)\n"
            code += "(base/'minimax-report.json').write_text(json.dumps({'outcome':'PASS','budget':{'calls':3,'attempts':3},'cumulativeIdentityBudget':{'calls':3,'attempts':3},'session':'session','mode':'live','kind':'minimax-research-only'}))\n"
            code += "import hashlib\nraw=(base/'minimax-report.json').read_bytes()\n(base/'minimax-integrity.json').write_text(json.dumps({'schema':1,'files':{'minimax-report.json':hashlib.sha256(raw).hexdigest()}}))\n"
        (self.bin/'producer').write_text(code)

    def prepared(self,report=True):
        self.producer(report)
        self.invoke('prepare','out');self.invoke('check','out');self.approve()
        (self.root/'.env').write_text('MINIMAX_KEY=SELECTED-CONTROLLED-SECRET\nOPENAI_KEY=UNSELECTED-CONTROLLED-SECRET\n')

    def test_actual_origin_requires_successful_reviewed_child_and_seal(self):
        self.prepared();self.approve(routes=['mm-candidate-m3-chat'])
        # Synthetic producer reports PASS but omits actual identity/seal: a
        # reviewed launch cannot anchor this as a valid actual candidate.
        self.invoke('run-selected','out','mm-candidate-m3-chat',ok=False)
        intent=json.loads((self.root/'.evidence/minimax-launches/mm-candidate-m3-chat.json').read_text())
        self.assertNotIn('origin',intent)

    def test_malformed_report_never_prints_unknown_fields(self):
        self.prepared();self.invoke('run-selected','out','mm-research-m3-chat')
        path=self.root/'out/mm-research-m3-chat/session/minimax-report.json'
        path.write_text(json.dumps({'outcome':'UNEXPECTED_STATUS_SENTINEL',
                                  'budget':{'calls':3,'attempts':3,'private':'REPORT_SECRET'}}))
        result=self.invoke('summarize','out')
        self.assertNotIn('REPORT_SECRET',result.stdout)
        self.assertNotIn('UNEXPECTED_STATUS_SENTINEL',result.stdout)
        rows=json.loads(result.stdout)['routes']
        row=next(r for r in rows if r['route']=='mm-research-m3-chat')
        self.assertEqual((row['state'],row['conservativeCalls']),('UNKNOWN',20))

    def test_selected_process_isolation_redaction_and_directory_restart(self):
        self.prepared()
        self.invoke('run-selected','out','mm-research-m3-chat')
        output=(self.root/'out/mm-research-m3-chat.log').read_text()
        self.assertNotIn('SELECTED-CONTROLLED-SECRET',output)
        self.assertNotIn('UNSELECTED-CONTROLLED-SECRET',output)
        self.assertNotIn('AMBIENT-SECRET',output)
        self.assertNotIn('HTTP_PROXY',output)
        self.assertNotIn('OPENAI_API_KEY',output)
        self.assertIn('BARNESS_MINIMAX_KEY',output)
        self.assertIn('[SELECTED-CREDENTIAL]',output)
        intent=json.loads((self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').read_text())
        self.assertTrue(intent['credentialsRead'])
        self.assertNotIn(str(self.root),output)
        receipt=(self.root/'out/mm-research-m3-chat-command.json').read_text()
        self.assertNotIn(str(self.root),receipt)
        data=json.loads(receipt)
        self.assertEqual([k for k in data['environmentKeys'] if k.endswith('_KEY')],['BARNESS_MINIMAX_KEY'])
        self.invoke('prepare','other');self.invoke('check','other');self.approve('other')
        retained=self.root/'.evidence/minimax-research-captures/mm-research-m3-chat/session'
        original=self.root/'out/mm-research-m3-chat/session'
        self.assertEqual((retained/'minimax-report.json').read_bytes(),(original/'minimax-report.json').read_bytes())
        self.assertEqual((retained/'minimax-integrity.json').read_bytes(),(original/'minimax-integrity.json').read_bytes())
        later=json.loads(self.invoke('preview','other','mm-candidate-m3-chat').stdout)
        self.assertEqual(later['environment']['BARNESS_MINIMAX_PREDECESSOR_ROOT'],'$REPO/.evidence/minimax-research-captures/mm-research-m3-chat')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','other','mm-research-m3-chat',ok=False)
        self.assertFalse((self.root/'other/mm-research-m3-chat-command.json').exists())

    def test_unknown_research_consumes_full_identity_cap(self):
        self.prepared(report=False)
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False)
        data=json.loads(self.invoke('summarize','out').stdout)
        stage=next(r for r in data['minimax'] if r['stage']=='candidate')
        self.assertEqual((stage['conservativeCalls'],stage['conservativeAttempts']),(20,24))
        self.approve(routes=['mm-candidate-m3-chat'])
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','out','mm-candidate-m3-chat',ok=False)
        self.assertFalse((self.root/'out/mm-candidate-m3-chat-command.json').exists())

    def test_verifier_origin_index_is_not_paid_generation_input(self):
        self.prepared()
        index=self.root/'ai/live/testdata/minimax-formal/origins.json'
        index.parent.mkdir(parents=True)
        index.write_text('{"schema":1,"entries":{}}')
        self.invoke('preview','out','mm-candidate-m3-chat')
        index.write_text('{"schema":1,"entries":{"curated-verifier-only":{}}}')
        self.invoke('preview','out','mm-candidate-m3-chat')
        self.assertFalse((self.root/'.evidence/minimax-launches').exists())

    def test_source_binary_review_drift_refuses_before_intent(self):
        self.prepared()
        runtime=self.root/'ai/runtime.go'
        runtime.write_text('package changed\n')
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False)
        runtime.write_text('package ai\n')
        extra=self.root/'ai/new.go';extra.write_text('package ai\n')
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False);extra.unlink()
        binary=self.root/'out/live.test';original=binary.read_bytes();binary.write_bytes(original+b'# altered\n')
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False);binary.write_bytes(original)
        review=self.root/'out/prerelease-source-review.json';data=json.loads(review.read_text())
        data['sourceManifestSHA256']='wrong';review.write_text(json.dumps(data))
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').exists())

    def test_known_research_plus_candidate_cannot_exceed_identity_cap(self):
        self.prepared();self.invoke('run-selected','out','mm-research-m3-chat')
        # A controlled current candidate claims20after earlier3: launcher must
        # fail even if the aggregate six-identity stage still has spare room.
        base=self.root/'out/mm-candidate-m3-chat/session';base.mkdir(parents=True)
        (base/'minimax-report.json').write_text(json.dumps({'outcome':'PASS','budget':{'calls':20,'attempts':20}}))
        (self.root/'out/mm-candidate-m3-chat-command.json').write_text('{"state":"COMPLETED"}')
        (self.root/'.evidence/minimax-launches/mm-candidate-m3-chat.json').write_text('{"state":"COMPLETED"}')
        self.invoke('summarize','out',ok=False)

    def test_timeout_receipt_cannot_borrow_partial_report_counts(self):
        self.prepared();self.invoke('run-selected','out','mm-research-m3-chat')
        receipt=self.root/'out/mm-research-m3-chat-command.json'
        data=json.loads(receipt.read_text());data['state']='TIMEOUT';receipt.write_text(json.dumps(data))
        intent=self.root/'.evidence/minimax-launches/mm-research-m3-chat.json'
        data=json.loads(intent.read_text());data['state']='TIMEOUT';intent.write_text(json.dumps(data))
        result=json.loads(self.invoke('summarize','out').stdout)
        row=next(r for r in result['routes'] if r['route']=='mm-research-m3-chat')
        self.assertEqual((row['state'],row['conservativeCalls'],row['conservativeAttempts']),('UNKNOWN',20,24))

    def test_actual_producer_status_formats_keep_known_component_counters(self):
        (self.root/'.env').mkdir();self.invoke('prepare','out')
        scenarios={'expected':['text','optional'],'scenarios':[{'id':'text','outcome':'PASS'},{'id':'optional','outcome':'UNSUPPORTED'}], 'budget':{'callsUsed':2,'httpAttempts':2,'attemptsReserved':2}}
        cases=[('openai-chat','live/openai-chat/session/live-report.json',scenarios,2),
               ('glm-chat','glm-chat/session/formal-glm-report.json',{'outcome':'FAIL','budget':{'callsUsed':2,'httpAttempts':2,'attemptsReserved':2}},2),
               ('text-jev','text-jev/session/evaluation-report.json',{'status':'COMPLETE','budget':{'calls':24,'attempts':24}},24),
               ('visual-jev','visual-jev/report.json',{'status':'COMPLETE','budget':{'calls':12,'attempts':12}},12),
               ('protocol114','protocol114/session/multimodal-report.json',{'outcome':'PASS','budget':{'actual':{'calls':12,'attempts':12}}},12)]
        for route,path,report,calls in cases:
            target=self.root/'out'/path;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(json.dumps(report))
            (self.root/'out'/(route+'-command.json')).write_text('{"state":"COMPLETED"}')
            marker=self.root/'.evidence/minimax-launches'/(route+'.json');marker.parent.mkdir(parents=True,exist_ok=True);marker.write_text('{"state":"COMPLETED"}')
        result=json.loads(self.invoke('summarize','out').stdout)
        for route,_,_,calls in cases:
            row=next(r for r in result['routes'] if r['route']==route)
            self.assertEqual((row['state'],row['status'],row['calls'],row['attempts']),('REPORTED','FAIL' if route=='glm-chat' else 'PASS',calls,calls))
            if route in ('openai-chat','glm-chat'):
                self.assertEqual(row['currentReportBudget']['attemptsReserved'],2)
        report=self.root/'out/live/openai-chat/session/live-report.json'
        import hashlib
        marker=self.root/'.evidence/minimax-launches/openai-chat.json'
        marker.write_text(json.dumps({'state':'COMPLETED','reportedFileSha256':hashlib.sha256(report.read_bytes()).hexdigest()}))
        value=json.loads(report.read_text());value['budget']['callsUsed']=0;value['budget']['httpAttempts']=0;report.write_text(json.dumps(value))
        changed=json.loads(self.invoke('summarize','out').stdout)
        row=next(r for r in changed['routes'] if r['route']=='openai-chat')
        self.assertEqual((row['state'],row['conservativeCalls']),('UNKNOWN',32))

    def test_all_commands_keep_component_caps_and_exact_seventeen_inputs(self):
        (self.root/'.env').mkdir()
        self.invoke('prepare','out')
        plan=json.loads((self.root/'out/plan.json').read_text())
        old=[j for j in plan['routes'].values() if j['provider']!='minimax']
        self.assertEqual((sum(j['maxCalls'] for j in old),sum(j['maxAttempts'] for j in old)),(374,398))
        self.assertEqual(len([j for j in old if j['name'].startswith(('text-','visual-'))]),4)
        mm=json.loads(self.invoke('preview','out','mm-candidate-m3-responses').stdout)
        self.assertEqual(mm['environment']['BARNESS_MINIMAX_LEDGER'],'$REPO/.evidence/minimax-stages/candidate.jsonl')
        self.assertEqual((mm['plannedCalls'],mm['maxRequestedOutput'],mm['maxReservedInput']),(13,40000,120000))
        self.assertIn('-test.run=^TestMiniMaxCandidate$',mm['argv'])
        base=self.root/'out'
        # The final combination is independently specified by R27 and the
        # renewal contract:10generic+Jev1.14+GLM3+M3threeprotocols.
        names=['openai-responses','openai-chat','anthropic-messages','google-gemini',
               'deepseek-responses','deepseek-chat','openai-images','google-interactions-image',
               'openai-decisions','typesafe-classifier']
        for n in names:
            path=base/'live'/n/'session';path.mkdir(parents=True);(path/'live-report.json').write_text('{}')
        for n in ['protocol114','glm-chat','glm-responses','glm-messages',
                  'mm-formal-m3-chat','mm-formal-m3-responses','mm-formal-m3-messages','text-jev','text-luna']:
            path=base/n/'session';path.mkdir(parents=True)
            for leaf in ('manifest.json','live-report.json','minimax-report.json'):(path/leaf).write_text('{}')
        (base/'support').mkdir();(base/'support/live-report.json').write_text('{}')
        view=json.loads(self.invoke('keyless-plan','out').stdout)
        self.assertEqual(len(view['liveInputs']),17)
        self.assertEqual(view['gate'].count('-live'),17)
        self.assertNotIn('-bundle',view['gate']);self.assertNotIn('-race-bundle',view['gate'])
        for flag in ('-evaluation-jev','-evaluation-luna','-visual-evaluation-jev','-visual-evaluation-decisions'):
            self.assertIn(flag,view['gate'])
        self.assertFalse(view['credentialsRead'])

    def test_stale_output_refuses_before_key_or_budget_reservation(self):
        self.prepared()
        stale=self.root/'out/mm-research-m3-chat/session';stale.mkdir(parents=True)
        (stale/'minimax-report.json').write_text('{"outcome":"PASS","budget":{"calls":0,"attempts":0}}')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','out','mm-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').exists())
        self.assertFalse((self.root/'out/mm-research-m3-chat-command.json').exists())


if __name__ == '__main__': unittest.main(verbosity=2)
