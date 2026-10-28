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


def chat_bundle(test):
    """Native-shaped Google chat captures for the CLI seam, never vendor proof."""
    bundle, stage, report = live_bundle(test, 'google-gemini')
    import shutil
    for path in bundle.iterdir():
        if path.is_dir():
            shutil.rmtree(path)
        elif path.name in ('process-grant.json', 'process-authorization-receipt.json'):
            path.unlink()
    expected = ['text-stream', 'text-complete', 'tool-round-trip', 'cancel-after-first-frame', 'reasoning-history', 'image', 'xai-origin-history']
    endpoint, model = 'https://generativelanguage.googleapis.com/v1beta', 'gemini-3.8-flash'
    report.update(operation='chat', provider='google', api='google-generative-ai', model=model, endpoint=endpoint,
                  expected=expected, scenarios=[], budget={'maxCalls': 32, 'maxRetriesPerScenario': 1,
                  'maxOutputTokensPerCall': 4096, 'callsUsed': 7, 'httpAttempts': 7,
                  'inputTokensUsed': 7, 'outputTokensUsed': 7})
    cases = []
    for ident in expected:
        case = bundle / ident
        metadata = {'tenantId': 'fixture', 'requestId': ident, 'bindingId': 'google-gemini', 'resolved': True,
                    'operation': 'chat', 'providerId': 'google', 'api': 'google-generative-ai', 'modelId': model,
                    'catalogVersion': report['catalogVersion'], 'catalogHash': report['catalogHash'],
                    'attempts': [{'attemptId': ident + '#1', 'httpStatus': 200, 'providerRequestId': ident}]}
        message = {'usage': {'input': 1, 'output': 1}}
        body = xai_history_fixture(case, metadata, message) if ident == 'xai-origin-history' else '{}'
        write_json(case / 'attempt-1-call-result.json', {'error': '', 'result': {
            'metadata': metadata, 'message': message}})
        write_json(case / 'attempt-1-call-exchanges.json', [{'method': 'POST',
            'url': endpoint + '/models/' + model + ':streamGenerateContent',
            'requestBody': body, 'status': 200, 'responseHeaders': {}, 'responseBody': '{}'}])
        write_json(case / 'observations-attempt-1-call.json', [{'kind': 'call_finished', 'call': metadata}])
        write_json(case / 'assertions.json', {'case_id': ident, 'status': 'PASS', 'assertions': [{'pass': True}]})
        cases.append({'id': ident, 'dir': ident, 'status': 'PASS',
                      'artifacts': {p.name: digest(p) for p in case.iterdir()}})
        report['scenarios'].append({'id': ident, 'caseId': ident, 'model': model, 'outcome': 'PASS',
                                    'calls': 1, 'attempts': 1, 'providerRequestIds': [ident]})
    write_json(bundle / 'live-report.json', report)
    manifest = json.loads((bundle / 'manifest.json').read_text())
    # The helper derives typed native scope and burns only this fixture's
    # independent Git store; the real checkout's grant namespace is inaccessible.
    claim = json.loads(test.call(test.claim_fixture, '-binary', test.binary, '-out', bundle, expected=0).stdout)
    report.update(startedAt=claim['startedAt'], finishedAt=claim['finishedAt'], sdk=claim['sdk'])
    write_json(bundle / 'live-report.json', report)
    manifest.update(package='-tags live ./ai/live', cases=cases,
                    started_at=report['startedAt'], finished_at=report['finishedAt'])
    manifest['versions'].update(live_sdk=report['sdk'], process_build_source=claim['sourceSha256'])
    write_json(bundle / 'manifest.json', manifest)
    stage['requiredCases'] = expected
    refresh_chat_receipt(bundle, stage)
    return bundle, stage, report


def refresh_chat_receipt(bundle, stage):
    """Seal the supplied test input, including intentional independent mutations."""
    manifest = json.loads((bundle / 'manifest.json').read_text())
    for case in manifest['cases']:
        case['artifacts'] = {p.name: digest(p) for p in (bundle / case['dir']).iterdir() if p.is_file()}
    write_json(bundle / 'manifest.json', manifest)
    path = bundle.parent.parent / 'receipt.json'
    receipt = json.loads(path.read_text())
    receipt['bundles'][0].update(package=manifest['package'], manifestSHA256=digest(bundle / 'manifest.json'))
    write_json(path, receipt)
    stage['receipt']['sha256'] = digest(path)


