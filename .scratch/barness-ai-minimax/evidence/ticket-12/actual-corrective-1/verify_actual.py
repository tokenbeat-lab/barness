"""Portable byte checks only; no credentials, captured commands or vendor calls."""
from pathlib import Path, PurePosixPath
from collections import Counter
import argparse
import gzip
import hashlib
import io
import json
import os
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('index', type=Path)
parser.add_argument('--prepay-index', type=Path, required=True)
parser.add_argument('--prepay-archive', type=Path, required=True)
parser.add_argument('--prepay-inventory', type=Path, required=True)
parser.add_argument('--lineage-archive', type=Path, action='append', default=[])
args = parser.parse_args()
checks = 0


def need(ok, detail):
    global checks
    if not ok:
        raise ValueError(detail)
    checks += 1


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for raw in iter(lambda: f.read(1 << 20), b''):
            h.update(raw)
    return h.hexdigest()


index = json.loads(args.index.read_bytes())
archive = args.index.parent / index['archive']
need(digest(archive) == index['archiveSha256'] and archive.stat().st_size == index['archiveBytes'],
     'actual archive identity')
files = {}
total = 0
with tarfile.open(archive, 'r|gz') as tar:
    for member in tar:
        name = member.name
        need(member.isfile() and not name.startswith('/') and '..' not in PurePosixPath(name).parts
             and name not in files and len(files) < 4000 and member.size <= 16 << 20,
             'bounded regular actual member')
        raw = tar.extractfile(member).read()
        total += len(raw)
        need(total <= 128 << 20, 'bounded actual expanded total')
        pin = index['files'][name]
        need(len(raw) == pin['bytes'] and sha(raw) == pin['sha256'], 'actual member exact bytes')
        files[name] = raw
need(set(files) == set(index['files']), 'complete actual archive inventory')
source = json.loads(files['collection/source-at-start.json'])
need(source['sourceCommit'] == index['sourceCommit'] and len(source['files']) == 1034,
     'committed source manifest')
for name, pin in source['files'].items():
    need(sha(files['source/' + index['sourceCommit'] + '/' + name]) == pin, 'committed source byte pin')
human = json.loads(files['collection/human-approval.json'])
review = json.loads(files['collection/prerelease-source-review.json'])
for receipt in (human, review):
    need(receipt['decision'] == 'APPROVED' and receipt['sourceCommit'] == index['sourceCommit'],
         'retained exact approval state')
    for key, name in [('sourceManifestSHA256', 'source-at-start.json'),
                      ('binariesManifestSHA256', 'binaries.json'), ('planSHA256', 'plan.json')]:
        need(receipt[key] == sha(files['collection/' + name]), 'retained approval byte binding')
need(human['checksSHA256'] == sha(files['collection/checks.json']) and
     human['replyReference'].endswith('sha256:' + sha(files['collection/user-approval-reply.json'])),
     'retained human reply and passed-check binding')
for report in index['reports']:
    prefix = report['archiveDirectory'] + '/'
    own = {name[len(prefix):]: raw for name, raw in files.items() if name.startswith(prefix)}
    intent = json.loads(files['launches/' + report['job'] + '.json'])
    r = json.loads(own['formal-glm-report.json'])
    need(intent['state'] == 'COMPLETED' and intent['sourceCommit'] == index['sourceCommit']
         and intent['launcherAudit'] == {'passed': True, 'findings': []}, 'original closed audited launch')
    need(sha(own['formal-glm-report.json']) == report['reportSha256'] == intent['reportedFileSha256']
         and r['outcome'] == report['outcome'] and r['complete'] == report['complete'],
         'original own report and outcome')
    for key, name in [('humanApprovalSha256', 'human-approval.json'),
                      ('sourceManifestSha256', 'source-at-start.json'),
                      ('binariesManifestSha256', 'binaries.json'), ('planSha256', 'plan.json')]:
        need(intent[key] == sha(files['collection/' + name]), 'canonical approved-source binding')
    if report['sealFile']:
        seal = json.loads(own[report['sealFile']])
        expected = set(seal['files']) | {report['sealFile']}
        if 'live-report.json' in own:
            expected.add('live-report.json')
        need(set(own) == expected, 'exact existing producer seal inventory')
        for name, pin in seal['files'].items():
            need(sha(own[name]) == pin, 'existing sealed own file')
        need(sha(own[report['sealFile']]) == report['sealSha256'], 'producer seal hash')
    if report['supportEligible']:
        pins = intent['capturePins']
        need(r['outcome'] == 'PASS' and r['complete'] is True and report['canonicalCapturePinsPresent'],
             'individual successful complete qualification')
        for key, name in [('formalReportSha256', 'formal-glm-report.json'),
                          ('integritySha256', 'glm-integrity.json'),
                          ('supportReportSha256', 'live-report.json'),
                          ('executionSourceSha256', 'glm-execution-source.json')]:
            need(pins[key] == sha(own[name]), 'original successful canonical capture pin')
    else:
        need(r['outcome'] == 'FAIL' and r['complete'] is False and 'capturePins' not in intent
             and report['forensicSnapshotOnly'] is True and report['publishedSupportReportSha256'] is None,
             'failed view and seal never mint qualification')
        if report['publishedViewProduced']:
            need(sha(own['live-report.json']) == report['publishedViewSha256'], 'unchanged failed published view')
        for model in ('glm-5.3', 'glm-5.2'):
            need(json.loads(own['formal-' + model + '-report.json'])['complete'] is False,
                 'original incomplete model-qualified views')
    manifest = json.loads(own['manifest.json'])
    for case in manifest['cases']:
        for kind in ('artifacts', 'fixtures'):
            for name, pin in case.get(kind, {}).items():
                need(sha(own[case['dir'] + '/' + name]) == pin, 'own actual artifact reference')
    execution = json.loads(own['glm-execution-source.json'])
    for name, pin in execution.items():
        need(sha(files['source/' + index['sourceCommit'] + '/' + name]) == pin, 'own execution source byte pin')
    official_raw = own['GLM11-sources/sources.json']
    need(sha(official_raw) == r['sourceManifestHash'], 'own official source manifest')
    for pin in json.loads(official_raw)['sources']:
        need(sha(own['GLM11-sources/' + pin['file']]) == pin['sha256'], 'own fixed official source')
