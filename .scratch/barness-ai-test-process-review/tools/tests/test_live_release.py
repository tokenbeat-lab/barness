import json
from live_fixture import live_bundle
from test_release import ReleaseRecordProcess as Base

# Inherited setup builds the actual public CLI. All records live in a fake repository.
class LiveQualificationProcess(Base):
    def test_zero_call_pass_labels_do_not_substitute_for_native_live_qualification(self):
        from process_support import digest
        bundle, stage, _ = live_bundle(self)
        out = bundle.parent.parent / 'consumer-rejected'
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', out, expected=1)
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
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', out, expected=0)
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
                result = self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', out, expected=expected)
                self.assertEqual(result.stdout, (out / 'stdout.jsonl').read_text())
                self.assertEqual(result.stderr, (out / 'stderr.log').read_text())
                self.assertEqual(result.returncode, int((out / 'exit-code').read_text()))

    def test_consumer_receipt_accepts_equivalent_explicit_flag_order(self):
        from process_support import digest
        bundle, stage, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        out = bundle.parent.parent / 'consumer-reordered'
        self.call(self.binary, '-out=' + str(out), '-live-bundle=' + str(bundle.relative_to(self.repo)), expected=0)
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
                self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', out, expected=1)
                self.assertFalse(json.loads((out / 'live-consumer.json').read_text())['pass'])

    def test_consumer_output_cannot_modify_native_bundle_through_direct_or_symbolic_path(self):
        bundle, _, _ = live_bundle(self, 'typesafe-classifier', qualified=True)
        before = {str(p.relative_to(bundle)): p.read_bytes() for p in bundle.rglob('*') if p.is_file()}
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', bundle / 'injected', expected=1)
        self.assertFalse((bundle / 'injected').exists())
        alias = self.repo / 'bundle-alias'
        alias.symlink_to(bundle, target_is_directory=True)
        self.call(self.binary, '-live-bundle', bundle.relative_to(self.repo), '-out', alias / 'injected', expected=1)
        self.assertFalse((bundle / 'injected').exists())
        after = {str(p.relative_to(bundle)): p.read_bytes() for p in bundle.rglob('*') if p.is_file()}
        self.assertEqual(before, after)


# Reuse setup and fixture helpers, without running the base class's independent
# scenarios a second time under this module's discovery.
def load_tests(loader, tests, pattern):
    import unittest
    return unittest.TestSuite(LiveQualificationProcess(name) for name in LiveQualificationProcess.__dict__
                              if name.startswith('test_'))
