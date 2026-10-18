#!/usr/bin/env python3
"""Read-only interrupted launch/observer/preparation history; no release certificate."""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import stat
import sys
import types

BASE=Path(__file__).resolve().parent
HELPER_SHA='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'


def helper():
    path=BASE.parent/'implementation/verify.py'
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:
        info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER_SHA:
        raise ValueError('pinned bounded verifier differs')
    module=types.ModuleType('bounded_capsule');module.__file__=str(path)
    exec(compile(raw,str(path),'exec'),module.__dict__);module.BASE=BASE
    return module


SOURCE='8fbd17006677774368d4d77d924a8672fd205c51'
CURRENT_HELPER='a1ecf4a0b2285732aada1a9d7026f693f7c05cef5b73ddb21b1a2bfe13226d4a'
CURRENT_OBSERVER='aec940df7db842009d9c977142dc35f604022b380ce383af7ce2dd85383879d7'
COMPILED='c54a7d07f5780711af334a4369dcfd1dbb2b3b8bd440f09962b88831d11a906a'
PHASE_HELPER='f7c2b86676e7a46e1aabeb4c37300eaa75663060a23420d0d518f98ce04cc29f'
ALLOW=('plan.json','prepare-plan.json','frozen-inputs.json','build-public-releasegate-command.json','build-public-releasegate.log',
       'compiled-binary.json','releasegate','observer-bin/go','barness-routed-final-keyless-compiled.py','barness-routed-go-recorder.py','parent-review-receipt.json')


