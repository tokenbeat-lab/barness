"""Read-only closeout of the first paid attempt and one human-requested retry."""
from pathlib import Path
from datetime import datetime,timezone
from decimal import Decimal
import json,hashlib,sys
R=Path(__file__).resolve().parents[2];B=R/'.evidence/provider-minimal-paid-supplement';S=R/'.scratch/provider-minimal-paid-supplement'
sys.path.insert(0,str(R/'tools/verification'));from receipt_files import source_snapshot
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();load=lambda p:json.loads(p.read_text())
def ref(p):return {'path':str(p.relative_to(R)),'sha256':sha(p)}
def save(p,d):
 with p.open('x') as f:json.dump(d,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
pre=load(B/'preflight.json');assert source_snapshot(R)==pre['source']
originalPublic=load(B/'historical-authority-before.json');roots=load(R/'.evidence/minimax-thinking-supplement-plan/store-roots.json');actual={str(p):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}
assert all(actual.get(n)==h for n,h in originalPublic.items())
assert sha(R/'ai/catalog_kimi.json')=='c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109'
idx=R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json';assert sha(idx)=='818152762619e44b4a8d60946e944ea4201ef9fd19f9725ce288083e50d99c38'
for f in load(idx)['scopeDocuments']:assert sha(R/f['path'])==f['sha256']
rows=[];expectedNew=set();reserve=Decimal(0);sealed=0;gaps=[];recoveredRefs=[]
for outer in [B/'minimax-attempt-index.json',B/'retry-minimax-v1/attempt-index.json']:
 info=load(outer);receiptPath=R/info['nativeReceipt']['path'];assert sha(receiptPath)==info['nativeReceipt']['sha256'];receipt=load(receiptPath)
 assert receipt['source']==pre['source'] and receipt['status']=='FAIL' and receipt['exitCode']==1
 assert all(sha(R/n)==h for n,h in receipt['selectedInputs'].items())
 manifests=list(receiptPath.parent.glob('bundles/*/manifest.json'));assert len(manifests)==1;native=manifests[0].parent
 report=load(native/'minimax-report.json');audit=load(native/'audit.json');assert audit['findings']==[]
 assert report['mode']=='live' and report['stage']=='candidate' and report['outcome']=='FAIL' and report['complete']==False
 assert report['failure']=='MiniMax single billable counter unreported: cache_creation_input_tokens'
 assert report['budget']['calls']==report['budget']['attempts']==report['budget']['reservedAttempts']==1
 assert len(report['calls'])==1 and report['calls'][0]['id']=='thinking-on' and report['calls'][0]['outcome']=='PASS'
 usage=report['calls'][0]['usagePresence'];assert all('cache_creation_input_tokens' not in u for u in usage)
 wireBudget=load(native/'MM07-thinking-on/thinking-budget.json');assert wireBudget['maxUSD']==0.01 and wireBudget['reservedUSD']==0.006144
 reserve+=Decimal(str(wireBudget['reservedUSD']))
 seal=load(native/'minimax-integrity.json')
 bad=[]
 for n,h in seal['files'].items():
  actualHash=sha(native/n);sealed+=1
  if actualHash!=h:bad.append({'file':n,'sealSHA256':h,'finalSHA256':actualHash})
 assert len(bad)==1 and bad[0]['file']=='minimax-report.json'
 # Restore only into a separate diagnostic path and verify its exact hash
 # against the retained old seal. It is a provisional pre-validation report,
 # never a successful qualification and never substituted into the native pack.
 raw=(native/'minimax-report.json').read_bytes()
 provisional=raw.replace(b'  "outcome": "FAIL",\n',b'  "outcome": "PASS",\n',1).replace(b'  "failure": "MiniMax single billable counter unreported: cache_creation_input_tokens",\n',b'',1)
 assert hashlib.sha256(provisional).hexdigest()==bad[0]['sealSHA256']
 recovered=B/'diagnostics'/('pre-verification-report-'+native.name+'.json')
 with recovered.open('xb') as stream:stream.write(provisional)
 recoveredRefs.append({'recoveredBytes':ref(recovered),'role':'diagnostic pre-validation provisional report; not a qualification or original packet replacement','hashMatchesOriginalSeal':True,'finalOutcomeStill':'FAIL'})
 gaps.append({'nativeReport':ref(native/'minimax-report.json'),'nativeSeal':ref(native/'minimax-integrity.json'),'mismatches':bad,'cause':'finishMiniMax seals before VerifyThinkingSupplement and rewrites FAIL report on validation error without resealing','provisionalReportDiagnostic':ref(recovered)})
 grant=load(native/'process-grant.json');claim=load(native/'process-authorization-receipt.json');assert sha(native/'process-grant.json')==claim['grantSha256']
 assert sha(R/info['grant']['path'])==info['grant']['sha256']==claim['grantSha256']
 store=R/'.git/barness-live-process-grants';idKey='id-'+hashlib.sha256(grant['id'].encode()).hexdigest()+'.json';scopeKey='scope-'+claim['scopeSha256']+'.json';ledgerKey='ledger-'+claim['scopeSha256']+'.jsonl'
 for name in [idKey,scopeKey]:
  p=store/name;assert load(p)==claim;expectedNew.add(str(p))
 ledger=store/ledgerKey;nativeLedger=load(native/'minimax-ledger.json')['receipts'];assert [json.loads(x) for x in ledger.read_text().splitlines() if x]==nativeLedger;expectedNew.add(str(ledger));assert len(nativeLedger)==2
 rows.append({'producer':ref(receiptPath),'grant':info['grant'],'nativeReport':ref(native/'minimax-report.json'),'manifest':ref(native/'manifest.json'),'audit':ref(native/'audit.json'),'integrity':ref(native/'minimax-integrity.json'),'thinkingBudget':ref(native/'MM07-thinking-on/thinking-budget.json'),'claim':ref(native/'process-authorization-receipt.json'),'authoritativeLedger':ref(ledger),'actualArgv':receipt['argv'],'startedAt':receipt['startedAt'],'finishedAt':receipt['finishedAt'],'elapsedSeconds':receipt['elapsedSeconds'],'exitCode':1,'captureStatus':'FAIL','nativeOutcome':'FAIL','sceneFlowOutcome':'PASS','missingBillableCounter':'cache_creation_input_tokens','calls':1,'httpAttempts':1,'grantMaxUSD':'0.01','conservativeReservedUSD':'0.006144','actualInvoiceUSD':None,'qualificationPassed':False,'combinationAttempted':False,'nativeSealMatchesFinalReport':False,'integrityGap':bad,'diagnosticProvisionalReport':ref(recovered)})
assert set(actual)-set(originalPublic)==expectedNew
assert len(expectedNew)==6 and len(actual)==len(originalPublic)+6
assert not list((B/'grants').glob('kimi*'))
(S/'status.md').write_text('''# 七个最小付费补采当前状态

2026-10-28：STOPPED。本轮首条MiniMax thinking-on及用户后续明确要求的一次同范围重试，均实际FAIL/exit1，故整个剩余付费序列停止。Kimi六路没有签发grant、没有调用，不恢复暂停轮次。失败原因均为原始SSE用量未报告cache_creation_input_tokens；场景正文/流完成，但固定单项计费用量合同未满足，不能取得补项资格，不能与原13项组合通过。

[交付与两次原件](../../.evidence/provider-minimal-paid-supplement/delivery-v1.json)、[原七路批准](../../.evidence/provider-minimal-paid-supplement/approval.json)、[原失败STOP](../../.evidence/provider-minimal-paid-supplement/STOP.json)、[仅一次MiniMax重试批准](../../.evidence/provider-minimal-paid-supplement/retry-minimax-v1/approval.json)。原STOP未清除，重试未恢复Kimi权限。

累计真实2调用/2HTTP、2份独立新grant，授权上限合计USD0.02、保守已知费率预留合计USD0.012288；这些不是已确认账单金额，实际账单未知。两次均零SDK/环境/请求/工具自动重试；第二次来自人类新指令。原全局USD0.61没有提高，剩余未签授权额度按授权上限计为USD0.59，不能擅自执行六份各USD0.10的Kimi预算或继续第三次MiniMax。

旧公共授权/账本8156文件的路径与字节全部保留；仅在真实Git授权域追加本次两份新claim对应的6个ID/scope/ledger文件。两份失败原生包、manifest、audit、封存、日志、数字退出、授权、消费与预算均保留。原v5事实确认修订1七份文件、冻结二进制、源码/SDK/型号/目录/价格保持，源码仍匹配已通过完整普通证据，不重建、不修改消费者，不重跑普通。

真实缺口仍为MiniMax thinking-on与Kimi六项；22/29路径完整适用，MiniMax13/14、Kimi各9/10，原54+13项适用性保持。release.pass=false。race/压力、Kimi目录激活、旧暂停恢复、额外Provider调用均未执行。

另发现两个失败包都有原生收尾完整性缺口：finishMiniMax先封存验证前报告，随后用量校验失败改写报告但未更新seal；各包只有minimax-report.json与seal不符，其他封存文件一致。原FAIL报告和seal均保持原样；仅在diagnostics路径复原验证前报告字节，并以原seal SHA核对成功，用于追溯收尾缺陷，不能作为PASS资格或替换原包。

下一步先离线审阅上述收尾缺陷、MiniMax缓存写计数省略语义与当前消费者规则，区分实际计费证据缺口和过宽拦截；现有捕获是否可复用另行判定。不得把缺失计数改写为0、把失败改成PASS或继续盲目付费重试；后续付费需重新明确范围及累计预算。
''')
issue=S/'issues/01-seven-paid-supplements.md';issue.write_text(issue.read_text().replace('Status: claimed','Status: ready-for-human')+'\n## 执行结果\n\n首项及一次人类明确重试均因缺cache_creation_input_tokens失败，已按合同停止；Kimi未签发/未执行。需先离线审阅用量合同，再另行决定是否修消费者或重新批准补采及累计预算。原件见../status.md。\n')
result={'kind':'paid-minimal-supplements-stopped-after-first-route-and-one-human-requested-retry','recordedAt':datetime.now(timezone.utc).isoformat(),'status':'STOPPED','source':pre['source'],'sourceUnchanged':True,'currentState':ref(S/'status.md'),'initialApproval':ref(B/'approval.json'),'retryApproval':ref(B/'retry-minimax-v1/approval.json'),'originalStop':ref(B/'STOP.json'),'retryStop':ref(B/'retry-minimax-v1/STOP.json'),'attempts':rows,'actualCalls':2,'actualHTTPAttempts':2,'paidGrantsIssued':2,'newAuthoritativeClaims':2,'grantCapsUSD':'0.02','knownRateReservedUSD':str(reserve),'actualInvoiceUSD':None,'originalGlobalCapUSD':'0.61','remainingUnissuedGlobalCapUSD':'0.59','kimiGrantsIssued':0,'kimiCalls':0,'remainingSixRoutesStatus':'NOT_STARTED_NOT_REAUTHORIZED_AFTER_STOP','oldPublicAuthorityFilesPreserved':len(originalPublic),'newPublicAuthorityFiles':{str(Path(n).relative_to(R)):actual[n] for n in sorted(expectedNew)},'allOldAuthorityBytesUnchanged':True,'sealedNewNativeFilesChecked':sealed,'nativeSealsMatchFinalReports':False,'integrityGaps':gaps,'diagnosticRecoveredReports':recoveredRefs,'oldV5ConfirmedIndex':ref(idx),'catalogActive':False,'raceExecuted':False,'pressureExecuted':False,'pausedResumed':False,'ordinaryRerun':False,'releasePass':False,'currentQualifiedPaths':22,'totalPaths':29,'newSingleFactQualifications':0,'newCombinationQualifications':0,'automaticPaidRetryOrContinuationForbidden':True}
save(B/'delivery-v1.json',result)
print(json.dumps({k:result[k] for k in ['status','actualCalls','actualHTTPAttempts','paidGrantsIssued','grantCapsUSD','knownRateReservedUSD','kimiGrantsIssued','oldPublicAuthorityFilesPreserved','allOldAuthorityBytesUnchanged','releasePass']}))
