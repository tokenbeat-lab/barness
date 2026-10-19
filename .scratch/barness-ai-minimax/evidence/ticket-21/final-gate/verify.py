"""Prospective public readback of the three fixed closed gate capsules; no capture."""
import argparse,hashlib,json,os,posixpath,re,stat,sys,types
from pathlib import Path
sys.dont_write_bytecode=True
SOURCE='78843eb99905e95c8fa534a75897e0dc8787b1da'
GATE='.evidence/minimax-current-visual-gate-1'
SPECS={'ordinary':GATE+'/gate/offline','race':GATE+'/gate/race','outer':GATE}
SMALL='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
FULL='3017cba1d4ac09e0a203c68ac768a4d508b861fa862aa04c9ca4550efebd9328'
CANDIDATE='5913236998b006a14f3e611e71f584e3c6c6c17859d8bf319d51655dce5fc899'
DEFAULT='5a144d8e3f1ab398dd54fcff8da039d0335dce720326e6db722b95d14b220754'
# Same literal closure authority as the writer; unresolved until actual outer closes.
CLOSED={'resultSHA256': '454b0f928e3d6b2cfbb87019ad3dbbd371326b4655110278c0869a37b72fc211',
 'planSHA256': 'b71df75d0c3adc0f09d73d33a058caef80e718cd4b29a0b8a0f33798edc95a33',
 'sourceSHA256': '319357cf2cb40a70414bb09fd9016b76bbc663170e23d930dba839a5fe96715d',
 'reviewSHA256': '55fc419067e2cc14790b893c39b75c2b0baf5a86c7d23526b65539083854ebf4',
 'outerCommandSHA256': 'f37ecc39aa553f30ce019715fb16c359bd7bb12b44e9c7f67f49ffd179cba175',
 'reportSHA256': 'badb24706cad2629d6553a878b4afdfcfdadb071397ce8153597eaa43bcf4a77',
 'receiptSHA256': {'vet': '584e0087226e83b21d3f59e298a24d0e40ee89fadcf9a83d90874fc5b3f8f394',
                   'vet-live': 'a3df502d79f8bc43923360652c20e38ac537eaa80ae587b92f1f9a3bc75dc31a',
                   'ordinary': 'd247f3b1063f24fe7ed10a215973b52cd29d8d04259dcf38963bd1a5276851d3',
                   'race': '56a16276f4812c0e8b1e7e083e97bf67533d766bff7f5fb56b9e2bed06324521'},
 'outerExit': 0,
 'numericExits': {'vet': 0, 'vet-live': 0, 'ordinary': 0, 'race': 0}}
