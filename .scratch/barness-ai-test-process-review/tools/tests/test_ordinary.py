import os
from pathlib import Path
import shutil
import shlex
import subprocess
import tempfile
import unittest
from process_support import TOOLS

PROJECT = TOOLS.parents[2]


class OrdinaryEntrypointProcess(unittest.TestCase):
    def test_minimax_consumer_never_starts_a_hidden_load_process(self):
        with tempfile.TemporaryDirectory(prefix='barness-ordinary-boundary-') as tmp:
            root = Path(tmp)
            (root / 'go').write_text('#!/bin/sh\ncase " $* " in *TestMiniMaxSharedDesignLoad*) echo hidden-load > ' + shlex.quote(str(root / 'started')) + '; exit 91;; *) exec ' + shlex.quote(shutil.which('go')) + ' "$@";; esac\n')
            (root / 'go').chmod(0o755)
            env = {k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ}
            env['PATH'] = str(root) + os.pathsep + env['PATH']
            env['BARNESS_BOUNDARY_MARKER'] = str(root / 'started')
            env['BARNESS_AI_EVIDENCE_DIR'] = str(root / 'bundles')
            binary = root / 'ordinary.test'
            built = subprocess.run([shutil.which('go'), 'test', '-c', '-o', binary, './ai/e2e'], cwd=PROJECT,
                                   env={k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ},
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
            env = {k: os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ}
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
