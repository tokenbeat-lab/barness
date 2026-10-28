"""Real option producer/consumer proof without a Go -> Python -> Go loop.

Failure inventory: mapped case paths must retain their identity, actual source
and E2E must pass, and missing paid/human stages must still prevent release.
The top-level tool suite owns collection; no vendor credentials are inherited.
"""
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from process_support import ARTIFACT, TOOLS, digest, write_json

sys.path.insert(0, str(TOOLS))
from run_process import environment, isolate_go_home

PROJECT = TOOLS.parents[1]


class OptionProducerProcess(unittest.TestCase):
    def test_actual_option_bundle_passes_source_and_e2e_only(self):
        with tempfile.TemporaryDirectory(prefix='barness-option-process-') as temp:
            env, _ = environment('e2e')
            isolate_go_home(env, 'e2e', Path(temp))
            commands = []

            def call(argv, expected, timeout=180):
                result = subprocess.run([str(x) for x in argv], cwd=PROJECT,
                    env=env, capture_output=True, text=True, timeout=timeout)
                commands.append({'argv': [str(x) for x in argv],
                                 'exitCode': result.returncode})
                self.assertEqual(result.returncode, expected, result.stderr + result.stdout)
                return result

            produced = call([sys.executable, TOOLS / 'run.py', '--feature',
                'verification-option-producer', '--stage', 'e2e', '--', 'go',
                'test', '-json', './ai/e2e', '-run', '^TestXAIReasoningAndBudget$',
                '-count=1', '-timeout=2m'], 0)
            collected = json.loads(produced.stdout)
            self.assertEqual(collected['status'], 'PASS')
            attempt = PROJECT / collected['attempt']
            receipt_path = attempt / 'receipt.json'
            receipt = json.loads(receipt_path.read_text())
            self.assertEqual(len(receipt['bundles']), 1)
            binary = Path(temp) / 'releasegate'
            call(['go', 'build', '-o', binary, './ai/release/cmd/releasegate'], 0)
            state = attempt / 'option-state.md'
            state.write_text('Controlled option proof; paid and human stages pending.\n')
            record = {'schema': 2, 'feature': 'verification-option-producer',
                'currentState': {'path': str(state.relative_to(PROJECT)), 'sha256': digest(state)},
                'source': receipt['source'], 'selectedInputs': receipt['selectedInputs'],
                'impact': {'liveRoutes': ['xai-responses'], 'pressureRequired': True,
                           'reason': 'Controlled process regression; feature stages pending'},
                'e2e': {'status': 'PASS', 'receipt': {
                    'path': str(receipt_path.relative_to(PROJECT)), 'sha256': digest(receipt_path)},
                    'requiredCases': ['XAI-E04-mapped-options-full/default/0',
                                      'XAI-E04-mapped-options-simple/default/128000']},
                'live': {'status': 'NOT_RUN'}, 'race': {'status': 'NOT_RUN'},
                'pressure': {'status': 'NOT_RUN'}}
            index = write_json(attempt / 'option-record.json', record)
            out = attempt / 'option-consumer'
            call([binary, '-record', index, '-out', out], 1)
            report = json.loads((out / 'release-report.json').read_text())
            checks = {c['id']: c['pass'] for c in report['checks']}
            self.assertEqual(checks, {'source': True, 'e2e': True,
                                     'live': False, 'race': False, 'pressure': False})
            self.assertFalse(report['pass'])
            evidence = {'kind': 'real-controlled-option-producer-consumer',
                'commands': commands, 'binarySHA256': digest(binary),
                'receipt': str(receipt_path.relative_to(PROJECT)),
                'receiptSHA256': digest(receipt_path), 'vendorCalls': 0,
                'report': str((out / 'release-report.json').relative_to(PROJECT))}
            write_json(attempt / 'option-process-proof.json', evidence)
            if ARTIFACT:
                write_json(Path(ARTIFACT) / 'option-producer-proof.json', evidence)
