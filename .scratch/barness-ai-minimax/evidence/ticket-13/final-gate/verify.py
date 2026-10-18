#!/usr/bin/env python3
"""Read-only finite verification of the failed compiled A2 full gate at8fbd."""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import stat
import sys
import types

BASE=Path(__file__).resolve().parent
SOURCE='8fbd17006677774368d4d77d924a8672fd205c51'
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
COMPILED='c54a7d07f5780711af334a4369dcfd1dbb2b3b8bd440f09962b88831d11a906a'
RUNNER='a1ecf4a0b2285732aada1a9d7026f693f7c05cef5b73ddb21b1a2bfe13226d4a'
OBSERVER='aec940df7db842009d9c977142dc35f604022b380ce383af7ce2dd85383879d7'
EXITS={'vet':0,'vet-live':0,'ordinary':1,'race':1}
FAILURES=['GLM12-formal-import-chat','GLM12-formal-import-messages','GLM12-formal-import-responses','GLM12-reported-known-usage-chat','GLM12-reported-known-usage-messages','GLM12-reported-known-usage-responses','MM10-release-contract','P0-GLM13-release-trace-complete']


def helper():
    path=BASE.parent/'implementation/verify.py';fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:raise ValueError('pinned bounded verifier differs')
    module=types.ModuleType('bounded_fullgate');module.__file__=str(path);exec(compile(raw,str(path),'exec'),module.__dict__);return module


