#!/usr/bin/env python3
"""Numeric receipt for one unmodified-input current public factory construction."""
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
BASE=Path(__file__).resolve().parent;REPO=BASE.parents[5];sha=lambda v:hashlib.sha256(v).hexdigest()
def main():
 root=REPO/'.evidence/minimax-ticket17-resource-factory-2';out=root/'legitimate-driver';out.mkdir(exist_ok=False);producer=root/'green-index-size/source/.scratch/barness-ai-minimax/evidence/ticket-17/retention-fixes/resource-probe.py';destination=root/'legitimate-work';argv=[sys.executable,'-B',str(producer),'--common-root','/Users/cyber/RestoX/barness','--value','false','--out',str(destination)];before=sha(producer.read_bytes());start=time.monotonic()
 with (out/'stdout.log').open('xb') as stdout,(out/'stderr.log').open('xb') as stderr:run=subprocess.run(argv,stdout=stdout,stderr=stderr,timeout=240,start_new_session=True,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
 after=sha(producer.read_bytes());verification=json.loads((destination/'verification.json').read_bytes()) if (destination/'verification.json').exists() else None;construction=json.loads((destination/'construction.json').read_bytes()) if (destination/'construction.json').exists() else None;row=dict(argv=argv,exitCode=run.returncode,seconds=time.monotonic()-start,sourceBeforeSHA256=before,sourceAfterSHA256=after,sourceUnchanged=before==after,publicVerification=verification,construction=construction,credentialsRead=False,paidCalls=0,releaseCertified=False);(out/'receipt.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True)
 if run.returncode!=0 or before!=after or verification is None or type(verification['exitCode']) is not int or verification['exitCode']!=1:raise ValueError('normal factory must succeed while counterfeit resource verifier refuses')
if __name__=='__main__':main()
