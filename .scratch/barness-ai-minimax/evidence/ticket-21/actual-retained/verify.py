"""Bounded read-only custody of the closed36/36 round; no captured execution."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import sys
import types

sys.dont_write_bytecode = True
BASE = Path(__file__).resolve().parent
SOURCE = '78843eb99905e95c8fa534a75897e0dc8787b1da'
ROUND = 'visual-source-renewal-1'
COLLECTION = '.evidence/minimax-current-visual-renewal-1'
PIN_AUTHORITY = '184e22962fb06f342154a70a514596f372488bc3910eaf01e38caec9a1a0cf42'
HUMAN = 'de47fa2b8002f2de0ae8ca3a890d52614b612b3f160cacf104ab45328dbfaae5'
SOURCE_PIN = '319357cf2cb40a70414bb09fd9016b76bbc663170e23d930dba839a5fe96715d'
AUTH = 'c9d8e12102a6abed069f548b6daa0aa516deb0e2577e8f6a3ebddfe662e451fe'
CONSUMPTION = 'bdff7850315a24256d1f6af76b162126f34c288e37e0e3d5b020b14a973dd87c'
FINAL = '1fc0c79553100c9fec81e8b92513e9d678c79a6c2e7b07d8af63a96d7bc3af31'
CANDIDATE = '5913236998b006a14f3e611e71f584e3c6c6c17859d8bf319d51655dce5fc899'
DEFAULT = '5a144d8e3f1ab398dd54fcff8da039d0335dce720326e6db722b95d14b220754'
LINK_TARGET = '82b51b8fc7c9271800b022cb5281e74a0ee89cfe8abc5b5f0553463278661aee'
ARCHIVE_SHA = 'a95ffa0ca3aba1c816a2b956048037a9578af4602f69756fc1972bf650de537b'
ARCHIVE_BYTES = 86782794
PROGRAM_AUTHORITY = '698f62d6fa943f901d57df454a3bd15381f542d891b512afcfca75f770b5d0c8'
NONAUTHORITATIVE = {
    'barness-ticket21-prepared-independent-proof.json': 'f6626c005e031ac1a6cd6fabccbacd094c0795f3f5f272fe18f932f67e2cc533',
    'barness-ticket21-review-prepared.py': '6aa364fb70bb17ce03266c32a27df7d29c3def194f122324981d564e49534ad1',
    'barness-ticket21-review-prepared-post.py': '9ef60ddf9aac6b9c9f4a96a635b63a3e53ef76e8d150fc4e7995019ddf4bb131'}
ORIGINAL_INVENTORY = dict(file='inventory.json.gz', compressedSHA256='19020d5cab381cd6826d386d314b23f8c03a4f581d37da8a81368d743b45ddda',
    decodedBytes=6869676, decodedSHA256='c2b8101f29abc1e297dca96f2b30ce47813b66427877e8bb091b314f6579ddf0', files=29941)
ORIGINAL_SOURCE = dict(file='source-proof.json.gz', compressedSHA256='9889793735a840e78ff3daffc9f4569eedc993870d55910892722752af15b32d',
    decodedBytes=119914, decodedSHA256=SOURCE_PIN, archiveMember='actual/source-at-start.json')
PROTO = 'protocol114/20261019T091724.020745000Z/'
JOBS = ('protocol114', 'visual-jev', 'visual-decisions')
REPORTS = dict(protocol114=PROTO + 'multimodal-report.json', **{'visual-' + r: 'visual-' + r + '/report.json' for r in ('jev', 'decisions')})
PINS = {'ticket-13/implementation/verify.py': 'b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f',
        'ticket-13/final-gate/capsule.py': '3017cba1d4ac09e0a203c68ac768a4d508b861fa862aa04c9ca4550efebd9328'}


def readers():
    modules = []
    for name, sha in PINS.items():
        path = BASE.parent.parent / name
        with os.fdopen(os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK), 'rb') as stream:
            info = os.fstat(stream.fileno())
            if not stat.S_ISREG(info.st_mode) or info.st_size > 64 << 10: raise ValueError('bounded fixed reader required')
            raw = stream.read((64 << 10) + 1)
        if len(raw) != info.st_size or hashlib.sha256(raw).hexdigest() != sha: raise ValueError('fixed reader differs')
        module = types.ModuleType(name); module.__file__ = str(path)
        exec(compile(raw, str(path), 'exec'), module.__dict__); modules.append(module)
    return modules


def integer(value, expected):
    if type(value) is not int or value != expected: raise ValueError('exact integer measurement differs')


def zeros(value, keys):
    if not isinstance(value, dict) or set(value) != set(keys): raise ValueError('resource keys differ')
    for v in value.values(): integer(v, 0)


def original_inventory(b, full):
    value = b.doc(full.decoded(b, BASE, ORIGINAL_INVENTORY, full.LIMITS['inventoryDecodedBytes']))
    integer(len(value), 29941)
    return value


def program_pins(b):
    raw = b.read(BASE, 'closed-program-authority.json', 1 << 20)
    if b.sha(raw) != PROGRAM_AUTHORITY: raise ValueError('independent program authority differs')
    independent = b.doc(raw)['expectedSHA256']
    if len(independent) != 23 or set(independent) & set(NONAUTHORITATIVE): raise ValueError('independent/captured program scopes differ')
    return independent | NONAUTHORITATIVE


def facts(b, pins, seen):
    def raw(name): return seen[name]
    def doc(name): return b.doc(raw(name))
    def pin(name, sha):
        if pins[name]['sha256'] != sha: raise ValueError('fixed original authority differs')
    authority = doc('metadata/closed-pins.json'); pin('metadata/closed-pins.json', PIN_AUTHORITY)
    programs = program_pins(b)
    if {n.removeprefix('metadata/closed-programs/') for n in pins if n.startswith('metadata/closed-programs/')} != set(programs): raise ValueError('fixed26 captured program inventory differs')
    for name, sha in programs.items(): pin('metadata/closed-programs/' + name, sha)
    integer(len(authority['closedFiles']), 43)
    for relative, sha in authority['closedFiles'].items():
        if not relative.startswith(COLLECTION + '/'): raise ValueError('foreign closure pin')
        pin('actual/' + relative.removeprefix(COLLECTION + '/'), sha)
    pin('actual/source-at-start.json', SOURCE_PIN); source = doc('actual/source-at-start.json')
    if source['sourceCommit'] != SOURCE or len(source['files']) != 1078: raise ValueError('original complete source freeze differs')
    pin('actual/execution-authorization.json', AUTH)
    pin('actual/consumption.json', CONSUMPTION)
    pin('actual/finalization/result.json', FINAL)
    pin('canonical/human-approval.json', HUMAN); human = doc('canonical/human-approval.json')
    if human['decision'] != 'APPROVED' or human['renewalId'] != ROUND or human['sourceCommit'] != SOURCE or human['requestReference'] != 'request_user_input_async/call_cbb7af167e4f4c439e3a73ed647ef1ae/0': raise ValueError('canonical direct human authority differs')
    integer(human['maxCalls'], 36); integer(human['maxAttempts'], 36)
    totals = dict(calls=0, attempts=0); native = {}
    for job in JOBS:
        marker = doc('canonical/' + job + '.json'); collection_marker = 'actual/' + ROUND + '-' + job + '-intent.json'
        if raw('canonical/' + job + '.json') != raw(collection_marker): raise ValueError('canonical and collection receipt differ')
        command_name = 'actual/' + job + '-command.json'; command = doc(command_name)
        report_name = 'actual/' + REPORTS[job]; report = doc(report_name)
        if marker['state'] != 'COMPLETED' or marker['route'] != ROUND + '-' + job or marker['validationRound'] != ROUND or marker['sourceCommit'] != SOURCE or marker['collectionRoot'] != '$REPO/' + COLLECTION or marker['authorizationSha256'] != AUTH: raise ValueError('canonical native authority differs')
        pin(command_name, marker['commandReceiptSha256']); pin(report_name, marker['reportedFileSha256'])
        if command['state'] != 'COMPLETED' or command['credentialsRead'] is not True or marker['credentialsRead'] is not True: raise ValueError('actual paid process provenance differs')
        integer(command['exitCode'], 0); integer(marker['exitCode'], 0)
        if job == 'protocol114':
            if report['mode'] != 'live' or report['route'] != 'jev' or report['outcome'] != 'PASS' or report['complete'] is not True or report['auditPassed'] is not True: raise ValueError('protocol is incomplete')
            if len(report['sourceHashes']) != 878 or len(report['expected']) != 12 or set(report['scenarios']) != set(report['expected']): raise ValueError('complete protocol source/scenario set differs')
            manifest = doc('actual/' + PROTO + 'manifest.json')
            expected = {'MM11-' + n for n in report['expected']}; calls = attempts = 0
            if {v['id'] for v in manifest['cases']} != expected | {'MM11-configuration'}: raise ValueError('native protocol manifest denominator differs')
            for case in manifest['cases']:
                if case['status'] != 'PASS': raise ValueError('native protocol case failed')
                prefix = 'actual/' + PROTO + case['dir'] + '/'
                for relative, sha in case['artifacts'].items(): b.relative(relative); pin(prefix + relative, sha)
                if case['id'] in expected:
                    calls += 1
                    attempts += len(doc(prefix + 'result.json')['result']['metadata']['attempts'])
            zeros(doc('actual/' + PROTO + 'MM11-configuration/resources.json'), ('bodies', 'calls', 'events', 'permits', 'waiters'))
            budget = report['budget']['actual']
        else:
            if report['mode'] != 'live' or report['catalogKind'] != 'formal' or report['route'] != job.removeprefix('visual-') or report['status'] != 'COMPLETE' or report['realQuality'] != 'COMPLETE': raise ValueError('formal visual is incomplete')
            samples = report['samples']
            if len(samples) != 12 or {s['id'] for s in samples} != {'v%02d' % n for n in range(1, 13)}: raise ValueError('visual denominator differs')
            calls = len(samples); attempts = sum(len(s['result']['metadata']['attempts']) for s in samples)
            for sample in samples:
                zeros(sample['resources'], ('bodies', 'calls', 'permits', 'waiters'))
            manifest = doc('actual/' + job + '/manifest.json')
            for relative, sha in manifest['artifacts'].items(): b.relative(relative); pin('actual/' + job + '/' + relative, sha)
            budget = report['budget']
        integer(calls, 12); integer(attempts, 12)
        integer(budget['calls'], calls); integer(budget['attempts'], attempts)
        audit = doc('actual/' + (PROTO if job == 'protocol114' else job + '/') + 'audit.json')
        if audit['findings'] != []: raise ValueError('actual native audit failed')
        totals['calls'] += calls; totals['attempts'] += attempts
        native[job] = dict(calls=calls, attempts=attempts, status='PASS', samples=12)
    current = doc('actual/consumption.json')['current']
    for k in ('calls', 'attempts'): integer(current[k], totals[k])
    checks = doc('actual/checks.json')
    if checks['passed'] is not True or len(checks['commands']) != 6: raise ValueError('six checks incomplete')
    for name, code in checks['commands'].items():
        integer(code, 0); command = doc('actual/' + name + '-command.json'); integer(command['exitCode'], 0)
        if command['state'] != 'COMPLETED' or command['credentialsRead'] is not False: raise ValueError('keyless check provenance differs')
    final = doc('actual/finalization/result.json')
    if final['state'] != 'COMPLETED' or final['credentialsRead'] is not False or final['strictElevenCurrent'] is not True or final['defaultPublished'] is not False or final['releaseCertified'] is not False: raise ValueError('pre-gate finalization scope differs')
    integer(final['currentPaidCalls'], 36); integer(final['currentPaidAttempts'], 36)
    integer(final['cumulativeRenewalCalls'], 873); integer(final['cumulativeRenewalAttempts'], 870)
    integer(len(final['commands']), 12)
    for name, sha in final['commands'].items():
        relative = 'actual/finalization/' + name + '-command.json'; pin(relative, sha); command = doc(relative)
        integer(command['exitCode'], 0)
        if command['state'] != 'COMPLETED' or command['credentialsRead'] is not False: raise ValueError('keyless finalization provenance differs')
    candidate = doc('actual/finalization/matrix-candidate.json'); before = doc('actual/finalization/matrix-before.json')
    pin('actual/finalization/matrix-candidate.json', CANDIDATE); pin('actual/finalization/matrix-before.json', DEFAULT)
    if {k: v for k, v in before.items() if k != 'rows'} != {k: v for k, v in candidate.items() if k != 'rows'} or [v['combo'] for v in before['rows']] != [v['combo'] for v in candidate['rows']] or len(candidate['rows']) != 17: raise ValueError('candidate identity/order differs')
    changed = [a['combo'] for a, z in zip(before['rows'], candidate['rows']) if a != z]
    if changed != ['typesafe-classifier-jev-1.14.0']: raise ValueError('candidate changed other rows')
    support = doc('actual/jev-support/live-report.json')
    if support['combo'] != changed[0] or len(support['scenarios']) != 12: raise ValueError('new own Jev support differs')
    inputs = doc('actual/finalization/input-directories.json')
    if len(inputs['live']) != 17 or len(set(inputs['live'])) != 17 or len(inputs['effects']) != 4 or len(set(inputs['effects'])) != 4: raise ValueError('current17+4 input proof differs')
    link = doc('metadata/intentional-link.json')
    if link['type'] != 'symlink' or link['path'] != 'no-key-live/20261019T085824.163169000Z/formal-corrupt-symlink/formal-integrity.json' or hashlib.sha256(bytes.fromhex(link['targetHex'])).hexdigest() != LINK_TARGET or link['targetSHA256'] != LINK_TARGET or 'actual/' + link['path'] in pins: raise ValueError('exact metadata-only link projection differs')
    failed = doc('controls-1/source-proof-command.json'); integer(failed['exitCode'], 1)
    if failed['credentialsRead'] is not False: raise ValueError('failed source-proof credential provenance differs')
    return dict(native=native, actualCalls=totals['calls'], actualAttempts=totals['attempts'], cumulativeRenewalCalls=873, cumulativeRenewalAttempts=870,
        keylessChecks=6, keylessFinalizationCommands=12, changedCandidateRows=changed, preservedRows=16, defaultPublished=False, releaseCertified=False,
        independentlyHashBoundCapturedData=23, nonauthoritativeCapturedData=sorted(NONAUTHORITATIVE), capturedDataIsNotExecutionAuthority=True)


def main():
    p = argparse.ArgumentParser(description=__doc__); p.add_argument('--common-root', type=Path, required=True); a = p.parse_args()
    b, full = readers(); index = b.doc(b.read(BASE, 'index.json', 1 << 20))
    integer(index.get('schema'), 1)
    if index['sourceCommit'] != SOURCE or index['round'] != ROUND: raise ValueError('fixed custody source/round differs')
    cap = index['capsule']
    if cap['sourceCommit'] != SOURCE or cap['archive']['file'] != '.evidence/minimax-retained/ticket-21-three-renewal/actual.tar.gz' or cap['sourceSnapshot']['archiveMember'] != 'actual/source-at-start.json': raise ValueError('fixed actual archive/source authority differs')
    integer(cap['archive']['bytes'], ARCHIVE_BYTES)
    if cap['archive']['sha256'] != ARCHIVE_SHA or cap['archive']['prefix'] != 'ticket21-actual' or cap['inventory'] != ORIGINAL_INVENTORY or cap['sourceSnapshot'] != ORIGINAL_SOURCE: raise ValueError('original sealed archive/inventory/source binding differs')
    for key in ('decodedBytes', 'files'): integer(cap['inventory'][key], ORIGINAL_INVENTORY[key])
    integer(cap['sourceSnapshot']['decodedBytes'], ORIGINAL_SOURCE['decodedBytes'])
    pins, seen, summary = full.verify(b, BASE, a.common_root, cap)
    result = facts(b, pins, seen)
    print(json.dumps(dict(verified=True, archive=summary, originalRound=result, retentionPaidCalls=0, retentionCredentialsRead=False, capturedProgramsExecuted=False), indent=2))


if __name__ == '__main__': main()
