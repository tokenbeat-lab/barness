"""Portable byte/reference checks for closed L evidence; never execute raw commands."""
from pathlib import Path, PurePosixPath
from collections import Counter
import argparse
import gzip
import json
import tarfile
from retention_core import (sha, digest, stable, safe_name, verify_tar,
                            retained_main_manifests, retained_references)

SOURCE = '3d3a1fb28e143a2c29d684f78edcad94c0217243'


def basename(base, name):
    if not name or Path(name).name != name or name in ('.', '..'):
        raise ValueError('portable archive/index must be a basename')
    return base / name


def members(archive, names=None):
    result = {}
    total = 0
    with tarfile.open(archive, 'r|gz') as tar:
        for member in tar:
            if names is not None and member.name not in names:
                continue
            if not member.isfile() or member.size > 32 << 20 or member.name in result:
                raise ValueError('unsafe selected regular member')
            total += member.size
            if total > 96 << 20:
                raise ValueError('selected metadata total bound')
            result[safe_name(member.name)] = tar.extractfile(member).read()
    if names is not None and set(result) != names:
        raise ValueError('selected immutable members absent')
    return result


def pin_bounds(pins):
    if not isinstance(pins, dict) or len(pins) > 400000:
        raise ValueError('inventory bound')
    total = 0
    for name, pin in pins.items():
        safe_name(name)
        if type(pin['bytes']) is not int or not 0 <= pin['bytes'] <= 128 << 20:
            raise ValueError('file size bound')
        total += pin['bytes']
    if total > 4 << 30:
        raise ValueError('inventory expanded size bound')


p = argparse.ArgumentParser(description=__doc__)
p.add_argument('index', type=Path)
p.add_argument('--external-map', type=Path)
p.add_argument('--require-external', action='store_true')
a = p.parse_args()
v = json.loads(stable(a.index, 32 << 20))
base = a.index.parent
if (v['kind'] != 'closed-legacy-paid-retention-not-release-certification' or v['sourceCommit'] != SOURCE
        or v['sourceFiles'] != 1034 or v['releaseCertifiedByCurator'] is not False
        or v['actualKeylessACollected'] is not False or v['supportDirectoryCollected'] is not False):
    raise ValueError('original-source/actual-scope claim differs')
archive = basename(base, v['archive'])
if digest(archive) != v['archiveSha256'] or archive.stat().st_size != v['archiveBytes']:
    raise ValueError('small original paid/source archive differs')
pin_bounds(v['files'])
summary = verify_tar(archive, v['files'])
if summary != v['archiveRoundtrip']:
    raise ValueError('archive total differs')
m = members(archive)
source = json.loads(m['source-at-start.json'])
if source['sourceCommit'] != SOURCE or len(source['files']) != 1034:
    raise ValueError('source manifest identity differs')
if {name for name in m if name.startswith('source/')} != {'source/' + name for name in source['files']}:
    raise ValueError('exact selected source byte set differs')
for name, expected in source['files'].items():
    if Path(name).name == '.env' or sha(m['source/' + name]) != expected:
        raise ValueError('selected committed source bytes differ')
for name, pin in v['toolFiles'].items():
    raw = m['curation-tools/' + name]
    if sha(raw) != pin['sha256'] or len(raw) != pin['bytes']:
        raise ValueError('retained curation tool differs')
for name in ['verify_paid.py', 'retention_core.py']:
    if sha(stable(Path(__file__).parent / name, 2 << 20)) != v['toolFiles'][name]['sha256']:
        raise ValueError('executed portable tool differs from retained tool')
review = json.loads(m['metadata/L/prerelease-source-review.json'])
plan = json.loads(m['metadata/L/plan.json'])
if review['decision'] != 'APPROVED' or review['sourceCommit'] != SOURCE:
    raise ValueError('original source approval identity differs')
for key, file in [('sourceManifestSHA256', 'source-at-start.json'), ('planSHA256', 'plan.json'),
                  ('binariesManifestSHA256', 'binaries.json'), ('checksSHA256', 'checks.json')]:
    if review[key] != sha(m['metadata/L/' + file]):
        raise ValueError('source approval original binding differs')
