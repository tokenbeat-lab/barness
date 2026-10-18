"""Retain closed L paid evidence under /tmp. Never execute a captured command."""
from pathlib import Path
from collections import Counter
import argparse
import datetime
import json
from retention_core import (sha, digest, stable, scan, pack_bytes, pack_tree, verify_tar,
                            main_manifests, retained_main_manifests, retained_references, committed_sources)

SOURCE = '3d3a1fb28e143a2c29d684f78edcad94c0217243'
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('config', type=Path)
p.add_argument('--parent-ready', required=True, action='store_true')
a = p.parse_args()
config_raw = stable(a.config, 2 << 20)
c = json.loads(config_raw)
repo = Path(c['repo'])
l = repo / '.evidence/minimax-global-legacy-final'
out = Path(c['out']).resolve()
if not out.is_relative_to(Path('/tmp').resolve()) or c['sourceCommit'] != SOURCE:
    raise ValueError('only /tmp outputs and original3d3 source are allowed')
out.mkdir(parents=True, exist_ok=True)
items = {}

def add(name, raw):
    if name in items:
        raise ValueError('duplicate capsule selection')
    items[name] = raw


def load(name):
    return json.loads(stable(l / name, 2 << 20))


review = load('prerelease-source-review.json')
plan = load('plan.json')
checks = load('checks.json')
if review['decision'] != 'APPROVED' or review['sourceCommit'] != SOURCE:
    raise ValueError('original approved source review required')
for key, file in [('sourceManifestSHA256', 'source-at-start.json'), ('planSHA256', 'plan.json'),
                  ('binariesManifestSHA256', 'binaries.json'), ('checksSHA256', 'checks.json')]:
    if review[key] != sha(stable(l / file, 2 << 20)):
        raise ValueError('original approved fixed metadata binding differs')
if not checks['passed'] or checks['commands'] != {'vet': 0, 'vet-live': 0, 'no-key-live': 0}:
    raise ValueError('original passed preflight commands required')
routes = review['authorizedRoutes']
if routes != c['routes'] or len(routes) != 15 or plan['automaticProcessReruns']:
    raise ValueError('original fixed15/no-rerun contract differs')
canonical = []
for route in routes:
    raw = stable(repo / '.evidence/minimax-launches' / (route + '.json'), 2 << 20)
    receipt = json.loads(raw)
    if (receipt['state'] != 'COMPLETED' or receipt['sourceCommit'] != SOURCE
            or receipt['launcherAudit'] != {'passed': True, 'findings': []}
            or raw != stable(l / (route + '-intent.json'), 2 << 20)):
        raise ValueError('closed canonical receipt/audit/local identity differs')
    add('canonical/' + route + '.json', raw)
    canonical.append({'route': route, 'sha256': sha(raw), 'budget': receipt['budget'],
                      'exitCode': receipt['exitCode'], 'reportedFileSha256': receipt['reportedFileSha256']})
source_raw = stable(l / 'source-at-start.json', 2 << 20)
source = json.loads(source_raw)
if source['sourceCommit'] != SOURCE or len(source['files']) != 1034:
    raise ValueError('exact1034 selected original source required')
items.update(committed_sources(repo, SOURCE, source))
add('source-at-start.json', source_raw)
excluded = []
binaries = []
for file, pin in load('binaries.json').items():
    raw = stable(l / file)
    if sha(raw) != pin:
        raise ValueError('prepared binary differs')
    binaries.append({'file': file, 'sha256': pin, 'bytes': len(raw), 'included': False,
                     'original': '$REPO/.evidence/minimax-global-legacy-final/' + file})
    excluded.append(l / file)
for file, expected in load('prepared.json')['authorizationFiles'].items():
    if sha(stable(l / file, 2 << 20)) != expected:
        raise ValueError('original fixed authorization differs')
metadata_paths = {}
for label, folder in [('L', l), ('L-parent', Path(str(l) + '-parent'))]:
    for file in sorted(folder.iterdir()):
        if not file.is_file() or file in excluded:
            continue
        if file.suffix not in ('.json', '.log', '.md', '.txt', '.py'):
            raise ValueError('unclassified top-level metadata')
        key = 'metadata/' + label + '/' + file.name
        add(key, stable(file, 32 << 20))
        metadata_paths[key] = file
