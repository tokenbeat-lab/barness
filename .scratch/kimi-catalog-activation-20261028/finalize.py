"""Bind active catalog, current formal consumers, ordinary evidence and history."""
from pathlib import Path
from datetime import datetime,timezone
import hashlib,json,sys,difflib
R=Path(__file__).resolve().parents[2];S=R/'.scratch/kimi-catalog-activation-20261028';B=R/'.evidence/kimi-catalog-activation-20261028'
sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import sha,save,source_snapshot
load=lambda p:json.loads(p.read_text());ref=lambda p:{'path':str(p.relative_to(R)),'sha256':sha(p)}
ordinaryPath=R/sys.argv[1];ordinary=load(ordinaryPath)
assert ordinary['status']=='PASS' and ordinary['exitCode']==0
source=source_snapshot(R);assert source==ordinary['source']
assert all(sha(R/n)==h for n,h in ordinary['selectedInputs'].items())
formal=load(B/'formal-qualification-v4.json');assert formal['pass_'] and formal['source']==source and len(formal['consumers'])==6
for row in formal['consumers']:
 for field in ['outer','native','capture','originalComposition']:
  vr=row[field];assert sha(R/vr['path'])==vr['sha256']
 outer=load(R/row['outer']['path']);native=load(R/row['native']['path'])
 assert outer['schema']==2 and outer['pass'] and outer['exitCode']==0 and outer['source']==source and native['pass']
 for n,h in outer['files'].items():assert sha((R/row['outer']['path']).parent/n)==h
activation=load(B/'activation.json');proof=load(R/activation['intent']['path']);catalog=load(R/'ai/catalog_kimi.json')
assert activation['pass_'] and catalog['active'] and len(catalog['models'])==6
assert sha(R/'ai/catalog_kimi.json')==activation['catalogAfter']['sha256']==proof['catalogSha256']
assert proof['catalog']==catalog and len(proof['candidates'])==6
assert sum(len(c['qualification']['measurements']) for c in proof['candidates'])==12
assert all(len(c['qualification']['coverage'])==10 for c in proof['candidates'])
expected={'ai/e2e/kimi_responses_collision_test.go','ai/catalog_kimi.json','ai/release/catalog-snapshot.json','ai/e2e/kimi_qualification_promotion_test.go','ai/e2e/kimi_host_rollback_test.go','ai/e2e/kimi_composition_promotion_test.go','ai/e2e/kimi_history_composition_test.go','ai/examples/chineseeval/testdata/jev-1.13.0-v1.json','ai/examples/chineseeval/testdata/jev-route-v1.json','ai/examples/chineseeval/testdata/luna-route-v1.json'}
before=load(B/'source-before.json');changes={n:dict(before=before['files'].get(n),after=source['files'].get(n)) for n in before['files'].keys()|source['files'].keys() if before['files'].get(n)!=source['files'].get(n)}
assert set(changes)==expected,changes
snapshot=load(R/'ai/release/catalog-snapshot.json');oldSnapshot=load(B/'before/ai__release__catalog-snapshot.json')
new=snapshot['catalog'];without={**new,'models':[m for m in new['models'] if m['provider'] not in ['moonshot','moonshot-cn']]}
assert without==oldSnapshot['catalog']
assert [m for m in new['models'] if m['provider'] in ['moonshot','moonshot-cn']]==catalog['models']
inventories=[];required={'catalog-release-snapshot','P07-zh-evaluation-current-jev','P07-zh-evaluation-current-luna','KIMI-K16-promotion-proof-boundary','KIMI-moonshot-K02-K15-host-disable-inflight-rollback','KIMI-moonshot-cn-K02-K15-host-disable-inflight-rollback'}
seen=set()
for row in ordinary['bundles']:
 bundle=ordinaryPath.parent/row['path'];manifest=load(bundle/'manifest.json');audit=load(bundle/'audit.json')
 assert sha(bundle/'manifest.json')==row['manifestSHA256'] and sha(bundle/'audit.json')==row['auditSHA256'] and audit['findings']==[]
 counts={}
 for c in manifest['cases']:
  counts[c['status']]=counts.get(c['status'],0)+1
  if manifest['package']=='./ai/e2e' and c['status']=='PASS':seen.add(c['id'])
 assert not any(n not in ['PASS','NOT_RUN','UNSUPPORTED'] for n in counts)
 inventories.append(dict(package=manifest['package'],counts=counts,manifest=ref(bundle/'manifest.json'),audit=ref(bundle/'audit.json')))
assert required<=seen,required-seen
history=load(B/'protected-before.json');missing=[];changed=[]
for n,h in history.items():
 p=R/n
 if not p.is_file():missing.append(n)
 else:
  assert not p.is_symlink()
  with p.open('rb') as f:actual=hashlib.file_digest(f,'sha256').hexdigest()
  if actual!=h:changed.append(n)
