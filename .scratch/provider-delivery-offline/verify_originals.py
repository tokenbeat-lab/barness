"""Read-only custody checks for this finite audit; no execution or qualification authority."""
import collections
import datetime
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / '.evidence/provider-delivery-offline'
SELECTED = json.loads((OUT / 'selected-evidence-v1.json').read_text())


def sha(path):
    path = Path(path)
    assert path.is_file() and not path.is_symlink(), str(path)
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load(path):
    return json.loads(Path(path).read_text())


def stamp(value):
    return datetime.datetime.fromisoformat(value.replace('Z', '+00:00'))


def refs(value):
    found = []
    if isinstance(value, dict):
        if isinstance(value.get('path'), str) and isinstance(value.get('sha256'), str):
            path = Path(value['path'])
            path = path if path.is_absolute() else ROOT / path
            assert sha(path) == value['sha256'], str(path)
            found.append({'path': str(path.relative_to(ROOT)), 'sha256': value['sha256']})
        for v in value.values():
            found.extend(refs(v))
    elif isinstance(value, list):
        for v in value:
            found.extend(refs(v))
    return found


def manifest(base):
    d = load(base / 'manifest.json')
    for c in d.get('cases', []):
        for p, h in c['artifacts'].items():
            assert sha(base / c['dir'] / p) == h.removeprefix('sha256:')
    a = load(base / 'audit.json')
    assert a['findings'] == []
    return d