def outer(b,pins,m,archive):
    def doc(n):return b.doc(m[n])
    report=doc('outer/gate/release-report.json');summary=doc('outer/summary.json');plan=doc('outer/plan.json');command=doc('outer/actual-releasegate-command.json')
    if (report['commit']!=SOURCE or report['pass'] is not False or summary['sourceCommit']!=SOURCE or summary['releasePass'] is not False or summary['passVerdict'] is not False or
        summary['gateExit']!=1 or summary['releaseReportPresent'] is not True or summary['releaseReportError'] is not None or summary['allFourNumericExitsPresent'] is not True or
        summary['credentialsRead'] is not False or summary['paidCalls']!=0 or summary['automaticReruns'] is not False or
        command['state']!='COMPLETED' or command['exitCode']!=1 or command['credentialsRead'] is not False or command['credentialsInherited'] is not False):raise ValueError('original actual failed outer verdict differs')
    argv=command['argv']
    if not argv[0].endswith('/minimax-routed-final-compiled-1/releasegate') or argv.count('-live')!=17 or any(x in argv for x in ('-bundle','-race')):raise ValueError('actual full public CLI argv differs')
    if (plan['sourceCommit']!=SOURCE or plan['runnerSHA256']!=RUNNER or plan['observerSHA256']!=OBSERVER or plan['compiledCLISHA256']!=COMPILED or len(plan['canonicalActuals'])!=18 or
        pins['outer/barness-routed-final-keyless-compiled.py']['sha256']!=RUNNER or pins['outer/barness-routed-go-recorder.py']['sha256']!=OBSERVER or
        pins['outer/observer-bin/go']['sha256']!=OBSERVER or doc('outer/compiled-binary.json')['sha256']!=COMPILED or
        plan['buildReceiptSHA256']!=pins['outer/build-public-releasegate-command.json']['sha256'] or doc('outer/build-public-releasegate-command.json')['exitCode']!=0):raise ValueError('actual source/compiler/observer pins differ')
    review=doc('outer/parent-review-receipt.json')
    if (review['sourceCommit']!=SOURCE or review['planSHA256']!=pins['outer/plan.json']['sha256'] or review['helperSHA256']!=RUNNER or review['binarySHA256']!=COMPILED or
        review['reviews']['standards']['decision']!='APPROVED' or review['reviews']['spec']['decision']!='PASS' or plan['gate']!=argv or
        plan['originalPins']!=doc('outer/frozen-inputs.json') or len(plan['originalPins'])!=137 or
        plan['originalPins']['.evidence/minimax-routed-legacy-1/source-at-start.json']!=pins['source/source-at-start.json']['sha256']):raise ValueError('actual reviewed plan/input/source pins differ')
    reference=b.doc(b.read(BASE,'binary-reference.json',1<<20))
    if reference['sha256']!=COMPILED or reference['copied'] is not False or reference['existingMember']!='compiled-prepared/releasegate' or reference['existingImmutableArchive']!='.evidence/minimax-retained/ticket-13-keyless-launch-history/keyless-launch-history-raw.tar.gz':raise ValueError('existing compiled binary reference differs')
    for label,code in EXITS.items():
        receipt=doc('outer/gate-command-exits/'+label+'.json')
        if receipt['state']!='COMPLETED' or type(receipt['exitCode']) is not int or receipt['exitCode']!=code or receipt['credentialsRead'] is not False or receipt['cwd']!='$REPO' or summary['actualInternalCommandExits'][label]!={'state':'COMPLETED','exitCode':code}:raise ValueError('four actual numeric command exits differ')
    if len(summary['steps'])!=9 or any(v['exitCode']!=0 or v['credentialsRead'] is not False for v in summary['steps'][:-1]) or summary['steps'][-1]['exitCode']!=1:raise ValueError('eight public native verifiers differ')
    for step in summary['steps'][:-1]:
        v=doc('outer/'+step['name']+'-command.json')
        if v['state']!='COMPLETED' or v['exitCode']!=0 or v['credentialsRead'] is not False:raise ValueError('original public verifier receipt differs')
    native={v['id']:v['passed'] for v in report['commands']}
    if native!={'go-vet':True,'go-vet-live':True,'go-test':False,'go-test-race':False}:raise ValueError('public Boolean commands differ from numeric receipts')
    tails={v['id']:v.get('tail','') for v in report['commands']}
    for name in ('go-test','go-test-race'):
        if any(s not in tails[name] for s in ('TestDecisionsEvidenceRejectsSourceAndCaptureDrift/valid','artifact inventory invalid','TestGLMReleaseContract','TestReleasePlanCannotBeReduced')):raise ValueError('original extra release-package unit failures missing')
    live=report['live']['rows'];statuses=dict(collections.Counter(v['status'] for v in live))
    if len(live)!=17 or statuses!={'PASS':16,'FAIL':1} or [v['combo'] for v in live if v['status']!='PASS']!=['zhipu-responses']:raise ValueError('actual native17 status distribution differs')
    if len(report['audits'])!=23 or any(v.get('findings') or v.get('err') for v in report['audits']):raise ValueError('actual public redaction audit differs')
    if report['differential']['pass']!=868 or report['differential']['pending']!=0 or report['differential']['fail']!=0:raise ValueError('public frozen differential counters differ')
    return dict(sourceCommit=SOURCE,gateExit=1,releasePass=False,passVerdict=False,releaseReportPresent=True,actualNumericExits=EXITS,
        publicVerifiers=8,liveStatuses=statuses,liveFailure='zhipu-responses',failedGates=[v['id'] for v in report['gates'] if not v['pass']],
        traceRows=len(report['traceability']),traceStatuses=dict(collections.Counter(v['status'] for v in report['traceability'])),audits=23,auditFindings=0,
        additionalUnregisteredUnitFailure='TestDecisionsEvidenceRejectsSourceAndCaptureDrift/valid: artifact inventory invalid',
        otherUnregisteredMappingFailures=['TestGLMReleaseContract','TestReleasePlanCannotBeReduced'],credentialsRead=False,paidCalls=0,archive=archive)


