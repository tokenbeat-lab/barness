#!/usr/bin/env python3
"""Socket-free public-process checks; all approvals/keys are synthetic."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]

class FinalRound(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='provider-sites-operator-')
        self.root = Path(self.tmp.name)
        self.tools = self.root / '.scratch/barness-provider-sites/tools'
        shutil.copytree(HERE, self.tools, ignore=shutil.ignore_patterns('__pycache__'))
        shutil.copytree(REPO / '.scratch/barness-ai-minimax/tools', self.root / '.scratch/barness-ai-minimax/tools', ignore=shutil.ignore_patterns('__pycache__'))
        (self.root / 'ai').mkdir()
        (self.root / 'ai/live').mkdir()
        (self.root / 'ai/sentinel.go').write_text('package controlled\n')
        (self.root / 'ai/release').mkdir()
        (self.root / 'ai/release/catalog-snapshot.json').write_text(json.dumps(dict(hash='sha256:controlled', catalog=dict(version='controlled'))))
        (self.root / 'go.mod').write_text('module controlled\n')
        (self.root / 'go.sum').write_text('')
        self.env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR') if k in os.environ}
        self.env.update(HTTP_PROXY='http://forbidden.invalid', OPENAI_KEY='FORBIDDEN-AMBIENT')
        self.git('init', '-q')
        self.seed_research()
        self.commit()
        # A directory cannot be read as a credential file. Keyless operations
        # must still work; refusal before approval must not touch it.
        (self.root / '.env').mkdir()
        self.cli = [sys.executable, '-B', str(self.tools / 'final_round.py')]
        self.journal = []

    def tearDown(self):
        target = os.environ.get('BARNESS_OPERATOR_TEST_EVIDENCE')
        if target:
            p = Path(target) / self._testMethodName
            p.mkdir(parents=True, exist_ok=True)
            (p / 'processes.json').write_text(json.dumps(self.journal, indent=2) + '\n')
        self.tmp.cleanup()

    def seed_research(self):
        import hashlib
        base=self.root/'.evidence/provider-sites/final-acceptance-20261020';base.mkdir(parents=True)
        def put(path,value):
            path.parent.mkdir(parents=True,exist_ok=True);path.write_text(json.dumps(value))
            return hashlib.sha256(path.read_bytes()).hexdigest()
        plan='a'*64
        approval=put(base/'authorization.json',dict(decision='APPROVED',roundId='provider-sites-final-acceptance-1',researchPlanSha256=plan,requestText='Synthetic research-only fixture'))
        command=put(base/'research-command.receipt.json',dict(exitCode=0,authorizationSha256=approval))
        files={'authorization.json':approval,'research-command.receipt.json':command}
        for i in range(14):files['research/'+str(i)+'.json']=put(base/'research'/(str(i)+'.json'),dict(synthetic=True))
        put(base/'research-index.json',dict(classification='RAW_OWN_WIRE_RESEARCH_NOT_QUALIFICATION',calls=4,httpAttempts=4,retries=0,qualificationPassClaimed=False,keyAudit=dict(findings=0),budgetUSDReserved=.2,files=files))
        put(self.root/'.git/barness-regional-research'/(plan+'.json'),dict(planSha256=plan,authorizationSha256=approval))

    def git(self, *argv):
        return subprocess.check_output(['git', *argv], cwd=self.root, env=self.env, stderr=subprocess.DEVNULL).decode().strip()

    def commit(self):
        self.git('add', '.')
        self.git('-c', 'user.name=Controlled', '-c', 'user.email=controlled@example.invalid', 'commit', '-qm', 'controlled source')

    def runcli(self, *argv, ok=True):
        result = subprocess.run([*self.cli, *argv], cwd=self.root, env=self.env, text=True, capture_output=True, timeout=30)
        self.journal.append(dict(argv=list(argv), exitCode=result.returncode, stdout=result.stdout.replace(str(self.root), '$CONTROLLED_REPO'), stderr=result.stderr.replace(str(self.root), '$CONTROLLED_REPO')))
        self.assertEqual(result.returncode == 0, ok, result.stderr[-1000:] + result.stdout[:300])
        return result

    def fake_build(self, regional_reports=False):
        bindir = self.root / 'bin'
        bindir.mkdir()
        go = bindir / 'go'
        native_source = '\nif "-source" in sys.argv[1:]:\n import hashlib\n files={p.as_posix():hashlib.sha256(p.read_bytes()).hexdigest() for p in pathlib.Path("ai/internal/testkit/regional").rglob("*") if p.is_file()};print(json.dumps(dict(sourceSha256="a"*64,files=files)))\n'
        producer = '#!' + sys.executable + '\nimport sys,json,os,pathlib\nif os.getenv("BARNESS_REGIONAL_KEY"): print(json.dumps(dict(environmentKeys=sorted(os.environ))));pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"]).parent.joinpath("selected-child-started").touch()\na=sys.argv[1:]\nif "-grant-template" in a:\n c=a[a.index("-grant-template")+1];s=a[a.index("-stage")+1];print(json.dumps(dict(schema=1,mode="live",id="<new>",accountAlias="<own>",regionAlias="<own>",expiresAt="0001-01-01T00:00:00Z",combo=c,stage=s)))\n'
        if regional_reports:
            # This controlled executable tests operator filename/phase wiring;
            # the real native validator has independent capture/authority E2E.
            producer += r'''
if os.getenv("BARNESS_REGIONAL_KEY"):
 import subprocess
 base=pathlib.Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])/"controlled-session";base.mkdir(parents=True)
 combo=os.environ["BARNESS_REGIONAL_ROUTE"];count=18 if combo.startswith("zai-") else 9
 report=dict(combo=combo,gitCommit=subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip(),budget=dict(callsUsed=count,httpAttempts=count),expected=[str(i) for i in range(count)],scenarios=[dict(id=str(i),outcome="PASS") for i in range(count)])
 (base/"regional-report.json").write_text(json.dumps(report))
 usd=pathlib.Path(os.environ["BARNESS_FINAL_BUDGET"]);value=json.loads(usd.read_text());value.update(reservedMicrousd=1000,attempts=count);usd.write_text(json.dumps(value))
if "-qualify" in a:
 base=pathlib.Path(a[-1]);assert (base/"regional-report.json").is_file() and not (base/"live-report.json").exists()
 print((base/"regional-report.json").read_text())
if "-receipt" in a:
 assert len(a)==8 and all((pathlib.Path(p)/"regional-report.json").is_file() for p in a[2:])
 pathlib.Path(a[1]).write_text(json.dumps(dict(kind="controlled-operator-fixture")))
 pathlib.Path("ai/catalog_regional.json").write_text(json.dumps(dict(active=True)))
'''
            (self.root/'ai/catalog_regional.json').write_text(json.dumps(dict(active=False)))
        producer += native_source
        go.write_text('#!' + sys.executable + '\nimport sys,json,pathlib\na=sys.argv[1:]\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);p.write_text(' + repr(producer) + ');p.chmod(0o755)\n' + native_source)
        go.chmod(0o755)
        self.env['PATH'] = str(bindir) + os.pathsep + self.env['PATH']
        self.commit()

    def test_prepare_freezes_current_source_and_refuses_paid_start_without_approval(self):
        self.fake_build()
        self.runcli('prepare', '.evidence/provider-sites/current', '--phase', 'glm')
        base = self.root / '.evidence/provider-sites/current'
        plan = json.loads((base / 'plan.json').read_text())
        self.assertEqual(set(plan['routes']), {'glm-chat', 'glm-responses', 'glm-messages'})
        template = json.loads((base / 'execution-authorization.template.json').read_text())
        self.assertEqual(template['decision'], 'PENDING')
        self.assertEqual((template['maxCalls'], template['maxAttempts']), (726, 762))
        self.runcli('preview', '.evidence/provider-sites/current', 'glm-chat')
        self.runcli('run-selected', '.evidence/provider-sites/current', 'glm-chat', ok=False)
        self.assertTrue((self.root / '.env').is_dir())
        self.assertFalse((self.root / '.evidence/minimax-launches/provider-sites-final-acceptance-1-glm-chat.json').exists())
        (self.root / 'ai/sentinel.go').write_text('package changed\n')
        self.runcli('preview', '.evidence/provider-sites/current', 'glm-chat', ok=False)

    def test_native_non_json_source_add_modify_delete_refuses_before_secret(self):
        self.fake_build()
        source=self.root/'ai/internal/testkit/regional/capture.gz'
        source.parent.mkdir(parents=True);source.write_bytes(b'controlled embedded original')
        self.commit()
        self.runcli('prepare','.evidence/provider-sites/current','--phase','candidate')
        original=source.read_bytes()
        source.write_bytes(b'controlled changed embedded original')
        self.runcli('preview','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        source.write_bytes(original)
        extra=source.with_name('extra.gz');extra.write_bytes(b'controlled added original')
        self.runcli('preview','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        extra.unlink();source.unlink()
        self.runcli('preview','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        source.write_bytes(original)
        self.runcli('preview','.evidence/provider-sites/current','candidate-zai-chat')
        self.assertTrue((self.root/'.env').is_dir())

    def test_native_grant_and_promotion_dispatch_without_final_matrix_inputs(self):
        self.fake_build()
        self.runcli('prepare','.evidence/provider-sites/candidate','--phase','candidate')
        result=self.runcli('native-grant','.evidence/provider-sites/candidate',ok=False)
        self.assertIn('requires the final remaining-source freeze',result.stderr)
        result=self.runcli('promote','.evidence/provider-sites/candidate',ok=False)
        self.assertIn('minimax-launches',result.stderr)
        self.assertNotIn('all23+four',result.stderr)
        self.runcli('prepare','.evidence/provider-sites/remaining','--phase','remaining')
        result=self.runcli('promote','.evidence/provider-sites/remaining',ok=False)
        self.assertIn('requires this frozen six-candidate phase',result.stderr)
        result=self.runcli('native-grant','.evidence/provider-sites/remaining',ok=False)
        self.assertIn('prerelease-source-review.json',result.stderr)
        self.assertFalse((self.root/'.evidence/provider-sites/candidate/finalization').exists())
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-1/grant.json').exists())
        self.assertTrue((self.root/'.env').is_dir())

    def approve(self, base, reserve=1000000):
        import hashlib
        plan = json.loads((base / 'plan.json').read_text())
        review = json.loads((base / 'prerelease-source-review.template.json').read_text())
        for name in ('standardsReview', 'specReview'):
            proof = base / (name + '.md')
            proof.write_text('Synthetic independent ordinary process fixture review.\n')
            review[name] = dict(path=str(proof.relative_to(self.root)), sha256=hashlib.sha256(proof.read_bytes()).hexdigest())
        review.update(decision='APPROVED', reviewedBy='synthetic independent reviewers', authorizedRoutes=list(plan['routes']))
        (base / 'prerelease-source-review.json').write_text(json.dumps(review))
        auth = json.loads((base / 'execution-authorization.template.json').read_text())
        auth.update(decision='APPROVED', requestText='Synthetic approve new fixed726/762 US30 standard only', requestReference='controlled test reply', maxUSD='30', diagnosticReservedMicrousd=200000)
        for name, job in plan['routes'].items():
            auth['standardCredentialScopes'][job['sourceVariable']] = dict(provider=job['provider'], purpose='standard', confirmed=True, accountAlias='controlled-own-standard', regionAlias='controlled-own-region')
            auth['networkRoutes'][name] = 'direct'
            auth['processReservationsMicrousd'][name] = reserve
            if 'combo' in job:
                import datetime
                value=json.loads((base/'grants'/(name+'.template.json')).read_text())
                value.update(id='ps-final-1-'+name,accountAlias='controlled-own-standard',regionAlias='controlled-own-region',expiresAt=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(hours=1)).isoformat())
                grant=base/'grants'/(name+'.json');grant.write_text(json.dumps(value))
                auth['componentAuthorizationSHA256'][name]=hashlib.sha256(grant.read_bytes()).hexdigest()
        (base / 'execution-authorization.json').write_text(json.dumps(auth))

    def test_one_shot_claim_precedes_secret_and_failed_process_keeps_full_reservation(self):
        self.fake_build()
        self.runcli('prepare', '.evidence/provider-sites/current', '--phase', 'candidate')
        base = self.root / '.evidence/provider-sites/current'
        self.runcli('check', '.evidence/provider-sites/current')
        self.approve(base)
        # No actual credential can be read; this deliberate acquisition failure
        # still burns the durable claim and its whole conservative allowance.
        self.runcli('run-selected', '.evidence/provider-sites/current', 'candidate-zai-chat', ok=False)
        marker = self.root / '.evidence/minimax-launches/provider-sites-final-acceptance-1-candidate-zai-chat.json'
        self.assertTrue(marker.exists())
        receipt = json.loads(marker.read_text())
        self.assertEqual(receipt['reservedMicrousd'], 1000000)
        self.assertEqual(receipt['budget']['state'], 'UNKNOWN')
        self.runcli('run-selected', '.evidence/provider-sites/current', 'candidate-zai-chat', ok=False)
        summary = json.loads(self.runcli('summarize').stdout)
        self.assertEqual(summary['reservedMicrousd'], 1200000)
        self.assertEqual(summary['claimedCalls'], 28)
        self.assertTrue((self.root / '.env').is_dir())

    def test_candidate_native_report_settlement_and_six_route_promotion(self):
        self.fake_build(regional_reports=True)
        self.runcli('prepare','.evidence/provider-sites/current','--phase','candidate')
        base=self.root/'.evidence/provider-sites/current'
        self.runcli('check','.evidence/provider-sites/current');self.approve(base)
        (self.root/'.env').rmdir();(self.root/'.env').write_text('ZAI_KEY=SYNTHETIC-ZAI\nMINIMAX_CN_KEY=SYNTHETIC-CN\n')
        for combo in ('zai-chat','zai-responses','zai-messages','minimax-cn-chat','minimax-cn-responses','minimax-cn-messages'):
            name='candidate-'+combo
            result=json.loads(self.runcli('run-selected','.evidence/provider-sites/current',name).stdout)
            count=18 if combo.startswith('zai-') else 9
            self.assertEqual(result['budget'],dict(state='REPORTED',status='PASS',calls=count,attempts=count))
            self.assertFalse(list((base/name).rglob('live-report.json')))
        result=json.loads(self.runcli('promote','.evidence/provider-sites/current').stdout)
        self.assertEqual(result['activatedFile'],'ai/catalog_regional.json')
        self.assertEqual(json.loads((self.root/'ai/catalog_regional.json').read_text()),dict(active=True))
        self.assertTrue((self.root/'.evidence/provider-sites/final-acceptance-20261020/promotion/regional-promotion-lineage.json').is_file())

    def test_final_plan_requires_all23_own_inputs_and_four_effects(self):
        self.fake_build()
        self.runcli('prepare', '.evidence/provider-sites/remaining', '--phase', 'remaining')
        base=self.root/'.evidence/provider-sites/remaining'
        inventory=json.loads(self.runcli('inventory').stdout)
        names=[name for name in inventory['jobs'] if name.startswith('formal-')]
        # Literal old17 identities plus exactly six newly formal providers.
        old=['openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','typesafe-classifier-jev-1.14.0','zhipu-chat','zhipu-responses','zhipu-messages','minimax-chat','minimax-responses','minimax-messages']
        live={name:str((base/'inputs'/name).relative_to(self.root)) for name in old+[n.removeprefix('formal-') for n in names]}
        effects={name:str((base/'inputs'/name).relative_to(self.root)) for name in ('text-jev','text-luna','visual-jev','visual-decisions')}
        for name,relative in live.items():
            p=self.root/relative;p.mkdir(parents=True);(p/'live-report.json').write_text(json.dumps(dict(combo=name)))
        for name,relative in effects.items():
            p=self.root/relative;p.mkdir(parents=True);(p/('evaluation-report.json' if name.startswith('text-') else 'report.json')).write_text('{}')
        selected=base/'inputs.json';selected.write_text(json.dumps(dict(live=live,effects=effects)))
        result=json.loads(self.runcli('final-plan',str(base.relative_to(self.root)),'--inputs',str(selected.relative_to(self.root))).stdout)
        self.assertEqual(len(result['liveInputs']),23)
        self.assertEqual(result['preflight'].count('-live'),23)
        self.assertEqual(result['merge'][1],'-matrix')
        self.assertEqual(len(result['merge'][3:]),23)
        self.assertFalse(result['credentialsRead'])
        broken=dict(live);broken.pop('openai-chat');selected.write_text(json.dumps(dict(live=broken,effects=effects)))
        self.runcli('final-plan',str(base.relative_to(self.root)),'--inputs',str(selected.relative_to(self.root)),ok=False)
        self.assertFalse((base/'finalization').exists())

    def test_duplicate_selected_name_refuses_before_starting_child(self):
        self.fake_build()
        self.runcli('prepare','.evidence/provider-sites/current','--phase','candidate')
        base=self.root/'.evidence/provider-sites/current'
        self.runcli('check','.evidence/provider-sites/current');self.approve(base)
        (self.root/'.env').rmdir()
        (self.root/'.env').write_text('ZAI_KEY=SYNTHETIC-FIRST\nZAI_KEY=SYNTHETIC-SECOND\nMINIMAX_CN_KEY=UNSELECTED-SECRET\n')
        result=self.runcli('run-selected','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        self.assertFalse((base/'selected-child-started').exists())
        self.assertNotIn('SYNTHETIC-FIRST',result.stdout+result.stderr)
        self.assertNotIn('SYNTHETIC-SECOND',result.stdout+result.stderr)

    def test_one_selected_key_starts_only_its_approved_child_and_unknown_native_stays_unknown(self):
        self.fake_build();self.runcli('prepare','.evidence/provider-sites/current','--phase','candidate')
        base=self.root/'.evidence/provider-sites/current'
        self.runcli('check','.evidence/provider-sites/current');self.approve(base)
        (self.root/'.env').rmdir();(self.root/'.env').write_text('ZAI_KEY=SYNTHETIC-ONLY-SELECTED\nMINIMAX_CN_KEY=SYNTHETIC-NOT-SELECTED\n')
        self.runcli('run-selected','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        self.assertTrue((base/'selected-child-started').exists())
        command=json.loads((base/'candidate-zai-chat-command.json').read_text())
        self.assertEqual(command['exitCode'],0)
        self.assertTrue(command['credentialsRead'])
        self.assertIn('BARNESS_REGIONAL_KEY',command['environmentKeys'])
        self.assertNotIn('MINIMAX_CN_KEY',command['environmentKeys'])
        self.assertNotIn('OPENAI_KEY',command['environmentKeys'])
        self.assertNotIn('HTTP_PROXY',command['environmentKeys'])
        record=json.loads((self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-1-candidate-zai-chat.json').read_text())
        self.assertEqual(record['budget']['state'],'UNKNOWN')
        self.assertEqual(record['reservedMicrousd'],1000000)
        log=(base/'candidate-zai-chat.log').read_text()
        self.assertNotIn('SYNTHETIC-ONLY-SELECTED',log)
        self.assertNotIn('SYNTHETIC-NOT-SELECTED',log)

    def test_usd_exhaustion_and_wrong_provider_refuse_before_claim_or_key(self):
        self.fake_build();self.runcli('prepare','.evidence/provider-sites/current','--phase','candidate')
        base=self.root/'.evidence/provider-sites/current'
        self.runcli('check','.evidence/provider-sites/current');self.approve(base,reserve=30000000)
        self.runcli('run-selected','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        marker=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-1-candidate-zai-chat.json'
        self.assertFalse(marker.exists());self.assertFalse((base/'candidate-zai-chat-command.json').exists())
        auth=json.loads((base/'execution-authorization.json').read_text());auth['standardCredentialScopes']['ZAI_KEY']['provider']='zhipu'
        (base/'execution-authorization.json').write_text(json.dumps(auth))
        self.runcli('run-selected','.evidence/provider-sites/current','candidate-zai-chat',ok=False)
        self.assertFalse(marker.exists());self.assertTrue((self.root/'.env').is_dir())

    def test_inventory_is_fixed_complete_and_keyless(self):
        value = json.loads(self.runcli('inventory').stdout)
        self.assertEqual(value['round'], 'provider-sites-final-acceptance-1')
        self.assertEqual((value['maxCalls'], value['maxAttempts']), (726, 762))
        self.assertEqual(len(value['jobs']), 33)
        self.assertEqual(sum(j['maxCalls'] for j in value['jobs'].values()) + 4, 726)
        self.assertEqual(sum(j['maxAttempts'] for j in value['jobs'].values()) + 4, 762)
        self.assertEqual(value['jobs']['glm-chat']['sourceVariable'], 'ZHIPU_KEY')
        self.assertEqual(value['jobs']['google-gemini']['sourceVariable'], 'GEMINI_KEY')
        self.assertEqual(value['jobs']['candidate-minimax-cn-messages']['sourceVariable'], 'MINIMAX_CN_KEY')
        self.assertFalse(value['credentialsRead'])
        self.assertFalse(value['automaticProcessReruns'])
        self.assertTrue((self.root / '.env').is_dir())
        self.assertNotIn('FORBIDDEN-AMBIENT', json.dumps(value))

if __name__ == '__main__':
    unittest.main()
