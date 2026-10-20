# 10 最终验收的有限采集与来源发布

Status: claimed
Type: task

## Context

[07 最终验收](07-final-acceptance.md)、[规格 R16 / R18 / R20](../spec.md)；具体待批准的有限计划在根 `.evidence/provider-sites/final-acceptance-20261020/plan/plan.md`。旧操作脚本及其 grant 均为关闭轮次，不能重新执行旧权限或改写旧原件。九身份晋升后目录 Hash 改变，最终需 23 条自身支持和四份效果。

实现可按已约定的固定轮次接口与 09 并行；实际国际 MiniMax 采集仍依赖 09 的原生授权通过，最终执行统一由 07 阻断和调度。

## Required behavior

准备本功能单一、具体的最终轮次操作入口，固定 `provider-sites-final-acceptance-1`，至多 4 个独立研究调用、6 candidate、6 formal、17 旧支持和 4 效果；最多 726 Calls / 762 HTTP Attempts，无自动进程重启、换型号、换站、换 Key 或新轮次。研究使用本轮单独 plan/helper，仅供诊断。授权、源码、工具、二进制及本阶段选定输入在付费前固定并审查；人类预算未批准前不能读秘密或访问厂商。

每个子进程仅注入自己的已确认 standard Key 与显式网络配置，清理其他 Key 与 ambient proxy。变量映射固定为 `OPENAI_KEY / ANTHROPIC_KEY / GEMINI_KEY / DEEPSEEK_KEY / TYPESAFE_KEY / ZHIPU_KEY / ZAI_KEY / MINIMAX_KEY / MINIMAX_CN_KEY`，不扫描、猜测或尝试备用配置。保留真实 argv、数字退出、起止时间、原生输出、来源/二进制/plan/grant/原件 Hash 及各原生消费者结论；失败阻断依赖阶段并保持已消费预算，不自动付费重跑。每阶段可恢复查询已完成事实，但不得重新启动已 claim 的进程。

提供明确的剩余美元预留与检查，区分保守估算、缺用量及最终账单；不能声称旧通用图片/分类器/效果程序具备不存在的 token 上限。已知率/模态价格、输入/输出/问题/图片/状态固定上限必须逐类有根据；未能建立安全的预发送上限则停止该部分并报告，而非隐含无限账单授权。US$30 只是当前待批准总额，实际额度来自本轮人类原文。

GLM 新原件发布只扩展当前原生消费者要求的固定新轮次，不修改旧圆次或其授权。使用新 canonical launcher 收据创建新不可变 raw archive 与当前 selectors/views；全部旧 ZIP、捕获、账本与收据保持。不能只刷新 Hash 或缩小执行来源清单。最终矩阵采用公共 supportmatrix importer，必须验证全部 23 条当前自身报告；旧只更新七条、保留十条的 17 行 finalizer 不适用。

## Ownership / validation

负责本功能 `.scratch/barness-provider-sites/tools/` 新有限操作入口、必要的旧 `renewal.py / renewal_glm.py` 新轮次发布分支及直接公共 CLI / 进程 E2E，不负责生产协议、候选/正式目录、09 原生授权实现或旧原件。按验证策略先列失败方式，在公共 CLI/进程接缝逐片 RED→GREEN；只运行定向普通离线 E2E，不读 Key、不发真实网络、不执行 race。

## Acceptance

来源冻结、一次性授权/持久 claim、失败/取消账务、各阶段实际原生消费与 23 行发布可独立审查。执行本身在 07 记录，准备就绪不视为真实资格或发布 PASS。所有代码文件不超过 500 行，避免新的通用工作流框架。

## Answer

已实现固定 `provider-sites-final-acceptance-1` 的单轮入口，33 个完整进程及原始研究 4/4 合计上限 726/762。每阶段冻结已提交源码、实际原生 selector、十个二进制、工具链、固定 argv/plan、双轴审查与用户原文；common-root authority 固定总额与研究事实，独占 route intent、美元预留和消费跨阶段、目录、工作树保留。进入真实子进程前，校验对应 standard 资格与原生权限，国际 M3 再走只读 native preflight；只读选中变量、拒绝重复定义，不向子进程泄露其他 Key 或 ambient proxy。失败/未知保留整笔额度并阻断继续。

研究不会重跑：只导入根本轮原始 approval/index/command/raw pins/common Git claim，固定已耗 4 Calls/4 HTTP 与 USD0.2。transport guard 仅本轮显式启用，在三个 Host 装配中对最终 wire 和已冻结最高价格预发送预留；1h cache write 纳入既有 `2*Input` 语义，图片采用官方 calculator/总输出硬限与明确保守输入 allowance，未知取消不释放。它是 known-rate 运行控制，非最终账单保证。各类依据、USD28.6 建议分配和边界见 [工具说明](../tools/README.md)。

六候选通过后使用已冻结公共晋升 CLI；activation/catalog pins、GLM raw selectors/views 分阶段提交并重新冻结。GLM 新原件发布只接受本轮 canonical launch receipts；旧轮次保持关闭。国际 M3 三协议执行期间 HEAD 冻结，随后才发布派生矩阵。最终入口使用公共 importer 验证 23 条自身输入和四效果，发布 tracked matrix 后才能提交、冻结最终 CompleteSource 并执行普通/READY/人工 race。不会自动运行 race。

公共进程与公共 Client 按失败清单逐片 RED→GREEN；保留实际源码、argv、数字退出、UTC、原生 bundle/audit 和回归记录于 [永久验证索引](../../../.evidence/provider-sites/final-acceptance-20261020/collection-preparation/index.json)。实现准备等待父任务独立审查；此工单未读取真实 Key、未发厂商请求、未执行 race，亦不宣称任何真实资格或整体验收已通过。唯一当前状态仍由 [spec](../spec.md#当前状态) 维护。
