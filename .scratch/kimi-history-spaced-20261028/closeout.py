"""Read-only closeout after all six routes pass or the first route fails."""
from common import *
from decimal import Decimal

pre=load(B/'preflight.json')
results=[]
for route in ORDER:
 p=B/'results'/f'{route}.json'
 if p.exists():
  result=load(p)
  for field in ['grant','producer','composition','candidateReceipt','candidateResult','transitiveInputs']:check_ref(result[field])
  for n,h in load(R/result['transitiveInputs']['path']).items():assert sha(R/n)==h,n
  results.append(result)
grants=[]
if (B/'grants').exists():
 for p in sorted((B/'grants').glob('*.json')):
  g=load(p);assert p.stem in ORDER and g['combo']==p.stem
  grants.append(dict(route=p.stem,grant=ref(p),id=g['id'],maxUSD=g['maxUSD'],maxCalls=g['maxCalls'],maxHTTPAttempts=g['maxAttempts'],expiresAt=g['expiresAt']))
assert len(grants)<=4 and all(g['maxUSD']==0.10 and g['maxCalls']==g['maxHTTPAttempts']==1 for g in grants)
complete=len(results)==6 and not (B/'STOP.json').exists()
assert [r['route'] for r in results]==ORDER[:len(results)]
if not complete:assert (B/'STOP.json').exists(),'incomplete sequence without a stop is not a closeout'
# Preserve the real completed or failed native measurements independently.
attempts=[]
for route in ORDER:
 p=B/'attempt-indices'/f'{route}.json'
 if not p.exists():continue
 idx=load(p);rp=check_ref(idx['nativeReceipt']);producer=load(rp)
 native=list((rp.parent/'bundles').glob('*/regional-report.json'))
 measurements=[]
 for reportPath in native:
  report=load(reportPath);measurements.append(dict(report=ref(reportPath),budget=report.get('budget'),regional=report.get('regional'),scenarios=report.get('scenarios')))
 attempts.append(dict(route=route,index=ref(p),nativeReceipt=ref(rp),nativeExitCode=producer['exitCode'],captureStatus=producer['status'],measurements=measurements))
# Historical snapshots are immutable; this phase never writes to their roots.
protected=load(B/'protected-before.json');missing=[];changed=[]
for n,h in protected.items():
 p=R/n
 if not p.is_file():missing.append(n)
 elif sha(p)!=h:changed.append(n)
assert not missing and not changed,'historical protected bytes changed'
for vr in load(B/'previous-records.json')['preserved']:check_ref(vr)
oldAuthority=load(B/'authority-before.json');authority=stores()
assert all(authority.get(n)==h for n,h in oldAuthority.items()),'old authority changed'
added={n:h for n,h in authority.items() if n not in oldAuthority}
expected=set();reservationRecords=[]
for row in grants:
 idname='id-'+hashlib.sha256(row['id'].encode()).hexdigest()+'.json'
 path=R/'.git/barness-kimi-grants'/idname
 if path.exists():
  ar=load(path);names=[idname,'scope-'+ar['scopeSha256']+'.json']
  for name in names:
   p=R/'.git/barness-kimi-grants'/name
   if p.exists():
    expected.add(str(p.relative_to(R)));reservationRecords.append(ref(p))
assert set(added)==expected,'unexpected authority additions'
if complete:assert len(grants)==4 and len(added)==8
f=frozen();assert f['source']==pre['source']
assert load(R/'ai/catalog_kimi.json')['active'] is False
previousPaid=load(B/'previous-paid-history-before.json')
for n,h in previousPaid.items():
 path=R/n
 assert path.is_file() and not path.is_symlink(),n
 with path.open('rb') as stream:assert hashlib.file_digest(stream,'sha256').hexdigest()==h,n
cooldowns=[]
for g in grants:
 route=g['route'];cp=B/'cooldowns'/f'{route}.json';gate=load(cp);issuance=load(B/'issuance'/f'{route}.json')
 assert datetime.fromisoformat(issuance['issuedAt'])>=datetime.fromisoformat(gate['earliestGrantAt'])
 assert gate['requiredSeconds']==300 and gate['elapsedSeconds']>=300
 cooldowns.append(ref(cp))
save(B/'retention.json',dict(pass_=True,protectedFiles=len(protected),missing=missing,changed=changed,oldAuthorityFiles=len(oldAuthority),oldAuthorityExactBytesUnchanged=True,newAuthorityFiles=added,permanentReservationRecords=reservationRecords,previousSTOPsPreserved=True))
# Each route contributes two actual measurements; never synthesize one source,
# commit, budget, authorization or timestamp for old-nine plus new-one.
routeRows=[]
for r in results:
 q=load(check_ref(r['candidateResult']))
 assert len(q['measurements'])==2 and len(q['coverage'])==10
 routeRows.append(dict(route=r['route'],newMeasurementThisApproval=r['route'] in REMAINING,reusedChatCandidate=r['route'] in ORDER[:2],reusedFrom=r.get('reusedFrom'),qualification=q['qualification'],candidateResult=r['candidateResult'],candidateReceipt=r['candidateReceipt'],composition=r['composition'],transitiveInputs=r['transitiveInputs'],coverage=q['coverage'],measurements=q['measurements']))
reserved=sum((Decimal(str(r['measurements'][1]['original']['regional']['reservedUSD'])) for r in results if r['route'] in REMAINING),Decimal(0))
knownCalls=[];knownHTTP=[];unknown=[]
for a in attempts:
 if len(a['measurements'])==1 and isinstance(a['measurements'][0].get('budget'),dict):
  budget=a['measurements'][0]['budget'];knownCalls.append(budget.get('callsUsed'));knownHTTP.append(budget.get('httpAttempts'))
 else:unknown.append(a['route'])