for entry in c['parentAuditFiles']:
    file = Path(entry['path'])
    raw = stable(file, 16 << 20)
    if entry.get('sha256') and sha(raw) != entry['sha256']:
        raise ValueError('exact parent audit pin differs')
    add('parent-audits/' + file.name, raw)
externals = []
for entry in c['externalArchives']:
    file = Path(entry['path'])
    if digest(file) != entry['sha256']:
        raise ValueError('historical immutable external archive differs')
    externals.append({key: entry[key] for key in ['sha256', 'scope']} |
                     {'file': file.name, 'bytes': file.stat().st_size, 'included': False})
origins = []
for entry in c['historicalReceipts']:
    index_raw = stable(Path(entry['index']), 8 << 20)
    index = json.loads(index_raw)
    file = repo / '.evidence/minimax-launches' / (entry['route'] + '.json')
    raw = stable(file, 2 << 20)
    v = json.loads(raw)
    if (index['archiveSha256'] != entry['archiveSha256']
            or index['files'][entry['member']]['sha256'] != sha(raw)
            or v['state'] != 'COMPLETED' or v['launcherAudit'] != {'passed': True, 'findings': []}):
        raise ValueError('original historical canonical boundary differs')
    add('historical-canonical/' + entry['route'] + '.json', raw)
    origins.append({'route': entry['route'], 'sha256': sha(raw), 'sourceCommit': v['sourceCommit'],
                    'status': v['budget']['status'], 'calls': v['budget']['calls'],
                    'attempts': v['budget']['attempts'], 'archive': index['archive'],
                    'archiveSha256': index['archiveSha256'], 'index': Path(entry['index']).name,
                    'member': entry['member'], 'hasOrigin': 'origin' in v, 'hasCapturePins': 'capturePins' in v})
if len(origins) != 16:
    raise ValueError('exact16 prior canonical histories required')
for group in ['originalReceipts', 'correctiveReceipts']:
    for route, pin in plan['carriedGLM'][group].items():
        if sha(items['historical-canonical/' + route + '.json']) != pin['receiptSHA256']:
            raise ValueError('fixed-plan carried GLM original receipt differs')
embedded = []
reports = []
for group in c['embeddedRawGroups']:
    base = l / group['relative']
    if base.resolve() == (l / 'support').resolve() or not base.resolve().is_relative_to(l.resolve()):
        raise ValueError('derived support or foreign root excluded')
    files = scan(base)
    inventory = {name: pin for name, (_, pin) in files.items()}
    members = {}
    for name, (file, pin) in files.items():
        if pin.get('kind'):
            raise ValueError('unexpected link in actual paid/NOT_RUN capsule')
        raw = stable(file)
        if sha(raw) != pin['sha256']:
            raise ValueError('closed actual source changed')
        members[name] = raw
        add('captures/' + group['name'] + '/' + name, raw)
    manifests = retained_main_manifests(members, inventory)
    refs = retained_references(members, inventory)
    if {name: pin for name, (_, pin) in scan(base).items()} != inventory:
        raise ValueError('actual capture tree changed')
    embedded.append({'name': group['name'], 'files': len(inventory),
                     'bytes': sum(pin['bytes'] for pin in inventory.values()),
                     'mainManifests': manifests, 'producerReferences': refs})
    if group['name'] in routes:
        receipt = next(v for v in canonical if v['route'] == group['name'])
        matches = [name for name, pin in inventory.items() if pin['sha256'] == receipt['reportedFileSha256']]
        if len(matches) != 1:
            raise ValueError('exact original canonical primary report required')
        name = matches[0]
        v = json.loads(members[name])
        scenarios = v.get('scenarios') or v.get('samples') or []
        reports.append({'route': group['name'], 'file': 'captures/' + group['name'] + '/' + name,
                        'sha256': sha(members[name]), 'model': v.get('model'), 'endpoint': v.get('endpoint'),
                        'sourceCommit': v.get('gitCommit'), 'outcome': receipt['budget']['status'],
                        'complete': v.get('complete'), 'actualBudget': v.get('budget'),
                        'expected': v.get('expected'), 'recordedScenarios': len(scenarios),
                        'scenarioOutcomes': dict(Counter(s.get('outcome', s.get('status', 'UNSPECIFIED')) for s in scenarios))})
    else:
        not_run = [json.loads(raw) for name, raw in members.items() if name.endswith('/minimax-report.json')]
        if len(not_run) != 1:
            raise ValueError('exact current NOT_RUN report required')
        v = not_run[0]
        if (v['identity']['model'] != 'MiniMax-M3.1-Flash-Preview' or v['outcome'] != 'NOT_RUN'
                or v['gitCommit'] != SOURCE or v['budget']['calls'] != 0 or v['budget']['attempts'] != 0):
            raise ValueError('original current M3.1 NOT_RUN identity/accounting differs')