located = {}
missing = []
checks = []
common = Path(subprocess.check_output(['git', '-C', str(ROOT), 'rev-parse', '--path-format=absolute', '--git-common-dir'], text=True).strip())
for row in SELECTED['formal'] + SELECTED['kimi']:
    base = ROOT / row['bundle']
    is_kimi = row in SELECTED['kimi']
    report = load(base / ('regional-report.json' if (base / 'regional-report.json').exists() else 'live-report.json'))
    m = manifest(base)
    item = {'route': row['route'], 'bundle': row['bundle'], 'manifestSHA256': sha(base / 'manifest.json'), 'scenarioCount': len(report['scenarios']), 'scenarioOutcomes': dict(collections.Counter(x['outcome'] for x in report['scenarios'])), 'auditFindings': 0}
    for seal in ['regional-integrity.json', 'glm-integrity.json', 'formal-integrity.json', 'decisions-integrity.json']:
        if (base / seal).is_file():
            d = load(base / seal)
            if isinstance(d.get('files'), dict):
                for p, h in d['files'].items():
                    assert sha(base / p) == h.removeprefix('sha256:'), (row['route'], p)
                item['sealedFiles'] = len(d['files'])
    source = None
    for name in ['regional-execution-source.json', 'glm-execution-source.json']:
        if (base / name).is_file():
            source = load(base / name)
            item['sourceMap'] = {'path': str((base / name).relative_to(ROOT)), 'sha256': sha(base / name), 'entries': len(source)}
    if row['route'] == 'typesafe-classifier-jev-1.14.0':
        p = load(base / 'protocol/multimodal-report.json')
        source = {str(Path('ai/live', name)): h for name, h in p['sourceHashes'].items()}
        import os
        source = {os.path.normpath(k): v for k, v in source.items()}
    if source:
        commit = report['gitCommit']
        locations = collections.Counter()
        for name, h in source.items():
            key = (name, h)
            if key not in located:
                now = ROOT / name
                if now.is_file() and not now.is_symlink() and sha(now) == h:
                    located[key] = {'kind': 'current-identical', 'path': name}
                else:
                    r = subprocess.run(['git', '-C', str(ROOT), 'cat-file', 'blob', commit + ':' + name], capture_output=True)
                    if r.returncode == 0 and hashlib.sha256(r.stdout).hexdigest() == h:
                        located[key] = {'kind': 'retained-git-blob', 'commit': commit, 'path': name}
                    else:
                        located[key] = {'kind': 'MISSING', 'commit': commit, 'path': name}
                        missing.append({'route': row['route'], 'path': name, 'sha256': h, 'commitChecked': commit})
            locations[located[key]['kind']] += 1
        item['sourceLocations'] = dict(locations)
    if is_kimi:
        reference_checks = refs(row['original'])
        qualification = load(ROOT / row['original']['qualificationCommand']['path'])
        reference_checks.extend(refs(qualification))
        assert qualification['exitCode'] == 0
        old = load(ROOT / qualification['stdout']['path'])
        assert old == report
        producer = load(ROOT / row['original']['receipt']['path'])
        assert producer['exitCode'] == 0 and producer['status'] == 'PASS'
        for p, h in producer['files'].items():
            assert sha((ROOT / row['original']['receipt']['path']).parent / p) == h
        for p, h in producer['selectedInputs'].items():
            now = ROOT / p
            if now.is_file() and sha(now) == h:
                continue
            r = subprocess.run(['git', '-C', str(ROOT), 'cat-file', 'blob', producer['source']['commit'] + ':' + p], capture_output=True)
            assert r.returncode == 0 and hashlib.sha256(r.stdout).hexdigest() == h, ('original input', p)
        g = load(base / 'regional-grant.json')
        native = report['regional']
        assert g['mode'] == native['mode'] == 'live' and g['stage'] == native['stage'] == 'candidate'
        assert g['id'] == native['grantId'] and sha(base / 'regional-grant.json') == native['grantSha256']
        assert g['maxCalls'] == g['maxAttempts'] == len(report['expected']) == len(report['scenarios']) == 9
        assert stamp(report['finishedAt']) >= stamp(report['startedAt'])
        assert stamp(report['startedAt']) < stamp(g['expiresAt']) <= stamp(report['startedAt']) + datetime.timedelta(hours=24)
        canonical = json.dumps(source, sort_keys=True, separators=(',', ':')).encode()
        assert hashlib.sha256(canonical).hexdigest() == g['sourceSha256'] == native['sourceSha256'] == native['buildSourceSha256']
        assert sha(Path(producer['argv'][0])) == native['binarySha256'] == g['binarySha256']
        origin = load(base / 'regional-authorization-receipt.json')
        assert origin['scopeSha256'] and origin['grantId'] == g['id']
        proof = load(ROOT / row['original']['ledgerProof']['path'])
        assert len(proof['records']) == 2
        for record in proof['records']:
            trusted = common / record['trustedGitCommonRelativePath']
            preserved = record['preservedCopy']
            want = preserved['sha256'] if isinstance(preserved, dict) else record['sha256']
            preserved = preserved['path'] if isinstance(preserved, dict) else preserved
            assert sha(trusted) == sha(ROOT / preserved) == want
            assert load(trusted) == origin
        assert m['versions']['regional_stage'] == 'candidate' and m['versions']['regional_mode'] == 'live'
        assert set(x['id'] for x in report['scenarios']) == set(report['expected'])
        assert item['scenarioOutcomes'] == {'PASS': 9}
        item.update(originalQualificationExitCode=0, producerExitCode=0, pinnedReferencesVerified=len(reference_checks), producerSelectedInputsVerified=len(producer['selectedInputs']), originalBinaryHashedNotExecuted=True, trustedLedgerRecordsVerified=2, originalStage='candidate', originalBudgetCalls=9, currentNewScene='kimi-k3/xai-origin-history')
    checks.append(item)

result = {'kind': 'finite-original-custody-and-contract-audit', 'routes': checks, 'sourceBytes': [{'path': name, 'sha256': h, 'retained': location} for (name, h), location in sorted(located.items())], 'missingOriginalSourceBytes': missing, 'vendorCalls': 0, 'claimAcquisitions': 0, 'oldBinariesExecuted': False}
path = OUT / 'originals-audit-v1.json'
assert not path.exists()
path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({'routes': len(checks), 'sourceBytes': len(located), 'missingOriginalSourceBytes': missing, 'historicalKimiScenarios': sum(x['scenarioCount'] for x in checks if x['route'].startswith('kimi-'))}, ensure_ascii=False))
