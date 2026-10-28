"""Prepare non-executable paid-scope proposals and a fixed 29-path matrix."""
import collections,hashlib,json,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/provider-final-readiness';SCRATCH=ROOT/'.scratch/provider-final-readiness'
sys.path.insert(0,str(ROOT/'tools/verification'));from receipt_files import source_snapshot
load=lambda p:json.loads(Path(p).read_text())
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def save(p,x):
 with Path(p).open('x') as f:json.dump(x,f,ensure_ascii=False,indent=2,sort_keys=True);f.write('\n')
indexPath=ROOT/'.evidence/kimi-history-supplement/current-qualification-index-v4.json';index=load(indexPath)
facts=load(BASE/'native-facts-v2/stdout.jsonl');fmap={r['route']:r for r in facts['routes']}
source=source_snapshot(ROOT)
# Existing normal live-tag producer retained by the completed public E2E.
ordinary=ROOT/'.evidence/provider-final-readiness/e2e/20261027T133012.910360Z-a311m5zr/receipt.json';oldrun=load(ordinary);main=next(x for x in oldrun['bundles'] if x['package']=='./ai/e2e');mainDir=ordinary.parent/main['path'];manifest=load(mainDir/'manifest.json')
pointers={}
for c in manifest['cases']:
 if c['id'].startswith('KHS-composition-'):pointers[c['id'].removeprefix('KHS-composition-')]=load(mainDir/c['dir']/'composition-pointer.json')['descriptor']
proposals=SCRATCH/'kimi-scopes-v2';proposals.mkdir();scopeRefs=[]
for old in index['partialKimiRecords']:
 route=old['route'];factsRow=fmap[route];controlled=pointers[route]['supplement'];r=load(ROOT/controlled['bundle']/'regional-report.json');producer=load(ROOT/controlled['producer']['path']);binary=Path(producer['argv'][0])
 assert sha(binary)==r['regional']['binarySha256']
 assert load(ROOT/controlled['bundle']/'regional-execution-source.json')==facts['nativeSource']
 assert r['regional']['buildSourceSha256']==r['regional']['sourceSha256']==facts['nativeSourceSHA256']
 baseline=ROOT/old['bundle'];grant=load(baseline/'regional-grant.json');br=load(baseline/'regional-report.json');scene=br['scenarios'][0];meta=load(baseline/scene['caseId']/'result.json')['metadata']
 pending={'documentType':'KIMI_PAID_SCOPE_PROPOSAL_NOT_A_GRANT','schemaVersion':1,'status':'AWAITING_EXPLICIT_PAID_APPROVAL','approved':False,'grantIssued':False,'executable':False,'route':route,'model':'kimi-k3','purpose':'xai-history-supplement','facts':factsRow,'ordinarySource':source,'nativeSourceSHA256':facts['nativeSourceSHA256'],'frozenProducer':ref(binary),'frozenProducerEvidence':{'nativeReport':ref(ROOT/controlled['bundle']/'regional-report.json'),'producerReceipt':controlled['producer'],'binaryOrigin':'retained normal live-tag binary; controlled execution is mechanism proof only'},'baseline':{'bundle':old['bundle'],'manifest':ref(baseline/'manifest.json'),'report':ref(baseline/'regional-report.json'),'seal':ref(baseline/'regional-integrity.json'),'retainedActualCases':9,'applicability':old['baselineQualification']},'identityToConfirm':{'provider':factsRow['provider'],'accountAlias':grant['accountAlias'],'regionAlias':grant['regionAlias'],'tenantID':meta['tenantId'],'accountScopeID':meta['accountScopeId'],'confirmation':'REQUIRED from the trusted Host before authorization; identical aliases across markets do not prove identical accounts','accountPurpose':'own standard API only'},'executionScope':{'scenarioIDs':['kimi-k3/xai-origin-history'],'entry':'^TestKimiHistorySupplement$','testCount':1,'testTimeoutSeconds':240,'modeRequiredIfAuthorized':'live','stageRequiredIfAuthorized':'candidate','calls':1,'httpAttempts':1,'sdkRetries':0,'environmentRetries':0,'maxReservedInputTokens':4096,'maxRequestedOutputTokens':4096,'callSeconds':120,'processSeconds':240,'callIntervalMs':60000,'proposedMaxUSD':'0.10','currency':'USD','priceBasis':'kimi-prices-2026-10-22.1, retained estimate, not a current invoice'},'authorizationStillRequired':{'newIndependentGrant':True,'expectedSchema':2,'expectedContract':'kimi/'+route+'/xai-origin-history/v1','id':'NOT_ISSUED','expiresAt':'NOT_ISSUED','recommendedValidityMinutes':30,'maximumValidityHours':24,'sixScopesTogetherMaxUSD':'0.60','noOldBudgetReuse':True},'stopConditions':['failure/cancellation/timeout/native validation rejection: stop remaining paid work','price/model/SDK/source/binary/account/baseline/input differs: return to offline review','no automatic retry/relabel/renewal or paused batch recovery'],'excludes':['MiniMax','xAI requests','directory activation','formal full rerun','race','pressure','paused-round resume']}
 # This deliberately lacks the RegionalGrant schema and top-level fields.
 assert 'schema' not in pending and 'id' not in pending and 'expiresAt' not in pending
 p=proposals/(route+'.json');save(p,pending);scopeRefs.append(ref(p))
