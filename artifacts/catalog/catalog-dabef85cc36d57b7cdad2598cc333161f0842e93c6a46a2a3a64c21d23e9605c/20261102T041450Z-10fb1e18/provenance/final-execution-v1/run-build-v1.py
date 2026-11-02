#!/usr/bin/env python3
"""Capture bounded generation, build and vet against frozen ordinary inputs."""
import json,os,pathlib,subprocess,sys,time
root=pathlib.Path.cwd();sys.path.insert(0,str(root/'tools/verification'))
from receipt_files import sha,source_snapshot
from run_process import environment,isolate_go_home
out=root/'artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/build-v1';out.mkdir()
manifest=root/'artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/ordinary-inputs-v6.json'
inputs=json.loads(manifest.read_text()); inputs[str(manifest.relative_to(root))]=sha(manifest); script=pathlib.Path(__file__).resolve(); inputs[str(script.relative_to(root))]=sha(script); source=source_snapshot(root)
if source['commit']!='8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11' or any(sha(root/n)!=h for n,h in inputs.items()):raise SystemExit('frozen input mismatch')
env,_=environment('e2e');isolate_go_home(env,'e2e',out)
for name,argv in [('generate',['go','generate','./ai']),('build',['go','build','./ai/...','./catalog/...','./cmd/catalogctl','./cmd/cataloggen']),('vet',['go','vet','./ai/...','./catalog/...','./cmd/catalogctl','./cmd/cataloggen'])]:
 start=time.time();proc=subprocess.run(argv,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=1200);end=time.time()
 stdout=out/(name+'.stdout');stderr=out/(name+'.stderr');stdout.write_bytes(proc.stdout);stderr.write_bytes(proc.stderr)
 stable=source_snapshot(root)==source and all(sha(root/n)==h for n,h in inputs.items())
 record={'argv':argv,'exitCode':proc.returncode,'startedUnix':start,'finishedUnix':end,'elapsedSeconds':end-start,'commit':source['commit'],'sourceAndSelectedInputsStable':stable,'inputManifestSha256':sha(manifest),'stdoutSha256':sha(stdout),'stderrSha256':sha(stderr),'goenv':env['GOENV'],'goflags':env['GOFLAGS'],'home':'isolated','timeoutSeconds':1200}
 (out/(name+'.json')).write_text(json.dumps(record,indent=2)+'\n');print(name,proc.returncode,'stable',stable,flush=True)
 if proc.returncode!=0 or not stable:raise SystemExit(proc.returncode or 1)
