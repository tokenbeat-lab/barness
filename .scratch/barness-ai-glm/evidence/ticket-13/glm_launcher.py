#!/usr/bin/env python3
"""One fixed formal GLM route at the reviewed current source; one selected key."""
import argparse
import datetime as dt
import json
from pathlib import Path
import subprocess
import sys
import time
import zipfile
sys.dont_write_bytecode=True
import commands as c
import run as plan


def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('out');p.add_argument('route',choices=('chat','responses','messages'));a=p.parse_args()
 if Path(a.out).is_absolute() or '..' in Path(a.out).parts:p.error('prepared relative collection root required')
 round=plan.Round(a.out);round.guard();out=round.out
 freeze=c.doc(out/'source-at-start.json');review=c.doc(out/'prerelease-source-review.json')
 if review.get('sourceCommit')!=freeze['gitCommit'] or review.get('decision')!='APPROVED':
  raise RuntimeError('reviewed exact source required before collection')
 binary=out/'live.test';base=out/('glm-'+a.route);base.mkdir(exist_ok=False,mode=0o700)
 receipt=dict(schema=1,route=a.route,sourceCommit=freeze['gitCommit'],
  binarySHA256=c.digest(binary),sourceManifestSHA256=c.digest(c.ROOT/'ai/live/testdata/glm-candidate/sources.json'),
  command=['reviewed-live-binary','-test.run=^TestGLMFormal$','-test.timeout=1800s','-test.v'],
  maxCalls=24,maxReservedAndActualAttempts=32,plannedCalls=22,maxOutputTokens=4096,
  callTimeoutSeconds=120,processTimeoutSeconds=1800,automaticReruns=False,direct=True,
  sourceVariable='GLM_KEY',injectedCredential='BARNESS_GLM_KEY',credentialsInherited=False,
  accountAlias='local-glm',regionAlias='cn-standard-endpoint',accountRegionPermissionsBilling='unknown',
  startedAt=dt.datetime.now(dt.timezone.utc).isoformat(),state='STARTING')
 # Unique directory/intent is created before reading the selected credential.
 c.dump(base/'receipt.json',receipt)
 key=c.secret('GLM_KEY')
 env=c.ENV|{'BARNESS_AI_LIVE':'1','BARNESS_GLM_ROUTE':a.route,'BARNESS_GLM_FORMAL':'1',
  'BARNESS_GLM_KEY':key,'BARNESS_AI_EVIDENCE_DIR':str(base)}
 receipt['environmentKeys']=sorted(env);began=time.monotonic();process=None
 try:
  with (base/'process.log').open('wb') as output:
   process=subprocess.Popen([str(binary),*receipt['command'][1:]],cwd=c.ROOT/'ai/live',env=env,
    stdout=output,stderr=subprocess.STDOUT,start_new_session=True)
   try:
    code=process.wait(timeout=1800);receipt.update(state='COMPLETED',exitCode=code)
   except subprocess.TimeoutExpired:
    c.stop(process);receipt.update(state='TIMEOUT',exitCode=124)
 except BaseException as error:
  if process is not None:c.stop(process)
  receipt.update(state='INTERRUPTED',exitCode=None,exceptionType=type(error).__name__)
  raise
 finally:
  receipt.update(finishedAt=dt.datetime.now(dt.timezone.utc).isoformat(),seconds=time.monotonic()-began)
  findings=[]
  for file in base.rglob('*'):
   if file.is_file() and key.encode() in file.read_bytes():
    findings.append(dict(file=file.relative_to(base).as_posix(),rule='exact-selected-secret'))
  receipt['launcherAudit']=dict(passed=not findings,findings=findings)
  c.dump(base/'receipt.json',receipt);env.pop('BARNESS_GLM_KEY',None);key=''
 round.guard()
 inventory={file.relative_to(base).as_posix():c.digest(file) for file in sorted(base.rglob('*')) if file.is_file()}
 c.dump(base/'SHA256SUMS.json',inventory)
 archive=base.with_suffix('.zip')
 with zipfile.ZipFile(archive,'x',zipfile.ZIP_DEFLATED) as z:
  for file in sorted(base.rglob('*')):
   if file.is_file():z.write(file,file.relative_to(base))
 summary=dict(route=a.route,exitCode=receipt['exitCode'],archive=str(archive.relative_to(c.ROOT)),
  archiveSHA256=c.digest(archive),files=len(inventory),seconds=receipt['seconds'],auditPassed=not findings)
 c.dump(out/('glm-'+a.route+'-summary.json'),summary);print(json.dumps(summary))
 if receipt['exitCode']!=0 or findings:raise SystemExit(1)
if __name__=='__main__':main()
