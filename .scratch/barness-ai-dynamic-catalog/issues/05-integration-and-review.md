# 全规格集成、双轴审查及普通验证

Type: task
Status: claimed
Blocked by: 01, 02, 03, 04

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责集成收尾与公共端到端缺口、用户文档、唯一当前状态和可重复证据。以初始提交为固定点按 code-review 技能双轴审查，统一修复，再冻结普通来源并执行完整受影响普通包/构建/vet，记录真实 SHA、哈希和未执行阶段。不得把受控通过写成厂商资格或完整发布通过。race/压力适用时仅列具体人工选择请求，未授权不执行。

## Acceptance and verification

T01–T13；普通证据、全部必需实际场景及 audit，状态/索引一致。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

2026-11-01：开始来源适用性与文档子范围。源代码基准 0e9ef483cb132c681e7a09b22e230190b19b9511；03/04 后续仍在实施，完整普通与双轴审查未执行。验证边界为现有公共 releasegate 资格 CLI、控制捕获及来源拒绝场景。当前状态由主 agent 统一维护。

## Checkpoint 1 — 来源适用性与文档

2026-11-01：有限来源审阅见 [source-review.md](../source-review.md)，不可变命令/退出/哈希索引见 [integration-verification-v1.json](../integration-verification-v1.json)。四项精确来源迁移仅适用 xAI Images；原审阅条目、原始测量、pin 与支持矩阵未改写。运行数据新 JSON 按实际型号投影核对，新增验证/helper 仍属于产品来源，未来编辑、改原来源或错路线继续拒绝。

- RED：100651 原历史原件 CLI 失败；100753 新精确迁移/分类失败；100826 首次修复仍暴露 classifier-kind 提取的源码漂移。全部实际 exit1 原件保留。
- GREEN：100924 定向通过；101100 扩展 Evidence/XAIImages 组 PASS/exit0，149 实际 Go 测试、235 原生 PASS、audit0；101407 安全失败报告修订后的聚焦组 PASS/exit0。完整路径/argv/manifest 与 audit 哈希见索引，不以父测试或零匹配代替。
- 文档：新增本地加载/候选检查应用/字段归属/价格/版本与生成指南，修正现有 README/contract 的旧字面量维护入口和默认快照指向。六个相关文档的本地链接核对无缺失。
- 未完成：03 最终 Provider 导出命令的指南补齐、03/04 最终代码合并后的完整普通/编译/vet、主 agent 双轴审查和最终状态/发布只读索引。保持 claimed，不声明 release.pass=true。

## Checkpoint 2 — 最终 CLI 指南准备

2026-11-01：已按 03 已提交实现补齐 `cost-update`、按 Provider 的 `rollback`、`--source-kind official`、冻结本地定义/决定及 `export --review-dir` 完整审阅包；按 04 补齐 `candidate --check --repo` 的严格只读状态查询。指南中的 28 个参数名与实际 CLI/生成器声明核对一致，六个相关文档本地链接无缺失，最大 476 行。尚未执行最终普通阶段；等待 03 合并与主 agent 明确冻结信号，再检查最终帮助/完整普通/编译/vet。当前状态仍只由主 agent 维护。

## Checkpoint 3 — 本地定义的实际能力与诊断

2026-11-01：主 agent 在最终冻结前确认两项具体缺口。分类器结构自洽但超出固定协议子集时现报告 unsupported_definition；非法限额仍为 invalid_definition。Google 声明未实现的 mask/透明背景/custom size/fidelity/size/quality，以及 OpenAI 未消费的 aspect/imageSize 或实际选项 guard 不允许的数量/quality/fidelity 声明，不再伪装加载成功。参考数复用既有协议 guard 常量。

新增公共 Load/Hash/NewClient 红绿场景，完整旧 Hash/NewClient 接受/错误文本保持；原 guard 条件未变、无能力扩展或新厂商规则。追加精确来源审阅只在原 xAI Images 范围生效，旧条目/旧文档保持；见 [v2](../source-review-v2.md)、[v3](../source-review-v3.md)。完整命令、失败/通过、实际退出及 manifest/audit 指针见 [integration-verification-v2.json](../integration-verification-v2.json)。

分类范围 104438 PASS/exit0，最终图片/分类/加载/证据及全 118 合约扩展范围 105117 PASS/exit0；保留 104121、104252、104838 三次独立失败。仍待 03 最终合并后的完整普通/build/vet 和主 agent 双轴审查；不声明完整发布通过。
