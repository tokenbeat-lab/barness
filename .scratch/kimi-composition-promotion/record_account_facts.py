"""Record direct human facts and append a non-executable v5 revision.
No collectors, grant/template tools, tests, or external services are invoked.
"""
from pathlib import Path
from copy import deepcopy
from datetime import datetime, timezone
from decimal import Decimal
import hashlib, json, sys
ROOT=Path(__file__).resolve().parents[2]
BASE=ROOT/'.evidence/kimi-composition-promotion'
SCRATCH=ROOT/'.scratch/kimi-composition-promotion'
sys.path.insert(0,str(ROOT/'tools/verification'))
from receipt_files import source_snapshot
load=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def ref(p):return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
now=datetime.now(timezone.utc).isoformat()
out=BASE/'account-confirmation-v1';out.mkdir()
indexPath=BASE/'pending-scopes-v5.json';index=load(indexPath)
freeze=load(BASE/'build-v2/freeze.json');ordinaryPath=BASE/'e2e/20261028T012642.072979Z-ud9sss5q/receipt.json';ordinary=load(ordinaryPath)
current=source_snapshot(ROOT)
assert current==freeze['source']==ordinary['source']
assert all(sha(ROOT/n)==h for n,h in ordinary['selectedInputs'].items())
oldRecord=load(BASE/'record-v1.json');oldStatus=SCRATCH/'status.md'
assert sha(oldStatus)==oldRecord['currentState']['sha256']
(out/'status-before.md').write_bytes(oldStatus.read_bytes())
publicPins=load(BASE/'initial/public-stores-before.json')
roots=load(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json')
publicActual={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}
assert publicActual=={str(ROOT/p):h for p,h in publicPins.items()}
# Pin old immutable proposals, indexes and reports before appending any revision.
protected={}
for p in [indexPath,BASE/'record-v1.json',BASE/'delivery-v1.json',BASE/'release-v1/release-report.json',SCRATCH/'pending-scopes-v5/README.md',SCRATCH/'delivery-matrix-v4.md']:
 protected[str(p.relative_to(ROOT))]=sha(p)
for f in index['scopeDocuments']:
 assert sha(ROOT/f['path'])==f['sha256'];protected[f['path']]=f['sha256']
assert len(index['scopeDocuments'])==7
save(out/'preserved-v5-before.json',protected)
quote='三个账户组均与旧证据一致，地区、型号权限及额度正常，均为自有 standard 按量 API；Kimi 每站三协议同账户，冻结费率仍适用，无额外收费。先记录这些事实确认并更新七份 v5 待批准范围，保留全部历史证据、授权和账本；暂不签发付费授权，不启动 live、race、压力测试，也不恢复暂停轮次。'
checklist=ROOT/'.scratch/kimi-formal-offline-review/account-rate-checklist.md'
groups=[]
for name,provider,alias,region,model,routes in [
 ('minimax-international','minimax','MINIMAX','global','MiniMax-M3',['minimax-messages']),
 ('kimi-cn','moonshot-cn','KIMI','CN','kimi-k3',['kimi-cn-chat','kimi-cn-responses','kimi-cn-messages']),
 ('kimi-international','moonshot','KIMI','GLOBAL','kimi-k3',['kimi-intl-chat','kimi-intl-responses','kimi-intl-messages'])]:
 groups.append({'group':name,'provider':provider,'accountAlias':alias,'regionAlias':region,'model':model,'routes':routes,'sameActualAccountAsOriginalEvidence':True,'regionUnchanged':True,'modelPermissionNormal':True,'availableQuotaNormal':True,'ownStandardMeteredAPI':True,'sameActualAccountAcrossThreeProtocolsAtThisSite':provider!='minimax','frozenRatesStillApplicable':True,'noAdditionalFees':True,'status':'USER_CONFIRMED'})
facts={'schema':1,'documentType':'USER_ACCOUNT_REGION_RATE_FACT_CONFIRMATION_NOT_AUTHORIZATION','recordedAt':now,'source':{'kind':'direct-user-message-in-current-conversation','verbatim':quote,'verbatimSHA256':hashlib.sha256(quote.encode()).hexdigest(),'attribution':'human user; no external vendor verification claimed'},'previousProposalIndex':ref(indexPath),'retainedRateAndIdentityChecklist':ref(checklist),'accountGroups':groups,'noCrossMarketAccountEqualityInferred':True,'onlyPreviouslyKnownFrozenRateClassesConfirmed':True,'minimaxIndependentPositiveCacheWritePriceRemainsUnknownAndRefused':True,'paidExecutionApproved':False,'paidGrantsIssued':0,'realClaims':0,'vendorCalls':0,'newGrantIDIssued':False,'newGrantExpiryIssued':False,'catalogActivationApproved':False,'liveExecuted':False,'raceExecuted':False,'pressureExecuted':False,'pausedResumed':False}
factsPath=out/'facts.json';save(factsPath,facts)
revised=SCRATCH/'pending-scopes-v5-confirmed-v1';revised.mkdir()
refs=[];changes=[];total=Decimal('0')
for old in index['scopeDocuments']:
 p=ROOT/old['path'];before=load(p);d=deepcopy(before)
 assert not any(d[k] for k in ['approved','grantIssued','executable'])
 assert d['authorizationStillRequired']['id']=='NOT_ISSUED' and d['authorizationStillRequired']['expiresAt']=='NOT_ISSUED'
 route=d['route'];group=next(g for g in groups if route in g['routes'])
 d['previousProposal']=ref(p)
 d['confirmations'].update(accountAndRegion='USER_CONFIRMED',ratesAndServiceTier='USER_CONFIRMED',modelPermissionsAndQuota='USER_CONFIRMED',confirmationEvidence=ref(factsPath),recordedAt=now,accountGroup=group['group'])
 d['confirmations']['factsConfirmationIsNotPaidAuthorization']=True
 d['identityToConfirm']['confirmation']='USER_CONFIRMED: same original own standard metered API account and region, normal model permissions/quota, frozen rates applicable and no additional fees; see confirmations.confirmationEvidence. Paid authorization remains NOT_GRANTED.'
 d['status']='AWAITING_EXPLICIT_PAID_APPROVAL'
 # Every machine scope, binary, original measurement, price/cap and stop rule
 # is kept verbatim. Only factual confirmation and predecessor metadata change.
 allowed={'confirmations','identityToConfirm','status','previousProposal'}
 changed={k for k in set(before)|set(d) if before.get(k)!=d.get(k)}
 assert changed<=allowed
 bIdentity=deepcopy(before['identityToConfirm']);nIdentity=deepcopy(d['identityToConfirm']);bIdentity.pop('confirmation');nIdentity.pop('confirmation');assert bIdentity==nIdentity
 for k in set(before)-allowed:assert d[k]==before[k],(route,k)
 assert d['ordinarySource']==current
 assert d['executionScope']['calls']==d['executionScope']['httpAttempts']==1
 total+=Decimal(d['executionScope']['proposedMaxUSD'])
 dest=revised/(route+'.json');save(dest,d);refs.append(ref(dest));changes.append({'route':route,'changedFields':sorted(changed),'previous':ref(p),'updated':ref(dest),'paidGrantIssued':False,'executable':False})
assert total==Decimal('0.61')
newIndex=deepcopy(index);newIndex.update(previousIndex=ref(indexPath),confirmationEvidence=ref(factsPath),proposalRevision='v5-facts-confirmed-1',readiness='FACTS_CONFIRMED_AWAITING_EXPLICIT_PAID_APPROVAL',accountAndRegionConfirmation='USER_CONFIRMED',ratesAndTierConfirmation='USER_CONFIRMED',scopeDocuments=refs)
assert newIndex['status']=='NOT_GRANTED' and newIndex['grantsIssued']==newIndex['realClaims']==newIndex['vendorCalls']==0
newIndexPath=BASE/'pending-scopes-v5-confirmed-v1.json';save(newIndexPath,newIndex)
(revised/'README.md').write_text('''# 七份待批准范围 v5：事实确认修订1

用户已在本轮直接确认：MiniMax国际、Kimi国内、Kimi国际三个账户组均与各自旧证据一致；地区、型号权限及额度正常；均为自有standard按量API；Kimi每站三协议同账户；冻结费率仍适用且无额外收费。确认依据见[事实原件](../../../.evidence/kimi-composition-promotion/account-confirmation-v1/facts.json)。事实仅按用户确认记录，没有声称新增厂商核验。

七份均更新为 accountAndRegion/ratesAndServiceTier/modelPermissionsAndQuota=USER_CONFIRMED；付费状态仍NOT_GRANTED／AWAITING_EXPLICIT_PAID_APPROVAL，approved=false、grantIssued=false、executable=false，授权ID和有效期仍NOT_ISSUED。

| 路线 | 仅补场景 | 调用/HTTP | 拟议USD |
| --- | --- | --- | --- |
| minimax-messages | thinking-on | 1/1 | 0.01 |
| kimi-cn-chat | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-chat | xai-origin-history | 1/1 | 0.10 |
| kimi-cn-responses | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-responses | xai-origin-history | 1/1 | 0.10 |
| kimi-cn-messages | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-messages | xai-origin-history | 1/1 | 0.10 |

合计仍为7调用／7HTTP、拟议USD0.61，七个独立预算，全部零重试。MiniMax保留2048输入/输出、调用120s/进程180s；Kimi保留4096输入/输出、调用120s/进程240s、真实首发等待60s。原54＋13项、基包manifest、SDK、目录、价格与正常冻结二进制绑定均未改。本次不新增未知MiniMax独立缓存写价格，出现正写量仍拒绝资格；其他停止条件不变。

[原v5](../pending-scopes-v5/README.md)与所有旧索引、失败、证据、授权和账本在原路径保留。原v5 README标题误写v4也作为原件保留，本修订标题已正确标注。今后需另行明确批准才可逐路线签新grant；本轮仅记录事实与修订提案，没有签付费授权、真实claim、厂商调用、目录激活、live/race/压力或暂停恢复。
''')
(SCRATCH/'account-facts-update-spec.md').write_text('''# v5账户地区费率事实确认更新

范围：直接记录用户已确认的三个账户组与冻结费率事实，为七份未批准v5生成不可覆盖的事实确认修订1。保留原v5路径/哈希与历史状态快照；仅改确认和前驱元数据，不改场景、预算、价格、SDK、原证据或正常冻结二进制。

验证为静态结构/逐字段差异、完整来源和185选定输入相等性、引用哈希、公共授权账本集合与字节保全；沿用原完整普通通过回执，不重跑测试、不重建二进制。确认不签发任何付费权限。新当前状态另绑定record-v2并只读判定，真实七项仍NOT_RUN、release.pass=false。旧人工阶段及暂停状态保持。
''')
issue=SCRATCH/'issues/02-v5-account-facts.md';issue.write_text('Status: claimed\nType: task\n\n# 记录账户地区费率确认并修订七份v5\n\n范围与离线验证见 ../account-facts-update-spec.md；不签付费授权、不执行测试。\n')
# A new matrix describes current facts; prior matrix and its indexes stay pinned.
oldMatrix=SCRATCH/'delivery-matrix-v4.md';text=oldMatrix.read_text().replace('交付矩阵 v4','交付矩阵 v5（账户费率事实已确认）').replace('pending-scopes-v5/README.md','pending-scopes-v5-confirmed-v1/README.md').replace('待账户/费率确认及独立批准','账户/费率已确认，待独立批准').replace('待独立批准；13＋1已实现','账户/费率已确认，待独立批准；13＋1已实现').replace('先确认三个账户组与费率，随后分别批准七个独立补项','三个账户组、地区/权限/额度及冻结费率已获用户确认；今后仍须分别明确批准七个独立补项')
newMatrix=SCRATCH/'delivery-matrix-v5.md';newMatrix.write_text(text)
# Update the feature's single mutable current state, retaining its original bytes.
status=oldStatus.read_text().replace('[七份待批准范围v5](pending-scopes-v5/README.md)','[七份v5事实确认修订1](pending-scopes-v5-confirmed-v1/README.md)').replace('[29路径矩阵v4](delivery-matrix-v4.md)','[29路径矩阵v5](delivery-matrix-v5.md)')
status=status.replace('账户/地区/费率保持PENDING，七份scope仍approved=false、executable=false、ID/有效期NOT_ISSUED。','三个账户组、地区/型号权限/额度、自有standard按量API、Kimi每站三协议同账户及冻结费率/无额外收费，均已由用户直接确认；七份scope事实状态为USER_CONFIRMED，付费仍NOT_GRANTED，approved=false、grantIssued=false、executable=false、ID/有效期NOT_ISSUED。')
status+='''\n## 本轮账户、地区与费率事实确认\n\n2026-10-28：本轮仅更新用户事实确认与提案元数据，未修改产品/采集/消费者源码、SDK、型号、价格、预算、旧证据或冻结二进制。原v5七份与旧索引保留，新当前范围为v5事实确认修订1。确认依据见[事实原件](../../.evidence/kimi-composition-promotion/account-confirmation-v1/facts.json)，变更与保全见[离线核验](../../.evidence/kimi-composition-promotion/account-confirmation-v1/verification.json)。\n\n原当前状态完整快照见[status-before](../../.evidence/kimi-composition-promotion/account-confirmation-v1/status-before.md)，与record-v1原currentState哈希一致；旧record/delivery/release报告原件保留。更新后的状态由[record-v2](../../.evidence/kimi-composition-promotion/record-v2.json)绑定，新的[只读发布报告](../../.evidence/kimi-composition-promotion/release-v2/release-report.json)继续拒绝未完成真实补项。原完整普通证据与当前2048文件来源、185选定输入保持一致，不重跑普通、受控、live、race或压力。\n\n账户事实前置已解除；下一项需要用户另行明确付费批准，拟议7次调用/7HTTP/USD0.61及各路线停止条件保持。真实补项、目录激活和历史人工阶段仍未完成；本次没有签发ID/有效期、付费授权、真实claim、厂商请求、目录激活或恢复暂停轮次。MiniMax未知独立正缓存写价格继续拒绝，不把“无额外收费”推定为新价格或免费计费项。\n'''
oldStatus.write_text(status)
record=deepcopy(oldRecord);record['currentState']=ref(oldStatus);assert record['source']==current;save(BASE/'record-v2.json',record)
# Recheck public authority and every old pinned object after the only permitted
# mutable state edit. No old proposal, grant, ledger or measurement is updated.
assert all(sha(ROOT/n)==h for n,h in protected.items())
assert {str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}==publicActual
assert source_snapshot(ROOT)==current and all(sha(ROOT/n)==h for n,h in ordinary['selectedInputs'].items())
for f in refs:
 d=load(ROOT/f['path']);assert not any(d[k] for k in ['approved','executable','grantIssued']);assert d['authorizationStillRequired']['id']==d['authorizationStillRequired']['expiresAt']=='NOT_ISSUED'
verify={'kind':'static-v5-human-facts-confirmation-update','pass':True,'recordedAt':now,'facts':ref(factsPath),'previousIndex':ref(indexPath),'currentIndex':ref(newIndexPath),'changes':changes,'source':current,'unchangedSourceFiles':len(current['files']),'ordinaryReceiptReused':ref(ordinaryPath),'ordinarySelectedInputsExact':len(ordinary['selectedInputs']),'noOrdinaryRerunRequired':'Only user fact/proposal metadata changed; source, execution scope and ordinary inputs are identical.','oldV5AndPriorReportsExact':True,'oldCurrentStateSnapshot':ref(out/'status-before.md'),'oldRecordCurrentStateSHA256':oldRecord['currentState']['sha256'],'publicAuthorityFiles':len(publicActual),'publicAuthorityExactSetAndBytesUnchanged':True,'totalProposedCalls':7,'totalProposedHTTPAttempts':7,'totalProposedUSD':str(total),'paidGrantsIssued':0,'realClaims':0,'vendorCalls':0,'testExecutions':0,'catalogActive':False,'liveExecuted':False,'raceExecuted':False,'pressureExecuted':False,'pausedResumed':False}
assert sha(ROOT/'ai/catalog_kimi.json')=='c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109'
save(out/'verification.json',verify)
print(json.dumps({'facts':str(factsPath.relative_to(ROOT)),'currentIndex':str(newIndexPath.relative_to(ROOT)),'scopes':len(refs),'confirmedGroups':len(groups),'USD':str(total),'publicAuthorityFiles':len(publicActual),'sourceExact':True,'testsExecuted':0,'paidGrantsIssued':0},ensure_ascii=False))