rows={}
for q in index['qualified']:
 receipt=ROOT/q['receipt'];proof=load(receipt);assert sha(receipt)==q['receiptSHA256'];bundle=ROOT/proof['bundle'];name=q['route'];route='typesafe-classifier' if name.startswith('typesafe-classifier') else name
 if (bundle/'live-report.json').exists():
  r=load(bundle/'live-report.json');provider,api,op=r['provider'],r['api'],r['operation'];cases=r['scenarios'];counts=dict(collections.Counter(c['outcome'] for c in cases));models=sorted({c.get('model',r['model']) for c in cases})
 else:
  r=load(bundle/'minimax-report.json');provider,api,op='minimax',r['identity']['api'],'chat';models=[r['identity']['model']];cases=r['expected'];counts={'PASS':len(cases)}
 if route not in rows:rows[route]={'route':route,'provider':provider,'api':api,'operation':op,'evidenceState':'APPLICABLE_COMPLETE','fullPathQualification':True,'actualEvidence':[],'minimumNewActualScenarios':[],'newPaidCallsThisTask':0,'deliveryState':'HISTORICAL_BASELINE_RETAINED_CURRENT_EVIDENCE_APPLICABLE'}
 rows[route]['actualEvidence'].append({'record':name,'models':models,'requiredContractItems':len(cases),'statuses':counts,'consumer':ref(receipt),'bundle':str(bundle.relative_to(ROOT))})
 if provider in ['zai','minimax-cn']:rows[route]['deliveryState']='QUALIFICATION_APPLICABLE_REGIONAL_PROJECT_CLOSURE_PENDING'
 if provider=='xai':rows[route]['deliveryState']='QUALIFICATION_APPLICABLE_GROK_RELEASE_PENDING'
for old,proposal in zip(index['partialKimiRecords'],scopeRefs):
 f=fmap[old['route']];rows[old['route']]={'route':old['route'],'provider':f['provider'],'api':f['api'],'operation':'chat','models':['kimi-k3'],'evidenceState':'PARTIAL_9_OF_10','fullPathQualification':False,'requiredContractItems':10,'applicableActualItems':9,'baseline':old['manifest'],'minimumNewActualScenarios':['kimi-k3/xai-origin-history'],'minimumProposedCalls':1,'proposedUSD':'0.10','proposal':proposal,'newPaidCallsThisTask':0,'deliveryState':'CANDIDATE_ONLY_MISSING_REAL_SCENE_AND_CATALOG_PROMOTION'}
