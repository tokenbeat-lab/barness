"""Seven non-executable proposals. No authorization identity or expiry issued."""
from pathlib import Path
import copy,hashlib,json,subprocess,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement';SCRATCH=ROOT/'.scratch/minimax-thinking-supplement'
sys.path.insert(0,str(ROOT/'tools/verification'));from run_process import environment
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(Path(p).read_text())
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
freeze=load(BASE/'build-v1/freeze.json');out=SCRATCH/'pending-scopes-v3';out.mkdir(exist_ok=True);assert not list(out.iterdir());meta=BASE/'proposal-templates-v2';meta.mkdir();env,_=environment('e2e')
cli=ROOT/freeze['processgrant']['path'];refs=[]
for old in sorted((ROOT/'.scratch/provider-final-readiness/kimi-scopes-v2').glob('*.json')):
 d=load(old);route=d['route'];baseline=d['baseline']['manifest']['sha256']
 argv=[str(cli),'-template',route,'-stage','candidate','-purpose','xai-history-supplement','-baseline-manifest-sha256',baseline]
 r=subprocess.run(argv,cwd=ROOT,env=env,capture_output=True);(meta/(route+'.stdout.json')).write_bytes(r.stdout);(meta/(route+'.stderr.log')).write_bytes(r.stderr)
 assert r.returncode==0,r.stderr.decode();g=json.loads(r.stdout)
 assert g['sourceSha256']==freeze['regionalSource']['sourceSha256']
 assert g['catalogHash']==d['facts']['candidateCatalogHash'] and g['sdk']==d['facts']['sdk']
 d['previousProposal']=ref(old);d['nativeSourceSHA256']=g['sourceSha256'];d['ordinarySource']=freeze['source'];d['frozenProducer']=freeze['kimi']
 d['frozenProducerEvidence']={'kind':'normal live-tag build; final controlled ordinary evidence is indexed in current status','freeze':ref(BASE/'build-v1/freeze.json'),'metadata':ref(BASE/'build-v1/08-kimi-metadata/stdout')}
 d['machineTemplateNotAuthorization']=ref(meta/(route+'.stdout.json'))
 d['sourceChangeReason']='shared grant and MiniMax collector/consumer changes only; old54 vendor items remain applicable; no runtime/SDK/model/price change or old authority renewal'
 save(out/(route+'.json'),d);refs.append(ref(out/(route+'.json')))
base=ROOT/'.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id';bundle=base/'bundles/20261027T000249.211139000Z'
p={'documentType':'MINIMAX_PAID_SCOPE_PROPOSAL_NOT_A_GRANT','approved':False,'grantIssued':False,'executable':False,'route':'minimax-messages','profile':'minimax-messages-thinking-on/v1','identity':{'provider':'minimax','api':'anthropic-messages','model':'MiniMax-M3','accountPurpose':'standard','endpoint':'https://api.minimax.io/anthropic'},'baseline':{'bundle':str(bundle.relative_to(ROOT)),'manifest':ref(bundle/'manifest.json'),'report':ref(bundle/'minimax-report.json'),'seal':ref(bundle/'minimax-integrity.json'),'producer':ref(base/'receipt.json'),'originalOutcome':'FAIL','originalExitCode':1,'originalLogicalCalls':15,'retainedApplicableItems':13,'currentReadOnlyApplicability':ref(BASE/'original-baseline-v5/receipt.json')},'machineScope':freeze['minimaxMachineScope'],'nativeSourceSHA256':freeze['minimaxSource']['sourceSha256'],'ordinarySource':freeze['source'],'frozenProducer':freeze['minimax'],'frozenConsumer':freeze['releasegate'],'frozenProducerEvidence':{'freeze':ref(BASE/'build-v1/freeze.json'),'metadata':ref(BASE/'build-v1/07-minimax-metadata/stdout')},'executionScope':{'entry':'^TestMiniMaxThinkingOnSupplement$','testCount':1,'testTimeoutSeconds':180,'callSeconds':120,'calls':1,'httpAttempts':1,'sdkRetries':0,'environmentRetries':0,'requestReruns':0,'toolReruns':0,'reservedInput':2048,'requestedOutput':2048,'currency':'USD','proposedMaxUSD':'0.01','conservativeKnownRateReservationUSD':'0.006144','priceBasis':'minimax-global-m3-standard-2026-10-18-direct-usd','notAnInvoiceCeiling':True},'identityToConfirm':{'accountAlias':'MINIMAX','regionAlias':'global','tenantID':'live-smoke','accountScopeID':'live-minimax-messages','accountPurpose':'own international standard API','confirmation':'Trusted Host must confirm the actual same own account, current model/tier/rates and payload before any new grant; aliases alone are not account proof'},'authorizationStillRequired':{'id':'NOT_ISSUED','expiresAt':'NOT_ISSUED','newIndependentGrant':True,'noOldBudgetReuse':True,'recommendedValidityMinutes':30,'maximumValidityHours':24,'onlyThisOneScene':True},'stopConditions':['first failure/cancellation/timeout: stop; consumed reservation is retained','missing/null billable counter or unknown positive cache write price: refuse qualification','wire/model/source/SDK/binary/account/baseline/price/cap differs: return to offline review; never raise cap or rerun'],'excludes':['all other13 MiniMax scenes','Kimi or other Providers','warmup/health vendor calls','full formal rerun','race','pressure','paused-round resume']}
save(out/'minimax-messages.json',p);refs.append(ref(out/'minimax-messages.json'))
save(BASE/'pending-scopes-v3.json',{'documentType':'PROPOSAL_INDEX_NOT_AUTHORIZATION','status':'NOT_GRANTED','scopeDocuments':refs,'minimax':{'calls':1,'attempts':1,'USD':'0.01'},'kimi':{'calls':6,'attempts':6,'USD':'0.60'},'totalProposedCalls':7,'totalProposedHTTPAttempts':7,'totalProposedUSD':'0.61','budgetsIndependent':True,'grantsIssued':0,'realClaims':0,'vendorCalls':0})
(out/'README.md').write_text('# 七份待批准范围 v3\n\nMiniMax 国际 Messages 单项 1 次／USD0.01；Kimi 六路线各 1 次／USD0.10，总 USD0.60。合计七个事实缺口／拟议 USD0.61，各范围独立，不共用余额。全部 approved=false、executable=false，ID/有效期未签发，无真实 claim。\n\n新冻结仅正常 live 标签构建；来源、SDK、目录与各自原 manifest 已绑定。MiniMax 受控补项二进制与本冻结字节相同；Kimi 新冻结受控回归见本功能唯一 status。旧 v1/v2 提案、原授权及账本全部保留。消费者来源变化不使旧54项或其他 Provider 资格过期。\n\nMiniMax 缺输入/输出/缓存读/缓存写任一计费counter、正缓存写价未知、失败或取消时停止；保留整个预留 USD0.006144 及消费，不自动重试。批准后仅用精确单项入口，原13项不重买。付费执行、目录激活、race/压力及暂停恢复均未获本轮授权。\n')
print(json.dumps({'proposals':len(refs),'USD':'0.61','executable':False}))
