"""Read-only retention audit and fixed-route delivery records; no paid launch."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,sys,time
R=Path(__file__).resolve().parents[2];B=R/'.evidence/minimax-thinking-resample';S=R/'.scratch/minimax-thinking-resample'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False);f.write('\n')
tic=time.monotonic();result=load(B/'composition-result-v1.json');assert result['pass'] and not (B/'STOP.json').exists();pre=load(B/'preflight.json');source=source_snapshot(R);assert source==pre['source']==result['source']
protected=load(B/'historical-files-before.json');missing=[];bad=[]
for n,h in protected.items():
 p=R/n
 if not p.is_file():missing.append(n)
 elif sha(p)!=h:bad.append(n)
assert not missing and not bad,(missing[:5],bad[:5])
info=load(B/'attempt-index.json');rp=R/info['nativeReceipt']['path'];receipt=load(rp);native=rp.parent/receipt['bundles'][0]['path'];report=load(native/'minimax-report.json');grant=load(B/'grant.json');claim=load(native/'process-authorization-receipt.json');assert report['processGrantSha256']==claim['grantSha256']==sha(B/'grant.json')==sha(native/'process-grant.json')
roots=load(R/'.evidence/minimax-thinking-supplement-plan/store-roots.json');actual={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()};before=load(B/'authority-before.json');assert all(actual.get(n)==h for n,h in before.items())
store=R/'.git/barness-live-process-grants';expected={str(store/('id-'+hashlib.sha256(grant['id'].encode()).hexdigest()+'.json')),str(store/('scope-'+claim['scopeSha256']+'.json')),str(store/('ledger-'+claim['scopeSha256']+'.jsonl'))};assert set(actual)-set(before)==expected and len(actual)==len(before)+3
ledger=store/('ledger-'+claim['scopeSha256']+'.jsonl');entries=[json.loads(s) for s in ledger.read_text().splitlines() if s];assert entries==load(native/'minimax-ledger.json')['receipts'] and len(entries)==2 and [r['kind'] for r in entries]==['reserve','attempt']
assert load(B/'new-native-files-before-consumption.json')=={str(p.relative_to(native)):sha(p) for p in native.rglob('*') if p.is_file()}
assert sha(R/'ai/catalog_kimi.json')==pre['catalogSHA256'];q=load(B/'inspection-v1/stdout.jsonl');assert len(q['coverage'])==14 and q['qualification']=='FORMAL_APPLICABLE'
assert report['budget']=={'calls':1,'reservedAttempts':1,'attempts':1,'requestedOutput':2048,'reservedInput':2048}
assert not load(native/'audit.json')['findings'] and all('cache_creation_input_tokens' not in u for u in report['calls'][0]['usagePresence'])
cap=load(native/'MM07-thinking-on/result.json')['result']['metadata']['attempts'][0];assert cap['usageReporting']=='partial';budget=load(native/'MM07-thinking-on/thinking-budget.json');assert budget['maxUSD']==0.01 and budget['reservedUSD']==0.006144
save(B/'retention-v1.json',{'pass':True,'historicalFilesChecked':len(protected),'historicalFilesUnchanged':True,'missing':missing,'changed':bad,'oldPublicAuthorityFiles':len(before),'allOldAuthoritySetMembersAndBytesUnchanged':True,'newPublicAuthorityFiles':{str(Path(n).relative_to(R)):actual[n] for n in sorted(expected)},'onlyOneNewClaimAndLedger':True,'newNativeUnchangedByConsumption':True,'oldBatchAndRetrySTOPUnchanged':True,'catalogInactiveAndUnchanged':True,'sourceUnchanged':True,'otherProviderPaidExecutions':0,'elapsedSeconds':time.monotonic()-tic})
# New matrix version only: preserve the old matrix and inherit unaffected rows.
prior=R/'.scratch/kimi-composition-promotion/delivery-matrix-v4.md';lines=prior.read_text().splitlines();rows=[l for l in lines if l.startswith('|')];rows=[l.replace('| minimax | minimax-messages | 13/14 | thinking-on 1项／USD0.01，待独立批准；13＋1已实现 |','| minimax | minimax-messages | MiniMax-M3 14/14（原13＋新1） | FORMAL_APPLICABLE；本次1调用/1HTTP，新增补采0 |').replace('1项／USD0.10，待账户/费率确认及独立批准；目录inactive','1项／USD0.10；账户/费率已确认，当前未获新执行权限；目录inactive') for l in rows]
assert len([l for l in rows if l.startswith('| minimax | minimax-messages')])==1
(S/'delivery-matrix-v5.md').write_text('# Provider 交付矩阵 v5\n\n12个Provider身份、29条路径，沿用旧22条适用性，本次MiniMax国际Messages新增1条，现为23/29。其余路线未重采或更新资格。Kimi六路仍各9/10，目录inactive。\n\n'+'\n'.join(rows)+'\n\n[本次原件](../../.evidence/minimax-thinking-resample/composition-result-v1.json)；旧v4矩阵及全部测量保留。本次只新增MiniMax thinking-on一次，独立grant上限USD0.01，不借用旧预算。MiniMax三个历史单项的累计真实调用为3次，但本次13＋1仅采用本次成功包。Kimi六个xai-origin-history缺口仍需新的独立批准，拟议各1调用/1HTTP/USD0.10、合计USD0.60，当前全部未执行；不恢复旧七路批次。账户/地区/费率沿用既有人类事实确认。全部Provider整体发布仍未完成，目录激活及原人工阶段保持独立；不因本次新grant或消费者记录变化牵动其他Provider付费复验。\n')
(B/'status-before-delivery-v1.md').write_bytes((S/'status.md').read_bytes())
(S/'status.md').write_text(f'''# MiniMax thinking-on 单项重采当前状态

2026-10-28：本次唯一付费重采及固定13＋1只读判定完成。新冻结二进制真实producer PASS/exit0，仅1调用/1HTTP；单项原报告live/candidate、complete=false保持。原13项未重采；组合FORMAL_APPLICABLE，14项coverage、两份独立measurement及原失败thinking-on均保留。外层schema2消费者PASS/exit0，传递输入2657项。 [组合原件](../../.evidence/minimax-thinking-resample/composition-result-v1.json)、[本次授权](../../.evidence/minimax-thinking-resample/approval.json)、[原生进程](../../{rp.relative_to(R)})、[保全](../../.evidence/minimax-thinking-resample/retention-v1.json)、[交付矩阵v5](delivery-matrix-v5.md)。

本次grant/claim各1份，grant上限USD0.01，已知最高费率预留USD0.006144；不是实际账单金额，实际账单未知。观测usage输入111、输出946（其中thinking591）、缓存读128；cache_creation计数仍未报告，UsagePartial仍partial，独立写费NOT_APPLICABLE，没有补零。此前两次单项仍FAIL/exit1及原seal/report缺口不改，也未被用于本次13＋1；仅保留作历史和回归。三次历史单项累计3调用/3HTTP、独立grant上限合计USD0.03，旧预算不退款、不恢复、不合并。

源码/SDK/目录/工具/测试与冻结完全相同，HEAD {source['commit']}，保留既有工作区修改。2051来源、187普通输入和既有八包普通证据完全适用，主包15357 PASS/214 NOT_RUN；本轮没有修改受测代码或重跑普通。原13项另由当前原生公共校验器和原producer/原授权/账本离线核验通过，新单项使用精确入口/count1/180s。SDK/环境/请求/工具自动重试及重跑均0，无探活/预热或其他厂商调用。

本次逐字节核验保全旧195287个历史文件，原8162个公共授权/账本成员及字节不变，仅追加本次真实ID/scope/ledger三个文件；新单项完整封存、audit及只读消费前后字节相同。旧批次STOP、旧重试STOP、七份v5事实确认修订1、所有旧授权及当前暂停状态保留。Kimi目录字节不变，active=false，Kimi本轮授权/调用均0。race/压力未执行；本次只执行既有冻结固定单项和只读消费，无代码、共享状态或设计负载变化，本功能race/压力有依据地NOT_APPLICABLE，不替代其他旧功能阶段。

本次路线已达到14/14；沿用原22条再加本次1条，29路径现23条完整适用，Kimi六路仍各9/10。该单路线完成不表示全部Provider整体发布完成，不激活Kimi、不恢复暂停轮次，不继续任何付费。只读release记录仅覆盖本次MiniMax范围及沿用的准确普通事实，整体Provider发布仍未完成。后续任何付费执行必须有新的明确批准。本次单项、组合及授权记录位于新目录，不覆盖旧版本；撤销当前资格引用不会删除任何证据、授权或账本。
''')
issue=S/'issues/01-single-resample.md';issue.write_text(issue.read_text().replace('Status: claimed','Status: resolved')+'\n## Answer\n\n本次1调用/1HTTP成功，固定13＋1正式适用通过；原件与剩余六路边界见../status.md。\n')
oldVerification=load(R/'.evidence/minimax-cache-billing-finish/verification-v1.json');ordinary=load(R/pre['ordinary']['path']);assert all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
stage={'status':'PASS','receipt':info['nativeReceipt'],'selection':{'by':'human','argv':receipt['argv']},'requiredCases':[q['coverage'][n]['caseId'] for n in q['expected']],'consumers':[ref(B/'consumer-v1/live-consumer.json')]}
record={'schema':2,'feature':'minimax-thinking-resample','currentState':ref(S/'status.md'),'source':source,'selectedInputs':ordinary['selectedInputs'],'impact':{'liveRoutes':['minimax-messages'],'raceRequired':False,'pressureRequired':False,'reason':'Execute only existing frozen MiniMax thinking-on once and consume fixed thirteen plus one. Product, SDK, tested code, authority algorithms and resource/load behavior unchanged. Reuse exact-source ordinary receipt; Kimi activation and other Provider work excluded.'},'e2e':{'status':'PASS','receipt':pre['ordinary'],'requiredCases':oldVerification['requiredCases']},'live':{'status':'PASS','routes':{'minimax-messages':stage}},'race':{'status':'NOT_APPLICABLE','reason':'No implementation or concurrent state change; existing one-call collector and bounded read-only composition only, no race execution authorized.'},'pressure':{'status':'NOT_APPLICABLE','reason':'No new runtime load or resource behavior; one fixed bounded call and read-only evidence, no pressure execution authorized.'}}
save(B/'record-v1.json',record)
save(B/'facts-v1.json',{'qualification':'FORMAL_APPLICABLE','actualCalls':1,'actualHTTPAttempts':1,'paidGrantsIssued':1,'authoritativeClaims':1,'grantMaxUSD':'0.01','knownRateReservedUSD':'0.006144','estimatedKnownUsageUSD':cap['usage']['cost']['total'],'actualInvoiceUSD':None,'usageReporting':'partial','independentCacheWriteFee':'NOT_APPLICABLE','cacheCreationCounter':'UNREPORTED','currentQualifiedPaths':23,'totalPaths':29,'remainingKimiMissingScenes':6,'remainingKimiProposedUSD':'0.60','KimiPaidAuthorityIssued':0,'otherProviderCalls':0,'raceExecuted':False,'pressureExecuted':False,'catalogActive':False,'pausedResumed':False,'allProviderReleasePass':False,'ordinaryReused':pre['ordinary'],'paidReceipt':info['nativeReceipt'],'composition':ref(B/'composition-result-v1.json'),'retention':ref(B/'retention-v1.json'),'matrix':ref(S/'delivery-matrix-v5.md')})
print(json.dumps({'retention':'PASS','historicalFilesUnchanged':len(protected),'oldAuthorityFilesUnchanged':len(before),'newAuthorityFiles':3,'currentQualifiedPaths':23,'remainingKimiScenes':6}))
