#!/usr/bin/env python3
"""Bounded, read-only replay of ticket14 keyless repair history; no release certificate."""
import argparse
import collections
import gzip
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import types

BASE=Path(__file__).resolve().parent
SOURCE='c2604db6f693f5c7d3a82c656512004ec85d2249'
OLD='803e5aa6b091bdb7557a68c824befe01289428ea688e0c58e6d9e2d5ca79c225'
NEW='f1de4ce6b8a81f08b0ce27c56bc01f6c8fae9e8a8ca51c114c44d5a55acd21b6'
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
CONTRACT=['go','test','./ai/e2e','-count=1','-run','^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract)$']
LINKS={
 'preflight-1/no-key-live/20261018T090917.114975000Z/formal-corrupt-symlink/formal-integrity.json':'7dd1e13c1f70db629337af0be1c5faa634208efd1ad82f8e63ce273e94138fc6',
 'preflight-2/no-key-live/20261018T091324.853082000Z/formal-corrupt-symlink/formal-integrity.json':'f0affa18293a736b78c9251711eb409c15b8a8a00d90ca6a79bdc5bfabafc50f'}


def helper():
    path=BASE.parents[1]/'ticket-13/implementation/verify.py'
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:raise ValueError('pinned bounded verifier differs')
    module=types.ModuleType('bounded_repair');module.__file__=str(path)
    exec(compile(raw,str(path),'exec'),module.__dict__);module.BASE=BASE;return module


def normalized(b,raw):
    value=b.doc(raw);items=[]
    for item in value['items']:
        row={'id':item['id'],'scenarios':sorted(item['scenarios']) if item.get('scenarios') is not None else None,'cases':item.get('cases','')}
        if item.get('requiredCases'):row['requiredCases']=sorted(item['requiredCases'])
        items.append(row)
    value2={'scenarios':dict(sorted(value['scenarios'].items())),'items':sorted(items,key=lambda item:item['id'])}
    data=json.dumps(value2,ensure_ascii=False,separators=(',',':')).replace('&','\\u0026').replace('<','\\u003c').replace('>','\\u003e').replace('\u2028','\\u2028').replace('\u2029','\\u2029').encode()
    return b.sha(data),value


def expanded(raw,limit):
    with gzip.GzipFile(fileobj=io.BytesIO(raw)) as stream:data=stream.read(limit+1)
    if len(data)>limit:raise ValueError('fixed inner gzip byte limit')
    return data


