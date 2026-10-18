"""DRAFT collector: execute only after explicit parent A-CLOSED notification."""
from pathlib import Path
from collections import Counter
import argparse
import datetime
import json
from retention_core import sha, digest, stable, pack_tree, pack_bytes, verify_tar, main_manifests
from a_core import SOURCE, selected_members, command_log_contract, check_reference_contract, prove_paid_source


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('config', type=Path)
    p.add_argument('--a-closed', action='store_true', required=True)
    a = p.parse_args()
    config_raw = stable(a.config, 2 << 20)
    c = json.loads(config_raw)
    if c['parentClosureConfirmed'] is not True or c['sourceCommit'] != SOURCE:
        raise ValueError('separate explicit parent A-CLOSED confirmation required')
    repo = Path(c['repo'])
    base = Path(c['collection'])
    if base != repo / '.evidence/minimax-global-final-keyless':
        raise ValueError('only original final-keyless A is eligible')
    out = Path(c['out']).resolve()
    if not out.is_relative_to(Path('/tmp').resolve()):
        raise ValueError('only /tmp outputs permitted')
    outer_raw = stable(base / 'actual-releasegate-command.json', 2 << 20)
    outer = json.loads(outer_raw)
    if outer.get('state') != 'COMPLETED' or outer.get('launcherAudit') != {'passed': True, 'findings': []}:
        raise ValueError('actual public gate has not closed')
    manifest_raw = stable(base / 'argv-manifest.json', 2 << 20)
    if sha(manifest_raw) != c['argvManifestSHA256']:
        raise ValueError('original published input plan changed')
    manifest = json.loads(manifest_raw)
    if (base / 'gate/release-report.json').exists() or (base / 'gate/release-report.json').is_symlink():
        raise ValueError('original failed-load window has no release report; do not splice later output')
    log_raw = stable(base / 'actual-releasegate.log', 32 << 20)
    if sha(log_raw) != c['closedLogSHA256']:
        raise ValueError('original closed gate log differs')
    actual_commands = command_log_contract(log_raw, outer, manifest)
    check_reference_contract(c, manifest)
    if manifest['plan']['sourceCommit'] != SOURCE or outer['argv'] != c['expectedOuterArgv']:
        raise ValueError('original paid collection/gate source or argv differs')
    summary_raw = stable(base / 'summary.json', 2 << 20)
    summary = json.loads(summary_raw)
    if (summary.get('gateExitCode') != 1 or summary.get('releasePass') is not None
            or summary.get('passVerdict') is not False or summary.get('paidCalls') != 0
            or summary.get('releaseReportPresent') is not False
            or summary.get('releaseReportSHA256') is not None
            or summary.get('releaseReportReadError') != 'FileNotFoundError'
            or sha(summary_raw) != c['closedSummarySHA256'] or sha(outer_raw) != c['closedOuterSHA256']):
        raise ValueError('original closed failed keyless summary differs')
    runner_raw = stable(Path(c['runnerPath']), 2 << 20)
    if sha(runner_raw) != c['runnerSHA256'] or manifest['runnerSHA256'] != c['runnerSHA256']:
        raise ValueError('original parent runner differs')
    for group in c['rawGroups']:
        root = Path(group['path'])
        if root != base / 'gate' / group['name'] or group['name'] not in ('offline', 'race'):
            raise ValueError('raw split must be original ordinary/race roots')
        audit = json.loads(stable(root / group['session'] / 'audit.json', 16 << 20))
        if audit.get('findings') != []:
            raise ValueError('closed suite must pass keyless audit before raw collection')
    for path in base.glob('*-command.json'):
        command = json.loads(stable(path, 2 << 20))
        if (command.get('state') != 'COMPLETED' or command.get('credentialsRead') is not False
                or command.get('launcherAudit') != {'passed': True, 'findings': []}):
            raise ValueError('original closed keyless command/audit required before logs')
    out.mkdir(parents=True, exist_ok=True)
    items = {}
    paths = {}
    def add(name, raw):
        if name in items:
            raise ValueError('duplicate metadata selection')
        items[name] = raw
    for label, directory in [('A', base), ('gate', base / 'gate')]:
        for file in sorted(directory.iterdir()):
            if not file.is_file():
                continue
            if file.suffix not in ('.json', '.md', '.log', '.txt'):
                raise ValueError('unclassified top-level closed A metadata')
            name = 'metadata/' + label + '/' + file.name
            add(name, stable(file, 32 << 20))
            paths[name] = file
    add('producer/parent-keyless-runner.py', runner_raw)
    paid_raw = stable(Path(c['paidIndexPath']), 8 << 20)
    if sha(paid_raw) != c['paidIndexSHA256']:
        raise ValueError('immutable paid L source proof index differs')
    paid = json.loads(paid_raw)
    add('source-proof/original-paid-L-index.json', paid_raw)
    externals = []
    external_paths = {}
    for entry in c['externalArchives']:
        path = Path(entry['path'])
        if digest(path) != entry['sha256'] or path.stat().st_size != entry['bytes']:
            raise ValueError('original immutable external archive boundary differs')
        external_paths[entry['file']] = path
        externals.append({key: entry[key] for key in ['file', 'sha256', 'bytes', 'scope']})
    proved = prove_paid_source(paid, external_paths[c['paidArchive']])
    add('source-proof/source-at-start.json', proved['source-at-start.json'])
    for name in c['publicSourceFiles']:
        add('source-proof/public-gate/' + name, proved['source/' + name])
    reference_results = []
    for entry in c['ownInputReferences'] + c['effectReferences']:
        row = dict(entry)
        if entry.get('expectedPresence') == 'MISSING':
            missing = repo / entry['path'].removeprefix('$REPO/')
            if missing.exists() or missing.is_symlink():
                raise ValueError('original expected missing support path now exists; investigate without relabelling')
            row['observedPresence'] = 'MISSING'
        else:
            raw = selected_members(external_paths[entry['archive']], {entry['member']})[entry['member']]
            if sha(raw) != entry['sha256']:
                raise ValueError('original own/effect archive member differs')
            row['observedPresence'] = 'EXTERNAL_ORIGINAL_BYTES_VERIFIED'
        reference_results.append(row)
    raw_groups = []
    failed_assertions = []
    for group in c['rawGroups']:
        root = Path(group['path'])
        manifest_path = root / group['manifest']
        main_raw = stable(manifest_path, 16 << 20)
        if sha(main_raw) != group['publishedManifestSHA256']:
            raise ValueError('original published suite manifest changed')
        main = json.loads(main_raw)
        outcomes = dict(Counter(case['status'] for case in main['cases']))
        failures = [case for case in main['cases'] if case['status'] == 'FAIL']
        if (main['git_commit'] != SOURCE or len(main['cases']) != group['expectedCaseCount']
                or outcomes != group['expectedOutcomes']
                or [case['id'] for case in failures] != group['failedCaseIds']):
            raise ValueError('original six actual failed cases/inventory differ')
        archive = out / ('A-' + group['name'] + '-3d3.tar.gz')
        raw_summary, inventory = pack_tree(root, archive)
        raw_summary.update({'name': group['name'], 'mainManifests': main_manifests(root, inventory),
                            'originalMainManifest': group['manifest'], 'originalMainManifestSHA256': sha(main_raw),
                            'originalOutcomes': outcomes,
                            'sourceCommit': SOURCE,
                            'intendedCommonRetention': '$COMMON/.evidence/minimax-retained/' + archive.name})
        raw_groups.append(raw_summary)
        derived = {'scope': 'Derived index only; exact full manifest remains inside raw archive',
                   'sourceCommit': SOURCE, 'originalManifest': group['manifest'],
                   'originalManifestSHA256': sha(main_raw), 'cases': len(main['cases']),
                   'caseOutcomes': outcomes, 'actualFailedCases': failures}
        add('suite-summaries/' + group['name'] + '/derived-index.json',
            json.dumps(derived, ensure_ascii=False, indent=2).encode() + b'\n')
        audit_raw = stable(root / group['session'] / 'audit.json', 16 << 20)
        add('suite-summaries/' + group['name'] + '/audit.json', audit_raw)
        for case in failures:
            member = group['session'] + '/' + case['dir'] + '/assertions.json'
            raw = stable(root / member, 16 << 20)
            if inventory[member]['sha256'] != sha(raw) or case['artifacts']['assertions.json'] != sha(raw):
                raise ValueError('actual failed assertion original artifact reference differs')
            name = 'failed-assertions/' + group['name'] + '/' + case['id'] + '.json'
            add(name, raw)
            failed_assertions.append({'suite': group['name'], 'caseId': case['id'], 'status': 'FAIL',
                                      'file': name, 'sha256': sha(raw), 'rawArchive': archive.name,
                                      'rawMember': member})
    for entry in c['parentAuditFiles']:
        raw = stable(Path(entry['path']), 16 << 20)
        if sha(raw) != entry['sha256']:
            raise ValueError('original parent audit pin differs')
        add('parent-audits/' + Path(entry['path']).name, raw)
    for name, path in paths.items():
        if stable(path, 32 << 20) != items[name]:
            raise ValueError('closed A metadata changed during collection')
    if stable(base / 'summary.json', 2 << 20) != summary_raw or stable(base / 'actual-releasegate-command.json', 2 << 20) != outer_raw:
        raise ValueError('original gate closure changed during collection')
    tools = {}
    for name in ['curate_a.py', 'verify_a.py', 'a_core.py', 'retention_core.py', 'README.md', 'FAILURE-INVENTORY.md']:
        raw = stable(Path(__file__).parent / name, 2 << 20)
        add('curation-tools/' + name, raw)
        tools[name] = {'sha256': sha(raw), 'bytes': len(raw)}
    add('curation-config.json', config_raw)
    archive = out / 'A-final-keyless-3d3-metadata.tar.gz'
    pack_bytes(archive, items)
    pins = {name: {'sha256': sha(raw), 'bytes': len(raw)} for name, raw in items.items()}
    roundtrip = verify_tar(archive, pins)
    if archive.stat().st_size > 8 << 20:
        raise ValueError('metadata capsule exceeded8MiB target')
    result = {'schema': 1, 'kind': 'original3d3-actual-failed-gate-byte-retention',
              'createdAt': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'sourceCommit': SOURCE,
              'sourceFilesProvedExternally': 1034, 'fullSourceByteArchive': paid['archive'],
              'fullSourceByteArchiveSHA256': paid['archiveSha256'],
              'archive': archive.name, 'archiveSha256': digest(archive), 'archiveBytes': archive.stat().st_size,
              'files': pins, 'archiveRoundtrip': roundtrip, 'rawGroups': raw_groups,
              'plannedOwnInputs': reference_results[:17], 'effectReferences': reference_results[17:],
              'failedAssertions': failed_assertions, 'originalCommandBannerEvidence': actual_commands,
              'individualInternalCommandExitCodesPersisted': False,
              'actualGateExitCode': outer['exitCode'], 'actualGatePass': None,
              'originalReleaseReportPresent': False, 'childCLIExitStatusInOriginalLog': 2,
              'actualGateTerminalState': 'FAILED_AT_OFFLINE_BUNDLE_LOAD_BEFORE_REPORT',
              'actualSummaryPassVerdict': summary['passVerdict'], 'externalArchives': externals,
              'toolFiles': tools, 'releaseCertifiedByCurator': False, 'originalOutcomesUpgraded': False,
              'missingReportsInvented': False, 'additionalRoundAuthorized': False, 'paidCallsByCurator': 0,
              'rootModified': False, 'gitModified': False, 'credentialsRead': False,
              'environmentRead': False, 'networkUsed': False, 'testsRun': False,
              'capturedCommandsExecuted': False, 'scope': 'Only original3d3 closed A. Later source fixes are separate. Byte integrity never repairs actual failed acceptance.'}
    with (out / 'A-final-keyless-3d3.json').open('x') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
        f.write('\n')
    print(json.dumps({'integrity': 'PASS', 'actualGatePass': None, 'releaseReportPresent': False, 'actualGateExitCode': 1,
                      'releaseCertifiedByCurator': False, 'rawGroups': len(raw_groups),
                      'metadataArchiveSha256': result['archiveSha256']}))


if __name__ == '__main__':
    main()
