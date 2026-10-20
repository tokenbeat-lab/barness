"""Keyless portable byte-integrity verification; never executes captured commands."""
from pathlib import Path
import argparse
import base64
import hashlib
import json
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('index', type=Path)
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
require(set(files) == set(index['files']), 'complete member inventory')
for name, pin in index['files'].items():
    require(sha(files[name]) == pin['sha256'] and len(files[name]) == pin['bytes'], name)
source = json.loads(files['collection/source-at-start.json'])
require(source['sourceCommit'] == index['sourceCommit'] and
        len(source['files']) == index['generationSourceFileCount'], 'generation source identity')
for name, pin in source['files'].items():
    require(sha(files['source/' + source['sourceCommit'] + '/' + name]) == pin, 'source ' + name)
for capture in index['candidates'] + index['historicalResearch']:
    prefix = capture['archiveDirectory'] + '/'
    own = {name[len(prefix):]: raw for name, raw in files.items() if name.startswith(prefix)}
    seal = json.loads(own['minimax-integrity.json'])
    require(set(own) == set(seal['files']) | {'minimax-integrity.json'}, 'complete capture ' + prefix)
    for name, pin in seal['files'].items():
        require(sha(own[name]) == pin, 'sealed ' + prefix + name)
    report = json.loads(own['minimax-report.json'])
    require(sha(own['minimax-report.json']) == capture['reportSha256'] and
            sha(own['minimax-integrity.json']) == capture['sealSha256'], 'capture roots ' + prefix)
    require(report['gitCommit'] == capture['sourceCommit'] and report['outcome'] == capture['outcome']
            and report['complete'] == capture['complete'], 'capture scope ' + prefix)
    source_prefix = ('source/' if capture in index['candidates'] else 'historical-source/')
    for name, pin in report['executionSource'].items():
        require(sha(files[source_prefix + report['gitCommit'] + '/' + name]) == pin,
                'execution source ' + prefix + name)
    official = json.loads(own['minimax-sources.json'])
    official_manifest = base64.b64decode(official['sources.json'], validate=True)
    require(sha(official_manifest) == report['sourceHash'], 'official manifest ' + prefix)
    for pin in json.loads(official_manifest):
        require(sha(base64.b64decode(official[pin['file']], validate=True)) == pin['sha256'],
                'official source ' + prefix + pin['file'])
receipts = [json.loads(raw) for raw in files['accounting/candidate.jsonl'].splitlines()]
previous = ''
for sequence, receipt in enumerate(receipts, 1):
    empty_hash = dict(receipt)
    empty_hash['hash'] = ''
    raw = json.dumps(empty_hash, ensure_ascii=False, separators=(',', ':')).encode()
    require(receipt['sequence'] == sequence and receipt['previous'] == previous and
            sha(raw) == receipt['hash'], 'ledger sequence ' + str(sequence))
    previous = receipt['hash']
require(sum(x['kind'] == 'reserve' for x in receipts) == index['stageBudgetSnapshot']['calls'] and
        sum(x['kind'] == 'attempt' for x in receipts) == index['stageBudgetSnapshot']['attempts'],
        'inclusive accounting totals')
print(json.dumps({'status': 'PASS', 'checks': checks, 'archiveFiles': len(files),
                  'sourceCommit': index['sourceCommit'], 'networkUsed': False,
                  'credentialsRead': False, 'capturedCommandsExecuted': False}))
