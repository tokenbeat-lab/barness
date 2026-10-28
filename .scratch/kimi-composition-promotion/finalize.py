"""Bind final ordinary evidence without converting controlled facts to live."""
from pathlib import Path
import json,hashlib,sys,collections
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/kimi-composition-promotion';SCRATCH=ROOT/'.scratch/kimi-composition-promotion'
sys.path.insert(0,str(ROOT/'tools/verification'));from receipt_files import source_snapshot
load=lambda p:json.loads(p.read_text());sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,ensure_ascii=False,indent=2);f.write('\n')
# Only a complete eight-package attempt qualifies. Earlier failures stay intact.
selected=load(BASE/'full-ordinary-selection-v3.json');expected=selected['argv'][selected['argv'].index('--')+1:]
rs=[]
for p in sorted((BASE/'e2e').glob('*/receipt.json')):
 r=load(p)
 if r['argv']==expected and r['status']=='PASS' and r['exitCode']==0:rs.append((p,r))
assert len(rs)==1,len(rs)
rp,r=rs[0];current=source_snapshot(ROOT);freeze=load(BASE/'build-v2/freeze.json')
assert current==r['source']==freeze['source']
assert all(sha(ROOT/n)==h for n,h in r['selectedInputs'].items())
main=[b for b in r['bundles'] if b['package']=='./ai/e2e'];assert len(main)==1
bundle=rp.parent/main[0]['path'];manifest=load(bundle/'manifest.json');counts=collections.Counter(c['status'] for c in manifest['cases']);assert counts['FAIL']==0
required=['KCP-six-paired-controlled-promotion','KIMI-K16-standalone-promotion-refused']+['KHS-composition-'+x for x in ['kimi-cn-chat','kimi-cn-messages','kimi-cn-responses','kimi-intl-chat','kimi-intl-messages','kimi-intl-responses']]
caseMap={c['id']:c for c in manifest['cases']}
for c in required:assert caseMap[c]['status']=='PASS' and load(bundle/c/'assertions.json')['assertions']
assert all(a['pass'] for a in load(bundle/required[0]/'assertions.json')['assertions'])
proposal=load(bundle/required[0]/'controlled-proposal.json')
assert proposal['schema']==2 and proposal['mode']=='controlled' and not proposal['catalog']['active'] and len(proposal['candidates'])==6
measurements=[m for c in proposal['candidates'] for m in c['qualification']['measurements']]
assert len(measurements)==12 and len({m['original']['gitCommit'] for m in measurements})==2
assert all(len(c['qualification']['coverage'])==10 and len(c['inputs'])>100 for c in proposal['candidates'])
aux=[]
for b in r['bundles']:
 p=rp.parent/b['path'];m=load(p/'manifest.json');assert not load(p/'audit.json')['findings']
 aux.append({'package':b['package'],'counts':dict(collections.Counter(c['status'] for c in m['cases'])),'manifest':ref(p/'manifest.json'),'audit':ref(p/'audit.json')})
controlled={}
for m in measurements:
 if m['kind']=='history-one' and m['original']['regional']['binarySha256']==freeze['kimi']['sha256']:
  controlled[m['original']['combo']]={'native':m['evidence'],'binarySHA256':freeze['kimi']['sha256'],'mode':'controlled'}
for p in rp.parent.rglob('minimax-report.json'):
 d=load(p)
 if d.get('kind')=='minimax-thinking-on-supplement' and d.get('mode')=='controlled' and d.get('outcome')=='PASS':
  g=load(p.parent/'process-grant.json')
  if g['scope']['binarySha256']==freeze['minimax']['sha256']:
   controlled['minimax-messages']={'report':ref(p),'producer':ref(p.parent.parent/'producer-receipt.json'),'binarySHA256':freeze['minimax']['sha256'],'mode':'controlled'};break
assert len(controlled)==7,controlled.keys()
save(BASE/'frozen-controlled-v1.json',{'kind':'exact-normal-binary-controlled-proofs','measurements':controlled,'promotionProof':ref(bundle/required[0]/'controlled-proposal.json'),'independentMeasurements':12,'vendorCalls':0})
for key in ['minimax','kimi','releasegate','regionalpromote','processgrant']:assert sha(ROOT/freeze[key]['path'])==freeze[key]['sha256']
for p in (BASE/'build-v2/07-minimax-metadata/stdout',BASE/'build-v2/08-kimi-metadata/stdout'):
 s=p.read_text();assert '-tags=live\n' in s and 'kimi_baseline_fixture' not in s
