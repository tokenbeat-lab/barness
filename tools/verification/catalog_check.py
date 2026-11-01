#!/usr/bin/env python3
"""Freeze and verify one controlled catalog candidate; never run vendor routes."""
import argparse
import base64
import json
import os
from pathlib import Path
import shutil
import sys

from receipt_files import json_file, save, sha, source_snapshot, within
from catalog_check_receipt import verify_receipt


def fixed_argv(directory, timeout):
    if type(timeout) is not int or not 1 <= timeout <= 1200:
        raise ValueError('timeout must be an integer in 1..1200 seconds')
    return ['go', 'test', '-json', './ai/e2e', '-run', '^TestCatalogCandidateContracts$',
            '-count=1', '-timeout=' + str(timeout) + 's', '-args', '-catalog-candidate',
            directory + '/snapshot.json']


def extra_sources(root):
    files = {}
    for scope in ('catalog', 'cmd'):
        base = within(root, scope)
        for parent, directories, names in os.walk(base, followlinks=False):
            directories[:] = [n for n in directories if n not in ('__pycache__', '.evidence', 'node_modules')]
            for name in directories:
                if (Path(parent) / name).is_symlink():
                    raise ValueError('symbolic catalog implementation rejected')
            for name in names:
                path = Path(parent) / name
                if name in ('.DS_Store', 'Thumbs.db') or path.suffix.lower() in ('.md', '.pyc', '.test', '.exe'):
                    continue
                relative = str(path.relative_to(root))
                files[relative] = sha(within(root, relative))
    return files


def prepare(root, directory, feature, timeout):
    argv = fixed_argv(directory, timeout)
    out = within(root, directory)
    candidate = json_file(out / 'candidate.json')
    if base64.b64decode(candidate['snapshot'], validate=True) != (out / 'snapshot.json').read_bytes():
        raise ValueError('frozen candidate snapshot differs')
    source, extra = source_snapshot(root), extra_sources(root)
    files = dict(source['files'])
    files.update(extra)
    for name, digest in files.items():
        target = out / 'source' / name
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open('xb') as destination, within(root, name).open('rb') as original:
            shutil.copyfileobj(original, destination)
        if sha(target) != digest:
            raise ValueError('source changed while freezing')
    if source_snapshot(root) != source or extra_sources(root) != extra:
        raise ValueError('implementation changed while freezing')
    frozen = {name: sha(out / name) for name in ('candidate.json', 'snapshot.json', 'baseline.json')}
    manifest = {'schema': 1, 'directory': directory, 'feature': feature, 'timeout': timeout,
                'source': source, 'extra': extra, 'originals': files, 'frozen': frozen, 'argv': argv}
    save(out / 'prepared.json', manifest)
    return {'prepared': str((out / 'prepared.json').relative_to(root)), 'inputs': selected_inputs(root, manifest)}


def selected_inputs(root, prepared):
    files = dict(prepared['extra'])
    for name, digest in prepared['frozen'].items():
        files[prepared['directory'] + '/' + name] = digest
    name = prepared['directory'] + '/prepared.json'
    files[name] = sha(within(root, name))
    return files


def inspect(root, prepared_name, receipt_name):
    prepared = json_file(within(root, prepared_name))
    if type(prepared.get('schema')) is not int or prepared['schema'] != 1:
        raise ValueError('unsupported prepared schema')
    if prepared.get('argv') != fixed_argv(prepared['directory'], prepared['timeout']):
        raise ValueError('prepared command is not the fixed candidate contract entry')
    out = within(root, prepared['directory'])
    if prepared_name != prepared['directory'] + '/prepared.json':
        raise ValueError('prepared location changed')
    if source_snapshot(root)['files'] != prepared['source']['files'] or extra_sources(root) != prepared['extra']:
        raise ValueError('actual implementation source set or bytes changed')
    originals = dict(prepared['source']['files'])
    originals.update(prepared['extra'])
    if originals != prepared['originals']:
        raise ValueError('incomplete frozen implementation inventory')
    for name, digest in originals.items():
        if sha(within(out / 'source', name)) != digest:
            raise ValueError('retained implementation bytes changed')
    for name, digest in prepared['frozen'].items():
        if sha(within(out, name)) != digest:
            raise ValueError('frozen candidate or baseline changed')
    candidate = json_file(out / 'candidate.json')
    snapshot = json_file(out / 'snapshot.json')
    if base64.b64decode(candidate['snapshot'], validate=True) != (out / 'snapshot.json').read_bytes():
        raise ValueError('candidate result bytes changed')
    result = verify_receipt(root, prepared, selected_inputs(root, prepared), receipt_name, snapshot)
    result.update(candidateId=candidate['id'], baseDigest=candidate['baseDigest'])
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='command', required=True)
    freeze = sub.add_parser('prepare')
    freeze.add_argument('--directory', required=True)
    freeze.add_argument('--feature', required=True)
    freeze.add_argument('--timeout', required=True, type=int)
    check = sub.add_parser('inspect')
    check.add_argument('--prepared', required=True)
    check.add_argument('--receipt', required=True)
    args = parser.parse_args()
    root = Path.cwd().resolve()
    try:
        if args.command == 'prepare':
            result = prepare(root, args.directory, args.feature, args.timeout)
        else:
            result = inspect(root, args.prepared, args.receipt)
        print(json.dumps(result, sort_keys=True))
        return 0
    except (OSError, ValueError, KeyError, TypeError) as error:
        print(json.dumps({'status': 'execution_failed', 'message': str(error)}))
        return 1


if __name__ == '__main__':
    sys.exit(main())
