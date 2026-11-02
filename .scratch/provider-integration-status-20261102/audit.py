#!/usr/bin/env python3
"""Read-only Provider inventory and evidence-reference audit; no vendor calls."""
import collections
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
OUT = Path(__file__).resolve().parent
sys.dont_write_bytecode = True
sys.path.insert(0, str(ROOT / 'tools/verification'))
from receipt_files import source_snapshot
from run_process import environment, isolate_go_home


def read(path):
    return json.loads((ROOT / path).read_text())


def sha(path):
    return hashlib.sha256((ROOT / path).read_bytes()).hexdigest()


checked = []


def reference(ref):
    actual = sha(ref['path'])
    if actual != ref['sha256']:
        raise ValueError('reference hash differs: ' + ref['path'])
    checked.append(dict(ref))
    return read(ref['path'])


head = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
branch = subprocess.check_output(['git', 'branch', '--show-current'], cwd=ROOT, text=True).strip()
with tempfile.TemporaryDirectory(prefix='barness-provider-inventory-') as temporary:
    temporary = Path(temporary)
    program = temporary / 'catalog.go'
    program.write_text('''package main
import ("encoding/json"; "os"; "github.com/tokenbeat-lab/barness/ai")
func main() {
 c := ai.BuiltinCatalog()
 h, err := c.Hash(); if err != nil { panic(err) }
 if err := json.NewEncoder(os.Stdout).Encode(struct {
  Hash string `json:"hash"`; Catalog ai.Catalog `json:"catalog"`
 }{h, c}); err != nil { panic(err) }
}
''')
    env, _ = environment('e2e')
    isolate_go_home(env, 'e2e', temporary)
    # Use installed dependencies only; inventory must not fetch remote modules.
    env.update(GOPROXY='off', GOSUMDB='off', GOTOOLCHAIN='local')
    started = time.monotonic()
    result = subprocess.run(['go', 'run', str(program)], cwd=ROOT, env=env,
                            capture_output=True, text=True, timeout=120)
    elapsed = time.monotonic() - started
    if result.returncode:
        raise RuntimeError('isolated public catalog probe failed')
    snapshot = json.loads(result.stdout)
    if snapshot != read('ai/catalog_builtin.json'):
        raise ValueError('compiled catalog differs from generated JSON')

(OUT / 'catalog.json').write_text(json.dumps(snapshot, ensure_ascii=False, indent=2) + '\n')
providers = collections.defaultdict(lambda: collections.defaultdict(list))
routes = set()
for operation, key in [('chat', 'models'), ('image', 'imageModels'),
                       ('classifier', 'classifierModels')]:
    for model in snapshot['catalog'].get(key, []):
        providers[model['provider']][operation].append({'api': model['api'], 'id': model['id']})
        routes.add((operation, model['provider'], model['api']))

base = read('.evidence/provider-final-readiness/delivery-matrix-v2.json')
complete = {}
for route in base['paths']:
    if not route['fullPathQualification']:
        continue
    for evidence in route['actualEvidence']:
        consumer = reference(evidence['consumer'])
        if consumer.get('pass') is not True or consumer.get('exitCode') != 0:
            raise ValueError('retained consumer is not PASS: ' + route['route'])
    complete[(route['operation'], route['provider'], route['api'])] = route['route']

minimax = read('.evidence/minimax-thinking-resample/composition-result-v1.json')
if not minimax['pass'] or minimax['qualification'] != 'FORMAL_APPLICABLE' or minimax['coverage'] != 14:
    raise ValueError('MiniMax Messages composition is incomplete')
complete[('chat', 'minimax', 'anthropic-messages')] = 'minimax-messages'

kimi = read('.evidence/kimi-catalog-activation-20261028/completion.json')
qualification = reference(kimi['formalQualification'])
if not kimi['catalogActive'] or not qualification['pass_'] or len(kimi['formalConsumers']) != 6:
    raise ValueError('Kimi activation or six-route qualification is incomplete')
for consumer in kimi['formalConsumers']:
    outer = reference(consumer['outer'])
    reference(consumer['native'])
    reference(consumer['capture'])
    reference(consumer['originalComposition'])
    if outer.get('pass') is not True or outer.get('exitCode') != 0:
        raise ValueError('Kimi formal consumer is not PASS')
