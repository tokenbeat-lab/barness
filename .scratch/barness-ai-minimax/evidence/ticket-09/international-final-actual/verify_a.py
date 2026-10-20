"""DRAFT portable A integrity verifier. Does not execute any captured command."""
from pathlib import Path, PurePosixPath
from collections import Counter
import argparse
import gzip
import json
from retention_core import sha, digest, stable, verify_tar, retained_main_manifests
from a_core import SOURCE, selected_members, command_log_contract, check_reference_contract, prove_paid_source


def local(base, name):
    if not name or Path(name).name != name or name in ('.', '..'):
        raise ValueError('archive/index must be a basename')
    return base / name


def inventory(path, expected):
    if digest(path) != expected:
        raise ValueError('full inventory hash differs')
    with gzip.open(path, 'rb') as f:
        raw = f.read((128 << 20) + 1)
    if len(raw) > 128 << 20:
        raise ValueError('decompressed full inventory bound')
    values = json.loads(raw)
    if len(values) > 400000 or sum(p['bytes'] for p in values.values()) > 4 << 30:
        raise ValueError('raw inventory per-group bound')
    return values


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('index', type=Path)
    p.add_argument('--external-map', type=Path, required=True,
                   help='Original external archive basenames -> immutable relocated archive files')
    a = p.parse_args()
    v = json.loads(stable(a.index, 16 << 20))
    if (v['kind'] != 'original3d3-actual-failed-gate-byte-retention' or v['sourceCommit'] != SOURCE
            or v['sourceFilesProvedExternally'] != 1034 or v['actualGatePass'] is not None
            or v['originalReleaseReportPresent'] is not False
            or v['actualGateExitCode'] != 1 or v['releaseCertifiedByCurator'] is not False):
        raise ValueError('original source/outcome/scope differs')
    base = a.index.parent
    archive = local(base, v['archive'])
    if digest(archive) != v['archiveSha256'] or archive.stat().st_size != v['archiveBytes']:
        raise ValueError('A metadata/source-proof capsule differs')
    if verify_tar(archive, v['files']) != v['archiveRoundtrip']:
        raise ValueError('metadata archive member totals differ')
    m = selected_members(archive, set(v['files']))
    for name, pin in v['toolFiles'].items():
        raw = m['curation-tools/' + name]
        if sha(raw) != pin['sha256'] or len(raw) != pin['bytes']:
            raise ValueError('retained tool differs')
    for name in ['verify_a.py', 'retention_core.py', 'a_core.py']:
        if sha(stable(Path(__file__).parent / name, 2 << 20)) != v['toolFiles'][name]['sha256']:
            raise ValueError('portable verification tools differ from retained tools')
    c = json.loads(m['curation-config.json'])
    manifest_raw = m['metadata/A/argv-manifest.json']
    manifest = json.loads(manifest_raw)
    if sha(manifest_raw) != c['argvManifestSHA256']:
        raise ValueError('original planned input/argv metadata differs')
    outer = json.loads(m['metadata/A/actual-releasegate-command.json'])
    if 'metadata/gate/release-report.json' in m:
        raise ValueError('a missing original release report was invented')
    log_raw = m['metadata/A/actual-releasegate.log']
    if sha(log_raw) != c['closedLogSHA256']:
        raise ValueError('original closed gate log differs')
    if command_log_contract(log_raw, outer, manifest) != v['originalCommandBannerEvidence']:
        raise ValueError('original command-banner evidence differs')
    check_reference_contract(c, manifest)
    summary = json.loads(m['metadata/A/summary.json'])
    if (summary['passVerdict'] is not False or summary['releasePass'] is not None
            or summary['releaseReportPresent'] is not False or summary['releaseReportSHA256'] is not None
            or summary['gateExitCode'] != 1
            or summary['paidCalls'] != 0 or summary['matrixBeforeSHA256'] != summary['matrixAfterSHA256']):
        raise ValueError('original failed A closure/matrix retention differs')
    for key, file in [('matrixBeforeSHA256', 'matrix-before.json'), ('matrixAfterSHA256', 'matrix-after.json')]:
        if summary[key] != sha(m['metadata/A/' + file]):
            raise ValueError('original unchanged matrix byte proof differs')
    mapping = json.loads(stable(a.external_map, 2 << 20))
    paths = {}
    for entry in v['externalArchives']:
        path = Path(mapping[entry['file']])
        if digest(path) != entry['sha256'] or path.stat().st_size != entry['bytes']:
            raise ValueError('original external archive boundary differs')
        paths[entry['file']] = path
    paid_raw = m['source-proof/original-paid-L-index.json']
    if sha(paid_raw) != c['paidIndexSHA256']:
        raise ValueError('immutable original L source proof index differs')
    paid = json.loads(paid_raw)
    proved = prove_paid_source(paid, paths[c['paidArchive']])
    if proved['source-at-start.json'] != m['source-proof/source-at-start.json']:
        raise ValueError('original selected1034source manifest differs')
    for name in c['publicSourceFiles']:
        if proved['source/' + name] != m['source-proof/public-gate/' + name]:
            raise ValueError('original exact public-gate code bytes differ')
    planned = v['plannedOwnInputs'] + v['effectReferences']
    if len(v['plannedOwnInputs']) != 17 or len(v['effectReferences']) != 4:
        raise ValueError('planned17/effects4 reference inventory differs')
    for row, original in zip(planned, c['ownInputReferences'] + c['effectReferences']):
        if any(row.get(key) != value for key, value in original.items()):
            raise ValueError('original planned reference metadata changed')
        if row.get('expectedPresence') == 'MISSING':
            if row['observedPresence'] != 'MISSING' or row['archive'] is not None or row['sha256'] is not None:
                raise ValueError('missing planned support input was invented')
            continue
        raw = selected_members(paths[row['archive']], {row['member']})[row['member']]
        if sha(raw) != row['sha256']:
            raise ValueError('original external own/effect report member differs')
    suites = []
    failures_seen = []
    for group in v['rawGroups']:
        raw_archive = local(base, group['archive'])
        if digest(raw_archive) != group['sha256'] or raw_archive.stat().st_size != group['bytes']:
            raise ValueError('complete ordinary/race raw boundary differs')
        pins = inventory(local(base, group['inventory']), group['inventorySha256'])
        total = verify_tar(raw_archive, pins)
        if any(total[key] != group[key] for key in ['entries', 'expandedBytes']):
            raise ValueError('complete raw member count/size differs')
        names = {name for name in pins if len(PurePosixPath(name).parts) <= 3
                 and PurePosixPath(name).name in ('manifest.json', 'audit.json') and not pins[name].get('kind')}
        failed = [row for row in v['failedAssertions'] if row['suite'] == group['name']]
        names.update(row['rawMember'] for row in failed)
        selected = selected_members(raw_archive, names)
        reproduced = retained_main_manifests(selected, pins)
        if reproduced != sorted(group['mainManifests'], key=lambda row: row['manifest']):
            raise ValueError('original main suite artifact refs/outcomes/audits differ')
        raw = selected[group['originalMainManifest']]
        if sha(raw) != group['originalMainManifestSHA256']:
            raise ValueError('original published manifest hash differs')
        main = json.loads(raw)
        outcomes = dict(Counter(case['status'] for case in main['cases']))
        if main['git_commit'] != SOURCE or outcomes != group['originalOutcomes']:
            raise ValueError('original suite source/count outcomes differ')
        cases = {case['id']: case for case in main['cases'] if case['status'] == 'FAIL'}
        if len(failed) != 6 or set(cases) != {row['caseId'] for row in failed}:
            raise ValueError('six real failed assertions per suite differ')
        for row in failed:
            raw = selected[row['rawMember']]
            if (sha(raw) != row['sha256'] or m[row['file']] != raw
                    or cases[row['caseId']]['artifacts']['assertions.json'] != row['sha256']):
                raise ValueError('exact real failed assertion bytes/reference differ')
            failures_seen.append(row['caseId'])
        suites.append({'name': group['name'], **total, 'caseOutcomes': outcomes,
                       'artifactReferences': sum(row['artifactReferences'] for row in reproduced)})
    if {row['name'] for row in suites} != {'offline', 'race'} or len(failures_seen) != 12:
        raise ValueError('complete two-suite/six-distinct-failure evidence absent')
    print(json.dumps({'integrity': 'PASS', 'scope': 'original3d3 archive-byte/reference proof; actual gate failed loading before report; release verdict absent',
                      'sourceFilesVerifiedInsideExternalPaidL': 1034, 'rawSuites': suites,
                      'plannedOwnInputs': 17, 'plannedMissingOwnInputs': 1, 'effects': 4,
                      'failedAssertions': 12, 'externalArchiveBoundaries': len(paths),
                      'actualGateExitCode': 1, 'childCLIExitStatusInOriginalLog': 2, 'actualGatePass': None,
                      'originalReleaseReportPresent': False, 'individualInternalCommandExitsPersisted': False,
                      'releaseCertifiedByCurator': False, 'originalOutcomesUpgraded': False,
                      'capturedCommandsExecuted': False, 'environmentRead': False, 'networkUsed': False}))


if __name__ == '__main__':
    main()
