"""Record the local script fix without running a preflight or paid producer."""
from common import *
import difflib

out=B/'type-error-fix-v1'
red=load(out/'red/report.json');green=load(out/'green/report.json')
assert not red['pass_'] and green['pass_']
assert len(green['results'])==34 and all(row['pass'] for row in green['results'])
source=source_snapshot(R);f=frozen()
assert source==load(out/'source-before.json')==f['source']
ordinary=load(R/'.evidence/minimax-cache-billing-finish/e2e/20261028T030752.619882Z-7o4zhdo3/receipt.json')
assert ordinary['source']==source and all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
history=load(out/'history-before.json')
def history_sha(p):
 with p.open('rb') as stream:return hashlib.file_digest(stream,'sha256').hexdigest()
missing=[n for n in history if not (R/n).is_file()]
changed=[n for n,h in history.items() if (R/n).is_file() and history_sha(R/n)!=h]
authority=stores();before=load(out/'authority-before.json')
assert not missing and not changed and before==authority
changes=[]
for name in ['common.py','preflight.py','offline_review.py']:
 old=out/('before-'+name+'.txt');current=S/name
 patch=''.join(difflib.unified_diff(old.read_text().splitlines(True),current.read_text().splitlines(True),fromfile='before/'+name,tofile='after/'+name))
 (out/(name+'.patch')).write_text(patch)
 (out/('after-'+name+'.txt')).write_bytes(current.read_bytes())
 changes.append(dict(path=str(current.relative_to(R)),before=ref(old),after=ref(current),retainedAfter=ref(out/('after-'+name+'.txt')),patch=ref(out/(name+'.patch'))))
previous=load(B/'completion-v1.json')
now=datetime.now(timezone.utc).isoformat()
state('2026-10-28：前置脚本账本副本类型错误已修复。本次仅离线修复，不恢复付费链；此前两次失败、STOP、已完成的离线54项审阅及六份不可执行机器范围均保留。\n\n原因：preflight.py始终把preservedCopy当路径字符串；国际Responses/Messages的旧记录实际为{path,sha256}引用对象。上次仅offline_review.py处理了该格式，实际入口没有接入。本次在common.py增加一处明确的账本记录读取校验，preflight.py与offline_review.py均调用；字符串格式从外层取哈希，对象格式从引用内取路径和哈希，随后独立核验副本和权威账本。非法格式/缺字段明确报错，哈希不符仍拒绝，不改写历史格式。\n\n先保留实际入口读取段的失败进程回归（退出1），再修改代码；34项离线子进程检查全部通过（退出0），包括两个入口分别读取六路线共12条真实账本记录、两种格式，以及错误类型/缺字段/副本或权威账本篡改拒绝。仅执行实际入口的账本读取段，未重跑会写STOP/scope的完整前置流程，也未重跑厂商采集。\n\n2051来源与冻结仍一致，187普通输入保持；46个既有执行证据文件及8165个授权/账本文件集合和字节不变。付费live、race、压力、正式消费、目录激活和暂停恢复均未执行，新grant/claim/凭据读取/厂商请求均0。并发/生命周期/负载未变，race/压力本次不适用。\n\n修复与证据：[verification.json](../../.evidence/kimi-history-paid-restart-20261028/type-error-fix-v1/verification.json)。先前原采集/停止事实见[completion-v1.json](../../.evidence/kimi-history-paid-restart-20261028/completion-v1.json)，本次不改其测量或状态。回滚仅恢复本次三个本地脚本的before快照；历史账本和证据无需回滚。')
issue=S/'issues/02-preserved-copy-type.md'
body=issue.read_text().replace('Status: claimed','Status: resolved',1)
body+='\n## Answer\n\n实际入口与离线审阅入口现共用verify_ledger_record。先失败、后修复；34项实际读取段子进程回归通过，六路线12条原账本均接受，篡改/缺字段仍拒绝。无厂商请求，旧STOP保留。详见../status.md及type-error-fix-v1/verification.json。\n'
issue.write_text(body)
save(out/'verification.json',dict(kind='KIMI_PREFLIGHT_PRESERVED_COPY_TYPE_FIX',recordedAt=now,status='FIXED_OFFLINE',cause='Actual preflight treated preservedCopy as a path string; two routes store a path/SHA256 object. Only the offline copy had previously been corrected.',changes=changes,red=dict(report=ref(out/'red/report.json'),argv=['python3','.scratch/kimi-history-paid-restart-20261028/verify_ledger_regression.py','--out','.evidence/kimi-history-paid-restart-20261028/type-error-fix-v1/red'],actualExitCode=1,sourceRetainedBefore=True),green=dict(report=ref(out/'green/report.json'),argv=['python3','.scratch/kimi-history-paid-restart-20261028/verify_ledger_regression.py','--out','.evidence/kimi-history-paid-restart-20261028/type-error-fix-v1/green'],actualExitCode=0,checks=34,actualEntryLoops=True,wholePreflightExecuted=False,retainedRecords=12),regressionScript=ref(S/'verify_ledger_regression.py'),recorder=ref(Path(__file__).resolve()),currentState=ref(S/'status.md'),frozenSourceUnchanged=True,frozenSourceFiles=len(source['files']),ordinarySelectedInputs=len(ordinary['selectedInputs']),oldEvidenceFiles=len(history),oldEvidenceMissing=missing,oldEvidenceChanged=changed,authorityFiles=len(authority),authorityExactSetAndBytesUnchanged=True,previousPaidOutcome=ref(B/'completion-v1.json'),stopPreserved=ref(B/'STOP.json'),oldStopPreserved=ref(R/'.evidence/kimi-history-paid-20261028/STOP.json'),grantsIssued=0,claims=0,credentialsRead=0,vendorCalls=0,httpAttempts=0,live='NOT_RUN',race='NOT_APPLICABLE; no concurrency change',pressure='NOT_APPLICABLE; bounded reference decoding',formalConsumption=False,catalogActivation=False,pausedResumed=False))
print(json.dumps(dict(status='FIXED_OFFLINE',checks=34,oldEvidenceUnchanged=len(history),authorityFilesUnchanged=len(authority),vendorCalls=0,verification=str((out/'verification.json').relative_to(R)))))