plan = json.loads(files['collection/plan.json'])
for name, pin in plan['originalReceipts'].items():
    raw = files['lineage/original-receipts/' + name + '.json']
    original = json.loads(raw)
    need(sha(raw) == pin['receiptSHA256'] and original['reportedFileSha256'] == pin['reportSHA256']
         and original['budget']['status'] == pin['status'], 'unchanged original accounting receipt')
consumption = json.loads(files['collection/corrective-consumption.json'])
need(consumption['lifetime'] == index['currentRenewalLifetime'] and consumption['lifetime']['calls'] == 132
     and consumption['lifetime']['attempts'] == 132 and index['qualification'] == 'PARTIAL'
     and index['completeThreeRouteRenewal'] is False, 'truthful current lifetime and partial outcome')

# The large keyless bundle is an external exact-byte boundary. Never dereference
# the intentional negative symlink or execute a captured subprocess command.
prepay = json.loads(args.prepay_index.read_bytes())
need(args.prepay_index.read_bytes() == files['lineage/approved-prepay-index.json'],
     'exact external passed prepay index bound by actual capsule')
need(digest(args.prepay_archive) == prepay['archiveSha256'] and
     args.prepay_archive.stat().st_size == prepay['archiveBytes'], 'passed prepay archive identity')
with args.prepay_inventory.open('rb') as stream:
    compressed = stream.read((8 << 20) + 1)
need(len(compressed) <= 8 << 20, 'passed prepay compressed inventory bound')
need(sha(compressed) == prepay['inventorySha256'], 'passed prepay inventory identity')
with gzip.GzipFile(fileobj=io.BytesIO(compressed)) as stream:
    inventory_raw = stream.read((32 << 20) + 1)
need(len(inventory_raw) <= 32 << 20, 'passed prepay inventory expanded bound')
inventory = json.loads(inventory_raw)
need(len(inventory) <= 40000, 'passed prepay entry bound')
seen = set()
selected = {}
expanded = 0
with tarfile.open(args.prepay_archive, 'r|gz') as tar:
    for member in tar:
        name = member.name
        need(name in inventory and name not in seen and not name.startswith('/')
             and '..' not in PurePosixPath(name).parts and len(PurePosixPath(name).parts) <= 24,
             'passed prepay safe unique entry')
        seen.add(name)
        pin = inventory[name]
        if member.issym():
            raw = os.fsencode(member.linkname)
            need(pin.get('kind') == 'intentional-negative-symlink' and len(raw) <= 4096,
                 'passed prepay negative symlink metadata')
        else:
            need(member.isfile() and not pin.get('kind') and member.size <= 64 << 20,
                 'passed prepay bounded regular entry')
            raw = tar.extractfile(member).read()
        expanded += len(raw)
        need(expanded <= 512 << 20 and len(raw) == pin['bytes'] and sha(raw) == pin['sha256'],
             'passed prepay exact entry bytes')
        if name in (prepay['manifest'], str(PurePosixPath(prepay['manifest']).parent / 'audit.json')):
            selected[name] = raw
need(seen == set(inventory), 'passed prepay complete file inventory')
manifest = json.loads(selected[prepay['manifest']])
need(Counter(case['status'] for case in manifest['cases']) == {'PASS': 336, 'NOT_RUN': 62},
     'passed prepay original case outcomes')
base = PurePosixPath(prepay['manifest']).parent
need(json.loads(selected[str(base / 'audit.json')]) == {'findings': []}, 'passed prepay audit zero')
refs = 0
for case in manifest['cases']:
    for name, pin in case.get('artifacts', {}).items():
        need(inventory[str(base / case['dir'] / name)]['sha256'] == pin, 'passed prepay actual artifact reference')
        refs += 1
need(refs == prepay['mainArtifactReferences'], 'passed prepay artifact count')
verified_lineage = []
expected = {entry['archive']: entry for entry in index['externalLineage']}
for path in args.lineage_archive:
    need(path.name in expected and digest(path) == expected[path.name]['sha256'], 'external retained lineage hash')
    verified_lineage.append(path.name)
print(json.dumps({'integrity': 'PASS', 'actualQualification': 'PARTIAL', 'checks': checks,
                  'actualFiles': len(files), 'generationSourceFiles': 1034,
                  'prepayArchiveEntries': len(seen), 'prepayMainArtifactReferences': refs,
                  'currentRenewalCalls': 132, 'currentRenewalAttempts': 132,
                  'failedResponsesSupportEligible': False, 'verifiedExternalLineage': verified_lineage,
                  'credentialsRead': False, 'networkUsed': False, 'testsRun': False,
                  'capturedCommandsExecuted': False}))