for identity in kimi['identities']:
    complete[('chat', identity['provider'], identity['api'])] = 'kimi-9-plus-1'

xai = read('.evidence/barness-ai-xai-imagine/release/20261029-07/release-report.json')
xai_consumer = read('.evidence/barness-ai-xai-imagine/live-consumer/20261029-06-03/live-consumer.json')
if not xai['pass'] or not xai_consumer['pass'] or xai_consumer['exitCode'] != 0:
    raise ValueError('xAI image qualification/release report is not PASS')
complete[('image', 'xai', 'xai-images')] = 'xai-images'
if routes != set(complete):
    raise ValueError('catalog routes and retained complete evidence routes differ')

record = read('.scratch/barness-ai-dynamic-catalog/release-record-v6.json')
receipt = reference(record['e2e']['receipt'])
current = source_snapshot(ROOT)['files']
expected = record['source']['files']
changed = [p for p, digest in expected.items() if not (ROOT / p).is_file() or sha(p) != digest]
added = sorted(set(current) - set(expected))
removed = sorted(set(expected) - set(current))
changed_inputs = [p for p, digest in record['selectedInputs'].items()
                  if not (ROOT / p).is_file() or sha(p) != digest]
if changed or added or removed or changed_inputs:
    raise ValueError('latest complete ordinary evidence differs from current execution source/inputs')
if receipt['status'] != 'PASS' or receipt['exitCode'] != 0:
    raise ValueError('latest ordinary evidence is not PASS')
for stage in ('race', 'pressure'):
    stage_receipt = reference(record[stage]['receipt'])
    if stage_receipt['status'] != 'PASS' or stage_receipt['exitCode'] != 0:
        raise ValueError('catalog ' + stage + ' evidence is not PASS')

matrix = read('ai/live/support-matrix.json')
matrix_routes = {(row['operation'], row['provider'], row['api']) for row in matrix['rows']}
summary = {
    'date': '2026-11-02', 'head': head, 'branch': branch,
    'scope': 'Fixed-date read-only inventory; retained evidence, not a new live qualification or release verdict',
    'probe': {'entry': 'ai.BuiltinCatalog + Catalog.Hash', 'exitCode': result.returncode,
              'elapsedSeconds': elapsed, 'networkDisabled': True},
    'catalogHash': snapshot['hash'], 'catalogVersion': snapshot['catalog']['version'],
    'providers': dict(providers), 'providerCount': len(providers),
    'identityCounts': {key: len(snapshot['catalog'].get(key, []))
                       for key in ('models', 'imageModels', 'classifierModels')},
    'routeCount': len(routes), 'routesWithRetainedCompleteOwnEvidence': len(complete),
    'routeEvidence': [{'operation': op, 'provider': p, 'api': api, 'evidence': complete[(op, p, api)]}
                      for op, p, api in sorted(routes)],
    'latestOrdinary': {'commit': record['source']['commit'], 'receipt': record['e2e']['receipt'],
                       'status': receipt['status'], 'exitCode': receipt['exitCode'],
                       'sourceFiles': len(expected), 'selectedInputs': len(record['selectedInputs']),
                       'sourceAndInputSetsEqual': True},
    'supportMatrix': {'rows': len(matrix['rows']), 'uniqueRoutes': len(matrix_routes),
                      'missingRoutes': [list(r) for r in sorted(routes - matrix_routes)],
                      'notRunRows': [r['combo'] for r in matrix['rows']
                                     if any(c['outcome'] == 'NOT_RUN' for c in r['capabilities'])]},
    'checkedReferences': checked,
    'limits': ['Retained consumer and stage-reference hashes checked; full transitive bundles not re-audited',
               'Original feature race/pressure/release gaps remain distinct from catalog-only PASS',
               'Catalog model inclusion does not establish real measurement of every model ID'],
    'newVendorCalls': 0, 'newTests': 0, 'newRaceOrPressureRuns': 0,
}
(OUT / 'inventory.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({k: summary[k] for k in ('head', 'providerCount', 'identityCounts', 'routeCount',
                                         'routesWithRetainedCompleteOwnEvidence', 'latestOrdinary',
                                         'supportMatrix')}, ensure_ascii=False, indent=2))
