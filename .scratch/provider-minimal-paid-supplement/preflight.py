"""Read-only preflight for the explicitly authorized seven fixed supplements."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,sys,subprocess
R=Path(__file__).resolve().parents[2];B=R/'.evidence/provider-minimal-paid-supplement';S=R/'.scratch/provider-minimal-paid-supplement'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
from run_process import environment,capture
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
P=R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json'
assert sha(P)=='818152762619e44b4a8d60946e944ea4201ef9fd19f9725ce288083e50d99c38'
index=load(P);freeze=load(R/'.evidence/kimi-composition-promotion/build-v2/freeze.json')
source=source_snapshot(R);assert source==freeze['source']
ordinaryPath=R/'.evidence/kimi-composition-promotion/e2e/20261028T012642.072979Z-ud9sss5q/receipt.json';ordinary=load(ordinaryPath)
assert ordinary['status']=='PASS' and ordinary['exitCode']==0 and source==ordinary['source'];assert all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
order=['minimax-messages','kimi-cn-chat','kimi-intl-chat','kimi-cn-responses','kimi-intl-responses','kimi-cn-messages','kimi-intl-messages'];docs={};pins={};refs={}
def check(d):
 if isinstance(d,dict):
  if isinstance(d.get('path'),str) and isinstance(d.get('sha256'),str):
   p=R/d['path'];assert p.is_file() and not p.is_symlink();refs[d['path']]=sha(p);assert refs[d['path']]==d['sha256'],d['path']
  for v in d.values():check(v)
 elif isinstance(d,list):
  for v in d:check(v)
check(index)
for f in index['scopeDocuments']:
 d=load(R/f['path']);check(d);docs[d['route']]=d;pins[d['route']]=f
 assert d['confirmations']['accountAndRegion']==d['confirmations']['ratesAndServiceTier']=='USER_CONFIRMED'
 assert not any(d[k] for k in ['approved','grantIssued','executable'])
 e=d['executionScope'];assert e['calls']==e['httpAttempts']==e['testCount']==1 and e['sdkRetries']==e['environmentRetries']==0
 assert e['proposedMaxUSD']==('0.01' if d['route']=='minimax-messages' else '0.10')
 assert e['testTimeoutSeconds']==(180 if d['route']=='minimax-messages' else 240)
assert set(docs)==set(order)
sealed=0
for d in docs.values():
 for name,h in load(R/d['baseline']['seal']['path'])['files'].items():assert sha(R/d['baseline']['bundle']/name)==h;sealed+=1
assert sha(R/'ai/catalog_kimi.json')=='c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109'
env,_=environment('e2e')
assert not any(k.startswith('GIT_') for k in env)
common=subprocess.check_output(['git','rev-parse','--path-format=absolute','--git-common-dir'],cwd=R,env=env,text=True).strip();assert Path(common).resolve()==(R/'.git').resolve()
roots=load(R/'.evidence/minimax-thinking-supplement-plan/store-roots.json');public={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}
oldPublic=load(R/'.evidence/kimi-composition-promotion/initial/public-stores-before.json');assert public=={str(R/p):h for p,h in oldPublic.items()}
save(B/'historical-authority-before.json',public)
checks=B/'preflight';checks.mkdir()
cli=R/freeze['processgrant']['path']
for name,args,expected in [('minimax-source',['-source','minimax-messages','-stage','candidate'],docs['minimax-messages']['nativeSourceSHA256']),('regional-source',['-source','regional'],docs['kimi-cn-chat']['nativeSourceSHA256'])]:
 out=checks/name;out.mkdir();argv=[str(cli)]+args;r=capture(argv,R,env,out,90,[]);save(out/'receipt.json',{'argv':argv,**r});assert r['exitCode']==0
 current=json.loads((out/'stdout.jsonl').read_text());assert current['sourceSha256']==expected
# Existing production read-only verifier, compiled against this unchanged source.
oldInspect=R/'.evidence/kimi-composition-promotion/original-baseline-v6/inspect-original'
mm=docs['minimax-messages'];out=checks/'minimax-thirteen';out.mkdir();argv=[str(oldInspect),mm['baseline']['bundle'],mm['baseline']['producer']['path']]
r=capture(argv,R,env,out,120,[]);save(out/'receipt.json',{'argv':argv,'binary':ref(oldInspect),**r});assert r['exitCode']==0
proof=json.loads((out/'stdout.jsonl').read_text());assert proof['pass'] and len(proof['applicable'])==13
(S/'spec.md').write_text('''# 七个最小付费补采执行\n\n用户已明确批准v5事实确认修订1：先MiniMax thinking-on USD0.01，再Kimi国内/国际Chat、Responses、Messages各USD0.10；最多7调用/7HTTP、合计USD0.61。逐路线串行、签当前一次性grant/30分钟有效期、预算独立。补采和对应只读组合验证通过才签下一条；任何启动/采集/用量/审计/组合拒绝立即停止所有剩余签发与付费，禁止自动重试/换ID/续期/扩大范围。\n\n执行只用绑定的正常冻结二进制、原生精确单项/count1/180s或240s；Kimi真实首发等待60s。原54/13项按原件复用；MiniMax13+1，Kimi live/candidate 9+1（目录inactive），不激活Kimi、不启动race/压力、不恢复暂停。原授权/账本只保留并追加本轮真实记录，不迁移/覆盖/删除。仅当前路线凭据进入子进程，密钥与.env不进入argv/证据/selected inputs。\n\n来源/完整普通沿用已通过且集合相等的2048来源/185输入，不重跑普通或重建冻结。新的执行、授权与消费者各自固定实际输入、真实进程退出及原件。当前唯一状态见status.md。\n''')
(S/'status.md').write_text('# 当前状态\n\n已取得七项单轮条件付费批准；只读preflight通过，尚未签发grant或请求厂商。MiniMax原13项当前原生消费者PASS；六路旧九项沿用完整当前来源适用性与封存核验。源码/SDK/目录/价格/冻结二进制匹配。真实执行范围与停止条件见spec.md，原件见.evidence/provider-minimal-paid-supplement。race/压力、目录激活和暂停恢复均禁止。\n')
(S/'issues/01-seven-paid-supplements.md').write_text('Status: claimed\nType: task\n\n# 按批准顺序补七个真实缺项\n\n仅现有单项入口及只读固定组合，失败即停，不重试、不恢复暂停，不激活目录。\n')
quote='我明确批准按上述最终授权提示词执行七份 v5 事实确认修订1的最小付费补采：MiniMax USD0.01，Kimi六路各USD0.10，合计最多7调用/7HTTP、USD0.61。严格按既定顺序逐路线串行、独立授权与预算，失败即停；保留全部历史证据、授权和账本，不启动race或压力测试，不激活Kimi目录，也不恢复暂停轮次。'
approval={'documentType':'EXPLICIT_HUMAN_SEVEN_CONDITIONAL_PAID_SCOPE_APPROVAL','recordedAt':datetime.now(timezone.utc).isoformat(),'source':'direct current human user message, incorporating preceding final authorization prompt','verbatim':quote,'scopeIndex':ref(P),'facts':index['confirmationEvidence'],'scopeDocuments':[pins[r] for r in order],'order':order,'budgetUSD':{'minimax-messages':'0.01',**{r:'0.10' for r in order[1:]}},'totalMaxCalls':7,'totalMaxHTTPAttempts':7,'totalMaxUSD':'0.61','grantLifetimeMinutes':30,'signOnlyCurrentRouteAfterPredecessorPass':True,'failureStopsEntireRemainingSequence':True,'SDKEnvironmentRequestToolRetries':0,'newGrantOrRetryAfterStopRequiresNewHumanApproval':True,'raceAuthorized':False,'pressureAuthorized':False,'catalogActivationAuthorized':False,'pausedResumeAuthorized':False,'priorProposalsPreserved':True}
save(B/'approval.json',approval)
save(B/'preflight.json',{'pass':True,'source':source,'ordinary':ref(ordinaryPath),'scopeIndex':ref(P),'checkedReferencePins':refs,'sealedBaselineFilesChecked':sealed,'publicAuthorityFiles':len(public),'trustedGitCommonDir':common,'catalogActive':False,'calls':0,'httpAttempts':0,'grantsIssued':0,'paidProducerExecutions':0})
print(json.dumps({'preflight':'PASS','sourceFiles':len(source['files']),'baselineFiles':sealed,'publicAuthorityFiles':len(public),'approvedOrder':order,'grantsIssued':0,'calls':0}))
