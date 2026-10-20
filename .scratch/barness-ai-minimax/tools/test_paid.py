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
# Immutable, redacted historical accounting fixtures; never current proof.
CN_INTENT = b'''{
  "state": "COMPLETED",
  "route": "mm-research-m3-chat",
  "sourceCommit": "9481a6ad8028025a7f35da19a3e488935c6cd897",
  "maxCalls": 20,
  "maxAttempts": 24,
  "automaticReruns": false,
  "collectionRoot": "$REPO/.evidence/minimax-ticket10-prepay-final",
  "sourceVariable": "MINIMAX_KEY",
  "credentialsRead": true,
  "exitCode": 1,
  "launcherAudit": {
    "passed": true,
    "findings": []
  },
  "finishedAt": "2026-10-17T23:54:19.193037+00:00",
  "budget": {
    "state": "REPORTED",
    "calls": 2,
    "attempts": 2,
    "status": "FAIL",
    "currentReportBudget": {
      "calls": 2,
      "reservedAttempts": 2,
      "attempts": 2,
      "requestedOutput": 5120,
      "reservedInput": 4096
    },
    "cumulativeIdentityBudget": {
      "calls": 2,
      "reservedAttempts": 2,
      "attempts": 2,
      "requestedOutput": 5120,
      "reservedInput": 4096
    }
  },
  "reportedFileSha256": "9eda3db2343c9ef75b06f6410ea1081323108e2627d46bc8c0406eecb222b286",
  "researchCapture": {
    "session": "20261017T235416.418778000Z",
    "reportSha256": "9eda3db2343c9ef75b06f6410ea1081323108e2627d46bc8c0406eecb222b286",
    "sealSha256": "4dddbc1f1819fb6dadab5a1b2aaa882fa254a78f2390662bad868139ef060abb",
    "files": 34,
    "bytes": 416742
  }
}
'''
CN_PREFIX = b'''{"session":"20261017T235416.418778000Z","sequence":1,"stage":"candidate","identity":{"route":"chat","model":"MiniMax-M3","purpose":"standard","api":"openai-completions","endpoint":"https://api.minimax.cn/v1"},"case":"research-full-stream","kind":"reserve","output":1024,"input":2048,"at":"2026-10-17T23:54:16.453306Z","previous":"","hash":"e9be5ccb57f85d6489c739574c99aa5dde9af9b7d30630fbf27610d56cf66d22"}
{"session":"20261017T235416.418778000Z","sequence":2,"stage":"candidate","identity":{"route":"chat","model":"MiniMax-M3","purpose":"standard","api":"openai-completions","endpoint":"https://api.minimax.cn/v1"},"case":"research-full-stream","kind":"attempt","output":0,"input":0,"at":"2026-10-17T23:54:16.460159Z","previous":"e9be5ccb57f85d6489c739574c99aa5dde9af9b7d30630fbf27610d56cf66d22","hash":"f44a51b40705e17b9e7e96284476795b937dd4c099a7779b22d9c45714cdf7a3"}
{"session":"20261017T235416.418778000Z","sequence":3,"stage":"candidate","identity":{"route":"chat","model":"MiniMax-M3","purpose":"standard","api":"openai-completions","endpoint":"https://api.minimax.cn/v1"},"case":"research-tool-call","kind":"reserve","output":4096,"input":2048,"at":"2026-10-17T23:54:18.705596Z","previous":"f44a51b40705e17b9e7e96284476795b937dd4c099a7779b22d9c45714cdf7a3","hash":"f9d37c734e0eea1251c19f39171247aa0cf5311771d2be4f61b3ac338e860e60"}
{"session":"20261017T235416.418778000Z","sequence":4,"stage":"candidate","identity":{"route":"chat","model":"MiniMax-M3","purpose":"standard","api":"openai-completions","endpoint":"https://api.minimax.cn/v1"},"case":"research-tool-call","kind":"attempt","output":0,"input":0,"at":"2026-10-17T23:54:18.712984Z","previous":"f9d37c734e0eea1251c19f39171247aa0cf5311771d2be4f61b3ac338e860e60","hash":"4277d99a1d84a227e24570a24afa79a859fb5bb8ed98ba1642c33b81bc93075e"}
'''


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
        self.history()
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

    def history(self):
        base=self.root/'.evidence'
        (base/'minimax-launches').mkdir(parents=True,exist_ok=True)
        (base/'minimax-launches/mm-research-m3-chat.json').write_bytes(CN_INTENT)
        (base/'minimax-stages').mkdir(parents=True,exist_ok=True)
        (base/'minimax-stages/candidate.jsonl').write_bytes(CN_PREFIX)

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
        view=json.loads(self.invoke('preview','out','mm-global-research-m3-chat').stdout)
        self.assertEqual(view['maxCalls'],20)
        self.assertEqual(view['plannedCalls'],3)
        self.assertFalse(view['credentialsRead'])
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json').exists())

    def approve(self,out='out',routes=None):
        base=self.root/out
        review=json.loads((base/'prerelease-source-review.template.json').read_text())
        review.update(decision='APPROVED',reviewedBy='controlled-parent',
                      authorizedRoutes=routes or ['mm-global-research-m3-chat'],
                      credentialScopes={'MINIMAX_KEY':'global-standard'})
        (base/'prerelease-source-review.json').write_text(json.dumps(review))

    def producer(self,report=True):
        code='#!'+sys.executable+'\nimport json,os\nfrom pathlib import Path\nprint(json.dumps(dict(os.environ)))\n'
        if report:
            code += "base=Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session';base.mkdir(parents=True)\n"
            code += "(base/'minimax-report.json').write_text(json.dumps({'outcome':'PASS','budget':{'calls':3,'attempts':3,'requestedOutput':3000,'reservedInput':6000},'cumulativeSlotBudget':{'calls':5,'attempts':5,'requestedOutput':8120,'reservedInput':10096},'session':'session','mode':'live','kind':'minimax-research-only'}))\n"
            code += "import hashlib\nraw=(base/'minimax-report.json').read_bytes()\n(base/'minimax-integrity.json').write_text(json.dumps({'schema':1,'files':{'minimax-report.json':hashlib.sha256(raw).hexdigest()}}))\n"
        (self.bin/'producer').write_text(code)

    def prepared(self,report=True):
        self.producer(report)
        self.invoke('prepare','out');self.invoke('check','out');self.approve()
        (self.root/'.env').write_text('MINIMAX_KEY=SELECTED-CONTROLLED-SECRET\nOPENAI_KEY=UNSELECTED-CONTROLLED-SECRET\n')

    def test_actual_origin_requires_successful_reviewed_child_and_seal(self):
        self.prepared();self.approve(routes=['mm-global-candidate-m3-chat'])
        # Synthetic producer reports PASS but omits actual identity/seal: a
        # reviewed launch cannot anchor this as a valid actual candidate.
        self.invoke('run-selected','out','mm-global-candidate-m3-chat',ok=False)
        intent=json.loads((self.root/'.evidence/minimax-launches/mm-global-candidate-m3-chat.json').read_text())
        self.assertNotIn('origin',intent)

    def test_malformed_report_never_prints_unknown_fields(self):
        self.prepared();self.invoke('run-selected','out','mm-global-research-m3-chat')
        path=self.root/'out/mm-global-research-m3-chat/session/minimax-report.json'
        path.write_text(json.dumps({'outcome':'UNEXPECTED_STATUS_SENTINEL',
                                  'budget':{'calls':3,'attempts':3,'private':'REPORT_SECRET'}}))
        result=self.invoke('summarize','out')
        self.assertNotIn('REPORT_SECRET',result.stdout)
        self.assertNotIn('UNEXPECTED_STATUS_SENTINEL',result.stdout)
        rows=json.loads(result.stdout)['routes']
        row=next(r for r in rows if r['route']=='mm-global-research-m3-chat')
        self.assertEqual((row['state'],row['conservativeCalls']),('UNKNOWN',20))

    def test_selected_process_isolation_redaction_and_directory_restart(self):
        self.prepared()
        self.invoke('run-selected','out','mm-global-research-m3-chat')
        output=(self.root/'out/mm-global-research-m3-chat.log').read_text()
        self.assertNotIn('SELECTED-CONTROLLED-SECRET',output)
        self.assertNotIn('UNSELECTED-CONTROLLED-SECRET',output)
        self.assertNotIn('AMBIENT-SECRET',output)
        self.assertNotIn('HTTP_PROXY',output)
        self.assertNotIn('OPENAI_API_KEY',output)
        self.assertIn('BARNESS_MINIMAX_KEY',output)
        self.assertIn('[SELECTED-CREDENTIAL]',output)
        intent=json.loads((self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json').read_text())
        self.assertTrue(intent['credentialsRead'])
        self.assertNotIn(str(self.root),output)
        receipt=(self.root/'out/mm-global-research-m3-chat-command.json').read_text()
        self.assertNotIn(str(self.root),receipt)
        data=json.loads(receipt)
        self.assertEqual([k for k in data['environmentKeys'] if k.endswith('_KEY')],['BARNESS_MINIMAX_KEY'])
        self.invoke('prepare','other');self.invoke('check','other');self.approve('other')
        retained=self.root/'.evidence/minimax-research-captures/mm-global-research-m3-chat/session'
        original=self.root/'out/mm-global-research-m3-chat/session'
        self.assertEqual((retained/'minimax-report.json').read_bytes(),(original/'minimax-report.json').read_bytes())
        self.assertEqual((retained/'minimax-integrity.json').read_bytes(),(original/'minimax-integrity.json').read_bytes())
        later=json.loads(self.invoke('preview','other','mm-global-candidate-m3-chat').stdout)
        self.assertEqual(later['environment']['BARNESS_MINIMAX_PREDECESSOR_ROOT'],'$REPO/.evidence/minimax-research-captures/mm-global-research-m3-chat')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','other','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'other/mm-global-research-m3-chat-command.json').exists())

    def test_unknown_research_consumes_full_identity_cap(self):
        self.prepared(report=False)
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        data=json.loads(self.invoke('summarize','out').stdout)
        stage=next(r for r in data['minimax'] if r['stage']=='candidate')
        self.assertEqual((stage['conservativeCalls'],stage['conservativeAttempts']),(20,24))
        self.approve(routes=['mm-global-candidate-m3-chat'])
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','out','mm-global-candidate-m3-chat',ok=False)
        self.assertFalse((self.root/'out/mm-global-candidate-m3-chat-command.json').exists())

    def test_verifier_origin_index_is_not_paid_generation_input(self):
        self.prepared()
        index=self.root/'ai/live/testdata/minimax-formal/origins.json'
        index.parent.mkdir(parents=True)
        index.write_text('{"schema":1,"entries":{}}')
        self.invoke('preview','out','mm-global-candidate-m3-chat')
        index.write_text('{"schema":1,"entries":{"curated-verifier-only":{}}}')
        self.invoke('preview','out','mm-global-candidate-m3-chat')
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-global-candidate-m3-chat.json').exists())

    def test_embedded_price_document_drift_refuses_before_intent(self):
        source=self.root/'ai/internal/testkit/minimax/pricing-source.md'
        source.parent.mkdir(parents=True)
        source.write_text('Synthetic fixed USD rate document\n')
        for args in (['add',str(source.relative_to(self.root))],
                     ['-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','-qm','rate source']):
            subprocess.run(['git',*args],cwd=self.root,env=self.env,check=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        self.invoke('prepare','out');self.invoke('check','out');self.approve()
        (self.root/'.env').mkdir()
        source.write_text('Changed after reviewed binary build\n')
        self.invoke('preview','out','mm-global-research-m3-chat',ok=False)
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def test_source_binary_review_drift_refuses_before_intent(self):
        self.prepared()
        runtime=self.root/'ai/runtime.go'
        runtime.write_text('package changed\n')
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        runtime.write_text('package ai\n')
        extra=self.root/'ai/new.go';extra.write_text('package ai\n')
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False);extra.unlink()
        binary=self.root/'out/live.test';original=binary.read_bytes();binary.write_bytes(original+b'# altered\n')
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False);binary.write_bytes(original)
        review=self.root/'out/prerelease-source-review.json';data=json.loads(review.read_text())
        data['sourceManifestSHA256']='wrong';review.write_text(json.dumps(data))
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json').exists())

    def test_known_research_plus_candidate_cannot_exceed_identity_cap(self):
        self.prepared();self.invoke('run-selected','out','mm-global-research-m3-chat')
        # A controlled current candidate claims20after earlier3: launcher must
        # fail even if the aggregate six-identity stage still has spare room.
        base=self.root/'out/mm-global-candidate-m3-chat/session';base.mkdir(parents=True)
        (base/'minimax-report.json').write_text(json.dumps({'outcome':'PASS','budget':{'calls':20,'attempts':20}}))
        (self.root/'out/mm-global-candidate-m3-chat-command.json').write_text('{"state":"COMPLETED"}')
        (self.root/'.evidence/minimax-launches/mm-global-candidate-m3-chat.json').write_text('{"state":"COMPLETED"}')
        self.invoke('summarize','out',ok=False)

    def test_timeout_receipt_cannot_borrow_partial_report_counts(self):
        self.prepared();self.invoke('run-selected','out','mm-global-research-m3-chat')
        receipt=self.root/'out/mm-global-research-m3-chat-command.json'
        data=json.loads(receipt.read_text());data['state']='TIMEOUT';receipt.write_text(json.dumps(data))
        intent=self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json'
        data=json.loads(intent.read_text());data['state']='TIMEOUT';intent.write_text(json.dumps(data))
        result=json.loads(self.invoke('summarize','out').stdout)
        row=next(r for r in result['routes'] if r['route']=='mm-global-research-m3-chat')
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
        mm=json.loads(self.invoke('preview','out','mm-global-candidate-m3-responses').stdout)
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
                  'mm-global-formal-m3-chat','mm-global-formal-m3-responses','mm-global-formal-m3-messages','text-jev','text-luna']:
            path=base/n/'session';path.mkdir(parents=True)
            for leaf in ('manifest.json','live-report.json','minimax-report.json'):(path/leaf).write_text('{}')
        (base/'support').mkdir();(base/'support/live-report.json').write_text('{}')
        view=json.loads(self.invoke('keyless-plan','out').stdout)
        self.assertEqual(len(view['liveInputs']),17)
        self.assertEqual(view['liveIdentities'][-3:],['minimax-chat','minimax-responses','minimax-messages'])
        self.assertEqual(view['liveIdentities'][:-3],names+['typesafe-classifier-jev-1.14.0']+['zhipu-'+r for r in ('chat','responses','messages')])
        self.assertEqual(view['gate'].count('-live'),17)
        self.assertNotIn('-bundle',view['gate']);self.assertNotIn('-race-bundle',view['gate'])
        for flag in ('-evaluation-jev','-evaluation-luna','-visual-evaluation-jev','-visual-evaluation-decisions'):
            self.assertIn(flag,view['gate'])
        self.assertFalse(view['credentialsRead'])
        # Execute the fixed verification sequence twice. The controlled Go
        # boundary enforces jevformal's real exclusive-output contract, so a
        # stale creation action cannot hide behind argv-count assertions.
        log=self.root/'verified-argv.jsonl'
        (self.bin/'go').write_text('#!'+sys.executable+'\nimport json,sys\nfrom pathlib import Path\n'
            'args=sys.argv[1:]\n'
            'with Path('+repr(str(log))+').open("a") as stream:stream.write(json.dumps(args)+"\\n")\n'
            'if "./ai/live/cmd/jevformal" in args and "-out" in args:\n'
            '    target=Path(args[args.index("-out")+1]);target.mkdir()\n')
        for _ in range(2):
            emitted=json.loads(self.invoke('keyless-plan','out').stdout)
            for argv in emitted['verify']:
                expanded=[arg.replace('$REPO',str(self.root)) for arg in argv]
                completed=subprocess.run(expanded,cwd=self.root,env=self.env,capture_output=True,text=True,timeout=10)
                self.assertEqual(completed.returncode,0,completed.stderr)
        actions=[json.loads(line) for line in log.read_text().splitlines()]
        jev=[a for a in actions if './ai/live/cmd/jevformal' in a]
        self.assertEqual(len(jev),2)
        self.assertTrue(all('-verify' in a and '-out' not in a for a in jev))
        self.assertEqual((base/'support/live-report.json').read_text(),'{}')

    def test_global_inventory_scopes_and_historical_balance_are_explicit(self):
        (self.root/'.env').mkdir();self.invoke('prepare','out');self.invoke('check','out')
        old=(self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').read_bytes()
        plan=json.loads((self.root/'out/plan.json').read_text())
        self.assertTrue(plan['previousTaskBalanceCarried'])
        self.assertEqual(plan['historicalAccounting']['knownCalls'],2)
        for route in ('chat','responses','messages'):
            job=json.loads(self.invoke('preview','out','mm-global-candidate-m3-'+route).stdout)
            self.assertEqual(job['environment']['BARNESS_MINIMAX_ACCOUNT_ALIAS'],'minimax-global-standard')
            self.assertEqual(job['environment']['BARNESS_MINIMAX_REGION_ALIAS'],'global')
            self.assertIn('mm-global-research-m3-',job['environment']['BARNESS_MINIMAX_PREDECESSOR_ROOT'])
        mplan=json.loads(self.invoke('preview','out','mm-global-research-m31-chat').stdout)
        self.assertEqual(mplan['environment']['BARNESS_MINIMAX_ACCOUNT_ALIAS'],'minimax-global-mplan')
        self.invoke('preview','out','mm-research-m3-chat',ok=False)
        self.approve();review=self.root/'out/prerelease-source-review.json'
        value=json.loads(review.read_text());value['credentialScopes']['MINIMAX_KEY']='cn-standard';review.write_text(json.dumps(value))
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'out/mm-global-research-m3-chat-command.json').exists())
        self.assertEqual((self.root/'.evidence/minimax-launches/mm-research-m3-chat.json').read_bytes(),old)
        self.approve(routes=['mm-global-research-m31-chat'])
        value=json.loads(review.read_text());value['credentialScopes']['MINIMAX_MPLAN_KEY']='cn-mplan';review.write_text(json.dumps(value))
        self.invoke('run-selected','out','mm-global-research-m31-chat',ok=False)
        self.assertFalse((self.root/'out/mm-global-research-m31-chat-command.json').exists())

    def test_history_is_counted_once_with_current_component_not_cumulative(self):
        self.prepared();self.invoke('run-selected','out','mm-global-research-m3-chat')
        result=json.loads(self.invoke('summarize','out').stdout)
        current=next(x for x in result['routes'] if x['route']=='mm-global-research-m3-chat')
        self.assertEqual(current['cumulativeSlotBudget']['calls'],5)
        self.assertEqual(current['budgetSlot'],'chat/MiniMax-M3/standard')
        self.assertEqual(current['identity']['endpoint'],'https://api.minimax.io/v1')
        self.assertNotIn('identity',result['historicalAccounting']['routes'][0])
        self.assertNotIn('cumulativeIdentityBudget',current)
        stage=next(x for x in result['minimax'] if x['stage']=='candidate')
        slot=stage['budgetSlots']['chat/MiniMax-M3/standard']
        self.assertEqual((slot['calls'],slot['attempts']),(5,5))
        self.assertEqual((slot['requestedOutput'],slot['reservedInput']),(8120,10096))
        data=json.loads(self.invoke('summarize','out').stdout)
        self.assertEqual(data['minimax'][0]['budgetSlots']['chat/MiniMax-M3/standard']['calls'],5)
        self.assertEqual(len(result['historicalAccounting']['routes']),1)
        self.assertEqual((result['historicalAccounting']['knownCalls'],result['historicalAccounting']['knownAttempts']),(2,2))
        self.assertEqual((self.root/'.evidence/minimax-stages/candidate.jsonl').read_bytes(),CN_PREFIX)

    def test_history_mutation_missing_or_partial_is_unknown_full_slot(self):
        (self.root/'.env').mkdir();self.invoke('prepare','out');self.invoke('check','out');self.approve()
        intent=self.root/'.evidence/minimax-launches/mm-research-m3-chat.json'
        ledger=self.root/'.evidence/minimax-stages/candidate.jsonl'
        for mutation in ('intent','ledger','missing-intent','missing-both'):
            self.history()
            if mutation=='intent':
                value=json.loads(intent.read_bytes());value['budget']['calls']=0;intent.write_text(json.dumps(value))
            elif mutation=='ledger':ledger.write_bytes(CN_PREFIX[:-1])
            elif mutation=='missing-intent':intent.unlink()
            else:intent.unlink();ledger.unlink()
            result=json.loads(self.invoke('summarize','out').stdout)
            stage=next(x for x in result['minimax'] if x['stage']=='candidate')
            self.assertEqual((stage['conservativeCalls'],stage['conservativeAttempts']),(20,24))
            self.assertEqual(result['historicalAccounting']['routes'][0]['state'],'UNKNOWN')
            self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
            self.assertFalse((self.root/'out/mm-global-research-m3-chat-command.json').exists())

    def test_history_and_global_requested_token_caps_share_one_slot(self):
        (self.root/'.env').mkdir();self.invoke('prepare','out')
        base=self.root/'out/mm-global-candidate-m3-chat/session';base.mkdir(parents=True)
        report=base/'minimax-report.json'
        (self.root/'out/mm-global-candidate-m3-chat-command.json').write_text('{"state":"COMPLETED"}')
        marker=self.root/'.evidence/minimax-launches/mm-global-candidate-m3-chat.json';marker.write_text('{"state":"COMPLETED"}')
        for budget in ({'calls':19,'attempts':19,'requestedOutput':1,'reservedInput':1},
                       {'calls':1,'attempts':1,'requestedOutput':35000,'reservedInput':1},
                       {'calls':1,'attempts':1,'requestedOutput':1,'reservedInput':116000}):
            report.write_text(json.dumps({'outcome':'PASS','budget':budget}))
            self.invoke('summarize','out',ok=False)

    def test_stale_output_refuses_before_key_or_budget_reservation(self):
        self.prepared()
        stale=self.root/'out/mm-global-research-m3-chat/session';stale.mkdir(parents=True)
        (stale/'minimax-report.json').write_text('{"outcome":"PASS","budget":{"calls":0,"attempts":0}}')
        (self.root/'.env').unlink();(self.root/'.env').mkdir()
        self.invoke('run-selected','out','mm-global-research-m3-chat',ok=False)
        self.assertFalse((self.root/'.evidence/minimax-launches/mm-global-research-m3-chat.json').exists())
        self.assertFalse((self.root/'out/mm-global-research-m3-chat-command.json').exists())


if __name__ == '__main__': unittest.main(verbosity=2)
