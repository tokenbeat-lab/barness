#!/usr/bin/env python3
"""Read-only Provider audit; does not execute tests or contact vendors."""
import collections
import csv
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]
OUT = Path(__file__).resolve().parent
os.chdir(ROOT)
sys.dont_write_bytecode = True
sys.path.insert(0, str(ROOT / 'tools/verification'))
from run_process import environment, isolate_go_home
from receipt_files import source_snapshot


def read(path):
    return json.loads(Path(path).read_text())


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


record_path = Path('.scratch/barness-ai-grok/verification/record-v18.json')
record = read(record_path)
receipt_path = Path(record['e2e']['receipt']['path'])
receipt = read(receipt_path)
primary = next(b for b in receipt['bundles'] if b['package'] == './ai/e2e')
manifest_path = receipt_path.parent / primary['path'] / 'manifest.json'
audit_path = manifest_path.parent / 'audit.json'
manifest = read(manifest_path)
providers = re.findall(r'Provider\w+\s+ProviderID\s*=\s*"([^"]+)"',
                       Path('ai/binding.go').read_text())

# The public catalog probe reaches current compiled code without credentials.
with tempfile.TemporaryDirectory(prefix='barness-provider-audit-') as temporary:
    temporary = Path(temporary)
    program = temporary / 'catalog.go'
    program.write_text('''package main
import ("encoding/json"; "os"; "github.com/tokenbeat-lab/barness/ai")
func main() { c := ai.BuiltinCatalog(); if _, e := c.Hash(); e != nil { panic(e) }; if e := json.NewEncoder(os.Stdout).Encode(c); e != nil { panic(e) } }
''')
    env, _ = environment('e2e')
    isolate_go_home(env, 'e2e', temporary)
    command = ['go', 'run', str(program)]
    result = subprocess.run(command, env=env, capture_output=True, text=True,
                            timeout=120, check=True)
    catalog = json.loads(result.stdout)
(OUT / 'catalog.json').write_text(json.dumps(catalog, indent=2) + '\n')

catalog_counts = {p: collections.Counter() for p in providers}
for operation, key in [('chat', 'models'), ('image', 'imageModels'),
                       ('classifier', 'classifierModels')]:
    for model in catalog.get(key, []):
        catalog_counts[model['provider']][operation + '/' + model['api']] += 1

# Count manifest cases, never parent Go PASS nodes or individual assertions.
# Families record ownership, not every Provider indirectly exercised by a case.
families = {'P01': 'openai', 'P02': 'anthropic', 'P03': 'google',
            'P04': 'openai', 'P05': 'deepseek', 'P06': 'deepseek',
            'P07': 'typesafe', 'P08': 'openai', 'P09': 'google',
            'P10': 'openai', 'UC02': 'openai', 'UC03': 'openai',
            'UC04': 'typesafe', 'UC05': 'typesafe', 'UC06': 'typesafe',
            'UC07': 'typesafe', 'UC08': 'typesafe', 'UC13': 'typesafe'}
route_suffixes = ['chat', 'responses', 'messages', 'openai-completions',
                  'openai-responses', 'anthropic-messages']
protocol_names = set(re.findall(r'API\w+\s+API\s*=\s*"([^"]+)"',
                                Path('ai/binding.go').read_text()))


def explicit_targets(test):
    # Match complete Provider/route subtest names; hostnames are not IDs.
    return {p for p in providers for part in test.split('/')
            if part == p or (part not in protocol_names and
                             part in {p + '-' + suffix for suffix in route_suffixes})}


def owner(case):
    case_id, test = case['id'], case['test']
    prefix = case_id.split('-')[0]
    if prefix in families:
        return families[prefix], 'case-family:' + prefix
    if prefix.startswith('GLM') or case_id.startswith('P0-GLM'):
        return 'zhipu', 'GLM-case-family'
    if re.fullmatch(r'MM\d+', prefix):
        return 'minimax', 'MiniMax-case-family'
    if prefix == 'XAI' or case_id.startswith('E17-xai-'):
        return 'xai', 'xAI-case-family'
    if re.fullmatch(r'OD\d+', prefix):
        return 'openai', 'Decisions-case-family'
    if prefix == 'KIMI' or prefix.startswith('RP') or prefix == 'regional':
        targets = explicit_targets(test)
        if case_id.startswith('KIMI-CN-INTL-'):
            return 'shared', 'Kimi-cross-market'
        if len(targets) == 1:
            return targets.pop(), 'explicit-target'
        if prefix == 'KIMI' and not targets:
            if case_id.startswith('KIMI-INTL-'):
                return 'moonshot', 'Kimi-INTL-case-family'
            if case_id.startswith('KIMI-CN-'):
                return 'moonshot-cn', 'Kimi-CN-case-family'
        return 'shared', 'multi-or-unspecified-target'
    return 'shared', 'shared-or-unspecified-case-family'


