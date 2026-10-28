"""Execute only the newly approved MiniMax thinking-on attempt, exactly once."""
from pathlib import Path
from datetime import datetime,timezone,timedelta
import hashlib,json,os,re,subprocess,sys,time,uuid,shlex
R=Path(__file__).resolve().parents[2];B=R/'.evidence/provider-minimal-paid-supplement';S=R/'.scratch/provider-minimal-paid-supplement'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
from run_process import environment
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
pre=load(B/'preflight.json');assert pre['pass'] and source_snapshot(R)==pre['source']
assert not (B/'STOP.json').exists() and not (B/'grants').exists(),'one attempt only; never resume/reissue'
index=load(R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json');f=next(x for x in index['scopeDocuments'] if x['path'].endswith('/minimax-messages.json'));p=R/f['path'];assert sha(p)==f['sha256'];d=load(p)
assert d['executionScope']['calls']==d['executionScope']['httpAttempts']==1 and d['executionScope']['proposedMaxUSD']=='0.01'
binary=R/d['frozenProducer']['path'];assert sha(binary)==d['frozenProducer']['sha256']
# Retrieve only the selected existing secret. Never execute/source .env or log it.
values=[]
for line in (R/'.env').read_text().splitlines():
 m=re.match(r'^\s*(?:export\s+)?MINIMAX_KEY\s*=\s*(.*)$',line)
 if m:
  raw=m[1].strip()
  if raw.startswith(('"',"'")):
   tokens=shlex.split(raw,comments=True,posix=True);assert len(tokens)==1;values.append(tokens[0])
  else:values.append(raw.split(' #',1)[0].strip())
assert len(values)==1 and values[0] and not values[0].startswith('$'),'selected credential unavailable/ambiguous; no grant issued'
key=values[0];env,_=environment('e2e')
# Check current machine template without issuing authority or burning a claim.
freeze=load(R/'.evidence/kimi-composition-promotion/build-v2/freeze.json');cli=R/freeze['processgrant']['path']
args=[str(cli),'-template','minimax-messages','-mode','live','-stage','candidate','-profile','minimax-messages-thinking-on/v1','-baseline-manifest-sha256',d['baseline']['manifest']['sha256'],'-binary',str(binary)]
c=subprocess.run(args,cwd=R,env=env,capture_output=True);assert c.returncode==0,'current scope template refused'
template=json.loads(c.stdout);assert template['scope']==d['machineScope']
# Bind every actual native source byte; Markdown lives in selectedInputs.
controlled=load(R/'.evidence/kimi-composition-promotion/frozen-controlled-v1.json')['measurements']['minimax-messages'];nativeSourcePath=(R/controlled['report']['path']).parent/'execution-source.json';native=load(nativeSourcePath)
assert all(sha(R/n)==h for n,h in native.items())
assert hashlib.sha256(json.dumps(native,sort_keys=True,separators=(',',':')).encode()).hexdigest()==d['nativeSourceSHA256']
out=B/'grants';out.mkdir();now=datetime.now(timezone.utc);gid='minimax-thinking-'+now.strftime('%Y%m%dT%H%M%SZ')+'-'+uuid.uuid4().hex[:12]
grant={'schema':1,'id':gid,'accountAlias':d['identityToConfirm']['accountAlias'],'regionAlias':d['identityToConfirm']['regionAlias'],'expiresAt':(now+timedelta(minutes=30)).isoformat().replace('+00:00','Z'),'scope':d['machineScope']}
grantPath=out/'minimax-messages.json';save(grantPath,grant);os.chmod(grantPath,0o600)
save(B/'minimax-grant-issuance.json',{'route':'minimax-messages','grant':ref(grantPath),'approval':ref(B/'approval.json'),'scopeProposal':f,'issuedAt':now.isoformat(),'expiresAt':grant['expiresAt'],'maxCalls':1,'maxHTTPAttempts':1,'maxUSD':'0.01','otherRouteGrantsIssued':0})
(S/'status.md').write_text('# 当前状态\n\nMiniMax thinking-on 单项已按本次批准签发唯一新grant，正在执行一次真实调用；Kimi六份尚未签发。当前范围1调用/1HTTP/USD0.01、零重试，退出或验收失败立即停止剩余序列。grant及本轮批准原件均保存，源码/冻结绑定未变。race/压力、目录激活和暂停恢复均禁止。\n')
env.update(BARNESS_AI_LIVE='1',BARNESS_MINIMAX_ROUTE='messages',BARNESS_MINIMAX_MODEL='MiniMax-M3',BARNESS_MINIMAX_PURPOSE='standard',BARNESS_MINIMAX_GRANT=str(grantPath),BARNESS_MINIMAX_ACCOUNT_ALIAS=grant['accountAlias'],BARNESS_MINIMAX_REGION_ALIAS=grant['regionAlias'],BARNESS_MINIMAX_ENTITLEMENT='confirmed',BARNESS_MINIMAX_KEY=key)
inputs=[n for n in native if n not in pre['source']['files']]
inputs+=[str(x.relative_to(R)) for x in [Path(__file__).resolve(),S/'spec.md',B/'approval.json',B/'preflight.json',grantPath,p,R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json',R/index['confirmationEvidence']['path'],R/d['baseline']['manifest']['path']]]
inputs=sorted(set(inputs));argv=['python3','tools/verification/run.py','--feature','provider-minimal-paid-supplement','--stage','live','--route','minimax-messages','--timeout','205']
for n in inputs:argv+=['--input',n]
argv+=['--',str(binary),'-test.run=^TestMiniMaxThinkingOnSupplement$','-test.count=1','-test.timeout=180s']
save(B/'minimax-command-selection.json',{'argv':argv,'inputs':{n:sha(R/n) for n in inputs},'credentialSourceVariable':'MINIMAX_KEY','credentialForwardedOnlyAs':'BARNESS_MINIMAX_KEY','credentialStored':False,'paidProcessLaunchLimit':1})
started=datetime.now(timezone.utc);wall=time.monotonic();r=subprocess.run(argv,cwd=R,env=env,capture_output=True)
# run.py redacts the selected secret from native logs; redact its own streams too.
stdout=r.stdout.decode(errors='replace').replace(key,'[REDACTED]');stderr=r.stderr.decode(errors='replace').replace(key,'[REDACTED]')
(B/'minimax-wrapper.stdout').write_text(stdout);(B/'minimax-wrapper.stderr').write_text(stderr)
rows=[json.loads(line) for line in stdout.splitlines() if line.startswith('{')];assert len(rows)==1,'capture result unavailable; stop without retry'
row=rows[0];receiptPath=R/row['attempt']/'receipt.json';receipt=load(receiptPath)
save(B/'minimax-attempt-index.json',{'route':'minimax-messages','startedAt':started.isoformat(),'finishedAt':datetime.now(timezone.utc).isoformat(),'elapsedSeconds':time.monotonic()-wall,'wrapperExitCode':r.returncode,'nativeReceipt':ref(receiptPath),'nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'grant':ref(grantPath),'KimiGrantsIssued':0})
if r.returncode or receipt['status']!='PASS' or receipt['exitCode']!=0:
 save(B/'STOP.json',{'kind':'stop-entire-remaining-paid-sequence','route':'minimax-messages','reason':'first paid process/capture did not pass','nativeReceipt':ref(receiptPath),'nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'issuedGrants':['minimax-messages'],'remainingSixGrantsIssued':False,'automaticRetryForbidden':True,'newApprovalRequiredForAnyFurtherPaidExecution':True})
 (S/'status.md').write_text('# 当前状态\n\n已按失败即停规则停止整个剩余付费序列。MiniMax首次单项进程/采集未通过，实际退出与原件见.evidence/provider-minimal-paid-supplement/minimax-attempt-index.json和STOP.json。本次MiniMax授权及消费保留，不重试、不换ID、不续期。Kimi六份未签发、未执行。race/压力、目录激活和暂停恢复均未执行。\n')
 print(json.dumps({'stopped':True,'route':'minimax-messages','wrapperExitCode':r.returncode,'nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'receipt':str(receiptPath.relative_to(R)),'KimiGrantsIssued':0}));raise SystemExit(1)
print(json.dumps({'producerPass':True,'route':'minimax-messages','nativeReceipt':str(receiptPath.relative_to(R)),'nextAction':'read-only native supplement and fixed13+1 validation before any Kimi grant'}))
