from datetime import datetime, timezone
import copy
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
from process_support import ProcessCase, TOOLS, digest, write_json

PROJECT = TOOLS.parents[2]


class ReleaseRecordProcess(ProcessCase):
    @classmethod
    def setUpClass(cls):
        cls.build = tempfile.TemporaryDirectory(prefix='barness-release-cli-')
        cls.binary = Path(cls.build.name) / 'releasegate'
        env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH') if k in os.environ}
        built = subprocess.run(['go', 'build', '-o', cls.binary, './ai/release/cmd/releasegate'],
                               cwd=PROJECT, env=env, capture_output=True, text=True, timeout=180)
        if built.returncode:
            raise RuntimeError(built.stderr)

    @classmethod
    def tearDownClass(cls):
        cls.build.cleanup()

    def setUp(self):
        super().setUp()
        self.source = {'commit': self.call('git', 'rev-parse', 'HEAD', expected=0).stdout.strip(),
                       'files': {'ai/example.go': digest(self.repo / 'ai/example.go'), 'go.mod': digest(self.repo / 'go.mod')}}
        self.record = {'schema': 1, 'feature': 'sample',
            'currentState': {'path': 'state.md', 'sha256': digest(self.repo / 'state.md')},
            'source': self.source, 'selectedInputs': {},
            'impact': {'liveRoutes': [], 'pressureRequired': False, 'reason': 'synthetic local change'},
            'e2e': self.stage('e2e'), 'race': self.stage('race'),
            'live': {'status': 'NOT_APPLICABLE', 'reason': 'no provider behavior changed'},
            'pressure': {'status': 'NOT_APPLICABLE', 'reason': 'no capacity path changed'}}
        self.record_path = self.repo / 'record.json'
        self.out_number = 0

    def stage(self, stage):
        attempt = self.repo / '.evidence/sample' / stage / 'attempt'
        bundle = attempt / 'bundles/native'
        write_json(bundle / 'fixture-case/assertions.json', {'assertions': [{'passed': True}]})
        write_json(bundle / 'manifest.json', {'package': './ai/e2e', 'git_commit': self.source['commit'], 'cases': [
            {'id': 'fixture-case', 'test': 'TestFixture', 'status': 'PASS', 'dir': 'fixture-case',
             'artifacts': {'assertions.json': digest(bundle / 'fixture-case/assertions.json')}}]})
        write_json(bundle / 'audit.json', {'findings': []})
        argv = ['go', 'test', '-json', './ai/e2e', '-count=1']
        if stage == 'race':
            argv.insert(2, '-race')
        toolchain = {'collectorPython': 'fixture', 'platform': 'FixtureOS', 'architecture': 'fixture', 'executableSHA256': 'a' * 64}
        command = write_json(attempt / 'command.json', {'toolchain': toolchain, 'schema': 1, 'argv': argv, 'stage': stage,
            'route': None, 'timeoutSeconds': 1500, 'environmentNames': ['PATH']})
        (attempt / 'stdout.jsonl').write_text('\n'.join(json.dumps(x) for x in [
            {'Action': 'run', 'Package': 'example.test/fixture/ai/e2e', 'Test': 'TestFixture'},
            {'Action': 'pass', 'Package': 'example.test/fixture/ai/e2e', 'Test': 'TestFixture'},
            {'Action': 'pass', 'Package': 'example.test/fixture/ai/e2e'}]) + '\n')
        (attempt / 'stderr.log').write_text('')
        (attempt / 'exit-code').write_text('0\n')
        stamp = datetime.now(timezone.utc).isoformat()
        receipt = {'toolchain': toolchain, 'schema': 1, 'feature': 'sample', 'stage': stage, 'route': None, 'status': 'PASS', 'exitCode': 0,
            'source': self.source, 'selectedInputs': {}, 'startedAt': stamp, 'finishedAt': stamp,
            'elapsedSeconds': 0, 'argv': argv, 'commandSHA256': digest(command), 'tests': ['TestFixture'],
            'bundles': [{'path': 'bundles/native', 'package': './ai/e2e',
                         'manifestSHA256': digest(bundle / 'manifest.json'), 'auditSHA256': digest(bundle / 'audit.json')}],
            'files': {p.name: digest(p) for p in attempt.iterdir() if p.is_file()}}
        path = write_json(attempt / 'receipt.json', receipt)
        value = {'status': 'PASS', 'receipt': {'path': str(path.relative_to(self.repo)), 'sha256': digest(path)},
                 'requiredCases': ['fixture-case']}
        if stage == 'race':
            value['selection'] = {'by': 'human', 'argv': argv}
        return value

    def gate(self, expected, env=None):
        write_json(self.record_path, self.record)
        self.out_number += 1
        out = self.repo / ('report-' + str(self.out_number))
        self.call(self.binary, '-record', self.record_path, '-out', out, expected=expected, env=env)
        report = json.loads((out / 'release-report.json').read_text())
        self.assertEqual(report['pass'], expected == 0)
        return report

    def test_complete_record_passes_without_starting_verification(self):
        # An executable Go sentinel would fail if the read-only CLI tried to run tests.
        bin_dir = self.repo / 'bin'
        bin_dir.mkdir()
        go = bin_dir / 'go'
        go.write_text('#!/bin/sh\necho forbidden > started-verification\nexit 91\n')
        go.chmod(0o755)
        self.gate(0, env={'PATH': str(bin_dir) + os.pathsep + os.environ['PATH']})
        self.assertFalse((self.repo / 'started-verification').exists())

    def test_missing_failed_not_run_and_fake_na_are_rejected(self):
        original = copy.deepcopy(self.record)
        for mode in ('missing', 'fail', 'not-run', 'no-human', 'fake-live-na', 'fake-pressure-na', 'no-reason', 'missing-impact'):
            with self.subTest(mode=mode):
                self.record = copy.deepcopy(original)
                if mode == 'missing': del self.record['e2e']
                if mode == 'fail': self.record['e2e']['status'] = 'FAIL'
                if mode == 'not-run': self.record['race']['status'] = 'NOT_RUN'
                if mode == 'no-human': del self.record['race']['selection']
                if mode == 'fake-live-na': self.record['impact']['liveRoutes'] = ['zai-chat']
                if mode == 'fake-pressure-na': self.record['impact']['pressureRequired'] = True
                if mode == 'no-reason': del self.record['live']['reason']
                if mode == 'missing-impact': del self.record['impact']['pressureRequired']
                self.gate(1)

    def test_source_artifact_audit_and_empty_tests_are_rejected(self):
        self.gate(0)
        target = self.repo / '.evidence/sample/e2e/attempt/bundles/native/fixture-case/assertions.json'
        target.write_text('{}')
        self.gate(1)
        (self.repo / 'ai/example.go').write_text('package changed\n')
        self.gate(1)

    def test_old_execution_entry_is_rejected_without_running_commands(self):
        result = self.call(self.binary, '-stage', 'check')
        self.assertNotEqual(result.returncode, 0)
        self.assertNotIn('READY_FOR_RACE', result.stdout)

    def test_case_alias_and_race_warning_cannot_be_rehashed_into_pass(self):
        write_json(self.record_path, self.record)
        raw = self.record_path.read_text().replace('"schema": 1', '"schema": 1, "Schema": 1', 1)
        self.record_path.write_text(raw)
        result = self.call(self.binary, '-record', self.record_path, '-out', self.repo / 'alias-report')
        self.assertNotEqual(result.returncode, 0, 'ambiguous JSON field accepted')
        attempt = self.repo / '.evidence/sample/race/attempt'
        (attempt / 'stderr.log').write_text('WARNING: DATA RACE\n')
        receipt_path = attempt / 'receipt.json'
        receipt = json.loads(receipt_path.read_text())
        receipt['files']['stderr.log'] = digest(attempt / 'stderr.log')
        write_json(receipt_path, receipt)
        self.record['race']['receipt']['sha256'] = digest(receipt_path)
        self.gate(1)

    def test_missing_audit_empty_tests_and_wrong_case_are_independent_rejections(self):
        original = copy.deepcopy(self.record)
        path = self.repo / '.evidence/sample/e2e/attempt/receipt.json'
        saved = path.read_text()
        for mode in ('empty-tests', 'wrong-case', 'source-empty', 'missing-audit', 'missing-toolchain'):
            with self.subTest(mode=mode):
                self.record = copy.deepcopy(original)
                path.write_text(saved)
                receipt = json.loads(saved)
                if mode == 'empty-tests': receipt['tests'] = []
                if mode == 'source-empty': receipt['source']['files'] = {}
                if mode == 'missing-toolchain': del receipt['toolchain']
                if mode == 'wrong-case': self.record['e2e']['requiredCases'] = ['absent-case']
                if mode == 'missing-audit': receipt['bundles'][0]['auditSHA256'] = '0' * 64
                write_json(path, receipt)
                self.record['e2e']['receipt']['sha256'] = digest(path)
                self.gate(1)
