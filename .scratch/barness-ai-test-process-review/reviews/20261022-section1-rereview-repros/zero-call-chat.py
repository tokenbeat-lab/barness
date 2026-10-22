import os,sys,pathlib,json,shutil
ROOT=pathlib.Path('/Users/cyber/RestoX/barness');ART=pathlib.Path('/tmp/barness-section1-spec-rereview')
sys.path.insert(0,str(ROOT/'.scratch/barness-ai-test-process-review/tools/tests'))
from test_release import ReleaseRecordProcess
from live_fixture import live_bundle
from process_support import digest,write_json
ReleaseRecordProcess.setUpClass()
t=ReleaseRecordProcess('test_complete_record_passes_without_starting_verification');t.setUp()
try:
    bundle,stage,report=live_bundle(t,'google-gemini',qualified=False)
    expected=['text-stream','text-complete','tool-round-trip','cancel-after-first-frame','reasoning-history','image']
    manifest=json.loads((bundle/'manifest.json').read_text());manifest['cases']=[]
    for c in list(bundle.iterdir()):
        if c.is_dir():shutil.rmtree(c)
    report.update(operation='chat',provider='google',api='google-generative-ai',model='nonexistent-model',endpoint='https://proxy.invalid',expected=expected,scenarios=[])
    for ident in expected:
        write_json(bundle/ident/'assertions.json',{'case_id':ident,'status':'PASS','assertions':[{'pass':True}]})
        manifest['cases'].append({'id':ident,'dir':ident,'status':'PASS','artifacts':{'assertions.json':digest(bundle/ident/'assertions.json')}})
        report['scenarios'].append({'id':ident,'caseId':ident,'model':'nonexistent-model','outcome':'PASS','calls':0,'attempts':0})
    manifest['package']='-tags live ./ai/live'
    write_json(bundle/'manifest.json',manifest);write_json(bundle/'live-report.json',report)
    receiptpath=bundle.parent.parent/'receipt.json';receipt=json.loads(receiptpath.read_text())
    receipt['bundles'][0].update(package=manifest['package'],manifestSHA256=digest(bundle/'manifest.json'))
    write_json(receiptpath,receipt);stage['receipt']['sha256']=digest(receiptpath);stage['requiredCases']=expected
    out=bundle.parent.parent/'consumer-chat'
    result=t.call(t.binary,'-live-bundle',bundle.relative_to(t.repo),'-out',out)
    proof=out/'live-consumer.json';stage['consumers']=[{'path':str(proof.relative_to(t.repo)),'sha256':digest(proof)}]
    gate=t.gate(0)
    summary={'consumerExit':result.returncode,'consumerPass':json.loads(proof.read_text())['pass'],'releasePass':gate['pass'],'model':report['model'],'endpoint':report['endpoint'],'budget':report['budget'],'calls':sum(s['calls'] for s in report['scenarios']),'attempts':sum(s['attempts'] for s in report['scenarios'])}
    write_json(ART/'zero-call-chat-summary.json',summary)
    shutil.copytree(t.repo,ART/'zero-call-chat-repo',ignore=shutil.ignore_patterns('.git'),dirs_exist_ok=True)
    print(json.dumps(summary))
finally:
    t.doCleanups();ReleaseRecordProcess.tearDownClass()
