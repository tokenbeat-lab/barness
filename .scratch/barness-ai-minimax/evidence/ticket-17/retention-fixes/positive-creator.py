#!/usr/bin/env python3
"""Reuse one isolated failed-before-write fixture for a normal bounded CLI success."""
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
BASE=Path(__file__).resolve().parent;REPO=BASE.parents[4];sha=lambda v:hashlib.sha256(v).hexdigest()
def main():
 out=REPO/'.evidence/minimax-ticket17-retention-fixes/positive-creator';out.mkdir(exist_ok=False);fixture=REPO/'.evidence/minimax-ticket17-retention-fixes/green-command-oversize-2/clone';front=fixture/'.scratch/barness-ai-minimax/evidence/ticket-17';original=REPO/'.evidence/minimax-ticket17/red-containers/command.json';private=fixture/'.evidence/minimax-ticket17/red-containers/command.json';private.unlink();os.link(original,private)
 before=dict(rawCommandSHA256=sha(original.read_bytes()),sourceProofSHA256=sha((BASE.parent/'source-proof.json').read_bytes()),originalIndexSHA256=sha((BASE.parent/'index.json').read_bytes()));common=out/'common';argv=[sys.executable,'-B',str(front/'freeze.py'),'--common-root',str(common)];start=time.monotonic()
 with (out/'stdout.log').open('xb') as stdout,(out/'stderr.log').open('xb') as stderr:run=subprocess.run(argv,stdout=stdout,stderr=stderr,timeout=240,env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ})
 after=dict(rawCommandSHA256=sha(original.read_bytes()),sourceProofSHA256=sha((BASE.parent/'source-proof.json').read_bytes()),originalIndexSHA256=sha((BASE.parent/'index.json').read_bytes()));index=json.loads((front/'index.json').read_bytes()) if (front/'index.json').exists() else None
 row=dict(argv=argv,exitCode=run.returncode,seconds=time.monotonic()-start,producerSHA256=sha((front/'freeze.py').read_bytes()),originalBefore=before,originalAfter=after,originalsUnchanged=before==after,privateOutput=index['archive'] if index else None,qualification='NON_AUTHORITATIVE_PRIVATE_REPLAY',credentialsRead=False,paidCalls=0);(out/'receipt.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True)
 if run.returncode!=0 or before!=after or index is None:raise ValueError('bounded legitimate producer must succeed')
if __name__=='__main__':main()