if len(reports) != 15 or len(embedded) != 18:
    raise ValueError('fixed15paid +3currentNOT_RUN captures required')
raw_summary, raw_inventory = pack_tree(l / 'no-key-live', out / 'legacy-preflight-3d3.tar.gz')
raw_summary['mainManifests'] = main_manifests(l / 'no-key-live', raw_inventory)
raw_summary['intendedCommonRetention'] = '$COMMON/.evidence/minimax-retained/legacy-preflight-3d3.tar.gz'
for name, file in metadata_paths.items():
    if items[name] != stable(file, 32 << 20):
        raise ValueError('selected closed metadata changed')
add('curation-config.json', config_raw)
tools = {}
for name in ['curate_paid.py', 'verify_paid.py', 'retention_core.py', 'FAILURE-INVENTORY.md', 'README.md']:
    raw = stable(Path(__file__).parent / name, 2 << 20)
    add('curation-tools/' + name, raw)
    tools[name] = {'sha256': sha(raw), 'bytes': len(raw)}
archive = out / 'legacy-paid-3d3.tar.gz'
pack_bytes(archive, items)
pins = {name: {'sha256': sha(raw), 'bytes': len(raw)} for name, raw in items.items()}
roundtrip = verify_tar(archive, pins)
if archive.stat().st_size > 8 << 20:
    raise ValueError('small complete actual/source capsule exceeds8MiB target')
consumption = load('consumption.json')
if (sum(v['budget']['calls'] for v in canonical) != 168 or sum(v['budget']['attempts'] for v in canonical) != 168
        or consumption['legacy']['calls'] != 300 or consumption['legacy']['attempts'] != 300):
    raise ValueError('original actual accounting differs')
result = {'schema': 1, 'kind': 'closed-legacy-paid-retention-not-release-certification',
          'createdAt': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'sourceCommit': SOURCE,
          'sourceFiles': 1034, 'archive': archive.name, 'archiveSha256': digest(archive),
          'archiveBytes': archive.stat().st_size, 'files': pins, 'archiveRoundtrip': roundtrip,
          'canonical15': canonical, 'paidReports': reports, 'embeddedRawGroups': embedded,
          'preflightRaw': raw_summary, 'externalArchives': externals, 'historicalOrigins': origins,
          'externalBinaries': binaries, 'toolFiles': tools,
          'paidOutcomeCounts': dict(Counter(v['outcome'] for v in reports)),
          'paidCalls': 168, 'paidAttempts': 168, 'originalGLM': {'calls': 66, 'attempts': 66},
          'correctiveR1': {'calls': 66, 'attempts': 66}, 'cumulativeGLM': {'calls': 132, 'attempts': 132},
          'legacyAccounting': consumption['legacy'], 'miniMaxStages': consumption['minimax'],
          'currentM31': {'routes': 3, 'status': 'NOT_RUN', 'calls': 0, 'attempts': 0},
          'automaticReruns': False, 'additionalRoundAuthorized': False, 'releaseCertifiedByCurator': False,
          'supportDirectoryCollected': False, 'actualKeylessACollected': False,
          'rawOutcomesUpgraded': False, 'rootModified': False, 'credentialsReadByCurator': False,
          'environmentReadByCurator': False, 'testsRunByCurator': False, 'networkUsedByCurator': False,
          'capturedCommandsExecuted': False,
          'scope': 'Exact closed L paid15/source/preflight and original16 canonical histories. PARTIAL actualFAIL; no support/release certification; active A excluded.'}
with (out / 'legacy-paid-3d3.json').open('x') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print(json.dumps({'integrity': 'PASS', 'archiveSha256': result['archiveSha256'],
                  'archiveBytes': result['archiveBytes'], 'entries': len(pins),
                  'paidOutcomes': result['paidOutcomeCounts'], 'releaseCertifiedByCurator': False}))
