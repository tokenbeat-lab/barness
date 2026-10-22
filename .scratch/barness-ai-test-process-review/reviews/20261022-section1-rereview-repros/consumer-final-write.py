import os,sys,pathlib,json,resource,shutil,subprocess
ROOT=pathlib.Path('/Users/cyber/RestoX/barness'); ART=pathlib.Path('/tmp/barness-section1-spec-rereview')
sys.path.insert(0,str(ROOT/'.scratch/barness-ai-test-process-review/tools/tests'))
from test_release import ReleaseRecordProcess
from live_fixture import live_bundle
from process_support import digest,write_json
ReleaseRecordProcess.setUpClass()
t=ReleaseRecordProcess('test_complete_record_passes_without_starting_verification');t.setUp()
try:
    bundle,stage,_=live_bundle(t,'typesafe-classifier',qualified=True)
    good=bundle.parent.parent/'consumer-good';bad=bundle.parent.parent/'consumer-fail'
    baseline=t.call(t.binary,'-live-bundle',bundle.relative_to(t.repo),'-out',good,expected=0)
    limit=len((good/'stdout.jsonl').read_bytes())+100
    def limited():resource.setrlimit(resource.RLIMIT_FSIZE,(limit,limit))
    result=t.call(t.binary,'-live-bundle',bundle.relative_to(t.repo),'-out',bad,child_setup=limited)
    summary={'actualExitCode':result.returncode,'savedExitCode':int((bad/'exit-code').read_text()),'actualStderr':result.stderr,'savedStderr':(bad/'stderr.log').read_text(),'fileSizeLimit':limit,'finalReceiptSize':(bad/'live-consumer.json').stat().st_size,'stdoutEqual':result.stdout==(bad/'stdout.jsonl').read_text()}
    (ART/'io-failure-actual.stdout').write_text(result.stdout);(ART/'io-failure-actual.stderr').write_text(result.stderr)
    write_json(ART/'io-failure-summary.json',summary)
    shutil.copytree(bad,ART/'io-failure-saved',dirs_exist_ok=True)
    print('IO_FAILURE',json.dumps(summary))
finally:
    t.doCleanups();ReleaseRecordProcess.tearDownClass()