checks = json.loads(m['metadata/L/checks.json'])
if not checks['passed'] or checks['commands'] != {'vet': 0, 'vet-live': 0, 'no-key-live': 0}:
    raise ValueError('original passed no-key preflight differs')
prepared = json.loads(m['metadata/L/prepared.json'])
for file, expected in prepared['authorizationFiles'].items():
    if sha(m['metadata/L/' + file]) != expected:
        raise ValueError('original protocol authorization differs')
if [item['route'] for item in v['canonical15']] != review['authorizedRoutes'] or len(v['canonical15']) != 15:
    raise ValueError('original fixed15 canonical inventory differs')
for item in v['canonical15']:
    raw = m['canonical/' + item['route'] + '.json']
    original = json.loads(raw)
    if (sha(raw) != item['sha256'] or raw != m['metadata/L/' + item['route'] + '-intent.json']
            or original['state'] != 'COMPLETED' or original['sourceCommit'] != SOURCE
            or original['launcherAudit'] != {'passed': True, 'findings': []}):
        raise ValueError('closed canonical identity/audit differs')
    for key in ['budget', 'exitCode', 'reportedFileSha256']:
        if item[key] != original[key]:
            raise ValueError('original canonical accounting/exit differs')
for group in v['embeddedRawGroups']:
    prefix = 'captures/' + group['name'] + '/'
    selected = {name.removeprefix(prefix): raw for name, raw in m.items() if name.startswith(prefix)}
    inventory = {name: {'bytes': len(raw), 'sha256': sha(raw)} for name, raw in selected.items()}
    if len(inventory) != group['files'] or sum(p['bytes'] for p in inventory.values()) != group['bytes']:
        raise ValueError('complete actual capture inventory differs')
    if (retained_main_manifests(selected, inventory) != group['mainManifests']
            or retained_references(selected, inventory) != group['producerReferences']):
        raise ValueError('actual outer artifact/map reference differs')
    if group['name'].startswith('m31-not-run-'):
        reports = [json.loads(raw) for name, raw in selected.items() if name.endswith('/minimax-report.json')]
        if len(reports) != 1:
            raise ValueError('exact current own NOT_RUN report absent')
        r = reports[0]
        if (r['identity']['model'] != 'MiniMax-M3.1-Flash-Preview' or r['outcome'] != 'NOT_RUN'
                or r['gitCommit'] != SOURCE or r['budget']['calls'] != 0 or r['budget']['attempts'] != 0):
            raise ValueError('current M3.1 identity/zero consumption differs')
if len(v['embeddedRawGroups']) != 18 or len(v['paidReports']) != 15:
    raise ValueError('paid15+currentNOT_RUN3 inventory differs')
for item in v['paidReports']:
    raw = m[item['file']]
    report = json.loads(raw)
    canonical = next(c for c in v['canonical15'] if c['route'] == item['route'])
    if sha(raw) != item['sha256'] or item['sha256'] != canonical['reportedFileSha256']:
        raise ValueError('canonical primary report hash differs')
    if item['outcome'] != canonical['budget']['status'] or item['actualBudget'] != report.get('budget'):
        raise ValueError('retained original outcome/usage budget differs')
    for key in ['model', 'endpoint', 'complete', 'expected']:
        if item[key] != report.get(key):
            raise ValueError('recorded primary report summary differs')
    scenarios = report.get('scenarios') or report.get('samples') or []
    if item['recordedScenarios'] != len(scenarios):
        raise ValueError('original fixed inventory count differs')
    outcomes = dict(Counter(s.get('outcome', s.get('status', 'UNSPECIFIED')) for s in scenarios))
    if outcomes != item['scenarioOutcomes']:
        raise ValueError('original scenario outcomes differ')
consumption = json.loads(m['metadata/L/consumption.json'])
if (sum(c['budget']['calls'] for c in v['canonical15']) != 168
        or sum(c['budget']['attempts'] for c in v['canonical15']) != 168
        or consumption['legacy'] != v['legacyAccounting'] or consumption['legacy']['calls'] != 300
        or consumption['legacy']['attempts'] != 300 or consumption['minimax'] != v['miniMaxStages']
        or v['paidOutcomeCounts'] != {'FAIL': 13, 'PASS': 2}):
    raise ValueError('actual168/168/cumulative300/300 original outcome accounting differs')
