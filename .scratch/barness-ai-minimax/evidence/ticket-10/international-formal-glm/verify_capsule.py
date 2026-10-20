"""Portable keyless byte verification, including truthful failed-forensic scope."""
from pathlib import Path
import argparse
import base64
import hashlib
import json
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('index', type=Path)
parser.add_argument('--candidate-archive', type=Path)
args = parser.parse_args()
index = json.loads(args.index.read_bytes())
archive = args.index.parent / index['archive']
checks = 0


def require(ok, detail):
    global checks
    if not ok:
        raise SystemExit('FAIL: ' + detail)
    checks += 1


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


require(sha(archive.read_bytes()) == index['archiveSha256'], 'archive SHA-256')
require(archive.stat().st_size == index['archiveBytes'], 'archive byte length')
files = {}
with tarfile.open(archive, 'r:gz') as tar:
    for member in tar:
        require(member.isfile() and not member.name.startswith('/') and
                '..' not in Path(member.name).parts, 'safe regular member ' + member.name)
        require(member.name not in files, 'unique member ' + member.name)
        files[member.name] = tar.extractfile(member).read()
require(set(files) == set(index['files']), 'complete archive inventory')
for name, pin in index['files'].items():
    require(sha(files[name]) == pin['sha256'] and len(files[name]) == pin['bytes'], name)
source = json.loads(files['collection/source-at-start.json'])
require(source['sourceCommit'] == index['sourceCommit'] and len(source['files']) == 1026,
        'generation source manifest')
for name, pin in source['files'].items():
    require(sha(files['source/' + source['sourceCommit'] + '/' + name]) == pin, 'source ' + name)
review = json.loads(files['collection/prerelease-source-review.json'])
for key, filename in [('sourceManifestSHA256', 'source-at-start.json'),
                      ('binariesManifestSHA256', 'binaries.json'), ('planSHA256', 'plan.json')]:
    require(review[key] == sha(files['collection/' + filename]), 'approval ' + filename)
