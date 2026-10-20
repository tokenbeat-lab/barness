#!/usr/bin/env python3
"""Explicit Host launcher. Never imports dotenv or reads another provider key."""
import argparse, hashlib, json, os, pathlib, re, subprocess, time, zipfile
ROOT=pathlib.Path('/Users/cyber/RestoX/barness')
SOURCE=pathlib.Path(__file__).resolve().parents[4]
OUT=ROOT/'.evidence/barness-ai-glm/ticket-12'
def sha(raw): return hashlib.sha256(raw).hexdigest()
def selected_key():
    path=ROOT/'.env'
    if not path.exists(): return ''
    with path.open() as stream:
        for line in stream:
            match=re.match(r'^\s*(?:export\s+)?GLM_KEY\s*=\s*(.*?)\s*$', line)
            if match:
                value=match.group(1)
                if len(value)>1 and value[0]==value[-1] and value[0] in "\"'": value=value[1:-1]
                return value
    return ''
def main():
    ap=argparse.ArgumentParser();ap.add_argument('--execute',action='store_true',required=True);ap.add_argument('--route',choices=['chat','responses','messages'],required=True);ap.add_argument('--binary',required=True);args=ap.parse_args()
    binary=pathlib.Path(args.binary).resolve();route=args.route
    OUT.mkdir(parents=True,exist_ok=True);stamp=time.strftime('%Y%m%dT%H%M%SZ',time.gmtime());base=OUT/(stamp+'-'+route);base.mkdir()
    key=selected_key()
    env={'PATH':'/usr/bin:/bin:/usr/sbin:/sbin','TMPDIR':'/tmp','BARNESS_AI_LIVE':'1','BARNESS_GLM_ROUTE':route,'BARNESS_GLM_FORMAL':'1','BARNESS_GLM_KEY':key,'BARNESS_AI_EVIDENCE_DIR':str(base)}
    command=[str(binary),'-test.run=^TestGLMFormal$','-test.timeout=1800s','-test.v']
    receipt={'schema':1,'route':route,'command':['formal-live-binary']+command[1:],'gitCommit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=SOURCE,text=True).strip(),'binarySHA256':sha(binary.read_bytes()),'sourceManifestSHA256':sha((SOURCE/'ai/live/testdata/glm-candidate/sources.json').read_bytes()),'startedAt':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),'accountAlias':'local-glm','regionAlias':'cn-standard-endpoint','accountRegionPermissionsBilling':'unknown','credentialSource':'only GLM_KEY selected by Host from root dotenv; one injected key; exact-secret absence audit'}
    start=time.monotonic()
    with (base/'process.log').open('wb') as out:
        try: result=subprocess.run(command,cwd=SOURCE/'ai/live',env=env,stdout=out,stderr=subprocess.STDOUT,timeout=1800);code=result.returncode
        except subprocess.TimeoutExpired: code=124
    receipt.update(exitCode=code,elapsedSeconds=round(time.monotonic()-start,3),finishedAt=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()))
    findings=[]
    if key:
        needle=key.encode()
        for file in base.rglob('*'):
            if file.is_file() and needle in file.read_bytes(): findings.append({'file':file.relative_to(base).as_posix(),'rule':'exact-selected-secret'})
    receipt['launcherAudit']={'passed':not findings,'findings':findings}
    (base/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
    inventory={file.relative_to(base).as_posix():sha(file.read_bytes()) for file in sorted(base.rglob('*')) if file.is_file()}
    (base/'SHA256SUMS.json').write_text(json.dumps(inventory,indent=2)+'\n')
    archive=base.with_suffix('.zip')
    with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED) as z:
        for file in sorted(base.rglob('*')):
            if file.is_file():z.write(file,file.relative_to(base))
    print(json.dumps({'route':route,'exitCode':code,'archive':str(archive),'archiveSHA256':sha(archive.read_bytes()),'launcherAuditPassed':not findings,'elapsedSeconds':receipt['elapsedSeconds']}))
    raise SystemExit(code if code else (1 if findings else 0))
if __name__=='__main__':main()
