"""Synthetic protocol documents for the read-only CLI seam; no vendor requests."""
import json
from process_support import digest, write_json
from test_release import PROJECT


def live_bundle(test, combo='openai-decisions', qualified=False):
    stage = test.stage('live/' + combo)
    attempt = test.repo / '.evidence/sample/live' / combo / 'attempt'
    bundle = attempt / 'bundles/native'
    catalog = json.loads((PROJECT / 'ai/release/catalog-snapshot.json').read_text())
    expected = ['text-choice', 'text-score', 'text-bool', 'mixed-questions', 'image-png', 'image-jpeg',
                'image-webp', 'cancel-in-flight', 'recovery']
    model, provider, api = 'gpt-6-luna', 'openai', 'openai-decisions'
    package = '-tags live -timeout 180s ./ai/live'
    budget, extra = {}, {}
    if combo == 'typesafe-classifier':
        expected = ['mixed-questions', 'single-choice', 'context-422']
        model, provider, api = 'jev-1.13.0', 'typesafe', 'typesafe-system-one'
        package = '-tags live ./ai/live'
        if qualified:
            budget = {'maxCalls': 6, 'maxRetriesPerScenario': 1, 'maxQuestions': 10, 'maxStateBytesPerCall': 131072,
                      'callsUsed': 3, 'questionsUsed': 5, 'httpAttempts': 3, 'stateBytesSubmitted': 80100}
    cases, scenarios = [], []
    for ident in expected:
        case = bundle / ident
        write_json(case / 'assertions.json', {'case_id': ident, 'status': 'PASS', 'assertions': [{'pass': True}]})
        scenario = {'id': ident, 'caseId': ident, 'model': model, 'outcome': 'PASS', 'calls': 0, 'attempts': 0}
        if qualified:
            scenario.update(calls=1, attempts=1, questions=3 if ident == 'mixed-questions' else 1,
                            providerRequestIds=['synthetic-request-id'])
            for filename in ('attempt-0-exchanges.json', 'attempt-0-result.json', 'observations-0.json'):
                write_json(case / filename, [] if filename.startswith('observations') else {'synthetic': True})
        cases.append({'id': ident, 'dir': ident, 'status': 'PASS',
                      'artifacts': {p.name: digest(p) for p in case.iterdir()}})
        scenarios.append(scenario)
    stamp = '2026-10-22T00:00:00Z'
    report = {'schema': 3, 'combo': combo, 'operation': 'classifier', 'provider': provider, 'api': api,
              'endpoint': 'https://api.typesafe.ai/v1' if provider == 'typesafe' else 'https://api.openai.com/v1',
              'model': model, 'sdk': 'direct HTTP', 'accountAlias': 'synthetic', 'gitCommit': test.source['commit'],
              'catalogVersion': catalog['catalog']['version'], 'catalogHash': catalog['hash'],
              'startedAt': stamp, 'finishedAt': stamp, 'budget': budget, 'expected': expected, 'scenarios': scenarios, **extra}
    write_json(bundle / 'live-report.json', report)
    write_json(bundle / 'manifest.json', {'package': package, 'git_commit': test.source['commit'], 'started_at': stamp,
        'finished_at': stamp, 'versions': {'live_combo': combo, 'live_sdk': 'direct HTTP',
        'model_catalog_version': report['catalogVersion'], 'model_catalog_hash': report['catalogHash']}, 'cases': cases})
    path = attempt / 'receipt.json'
    receipt = json.loads(path.read_text())
    argv = ['synthetic-live-binary', '-test.run=^TestSynthetic$']
    receipt.update(stage='live', route=combo, argv=argv)
    receipt['bundles'] = [{'path': 'bundles/native', 'package': package, 'manifestSHA256': digest(bundle / 'manifest.json'),
                           'auditSHA256': digest(bundle / 'audit.json')}]
    command_path = attempt / 'command.json'
    command = json.loads(command_path.read_text())
    command.update(stage='live', route=combo, argv=argv)
    write_json(command_path, command)
    receipt['files']['command.json'] = receipt['commandSHA256'] = digest(command_path)
    write_json(path, receipt)
    stage.update(receipt={'path': str(path.relative_to(test.repo)), 'sha256': digest(path)}, requiredCases=expected,
                 selection={'by': 'human', 'argv': argv},
                 consumers=[{'argv': ['true'], 'exitCode': 0, 'files': {'stdout.jsonl': digest(attempt / 'stdout.jsonl')}}])
    test.record['impact']['liveRoutes'] = [combo]
    test.record['live'] = {'status': 'PASS', 'routes': {combo: stage}}
    return bundle, stage, report