# Filled only after one capture and independent original/capsule metadata review.
# Per label: indexBytes/indexSHA256, archiveBytes/archiveSHA256,
# inventoryCompressedSHA256/inventoryDecodedSHA256, originalScope.
CAPSULES={'ordinary': {'indexBytes': 1849, 'indexSHA256': '5958cab38d235dde0dc1d810945a815c5bd7f34ac265aeedda284ad11663f44f', 'archiveBytes': 134251096, 'archiveSHA256': '71db259c0f740ed3ce6d2f6abe8a14dae88c6cc3bb9b81ed183cfa832ab0a11d', 'inventoryCompressedSHA256': '5a186b8de080a9e6936fb50a14dbc7946a73e24ddd08da3829c5524d9f9aa8c3', 'inventoryDecodedSHA256': '8327375d3cc8c6992c95678fb49a726c93f0557fa8818eaf910c15f95502e6da', 'originalScope': {'entries': 160913, 'payloadBytes': 1350851478, 'regularFiles': 142534}}, 'race': {'indexBytes': 1838, 'indexSHA256': 'd6cd5d6a23bdc43e3d4dce510ff4746cc1345a4df48f40d52d94d7224b224b6f', 'archiveBytes': 133909347, 'archiveSHA256': 'f9195ec6e8655f34967eff2b073eac49e111a59994899cdc460e600b6c123b17', 'inventoryCompressedSHA256': 'ba42311a19cf1c04e5dc5da3e46099d04375e7febd005971e3e3ed345467461d', 'inventoryDecodedSHA256': 'bdbec89c26fbe554080afe75672d6dd0d9187ea57b7b73118cfbdf71ccbc740f', 'originalScope': {'entries': 160975, 'payloadBytes': 1346333368, 'regularFiles': 142587}}, 'outer': {'indexBytes': 2091, 'indexSHA256': '3683879afe7123e7e5ef7b896c07e80f01ad7ce4c04af5749d1aec012ba950ff', 'archiveBytes': 12543298, 'archiveSHA256': '1fa7deb1f2e1284030da02510655a58849dee9bc3ce5657221349e9d6f72ae96', 'inventoryCompressedSHA256': '79dfcaac3e8d5eb43756729ea57f2d87897c1ca16ac360a02188da0c9a433ec6', 'inventoryDecodedSHA256': '8009a47446b69b8acec20664773e1ecaf62261512eb73831c955e00e3072e8aa', 'originalScope': {'entries': 34, 'payloadBytes': 35333452, 'regularFiles': 29}}}
CONTROLLER='f21a42a8b7c8f2ec2674997be1257017028750e40b9917f7edfff118c2908c9d'
COMMANDS={'vet':('go-vet',['vet','./...']),'vet-live':('go-vet-live',['vet','-tags','live','./...']),'ordinary':('go-test',['test','-count=1','-timeout=20m','./...']),'race':('go-test-race',['test','-race','-count=1','-timeout=75m','./...'])}
LINES={'vet':'go vet ./...','vet-live':'go vet -tags live ./...','ordinary':'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -count=1 -timeout=20m ./...','race':'BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race -count=1 -timeout=75m ./...'}
LIVE={'openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','typesafe-classifier-jev-1.14.0','zhipu-chat','zhipu-responses','zhipu-messages','minimax-chat','minimax-responses','minimax-messages'}

def load(relative,pin,repository):
    path=repository/'.scratch/barness-ai-minimax/evidence/ticket-13'/relative
    with os.fdopen(os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK),'rb') as stream:
        before=os.fstat(stream.fileno())
        if not stat.S_ISREG(before.st_mode) or before.st_size>64<<10:raise ValueError('bounded fixed helper required')
        raw=stream.read((64<<10)+1)
    if len(raw)!=before.st_size or hashlib.sha256(raw).hexdigest()!=pin:raise ValueError('fixed helper differs')
    m=types.ModuleType('fixed_helper');m.__file__=str(path);exec(compile(raw,str(path),'exec'),m.__dict__);return m

def exact_int(value,expected=None):
    if type(value) is not int or value<0 or expected is not None and value!=expected:raise ValueError('exact nonnegative integer required')

