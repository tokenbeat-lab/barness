import os
from pathlib import Path
import shutil
import shlex
import subprocess
import tempfile
import unittest
from process_support import safe_environment, TOOLS

PROJECT = TOOLS.parents[2]


class OrdinaryEntrypointProcess(unittest.TestCase):
    def test_minimax_consumer_never_starts_a_hidden_load_process(self):
        with tempfile.TemporaryDirectory(prefix='barness-ordinary-boundary-') as tmp:
            root = Path(tmp)
            (root / 'go').write_text('#!/bin/sh\ncase " $* " in *TestMiniMaxSharedDesignLoad*) echo hidden-load > ' + shlex.quote(str(root / 'started')) + '; exit 91;; *) exec ' + shlex.quote(shutil.which('go')) + ' "$@";; esac\n')
            (root / 'go').chmod(0o755)
            env = safe_environment()
            env['PATH'] = str(root) + os.pathsep + env['PATH']
            env['BARNESS_BOUNDARY_MARKER'] = str(root / 'started')
            env['BARNESS_AI_EVIDENCE_DIR'] = str(root / 'bundles')
            binary = root / 'ordinary.test'
            built = subprocess.run([shutil.which('go'), 'test', '-c', '-o', binary, './ai/e2e'], cwd=PROJECT,
                                   env=safe_environment(),
                                   capture_output=True, text=True, timeout=90)
            self.assertEqual(built.returncode, 0, built.stderr)
            result = subprocess.run([binary, '-test.run=^TestMiniMaxReleaseLoadConsumer$', '-test.count=1', '-test.timeout=60s'],
                cwd=PROJECT / 'ai/e2e', env=env, capture_output=True,text=True,timeout=90)
            artifact = os.environ.get('BARNESS_PROCESS_TEST_ARTIFACT_DIR')
            if artifact:
                dest = Path(artifact);dest.mkdir(parents=True,exist_ok=True)
                (dest / 'ordinary.stdout.log').write_text(result.stdout)
                (dest / 'ordinary.stderr.log').write_text(result.stderr)
                (dest / 'ordinary.exit-code').write_text(str(result.returncode)+'\n')
            self.assertFalse((root / 'started').exists(), 'ordinary consumer attempted a hidden pressure subprocess')
            self.assertEqual(result.returncode,0,result.stdout+result.stderr)

    def test_manual_pressure_entrypoints_are_not_run_without_opt_in(self):
        import json
        with tempfile.TemporaryDirectory(prefix='barness-manual-boundary-') as tmp:
            root = Path(tmp)
            env = safe_environment()
            binary = root / 'ordinary.test'
            built = subprocess.run([shutil.which('go'), 'test', '-c', '-o', binary, './ai/e2e'], cwd=PROJECT,
                                   env=env, capture_output=True, text=True, timeout=90)
            self.assertEqual(built.returncode, 0, built.stderr)
            env['BARNESS_AI_EVIDENCE_DIR'] = str(root / 'bundles')
            pattern = '^(TestAdmissionBurstPressure|TestGLMSharedDesignLoad|TestMultimodalRuntimeDesignLoad|TestMiniMaxSharedDesignLoad|TestPolicyPressure|TestMixedPressure|TestClassifierImagesPressure|TestClassifierImagesPressureImport|TestPressureLoopback|TestToolArgumentsScaling|TestPressureToolWriteIntegrity|TestPressureFixtureWriteBackpressure|TestMiniMaxReleasePressureLoadConsumer)$'
            result = subprocess.run([binary, '-test.v', '-test.run=' + pattern, '-test.count=1', '-test.timeout=30s'],
                                    cwd=PROJECT / 'ai/e2e', env=env, capture_output=True, text=True, timeout=40)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            manifests = list((root / 'bundles').glob('*/manifest.json'))
            self.assertEqual(len(manifests), 1)
            cases = json.loads(manifests[0].read_text())['cases']
            self.assertTrue(cases)
            self.assertTrue(all(c['status'] == 'NOT_RUN' for c in cases), 'manual capacity proof executed in ordinary stage')

    def test_isolated_go_home_preserves_host_home_redaction_in_native_artifacts(self):
        import json
        import pwd
        import sys
        env = safe_environment()
        result = subprocess.run([sys.executable, str(TOOLS / 'run.py'), '--feature', 'barness-ai-test-process-review',
            '--stage', 'e2e', '--', 'go', 'test', '-json', './ai/e2e', '-run',
            '^TestRegionalProviderReleaseGrantAliases$/^offset$', '-count=1', '-timeout=60s'],
            cwd=PROJECT, env=env, capture_output=True, text=True, timeout=90)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        receipt_path = PROJECT / json.loads(result.stdout)['attempt'] / 'receipt.json'
        receipt = json.loads(receipt_path.read_text())
        bundle = receipt_path.parent / next(b['path'] for b in receipt['bundles'] if b['package'] == './ai/e2e')
        home = pwd.getpwuid(os.getuid()).pw_dir.encode()
        exposed = any(home in p.read_bytes() for p in (bundle / 'RP18-grant-alias-offset').rglob('*.json'))
        self.assertFalse(exposed, 'isolated HOME lost the actual host-home redaction policy')
        artifact = os.environ.get('BARNESS_PROCESS_TEST_ARTIFACT_DIR')
        if artifact:
            (Path(artifact) / 'home-redaction-attempt.txt').write_text(str(receipt_path.relative_to(PROJECT)) + '\n')
