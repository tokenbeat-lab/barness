#!/usr/bin/env python3
"""Render a read-only, reproducible analysis of retained FAILs and session timing."""
import collections,csv,datetime,html,json,pathlib,re,tarfile
ROOT=pathlib.Path(__file__).resolve().parents[3]; OUT=pathlib.Path(__file__).resolve().parent
BASE=ROOT/'.scratch/barness-ai-minimax/evidence'
rows=json.loads((OUT/'raw-failed-cases.json').read_text())
full=json.loads((ROOT/'.evidence/minimax-global-final-keyless/gate/offline/20261018T045647.697909000Z/manifest.json').read_text())
final_status={c['id']:c['status'] for c in full['cases']}

def classify(r):
 path=r['evidence'][0];case=r['case'];test=r['test'] or '';detail=' '.join(x['name']+' '+x['detail'] for x in r['false_assertions'])
 ancestors=r.get('ancestor_cases',[])
 if ancestors and ancestors[-1]['status']=='PASS':
  return '受控负例（父测试PASS）','负例子进程按设计产生 FAIL；外层断言已确认拒绝行为正确。具体拒绝条件见内容及父用例。','不属于当前产品缺陷'
 if 'legacy-paid-3d3.tar.gz' in path:
  if any(s in path for s in ['typesafe-classifier','protocol114']):cause='上游 TypeSafe HTTP 451，明确返回 Typesafe is not available in your region；归一化 invalid_request 不代表本地参数错误。'
  elif 'anthropic-messages' in path:cause='上游 HTTP 403，error.type=forbidden / Request not allowed；不能据此断言 Key 无效。'
  else:cause='请求在取得连接/响应前超时；OpenAI/Google 通用路径的 ConnectTimeout 为10或15秒，没有可供协议解析的 HTTP 响应。无法从捕获进一步区分 DNS/TCP/TLS。'
  return '当前外部验收失败',cause,'保留FAIL；未重跑'
 if 'glm-corrective-actual-5bf2c16.tar.gz' in path:
  return '当前外部验收失败','glm-5.2/native-history 的完整 HTTP200/SSE 仅有5个字符的 reasoning，没有 message/output_text；适配器没有丢失正文。GLM11-sources 是同次父测试失败的汇总。','保留FAIL；未开启R2'
 if case.startswith(('GLM12-formal-import-','GLM12-reported-known-usage-')):
  return '来源/资格门禁失败','GLM 历史正式证据与当前 execution/consumer/importer source map 不一致；当前 R1 Responses 又有真实 FAIL，不能整体刷新支持矩阵。','仍阻塞整体验收'
 if 'whole-ordinary-failed-9481a6a' in path:
  if test.startswith(('TestReleaseRollbackArtifacts','TestReleaseRevocationSemanticIntegrity')):
   cause='测试准备要求的 runtime-evidence/runtime.zip 缺失。部分记录无断言/stderr；依据公共 helper 的执行顺序和缺失文件归因，非声称持有原始 t.Fatal。'
  elif test.startswith('TestJevBaselineEvidence'):cause='jev-baseline/baseline.zip 缺失/无效；有空断言记录，归因依据文件状态与准备路径。'
  elif test.startswith('TestDecisionsEvidence'):cause='text-choice-evidence/text-choice.zip 缺失/无效；有空断言记录，归因依据文件状态与准备路径。'
  elif test.startswith('TestJevFormalMatrixHistory'):cause='MiniMax capabilities 的空切片 [] 在重编码时省略，解码成 nil；reflect.DeepEqual 判为历史被修改。'
  else:cause='见保留的具体断言；该52项历史全量运行的归因详见 whole-ordinary-timing-investigation.md。'
  return '历史本地失败',cause,'后续同用例PASS（最终A）' if final_status.get(case)=='PASS' else '历史记录未改写'
 if 'preactivation-activated-whole' in path or 'preactivation-business-red' in path:
  if test.startswith('TestGLMFullRollback'):cause='私有测试 checkout 的 frozen runtime 目录符号链接被 Git 当作未跟踪文件；属于测试准备问题。'
  else:cause='业务配置身份哈希耦合了两项已独立验证的 catalog pin；激活 MiniMax 后误判既有业务配置变更。df342b9 解除该耦合。'
  return '历史本地失败',cause,'已修复；相关定向ordinary/race共24PASS'
 if 'actual-auth-failure' in path:
  return '历史外部失败','中国区端点真实返回 HTTP401 authorized_error；仅能确认该 Key/账户无法访问该端点。后按用户要求切换国际站。','中国区历史FAIL；国际站M3自身正式三路PASS'
 if 'international-complete-candidates' in path and case.startswith('MM07-'):
  cause='Chat 同时返回内容相等的两种 reasoning 镜像格式，旧解码器误判为冲突。' if 'full-stream' in case or 'tool-call' in case else 'Responses 原生历史把仅适用于输出的元数据带入输入，SDK encoder 在网络调用前拒绝。'
  return '历史适配失败',cause,'8d56410已修复；后续候选/正式PASS'
 if 'international-convergence' in path:
  if 'chat-mirror' in case:cause='真实 Chat 镜像格式需校验一致后仅输出一次，同时保留原生历史；旧处理拒绝一致镜像或丢回放字段。'
  elif 'encrypted-type' in case:cause='Responses 对未证实的 encrypted_content 非字符串形状未严格拒绝。'
  else:cause='Responses 回放误带输出专属字段，导致 SDK 输入编码失败；需在边界保留原始状态并投影合法输入。'
  return '开发期RED/历史适配失败',cause,'8d56410已修复'
 if 'international-formal-glm' in path:
  return '历史实际验收失败','glm-5.3/control-low 耗尽4096输出预算而没有最终文本；wire捕获在512KiB被截断，旧事件保留格式又超过约16MiB，无法形成合格原始证据。父 sources 随子测试失败。','捕获/保留实现已修复；本轮原FAIL不改写'
 if 'release-consumer' in path:
  if 'post-load' in case:cause='校验后重新按未经固定的路径打开输入，允许替换/符号链接或文件内容更换；修复为rooted且按原始hash复核的读取。'
  elif 'root-selection' in case:cause='初次选择根目录时未固定并比较打开前后的inode，符号链接/替换选择可被接受；后加根身份校验。'
  elif 'inventory-bound' in case:cause='原消费端对递归证据目录的文件数/单文件/总字节上限未完整约束。'
  elif 'directory-isolation' in case:cause='原消费端只处理平面目录，递归清单及case/祖先目录符号链接边界不完整。'
  elif 'catalog-release' in case:
   if 'child failed=true' in detail:cause='验证用子进程缺少既有 frozen node_modules 运行时准备。'
   else:cause='Catalog 差分生产者遗漏 Go SDK/toolchain provenance；纯元数据控制又错误使用 PIDIFF 命名空间。首版测试另有把零值 Request 当 JSON null 的错误断言。'
  else:cause='生产者保留空 effort 的尾随斜线case ID，物理目录被 filepath.Join 规范化，消费端拒绝原始dir；另有递归清单规则不一致。'
  return '开发期RED/历史本地失败',cause,'最终定向ordinary/race和双vet PASS；旧A不改写'
 if 'ticket-12' in path or 'glm-corrective-q-' in path or 'corrective-approved-prepay-' in path:
  if 'home' in case:cause='回放诊断把临时 HOME 的绝对路径写入证据，触发redaction审计；修复诊断路径脱敏。'
  elif 'event-budget-known-usage' in case:cause='事件证据溢出/拒绝时覆盖了已获得的Attempt usage，导致已知用量丢失；补正保留已知usage与消费。'
  elif 'capture' in case:cause='补正前捕获/事件证据的512KiB、事件容积或编码格式边界不足；失败优先测试复现合法大流及溢出场景。'
  else:cause='GLM补正的合成故障/来源拒绝复现；具体边界见失败断言。该记录本身不是一次真实厂商调用。'
  return '开发期RED/受控复现',cause,'后续补正离线验证PASS；原始记录保留'
 if 'ticket-11' in path:
  return '开发期RED','切换国际站后，端点/Provider身份、USD价格、历史消费槽位或oracle metadata尚沿用中国区规则。具体缺口见断言；国际站实现和后续验证已收敛。','后续国际站定向验证PASS'
 topic={'ticket-01':'默认thinking标签、采样默认值与配置边界','ticket-02':'MiniMax Chat 的请求字段、最终守卫、流终态、reasoning或Native State回放','ticket-03':'MiniMax Responses 的字段子集、原生输入和SSE边界','ticket-04':'MiniMax Messages 的thinking、工具/预算守卫、签名、缓存和原生历史','ticket-05':'共享错误分类、取消、永久quota、回调panic、Observer或资源释放','ticket-06':'冻结pi目录、差分分类/ledger及扩展元数据','ticket-08':'Host配置、离线回滚与回滚凭证'}
 ticket=next((x for x in topic if x in path),'')
 if ticket:
  if any(x in path for x in ['assembly-failure','attribution-failure','expectation-failure']):cause='测试装配或预期尚未正确匹配公共接口；并非厂商故障。具体错误见断言。'
  else:cause='失败优先开发时，'+topic[ticket]+'尚未满足断言所列协议义务；不是最终实现回归。'
  if not r['false_assertions']:cause+=' 原始报告无false断言正文，无法仅凭该报告精确到更细一行代码。'
  return '开发期RED/历史本地失败',cause,'后续同用例PASS（最终A）' if final_status.get(case)=='PASS' else '后续相关定向验证PASS；原记录保留'
 return '历史失败/需结合上下文','仅能确认下列断言失败；没有充分证据进一步精确归因，不猜测。','原记录保留'

