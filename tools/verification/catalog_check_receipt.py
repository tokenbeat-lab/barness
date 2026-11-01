"""Read-only finite catalog-contract receipt consumer, using native inventories."""
import hashlib
import json
from receipt_files import json_file, sha, verify_bundle, within


def required_cases(snapshot):
    result = {}
    for operation, collection in (('chat', 'models'), ('image', 'imageModels'), ('classifier', 'classifierModels')):
        for model in snapshot['catalog'].get(collection, []):
            identity = {'operation': operation, 'provider': model['provider'], 'api': model['api'], 'id': model['id']}
            key = '\0'.join(identity[name] for name in ('operation', 'provider', 'api', 'id'))
            name = hashlib.sha256(key.encode()).hexdigest()[:24]
            if operation == 'chat':
                scenes = ['load-query', 'basic', 'cache-usage', 'history-tools', 'input-modalities', 'parameters', 'authorization']
                if model.get('reasoning'):
                    scenes.append('reasoning')
            elif operation == 'image':
                scenes = ['load-query', 'generation', 'options', 'edit-boundaries', 'authorization']
            else:
                scenes = ['load-query', 'questions', 'formats', 'image-locations', 'limits', 'authorization']
            if operation != 'image':
                for index, _ in enumerate(model.get('cost', {}).get('tiers', [])):
                    scenes.extend('tier-%d-%s' % (index, b) for b in ('below', 'equal', 'above'))
            for scene in scenes:
                case = {'identity': identity, 'scene': scene, 'case': 'CC-' + name + '-' + scene,
                        'test': 'TestCatalogCandidateContracts/' + name + '/' + scene}
                if case['case'] in result:
                    raise ValueError('duplicate candidate identity')
                result[case['case']] = case
    if not result:
        raise ValueError('no candidate identities or contracts')
    return result


def verify_receipt(root, prepared, selected, receipt_name, snapshot):
    path = within(root, receipt_name)
    receipt = json_file(path)
    if receipt.get('status') != 'PASS' or receipt.get('exitCode') != 0 or receipt.get('stage') != 'e2e' or receipt.get('route') is not None:
        raise ValueError('controlled process did not complete successfully')
    if receipt.get('feature') != prepared['feature'] or receipt.get('argv') != prepared['argv']:
        raise ValueError('wrong controlled process entry')
    if receipt.get('source') != prepared['source'] or receipt.get('selectedInputs') != selected:
        raise ValueError('measured source or selected input set differs')
    if receipt.get('toolchain', {}).get('goEnv') != 'off' or receipt.get('toolchain', {}).get('goFlags') != '':
        raise ValueError('uncontrolled Go configuration')
    out = path.parent
    inventory = receipt.get('files')
    if set(inventory or {}) != {'command.json', 'stdout.jsonl', 'stderr.log', 'exit-code'}:
        raise ValueError('incomplete process capture')
    for name, digest in inventory.items():
        if sha(within(out, name)) != digest:
            raise ValueError('process artifact changed')
    if (out / 'exit-code').read_text().strip() != '0' or receipt.get('commandSHA256') != sha(out / 'command.json'):
        raise ValueError('actual process exit or command differs')
    command = json_file(out / 'command.json')
    if command.get('argv') != prepared['argv'] or command.get('stage') != 'e2e' or command.get('route') is not None:
        raise ValueError('process command differs')
    bundles = receipt.get('bundles')
    if not isinstance(bundles, list) or len(bundles) != 1 or bundles[0].get('package') != './ai/e2e':
        raise ValueError('one primary candidate native bundle required')
    bundle = within(out, bundles[0]['path'])
    manifest = verify_bundle(bundle, './ai/e2e')
    if sha(bundle / 'manifest.json') != bundles[0]['manifestSHA256'] or sha(bundle / 'audit.json') != bundles[0]['auditSHA256']:
        raise ValueError('native bundle reference changed')
    if manifest.get('git_commit') != prepared['source']['commit']:
        raise ValueError('native measured commit differs')
    cases = {case['id']: case for case in manifest['cases']}
    plan_case = cases.get('CATALOG-CANDIDATE-PLAN')
    if plan_case is None or plan_case['status'] != 'PASS':
        raise ValueError('missing native contract plan')
    plan = json_file(within(bundle, plan_case['dir'] + '/plan.json'))
    snapshot_name = prepared['directory'] + '/snapshot.json'
    if plan.get('snapshot') != 'sha256:' + sha(within(root, snapshot_name)) or plan.get('version') != snapshot['catalog']['version'] or plan.get('hash') != snapshot['hash']:
        raise ValueError('contract plan describes another snapshot')
    want = required_cases(snapshot)
    planned = {case['case']: case for case in plan.get('cases', [])}
    if len(planned) != len(plan.get('cases', [])) or planned != want:
        raise ValueError('missing, shortened or changed identity/scenario plan')
    expected = set(want) | {'CATALOG-CANDIDATE-PLAN'}
    if set(cases) != expected:
        raise ValueError('native contract case set differs')
    passed = set()
    failed = set()
    with (out / 'stdout.jsonl').open() as stream:
        for line in stream:
            event = json.loads(line)
            if event.get('Action') == 'pass' and event.get('Test'):
                passed.add(event['Test'])
            if event.get('Action') in ('fail', 'skip') and event.get('Test'):
                failed.add(event['Test'])
    required_tests = {case['test'] for case in want.values()} | {'TestCatalogCandidateContracts'}
    if not required_tests <= passed or required_tests & failed or not required_tests <= set(receipt.get('tests', [])):
        raise ValueError('required child scenarios did not actually pass')
    for name, planned_case in want.items():
        actual = cases[name]
        if actual.get('status') != 'PASS' or actual.get('test') != planned_case['test']:
            raise ValueError('required native scenario did not pass')
        assertions = json_file(within(bundle, actual['dir'] + '/assertions.json')).get('assertions')
        if not isinstance(assertions, list) or not assertions or any(item.get('pass') is not True for item in assertions):
            raise ValueError('required scenario lacks complete assertions')
    return {'status': 'applicable', 'conclusion': 'code_compatible', 'measuredCommit': prepared['source']['commit'],
            'requiredCases': sorted(expected), 'receiptSHA256': sha(path), 'snapshotSHA256': sha(within(root, snapshot_name))}
