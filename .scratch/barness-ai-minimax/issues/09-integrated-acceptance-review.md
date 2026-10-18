# 09 — 完整验收、双轴审查与交付判定

Status: claimed
Type: task
Updated: 2026-10-18
Blocked by: 08, 10

## Context

[交付规格](../spec.md)、[任务图](../map.md)。追溯R01–R28、A01–A20、US01–US80。

## Scope and acceptance

执行普通/race公共E2E、两种vet、实际冻结差分、全部原有与新增压力/业务效果、逐包脱敏和完整回滚；汇总US/R/A→实际有限case追溯。保持旧十三组合/十四自身身份的完整当前证据，加入MiniMax三组合/三M3正式身份与M3.1三候选离线状态。依据不可变来源作最终门禁，不重写捕获适用性。

全部实现后按code-review技能对固定基点ecfd4b4运行Standards与Spec并行审查，单一实现subagent修复全部问题，再完成必要重验。只有三份M3正式自身报告、全部门禁和M3.1三协议离线约束完整时才关闭spec与全部票；缺口保留真实状态和可复现指针。清理各实现worktree，报告集成分支。

## Comments

2026-10-18：先行整理文档与历史证据；最终验收仍被10阻塞，未关闭本票或spec。
[整理检查点](../evidence/ticket-09/README.md)将R01–R28、A01–A20、US01–US80映射到已有有限case，保留原始来源、状态和重放命令；03的33次受控red/green原始观察已归档。此整理不产生运行时或正式支持PASS。

当前为[PARTIAL交付](../delivery-status.md)。完整MiniMax自身离线普通/race验证与部分旧case定向重验可独立通过，但不替代三M3自身正式报告、旧十四身份/四效果续期、完整发布门禁与最终整体验收。不得关闭本票或将原ordinary的52个FAIL和中断退出重写为PASS。

实际M3 Chat研究已执行两次调用、两次Attempt，均为HTTP401认证失败；工具续接明确NOT_RUN，不能将已执行研究改写成NOT_RUN。该中国区失败作为历史保留；用户已改用国际站，先由11完成切换后继续10的自身真实验收；M3.1缺少获准M Plan资格，三个线上候选保持NOT_RUN。实际失败捕获与预算由10持有，09只引用其不可变来源。

2026-10-18当前：全部批准实际收集与原始完整门禁均已结束。其余15项原授权验证已完成168/168（两项DeepSeek PASS、13项FAIL）；3d3a1fb的完整普通/race各12385 PASS/6 FAIL、审计0，CLI随后读取bundle失败且没有发布报告。全部实际进程已关闭，旧能力本次续期累计300/300，17身份矩阵与完整交付仍未通过；消费者修复及其适用范围见交付状态。 原始双套每套869个PIDiff及29个Pressure标记case均PASS，不能因此推出发布PASS；内部各命令数字退出码未留存，保持未知。后续 release-only 读取修复及两轴复审与原始3d3全量结果分别保留。见[原始门禁](../evidence/ticket-09/international-final-actual/README.md)、[审查索引](../evidence/ticket-09/final-reviews/README.md)及[当前交付状态](../delivery-status.md)。本票保持claimed，不关闭spec。


2026-10-18区域路由轮次后：来源8fbd170的新完整公共CLI已实际完成双vet、ordinary/race，四项数字退出0/0/1/1；两套各12439PASS/8FAIL/1NOT_RUN、审计0，868个真实PIDiff及29个Pressure case均PASS。已生成pass=false报告；不改写3d3a1fb原来无报告的历史。其余15项与GLM Chat/Messages本轮PASS，GLM Responses仍真实FAIL，17矩阵记录16PASS/1FAIL。新增245/245、累计545/545全部关闭。离线新缺口由[14](14-release-preflight-gaps.md)补正，最终完整认证仍未通过。本票保持claimed。
