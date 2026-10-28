"""One explicitly requested retry of the failed MiniMax single scene only.
The original seven-route STOP is never cleared or resumed.
"""
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
assert (B/'STOP.json').exists();previous=load(B/'minimax-attempt-index.json');assert previous['nativeExitCode']==1 and previous['captureStatus']=='FAIL'
pre=load(B/'preflight.json');assert source_snapshot(R)==pre['source']
out=B/'retry-minimax-v1';out.mkdir();(out/'status-before.md').write_bytes((S/'status.md').read_bytes())
protected={str(p.relative_to(R)):sha(p) for p in [B/'STOP.json',B/'grants/minimax-messages.json',B/'approval.json',R/previous['nativeReceipt']['path']]}
indexPath=R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json';index=load(indexPath);scopeRef=next(x for x in index['scopeDocuments'] if x['path'].endswith('/minimax-messages.json'));d=load(R/scopeRef['path']);assert sha(R/scopeRef['path'])==scopeRef['sha256']
binary=R/d['frozenProducer']['path'];assert sha(binary)==d['frozenProducer']['sha256']
values=[]
for line in (R/'.env').read_text().splitlines():
 m=re.match(r'^\s*(?:export\s+)?MINIMAX_KEY\s*=\s*(.*)$',line)
 if m:
  raw=m[1].strip()
  if raw.startswith(('"',"'")):
   tokens=shlex.split(raw,comments=True,posix=True);assert len(tokens)==1;values.append(tokens[0])
  else:values.append(raw.split(' #',1)[0].strip())
