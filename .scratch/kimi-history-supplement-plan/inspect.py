"""Read retained Kimi evidence and budget inputs; never execute or authorize live."""
from decimal import Decimal as D
from pathlib import Path
import collections
import hashlib
import json

ROOT = Path(__file__).resolve().parents[2]
BASE = ROOT / '.evidence/kimi-history-supplement-plan'

def load(p):
    return json.loads(Path(p).read_text())

def sha(p):
    p = Path(p)
    assert p.is_file() and not p.is_symlink(), str(p)
    return hashlib.sha256(p.read_bytes()).hexdigest()

def ref(p):
    p = Path(p)
    if not p.is_absolute():
        p = ROOT / p
    assert p.resolve().is_relative_to(ROOT)
    return {'path': str(p.relative_to(ROOT)), 'sha256': sha(p)}

def save(name, value):
    with (BASE / name).open('x') as f:
        f.write(json.dumps(value, indent=2, sort_keys=True) + '\n')

selection = ROOT / '.evidence/provider-delivery-offline/selected-evidence-v1.json'
selected = load(selection)['kimi']
candidate_path = ROOT / 'ai/internal/testkit/kimi/candidate.json'
candidate = load(candidate_path)
prices_path = ROOT / 'ai/internal/testkit/kimi/prices.json'
old_audit = ROOT / '.evidence/provider-delivery-offline/originals-audit-v1.json'
current_ordinary = ROOT / '.evidence/provider-source-scope/e2e/20261027T111447.586011Z-28bkc2h1/receipt.json'
e2e = load(current_ordinary)
assert e2e['status'] == 'PASS' and e2e['exitCode'] == 0
bundle = current_ordinary.parent / next(b['path'] for b in e2e['bundles'] if b['package'] == './ai/e2e')
manifest = load(bundle / 'manifest.json')
controls = {c['id'].removeprefix('GROK-HISTORY-WIRE-'): c for c in manifest['cases'] if c['id'].startswith('GROK-HISTORY-WIRE-kimi')}
old_ids = ['kimi-k3/' + k for k in 'full-stream full-complete simple-stream simple-complete native-history tool-call tool-result cancel recovery'.split()]
supplemental_protection = {}

# Only follow explicit retained repository evidence references, never key names.
def retained_refs(value):
    out = []
    if isinstance(value, dict):
        if isinstance(value.get('path'), str) and isinstance(value.get('sha256'), str):
            p = Path(value['path'])
            if not p.is_absolute():
                p = ROOT / p
            assert p.resolve().is_relative_to(ROOT) and sha(p) == value['sha256'], str(p)
            out.append(ref(p))
            supplemental_protection[str(p.relative_to(ROOT))] = sha(p)
        for v in value.values():
            out.extend(retained_refs(v))
    elif isinstance(value, list):
        for v in value:
            out.extend(retained_refs(v))
    return out