def metadata(b,base,label):
    anchor=CAPSULES[label];raw=b.read(base/label,'index.json',1<<20)
    exact_int(anchor['indexBytes'])
    if len(raw)!=anchor['indexBytes'] or b.sha(raw)!=anchor['indexSHA256']:raise ValueError('independent original index anchor differs')
    cap=b.doc(raw)
    if cap['archive']['bytes']!=anchor['archiveBytes'] or cap['archive']['sha256']!=anchor['archiveSHA256'] or cap['inventory']['compressedSHA256']!=anchor['inventoryCompressedSHA256'] or cap['inventory']['decodedSHA256']!=anchor['inventoryDecodedSHA256'] or cap['originalScope']!=anchor['originalScope']:raise ValueError('independent original archive/inventory/scope anchor differs')
    exact_int(cap.get('schema'),1)
    if cap.get('kind')!='closed-current-visual-full-gate-custody' or cap.get('sourceCommit')!=SOURCE or cap.get('originalRoot')!=SPECS[label] or cap.get('originalMemberPrefix')!='original/':raise ValueError('fixed gate custody identity differs')
    if cap['archive']['file']!='.evidence/minimax-retained/ticket-21-final-gate-'+label+'/raw.tar.gz' or cap['archive']['prefix']!='current-visual-gate-'+label:raise ValueError('fixed original archive scope differs')
    exact_int(cap['archive']['bytes'])
    if cap['inventory']['file']!='inventory.json.gz' or cap['sourceSnapshot']['file']!='full-gate-source.json.gz' or cap['sourceSnapshot']['archiveMember']!='custody/full-gate-source.json':raise ValueError('fixed metadata filenames differ')
    for obj,key in [(cap['inventory'],'files'),(cap['inventory'],'decodedBytes'),(cap['sourceSnapshot'],'decodedBytes')]:exact_int(obj[key])
    if any(type(v) is not int for v in cap['limits'].values()):raise ValueError('typed fixed limits required')
    for key in ('regularFiles','payloadBytes','entries'):exact_int(cap['originalScope'][key])
    if cap.get('originalsUnchangedAfterSnapshot') is not True or cap.get('retentionCredentialsRead') is not False:raise ValueError('closed unchanged original provenance required')
    exact_int(cap.get('retentionPaidCalls'),0)
    return cap

def observation_policy(pins,label):
    required={'custody/full-gate-source.json','custody/closed-authority.json','custody/original-scope.json','custody/full-gate-plan.json','custody/full-gate-result.json'}
    if label in ('ordinary','race'):
        required.add('custody/native-command.json')
        for name in pins:
            if name.startswith('original/') and name.endswith('/manifest.json') and len(name.split('/'))==3:
                required.add(name)
                audit=name.removesuffix('manifest.json')+'audit.json'
                if audit in pins:required.add(audit)
    else:
        required.add('custody/retainer.py.txt')
        required.update('original/'+name for name in ('full-gate-source.json','full-gate-plan.json','full-gate-result.json','full-gate-source-review.json','full-releasegate-command.json'))
        required.update('original/gate-command-exits/'+name+'.json' for name in COMMANDS)
        required.update(name for name in pins if name in ('original/gate/release-report.json','original/gate/release-report.md','original/gate/matrix-candidate.json','original/full-gate-inputs.json'))
    return required

def one_manifest(b,full,pins,seen,name):
    manifest=b.doc(seen[name]);prefix=name.removesuffix('manifest.json')
    if manifest.get('git_commit')!=SOURCE or not isinstance(manifest.get('package'),str):raise ValueError('native suite package/recorded source differs')
    audit_path=prefix+'audit.json';audit=b.doc(seen[audit_path]) if audit_path in pins else None
    cases=manifest['cases']
    if not isinstance(cases,list) or len(cases)>full.LIMITS['files'] or audit is not None and not isinstance(audit.get('findings'),list):raise ValueError('bounded original suite/audit shape')
    counts={'PASS':0,'FAIL':0,'NOT_RUN':0};ids=set();refs=0;normalized=0
    for case in cases:
        identity=case['id'];status=case['status'];directory=case['dir'];artifacts=case.get('artifacts',{})
        if not isinstance(identity,str) or identity in ids or status not in counts or not isinstance(directory,str) or not isinstance(artifacts,dict):raise ValueError('unique original case/status/artifact shape')
        ids.add(identity);counts[status]+=1
        for leaf,pin in artifacts.items():
            if not isinstance(leaf,str) or any(p.startswith('/') or '..' in p.split('/') for p in (directory,leaf)):raise ValueError('escaping original reference')
            joined=prefix+directory+'/'+leaf;target=posixpath.normpath(joined);b.relative(target)
            if not target.startswith(prefix) or pins.get(target,{}).get('sha256')!=pin:raise ValueError('native filepath.Join artifact reference pin differs')
            normalized+=target!=joined;refs+=1
            if refs>full.LIMITS['entries']:raise ValueError('finite artifact reference bound')
    pidiff=[c for c in cases if c['id'].startswith('PIDIFF-')];pressure=[c for c in cases if 'pressure' in c['id'].lower()]
    fixed=[c for c in cases if re.match(r'^E08-(policy-pressure|admission-burst-pressure|mixed-pressure)-',c['id'])]
    return {'manifest':name,'manifestSHA256':b.sha(seen[name]),'package':manifest['package'],'cases':len(cases),'statuses':counts,'auditPresent':audit is not None,'auditFindings':audit['findings'] if audit is not None else None,'artifactReferences':refs,'normalizedArtifactReferencePaths':normalized,'truePIDiffCases':len(pidiff),'truePIDiffAllPass':all(c['status']=='PASS' for c in pidiff),'allNamedPressureCases':len(pressure),'fixedPressureRegexCases':len(fixed),'fixedPressureRegex':'^E08-(policy-pressure|admission-burst-pressure|mixed-pressure)-','namedPressureAllPass':all(c['status']=='PASS' for c in pressure),'failures':[c['id'] for c in cases if c['status']=='FAIL'],'notRun':[c['id'] for c in cases if c['status']=='NOT_RUN']}