for item in v['historicalOrigins']:
    raw = m['historical-canonical/' + item['route'] + '.json']
    receipt = json.loads(raw)
    old_index = json.loads(m['parent-audits/' + item['index']])
    if (sha(raw) != item['sha256'] or sha(raw) != old_index['files'][item['member']]['sha256']
            or old_index['archiveSha256'] != item['archiveSha256']
            or receipt['sourceCommit'] != item['sourceCommit']
            or receipt['budget']['status'] != item['status']):
        raise ValueError('original historical canonical boundary differs')
if len(v['historicalOrigins']) != 16:
    raise ValueError('original16 canonical histories absent')
for key in ['originalReceipts', 'correctiveReceipts']:
    for route, pin in plan['carriedGLM'][key].items():
        if sha(m['historical-canonical/' + route + '.json']) != pin['receiptSHA256']:
            raise ValueError('original fixed carried GLM receipt differs')
pre = v['preflightRaw']
pre_archive = basename(base, pre['archive'])
pre_inventory = basename(base, pre['inventory'])
if digest(pre_archive) != pre['sha256'] or pre_archive.stat().st_size != pre['bytes']:
    raise ValueError('complete preflight raw archive differs')
if digest(pre_inventory) != pre['inventorySha256']:
    raise ValueError('complete preflight full inventory differs')
with gzip.open(pre_inventory, 'rb') as f:
    raw = f.read((128 << 20) + 1)
if len(raw) > 128 << 20:
    raise ValueError('decompressed inventory bound')
pins = json.loads(raw)
pin_bounds(pins)
pre_summary = verify_tar(pre_archive, pins)
if any(pre_summary[k] != pre[k] for k in ['entries', 'expandedBytes']):
    raise ValueError('preflight inventory total differs')
selected = {name for name in pins if len(PurePosixPath(name).parts) <= 3
            and (PurePosixPath(name).name in ['manifest.json', 'audit.json']) and not pins[name].get('kind')}
pre_members = members(pre_archive, selected)
if retained_main_manifests(pre_members, pins) != sorted(pre['mainManifests'], key=lambda r: r['manifest']):
    raise ValueError('original outer preflight case outcomes/audit/reference counts differ')
external_map = json.loads(stable(a.external_map, 2 << 20)) if a.external_map else {}
external_results = []
for entry in v['externalArchives']:
    path = external_map.get(entry['file'])
    if not path:
        if a.require_external:
            raise ValueError('original external archive required but unavailable')
        external_results.append({'file': entry['file'], 'status': 'NOT_CHECKED'})
        continue
    if digest(path) != entry['sha256'] or Path(path).stat().st_size != entry['bytes']:
        raise ValueError('original external archive hash/size differs')
    origins = [r for r in v['historicalOrigins'] if r['archive'] == entry['file']]
    recovered = members(path, {r['member'] for r in origins}) if origins else {}
    for origin in origins:
        if recovered[origin['member']] != m['historical-canonical/' + origin['route'] + '.json']:
            raise ValueError('historical canonical differs from original sealed external archive bytes')
    external_results.append({'file': entry['file'], 'status': 'HASH_SIZE_AND_SELECTED_HISTORY_PASS',
                             'historicalCanonicalBytesCompared': len(origins)})
print(json.dumps({'integrity': 'PASS', 'scope': 'archive bytes/references only; actual paid outcome PARTIAL/FAIL',
                  'sourceCommit': SOURCE, 'sourceFiles': 1034, 'smallCapsule': summary,
                  'preflightRaw': pre_summary, 'preflightMainManifests': pre['mainManifests'],
                  'paidOutcomes': v['paidOutcomeCounts'], 'paidCalls': 168, 'paidAttempts': 168,
                  'legacyCalls': 300, 'legacyAttempts': 300, 'historicalCanonicalReceipts': 16,
                  'externalBoundaries': external_results, 'releaseCertifiedByCurator': False,
                  'excludedBinariesVerifiedByPortableTool': False, 'actualKeylessACollected': False,
                  'capturedCommandsExecuted': False, 'environmentRead': False, 'networkUsed': False}, ensure_ascii=False))