assert not missing and not changed,(missing,changed)
roots=load(R/'.evidence/kimi-history-supplement-plan/store-roots.json')
currentAuthority={str(p.relative_to(R)):sha(p) for root in roots for p in Path(root).rglob('*') if p.is_file()}
assert currentAuthority==load(B/'authority-before.json'),'public authority changed'
save(B/'retention.json',dict(pass_=True,protectedFiles=len(history),missing=missing,changed=changed,authorityFiles=len(currentAuthority),authorityExactSetAndBytesUnchanged=True,previousFailuresAndSTOPsPreserved=True))
for n in sorted(expected):
 old=B/'before'/n.replace('/','__');after=R/n
 (B/(n.replace('/','__')+'.patch')).write_text(''.join(difflib.unified_diff(old.read_text().splitlines(True),after.read_text().splitlines(True),fromfile='before/'+n,tofile='after/'+n)))
(B/'status-before-complete.md').write_bytes((S/'status.md').read_bytes())
(S/'status.md').write_text('# 当前状态\n\n2026-10-28：Kimi正式目录已激活并验证。ai/catalog_kimi.json的active=true、版本kimi-k3-formal-2026-10-22.1，包含moonshot/moonshot-cn × Chat/Responses/Messages六个kimi-k3身份，字段与已审候选完全一致。通过冻结regionalpromote消费六组真实9＋1、十二份独立测量及完整传递库存后原子写入，晋升命令退出0；独占intent与激活前目录精确备份均保留。\n\n当前内置发布快照仅新增六个Kimi身份，其他型号/图像/分类器/能力/价格/目录版本完全相同。三个当前评测配置仅同步全目录hash和配置revision，未改变任务、型号、指令或预算；历史录制配置及报告保持。五个公共E2E文件去除旧阶段的inactive假设并显式测试旧目录回滚。总计十处既有来源文件变化，无SDK/协议/计费/授权实现变化。\n\n先保留激活前RED、快照缺项与旧当前配置pin失败，再修复对应数据。最终完整普通五包退出0，所需目录/回滚/当前配置case全部PASS；全部原生库存和audit通过。最终来源与普通回执、只读消费者v4完全一致。六路正式schema2外层消费者均PASS/exit0，分别保留旧测量的真实stage/mode/来源/预算/授权和时间，不重采、不重封存。\n\n原'+str(len(history))+'个受保护文件及'+str(len(currentAuthority))+'个公共授权/账本文件集合和字节保持，旧过载失败、未知费用和全部STOP保留。新增厂商请求/真实grant/claim/凭据读取均0。race/压力无新增并发/生命周期/负载影响，本次不适用；不恢复暂停轮次，不宣称所有Provider完整发布。\n\n最终索引：[completion.json](../../.evidence/kimi-catalog-activation-20261028/completion.json)。回滚按保存的before快照恢复本次目录/发布快照、当前配置及对应测试，保留所有晋升、测试、资格回执和历史账本；已编译程序中的目录随重新构建生效。\n')
issue=S/'issues/01-activate-catalog.md';issue.write_text(issue.read_text().replace('Status: claimed','Status: resolved',1)+'\n## Answer\n\n六身份目录已通过原生流程激活；六路正式离线资格和最终普通验证通过，原件/账本不变。详见../status.md及completion.json。\n')
allAttempts=[]
for attempt in sorted((B/'e2e').iterdir()):
 if (attempt/'receipt.json').is_file():
  r=load(attempt/'receipt.json');allAttempts.append(dict(receipt=ref(attempt/'receipt.json'),status=r['status'],exitCode=r['exitCode']))
scripts={str(p.relative_to(R)):sha(p) for p in S.iterdir() if p.suffix in ['.py','.go']}
save(B/'completion.json',dict(kind='KIMI_CATALOG_ACTIVATED_AND_VERIFIED',recordedAt=datetime.now(timezone.utc).isoformat(),approval=ref(B/'approval.json'),currentState=ref(S/'status.md'),source=source,sourceChanges=changes,catalog=ref(R/'ai/catalog_kimi.json'),catalogActive=True,catalogVersion=catalog['version'],identities=[dict(provider=m['provider'],api=m['api'],model=m['id']) for m in catalog['models']],activation=ref(B/'activation.json'),promotionIntent=activation['intent'],catalogSnapshot=ref(R/'ai/release/catalog-snapshot.json'),snapshotUpdate=ref(B/'catalog-snapshot-update.json'),currentEvaluationPinUpdate=ref(B/'current-evaluation-pin-update.json'),ordinary=ref(ordinaryPath),requiredCases=sorted(required),inventories=inventories,formalQualification=ref(B/'formal-qualification-v4.json'),formalConsumers=formal['consumers'],allOrdinaryAttempts=allAttempts,retention=ref(B/'retention.json'),supportingScripts=scripts,historicalSpecVersions=ref(B/'spec-version-index.json'),newVendorCalls=0,newGrants=0,newClaims=0,credentialsRead=0,race='NOT_APPLICABLE: only catalog/config data and test expectations; existing atomic activation mechanism unchanged',pressure='NOT_APPLICABLE: no new workload or resource path',pausedResumed=False,allProviderReleaseClaimed=False))
print(json.dumps(dict(activated=True,models=6,formalRoutes=6,ordinaryExitCode=0,inventories=inventories,historyUnchanged=len(history),authorityUnchanged=len(currentAuthority),newVendorCalls=0,completion=str((B/'completion.json').relative_to(R)))))