def analyse(b,index,pins,m,archive):
    def doc(name):return b.doc(m[name])
    def digest(name):return pins[name]['sha256']
    old=doc('interrupted/summary.json');command=doc('interrupted/actual-releasegate-command.json');note=doc('interrupted/instrumentation-interruption.json');old_plan=doc('interrupted/plan.json')
    if (old['sourceCommit']!=SOURCE or old['credentialsRead'] is not False or old['paidCalls']!=0 or old['gateExit']!=-15 or
        old['passVerdict'] is not False or old['releasePass'] is not None or old['releaseReportPresent'] is not False or old['automaticReruns'] is not False or
        command['state']!='COMPLETED' or command['exitCode']!=-15 or command['credentialsRead'] is not False or
        command['sourceVariable'] is not None or note['sourceCommit']!=SOURCE or note['paidCallsAdded']!=0 or note['productSourcesChanged'] is not False or
        note['originalReleaseReportPresent'] is not False or note['recordedInternalNumericExits']!=[] or note['postStopSurvivingOwnedPIDs']!=[]):
        raise ValueError('original interrupted no-certificate outcome differs')
    if (any(n=='interrupted/gate/release-report.json' or n.startswith('interrupted/gate-command-exits/') for n in pins) or
        any(n.startswith('interrupted/gate/race/') for n in pins)):
        raise ValueError('interrupted certificate, numeric exits or race output invented')
    if [(s['name'],s['exitCode']) for s in old['steps']][-1]!=('actual-releasegate',-15) or len(old['steps'])!=12 or any(s['exitCode']!=0 for s in old['steps'][:-1]):raise ValueError('original11 successful keyless steps differ')
    if old_plan['runnerSHA256']!=digest('interrupted/barness-routed-final-keyless.py') or old_plan['observerSHA256']!=CURRENT_OBSERVER or digest('interrupted/barness-routed-go-recorder.py')!=CURRENT_OBSERVER or digest('interrupted/observer-bin/go')!=CURRENT_OBSERVER:
        raise ValueError('original interrupted helper/observer bytes differ')
    for name in pins:
        if name.startswith('interrupted/') and name.count('/')==1 and name.endswith('-command.json'):
            if doc(name)['credentialsRead'] is not False:raise ValueError('interrupted keyless command scope differs')
    probe=doc('probe/summary.json');green=doc('probe/green-exits/vet.json')
    if (probe['sourceCommit']!=SOURCE or probe['credentialsRead'] is not False or probe['paidCalls']!=0 or probe['fullGateExecuted'] is not False or
        probe['passed'] is not True or probe['redGoRunActualVetExit']!=0 or probe['redObserverRecordPresent'] is not False or
        probe['greenDirectActualVetExit']!=0 or probe['greenObserverRecordPresent'] is not True or 'probe/red-exits/vet.json' in pins or
        green['state']!='COMPLETED' or type(green['exitCode']) is not int or green['exitCode']!=0 or green['credentialsRead'] is not False or
        green['argv']!=['/usr/local/go/bin/go','vet','./...'] or green['cwd']!='$REPO'):
        raise ValueError('actual RED bypass/GREEN numeric vet proof differs')
    for field,leaf in (('probeSourceSHA256','probe.go'),('controllerSHA256','barness-routed-observer-probe-check.py'),('compiledProbeSHA256','probe'),('greenObserverRecordSHA256','green-exits/vet.json'),
                       ('goToolchainbase.goSHA256','go-toolchain-base.go'),('goToolchainenv.goSHA256','go-toolchain-env.go')):
        if probe[field]!=digest('probe/'+leaf):raise ValueError('fixed probe or toolchain byte pin differs')
    if (b'\tenv = AppendPATH(env)' not in m['probe/go-toolchain-base.go'] or b'cfg.GOROOTbin+string(os.PathListSeparator)+path' not in m['probe/go-toolchain-env.go']):raise ValueError('observed installed Go prepend cause differs')
    for name in ('red-go-run','build-same-probe','green-compiled-direct','compiled-helper-plan-readback'):
        record=doc('probe/'+name+'-command.json')
        if record['state']!='COMPLETED' or record['exitCode']!=0 or record['credentialsRead'] is not False:raise ValueError('actual keyless probe command differs')
    location='$REPO/.evidence/minimax-routed-observer-check-1/'
    expected_probe={'red-go-run':['/usr/local/go/bin/go','run',location+'probe.go'],
        'build-same-probe':['/usr/local/go/bin/go','build','-o',location+'probe',location+'probe.go'],
        'green-compiled-direct':[location+'probe']}
    if any(doc('probe/'+name+'-command.json')['argv']!=argv for name,argv in expected_probe.items()):raise ValueError('same-source actual probe argv differs')
    probe_inventory=doc('probe/inventory.json')['files']
    for name,pin in probe_inventory.items():
        if pins.get('probe/'+name)!=pin:raise ValueError('original probe inventory differs')
    for leaf in ('actual-releasegate-command.json','summary.json','instrumentation-interruption.json'):
        if m['probe/interrupted-'+leaf]!=m['interrupted/'+leaf]:raise ValueError('probe original interruption copy differs')
    plan=doc('compiled-prepared/plan.json');binary=doc('compiled-prepared/compiled-binary.json');build=doc('compiled-prepared/build-public-releasegate-command.json');review=doc('compiled-prepared/parent-review-receipt.json')
    prepared={n.removeprefix('compiled-prepared/') for n in pins if n.startswith('compiled-prepared/')}
    if prepared!=set(ALLOW) or index['compiledPreparedAllowlist']!=list(ALLOW):raise ValueError('frozen prepared allowlist differs; live A2 forbidden')
    if (plan['sourceCommit']!=SOURCE or plan['credentialsRead'] is not False or plan['paidCalls']!=0 or plan['automaticReruns'] is not False or
        plan['runnerSHA256']!=CURRENT_HELPER or digest('compiled-prepared/barness-routed-final-keyless-compiled.py')!=CURRENT_HELPER or
        plan['observerSHA256']!=CURRENT_OBSERVER or digest('compiled-prepared/observer-bin/go')!=CURRENT_OBSERVER or
        digest('compiled-prepared/barness-routed-go-recorder.py')!=CURRENT_OBSERVER or plan['compiledCLISHA256']!=COMPILED or
        digest('compiled-prepared/releasegate')!=COMPILED or binary['sha256']!=COMPILED or binary['bytes']!=pins['compiled-prepared/releasegate']['bytes'] or
        plan['buildReceiptSHA256']!=digest('compiled-prepared/build-public-releasegate-command.json') or
        build['state']!='COMPLETED' or build['exitCode']!=0 or build['credentialsRead'] is not False or
        not plan['gate'][0].endswith('/minimax-routed-final-compiled-1/releasegate') or len(plan['checks'])!=8 or
        len(plan['canonicalActuals'])!=18 or plan['canonicalActuals']!=old_plan['canonicalActuals'] or plan['gate'].count('-live')!=17 or
        any(arg in ('-bundle','-race') for arg in plan['gate']) or plan['existingSupportAndMatrixReused'] is not True):
        raise ValueError('reviewed frozen direct compilation plan/binary differs')
    if build['argv']!=['/usr/local/go/bin/go','build','-o','$REPO/.evidence/minimax-routed-final-compiled-1/releasegate','./ai/release/cmd/releasegate']:raise ValueError('actual public CLI build argv differs')
    if doc('compiled-prepared/frozen-inputs.json')!=plan['originalPins'] or doc('probe/compiled-helper-plan-readback.log')!=plan:raise ValueError('actual prepared plan readback/input pins differ')
    if (review['planSHA256']!=digest('compiled-prepared/plan.json') or review['helperSHA256']!=CURRENT_HELPER or review['binarySHA256']!=COMPILED or
        review['credentialsRead'] is not False or review['paidCalls']!=0 or review['reviews']['standards']['decision']!='APPROVED' or review['reviews']['spec']['decision']!='PASS'):
        raise ValueError('independent prepared plan review pins differ')
    if (probe['compiledHelperSHA256']!=CURRENT_HELPER or probe['observerSHA256']!=CURRENT_OBSERVER or probe['realGoSHA256']!=plan['realGoSHA256'] or
        old_plan['realGoSHA256']!=plan['realGoSHA256']):raise ValueError('fixed Go/helper/observer generation differs')
    historical=doc('historical83/index.json')
    expected83={'barness-routed-final-keyless.py':'913d9f2d37dff05580c34b22b6e7d0b24983298df91c3023f81aed6cbbc0754f',
                'barness-routed-go-recorder.py':'0cd1d2832598dde2cefa202de1a3adda51477ebf5a64db177418ce7bbe37f2de',
                'barness-routed-run-phase.py':'489beb838df9b3734b2672b3f81163ab46809b23ca0232764c2fa1605f3e19e8'}
    if set(historical['programs'])!=set(expected83) or any(digest('historical83/'+n)!=s or historical['programs'][n]['SHA256']!=s for n,s in expected83.items()):raise ValueError('historical reviewed83 helper hash differs')
    if historical['programs']['barness-routed-final-keyless.py']['executed'] is not False or historical['programs']['barness-routed-run-phase.py']['executed'] is not False:raise ValueError('historical83 authority cannot be upgraded')
    if digest('run-phase/barness-routed-run-phase.py')!=PHASE_HELPER:raise ValueError('actual G/L parent helper differs')
    for phase,source in (('glm','51c34338cdcabd77625d6265409ab910ae08c8c9'),('legacy',SOURCE)):
        record=doc('run-phase/'+phase+'-parent-plan.json')
        if record['runnerSHA256']!=PHASE_HELPER or record['parentCredentialsRead'] is not False or record['automaticReruns'] is not False or record['sourceCommit']!=source:raise ValueError('historical actual phase helper/source pin differs')
    raw=[v for n,v in pins.items() if n.startswith('interrupted/gate/offline/')]
    if not raw or any(n.startswith('interrupted/gate/offline/') and len(n.split('/'))<=5 and n.endswith('/manifest.json') for n in pins):raise ValueError('incomplete ordinary main manifest boundary differs')
    return dict(schema=1,sourceCommit=SOURCE,retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,
        originalA=dict(gateExit=-15,passVerdict=False,releasePass=None,releaseReportPresent=False,numericInternalExits=[],postStopSurvivingOwnedPIDs=[],
                       partialOrdinaryFiles=len(raw),partialOrdinaryBytes=sum(v['bytes'] for v in raw),releaseCertified=False),
        probe=dict(redActualVetExit=0,redObserverRecordPresent=False,greenActualVetExit=0,greenNumericRecordPresent=True,fullGateExecuted=False),
        compiledPreparation=dict(sourceCommit=SOURCE,helperSHA256=CURRENT_HELPER,binarySHA256=COMPILED,canonicalActualPins=18,frozenInputPins=len(plan['originalPins']),
                                 checks=8,liveInputs=plan['gate'].count('-live'),completedGateCaptured=False),
        historical83HelperBytesOnly=True,actualPhaseHelperSourceSHA256=PHASE_HELPER,phasePaidWorkCertifiedHere=False,archive=archive,
        scope='Interrupted launch bytes, actual vet observer proof and frozen compilation only; no release certificate or completed A2 output')


def main():
    import capsule
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
    for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        digest,name=line.split('  ',1)
        if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
    index=b.doc(b.read(BASE,'index.json',1<<20))
    if index['kind']!='keyless-interrupted-launch-and-observer-history' or index['sourceCommit']!=SOURCE or index['archive']['file']!='.evidence/minimax-retained/ticket-13-keyless-launch-history/keyless-launch-history-raw.tar.gz' or index['archive']['prefix']!='keyless-launch-history':raise ValueError('fixed history capsule identity/path differs')
    pins,payload,archive=capsule.verify(b,BASE,a.common_root,index);summary=analyse(b,index,pins,payload,archive)
    if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent no-certificate history summary differs')
    print(json.dumps(dict(verified=True,retentionCredentialsRead=False,retentionPaidCalls=0,releaseCertified=False,archive=archive,
                         scope=summary['scope']),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
