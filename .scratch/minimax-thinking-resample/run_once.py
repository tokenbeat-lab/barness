"""Issue and launch exactly this newly approved MiniMax single once. No loop."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,os,re,shlex,sys,time,uuid
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-thinking-resample';S=R/'.scratch/minimax-thinking-resample'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
from run_process import environment,capture
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False);f.write('\n')
pre=load(B/'preflight.json');a=load(B/'approval.json');assert pre['pass'] and source_snapshot(R)==pre['source']
assert not any((B/n).exists() for n in ['grant.json','issuance.json','execution','live','attempt-index.json','STOP.json']),'one attempt only; never resume/reissue'
assert datetime.now(timezone.utc)<datetime.fromisoformat(a['expiresAt'].replace('Z','+00:00'))
freeze=load(R/a['freeze']['path']);assert sha(R/a['freeze']['path'])==a['freeze']['sha256']
for key in ['minimax','processgrant','releasegate']:assert sha(R/freeze[key]['path'])==freeze[key]['sha256']
ordinary=load(R/pre['ordinary']['path']);assert ordinary['source']==pre['source'] and all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
native=load(B/'native-source.json');assert all(sha(R/n)==h for n,h in native.items())
# Retrieve only this selected secret; do not source/execute .env or serialize it.
values=[]
for line in (R/'.env').read_text().splitlines():
 m=re.match(r'^\s*(?:export\s+)?MINIMAX_KEY\s*=\s*(.*)$',line)
 if m:
  raw=m[1].strip()
  if raw.startswith(('"',"'")):
   tokens=shlex.split(raw,comments=True,posix=True);assert len(tokens)==1,'selected credential ambiguous';values.append(tokens[0])
  else:values.append(raw.split(' #',1)[0].strip())
assert len(values)==1 and values[0] and not values[0].startswith('$'),'selected credential unavailable; no grant issued'
key=values[0];now=datetime.now(timezone.utc);assert now<datetime.fromisoformat(a['expiresAt'].replace('Z','+00:00'))
grant={'schema':1,'id':'minimax-thinking-resample-'+now.strftime('%Y%m%dT%H%M%SZ')+'-'+uuid.uuid4().hex[:12],'accountAlias':'MINIMAX','regionAlias':'global','expiresAt':a['expiresAt'],'scope':load(B/'machine-template.json')['scope']}
assert grant['scope']==freeze['minimaxMachineScope'];grantPath=B/'grant.json';save(grantPath,grant);os.chmod(grantPath,0o600)
save(B/'issuance.json',{'grant':ref(grantPath),'approval':ref(B/'approval.json'),'issuedAt':now.isoformat(),'expiresAt':a['expiresAt'],'maxCalls':1,'maxHTTPAttempts':1,'maxUSD':'0.01','oldBudgetReused':False,'KimiGrantsIssued':0})
(S/'status.md').write_text('# MiniMax thinking-on 单项重采当前状态\n\n本次preflight及原13项离线核验通过，当前源码与既有完整普通/冻结来源相等。已签发本次唯一grant，仅一次thinking-on，1调用/1HTTP/USD0.01，正在执行；没有恢复旧七路序列或Kimi权限。失败即停，不重试或更换ID。旧证据、授权、账本、STOP及Kimi目录均保留，不启动race/压力或恢复暂停。\n')
env,_=environment('e2e');env.update(BARNESS_AI_LIVE='1',BARNESS_MINIMAX_ROUTE='messages',BARNESS_MINIMAX_MODEL='MiniMax-M3',BARNESS_MINIMAX_PURPOSE='standard',BARNESS_MINIMAX_GRANT=str(grantPath),BARNESS_MINIMAX_ACCOUNT_ALIAS='MINIMAX',BARNESS_MINIMAX_REGION_ALIAS='global',BARNESS_MINIMAX_ENTITLEMENT='confirmed',BARNESS_MINIMAX_KEY=key)
inputs=[n for n in native if n not in pre['source']['files']]
inputs += [str(p.relative_to(R)) for p in [Path(__file__).resolve(),S/'spec.md',B/'approval.json',grantPath,B/'machine-template.json',B/'native-source.json',R/a['freeze']['path'],R/pre['baseline']['manifest']['path'],R/'.evidence/kimi-composition-promotion/account-confirmation-v1/facts.json']]
inputs=sorted(set(inputs));binary=R/a['producer']['path'];argv=['python3','tools/verification/run.py','--feature','minimax-thinking-resample','--stage','live','--route','minimax-messages','--timeout','205']
for n in inputs:argv+=['--input',n]
argv+=['--',str(binary),'-test.run=^TestMiniMaxThinkingOnSupplement$','-test.count=1','-test.timeout=180s']
save(B/'command-selection.json',{'argv':argv,'inputs':{n:sha(R/n) for n in inputs},'credentialSourceVariable':'MINIMAX_KEY','credentialStored':False,'launchLimit':1})
out=B/'execution';out.mkdir();started=datetime.now(timezone.utc);tic=time.monotonic();result=capture(argv,R,env,out,230,[key]);files={p.name:sha(p) for p in out.iterdir() if p.is_file()};save(out/'receipt.json',{'argv':argv,'startedAt':started.isoformat(),'finishedAt':datetime.now(timezone.utc).isoformat(),'elapsedSeconds':time.monotonic()-tic,'files':files,**result})
receipts=list((B/'live/minimax-messages').glob('*/receipt.json'));receiptPath=receipts[0] if len(receipts)==1 else None;receipt=load(receiptPath) if receiptPath else {};passed=result['status']=='PASS' and result['exitCode']==0 and receipt.get('status')=='PASS' and receipt.get('exitCode')==0
save(B/'attempt-index.json',{'route':'minimax-messages','outerCapture':ref(out/'receipt.json'),'nativeReceipt':ref(receiptPath) if receiptPath else None,'nativeExitCode':receipt.get('exitCode'),'captureStatus':receipt.get('status','UNKNOWN'),'producerPass':passed,'grant':ref(grantPath),'KimiGrantsIssued':0,'paidLaunches':1})
if not passed:
 save(B/'STOP.json',{'kind':'single-paid-resample-failed-or-unknown','nativeReceipt':ref(receiptPath) if receiptPath else None,'nativeExitCode':receipt.get('exitCode'),'automaticRetryForbidden':True,'newGrantForbiddenWithoutNewHumanApproval':True,'KimiGrantsIssued':0})
 (S/'status.md').write_text('# MiniMax thinking-on 单项重采当前状态\n\nSTOPPED：本次唯一付费启动已结束但未获得成功producer；详见.evidence/minimax-thinking-resample/attempt-index.json与STOP.json。保留本次grant及实际claim/ledger，不重试、不换ID、不合并预算；后续仅只读核查与收口，不再发厂商请求。Kimi/其他Provider、race/压力、目录激活及暂停恢复均未执行。\n')
print(json.dumps({'producerPass':passed,'nativeExitCode':receipt.get('exitCode'),'captureStatus':receipt.get('status','UNKNOWN'),'receipt':str(receiptPath.relative_to(R)) if receiptPath else None,'paidLaunches':1,'automaticRetryForbidden':True}));raise SystemExit(0 if passed else 1)
