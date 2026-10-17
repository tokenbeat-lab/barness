#!/usr/bin/env python3
"""Seal/recheck ticket-14 evidence; never turn a failed release into PASS."""
import argparse
from collections import Counter
import hashlib
import json
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
RECEIPT = ROOT/'.scratch/jev-1.14-multimodal-box/14-evidence.json'
BINS = {'live.test','chineseeval','visualeval'}


def raw(path):
    if not path.is_file() or any(p.is_symlink() for p in (path,*path.parents)):
        raise ValueError('nonregular evidence file')
    return path.read_bytes()


def sha(path):
    return hashlib.sha256(raw(path)).hexdigest()


def doc(path):
    return json.loads(raw(path))


def relative(path):
    return path.relative_to(ROOT).as_posix()


def stored_identity(path):
    if path.is_symlink():
        if any(p.is_symlink() for p in path.parents):raise ValueError('linked parent')
        return dict(kind='symlink',targetSHA256=hashlib.sha256(os.fsencode(os.readlink(path))).hexdigest())
    return sha(path)


def inventory(base, files, allow_links=False):
    for name, want in files.items():
        p=Path(name)
        if p.is_absolute() or '..' in p.parts:raise ValueError('unsafe inventory path')
        if isinstance(want,dict):
            if not allow_links or want.get('kind')!='symlink' or stored_identity(base/p)!=want:
                raise ValueError('retained link identity mismatch')
        elif sha(base/p)!=want.removeprefix('sha256:'):
            raise ValueError('evidence inventory mismatch')
    return len(files)


def bundle(path, allow_incomplete=False):
    m=doc(path/'manifest.json');count=0
    for c in m.get('cases',[]):
        if not isinstance(c,dict) or not c.get('artifacts'):
            raise ValueError('case inventory missing')
        if c['dir']!=c['id'] or Path(c['dir']).name!=c['dir']:
            raise ValueError('unsafe case directory')
        directory=path/c['dir']
        if {p.name for p in directory.iterdir()}!=set(c['artifacts']):
            raise ValueError('incomplete case inventory')
        count+=inventory(directory,c['artifacts'])
        if not set(c.get('fixtures',{}))<=set(c['artifacts']):
            raise ValueError('fixture missing from stored-byte inventory')
    audit=doc(path/'audit.json') if (path/'audit.json').is_file() else None
    if audit is None and not allow_incomplete:raise ValueError('required audit missing')
    return dict(path=relative(path),package=m.get('package'),
                cases=dict(Counter(c['status'] for c in m.get('cases',[]))),
                verifiedArtifacts=count,auditFindings=len(audit['findings']) if audit is not None else None,
                auditState='RECORDED' if audit is not None else 'MISSING')


def visual_bundle(path):
    m=doc(path/'manifest.json');files=m.get('artifacts')
    if m.get('schema')!=1 or not isinstance(files,dict) or not files:
        raise ValueError('visual inventory missing')
    if tree_paths(path)!=set(files)|{'manifest.json'}:
        raise ValueError('visual inventory incomplete')
    inventory(path,files)
    audit=doc(path/'audit.json')
    return dict(kind='visual',file=relative(path/'manifest.json'),base=relative(path),
                count=len(files),auditFindings=len(audit['findings']))


def tree_paths(base):
    # Walk the actual tree, never trust a nested fixture's ownership claims.
    paths=set()
    for p in base.rglob('*'):
        if p.is_symlink():
            paths.add(p.relative_to(base).as_posix())
            continue
        if p.is_file() and not (p.parent==base and p.name in BINS):
            paths.add(p.relative_to(base).as_posix())
    return paths


