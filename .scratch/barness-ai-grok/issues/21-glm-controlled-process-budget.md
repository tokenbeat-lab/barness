# GLM 受控子进程时限与终止边界

Type: task
Status: resolved

## Contract

[Spec](../spec.md)与[测试SOP](../../../docs/agents/verification.md)。只修复GLM普通受控包装器使用短公共调用context导致完整native用例及收尾被截断的问题；为构建/启动、受控执行、manifest/audit收尾设置清晰有界的层次，并确保取消/超时等待实际子孙进程停止。不得通过预热缓存掩盖已知时间敏感失败、全局增大ctxFor、放宽用例/原件校验或跳过既有场景取绿。保留512上限与原辅助bundle边界。

## Actual failure and feedback

v8新增36授权输入后完整普通在654750/2064/165实际exit1，637.400秒；两根TestGrokHistoryGLMOfflineRecordProcess、TestGrokHistoryGLMControlledFullCapture失败。直接和内嵌GLM包装都在10.01秒失败，ctxFor(t)共用公共调用10秒context，内层go test却写180秒。原生确实启动；内嵌9个PASS后进入捕获校验，直接原生32个assertions全PASS但无manifest/audit，且有写入晚于父包装FAIL。故“全耗于编译”已否定；构建/启动加完整受控工作及收尾总预算不足是首要假设，资源竞争和真实卡住待独立受控区分。此前相同源码129普通完整PASS，直接包装8.69秒只剩约1.3秒余量。

真实普通原件：[FAIL](../../../.evidence/barness-ai-grok/e2e/20261027T012038.355447Z-1hopb8rs/receipt.json)。[详细诊断](../../../.evidence/barness-ai-grok/live-batches/gf8-20261027T012011Z-d0c21717/glm-preflight-readonly-diagnosis/barness-grok-v8-glm-preflight-diagnosis.md)。两个现有公共入口就是反馈缝；先用有界受控时序/延迟或可观察生命周期把截断及子孙释放复现为稳定RED，再改代码，不写镜像式时限数字单元测试。

## Ownership

单一implementer在现有native-proof独立工作树，先同步集成654750edb2e22a8700f56028c061df16487adec1。你不是唯一工作者，不得回退他人编辑；最终同步root集成tip。拥有最小ai/e2e的GLM包装器/离线包装预算与必要受控进程E2E/本地helper、说明和局部证据摘要。经真实外层取消RED后，另明确拥有既有tools/verification/run_process.py的有界进程组关闭及其公开CLI/进程回归；不新增执行器或授权框架。优先复用现有有界进程/进程组能力，避免新框架和推测抽象。文件<500行。

不拥有rootSpec/map/05/21状态、任何真实授权/grant/账本/原件、全局ctxFor、Provider生产/原生harness/库/目录/模型/SDK/预算/重试。既有采集器允许的增量仅限原有cancel/timeout/收尾的有界组关闭，不改变授权、环境、红字脱敏、输出上限或发布语义。不得读.env/真实Key、写真实commonGit授权store或发厂商请求；受控合成grants若需要只用既有私有namespace。不得编译新正式冻结镜像、全完整普通、race/压力/正式消费者或修改/激活v8权限。root负责合并/独立双轴/完整冻结与授权适用性判断。

## Verification

读取diagnosing-bugs技能、GLOSSARY和相关ADR/SOP；先列全失败方式并报告排名假设与RED设计。冷/慢启动和受控执行/收尾总时长应能超过旧10秒仍完整完成；缺场景/无manifest或无audit仍拒绝；外层期限/取消时真正停止并等待整个子树，取消后无持续写入/挂管道，根上下文限制仍生效。优先公开已有Go/collector入口；必要新增受控进程正反例先RED提交，再最小实现。用可控gate/时序，而非靠宿主缓存或无边界重复碰概率。保存实际argv/环境名/数字exit/source/完整bundle/audit/辅助32case，失败原件不覆盖。局部回归覆盖原始两公共失败入口及其真实GLM辅助合同，MM18辅助边界保持，不能把32case展平塞进一个case。尚未有真实能力或付费验收成功声明。

## Comments

- 2026-10-27：本次v8厂商路线0启动、0请求、0真实claim；原17grant原样未用。E2E-only修复预计不改342/407/538/539原生map及8镜像，但会改变全局普通来源，必须在实际diff冻结后独立评估，不自动重绑原v8批准pin，也不预判权限已消耗。当前停在普通阶段，无新增付费动作。

- 2026-10-27：真实外层采集取消RED78e2cf88证明run.py返回后心跳40→100，单纯E2E watchdog不能等待分离的inner PG。root明确扩展ownership至既有run_process.stop有界SIGTERM→组内owner清理/静默→剩余SIGKILL及等待；先RED后改，不动原生源码/镜像/授权。Python公开CLI回归随影响纳入，实际source差异与授权适用性由root独立复核。

## Answer

软件修复已完成，公开过程真实RED/GREEN、v17独立双轴0/0；实际dcf73abd6740d2107352909ba3cd3d07ad6684fb/2068/175完整四包16,038个Go测试、1572主必交/130入口及32+6辅助全部PASS，665.060秒/exit0/audit0。完整Python工具52/52通过，源字节适用性独立核验。[完整核验](../../../.evidence/barness-ai-grok/live-batches/gf8-20261027T012011Z-d0c21717/ordinary-validation-glm21-preconsent.json)及[当前状态](../spec.md#唯一当前状态)。软件关闭不消耗或重绑原未用v8权限，不产生本源真实支持资格；全部失败及原v16P2保留。