def suite(b,index,pins,m,archive):
    name=index['mainManifest'];manifest=b.doc(m[name]);prefix=name.rsplit('/',1)[0]+'/';cases=manifest['cases']
    if manifest['git_commit']!=SOURCE or len(cases)!=12448 or len({c['id'] for c in cases})!=12448:raise ValueError('actual complete-suite identity/count differs')
    statuses=dict(collections.Counter(c['status'] for c in cases));failed=sorted(c['id'] for c in cases if c['status']=='FAIL')
    if statuses!={'PASS':12439,'FAIL':8,'NOT_RUN':1} or failed!=FAILURES or [c['id'] for c in cases if c['status']=='NOT_RUN']!=['OD57-release-recorded-bundle-diagnostic'] or b.doc(m[prefix+'audit.json'])['findings']:raise ValueError('original eight failures/optional NOT_RUN/audit differ')
    refs=0
    for case in cases:
        for leaf,digest in case['artifacts'].items():
            directory=PurePosixPath(case['dir']);artifact=PurePosixPath(leaf)
            if directory.is_absolute() or artifact.is_absolute() or '..' in (*directory.parts,*artifact.parts):raise ValueError('unsafe native artifact reference')
            if pins.get(prefix+(directory/artifact).as_posix(),{}).get('sha256')!=digest:raise ValueError('complete-suite native artifact pin differs')
            refs+=1
    pi=[c for c in cases if c['id'].startswith('PIDIFF-')];pressure=[c for c in cases if 'pressure' in c['id'].lower()]
    if len(pi)!=868 or any(c['status']!='PASS' for c in pi) or len(pressure)!=29 or any(c['status']!='PASS' for c in pressure):raise ValueError('actual pi differential/pressure cases differ')
    failures=[]
    for case in cases:
        if case['status']=='PASS':continue
        path=prefix+case['dir']+'/assertions.json';failures.append(dict(id=case['id'],status=case['status'],assertionsSHA256=pins[path]['sha256'],assertions=b.doc(m[path])))
    return dict(suite=index['suite'],sourceCommit=SOURCE,exitCode=1,cases=len(cases),statuses=statuses,auditFindings=0,PIDiff=868,pressure=29,
                artifactReferences=refs,failures=failures,archive=archive)


def load_index(b,label):
    index=b.doc(b.read(BASE,label+'-index.json',1<<20));expected='outer' if label=='verdict' else label
    if (index['sourceCommit']!=SOURCE or index['suite']!=expected or index['kind']!='completed-'+expected+'-original-failed' or index['archive']['file']!='.evidence/minimax-retained/ticket-13-final-gate/'+('verdict' if label=='verdict' else label)+'-raw.tar.gz' or
        index['archive']['prefix']!='routed-compiled-'+('verdict' if label=='verdict' else label) or index['originalsUnchangedAfterSnapshot'] is not True):raise ValueError('fixed final-gate capsule identity/path differs')
    return index


def verify(common):
    import capsule
    b=helper()
    for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        digest,name=line.split('  ',1)
        if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
    expected=b.doc(b.read(BASE,'summary.json',1<<20));index=load_index(b,'verdict');pins,m,archive=capsule.verify(b,BASE,common,index);actual=outer(b,pins,m,archive)
    if actual!=expected['outer']:raise ValueError('independent outer failed verdict differs')
    suites=[]
    for label in ('ordinary','race'):
        index=load_index(b,label);pins,m,archive=capsule.verify(b,BASE,common,index);suites.append(suite(b,index,pins,m,archive))
    result=dict(schema=1,sourceCommit=SOURCE,outer=actual,suites=suites,retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,releaseCertified=False,
                scope='Actual failed complete gate at8fbd; later offline repairs are separate and never recertify these original executions')
    if result!=expected:raise ValueError('independent complete gate summary differs')
    return result


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();v=verify(a.common_root)
    print(json.dumps(dict(verified=True,sourceCommit=SOURCE,releaseCertified=False,releasePass=False,actualNumericExits=EXITS,suites=[{k:x[k] for k in ('suite','cases','statuses','PIDiff','pressure','artifactReferences','archive')} for x in v['suites']],retentionCredentialsRead=False,retentionPaidCalls=0),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
