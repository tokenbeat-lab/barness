import json
import sys
from process_support import ProcessCase, TOOLS


class StageRunProcess(ProcessCase):
    def test_ordinary_process_clears_heavy_flags_and_records_exact_bundle(self):
        self.run_stage(env={"BARNESS_AI_PRESSURE": "1", "BARNESS_AI_PIDIFF": "1", "BARNESS_AI_LIVE": "1",
                            "BARNESS_GLM_ROUTE": "synthetic-route", "OPENAI_API_KEY": "synthetic-secret"}, expected=0)
        receipt = json.loads(self.attempts()[0].read_text())
        self.assertEqual(receipt["status"], "PASS")
        self.assertEqual(receipt["exitCode"], 0)
        self.assertEqual(len(receipt["bundles"]), 1)
        attempt = self.attempts()[0].parent
        self.assertTrue((attempt / receipt["bundles"][0]["path"] / "manifest.json").is_file())
        output = (attempt / "stdout.jsonl").read_text()
        self.assertNotIn("synthetic-secret", output)
        self.assertNotIn('"BARNESS_AI_PRESSURE"', output)
        self.assertNotIn('"BARNESS_GLM_ROUTE"', output)

    def test_failed_run_survives_successful_retry(self):
        self.run_stage("fail", expected=1)
        first = self.attempts()[0]
        before = first.read_bytes()
        self.run_stage(expected=0)
        self.assertEqual(len(self.attempts()), 2)
        self.assertEqual(first.read_bytes(), before)
        self.assertEqual(json.loads(before)["exitCode"], 7)

    def test_incomplete_or_ambiguous_evidence_never_passes(self):
        for mode in ("empty", "no-cases", "ambiguous", "no-audit", "bad-audit", "tampered"):
            with self.subTest(mode=mode):
                result = self.run_stage(mode)
                self.assertNotEqual(result.returncode, 0)
                receipt = json.loads(self.attempts()[-1].read_text())
                self.assertNotEqual(receipt["status"], "PASS")

    def test_timeout_and_start_failure_leave_receipts(self):
        self.run_stage("timeout", expected=1)
        self.assertEqual(json.loads(self.attempts()[0].read_text())["status"], "TIMEOUT")
        self.call(sys.executable, TOOLS / "run.py", "--feature", "sample", "--stage", "e2e", "--",
                  "/nonexistent/barness-synthetic-program", expected=1)
        self.assertEqual(json.loads(self.attempts()[-1].read_text())["status"], "START_FAILED")
        self.assertIsNone(json.loads(self.attempts()[-1].read_text())["exitCode"])

    def test_invalid_feature_cannot_escape_root(self):
        result = self.call(sys.executable, TOOLS / "run.py", "--feature", "../../escape", "--stage", "e2e", "--", "true")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.repo.parent / "escape").exists())

    def test_selected_input_and_tool_identity_are_frozen(self):
        (self.repo / 'policy.json').write_text('{"limit": 4}\n')
        self.call(sys.executable, TOOLS / 'run.py', '--feature', 'sample', '--stage', 'e2e',
                  '--input', 'policy.json', '--', sys.executable, TOOLS / 'tests/fixture_child.py', 'valid', expected=0)
        receipt = json.loads(self.attempts()[0].read_text())
        self.assertIn('policy.json', receipt['selectedInputs'])
        self.assertEqual(len(receipt['toolchain']['executableSHA256']), 64)
        self.call(sys.executable, TOOLS / 'run.py', '--feature', 'sample', '--stage', 'e2e',
                  '--input', 'policy.json', '--', sys.executable, TOOLS / 'tests/fixture_child.py', 'input-change', expected=1)
        self.assertEqual(json.loads(self.attempts()[-1].read_text())['status'], 'INVALID_EVIDENCE')

    def test_multiple_packages_keep_all_bundles_but_one_primary(self):
        self.run_stage('auxiliary', expected=0)
        receipt = json.loads(self.attempts()[0].read_text())
        self.assertEqual(len(receipt['bundles']), 2)
        self.assertEqual(sum(b['package'] == './ai/e2e' for b in receipt['bundles']), 1)

    def test_source_inventory_includes_ignored_code_but_allows_prose_only_changes(self):
        (self.repo / '.gitignore').write_text('ai/hidden.go\n')
        (self.repo / 'ai/hidden.go').write_text('package ai\n// ignored code is still compiled\n')
        self.run_stage('valid', expected=0)
        receipt = json.loads(self.attempts()[0].read_text())
        self.assertIn('ai/hidden.go', receipt['source']['files'])
        self.run_stage('doc-change', expected=0)

    def test_cancellation_of_real_go_process_finalizes_failed_receipt(self):
        import os
        import signal
        import subprocess
        import time
        (self.repo / 'ai/cancel_test.go').write_text('package ai\nimport("testing";"time")\nfunc TestControlledCancellation(t *testing.T){time.Sleep(15*time.Second)}\n')
        env = {k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ}
        command = [sys.executable, str(TOOLS / 'run.py'), '--feature', 'sample', '--stage', 'e2e', '--',
                   'go', 'test', '-json', './ai', '-run', '^TestControlledCancellation$', '-count=1']
        process = subprocess.Popen(command, cwd=self.repo, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        try:
            deadline = time.monotonic() + 15
            started = False
            while time.monotonic() < deadline:
                for receipt_path in self.attempts():
                    log = receipt_path.parent / 'stdout.jsonl'
                    if log.exists() and 'TestControlledCancellation' in log.read_text():
                        started = True
                        break
                if started:
                    break
                time.sleep(.05)
            self.assertTrue(started, 'controlled Go child did not start')
            process.send_signal(signal.SIGTERM)
            process.communicate(timeout=12)
            self.assertNotEqual(process.returncode, 0)
            receipt = json.loads(self.attempts()[0].read_text())
            self.assertEqual(receipt['status'], 'CANCELLED')
            self.assertIsInstance(receipt['exitCode'], int)
        finally:
            if process.poll() is None:
                process.kill()
                process.communicate(timeout=5)