assert len(values)==1 and values[0] and not values[0].startswith('$');key=values[0]
now=datetime.now(timezone.utc)
approval={'documentType':'EXPLICIT_HUMAN_ONE_CURRENT_ROUTE_RETRY_APPROVAL','recordedAt':now.isoformat(),'humanMessageVerbatim':'重试','scopeInterpretation':'Exactly one new attempt of the currently failed MiniMax thinking-on scene, same independent USD0.01 cap; no restart of the stopped six Kimi routes and no increase to the original global USD0.61 cap.','previousFailure':previous['nativeReceipt'],'originalBatchStop':ref(B/'STOP.json'),'originalScope':scopeRef,'maxNewCalls':1,'maxNewHTTPAttempts':1,'maxNewUSD':'0.01','allSDKEnvironmentRequestToolRetries':0,'newGrantLifetimeMinutes':30,'KimiAuthorityGranted':False,'raceAuthorized':False,'pressureAuthorized':False,'catalogActivationAuthorized':False,'pausedResumeAuthorized':False,'originalFailureAndConsumptionPreserved':True}
save(out/'approval.json',approval)
(S/'retry-spec-v1.md').write_text('''# 当前MiniMax单项一次重试\n\n用户在首项失败后直接指令“重试”。仅新增一次当前MiniMax thinking-on的独立新grant/30分钟、1调用/1HTTP/USD0.01，沿用完全相同v5事实确认修订1范围及冻结二进制，零自动重试。原首次失败、授权、claim与账本保留，原七路STOP不清除；本次不授予Kimi六路恢复权限，也不提高原总预算。重试后只读记录结果；任何后续付费需重新核对剩余总预算与明确批准。\n''')
grant={'schema':1,'id':'minimax-thinking-retry-'+now.strftime('%Y%m%dT%H%M%SZ')+'-'+uuid.uuid4().hex[:12],'accountAlias':d['identityToConfirm']['accountAlias'],'regionAlias':d['identityToConfirm']['regionAlias'],'expiresAt':(now+timedelta(minutes=30)).isoformat().replace('+00:00','Z'),'scope':d['machineScope']}
grantPath=out/'grant.json';save(grantPath,grant);os.chmod(grantPath,0o600)
save(out/'issuance.json',{'grant':ref(grantPath),'approval':ref(out/'approval.json'),'issuedAt':now.isoformat(),'maxUSD':'0.01','KimiGrantsIssued':0})
(S/'status.md').write_text('# 当前状态\n\n首次MiniMax真实单项已FAIL/exit1，1调用/1HTTP，计费用量cache_creation_input_tokens未报告；原七路STOP保留。按用户后续“重试”，仅新增一次同范围MiniMax单项USD0.01，正在执行；Kimi六路仍未签发/未执行，未增加原全局预算，不恢复旧序列。race/压力、目录激活和暂停恢复均禁止。\n')
env,_=environment('e2e');env.update(BARNESS_AI_LIVE='1',BARNESS_MINIMAX_ROUTE='messages',BARNESS_MINIMAX_MODEL='MiniMax-M3',BARNESS_MINIMAX_PURPOSE='standard',BARNESS_MINIMAX_GRANT=str(grantPath),BARNESS_MINIMAX_ACCOUNT_ALIAS=grant['accountAlias'],BARNESS_MINIMAX_REGION_ALIAS=grant['regionAlias'],BARNESS_MINIMAX_ENTITLEMENT='confirmed',BARNESS_MINIMAX_KEY=key)
controlled=load(R/'.evidence/kimi-composition-promotion/frozen-controlled-v1.json')['measurements']['minimax-messages'];native=load((R/controlled['report']['path']).parent/'execution-source.json');assert all(sha(R/n)==h for n,h in native.items())
inputs=[n for n in native if n not in pre['source']['files']]+[str(p.relative_to(R)) for p in [Path(__file__).resolve(),S/'retry-spec-v1.md',out/'approval.json',grantPath,R/scopeRef['path'],indexPath,R/index['confirmationEvidence']['path'],R/d['baseline']['manifest']['path']]]
inputs=sorted(set(inputs));argv=['python3','tools/verification/run.py','--feature','provider-minimal-paid-supplement','--stage','live','--route','minimax-messages','--timeout','205']
for n in inputs:argv+=['--input',n]
argv+=['--',str(binary),'-test.run=^TestMiniMaxThinkingOnSupplement$','-test.count=1','-test.timeout=180s']
save(out/'command-selection.json',{'argv':argv,'inputs':{n:sha(R/n) for n in inputs},'credentialSourceVariable':'MINIMAX_KEY','credentialStored':False,'launchLimit':1})
started=datetime.now(timezone.utc);wall=time.monotonic();r=subprocess.run(argv,cwd=R,env=env,capture_output=True)
stdout=r.stdout.decode(errors='replace').replace(key,'[REDACTED]');stderr=r.stderr.decode(errors='replace').replace(key,'[REDACTED]');(out/'wrapper.stdout').write_text(stdout);(out/'wrapper.stderr').write_text(stderr)
rows=[json.loads(line) for line in stdout.splitlines() if line.startswith('{')];assert len(rows)==1,'capture unavailable: stop, never retry'
receiptPath=R/rows[0]['attempt']/'receipt.json';receipt=load(receiptPath)
save(out/'attempt-index.json',{'route':'minimax-messages','startedAt':started.isoformat(),'finishedAt':datetime.now(timezone.utc).isoformat(),'elapsedSeconds':time.monotonic()-wall,'wrapperExitCode':r.returncode,'nativeReceipt':ref(receiptPath),'nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'grant':ref(grantPath),'KimiGrantsIssued':0})
assert all(sha(R/n)==h for n,h in protected.items())
passed=r.returncode==0 and receipt['status']=='PASS' and receipt['exitCode']==0
save(out/'STOP.json',{'kind':'retry-finished-no-downstream-paid-authority','producerPass':passed,'nativeReceipt':ref(receiptPath),'nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'originalBatchStopStillInForce':True,'KimiGrantsIssued':0,'furtherPaidCallsRequireBudgetReviewAndHumanApproval':True,'automaticRetryForbidden':True})
(S/'status.md').write_text('# 当前状态\n\n首次MiniMax单项及用户明确要求的一次重试均已结束，结果与真实消费见.evidence/provider-minimal-paid-supplement原attempt和retry-minimax-v1。原失败、授权、claim、ledger及STOP均保留。Kimi六路仍未签发/未执行；没有自动重试、增加总预算、race/压力、目录激活或暂停恢复。后续只做离线核验和结果收口，不再发送付费请求。\n')
print(json.dumps({'retryProducerPass':passed,'route':'minimax-messages','nativeExitCode':receipt['exitCode'],'captureStatus':receipt['status'],'receipt':str(receiptPath.relative_to(R)),'KimiGrantsIssued':0,'furtherPaidExecutionStopped':True}));raise SystemExit(0 if passed else 1)