mini=ROOT/'.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id';mb=next((mini/'bundles').glob('*/minimax-report.json'));mr=load(mb);partial=load(BASE/'native-analysis-v3/stdout.jsonl');assert partial['pass'] and len(partial['cases'])==13
rows['minimax-messages']={'route':'minimax-messages','provider':'minimax','api':'anthropic-messages','operation':'chat','models':['MiniMax-M3'],'evidenceState':'PARTIAL_13_OF_14_ORIGINAL_FAIL','fullPathQualification':False,'requiredContractItems':14,'applicableActualItems':13,'originalReport':ref(mb),'originalProducer':ref(mini/'receipt.json'),'nativeCaseAnalysis':ref(BASE/'native-analysis-v3/receipt.json'),'minimumNewActualScenarios':['thinking-on'],'minimumFactGapCount':1,'actualPaidCallCap':None,'USD':None,'executionReadiness':'no single-scene producer/composition contract implemented for MiniMax; fixed current full plan has14 contract items','newPaidCallsThisTask':0,'deliveryState':'HISTORICAL_DELIVERY_RETAINED_NEW_THINKING_ON_CAPTURE_FAILED'}
assert len(rows)==29 and len({r['provider'] for r in rows.values()})==12
matrix={'kind':'provider-path-delivery-matrix','schema':1,'source':source,'previousQualificationIndex':ref(indexPath),'scope':'12 Provider identities,29 operation/provider/API paths;TypeSafe1.13/1.14 evidence share one path','counts':{'providers':12,'paths':29,'completeApplicablePaths':22,'completeRouteModelEvidenceRecords':23,'partialKimiPaths':6,'partialFailedMiniMaxPaths':1},'paths':[rows[k] for k in sorted(rows)],'noNewVendorEvidence':True,'grantsIssued':0,'realClaims':0,'pausedRoundsResumed':0,'releasePass':False,'releaseBlockers':['Kimi six actual scene gaps','Kimi own formal catalog review/promotion and formal read-only qualifications','MiniMax Messages thinking-on incomplete plus single-scene execution decision','impact-scoped human race/pressure and final release record;do not restore prior17-route batch']}
save(BASE/'delivery-matrix-v2.json',matrix)
lines=['# 29条 Provider 路径交付矩阵','','口径：12个Provider身份、29条操作×Provider×协议路径；TypeSafe两个版本共用一条路径。22条完整适用证据、6条Kimi各9/10、1条MiniMax Messages13/14。适用证据不是整体发布完成；所有旧事实保留，release.pass=false。','','来源与逐记录原件见 ../../.evidence/provider-final-readiness/delivery-matrix-v2.json。合同项数包含零请求预检/声明不支持，不等于付费调用数。','','| Provider | 路线 | 协议 | 真实证据/合同项 | 当前状态与最小缺口 |','| --- | --- | --- | --- | --- |']
for row in matrix['paths']:
 if 'actualEvidence' in row:
  counts='；'.join('/'.join(x['models'])+': '+str(x['requiredContractItems'])+'项 '+','.join(str(v)+' '+k for k,v in x['statuses'].items()) for x in row['actualEvidence']);state='完整适用；新增补采0'
  if row['provider']=='xai':state+='；Grok发布收尾待完'
  if row['provider'] in ('zai','minimax-cn'):state+='；地区项目收尾待完'
 else:
  counts=str(row['applicableActualItems'])+'/'+str(row['requiredContractItems']);state='缺 '+', '.join(row['minimumNewActualScenarios'])
  if row['provider'].startswith('moonshot'):state+='；目录inactive'
 lines.append('| '+row['provider']+' | '+row['route']+' | '+row['api']+' | '+counts+' | '+state+' |')
lines+=['','最小真实事实缺口共7个：Kimi6、MiniMax thinking-on1。已具备单场景执行合同的是Kimi6次/USD0.60提案；MiniMax的执行调用数和费用尚未核定，不写成已经可执行的7调用授权。race/压力按影响另选，旧暂停轮次不恢复。']
with (SCRATCH/'delivery-matrix-v2.md').open('x') as f:f.write('\n'.join(lines)+'\n')
save(BASE/'kimi-pending-scopes-v2.json',{'documentType':'PROPOSAL_INDEX_NOT_AUTHORIZATION','status':'NOT_GRANTED','scopeDocuments':scopeRefs,'proposedCalls':6,'proposedHTTPAttempts':6,'sdkAndEnvironmentRetries':0,'proposedTotalUSD':'0.60','priceBasis':'frozen kimi-prices-2026-10-22.1','proposedSerialNativeProcessSeconds':1440,'grantsIssued':0,'vendorCalls':0,'claims':0})
print(json.dumps({'paths':29,'completeApplicable':22,'KimiProposals':len(scopeRefs),'USD':'0.60','grantsIssued':0}))