for capture in index['reports']:
    prefix = capture['archiveDirectory'] + '/'
    own = {name[len(prefix):]: raw for name, raw in files.items() if name.startswith(prefix)}
    report = json.loads(own[capture['reportFile']])
    intent = json.loads(files['launches/' + capture['job'] + '.json'])
    require(intent['state'] == 'COMPLETED' and intent['launcherAudit'] == {'passed': True, 'findings': []},
            'original completed audited launch ' + prefix)
    require(sha(own[capture['reportFile']]) == capture['reportSha256'] == intent['reportedFileSha256'],
            'original report hash ' + prefix)
    require(report['gitCommit'] == capture['sourceCommit'] == index['sourceCommit']
            and report['outcome'] == capture['outcome'] and report['complete'] == capture['complete'],
            'original outcome and source scope ' + prefix)
    if capture['sealFile']:
        seal = json.loads(own[capture['sealFile']])
        require(set(own) == set(seal['files']) | {capture['sealFile'], 'live-report.json'},
                'complete sealed capture ' + prefix)
        for name, pin in seal['files'].items():
            require(sha(own[name]) == pin, 'sealed ' + prefix + name)
        require(sha(own[capture['sealFile']]) == capture['sealSha256'], 'seal root ' + prefix)
        require(sha(own['live-report.json']) == capture['publishedSupportReportSha256'],
                'published support view ' + prefix)
        require(capture['outcome'] == 'PASS' and capture['complete'] is True,
                'successful complete qualification ' + prefix)
        if capture['family'] == 'minimax':
            pins = intent['origin']
            require(pins['reportSha256'] == capture['reportSha256'] and pins['sealSha256'] == capture['sealSha256']
                    and pins['ledgerSha256'] == sha(own['minimax-ledger.json']), 'MiniMax canonical origin ' + prefix)
            published = json.loads(own['live-report.json'])
            require(published['minimax']['reportSha256'] == capture['reportSha256'] and
                    published['minimax']['sealSha256'] == capture['sealSha256'], 'MiniMax published anchors ' + prefix)
        else:
            pins = intent['capturePins']
            require(pins['formalReportSha256'] == capture['reportSha256'] and pins['integritySha256'] == capture['sealSha256']
                    and pins['supportReportSha256'] == capture['publishedSupportReportSha256']
                    and pins['executionSourceSha256'] == sha(own['glm-execution-source.json']), 'GLM canonical capture pins ' + prefix)
    else:
        require(capture['family'] == 'glm-forensic' and capture['forensicSnapshotOnly'] is True
                and capture['supportEligible'] is False and capture['outcome'] == 'FAIL'
                and capture['complete'] is False and capture['publishedSupportReportSha256'] is None
                and 'glm-integrity.json' not in own and 'live-report.json' not in own
                and 'capturePins' not in intent, 'failed unsealed forensic scope ' + prefix)
    manifest = json.loads(own['manifest.json'])
    for case in manifest['cases']:
        for kind in ('artifacts', 'fixtures'):
            for name, pin in case.get(kind, {}).items():
                require(sha(own[case['dir'] + '/' + name]) == pin, 'manifest ' + prefix + case['dir'] + '/' + name)
    execution = report['executionSource'] if capture['family'] == 'minimax' else json.loads(own['glm-execution-source.json'])
    for name, pin in execution.items():
        require(sha(files['source/' + index['sourceCommit'] + '/' + name]) == pin, 'execution source ' + prefix + name)
    if capture['family'] == 'minimax':
        official = json.loads(own['minimax-sources.json'])
        manifest_raw = base64.b64decode(official['sources.json'], validate=True)
        require(sha(manifest_raw) == report['sourceHash'], 'official MiniMax source manifest ' + prefix)
        for pin in json.loads(manifest_raw):
            require(sha(base64.b64decode(official[pin['file']], validate=True)) == pin['sha256'],
                    'official MiniMax source ' + prefix + pin['file'])
    else:
        manifest_raw = own['GLM11-sources/sources.json']
        require(sha(manifest_raw) == report['sourceManifestHash'], 'official GLM source manifest ' + prefix)
        for pin in json.loads(manifest_raw)['sources']:
            require(sha(own['GLM11-sources/' + pin['file']]) == pin['sha256'],
                    'official GLM source ' + prefix + pin['file'])
formal = [json.loads(raw) for raw in files['accounting/formal.jsonl'].splitlines()]
previous = ''
for sequence, receipt in enumerate(formal, 1):
    blank = dict(receipt)
    blank['hash'] = ''
    raw = json.dumps(blank, ensure_ascii=False, separators=(',', ':')).encode()
    require(receipt['sequence'] == sequence and receipt['previous'] == previous and sha(raw) == receipt['hash'],
            'formal linked receipt ' + str(sequence))
    previous = receipt['hash']
require(len(formal) == 75 and sum(x['kind'] == 'reserve' for x in formal) == 39 and
        sum(x['kind'] == 'attempt' for x in formal) == 36, 'formal stage accounting')
require(sha(files['accounting/candidate.jsonl']) == index['candidateLineage']['stageBudget']['ledgerSha256'],
        'unchanged historical inclusive candidate stage')
candidate_verified = False
if args.candidate_archive:
    require(sha(args.candidate_archive.read_bytes()) == index['candidateLineage']['archiveSha256'],
            'external retained candidate archive boundary')
    candidate_verified = True
print(json.dumps({'integrity': 'PASS', 'actualQualificationOutcome': index['actualQualificationOutcome'],
                  'checks': checks, 'archiveFiles': len(files), 'sourceCommit': index['sourceCommit'],
                  'candidateArchiveBoundaryVerified': candidate_verified, 'networkUsed': False,
                  'credentialsRead': False, 'capturedCommandsExecuted': False}))