def analyse(b,index,pins,m,archive):
    def doc(name):return b.doc(m[name])
    def pin(name):return pins[name]['sha256']
    def command(name,code,argv=None):
        v=doc(name)
        if (v['state']!='COMPLETED' or type(v['exitCode']) is not int or v['exitCode']!=code or v['credentialsRead'] is not False or
            v['credentialsInherited'] is not False or v['sourceVariable'] is not None or v['injectedCredential'] is not None or
            v['launcherAudit']!={'passed':True,'findings':[]} or any('KEY' in k or 'TOKEN' in k or 'SECRET' in k for k in v['environmentKeys'])):raise ValueError('original keyless command outcome differs')
        if argv is not None and v['argv']!=argv:raise ValueError('original public argv differs')
        return v
    proof=doc('tdd/mapping-delta-proof.json');before_raw=expanded(m['tdd/trace-before-ad204.json.gz'],8<<20);after_raw=expanded(m['tdd/trace-after-ad204.json.gz'],8<<20)
    old,before=normalized(b,before_raw);new,after=normalized(b,after_raw)
    if (old!=OLD or new!=NEW or proof['beforeMappingSHA256']!=old or proof['afterMappingSHA256']!=new or
        b.sha(before_raw)!=proof['beforeTraceSHA256'] or b.sha(after_raw)!=proof['afterTraceSHA256'] or
        len(before['items'])!=394 or len(after['items'])!=394 or
        {k:v for k,v in before.items() if k!='items'}!={k:v for k,v in after.items() if k!='items'}):raise ValueError('normalized mapping proof differs')
    changes=[]
    for a,z in zip(before['items'],after['items']):
        if a['id']!=z['id']:raise ValueError('trace item identity changed')
        if a==z:continue
        if ({k for k in set(a)|set(z) if a.get(k)!=z.get(k)}!={'cases','requiredCases'} or
            z['cases']!=a['cases'].replace(re.escape('PIDIFF-MM06-frozen-catalog'),re.escape('MM06-frozen-catalog-metadata')) or
            len(z['requiredCases'])!=len(a['requiredCases']) or sorted(z['requiredCases'])!=sorted(x.replace('PIDIFF-MM06-frozen-catalog','MM06-frozen-catalog-metadata') for x in a['requiredCases'])):raise ValueError('trace delta is not exact metadata rename')
        changes.append(a['id'])
    if changes!=['MM-R24','MM-T03','MM-A19','MM-US68','MM-US71','MM-US72','MM-US73'] or [p['id'] for p in proof['changedItems']]!=changes:raise ValueError('seven mapping delta identities differ')
    if (m['source/plan-before.go'].replace(OLD.encode(),NEW.encode())!=m['source/plan-after.go'] or m['source/plan-before.go'].count(OLD.encode())!=1 or
        m['source/plan-before.go']!=m['tdd/plan-before.go'] or m['source/plan-after.go']!=m['tdd/plan-after.go']):raise ValueError('one-line approved mapping hash change differs')
    original=doc('tdd/source-before.json');slice_source=doc('tdd/source-after-mapping-slice.json');final=doc('preflight-2/source-at-start.json')
    if final['sourceCommit']!=SOURCE or len(final['files'])!=1047 or set(original['files'])!=set(final['files']):raise ValueError('fixed source set differs')
    changed=sorted(n for n in final['files'] if original['files'][n]!=final['files'][n])
    expected=sorted(['ai/internal/testkit/release/plan.go','.scratch/barness-ai-minimax/tools/routed_validation.py','.scratch/barness-ai-minimax/tools/test_routed_validation.py'])
    if changed!=expected or [n for n in original['files'] if original['files'][n]!=slice_source['files'][n]]!=['ai/internal/testkit/release/plan.go']:raise ValueError('minimal source delta differs')
    for leaf,name in [('plan-after.go',expected[2]),('routed-after.py',expected[0]),('test-routed-after.py',expected[1])]:
        if pin('source/'+leaf)!=final['files'][name]:raise ValueError('retained final source body differs')
    if b.sha(after_raw)!=final['files']['ai/release/traceability.json']:raise ValueError('trace changed after metadata rename')
    if final!=doc('preflight-1/source-at-start.json'):raise ValueError('check-only attempts used different code')
    for n in ('tools-all','race-public-contracts','vet','vet-live'):
        if doc('tdd/source-'+n+'.json')['files']!=final['files']:raise ValueError('GREEN checks differ from committed final selected source')
    source_red=doc('tdd/source-preflight-red.json');source_green=doc('tdd/source-preflight-green.json')
    if [n for n in source_red['files'] if source_red['files'][n]!=source_green['files'][n]]!=['.scratch/barness-ai-minimax/tools/routed_validation.py']:raise ValueError('preflight RED/GREEN delta differs')
    if source_red['files']['.scratch/barness-ai-minimax/tools/routed_validation.py']!=pin('source/routed-before.py') or source_green['files']!=final['files']:raise ValueError('preflight source slice differs')
    for name,code in [('red-public-contracts',1),('green-public-contracts',0),('race-public-contracts',0)]:
        argv=CONTRACT if not name.startswith('race') else CONTRACT[:2]+['-race']+CONTRACT[2:];command('tdd/'+name+'-command.json',code,argv)
    for name,code in [('red-release-contract-preflight',1),('green-release-contract-preflight',0),('tools-all',0),('vet',0),('vet-live',0)]:command('tdd/'+name+'-command.json',code)
    for name,tests,passed in [('red-release-contract-preflight',2,False),('green-release-contract-preflight',2,True),('tools-all',60,True)]:
        raw=m['tdd/'+name+'.log']
        if re.search(rb'Ran '+str(tests).encode()+rb' tests in [0-9.]+s\s+'+(rb'OK\s*$' if passed else rb'FAILED \(failures=2\)\s*$'),raw) is None:raise ValueError('public CLI TDD test count/outcome differs')
    if b'existing public release contracts never executed' not in m['tdd/red-release-contract-preflight.log'] or b'"no-key-visual-proxy": 0' not in m['tdd/red-release-contract-preflight.log']:raise ValueError('original preflight omission RED differs')
    if m['tdd/vet.log'] or m['tdd/vet-live.log']:raise ValueError('dual vet GREEN log differs')
    checknames=['vet','vet-live','no-key-live','no-key-visual-proxy','no-key-release-contracts'];first=doc('preflight-1/checks.json');second=doc('preflight-2/checks.json')
    if first!={'passed':False,'commands':dict(zip(checknames[:3],[0,0,1])),'credentialsRead':False} or second!={'passed':True,'commands':dict.fromkeys(checknames,0),'credentialsRead':False}:raise ValueError('two actual preflight outcomes differ')
    for attempt,checks in [(1,first),(2,second)]:
        prefix='preflight-'+str(attempt)+'/'
        for name,code in checks['commands'].items():command(prefix+name+'-command.json',code,CONTRACT if name=='no-key-release-contracts' else None)
        for name in ('execution-authorization.template.json','prerelease-source-review.template.json'):
            v=doc(prefix+name)
            if v['decision']!='PENDING' or v['sourceCommit']!=SOURCE:raise ValueError('diagnostic-only pending authorization differs')
        if any(n==prefix+'execution-authorization.json' or n==prefix+'prerelease-source-review.json' or n.startswith(prefix+'intents/') for n in pins):raise ValueError('diagnostic-only capsule acquired paid approval or intent')
    if b'no such file or directory' not in m['preflight-1/no-key-live.log'] or b'TestGLMCaptureReplayProcess' not in m['preflight-1/no-key-live.log']:raise ValueError('original missing offline fixture failure differs')
    restore=doc('tdd/offline-fixture-restoration.json')
    if (len(restore['files'])!=6 or restore['sourceCommit']!=SOURCE or restore['sourceAndSelectorsUnchanged'] is not True or
        restore['actualProviderRequestsReplayed'] is not False or restore['newCapturesPublished'] is not False or restore['paidCalls']!=0 or restore['credentialsRead'] is not False):raise ValueError('exact historical ZIP restoration scope differs')
    for item in restore['files']:
        name='fixtures/'+item['archive'].removeprefix('.evidence/')
        selector_name='fixtures/'+Path(item['selector']).parent.name+'-selector.json';selector=doc(selector_name)
        route=selector['routes'][item['route']]
        if pin(name)!=item['sha256'] or pins[name]['bytes']!=item['bytes'] or route['archiveSHA256']!=item['sha256'] or route['archive']!=item['archive'] or pin(selector_name)!=restore['selectors'][item['selector']]:raise ValueError('historical restored ZIP or committed selector differs')
    links=doc('negative-fixture-links.json')
    if links['type']!='symlink-metadata-only-no-follow-no-extract' or links['links']!=index['negativeFixtureLinkMetadata'] or set(links['links'])!=set(LINKS):raise ValueError('negative-fixture metadata projection differs')
    for name,row in links['links'].items():
        if name in pins or row['sha256']!=LINKS[name] or b.sha(row['target'].encode())!=row['sha256'] or len(row['target'].encode())!=row['bytes'] or row['bytes']!=158:raise ValueError('negative-fixture link target differs')
    observed=doc('tdd/actual-five-check-observed-2.json')
    if (observed['sourceCommit']!=SOURCE or observed['checks']!=second or observed['authorization']!='PENDING' or observed['checkOnly'] is not True or
        observed['newRoundCreated'] is not False or observed['original18Unchanged'] is not True or len(observed['original18CanonicalPins'])!=18 or
        observed['credentialsRead'] is not False or observed['paidCalls']!=0):raise ValueError('closed original18/no new paid round proof differs')
    for key,leaf in [('sourceManifestSHA256','source-at-start.json'),('binariesManifestSHA256','binaries.json'),('planSHA256','plan.json'),('checksSHA256','checks.json')]:
        if observed[key]!=pin('preflight-2/'+leaf):raise ValueError('actual5 evidence pin differs')
    for leaf in ('mapping-delta-proof.json','offline-fixture-restoration.json','actual-five-check-observed-2.json'):
        if b.read(BASE,leaf,1<<20)!=m['tdd/'+leaf]:raise ValueError('frontdoor proof differs from original bytes')
    external=b.doc(b.read(BASE,'external-references.json',1<<20))
    if len(external['binaries'])!=6 or external['codeCommit']!=SOURCE or external['retentionCredentialsRead'] is not False or external['retentionPaidCalls']!=0:raise ValueError('unused prepared binary references differ')
    for row in external['binaries']:
        match=re.fullmatch(r'\$REPO/\.evidence/minimax-ticket14-check-only-([12])/(live.test|chineseeval|visualeval)',row['location'])
        if not match or row['copied'] is not False or row['bytes']>64<<20:raise ValueError('unused binary reference scope differs')
        prefix='preflight-'+match[1]+'/'
        if doc(prefix+'binaries.json')[match[2]]!=row['sha256'] or prefix+match[2] in pins:raise ValueError('unused binary hash or raw exclusion differs')
    manifests=[];references=0
    for name in sorted(m):
        if not name.endswith('/manifest.json'):continue
        value=doc(name);root=name.rsplit('/',1)[0]
        audit=doc(root+'/audit.json') if root+'/audit.json' in m else None
        if audit is None and (len(value['cases'])!=1 or value['cases'][0]['id'] not in ('P08-replay-selection','P09-live-replay-command')):raise ValueError('missing native main audit')
        if audit is not None and audit['findings']:raise ValueError('native E2E audit findings')
        statuses=dict(collections.Counter(c['status'] for c in value['cases']))
        for case in value['cases']:
            directory=PurePosixPath(case['dir'])
            if directory.is_absolute() or '..' in directory.parts:raise ValueError('unsafe native E2E case directory')
            for leaf,digest in case['artifacts'].items():
                b.relative(leaf);member=root+'/'+str(directory/leaf)
                if pins.get(member,{}).get('sha256')!=digest:raise ValueError('native E2E artifact pin differs')
                references+=1
        manifests.append(dict(member=name,cases=len(value['cases']),statuses=statuses,gitCommit=value['git_commit'],auditFindings=0 if audit is not None else None,auditPresent=audit is not None))
    for kind,want in [('red',{'FAIL':2,'PASS':4}),('green',{'PASS':6}),('race',{'PASS':6})]:
        chosen=[v for v in manifests if v['member'].startswith('tdd/'+kind+'/')]
        if len(chosen)!=1 or chosen[0]['statuses']!=want:raise ValueError('native six-case RED/GREEN contract proof differs')
    visual=[v for v in manifests if v['member']==index['visualProxyRoot']+'/manifest.json']
    fifth=[v for v in manifests if v['member'].startswith('preflight-2/no-key-release-contracts/')]
    if len(visual)!=1 or visual[0]['statuses']!={'PASS':16} or visual[0]['gitCommit']!=SOURCE or len(fifth)!=1 or fifth[0]['statuses']!={'PASS':6} or fifth[0]['gitCommit']!=SOURCE:raise ValueError('fresh actual public16 visual/6 contract cases differ')
    return dict(schema=1,codeCommit=SOURCE,selectedFiles=1047,sourceFilesChanged=changed,mappingDeltaIDs=changes,
        publicContractRED={'exitCode':1,'cases':6,'PASS':4,'FAIL':2},publicContractGREEN={'ordinaryExit':0,'raceExit':0,'casesEach':6,'PASS':6},
        publicPreflightCLI={'redExit':1,'redFailures':2,'greenExit':0,'greenTests':2},toolsGREEN={'exitCode':0,'tests':60},dualVet=[0,0],
        actualPreflightFirst=first,actualPreflightRestored=second,restoredOfflineZIPs=6,restoredZIPBytes=sum(v['bytes'] for v in restore['files']),
        nativeManifests=manifests,artifactReferencesVerified=references,negativeFixtureLinksMetadataOnly=2,unusedBinariesHashReferences=6,
        authorization='PENDING',newRoundCreated=False,original18Unchanged=True,
        retentionCredentialsRead=False,retentionPaidCalls=0,capturedCommandsExecuted=False,releaseCertified=False,archive=archive,
        scope='Offline mapping hash and preflight repair only; original paid failures and frozen source expiry remain; no new paid authority')


def main():
    import capsule
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
    for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        digest,name=line.split('  ',1)
        if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
    index=b.doc(b.read(BASE,'index.json',1<<20))
    if (index['kind']!='offline-mapping-pin-and-routed-preflight-repair' or index['sourceCommit']!=SOURCE or index['codeCommit']!=SOURCE or
        index['archive']['file']!='.evidence/minimax-retained/ticket-14-release-mapping-pin/ticket14-repair-raw.tar.gz' or index['archive']['prefix']!='release-mapping-repair'):raise ValueError('fixed repair capsule identity/path differs')
    pins,m,archive=capsule.verify(b,BASE,a.common_root,index);summary=analyse(b,index,pins,m,archive)
    if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent repair summary differs')
    print(json.dumps(dict(verified=True,codeCommit=SOURCE,retentionCredentialsRead=False,retentionPaidCalls=0,releaseCertified=False,archive=archive,scope=summary['scope']),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
