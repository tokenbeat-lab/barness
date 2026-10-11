#!/usr/bin/env python3
"""Nine isolated live processes. Each report is merged only into its own row."""
import json,os,subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[3]
out=root/'.evidence/barness-ai-thinking-sampling/current-release'
out.mkdir(parents=True,exist_ok=True)
env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','LANG','SSL_CERT_FILE') if k in os.environ}
results=[]
for combo in ['openai-responses','anthropic-messages','google-gemini','openai-chat','deepseek-responses','deepseek-chat','typesafe-classifier','openai-images','google-interactions-image']:
 directory=out/'live'/combo
 command=['python3',str(Path(__file__).with_name('run-live.py')),'--live','--combo',combo,'--binary',str(root/'.evidence/barness-ai-thinking-sampling/live.test'),'--out',str(directory),'--alias','local-'+combo.split('-')[0]+'@region-permissions-unconfirmed']
 r=subprocess.run(command,cwd=root,env=env)
 reports=list(directory.glob('*/live-report.json'))
 merged=None
 if len(reports)==1:
  merge=subprocess.run(['go','run','./ai/live/cmd/supportmatrix',str(reports[0])],cwd=root,env=env,capture_output=True)
  merged=merge.returncode
  (directory/'merge.log').write_bytes((merge.stdout+merge.stderr).replace(str(root).encode(),b'[REPO]').replace(str(Path.home()).encode(),b'[HOME]'))
 results.append({'combo':combo,'exitCode':r.returncode,'mergeExitCode':merged,'bundle':str(reports[0].parent.relative_to(root)) if len(reports)==1 else None})
 (out/'smoke-commands.json').write_text(json.dumps(results,indent=2)+'\n')
 print(combo,'exit',r.returncode,'merge',merged,flush=True)