case_counts = collections.defaultdict(collections.Counter)
case_families = collections.defaultdict(collections.Counter)
with (OUT / 'cases.csv').open('w', newline='') as stream:
    writer = csv.writer(stream)
    writer.writerow(['id', 'status', 'test', 'owner', 'ownership_rule'])
    for case in manifest['cases']:
        provider, rule = owner(case)
        case_counts[provider][case['status']] += 1
        case_families[provider][case['id'].split('-')[0]] += 1
        writer.writerow([case['id'], case['status'], case['test'], provider, rule])

code = collections.defaultdict(lambda: {'files': 0, 'lines': 0, 'testFunctions': 0})
testkit = collections.defaultdict(lambda: {'files': 0, 'lines': 0})
oversized = []
for base in ['ai', 'tools/verification']:
    for path in sorted(Path(base).rglob('*')):
        if not path.is_file() or path.suffix not in ['.go', '.py']:
            continue
        if '__pycache__' in path.parts:
            continue
        content = path.read_text()
        lines = len(content.splitlines())
        if path.parent == Path('ai'):
            group = 'ai/root-tests' if path.name.endswith('_test.go') else 'ai/runtime'
        else:
            group = '/'.join(path.parts[:3] if path.parts[1] == 'internal' else path.parts[:2])
        code[group]['files'] += 1
        code[group]['lines'] += lines
        code[group]['testFunctions'] += len(re.findall(r'^func Test\w+\(', content, re.M))
        if str(path).startswith('ai/internal/testkit/'):
            testkit[path.parts[3]]['files'] += 1
            testkit[path.parts[3]]['lines'] += lines
        if lines > 500:
            oversized.append(str(path))

current_source = source_snapshot(ROOT)
expected_source = record['source']['files']
changed = [p for p, sha in expected_source.items()
           if not Path(p).is_file() or digest(p) != sha]
added = sorted(set(current_source['files']) - set(expected_source))
removed = sorted(set(expected_source) - set(current_source['files']))
changed_inputs = [p for p, sha in record['selectedInputs'].items()
                  if not Path(p).is_file() or digest(p) != sha]
go_passes = collections.Counter()
for line in (receipt_path.parent / 'stdout.jsonl').open():
    try:
        event = json.loads(line)
    except json.JSONDecodeError:
        continue
    if event.get('Action') == 'pass' and event.get('Test'):
        go_passes[event['Package']] += 1
live_checkpoint = Path('.evidence/barness-ai-grok/live-batches/'
                       'gf6-20261026T233415Z-ceb8d8c1/completed-checkpoint-v1.json')
live = read(live_checkpoint)
matrix = read('ai/live/support-matrix.json')

summary = {
    'head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
    'providers': providers,
    'catalogVersion': catalog['version'],
    'catalogIdentities': {p: dict(v) for p, v in catalog_counts.items()},
    'caseOwnership': {p: dict(case_counts[p]) for p in providers + ['shared']},
    'caseFamilies': {p: dict(v) for p, v in case_families.items()},
    'totalNativeCases': len(manifest['cases']),
    'nativeStatuses': dict(collections.Counter(c['status'] for c in manifest['cases'])),
    'goPassNodes': dict(go_passes),
    'latestCompletedLiveBatch': {
        'checkpoint': str(live_checkpoint), 'status': live['status'],
        'activity': live['activity'],
        'routes': [{k: r.get(k) for k in ['route', 'status', 'scenariosPass',
                                          'scenariosUnsupported', 'activity']}
                   for r in live['progress']['routes']],
    },
    'persistedSupportMatrix': [{
        'combo': r['combo'], 'provider': r['provider'], 'api': r['api'],
        'model': r['model'], 'allPassedAt': r.get('allPassedAt'),
        'outcomes': dict(collections.Counter(c['outcome'] for c in r['capabilities'])),
    } for r in matrix['rows']],
    'code': dict(code), 'testkit': dict(testkit), 'filesOver500Lines': oversized,
    'evidence': {
        'record': str(record_path), 'receipt': str(receipt_path),
        'manifest': str(manifest_path), 'audit': str(audit_path),
        'testedCommit': receipt['source']['commit'],
        'argv': receipt['argv'], 'exitCode': receipt['exitCode'],
        'elapsedSeconds': receipt['elapsedSeconds'],
        'receiptHashMatches': digest(receipt_path) == record['e2e']['receipt']['sha256'],
        'manifestHashMatches': digest(manifest_path) == primary['manifestSHA256'],
        'auditHashMatches': digest(audit_path) == primary['auditSHA256'],
        'sourceFiles': len(expected_source), 'inputs': len(record['selectedInputs']),
        'sourceChanged': changed, 'sourceAdded': added, 'sourceRemoved': removed,
        'inputsChanged': changed_inputs,
    },
    'probe': {'kind': 'public-BuiltinCatalog-and-Hash', 'exitCode': result.returncode,
              'goEnvironment': 'GOENV=off; GOFLAGS=; isolated HOME; allowlisted environment'},
}
(OUT / 'inventory.json').write_text(json.dumps(summary, indent=2, ensure_ascii=False) + '\n')
print(json.dumps({k: v for k, v in summary.items() if k not in ['caseFamilies', 'testkit']},
                 ensure_ascii=False, indent=2))