def xai_history_fixture(case, metadata, message):
    """Fixed synthetic Host contract from grokhistory.Request, not vendor evidence.

    The independent compiled consumer checks this public ai.Request JSON and
    source NativeStateEnvelope proof, plus the destination's Gemini wire form.
    No proof is imported from another provider or from retained live captures.
    """
    import base64
    import copy
    compact = lambda value: json.dumps(value, separators=(',', ':'))
    def call(suffix, label):
        pair = ['same-prefix_' * 8 + suffix, 'SYNTHETIC_XAI_FOREIGN_ITEM_' + suffix]
        ident = 'xar1|' + base64.urlsafe_b64encode(compact(pair).encode()).decode().rstrip('=')
        return {'id': ident, 'name': 'inspect_xai_history', 'arguments': {'label': label}}
    first, second = call('a+b', 'first'), call('a/b', 'second')
    def assistant(content):
        return {'content': content, 'api': 'openai-responses', 'provider': 'xai', 'model': 'grok-4.7',
                'usage': {**dict.fromkeys(('input', 'output', 'cacheRead', 'cacheWrite', 'totalTokens'), 0),
                          'cost': dict.fromkeys(('input', 'output', 'cacheRead', 'cacheWrite', 'total'), 0)},
                'stopReason': 'toolUse', 'timestamp': 0}
    def user(text): return {'content': [{'text': text}]}
    def result(call, marker):
        return {'toolCallId': call['id'], 'toolName': call['name'],
                'content': [{'text': 'SYNTHETIC_XAI_RESULT_' + marker}], 'isError': False}
    host = {'SystemPrompt': '', 'Messages': [user('Inspect the synthetic stored history.'),
        assistant([{'thinking': 'SYNTHETIC_XAI_VISIBLE_REASONING', 'thinkingSignature': compact({
            'type': 'reasoning', 'id': 'SYNTHETIC_XAI_FOREIGN_REASONING_RESOURCE',
            'summary': [], 'encrypted_content': 'SYNTHETIC_XAI_FOREIGN_CIPHER'})},
            {'text': 'SYNTHETIC_XAI_VISIBLE_ANSWER', 'textSignature': compact({
                'v': 1, 'id': 'SYNTHETIC_XAI_FOREIGN_MESSAGE_RESOURCE'})}, first, second]),
        result(first, 'FIRST'), result(second, 'SECOND'), user('Repeat the first inspection in another turn.'),
        assistant([first]), result(first, 'REUSED'),
        user('The tools have finished. Reply with exactly history ready, without calling tools.')],
        'Tools': [{'name': 'inspect_xai_history', 'description': 'Inspect a synthetic stored history label.',
                   'parameters': {'type': 'object', 'properties': {'label': {'type': 'string'}},
                                  'required': ['label'], 'additionalProperties': False}}]}
    write_json(case / 'public-request.json', host)
    write_json(case / 'immutable-input.json', copy.deepcopy(host))
    envelope = {'tenantId': metadata['tenantId'], 'accountScopeId': 'synthetic-xai-source-account',
                'providerId': 'xai', 'api': 'openai-responses', 'modelId': 'grok-4.7'}
    write_json(case / 'history-provenance.json', {'before': [envelope, envelope], 'after': [envelope, envelope]})
    metadata.update(accountScopeId='live-google-gemini', nativeStateDowngrades={'crossModel': 2})
    message.update(api=metadata['api'], provider=metadata['providerId'], model=metadata['modelId'], stopReason='stop')
    parts = [{'text': 'SYNTHETIC_XAI_VISIBLE_REASONING SYNTHETIC_XAI_VISIBLE_ANSWER'}]
    for index, (label, marker) in enumerate(zip(('first', 'second', 'first'), ('FIRST', 'SECOND', 'REUSED'))):
        ident = 'synthetic_call_' + str(index)
        parts.extend([{'functionCall': {'id': ident, 'name': 'inspect_xai_history', 'args': {'label': label}}},
                      {'functionResponse': {'id': ident, 'name': 'inspect_xai_history',
                                            'response': {'result': 'SYNTHETIC_XAI_RESULT_' + marker}}}])
    return compact({'contents': [{'role': 'model', 'parts': parts}]})