pending=load(BASE/'pending-scopes-v5.json');assert len(pending['scopeDocuments'])==7
for refd in pending['scopeDocuments']:
 p=ROOT/refd['path'];assert sha(p)==refd['sha256'];d=load(p)
 assert not d['approved'] and not d['executable'] and not d['grantIssued'] and d['authorizationStillRequired']['id']=='NOT_ISSUED'
 assert d['confirmations']['accountAndRegion']=='PENDING' and d['confirmations']['ratesAndServiceTier']=='PENDING' and d['ordinarySource']==current
assert load(BASE/'retention-v2.json')['pass']
assert load(BASE/'frozen-promotion-v2/receipt.json')['exitCode']==0
for p in [BASE/'proposal-validation-v2/receipt.json',BASE/'kimi-original-producers-v2/receipt.json',BASE/'original-baseline-v6/receipt.json']:assert load(p)['exitCode']==0
assert all(x['exitCode']==0 for x in load(BASE/'tools-vet-v2/receipt.json')['results'])
before=load(BASE/'initial/source-before.json')['files'];now=current['files'];diff=[n for n in sorted(set(before)|set(now)) if before.get(n)!=now.get(n)]
product=[n for n in diff if n in ['go.mod','go.sum'] or n.startswith('ai/') and not n.startswith(('ai/e2e/','ai/internal/testkit/','ai/live/','ai/release/'))];assert not product
attempts=[]
for p in sorted((BASE/'e2e').glob('*/receipt.json')):
 d=load(p);attempts.append({'receipt':ref(p),'status':d['status'],'exitCode':d['exitCode'],'argv':d['argv'],'elapsedSeconds':d['elapsedSeconds']})