rows = []
for chosen in selected:
    route = chosen['route']
    original = ROOT / chosen['bundle']
    report = load(original / 'regional-report.json')
    m = load(original / 'manifest.json')
    seal = load(original / 'regional-integrity.json')
    assert report['expected'] == old_ids
    assert [s['id'] for s in report['scenarios']] == old_ids
    assert len(m['cases']) == 9 and all(c['status'] == 'PASS' for c in m['cases'])
    assert all(s['outcome'] == 'PASS' and s['calls'] == 1 and s['attempts'] == 1 and s.get('retries', 0) == 0 for s in report['scenarios'])
    assert report['regional']['stage'] == 'candidate' and report['regional']['mode'] == 'live'
    assert report['budget']['maxCalls'] == 9 and report['budget']['maxAttempts'] == 9
    assert report['budget']['maxRequestedOutputTokens'] == 36864 and report['budget']['maxReservedInputTokens'] == 131072
    assert load(original / 'audit.json')['findings'] == []
    for c in m['cases']:
        for n, h in c['artifacts'].items():
            assert sha(original / c['dir'] / n) == h
    for n, h in seal['files'].items():
        assert sha(original / n) == h
    origin_refs = retained_refs(chosen['original'])
    ledger = load(ROOT / chosen['original']['ledgerProof']['path'])
    ledger_refs = retained_refs(ledger)
    native_catalog = load(original / 'regional-catalog.json')
    select_model = lambda c: next(x for x in c['models'] if x['provider'] == report['provider'] and x['api'] == report['api'] and x['id'] == 'kimi-k3')
    old_model, new_model = dict(select_model(native_catalog)), dict(select_model(candidate))
    old_model.pop('name', None); new_model.pop('name', None)
    assert old_model == new_model
    cost = new_model['cost']
    rate = max(D(str(cost.get(k, 0))) for k in ('input', 'cacheRead', 'cacheWrite'))
    rate = max(rate, D(str(cost['input'])) * 2)
    bounded = D(4096) * rate / D(1000000) + D(4096) * D(str(cost['output'])) / D(1000000)
    assert bounded < D('0.10')
    c = controls[route]
    requests_file = bundle / c['dir'] / 'requests.json'
    assert c['status'] == 'PASS' and sha(requests_file) == c['artifacts']['requests.json']
    captured = load(requests_file)
    assert len(captured) == 1
    body = captured[0]['body']
    raw = json.dumps(body, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode()
    preliminary_input = max(2048, len(raw) + 1024)
    assert preliminary_input < 4096
    rows.append({'route': route, 'provider': report['provider'], 'api': report['api'], 'endpoint': report['endpoint'], 'model': 'kimi-k3', 'accountAlias': report['accountAlias'], 'regionAlias': report['regional']['regionAlias'], 'sdk': report['sdk'], 'baseline': {'bundle': chosen['bundle'], 'manifest': ref(original / 'manifest.json'), 'report': ref(original / 'regional-report.json'), 'seal': ref(original / 'regional-integrity.json'), 'originalCommit': report['gitCommit'], 'originalFinishedAt': report['finishedAt'], 'originalStage': 'candidate', 'originalMode': 'live', 'expected': old_ids, 'passCount': 9, 'newlyExecutedCalls': 0, 'originReferences': origin_refs, 'trustedLedgerReferences': ledger_refs}, 'currentUsedModelProjectionEqual': True, 'missing': ['kimi-k3/xai-origin-history'], 'existingControlledWire': {'requests': ref(requests_file), 'inputReservation': preliminary_input, 'note': 'Existing ordinary StreamSimple/64 fixture, not the proposed full/4096 producer. Exact new payload must fit before issuing any paid grant.'}, 'proposedNewBudget': {'maxCalls': 1, 'maxAttempts': 1, 'maxRetries': 0, 'maxRequestedOutputTokens': 4096, 'maxOutputTokensPerCall': 4096, 'maxReservedInputTokens': 4096, 'callSeconds': 120, 'processSeconds': 240, 'callIntervalMs': 60000, 'maxUSD': '0.10', 'conservativeReservationAtTokenCapsUSD': str(bounded)}, 'proposalStatus': 'UNAUTHORIZED_NOT_IMPLEMENTED_NOT_EXECUTED'})

assert len(rows) == 6 and sum(r['baseline']['passCount'] for r in rows) == 54
assert load(ROOT / 'ai/catalog_kimi.json')['active'] is False
save('route-evidence-and-budget-v1.json', {'kind': 'design-only-proposal', 'authorizationStatus': 'NOT_GRANTED', 'implementationStatus': 'NOT_IMPLEMENTED', 'selectedInput': ref(selection), 'priorOriginalsAudit': ref(old_audit), 'ordinaryControlledEvidence': ref(current_ordinary), 'candidate': ref(candidate_path), 'prices': ref(prices_path), 'priceVersion': load(prices_path)['version'], 'fxNote': 'Retained 2026-10-16 conversion 1 USD = 6.7220 CNY; not a current FX quote.', 'routes': rows, 'oldActualCases': 54, 'minimumNewCases': 6, 'newCallCapIfSeparatelyAuthorized': 6, 'newUSDCapIfSeparatelyAuthorized': '0.60', 'conservativeReservationAtAllTokenCapsUSD': str(sum(D(r['proposedNewBudget']['conservativeReservationAtTokenCapsUSD']) for r in rows)), 'kimiCatalogCurrentlyActive': False, 'vendorCallsThisTurn': 0, 'newGrantsThisTurn': 0, 'claimsThisTurn': 0, 'raceExecuted': False, 'pressureExecuted': False, 'resumedPausedRounds': 0, 'currentQualification': 'PARTIAL_9_OF_10'})
save('supplemental-protection-before.json', supplemental_protection)
print('six original bundles: 54 actual PASS, captured bytes/seals/references intact; used model projections equal')
print('proposal token-capped reservation USD', sum(D(r['proposedNewBudget']['conservativeReservationAtTokenCapsUSD']) for r in rows), 'approval cap USD 0.60 NOT_GRANTED')
print('no producer execution, credentials, grants, claims, race or pressure')