def suite(b,full,pins,seen,label):
    main=[];siblings=[]
    for name in pins:
        if not (name.startswith('original/') and name.endswith('/manifest.json') and len(name.split('/'))==3):continue
        try:manifest=b.doc(seen[name])
        except (ValueError,TypeError) as error:
            siblings.append({'manifest':name,'parseError':type(error).__name__});continue
        if manifest.get('package')=='./ai/e2e':main.append(name)
        elif isinstance(manifest.get('package'),str):siblings.append(one_manifest(b,full,pins,seen,name))
        else:siblings.append({'manifest':name,'package':None,'notEvidenceRun':True})
    if len(main)!=1:raise ValueError('exact one package ./ai/e2e main manifest required')
    summary=one_manifest(b,full,pins,seen,main[0])
    if summary['auditPresent'] is not True:raise ValueError('main native audit missing')
    return {'scope':label,'main':summary,'siblings':siblings}

def outer(b,pins,seen):
    result=b.doc(seen['original/full-gate-result.json']);plan=b.doc(seen['original/full-gate-plan.json']);review=b.doc(seen['original/full-gate-source-review.json']);outer=b.doc(seen['original/full-releasegate-command.json'])
    for name,key in [('full-gate-result.json','resultSHA256'),('full-gate-plan.json','planSHA256'),('full-gate-source.json','sourceSHA256'),('full-gate-source-review.json','reviewSHA256'),('full-releasegate-command.json','outerCommandSHA256')]:
        if b.sha(seen['original/'+name])!=CLOSED[key]:raise ValueError('original closure authority pin differs')
    if result.get('state')!='COMPLETED' or result.get('sourceCommit')!=SOURCE or type(result.get('outerExit')) is not int or result['outerExit']!=CLOSED['outerExit'] or type(outer.get('exitCode')) is not int or outer['exitCode']!=result['outerExit'] or outer.get('state')!='COMPLETED':raise ValueError('actual numeric outer differs')
    if review.get('decision')!='APPROVED' or review.get('sourceCommit')!=SOURCE or review.get('planSHA256')!=CLOSED['planSHA256'] or plan.get('sourceCommit')!=SOURCE or plan.get('sourceManifestSHA256')!=CLOSED['sourceSHA256']:raise ValueError('approved original plan/source binding differs')
    if plan.get('controllerSHA256')!=CONTROLLER or pins.get('original/controller.py.txt',{}).get('sha256')!=CONTROLLER:raise ValueError('original executed controller differs')
    if plan.get('selectedMatrixSHA256')!=CANDIDATE or plan.get('preservedDefaultMatrixSHA256')!=DEFAULT or plan.get('adoptedOldBundles') is not False:raise ValueError('candidate/default/fresh-gate distinction differs')
    receipts={}
    for name,(_,argv) in COMMANDS.items():
        key='original/gate-command-exits/'+name+'.json';value=b.doc(seen[key])
        if b.sha(seen[key])!=CLOSED['receiptSHA256'][name] or value.get('state')!='COMPLETED' or type(value.get('exitCode')) is not int or value['exitCode']!=CLOSED['numericExits'][name] or value.get('argv')!=['/usr/local/go/bin/go',*argv] or value.get('cwd')!='$REPO' or value.get('credentialsRead') is not False or value!=result['commands'][name]:raise ValueError('exact numeric native command binding differs')
        receipts[name]=value
    inputs=b.doc(seen['original/full-gate-inputs.json'])
    if b.sha(seen['original/full-gate-inputs.json'])!=plan.get('inputsManifestSHA256') or set(inputs)!=LIVE|{'text-jev','text-luna','visual-jev','visual-decisions'}:raise ValueError('fixed21 input manifest differs')
    for row in inputs.values():
        b.relative(row['root'])
        if not isinstance(row['files'],dict) or len(row['files'])>32768:raise ValueError('bounded native input tree pin')
        for name,pin in row['files'].items():
            b.relative(name)
            if not isinstance(pin,str) or re.fullmatch('[a-f0-9]{64}',pin) is None:raise ValueError('native input file pin shape')
    report_present='original/gate/release-report.json' in pins
    report_raw=seen['original/gate/release-report.json'] if report_present else None
    if report_present!=(CLOSED['reportSHA256'] is not None) or report_present and b.sha(report_raw)!=CLOSED['reportSHA256']:raise ValueError('original canonical report/absence anchor differs')
    report=None;report_error=None;bindings=False
    if report_raw is not None:
        try:
            report=b.doc(report_raw);rows=report['commands']
            if type(report.get('pass')) is not bool or report.get('commit')!=SOURCE or not isinstance(rows,list) or len(rows)!=4:raise ValueError('public report source/pass/four-row shape')
            commands={v['id']:v for v in rows}
            if set(commands)!={v[0] for v in COMMANDS.values()}:raise ValueError('four report command identities')
            for name,(identity,_) in COMMANDS.items():
                row=commands[identity]
                if type(row.get('passed')) is not bool or row['passed']!=(receipts[name]['exitCode']==0) or row.get('line')!=LINES[name]:raise ValueError('native report command binding')
            bindings=True
        except (ValueError,TypeError,KeyError,AttributeError) as error:report_error={'type':type(error).__name__,'message':str(error)[:2048]}
    zero=all(v['exitCode']==0 for v in receipts.values());matrix=seen['original/gate/matrix-candidate.json'] if 'original/gate/matrix-candidate.json' in pins else None
    matrix_valid=matrix is not None and b.sha(matrix)==CANDIDATE
    certified=result['outerExit']==0 and report is not None and report.get('pass') is True and zero and bindings and matrix_valid and result.get('selectedSourceAndInputsUnchanged') is True and result.get('postGuardPassed') is True
    if result.get('certified') is not certified:raise ValueError('reported certification differs from numeric/postguard facts')
    return {'outerExit':result['outerExit'],'numericExits':{n:v['exitCode'] for n,v in receipts.items()},'reportPresent':report_raw is not None,'reportedReleasePass':report.get('pass') if report is not None else None,'reportParsingOrBindingError':report_error,'allFourSucceeded':zero,'commandBindingsValid':bindings,'candidateMatrixCopyValid':matrix_valid,'selectedSourceAndInputsUnchanged':result.get('selectedSourceAndInputsUnchanged'),'releaseCertifiedAtRecordedSource':certified,'defaultPublicationClaimed':False,'fullInputTrees':len(inputs)}

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,required=True);p.add_argument('--metadata-root',type=Path,default=Path(__file__).resolve().parent);p.add_argument('--repository-root',type=Path);a=p.parse_args()
    if CLOSED is None or CAPSULES is None:raise ValueError('DRAFT: closed/capsule authority unresolved; no capsule reads permitted')
    if set(CAPSULES)!=set(SPECS):raise ValueError('fixed three independently reviewed capsule anchors required')
    repository=a.repository_root
    if repository is None:
        installed=Path(__file__).resolve()
        if tuple(installed.parts[-6:-1])!=('.scratch','barness-ai-minimax','evidence','ticket-21','final-gate'):raise ValueError('installed tracked reader path or explicit repository root required')
        repository=installed.parents[5]
    b=load('implementation/verify.py',SMALL,repository);full=load('final-gate/capsule.py',FULL,repository)
    checked={};observed={};union=set();counts=0
    for label in SPECS:
        cap=metadata(b,a.metadata_root,label);pins,seen,result=full.verify(b,a.metadata_root/label,a.common_root,cap)
        required=observation_policy(pins,label)
        if set(cap['observe'])!=required or set(seen)!=required:raise ValueError('anchored-inventory observation policy differs')
        if b.sha(seen['custody/full-gate-plan.json'])!=CLOSED['planSHA256'] or b.sha(seen['custody/full-gate-result.json'])!=CLOSED['resultSHA256']:raise ValueError('suite/outer inline authority differs')
        if label in ('ordinary','race') and b.sha(seen['custody/native-command.json'])!=CLOSED['receiptSHA256'][label]:raise ValueError('suite inline native receipt differs')
        source=b.doc(seen['custody/full-gate-source.json'])
        if b.sha(seen['custody/full-gate-source.json'])!=CLOSED['sourceSHA256'] or source.get('sourceCommit')!=SOURCE or len(source.get('files',{}))!=1078 or b.doc(seen['custody/closed-authority.json'])!=CLOSED:raise ValueError('recorded1078 source authority differs')
        original={SPECS[label]+'/'+name.removeprefix('original/') for name in pins if name.startswith('original/')}
        if len(original)!=cap['originalScope']['regularFiles'] or sum(v['bytes'] for n,v in pins.items() if n.startswith('original/'))!=cap['originalScope']['payloadBytes'] or union&original:raise ValueError('disjoint complete regular original inventory differs')
        if label=='outer' and any(name.startswith((GATE+'/gate/offline/',GATE+'/gate/race/')) for name in original):raise ValueError('outer illegally duplicates a complete suite subtree')
        if b.doc(seen['custody/original-scope.json'])!=dict(label=label,root=SPECS[label],**cap['originalScope']):raise ValueError('anchored original scope projection differs')
        union|=original;counts+=len(original);observed[label]=(pins,seen)
        checked[label]={'archive':result,'originalFiles':len(original)}
        if label in ('ordinary','race'):checked[label]['suite']=suite(b,full,pins,seen,label)
        print(json.dumps({'scope':label,**checked[label]}),flush=True)
    outer_seen=observed['outer'][1]
    for label in ('ordinary','race'):
        seen=observed[label][1]
        for name in ('full-gate-source.json','full-gate-plan.json','full-gate-result.json'):
            if seen['custody/'+name]!=outer_seen['original/'+name]:raise ValueError('suite and outer original authority bytes differ')
        if seen['custody/native-command.json']!=outer_seen['original/gate-command-exits/'+label+'.json']:raise ValueError('suite and original numeric receipt bytes differ')
    if counts!=sum(CAPSULES[label]['originalScope']['regularFiles'] for label in SPECS):raise ValueError('fixed master original union differs')
    verdict=outer(b,*observed['outer'])
    print(json.dumps({'custodyVerified':True,'recordedSourceCommit':SOURCE,'capsules':checked,'disjointOriginalRegularFiles':counts,'gate':verdict,'publicationIncluded':False,'historyIncluded':False,'retentionPaidCalls':0,'retentionCredentialsRead':False,'capturedProgramsExecuted':False},indent=2))

if __name__=='__main__':main()