for r in rows:
 r['kind']='case';r['classification'],r['root_cause'],r['disposition']=classify(r)
 r['content']='\n'.join(a['name']+': '+a['detail'] for a in r['false_assertions']) or 'manifest记录FAIL；未保留false断言正文（可能在准备阶段或父测试因子测试失败）。'
# The two full runs have already extracted immutable six-failure lists; avoid expanding their 2.6GB raw trees.
idx=json.loads((BASE/'ticket-09/international-final-actual/A-final-keyless-3d3.json').read_text())
for f in idx['failedAssertions']:
 p=BASE/'ticket-09/international-final-actual/A-final-keyless-3d3-metadata.tar.gz'
 with tarfile.open(p) as t:d=json.load(t.extractfile(f['file']))
 r={'kind':'case','case':f['caseId'],'test':d.get('test'),'started_at':'2026-10-18T04:56:47.697909Z' if f['suite']=='offline' else '2026-10-18T05:06:44.512932Z','source':idx['sourceCommit'],'evidence':[str(p.relative_to(ROOT))+'::'+f['file']],'content':'\n'.join(x['name']+': '+x.get('detail','') for x in d.get('assertions',[]) if x.get('pass') is False),'classification':'来源/资格门禁失败','root_cause':'当前 GLM consumer/importer/execution source 与已存正式证据来源不一致；R1 Responses也有真实FAIL，无法整体续期。','disposition':'最终全量普通/race各保留6FAIL；未取得完整发布认证'}
 rows.append(r)
