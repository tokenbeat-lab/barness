import json
import sys
from process_support import TOOLS
from live_fixture import live_bundle
from test_release import ReleaseRecordProcess as Base

# Inherited setup builds the actual public CLI. All records live in a fake repository.
class LiveQualificationProcess(Base):
    @classmethod
    def setUpClass(cls):
        super().setUpClass()
        import subprocess
        from process_support import safe_environment
        from test_release import PROJECT
        cls.claim_fixture = cls.build.name + '/process-grant-fixture'
        built = subprocess.run(['go', 'build', '-o', cls.claim_fixture,
                                './ai/e2e/testdata/process-grant-fixture'], cwd=PROJECT,
                               env=safe_environment(), capture_output=True, text=True, timeout=180)
        if built.returncode:
            raise RuntimeError(built.stderr)

    def setUp(self):
        super().setUp()
        from process_support import digest
        # Typed native scope includes both module files. Rebuild these purely
        # synthetic stage documents after completing their private source set.
        (self.repo / 'go.sum').write_text('')
        self.source['files']['go.sum'] = digest(self.repo / 'go.sum')
        for stage in ('e2e', 'race'):
            # The synthetic builder inventories pre-existing top-level files;
            # remove its old document before rebuilding to avoid a self-hash.
            (self.repo / '.evidence/sample' / stage / 'attempt/receipt.json').unlink()
            self.record[stage] = self.stage(stage)

    def test_claim_fixture_rejects_unowned_git_and_output_boundaries(self):
        targets = []
        ordinary = self.repo / 'ordinary-checkout'
        ordinary.mkdir()
        targets.append((ordinary, ordinary / 'bundle'))
        linked = self.repo / 'barness-process-e2e-linked'
        self.call('git', 'worktree', 'add', '--detach', str(linked), 'HEAD', expected=0)
        targets.append((linked, linked / 'bundle'))
        symbolic = self.repo / 'barness-process-e2e-symbolic'
        symbolic.mkdir()
        (symbolic / '.git').symlink_to(self.repo / '.git', target_is_directory=True)
        targets.append((symbolic, symbolic / 'bundle'))
        targets.append((self.repo, self.repo.parent / ('outside-' + self.repo.name)))
        (self.repo / 'escape').symlink_to(self.repo.parent, target_is_directory=True)
        targets.append((self.repo, self.repo / 'escape' / 'bundle'))
        for cwd, out in targets:
            with self.subTest(cwd=cwd.name, out=str(out)):
                self.call(sys.executable, '-c',
                          'import os,sys; os.chdir(sys.argv[1]); os.execv(sys.argv[2],sys.argv[2:])',
                          cwd, self.claim_fixture, '-binary', self.binary, '-out', out, expected=1)
                self.assertFalse(out.exists())
                self.assertFalse((self.repo / '.git/barness-live-process-grants').exists())

    def consume(self, bundle, out, expected=None, binary=None, env=None):
        return self.call(sys.executable, TOOLS / 'consume_live.py', '--binary', binary or self.binary,
                         '--bundle', bundle.relative_to(self.repo), '--out', out.relative_to(self.repo),
                         expected=expected, env=env)

    def test_outer_capture_records_actual_native_failure_and_cannot_claim_pass(self):
        import resource
        from process_support import digest
        bundle, stage, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        good = bundle.parent.parent / 'native-before-failure'
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', good, expected=0)
        limit = (good / 'live-consumer.json').stat().st_size - 64
        # An external launcher applies a real file-size failure only to the
        # native process; the outer capture retains its actual streams/exit.
        launcher = self.repo / 'limited/releasegate'
        launcher.parent.mkdir()
        launcher.write_text('#!' + sys.executable + '\nimport os,resource,signal,sys\n'
            'signal.signal(signal.SIGXFSZ, signal.SIG_IGN)\n'
            'resource.setrlimit(resource.RLIMIT_FSIZE, (' + str(limit) + ',' + str(limit) + '))\n'
            'os.execv(' + repr(str(self.binary)) + ', [' + repr(str(self.binary)) + '] + sys.argv[1:])\n')
        launcher.chmod(0o700)
        out = bundle.parent.parent / 'consumer-captured-failure'
        self.consume(bundle, out, binary=launcher, expected=1)
        self.assertEqual((out / 'stdout.jsonl').read_text(), '')
        self.assertEqual((out / 'stderr.log').read_text(), 'releasegate: native live consumer failed\n')
        self.assertEqual((out / 'exit-code').read_text(), '1\n')
        proof = out / 'live-consumer.json'
        self.assertFalse(json.loads(proof.read_text())['pass'])
        stage['consumers'] = [{'path': str(proof.relative_to(self.repo)), 'sha256': digest(proof)}]
        self.gate(1)

    def test_consumer_final_write_failure_does_not_leave_success_capture(self):
        import resource
        import signal
        bundle, _, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        good = bundle.parent.parent / 'consumer-before-failure'
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', good, expected=0)
        # Allow timestamp precision to vary without making the failure flaky.
        limit = (good / 'live-consumer.json').stat().st_size - 64
        def disk_failure():
            signal.signal(signal.SIGXFSZ, signal.SIG_IGN)
            resource.setrlimit(resource.RLIMIT_FSIZE, (limit, limit))
        bad = bundle.parent.parent / 'consumer-write-failed'
        result = self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', bad,
                           child_setup=disk_failure, expected=1)
        self.assertFalse((bad / 'exit-code').exists(), 'native consumer manufactured its own exit')
        self.assertFalse((bad / 'stderr.log').exists(), 'native consumer manufactured its own stderr')

    def test_outer_finalization_failure_keeps_nonpassing_record_and_real_child_exit(self):
        import resource
        import signal
        from process_support import digest
        bundle, stage, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        good = bundle.parent.parent / 'outer-before-failure'
        self.consume(bundle, good, expected=0)
        limit = (good / 'live-consumer.json').stat().st_size - 64
        def disk_failure():
            signal.signal(signal.SIGXFSZ, signal.SIG_IGN)
            resource.setrlimit(resource.RLIMIT_FSIZE, (limit, limit))
        out = bundle.parent.parent / 'outer-write-failed'
        self.call(sys.executable, TOOLS / 'consume_live.py', '--binary', self.binary,
                  '--bundle', bundle.relative_to(self.repo), '--out', out.relative_to(self.repo),
                  child_setup=disk_failure, expected=1)
        proof = out / 'live-consumer.json'
        value = json.loads(proof.read_text())
        self.assertEqual(value['status'], 'RUNNING')
        self.assertFalse(value['pass'])
        self.assertIsNone(value['exitCode'])
        # The saved exit belongs to the actual native child. Its success is
        # retained truthfully but cannot turn the incomplete capture into PASS.
        self.assertEqual((out / 'exit-code').read_text(), '0\n')
        self.assertEqual((out / 'stdout.jsonl').read_bytes(), (out / 'native/live-consumer.json').read_bytes())
        stage['consumers'] = [{'path': str(proof.relative_to(self.repo)), 'sha256': digest(proof)}]
        self.gate(1)

    def test_chat_qualification_rejects_zero_calls_bad_identity_budget_and_capture_disagreement(self):
        from live_fixture import chat_bundle, refresh_chat_receipt
        from process_support import digest, write_json
        for mode in ('zero-calls', 'empty-budget', 'exceeded-budget', 'wrong-model', 'wrong-endpoint',
                     'wrong-provider', 'wrong-scenario-model', 'missing-capture', 'empty-capture',
                     'wrong-attribution', 'wrong-observer', 'reused-call', 'wrong-counters'):
            with self.subTest(mode=mode):
                bundle, stage, report = chat_bundle(self)
                case = bundle / 'text-stream'
                if mode == 'zero-calls':
                    for s in report['scenarios']: s.update(calls=0, attempts=0)
                    report['budget'].update(callsUsed=0, httpAttempts=0)
                    for p in bundle.glob('*/**/*.json'):
                        if p.name != 'assertions.json': p.unlink()
                if mode == 'empty-budget': report['budget'] = {}
                if mode == 'exceeded-budget': report['budget']['callsUsed'] = 33
                if mode == 'wrong-model': report['model'] = 'nonexistent-model'
                if mode == 'wrong-endpoint': report['endpoint'] = 'https://proxy.invalid'
                if mode == 'wrong-provider': report['provider'] = 'openai'
                if mode == 'wrong-scenario-model': report['scenarios'][0]['model'] = 'nonexistent-model'
                if mode == 'missing-capture': (case / 'attempt-1-call-result.json').unlink()
                if mode == 'empty-capture': write_json(case / 'attempt-1-call-exchanges.json', [])
                if mode == 'wrong-attribution':
                    p = case / 'attempt-1-call-result.json'
                    value = json.loads(p.read_text()); value['result']['metadata']['modelId'] = 'substituted'
                    write_json(p, value)
                if mode == 'wrong-observer': write_json(case / 'observations-attempt-1-call.json', [])
                if mode == 'reused-call':
                    for p in (bundle / 'text-complete').glob('*call*.json'):
                        p.write_bytes((case / p.name).read_bytes())
                if mode == 'wrong-counters': report['budget']['httpAttempts'] = 7
                write_json(bundle / 'live-report.json', report)
                refresh_chat_receipt(bundle, stage)
                out = bundle.parent.parent / ('consumer-chat-' + mode)
                self.consume(bundle, out, expected=1)
                proof = out / 'live-consumer.json'
                self.assertFalse(json.loads(proof.read_text())['pass'])
                stage['consumers'] = [{'path': str(proof.relative_to(self.repo)), 'sha256': digest(proof)}]
                self.gate(1)

    def test_complete_chat_captures_are_consumed_readonly(self):
        from live_fixture import chat_bundle
        from process_support import digest
        bundle, stage, _ = chat_bundle(self)
        before = {str(p.relative_to(bundle)): digest(p) for p in bundle.rglob('*') if p.is_file()}
        out = bundle.parent.parent / 'consumer-chat-valid'
        self.consume(bundle, out, expected=0)
        proof = out / 'live-consumer.json'
        stage['consumers'] = [{'path': str(proof.relative_to(self.repo)), 'sha256': digest(proof)}]
        try:
            self.gate(0)
        except AssertionError:
            self.fail((self.repo / ('report-' + str(self.out_number)) / 'release-report.json').read_text())
        self.assertEqual(before, {str(p.relative_to(bundle)): digest(p) for p in bundle.rglob('*') if p.is_file()})

    def test_chat_history_requires_complete_host_provenance_and_downgrade_proof(self):
        from live_fixture import chat_bundle, refresh_chat_receipt
        from process_support import write_json
        for mode in ('missing-grant', 'label-only', 'provenance', 'downgrade'):
            with self.subTest(mode=mode):
                bundle, stage, _ = chat_bundle(self)
                case = bundle / 'xai-origin-history'
                self.assertTrue(case.is_dir(), 'complete mandatory history fixture is missing')
                if mode == 'missing-grant':
                    (bundle / 'process-grant.json').unlink()
                if mode == 'label-only':
                    for name in ('public-request.json', 'immutable-input.json', 'history-provenance.json'):
                        (case / name).unlink()
                if mode == 'provenance':
                    path = case / 'history-provenance.json'
                    value = json.loads(path.read_text())
                    value['after'][0]['accountScopeId'] = 'different-source'
                    write_json(path, value)
                if mode == 'downgrade':
                    path = case / 'attempt-1-call-result.json'
                    value = json.loads(path.read_text())
                    value['result']['metadata']['nativeStateDowngrades'] = {}
                    write_json(path, value)
                    write_json(case / 'observations-attempt-1-call.json',
                               [{'kind': 'call_finished', 'call': value['result']['metadata']}])
                refresh_chat_receipt(bundle, stage)
                self.consume(bundle, bundle.parent.parent / ('history-invalid-' + mode), expected=1)

    def test_zero_call_pass_labels_do_not_substitute_for_native_live_qualification(self):
        from process_support import digest
        bundle, stage, _ = live_bundle(self)
        out = bundle.parent.parent / 'consumer-rejected'
        self.consume(bundle, out, expected=1)
        path = out / 'live-consumer.json'
        self.assertFalse(json.loads(path.read_text())['pass'])
        stage['consumers'] = [{'path': str(path.relative_to(self.repo)), 'sha256': digest(path)}]
        report = self.gate(1)
        self.assertFalse(next(c['pass'] for c in report['checks'] if c['id'] == 'live'))

    def test_real_readonly_consumer_receipt_binds_result_source_and_exact_native_files(self):
        from process_support import digest, write_json
        import copy
        bundle, stage, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        # A native-valid report still cannot use the old self-reported `true` consumer.
        self.gate(1)
        out = bundle.parent.parent / 'consumer-actual'
        self.consume(bundle, out, expected=0)
        proof_path = out / 'live-consumer.json'
        proof = json.loads(proof_path.read_text())
        self.assertTrue(proof['pass'])
        self.assertEqual(proof['exitCode'], 0)
        self.assertEqual(proof['consumer'], 'releasegate/native-live-v1')
        stage['consumers'] = [{'path': str(proof_path.relative_to(self.repo)), 'sha256': digest(proof_path)}]
        self.gate(0)
        original = copy.deepcopy(proof)
        for mode in ('wrong-result', 'wrong-bundle', 'wrong-source', 'wrong-argv', 'missing-log', 'nonzero'):
            with self.subTest(mode=mode):
                changed = copy.deepcopy(original)
                if mode == 'wrong-result': changed['reportSHA256'] = '0' * 64
                if mode == 'wrong-bundle': changed['bundle'] = '.evidence/elsewhere'
                if mode == 'wrong-source': changed['source']['commit'] = '0' * 40
                if mode == 'wrong-argv': changed['argv'] = ['true']
                if mode == 'missing-log': changed['files'] = {}
                if mode == 'nonzero': changed['exitCode'] = 7
                write_json(proof_path, changed)
                stage['consumers'][0]['sha256'] = digest(proof_path)
                self.gate(1)
        write_json(proof_path, original)
        stage['consumers'][0]['sha256'] = digest(proof_path)
        # Changing a native input invalidates an otherwise intact old consumer receipt.
        write_json(bundle / 'extra.json', {'changed': True})
        self.gate(1)

    def test_consumer_capture_equals_actual_public_process_streams_and_exit(self):
        for combo, qualified, expected in (('typesafe-classifier', True, 0), ('openai-decisions', False, 1)):
            with self.subTest(combo=combo):
                bundle, _, _ = live_bundle(self, combo, qualified=qualified)
                out = bundle.parent.parent / 'consumer-streams'
                result = self.consume(bundle, out, expected=expected)
                self.assertEqual(result.stdout, (out / 'stdout.jsonl').read_text())
                self.assertEqual(result.stderr, (out / 'stderr.log').read_text())
                self.assertEqual(result.returncode, int((out / 'exit-code').read_text()))

    def test_consumer_receipt_accepts_equivalent_explicit_flag_order(self):
        from process_support import digest
        bundle, stage, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        out = bundle.parent.parent / 'consumer-reordered'
        self.call(sys.executable, TOOLS / 'consume_live.py', '--out=' + str(out.relative_to(self.repo)),
                  '--bundle=' + str(bundle.relative_to(self.repo)), '--binary=' + str(self.binary), expected=0)
        proof = out / 'live-consumer.json'
        stage['consumers'] = [{'path': str(proof.relative_to(self.repo)), 'sha256': digest(proof)}]
        self.gate(0)

    def test_native_valid_report_cannot_hide_budget_request_or_identity_failures(self):
        from process_support import digest, write_json
        import copy
        bundle, stage, original = live_bundle(self, 'typesafe-classifier', qualified=True)
        for mode in ('budget', 'no-request', 'wrong-model', 'wrong-target'):
            with self.subTest(mode=mode):
                report = copy.deepcopy(original)
                if mode == 'budget': report['budget']['callsUsed'] = 7
                if mode == 'no-request': report['scenarios'][0]['providerRequestIds'] = []
                if mode == 'wrong-model': report['model'] = 'jev-latest'
                if mode == 'wrong-target': report['endpoint'] = 'https://proxy.invalid'
                write_json(bundle / 'live-report.json', report)
                out = bundle.parent.parent / ('consumer-' + mode)
                self.consume(bundle, out, expected=1)
                self.assertFalse(json.loads((out / 'live-consumer.json').read_text())['pass'])

    def test_consumer_output_cannot_modify_native_bundle_through_direct_or_symbolic_path(self):
        bundle, _, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        before = {str(p.relative_to(bundle)): p.read_bytes() for p in bundle.rglob('*') if p.is_file()}
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', bundle / 'injected', expected=1)
        self.assertFalse((bundle / 'injected').exists())
        self.consume(bundle, bundle / 'outer-injected', expected=1)
        self.assertFalse((bundle / 'outer-injected').exists())
        alias = self.repo / 'bundle-alias'
        alias.symlink_to(bundle, target_is_directory=True)
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', alias / 'injected', expected=1)
        self.assertFalse((bundle / 'injected').exists())
        self.consume(bundle, alias / 'outer-injected', expected=1)
        self.assertFalse((bundle / 'outer-injected').exists())
        after = {str(p.relative_to(bundle)): p.read_bytes() for p in bundle.rglob('*') if p.is_file()}
        self.assertEqual(before, after)


# Reuse setup and fixture helpers, without running the base class's independent
# scenarios a second time under this module's discovery.
def load_tests(loader, tests, pattern):
    import unittest
    return unittest.TestSuite(LiveQualificationProcess(name) for name in LiveQualificationProcess.__dict__
                              if name.startswith('test_'))
