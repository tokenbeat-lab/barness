"""Independent final evidence sealing; requires the actual complete gate PASS."""
import collections
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import stat
import subprocess
import sys
import tarfile

WORK = Path('/tmp/barness-glm-13')
ROOT = Path('/Users/cyber/RestoX/barness')
SOURCE = '3d3cd0fac02a9470a382e825a5e2e114d9e45ee7'
OUT = ROOT / '.evidence/barness-ai-glm/ticket-13'
COLLECTION = Path('.evidence/glm13-corrective-current-r1')


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        while chunk := stream.read(1 << 20):
            h.update(chunk)
    return h.hexdigest()


def doc(path):
    return json.loads(path.read_bytes())


def dump(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def inventory(path):
    # The no-key adversarial harness deliberately creates one symlink fixture.
    # Preserve its lstat/readlink metadata, never its external target bytes.
    result = {}
    for parent, dirs, files in os.walk(path, followlinks=False):
        for name in dirs + files:
            p = Path(parent) / name
            info = p.lstat()
            rel = str(p.relative_to(path))
            if stat.S_ISLNK(info.st_mode):
                assert rel.split('/')[0] in ('live-harness', 'live-harness-confirmation-r1', 'diagnose-jev-isolated') and rel.endswith('/formal-corrupt-symlink/formal-integrity.json'), rel
                result[rel] = dict(kind='symlink-negative-fixture', target=os.readlink(p), mode=stat.S_IMODE(info.st_mode))
            elif stat.S_ISREG(info.st_mode):
                result[rel] = dict(kind='regular', sha256=digest(p), bytes=info.st_size, mode=stat.S_IMODE(info.st_mode))
            else:
                assert stat.S_ISDIR(info.st_mode), rel
    return result


def canonical(path):
    p = Path(path)
    if not p.is_absolute():
        p = WORK / p
    # The recorded worker paths remain immutable provenance; only this new
    # index supplies the durable counterpart, without rewriting old reports.
    return str(ROOT / p.resolve().relative_to(WORK.resolve()))


def verify_reused_inputs(paths):
    # Only the ten reused own-live bundles and two text-effect bundles are
    # rescanned. The full failed historical E2E tree remains archival proof.
    base = ROOT / '.evidence/glm13-legacy-current-r3'
    pinned = OUT / 'historical-r3-files.json'
    historical = doc(pinned)
    roots = sorted({Path(p) for p in paths if Path(p).is_relative_to(base)})
    assert len(roots) == 12, roots
    verified = {}
    for root in roots:
        prefix = str(root.relative_to(base)) + '/'
        expected = {name.removeprefix(prefix): value for name, value in historical.items()
                    if name.startswith(prefix)}
        assert expected and inventory(root) == expected, root
        verified[prefix] = dict(path=str(root), files=len(expected), exactSetHashMode=True)
    return dict(inventory='historical-r3-files.json', inventorySHA256=digest(pinned),
                completeReusedSubtrees=12, bundles=verified)


if sys.argv[1:] == ['--verify']:
    tracked_index = ROOT / '.scratch/barness-ai-glm/evidence/ticket-13/index.json'
    if not tracked_index.is_file():
        tracked_index = WORK / '.scratch/barness-ai-glm/evidence/ticket-13/index.json'
    index = doc(tracked_index)
    assert index['passVerdict'] and index['sourceCommit'] == SOURCE
    assert digest(OUT / index['actualFullRelease']['seal']) == index['actualFullRelease']['sealSHA256']
    assert digest(OUT / index['durableInventory']['path']) == index['durableInventory']['sha256']
    owned_payloads = doc(OUT / index['durableInventory']['path'])
    assert len(owned_payloads) == index['durableInventory']['files']
    assert all(digest(OUT / name) == sha for name, sha in owned_payloads.items())
    for name, sha in index['proofIndexes'].items():
        assert digest(OUT / name) == sha, name
    seal = doc(OUT / 'current-release-seal.json')
    assert seal['passVerdict'] and seal['sourceCommit'] == SOURCE
    for name, tree in seal['trees'].items():
        expected = doc(Path(tree['inventory']))
        assert digest(Path(tree['inventory'])) == tree['inventorySHA256']
        assert len(expected) == tree['files']
        assert inventory(Path(tree['path'])) == expected, name
        assert digest(Path(tree['archive'])) == tree['archiveSHA256'], name
        print(name, len(expected), 'canonical files and archive verified', flush=True)
    for name, sha in seal['payloadPins'].items():
        assert digest(OUT / name) == sha, name
    for row in seal['live']:
        root = Path(row['path'])
        assert row['status'] == 'PASS'
        assert digest(root / row['report']) == row['reportSHA256'], row['combo']
        assert digest(root / 'audit.json') == row['auditSHA256'], row['combo']
    effects = doc(OUT / 'current-effect-summary.json')
    assert effects['independentEffects'] == len(effects['effects']) == 4
    for effect in effects['effects'].values():
        assert digest(Path(effect['report'])) == effect['sha256']
    reused = verify_reused_inputs([row['path'] for row in seal['live']] +
        [str(Path(effect['report']).parent) for effect in effects['effects'].values()])
    assert reused == seal['reusedHistoricalInputs']
    assert digest(OUT / seal['trace']['receipt']) == seal['trace']['sha256']
    trace_receipt = doc(OUT / seal['trace']['receipt'])
    assert len(trace_receipt['items']) == 99 and len(trace_receipt['cases']) == 643
    for mode, bundle in trace_receipt['bundles'].items():
        assert digest(Path(bundle['manifest'])) == bundle['sha256']
        manifest = doc(Path(bundle['manifest']))
        assert all(c['status'] == 'PASS' for c in manifest['cases'])
    for case in trace_receipt['cases'].values():
        for record in case.values():
            assert record['status'] == 'PASS'
            assert all(digest(Path(record['path']) / p) == sha for p, sha in record['artifacts'].items())
    print('Complete canonical current release verification PASS', flush=True)
    raise SystemExit(0)
assert not sys.argv[1:], 'usage: glm13-final-seal.py [--verify]'


gate = WORK / COLLECTION / 'gate'
summary = doc(WORK / COLLECTION / 'gate-summary.json')
report = doc(gate / 'release-report.json')
assert summary == {'exitCode': 0, 'passVerdict': True, 'reportAvailable': True}, summary
assert report['pass'] and report['commit'] == SOURCE
assert len(report['gates']) == 10 and all(g['pass'] for g in report['gates'])
assert all(c['passed'] for c in report['commands'])
assert all(s['current'] for s in report['snapshots'])
assert len(report['live']['rows']) == 14
assert all(row['status'] == 'PASS' for row in report['live']['rows'])
assert all(not a.get('error') and not a['findings'] for a in report['audits'])
assert all(e['status'] == 'PASS' for e in report['evidence'])
live_receipts = {}
for audit in report['audits']:
    if not audit.get('combo'):
        continue
    original = Path(audit['bundle'])
    if not original.is_absolute():
        original = WORK / original
    live_receipts[audit['combo']] = dict(path=canonical(original),
        report='live-report.json', reportSHA256=digest(original / 'live-report.json'),
        auditSHA256=digest(original / 'audit.json'))
assert set(live_receipts) == {row['combo'] for row in report['live']['rows']}
trace = doc(WORK / 'ai/release/traceability.json')
items = [t for t in trace['items'] if re.fullmatch(r'GLM-(?:R\d{2}|A\d{2}|US\d{2})', t['id'])]
assert len(items) == 99
expected_items = {f'GLM-{prefix}{n:02d}' for prefix, count in [('R', 22), ('A', 15), ('US', 62)]
    for n in range(1, count + 1)}
assert {t['id'] for t in items} == expected_items
required = {c for t in items for c in t['requiredCases']}
assert len(required) == 643
gate_trace = {t['id']: t for t in report['traceability']}
assert set(gate_trace) == {t['id'] for t in trace['items']}
assert all(gate_trace[t['id']]['status'] == 'PASS' for t in items)
assert all(t['status'] == 'PASS' for t in report['traceability'])

# Verify every frozen code/config file, including the actual launchers, remains
# the exact paid-reviewed byte set. Docs and the generated matrix are excluded
# by the frozen producer's source contract, not by this sealer.
freeze = doc(WORK / COLLECTION / 'source-at-start.json')
assert freeze['gitCommit'] == SOURCE
assert all(digest(WORK / p) == sha for p, sha in freeze['files'].items())

bundle_records = {}
case_records = {case: {} for case in sorted(required)}
for mode in ('offline', 'race'):
    manifests = list((gate / mode).glob('*/manifest.json'))
    assert len(manifests) == 1, mode
    mpath = manifests[0]
    manifest = doc(mpath)
    assert manifest['git_commit'] == SOURCE
    cases = {c['id']: c for c in manifest['cases']}
    assert len(cases) == len(manifest['cases'])
    assert all(c['status'] == 'PASS' for c in cases.values())
    assert required <= cases.keys()
    bundle_records[mode] = dict(path=canonical(mpath.parent),
        manifest=canonical(mpath), sha256=digest(mpath),
        cases=len(cases), statuses=dict(collections.Counter(c['status'] for c in cases.values())),
        startedAt=manifest['started_at'], finishedAt=manifest['finished_at'])
    for case in sorted(required):
        c = cases[case]
        for name, sha in c['artifacts'].items():
            assert digest(mpath.parent / c['dir'] / name) == sha, (mode, case, name)
        case_records[case][mode] = dict(status=c['status'], test=c['test'], replay=c['replay'],
            path=canonical(mpath.parent / c['dir']), artifacts=c['artifacts'])

consumption = doc(WORK / COLLECTION / 'consumption.json')['total']
assert consumption['calls'] == consumption['attempts'] == 36
assert consumption['maxCalls'] == consumption['maxAttempts'] == 36
assert doc(WORK / COLLECTION / 'live-summary.json')['allCommandsSucceeded']
assert doc(OUT / 'historical-r3-preservation.json')['releasePassed'] is False

# Copy immutable, completed command outputs to the same repository-relative
# location in ROOT. Every payload is compared after copying, then archived.
trees = {}
for name, rel, final in (
    ('glm-source-collection', Path('.evidence/glm13-current-r4'), True),
    ('complete-current-release', COLLECTION, True),
    ('nonfinal-consumer-diagnostic', Path('.evidence/glm13-consumer-diagnostic-r2'), False)):
    src, dst = WORK / rel, ROOT / rel
    before = inventory(src)
    if not dst.exists():
        # APFS clone preserves immutable bytes efficiently; portable fallback
        # keeps links as links and never follows the negative fixture.
        if sys.platform == 'darwin':
            subprocess.run(['/bin/cp', '-cR', str(src), str(dst)], check=True)
        else:
            shutil.copytree(src, dst, copy_function=shutil.copy2, symlinks=True)
    after = inventory(dst)
    assert before == after
    inv = OUT / (name + '-files.json')
    dump(inv, after)
    archive = OUT / (name + '.tar.gz')
    assert not archive.exists(), str(archive)
    with tarfile.open(archive, 'w:gz', compresslevel=1) as tf:
        tf.add(dst, arcname=rel.name)
    trees[name] = dict(path=str(dst), files=len(after), exactCopyVerified=True, finalReleaseEvidence=final,
        inventory=str(inv), inventorySHA256=digest(inv),
        archive=str(archive), archiveSHA256=digest(archive))
    print(name, len(after), 'files sealed', flush=True)

# These postpaid indexes preserve the actual successful operation as well as
# its historical failed transaction; no source or paid launcher is edited.
dump(OUT / 'corrective-paid-index.json', dict(sourceCommit=SOURCE,
    canonicalRoot=str(ROOT / COLLECTION), total=consumption, providerCaps={'typesafe':24, 'openai':12},
    allNineProducerVerifierCommandsPass=True, cumulativeTicket13ActualHTTP=282,
    originalRoundActualHTTP=246, newCorrectiveHTTP=36, automaticReruns=False,
    fullCollection=trees['complete-current-release'],
    matrixResolution=dict(path=canonical(WORK / COLLECTION / 'matrix-merge-resolution.json'),
        sha256=digest(WORK / COLLECTION / 'matrix-merge-resolution.json')),
    originalAll14Merge='FAIL: duplicate timestamps; no matrix write; retained separately',
    actualSuccessfulMerge='unchanged public CLI merges sole newer Jev report; other13 byte-equivalent'))
projection = ROOT / '.evidence/glm13-consumer-diagnostic-r2'
dump(OUT / 'consumer-diagnostic-r2-index.json', dict(finalReleaseEvidence=False,
    sourceCommit=SOURCE, originalBundleSource='dbc6c4e048e8578997839134af2203f3d102ddeb',
    scope='NONFINAL leaf-ownership projections, not corrected actual test execution',
    canonicalRoot=str(projection), tree=trees['nonfinal-consumer-diagnostic'],
    pins={n:digest(projection / n) for n in ('NONFINAL.json', 'process.json', 'process.log',
        'assessment.json', 'offline/manifest.json', 'race/manifest.json', 'report/release-report.json')},
    expectedWholeVerdict=False, publicCLIExit=1, allOtherKnownBlockersChecked=True))
effects = {}
for route in ('jev', 'luna'):
    bundle = next((ROOT / '.evidence/glm13-legacy-current-r3' / ('text-' + route)).glob('*/evaluation-report.json'))
    measured = doc(bundle)
    effects['chinese-' + route] = dict(report=str(bundle), sha256=digest(bundle),
        sourceCommit='dbc6c4e048e8578997839134af2203f3d102ddeb', reused=True,
        applicability='complete exact evaluationSource set unchanged; public current verifier PASS',
        metrics=measured['metrics'], budget=measured['budget'])
for route in ('jev', 'decisions'):
    bundle = ROOT / COLLECTION / ('visual-' + route) / 'report.json'
    measured = doc(bundle)
    effects['visual-' + route] = dict(report=str(bundle), sha256=digest(bundle),
        sourceCommit=SOURCE, reused=False, metrics=measured['metrics'], budget=measured['budget'])
dump(OUT / 'current-effect-summary.json', dict(sourceCommit=SOURCE, independentEffects=4,
    fixedDatasetOnly=True, newQualityThreshold=False, productionAccuracyClaim=False, effects=effects))
reused = verify_reused_inputs([row['path'] for row in live_receipts.values()] +
    [str(Path(effect['report']).parent) for effect in effects.values()])

storymap = doc(WORK / '.scratch/barness-ai-glm/evidence/ticket-13/trace-summary.json')['storiesToRequirements']
requirement_tickets = {
    'R01': [1, 2, 4, 6], 'R02': [1, 2, 6, 9, 13], 'R03': [2, 6, 11, 13],
    'R04': [2, 4, 6, 13], 'R05': [1, 3, 5, 7, 9, 10, 13],
    'R06': [11, 12], 'R07': [2, 3, 4, 5, 6, 7], 'R08': [4, 5],
    'R09': [4, 5, 13], 'R10': [5, 9], 'R11': [2, 3, 9, 13],
    'R12': [3, 9, 13], 'R13': [6, 7, 11], 'R14': [6, 7, 9],
    'R15': [3, 5, 7, 9, 13], 'R16': [5, 7, 9], 'R17': [8, 11, 12],
    'R18': [8, 13], 'R19': [9, 13], 'R20': [10], 'R21': [11, 12, 13],
    'R22': [13],
}


def item_receipt(t):
    key = t['id'].removeprefix('GLM-')
    rs = [key] if key in requirement_tickets else storymap.get(key, sorted(set(re.findall(r'\bR\d{2}\b', t['requirement']))))
    assert rs and all(r in requirement_tickets for r in rs), t['id']
    tickets = []
    for number in sorted({n for r in rs for n in requirement_tickets[r]}):
        paths = list((WORK / '.scratch/barness-ai-glm/issues').glob(f'{number:02d}-*.md'))
        assert len(paths) == 1
        tickets.append(str(paths[0].relative_to(WORK)))
    return dict(id=t['id'], source=t['source'], requirement=t['requirement'],
        requirements=rs, tickets=tickets, requiredCases=t['requiredCases'],
        releaseEvidence=gate_trace[t['id']]['evidence'], status=gate_trace[t['id']]['status'])


trace_receipt = dict(sourceCommit=SOURCE, tracePath='ai/release/traceability.json',
    traceSHA256=digest(WORK / 'ai/release/traceability.json'), requirements=22,
    acceptanceGroups=15, stories=62, itemCount=99, finiteCaseIDs=643,
    allRequiredCasesNormalAndRacePass=True, bundles=bundle_records, ownLiveReceipts=live_receipts,
    fullRollbackReceipt=dict(path=str(OUT / 'rollback-index.json'), sha256=digest(OUT / 'rollback-index.json')),
    items=[item_receipt(t) for t in items], cases=case_records)
dump(OUT / 'complete-trace-receipts.json', trace_receipt)
shutil.copyfile(WORK / 'ai/release/catalog-snapshot.json', OUT / 'catalog-snapshot.json')
shutil.copyfile(WORK / 'ai/live/support-matrix.json', OUT / 'support-matrix.json')
shutil.copyfile(WORK / 'ai/live/testdata/glm-candidate/sources.json', OUT / 'glm-price-source-snapshot.json')
shutil.copyfile(WORK / 'ai/live/testdata/glm-formal/known-usage.json', OUT / 'known-usage-snapshot.json')
shutil.copyfile(gate / 'release-report.json', OUT / 'release-report.json')
shutil.copyfile(gate / 'release-report.md', OUT / 'release-report.md')
dump(OUT / 'current-release-seal.json', dict(passVerdict=True, sourceCommit=SOURCE,
    trees=trees, bundles=bundle_records, commands=report['commands'], gates=report['gates'],
    audits=len(report['audits']), snapshots=report['snapshots'],
    trace=dict(items=len(items), cases=len(required), allPass=True,
        receipt='complete-trace-receipts.json', sha256=digest(OUT / 'complete-trace-receipts.json')),
    live=[dict(combo=r['combo'], status=r['status'], **live_receipts[r['combo']]) for r in report['live']['rows']],
    evidence=[dict(name=e['name'], status=e['status'], path=canonical(e['bundle']), detail=e.get('detail')) for e in report['evidence']],
    payloadPins={name:digest(OUT / name) for name in ['catalog-snapshot.json', 'support-matrix.json',
        'glm-price-source-snapshot.json', 'known-usage-snapshot.json', 'release-report.json', 'release-report.md']},
    paid=dict(glmLogicalCalls=66, glmHTTP=66, legacyLogicalCalls=180, legacyHTTP=180,
        correctiveLogicalCalls=36, correctiveHTTP=36, newActualHTTP=282,
        originalRoundClosed=True, correctiveMaxCalls=36, correctiveMaxAttempts=36,
        cumulativePlannedMaxCalls=410, cumulativePlannedMaxAttempts=434,
        automaticReruns=0, providerFallbacks=0,
        accountRegionPermissionsBilling='UNKNOWN; reported usage and USD estimates are not invoices'),
    originalFailedFullCLI=dict(receipt='historical-r3-preservation.json',
        sha256=digest(OUT / 'historical-r3-preservation.json'), releasePassed=False),
    sourceApplicability=dict(path=canonical(WORK / COLLECTION / 'source-applicability.json'),
        sha256=digest(WORK / COLLECTION / 'source-applicability.json')),
    reusedHistoricalInputs=reused,
    immutableReportsPreserved=True, originalWorkerPaths='provenance only; canonical paths indexed here'))
shutil.copyfile(Path(__file__), OUT / Path(__file__).name)
print('Complete actual current release PASS sealed', flush=True)