# Text/visual evaluation failures live in sample reports, rather than evidence Case rows.
p=BASE/'ticket-10/international-legacy-final/legacy-paid-3d3.tar.gz'
with tarfile.open(p) as t:
 for m in t:
  if not m.isfile() or not (m.name.endswith('evaluation-report.json') or re.search(r'captures/visual-[^/]+/report.json$',m.name)):continue
  d=json.load(t.extractfile(m));route=m.name.split('/')[1]
  for s in d.get('samples',[]):
   if s.get('status') not in ['failed','cancelled']:continue
   md=s.get('metadata',s.get('result',{}).get('metadata',{}));attempts=md.get('attempts',[])
   cause='TypeSafe实际HTTP451明确地区拒绝：Typesafe is not available in your region。' if route.endswith('jev') else '约5秒请求期限/连接策略在收到HTTP响应前到期；无法据此区分连接watchdog与TLS握手细因。'
   rows.append({'kind':'evaluation_sample','case':route+'/'+s['id'],'test':route,'started_at':s.get('startedAt'),'source':'3d3a1fb28e143a2c29d684f78edcad94c0217243','evidence':[str(p.relative_to(ROOT))+'::'+m.name],'classification':'当前外部验收失败','content':json.dumps({'sample':s['id'],'status':s['status'],'code':s.get('code',s.get('error',{}).get('code')),'phase':s.get('phase',s.get('error',{}).get('phase')),'HTTPStatus':[a.get('httpStatus') for a in attempts]},ensure_ascii=False),'root_cause':cause,'disposition':'失败样本原样保留；无质量结果；未重跑'})

