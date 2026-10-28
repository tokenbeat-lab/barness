"""Rebind seven unapproved proposals; no ID, expiry, authority or invocation."""
from pathlib import Path
import hashlib,json,subprocess,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/kimi-composition-promotion';SCRATCH=ROOT/'.scratch/kimi-composition-promotion'
sys.path.insert(0,str(ROOT/'tools/verification'));from run_process import environment
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
freeze=load(BASE/'build-v2/freeze.json');out=SCRATCH/'pending-scopes-v5';out.mkdir(exist_ok=True);assert not list(out.iterdir());meta=BASE/'proposal-templates-v3';meta.mkdir();env,_=environment('e2e')
refs=[]
for old in sorted((ROOT/'.scratch/kimi-composition-promotion/pending-scopes-v4').glob('*.json')):
 d=load(old);route=d['route'];is_mm=route=='minimax-messages';which='minimax' if is_mm else 'kimi'
 assert not d['approved'] and not d['grantIssued'] and not d['executable']
 assert d['authorizationStillRequired']['id']=='NOT_ISSUED' and d['authorizationStillRequired']['expiresAt']=='NOT_ISSUED'
 d['previousProposal']=ref(old);d['ordinarySource']=freeze['source'];d['frozenProducer']=freeze[which];d['frozenConsumer']=freeze['releasegate']
 d['nativeSourceSHA256']=freeze['minimaxSource' if is_mm else 'regionalSource']['sourceSha256']
 d['frozenProducerEvidence']={'freeze':ref(BASE/'build-v2/freeze.json'),'metadata':ref(BASE/('build-v2/07-minimax-metadata/stdout' if is_mm else 'build-v2/08-kimi-metadata/stdout'))}
 d['confirmations']={'accountAndRegion':'PENDING','ratesAndServiceTier':'PENDING','checklist':ref(ROOT/'.scratch/kimi-formal-offline-review/account-rate-checklist.md'),'factsConfirmationIsNotPaidAuthorization':True}
 d['sourceChangeReason']='Kimi-only promotion consumer and public CLI tests; remove cross-measurement equality, retain fixed pairs and independent native applicability; no product/SDK/model/price change; old54 and MiniMax13 remain applicable, no old authority renewed'
 if is_mm:
  d['machineScope']=freeze['minimaxMachineScope']
  d['baseline']['currentReadOnlyApplicability']=ref(BASE/'original-baseline-v6/receipt.json')
 else:
  cli=ROOT/freeze['processgrant']['path'];argv=[str(cli),'-template',route,'-stage','candidate','-purpose','xai-history-supplement','-baseline-manifest-sha256',d['baseline']['manifest']['sha256']]
  r=subprocess.run(argv,cwd=ROOT,env=env,capture_output=True)
  (meta/(route+'.stdout.json')).write_bytes(r.stdout);(meta/(route+'.stderr.log')).write_bytes(r.stderr);(meta/(route+'.exit-code')).write_text(str(r.returncode)+'\n')
  assert r.returncode==0,r.stderr.decode();g=json.loads(r.stdout)
  assert g['id']=='<new human authorization required>' and g['expiresAt']=='0001-01-01T00:00:00Z'
  assert g['sourceSha256']==d['nativeSourceSHA256'] and g['catalogHash']==d['facts']['candidateCatalogHash'] and g['sdk']==d['facts']['sdk']
  d['machineTemplateNotAuthorization']=ref(meta/(route+'.stdout.json'))
  d['promotionReviewOnly']={'binary':freeze['regionalpromote'],'contract':'fixed-six-nine-plus-one/schema2','measurements':12,'catalogActivationAuthorized':False}
 save(out/(route+'.json'),d);refs.append(ref(out/(route+'.json')))
save(BASE/'pending-scopes-v5.json',{'documentType':'PROPOSAL_INDEX_NOT_AUTHORIZATION','status':'NOT_GRANTED','scopeDocuments':refs,'minimax':{'calls':1,'attempts':1,'USD':'0.01'},'kimi':{'calls':6,'attempts':6,'USD':'0.60'},'totalProposedCalls':7,'totalProposedHTTPAttempts':7,'totalProposedUSD':'0.61','budgetsIndependent':True,'accountAndRegionConfirmation':'PENDING','ratesAndTierConfirmation':'PENDING','grantsIssued':0,'realClaims':0,'vendorCalls':0})
(out/'README.md').write_text('''# 七份待批准范围 v4

本轮只重建消费者变更所需的正常标签冻结二进制与未批准绑定。MiniMax thinking-on 1次／USD0.01，Kimi两站六路线各补xai-origin-history 1次／USD0.10；合计7次调用、7次HTTP、拟议USD0.61，七个独立预算，均零重试。原54＋13项事实继续独立适用，不重买旧场景。

全部 approved=false、grantIssued=false、executable=false；ID/有效期 NOT_ISSUED。账户、地区、费率与服务档位仍 PENDING，最少事实确认沿用[账户费率清单](../../kimi-formal-offline-review/account-rate-checklist.md)。事实确认不等于付费授权。MiniMax精确180s/2048输入输出，Kimi精确240s/4096输入输出且真实首发等待60s，调用均120s；停止条件、账本、失败耗用及账户隔离均保留。

Kimi晋升入口只接受固定六组9＋1，schema2审阅证明保留12份测量，取消六路同source/build/commit要求而逐份核验适用性。本轮仅受控inactive审阅，生产目录active=false。v1/v2/v3/v4提案、所有旧原件和授权账本保留。本轮不签付费授权、不激活目录、不启动live/race/压力、不恢复暂停轮次。
''')
print(json.dumps({'proposals':len(refs),'USD':'0.61','executable':False,'account':'PENDING','rates':'PENDING'}))
