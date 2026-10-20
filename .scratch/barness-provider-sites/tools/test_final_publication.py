#!/usr/bin/env python3
"""Public new-round GLM publication refuses every closed canonical fallback."""
import json
from pathlib import Path
import os
import shutil
import subprocess
import sys
import unittest

HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE.parents[2]/'.scratch/barness-ai-minimax/tools'))
import test_renewal

class FinalPublication(unittest.TestCase):
    def setUp(self):
        self.fixture=test_renewal.Renewal()
        self.fixture.setUp()
        self.root=self.fixture.root
        shutil.copytree(HERE,self.root/'.scratch/barness-provider-sites/tools',ignore=shutil.ignore_patterns('__pycache__'))
        self.fixture.commit()
        self.journal=[]
    def tearDown(self):
        target=os.environ.get('BARNESS_OPERATOR_TEST_EVIDENCE')
        if target:
            out=Path(target)/self._testMethodName;out.mkdir(parents=True,exist_ok=True)
            (out/'processes.json').write_text(json.dumps(self.journal,indent=2)+'\n')
        self.fixture.tearDown()
    def invoke(self,*argv,ok):
        p=subprocess.run([sys.executable,'-B',str(self.fixture.tools/'renewal.py'),*argv],cwd=self.root,env=self.fixture.env,text=True,capture_output=True,timeout=30)
        self.journal.append(dict(argv=list(argv),exitCode=p.returncode,stdout=p.stdout.replace(str(self.root),'$CONTROLLED_REPO'),stderr=p.stderr.replace(str(self.root),'$CONTROLLED_REPO')))
        self.assertEqual(p.returncode==0,ok,p.stderr[-900:]);return p
    def seed_new(self):
        values=self.fixture.captures()
        code=r"""
import sys,shutil
from pathlib import Path
sys.path.insert(0,'.scratch/barness-provider-sites/tools')
import final_contract as r,final_state as s,final_execute as e
c=r.c;out=c.ROOT/'.evidence/provider-sites/glm-native';out.mkdir(parents=True)
commit=c.git('rev-parse','HEAD');jobs={n:j for n,j in r.jobs(out).items() if j['phase']=='glm'}
for name,job in jobs.items():
 source=c.ROOT/'raw'/job['glmRoute']/'session';target=out/name/'session';shutil.copytree(source,target)
 report=c.doc(target/'formal-glm-report.json');report['budget']=dict(callsUsed=22,httpAttempts=22);c.dump(target/'formal-glm-report.json',report)
c.dump(out/'source-at-start.json',dict(sourceCommit=commit,files={}))
c.dump(out/'binaries.json',{})
c.dump(out/'plan.json',c.redact(r.inventory(out)|dict(phase='glm',routes=jobs)))
prepared=dict(schema=1,round=r.ROUND,phase='glm',credentialsRead=False,sourceCommit=commit,sourceManifestSHA256=c.digest(out/'source-at-start.json'),binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'))
c.dump(out/'prepared.json',prepared)
proofs={}
for name in ('standardsReview','specReview'):
 path=out/(name+'.md');path.write_text('Synthetic independent source review.\n');proofs[name]=dict(path=str(path.relative_to(c.ROOT)),sha256=c.digest(path))
c.dump(out/'prerelease-source-review.json',prepared|dict(decision='APPROVED',reviewedBy='synthetic independent reviewers',authorizedRoutes=list(jobs),**proofs))
c.dump(out/'execution-authorization.json',prepared|dict(decision='APPROVED',requestText='Synthetic exact fixed726/762 US30 standard',requestReference='synthetic own new reply',maxCalls=726,maxAttempts=762,maxUSD='30',automaticProcessReruns=False,standardCredentialScopes={'ZHIPU_KEY':dict(provider='zhipu',purpose='standard',confirmed=True,accountAlias='own-cn',regionAlias='cn-standard')},networkRoutes={n:'direct' for n in jobs},processReservationsMicrousd={n:1000000 for n in jobs},diagnosticReservedMicrousd=200000))
checks=dict(passed=True,commands={},commandReceipts={},credentialsRead=False)
for name,argv in dict(vet=['go','vet','./...'],**{'vet-live':['go','vet','-tags','live','./...'],'no-key-live':['go','test','-tags','live','-count=1','-timeout=20m','./ai/live']}).items():
 c.dump(out/(name+'-command.json'),dict(argv=argv,state='COMPLETED',exitCode=0,credentialsRead=False,credentialsInherited=False));checks['commands'][name]=0;checks['commandReceipts'][name]=c.digest(out/(name+'-command.json'))
c.dump(out/'checks.json',checks)
for name,job in jobs.items():
 base=out/name/'session';c.dump(out/(name+'-command.json'),c.redact(dict(argv=job['argv'],state='COMPLETED',exitCode=0,credentialsRead=True)))
 c.dump(out/(name+'-consume-command.json'),dict(state='COMPLETED',exitCode=0,credentialsRead=False))
 pins=dict(schema=1,**{key:c.digest(base/leaf) for key,leaf in [('formalReportSha256','formal-glm-report.json'),('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),('executionSourceSha256','glm-execution-source.json')]})
 record=dict(state='COMPLETED',route=r.ROUND+'-'+name,baseName=name,validationRound=r.ROUND,phase='glm',sourceCommit=commit,sourceVariable='ZHIPU_KEY',networkRoute='direct',automaticReruns=False,maxCalls=24,maxAttempts=32,credentialsRead=True,exitCode=0,launcherAudit=dict(passed=True),collectionRoot=str(out),sourceManifestSha256=prepared['sourceManifestSHA256'],binariesManifestSha256=prepared['binariesManifestSHA256'],planSha256=prepared['planSHA256'],sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),authorizationSha256=c.digest(out/'execution-authorization.json'),commandReceiptSha256=c.digest(out/(name+'-command.json')),nativeConsumerReceiptSha256=c.digest(out/(name+'-consume-command.json')),capturePins=pins,budget=dict(state='REPORTED',status='PASS',calls=22,attempts=22))
 c.dump(e.marker(name),c.redact(record))
"""
        p=subprocess.run([sys.executable,'-B','-c',code],cwd=self.root,env=self.fixture.env,text=True,capture_output=True,timeout=20)
        self.assertEqual(p.returncode,0,p.stderr)
        return {route:'.evidence/provider-sites/glm-native/glm-'+route for route in values}

    def test_full_new_canonical_capture_publishes_new_archives_and_rejects_pin_changes(self):
        values=self.seed_new()
        args=['glm-archive','.evidence/minimax-provider-sites-own','--round','provider-sites-final-acceptance-1']
        for route,value in values.items():args+=['--'+route,value]
        self.invoke(*args,ok=True)
        target=self.root/'ai/live/testdata/glm-formal/live-captures.json';before=target.read_bytes()
        index=json.loads(before)
        self.assertEqual(index['collectionRound'],'provider-sites-final-acceptance-1')
        self.assertEqual(set(index['routes']),{'chat','responses','messages'})
        marker=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-1-glm-messages.json'
        value=json.loads(marker.read_text());value['sourceVariable']='GLM_KEY';marker.write_text(json.dumps(value))
        args[1]='.evidence/minimax-provider-sites-borrowed-key'
        self.invoke(*args,ok=False)
        self.assertEqual(target.read_bytes(),before)

    def test_new_round_never_borrows_original_or_closed_round_capture_receipts(self):
        values=self.fixture.captures()
        args=['glm-archive','.evidence/minimax-provider-sites-new','--round','provider-sites-final-acceptance-1']
        for route,value in values.items():args+=['--'+route,value]
        before=(self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes()
        result=self.invoke(*args,ok=False)
        self.assertNotIn('invalid choice',result.stderr)
        self.assertIn('provider-sites-final-acceptance-1-glm-chat.json',result.stderr)
        self.assertEqual((self.root/'ai/live/testdata/glm-formal/live-captures.json').read_bytes(),before)

if __name__=='__main__':unittest.main()