def extra(case,content,cause,state,evidence,category='汇总/门禁失败',kind='gate'):
 rows.append({'kind':kind,'case':case,'test':case,'started_at':None,'source':None,'evidence':[evidence],'classification':category,'content':content,'root_cause':cause,'disposition':state})
# Actual original gate orchestration failures, preserving their exact log meanings.
for p in sorted((ROOT/'.evidence/minimax-global-final-keyless').glob('*-command.json')):
 d=json.loads(p.read_text())
 if d.get('exitCode')!=1:continue
 name=d.get('name',p.stem);log=p.with_name(p.name.replace('-command.json','.log'));body=log.read_text() if log.exists() else ''
 if name=='actual-releasegate':
  body='releasegate: release: invalid or duplicate evidence case directory; exit status 2; outer go run exit 1; no release-report'
  cause='8个空effort测试ID保留一个尾随/，物理目录已被规范化；证据生产者和消费端目录规则不一致。';state='消费端已修复；旧全量结果不改写，仍无完整发布认证'
 elif name=='derive-jev-support':cause='Jev协议首调用HTTP451，原始formal protocol不完整/FAIL，无法派生正式支持投影。';state='派生正确拒绝；support目录未生成'
 elif name=='verify-jev-support':cause='上一步派生失败导致support/live-report.json不存在。';state='派生失败的连锁结果'
 elif name.startswith('verify-visual-'):cause='原始visual样本全部在HTTP451或无响应期限边界失败，无法满足require-live完整真实效果验收。';state='正确保留FAIL'
 else:cause='Jev正式support投影不存在，17组合临时矩阵导入失败。';state='默认矩阵未合并；来源未伪造'
 extra('original-finalizer/'+name,body[-6000:],cause,state,str(p.relative_to(ROOT)))
# Diagnostic false gates are an interpretation of old inputs, never a new run.
for z in json.loads((OUT/'gate-reports.json').read_text()):
 for g in z['report']['gates']:
  if g['pass']:continue
  for number,problem in enumerate(g.get('problems',[]),1):
   if 'not run' in problem.lower() or 'is NOT_RUN' in problem:continue
   if g['id']=='differential':cause='旧catalog差分缺Go provenance；纯元数据case误用PIDIFF前缀。未来生产者已修复；旧记录保持原样。'
   elif g['id']=='live':cause='旧实际FAIL与未更新默认矩阵/当前完整来源要求不符；其中DeepSeek或MiniMax自身PASS不自动等于全局矩阵通过。'
   elif g['id']=='redaction-audit':cause='缺失Jev support目录/manifest导致无法完成bundle audit；不是发现Key或内容泄露。'
   elif g['id']=='traceability':cause='底层GLM来源/资格及真实业务效果FAIL向需求追踪项传播；追踪FAIL不是新增独立根因。'
   elif g['id']=='p0-offline':cause='旧GLM六项来源失败的门禁汇总。'
   else:cause='旧真实效果FAIL、来源失败或差分元数据缺陷的门禁汇总。'
   extra('diagnostic/'+g['id']+'/'+str(number),problem,cause,'adopted historical inputs，commands=null；诊断pass=false，非全量重跑',z['archive']+'::'+z['member'])
