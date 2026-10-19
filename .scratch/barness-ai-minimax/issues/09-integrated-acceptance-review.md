# 09 — 完整验收、双轴审查与交付判定

Status: claimed
Type: task
Updated: 2026-10-19
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

2026-10-18离线合入检查点：已审查修复与有限归档在 `7231fdad` 合入，13按固定执行/证据范围resolved，14/16按离线范围resolved。六项无Key检查与本地GLM91秒边界证明不替代真实支持；旧545/545、GLM Responses FAIL和8fbd全量false保持。[15](15-source-scoped-revalidation.md)准备固定6进程，最终来源/二进制/计划及无Key检查审查就绪后才请求新108/132预算；本票与spec继续开放。


### 2026-10-18 固定六项关闭后的最新交付检查点

2026-10-18 最新固定六项真实验收已按人类新增108calls/132HTTPAttempts批准完成：中国区GLM直连，国际protocol114及两项视觉经127.0.0.1:7890。Chat、Responses及三项效果collection通过，Messages21PASS/1FAIL，新增102/102，连原545/545累计647/647。该固定合同全部关闭，未用6calls/30Attempts不转移，无自动下一轮或新增付费授权。严格GLM续期和新四报告装配拒绝，原selector与17身份矩阵SHA保持；原矩阵16PASS/1FAIL为旧Responses记录，最新Messages原始失败未导入。

来源7931的首次完整无Key公共CLI及post-guards均已结束：八项输入校验0，双vet/ordinary/race数字退出0/0/1/1，CLI1、报告pass=false。普通/race各12451项（12445PASS/6FAIL/0NOT_RUN），审计0、868真实PIDIFF PASS；29个pressure命名用例通过，其中固定规则选择15项。六项失败为GLM三协议正式导入/已报告用量检查；未采用旧bundle或绕过缺失Messages报告。

最终双轴审查发现的Messages最终payload形状及文档问题已由票17离线补正，源码2306917、文档295728a均复审零发现。44公共用例为42非法输入零Provider/Attempt拒绝、两合法业务JSON原值发送；定向普通/race各1663PASS/8NOT_RUN，同8项mandatory显式开关另各8PASS，双vet0。实际测量为ca67+WIP，1027编译/JSON/mod/sum文件等于2306，不声称提交后重跑。修复后M3/GLM/文本/protocol/视觉全部11组严格来源当前相等性不成立；历史证据保持，generic10按既有受影响规则另行核对不变。旧门禁不认证当前源码。

本票保持claimed；M3原来源8900034三份自身正式PASS39/36、本轮未重跑，M3.1线上三候选NOT_RUN0/0仍无获准MPlan资格。spec继续开放，不以受控通过替代修复后当前来源的必交证明。见[当前交付](../delivery-status.md)、[六项原件](../evidence/ticket-15/new-actual/glm/README.md)、[全量门禁](../evidence/ticket-15/new-actual/final-gate/README.md)及[最终审查](../evidence/ticket-09/final-closure/README.md)。

2026-10-19 本轮最新检查点：十一组真实复验及当前来源公开消费者全部PASS，新增189/186、固定216/252关闭。来源01341完整门禁已实际结束，双vet/ordinary/race=0/0/0/1，CLI1、报告false；ordinary12495PASS/audit0，race12494PASS/1FAIL/audit1。唯一VQ12-failure-decisions-plain-deadline及其零字节request.json审计失败正在独立工作树离线修复，原17支持与四效果均PASS、测量前后来源/输入未变。当前root源码尚未修改，十一组证明保留；完整交付不宣称通过。见[交付状态](../delivery-status.md)。

2026-10-19 当前收尾检查点：票21新增36/36已闭合，三项各12/12 PASS；M3/GLM/文本八组完整原生来源集合不变，当前严格资格11/11。来源78843新完整门禁四命令0/0/0/0、CLI0、报告true；普通/race各12495PASS/audit0，最终源码及实际结果两轴均0发现。精确591323候选已实际exit0原子发布，仅Jev一行改变、其余16行对象与时间不变。原01341门禁FAIL及所有原有限授权轮次的失败、消费和未用额度保持；当前仅完整有界归档公开读回、最终文档审查及工单关闭收尾。见[当前交付](../delivery-status.md)及[当前实际结果](../evidence/ticket-21/actual/README.md)。