def recorder_roots(base):
    # These are the runner's actual process outputs. Captured manifests and
    # intentional corrupt nested fixtures remain opaque bytes in tree indexes.
    patterns=('*/manifest.json','gate/offline/*/manifest.json','gate/race/*/manifest.json',
              'live-harness/*/manifest.json','preflight/*/*/manifest.json','text-*/*/manifest.json')
    paths=set(p for pattern in patterns for p in base.glob(pattern))
    for p in sorted(paths):
        if not p.parent.name[:8].isdigit():continue
        m=doc(p)
        if m.get('package') and 'cases' in m:yield p.parent


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record',help='final verification base, relative to repository')
    parser.add_argument('--verify',action='store_true')
    args=parser.parse_args()
    if args.verify:
        receipt=doc(RECEIPT);inventory(ROOT,receipt['files'])
        if sorted(relative(p) for p in (ROOT/'ai').rglob('*.go'))!=receipt['sourcePaths']:
            raise ValueError('complete Go source set changed')
        for run in receipt['runs']:
            if bundle(ROOT/run['path'])!=run:raise ValueError('bundle verdict drift')
        for run in receipt.get('subprocessRuns',[]):
            if bundle(ROOT/run['path'],allow_incomplete=True)!=run:raise ValueError('child recorder drift')
        for run in receipt['history']:
            if bundle(ROOT/run['path'],allow_incomplete=True)!=run:raise ValueError('history drift')
        for item in receipt['inventories']:
            if item.get('kind')=='visual':
                if visual_bundle(ROOT/item['base'])!=item:raise ValueError('visual package drift')
            else:inventory(ROOT/item['base'],doc(ROOT/item['file']))
        for item in receipt['trees']:
            base=ROOT/item['base'];files=doc(ROOT/item['file'])
            if tree_paths(base)!=set(files):raise ValueError('retained tree file set drift')
            inventory(base,files,allow_links=True)
        print('ticket-14 source, complete retained trees, primary inventories, audits and NOT_RUN verdicts verified')
        return
    if not args.record:parser.error('use --record or --verify')
    base=ROOT/args.record;report=doc(base/'gate/release-report.json')
    if report['pass']:raise ValueError('unexpected overall PASS while historical blocker and real evidence absent')
    if any(r['status']!='NOT_RUN' for r in report['live']['rows']) or len(report['live']['rows'])!=11:
        raise ValueError('real-route preflight was misrepresented')
    pressure=next(s for s in report['snapshots'] if s['name']=='pressure-stream-integrity')
    original=ROOT/'.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md'
    if not pressure['detail'].endswith('sha256:'+sha(original)):
        raise ValueError('original tracker changed during execution')
    receipt=dict(schema=1,kind='current-catalog-release-integration',overall='BLOCKED',
                 realCalls=0,realAttempts=0,files={},runs=[],subprocessRuns=[],history=[],inventories=[],trees=[],preflights=[],
                 commands=doc(base/'commands.json'),gates=report['gates'],live=report['live'],
                 executionCommit=report['commit'],goContentCommit=report['commit'],pressureState=pressure,
                 sourcePaths=sorted(relative(p) for p in (ROOT/'ai').rglob('*.go')),
                 note='Controlled HTTP only; real/effect routes NOT_RUN. Nested captured/corrupt fixture manifests are retained bytes, never accepted process verdicts.')
    def seal(path):receipt['files'][relative(path)]=sha(path)
    harness=sorted((base/'live-harness').glob('*/manifest.json'))
    # TestMain starts the parent recorder before its deliberate crash children.
    # Those child outputs are retained facts, never completed acceptance runs.
    children={p.parent for p in harness[1:]}
    for path in recorder_roots(base):
        if path in children:receipt['subprocessRuns'].append(bundle(path,allow_incomplete=True))
        else:receipt['runs'].append(bundle(path))
    for name in ('visual-decisions','visual-jev'):
        receipt['inventories'].append(visual_bundle(base/name))
    for p in sorted(base.glob('text-*/*/files.json')):
        files=doc(p);inventory(p.parent,files)
        receipt['inventories'].append(dict(file=relative(p),base=relative(p.parent),count=len(files)))
    for item in doc(base/'gate/fixtures.json'):
        p=ROOT/item['path']
        if len(raw(p))!=item['bytes'] or sha(p)!=item['sha256']:
            raise ValueError('checked-in fixture changed after execution')
        seal(p)
    parent=ROOT/'.evidence/ticket14';indexes=parent/'integrity';indexes.mkdir(exist_ok=True)
    for directory in sorted(parent.iterdir()):
        if directory.is_symlink():raise ValueError('linked evidence root')
        if not directory.is_dir() or directory==indexes:continue
        if directory!=base:
            for path in recorder_roots(directory):
                receipt['history'].append(bundle(path,allow_incomplete=True))
        # Complete immutable stored-byte index, including nested protocol/effect
        # packages, failed captures and audits. Exclude only rebuildable binaries.
        paths=tree_paths(directory);files={name:stored_identity(directory/name) for name in sorted(paths)}
        index=indexes/(directory.name+'.json')
        index.write_text(json.dumps(files,sort_keys=True,separators=(',',':'))+'\n')
        seal(index);receipt['trees'].append(dict(base=relative(directory),file=relative(index),count=len(files)))
    for p in sorted(parent.glob('*')):
        if p.is_file():seal(p)
    for p in sorted((ROOT/'ai').rglob('*.go')):seal(p)
    for name in ('go.mod','go.sum','ai/release/traceability.json','ai/release/classifier-extensions.json',
                 'ai/release/catalog-snapshot.json','ai/live/support-matrix.json',
                 'ai/internal/testkit/release/testdata/multimodal-load.json',
                 'ai/internal/testkit/pioracle/node/package-lock.json',
                 'ai/internal/testkit/pioracle/node/runner.mjs',
                 'ai/internal/testkit/pioracle/node/package.json',
                 'ai/internal/testkit/pioracle/node/provenance.json',
                 'ai/internal/testkit/pioracle/node/model-data.sha256',
                 'ai/e2e/testdata/pidiff/ledger.json',
                 '.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md',
                 '.scratch/jev-1.14-multimodal-box/run-14-verification.py',
                 '.scratch/jev-1.14-multimodal-box/14-code-review.md',
                 '.scratch/jev-1.14-multimodal-box/14-verification.md',
                 '.scratch/jev-1.14-multimodal-box/issues/14-current-catalog-release-integration.md',
                 '.scratch/barness-ai-release-verification/issues/01-unary-race-timeout-stop-reason.md',
                 '.scratch/barness-ai-release-verification/issues/02-decisions-notrun-case-inventory.md',
                 'docs/barness-ai/current-catalog-release.md'):
        seal(ROOT/name)
    seal(Path(__file__).resolve())
    for directory in sorted((base/'preflight').iterdir()):
        manifests=list(directory.glob('*/manifest.json'))
        if len(manifests)!=1:raise ValueError('route preflight package missing')
        package=manifests[0].parent;formal=directory.name=='typesafe-classifier-jev-1.14.0'
        r=doc(package/('multimodal-report.json' if formal else 'live-report.json'));b=r['budget']
        calls,attempts=(b['actual']['calls'],b['actual']['attempts']) if formal else (b['callsUsed'],b['httpAttempts'])
        if calls!=0 or attempts!=0:raise ValueError('preflight unexpectedly spent real requests')
        if formal:
            if r['outcome']!='NOT_RUN' or r['complete']:raise ValueError('formal preflight status changed')
        elif any(case['outcome']!='NOT_RUN' for case in r['scenarios']):raise ValueError('route preflight status changed')
        receipt['preflights'].append(dict(route=directory.name,calls=calls,attempts=attempts,package=relative(package)))
    for p in sorted(base.glob('visual-*/report.json')):
        r=doc(p)
        if r['status']!='NOT_RUN' or r['realQuality']!='NOT_RUN' or r['budget']['calls']!=0:
            raise ValueError('visual preflight makes real-quality claim')
    for p in sorted(base.glob('text-*/*/evaluation-report.json')):
        if doc(p)['status']!='NOT_RUN':raise ValueError('text preflight status changed')
    RECEIPT.write_text('{\n'+',\n'.join('  '+json.dumps(k)+': '+json.dumps(v,ensure_ascii=False,separators=(',',':')) for k,v in receipt.items())+'\n}\n')
    print('sealed',len(receipt['runs']),'primary current packages;',len(receipt['history']),'historical packages;',len(receipt['trees']),'complete trees; overall BLOCKED')


if __name__=='__main__':main()