extra('baseline/redaction-home','go test ./...: package ai/e2e FAIL; audit: P0-GLM13-full-rollback/commands.json [environment] the home directory appears 2 time(s)','初始基线的回滚命令证据包含HOME绝对路径；测试体PASS，但审计使包返回失败。','既有基线问题；非MiniMax业务适配失败','/tmp/barness-minimax-baseline.log','历史审计失败')
extra('corrective-Q-9dc/redaction-home','335PASS/62NOT_RUN，audit findings=2，Go exit1','新capture replay诊断错误把临时HOME路径写入证据。','ecce867修复；5bf最终336PASS/62NOT_RUN/audit0',str((BASE/'ticket-12/glm-corrective-q-9dc-audit-failure.json').relative_to(ROOT)),'历史审计失败')
for case,content,cause in [
 ('portable-verifier/Q-old','96MiB有限gzip膨胀复现：旧版peak RSS约216MiB','gzip.decompress在32MiB扩展上限检查之前完整分配；Q还在解压后才核验compressed hash。'),
 ('portable-verifier/actual-old','96MiB有限gzip膨胀复现：旧版peak RSS约260MiB','同样先完整gzip解压、后查长度；应先bounded read/hash，再最多解压32MiB+1。'),
 ('portable-verifier/calibration','初始actual用例128MiB测试RSS阈值失败；修复后基线仍约134MiB','测试阈值未覆盖实际验证器正常胶囊解析基线；最终测试先登记192MiB阈值，生产8/32MiB上限不变。'),
 ('release-consumer/catalog-compile-preparation','catalog-red-compile-preparation: go test exit1','新失败优先测试的编译准备错误；未进入厂商调用或正式验收。'),
 ('release-consumer/iterator-red','TestInventoryIteratorRejectsDuplicateNames: go test exit1','streamed ReadDir计数没有seen-set，重复目录项可绕过唯一性计数；09fea1b修复。'),
 ('release-consumer/diagnostic-output-audit','audit CLI REFUSED: output directory is not an evidence bundle (missing manifest)','把普通诊断输出目录当完整bundle调用audit CLI；后用有限file overlay审计8文件0findings，保留原拒绝。')]:
 extra(case,content,cause,'本地已修复/校正；原始失败保留',str((BASE/'ticket-12/retention-bounds/failure-inventory.md').relative_to(ROOT)) if case.startswith('portable') else str((BASE/'ticket-09/release-consumer/commands-summary.json').relative_to(ROOT)),'开发期RED/历史本地失败','operational')
rows.sort(key=lambda r:(r.get('started_at') or 'Z',r['case']))
for n,r in enumerate(rows,1):r['number']=n
(OUT/'all-failures.json').write_text(json.dumps({'scope':'本次实施时间范围内保留的118份归档中的FAIL case，另补最终A两套六项失败、72条业务样本、实际finalizer与诊断门禁、已知审计/工具失败。按原始manifest语义身份去重重复归档；不同运行和父子记录保留。负例子进程FAIL且父测试PASS单独标注；NOT_RUN和成功拒绝的PASS不计入FAIL。不是独立bug总数。','limitations':['这是保留证据的穷举清单，不把未留存stderr的准备失败编造为确定原文。','测试开发期RED的根因分辨到协议/模块义务；不能仅从断言为每次中间修改推断唯一代码行。','源代码和日志未重新执行；未访问.env或进行厂商调用。'],'counts':dict(collections.Counter(r['classification'] for r in rows)),'rows':rows},ensure_ascii=False,indent=2)+'\n')
fields=['number','kind','classification','case','test','started_at','source','content','root_cause','disposition','evidence']
with (OUT/'all-failures.csv').open('w',encoding='utf-8-sig',newline='') as f:
 w=csv.DictWriter(f,fieldnames=fields,extrasaction='ignore');w.writeheader()
 for r in rows:w.writerow({**r,'evidence':'\n'.join(r['evidence'])})
