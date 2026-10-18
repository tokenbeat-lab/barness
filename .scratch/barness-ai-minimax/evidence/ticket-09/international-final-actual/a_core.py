"""Pure bounded archive/reference helpers. No process, environment or network use."""
from pathlib import Path
import json
import tarfile
from retention_core import safe_name, sha

SOURCE = '3d3a1fb28e143a2c29d684f78edcad94c0217243'


def selected_members(path, names):
    result = {}
    total = 0
    with tarfile.open(path, 'r|gz') as tar:
        for member in tar:
            if member.name not in names:
                continue
            if not member.isfile() or member.size > 32 << 20 or member.name in result:
                raise ValueError('unsafe selected regular archive member')
            total += member.size
            if total > 96 << 20:
                raise ValueError('selected metadata total bound')
            result[safe_name(member.name)] = tar.extractfile(member).read()
    if set(result) != set(names):
        raise ValueError('selected original archive members absent')
    return result


def command_log_contract(log_raw, outer, manifest):
    if (outer.get('state') != 'COMPLETED' or outer.get('exitCode') != 1
            or outer.get('credentialsRead') is not False
            or outer.get('launcherAudit') != {'passed': True, 'findings': []}):
        raise ValueError('original closed failed outer gate/audit receipt required')
    argv = outer.get('argv')
    if argv != manifest['argv']['gate']:
        raise ValueError('original outer gate argv differs from preregistered plan')
    if any(arg in ('-bundle', '-race-bundle') or arg.startswith(('-bundle=', '-race-bundle=')) for arg in argv):
        raise ValueError('supplied suite bundle cannot prove real public execution')
    paths = [argv[n+1] for n, arg in enumerate(argv) if arg == '-live']
    if len(paths) != 17 or paths != [row['path'] for row in manifest['plan']['inputs']]:
        raise ValueError('original17 planned live inputs differ')
    expected = [
        ('go-vet', 'go vet ./...'),
        ('go-vet-live', 'go vet -tags live ./...'),
        ('go-test', 'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -count=1 -timeout=20m ./...'),
        ('go-test-race', 'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race -count=1 -timeout=75m ./...')]
    text = log_raw.decode('utf-8')
    banners = [line for line in text.splitlines() if line.startswith('releasegate: ')]
    terminal = 'releasegate: release: invalid or duplicate evidence case directory'
    if banners != ['releasegate: ' + line for _, line in expected] + [terminal]:
        raise ValueError('original four command banners/terminal load error differ')
    if not text.endswith(terminal + '\nexit status 2\n'):
        raise ValueError('original pre-report CLI failure closure differs')
    result = []
    for role, line in expected:
        banner = ('releasegate: ' + line + '\n').encode()
        offset = log_raw.index(banner)
        result.append({'role': role, 'originalBanner': banner.decode().rstrip('\n'),
                       'logByteOffset': offset, 'bannerBytes': len(banner),
                       'individualExitCode': None,
                       'scope': 'Exact original log banner only; numeric internal command result was not persisted.'})
    return result


def check_reference_contract(config, manifest):
    refs = config['ownInputReferences']
    expected = manifest['plan']['inputs']
    if len(refs) != 17 or [(row['identity'], row['path']) for row in refs] != [(row['identity'], row['path']) for row in expected]:
        raise ValueError('all17 original planned input references required')
    missing = [row for row in refs if row['expectedPresence'] == 'MISSING']
    if len(missing) != 1 or missing[0]['identity'] != 'typesafe-classifier-jev-1.14.0':
        raise ValueError('honest missing L/support planned role required')
    for row in refs:
        if row['expectedPresence'] == 'MISSING':
            continue
        pin_key = (row['path'] + '/live-report.json').removeprefix('$REPO/')
        if manifest['plan']['originalPins'][pin_key] != row['sha256']:
            raise ValueError('planned own original report pin differs')
    if len(config['effectReferences']) != 4:
        raise ValueError('all4 original effect references required')
    for row in config['effectReferences']:
        kind, role = row['identity'].split('-', 1)
        if manifest['plan'][kind][role] != row['path']:
            raise ValueError('original effect planned path differs')


def prove_paid_source(paid, archive):
    if paid['sourceCommit'] != SOURCE or paid['sourceFiles'] != 1034:
        raise ValueError('original paid source-proof identity differs')
    names = {name for name in paid['files'] if name.startswith('source/')}
    names.add('source-at-start.json')
    found = selected_members(archive, names)
    source = json.loads(found['source-at-start.json'])
    if source['sourceCommit'] != SOURCE or len(source['files']) != 1034:
        raise ValueError('original source-at-start identity differs')
    if names != {'source/' + name for name in source['files']} | {'source-at-start.json'}:
        raise ValueError('exact external selected source set differs')
    for name, raw in found.items():
        if name not in paid['files'] or sha(raw) != paid['files'][name]['sha256']:
            raise ValueError('original external source-proof file differs')
        if name.startswith('source/') and sha(raw) != source['files'][name.removeprefix('source/')]:
            raise ValueError('original committed source pin differs')
    return found