save(BASE/'attempts-v1.json',{'attempts':attempts,'notes':['RED: old consumer cannot find regional-report.json in fixed descriptor directory','first GREEN attempt exposed a new fixture relative-path error; fixture corrected and failed attempt retained','first full run stopped at TestGLMFullRollback because concurrent scratch document generation changed worktree; final rerun keeps worktree quiet','read-only proposal preparation initially expected blank ID rather than explicit invalid placeholder; both templates retained','second full run intentionally terminated (-15) after finding transitive-original output-path gap; additional public RED/GREEN retains the failure and required guard',
'first frozen controlled replay supplied relative descriptor references instead of the absolute input directories used by the public E2E; output path guard refused it. Corrected replay uses the same absolute-argument contract, no implementation change',
'retention first scanner compared absolute and relative authority keys; canonicalized without changing authorities']})
v={'kind':'fixed-six-kimi-composition-promotion','source':current,'sourceChanges':diff,'productSourceChanges':product,'ordinary':ref(rp),'nativeInventories':aux,'ordinaryMainCounts':dict(counts),'selectedInputs':len(r['selectedInputs']),'selectedInputsExact':True,'requiredCases':required,'pairedControlledProof':ref(bundle/required[0]/'controlled-proposal.json'),'independentControlledMeasurements':12,'controlledPromotionAssertions':len(load(bundle/required[0]/'assertions.json')['assertions']),'frozenBuilds':ref(BASE/'build-v2/freeze.json'),'exactFrozenPromotion':ref(BASE/'frozen-promotion-v2/receipt.json'),'exactBinaryControlled':ref(BASE/'frozen-controlled-v1.json'),'pendingScopes':ref(BASE/'pending-scopes-v5.json'),'proposalCannotAuthorize':ref(BASE/'proposal-validation-v2/receipt.json'),'oldKimi54':ref(BASE/'kimi-original-producers-v2/receipt.json'),'oldMiniMax13':ref(BASE/'original-baseline-v6/receipt.json'),'toolsVet':ref(BASE/'tools-vet-v2/receipt.json'),'retention':ref(BASE/'retention-v2.json'),'attempts':ref(BASE/'attempts-v1.json'),'live':'NOT_RUN','race':'NOT_EXECUTED','pressure':'NOT_EXECUTED','remainingVendorFacts':7,'catalogActive':False,'paidGrantsIssued':0,'realClaims':0,'vendorCalls':0,'pausedResumed':False,'releasePass':False}
save(BASE/'verification-v1.json',v)
(BASE/'status-before-delivery-v1.md').write_bytes((SCRATCH/'status.md').read_bytes())
(SCRATCH/'status.md').write_text(f'''# Kimi 固定组合晋升当前状态

2026-10-28：实施、受控与完整普通验证已完成。公共 regionalpromote -kimi 只接受固定六组9＋1；schema2专用证明保留十二份独立测量、每组十项coverage及完整传递库存。删除旧单包晋升与六路同source/build/binary/commit拦截，不新增通用框架。各原测量的来源/SDK/型号/账户/地区/租户、授权/预算/时间、真实producer及权威账本仍分别核验。其他地区晋升合同保持。

[完整验证](../../.evidence/kimi-composition-promotion/verification-v1.json)、[七份待批准范围v5](pending-scopes-v5/README.md)、[29路径矩阵v4](delivery-matrix-v4.md)、[保全核验](../../.evidence/kimi-composition-promotion/retention-v2.json)。本状态为本功能唯一当前状态；旧功能状态与所有旧版本保留。

普通测试实际退出0，主包{dict(counts)}；完整八包回执 {rp.relative_to(ROOT)}，墙钟{r['elapsedSeconds']:.3f}s，来源{len(current['files'])}文件、选定输入{len(r['selectedInputs'])}文件均与冻结/当前集合相等。受控晋升{v['controlledPromotionAssertions']}项断言全部通过，包括不同实际commit/binary和全部拒绝边界。{len(required)}项有限必需case绑定原生库存。工具56项及normal/live-tag compile-only vet退出0；未运行厂商live。

受测和当前HEAD均{current['commit']}；工作区原有修改保留，本轮源码差异见verification-v1，不仅凭HEAD声称一致。Kimi/MiniMax正常live标签二进制、releasegate、regionalpromote、processgrant已冻结，完整普通中七个单项受控进程与冻结字节相同；旧54项及13项由当前原生消费者离线通过。REGIONAL.md原字节另存initial，旧ADR/SOP两份历史修订也有原件，保全扫描没有未解释变更；六处公共授权/账本8156文件集合及字节相同。

真实缺口仍为Kimi六个xai-origin-history与MiniMax一个thinking-on，共7项NOT_RUN；22/29路径完整适用，Kimi六路各9/10、MiniMax13/14。账户/地区/费率保持PENDING，七份scope仍approved=false、executable=false、ID/有效期NOT_ISSUED。拟议7调用/7HTTP/USD0.61，各自预算独立。仅事实确认不授予付费权限。

生产ai/catalog_kimi.json始终active=false且字节未变；没有签发付费授权、真实claim、厂商请求、live/race/压力执行或暂停恢复。此变更是现有激活锁/原子替换外围的只读证据消费接线，没有新增共享可变状态或负载；本功能race/压力有依据地NOT_APPLICABLE，不替代旧功能尚缺人工阶段。六路真实补项、目录独立审阅及真实资格尚未完成，release.pass=false，见只读release报告。

所有RED、夹具路径修复、并行scratch写入触发回滚保护的完整失败attempt、模板与核验脚本失败均保留并说明；最终普通运行期间没有工作区写入。最终冻结regionalpromote另行只读复现同一受控正例，输出与普通回归证明逐字节一致。回滚只撤本轮差异并停用新未批准scope/二进制，不删除原件、账本或新旧失败，也不加兼容shim。
''')
issue=SCRATCH/'issues/01-composition-promotion.md';issue.write_text(issue.read_text().replace('Status: claimed','Status: resolved')+'\n## Answer\n\n固定六组组合晋升已实现并通过完整普通验证，详见 ../status.md。真实七项、目录激活及付费权限均未执行。\n')
routes=sorted(c['combo'] for c in proposal['candidates'])
record={'schema':2,'feature':'kimi-composition-promotion','currentState':ref(SCRATCH/'status.md'),'source':current,'selectedInputs':r['selectedInputs'],'impact':{'liveRoutes':routes,'raceRequired':False,'pressureRequired':False,'reason':'Read-only Kimi fixed-pair evidence consumption; existing activation lock and atomic replacement unchanged, no new shared state or resource/load path. Actual six supplements remain absent; other providers product/SDK/model behavior unchanged.'},'e2e':{'status':'PASS','receipt':ref(rp),'requiredCases':required},'live':{'status':'NOT_RUN','reason':'six actual supplements and activation not authorized','routes':{x:{'status':'NOT_RUN','reason':'missing xai-origin-history actual supplement'} for x in routes}},'race':{'status':'NOT_APPLICABLE','reason':'read-only evidence consumer; no shared mutation changes, existing activation write/lock untouched; user prohibited execution'},'pressure':{'status':'NOT_APPLICABLE','reason':'fixed six bounded input compositions; no runtime resource or load path changes; user prohibited execution'}}
save(BASE/'record-v1.json',record)
print({'mainCounts':dict(counts),'requiredCases':len(required),'controlledFrozen':len(controlled),'sourceFiles':len(current['files']),'selectedInputs':len(r['selectedInputs']),'elapsedSeconds':r['elapsedSeconds']})