summary={'raw_archived_case_failures':2234,'raw_archived_unique_case_ids':718,'expected_child_failures_inside_passing_parent':sum(r['classification']=='受控负例（父测试PASS）' for r in rows),'all_inventory_rows_including_gates_samples_audits':len(rows),'classification_counts':dict(collections.Counter(r['classification'] for r in rows)),'final_whole_ordinary':{'PASS':12385,'FAIL':6},'final_whole_race':{'PASS':12385,'FAIL':6},'legacy_routes':{'PASS':2,'FAIL':13}}
(OUT/'summary.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
# Standalone local report: compact rows with expandable exact assertions, filterable without network.
data=json.dumps([{k:r.get(k) for k in fields} for r in rows],ensure_ascii=False).replace('</','<\\/')
page='''<!doctype html><html lang="zh-CN"><meta charset="utf-8"><title>MiniMax 会话耗时与 FAIL 清单</title>
<style>body{font:15px/1.65 system-ui,sans-serif;margin:32px;color:#172132;background:#f5f7fb}main{max-width:1500px;margin:auto}h1{font-size:28px}table{border-collapse:collapse;width:100%;background:white}td,th{padding:12px;border:1px solid #dbe1e9;text-align:left;vertical-align:top}th{background:#eaf0f7;position:sticky;top:0}input,select{font:inherit;padding:9px;margin:5px;border:1px solid #b7c4d8;border-radius:6px}input{width:min(650px,80%)}pre{white-space:pre-wrap;word-break:break-word;max-height:420px;overflow:auto;font:12px/1.5 ui-monospace,monospace}small{color:#526277}.box{background:white;padding:20px;border-radius:10px;margin:20px 0}td:nth-child(1){width:80px}td:nth-child(2){width:280px}td:nth-child(3){width:38%}summary{cursor:pointer}strong{color:#173d7c}</style><main>
<h1>会话耗时与 FAIL 清单</h1><p>统计范围：2026-10-18 06:02:53—14:44:33（上海时间），排除本次分析。总墙钟 <strong>8小时41分40秒</strong>；三个执行轮次合计约8小时31分56秒；轮次间间隔合计9分44秒。并行工作不累加为墙钟。</p>
<table><tr><th>阶段</th><th>时间</th><th>耗时</th><th>主要内容</th></tr><tr><td>初始中国区实施</td><td>06:02:53—08:50:08</td><td>2:47:15</td><td>三协议、默认行为、隔离/资源/错误/回滚、冻结pi与受控验收、China401及历史门禁准备失败。</td></tr><tr><td>切换国际站后的剩余实施与验收</td><td>08:52:49—12:12:10</td><td><strong>3:19:21（最长轮次）</strong></td><td>端点/身份/USD/历史预算切换，真实wire修正，候选及正式M3，GLM原轮失败后的捕获、用量、脱敏修正和补正预验。</td></tr><tr><td>获批R1到最终交付</td><td>12:19:13—14:44:33</td><td>2:25:20</td><td>GLM R1、剩余15路线、最终全量门禁，以及之后消费端/差分生产者修复、复审和归档。</td></tr></table>
<p><strong>归因更正：</strong>3小时19分21秒是切换国际站后整轮剩余工作的墙钟，不是端点接入耗时。配置迁移及来源冻结在09:29:34前完成（约36分45秒，且包含身份、价格、账本、文档和测试工作）；之后约2小时42分36秒主要处理既有适配缺陷、目录激活和GLM验收准备。没有证据将这些协议缺陷归因为中国区/国际站差异。</p>
<div class="box"><strong>最长单次执行：最终全量发布门禁47分05秒（外层命令47分10秒）</strong><p>ordinary ai/e2e 588.256秒，race ai/e2e 2219.942秒（37分钟，约占门禁79%）。每套12391 cases，约1.3GB、131992文件；串行证据写入/脱敏审计及race插桩增加代价。无性能profile，不能进一步给出CPU、I/O、子进程的精确份额。旧普通全量的manifest已完成，剩余tail有证据指向审计，未发现E03/decimal挂死。</p><p>15路线串行真实验收19分33秒；原GLM三个命令合计5分34秒，R1合计4分45秒；M3三份正式调用命令合计约53秒。不能把数小时主要归因于模型推理。13:43:47门禁结束到14:44:33交付另有约61分钟，主要是证据消费、来源/差分问题修复与归档复审。</p></div>
<p>下表按保留证据逐项列出。<strong>FAIL记录数不是bug数</strong>：包括开发期RED、同一用例在不同轮次的复现、父子汇总及门禁传播。1441条记录是父测试PASS的受控负例。NOT_RUN不算FAIL；DeepSeek预期HTTP400控制仍为PASS。</p>
<p>完整数据：<a href="all-failures.csv">CSV</a> · <a href="all-failures.json">JSON</a> · <a href="timing.json">时序元数据</a> · <a href="scan-index.json">118份归档扫描范围</a> · <a href="nonzero-commands.json">主线程49次命令非零退出</a></p>
<input id="query" placeholder="搜索用例、断言、根因或证据路径"><select id="category"><option value="">所有类别</option></select><label><input id="controls" type="checkbox" style="width:auto">显示父测试PASS的受控负例</label><p id="count"></p>
<table><thead><tr><th>编号/类别</th><th>用例/状态</th><th>失败内容（展开原始断言）</th><th>根因</th></tr></thead><tbody id="body"></tbody></table><p><small>只读取既有证据，不重跑、不读取凭据。空断言的历史准备失败与诊断所能确定的边界均明确保留。源码未改动。</small></p></main><script>const rows=DATA;const q=document.querySelector('#query'),c=document.querySelector('#category'),controls=document.querySelector('#controls'),body=document.querySelector('#body');const esc=s=>String(s??'').replace(/[&<>"']/g,ch=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));for(const k of [...new Set(rows.map(r=>r.classification))].sort()){const o=document.createElement('option');o.value=k;o.textContent=k;c.append(o)}function render(){const term=q.value.toLowerCase();const selected=rows.filter(r=>(controls.checked||r.classification!=='受控负例（父测试PASS）')&&(!c.value||r.classification===c.value)&&JSON.stringify(r).toLowerCase().includes(term));document.querySelector('#count').textContent=`显示 ${selected.length} / ${rows.length} 条。为保持滚动流畅，最多渲染前300条；输入关键词可缩小范围，CSV/JSON包含全部。`;body.innerHTML=selected.slice(0,300).map(r=>`<tr><td>${r.number}<br><small>${esc(r.classification)}</small></td><td><b>${esc(r.case)}</b><br><small>${esc(r.test)}<br>${esc(r.disposition)}</small></td><td><details><summary>${esc(r.content.slice(0,160))}</summary><pre>${esc(r.content)}</pre><small>${esc(r.evidence.join(' | '))}</small></details></td><td>${esc(r.root_cause)}</td></tr>`).join('')}q.oninput=c.onchange=controls.onchange=render;render();</script></html>'''.replace('DATA',data)
(OUT/'report.html').write_text(page)
print(json.dumps(summary,ensure_ascii=False,indent=2))
