"""Retain the stopped paid attempt and independently successful offline review."""
from common import *
O=B/'offline-review-v1'
p=load(O/'preflight.json');protected=load(O/'protected-before.json')
missing=[];changed=[]
for n,h in protected.items():
 path=R/n
 if not path.is_file():missing.append(n)
 elif sha(path)!=h:changed.append(n)
a=load(O/'authority-before.json');now=stores()
aMissing=[n for n in a if n not in now];aChanged=[n for n in a.keys()&now.keys() if a[n]!=now[n]];aAdded=[n for n in now if n not in a]
f=frozen()
scopes=[]
for vr in p['machineScopes']:
 sp=check_ref(vr);d=load(sp);g=d['machineTemplate']
 scopes.append(dict(route=d['route'],scope=vr,grantIssued=False,grantExecutable=False,expiresAt=g['expiresAt'],baselineManifestSHA256=g['baselineManifestSha256'],sourceSHA256=g['sourceSha256'],binarySHA256=g['binarySha256'],maxUSD=g['maxUSD'],reservationAt4096InputOutputUSD=d['reservationAt4096InputOutputUSD']))
retention=dict(kind='read-only-retention-after-stopped-kimi-preflight',protectedFiles=len(protected),missing=missing,changed=changed,authorityFilesBefore=len(a),authorityFilesNow=len(now),authorityMissing=aMissing,authorityChanged=aChanged,authorityAdded=aAdded,pass_=not (missing or changed or aMissing or aChanged),newGrantsIssued=0,newClaims=0,newHTTPAttempts=0,oldSTOPPreserved=ref(R/'.evidence/kimi-history-paid-20261028/STOP.json'),restartSTOPPreserved=ref(B/'STOP.json'))
save(B/'retention-v1.json',retention)
summary=dict(kind='STOPPED_PAID_SEQUENCE_OFFLINE_REVIEW_COMPLETE',recordedAt=datetime.now(timezone.utc).isoformat(),humanRestartApproval=ref(B/'approval.json'),paidStop=ref(B/'STOP.json'),failure=ref(B/'preflight-failure.json'),failedPreflightOriginal=ref(B/'preflight-failed-original.py.txt'),failedPreflightActualExitCode=1,offlineReview=ref(O/'preflight.json'),offlineReviewArgv=['python3','.scratch/kimi-history-paid-restart-20261028/offline_review.py'],offlineReviewActualExitCode=0,offlineReviewExitSource='observed Codex exec_command session 36375 completion',source=f['source'],ordinary=p['ordinary'],ordinarySelectedInputs=p['ordinarySelectedInputs'],original54OfflineReread='PASS',baselineCases=54,scopes=scopes,retention=ref(B/'retention-v1.json'),calls=0,httpAttempts=0,grantsIssued=0,realClaims=0,credentialsRead=0,paidProducerExecutions=0,newAPICostUSD='0.00',candidateCompositionsProduced=0,candidateCoveragePerRoute='original9/10; supplement NOT_RUN',catalogActive=load(R/'ai/catalog_kimi.json')['active'],formalQualificationConsumption=False,catalogPromotion=False,race=False,pressure=False,pausedResumed=False,remainingPaidWorkStopped=True,newHumanApprovalRequired=True,reason='Original detailed instruction requires stopping all remaining paid work after any preflight failure; successful offline correction does not erase the TypeError/STOP.')
save(B/'completion-v1.json',summary)
state('2026-10-28：按用户“使用Kimi前置文书，重新开始”建立独立本次记录，精确选择六份Kimi文书。首次前置读取账本留存副本时发生TypeError（国际Responses/Messages的preservedCopy为文件引用对象，其余为字符串路径），实际退出1；本次付费链已按“任何前置核验失败立即停止全部剩余付费”停止，STOP和失败脚本原件保留。\n\n随后仅做离线修正审阅，退出0：2051来源及187普通输入与指定冻结一致；原54项完整原件/来源/真实producer/原授权和权威账本重新核对，534份封存文件有效。六份新冻结机器范围已生成，均无ID和有效期、不可执行，不复用旧source/binary绑定。沿用已有原生语义/54项公共回放证明及其完整原输入，未伪造新测量，也未重采或重新封存原九项。\n\n本次grant/claim/凭据读取/厂商调用/HTTP均0，新增API消费USD0.00；六路仍各9/10，补项NOT_RUN。目录active=false，正式消费、晋升、race、压力、其他Provider请求与暂停恢复均未执行。成功离线修正不自动恢复已停止付费链；后续付费须新明确批准。\n\n完整索引：[completion-v1.json](../../.evidence/kimi-history-paid-restart-20261028/completion-v1.json)；离线原件：[preflight.json](../../.evidence/kimi-history-paid-restart-20261028/offline-review-v1/preflight.json)；保全：[retention-v1.json](../../.evidence/kimi-history-paid-restart-20261028/retention-v1.json)。只新增本次记录和脚本，生产源码/冻结工具不改。')
with (S/'issues/01-six-route-supplement.md').open('a') as out:out.write('\n## Comments\n\n2026-10-28：六份Kimi文书选择正确；前置脚本类型错误触发本次STOP。仅离线修正审阅通过，原54项/新六份不可执行scope就绪，grant与厂商请求0。未恢复付费，详见../status.md。\n')
print(json.dumps(dict(status='STOPPED_PAID_OFFLINE_READY',baselineCases=54,machineScopes=len(scopes),protectedFiles=len(protected),missing=len(missing),changed=len(changed),authorityMissing=len(aMissing),authorityChanged=len(aChanged),calls=0,grants=0,completion=str((B/'completion-v1.json').relative_to(R)))))
