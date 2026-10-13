from pathlib import Path
import os,subprocess,json,datetime,hashlib
root=Path.cwd()
key_values=[]
for line in (root/'.env').read_text().splitlines():
    name,sep,value=line.partition('=')
    if sep and name.strip()=='OPENAI_KEY':
        value=value.strip()
        if len(value)>=2 and value[0]==value[-1] and value[0] in ['"',"'"]:value=value[1:-1]
        key_values.append(value)
if len(key_values)!=1 or not key_values[0]:raise SystemExit('OPENAI_KEY is missing or ambiguous')
key=key_values[0]
base=root/'.evidence/decisions-candidate-live-authorized'
base.mkdir(parents=True,exist_ok=True)
env={k:os.environ[k] for k in ['PATH','HOME','TMPDIR'] if k in os.environ}
env.update({'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':'openai-decisions-candidate','BARNESS_AI_LIVE_KEY_OPENAI_DECISIONS_CANDIDATE':key,'BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT':'1','BARNESS_AI_LIVE_PROXY':'http://127.0.0.1:7890','BARNESS_AI_EVIDENCE_DIR':str(base)})
cmd=['go','test','-tags','live','-timeout','180s','./ai/live','-run','^TestCandidateDecisions$','-count=1','-json']
start=datetime.datetime.now(datetime.timezone.utc).isoformat()
r=subprocess.run(cmd,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=210)
output=r.stdout.decode().replace(key,'[REDACTED]').replace(str(Path.home()),'<host-home>')
Path('/tmp/barness-decisions-authorized-live.jsonl').write_text(output)
receipt={'command':cmd,'startedAt':start,'finishedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'exitCode':r.returncode,'credentialSource':'.env:OPENAI_KEY','proxy':'http://127.0.0.1:7890','accountVerification':'unknown','regionVerification':'unknown','permissionVerification':'unknown','userAuthorization':'2026-10-13: use .env OPENAI_KEY via 127.0.0.1:7890; unknown account/region/permissions must not block','logSHA256':hashlib.sha256(output.encode()).hexdigest()}
Path('/tmp/barness-decisions-authorized-live-receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({'exitCode':r.returncode,'log':'/tmp/barness-decisions-authorized-live.jsonl'}))