if complete:
 assert knownCalls==[1]*4 and knownHTTP==[1]*4 and not unknown and reserved<=Decimal('0.40')
status='SIX_CANDIDATE_COMPOSITIONS_COMPLETE' if complete else 'STOPPED_AFTER_FAILURE'
textState='六路全部完成' if complete else '本轮已按失败即停停止'
if (B/'failure-diagnosis.json').exists():
 textState+='：'+load(B/'failure-diagnosis.json')['route']+'原生采集失败，原件与计费状态见failure-diagnosis.json，缺失用量不作为明确零计费'
(B/'status-before-closeout.md').write_bytes((S/'status.md').read_bytes())
state('2026-10-28：'+textState+'。本次按加大间隔的明确继续指令，仅新增Kimi/kimi-k3/live/candidate的国内Responses、国际Responses、国内Messages、国际Messages；两路Chat只读沿用。每路签grant前额外冷却300秒，进程内首等60秒保持。逐路新schema2授权，30分钟有效期，1调用/1HTTP/USD0.10独立预算；前一路采集和候选9＋1通过后才签下一路。\n\n成功候选组合'+str(len(results))+'/6；新grant '+str(len(grants))+'，已留存原生进程'+str(len(attempts))+'。成功的新单项均complete=false；每组保留旧九项＋新一项、十项覆盖和两份实际测量，不伪造统一source/build/commit/时间/预算。完整命令、真实数字退出、授权/账本、输入和捕获见[completion.json](../../.evidence/kimi-history-spaced-20261028/completion.json)。\n\n前置复核原54项/534封存文件、2051当前来源与187普通输入通过。保全'+str(len(protected))+'个受保护历史文件无缺失或变化，原授权账本'+str(len(oldAuthority))+'文件字节不变，本次仅新增'+str(len(added))+'份永久reservation。全部旧失败和STOP保留。\n\n本次已停止后续付费工作，不自动重试/续期/换ID；'+('六路候选结果完整；本次新增四路的发送前保守预留合计USD '+str(reserved)+'（不是厂商账单硬封顶）。' if complete else '失败/未知与实际消费保留，未开始路线不再签grant，继续须新明确批准。')+'目录active=false；正式资格消费、晋升激活、race、压力、MiniMax或其他Provider请求、xAI请求及暂停恢复均未执行。不宣称正式目录交付或完整发布通过。')
summary=dict(kind=status,recordedAt=datetime.now(timezone.utc).isoformat(),approval=ref(B/'approval.json'),scope=ref(S/'spec.md'),currentState=ref(S/'status.md'),preflight=ref(B/'preflight.json'),freeze=ref(FREEZE),order=ORDER,routes=routeRows,successfulCandidateCompositions=len(results),independentMeasurements=len(routeRows)*2,grants=grants,attempts=attempts,totalKnownCalls=sum(x for x in knownCalls if isinstance(x,int)),totalKnownHTTPAttempts=sum(x for x in knownHTTP if isinstance(x,int)),unknownAttemptCounts=unknown,perRouteApprovedMaxUSD='0.10',totalApprovedMaxUSD='0.40',successfulNewMeasurementsReservedUSD=str(reserved),reservationIsNotVendorInvoiceCap=True,retention=ref(B/'retention.json'),catalogActive=False,formalQualificationConsumption=False,catalogPromotion=False,raceExecuted=False,pressureExecuted=False,otherProviderRequests=False,xaiRequests=False,pausedResumed=False,paidWorkStopped=True,releasePass=False)
if (B/'STOP.json').exists():summary['stop']=ref(B/'STOP.json')
if (B/'failure-diagnosis.json').exists():summary['failureDiagnosis']=ref(B/'failure-diagnosis.json')
if (B/'usage-and-reservations.json').exists():summary['usageAndReservations']=ref(B/'usage-and-reservations.json')
summary['remainingUnsignedRoutes']=[r for r in REMAINING if not (B/'grants'/f'{r}.json').exists()]
summary['routeOutcomes']={route:('CANDIDATE_COMPLETE' if any(r['route']==route for r in results) else 'FAILED' if any(a['route']==route for a in attempts) else 'NOT_RUN') for route in ORDER}
summary['externalCooldowns']=cooldowns
summary['previousPaidFilesUnchanged']=len(previousPaid)
summary['reusedChatRoutes']=ORDER[:2]
summary['newOrder']=REMAINING
summary['previousFailedMeasurementPreserved']=ref(R/'.evidence/kimi-history-live-20261028/failure-diagnosis.json')
summary['previousSequenceOutcome']=ref(R/'.evidence/kimi-history-live-20261028/completion.json')
summary['allSentReservationsUSD']=str(sum((Decimal(str(m['regional']['reservedUSD'])) for a in attempts for m in a['measurements'] if m.get('regional')),Decimal(0)))
save(B/'completion.json',summary)
issue=S/'issues/01-remaining-four.md';body=issue.read_text()
if complete:body=body.replace('Status: claimed','Status: resolved',1)
issue.write_text(body+'\n## Comments\n\n本次结果：'+status+'；候选组合'+str(len(results))+'/6。详见../status.md及completion.json，所有历史事实保留。\n')
print(json.dumps(dict(status=status,candidateCompositions=len(results),independentMeasurements=len(routeRows)*2,grants=len(grants),calls=summary['totalKnownCalls'],httpAttempts=summary['totalKnownHTTPAttempts'],catalogActive=False,completion=str((B/'completion.json').relative_to(R)))))
